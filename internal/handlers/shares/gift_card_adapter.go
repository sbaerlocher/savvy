// Package shares provides unified share handling logic for different resource types.
// It implements the adapter pattern to eliminate code duplication across Card, Voucher, and Gift Card shares.
package shares

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"savvy/internal/services"
)

// GiftCardShareAdapter implements ShareAdapter for GiftCard resources.
// Gift cards support granular permissions including CanEditTransactions.
// Delegates all share operations to ShareService for consistent business logic.
type GiftCardShareAdapter struct {
	shareService services.ShareServiceInterface
	authzService services.AuthzServiceInterface
	userService  services.UserServiceInterface
}

// NewGiftCardShareAdapter creates a new gift card share adapter.
func NewGiftCardShareAdapter(shareService services.ShareServiceInterface, authzService services.AuthzServiceInterface, userService services.UserServiceInterface) *GiftCardShareAdapter {
	return &GiftCardShareAdapter{
		shareService: shareService,
		authzService: authzService,
		userService:  userService,
	}
}

// ResourceType returns the resource type identifier.
func (a *GiftCardShareAdapter) ResourceType() string {
	return "gift_cards"
}

// ResourceName returns the human-readable resource name.
func (a *GiftCardShareAdapter) ResourceName() string {
	return "Gift Card"
}

// CheckOwnership verifies if the user owns the gift card.
func (a *GiftCardShareAdapter) CheckOwnership(ctx context.Context, userID, resourceID uuid.UUID) (bool, error) {
	perms, err := a.authzService.CheckGiftCardAccess(ctx, userID, resourceID)
	if err != nil {
		// If forbidden, user doesn't own it
		if errors.Is(err, services.ErrForbidden) {
			return false, nil
		}
		return false, err
	}
	return perms.IsOwner, nil
}

// ListShares returns all shares for a gift card.
func (a *GiftCardShareAdapter) ListShares(ctx context.Context, resourceID uuid.UUID) ([]ShareView, error) {
	shares, err := a.shareService.GetGiftCardShares(ctx, resourceID)
	if err != nil {
		return nil, err
	}

	views := make([]ShareView, len(shares))
	for i, share := range shares {
		views[i] = ShareView{
			ID:                  share.ID,
			ResourceID:          share.GiftCardID,
			SharedWith:          share.SharedWithUser,
			CanEdit:             share.CanEdit,
			CanDelete:           share.CanDelete,
			CanEditTransactions: share.CanEditTransactions,
			CreatedAt:           share.CreatedAt,
		}
	}
	return views, nil
}

// CreateShare creates a new gift card share.
func (a *GiftCardShareAdapter) CreateShare(ctx context.Context, req CreateShareRequest) error {
	// Resolve email to user ID
	sharedUser, err := a.userService.GetUserByEmail(ctx, req.SharedWithEmail)
	if err != nil {
		return errors.New("user not found")
	}

	// Delegate to ShareService (handles ownership check, self-share prevention,
	// duplicate detection, notification creation)
	return a.shareService.CreateGiftCardShare(ctx, req.UserID, req.ResourceID, sharedUser.ID, req.CanEdit, req.CanDelete, req.CanEditTransactions)
}

// UpdateShare updates share permissions.
func (a *GiftCardShareAdapter) UpdateShare(ctx context.Context, req UpdateShareRequest) error {
	return a.shareService.UpdateGiftCardShare(ctx, req.CallerUserID, req.ResourceID, req.SharedWithID, req.CanEdit, req.CanDelete, req.CanEditTransactions)
}

// DeleteShare removes a gift card share.
func (a *GiftCardShareAdapter) DeleteShare(ctx context.Context, callerUserID, sharedWithID, resourceID uuid.UUID) error {
	return a.shareService.DeleteGiftCardShare(ctx, callerUserID, resourceID, sharedWithID)
}

// SupportsEdit returns true for gift cards (share permissions can be edited).
func (a *GiftCardShareAdapter) SupportsEdit() bool {
	return true
}

// HasTransactionPermission returns true for gift cards (supports CanEditTransactions).
func (a *GiftCardShareAdapter) HasTransactionPermission() bool {
	return true
}
