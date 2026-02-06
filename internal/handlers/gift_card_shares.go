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

// GiftCardSharesHandler handles gift card sharing operations using the unified share handler.
// Gift cards support granular permissions including CanEditTransactions.
// Eliminates code duplication by delegating to shares.BaseShareHandler.
type GiftCardSharesHandler struct {
	baseHandler         *shares.BaseShareHandler
	db                  *gorm.DB
	authzService        services.AuthzServiceInterface
	userService         services.UserServiceInterface
	notificationService services.NotificationServiceInterface
}

// NewGiftCardSharesHandler creates a new gift card shares handler.
func NewGiftCardSharesHandler(db *gorm.DB, authzService services.AuthzServiceInterface, userService services.UserServiceInterface, notificationService services.NotificationServiceInterface) *GiftCardSharesHandler {
	adapter := shares.NewGiftCardShareAdapter(db, authzService, userService, notificationService)
	return &GiftCardSharesHandler{
		baseHandler:         shares.NewBaseShareHandler(adapter, userService),
		db:                  db,
		authzService:        authzService,
		userService:         userService,
		notificationService: notificationService,
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
func (h *GiftCardSharesHandler) NewInline(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	giftCardID := c.Param("id")

	giftCardUUID, err := uuid.Parse(giftCardID)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid gift card ID")
	}

	perms, err := h.authzService.CheckGiftCardAccess(c.Request().Context(), user.ID, giftCardUUID)
	if err != nil || !perms.IsOwner {
		return c.String(http.StatusNotFound, "Gift card not found")
	}

	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = ""
	}

	component := templates.GiftCardShareInlineForm(c.Request().Context(), csrfToken, giftCardID)
	return component.Render(c.Request().Context(), c.Response().Writer)
}

// Cancel closes the inline share form without saving.
// Delegates to BaseShareHandler for unified cancel logic.
func (h *GiftCardSharesHandler) Cancel(c echo.Context) error {
	return h.baseHandler.Cancel(c)
}

// EditInline renders the inline share edit form.
func (h *GiftCardSharesHandler) EditInline(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	giftCardID := c.Param("id")
	shareID := c.Param("share_id")

	giftCardUUID, err := uuid.Parse(giftCardID)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid gift card ID")
	}

	perms, err := h.authzService.CheckGiftCardAccess(c.Request().Context(), user.ID, giftCardUUID)
	if err != nil || !perms.IsOwner {
		return c.String(http.StatusNotFound, "Gift card not found")
	}

	var share models.GiftCardShare
	if err := h.db.Where("id = ? AND gift_card_id = ?", shareID, giftCardID).
		Preload("SharedWithUser").First(&share).Error; err != nil {
		return c.String(http.StatusNotFound, "Share not found")
	}

	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = ""
	}

	component := templates.GiftCardShareInlineEdit(c.Request().Context(), csrfToken, giftCardID, share)
	return component.Render(c.Request().Context(), c.Response().Writer)
}

// CancelEdit closes the inline edit form without saving.
func (h *GiftCardSharesHandler) CancelEdit(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	giftCardID := c.Param("id")
	shareID := c.Param("share_id")

	giftCardUUID, err := uuid.Parse(giftCardID)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid gift card ID")
	}

	perms, err := h.authzService.CheckGiftCardAccess(c.Request().Context(), user.ID, giftCardUUID)
	if err != nil || !perms.IsOwner {
		return c.String(http.StatusNotFound, "Gift card not found")
	}

	var share models.GiftCardShare
	if err := h.db.Where("id = ? AND gift_card_id = ?", shareID, giftCardID).
		Preload("SharedWithUser").First(&share).Error; err != nil {
		return c.String(http.StatusNotFound, "Share not found")
	}

	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = ""
	}

	component := templates.GiftCardShareDisplay(c.Request().Context(), csrfToken, giftCardID, share, true)
	return component.Render(c.Request().Context(), c.Response().Writer)
}
