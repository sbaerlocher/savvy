// Package main provides an E2E test wrapper that runs database migrations and seeding before starting the server.
package main

import (
	"context"
	"log"
	"os"
	"os/exec"
	"savvy/internal/config"
	"savvy/internal/database"
	"savvy/internal/setup"
	"syscall"
)

func main() {
	log.Println("🎭 E2E Wrapper: Starting setup...")

	// Load configuration
	cfg := config.Load()

	// Initialize database connection
	log.Println("📦 Connecting to database...")
	if err := database.Connect(cfg.DatabaseURL, cfg.LogLevel); err != nil {
		log.Printf("❌ Database connection failed: %v", err)
		os.Exit(1)
	}

	// Run database migrations
	log.Println("🚀 Running database migrations...")
	if err := setup.RunMigrations(cfg); err != nil {
		log.Printf("❌ Migration failed: %v", err)
		os.Exit(1)
	}

	log.Println("✅ Migrations completed successfully")

	// Run seed binary (it will use the same database connection)
	log.Println("🌱 Seeding database...")
	ctx := context.Background()
	seedCmd := exec.CommandContext(ctx, "/app/seed")
	seedCmd.Stdout = os.Stdout
	seedCmd.Stderr = os.Stderr

	if err := seedCmd.Run(); err != nil {
		log.Printf("❌ Seeding failed: %v", err)
		os.Exit(1)
	}

	log.Println("✅ Seeding completed successfully")
	log.Println("🚀 Starting server...")

	// Replace current process with server binary (exec)
	// This ensures the server receives signals correctly (SIGTERM, SIGINT)
	// #nosec G204 - This is intentional process replacement in E2E test wrapper
	if err := syscall.Exec("/app/savvy", []string{"/app/savvy"}, os.Environ()); err != nil {
		log.Printf("❌ Failed to start server: %v", err)
		os.Exit(1)
	}
}
