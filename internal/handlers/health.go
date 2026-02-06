// Package handlers contains HTTP request handlers for the savvy system.
package handlers

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"gorm.io/gorm"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	db *gorm.DB
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{
		db: db,
	}
}

// Health returns basic health status
func (h *HealthHandler) Health(c echo.Context) error {
	// Get service version from context (set in main.go)
	version := c.Get("service_version")
	if version == nil {
		version = "unknown"
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "healthy",
		"version": version,
		"service": "savvy",
	})
}

// Ready checks if the service is ready to accept requests
func (h *HealthHandler) Ready(c echo.Context) error {
	// Check database connection
	sqlDB, err := h.db.DB()
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"status": "not ready",
			"reason": "database connection error",
		})
	}

	// Ping database
	if err := sqlDB.Ping(); err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"status": "not ready",
			"reason": "database ping failed",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": "ready",
	})
}
