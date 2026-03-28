// Package handlers contains HTTP request handlers.
package handlers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"savvy/internal/config"
	"strings"

	"github.com/labstack/echo/v5"
)

// maxOTelBodySize limits the request body size for OTel proxy requests (1 MB).
const maxOTelBodySize = 1 << 20

// OTelProxyHandler handles OTLP proxy requests from frontend
type OTelProxyHandler struct {
	config *config.Config
}

// NewOTelProxyHandler creates a new OTEL proxy handler
func NewOTelProxyHandler(cfg *config.Config) *OTelProxyHandler {
	return &OTelProxyHandler{
		config: cfg,
	}
}

// ProxyTraces proxies OTLP trace requests to the OTEL Collector
func (h *OTelProxyHandler) ProxyTraces(c *echo.Context) error {
	return h.proxyOTLP(c, "traces")
}

// ProxyLogs proxies OTLP log requests to the OTEL Collector
func (h *OTelProxyHandler) ProxyLogs(c *echo.Context) error {
	return h.proxyOTLP(c, "logs")
}

// ProxyMetrics proxies OTLP metric requests to the OTEL Collector
func (h *OTelProxyHandler) ProxyMetrics(c *echo.Context) error {
	return h.proxyOTLP(c, "metrics")
}

// proxyOTLP forwards OTLP requests to the OTEL Collector
func (h *OTelProxyHandler) proxyOTLP(c *echo.Context, signalType string) error {
	// Skip if OTEL is disabled
	if !h.config.OTelEnabled {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"error": "OpenTelemetry is disabled",
		})
	}

	// Read request body with size limit to prevent DoS
	body, err := io.ReadAll(io.LimitReader(c.Request().Body, maxOTelBodySize))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Failed to read request body",
		})
	}

	// Build OTLP endpoint URL
	// Format: http://otel-collector:4318/v1/{signals}
	// Example: http://otel-collector:4318/v1/traces
	// Note: OTelEndpoint is "otel-collector:4318" or "localhost:4318"
	endpoint := h.config.OTelEndpoint
	// Ensure endpoint has http:// prefix
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		endpoint = "http://" + endpoint
	}
	otlpURL := fmt.Sprintf("%s/v1/%s", endpoint, signalType)

	// Create HTTP request to OTEL Collector
	req, err := http.NewRequestWithContext(c.Request().Context(), "POST", otlpURL, bytes.NewReader(body))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create request",
		})
	}

	// Set a whitelisted Content-Type rather than forwarding the user-supplied value.
	// OTLP collectors accept "application/x-protobuf" (binary) and "application/json".
	ct := c.Request().Header.Get("Content-Type")
	switch ct {
	case "application/x-protobuf", "application/json":
		// accepted OTLP content types
	default:
		ct = "application/x-protobuf"
	}
	req.Header.Set("Content-Type", ct)

	// Send request to OTEL Collector
	client := &http.Client{}
	resp, err := client.Do(req) // #nosec G704 -- URL is from trusted config, not user input
	if err != nil {
		return c.JSON(http.StatusBadGateway, map[string]string{
			"error": "Failed to reach OTEL Collector",
		})
	}
	defer func() { _ = resp.Body.Close() }()

	// Read response body with size limit
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxOTelBodySize))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to read response",
		})
	}

	// Return response from OTEL Collector
	return c.Blob(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
}
