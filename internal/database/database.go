// Package database manages PostgreSQL database connections and migrations using GORM.
package database

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"savvy/internal/audit"
	"savvy/internal/models"
	"savvy/internal/telemetry"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB is the global database instance used throughout the application
var DB *gorm.DB

// Connect establishes a connection to the PostgreSQL database.
// The logLevel parameter controls GORM's SQL query logging and should match
// the application's LOG_LEVEL (e.g., "DEBUG", "INFO", "WARN", "ERROR").
func Connect(databaseURL string, logLevel string) error {
	var err error
	DB, err = gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger:      logger.Default.LogMode(mapLogLevel(logLevel)),
		PrepareStmt: true, // Use prepared statements for better performance
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool to prevent resource exhaustion
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	// Connection pool settings
	sqlDB.SetMaxOpenConns(25)                  // Maximum open connections
	sqlDB.SetMaxIdleConns(5)                   // Maximum idle connections
	sqlDB.SetConnMaxLifetime(5 * time.Minute)  // Connection max lifetime
	sqlDB.SetConnMaxIdleTime(10 * time.Minute) // Idle connection timeout

	slog.Info("✓ Database connected with connection pool configured")
	return nil
}

// EnableTelemetry enables OpenTelemetry tracing for GORM
func EnableTelemetry(serviceName string) error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	plugin := telemetry.NewGORMTelemetryPlugin(serviceName)
	if err := DB.Use(plugin); err != nil {
		return fmt.Errorf("failed to register telemetry plugin: %w", err)
	}

	slog.Info("✓ Database telemetry enabled")
	return nil
}

// EnableAuditLogging enables automatic audit logging for deletions
func EnableAuditLogging() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	if err := audit.SetupAuditHooks(DB); err != nil {
		return fmt.Errorf("failed to setup audit hooks: %w", err)
	}

	slog.Info("✓ Audit logging enabled")
	return nil
}

// AutoMigrate runs GORM auto-migration for all models
func AutoMigrate() error {
	err := DB.AutoMigrate(
		&models.User{},
		&models.Card{},
		&models.CardShare{},
		&models.Voucher{},
		&models.VoucherShare{},
		&models.GiftCard{},
		&models.GiftCardTransaction{},
		&models.GiftCardShare{},
		&models.Merchant{},
		&models.UserFavorite{},
		&models.AuditLog{},
	)
	if err != nil {
		return fmt.Errorf("failed to auto-migrate: %w", err)
	}

	slog.Info("✓ Database auto-migration completed")
	return nil
}

// mapLogLevel converts an application log level string to a GORM logger level.
// DEBUG → Info (log all queries), INFO/WARN → Warn (slow queries + errors), ERROR → Error (errors only).
func mapLogLevel(level string) logger.LogLevel {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return logger.Info
	case "INFO", "WARN", "WARNING":
		return logger.Warn
	case "ERROR":
		return logger.Error
	default:
		return logger.Warn
	}
}

// IsDuplicateError checks if an error is a PostgreSQL unique constraint violation
func IsDuplicateError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return strings.Contains(errMsg, "duplicate key") ||
		strings.Contains(errMsg, "unique constraint") ||
		strings.Contains(errMsg, "violates unique") ||
		strings.Contains(errMsg, "SQLSTATE 23505")
}

// ErrDuplicateKey is returned when a unique constraint is violated
var ErrDuplicateKey = errors.New("duplicate key value violates unique constraint")
