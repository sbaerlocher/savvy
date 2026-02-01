// Package cards contains HTTP request handlers for card operations.
package cards

import (
	"net/http"
	"savvy/internal/models"
	"savvy/internal/templates"
	"savvy/internal/views"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Show displays a single card
func (h *Handler) Show(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	isImpersonating := c.Get("is_impersonating") != nil

	cardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/cards")
	}

	// Check authorization
	perms, err := h.authzService.CheckCardAccess(c.Request().Context(), user.ID, cardID)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/cards")
	}

	card, err := h.cardService.GetCard(c.Request().Context(), cardID)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/cards")
	}

	var shares []models.CardShare
	if perms.IsOwner {
		shares, _ = h.shareService.GetCardShares(c.Request().Context(), cardID)
	}

	merchants, err := h.merchantService.GetAllMerchants(c.Request().Context())
	if err != nil {
		merchants = []models.Merchant{}
	}

	// Check if card is favorited by current user
	isFavorite, _ := h.favoriteService.IsFavorite(c.Request().Context(), user.ID, "card", cardID)

	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = ""
	}

	view := views.CardShowView{
		Card:      *card,
		Merchants: merchants,
		Shares:    shares,
		User:      user,
		Permissions: views.CardPermissions{
			CanEdit:    perms.CanEdit,
			CanDelete:  perms.CanDelete,
			IsFavorite: isFavorite,
		},
		IsImpersonating: isImpersonating,
	}

	return templates.CardsShow(c.Request().Context(), csrfToken, view).Render(c.Request().Context(), c.Response().Writer)
}
