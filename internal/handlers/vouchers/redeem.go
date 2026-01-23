package vouchers

import (
	"net/http"
	"savvy/internal/database"
	"savvy/internal/models"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Redeem handles voucher redemption (increments used_count)
func (h *Handler) Redeem(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	voucherID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/vouchers")
	}

	// Check authorization (owner or shared - view access is sufficient for redeem)
	_, err = h.authzService.CheckVoucherAccess(c.Request().Context(), user.ID, voucherID)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/vouchers")
	}

	var voucher models.Voucher
	if err := database.DB.First(&voucher, voucherID).Error; err != nil {
		return c.Redirect(http.StatusSeeOther, "/vouchers")
	}

	// Try to redeem
	if err := voucher.Redeem(); err != nil {
		// Voucher cannot be redeemed (expired or usage limit reached)
		return c.Redirect(http.StatusSeeOther, "/vouchers/"+voucher.ID.String()+"?error=cannot_redeem")
	}

	// Save updated voucher
	if err := database.DB.Save(&voucher).Error; err != nil {
		c.Logger().Errorf("Failed to redeem voucher: %v", err)
		return c.Redirect(http.StatusSeeOther, "/vouchers/"+voucher.ID.String()+"?error=database_error")
	}

	return c.Redirect(http.StatusSeeOther, "/vouchers/"+voucher.ID.String()+"?success=redeemed")
}
