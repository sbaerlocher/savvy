// Package repository provides data access layer implementations.
package repository

import (
	"context"
	"savvy/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NotificationRepository defines the interface for notification data access
type NotificationRepository interface {
	Create(ctx context.Context, notification *models.Notification) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Notification, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Notification, error)
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error)
	MarkAsRead(ctx context.Context, userID, notificationID uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	Delete(ctx context.Context, userID, notificationID uuid.UUID) error
	// ArchiveOldRead archives read notifications older than cutoff and returns the count.
	ArchiveOldRead(ctx context.Context, cutoff time.Time) (int64, error)
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
		Where("user_id = ? AND archived_at IS NULL", userID).
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
		Where("user_id = ? AND is_read = FALSE AND archived_at IS NULL", userID).
		Count(&count).Error
	return count, err
}

// MarkAsRead marks a notification as read (scoped to user for ownership check)
func (r *notificationRepository) MarkAsRead(ctx context.Context, userID, notificationID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": gorm.Expr("CURRENT_TIMESTAMP"),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
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

// Delete soft deletes a notification (scoped to user for ownership check)
func (r *notificationRepository) Delete(ctx context.Context, userID, notificationID uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", notificationID, userID).Delete(&models.Notification{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ArchiveOldRead archives read notifications created before cutoff by stamping
// archived_at. Archived rows drop out of the main list but stay in the table.
func (r *notificationRepository) ArchiveOldRead(ctx context.Context, cutoff time.Time) (int64, error) {
	result := r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("is_read = TRUE AND archived_at IS NULL AND created_at < ?", cutoff).
		Update("archived_at", gorm.Expr("CURRENT_TIMESTAMP"))
	return result.RowsAffected, result.Error
}
