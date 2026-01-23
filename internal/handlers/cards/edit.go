// Package cards contains HTTP request handlers for card operations.
package cards

import (
	"net/http"
	"savvy/internal/database"
	"savvy/internal/models"
	"savvy/internal/templates"
	"savvy/internal/views"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
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

	var card models.Card
	if err := database.DB.Where("id = ?", cardID).First(&card).Error; err != nil {
		return c.Redirect(http.StatusSeeOther, "/cards")
	}

	// Load all merchants for dropdown
	var merchants []models.Merchant
	database.DB.Order("name ASC").Find(&merchants)

	csrfToken := c.Get("csrf").(string)

	view := views.CardEditView{
		Card:            card,
		Merchants:       merchants,
		User:            user,
		IsImpersonating: isImpersonating,
	}

	return templates.CardsEdit(c.Request().Context(), csrfToken, view).Render(c.Request().Context(), c.Response().Writer)
}
