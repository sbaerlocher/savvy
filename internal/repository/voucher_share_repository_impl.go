// Package repository defines data access interfaces.
package repository

import (
	"context"
	"savvy/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GormVoucherShareRepository implements VoucherShareRepository using GORM.
type GormVoucherShareRepository struct {
	db *gorm.DB
}

// NewVoucherShareRepository creates a new voucher share repository.
func NewVoucherShareRepository(db *gorm.DB) VoucherShareRepository {
	return &GormVoucherShareRepository{db: db}
}

// Create persists a new voucher share.
func (r *GormVoucherShareRepository) Create(ctx context.Context, share *models.VoucherShare) error {
	return r.db.WithContext(ctx).Create(share).Error
}

// GetByVoucherAndUser returns the share record for a specific voucher and user.
func (r *GormVoucherShareRepository) GetByVoucherAndUser(ctx context.Context, voucherID, userID uuid.UUID) (*models.VoucherShare, error) {
	var share models.VoucherShare
	err := r.db.WithContext(ctx).
		Where("voucher_id = ? AND shared_with_id = ?", voucherID, userID).
		First(&share).Error
	if err != nil {
		return nil, err
	}
	return &share, nil
}

// GetByVoucherID returns all shares for a given voucher.
func (r *GormVoucherShareRepository) GetByVoucherID(ctx context.Context, voucherID uuid.UUID) ([]models.VoucherShare, error) {
	var shares []models.VoucherShare
	err := r.db.WithContext(ctx).
		Where("voucher_id = ?", voucherID).
		Where("deleted_at IS NULL").
		Preload("SharedWithUser").
		Find(&shares).Error
	return shares, err
}

// Update saves changes to an existing voucher share.
func (r *GormVoucherShareRepository) Update(ctx context.Context, share *models.VoucherShare) error {
	return r.db.WithContext(ctx).Save(share).Error
}

// DeleteByVoucherAndUser removes a share for a specific voucher and user.
func (r *GormVoucherShareRepository) DeleteByVoucherAndUser(ctx context.Context, voucherID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("voucher_id = ? AND shared_with_id = ?", voucherID, userID).
		Delete(&models.VoucherShare{}).Error
}

// DeleteByVoucherID removes all shares for a given voucher.
func (r *GormVoucherShareRepository) DeleteByVoucherID(ctx context.Context, voucherID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("voucher_id = ?", voucherID).
		Delete(&models.VoucherShare{}).Error
}

// CountByUser returns the number of vouchers shared with a user.
func (r *GormVoucherShareRepository) CountByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.VoucherShare{}).
		Where("shared_with_id = ?", userID).
		Count(&count).Error
	return count, err
}

// CountByVoucherIDs returns the number of shares per voucher for the given voucher IDs.
func (r *GormVoucherShareRepository) CountByVoucherIDs(ctx context.Context, voucherIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
	if len(voucherIDs) == 0 {
		return make(map[uuid.UUID]int64), nil
	}

	type result struct {
		VoucherID uuid.UUID `gorm:"column:voucher_id"`
		Count     int64     `gorm:"column:count"`
	}
	var results []result

	err := r.db.WithContext(ctx).
		Model(&models.VoucherShare{}).
		Select("voucher_id, COUNT(*) as count").
		Where("voucher_id IN ? AND deleted_at IS NULL", voucherIDs).
		Group("voucher_id").
		Find(&results).Error
	if err != nil {
		return nil, err
	}

	counts := make(map[uuid.UUID]int64, len(results))
	for _, r := range results {
		counts[r.VoucherID] = r.Count
	}
	return counts, nil
}

// GetSharedUserIDs returns distinct user IDs that have vouchers shared by the owner.
func (r *GormVoucherShareRepository) GetSharedUserIDs(ctx context.Context, ownerID uuid.UUID) ([]uuid.UUID, error) {
	var userIDs []uuid.UUID
	err := r.db.WithContext(ctx).
		Table("voucher_shares").
		Select("DISTINCT shared_with_id").
		Joins("JOIN vouchers ON vouchers.id = voucher_shares.voucher_id").
		Where("vouchers.user_id = ? AND voucher_shares.deleted_at IS NULL AND vouchers.deleted_at IS NULL", ownerID).
		Pluck("shared_with_id", &userIDs).Error
	return userIDs, err
}
