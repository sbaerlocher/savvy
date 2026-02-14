// Package telemetry provides OpenTelemetry metrics for the application.
package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.28.0"
)

// InitMetrics initializes OpenTelemetry metrics with OTLP exporter
// Returns a shutdown function that must be called on server shutdown
func InitMetrics(serviceName, serviceVersion, otlpEndpoint string, enabled bool) (func(context.Context) error, error) {
	// If telemetry is disabled, return a no-op shutdown function
	if !enabled {
		slog.Info("OpenTelemetry Metrics disabled")
		return func(_ context.Context) error { return nil }, nil
	}

	// Create resource with service information
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			"", // Empty schema URL to avoid conflicts
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Configure OTLP HTTP exporter for metrics
	// Remove http:// or https:// prefix from endpoint (WithEndpoint expects host:port only)
	endpoint := otlpEndpoint
	endpoint = strings.TrimPrefix(endpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")

	exporter, err := otlpmetrichttp.New(
		context.Background(),
		otlpmetrichttp.WithEndpoint(endpoint),
		otlpmetrichttp.WithInsecure(), // Use TLS in production
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP metric exporter: %w", err)
	}

	// Create meter provider with periodic reader
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(
				exporter,
				sdkmetric.WithInterval(15*time.Second), // Export metrics every 15 seconds
			),
		),
		sdkmetric.WithResource(res),
	)

	// Set global meter provider
	otel.SetMeterProvider(mp)

	slog.Info("OpenTelemetry Metrics enabled", "endpoint", otlpEndpoint)

	// Return shutdown function
	return func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		return mp.Shutdown(ctx)
	}, nil
}
