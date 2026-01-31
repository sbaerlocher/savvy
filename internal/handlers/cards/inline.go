// Package cards contains HTTP request handlers for card operations.
package cards

import (
	"net/http"
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

	card, err := h.cardService.GetCard(c.Request().Context(), cardID)
	if err != nil {
		return c.String(http.StatusNotFound, "Card not found")
	}

	// Load all merchants for dropdown
	merchants, err := h.merchantService.GetAllMerchants(c.Request().Context())
	if err != nil {
		c.Logger().Errorf("Failed to load merchants: %v", err)
		merchants = []models.Merchant{}
	}

	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = ""
	}
	component := templates.CardDetailEdit(c.Request().Context(), csrfToken, *card, merchants)
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

	card, err := h.cardService.GetCard(c.Request().Context(), cardID)
	if err != nil {
		return c.String(http.StatusNotFound, "Card not found")
	}

	canEdit := perms.CanEdit

	// Check if card is favorited by current user
	isFavorite, err := h.favoriteService.IsFavorite(c.Request().Context(), user.ID, "card", cardID)
	if err != nil {
		c.Logger().Errorf("Failed to check favorite status: %v", err)
		isFavorite = false
	}

	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = ""
	}
	component := templates.CardDetailView(c.Request().Context(), csrfToken, *card, user, canEdit, isFavorite)
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

	card, err := h.cardService.GetCard(c.Request().Context(), cardID)
	if err != nil {
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
			merchant, err := h.merchantService.GetMerchantByID(c.Request().Context(), merchantID)
			if err == nil {
				card.MerchantName = merchant.Name
			}
		}
	} else {
		// New merchant name entered or no selection
		card.MerchantID = nil
		card.MerchantName = c.FormValue("merchant_name")
	}

	if err := h.cardService.UpdateCard(c.Request().Context(), card); err != nil {
		return c.String(http.StatusInternalServerError, "Error updating card")
	}

	// Reload with merchant and user for display
	card, err = h.cardService.GetCard(c.Request().Context(), cardID)
	if err != nil {
		return c.String(http.StatusNotFound, "Card not found")
	}

	// Check if card is favorited by current user
	isFavorite, err := h.favoriteService.IsFavorite(c.Request().Context(), user.ID, "card", cardID)
	if err != nil {
		c.Logger().Errorf("Failed to check favorite status: %v", err)
		isFavorite = false
	}

	canEdit := perms.CanEdit
	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = ""
	}
	component := templates.CardDetailView(c.Request().Context(), csrfToken, *card, user, canEdit, isFavorite)
	return component.Render(c.Request().Context(), c.Response().Writer)
}
