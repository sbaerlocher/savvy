// Package handlers contains HTTP request handlers for the savvy system.
package handlers

import (
	"net/http"
	"savvy/internal/database"
	"savvy/internal/models"
	"savvy/internal/services"
	"savvy/internal/templates"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// VoucherSharesHandler handles voucher sharing operations
type VoucherSharesHandler struct {
	authzService services.AuthzServiceInterface
}

// NewVoucherSharesHandler creates a new voucher shares handler
func NewVoucherSharesHandler(authzService services.AuthzServiceInterface) *VoucherSharesHandler {
	return &VoucherSharesHandler{
		authzService: authzService,
	}
}

// Create creates a new share
func (h *VoucherSharesHandler) Create(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	voucherID := c.Param("id")
	voucherUUID, err := uuid.Parse(voucherID)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid voucher ID")
	}

	// Check authorization (only owners can create shares)
	perms, err := h.authzService.CheckVoucherAccess(c.Request().Context(), user.ID, voucherUUID)
	if err != nil {
		return c.String(http.StatusNotFound, "Voucher not found")
	}
	if !perms.IsOwner {
		return c.String(http.StatusForbidden, "Not authorized")
	}

	// Parse form
	email := strings.ToLower(strings.TrimSpace(c.FormValue("shared_with_email")))

	// Check if HTMX request
	isHTMX := c.Request().Header.Get("HX-Request") == "true"

	// Validate email exists (case-insensitive)
	var sharedUser models.User
	if err := database.DB.Where("LOWER(email) = ?", email).First(&sharedUser).Error; err != nil {
		if isHTMX {
			csrfToken := c.Get("csrf").(string)
			component := templates.VoucherShareInlineFormError(c.Request().Context(), csrfToken, voucherID, "Benutzer nicht gefunden")
			return component.Render(c.Request().Context(), c.Response().Writer)
		}
		return c.String(http.StatusBadRequest, "User not found")
	}

	// Check if already shared
	var existingShare models.VoucherShare
	if err := database.DB.Where("voucher_id = ? AND shared_with_id = ?", voucherID, sharedUser.ID).First(&existingShare).Error; err == nil {
		if isHTMX {
			csrfToken := c.Get("csrf").(string)
			component := templates.VoucherShareInlineFormError(c.Request().Context(), csrfToken, voucherID, "Bereits mit diesem Benutzer geteilt")
			return component.Render(c.Request().Context(), c.Response().Writer)
		}
		return c.String(http.StatusConflict, "Already shared with this user")
	}

	// Create share (vouchers are always read-only)
	share := models.VoucherShare{
		VoucherID:    voucherUUID,
		SharedWithID: sharedUser.ID,
	}

	if err := database.DB.Create(&share).Error; err != nil {
		return c.String(http.StatusInternalServerError, "Error creating share")
	}

	// For HTMX, return empty string to close the form and trigger page reload
	if isHTMX {
		c.Response().Header().Set("HX-Refresh", "true")
		return c.String(http.StatusOK, "")
	}

	return c.String(http.StatusOK, "Share created")
}

// Delete removes a share
func (h *VoucherSharesHandler) Delete(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	voucherID := c.Param("id")
	shareID := c.Param("share_id")

	voucherUUID, err := uuid.Parse(voucherID)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid voucher ID")
	}

	// Check authorization (only owners can delete shares)
	perms, err := h.authzService.CheckVoucherAccess(c.Request().Context(), user.ID, voucherUUID)
	if err != nil {
		return c.String(http.StatusNotFound, "Voucher not found")
	}
	if !perms.IsOwner {
		return c.String(http.StatusForbidden, "Not authorized")
	}

	// Get share first (for audit logging)
	var share models.VoucherShare
	if err := database.DB.Where("id = ?", shareID).First(&share).Error; err != nil {
		return c.String(http.StatusNotFound, "Share not found")
	}

	// Delete share with user context for audit logging
	return deleteShareWithAudit(c, user.ID, &share)
}

// NewInline returns the inline share form
func (h *VoucherSharesHandler) NewInline(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	voucherID := c.Param("id")

	voucherUUID, err := uuid.Parse(voucherID)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid voucher ID")
	}

	// Check authorization (only owners can create shares)
	perms, err := h.authzService.CheckVoucherAccess(c.Request().Context(), user.ID, voucherUUID)
	if err != nil {
		return c.String(http.StatusNotFound, "Voucher not found")
	}
	if !perms.IsOwner {
		return c.String(http.StatusForbidden, "Not authorized")
	}

	csrfToken := c.Get("csrf").(string)
	component := templates.VoucherShareInlineForm(c.Request().Context(), csrfToken, voucherID)
	return component.Render(c.Request().Context(), c.Response().Writer)
}

// Cancel closes the inline form
func (h *VoucherSharesHandler) Cancel(c echo.Context) error {
	return c.String(http.StatusOK, "")
}
