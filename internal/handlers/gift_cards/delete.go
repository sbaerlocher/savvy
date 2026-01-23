package gift_cards

import (
	"net/http"
	"savvy/internal/audit"
	"savvy/internal/database"
	"savvy/internal/models"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Delete deletes a gift card
func (h *Handler) Delete(c echo.Context) error {
	user := c.Get("current_user").(*models.User)

	giftCardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/gift-cards")
	}

	// Check authorization
	perms, err := h.authzService.CheckGiftCardAccess(c.Request().Context(), user.ID, giftCardID)
	if err != nil || !perms.CanDelete {
		return c.Redirect(http.StatusSeeOther, "/gift-cards")
	}

	var giftCard models.GiftCard
	if err := database.DB.Where("id = ?", giftCardID).First(&giftCard).Error; err != nil {
		return c.Redirect(http.StatusSeeOther, "/gift-cards")
	}

	// Add user context for audit logging
	ctx := audit.AddUserIDToContext(c.Request().Context(), user.ID)
	if err := database.DB.WithContext(ctx).Delete(&giftCard).Error; err != nil {
		return c.Redirect(http.StatusSeeOther, "/gift-cards")
	}

	// Always use HX-Redirect header for consistent behavior
	c.Response().Header().Set("HX-Redirect", "/gift-cards")
	return c.Redirect(http.StatusSeeOther, "/gift-cards")
}
