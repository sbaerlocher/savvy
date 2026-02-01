// Package giftcards provides HTTP handlers for gift card management operations.
package giftcards

import (
	"net/http"
	"savvy/internal/database"
	"savvy/internal/models"
	"savvy/internal/validation"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Create creates a new gift card
func (h *Handler) Create(c echo.Context) error {
	user := c.Get("current_user").(*models.User)

	initialBalance, _ := strconv.ParseFloat(c.FormValue("initial_balance"), 64)
	cardNumber := c.FormValue("card_number")

	var expiresAt *time.Time
	if expiresAtStr := c.FormValue("expires_at"); expiresAtStr != "" {
		parsed, err := validation.ParseAndValidateDate(expiresAtStr, true)
		if err != nil {
			c.Logger().Errorf("Expires_at validation failed: %v", err)
			return c.Redirect(http.StatusSeeOther, "/gift-cards/new?error=invalid_date")
		}
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

	if merchantIDStr != "" && merchantIDStr != newMerchantValue {
		merchantID, err := uuid.Parse(merchantIDStr)
		if err == nil {
			giftCard.MerchantID = &merchantID
			merchant, err := h.merchantService.GetMerchantByID(c.Request().Context(), merchantID)
			if err == nil {
				giftCard.MerchantName = merchant.Name
			}
		}
	} else {
		giftCard.MerchantName = merchantNameStr
	}

	if giftCard.BarcodeType == "" {
		giftCard.BarcodeType = "CODE128"
	}

	if giftCard.Currency == "" {
		giftCard.Currency = "CHF"
	}

	if err := h.giftCardService.CreateGiftCard(c.Request().Context(), &giftCard); err != nil {
		if database.IsDuplicateError(err) {
			c.Logger().Warnf("Duplicate gift card number detected by database constraint: %s", cardNumber)
			return c.Redirect(http.StatusSeeOther, "/gift-cards/new?error=card_number_exists")
		}
		c.Logger().Errorf("Failed to create gift card: %v", err)
		return c.Redirect(http.StatusSeeOther, "/gift-cards/new?error=database_error")
	}

	shareEmail := c.FormValue("share_with_email")
	if shareEmail != "" {
		sharedUser, err := h.userService.GetUserByEmail(c.Request().Context(), shareEmail)
		if err == nil {
			canEdit := c.FormValue("share_can_edit") == trueStringValue
			canDelete := c.FormValue("share_can_delete") == trueStringValue
			canEditTransactions := c.FormValue("share_can_edit_transactions") == trueStringValue

			if err := h.shareService.CreateGiftCardShare(c.Request().Context(), giftCard.ID, sharedUser.ID, canEdit, canDelete, canEditTransactions); err != nil {
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
