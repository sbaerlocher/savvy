// Package repository defines data access interfaces.
//
//nolint:dupl // Wrapper methods required for interface compliance with Go generics
package repository

import (
	"context"
	"savvy/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GormVoucherRepository implements VoucherRepository using GORM.
type GormVoucherRepository struct {
	*BaseRepository[models.Voucher]
}

// NewVoucherRepository creates a new voucher repository.
func NewVoucherRepository(db *gorm.DB) VoucherRepository {
	return &GormVoucherRepository{
		BaseRepository: NewBaseRepository[models.Voucher](db, &ShareConfig{
			ShareTableName:   "voucher_shares",
			ResourceIDColumn: "voucher_id",
			TableName:        "vouchers",
		}),
	}
}

// Create creates a new voucher.
func (r *GormVoucherRepository) Create(ctx context.Context, voucher *models.Voucher) error {
	return r.BaseRepository.Create(ctx, voucher)
}

// GetByID retrieves a voucher by ID with optional preloads.
func (r *GormVoucherRepository) GetByID(ctx context.Context, id uuid.UUID, preloads ...string) (*models.Voucher, error) {
	return r.BaseRepository.GetByID(ctx, id, preloads...)
}

// GetByUserID retrieves all vouchers for a user.
func (r *GormVoucherRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.Voucher, error) {
	return r.BaseRepository.GetByUserID(ctx, userID)
}

// GetSharedWithUser retrieves vouchers shared with a user.
func (r *GormVoucherRepository) GetSharedWithUser(ctx context.Context, userID uuid.UUID) ([]models.Voucher, error) {
	return r.BaseRepository.GetSharedWithUser(ctx, userID)
}

// Update updates a voucher.
func (r *GormVoucherRepository) Update(ctx context.Context, voucher *models.Voucher) error {
	return r.BaseRepository.Update(ctx, voucher)
}

// Delete soft-deletes a voucher.
func (r *GormVoucherRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.BaseRepository.Delete(ctx, id)
}

// Count counts vouchers for a user.
func (r *GormVoucherRepository) Count(ctx context.Context, userID uuid.UUID) (int64, error) {
	return r.BaseRepository.Count(ctx, userID)
}
