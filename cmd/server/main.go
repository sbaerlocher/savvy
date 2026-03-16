// Package main is the entry point for the savvy system server.
package main

import (
	"context"
	"crypto/tls"
	"flag"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"savvy/internal/config"
	"savvy/internal/database"
	"savvy/internal/handlers"
	"savvy/internal/repository"
	"savvy/internal/services"
	"savvy/internal/setup"
	"time"
)

var (
	healthCheck = flag.Bool("health", false, "perform health check and exit")
	healthPort  = flag.String("port", "3000", "server port for health check")
)

func main() {
	flag.Parse()

	// If health check flag is set, perform check and exit
	if *healthCheck {
		os.Exit(performHealthCheck(*healthPort))
	}

	os.Exit(run())
}

// performHealthCheck makes HTTP request to /health endpoint
func performHealthCheck(port string) int {
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS13,
			},
		},
	}

	url := "http://127.0.0.1:" + port + "/health"
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		return 1
	}
	resp, err := client.Do(req) //nolint:gosec // G704: URL is hardcoded to localhost health check
	if err != nil {
		log.Printf("Health check failed: %v", err)
		return 1
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Health check returned status %d", resp.StatusCode) // #nosec G706 -- integer value, no injection risk
		return 1
	}

	return 0
}

func run() int {
	// Load config
	cfg := config.Load()

	// Validate configuration constraints (all environments)
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
		return 1
	}

	// Validate production secrets before starting
	if err := cfg.ValidateProduction(); err != nil {
		log.Fatalf("Production validation failed: %v", err)
		return 1
	}

	// Initialize all dependencies (logging, telemetry, database, migrations, OAuth, etc.)
	shutdown, oauthProvider, err := setup.InitAllDependencies(cfg)
	if err != nil {
		log.Printf("Failed to initialize dependencies: %v", err)
		return 1
	}

	// Initialize rate limiters
	rateLimiters := setup.InitRateLimiters()
	defer setup.Shutdown(shutdown, rateLimiters)

	// Initialize email service
	emailService := setup.InitEmailService(cfg)

	// Initialize service container
	serviceContainer := services.NewContainer(database.DB, emailService)

	// Initialize health check service and handler
	healthService := services.NewHealthCheckService(database.DB, emailService, cfg)
	healthHandler := handlers.NewHealthHandler(healthService)

	// Set email service on notification service for share/transfer emails (with unsubscribe token support)
	serviceContainer.NotificationService.SetEmailService(emailService, serviceContainer.EmailTokenService, cfg.FrontendURL)

	// Initialize push notification service (requires VAPID config from environment)
	if cfg.IsPushEnabled() {
		pushRepo := repository.NewPushSubscriptionRepository(database.DB)
		userRepo := repository.NewUserRepository(database.DB)
		pushService := services.NewPushService(pushRepo, userRepo, cfg.VAPIDPublicKey, cfg.VAPIDPrivateKey, cfg.VAPIDSubject)
		serviceContainer.PushService = pushService
		serviceContainer.NotificationService.SetPushService(pushService)
		slog.Info("Push notifications enabled")
	}

	// Initialize expiry reminder service
	if cfg.EnableExpiryReminders {
		reminderRepo := repository.NewReminderRepository(database.DB)
		notifRepo := repository.NewNotificationRepository(database.DB)
		voucherRepo := repository.NewVoucherRepository(database.DB)
		giftCardRepo := repository.NewGiftCardRepository(database.DB)
		voucherShareRepo := repository.NewVoucherShareRepository(database.DB)
		giftCardShareRepo := repository.NewGiftCardShareRepository(database.DB)
		reminderService := services.NewReminderService(
			reminderRepo, voucherRepo, giftCardRepo,
			voucherShareRepo, giftCardShareRepo,
			notifRepo, serviceContainer.PushService, emailService,
			serviceContainer.EmailTokenService,
			cfg.ReminderDaysBefore,
			cfg.Location,
			cfg.FrontendURL,
		)
		serviceContainer.ReminderService = reminderService
		slog.Info("Expiry reminders enabled", "days_before", cfg.ReminderDaysBefore, "check_time", cfg.ReminderCheckTime, "timezone", cfg.Timezone)
	}

	// Initialize TOTP 2FA service
	if cfg.Is2FAEnabled() {
		totpRepo := repository.NewTOTPRepository(database.DB)
		totpService, err := services.NewTOTPService(totpRepo, cfg.TOTPEncryptionKey, cfg.TOTPIssuer)
		if err != nil {
			log.Printf("Failed to initialize TOTP service: %v", err)
			return 1
		}
		serviceContainer.TOTPService = totpService
		slog.Info("Two-factor authentication enabled", "issuer", cfg.TOTPIssuer)
	}

	// Create and configure Echo server
	serverConfig := &setup.ServerConfig{
		Config:        cfg,
		HealthHandler: healthHandler,
	}
	e := setup.NewEchoServer(serverConfig)

	// Register all routes
	routeConfig := &setup.RouteConfig{
		Echo:             e,
		Config:           cfg,
		ServiceContainer: serviceContainer,
		RateLimiters:     rateLimiters,
		OAuthProvider:    oauthProvider,
		EmailService:     emailService,
		HealthService:    healthService,
	}
	setup.RegisterRoutes(routeConfig)

	// Create a context for the application lifecycle
	appCtx, appCancel := context.WithCancel(context.Background())
	defer appCancel()

	// Start expiry reminder background job (runs daily at configured time)
	if serviceContainer.ReminderService != nil {
		go func() {
			for {
				nextRun := nextScheduledTime(cfg.ReminderCheckTime, cfg.Location)
				waitDuration := time.Until(nextRun)
				slog.Info("Next expiry reminder check scheduled", "at", nextRun.Format(time.RFC3339), "in", waitDuration.Round(time.Minute))

				timer := time.NewTimer(waitDuration)
				select {
				case <-timer.C:
					if err := serviceContainer.ReminderService.CheckAndSendReminders(appCtx); err != nil {
						slog.Error("Expiry reminder check failed", "error", err)
					}
				case <-appCtx.Done():
					timer.Stop()
					return
				}
			}
		}()
		slog.Info("Expiry reminder background job started", "daily_at", cfg.ReminderCheckTime, "timezone", cfg.Timezone)
	}

	// Start session cleanup goroutine (runs every hour)
	if serviceContainer.SessionService != nil {
		go func() {
			ticker := time.NewTicker(1 * time.Hour)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					count, err := serviceContainer.SessionService.CleanupExpired(appCtx)
					if err != nil {
						slog.Error("Session cleanup failed", "error", err)
					} else if count > 0 {
						slog.Info("Expired sessions cleaned up", "count", count)
					}
				case <-appCtx.Done():
					return
				}
			}
		}()
		slog.Info("Session cleanup background job started", "interval", "1h")
	}

	// Start metrics collector goroutine with lifecycle context
	setup.StartMetricsCollector(appCtx, database.DB)

	// Start separate metrics server for Prometheus (industry best practice)
	// Metrics are isolated on a separate port to prevent exposure via ingress
	metricsServer := setup.StartMetricsServer(appCtx, cfg.MetricsPort)
	_ = metricsServer // Graceful shutdown handled in StartMetricsServer via context

	// Start server with graceful shutdown
	slog.Info("Server starting", "port", cfg.ServerPort, "metrics_port", cfg.MetricsPort)

	// Start server in a goroutine
	go func() {
		if err := e.Start(":" + cfg.ServerPort); err != nil {
			slog.Info("Server shutdown", "error", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Println("Shutting down server...")

	// Cancel the application context to stop background goroutines
	appCancel()

	// Shutdown the HTTP server with configurable timeout
	shutdownTimeout := time.Duration(cfg.ShutdownTimeoutSeconds) * time.Second
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()
	if err := e.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error shutting down server: %v", err)
		return 1
	}

	log.Println("Server gracefully stopped")
	return 0
}

// nextScheduledTime calculates the next occurrence of the given HH:MM time in the given timezone.
// If today's target time has already passed, it returns tomorrow's target time.
func nextScheduledTime(checkTime string, loc *time.Location) time.Time {
	now := time.Now().In(loc)

	// Parse HH:MM (already validated in config)
	parsed, _ := time.Parse("15:04", checkTime)
	hour, minute := parsed.Hour(), parsed.Minute()

	// Build today's target time in the configured timezone
	target := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, loc)

	// If today's time has already passed, schedule for tomorrow
	if !target.After(now) {
		target = target.AddDate(0, 0, 1)
	}

	return target
}
