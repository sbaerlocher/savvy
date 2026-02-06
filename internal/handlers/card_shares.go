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

// CardSharesHandler handles card sharing operations using the unified share handler.
// Eliminates code duplication by delegating to shares.BaseShareHandler.
type CardSharesHandler struct {
	baseHandler         *shares.BaseShareHandler
	db                  *gorm.DB
	authzService        services.AuthzServiceInterface
	userService         services.UserServiceInterface
	notificationService services.NotificationServiceInterface
}

// NewCardSharesHandler creates a new card shares handler.
func NewCardSharesHandler(db *gorm.DB, authzService services.AuthzServiceInterface, userService services.UserServiceInterface, notificationService services.NotificationServiceInterface) *CardSharesHandler {
	adapter := shares.NewCardShareAdapter(db, authzService, userService, notificationService)
	return &CardSharesHandler{
		baseHandler:         shares.NewBaseShareHandler(adapter, userService),
		db:                  db,
		authzService:        authzService,
		userService:         userService,
		notificationService: notificationService,
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
func (h *CardSharesHandler) NewInline(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	cardID := c.Param("id")

	cardUUID, err := uuid.Parse(cardID)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid card ID")
	}

	perms, err := h.authzService.CheckCardAccess(c.Request().Context(), user.ID, cardUUID)
	if err != nil || !perms.IsOwner {
		return c.String(http.StatusNotFound, "Card not found")
	}

	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = ""
	}

	component := templates.CardShareInlineForm(c.Request().Context(), csrfToken, cardID)
	return component.Render(c.Request().Context(), c.Response().Writer)
}

// Cancel closes the inline share form without saving.
// Delegates to BaseShareHandler for unified cancel logic.
func (h *CardSharesHandler) Cancel(c echo.Context) error {
	return h.baseHandler.Cancel(c)
}

// EditInline renders the inline share edit form.
func (h *CardSharesHandler) EditInline(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	cardID := c.Param("id")
	shareID := c.Param("share_id")

	cardUUID, err := uuid.Parse(cardID)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid card ID")
	}

	perms, err := h.authzService.CheckCardAccess(c.Request().Context(), user.ID, cardUUID)
	if err != nil || !perms.IsOwner {
		return c.String(http.StatusNotFound, "Card not found")
	}

	var share models.CardShare
	if err := h.db.Where("id = ? AND card_id = ?", shareID, cardID).
		Preload("SharedWithUser").First(&share).Error; err != nil {
		return c.String(http.StatusNotFound, "Share not found")
	}

	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = ""
	}

	component := templates.CardShareInlineEdit(c.Request().Context(), csrfToken, cardID, share)
	return component.Render(c.Request().Context(), c.Response().Writer)
}

// CancelEdit closes the inline edit form without saving.
func (h *CardSharesHandler) CancelEdit(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	cardID := c.Param("id")
	shareID := c.Param("share_id")

	cardUUID, err := uuid.Parse(cardID)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid card ID")
	}

	perms, err := h.authzService.CheckCardAccess(c.Request().Context(), user.ID, cardUUID)
	if err != nil || !perms.IsOwner {
		return c.String(http.StatusNotFound, "Card not found")
	}

	var share models.CardShare
	if err := h.db.Where("id = ? AND card_id = ?", shareID, cardID).
		Preload("SharedWithUser").First(&share).Error; err != nil {
		return c.String(http.StatusNotFound, "Share not found")
	}

	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = ""
	}

	component := templates.CardShareDisplay(c.Request().Context(), csrfToken, cardID, share, true)
	return component.Render(c.Request().Context(), c.Response().Writer)
}
