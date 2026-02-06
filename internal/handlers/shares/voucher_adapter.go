// Package shares provides unified share handling logic for different resource types.
// It implements the adapter pattern to eliminate code duplication across Card, Voucher, and Gift Card shares.
package shares

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"savvy/internal/models"
	"savvy/internal/services"
)

// VoucherShareAdapter implements ShareAdapter for Voucher resources.
// Vouchers support read-only sharing only (no permission editing).
type VoucherShareAdapter struct {
	db                  *gorm.DB
	authzService        services.AuthzServiceInterface
	userService         services.UserServiceInterface
	notificationService services.NotificationServiceInterface
}

// NewVoucherShareAdapter creates a new voucher share adapter.
func NewVoucherShareAdapter(db *gorm.DB, authzService services.AuthzServiceInterface, userService services.UserServiceInterface, notificationService services.NotificationServiceInterface) *VoucherShareAdapter {
	return &VoucherShareAdapter{
		db:                  db,
		authzService:        authzService,
		userService:         userService,
		notificationService: notificationService,
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
	sharedUser, err := a.userService.GetUserByEmail(ctx, req.SharedWithEmail)
	if err != nil {
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

	if err := a.db.WithContext(ctx).Create(&share).Error; err != nil {
		return err
	}

	// Create notification for the shared user
	var voucher models.Voucher
	if err := a.db.WithContext(ctx).Where("id = ?", req.ResourceID).First(&voucher).Error; err == nil {
		if voucher.UserID != nil {
			var ownerUser models.User
			if err := a.db.WithContext(ctx).Where("id = ?", *voucher.UserID).First(&ownerUser).Error; err == nil {
				// Best effort notification - don't fail the share if notification fails
				// Vouchers are always read-only (no permissions)
				if err := a.notificationService.CreateShareNotification(
					ctx,
					sharedUser.ID,
					*voucher.UserID,
					ownerUser.DisplayName(),
					"voucher",
					req.ResourceID,
					nil, // No permissions for vouchers (always read-only)
				); err != nil {
					slog.Warn("Failed to create share notification for voucher",
						"voucher_id", req.ResourceID,
						"shared_with_id", sharedUser.ID,
						"error", err)
				}
			}
		}
	}

	return nil
}

// UpdateShare is not supported for vouchers (read-only sharing).
func (a *VoucherShareAdapter) UpdateShare(_ context.Context, _ UpdateShareRequest) error {
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
