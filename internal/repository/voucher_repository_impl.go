// Package repository defines data access interfaces.
//
//nolint:dupl // Wrapper methods required for interface compliance with Go generics
package repository

import (
	"context"
	"savvy/internal/models"
	"time"

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

// Create creates a new voucher in the database.
func (r *GormVoucherRepository) Create(ctx context.Context, voucher *models.Voucher) error {
	return r.BaseRepository.Create(ctx, voucher)
}

// GetByID retrieves a voucher by its ID with optional preloads.
func (r *GormVoucherRepository) GetByID(ctx context.Context, id uuid.UUID, preloads ...string) (*models.Voucher, error) {
	return r.BaseRepository.GetByID(ctx, id, preloads...)
}

// GetByUserID retrieves all vouchers owned by a specific user.
func (r *GormVoucherRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.Voucher, error) {
	return r.BaseRepository.GetByUserID(ctx, userID)
}

// GetSharedWithUser retrieves all vouchers shared with a specific user.
func (r *GormVoucherRepository) GetSharedWithUser(ctx context.Context, userID uuid.UUID) ([]models.Voucher, error) {
	return r.BaseRepository.GetSharedWithUser(ctx, userID)
}

// Update updates an existing voucher in the database.
func (r *GormVoucherRepository) Update(ctx context.Context, voucher *models.Voucher) error {
	return r.BaseRepository.Update(ctx, voucher)
}

// Delete soft-deletes a voucher from the database.
func (r *GormVoucherRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.BaseRepository.Delete(ctx, id)
}

// Count returns the total number of vouchers owned by a user.
func (r *GormVoucherRepository) Count(ctx context.Context, userID uuid.UUID) (int64, error) {
	return r.BaseRepository.Count(ctx, userID)
}

// GetAllForUserPaginated retrieves all vouchers (owned + shared) with pagination.
func (r *GormVoucherRepository) GetAllForUserPaginated(ctx context.Context, userID uuid.UUID, params PaginationParams) (*PaginatedResult[models.Voucher], error) {
	return r.BaseRepository.GetAllForUserPaginated(ctx, userID, params)
}

// GetExpiringVouchers retrieves vouchers expiring within the given number of days.
func (r *GormVoucherRepository) GetExpiringVouchers(ctx context.Context, withinDays int) ([]models.Voucher, error) {
	// Use date-based comparison in UTC since dates are stored as end-of-day UTC
	// (T23:59:59Z). Timestamp-based comparison with NOW() caused timezone issues
	// where converting to local timezone shifted the date forward by one day.
	now := time.Now().UTC()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	windowEnd := startOfToday.AddDate(0, 0, withinDays+1)

	var vouchers []models.Voucher
	err := r.BaseRepository.db.WithContext(ctx).
		Preload("Merchant").
		Preload("User").
		Where("user_id IS NOT NULL").
		Where("valid_until >= ?", startOfToday).
		Where("valid_until < ?", windowEnd).
		Find(&vouchers).Error
	return vouchers, err
}

// GetVouchersStartingTomorrow retrieves vouchers whose valid_from date is tomorrow.
func (r *GormVoucherRepository) GetVouchersStartingTomorrow(ctx context.Context) ([]models.Voucher, error) {
	now := time.Now().UTC()
	startOfTomorrow := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, 1)
	startOfDayAfter := startOfTomorrow.AddDate(0, 0, 1)

	var vouchers []models.Voucher
	err := r.BaseRepository.db.WithContext(ctx).
		Preload("Merchant").
		Preload("User").
		Where("user_id IS NOT NULL").
		Where("valid_from >= ?", startOfTomorrow).
		Where("valid_from < ?", startOfDayAfter).
		Find(&vouchers).Error
	return vouchers, err
}

// FindByVoucherCode finds a voucher by code for a specific user.
func (r *GormVoucherRepository) FindByVoucherCode(ctx context.Context, voucherCode string, userID uuid.UUID) (*models.Voucher, error) {
	var voucher models.Voucher
	err := r.db.WithContext(ctx).
		Where("code = ? AND user_id = ?", voucherCode, userID).
		First(&voucher).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // No duplicate found
		}
		return nil, err
	}

	return &voucher, nil
}

// Search searches vouchers by query (merchant name, code, description).
