// Package merchants contains HTTP request handlers for merchant operations.
package merchants

import (
	"savvy/internal/models"
	"savvy/internal/templates"

	"github.com/labstack/echo/v5"
)

// Index lists all merchants
func (h *Handler) Index(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	isImpersonating := c.Get("is_impersonating") != nil

	merchants, err := h.merchantService.GetAllMerchants(c.Request().Context())
	if err != nil {
		return err
	}

	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = ""
	}
	return templates.MerchantsIndex(c.Request().Context(), csrfToken, merchants, user, isImpersonating).Render(c.Request().Context(), c.Response().Writer)
}
