// Package vouchers provides HTTP handlers for voucher management operations.
package vouchers

import (
	"net/http"
	"savvy/internal/database"
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

	var voucher models.Voucher
	// Preload merchant for color and user for ownership display
	if err := database.DB.Where("id = ?", voucherID).Preload("Merchant").Preload("User").First(&voucher).Error; err != nil {
		return c.Redirect(http.StatusSeeOther, "/vouchers")
	}

	// Load shares with users (only if owner)
	var shares []models.VoucherShare
	if perms.IsOwner {
		database.DB.Where("voucher_id = ?", voucherID).Preload("SharedWithUser").Find(&shares)
	}

	// Load all merchants for dropdown
	var merchants []models.Merchant
	database.DB.Order("name ASC").Find(&merchants)

	// Check if voucher is favorited by current user
	var favorite models.UserFavorite
	isFavorite := database.DB.Where("user_id = ? AND resource_type = ? AND resource_id = ?",
		user.ID, "voucher", voucherID).First(&favorite).Error == nil

	csrfToken := c.Get("csrf").(string)

	view := views.VoucherShowView{
		Voucher:     voucher,
		Merchants:   merchants,
		Shares:      shares,
		User:        user,
		Permissions: views.VoucherPermissions{
			CanEdit:    perms.CanEdit,
			CanDelete:  perms.CanDelete,
			IsFavorite: isFavorite,
		},
		IsImpersonating: isImpersonating,
	}

	return templates.VouchersShow(c.Request().Context(), csrfToken, view).Render(c.Request().Context(), c.Response().Writer)
}
