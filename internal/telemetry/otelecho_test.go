package telemetry_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"savvy/internal/telemetry"
)

// setupTestTracer registers an in-memory TracerProvider and W3C propagator globally.
func setupTestTracer() *tracetest.SpanRecorder {
	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	return recorder
}

func TestOTelMiddleware_SkipperPreventsSpan(t *testing.T) {
	recorder := setupTestTracer()

	e := echo.New()
	cfg := telemetry.OTelMiddlewareConfig{
		Skipper: func(_ *echo.Context) bool { return true },
	}
	mw := telemetry.OTelMiddleware("test-server", cfg)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	called := false
	err := mw(func(_ *echo.Context) error {
		called = true
		return nil
	})(c)

	require.NoError(t, err)
	assert.True(t, called, "handler must be called even when skipped")
	assert.Empty(t, recorder.Ended(), "no span should be created when skipped")
}

func TestOTelMiddleware_SpanNameIncludesMethodAndRoute(t *testing.T) {
	recorder := setupTestTracer()

	e := echo.New()
	cfg := telemetry.OTelMiddlewareConfig{}
	mw := telemetry.OTelMiddleware("test-server", cfg)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cards/123", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/cards/:id") // simulate matched route pattern

	err := mw(func(c *echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})(c)
	require.NoError(t, err)

	spans := recorder.Ended()
	require.Len(t, spans, 1)
	assert.Equal(t, "GET /api/v1/cards/:id", spans[0].Name())
}

func TestOTelMiddleware_ResponseStatusAttribute(t *testing.T) {
	recorder := setupTestTracer()

	e := echo.New()
	cfg := telemetry.OTelMiddlewareConfig{}
	mw := telemetry.OTelMiddleware("test-server", cfg)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := mw(func(c *echo.Context) error {
		return c.String(http.StatusCreated, "created")
	})(c)

	require.NoError(t, err)

	spans := recorder.Ended()
	require.Len(t, spans, 1)

	var found bool
	for _, attr := range spans[0].Attributes() {
		if string(attr.Key) == "http.response.status_code" {
			assert.EqualValues(t, 201, attr.Value.AsInt64())
			found = true
		}
	}
	assert.True(t, found, "http.response.status_code attribute must be present")
}

func TestOTelMiddleware_ContextPropagation(t *testing.T) {
	recorder := setupTestTracer()

	e := echo.New()
	cfg := telemetry.OTelMiddlewareConfig{}
	mw := telemetry.OTelMiddleware("test-server", cfg)

	// Use a fixed, valid W3C trace ID injected via traceparent header
	const fixedTraceID = "aaaabbbbccccdddd1111222233334444"
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("traceparent", "00-"+fixedTraceID+"-1234567890abcdef-01")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := mw(func(c *echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})(c)

	require.NoError(t, err)

	spans := recorder.Ended()
	require.Len(t, spans, 1)
	assert.Equal(t, fixedTraceID, spans[0].SpanContext().TraceID().String(),
		"child span should share the trace ID from the incoming traceparent header")
}

func TestOTelMiddleware_UnknownMethodFallback(t *testing.T) {
	recorder := setupTestTracer()

	e := echo.New()
	cfg := telemetry.OTelMiddlewareConfig{}
	mw := telemetry.OTelMiddleware("test-server", cfg)

	req := httptest.NewRequest("CUSTOM", "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := mw(func(c *echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})(c)

	require.NoError(t, err)

	spans := recorder.Ended()
	require.Len(t, spans, 1)
	assert.Equal(t, "HTTP", spans[0].Name(), "unknown HTTP method should use 'HTTP' as span name")
}

func TestOTelMiddleware_ErrorSetsSpanStatus(t *testing.T) {
	recorder := setupTestTracer()

	e := echo.New()
	cfg := telemetry.OTelMiddlewareConfig{}
	mw := telemetry.OTelMiddleware("test-server", cfg)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Return an HTTP error — echo.NewHTTPError is the standard way
	handlerErr := echo.NewHTTPError(http.StatusInternalServerError, "something broke")
	_ = mw(func(_ *echo.Context) error {
		return handlerErr
	})(c)

	spans := recorder.Ended()
	require.Len(t, spans, 1)
	assert.Equal(t, codes.Error, spans[0].Status().Code, "error span status must be codes.Error")
}
