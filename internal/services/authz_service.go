// Package services contains business logic.
package services

import (
	"context"
	"errors"
	"fmt"
	"savvy/internal/repository"

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
	cardRepo          repository.CardRepository
	voucherRepo       repository.VoucherRepository
	giftCardRepo      repository.GiftCardRepository
	cardShareRepo     repository.CardShareRepository
	voucherShareRepo  repository.VoucherShareRepository
	giftCardShareRepo repository.GiftCardShareRepository
}

// NewAuthzService creates a new authorization service
func NewAuthzService(
	cardRepo repository.CardRepository,
	voucherRepo repository.VoucherRepository,
	giftCardRepo repository.GiftCardRepository,
	cardShareRepo repository.CardShareRepository,
	voucherShareRepo repository.VoucherShareRepository,
	giftCardShareRepo repository.GiftCardShareRepository,
) AuthzServiceInterface {
	return &AuthzService{
		cardRepo:          cardRepo,
		voucherRepo:       voucherRepo,
		giftCardRepo:      giftCardRepo,
		cardShareRepo:     cardShareRepo,
		voucherShareRepo:  voucherShareRepo,
		giftCardShareRepo: giftCardShareRepo,
	}
}

// CheckCardAccess checks if a user has access to a card and returns permissions
func (s *AuthzService) CheckCardAccess(ctx context.Context, userID, cardID uuid.UUID) (*ResourcePermissions, error) {
	card, err := s.cardRepo.GetByID(ctx, cardID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrForbidden
		}
		return nil, fmt.Errorf("check card access: %w", err)
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
	share, err := s.cardShareRepo.GetByCardAndUser(ctx, cardID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrForbidden
		}
		return nil, fmt.Errorf("check card share access: %w", err)
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
	voucher, err := s.voucherRepo.GetByID(ctx, voucherID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrForbidden
		}
		return nil, fmt.Errorf("check voucher access: %w", err)
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
	_, err = s.voucherShareRepo.GetByVoucherAndUser(ctx, voucherID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrForbidden
		}
		return nil, fmt.Errorf("check voucher share access: %w", err)
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
	giftCard, err := s.giftCardRepo.GetByID(ctx, giftCardID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrForbidden
		}
		return nil, fmt.Errorf("check gift card access: %w", err)
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
	share, err := s.giftCardShareRepo.GetByGiftCardAndUser(ctx, giftCardID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrForbidden
		}
		return nil, fmt.Errorf("check gift card share access: %w", err)
	}

	return &ResourcePermissions{
		CanView:             true,
		CanEdit:             share.CanEdit,
		CanDelete:           share.CanDelete,
		CanEditTransactions: share.CanEditTransactions,
		IsOwner:             false,
	}, nil
}
