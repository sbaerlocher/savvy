// Package main is the entry point for the savvy system server.
package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"savvy/internal/config"
	"savvy/internal/database"
	"savvy/internal/handlers"
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
	}

	url := "http://127.0.0.1:" + port + "/health"
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		return 1
	}
	resp, err := client.Do(req)
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
		log.Printf("Health check returned status %d", resp.StatusCode)
		return 1
	}

	return 0
}

func run() int {
	// Load config
	cfg := config.Load()

	// Validate production secrets before starting
	if err := cfg.ValidateProduction(); err != nil {
		log.Fatalf("❌ Production validation failed: %v", err)
		return 1
	}

	// Initialize all dependencies (logging, telemetry, database, migrations, OAuth, etc.)
	shutdown, err := setup.InitAllDependencies(cfg)
	if err != nil {
		log.Printf("Failed to initialize dependencies: %v", err)
		return 1
	}
	defer setup.Shutdown(shutdown)

	// Initialize service container and health handler
	serviceContainer := services.NewContainer(database.DB)
	healthHandler := handlers.NewHealthHandler(database.DB)

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
	}
	setup.RegisterRoutes(routeConfig)

	// Start metrics collector goroutine
	setup.StartMetricsCollector()

	// Start server with graceful shutdown
	slog.Info("Server starting", "port", cfg.ServerPort)

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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		log.Printf("Error shutting down server: %v", err)
		return 1
	}

	log.Println("Server gracefully stopped")
	return 0
}
