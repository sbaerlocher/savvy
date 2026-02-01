// Package merchants contains HTTP request handlers for merchant operations.
package merchants

import (
	"savvy/internal/models"
	"savvy/internal/templates"

	"github.com/labstack/echo/v4"
)

// New shows the form to create a new merchant
func (h *Handler) New(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	isImpersonating := c.Get("is_impersonating") != nil
	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = ""
	}

	return templates.MerchantsNew(c.Request().Context(), csrfToken, user, isImpersonating).Render(c.Request().Context(), c.Response().Writer)
}
