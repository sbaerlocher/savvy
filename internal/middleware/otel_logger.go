// Package middleware contains Echo middleware for authentication, sessions, and observability.
package middleware

import (
	"github.com/labstack/echo/v5"
	"go.opentelemetry.io/otel/trace"
)

// OTelLogger is a middleware that adds trace context to Echo's logger
func OTelLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get trace context from request
			span := trace.SpanFromContext(c.Request().Context())
			spanContext := span.SpanContext()

			// Add trace and span IDs to logger if available
			if spanContext.IsValid() {
				c.Logger().SetPrefix(
					"[trace_id=" + spanContext.TraceID().String() +
						" span_id=" + spanContext.SpanID().String() + "] ",
				)
			}

			return next(c)
		}
	}
}
