// Package setup contains setup logic for initializing the Echo server.
package setup

import (
	"context"
	"log/slog"
	"savvy/internal/config"
	"savvy/internal/handlers"
	"savvy/internal/metrics"
	"savvy/internal/middleware"
	"time"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"gorm.io/gorm"
)

// ServerConfig holds the configuration for server setup.
type ServerConfig struct {
	Config        *config.Config
	HealthHandler *handlers.HealthHandler
}

// NewEchoServer creates and configures a new Echo server instance.
func NewEchoServer(sc *ServerConfig) *echo.Echo {
	e := echo.New()

	// Configure middleware
	configureMiddleware(e, sc)

	// Configure observability endpoints (public, before auth middleware)
	configureObservabilityEndpoints(e, sc)

	// Configure authentication middleware
	configureAuthMiddleware(e, sc)

	return e
}

// configureMiddleware sets up all middleware for the Echo server.
func configureMiddleware(e *echo.Echo, sc *ServerConfig) {
	cfg := sc.Config

	// Trust proxy headers (X-Forwarded-Proto, X-Forwarded-For, X-Real-IP)
	// IMPORTANT: Only enable in production behind reverse proxy (Traefik)
	if cfg.IsProduction() {
		e.IPExtractor = echo.ExtractIPFromXFFHeader()

		e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				if proto := c.Request().Header.Get("X-Forwarded-Proto"); proto == "https" {
					c.Request().URL.Scheme = "https"
				}
				return next(c)
			}
		})
	}

	// OpenTelemetry Middleware (must be first for proper tracing)
	if cfg.OTelEnabled {
		e.Use(otelecho.Middleware(
			cfg.ServiceName,
			otelecho.WithSkipper(func(c echo.Context) bool {
				// Skip tracing for health checks and metrics endpoints
				path := c.Request().URL.Path
				return path == "/health" || path == "/ready" || path == "/metrics"
			}),
		))
		e.Use(middleware.OTelLogger()) // Add trace IDs to logs
	}

	// Request logging
	e.Use(echomiddleware.RequestLoggerWithConfig(echomiddleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogError:    true,
		LogMethod:   true,
		LogLatency:  true,
		HandleError: true,
		LogValuesFunc: func(_ echo.Context, v echomiddleware.RequestLoggerValues) error {
			attrs := []any{
				"uri", v.URI,
				"method", v.Method,
				"status", v.Status,
				"latency", v.Latency,
			}

			if v.Error != nil {
				attrs = append(attrs, "error", v.Error)
			}

			// Log 5xx responses at ERROR level regardless of whether v.Error is set
			// (handlers that write responses directly via c.String/c.JSON return nil)
			if v.Status >= 500 || v.Error != nil {
				slog.Error("request", attrs...)
			} else {
				slog.Info("request", attrs...)
			}
			return nil
		},
	}))

	// Body size limit to prevent memory exhaustion (4 MB)
	e.Use(echomiddleware.BodyLimit("4M"))

	// Recovery middleware
	e.Use(echomiddleware.Recover())

	// Security headers (CSP, XSS Protection, HSTS, etc.)
	e.Use(middleware.SecurityHeaders(cfg))

	// Prometheus metrics
	e.Use(metrics.Middleware())

	// CSRF Protection is handled per-route-group in routes.go (CSRFApiMiddleware for SPA)
}

// configureObservabilityEndpoints registers health endpoints.
// Note: /metrics is now served on a separate port for security (see StartMetricsServer)
func configureObservabilityEndpoints(e *echo.Echo, sc *ServerConfig) {
	e.GET("/health", sc.HealthHandler.Health)
	e.GET("/ready", sc.HealthHandler.Ready)
}

// configureAuthMiddleware sets up authentication and session middleware.
func configureAuthMiddleware(e *echo.Echo, sc *ServerConfig) {
	cfg := sc.Config

	// Note: SetCurrentUserWithService is applied per route group in routes.go with UserService injection
	// This allows proper dependency injection following Clean Architecture principles
	e.Use(middleware.SessionTracking) // Track active sessions for metrics
	e.Use(middleware.LanguageDetection)

	// Set service version and config in context
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("service_version", cfg.ServiceVersion)
			c.Set("config", cfg) // Make config available in Echo context

			// Inject config into Request Context for templates
			ctx := context.WithValue(c.Request().Context(), config.ConfigContextKey, cfg)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	})
}

// StartMetricsCollector starts a goroutine that periodically updates metrics.
// It can be gracefully stopped via the provided context.
// The db parameter is used for metrics queries instead of the global database.DB.
func StartMetricsCollector(ctx context.Context, db *gorm.DB) {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		// Initial metrics update
		updateMetrics(ctx, db)

		for {
			select {
			case <-ticker.C:
				updateMetrics(ctx, db)
			case <-ctx.Done():
				slog.Info("Metrics collector shutting down")
				return
			}
		}
	}()
}

// updateMetrics updates Prometheus gauges for resource counts and DB stats.
// Uses a timeout context to prevent hanging on database issues.
func updateMetrics(parentCtx context.Context, db *gorm.DB) {
	// Create a timeout context for DB queries (5 seconds max)
	ctx, cancel := context.WithTimeout(parentCtx, 5*time.Second)
	defer cancel()

	// Update resource counts with context (no tracing for background metrics)
	// Note: users count removed for privacy/DSGVO compliance
	var cardsCount, vouchersCount, giftCardsCount int64
	db.WithContext(ctx).Table("cards").Count(&cardsCount)
	db.WithContext(ctx).Table("vouchers").Count(&vouchersCount)
	db.WithContext(ctx).Table("gift_cards").Count(&giftCardsCount)

	metrics.UpdateResourceCounts(cardsCount, vouchersCount, giftCardsCount)

	// Update vouchers by status (active vs expired based on valid_until)
	var vouchersActive, vouchersExpired int64
	db.WithContext(ctx).Table("vouchers").Where("deleted_at IS NULL AND valid_until >= ?", time.Now()).Count(&vouchersActive)
	db.WithContext(ctx).Table("vouchers").Where("deleted_at IS NULL AND valid_until < ?", time.Now()).Count(&vouchersExpired)
	metrics.UpdateVouchersByStatus(vouchersActive, vouchersExpired)

	// Update gift cards by computed status (active, expired, redeemed)
	var gcActive, gcExpired, gcRedeemed int64
	db.WithContext(ctx).Table("gift_cards").Where("deleted_at IS NULL AND current_balance > 0 AND (expires_at IS NULL OR expires_at >= ?)", time.Now()).Count(&gcActive)
	db.WithContext(ctx).Table("gift_cards").Where("deleted_at IS NULL AND current_balance > 0 AND expires_at < ?", time.Now()).Count(&gcExpired)
	db.WithContext(ctx).Table("gift_cards").Where("deleted_at IS NULL AND current_balance <= 0").Count(&gcRedeemed)
	metrics.UpdateGiftCardsByStatus(gcActive, gcExpired, gcRedeemed)

	// Update shares counts
	var cardShares, voucherShares, giftCardShares int64
	db.WithContext(ctx).Table("card_shares").Where("deleted_at IS NULL").Count(&cardShares)
	db.WithContext(ctx).Table("voucher_shares").Where("deleted_at IS NULL").Count(&voucherShares)
	db.WithContext(ctx).Table("gift_card_shares").Where("deleted_at IS NULL").Count(&giftCardShares)
	metrics.UpdateSharesCounts(cardShares, voucherShares, giftCardShares)

	// Update notification metrics
	var nm metrics.NotificationMetrics
	db.WithContext(ctx).Table("push_subscriptions").Count(&nm.PushSubscriptions)
	db.WithContext(ctx).Table("push_subscriptions").Distinct("user_id").Count(&nm.PushSubscribedUsers)
	db.WithContext(ctx).Table("users").Where("email_verified = ?", true).Count(&nm.EmailVerifiedUsers)
	db.WithContext(ctx).Table("users").Where("push_notifications_enabled = ?", true).Count(&nm.PushNotificationsEnabled)
	db.WithContext(ctx).Table("users").Where("email_notifications_enabled = ?", true).Count(&nm.EmailNotificationsEnabled)
	db.WithContext(ctx).Table("users").Where("push_reminders_enabled = ?", true).Count(&nm.PushRemindersEnabled)
	db.WithContext(ctx).Table("users").Where("push_sharing_enabled = ?", true).Count(&nm.PushSharingEnabled)
	db.WithContext(ctx).Table("users").Where("email_reminders_enabled = ?", true).Count(&nm.EmailRemindersEnabled)
	db.WithContext(ctx).Table("users").Where("email_sharing_enabled = ?", true).Count(&nm.EmailSharingEnabled)
	metrics.UpdateNotificationMetrics(nm)

	// Cleanup inactive sessions (sessions are re-counted via middleware)
	middleware.CleanupInactiveSessions()

	// Update DB connection pool metrics
	// Note: DB() returns the underlying *sql.DB without executing queries
	// Stats() reads internal connection pool metrics (no DB queries, no traces)
	sqlDB, err := db.DB()
	if err != nil {
		slog.Warn("Failed to get database connection pool stats", "error", err)
		return
	}

	stats := sqlDB.Stats()
	metrics.UpdateDBMetrics(stats.InUse, stats.Idle)
}
