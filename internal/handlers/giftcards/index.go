// Package giftcards provides HTTP handlers for gift card management operations.
package giftcards

import (
	"savvy/internal/models"
	"savvy/internal/templates"
	"savvy/internal/views"

	"github.com/labstack/echo/v5"
)

// Index lists all gift cards for the current user (owned + shared)
func (h *Handler) Index(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	isImpersonating := c.Get("is_impersonating") != nil

	// Get all gift cards (owned + shared) via service
	allGiftCards, err := h.giftCardService.GetUserGiftCards(c.Request().Context(), user.ID)
	if err != nil {
		return err
	}

	view := views.GiftCardIndexView{
		GiftCards:       allGiftCards,
		User:            user,
		IsImpersonating: isImpersonating,
	}

	return templates.GiftCardsIndex(c.Request().Context(), view).Render(c.Request().Context(), c.Response().Writer)
}
