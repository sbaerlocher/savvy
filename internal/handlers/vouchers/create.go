// Package vouchers provides HTTP handlers for voucher management operations.
package vouchers

import (
	"net/http"
	"savvy/internal/database"
	"savvy/internal/models"
	"savvy/internal/validation"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Create creates a new voucher
func (h *Handler) Create(c echo.Context) error {
	user := c.Get("current_user").(*models.User)

	// Parse value
	value, _ := strconv.ParseFloat(c.FormValue("value"), 64)
	minPurchaseAmount, _ := strconv.ParseFloat(c.FormValue("min_purchase_amount"), 64)

	// Parse usage limit type
	usageLimitType := c.FormValue("usage_limit_type")
	code := c.FormValue("code")

	// Parse dates (date only, set time to start/end of day)
	validFrom, validUntil, err := validation.ParseAndValidateDateRange(
		c.FormValue("valid_from"),
		c.FormValue("valid_until"),
		false, // don't allow past dates for new vouchers
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
		UsedCount:         0,
		BarcodeType:       c.FormValue("barcode_type"),
	}

	// Handle merchant selection
	if merchantIDStr != "" && merchantIDStr != "new" {
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
		// New merchant name entered
		voucher.MerchantName = merchantNameStr
	}

	// Default barcode type if not provided
	if voucher.BarcodeType == "" {
		voucher.BarcodeType = "QR"
	}

	if err := h.voucherService.CreateVoucher(c.Request().Context(), &voucher); err != nil {
		// Check if it's a duplicate key error (race condition caught by DB)
		if database.IsDuplicateError(err) {
			c.Logger().Warnf("Duplicate voucher code detected by database constraint: %s", code)
			return c.Redirect(http.StatusSeeOther, "/vouchers/new?error=code_exists")
		}
		c.Logger().Errorf("Failed to create voucher: %v", err)
		return c.Redirect(http.StatusSeeOther, "/vouchers/new?error=database_error")
	}

	// Handle sharing if email provided (vouchers are always read-only when shared)
	// Note: Share creation still uses database.DB directly as ShareService.ShareVoucher is not fully implemented
	shareEmail := c.FormValue("share_with_email")
	if shareEmail != "" {
		// Find user by email (normalize email)
		shareEmail = strings.ToLower(strings.TrimSpace(shareEmail))

		var sharedUser models.User
		if err := database.DB.Where("LOWER(email) = ?", shareEmail).First(&sharedUser).Error; err == nil {
			// User exists, create share (vouchers have no edit permissions)
			share := models.VoucherShare{
				VoucherID:    voucher.ID,
				SharedWithID: sharedUser.ID,
			}

			if err := database.DB.Create(&share).Error; err != nil {
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
