// Package vouchers provides HTTP handlers for voucher management operations.
package vouchers

import (
	"net/http"
	"savvy/internal/audit"
	"savvy/internal/database"
	"savvy/internal/models"
	"savvy/internal/validation"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Update updates a voucher
func (h *Handler) Update(c echo.Context) error {
	user := c.Get("current_user").(*models.User)

	voucherID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/vouchers")
	}

	// Check authorization
	perms, err := h.authzService.CheckVoucherAccess(c.Request().Context(), user.ID, voucherID)
	if err != nil || !perms.CanEdit {
		return c.Redirect(http.StatusSeeOther, "/vouchers")
	}

	voucher, err := h.voucherService.GetVoucher(c.Request().Context(), voucherID)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/vouchers")
	}

	// Parse value
	value, _ := strconv.ParseFloat(c.FormValue("value"), 64)
	minPurchaseAmount, _ := strconv.ParseFloat(c.FormValue("min_purchase_amount"), 64)

	// Parse usage limit type
	usageLimitType := c.FormValue("usage_limit_type")

	// Parse dates (date only, set time to start/end of day)
	validFrom, validUntil, err := validation.ParseAndValidateDateRange(
		c.FormValue("valid_from"),
		c.FormValue("valid_until"),
		true, // allow past dates for existing vouchers (editing)
	)
	if err != nil {
		c.Logger().Errorf("Date validation failed: %v", err)
		return c.Redirect(http.StatusSeeOther, "/vouchers/"+voucher.ID.String()+"/edit?error=invalid_date")
	}

	// Handle merchant selection
	merchantIDStr := c.FormValue("merchant_id")
	if merchantIDStr != "" && merchantIDStr != newMerchantValue {
		// Existing merchant selected from dropdown
		merchantID, err := uuid.Parse(merchantIDStr)
		if err == nil {
			voucher.MerchantID = &merchantID
			// Load merchant to get name
			merchant, err := h.merchantService.GetMerchantByID(c.Request().Context(), merchantID)
			if err == nil {
				voucher.MerchantName = merchant.Name
			}
		}
	} else {
		// New merchant name entered or no selection
		voucher.MerchantID = nil
		voucher.MerchantName = c.FormValue("merchant_name")
	}

	// Update fields
	voucher.Code = c.FormValue("code")
	voucher.Type = c.FormValue("type")
	voucher.Value = value
	voucher.Description = c.FormValue("description")
	voucher.MinPurchaseAmount = minPurchaseAmount
	voucher.ValidFrom = validFrom
	voucher.ValidUntil = validUntil
	voucher.UsageLimitType = usageLimitType
	voucher.BarcodeType = c.FormValue("barcode_type")

	if err := h.voucherService.UpdateVoucher(c.Request().Context(), voucher); err != nil {
		return c.Redirect(http.StatusSeeOther, "/vouchers/"+voucher.ID.String()+"/edit")
	}

	// Log update to audit log (only if DB is available)
	if database.DB != nil {
		if err := audit.LogUpdateFromContext(c, database.DB, "vouchers", voucher.ID, *voucher); err != nil {
			c.Logger().Errorf("Failed to log voucher update: %v", err)
		}
	}

	return c.Redirect(http.StatusSeeOther, "/vouchers/"+voucher.ID.String())
}
