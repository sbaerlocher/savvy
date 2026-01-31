// Package cards contains HTTP request handlers for card operations.
package cards

import (
	"savvy/internal/models"
	"savvy/internal/templates"
	"savvy/internal/views"

	"github.com/labstack/echo/v4"
)

// Index lists all cards for the current user (owned + shared).
func (h *Handler) Index(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	isImpersonating := c.Get("is_impersonating") != nil

	// Get all cards (owned + shared) via service
	allCards, err := h.cardService.GetUserCards(c.Request().Context(), user.ID)
	if err != nil {
		return err
	}

	view := views.CardIndexView{
		Cards:           allCards,
		User:            user,
		IsImpersonating: isImpersonating,
	}

	return templates.CardsIndex(c.Request().Context(), view).Render(c.Request().Context(), c.Response().Writer)
}
