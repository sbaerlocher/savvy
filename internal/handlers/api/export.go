// Package api contains JSON API handlers for the SvelteKit frontend.
package api

//
//nolint:revive // "api" is a meaningful package name for API handlers

import (
	"fmt"
	"net/http"
	"savvy/internal/models"
	"savvy/internal/services"
	"time"

	"github.com/labstack/echo/v4"
)

// ExportHandler handles data export API endpoints.
type ExportHandler struct {
	exportService services.ExportServiceInterface
}

// NewExportHandler creates a new export API handler.
func NewExportHandler(exportService services.ExportServiceInterface) *ExportHandler {
	return &ExportHandler{exportService: exportService}
}

// ExportData exports all user data as a JSON download.
// GET /api/v1/export
func (h *ExportHandler) ExportData(c echo.Context) error {
	user, ok := c.Get("current_user").(*models.User)
	if !ok || user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "Not authenticated",
		})
	}

	data, err := h.exportService.ExportUserData(c.Request().Context(), user.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "export_failed",
			Message: "Failed to export data",
		})
	}

	// Set download headers
	filename := fmt.Sprintf("savvy-export-%s.json", time.Now().Format("2006-01-02"))
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	c.Response().Header().Set("Content-Type", "application/json")

	return c.JSON(http.StatusOK, data)
}
