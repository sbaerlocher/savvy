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

// CardSharesHandler handles card sharing operations
type CardSharesHandler struct {
	authzService services.AuthzServiceInterface
}

// NewCardSharesHandler creates a new card shares handler
func NewCardSharesHandler(authzService services.AuthzServiceInterface) *CardSharesHandler {
	return &CardSharesHandler{
		authzService: authzService,
	}
}

// Create creates a new share
func (h *CardSharesHandler) Create(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	cardID := c.Param("id")
	cardUUID, err := uuid.Parse(cardID)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid card ID")
	}

	// Check authorization (only owners can create shares)
	perms, err := h.authzService.CheckCardAccess(c.Request().Context(), user.ID, cardUUID)
	if err != nil {
		return c.String(http.StatusNotFound, "Card not found")
	}
	if !perms.IsOwner {
		return c.String(http.StatusForbidden, "Not authorized")
	}

	// Parse form
	email := strings.ToLower(strings.TrimSpace(c.FormValue("shared_with_email")))
	canEdit := c.FormValue("can_edit") == "on"
	canDelete := c.FormValue("can_delete") == "on"

	// Check if HTMX request
	isHTMX := c.Request().Header.Get("HX-Request") == trueStringValue

	// Validate email exists (case-insensitive)
	var sharedUser models.User
	if err := database.DB.Where("LOWER(email) = ?", email).First(&sharedUser).Error; err != nil {
		if isHTMX {
			csrfToken := c.Get("csrf").(string)
			component := templates.CardShareInlineFormError(c.Request().Context(), csrfToken, cardID, "Benutzer nicht gefunden")
			return component.Render(c.Request().Context(), c.Response().Writer)
		}
		return c.String(http.StatusBadRequest, "User not found")
	}

	// Check if already shared
	var existingShare models.CardShare
	if err := database.DB.Where("card_id = ? AND shared_with_id = ?", cardID, sharedUser.ID).First(&existingShare).Error; err == nil {
		if isHTMX {
			csrfToken := c.Get("csrf").(string)
			component := templates.CardShareInlineFormError(c.Request().Context(), csrfToken, cardID, "Bereits mit diesem Benutzer geteilt")
			return component.Render(c.Request().Context(), c.Response().Writer)
		}
		return c.String(http.StatusConflict, "Already shared with this user")
	}

	// Create share
	share := models.CardShare{
		CardID:       cardUUID,
		SharedWithID: sharedUser.ID,
		CanEdit:      canEdit,
		CanDelete:    canDelete,
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
func (h *CardSharesHandler) Update(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	cardID := c.Param("id")
	shareID := c.Param("share_id")

	cardUUID, err := uuid.Parse(cardID)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid card ID")
	}

	// Check authorization (only owners can update shares)
	perms, err := h.authzService.CheckCardAccess(c.Request().Context(), user.ID, cardUUID)
	if err != nil {
		return c.String(http.StatusNotFound, "Card not found")
	}
	if !perms.IsOwner {
		return c.String(http.StatusForbidden, "Not authorized")
	}

	// Get share
	var share models.CardShare
	if err := database.DB.Where("id = ? AND card_id = ?", shareID, cardID).First(&share).Error; err != nil {
		return c.String(http.StatusNotFound, "Share not found")
	}

	// Update permissions
	share.CanEdit = c.FormValue("can_edit") == "on"
	share.CanDelete = c.FormValue("can_delete") == "on"

	if err := database.DB.Save(&share).Error; err != nil {
		return c.String(http.StatusInternalServerError, "Error updating share")
	}

	// Check if HTMX request
	isHTMX := c.Request().Header.Get("HX-Request") == trueStringValue

	if isHTMX {
		// Reload share with user for display
		database.DB.Preload("SharedWithUser").First(&share, "id = ?", shareID)
		csrfToken := c.Get("csrf").(string)
		component := templates.CardShareDisplay(c.Request().Context(), csrfToken, cardID, share, perms.IsOwner)
		return component.Render(c.Request().Context(), c.Response().Writer)
	}

	return c.String(http.StatusOK, "Share updated")
}

// Delete removes a share
func (h *CardSharesHandler) Delete(c echo.Context) error {
	var share models.CardShare
	return handleDeleteShare(c, h.authzService.CheckCardAccess, &share, "id", "card")
}

// NewInline returns the inline share form
func (h *CardSharesHandler) NewInline(c echo.Context) error {
	return handleNewInlineShare(c, h.authzService.CheckCardAccess, templates.CardShareInlineForm, "id", "card")
}

// Cancel closes the inline form
func (h *CardSharesHandler) Cancel(c echo.Context) error {
	return c.String(http.StatusOK, "")
}

// EditInline returns the inline edit form
func (h *CardSharesHandler) EditInline(c echo.Context) error {
	var share models.CardShare
	data, err := loadShareWithAuthz(c, h.authzService.CheckCardAccess, &share, "card_id", "card", true)
	if err != nil {
		return err
	}
	component := templates.CardShareInlineEdit(data.Context, data.CSRF, data.ResID, share)
	return component.Render(data.Context, c.Response().Writer)
}

// CancelEdit cancels inline editing and returns to display
func (h *CardSharesHandler) CancelEdit(c echo.Context) error {
	var share models.CardShare
	data, err := loadShareWithAuthz(c, h.authzService.CheckCardAccess, &share, "card_id", "card", false)
	if err != nil {
		return err
	}
	component := templates.CardShareDisplay(data.Context, data.CSRF, data.ResID, share, data.Perms.IsOwner)
	return component.Render(data.Context, c.Response().Writer)
}
