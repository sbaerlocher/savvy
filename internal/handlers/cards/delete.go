// Package cards contains HTTP request handlers for card operations.
package cards

import (
	"net/http"
	"savvy/internal/audit"
	"savvy/internal/database"
	"savvy/internal/models"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Delete deletes a card
func (h *Handler) Delete(c echo.Context) error {
	user := c.Get("current_user").(*models.User)

	cardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/cards")
	}

	// Check authorization
	perms, err := h.authzService.CheckCardAccess(c.Request().Context(), user.ID, cardID)
	if err != nil || !perms.CanDelete {
		return c.Redirect(http.StatusSeeOther, "/cards")
	}

	var card models.Card
	if err := database.DB.Where("id = ?", cardID).First(&card).Error; err != nil {
		return c.Redirect(http.StatusSeeOther, "/cards")
	}

	// Add user context for audit logging
	ctx := audit.AddUserIDToContext(c.Request().Context(), user.ID)
	if err := database.DB.WithContext(ctx).Delete(&card).Error; err != nil {
		return c.Redirect(http.StatusSeeOther, "/cards")
	}

	// Always use HX-Redirect header for consistent behavior
	c.Response().Header().Set("HX-Redirect", "/cards")
	return c.Redirect(http.StatusSeeOther, "/cards")
}
