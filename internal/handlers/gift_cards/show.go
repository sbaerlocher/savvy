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

// Show displays a single gift card with transactions
func (h *Handler) Show(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	isImpersonating := c.Get("is_impersonating") != nil

	giftCardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/gift-cards")
	}

	// Check authorization
	perms, err := h.authzService.CheckGiftCardAccess(c.Request().Context(), user.ID, giftCardID)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/gift-cards")
	}

	var giftCard models.GiftCard
	// Preload merchant, user, and transactions
	if err := database.DB.Where("id = ?", giftCardID).Preload("Merchant").Preload("User").Preload("Transactions").First(&giftCard).Error; err != nil {
		return c.Redirect(http.StatusSeeOther, "/gift-cards")
	}

	// Load shares with users (only if owner)
	var shares []models.GiftCardShare
	if perms.IsOwner {
		database.DB.Where("gift_card_id = ?", giftCardID).Preload("SharedWithUser").Find(&shares)
	}

	// Load all merchants for dropdown (used in inline edit)
	var merchants []models.Merchant
	database.DB.Order("name ASC").Find(&merchants)

	// Check if gift card is favorited by current user
	var favorite models.UserFavorite
	isFavorite := database.DB.Where("user_id = ? AND resource_type = ? AND resource_id = ?",
		user.ID, "gift_card", giftCardID).First(&favorite).Error == nil

	csrfToken := c.Get("csrf").(string)

	view := views.GiftCardShowView{
		GiftCard:    giftCard,
		Merchants:   merchants,
		Shares:      shares,
		User:        user,
		Permissions: views.GiftCardPermissions{
			CanEdit:             perms.CanEdit,
			CanDelete:           perms.CanDelete,
			CanEditTransactions: perms.CanEditTransactions,
			IsFavorite:          isFavorite,
		},
		IsImpersonating: isImpersonating,
	}

	return templates.GiftCardsShow(c.Request().Context(), csrfToken, view).Render(c.Request().Context(), c.Response().Writer)
}
