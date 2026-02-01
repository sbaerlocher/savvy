// Package cards contains HTTP request handlers for card operations.
package cards

import (
	"net/http"
	"savvy/internal/database"
	"savvy/internal/models"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Create creates a new card
func (h *Handler) Create(c echo.Context) error {
	user := c.Get("current_user").(*models.User)

	merchantIDStr := c.FormValue("merchant_id")
	merchantNameStr := c.FormValue("merchant_name")
	cardNumber := c.FormValue("card_number")

	card := models.Card{
		UserID:      &user.ID,
		Program:     c.FormValue("program"),
		CardNumber:  cardNumber,
		BarcodeType: c.FormValue("barcode_type"),
		Status:      "active",
		Notes:       c.FormValue("notes"),
	}

	if merchantIDStr != "" && merchantIDStr != "new" {
		merchantID, err := uuid.Parse(merchantIDStr)
		if err == nil {
			card.MerchantID = &merchantID
			merchant, err := h.merchantService.GetMerchantByID(c.Request().Context(), merchantID)
			if err == nil {
				card.MerchantName = merchant.Name
				c.Logger().Printf("Loaded merchant: %s", merchant.Name)
			} else {
				c.Logger().Printf("Failed to load merchant: %v", err)
			}
		}
	} else {
		card.MerchantName = merchantNameStr
		c.Logger().Printf("Using new merchant name: %s", merchantNameStr)
	}

	if card.BarcodeType == "" {
		card.BarcodeType = "CODE128"
	}

	c.Logger().Printf("Creating card: MerchantID=%v, MerchantName=%s", card.MerchantID, card.MerchantName)

	if err := h.cardService.CreateCard(c.Request().Context(), &card); err != nil {
		if database.IsDuplicateError(err) {
			c.Logger().Warnf("Duplicate card_number detected by database constraint: %s", cardNumber)
			return c.Redirect(http.StatusSeeOther, "/cards/new?error=card_number_exists")
		}
		c.Logger().Errorf("Failed to create card: %v", err)
		return c.Redirect(http.StatusSeeOther, "/cards/new?error=database_error")
	}

	shareEmail := c.FormValue("share_with_email")
	if shareEmail != "" {
		sharedUser, err := h.userService.GetUserByEmail(c.Request().Context(), shareEmail)
		if err == nil {
			canEdit := c.FormValue("share_can_edit") == "true"
			canDelete := c.FormValue("share_can_delete") == "true"

			if err := h.shareService.CreateCardShare(c.Request().Context(), card.ID, sharedUser.ID, canEdit, canDelete); err != nil {
				c.Logger().Warnf("Failed to create card share: %v", err)
			} else {
				c.Logger().Printf("Card shared with %s", shareEmail)
			}
		} else {
			c.Logger().Warnf("Share email not found: %s", shareEmail)
		}
	}

	return c.Redirect(http.StatusSeeOther, "/cards/"+card.ID.String())
}
