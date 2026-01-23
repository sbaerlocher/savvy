package gift_cards

import (
	"net/http"
	"savvy/internal/database"
	"savvy/internal/models"
	"savvy/internal/templates"
	"savvy/internal/validation"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// EditInline returns the inline edit form
func (h *Handler) EditInline(c echo.Context) error {
	user := c.Get("current_user").(*models.User)

	giftCardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.String(http.StatusNotFound, "Gift card not found")
	}

	// Check authorization
	perms, err := h.authzService.CheckGiftCardAccess(c.Request().Context(), user.ID, giftCardID)
	if err != nil || !perms.CanEdit {
		return c.String(http.StatusForbidden, "Not authorized")
	}

	var giftCard models.GiftCard
	if err := database.DB.Where("id = ?", giftCardID).Preload("Merchant").First(&giftCard).Error; err != nil {
		return c.String(http.StatusNotFound, "Gift card not found")
	}

	// Load all merchants for dropdown
	var merchants []models.Merchant
	database.DB.Order("name ASC").Find(&merchants)

	csrfToken := c.Get("csrf").(string)
	component := templates.GiftCardDetailEdit(c.Request().Context(), csrfToken, giftCard, merchants)
	return component.Render(c.Request().Context(), c.Response().Writer)
}

// CancelEdit cancels inline editing and returns to display
func (h *Handler) CancelEdit(c echo.Context) error {
	user := c.Get("current_user").(*models.User)

	giftCardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.String(http.StatusNotFound, "Gift card not found")
	}

	// Check authorization
	perms, err := h.authzService.CheckGiftCardAccess(c.Request().Context(), user.ID, giftCardID)
	if err != nil {
		return c.String(http.StatusNotFound, "Gift card not found")
	}

	var giftCard models.GiftCard
	if err := database.DB.Where("id = ?", giftCardID).Preload("Merchant").Preload("User").First(&giftCard).Error; err != nil {
		return c.String(http.StatusNotFound, "Gift card not found")
	}

	canEdit := perms.CanEdit

	// Check if gift card is favorited by current user
	var favorite models.UserFavorite
	isFavorite := database.DB.Where("user_id = ? AND resource_type = ? AND resource_id = ?",
		user.ID, "gift_card", giftCardID).First(&favorite).Error == nil

	csrfToken := c.Get("csrf").(string)
	component := templates.GiftCardDetailView(c.Request().Context(), csrfToken, giftCard, user, canEdit, isFavorite)
	return component.Render(c.Request().Context(), c.Response().Writer)
}

// UpdateInline updates a gift card and returns the view
func (h *Handler) UpdateInline(c echo.Context) error {
	user := c.Get("current_user").(*models.User)

	giftCardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.String(http.StatusNotFound, "Gift card not found")
	}

	// Check authorization
	perms, err := h.authzService.CheckGiftCardAccess(c.Request().Context(), user.ID, giftCardID)
	if err != nil || !perms.CanEdit {
		return c.String(http.StatusForbidden, "Not authorized")
	}

	var giftCard models.GiftCard
	if err := database.DB.Where("id = ?", giftCardID).First(&giftCard).Error; err != nil {
		return c.String(http.StatusNotFound, "Gift card not found")
	}

	// Parse balance
	initialBalance, _ := strconv.ParseFloat(c.FormValue("initial_balance"), 64)

	// Parse expiration date
	var expiresAt *time.Time
	if expiresAtStr := c.FormValue("expires_at"); expiresAtStr != "" {
		parsed, err := validation.ParseAndValidateDate(expiresAtStr, true) // allow past for flexibility
		if err != nil {
			c.Logger().Errorf("Expires_at validation failed: %v", err)
			return c.String(http.StatusBadRequest, "Invalid expiration date")
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
	giftCard.Status = c.FormValue("status")
	giftCard.BarcodeType = c.FormValue("barcode_type")
	giftCard.Notes = c.FormValue("notes")

	// Handle merchant selection
	if merchantIDStr != "" && merchantIDStr != "new" {
		// Existing merchant selected from dropdown
		merchantID, err := uuid.Parse(merchantIDStr)
		if err == nil {
			giftCard.MerchantID = &merchantID
			// Load merchant to get name
			var merchant models.Merchant
			if err := database.DB.Where("id = ?", merchantID).First(&merchant).Error; err == nil {
				giftCard.MerchantName = merchant.Name
			}
		}
	} else {
		// New merchant name entered or no selection
		giftCard.MerchantID = nil
		giftCard.MerchantName = merchantNameStr
	}

	if err := database.DB.Save(&giftCard).Error; err != nil {
		return c.String(http.StatusInternalServerError, "Failed to update gift card")
	}

	// Reload with merchant and user preloaded
	database.DB.Where("id = ?", giftCardID).Preload("Merchant").Preload("User").First(&giftCard)

	// Check if gift card is favorited by current user
	var favorite models.UserFavorite
	isFavorite := database.DB.Where("user_id = ? AND resource_type = ? AND resource_id = ?",
		user.ID, "gift_card", giftCardID).First(&favorite).Error == nil

	canEdit := perms.CanEdit
	csrfToken := c.Get("csrf").(string)
	component := templates.GiftCardDetailView(c.Request().Context(), csrfToken, giftCard, user, canEdit, isFavorite)
	return component.Render(c.Request().Context(), c.Response().Writer)
}
