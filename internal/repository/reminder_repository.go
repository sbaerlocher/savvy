// Package repository defines data access interfaces.
package repository

import (
	"context"
	"savvy/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ReminderRepository defines the interface for expiry reminder tracking.
type ReminderRepository interface {
	// HasBeenSent checks if a reminder has already been sent for a resource
	HasBeenSent(ctx context.Context, userID uuid.UUID, resourceType string, resourceID uuid.UUID, daysBefore int) (bool, error)

	// MarkSent records that a reminder has been sent
	MarkSent(ctx context.Context, reminder *models.ExpiryReminderSent) error
}

// reminderRepository implements ReminderRepository using GORM.
type reminderRepository struct {
	db *gorm.DB
}

// NewReminderRepository creates a new reminder repository.
func NewReminderRepository(db *gorm.DB) ReminderRepository {
	return &reminderRepository{db: db}
}

// HasBeenSent checks if a reminder has already been sent.
func (r *reminderRepository) HasBeenSent(ctx context.Context, userID uuid.UUID, resourceType string, resourceID uuid.UUID, daysBefore int) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.ExpiryReminderSent{}).
		Where("user_id = ? AND resource_type = ? AND resource_id = ? AND days_before = ?",
			userID, resourceType, resourceID, daysBefore).
		Count(&count).Error
	return count > 0, err
}

// MarkSent records that a reminder has been sent.
func (r *reminderRepository) MarkSent(ctx context.Context, reminder *models.ExpiryReminderSent) error {
	return r.db.WithContext(ctx).Create(reminder).Error
}
