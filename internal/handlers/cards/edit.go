// Package cards contains HTTP request handlers for card operations.
package cards

import (
	"net/http"
	"savvy/internal/models"
	"savvy/internal/templates"
	"savvy/internal/views"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

// Edit shows the form to edit a card
func (h *Handler) Edit(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	isImpersonating := c.Get("is_impersonating") != nil

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

	merchants, err := h.merchantService.GetAllMerchants(c.Request().Context())
	if err != nil {
		merchants = []models.Merchant{}
	}

	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = ""
	}

	view := views.CardEditView{
		Card:            *card,
		Merchants:       merchants,
		User:            user,
		IsImpersonating: isImpersonating,
	}

	return templates.CardsEdit(c.Request().Context(), csrfToken, view).Render(c.Request().Context(), c.Response().Writer)
}
