package shares

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"savvy/internal/models"
	"savvy/internal/services"
)

// CardShareAdapter implements ShareAdapter for Card resources.
type CardShareAdapter struct {
	db           *gorm.DB
	authzService services.AuthzServiceInterface
	userService  services.UserServiceInterface
}

// NewCardShareAdapter creates a new card share adapter.
func NewCardShareAdapter(db *gorm.DB, authzService services.AuthzServiceInterface, userService services.UserServiceInterface) *CardShareAdapter {
	return &CardShareAdapter{
		db:           db,
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
	var shares []models.CardShare
	if err := a.db.WithContext(ctx).Where("card_id = ?", resourceID).
		Preload("SharedWithUser").
		Order("created_at DESC").
		Find(&shares).Error; err != nil {
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
	// Validate email exists
	sharedUser, err := a.userService.GetUserByEmail(ctx, req.SharedWithEmail)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	// Check if already shared
	var existingShare models.CardShare
	if err := a.db.WithContext(ctx).Where("card_id = ? AND shared_with_id = ?",
		req.ResourceID, sharedUser.ID).First(&existingShare).Error; err == nil {
		return errors.New("already shared with this user")
	}

	// Create share
	share := models.CardShare{
		CardID:       req.ResourceID,
		SharedWithID: sharedUser.ID,
		CanEdit:      req.CanEdit,
		CanDelete:    req.CanDelete,
	}

	return a.db.WithContext(ctx).Create(&share).Error
}

// UpdateShare updates share permissions.
func (a *CardShareAdapter) UpdateShare(ctx context.Context, req UpdateShareRequest) error {
	var share models.CardShare
	if err := a.db.WithContext(ctx).Where("id = ? AND card_id = ?",
		req.ShareID, req.ResourceID).First(&share).Error; err != nil {
		return err
	}

	share.CanEdit = req.CanEdit
	share.CanDelete = req.CanDelete

	return a.db.WithContext(ctx).Save(&share).Error
}

// DeleteShare removes a share.
func (a *CardShareAdapter) DeleteShare(ctx context.Context, shareID uuid.UUID) error {
	return a.db.WithContext(ctx).Delete(&models.CardShare{}, "id = ?", shareID).Error
}

// SupportsEdit returns true for cards (share permissions can be edited).
func (a *CardShareAdapter) SupportsEdit() bool {
	return true
}

// HasTransactionPermission returns false for cards (no transaction permission).
func (a *CardShareAdapter) HasTransactionPermission() bool {
	return false
}
