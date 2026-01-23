package vouchers

import (
	"net/http"
	"savvy/internal/database"
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
		return c.String(http.StatusBadRequest, "Invalid voucher ID")
	}

	// Check authorization
	perms, err := h.authzService.CheckVoucherAccess(c.Request().Context(), user.ID, voucherID)
	if err != nil || !perms.CanEdit {
		return c.String(http.StatusForbidden, "No edit permission")
	}

	var voucher models.Voucher
	if err := database.DB.Where("id = ?", voucherID).Preload("Merchant").First(&voucher).Error; err != nil {
		return c.String(http.StatusNotFound, "Voucher not found")
	}

	// Load all merchants for dropdown
	var merchants []models.Merchant
	database.DB.Order("name ASC").Find(&merchants)

	csrfToken := c.Get("csrf").(string)
	return templates.VoucherDetailEdit(c.Request().Context(), csrfToken, voucher, merchants).Render(c.Request().Context(), c.Response().Writer)
}

// CancelEdit returns the view mode for a voucher
func (h *Handler) CancelEdit(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	voucherID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid voucher ID")
	}

	// Check authorization
	perms, err := h.authzService.CheckVoucherAccess(c.Request().Context(), user.ID, voucherID)
	if err != nil {
		return c.String(http.StatusForbidden, "No access")
	}

	var voucher models.Voucher
	if err := database.DB.Where("id = ?", voucherID).Preload("Merchant").Preload("User").First(&voucher).Error; err != nil {
		return c.String(http.StatusNotFound, "Voucher not found")
	}

	canEdit := perms.CanEdit

	// Check if voucher is favorited by current user
	var favorite models.UserFavorite
	isFavorite := database.DB.Where("user_id = ? AND resource_type = ? AND resource_id = ?",
		user.ID, "voucher", voucherID).First(&favorite).Error == nil

	csrfToken := c.Get("csrf").(string)
	return templates.VoucherDetailView(c.Request().Context(), csrfToken, voucher, canEdit, user, isFavorite).Render(c.Request().Context(), c.Response().Writer)
}

// UpdateInline updates a voucher and returns the view mode
func (h *Handler) UpdateInline(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	voucherID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid voucher ID")
	}

	// Check authorization
	perms, err := h.authzService.CheckVoucherAccess(c.Request().Context(), user.ID, voucherID)
	if err != nil || !perms.CanEdit {
		return c.String(http.StatusForbidden, "No edit permission")
	}

	var voucher models.Voucher
	if err := database.DB.Where("id = ?", voucherID).First(&voucher).Error; err != nil {
		return c.String(http.StatusNotFound, "Voucher not found")
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
		return c.String(http.StatusBadRequest, "Invalid date range")
	}

	// Handle merchant selection
	merchantIDStr := c.FormValue("merchant_id")
	if merchantIDStr != "" && merchantIDStr != "new" {
		// Existing merchant selected from dropdown
		merchantID, err := uuid.Parse(merchantIDStr)
		if err == nil {
			voucher.MerchantID = &merchantID
			// Load merchant to get name
			var merchant models.Merchant
			if err := database.DB.Where("id = ?", merchantID).First(&merchant).Error; err == nil {
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

	if err := database.DB.Save(&voucher).Error; err != nil {
		c.Logger().Errorf("Failed to update voucher: %v", err)
		return c.String(http.StatusInternalServerError, "Failed to update voucher")
	}

	// Reload with merchant and user
	database.DB.Where("id = ?", voucherID).Preload("Merchant").Preload("User").First(&voucher)

	// Check if voucher is favorited by current user
	var favorite models.UserFavorite
	isFavorite := database.DB.Where("user_id = ? AND resource_type = ? AND resource_id = ?",
		user.ID, "voucher", voucherID).First(&favorite).Error == nil

	canEdit := perms.CanEdit
	csrfToken := c.Get("csrf").(string)
	return templates.VoucherDetailView(c.Request().Context(), csrfToken, voucher, canEdit, user, isFavorite).Render(c.Request().Context(), c.Response().Writer)
}
