// Package vouchers provides HTTP handlers for voucher management operations.
package vouchers

import (
	"savvy/internal/database"
	"savvy/internal/models"
	"savvy/internal/templates"
	"savvy/internal/views"

	"github.com/labstack/echo/v4"
)

// New shows the form to create a new voucher
func (h *Handler) New(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	isImpersonating := c.Get("is_impersonating") != nil
	csrfToken := c.Get("csrf").(string)

	// Load all merchants for dropdown
	var merchants []models.Merchant
	database.DB.Order("name ASC").Find(&merchants)

	view := views.VoucherEditView{
		Voucher:         models.Voucher{},
		Merchants:       merchants,
		User:            user,
		IsImpersonating: isImpersonating,
	}

	return templates.VouchersNew(c.Request().Context(), csrfToken, view).Render(c.Request().Context(), c.Response().Writer)
}
