// Package cards contains HTTP request handlers for card operations.
package cards

import (
	"net/http"
	"savvy/internal/database"
	"savvy/internal/models"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Create creates a new card
func (h *Handler) Create(c echo.Context) error {
	user := c.Get("current_user").(*models.User)

	merchantIDStr := c.FormValue("merchant_id")
	merchantNameStr := c.FormValue("merchant_name")
	cardNumber := c.FormValue("card_number")

	// Debug logging
	c.Logger().Printf("CardsCreate - merchant_id: '%s', merchant_name: '%s'", merchantIDStr, merchantNameStr)

	card := models.Card{
		UserID:      &user.ID,
		Program:     c.FormValue("program"),
		CardNumber:  cardNumber,
		BarcodeType: c.FormValue("barcode_type"),
		Status:      "active",
		Notes:       c.FormValue("notes"),
	}

	// Handle merchant selection
	if merchantIDStr != "" && merchantIDStr != "new" {
		// Existing merchant selected from dropdown
		merchantID, err := uuid.Parse(merchantIDStr)
		if err == nil {
			card.MerchantID = &merchantID
			// Load merchant to get name
			var merchant models.Merchant
			if err := database.DB.Where("id = ?", merchantID).First(&merchant).Error; err == nil {
				card.MerchantName = merchant.Name
				c.Logger().Printf("Loaded merchant: %s", merchant.Name)
			} else {
				c.Logger().Printf("Failed to load merchant: %v", err)
			}
		}
	} else {
		// New merchant name entered
		card.MerchantName = merchantNameStr
		c.Logger().Printf("Using new merchant name: %s", merchantNameStr)
	}

	// Default barcode type if not provided
	if card.BarcodeType == "" {
		card.BarcodeType = "CODE128"
	}

	c.Logger().Printf("Creating card: MerchantID=%v, MerchantName=%s", card.MerchantID, card.MerchantName)

	if err := database.DB.Create(&card).Error; err != nil {
		// Check if it's a duplicate key error (race condition caught by DB)
		if database.IsDuplicateError(err) {
			c.Logger().Warnf("Duplicate card_number detected by database constraint: %s", cardNumber)
			return c.Redirect(http.StatusSeeOther, "/cards/new?error=card_number_exists")
		}
		c.Logger().Errorf("Failed to create card: %v", err)
		return c.Redirect(http.StatusSeeOther, "/cards/new?error=database_error")
	}

	// Handle sharing if email provided
	shareEmail := c.FormValue("share_with_email")
	if shareEmail != "" {
		// Find user by email (normalize email)
		shareEmail = strings.ToLower(strings.TrimSpace(shareEmail))

		var sharedUser models.User
		if err := database.DB.Where("LOWER(email) = ?", shareEmail).First(&sharedUser).Error; err == nil {
			// User exists, create share
			canEdit := c.FormValue("share_can_edit") == "true"
			canDelete := c.FormValue("share_can_delete") == "true"

			share := models.CardShare{
				CardID:       card.ID,
				SharedWithID: sharedUser.ID,
				CanEdit:      canEdit,
				CanDelete:    canDelete,
			}

			if err := database.DB.Create(&share).Error; err != nil {
				c.Logger().Warnf("Failed to create card share: %v", err)
				// Don't fail the whole request, just log the error
			} else {
				c.Logger().Printf("Card shared with %s", shareEmail)
			}
		} else {
			c.Logger().Warnf("Share email not found: %s", shareEmail)
			// Could add flash message here
		}
	}

	return c.Redirect(http.StatusSeeOther, "/cards/"+card.ID.String())
}
