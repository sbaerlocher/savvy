// Package setup contains setup logic for initializing server dependencies.
package setup

import (
	"context"
	"log"
	"log/slog"
	"os"
	"savvy/internal/assets"
	"savvy/internal/config"
	"savvy/internal/database"
	"savvy/internal/handlers"
	"savvy/internal/i18n"
	"savvy/internal/middleware"
	"savvy/internal/migrations"
	"savvy/internal/oauth"
	"savvy/internal/security"
	"savvy/internal/telemetry"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
)

// InitLogger initializes structured logging based on environment.
func InitLogger(cfg *config.Config) {
	logLevel := parseLogLevel(cfg.LogLevel)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)

	slog.Info("Starting savvy system",
		"version", cfg.ServiceVersion,
		"environment", cfg.Environment,
		"log_level", cfg.LogLevel,
		"otel_enabled", cfg.OTelEnabled,
	)
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
		log.Printf("Warning: Invalid LOG_LEVEL '%s', defaulting to INFO. Valid values: DEBUG, INFO, WARN, ERROR", level)
		return slog.LevelInfo
	}
}

// InitTelemetry initializes OpenTelemetry tracing.
// Returns a shutdown function that must be called on server shutdown.
func InitTelemetry(cfg *config.Config) (func(context.Context) error, error) {
	otelEnabled := cfg.OTelEnabled
	if cfg.IsProduction() && !cfg.OTelEnabled {
		slog.Warn("OTel is disabled in production - metrics and tracing unavailable")
	}

	shutdown, err := telemetry.InitTracer(
		cfg.ServiceName,
		cfg.ServiceVersion,
		cfg.OTelEndpoint,
		otelEnabled,
	)
	if err != nil {
		return nil, err
	}

	return shutdown, nil
}

// InitSessionStore initializes the session store with secure configuration.
func InitSessionStore(cfg *config.Config) {
	middleware.InitSessionStore(cfg.SessionSecret, cfg.IsProduction())
}

// InitI18n initializes internationalization with embedded locale files.
func InitI18n() error {
	if err := i18n.Init(assets.Locales); err != nil {
		return err
	}
	return nil
}

// InitSecurity initializes the security package for token-based barcode access.
func InitSecurity(cfg *config.Config) {
	security.Init(cfg.SessionSecret)
}

// InitDatabase connects to the database and optionally enables telemetry.
func InitDatabase(cfg *config.Config) error {
	if err := database.Connect(cfg.DatabaseURL); err != nil {
		return err
	}

	// Enable database telemetry if OTel is enabled
	if cfg.OTelEnabled {
		if err := database.EnableTelemetry(cfg.ServiceName); err != nil {
			log.Printf("Warning: Failed to enable database telemetry: %v", err)
		}
	}

	return nil
}

// RunMigrations executes database migrations using Gormigrate.
func RunMigrations(cfg *config.Config) error {
	if !cfg.AutoMigrate {
		log.Println("⚠️  AutoMigrate disabled (AUTO_MIGRATE=false)")
		log.Println("   Run migrations manually: make migrate-up")
		return nil
	}

	log.Println("🚀 Running database migrations...")
	m := gormigrate.New(database.DB, gormigrate.DefaultOptions, migrations.GetMigrations())
	if err := m.Migrate(); err != nil {
		return err
	}
	log.Println("✅ Migrations applied successfully")
	return nil
}

// InitAuditLogging enables audit logging for all deletions.
func InitAuditLogging() {
	if err := database.EnableAuditLogging(); err != nil {
		log.Printf("Warning: Failed to enable audit logging: %v", err)
	}
}

// InitOAuth initializes OAuth provider if configured.
func InitOAuth(cfg *config.Config) {
	if !cfg.IsOAuthEnabled() {
		handlers.InitOAuth(nil, cfg)
		log.Printf("ℹ️  OAuth not configured (set OAUTH_CLIENT_ID, OAUTH_CLIENT_SECRET, OAUTH_ISSUER to enable)")
		return
	}

	oauthProvider, err := oauth.NewProvider(cfg)
	if err != nil {
		log.Printf("Warning: Failed to initialize OAuth: %v", err)
		handlers.InitOAuth(nil, cfg) // Disable OAuth but pass config
		return
	}

	handlers.InitOAuth(oauthProvider, cfg)
	log.Printf("✅ OAuth initialized successfully")
	if len(cfg.OAuthAdminEmails) > 0 {
		log.Printf("   Admin emails: %v", cfg.OAuthAdminEmails)
	}
	if cfg.OAuthAdminGroup != "" {
		log.Printf("   Admin group: %s", cfg.OAuthAdminGroup)
	}
}

// InitAllDependencies initializes all application dependencies in the correct order.
// Returns the telemetry shutdown function that must be called on server shutdown.
func InitAllDependencies(cfg *config.Config) (func(context.Context) error, error) {
	// 1. Logging
	InitLogger(cfg)

	// 2. Telemetry
	shutdown, err := InitTelemetry(cfg)
	if err != nil {
		return nil, err
	}

	// 3. Session Store
	InitSessionStore(cfg)

	// 4. i18n
	if err := InitI18n(); err != nil {
		return shutdown, err
	}

	// 5. Security
	InitSecurity(cfg)

	// 6. Database
	if err := InitDatabase(cfg); err != nil {
		return shutdown, err
	}

	// 7. Migrations
	if err := RunMigrations(cfg); err != nil {
		return shutdown, err
	}

	// 8. Audit Logging
	InitAuditLogging()

	// 9. OAuth
	InitOAuth(cfg)

	return shutdown, nil
}

// Shutdown performs graceful cleanup of resources.
func Shutdown(shutdownFn func(context.Context) error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := shutdownFn(ctx); err != nil {
		log.Printf("Error shutting down telemetry: %v", err)
	}
}
