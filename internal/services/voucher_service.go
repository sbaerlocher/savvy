// Package services contains business logic.
package services

import (
	"context"
	"errors"
	"savvy/internal/models"
	"savvy/internal/repository"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// VoucherServiceInterface defines the interface for voucher business logic.
type VoucherServiceInterface interface {
	CreateVoucher(ctx context.Context, voucher *models.Voucher) error
	GetVoucher(ctx context.Context, id uuid.UUID) (*models.Voucher, error)
	GetUserVouchers(ctx context.Context, userID uuid.UUID) ([]models.Voucher, error)
	UpdateVoucher(ctx context.Context, voucher *models.Voucher) error
	DeleteVoucher(ctx context.Context, id uuid.UUID) error
	CountUserVouchers(ctx context.Context, userID uuid.UUID) (int64, error)
	CanRedeemVoucher(ctx context.Context, voucherID uuid.UUID) (bool, error)
}

// VoucherService implements VoucherServiceInterface.
type VoucherService struct {
	repo repository.VoucherRepository
}

// NewVoucherService creates a new voucher service.
func NewVoucherService(repo repository.VoucherRepository) VoucherServiceInterface {
	return &VoucherService{repo: repo}
}

// CreateVoucher creates a new voucher.
func (s *VoucherService) CreateVoucher(ctx context.Context, voucher *models.Voucher) error {
	// Business logic: validate voucher
	if voucher.MerchantName == "" {
		return errors.New("merchant name is required")
	}

	if voucher.Code == "" {
		return errors.New("voucher code is required")
	}

	if voucher.Type == "" {
		return errors.New("voucher type is required")
	}

	if voucher.Value <= 0 {
		return errors.New("voucher value must be positive")
	}

	// Validate dates
	if !voucher.ValidFrom.IsZero() && !voucher.ValidUntil.IsZero() {
		if voucher.ValidFrom.After(voucher.ValidUntil) {
			return errors.New("valid_from must be before valid_until")
		}
	}

	return s.repo.Create(ctx, voucher)
}

// GetVoucher retrieves a voucher by ID.
func (s *VoucherService) GetVoucher(ctx context.Context, id uuid.UUID) (*models.Voucher, error) {
	return s.repo.GetByID(ctx, id, "Merchant", "User")
}

// GetUserVouchers retrieves all vouchers for a user (owned + shared).
func (s *VoucherService) GetUserVouchers(ctx context.Context, userID uuid.UUID) ([]models.Voucher, error) {
	// Get owned vouchers
	ownedVouchers, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get shared vouchers
	sharedVouchers, err := s.repo.GetSharedWithUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Combine and return
	return append(ownedVouchers, sharedVouchers...), nil
}

// UpdateVoucher updates a voucher.
func (s *VoucherService) UpdateVoucher(ctx context.Context, voucher *models.Voucher) error {
	// Business logic: validate voucher
	if voucher.MerchantName == "" {
		return errors.New("merchant name is required")
	}

	if voucher.Code == "" {
		return errors.New("voucher code is required")
	}

	if voucher.Type == "" {
		return errors.New("voucher type is required")
	}

	if voucher.Value <= 0 {
		return errors.New("voucher value must be positive")
	}

	// Validate dates
	if !voucher.ValidFrom.IsZero() && !voucher.ValidUntil.IsZero() {
		if voucher.ValidFrom.After(voucher.ValidUntil) {
			return errors.New("valid_from must be before valid_until")
		}
	}

	return s.repo.Update(ctx, voucher)
}

// DeleteVoucher deletes a voucher.
func (s *VoucherService) DeleteVoucher(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// CountUserVouchers counts vouchers for a user.
func (s *VoucherService) CountUserVouchers(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.repo.Count(ctx, userID)
}

// CanRedeemVoucher checks if a voucher can be redeemed.
func (s *VoucherService) CanRedeemVoucher(ctx context.Context, voucherID uuid.UUID) (bool, error) {
	voucher, err := s.GetVoucher(ctx, voucherID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	// Check validity dates
	now := time.Now()
	if !voucher.ValidFrom.IsZero() && now.Before(voucher.ValidFrom) {
		return false, nil
	}

	if !voucher.ValidUntil.IsZero() && now.After(voucher.ValidUntil) {
		return false, nil
	}

	// Check usage limits
	return voucher.CanRedeem(), nil
}
