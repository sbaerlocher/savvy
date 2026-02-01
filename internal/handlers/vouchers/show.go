// Package vouchers provides HTTP handlers for voucher management operations.
package vouchers

import (
	"net/http"
	"savvy/internal/models"
	"savvy/internal/templates"
	"savvy/internal/views"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Show displays a single voucher
func (h *Handler) Show(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	isImpersonating := c.Get("is_impersonating") != nil

	voucherID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/vouchers")
	}

	// Check authorization
	perms, err := h.authzService.CheckVoucherAccess(c.Request().Context(), user.ID, voucherID)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/vouchers")
	}


	voucher, err := h.voucherService.GetVoucher(c.Request().Context(), voucherID)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/vouchers")
	}


	var shares []models.VoucherShare
	if perms.IsOwner {
		shares, _ = h.shareService.GetVoucherShares(c.Request().Context(), voucherID)
	}


	merchants, err := h.merchantService.GetAllMerchants(c.Request().Context())
	if err != nil {
		merchants = []models.Merchant{} // Fallback to empty list
	}


	isFavorite, _ := h.favoriteService.IsFavorite(c.Request().Context(), user.ID, "voucher", voucherID)

	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = ""
	}

	view := views.VoucherShowView{
		Voucher:   *voucher,
		Merchants: merchants,
		Shares:    shares,
		User:      user,
		Permissions: views.VoucherPermissions{
			CanEdit:    perms.CanEdit,
			CanDelete:  perms.CanDelete,
			IsFavorite: isFavorite,
		},
		IsImpersonating: isImpersonating,
	}

	return templates.VouchersShow(c.Request().Context(), csrfToken, view).Render(c.Request().Context(), c.Response().Writer)
}
