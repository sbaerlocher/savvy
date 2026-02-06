// Package services contains business logic.
package services

import (
	"context"
	"savvy/internal/models"
	"savvy/internal/repository"

	"github.com/google/uuid"
)

// NotificationServiceInterface defines the interface for notification business logic
type NotificationServiceInterface interface {
	CreateShareNotification(ctx context.Context, recipientID, fromUserID uuid.UUID, fromUserName, resourceType string, resourceID uuid.UUID, permissions map[string]bool) error
	CreateTransferNotification(ctx context.Context, recipientID, fromUserID uuid.UUID, fromUserName, resourceType string, resourceID uuid.UUID) error
	GetUserNotifications(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Notification, error)
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error)
	MarkAsRead(ctx context.Context, notificationID uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	DeleteNotification(ctx context.Context, notificationID uuid.UUID) error
}

// NotificationService implements NotificationServiceInterface
type NotificationService struct {
	repo repository.NotificationRepository
}

// NewNotificationService creates a new notification service
func NewNotificationService(repo repository.NotificationRepository) NotificationServiceInterface {
	return &NotificationService{repo: repo}
}

// CreateShareNotification creates a notification when a resource is shared with a user
func (s *NotificationService) CreateShareNotification(
	ctx context.Context,
	recipientID, fromUserID uuid.UUID,
	fromUserName, resourceType string,
	resourceID uuid.UUID,
	permissions map[string]bool,
) error {
	// Build metadata
	metadata := models.NotificationMetadata{
		"from_user_id":   fromUserID.String(),
		"from_user_name": fromUserName,
	}

	// Add permissions if provided
	if permissions != nil {
		metadata["permissions"] = permissions
	}

	notification := &models.Notification{
		UserID:       recipientID,
		Type:         models.NotificationTypeShareReceived,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Metadata:     metadata,
		IsRead:       false,
	}

	return s.repo.Create(ctx, notification)
}

// CreateTransferNotification creates a notification when resource ownership is transferred to a user
func (s *NotificationService) CreateTransferNotification(
	ctx context.Context,
	recipientID, fromUserID uuid.UUID,
	fromUserName, resourceType string,
	resourceID uuid.UUID,
) error {
	notification := &models.Notification{
		UserID:       recipientID,
		Type:         models.NotificationTypeTransferReceived,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Metadata: models.NotificationMetadata{
			"from_user_id":   fromUserID.String(),
			"from_user_name": fromUserName,
		},
		IsRead: false,
	}

	return s.repo.Create(ctx, notification)
}

// GetUserNotifications retrieves all notifications for a user with pagination
func (s *NotificationService) GetUserNotifications(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Notification, error) {
	return s.repo.GetByUserID(ctx, userID, limit, offset)
}

// GetUnreadCount returns the number of unread notifications for a user
func (s *NotificationService) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.repo.GetUnreadCount(ctx, userID)
}

// MarkAsRead marks a notification as read
func (s *NotificationService) MarkAsRead(ctx context.Context, notificationID uuid.UUID) error {
	return s.repo.MarkAsRead(ctx, notificationID)
}

// MarkAllAsRead marks all unread notifications as read for a user
func (s *NotificationService) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	return s.repo.MarkAllAsRead(ctx, userID)
}

// DeleteNotification deletes a notification
func (s *NotificationService) DeleteNotification(ctx context.Context, notificationID uuid.UUID) error {
	return s.repo.Delete(ctx, notificationID)
}
