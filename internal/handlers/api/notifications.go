// Package api contains JSON API handlers for the SvelteKit frontend.
//
//nolint:revive // "api" is a meaningful package name for API handlers
package api

import (
	"net/http"
	"savvy/internal/audit"
	"savvy/internal/models"
	"savvy/internal/services"
	"strconv"

	"github.com/labstack/echo/v5"
)

// NotificationsHandler handles notification-related API requests
type NotificationsHandler struct {
	notificationService services.NotificationServiceInterface
}

// NewNotificationsHandler creates a new notifications API handler
func NewNotificationsHandler(notificationService services.NotificationServiceInterface) *NotificationsHandler {
	return &NotificationsHandler{
		notificationService: notificationService,
	}
}

// NotificationDTO represents a notification in API responses
type NotificationDTO struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	ResourceType string                 `json:"resource_type"`
	ResourceID   string                 `json:"resource_id"`
	Metadata     map[string]interface{} `json:"metadata"`
	IsRead       bool                   `json:"is_read"`
	ReadAt       *string                `json:"read_at,omitempty"`
	CreatedAt    string                 `json:"created_at"`
} // NotificationDTO

// toNotificationDTO converts a notification model to a DTO
func toNotificationDTO(n *models.Notification) NotificationDTO {
	dto := NotificationDTO{
		ID:           n.ID.String(),
		Type:         string(n.Type),
		ResourceType: n.ResourceType,
		ResourceID:   n.ResourceID.String(),
		Metadata:     n.Metadata,
		IsRead:       n.IsRead,
		CreatedAt:    n.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if n.ReadAt != nil {
		readAt := n.ReadAt.Format("2006-01-02T15:04:05Z07:00")
		dto.ReadAt = &readAt
	}

	return dto
}

// List retrieves all notifications for the current user
// GET /api/v1/notifications
func (h *NotificationsHandler) List(c echo.Context) error {
	user := c.Get("current_user").(*models.User)

	// Parse pagination parameters
	limit := 50 // Default limit
	if l := c.QueryParam("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	offset := 0
	if o := c.QueryParam("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Get notifications
	notifications, err := h.notificationService.GetUserNotifications(c.Request().Context(), user.ID, limit, offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch notifications")
	}

	// Convert to DTOs
	dtos := make([]NotificationDTO, len(notifications))
	for i, n := range notifications {
		dtos[i] = toNotificationDTO(&n)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"notifications": dtos,
	})
}

// GetUnreadCount retrieves the count of unread notifications
// GET /api/v1/notifications/unread-count
func (h *NotificationsHandler) GetUnreadCount(c echo.Context) error {
	user := c.Get("current_user").(*models.User)

	count, err := h.notificationService.GetUnreadCount(c.Request().Context(), user.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch unread count")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"count": count,
	})
}

// MarkAsRead marks a notification as read
// POST /api/v1/notifications/:id/read
func (h *NotificationsHandler) MarkAsRead(c echo.Context) error {
	notificationID, err := parseResourceID(c, "notification")
	if err != nil {
		return err
	}

	user := c.Get("current_user").(*models.User)

	// Mark as read (scoped to authenticated user)
	if err := h.notificationService.MarkAsRead(c.Request().Context(), user.ID, notificationID); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Notification not found")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Notification marked as read",
	})
}

// MarkAllAsRead marks all notifications as read for the current user
// POST /api/v1/notifications/read-all
func (h *NotificationsHandler) MarkAllAsRead(c echo.Context) error {
	user := c.Get("current_user").(*models.User)

	if err := h.notificationService.MarkAllAsRead(c.Request().Context(), user.ID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to mark all notifications as read")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "All notifications marked as read",
	})
}

// Delete deletes a notification
// DELETE /api/v1/notifications/:id
func (h *NotificationsHandler) Delete(c echo.Context) error {
	notificationID, err := parseResourceID(c, "notification")
	if err != nil {
		return err
	}

	// Add audit context (user ID, IP address, user agent) for audit logging
	user := c.Get("current_user").(*models.User)
	ctx := audit.AddAuditContextToContext(c.Request().Context(), user.ID, c.RealIP(), c.Request().UserAgent())

	// Delete notification (scoped to authenticated user)
	if err := h.notificationService.DeleteNotification(ctx, user.ID, notificationID); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Notification not found")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Notification deleted",
	})
}
