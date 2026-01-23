// Package services contains business logic.
package services

import (
	"context"
	"errors"
	"savvy/internal/models"
	"savvy/internal/repository"

	"github.com/google/uuid"
)

// ShareServiceInterface defines the interface for sharing business logic.
type ShareServiceInterface interface {
	// ShareCard shares a card with another user.
	ShareCard(ctx context.Context, cardID, sharedWithID uuid.UUID, canEdit, canDelete bool) error

	// ShareVoucher shares a voucher with another user (always read-only).
	ShareVoucher(ctx context.Context, voucherID, sharedWithID uuid.UUID) error

	// ShareGiftCard shares a gift card with another user.
	ShareGiftCard(ctx context.Context, giftCardID, sharedWithID uuid.UUID, canEdit, canDelete, canEditTransactions bool) error

	// UpdateCardShare updates card share permissions.
	UpdateCardShare(ctx context.Context, shareID uuid.UUID, canEdit, canDelete bool) error

	// UpdateGiftCardShare updates gift card share permissions.
	UpdateGiftCardShare(ctx context.Context, shareID uuid.UUID, canEdit, canDelete, canEditTransactions bool) error

	// RevokeCardShare revokes a card share.
	RevokeCardShare(ctx context.Context, shareID uuid.UUID) error

	// RevokeVoucherShare revokes a voucher share.
	RevokeVoucherShare(ctx context.Context, shareID uuid.UUID) error

	// RevokeGiftCardShare revokes a gift card share.
	RevokeGiftCardShare(ctx context.Context, shareID uuid.UUID) error

	// GetCardShares retrieves all shares for a card.
	GetCardShares(ctx context.Context, cardID uuid.UUID) ([]models.CardShare, error)

	// GetVoucherShares retrieves all shares for a voucher.
	GetVoucherShares(ctx context.Context, voucherID uuid.UUID) ([]models.VoucherShare, error)

	// GetGiftCardShares retrieves all shares for a gift card.
	GetGiftCardShares(ctx context.Context, giftCardID uuid.UUID) ([]models.GiftCardShare, error)

	// HasCardAccess checks if user has access to a card.
	HasCardAccess(ctx context.Context, cardID, userID uuid.UUID) (bool, error)

	// HasVoucherAccess checks if user has access to a voucher.
	HasVoucherAccess(ctx context.Context, voucherID, userID uuid.UUID) (bool, error)

	// HasGiftCardAccess checks if user has access to a gift card.
	HasGiftCardAccess(ctx context.Context, giftCardID, userID uuid.UUID) (bool, error)
}

// ShareService implements sharing business logic.
type ShareService struct {
	cardRepo     repository.CardRepository
	voucherRepo  repository.VoucherRepository
	giftCardRepo repository.GiftCardRepository
}

// NewShareService creates a new share service.
func NewShareService(
	cardRepo repository.CardRepository,
	voucherRepo repository.VoucherRepository,
	giftCardRepo repository.GiftCardRepository,
) ShareServiceInterface {
	return &ShareService{
		cardRepo:     cardRepo,
		voucherRepo:  voucherRepo,
		giftCardRepo: giftCardRepo,
	}
}

// ShareCard shares a card with another user.
func (s *ShareService) ShareCard(ctx context.Context, cardID, sharedWithID uuid.UUID, canEdit, canDelete bool) error {
	// Verify card exists
	card, err := s.cardRepo.GetByID(ctx, cardID)
	if err != nil {
		return errors.New("card not found")
	}

	// Prevent sharing with owner
	if card.UserID != nil && *card.UserID == sharedWithID {
		return errors.New("cannot share with card owner")
	}

	// Business logic for share creation would go here
	// For now, this returns an error indicating DB implementation needed
	return errors.New("share creation not implemented - use handlers directly")
}

// ShareVoucher shares a voucher with another user (always read-only).
func (s *ShareService) ShareVoucher(ctx context.Context, voucherID, sharedWithID uuid.UUID) error {
	voucher, err := s.voucherRepo.GetByID(ctx, voucherID)
	if err != nil {
		return errors.New("voucher not found")
	}

	if voucher.UserID != nil && *voucher.UserID == sharedWithID {
		return errors.New("cannot share with voucher owner")
	}

	return errors.New("share creation not implemented - use handlers directly")
}

// ShareGiftCard shares a gift card with another user.
func (s *ShareService) ShareGiftCard(ctx context.Context, giftCardID, sharedWithID uuid.UUID, canEdit, canDelete, canEditTransactions bool) error {
	giftCard, err := s.giftCardRepo.GetByID(ctx, giftCardID)
	if err != nil {
		return errors.New("gift card not found")
	}

	if giftCard.UserID != nil && *giftCard.UserID == sharedWithID {
		return errors.New("cannot share with gift card owner")
	}

	return errors.New("share creation not implemented - use handlers directly")
}

// UpdateCardShare updates card share permissions.
func (s *ShareService) UpdateCardShare(ctx context.Context, shareID uuid.UUID, canEdit, canDelete bool) error {
	return errors.New("share update not implemented - use handlers directly")
}

// UpdateGiftCardShare updates gift card share permissions.
func (s *ShareService) UpdateGiftCardShare(ctx context.Context, shareID uuid.UUID, canEdit, canDelete, canEditTransactions bool) error {
	return errors.New("share update not implemented - use handlers directly")
}

// RevokeCardShare revokes a card share.
func (s *ShareService) RevokeCardShare(ctx context.Context, shareID uuid.UUID) error {
	return errors.New("share revocation not implemented - use handlers directly")
}

// RevokeVoucherShare revokes a voucher share.
func (s *ShareService) RevokeVoucherShare(ctx context.Context, shareID uuid.UUID) error {
	return errors.New("share revocation not implemented - use handlers directly")
}

// RevokeGiftCardShare revokes a gift card share.
func (s *ShareService) RevokeGiftCardShare(ctx context.Context, shareID uuid.UUID) error {
	return errors.New("share revocation not implemented - use handlers directly")
}

// GetCardShares retrieves all shares for a card.
func (s *ShareService) GetCardShares(ctx context.Context, cardID uuid.UUID) ([]models.CardShare, error) {
	return nil, errors.New("get shares not implemented - use handlers directly")
}

// GetVoucherShares retrieves all shares for a voucher.
func (s *ShareService) GetVoucherShares(ctx context.Context, voucherID uuid.UUID) ([]models.VoucherShare, error) {
	return nil, errors.New("get shares not implemented - use handlers directly")
}

// GetGiftCardShares retrieves all shares for a gift card.
func (s *ShareService) GetGiftCardShares(ctx context.Context, giftCardID uuid.UUID) ([]models.GiftCardShare, error) {
	return nil, errors.New("get shares not implemented - use handlers directly")
}

// HasCardAccess checks if user has access to a card (owner or shared).
func (s *ShareService) HasCardAccess(ctx context.Context, cardID, userID uuid.UUID) (bool, error) {
	card, err := s.cardRepo.GetByID(ctx, cardID)
	if err != nil {
		return false, err
	}

	// Owner has access
	if card.UserID != nil && *card.UserID == userID {
		return true, nil
	}

	// Check shared access (would need share repository)
	return false, errors.New("share access check not implemented - use handlers directly")
}

// HasVoucherAccess checks if user has access to a voucher (owner or shared).
func (s *ShareService) HasVoucherAccess(ctx context.Context, voucherID, userID uuid.UUID) (bool, error) {
	voucher, err := s.voucherRepo.GetByID(ctx, voucherID)
	if err != nil {
		return false, err
	}

	if voucher.UserID != nil && *voucher.UserID == userID {
		return true, nil
	}

	return false, errors.New("share access check not implemented - use handlers directly")
}

// HasGiftCardAccess checks if user has access to a gift card (owner or shared).
func (s *ShareService) HasGiftCardAccess(ctx context.Context, giftCardID, userID uuid.UUID) (bool, error) {
	giftCard, err := s.giftCardRepo.GetByID(ctx, giftCardID)
	if err != nil {
		return false, err
	}

	if giftCard.UserID != nil && *giftCard.UserID == userID {
		return true, nil
	}

	return false, errors.New("share access check not implemented - use handlers directly")
}
