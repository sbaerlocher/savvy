// Package handlers contains HTTP request handlers for the savvy system.
package handlers

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"savvy/internal/handlers/shares"
)

// CardSharesHandler handles card sharing operations using the unified share handler.
// Eliminates code duplication by delegating to shares.BaseShareHandler.
type CardSharesHandler struct {
	baseHandler *shares.BaseShareHandler
}

// NewCardSharesHandler creates a new card shares handler.
func NewCardSharesHandler(db *gorm.DB) *CardSharesHandler {
	adapter := shares.NewCardShareAdapter(db)
	return &CardSharesHandler{
		baseHandler: shares.NewBaseShareHandler(adapter),
	}
}

// Create creates a new card share.
// Delegates to BaseShareHandler for unified share creation logic.
func (h *CardSharesHandler) Create(c echo.Context) error {
	return h.baseHandler.Create(c)
}

// Update updates share permissions (CanEdit, CanDelete).
// Delegates to BaseShareHandler for unified update logic.
func (h *CardSharesHandler) Update(c echo.Context) error {
	return h.baseHandler.Update(c)
}

// Delete removes a card share.
// Delegates to BaseShareHandler for unified deletion logic.
func (h *CardSharesHandler) Delete(c echo.Context) error {
	return h.baseHandler.Delete(c)
}

// NewInline renders the inline share creation form.
// Delegates to BaseShareHandler for unified form rendering.
func (h *CardSharesHandler) NewInline(c echo.Context) error {
	return h.baseHandler.NewInline(c)
}

// Cancel closes the inline share form without saving.
// Delegates to BaseShareHandler for unified cancel logic.
func (h *CardSharesHandler) Cancel(c echo.Context) error {
	return h.baseHandler.Cancel(c)
}

// EditInline renders the inline share edit form.
// Delegates to BaseShareHandler for unified edit form rendering.
func (h *CardSharesHandler) EditInline(c echo.Context) error {
	return h.baseHandler.EditInline(c)
}

// CancelEdit closes the inline edit form without saving.
// Delegates to BaseShareHandler for unified cancel edit logic.
func (h *CardSharesHandler) CancelEdit(c echo.Context) error {
	return h.baseHandler.CancelEdit(c)
}
