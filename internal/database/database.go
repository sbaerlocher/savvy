// Package database manages PostgreSQL database connections and migrations using GORM.
package database

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"savvy/internal/audit"
	"savvy/internal/models"
	"savvy/internal/telemetry"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB is the global database instance used throughout the application
var DB *gorm.DB

// Connect establishes a connection to the PostgreSQL database
func Connect(databaseURL string) error {
	var err error
	DB, err = gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("✓ Database connected")
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

	log.Println("✓ Database telemetry enabled")
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

	log.Println("✓ Audit logging enabled")
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

	log.Println("✓ Database auto-migration completed")
	return nil
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
