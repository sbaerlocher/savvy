// Package cards contains HTTP request handlers for card operations.
package cards

import (
	"net/http"
	"savvy/internal/database"
	"savvy/internal/models"
	"savvy/internal/templates"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// EditInline returns the inline edit form
func (h *Handler) EditInline(c echo.Context) error {
	user := c.Get("current_user").(*models.User)

	cardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.String(http.StatusNotFound, "Card not found")
	}

	// Check authorization
	perms, err := h.authzService.CheckCardAccess(c.Request().Context(), user.ID, cardID)
	if err != nil || !perms.CanEdit {
		return c.String(http.StatusForbidden, "Not authorized")
	}

	var card models.Card
	if err := database.DB.Where("id = ?", cardID).Preload("Merchant").First(&card).Error; err != nil {
		return c.String(http.StatusNotFound, "Card not found")
	}

	// Load all merchants for dropdown
	var merchants []models.Merchant
	database.DB.Order("name ASC").Find(&merchants)

	csrfToken := c.Get("csrf").(string)
	component := templates.CardDetailEdit(c.Request().Context(), csrfToken, card, merchants)
	return component.Render(c.Request().Context(), c.Response().Writer)
}

// CancelEdit cancels inline editing and returns to display
func (h *Handler) CancelEdit(c echo.Context) error {
	user := c.Get("current_user").(*models.User)

	cardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.String(http.StatusNotFound, "Card not found")
	}

	// Check authorization
	perms, err := h.authzService.CheckCardAccess(c.Request().Context(), user.ID, cardID)
	if err != nil {
		return c.String(http.StatusNotFound, "Card not found")
	}

	var card models.Card
	if err := database.DB.Where("id = ?", cardID).Preload("Merchant").Preload("User").First(&card).Error; err != nil {
		return c.String(http.StatusNotFound, "Card not found")
	}

	canEdit := perms.CanEdit

	// Check if card is favorited by current user
	var favorite models.UserFavorite
	isFavorite := database.DB.Where("user_id = ? AND resource_type = ? AND resource_id = ?",
		user.ID, "card", cardID).First(&favorite).Error == nil

	csrfToken := c.Get("csrf").(string)
	component := templates.CardDetailView(c.Request().Context(), csrfToken, card, user, canEdit, isFavorite)
	return component.Render(c.Request().Context(), c.Response().Writer)
}

// UpdateInline updates a card and returns the view
func (h *Handler) UpdateInline(c echo.Context) error {
	user := c.Get("current_user").(*models.User)

	cardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.String(http.StatusNotFound, "Card not found")
	}

	// Check authorization
	perms, err := h.authzService.CheckCardAccess(c.Request().Context(), user.ID, cardID)
	if err != nil || !perms.CanEdit {
		return c.String(http.StatusForbidden, "Not authorized")
	}

	var card models.Card
	if err := database.DB.Where("id = ?", cardID).First(&card).Error; err != nil {
		return c.String(http.StatusNotFound, "Card not found")
	}

	// Update fields
	card.Program = c.FormValue("program")
	card.CardNumber = c.FormValue("card_number")
	card.BarcodeType = c.FormValue("barcode_type")
	card.Status = c.FormValue("status")
	card.Notes = c.FormValue("notes")

	// Handle merchant selection
	merchantIDStr := c.FormValue("merchant_id")
	if merchantIDStr != "" && merchantIDStr != "new" {
		// Existing merchant selected from dropdown
		merchantID, err := uuid.Parse(merchantIDStr)
		if err == nil {
			card.MerchantID = &merchantID
			// Load merchant to get name
			var merchant models.Merchant
			if err := database.DB.Where("id = ?", merchantID).First(&merchant).Error; err == nil {
				card.MerchantName = merchant.Name
			}
		}
	} else {
		// New merchant name entered or no selection
		card.MerchantID = nil
		card.MerchantName = c.FormValue("merchant_name")
	}

	if err := database.DB.Save(&card).Error; err != nil {
		return c.String(http.StatusInternalServerError, "Error updating card")
	}

	// Reload with merchant and user for display
	database.DB.Where("id = ?", cardID).Preload("Merchant").Preload("User").First(&card)

	// Check if card is favorited by current user
	var favorite models.UserFavorite
	isFavorite := database.DB.Where("user_id = ? AND resource_type = ? AND resource_id = ?",
		user.ID, "card", cardID).First(&favorite).Error == nil

	canEdit := perms.CanEdit
	csrfToken := c.Get("csrf").(string)
	component := templates.CardDetailView(c.Request().Context(), csrfToken, card, user, canEdit, isFavorite)
	return component.Render(c.Request().Context(), c.Response().Writer)
}
