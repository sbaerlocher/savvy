// Package setup contains setup logic for initializing server dependencies.
package setup

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"savvy/internal/assets"
	"savvy/internal/config"
	"savvy/internal/database"
	"savvy/internal/email"
	"savvy/internal/i18n"
	"savvy/internal/middleware"
	"savvy/internal/migrations"
	"savvy/internal/oauth"
	"savvy/internal/repository"
	"savvy/internal/telemetry"
	"strconv"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"golang.org/x/time/rate"
)

// InitLogger initializes structured logging with OpenTelemetry integration.
// Returns the logger and shutdown function.
func InitLogger(cfg *config.Config) (*slog.Logger, func(context.Context) error, error) {
	logLevel := parseLogLevel(cfg.LogLevel)

	// Initialize full telemetry (logs, traces, metrics) via OTEL
	logger, shutdown, err := telemetry.InitTelemetryFull(
		cfg.ServiceName,
		cfg.ServiceVersion,
		cfg.OTelEndpoint,
		cfg.OTelEnabled,
		logLevel,
	)
	if err != nil {
		// Fallback to basic logger if OTEL fails
		basicLogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: logLevel,
		}))
		slog.SetDefault(basicLogger)
		return basicLogger, func(_ context.Context) error { return nil }, fmt.Errorf("telemetry init failed: %w", err)
	}

	// Set as default logger
	slog.SetDefault(logger)

	slog.Info("Starting savvy system",
		"version", cfg.ServiceVersion,
		"environment", cfg.Environment,
		"log_level", cfg.LogLevel,
		"otel_enabled", cfg.OTelEnabled,
	)

	// Return a wrapper that converts telemetry.Shutdown to simple shutdown function
	shutdownFn := func(ctx context.Context) error {
		return shutdown.Shutdown(ctx)
	}

	return logger, shutdownFn, nil
}

// parseLogLevel converts string log level to slog.Level.
// Valid values: DEBUG, INFO, WARN, ERROR (case-insensitive).
// Defaults to INFO if invalid.
func parseLogLevel(level string) slog.Level {
	switch level {
	case "DEBUG", "debug":
		return slog.LevelDebug
	case "INFO", "info":
		return slog.LevelInfo
	case "WARN", "warn", "WARNING", "warning":
		return slog.LevelWarn
	case "ERROR", "error":
		return slog.LevelError
	default:
		slog.Warn("Invalid LOG_LEVEL, defaulting to INFO", "level", level)
		return slog.LevelInfo
	}
}

// InitSessionStore initializes the PostgreSQL-backed session store.
func InitSessionStore(cfg *config.Config) {
	repo := repository.NewSessionRepository(database.DB)
	middleware.InitSessionStore(repo, cfg.SessionMaxAge)
}

// InitI18n initializes internationalization with embedded locale files.
func InitI18n() error {
	if err := i18n.Init(assets.Locales); err != nil {
		return err
	}
	return nil
}

// InitDatabase connects to the database and optionally enables telemetry.
func InitDatabase(cfg *config.Config) error {
	if err := database.Connect(cfg.DatabaseURL, cfg.LogLevel); err != nil {
		return err
	}

	// ✅ GORM telemetry with parent-span validation to prevent orphaned traces
	//
	// HOW IT WORKS:
	// - GORM plugin checks for active parent span before creating child spans (gorm.go:92-96)
	// - HTTP requests (with parent span from otelecho middleware) → GORM spans created as children
	// - Background queries (using context.Background()) → GORM spans skipped (no parent span)
	//
	// BENEFITS:
	// - Clean trace hierarchy: HTTP Request → Handler → Service → GORM Query
	// - No orphaned spans from background metrics collection
	// - Full DB instrumentation for actual user requests
	//
	// TRACE EXAMPLE:
	//   GET /api/v1/cards (65.94ms)
	//   ├── gorm:query (44.08ms)  ← SELECT * FROM cards
	//   ├── gorm:query (14.75ms)  ← SELECT * FROM merchants
	//   └── gorm:query (4.08ms)   ← SELECT * FROM card_shares
	if cfg.OTelEnabled {
		if err := database.EnableTelemetry(cfg.ServiceName); err != nil {
			slog.Warn("Failed to enable database telemetry", "error", err)
		}
	}

	return nil
}

// RunMigrations executes database migrations using Gormigrate.
func RunMigrations(cfg *config.Config) error {
	if !cfg.AutoMigrate {
		slog.Warn("AutoMigrate disabled (AUTO_MIGRATE=false), run migrations manually: make migrate-up")
		return nil
	}

	slog.Info("Running database migrations...")
	m := gormigrate.New(database.DB, gormigrate.DefaultOptions, migrations.GetMigrations())
	if err := m.Migrate(); err != nil {
		return err
	}
	slog.Info("Migrations applied successfully")
	return nil
}

// InitAuditLogging enables audit logging for all deletions.
func InitAuditLogging() {
	if err := database.EnableAuditLogging(); err != nil {
		slog.Warn("Failed to enable audit logging", "error", err)
	}
}

// InitOAuth initializes OAuth provider if configured.
// Returns the provider (nil if OAuth is not configured or initialization fails).
func InitOAuth(cfg *config.Config) *oauth.Provider {
	if !cfg.IsOAuthEnabled() {
		slog.Info("OAuth not configured (set OAUTH_CLIENT_ID, OAUTH_CLIENT_SECRET, OAUTH_ISSUER to enable)")
		return nil
	}

	oauthProvider, err := oauth.NewProvider(cfg)
	if err != nil {
		slog.Warn("Failed to initialize OAuth", "error", err)
		return nil
	}

	slog.Info("OAuth initialized successfully",
		"admin_emails", len(cfg.OAuthAdminEmails),
		"admin_group", cfg.OAuthAdminGroup,
	)
	return oauthProvider
}

// InitEmailService initializes the email service based on configuration.
// Returns SMTP service if configured, otherwise a log-based fallback.
func InitEmailService(cfg *config.Config) email.ServiceInterface {
	if cfg.IsSMTPEnabled() {
		smtpCfg := email.SMTPConfig{
			Host:        cfg.SMTPHost,
			Port:        cfg.SMTPPort,
			Username:    cfg.SMTPUsername,
			Password:    cfg.SMTPPassword,
			FromEmail:   cfg.SMTPFromEmail,
			FromName:    cfg.SMTPFromName,
			UseTLS:      cfg.SMTPUseTLS,
			FrontendURL: cfg.FrontendURL,
		}

		svc, err := email.NewSMTPEmailService(smtpCfg)
		if err != nil {
			slog.Warn("Failed to initialize SMTP email service, falling back to log-based", "error", err)
			return email.NewLogEmailService()
		}

		slog.Info("SMTP email service initialized", "host", cfg.SMTPHost, "port", cfg.SMTPPort)
		return svc
	}

	slog.Info("SMTP not configured (set SMTP_HOST, SMTP_FROM_EMAIL to enable)")
	return email.NewLogEmailService()
}

// InitAllDependencies initializes all application dependencies in the correct order.
// Returns the telemetry shutdown function and the OAuth provider (nil if not configured).
func InitAllDependencies(cfg *config.Config) (func(context.Context) error, *oauth.Provider, error) {
	// 1. Logging + Telemetry (combined initialization)
	logger, shutdown, err := InitLogger(cfg)
	if err != nil {
		slog.Warn("Telemetry initialization failed", "error", err)
		// Continue with basic logger, shutdown will be no-op
	}
	_ = logger // Logger is set as default via slog.SetDefault in InitLogger

	// 2. i18n
	if err := InitI18n(); err != nil {
		return shutdown, nil, err
	}

	// 3. Database
	if err := InitDatabase(cfg); err != nil {
		return shutdown, nil, err
	}

	// 4. Migrations
	if err := RunMigrations(cfg); err != nil {
		return shutdown, nil, err
	}

	// 5. Session Store (requires database for PostgreSQL-backed sessions)
	InitSessionStore(cfg)

	// 6. Audit Logging
	InitAuditLogging()

	// 7. OAuth
	oauthProvider := InitOAuth(cfg)

	return shutdown, oauthProvider, nil
}

// RateLimiters holds all rate limiters used by the application.
type RateLimiters struct {
	Global        *middleware.IPRateLimiter
	Auth          *middleware.IPRateLimiter
	PasswordReset *middleware.IPRateLimiter
	TwoFAChallenge *middleware.IPRateLimiter // Stricter limiter for 2FA challenge endpoint
	User          *middleware.UserRateLimiter
}

// InitRateLimiters creates and returns rate limiters for global and auth endpoints.
func InitRateLimiters() *RateLimiters {
	// Read rate limit configuration from environment with defaults
	globalRate := getEnvAsFloat("RATE_LIMIT_GLOBAL_RATE", 100)
	globalBurst := getEnvAsInt("RATE_LIMIT_GLOBAL_BURST", 20)
	authRate := getEnvAsFloat("RATE_LIMIT_AUTH_RATE", 5)
	authBurst := getEnvAsInt("RATE_LIMIT_AUTH_BURST", 3)

	passwordResetRate := getEnvAsFloat("RATE_LIMIT_PASSWORD_RESET_RATE", 1.0/60.0) // 1 req per 60s
	passwordResetBurst := getEnvAsInt("RATE_LIMIT_PASSWORD_RESET_BURST", 1)

	// 2FA challenge: 1 req per 3s per IP (burst 2) — stricter than auth to slow brute-force
	twoFAChallengeRate := getEnvAsFloat("RATE_LIMIT_2FA_CHALLENGE_RATE", 1.0/3.0)
	twoFAChallengeB := getEnvAsInt("RATE_LIMIT_2FA_CHALLENGE_BURST", 2)

	userRate := getEnvAsFloat("RATE_LIMIT_USER_RATE", 30)
	userBurst := getEnvAsInt("RATE_LIMIT_USER_BURST", 20)

	globalLimiter := middleware.NewIPRateLimiter(rate.Limit(globalRate), globalBurst)
	authLimiter := middleware.NewIPRateLimiter(rate.Limit(authRate), authBurst)
	passwordResetLimiter := middleware.NewIPRateLimiter(rate.Limit(passwordResetRate), passwordResetBurst)
	twoFAChallengeLimiter := middleware.NewIPRateLimiter(rate.Limit(twoFAChallengeRate), twoFAChallengeB)
	userLimiter := middleware.NewUserRateLimiter(rate.Limit(userRate), userBurst)

	slog.Info("Rate limiters initialized",
		"global_rate", fmt.Sprintf("%.0f req/s", globalRate),
		"global_burst", globalBurst,
		"auth_rate", fmt.Sprintf("%.0f req/s", authRate),
		"auth_burst", authBurst,
		"password_reset_rate", fmt.Sprintf("%.4f req/s (1 per %.0fs)", passwordResetRate, 1.0/passwordResetRate),
		"password_reset_burst", passwordResetBurst,
		"2fa_challenge_rate", fmt.Sprintf("%.4f req/s (1 per %.0fs)", twoFAChallengeRate, 1.0/twoFAChallengeRate),
		"2fa_challenge_burst", twoFAChallengeB,
		"user_rate", fmt.Sprintf("%.0f req/s", userRate),
		"user_burst", userBurst)

	return &RateLimiters{
		Global:         globalLimiter,
		Auth:           authLimiter,
		PasswordReset:  passwordResetLimiter,
		TwoFAChallenge: twoFAChallengeLimiter,
		User:           userLimiter,
	}
}

// getEnvAsFloat reads an environment variable as float64 with a default value.
func getEnvAsFloat(key string, defaultVal float64) float64 {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultVal
	}
	val, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		slog.Warn("Invalid environment variable value, using default", // #nosec G706 -- structured logging, not format string injection
			"key", key,
			"value", valStr,
			"default", defaultVal,
			"error", err)
		return defaultVal
	}
	return val
}

// getEnvAsInt reads an environment variable as int with a default value.
func getEnvAsInt(key string, defaultVal int) int {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		slog.Warn("Invalid environment variable value, using default", // #nosec G706 -- structured logging, not format string injection
			"key", key,
			"value", valStr,
			"default", defaultVal,
			"error", err)
		return defaultVal
	}
	return val
}

// Shutdown performs graceful cleanup of resources.
func Shutdown(shutdownFn func(context.Context) error, rateLimiters *RateLimiters) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown telemetry
	if err := shutdownFn(ctx); err != nil {
		slog.Error("Error shutting down telemetry", "error", err)
	}

	// Shutdown rate limiters
	if rateLimiters != nil {
		rateLimiters.Global.Shutdown()
		rateLimiters.Auth.Shutdown()
		rateLimiters.PasswordReset.Shutdown()
		rateLimiters.TwoFAChallenge.Shutdown()
		rateLimiters.User.Shutdown()
		slog.Info("Rate limiters stopped")
	}
}
