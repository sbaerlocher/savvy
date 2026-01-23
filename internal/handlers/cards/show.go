// Package cards contains HTTP request handlers for card operations.
package cards

import (
	"net/http"
	"savvy/internal/database"
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

	// Load card with relations
	var card models.Card
	if err := database.DB.Where("id = ?", cardID).Preload("Merchant").Preload("User").First(&card).Error; err != nil {
		return c.Redirect(http.StatusSeeOther, "/cards")
	}

	// Load shares with users (only if owner)
	var shares []models.CardShare
	if perms.IsOwner {
		database.DB.Where("card_id = ?", cardID).Preload("SharedWithUser").Find(&shares)
	}

	// Load all merchants for edit mode
	var merchants []models.Merchant
	database.DB.Order("name ASC").Find(&merchants)

	// Check if card is favorited by current user
	var favorite models.UserFavorite
	isFavorite := database.DB.Where("user_id = ? AND resource_type = ? AND resource_id = ?",
		user.ID, "card", cardID).First(&favorite).Error == nil

	csrfToken := c.Get("csrf").(string)

	view := views.CardShowView{
		Card:        card,
		Merchants:   merchants,
		Shares:      shares,
		User:        user,
		Permissions: views.CardPermissions{
			CanEdit:    perms.CanEdit,
			CanDelete:  perms.CanDelete,
			IsFavorite: isFavorite,
		},
		IsImpersonating: isImpersonating,
	}

	return templates.CardsShow(c.Request().Context(), csrfToken, view).Render(c.Request().Context(), c.Response().Writer)
}
