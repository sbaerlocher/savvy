// Package shares provides unified share handling logic for different resource types.
// It implements the adapter pattern to eliminate code duplication across Card, Voucher, and Gift Card shares.
package shares

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"savvy/internal/services"
)

// VoucherShareAdapter implements ShareAdapter for Voucher resources.
// Vouchers support read-only sharing only (no permission editing).
// Delegates all share operations to ShareService for consistent business logic.
type VoucherShareAdapter struct {
	shareService services.ShareServiceInterface
	authzService services.AuthzServiceInterface
	userService  services.UserServiceInterface
}

// NewVoucherShareAdapter creates a new voucher share adapter.
func NewVoucherShareAdapter(shareService services.ShareServiceInterface, authzService services.AuthzServiceInterface, userService services.UserServiceInterface) *VoucherShareAdapter {
	return &VoucherShareAdapter{
		shareService: shareService,
		authzService: authzService,
		userService:  userService,
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
	shares, err := a.shareService.GetVoucherShares(ctx, resourceID)
	if err != nil {
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
	// Resolve email to user ID
	sharedUser, err := a.userService.GetUserByEmail(ctx, req.SharedWithEmail)
	if err != nil {
		return errors.New("user not found")
	}

	// Delegate to ShareService (handles ownership check, self-share prevention,
	// duplicate detection, notification creation)
	return a.shareService.CreateVoucherShare(ctx, req.UserID, req.ResourceID, sharedUser.ID)
}

// UpdateShare is not supported for vouchers (read-only sharing).
func (a *VoucherShareAdapter) UpdateShare(_ context.Context, _ UpdateShareRequest) error {
	return errors.New("updating voucher shares is not supported (read-only sharing)")
}

// DeleteShare removes a voucher share.
func (a *VoucherShareAdapter) DeleteShare(ctx context.Context, callerUserID, sharedWithID, resourceID uuid.UUID) error {
	return a.shareService.DeleteVoucherShare(ctx, callerUserID, resourceID, sharedWithID)
}

// SupportsEdit returns false for vouchers (read-only sharing).
func (a *VoucherShareAdapter) SupportsEdit() bool {
	return false
}

// HasTransactionPermission returns false for vouchers (no transaction permission).
func (a *VoucherShareAdapter) HasTransactionPermission() bool {
	return false
}
