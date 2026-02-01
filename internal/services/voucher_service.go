// Package services contains business logic.
package services

import (
	"context"
	"errors"
	"savvy/internal/models"
	"savvy/internal/repository"

	"github.com/google/uuid"
)

// VoucherServiceInterface defines the interface for voucher business logic.
type VoucherServiceInterface interface {
	CreateVoucher(ctx context.Context, voucher *models.Voucher) error
	GetVoucher(ctx context.Context, id uuid.UUID) (*models.Voucher, error)
	GetUserVouchers(ctx context.Context, userID uuid.UUID) ([]models.Voucher, error)
	UpdateVoucher(ctx context.Context, voucher *models.Voucher) error
	DeleteVoucher(ctx context.Context, id uuid.UUID) error
	CountUserVouchers(ctx context.Context, userID uuid.UUID) (int64, error)
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
	ownedVouchers, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	sharedVouchers, err := s.repo.GetSharedWithUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	return append(ownedVouchers, sharedVouchers...), nil
}

// UpdateVoucher updates a voucher.
func (s *VoucherService) UpdateVoucher(ctx context.Context, voucher *models.Voucher) error {
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
