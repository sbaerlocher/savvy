package handlers

import (
	"net/http"
	"savvy/internal/database"

	"github.com/labstack/echo/v4"
)

// Health returns basic health status
func Health(c echo.Context) error {
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
func Ready(c echo.Context) error {
	// Check database connection
	sqlDB, err := database.DB.DB()
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
