// Package repository provides data access layer implementations.
package repository

import (
	"context"
	"savvy/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NotificationRepository defines the interface for notification data access
type NotificationRepository interface {
	Create(ctx context.Context, notification *models.Notification) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Notification, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Notification, error)
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error)
	MarkAsRead(ctx context.Context, notificationID uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	Delete(ctx context.Context, notificationID uuid.UUID) error
}

// notificationRepository implements NotificationRepository
type notificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository creates a new notification repository
func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

// Create creates a new notification
func (r *notificationRepository) Create(ctx context.Context, notification *models.Notification) error {
	return r.db.WithContext(ctx).Create(notification).Error
}

// GetByID retrieves a notification by ID
func (r *notificationRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Notification, error) {
	var notification models.Notification
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&notification).Error
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

// GetByUserID retrieves all notifications for a user with pagination
func (r *notificationRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Notification, error) {
	var notifications []models.Notification
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notifications).Error
	return notifications, err
}

// GetUnreadCount returns the number of unread notifications for a user
func (r *notificationRepository) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("user_id = ? AND is_read = FALSE", userID).
		Count(&count).Error
	return count, err
}

// MarkAsRead marks a notification as read
func (r *notificationRepository) MarkAsRead(ctx context.Context, notificationID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("id = ?", notificationID).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": gorm.Expr("CURRENT_TIMESTAMP"),
		}).Error
}

// MarkAllAsRead marks all unread notifications as read for a user
func (r *notificationRepository) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("user_id = ? AND is_read = FALSE", userID).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": gorm.Expr("CURRENT_TIMESTAMP"),
		}).Error
}

// Delete soft deletes a notification
func (r *notificationRepository) Delete(ctx context.Context, notificationID uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Notification{}, "id = ?", notificationID).Error
}
