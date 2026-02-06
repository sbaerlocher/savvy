// Package vouchers provides HTTP handlers for voucher management operations.
package vouchers

import (
	"net/http"
	"savvy/internal/audit"
	"savvy/internal/models"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

// Delete deletes a voucher
func (h *Handler) Delete(c echo.Context) error {
	user := c.Get("current_user").(*models.User)

	voucherID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/vouchers")
	}

	// Check authorization
	perms, err := h.authzService.CheckVoucherAccess(c.Request().Context(), user.ID, voucherID)
	if err != nil || !perms.CanDelete {
		return c.Redirect(http.StatusSeeOther, "/vouchers")
	}

	// Add user context for audit logging
	ctx := audit.AddUserIDToContext(c.Request().Context(), user.ID)
	if err := h.voucherService.DeleteVoucher(ctx, voucherID); err != nil {
		return c.Redirect(http.StatusSeeOther, "/vouchers")
	}

	// Always use HX-Redirect header for consistent behavior
	c.Response().Header().Set("HX-Redirect", "/vouchers")
	return c.Redirect(http.StatusSeeOther, "/vouchers")
}
