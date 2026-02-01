// Package handlers contains HTTP request handlers for the savvy system.
package handlers

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"net/http"
	"savvy/internal/handlers/shares"
	"savvy/internal/models"
	"savvy/internal/services"
	"savvy/internal/templates"
)

// VoucherSharesHandler handles voucher sharing operations using the unified share handler.
// Vouchers support read-only sharing only (no permission editing).
// Eliminates code duplication by delegating to shares.BaseShareHandler.
type VoucherSharesHandler struct {
	baseHandler  *shares.BaseShareHandler
	db           *gorm.DB
	authzService services.AuthzServiceInterface
	userService  services.UserServiceInterface
}

// NewVoucherSharesHandler creates a new voucher shares handler.
func NewVoucherSharesHandler(db *gorm.DB, authzService services.AuthzServiceInterface, userService services.UserServiceInterface) *VoucherSharesHandler {
	adapter := shares.NewVoucherShareAdapter(db, authzService, userService)
	return &VoucherSharesHandler{
		baseHandler:  shares.NewBaseShareHandler(adapter, userService),
		db:           db,
		authzService: authzService,
		userService:  userService,
	}
}

// Create creates a new voucher share (read-only).
// Delegates to BaseShareHandler for unified share creation logic.
func (h *VoucherSharesHandler) Create(c echo.Context) error {
	return h.baseHandler.Create(c)
}

// Delete removes a voucher share.
// Delegates to BaseShareHandler for unified deletion logic.
func (h *VoucherSharesHandler) Delete(c echo.Context) error {
	return h.baseHandler.Delete(c)
}

// NewInline renders the inline share creation form.
func (h *VoucherSharesHandler) NewInline(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	voucherID := c.Param("id")

	voucherUUID, err := uuid.Parse(voucherID)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid voucher ID")
	}

	// Check if user owns the voucher using AuthzService
	perms, err := h.authzService.CheckVoucherAccess(c.Request().Context(), user.ID, voucherUUID)
	if err != nil || !perms.IsOwner {
		return c.String(http.StatusNotFound, "Voucher not found")
	}

	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = ""
	}

	component := templates.VoucherShareInlineForm(c.Request().Context(), csrfToken, voucherID)
	return component.Render(c.Request().Context(), c.Response().Writer)
}

// Cancel closes the inline share form without saving.
// Delegates to BaseShareHandler for unified cancel logic.
func (h *VoucherSharesHandler) Cancel(c echo.Context) error {
	return h.baseHandler.Cancel(c)
}

// Note: EditInline and CancelEdit are NOT implemented for vouchers
// because voucher shares are read-only (no permission editing).
