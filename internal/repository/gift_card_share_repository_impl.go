// Package repository defines data access interfaces.
package repository

import (
	"context"
	"savvy/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GormGiftCardShareRepository implements GiftCardShareRepository using GORM.
type GormGiftCardShareRepository struct {
	db *gorm.DB
}

// NewGiftCardShareRepository creates a new gift card share repository.
func NewGiftCardShareRepository(db *gorm.DB) GiftCardShareRepository {
	return &GormGiftCardShareRepository{db: db}
}

// Create persists a new gift card share.
func (r *GormGiftCardShareRepository) Create(ctx context.Context, share *models.GiftCardShare) error {
	return r.db.WithContext(ctx).Create(share).Error
}

// GetByGiftCardAndUser returns the share record for a specific gift card and user.
func (r *GormGiftCardShareRepository) GetByGiftCardAndUser(ctx context.Context, giftCardID, userID uuid.UUID) (*models.GiftCardShare, error) {
	var share models.GiftCardShare
	err := r.db.WithContext(ctx).
		Where("gift_card_id = ? AND shared_with_id = ?", giftCardID, userID).
		First(&share).Error
	if err != nil {
		return nil, err
	}
	return &share, nil
}

// GetByGiftCardID returns all shares for a given gift card.
func (r *GormGiftCardShareRepository) GetByGiftCardID(ctx context.Context, giftCardID uuid.UUID) ([]models.GiftCardShare, error) {
	var shares []models.GiftCardShare
	err := r.db.WithContext(ctx).
		Where("gift_card_id = ?", giftCardID).
		Where("deleted_at IS NULL").
		Preload("SharedWithUser").
		Find(&shares).Error
	return shares, err
}

// Update saves changes to an existing gift card share.
func (r *GormGiftCardShareRepository) Update(ctx context.Context, share *models.GiftCardShare) error {
	return r.db.WithContext(ctx).Save(share).Error
}

// DeleteByGiftCardAndUser removes a share for a specific gift card and user.
func (r *GormGiftCardShareRepository) DeleteByGiftCardAndUser(ctx context.Context, giftCardID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("gift_card_id = ? AND shared_with_id = ?", giftCardID, userID).
		Delete(&models.GiftCardShare{}).Error
}

// DeleteByGiftCardID removes all shares for a given gift card.
func (r *GormGiftCardShareRepository) DeleteByGiftCardID(ctx context.Context, giftCardID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("gift_card_id = ?", giftCardID).
		Delete(&models.GiftCardShare{}).Error
}

// CountByUser returns the number of gift cards shared with a user.
func (r *GormGiftCardShareRepository) CountByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.GiftCardShare{}).
		Where("shared_with_id = ?", userID).
		Count(&count).Error
	return count, err
}

// CountByGiftCardIDs returns the number of shares per gift card for the given gift card IDs.
func (r *GormGiftCardShareRepository) CountByGiftCardIDs(ctx context.Context, giftCardIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
	if len(giftCardIDs) == 0 {
		return make(map[uuid.UUID]int64), nil
	}

	type result struct {
		GiftCardID uuid.UUID `gorm:"column:gift_card_id"`
		Count      int64     `gorm:"column:count"`
	}
	var results []result

	err := r.db.WithContext(ctx).
		Model(&models.GiftCardShare{}).
		Select("gift_card_id, COUNT(*) as count").
		Where("gift_card_id IN ? AND deleted_at IS NULL", giftCardIDs).
		Group("gift_card_id").
		Find(&results).Error
	if err != nil {
		return nil, err
	}

	counts := make(map[uuid.UUID]int64, len(results))
	for _, r := range results {
		counts[r.GiftCardID] = r.Count
	}
	return counts, nil
}

// GetSharedUserIDs returns distinct user IDs that have gift cards shared by the owner.
func (r *GormGiftCardShareRepository) GetSharedUserIDs(ctx context.Context, ownerID uuid.UUID) ([]uuid.UUID, error) {
	var userIDs []uuid.UUID
	err := r.db.WithContext(ctx).
		Table("gift_card_shares").
		Select("DISTINCT shared_with_id").
		Joins("JOIN gift_cards ON gift_cards.id = gift_card_shares.gift_card_id").
		Where("gift_cards.user_id = ? AND gift_card_shares.deleted_at IS NULL AND gift_cards.deleted_at IS NULL", ownerID).
		Pluck("shared_with_id", &userIDs).Error
	return userIDs, err
}
