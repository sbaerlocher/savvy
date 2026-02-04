// Package services contains business logic.
package services

import (
	"context"
	"errors"
	"savvy/internal/models"
	"savvy/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ShareServiceInterface defines the interface for share business logic.
type ShareServiceInterface interface {
	CreateCardShare(ctx context.Context, cardID, sharedWithID uuid.UUID, canEdit, canDelete bool) error
	CreateVoucherShare(ctx context.Context, voucherID, sharedWithID uuid.UUID) error
	CreateGiftCardShare(ctx context.Context, giftCardID, sharedWithID uuid.UUID, canEdit, canDelete, canEditTransactions bool) error
	GetCardShares(ctx context.Context, cardID uuid.UUID) ([]models.CardShare, error)
	GetVoucherShares(ctx context.Context, voucherID uuid.UUID) ([]models.VoucherShare, error)
	GetGiftCardShares(ctx context.Context, giftCardID uuid.UUID) ([]models.GiftCardShare, error)
	GetSharedUsers(ctx context.Context, userID uuid.UUID, searchQuery string) ([]models.User, error)
}

// ShareService implements ShareServiceInterface.
type ShareService struct {
	cardRepo     repository.CardRepository
	voucherRepo  repository.VoucherRepository
	giftCardRepo repository.GiftCardRepository
	db           *gorm.DB
}

// NewShareService creates a new share service.
func NewShareService(
	cardRepo repository.CardRepository,
	voucherRepo repository.VoucherRepository,
	giftCardRepo repository.GiftCardRepository,
	db *gorm.DB,
) ShareServiceInterface {
	return &ShareService{
		cardRepo:     cardRepo,
		voucherRepo:  voucherRepo,
		giftCardRepo: giftCardRepo,
		db:           db,
	}
}

// CreateCardShare creates a new card share.
func (s *ShareService) CreateCardShare(ctx context.Context, cardID, sharedWithID uuid.UUID, canEdit, canDelete bool) error {
	// Business logic: validate share
	if cardID == uuid.Nil {
		return errors.New("card ID is required")
	}
	if sharedWithID == uuid.Nil {
		return errors.New("shared with user ID is required")
	}

	// Verify card exists
	card, err := s.cardRepo.GetByID(ctx, cardID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("card not found")
		}
		return err
	}

	// Prevent sharing with owner
	if card.UserID != nil && *card.UserID == sharedWithID {
		return errors.New("cannot share card with its owner")
	}

	share := models.CardShare{
		CardID:       cardID,
		SharedWithID: sharedWithID,
		CanEdit:      canEdit,
		CanDelete:    canDelete,
	}

	return s.db.WithContext(ctx).Create(&share).Error
}

// CreateVoucherShare creates a new voucher share.
// Note: Vouchers are always read-only, so no edit/delete permissions.
func (s *ShareService) CreateVoucherShare(ctx context.Context, voucherID, sharedWithID uuid.UUID) error {
	// Business logic: validate share
	if voucherID == uuid.Nil {
		return errors.New("voucher ID is required")
	}
	if sharedWithID == uuid.Nil {
		return errors.New("shared with user ID is required")
	}

	// Verify voucher exists
	voucher, err := s.voucherRepo.GetByID(ctx, voucherID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("voucher not found")
		}
		return err
	}

	// Prevent sharing with owner
	if voucher.UserID != nil && *voucher.UserID == sharedWithID {
		return errors.New("cannot share voucher with its owner")
	}

	share := models.VoucherShare{
		VoucherID:    voucherID,
		SharedWithID: sharedWithID,
		// Vouchers are always read-only
	}

	return s.db.WithContext(ctx).Create(&share).Error
}

// CreateGiftCardShare creates a new gift card share.
func (s *ShareService) CreateGiftCardShare(ctx context.Context, giftCardID, sharedWithID uuid.UUID, canEdit, canDelete, canEditTransactions bool) error {
	// Business logic: validate share
	if giftCardID == uuid.Nil {
		return errors.New("gift card ID is required")
	}
	if sharedWithID == uuid.Nil {
		return errors.New("shared with user ID is required")
	}

	// Verify gift card exists
	giftCard, err := s.giftCardRepo.GetByID(ctx, giftCardID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("gift card not found")
		}
		return err
	}

	// Prevent sharing with owner
	if giftCard.UserID != nil && *giftCard.UserID == sharedWithID {
		return errors.New("cannot share gift card with its owner")
	}

	share := models.GiftCardShare{
		GiftCardID:          giftCardID,
		SharedWithID:        sharedWithID,
		CanEdit:             canEdit,
		CanDelete:           canDelete,
		CanEditTransactions: canEditTransactions,
	}

	return s.db.WithContext(ctx).Create(&share).Error
}

// GetCardShares retrieves all active (non-deleted) shares for a card.
func (s *ShareService) GetCardShares(ctx context.Context, cardID uuid.UUID) ([]models.CardShare, error) {
	var shares []models.CardShare
	if err := s.db.WithContext(ctx).
		Where("card_id = ?", cardID).
		Where("deleted_at IS NULL").
		Preload("SharedWithUser").
		Find(&shares).Error; err != nil {
		return nil, err
	}
	return shares, nil
}

// GetVoucherShares retrieves all active (non-deleted) shares for a voucher.
func (s *ShareService) GetVoucherShares(ctx context.Context, voucherID uuid.UUID) ([]models.VoucherShare, error) {
	var shares []models.VoucherShare
	if err := s.db.WithContext(ctx).
		Where("voucher_id = ?", voucherID).
		Where("deleted_at IS NULL").
		Preload("SharedWithUser").
		Find(&shares).Error; err != nil {
		return nil, err
	}
	return shares, nil
}

// GetGiftCardShares retrieves all active (non-deleted) shares for a gift card.
func (s *ShareService) GetGiftCardShares(ctx context.Context, giftCardID uuid.UUID) ([]models.GiftCardShare, error) {
	var shares []models.GiftCardShare
	if err := s.db.WithContext(ctx).
		Where("gift_card_id = ?", giftCardID).
		Where("deleted_at IS NULL").
		Preload("SharedWithUser").
		Find(&shares).Error; err != nil {
		return nil, err
	}
	return shares, nil
}

// GetSharedUsers retrieves all unique users that the given user has shared resources with.
// Optionally filters by search query (email, first name, or last name).
func (s *ShareService) GetSharedUsers(ctx context.Context, userID uuid.UUID, searchQuery string) ([]models.User, error) {
	var userIDs []uuid.UUID

	// Card shares (only active shares on active cards)
	var cardSharedUserIDs []uuid.UUID
	s.db.WithContext(ctx).Table("card_shares").
		Select("DISTINCT shared_with_id").
		Joins("JOIN cards ON cards.id = card_shares.card_id").
		Where("cards.user_id = ? AND card_shares.deleted_at IS NULL AND cards.deleted_at IS NULL", userID).
		Pluck("shared_with_id", &cardSharedUserIDs)
	userIDs = append(userIDs, cardSharedUserIDs...)

	// Voucher shares (only active shares on active vouchers)
	var voucherSharedUserIDs []uuid.UUID
	s.db.WithContext(ctx).Table("voucher_shares").
		Select("DISTINCT shared_with_id").
		Joins("JOIN vouchers ON vouchers.id = voucher_shares.voucher_id").
		Where("vouchers.user_id = ? AND voucher_shares.deleted_at IS NULL AND vouchers.deleted_at IS NULL", userID).
		Pluck("shared_with_id", &voucherSharedUserIDs)
	userIDs = append(userIDs, voucherSharedUserIDs...)

	// Gift card shares (only active shares on active gift cards)
	var giftCardSharedUserIDs []uuid.UUID
	s.db.WithContext(ctx).Table("gift_card_shares").
		Select("DISTINCT shared_with_id").
		Joins("JOIN gift_cards ON gift_cards.id = gift_card_shares.gift_card_id").
		Where("gift_cards.user_id = ? AND gift_card_shares.deleted_at IS NULL AND gift_cards.deleted_at IS NULL", userID).
		Pluck("shared_with_id", &giftCardSharedUserIDs)
	userIDs = append(userIDs, giftCardSharedUserIDs...)

	// Remove duplicates from userIDs
	uniqueUserIDs := make(map[uuid.UUID]bool)
	var filteredUserIDs []uuid.UUID
	for _, id := range userIDs {
		if !uniqueUserIDs[id] {
			uniqueUserIDs[id] = true
			filteredUserIDs = append(filteredUserIDs, id)
		}
	}

	// If no shared users, return empty array
	if len(filteredUserIDs) == 0 {
		return []models.User{}, nil
	}

	// Fetch users by IDs
	var users []models.User
	query := s.db.WithContext(ctx).Where("id IN ?", filteredUserIDs)

	// Apply search filter if provided
	if searchQuery != "" {
		query = query.Where("email ILIKE ? OR first_name ILIKE ? OR last_name ILIKE ?",
			"%"+searchQuery+"%", "%"+searchQuery+"%", "%"+searchQuery+"%")
	}

	if err := query.Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}
