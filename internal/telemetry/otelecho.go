// Package telemetry provides OpenTelemetry instrumentation for the savvy application.
package telemetry

import (
	"fmt"
	"slices"
	"strings"

	"github.com/labstack/echo/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

const otelScopeName = "savvy/internal/telemetry/otelecho"

// OTelMiddlewareConfig holds the configuration for the OTel middleware.
type OTelMiddlewareConfig struct {
	// Skipper defines a function to skip middleware.
	Skipper func(c *echo.Context) bool
}

// OTelMiddleware returns an Echo v5 compatible OpenTelemetry tracing middleware.
func OTelMiddleware(serverName string, cfg OTelMiddlewareConfig) echo.MiddlewareFunc {
	tracer := otel.GetTracerProvider().Tracer(
		otelScopeName,
		oteltrace.WithInstrumentationVersion("v1"),
	)
	propagators := otel.GetTextMapPropagator()

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if cfg.Skipper != nil && cfg.Skipper(c) {
				return next(c)
			}

			request := c.Request()
			savedCtx := request.Context()
			defer func() {
				c.SetRequest(request.WithContext(savedCtx))
			}()

			ctx := propagators.Extract(savedCtx, propagation.HeaderCarrier(request.Header))

			method := strings.ToUpper(request.Method)
			if !slices.Contains([]string{
				"GET", "HEAD", "POST", "PUT", "PATCH",
				"DELETE", "CONNECT", "OPTIONS", "TRACE",
			}, method) {
				method = "HTTP"
			}

			path := c.Path()
			spanName := method
			if path != "" {
				spanName = method + " " + path
			}

			attrs := []attribute.KeyValue{
				semconv.HTTPRequestMethodKey.String(method),
				semconv.ServerAddress(serverName),
				semconv.URLPath(request.URL.Path),
				// URLQuery is intentionally omitted to avoid capturing PII/tokens in traces
			}

			ctx, span := tracer.Start(ctx, spanName,
				oteltrace.WithSpanKind(oteltrace.SpanKindServer),
				oteltrace.WithAttributes(attrs...),
			)
			defer span.End()

			c.SetRequest(request.WithContext(ctx))

			err := next(c)

			_, status := echo.ResolveResponseStatus(c.Response(), err)
			span.SetAttributes(semconv.HTTPResponseStatusCode(status))

			if err != nil {
				span.SetAttributes(attribute.String("echo.error", err.Error()))
				span.SetStatus(codes.Error, err.Error())
			} else if status >= 500 {
				span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", status))
			}

			return err
		}
	}
}
