// Package vouchers provides HTTP handlers for voucher management operations.
package vouchers

import (
	"net/http"
	"savvy/internal/i18n"
	"savvy/internal/models"
	"savvy/internal/templates"
	"savvy/internal/validation"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// EditInline returns the inline edit form for a voucher
func (h *Handler) EditInline(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	voucherID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.String(http.StatusBadRequest, i18n.T(c.Request().Context(), "error.invalid_voucher_id"))
	}

	// Check authorization
	perms, err := h.authzService.CheckVoucherAccess(c.Request().Context(), user.ID, voucherID)
	if err != nil || !perms.CanEdit {
		return c.String(http.StatusForbidden, i18n.T(c.Request().Context(), "error.no_edit_permission"))
	}

	voucher, err := h.voucherService.GetVoucher(c.Request().Context(), voucherID)
	if err != nil {
		return c.String(http.StatusNotFound, i18n.T(c.Request().Context(), "error.voucher_not_found"))
	}

	merchants, err := h.merchantService.GetAllMerchants(c.Request().Context())
	if err != nil {
		c.Logger().Errorf("Failed to load merchants: %v", err)
		merchants = []models.Merchant{}
	}

	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = ""
	}
	return templates.VoucherDetailEdit(c.Request().Context(), csrfToken, *voucher, merchants).Render(c.Request().Context(), c.Response().Writer)
}

// CancelEdit returns the view mode for a voucher
func (h *Handler) CancelEdit(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	voucherID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.String(http.StatusBadRequest, i18n.T(c.Request().Context(), "error.invalid_voucher_id"))
	}

	// Check authorization
	perms, err := h.authzService.CheckVoucherAccess(c.Request().Context(), user.ID, voucherID)
	if err != nil {
		return c.String(http.StatusForbidden, i18n.T(c.Request().Context(), "error.no_access"))
	}

	voucher, err := h.voucherService.GetVoucher(c.Request().Context(), voucherID)
	if err != nil {
		return c.String(http.StatusNotFound, i18n.T(c.Request().Context(), "error.voucher_not_found"))
	}

	canEdit := perms.CanEdit

	isFavorite, err := h.favoriteService.IsFavorite(c.Request().Context(), user.ID, "voucher", voucherID)
	if err != nil {
		c.Logger().Errorf("Failed to check favorite status: %v", err)
		isFavorite = false
	}

	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = ""
	}
	return templates.VoucherDetailView(c.Request().Context(), csrfToken, *voucher, canEdit, user, isFavorite).Render(c.Request().Context(), c.Response().Writer)
}

// UpdateInline updates a voucher and returns the view mode
func (h *Handler) UpdateInline(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	voucherID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.String(http.StatusBadRequest, i18n.T(c.Request().Context(), "error.invalid_voucher_id"))
	}

	// Check authorization
	perms, err := h.authzService.CheckVoucherAccess(c.Request().Context(), user.ID, voucherID)
	if err != nil || !perms.CanEdit {
		return c.String(http.StatusForbidden, i18n.T(c.Request().Context(), "error.no_edit_permission"))
	}

	voucher, err := h.voucherService.GetVoucher(c.Request().Context(), voucherID)
	if err != nil {
		return c.String(http.StatusNotFound, i18n.T(c.Request().Context(), "error.voucher_not_found"))
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
		return c.String(http.StatusBadRequest, i18n.T(c.Request().Context(), "error.invalid_date_range"))
	}

	// Handle merchant selection
	merchantIDStr := c.FormValue("merchant_id")
	if merchantIDStr != "" && merchantIDStr != "new" {
		// Existing merchant selected from dropdown
		merchantID, err := uuid.Parse(merchantIDStr)
		if err == nil {
			voucher.MerchantID = &merchantID

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
		c.Logger().Errorf("Failed to update voucher: %v", err)
		return c.String(http.StatusInternalServerError, i18n.T(c.Request().Context(), "error.updating_voucher"))
	}

	// Reload with merchant and user
	voucher, err = h.voucherService.GetVoucher(c.Request().Context(), voucherID)
	if err != nil {
		return c.String(http.StatusNotFound, i18n.T(c.Request().Context(), "error.voucher_not_found"))
	}

	isFavorite, err := h.favoriteService.IsFavorite(c.Request().Context(), user.ID, "voucher", voucherID)
	if err != nil {
		c.Logger().Errorf("Failed to check favorite status: %v", err)
		isFavorite = false
	}

	canEdit := perms.CanEdit
	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = ""
	}
	return templates.VoucherDetailView(c.Request().Context(), csrfToken, *voucher, canEdit, user, isFavorite).Render(c.Request().Context(), c.Response().Writer)
}
