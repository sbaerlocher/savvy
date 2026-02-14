// Package repository defines data access interfaces.
package repository

import (
	"context"
	"savvy/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GormCardShareRepository implements CardShareRepository using GORM.
type GormCardShareRepository struct {
	db *gorm.DB
}

// NewCardShareRepository creates a new card share repository.
func NewCardShareRepository(db *gorm.DB) CardShareRepository {
	return &GormCardShareRepository{db: db}
}

// Create persists a new card share.
func (r *GormCardShareRepository) Create(ctx context.Context, share *models.CardShare) error {
	return r.db.WithContext(ctx).Create(share).Error
}

// GetByCardAndUser returns the share record for a specific card and user.
func (r *GormCardShareRepository) GetByCardAndUser(ctx context.Context, cardID, userID uuid.UUID) (*models.CardShare, error) {
	var share models.CardShare
	err := r.db.WithContext(ctx).
		Where("card_id = ? AND shared_with_id = ?", cardID, userID).
		First(&share).Error
	if err != nil {
		return nil, err
	}
	return &share, nil
}

// GetByCardID returns all shares for a given card.
func (r *GormCardShareRepository) GetByCardID(ctx context.Context, cardID uuid.UUID) ([]models.CardShare, error) {
	var shares []models.CardShare
	err := r.db.WithContext(ctx).
		Where("card_id = ?", cardID).
		Where("deleted_at IS NULL").
		Preload("SharedWithUser").
		Find(&shares).Error
	return shares, err
}

// Update saves changes to an existing card share.
func (r *GormCardShareRepository) Update(ctx context.Context, share *models.CardShare) error {
	return r.db.WithContext(ctx).Save(share).Error
}

// DeleteByCardAndUser removes a share for a specific card and user.
func (r *GormCardShareRepository) DeleteByCardAndUser(ctx context.Context, cardID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("card_id = ? AND shared_with_id = ?", cardID, userID).
		Delete(&models.CardShare{}).Error
}

// DeleteByCardID removes all shares for a given card.
func (r *GormCardShareRepository) DeleteByCardID(ctx context.Context, cardID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("card_id = ?", cardID).
		Delete(&models.CardShare{}).Error
}

// CountByUser returns the number of cards shared with a user.
func (r *GormCardShareRepository) CountByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.CardShare{}).
		Where("shared_with_id = ?", userID).
		Count(&count).Error
	return count, err
}

// CountByCardIDs returns the number of shares per card for the given card IDs.
func (r *GormCardShareRepository) CountByCardIDs(ctx context.Context, cardIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
	if len(cardIDs) == 0 {
		return make(map[uuid.UUID]int64), nil
	}

	type result struct {
		CardID uuid.UUID `gorm:"column:card_id"`
		Count  int64     `gorm:"column:count"`
	}
	var results []result

	err := r.db.WithContext(ctx).
		Model(&models.CardShare{}).
		Select("card_id, COUNT(*) as count").
		Where("card_id IN ? AND deleted_at IS NULL", cardIDs).
		Group("card_id").
		Find(&results).Error
	if err != nil {
		return nil, err
	}

	counts := make(map[uuid.UUID]int64, len(results))
	for _, r := range results {
		counts[r.CardID] = r.Count
	}
	return counts, nil
}

// GetSharedUserIDs returns distinct user IDs that have cards shared by the owner.
func (r *GormCardShareRepository) GetSharedUserIDs(ctx context.Context, ownerID uuid.UUID) ([]uuid.UUID, error) {
	var userIDs []uuid.UUID
	err := r.db.WithContext(ctx).
		Table("card_shares").
		Select("DISTINCT shared_with_id").
		Joins("JOIN cards ON cards.id = card_shares.card_id").
		Where("cards.user_id = ? AND card_shares.deleted_at IS NULL AND cards.deleted_at IS NULL", ownerID).
		Pluck("shared_with_id", &userIDs).Error
	return userIDs, err
}
