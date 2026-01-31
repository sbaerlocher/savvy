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

	// Load gift card with relations via service
	giftCard, err := h.giftCardService.GetGiftCard(c.Request().Context(), giftCardID)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/gift-cards")
	}

	// Load shares with users (only if owner)
	var shares []models.GiftCardShare
	if perms.IsOwner {
		shares, _ = h.shareService.GetGiftCardShares(c.Request().Context(), giftCardID)
	}

	// Load all merchants for dropdown (used in inline edit)
	merchants, err := h.merchantService.GetAllMerchants(c.Request().Context())
	if err != nil {
		merchants = []models.Merchant{} // Fallback to empty list
	}

	// Check if gift card is favorited by current user
	isFavorite, _ := h.favoriteService.IsFavorite(c.Request().Context(), user.ID, "gift_card", giftCardID)

	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = ""
	}

	view := views.GiftCardShowView{
		GiftCard:  *giftCard,
		Merchants: merchants,
		Shares:    shares,
		User:      user,
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
