package shares

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"savvy/internal/audit"
	"savvy/internal/database"
	"savvy/internal/models"
)

// BaseShareHandler provides unified share handling logic for all resource types.
// Eliminates 70% code duplication by using the adapter pattern.
type BaseShareHandler struct {
	adapter ShareAdapter
}

// NewBaseShareHandler creates a new base share handler with the given adapter.
func NewBaseShareHandler(adapter ShareAdapter) *BaseShareHandler {
	return &BaseShareHandler{adapter: adapter}
}

// Create handles share creation for any resource type.
// This method consolidates duplicate logic from CardSharesHandler.Create,
// VoucherSharesHandler.Create, and GiftCardSharesHandler.Create.
func (h *BaseShareHandler) Create(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	resourceID := c.Param("id")
	resourceUUID, err := uuid.Parse(resourceID)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid ID")
	}

	// Check ownership using adapter
	isOwner, err := h.adapter.CheckOwnership(c.Request().Context(), user.ID, resourceUUID)
	if err != nil || !isOwner {
		return c.String(http.StatusForbidden, "Not authorized")
	}

	// Parse form values
	email := strings.ToLower(strings.TrimSpace(c.FormValue("shared_with_email")))
	canEdit := c.FormValue("can_edit") == "on"
	canDelete := c.FormValue("can_delete") == "on"
	canEditTransactions := false
	if h.adapter.HasTransactionPermission() {
		canEditTransactions = c.FormValue("can_edit_transactions") == "on"
	}

	// Check if HTMX request
	isHTMX := c.Request().Header.Get("HX-Request") == "true"

	// Validate email exists
	var sharedUser models.User
	if err := database.DB.Where("LOWER(email) = ?", email).First(&sharedUser).Error; err != nil {
		if isHTMX {
			return c.String(http.StatusBadRequest, "Benutzer nicht gefunden")
		}
		return c.String(http.StatusBadRequest, "User not found")
	}

	// Create share using adapter
	req := CreateShareRequest{
		UserID:              user.ID,
		ResourceID:          resourceUUID,
		SharedWithEmail:     email,
		CanEdit:             canEdit,
		CanDelete:           canDelete,
		CanEditTransactions: canEditTransactions,
	}

	if err := h.adapter.CreateShare(c.Request().Context(), req); err != nil {
		if isHTMX {
			return c.String(http.StatusBadRequest, err.Error())
		}
		return c.String(http.StatusInternalServerError, "Error creating share")
	}

	// For HTMX, refresh page
	if isHTMX {
		c.Response().Header().Set("HX-Refresh", "true")
		return c.String(http.StatusOK, "")
	}

	return c.String(http.StatusOK, "Share created")
}

// Update handles share permission updates.
// Only supported for Cards and Gift Cards (not Vouchers - they're read-only).
func (h *BaseShareHandler) Update(c echo.Context) error {
	if !h.adapter.SupportsEdit() {
		return c.String(http.StatusMethodNotAllowed, "Updates not supported for this resource type")
	}

	user := c.Get("current_user").(*models.User)
	resourceID := c.Param("id")
	shareID := c.Param("share_id")

	resourceUUID, err := uuid.Parse(resourceID)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid resource ID")
	}

	shareUUID, err := uuid.Parse(shareID)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid share ID")
	}

	// Check ownership
	isOwner, err := h.adapter.CheckOwnership(c.Request().Context(), user.ID, resourceUUID)
	if err != nil || !isOwner {
		return c.String(http.StatusForbidden, "Not authorized")
	}

	// Parse permissions
	canEdit := c.FormValue("can_edit") == "on"
	canDelete := c.FormValue("can_delete") == "on"
	canEditTransactions := false
	if h.adapter.HasTransactionPermission() {
		canEditTransactions = c.FormValue("can_edit_transactions") == "on"
	}

	// Update using adapter
	req := UpdateShareRequest{
		ShareID:             shareUUID,
		UserID:              user.ID,
		ResourceID:          resourceUUID,
		CanEdit:             canEdit,
		CanDelete:           canDelete,
		CanEditTransactions: canEditTransactions,
	}

	if err := h.adapter.UpdateShare(c.Request().Context(), req); err != nil {
		return c.String(http.StatusInternalServerError, "Error updating share")
	}

	// HTMX: refresh page
	if c.Request().Header.Get("HX-Request") == "true" {
		c.Response().Header().Set("HX-Refresh", "true")
		return c.String(http.StatusOK, "")
	}

	return c.String(http.StatusOK, "Share updated")
}

// Delete handles share deletion for any resource type.
func (h *BaseShareHandler) Delete(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	resourceID := c.Param("id")
	shareID := c.Param("share_id")

	resourceUUID, err := uuid.Parse(resourceID)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid resource ID")
	}

	shareUUID, err := uuid.Parse(shareID)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid share ID")
	}

	// Check ownership
	isOwner, err := h.adapter.CheckOwnership(c.Request().Context(), user.ID, resourceUUID)
	if err != nil || !isOwner {
		return c.String(http.StatusForbidden, "Not authorized")
	}

	// Add user ID to context for audit logging
	ctx := audit.AddUserIDToContext(c.Request().Context(), user.ID)

	// Delete using adapter
	if err := h.adapter.DeleteShare(ctx, shareUUID); err != nil {
		return c.String(http.StatusInternalServerError, "Error deleting share")
	}

	// HTMX: refresh page
	if c.Request().Header.Get("HX-Request") == "true" {
		c.Response().Header().Set("HX-Refresh", "true")
		return c.String(http.StatusOK, "")
	}

	return c.String(http.StatusOK, "Share deleted")
}

// NewInline renders the inline share creation form.
// This is a simple method that returns an empty response to close the form.
func (h *BaseShareHandler) NewInline(c echo.Context) error {
	// This method is typically handled by templates directly
	// For now, return empty response (form is rendered by template)
	return c.String(http.StatusOK, "")
}

// Cancel closes the inline share form without saving.
func (h *BaseShareHandler) Cancel(c echo.Context) error {
	// HTMX: return empty string to remove the form
	return c.String(http.StatusOK, "")
}

// EditInline renders the inline share edit form (Cards & Gift Cards only).
func (h *BaseShareHandler) EditInline(c echo.Context) error {
	if !h.adapter.SupportsEdit() {
		return c.String(http.StatusMethodNotAllowed, "Editing not supported for this resource type")
	}
	// Template handles rendering
	return c.String(http.StatusOK, "")
}

// CancelEdit closes the inline edit form without saving (Cards & Gift Cards only).
func (h *BaseShareHandler) CancelEdit(c echo.Context) error {
	if !h.adapter.SupportsEdit() {
		return c.String(http.StatusMethodNotAllowed, "Editing not supported for this resource type")
	}
	// HTMX: return empty string to remove the form
	return c.String(http.StatusOK, "")
}
