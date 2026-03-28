package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"savvy/internal/config"
)

// errorReader is an io.Reader that always returns an error.
type errorReader struct{}

func (errorReader) Read(_ []byte) (int, error) {
	return 0, errors.New("read error")
}

func createOTelTestContext(path, body string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, "application/x-protobuf")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

func TestOTelProxy_Disabled(t *testing.T) {
	cfg := &config.Config{OTelEnabled: false}
	handler := NewOTelProxyHandler(cfg)
	c, rec := createOTelTestContext("/api/v1/otel/traces", "")

	err := handler.ProxyTraces(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.Contains(t, rec.Body.String(), "OpenTelemetry is disabled")
}

func TestOTelProxy_DisabledLogs(t *testing.T) {
	cfg := &config.Config{OTelEnabled: false}
	handler := NewOTelProxyHandler(cfg)
	c, rec := createOTelTestContext("/api/v1/otel/logs", "")

	err := handler.ProxyLogs(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
}

func TestOTelProxy_DisabledMetrics(t *testing.T) {
	cfg := &config.Config{OTelEnabled: false}
	handler := NewOTelProxyHandler(cfg)
	c, rec := createOTelTestContext("/api/v1/otel/metrics", "")

	err := handler.ProxyMetrics(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
}

func TestOTelProxy_SuccessfulProxy(t *testing.T) {
	// Start a mock OTel collector
	collector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/traces", r.URL.Path)
		assert.Equal(t, "application/x-protobuf", r.Header.Get("Content-Type"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"partialSuccess":{}}`))
	}))
	defer collector.Close()

	cfg := &config.Config{
		OTelEnabled:  true,
		OTelEndpoint: collector.URL, // Already has http:// prefix
	}
	handler := NewOTelProxyHandler(cfg)
	c, rec := createOTelTestContext("/api/v1/otel/traces", `{"resourceSpans":[]}`)

	err := handler.ProxyTraces(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "partialSuccess")
}

func TestOTelProxy_EndpointWithoutScheme(t *testing.T) {
	collector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer collector.Close()

	// Strip http:// prefix to test auto-prefixing
	endpoint := strings.TrimPrefix(collector.URL, "http://")
	cfg := &config.Config{
		OTelEnabled:  true,
		OTelEndpoint: endpoint,
	}
	handler := NewOTelProxyHandler(cfg)
	c, rec := createOTelTestContext("/api/v1/otel/logs", `{}`)

	err := handler.ProxyLogs(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestOTelProxy_EmptyEndpoint(t *testing.T) {
	// Empty endpoint should not panic (previously would panic with slice indexing)
	cfg := &config.Config{
		OTelEnabled:  true,
		OTelEndpoint: "",
	}
	handler := NewOTelProxyHandler(cfg)
	c, rec := createOTelTestContext("/api/v1/otel/traces", `{}`)

	err := handler.ProxyTraces(c)

	// Should fail gracefully (bad gateway or internal error), not panic
	assert.NoError(t, err)
	assert.True(t, rec.Code == http.StatusBadGateway || rec.Code == http.StatusInternalServerError)
}

func TestOTelProxy_ShortEndpoint(t *testing.T) {
	// Short endpoint like "a" should not panic (previously would panic with endpoint[:7])
	cfg := &config.Config{
		OTelEnabled:  true,
		OTelEndpoint: "a",
	}
	handler := NewOTelProxyHandler(cfg)
	c, rec := createOTelTestContext("/api/v1/otel/metrics", `{}`)

	err := handler.ProxyMetrics(c)

	assert.NoError(t, err)
	// Will fail to connect but should not panic.
	// Status depends on DNS resolution: 502 (connection refused) or 404 (host resolved but no route).
	assert.True(t, rec.Code >= 400, "expected error status code, got %d", rec.Code)
}

func TestOTelProxy_HttpsEndpoint(t *testing.T) {
	// Verify https:// prefix is preserved
	cfg := &config.Config{
		OTelEnabled:  true,
		OTelEndpoint: "https://unreachable.example.com:4318",
	}
	handler := NewOTelProxyHandler(cfg)
	c, rec := createOTelTestContext("/api/v1/otel/traces", `{}`)

	err := handler.ProxyTraces(c)

	assert.NoError(t, err)
	// Will fail to connect but should use https prefix, not double-prefix
	assert.Equal(t, http.StatusBadGateway, rec.Code)
}

func TestOTelProxy_CollectorUnreachable(t *testing.T) {
	cfg := &config.Config{
		OTelEnabled:  true,
		OTelEndpoint: "http://localhost:19999", // Non-existent port
	}
	handler := NewOTelProxyHandler(cfg)
	c, rec := createOTelTestContext("/api/v1/otel/traces", `{"data":"test"}`)

	err := handler.ProxyTraces(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadGateway, rec.Code)
	assert.Contains(t, rec.Body.String(), "Failed to reach OTEL Collector")
}

func TestOTelProxy_BodySizeLimit(t *testing.T) {
	var receivedBodySize int
	collector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := make([]byte, 2<<20) // 2 MB buffer
		n, _ := r.Body.Read(body)
		receivedBodySize = n
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer collector.Close()

	cfg := &config.Config{
		OTelEnabled:  true,
		OTelEndpoint: collector.URL,
	}
	handler := NewOTelProxyHandler(cfg)

	// Send body larger than maxOTelBodySize (1 MB)
	largeBody := strings.Repeat("x", 2<<20) // 2 MB
	c, rec := createOTelTestContext("/api/v1/otel/traces", largeBody)

	err := handler.ProxyTraces(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	// The body forwarded should be truncated to maxOTelBodySize (1 MB)
	assert.LessOrEqual(t, receivedBodySize, maxOTelBodySize)
}

func TestOTelProxy_CollectorErrorResponse(t *testing.T) {
	collector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":"rate limited"}`))
	}))
	defer collector.Close()

	cfg := &config.Config{
		OTelEnabled:  true,
		OTelEndpoint: collector.URL,
	}
	handler := NewOTelProxyHandler(cfg)
	c, rec := createOTelTestContext("/api/v1/otel/traces", `{}`)

	err := handler.ProxyTraces(c)

	assert.NoError(t, err)
	// Should forward the collector's status code
	assert.Equal(t, http.StatusTooManyRequests, rec.Code)
	assert.Contains(t, rec.Body.String(), "rate limited")
}

func TestOTelProxy_ContentTypeForwarding(t *testing.T) {
	var receivedContentType string
	collector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedContentType = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer collector.Close()

	cfg := &config.Config{
		OTelEnabled:  true,
		OTelEndpoint: collector.URL,
	}
	handler := NewOTelProxyHandler(cfg)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/otel/traces", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.ProxyTraces(c)

	assert.NoError(t, err)
	assert.Equal(t, "application/json", receivedContentType)
}

func TestOTelProxy_RequestBodyReadError(t *testing.T) {
	cfg := &config.Config{
		OTelEnabled:  true,
		OTelEndpoint: "http://localhost:4318",
	}
	handler := NewOTelProxyHandler(cfg)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/otel/traces", errorReader{})
	req.Header.Set("Content-Type", "application/x-protobuf")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.ProxyTraces(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Failed to read request body")
}

func TestMaxOTelBodySize(t *testing.T) {
	// Verify the constant is 1 MB
	assert.Equal(t, 1<<20, maxOTelBodySize)
}
