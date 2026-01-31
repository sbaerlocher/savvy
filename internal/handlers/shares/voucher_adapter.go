package shares

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"savvy/internal/database"
	"savvy/internal/models"
	"savvy/internal/services"
)

// VoucherShareAdapter implements ShareAdapter for Voucher resources.
// Vouchers support read-only sharing only (no permission editing).
type VoucherShareAdapter struct {
	db           *gorm.DB
	authzService services.AuthzServiceInterface
}

// NewVoucherShareAdapter creates a new voucher share adapter.
func NewVoucherShareAdapter(db *gorm.DB, authzService services.AuthzServiceInterface) *VoucherShareAdapter {
	return &VoucherShareAdapter{
		db:           db,
		authzService: authzService,
	}
}

// ResourceType returns the resource type identifier.
func (a *VoucherShareAdapter) ResourceType() string {
	return "vouchers"
}

// ResourceName returns the human-readable resource name.
func (a *VoucherShareAdapter) ResourceName() string {
	return "Voucher"
}

// CheckOwnership verifies if the user owns the voucher.
func (a *VoucherShareAdapter) CheckOwnership(ctx context.Context, userID, resourceID uuid.UUID) (bool, error) {
	perms, err := a.authzService.CheckVoucherAccess(ctx, userID, resourceID)
	if err != nil {
		// If forbidden, user doesn't own it
		if errors.Is(err, services.ErrForbidden) {
			return false, nil
		}
		return false, err
	}
	return perms.IsOwner, nil
}

// ListShares returns all shares for a voucher.
func (a *VoucherShareAdapter) ListShares(ctx context.Context, resourceID uuid.UUID) ([]ShareView, error) {
	var shares []models.VoucherShare
	if err := a.db.WithContext(ctx).Where("voucher_id = ?", resourceID).
		Preload("SharedWithUser").
		Order("created_at DESC").
		Find(&shares).Error; err != nil {
		return nil, err
	}

	views := make([]ShareView, len(shares))
	for i, share := range shares {
		views[i] = ShareView{
			ID:         share.ID,
			ResourceID: share.VoucherID,
			SharedWith: share.SharedWithUser,
			// Vouchers are always read-only (no CanEdit, CanDelete permissions)
			CanEdit:   false,
			CanDelete: false,
			CreatedAt: share.CreatedAt,
		}
	}
	return views, nil
}

// CreateShare creates a new voucher share (read-only).
func (a *VoucherShareAdapter) CreateShare(ctx context.Context, req CreateShareRequest) error {
	// Validate email exists
	var sharedUser models.User
	if err := database.DB.Where("LOWER(email) = ?", req.SharedWithEmail).First(&sharedUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	// Check if already shared
	var existingShare models.VoucherShare
	if err := a.db.WithContext(ctx).Where("voucher_id = ? AND shared_with_id = ?",
		req.ResourceID, sharedUser.ID).First(&existingShare).Error; err == nil {
		return errors.New("already shared with this user")
	}

	// Create share (vouchers are always read-only, ignore CanEdit/CanDelete from request)
	share := models.VoucherShare{
		VoucherID:    req.ResourceID,
		SharedWithID: sharedUser.ID,
	}

	return a.db.WithContext(ctx).Create(&share).Error
}

// UpdateShare is not supported for vouchers (read-only sharing).
func (a *VoucherShareAdapter) UpdateShare(ctx context.Context, req UpdateShareRequest) error {
	return errors.New("updating voucher shares is not supported (read-only sharing)")
}

// DeleteShare removes a share.
func (a *VoucherShareAdapter) DeleteShare(ctx context.Context, shareID uuid.UUID) error {
	return a.db.WithContext(ctx).Delete(&models.VoucherShare{}, "id = ?", shareID).Error
}

// SupportsEdit returns false for vouchers (read-only sharing).
func (a *VoucherShareAdapter) SupportsEdit() bool {
	return false
}

// HasTransactionPermission returns false for vouchers (no transaction permission).
func (a *VoucherShareAdapter) HasTransactionPermission() bool {
	return false
}
