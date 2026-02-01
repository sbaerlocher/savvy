// Package merchants contains HTTP request handlers for merchant operations.
package merchants

import (
	"net/http"
	"savvy/internal/models"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Delete deletes a merchant
func (h *Handler) Delete(c echo.Context) error {
	user := c.Get("current_user").(*models.User)

	merchantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/merchants")
	}

	// Verify merchant exists before deleting
	_, err = h.merchantService.GetMerchantByID(c.Request().Context(), merchantID)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/merchants")
	}

	// Add user context for audit logging
	ctx := c.Request().Context()

	// Delete the merchant
	if err := h.merchantService.DeleteMerchant(ctx, merchantID); err != nil {
		c.Logger().Errorf("Failed to delete merchant %s: %v", merchantID, err)
		return c.Redirect(http.StatusSeeOther, "/merchants")
	}

	// Log for audit trail
	c.Logger().Infof("User %s deleted merchant %s", user.ID, merchantID)

	// Always use HX-Redirect header for consistent behavior
	c.Response().Header().Set("HX-Redirect", "/merchants")
	return c.Redirect(http.StatusSeeOther, "/merchants")
}
