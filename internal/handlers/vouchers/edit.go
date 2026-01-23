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

// Edit shows the form to edit a voucher
func (h *Handler) Edit(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	isImpersonating := c.Get("is_impersonating") != nil

	voucherID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/vouchers")
	}

	// Check authorization
	perms, err := h.authzService.CheckVoucherAccess(c.Request().Context(), user.ID, voucherID)
	if err != nil || !perms.CanEdit {
		return c.Redirect(http.StatusSeeOther, "/vouchers")
	}

	var voucher models.Voucher
	if err := database.DB.Where("id = ?", voucherID).First(&voucher).Error; err != nil {
		return c.Redirect(http.StatusSeeOther, "/vouchers")
	}

	// Load all merchants for dropdown
	var merchants []models.Merchant
	database.DB.Order("name ASC").Find(&merchants)

	csrfToken := c.Get("csrf").(string)

	view := views.VoucherEditView{
		Voucher:         voucher,
		Merchants:       merchants,
		User:            user,
		IsImpersonating: isImpersonating,
	}

	return templates.VouchersEdit(c.Request().Context(), csrfToken, view).Render(c.Request().Context(), c.Response().Writer)
}
