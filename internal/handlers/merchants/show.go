// Package merchants contains HTTP request handlers for merchant operations.
package merchants

import (
	"net/http"
	"savvy/internal/models"
	"savvy/internal/templates"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Show displays a single merchant
func (h *Handler) Show(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	isImpersonating := c.Get("is_impersonating") != nil

	// Get merchant ID from URL parameter
	merchantIDStr := c.Param("id")
	merchantID, err := uuid.Parse(merchantIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid merchant ID")
	}

	// Get merchant from service
	merchant, err := h.merchantService.GetMerchantByID(c.Request().Context(), merchantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Merchant not found")
	}

	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = ""
	}

	return templates.MerchantsShow(c.Request().Context(), csrfToken, *merchant, user, isImpersonating).Render(c.Request().Context(), c.Response().Writer)
}
