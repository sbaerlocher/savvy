// Package handlers contains HTTP request handlers for the savvy system.
package handlers

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"savvy/internal/handlers/shares"
)

// VoucherSharesHandler handles voucher sharing operations using the unified share handler.
// Vouchers support read-only sharing only (no permission editing).
// Eliminates code duplication by delegating to shares.BaseShareHandler.
type VoucherSharesHandler struct {
	baseHandler *shares.BaseShareHandler
}

// NewVoucherSharesHandler creates a new voucher shares handler.
func NewVoucherSharesHandler(db *gorm.DB) *VoucherSharesHandler {
	adapter := shares.NewVoucherShareAdapter(db)
	return &VoucherSharesHandler{
		baseHandler: shares.NewBaseShareHandler(adapter),
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
// Delegates to BaseShareHandler for unified form rendering.
func (h *VoucherSharesHandler) NewInline(c echo.Context) error {
	return h.baseHandler.NewInline(c)
}

// Cancel closes the inline share form without saving.
// Delegates to BaseShareHandler for unified cancel logic.
func (h *VoucherSharesHandler) Cancel(c echo.Context) error {
	return h.baseHandler.Cancel(c)
}

// Note: EditInline and CancelEdit are NOT implemented for vouchers
// because voucher shares are read-only (no permission editing).
