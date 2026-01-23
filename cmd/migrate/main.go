// Package main provides a CLI tool for managing database migrations.
// This is similar to Laravel's "php artisan migrate" commands.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/go-gormigrate/gormigrate/v2"
	"savvy/internal/config"
	"savvy/internal/database"
	"savvy/internal/migrations"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Load configuration
	cfg := config.Load()

	// Connect to database
	if err := database.Connect(cfg.DatabaseURL); err != nil {
		log.Fatalf("❌ Database connection failed: %v", err)
	}

	// Initialize Gormigrate
	m := gormigrate.New(database.DB, gormigrate.DefaultOptions, migrations.GetMigrations())

	command := os.Args[1]

	switch command {
	case "up":
		migrateUp(m)
	case "down":
		migrateDown(m)
	case "reset":
		migrateReset(m)
	case "to":
		if len(os.Args) < 3 {
			fmt.Println("❌ Error: Version required")
			fmt.Println("Usage: go run cmd/migrate/main.go to VERSION")
			os.Exit(1)
		}
		migrateTo(m, os.Args[2])
	case "status":
		migrateStatus(m)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Printf("❌ Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func migrateUp(m *gormigrate.Gormigrate) {
	fmt.Println("🚀 Running migrations...")

	if err := m.Migrate(); err != nil {
		log.Fatalf("❌ Migration failed: %v", err)
	}

	fmt.Println("✅ Migrations applied successfully")
}

func migrateDown(m *gormigrate.Gormigrate) {
	fmt.Println("⏪ Rolling back last migration...")

	if err := m.RollbackLast(); err != nil {
		log.Fatalf("❌ Rollback failed: %v", err)
	}

	fmt.Println("✅ Rolled back last migration")
}

func migrateReset(m *gormigrate.Gormigrate) {
	fmt.Println("⚠️  WARNING: This will rollback ALL migrations!")
	fmt.Print("Type 'yes' to continue: ")

	var confirm string
	fmt.Scanln(&confirm)

	if confirm != "yes" {
		fmt.Println("❌ Aborted")
		os.Exit(0)
	}

	fmt.Println("🔄 Rolling back all migrations...")

	if err := m.RollbackTo("0"); err != nil {
		log.Fatalf("❌ Reset failed: %v", err)
	}

	fmt.Println("✅ All migrations rolled back")
}

func migrateTo(m *gormigrate.Gormigrate, version string) {
	fmt.Printf("🎯 Migrating to version: %s\n", version)

	if err := m.MigrateTo(version); err != nil {
		log.Fatalf("❌ Migration to %s failed: %v", version, err)
	}

	fmt.Printf("✅ Migrated to version %s\n", version)
}

func migrateStatus(_ *gormigrate.Gormigrate) {
	// Query the gormigrate migrations table
	var migrations []struct {
		ID string
	}

	if err := database.DB.Table("migrations").Select("id").Order("id ASC").Find(&migrations).Error; err != nil {
		log.Fatalf("❌ Failed to fetch migration status: %v", err)
	}

	if len(migrations) == 0 {
		fmt.Println("📋 No migrations applied yet")
		return
	}

	fmt.Println("📋 Applied migrations:")
	for _, migration := range migrations {
		fmt.Printf("  ✓ %s\n", migration.ID)
	}

	fmt.Printf("\n✅ Total: %d migration(s)\n", len(migrations))
}

func printUsage() {
	fmt.Print(`
Savvy System - Database Migration Tool

USAGE:
    go run cmd/migrate/main.go [COMMAND]

COMMANDS:
    up              Apply all pending migrations (like: php artisan migrate)
    down            Rollback the last migration (like: php artisan migrate:rollback)
    reset           Rollback all migrations (like: php artisan migrate:reset)
    to VERSION      Migrate to a specific version
    status          Show applied migrations (like: php artisan migrate:status)
    help            Show this help message

EXAMPLES:
    go run cmd/migrate/main.go up
    go run cmd/migrate/main.go down
    go run cmd/migrate/main.go reset
    go run cmd/migrate/main.go to 202601230001_init_schema
    go run cmd/migrate/main.go status

MAKEFILE SHORTCUTS:
    make migrate-up         Same as: up
    make migrate-down       Same as: down
    make migrate-reset      Same as: reset
    make migrate-status     Same as: status
`)
}
