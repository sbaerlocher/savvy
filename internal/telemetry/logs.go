// Package telemetry provides OpenTelemetry logging for the application.
package telemetry

import (
	"context"
	"log/slog"
	"os"

	"savvy/internal/logsafe"
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

	// Wrap the base handler so user-controlled string values are stripped of
	// control characters (log-injection defense), then add service metadata.
	enrichedHandler := &enrichedLogHandler{
		handler:        NewSanitizingHandler(handler),
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
	// Add service info. Log-injection sanitization is handled by the wrapped
	// SanitizingHandler, so these enrichment attributes (trusted constants) are
	// simply appended here.
	record.AddAttrs(
		slog.String("service.name", h.serviceName),
		slog.String("service.version", h.serviceVersion),
	)

	// Add trace context if available (set by otelecho middleware)
	// The trace ID is automatically added by the OTEL middleware in Echo
	return h.handler.Handle(ctx, record)
}

// NewSanitizingHandler wraps h so that control characters (notably CR/LF) are
// stripped from every string attribute value and from the message before the
// record is emitted. This is defense-in-depth against log injection: user
// input embedded in a log line cannot forge additional lines. Use it around
// any slog handler that may receive user-controlled values.
func NewSanitizingHandler(h slog.Handler) slog.Handler {
	return &sanitizingHandler{handler: h}
}

// sanitizingHandler strips control characters from string values in every
// record it forwards.
type sanitizingHandler struct {
	handler slog.Handler
}

func (h *sanitizingHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *sanitizingHandler) Handle(ctx context.Context, record slog.Record) error {
	safe := slog.NewRecord(record.Time, record.Level, logsafe.String(record.Message), record.PC)
	record.Attrs(func(a slog.Attr) bool {
		safe.AddAttrs(sanitizeAttr(a))
		return true
	})
	return h.handler.Handle(ctx, safe)
}

func (h *sanitizingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	sanitized := make([]slog.Attr, len(attrs))
	for i, a := range attrs {
		sanitized[i] = sanitizeAttr(a)
	}
	return &sanitizingHandler{handler: h.handler.WithAttrs(sanitized)}
}

func (h *sanitizingHandler) WithGroup(name string) slog.Handler {
	return &sanitizingHandler{handler: h.handler.WithGroup(name)}
}

// sanitizeAttr recursively strips control characters from string attribute
// values, descending into groups so nested string values are covered too.
func sanitizeAttr(a slog.Attr) slog.Attr {
	switch a.Value.Kind() {
	case slog.KindString:
		return slog.String(a.Key, logsafe.String(a.Value.String()))
	case slog.KindGroup:
		group := a.Value.Group()
		out := make([]slog.Attr, len(group))
		for i, g := range group {
			out[i] = sanitizeAttr(g)
		}
		return slog.Attr{Key: a.Key, Value: slog.GroupValue(out...)}
	case slog.KindLogValuer:
		return sanitizeAttr(slog.Attr{Key: a.Key, Value: a.Value.Resolve()})
	default:
		return a
	}
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
