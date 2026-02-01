// Package merchants contains HTTP request handlers for merchant operations.
package merchants

import (
	"net/http"
	"savvy/internal/models"
	"savvy/internal/templates"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Edit shows the form to edit a merchant
func (h *Handler) Edit(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	isImpersonating := c.Get("is_impersonating") != nil

	merchantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/merchants")
	}

	merchant, err := h.merchantService.GetMerchantByID(c.Request().Context(), merchantID)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/merchants")
	}

	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = ""
	}
	return templates.MerchantsEdit(c.Request().Context(), csrfToken, *merchant, user, isImpersonating).Render(c.Request().Context(), c.Response().Writer)
}
