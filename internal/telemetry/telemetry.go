// Package telemetry provides OpenTelemetry observability (logs, traces, metrics) for the application.
package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.28.0"
)

// Shutdown holds all shutdown functions for telemetry components
type Shutdown struct {
	shutdownLogger  func(context.Context) error
	shutdownTracer  func(context.Context) error
	shutdownMetrics func(context.Context) error
}

// Shutdown gracefully shuts down all telemetry components
func (ts *Shutdown) Shutdown(ctx context.Context) error {
	var firstErr error

	// Shutdown metrics
	if ts.shutdownMetrics != nil {
		if err := ts.shutdownMetrics(ctx); err != nil {
			slog.Error("Error shutting down metrics", "error", err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	// Shutdown tracer
	if ts.shutdownTracer != nil {
		if err := ts.shutdownTracer(ctx); err != nil {
			slog.Error("Error shutting down tracer", "error", err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	// Shutdown logger
	if ts.shutdownLogger != nil {
		if err := ts.shutdownLogger(ctx); err != nil {
			slog.Error("Error shutting down logger", "error", err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	return firstErr
}

// InitTelemetryFull initializes all OpenTelemetry components (logs, traces, metrics)
// Returns a logger, shutdown function, and error
func InitTelemetryFull(serviceName, serviceVersion, otlpEndpoint string, enabled bool, logLevel slog.Level) (*slog.Logger, *Shutdown, error) {
	shutdown := &Shutdown{}

	// 1. Initialize Logger (with OTLP export)
	logger, shutdownLogger, err := InitLogger(serviceName, serviceVersion, otlpEndpoint, enabled, logLevel)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize logger: %w", err)
	}
	shutdown.shutdownLogger = shutdownLogger

	// 2. Initialize Tracer
	shutdownTracer, err := InitTracer(serviceName, serviceVersion, otlpEndpoint, enabled)
	if err != nil {
		return nil, shutdown, fmt.Errorf("failed to initialize tracer: %w", err)
	}
	shutdown.shutdownTracer = shutdownTracer

	// 3. Initialize Metrics
	shutdownMetrics, err := InitMetrics(serviceName, serviceVersion, otlpEndpoint, enabled)
	if err != nil {
		return nil, shutdown, fmt.Errorf("failed to initialize metrics: %w", err)
	}
	shutdown.shutdownMetrics = shutdownMetrics

	return logger, shutdown, nil
}

// InitTracer initializes OpenTelemetry tracing with OTLP exporter
func InitTracer(serviceName, serviceVersion, otlpEndpoint string, enabled bool) (func(context.Context) error, error) {
	// If telemetry is disabled, return a no-op shutdown function
	if !enabled {
		slog.Info("OpenTelemetry disabled (set OTEL_ENABLED=true to enable)")
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

	// Configure OTLP HTTP exporter
	// Remove http:// or https:// prefix from endpoint (WithEndpoint expects host:port only)
	endpoint := otlpEndpoint
	endpoint = strings.TrimPrefix(endpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")

	exporter, err := otlptracehttp.New(
		context.Background(),
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(), // Use TLS in production
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()), // Sample all traces (adjust in production)
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	// Set global propagator for distributed tracing
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	slog.Info("OpenTelemetry enabled", "endpoint", otlpEndpoint, "service", serviceName)

	// Return shutdown function
	return func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		return tp.Shutdown(ctx)
	}, nil
}
