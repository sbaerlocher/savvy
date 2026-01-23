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

// GiftCardSharesHandler handles gift card sharing operations
type GiftCardSharesHandler struct {
	authzService services.AuthzServiceInterface
}

// NewGiftCardSharesHandler creates a new gift card shares handler
func NewGiftCardSharesHandler(authzService services.AuthzServiceInterface) *GiftCardSharesHandler {
	return &GiftCardSharesHandler{
		authzService: authzService,
	}
}

// Create creates a new share
func (h *GiftCardSharesHandler) Create(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	giftCardID := c.Param("id")
	giftCardUUID, err := uuid.Parse(giftCardID)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid gift card ID")
	}

	// Check authorization (only owners can create shares)
	perms, err := h.authzService.CheckGiftCardAccess(c.Request().Context(), user.ID, giftCardUUID)
	if err != nil {
		return c.String(http.StatusNotFound, "Gift card not found")
	}
	if !perms.IsOwner {
		return c.String(http.StatusForbidden, "Not authorized")
	}

	// Parse form
	email := strings.ToLower(strings.TrimSpace(c.FormValue("shared_with_email")))
	canEdit := c.FormValue("can_edit") == "on"
	canDelete := c.FormValue("can_delete") == "on"
	canEditTransactions := c.FormValue("can_edit_transactions") == "on"

	// Check if HTMX request
	isHTMX := c.Request().Header.Get("HX-Request") == "true"

	// Validate email exists (case-insensitive)
	var sharedUser models.User
	if err := database.DB.Where("LOWER(email) = ?", email).First(&sharedUser).Error; err != nil {
		if isHTMX {
			csrfToken := c.Get("csrf").(string)
			component := templates.GiftCardShareInlineFormError(c.Request().Context(), csrfToken, giftCardID, "Benutzer nicht gefunden")
			return component.Render(c.Request().Context(), c.Response().Writer)
		}
		return c.String(http.StatusBadRequest, "User not found")
	}

	// Check if already shared
	var existingShare models.GiftCardShare
	if err := database.DB.Where("gift_card_id = ? AND shared_with_id = ?", giftCardID, sharedUser.ID).First(&existingShare).Error; err == nil {
		if isHTMX {
			csrfToken := c.Get("csrf").(string)
			component := templates.GiftCardShareInlineFormError(c.Request().Context(), csrfToken, giftCardID, "Bereits mit diesem Benutzer geteilt")
			return component.Render(c.Request().Context(), c.Response().Writer)
		}
		return c.String(http.StatusConflict, "Already shared with this user")
	}

	// Create share
	share := models.GiftCardShare{
		GiftCardID:          giftCardUUID,
		SharedWithID:        sharedUser.ID,
		CanEdit:             canEdit,
		CanDelete:           canDelete,
		CanEditTransactions: canEditTransactions,
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

// Update updates share permissions
func (h *GiftCardSharesHandler) Update(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	giftCardID := c.Param("id")
	shareID := c.Param("share_id")

	giftCardUUID, err := uuid.Parse(giftCardID)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid gift card ID")
	}

	// Check authorization (only owners can update shares)
	perms, err := h.authzService.CheckGiftCardAccess(c.Request().Context(), user.ID, giftCardUUID)
	if err != nil {
		return c.String(http.StatusNotFound, "Gift card not found")
	}
	if !perms.IsOwner {
		return c.String(http.StatusForbidden, "Not authorized")
	}

	// Get share
	var share models.GiftCardShare
	if err := database.DB.Where("id = ? AND gift_card_id = ?", shareID, giftCardID).First(&share).Error; err != nil {
		return c.String(http.StatusNotFound, "Share not found")
	}

	// Update permissions
	share.CanEdit = c.FormValue("can_edit") == "on"
	share.CanDelete = c.FormValue("can_delete") == "on"
	share.CanEditTransactions = c.FormValue("can_edit_transactions") == "on"

	if err := database.DB.Save(&share).Error; err != nil {
		return c.String(http.StatusInternalServerError, "Error updating share")
	}

	// Check if HTMX request
	isHTMX := c.Request().Header.Get("HX-Request") == "true"

	if isHTMX {
		// Reload share with user for display
		database.DB.Preload("SharedWithUser").First(&share, "id = ?", shareID)
		csrfToken := c.Get("csrf").(string)
		component := templates.GiftCardShareDisplay(c.Request().Context(), csrfToken, giftCardID, share, perms.IsOwner)
		return component.Render(c.Request().Context(), c.Response().Writer)
	}

	return c.String(http.StatusOK, "Share updated")
}

// Delete removes a share
func (h *GiftCardSharesHandler) Delete(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	giftCardID := c.Param("id")
	shareID := c.Param("share_id")

	giftCardUUID, err := uuid.Parse(giftCardID)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid gift card ID")
	}

	// Check authorization (only owners can delete shares)
	perms, err := h.authzService.CheckGiftCardAccess(c.Request().Context(), user.ID, giftCardUUID)
	if err != nil {
		return c.String(http.StatusNotFound, "Gift card not found")
	}
	if !perms.IsOwner {
		return c.String(http.StatusForbidden, "Not authorized")
	}

	// Get share first (for audit logging)
	var share models.GiftCardShare
	if err := database.DB.Where("id = ?", shareID).First(&share).Error; err != nil {
		return c.String(http.StatusNotFound, "Share not found")
	}

	// Delete share with user context for audit logging
	return deleteShareWithAudit(c, user.ID, &share)
}

// NewInline returns the inline share form
func (h *GiftCardSharesHandler) NewInline(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	giftCardID := c.Param("id")

	giftCardUUID, err := uuid.Parse(giftCardID)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid gift card ID")
	}

	// Check authorization (only owners can create shares)
	perms, err := h.authzService.CheckGiftCardAccess(c.Request().Context(), user.ID, giftCardUUID)
	if err != nil {
		return c.String(http.StatusNotFound, "Gift card not found")
	}
	if !perms.IsOwner {
		return c.String(http.StatusForbidden, "Not authorized")
	}

	csrfToken := c.Get("csrf").(string)
	component := templates.GiftCardShareInlineForm(c.Request().Context(), csrfToken, giftCardID)
	return component.Render(c.Request().Context(), c.Response().Writer)
}

// Cancel closes the inline form
func (h *GiftCardSharesHandler) Cancel(c echo.Context) error {
	return c.String(http.StatusOK, "")
}

// EditInline returns the inline edit form
func (h *GiftCardSharesHandler) EditInline(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	giftCardID := c.Param("id")
	shareID := c.Param("share_id")

	giftCardUUID, err := uuid.Parse(giftCardID)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid gift card ID")
	}

	// Check authorization (only owners can edit shares)
	perms, err := h.authzService.CheckGiftCardAccess(c.Request().Context(), user.ID, giftCardUUID)
	if err != nil {
		return c.String(http.StatusNotFound, "Gift card not found")
	}
	if !perms.IsOwner {
		return c.String(http.StatusForbidden, "Not authorized")
	}

	// Get share with user
	var share models.GiftCardShare
	if err := database.DB.Where("id = ? AND gift_card_id = ?", shareID, giftCardID).Preload("SharedWithUser").First(&share).Error; err != nil {
		return c.String(http.StatusNotFound, "Share not found")
	}

	csrfToken := c.Get("csrf").(string)
	component := templates.GiftCardShareInlineEdit(c.Request().Context(), csrfToken, giftCardID, share)
	return component.Render(c.Request().Context(), c.Response().Writer)
}

// CancelEdit cancels inline editing and returns to display
func (h *GiftCardSharesHandler) CancelEdit(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	giftCardID := c.Param("id")
	shareID := c.Param("share_id")

	giftCardUUID, err := uuid.Parse(giftCardID)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid gift card ID")
	}

	// Check authorization
	perms, err := h.authzService.CheckGiftCardAccess(c.Request().Context(), user.ID, giftCardUUID)
	if err != nil {
		return c.String(http.StatusNotFound, "Gift card not found")
	}

	// Get share with user
	var share models.GiftCardShare
	if err := database.DB.Where("id = ? AND gift_card_id = ?", shareID, giftCardID).Preload("SharedWithUser").First(&share).Error; err != nil {
		return c.String(http.StatusNotFound, "Share not found")
	}

	csrfToken := c.Get("csrf").(string)
	component := templates.GiftCardShareDisplay(c.Request().Context(), csrfToken, giftCardID, share, perms.IsOwner)
	return component.Render(c.Request().Context(), c.Response().Writer)
}
