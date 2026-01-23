// Package repository defines data access interfaces.
package repository

import (
	"context"
	"savvy/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GormVoucherRepository implements VoucherRepository using GORM.
type GormVoucherRepository struct {
	db *gorm.DB
}

// NewVoucherRepository creates a new voucher repository.
func NewVoucherRepository(db *gorm.DB) VoucherRepository {
	return &GormVoucherRepository{db: db}
}

// Create creates a new voucher.
func (r *GormVoucherRepository) Create(ctx context.Context, voucher *models.Voucher) error {
	return r.db.WithContext(ctx).Create(voucher).Error
}

// GetByID retrieves a voucher by ID with optional preloads.
func (r *GormVoucherRepository) GetByID(ctx context.Context, id uuid.UUID, preloads ...string) (*models.Voucher, error) {
	var voucher models.Voucher
	query := r.db.WithContext(ctx)

	for _, preload := range preloads {
		query = query.Preload(preload)
	}

	if err := query.First(&voucher, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return &voucher, nil
}

// GetByUserID retrieves all vouchers for a user.
func (r *GormVoucherRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.Voucher, error) {
	var vouchers []models.Voucher
	err := r.db.WithContext(ctx).
		Preload("Merchant").
		Preload("User").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&vouchers).Error

	return vouchers, err
}

// GetSharedWithUser retrieves vouchers shared with a user.
func (r *GormVoucherRepository) GetSharedWithUser(ctx context.Context, userID uuid.UUID) ([]models.Voucher, error) {
	var vouchers []models.Voucher
	err := r.db.WithContext(ctx).
		Preload("Merchant").
		Preload("User").
		Joins("INNER JOIN voucher_shares ON voucher_shares.voucher_id = vouchers.id").
		Where("voucher_shares.shared_with_id = ?", userID).
		Order("vouchers.created_at DESC").
		Find(&vouchers).Error

	return vouchers, err
}

// Update updates a voucher.
func (r *GormVoucherRepository) Update(ctx context.Context, voucher *models.Voucher) error {
	return r.db.WithContext(ctx).Save(voucher).Error
}

// Delete soft-deletes a voucher.
func (r *GormVoucherRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Voucher{}, "id = ?", id).Error
}

// Count counts vouchers for a user.
func (r *GormVoucherRepository) Count(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Voucher{}).
		Where("user_id = ?", userID).
		Count(&count).Error

	return count, err
}
