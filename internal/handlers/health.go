// Package handlers contains HTTP request handlers for the savvy system.
package handlers

import (
	"net/http"
	"savvy/internal/services"

	"github.com/labstack/echo/v4"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	healthService services.HealthCheckServiceInterface
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(healthService services.HealthCheckServiceInterface) *HealthHandler {
	return &HealthHandler{
		healthService: healthService,
	}
}

// Health returns minimal health status (public endpoint - minimal information disclosure)
func (h *HealthHandler) Health(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": "ok",
	})
}

// Ready checks if the service is ready to accept requests
func (h *HealthHandler) Ready(c echo.Context) error {
	ctx := c.Request().Context()

	report, err := h.healthService.CheckReadiness(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"status": "error",
			"error":  err.Error(),
		})
	}

	// Determine HTTP status code based on overall status
	var httpStatus int
	switch report.Status {
	case "ready", "degraded":
		httpStatus = http.StatusOK // 200 - Pod receives traffic
	case "not_ready":
		httpStatus = http.StatusServiceUnavailable // 503 - Pod does not receive traffic
	default:
		httpStatus = http.StatusInternalServerError // 500 - Unexpected
	}

	return c.JSON(httpStatus, report)
}
