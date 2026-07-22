// Package services contains business logic.
package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"savvy/internal/logsafe"
	"savvy/internal/models"
	"savvy/internal/repository"
	"savvy/internal/validation"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// VoucherServiceInterface defines the interface for voucher business logic.
type VoucherServiceInterface interface {
	CreateVoucher(ctx context.Context, voucher *models.Voucher) error
	GetVoucher(ctx context.Context, id uuid.UUID) (*models.Voucher, error)
	GetUserVouchers(ctx context.Context, userID uuid.UUID) ([]models.Voucher, error)
	GetUserVouchersPaginated(ctx context.Context, userID uuid.UUID, page, perPage int) (*repository.PaginatedResult[models.Voucher], error)
	UpdateVoucher(ctx context.Context, voucher *models.Voucher) error
	DeleteVoucher(ctx context.Context, id uuid.UUID) error
	CountUserVouchers(ctx context.Context, userID uuid.UUID) (int64, error)
	CheckDuplicate(ctx context.Context, voucherCode string, userID uuid.UUID, excludeID *uuid.UUID) (*models.Voucher, error)
	FindDeletedDuplicate(ctx context.Context, code string, userID uuid.UUID) (*models.Voucher, error)
	RestoreVoucher(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.Voucher, error)
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

	if validation.VoucherValueRequired(voucher.Type) {
		if voucher.Value <= 0 {
			return errors.New("voucher value must be positive")
		}
	} else {
		// Free vouchers are gratis; enforce the value-0 invariant server-side
		// so a client or import cannot persist a hidden non-zero value.
		voucher.Value = 0
	}

	if !voucher.ValidFrom.IsZero() && !voucher.ValidUntil.IsZero() {
		if voucher.ValidFrom.After(voucher.ValidUntil) {
			return errors.New("valid_from must be before valid_until")
		}
	}

	if err := s.repo.Create(ctx, voucher); err != nil {
		return fmt.Errorf("create voucher: %w", err)
	}

	slog.Info("Voucher created", "voucher_id", logsafe.UUID(voucher.ID), "merchant", logsafe.String(voucher.MerchantName))
	return nil
}

// GetVoucher retrieves a voucher by ID.
func (s *VoucherService) GetVoucher(ctx context.Context, id uuid.UUID) (*models.Voucher, error) {
	voucher, err := s.repo.GetByID(ctx, id, "Merchant", "User")
	if err != nil {
		return nil, fmt.Errorf("get voucher %s: %w", id, err)
	}
	return voucher, nil
}

// GetUserVouchers retrieves all vouchers for a user (owned + shared).
func (s *VoucherService) GetUserVouchers(ctx context.Context, userID uuid.UUID) ([]models.Voucher, error) {
	ownedVouchers, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get owned vouchers: %w", err)
	}

	sharedVouchers, err := s.repo.GetSharedWithUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get shared vouchers: %w", err)
	}

	return append(ownedVouchers, sharedVouchers...), nil
}

// GetUserVouchersPaginated retrieves paginated vouchers for a user (owned + shared).
func (s *VoucherService) GetUserVouchersPaginated(ctx context.Context, userID uuid.UUID, page, perPage int) (*repository.PaginatedResult[models.Voucher], error) {
	params := repository.PaginationParams{Page: page, PerPage: perPage}
	result, err := s.repo.GetAllForUserPaginated(ctx, userID, params)
	if err != nil {
		return nil, fmt.Errorf("get paginated vouchers: %w", err)
	}
	return result, nil
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

	if validation.VoucherValueRequired(voucher.Type) {
		if voucher.Value <= 0 {
			return errors.New("voucher value must be positive")
		}
	} else {
		// Free vouchers are gratis; enforce the value-0 invariant server-side
		// so a client or import cannot persist a hidden non-zero value.
		voucher.Value = 0
	}

	if !voucher.ValidFrom.IsZero() && !voucher.ValidUntil.IsZero() {
		if voucher.ValidFrom.After(voucher.ValidUntil) {
			return errors.New("valid_from must be before valid_until")
		}
	}

	if err := s.repo.Update(ctx, voucher); err != nil {
		return fmt.Errorf("update voucher %s: %w", voucher.ID, err)
	}

	slog.Info("Voucher updated", "voucher_id", voucher.ID)
	return nil
}

// DeleteVoucher deletes a voucher.
func (s *VoucherService) DeleteVoucher(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete voucher %s: %w", id, err)
	}

	slog.Info("Voucher deleted", "voucher_id", id)
	return nil
}

// CountUserVouchers counts vouchers for a user.
func (s *VoucherService) CountUserVouchers(ctx context.Context, userID uuid.UUID) (int64, error) {
	count, err := s.repo.Count(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("count user vouchers: %w", err)
	}
	return count, nil
}

// FindDeletedDuplicate returns a soft-deleted voucher with the same code owned by the user, or nil.
func (s *VoucherService) FindDeletedDuplicate(ctx context.Context, code string, userID uuid.UUID) (*models.Voucher, error) {
	return s.repo.FindDeletedByCode(ctx, code, userID)
}

// RestoreVoucher clears deleted_at for the user's soft-deleted voucher and returns the restored voucher.
// Returns (nil, nil) when there is no restorable twin for this user (id unknown or not owned by user);
// (restoredVoucher, nil) on success; (nil, err) on real DB error.
// Zero-row-restore guard: after RestoreByID, GetByID is called; if the record is still
// not found (nothing was actually undeleted), we return (nil, nil) instead of an error.
func (s *VoucherService) RestoreVoucher(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.Voucher, error) {
	if err := s.repo.RestoreByID(ctx, id, userID); err != nil {
		return nil, fmt.Errorf("restore voucher: %w", err)
	}
	restored, err := s.repo.GetByID(ctx, id, "Merchant", "User")
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Nothing was restored (wrong id or not owned by this user) — signal 404.
			return nil, nil
		}
		return nil, fmt.Errorf("load restored voucher: %w", err)
	}
	// Guard against cross-user reads: RestoreByID is user-scoped and no-ops on a
	// foreign id, but GetByID fetches by id only. Without this check a user could
	// read another user's active voucher via the restore endpoint. Signal 404 instead.
	if restored.UserID == nil || *restored.UserID != userID {
		return nil, nil
	}
	return restored, nil
}

// CheckDuplicate checks if a voucher with the same code already exists for the user.
// Returns the existing voucher if found, nil otherwise.
// excludeID is used during updates to exclude the voucher being updated.
func (s *VoucherService) CheckDuplicate(ctx context.Context, voucherCode string, userID uuid.UUID, excludeID *uuid.UUID) (*models.Voucher, error) {
	existing, err := s.repo.FindByVoucherCode(ctx, voucherCode, userID)
	if err != nil {
		return nil, fmt.Errorf("check duplicate: %w", err)
	}

	// No duplicate found
	if existing == nil {
		return nil, nil
	}

	// If excludeID is provided (update case), check if it's the same voucher
	if excludeID != nil && existing.ID == *excludeID {
		return nil, nil // Same voucher, not a duplicate
	}

	// Duplicate found
	return existing, nil
}
