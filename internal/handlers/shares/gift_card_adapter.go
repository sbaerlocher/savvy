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

// GiftCardShareAdapter implements ShareAdapter for GiftCard resources.
// Gift cards support granular permissions including CanEditTransactions.
type GiftCardShareAdapter struct {
	db                  *gorm.DB
	authzService        services.AuthzServiceInterface
	userService         services.UserServiceInterface
	notificationService services.NotificationServiceInterface
}

// NewGiftCardShareAdapter creates a new gift card share adapter.
func NewGiftCardShareAdapter(db *gorm.DB, authzService services.AuthzServiceInterface, userService services.UserServiceInterface, notificationService services.NotificationServiceInterface) *GiftCardShareAdapter {
	return &GiftCardShareAdapter{
		db:                  db,
		authzService:        authzService,
		userService:         userService,
		notificationService: notificationService,
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
	var shares []models.GiftCardShare
	if err := a.db.WithContext(ctx).Where("gift_card_id = ?", resourceID).
		Preload("SharedWithUser").
		Order("created_at DESC").
		Find(&shares).Error; err != nil {
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
			CanEditTransactions: share.CanEditTransactions, // Gift card specific permission
			CreatedAt:           share.CreatedAt,
		}
	}
	return views, nil
}

// CreateShare creates a new gift card share.
func (a *GiftCardShareAdapter) CreateShare(ctx context.Context, req CreateShareRequest) error {
	// Validate email exists
	sharedUser, err := a.userService.GetUserByEmail(ctx, req.SharedWithEmail)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	// Check if already shared
	var existingShare models.GiftCardShare
	if err := a.db.WithContext(ctx).Where("gift_card_id = ? AND shared_with_id = ?",
		req.ResourceID, sharedUser.ID).First(&existingShare).Error; err == nil {
		return errors.New("already shared with this user")
	}

	// Create share with gift card specific permissions
	share := models.GiftCardShare{
		GiftCardID:          req.ResourceID,
		SharedWithID:        sharedUser.ID,
		CanEdit:             req.CanEdit,
		CanDelete:           req.CanDelete,
		CanEditTransactions: req.CanEditTransactions, // Gift card specific
	}

	if err := a.db.WithContext(ctx).Create(&share).Error; err != nil {
		return err
	}

	// Create notification for the shared user
	var giftCard models.GiftCard
	if err := a.db.WithContext(ctx).Where("id = ?", req.ResourceID).First(&giftCard).Error; err == nil {
		if giftCard.UserID != nil {
			var ownerUser models.User
			if err := a.db.WithContext(ctx).Where("id = ?", *giftCard.UserID).First(&ownerUser).Error; err == nil {
				// Best effort notification - don't fail the share if notification fails
				if err := a.notificationService.CreateShareNotification(
					ctx,
					sharedUser.ID,
					*giftCard.UserID,
					ownerUser.DisplayName(),
					"gift_card",
					req.ResourceID,
					map[string]bool{
						"can_edit":              req.CanEdit,
						"can_delete":            req.CanDelete,
						"can_edit_transactions": req.CanEditTransactions,
					},
				); err != nil {
					slog.Warn("Failed to create share notification for gift card",
						"gift_card_id", req.ResourceID,
						"shared_with_id", sharedUser.ID,
						"error", err)
				}
			}
		}
	}

	return nil
}

// UpdateShare updates share permissions.
func (a *GiftCardShareAdapter) UpdateShare(ctx context.Context, req UpdateShareRequest) error {
	var share models.GiftCardShare
	if err := a.db.WithContext(ctx).Where("id = ? AND gift_card_id = ?",
		req.ShareID, req.ResourceID).First(&share).Error; err != nil {
		return err
	}

	share.CanEdit = req.CanEdit
	share.CanDelete = req.CanDelete
	share.CanEditTransactions = req.CanEditTransactions // Gift card specific

	return a.db.WithContext(ctx).Save(&share).Error
}

// DeleteShare removes a share.
func (a *GiftCardShareAdapter) DeleteShare(ctx context.Context, shareID uuid.UUID) error {
	return a.db.WithContext(ctx).Delete(&models.GiftCardShare{}, "id = ?", shareID).Error
}

// SupportsEdit returns true for gift cards (share permissions can be edited).
func (a *GiftCardShareAdapter) SupportsEdit() bool {
	return true
}

// HasTransactionPermission returns true for gift cards (supports CanEditTransactions).
func (a *GiftCardShareAdapter) HasTransactionPermission() bool {
	return true
}
