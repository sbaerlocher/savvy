package gift_cards

import (
	"net/http"
	"savvy/internal/database"
	"savvy/internal/models"
	"savvy/internal/templates"
	"savvy/internal/views"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Edit shows the form to edit a gift card
func (h *Handler) Edit(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	isImpersonating := c.Get("is_impersonating") != nil

	giftCardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/gift-cards")
	}

	// Check authorization
	perms, err := h.authzService.CheckGiftCardAccess(c.Request().Context(), user.ID, giftCardID)
	if err != nil || !perms.CanEdit {
		return c.Redirect(http.StatusSeeOther, "/gift-cards")
	}

	var giftCard models.GiftCard
	if err := database.DB.Where("id = ?", giftCardID).First(&giftCard).Error; err != nil {
		return c.Redirect(http.StatusSeeOther, "/gift-cards")
	}

	// Load all merchants for dropdown
	var merchants []models.Merchant
	database.DB.Order("name ASC").Find(&merchants)

	csrfToken := c.Get("csrf").(string)

	view := views.GiftCardEditView{
		GiftCard:        giftCard,
		Merchants:       merchants,
		User:            user,
		IsImpersonating: isImpersonating,
	}

	return templates.GiftCardsEdit(c.Request().Context(), csrfToken, view).Render(c.Request().Context(), c.Response().Writer)
}
