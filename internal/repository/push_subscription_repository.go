// Package repository provides data access layer implementations.
package repository

import (
	"context"
	"savvy/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PushSubscriptionRepository defines the interface for push subscription data access.
type PushSubscriptionRepository interface {
	Create(ctx context.Context, sub *models.PushSubscription) error
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.PushSubscription, error)
	DeleteByEndpoint(ctx context.Context, endpoint string) error
}

// pushSubscriptionRepository implements PushSubscriptionRepository.
type pushSubscriptionRepository struct {
	db *gorm.DB
}

// NewPushSubscriptionRepository creates a new push subscription repository.
func NewPushSubscriptionRepository(db *gorm.DB) PushSubscriptionRepository {
	return &pushSubscriptionRepository{db: db}
}

// Create creates or updates a push subscription (upsert by endpoint).
func (r *pushSubscriptionRepository) Create(ctx context.Context, sub *models.PushSubscription) error {
	// Upsert: if endpoint already exists, update keys and user
	return r.db.WithContext(ctx).
		Where("endpoint = ?", sub.Endpoint).
		Assign(models.PushSubscription{
			UserID:    sub.UserID,
			P256dhKey: sub.P256dhKey,
			AuthKey:   sub.AuthKey,
			UserAgent: sub.UserAgent,
		}).
		FirstOrCreate(sub).Error
}

// GetByUserID retrieves all push subscriptions for a user.
func (r *pushSubscriptionRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.PushSubscription, error) {
	var subs []models.PushSubscription
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&subs).Error
	return subs, err
}

// DeleteByEndpoint removes a push subscription by endpoint.
func (r *pushSubscriptionRepository) DeleteByEndpoint(ctx context.Context, endpoint string) error {
	return r.db.WithContext(ctx).Where("endpoint = ?", endpoint).Delete(&models.PushSubscription{}).Error
}
