// Package repository defines data access interfaces.
package repository

import (
	"context"
	"savvy/internal/models"

	"github.com/google/uuid"
)

// VoucherShareRepository defines data access operations for voucher shares.
type VoucherShareRepository interface {
	Create(ctx context.Context, share *models.VoucherShare) error
	GetByVoucherAndUser(ctx context.Context, voucherID, userID uuid.UUID) (*models.VoucherShare, error)
	GetByVoucherID(ctx context.Context, voucherID uuid.UUID) ([]models.VoucherShare, error)
	Update(ctx context.Context, share *models.VoucherShare) error
	DeleteByVoucherAndUser(ctx context.Context, voucherID, userID uuid.UUID) error
	DeleteByVoucherID(ctx context.Context, voucherID uuid.UUID) error
	CountByUser(ctx context.Context, userID uuid.UUID) (int64, error)
	CountByVoucherIDs(ctx context.Context, voucherIDs []uuid.UUID) (map[uuid.UUID]int64, error)
	GetSharedUserIDs(ctx context.Context, ownerID uuid.UUID) ([]uuid.UUID, error)
}
