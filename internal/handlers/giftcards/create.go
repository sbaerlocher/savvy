// Package giftcards provides HTTP handlers for gift card management operations.
package giftcards

import (
	"net/http"
	"savvy/internal/database"
	"savvy/internal/models"
	"savvy/internal/validation"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Create creates a new gift card
func (h *Handler) Create(c echo.Context) error {
	user := c.Get("current_user").(*models.User)

	// Parse balance
	initialBalance, _ := strconv.ParseFloat(c.FormValue("initial_balance"), 64)
	cardNumber := c.FormValue("card_number")

	// Parse expiration date
	var expiresAt *time.Time
	if expiresAtStr := c.FormValue("expires_at"); expiresAtStr != "" {
		parsed, err := validation.ParseAndValidateDate(expiresAtStr, true) // allow past for flexibility
		if err != nil {
			c.Logger().Errorf("Expires_at validation failed: %v", err)
			return c.Redirect(http.StatusSeeOther, "/gift-cards/new?error=invalid_date")
		}
		// Set to end of day
		parsed = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 0, time.UTC)
		expiresAt = &parsed
	}

	merchantIDStr := c.FormValue("merchant_id")
	merchantNameStr := c.FormValue("merchant_name")

	giftCard := models.GiftCard{
		UserID:         &user.ID,
		CardNumber:     cardNumber,
		InitialBalance: initialBalance,
		Currency:       c.FormValue("currency"),
		PIN:            c.FormValue("pin"),
		ExpiresAt:      expiresAt,
		Status:         "active",
		BarcodeType:    c.FormValue("barcode_type"),
		Notes:          c.FormValue("notes"),
	}

	// Handle merchant selection
	if merchantIDStr != "" && merchantIDStr != newMerchantValue {
		// Existing merchant selected from dropdown
		merchantID, err := uuid.Parse(merchantIDStr)
		if err == nil {
			giftCard.MerchantID = &merchantID
			// Load merchant to get name
			merchant, err := h.merchantService.GetMerchantByID(c.Request().Context(), merchantID)
			if err == nil {
				giftCard.MerchantName = merchant.Name
			}
		}
	} else {
		// New merchant name entered
		giftCard.MerchantName = merchantNameStr
	}

	// Default barcode type if not provided
	if giftCard.BarcodeType == "" {
		giftCard.BarcodeType = "CODE128"
	}

	// Default currency if not provided
	if giftCard.Currency == "" {
		giftCard.Currency = "CHF"
	}

	if err := h.giftCardService.CreateGiftCard(c.Request().Context(), &giftCard); err != nil {
		// Check if it's a duplicate key error (race condition caught by DB)
		if database.IsDuplicateError(err) {
			c.Logger().Warnf("Duplicate gift card number detected by database constraint: %s", cardNumber)
			return c.Redirect(http.StatusSeeOther, "/gift-cards/new?error=card_number_exists")
		}
		c.Logger().Errorf("Failed to create gift card: %v", err)
		return c.Redirect(http.StatusSeeOther, "/gift-cards/new?error=database_error")
	}

	// Handle sharing if email provided
	// Note: Share creation still uses database.DB directly as ShareService.ShareGiftCard is not fully implemented
	shareEmail := c.FormValue("share_with_email")
	if shareEmail != "" {
		// Find user by email (normalize email)
		shareEmail = strings.ToLower(strings.TrimSpace(shareEmail))

		var sharedUser models.User
		if err := database.DB.Where("LOWER(email) = ?", shareEmail).First(&sharedUser).Error; err == nil {
			// User exists, create share
			canEdit := c.FormValue("share_can_edit") == trueStringValue
			canDelete := c.FormValue("share_can_delete") == trueStringValue
			canEditTransactions := c.FormValue("share_can_edit_transactions") == trueStringValue

			share := models.GiftCardShare{
				GiftCardID:          giftCard.ID,
				SharedWithID:        sharedUser.ID,
				CanEdit:             canEdit,
				CanDelete:           canDelete,
				CanEditTransactions: canEditTransactions,
			}

			if err := database.DB.Create(&share).Error; err != nil {
				c.Logger().Warnf("Failed to create gift card share: %v", err)
			} else {
				c.Logger().Printf("Gift card shared with %s", shareEmail)
			}
		} else {
			c.Logger().Warnf("Share email not found: %s", shareEmail)
		}
	}

	return c.Redirect(http.StatusSeeOther, "/gift-cards/"+giftCard.ID.String())
}
