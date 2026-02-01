// Package giftcards provides HTTP handlers for gift card management operations.
package giftcards

import (
	"net/http"
	"savvy/internal/audit"
	"savvy/internal/models"
	"savvy/internal/validation"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Update updates a gift card
func (h *Handler) Update(c echo.Context) error {
	user := c.Get("current_user").(*models.User)

	giftCardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/gift-cards")
	}

	// Check authorization
	perms, err := h.authzService.CheckGiftCardAccess(c.Request().Context(), user.ID, giftCardID)
	if err != nil || !perms.CanEdit {
		return c.Redirect(http.StatusSeeOther, "/gift-cards")
	}

	giftCard, err := h.giftCardService.GetGiftCard(c.Request().Context(), giftCardID)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/gift-cards")
	}

	initialBalance, _ := strconv.ParseFloat(c.FormValue("initial_balance"), 64)

	var expiresAt *time.Time
	if expiresAtStr := c.FormValue("expires_at"); expiresAtStr != "" {
		parsed, err := validation.ParseAndValidateDate(expiresAtStr, true)
		if err != nil {
			c.Logger().Errorf("Expires_at validation failed: %v", err)
			return c.Redirect(http.StatusSeeOther, "/gift-cards/"+giftCard.ID.String()+"/edit?error=invalid_date")
		}
		parsed = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 0, time.UTC)
		expiresAt = &parsed
	}

	merchantIDStr := c.FormValue("merchant_id")
	merchantNameStr := c.FormValue("merchant_name")

	giftCard.CardNumber = c.FormValue("card_number")
	giftCard.InitialBalance = initialBalance
	giftCard.Currency = c.FormValue("currency")
	giftCard.PIN = c.FormValue("pin")
	giftCard.ExpiresAt = expiresAt
	giftCard.Status = c.FormValue("status")
	giftCard.BarcodeType = c.FormValue("barcode_type")
	giftCard.Notes = c.FormValue("notes")

	if merchantIDStr != "" && merchantIDStr != "new" {
		merchantID, err := uuid.Parse(merchantIDStr)
		if err == nil {
			giftCard.MerchantID = &merchantID
			merchant, err := h.merchantService.GetMerchantByID(c.Request().Context(), merchantID)
			if err == nil {
				giftCard.MerchantName = merchant.Name
			}
		}
	} else {
		giftCard.MerchantID = nil
		giftCard.MerchantName = merchantNameStr
	}

	if err := h.giftCardService.UpdateGiftCard(c.Request().Context(), giftCard); err != nil {
		return c.Redirect(http.StatusSeeOther, "/gift-cards/"+giftCard.ID.String()+"/edit")
	}

	if h.db != nil {
		if err := audit.LogUpdateFromContext(c, h.db, "gift_cards", giftCard.ID, *giftCard); err != nil {
			c.Logger().Errorf("Failed to log gift card update: %v", err)
		}
	}

	return c.Redirect(http.StatusSeeOther, "/gift-cards/"+giftCard.ID.String())
}
