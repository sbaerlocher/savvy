// Package cards contains HTTP request handlers for card operations.
package cards

import (
	"savvy/internal/database"
	"savvy/internal/models"
	"savvy/internal/templates"
	"savvy/internal/views"

	"github.com/labstack/echo/v4"
)

// Index lists all cards for the current user (owned + shared).
func (h *Handler) Index(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	isImpersonating := c.Get("is_impersonating") != nil

	// Get owned cards with user info and merchant
	var ownedCards []models.Card
	if err := database.DB.Where("user_id = ?", user.ID).Preload("User").Preload("Merchant").Order("created_at DESC").Find(&ownedCards).Error; err != nil {
		return err
	}

	// Get shared cards with owner info and merchant
	var shares []models.CardShare
	database.DB.Where("shared_with_id = ?", user.ID).Preload("Card").Preload("Card.User").Preload("Card.Merchant").Find(&shares)

	// Extract shared cards
	var sharedCards []models.Card
	for _, share := range shares {
		if share.Card != nil {
			sharedCards = append(sharedCards, *share.Card)
		}
	}

	// Combine owned and shared
	ownedCards = append(ownedCards, sharedCards...)
	allCards := ownedCards

	view := views.CardIndexView{
		Cards:           allCards,
		User:            user,
		IsImpersonating: isImpersonating,
	}

	return templates.CardsIndex(c.Request().Context(), view).Render(c.Request().Context(), c.Response().Writer)
}
