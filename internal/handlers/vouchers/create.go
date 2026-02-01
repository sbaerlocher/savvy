// Package vouchers provides HTTP handlers for voucher management operations.
package vouchers

import (
	"net/http"
	"savvy/internal/database"
	"savvy/internal/models"
	"savvy/internal/validation"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Create creates a new voucher
func (h *Handler) Create(c echo.Context) error {
	user := c.Get("current_user").(*models.User)

	value, _ := strconv.ParseFloat(c.FormValue("value"), 64)
	minPurchaseAmount, _ := strconv.ParseFloat(c.FormValue("min_purchase_amount"), 64)
	usageLimitType := c.FormValue("usage_limit_type")
	code := c.FormValue("code")

	validFrom, validUntil, err := validation.ParseAndValidateDateRange(
		c.FormValue("valid_from"),
		c.FormValue("valid_until"),
		false,
	)
	if err != nil {
		c.Logger().Errorf("Date validation failed: %v", err)
		return c.Redirect(http.StatusSeeOther, "/vouchers/new?error=invalid_date")
	}

	merchantIDStr := c.FormValue("merchant_id")
	merchantNameStr := c.FormValue("merchant_name")

	voucher := models.Voucher{
		UserID:            &user.ID,
		Code:              code,
		Type:              c.FormValue("type"),
		Value:             value,
		Description:       c.FormValue("description"),
		MinPurchaseAmount: minPurchaseAmount,
		ValidFrom:         validFrom,
		ValidUntil:        validUntil,
		UsageLimitType:    usageLimitType,
		BarcodeType:       c.FormValue("barcode_type"),
	}

	if merchantIDStr != "" && merchantIDStr != "new" {
		merchantID, err := uuid.Parse(merchantIDStr)
		if err == nil {
			voucher.MerchantID = &merchantID
			merchant, err := h.merchantService.GetMerchantByID(c.Request().Context(), merchantID)
			if err == nil {
				voucher.MerchantName = merchant.Name
			}
		}
	} else {
		voucher.MerchantName = merchantNameStr
	}

	if voucher.BarcodeType == "" {
		voucher.BarcodeType = "QR"
	}

	if err := h.voucherService.CreateVoucher(c.Request().Context(), &voucher); err != nil {
		if database.IsDuplicateError(err) {
			c.Logger().Warnf("Duplicate voucher code detected by database constraint: %s", code)
			return c.Redirect(http.StatusSeeOther, "/vouchers/new?error=code_exists")
		}
		c.Logger().Errorf("Failed to create voucher: %v", err)
		return c.Redirect(http.StatusSeeOther, "/vouchers/new?error=database_error")
	}

	shareEmail := c.FormValue("share_with_email")
	if shareEmail != "" {
		sharedUser, err := h.userService.GetUserByEmail(c.Request().Context(), shareEmail)
		if err == nil {
			if err := h.shareService.CreateVoucherShare(c.Request().Context(), voucher.ID, sharedUser.ID); err != nil {
				c.Logger().Warnf("Failed to create voucher share: %v", err)
			} else {
				c.Logger().Printf("Voucher shared with %s", shareEmail)
			}
		} else {
			c.Logger().Warnf("Share email not found: %s", shareEmail)
		}
	}

	return c.Redirect(http.StatusSeeOther, "/vouchers/"+voucher.ID.String())
}
