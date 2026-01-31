// Package giftcards provides HTTP handlers for gift card management operations.
package giftcards

import (
	"savvy/internal/database"
	"savvy/internal/models"
	"savvy/internal/templates"
	"savvy/internal/views"

	"github.com/labstack/echo/v4"
)

// New shows the form to create a new gift card
func (h *Handler) New(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	isImpersonating := c.Get("is_impersonating") != nil
	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = ""
	}

	// Load all merchants for dropdown
	var merchants []models.Merchant
	database.DB.Order("name ASC").Find(&merchants)

	view := views.GiftCardEditView{
		GiftCard:        models.GiftCard{},
		Merchants:       merchants,
		User:            user,
		IsImpersonating: isImpersonating,
	}

	return templates.GiftCardsNew(c.Request().Context(), csrfToken, view).Render(c.Request().Context(), c.Response().Writer)
}
