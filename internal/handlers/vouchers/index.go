package vouchers

import (
	"savvy/internal/database"
	"savvy/internal/models"
	"savvy/internal/templates"
	"savvy/internal/views"

	"github.com/labstack/echo/v4"
)

// Index lists all vouchers for the current user (owned + shared)
func (h *Handler) Index(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	isImpersonating := c.Get("is_impersonating") != nil

	// Get owned vouchers with merchant info
	var ownedVouchers []models.Voucher
	if err := database.DB.Where("user_id = ?", user.ID).Preload("User").Preload("Merchant").Order("valid_until DESC").Find(&ownedVouchers).Error; err != nil {
		return err
	}

	// Get shared voucher IDs with merchant info
	var shares []models.VoucherShare
	database.DB.Where("shared_with_id = ?", user.ID).Preload("Voucher").Preload("Voucher.User").Preload("Voucher.Merchant").Find(&shares)

	// Extract shared vouchers
	var sharedVouchers []models.Voucher
	for _, share := range shares {
		if share.Voucher != nil {
			sharedVouchers = append(sharedVouchers, *share.Voucher)
		}
	}

	// Combine owned and shared
	ownedVouchers = append(ownedVouchers, sharedVouchers...)
	allVouchers := ownedVouchers

	view := views.VoucherIndexView{
		Vouchers:        allVouchers,
		User:            user,
		IsImpersonating: isImpersonating,
	}

	return templates.VouchersIndex(c.Request().Context(), view).Render(c.Request().Context(), c.Response().Writer)
}
