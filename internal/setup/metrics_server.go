// Package setup provides initialization and configuration for the server.
package setup

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// StartMetricsServer starts a separate HTTP server for Prometheus metrics.
// This follows the industry best practice of isolating metrics endpoints
// from the main application server for security reasons.
//
// The metrics server runs on a separate port (default: 9090) and only
// exposes the /metrics endpoint. This prevents accidental exposure of
// metrics through the main application's ingress/load balancer.
//
// Reference: https://prometheus.io/docs/practices/naming/
func StartMetricsServer(ctx context.Context, port string) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		slog.Info("Metrics server starting", "port", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Metrics server error", "error", err)
		}
	}()

	// Handle graceful shutdown
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.Error("Error shutting down metrics server", "error", err)
		} else {
			slog.Info("Metrics server stopped")
		}
	}()

	return server
}
