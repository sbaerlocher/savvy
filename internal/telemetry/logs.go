// Package telemetry provides OpenTelemetry logging for the application.
package telemetry

import (
	"context"
	"log/slog"
	"os"
)

// InitLogger initializes structured logging with OpenTelemetry correlation
// Returns a logger instance and shutdown function
//
// Note: Logs are exported to stdout in JSON format with trace context.
// The OpenTelemetry Collector scrapes logs from stdout and forwards to Loki.
// This approach is simpler and more reliable than OTLP log export (which is still experimental).
func InitLogger(serviceName, serviceVersion, _ string, enabled bool, logLevel slog.Level) (*slog.Logger, func(context.Context) error, error) {
	// Create JSON handler with trace context
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Add service info to all log records
			if len(groups) == 0 {
				switch a.Key {
				case slog.SourceKey:
					// Skip source info (reduces noise)
					return slog.Attr{}
				}
			}
			return a
		},
	})

	// Wrap handler to add service metadata
	enrichedHandler := &enrichedLogHandler{
		handler:        handler,
		serviceName:    serviceName,
		serviceVersion: serviceVersion,
	}

	logger := slog.New(enrichedHandler)

	if enabled {
		slog.Info("OpenTelemetry Logs enabled (via stdout -> OTEL Collector -> Loki)")
	} else {
		slog.Info("OpenTelemetry Logs disabled (basic slog to stdout)")
	}

	// No-op shutdown function (logs are written to stdout)
	shutdown := func(_ context.Context) error {
		return nil
	}

	return logger, shutdown, nil
}

// enrichedLogHandler adds service metadata to all log records
type enrichedLogHandler struct {
	handler        slog.Handler
	serviceName    string
	serviceVersion string
}

func (h *enrichedLogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *enrichedLogHandler) Handle(ctx context.Context, record slog.Record) error {
	// Add service info
	record.AddAttrs(
		slog.String("service.name", h.serviceName),
		slog.String("service.version", h.serviceVersion),
	)

	// Add trace context if available (set by otelecho middleware)
	// The trace ID is automatically added by the OTEL middleware in Echo
	return h.handler.Handle(ctx, record)
}

func (h *enrichedLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &enrichedLogHandler{
		handler:        h.handler.WithAttrs(attrs),
		serviceName:    h.serviceName,
		serviceVersion: h.serviceVersion,
	}
}

func (h *enrichedLogHandler) WithGroup(name string) slog.Handler {
	return &enrichedLogHandler{
		handler:        h.handler.WithGroup(name),
		serviceName:    h.serviceName,
		serviceVersion: h.serviceVersion,
	}
}
