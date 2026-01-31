// Package vouchers provides HTTP handlers for voucher management operations.
package vouchers

import (
	"savvy/internal/models"
	"savvy/internal/templates"
	"savvy/internal/views"

	"github.com/labstack/echo/v4"
)

// Index lists all vouchers for the current user (owned + shared)
func (h *Handler) Index(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	isImpersonating := c.Get("is_impersonating") != nil

	// Get all vouchers (owned + shared) via service
	allVouchers, err := h.voucherService.GetUserVouchers(c.Request().Context(), user.ID)
	if err != nil {
		return err
	}

	view := views.VoucherIndexView{
		Vouchers:        allVouchers,
		User:            user,
		IsImpersonating: isImpersonating,
	}

	return templates.VouchersIndex(c.Request().Context(), view).Render(c.Request().Context(), c.Response().Writer)
}
