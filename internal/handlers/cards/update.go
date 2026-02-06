// Package cards contains HTTP request handlers for card operations.
package cards

import (
	"net/http"
	"savvy/internal/audit"
	"savvy/internal/models"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

// Update updates a card
func (h *Handler) Update(c echo.Context) error {
	user := c.Get("current_user").(*models.User)

	cardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/cards")
	}

	// Check authorization
	perms, err := h.authzService.CheckCardAccess(c.Request().Context(), user.ID, cardID)
	if err != nil || !perms.CanEdit {
		return c.Redirect(http.StatusSeeOther, "/cards")
	}

	card, err := h.cardService.GetCard(c.Request().Context(), cardID)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/cards")
	}

	card.Program = c.FormValue("program")
	card.CardNumber = c.FormValue("card_number")
	card.BarcodeType = c.FormValue("barcode_type")
	card.Status = c.FormValue("status")
	card.Notes = c.FormValue("notes")

	merchantIDStr := c.FormValue("merchant_id")
	if merchantIDStr != "" && merchantIDStr != newMerchantValue {
		merchantID, err := uuid.Parse(merchantIDStr)
		if err == nil {
			card.MerchantID = &merchantID
			merchant, err := h.merchantService.GetMerchantByID(c.Request().Context(), merchantID)
			if err == nil {
				card.MerchantName = merchant.Name
			}
		}
	} else {
		card.MerchantID = nil
		card.MerchantName = c.FormValue("merchant_name")
	}

	if err := h.cardService.UpdateCard(c.Request().Context(), card); err != nil {
		return c.Redirect(http.StatusSeeOther, "/cards/"+card.ID.String()+"/edit")
	}

	if h.db != nil {
		if err := audit.LogUpdateFromContext(c, h.db, "cards", card.ID, *card); err != nil {
			c.Logger().Errorf("Failed to log card update: %v", err)
		}
	}

	return c.Redirect(http.StatusSeeOther, "/cards/"+card.ID.String())
}
