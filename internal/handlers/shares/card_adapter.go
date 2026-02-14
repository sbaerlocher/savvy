// Package shares provides unified share handling logic for different resource types.
// It implements the adapter pattern to eliminate code duplication across Card, Voucher, and Gift Card shares.
package shares

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"savvy/internal/services"
)

// CardShareAdapter implements ShareAdapter for Card resources.
// Delegates all share operations to ShareService for consistent business logic.
type CardShareAdapter struct {
	shareService services.ShareServiceInterface
	authzService services.AuthzServiceInterface
	userService  services.UserServiceInterface
}

// NewCardShareAdapter creates a new card share adapter.
func NewCardShareAdapter(shareService services.ShareServiceInterface, authzService services.AuthzServiceInterface, userService services.UserServiceInterface) *CardShareAdapter {
	return &CardShareAdapter{
		shareService: shareService,
		authzService: authzService,
		userService:  userService,
	}
}

// ResourceType returns the resource type identifier.
func (a *CardShareAdapter) ResourceType() string {
	return "cards"
}

// ResourceName returns the human-readable resource name.
func (a *CardShareAdapter) ResourceName() string {
	return "Card"
}

// CheckOwnership verifies if the user owns the card.
func (a *CardShareAdapter) CheckOwnership(ctx context.Context, userID, resourceID uuid.UUID) (bool, error) {
	perms, err := a.authzService.CheckCardAccess(ctx, userID, resourceID)
	if err != nil {
		// If forbidden, user doesn't own it
		if errors.Is(err, services.ErrForbidden) {
			return false, nil
		}
		return false, err
	}
	return perms.IsOwner, nil
}

// ListShares returns all shares for a card.
func (a *CardShareAdapter) ListShares(ctx context.Context, resourceID uuid.UUID) ([]ShareView, error) {
	shares, err := a.shareService.GetCardShares(ctx, resourceID)
	if err != nil {
		return nil, err
	}

	views := make([]ShareView, len(shares))
	for i, share := range shares {
		views[i] = ShareView{
			ID:         share.ID,
			ResourceID: share.CardID,
			SharedWith: share.SharedWithUser,
			CanEdit:    share.CanEdit,
			CanDelete:  share.CanDelete,
			CreatedAt:  share.CreatedAt,
		}
	}
	return views, nil
}

// CreateShare creates a new card share.
func (a *CardShareAdapter) CreateShare(ctx context.Context, req CreateShareRequest) error {
	// Resolve email to user ID
	sharedUser, err := a.userService.GetUserByEmail(ctx, req.SharedWithEmail)
	if err != nil {
		return errors.New("user not found")
	}

	// Delegate to ShareService (handles ownership check, self-share prevention,
	// duplicate detection, notification creation)
	return a.shareService.CreateCardShare(ctx, req.UserID, req.ResourceID, sharedUser.ID, req.CanEdit, req.CanDelete)
}

// UpdateShare updates share permissions.
func (a *CardShareAdapter) UpdateShare(ctx context.Context, req UpdateShareRequest) error {
	return a.shareService.UpdateCardShare(ctx, req.CallerUserID, req.ResourceID, req.SharedWithID, req.CanEdit, req.CanDelete)
}

// DeleteShare removes a card share.
func (a *CardShareAdapter) DeleteShare(ctx context.Context, callerUserID, sharedWithID, resourceID uuid.UUID) error {
	return a.shareService.DeleteCardShare(ctx, callerUserID, resourceID, sharedWithID)
}

// SupportsEdit returns true for cards (share permissions can be edited).
func (a *CardShareAdapter) SupportsEdit() bool {
	return true
}

// HasTransactionPermission returns false for cards (no transaction permission).
func (a *CardShareAdapter) HasTransactionPermission() bool {
	return false
}
