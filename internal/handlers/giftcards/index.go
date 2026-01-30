// Package giftcards provides HTTP handlers for gift card management operations.
package giftcards

import (
	"savvy/internal/database"
	"savvy/internal/models"
	"savvy/internal/templates"
	"savvy/internal/views"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// Index lists all gift cards for the current user (owned + shared)
func (h *Handler) Index(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	isImpersonating := c.Get("is_impersonating") != nil

	// Get owned gift cards with merchant info
	var ownedGiftCards []models.GiftCard
	if err := database.DB.Where("user_id = ?", user.ID).Preload("User").Preload("Merchant").Preload("Transactions", func(db *gorm.DB) *gorm.DB {
		return db.Order("transaction_date DESC").Limit(10)
	}).Order("created_at DESC").Find(&ownedGiftCards).Error; err != nil {
		return err
	}

	// Get shared gift card IDs with merchant info
	var shares []models.GiftCardShare
	database.DB.Where("shared_with_id = ?", user.ID).Preload("GiftCard").Preload("GiftCard.User").Preload("GiftCard.Merchant").Preload("GiftCard.Transactions", func(db *gorm.DB) *gorm.DB {
		return db.Order("transaction_date DESC").Limit(10)
	}).Find(&shares)

	// Extract shared gift cards
	var sharedGiftCards []models.GiftCard
	for _, share := range shares {
		if share.GiftCard != nil {
			sharedGiftCards = append(sharedGiftCards, *share.GiftCard)
		}
	}

	// Combine owned and shared
	ownedGiftCards = append(ownedGiftCards, sharedGiftCards...)
	allGiftCards := ownedGiftCards

	view := views.GiftCardIndexView{
		GiftCards:       allGiftCards,
		User:            user,
		IsImpersonating: isImpersonating,
	}

	return templates.GiftCardsIndex(c.Request().Context(), view).Render(c.Request().Context(), c.Response().Writer)
}
