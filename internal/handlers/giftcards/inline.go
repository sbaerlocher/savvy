// Package giftcards provides HTTP handlers for gift card management operations.
package giftcards

import (
	"net/http"
	"savvy/internal/i18n"
	"savvy/internal/models"
	"savvy/internal/templates"
	"savvy/internal/validation"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

// EditInline returns the inline edit form
func (h *Handler) EditInline(c echo.Context) error {
	user := c.Get("current_user").(*models.User)

	giftCardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.String(http.StatusNotFound, i18n.T(c.Request().Context(), "error.gift_card_not_found"))
	}

	// Check authorization
	perms, err := h.authzService.CheckGiftCardAccess(c.Request().Context(), user.ID, giftCardID)
	if err != nil || !perms.CanEdit {
		return c.String(http.StatusForbidden, i18n.T(c.Request().Context(), "error.unauthorized"))
	}

	giftCard, err := h.giftCardService.GetGiftCard(c.Request().Context(), giftCardID)
	if err != nil {
		return c.String(http.StatusNotFound, i18n.T(c.Request().Context(), "error.gift_card_not_found"))
	}

	merchants, err := h.merchantService.GetAllMerchants(c.Request().Context())
	if err != nil {
		c.Logger().Errorf("Failed to load merchants: %v", err)
		merchants = []models.Merchant{}
	}

	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = ""
	}
	component := templates.GiftCardDetailEdit(c.Request().Context(), csrfToken, *giftCard, merchants)
	return component.Render(c.Request().Context(), c.Response().Writer)
}

// CancelEdit cancels inline editing and returns to display
func (h *Handler) CancelEdit(c echo.Context) error {
	user := c.Get("current_user").(*models.User)

	giftCardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.String(http.StatusNotFound, i18n.T(c.Request().Context(), "error.gift_card_not_found"))
	}

	// Check authorization
	perms, err := h.authzService.CheckGiftCardAccess(c.Request().Context(), user.ID, giftCardID)
	if err != nil {
		return c.String(http.StatusNotFound, i18n.T(c.Request().Context(), "error.gift_card_not_found"))
	}

	giftCard, err := h.giftCardService.GetGiftCard(c.Request().Context(), giftCardID)
	if err != nil {
		return c.String(http.StatusNotFound, i18n.T(c.Request().Context(), "error.gift_card_not_found"))
	}

	canEdit := perms.CanEdit

	isFavorite, err := h.favoriteService.IsFavorite(c.Request().Context(), user.ID, "gift_card", giftCardID)
	if err != nil {
		c.Logger().Errorf("Failed to check favorite status: %v", err)
		isFavorite = false
	}

	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = ""
	}
	component := templates.GiftCardDetailView(c.Request().Context(), csrfToken, *giftCard, user, canEdit, isFavorite)
	return component.Render(c.Request().Context(), c.Response().Writer)
}

// UpdateInline updates a gift card and returns the view
func (h *Handler) UpdateInline(c echo.Context) error {
	user := c.Get("current_user").(*models.User)

	giftCardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.String(http.StatusNotFound, i18n.T(c.Request().Context(), "error.gift_card_not_found"))
	}

	// Check authorization
	perms, err := h.authzService.CheckGiftCardAccess(c.Request().Context(), user.ID, giftCardID)
	if err != nil || !perms.CanEdit {
		return c.String(http.StatusForbidden, i18n.T(c.Request().Context(), "error.unauthorized"))
	}

	giftCard, err := h.giftCardService.GetGiftCard(c.Request().Context(), giftCardID)
	if err != nil {
		return c.String(http.StatusNotFound, i18n.T(c.Request().Context(), "error.gift_card_not_found"))
	}

	// Parse balance
	initialBalance, _ := strconv.ParseFloat(c.FormValue("initial_balance"), 64)

	// Parse expiration date
	var expiresAt *time.Time
	if expiresAtStr := c.FormValue("expires_at"); expiresAtStr != "" {
		parsed, err := validation.ParseAndValidateDate(expiresAtStr, true) // allow past for flexibility
		if err != nil {
			c.Logger().Errorf("Expires_at validation failed: %v", err)
			return c.String(http.StatusBadRequest, i18n.T(c.Request().Context(), "error.invalid_expiration_date"))
		}
		// Set to end of day
		parsed = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 0, time.UTC)
		expiresAt = &parsed
	}

	merchantIDStr := c.FormValue("merchant_id")
	merchantNameStr := c.FormValue("merchant_name")

	// Update fields
	giftCard.CardNumber = c.FormValue("card_number")
	giftCard.InitialBalance = initialBalance
	giftCard.Currency = c.FormValue("currency")
	giftCard.PIN = c.FormValue("pin")
	giftCard.ExpiresAt = expiresAt
	giftCard.BarcodeType = c.FormValue("barcode_type")
	giftCard.Notes = c.FormValue("notes")

	// Handle merchant selection
	if merchantIDStr != "" && merchantIDStr != "new" {
		// Existing merchant selected from dropdown
		merchantID, err := uuid.Parse(merchantIDStr)
		if err == nil {
			giftCard.MerchantID = &merchantID

			merchant, err := h.merchantService.GetMerchantByID(c.Request().Context(), merchantID)
			if err == nil {
				giftCard.MerchantName = merchant.Name
			}
		}
	} else {
		// New merchant name entered or no selection
		giftCard.MerchantID = nil
		giftCard.MerchantName = merchantNameStr
	}

	if err := h.giftCardService.UpdateGiftCard(c.Request().Context(), giftCard); err != nil {
		return c.String(http.StatusInternalServerError, i18n.T(c.Request().Context(), "error.updating_gift_card"))
	}

	// Reload with merchant and user preloaded
	giftCard, err = h.giftCardService.GetGiftCard(c.Request().Context(), giftCardID)
	if err != nil {
		return c.String(http.StatusNotFound, i18n.T(c.Request().Context(), "error.gift_card_not_found"))
	}

	isFavorite, err := h.favoriteService.IsFavorite(c.Request().Context(), user.ID, "gift_card", giftCardID)
	if err != nil {
		c.Logger().Errorf("Failed to check favorite status: %v", err)
		isFavorite = false
	}

	canEdit := perms.CanEdit
	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = ""
	}
	component := templates.GiftCardDetailView(c.Request().Context(), csrfToken, *giftCard, user, canEdit, isFavorite)
	return component.Render(c.Request().Context(), c.Response().Writer)
}
