// Package giftcards provides HTTP handlers for gift card management operations.
package giftcards

import (
	"net/http"
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

	giftCard, err := h.giftCardService.GetGiftCard(c.Request().Context(), giftCardID)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/gift-cards")
	}

	// Load all merchants for dropdown
	merchants, err := h.merchantService.GetAllMerchants(c.Request().Context())
	if err != nil {
		merchants = []models.Merchant{} // Fallback to empty list
	}

	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = ""
	}

	view := views.GiftCardEditView{
		GiftCard:        *giftCard,
		Merchants:       merchants,
		User:            user,
		IsImpersonating: isImpersonating,
	}

	return templates.GiftCardsEdit(c.Request().Context(), csrfToken, view).Render(c.Request().Context(), c.Response().Writer)
}
