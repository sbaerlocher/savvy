// Package repository defines data access interfaces.
package repository

import (
	"context"
	"savvy/internal/models"

	"github.com/google/uuid"
)

// VoucherRepository defines the interface for voucher data access.
type VoucherRepository interface {
	// Create creates a new voucher
	Create(ctx context.Context, voucher *models.Voucher) error

	// GetByID retrieves a voucher by ID with optional preloads
	GetByID(ctx context.Context, id uuid.UUID, preloads ...string) (*models.Voucher, error)

	// GetByUserID retrieves all vouchers for a user
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.Voucher, error)

	// GetSharedWithUser retrieves vouchers shared with a user
	GetSharedWithUser(ctx context.Context, userID uuid.UUID) ([]models.Voucher, error)

	// Update updates a voucher
	Update(ctx context.Context, voucher *models.Voucher) error

	// Delete soft-deletes a voucher
	Delete(ctx context.Context, id uuid.UUID) error

	// Count counts vouchers for a user
	Count(ctx context.Context, userID uuid.UUID) (int64, error)
}
