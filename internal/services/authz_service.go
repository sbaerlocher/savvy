// Package services contains business logic.
package services

import (
	"context"
	"errors"
	"savvy/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ResourcePermissions represents access permissions for a resource
type ResourcePermissions struct {
	CanView             bool
	CanEdit             bool
	CanDelete           bool
	CanEditTransactions bool // Only used for GiftCards
	IsOwner             bool
}

// ErrForbidden is returned when user doesn't have access to a resource
var ErrForbidden = errors.New("access forbidden")

// AuthzServiceInterface defines the interface for authorization checks
type AuthzServiceInterface interface {
	CheckCardAccess(ctx context.Context, userID, cardID uuid.UUID) (*ResourcePermissions, error)
	CheckVoucherAccess(ctx context.Context, userID, voucherID uuid.UUID) (*ResourcePermissions, error)
	CheckGiftCardAccess(ctx context.Context, userID, giftCardID uuid.UUID) (*ResourcePermissions, error)
}

// AuthzService implements authorization checks for resources
type AuthzService struct {
	db *gorm.DB
}

// NewAuthzService creates a new authorization service
func NewAuthzService(db *gorm.DB) AuthzServiceInterface {
	return &AuthzService{db: db}
}

// CheckCardAccess checks if a user has access to a card and returns permissions
func (s *AuthzService) CheckCardAccess(ctx context.Context, userID, cardID uuid.UUID) (*ResourcePermissions, error) {
	var card models.Card
	if err := s.db.WithContext(ctx).First(&card, "id = ?", cardID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrForbidden
		}
		return nil, err
	}

	// Check ownership
	if card.UserID != nil && *card.UserID == userID {
		return &ResourcePermissions{
			CanView:   true,
			CanEdit:   true,
			CanDelete: true,
			IsOwner:   true,
		}, nil
	}

	// Check shared access
	var share models.CardShare
	if err := s.db.WithContext(ctx).Where("card_id = ? AND shared_with_id = ?", cardID, userID).First(&share).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrForbidden
		}
		return nil, err
	}

	return &ResourcePermissions{
		CanView:   true,
		CanEdit:   share.CanEdit,
		CanDelete: share.CanDelete,
		IsOwner:   false,
	}, nil
}

// CheckVoucherAccess checks if a user has access to a voucher and returns permissions
func (s *AuthzService) CheckVoucherAccess(ctx context.Context, userID, voucherID uuid.UUID) (*ResourcePermissions, error) {
	var voucher models.Voucher
	if err := s.db.WithContext(ctx).First(&voucher, "id = ?", voucherID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrForbidden
		}
		return nil, err
	}

	// Check ownership
	if voucher.UserID != nil && *voucher.UserID == userID {
		return &ResourcePermissions{
			CanView:   true,
			CanEdit:   true,
			CanDelete: true,
			IsOwner:   true,
		}, nil
	}

	// Check shared access (vouchers are read-only when shared)
	var share models.VoucherShare
	if err := s.db.WithContext(ctx).Where("voucher_id = ? AND shared_with_id = ?", voucherID, userID).First(&share).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrForbidden
		}
		return nil, err
	}

	return &ResourcePermissions{
		CanView:   true,
		CanEdit:   false, // Vouchers are always read-only when shared
		CanDelete: false,
		IsOwner:   false,
	}, nil
}

// CheckGiftCardAccess checks if a user has access to a gift card and returns permissions
func (s *AuthzService) CheckGiftCardAccess(ctx context.Context, userID, giftCardID uuid.UUID) (*ResourcePermissions, error) {
	var giftCard models.GiftCard
	if err := s.db.WithContext(ctx).First(&giftCard, "id = ?", giftCardID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrForbidden
		}
		return nil, err
	}

	// Check ownership
	if giftCard.UserID != nil && *giftCard.UserID == userID {
		return &ResourcePermissions{
			CanView:             true,
			CanEdit:             true,
			CanDelete:           true,
			CanEditTransactions: true,
			IsOwner:             true,
		}, nil
	}

	// Check shared access
	var share models.GiftCardShare
	if err := s.db.WithContext(ctx).Where("gift_card_id = ? AND shared_with_id = ?", giftCardID, userID).First(&share).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrForbidden
		}
		return nil, err
	}

	return &ResourcePermissions{
		CanView:             true,
		CanEdit:             share.CanEdit,
		CanDelete:           share.CanDelete,
		CanEditTransactions: share.CanEditTransactions,
		IsOwner:             false,
	}, nil
}
