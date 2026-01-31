// Package handlers contains HTTP request handlers for the savvy system.
package handlers

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"savvy/internal/handlers/shares"
)

// GiftCardSharesHandler handles gift card sharing operations using the unified share handler.
// Gift cards support granular permissions including CanEditTransactions.
// Eliminates code duplication by delegating to shares.BaseShareHandler.
type GiftCardSharesHandler struct {
	baseHandler *shares.BaseShareHandler
}

// NewGiftCardSharesHandler creates a new gift card shares handler.
func NewGiftCardSharesHandler(db *gorm.DB) *GiftCardSharesHandler {
	adapter := shares.NewGiftCardShareAdapter(db)
	return &GiftCardSharesHandler{
		baseHandler: shares.NewBaseShareHandler(adapter),
	}
}

// Create creates a new gift card share.
// Delegates to BaseShareHandler for unified share creation logic.
func (h *GiftCardSharesHandler) Create(c echo.Context) error {
	return h.baseHandler.Create(c)
}

// Update updates share permissions (CanEdit, CanDelete, CanEditTransactions).
// Delegates to BaseShareHandler for unified update logic.
func (h *GiftCardSharesHandler) Update(c echo.Context) error {
	return h.baseHandler.Update(c)
}

// Delete removes a gift card share.
// Delegates to BaseShareHandler for unified deletion logic.
func (h *GiftCardSharesHandler) Delete(c echo.Context) error {
	return h.baseHandler.Delete(c)
}

// NewInline renders the inline share creation form.
// Delegates to BaseShareHandler for unified form rendering.
func (h *GiftCardSharesHandler) NewInline(c echo.Context) error {
	return h.baseHandler.NewInline(c)
}

// Cancel closes the inline share form without saving.
// Delegates to BaseShareHandler for unified cancel logic.
func (h *GiftCardSharesHandler) Cancel(c echo.Context) error {
	return h.baseHandler.Cancel(c)
}

// EditInline renders the inline share edit form.
// Delegates to BaseShareHandler for unified edit form rendering.
func (h *GiftCardSharesHandler) EditInline(c echo.Context) error {
	return h.baseHandler.EditInline(c)
}

// CancelEdit closes the inline edit form without saving.
// Delegates to BaseShareHandler for unified cancel edit logic.
func (h *GiftCardSharesHandler) CancelEdit(c echo.Context) error {
	return h.baseHandler.CancelEdit(c)
}
