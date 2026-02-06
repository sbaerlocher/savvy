// Package services contains business logic.
package services

import (
	"context"
	"errors"
	"log/slog"
	"savvy/internal/models"
	"savvy/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TransferServiceInterface defines the interface for ownership transfer business logic.
type TransferServiceInterface interface {
	TransferCardOwnership(ctx context.Context, cardID, newOwnerID, currentOwnerID uuid.UUID) error
	TransferVoucherOwnership(ctx context.Context, voucherID, newOwnerID, currentOwnerID uuid.UUID) error
	TransferGiftCardOwnership(ctx context.Context, giftCardID, newOwnerID, currentOwnerID uuid.UUID) error
}

// TransferService implements TransferServiceInterface.
type TransferService struct {
	db                  *gorm.DB
	cardRepo            repository.CardRepository
	voucherRepo         repository.VoucherRepository
	giftCardRepo        repository.GiftCardRepository
	notificationService NotificationServiceInterface
}

// NewTransferService creates a new transfer service.
func NewTransferService(
	db *gorm.DB,
	cardRepo repository.CardRepository,
	voucherRepo repository.VoucherRepository,
	giftCardRepo repository.GiftCardRepository,
	notificationService NotificationServiceInterface,
) TransferServiceInterface {
	return &TransferService{
		db:                  db,
		cardRepo:            cardRepo,
		voucherRepo:         voucherRepo,
		giftCardRepo:        giftCardRepo,
		notificationService: notificationService,
	}
}

// TransferCardOwnership transfers ownership of a card to a new owner.
// - Validates new owner exists
// - Updates user_id
// - Deletes ALL existing shares (clean slate)
// - Returns error if current user is not owner
func (s *TransferService) TransferCardOwnership(ctx context.Context, cardID, newOwnerID, currentOwnerID uuid.UUID) error {
	// 1. Validate new owner exists
	var newOwner models.User
	if err := s.db.WithContext(ctx).Where("id = ?", newOwnerID).First(&newOwner).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("new owner not found")
		}
		return err
	}

	// 2. Get resource and verify current user is owner (not just shared access)
	card, err := s.cardRepo.GetByID(ctx, cardID)
	if err != nil {
		return err
	}
	if card.UserID == nil || *card.UserID != currentOwnerID {
		return errors.New("only owner can transfer")
	}

	// 3. Prevent self-transfer
	if newOwnerID == currentOwnerID {
		return errors.New("cannot transfer to yourself")
	}

	// 4. Update ownership
	card.UserID = &newOwnerID
	if err := s.cardRepo.Update(ctx, card); err != nil {
		return err
	}

	// 5. Delete ALL shares (clean slate)
	if err := s.db.WithContext(ctx).Where("card_id = ?", cardID).Delete(&models.CardShare{}).Error; err != nil {
		return err
	}

	// 6. Create notification for new owner
	var currentUser models.User
	if err := s.db.WithContext(ctx).Where("id = ?", currentOwnerID).First(&currentUser).Error; err == nil {
		// Best effort notification - don't fail the transfer if notification fails
		if err := s.notificationService.CreateTransferNotification(
			ctx,
			newOwnerID,
			currentOwnerID,
			currentUser.DisplayName(),
			"card",
			cardID,
		); err != nil {
			slog.Warn("Failed to create transfer notification for card",
				"card_id", cardID,
				"new_owner_id", newOwnerID,
				"error", err)
		}
	}

	return nil
}

// TransferVoucherOwnership transfers ownership of a voucher to a new owner.
// - Validates new owner exists
// - Updates user_id
// - Deletes ALL existing shares (clean slate)
// - Returns error if current user is not owner
func (s *TransferService) TransferVoucherOwnership(ctx context.Context, voucherID, newOwnerID, currentOwnerID uuid.UUID) error {
	// 1. Validate new owner exists
	var newOwner models.User
	if err := s.db.WithContext(ctx).Where("id = ?", newOwnerID).First(&newOwner).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("new owner not found")
		}
		return err
	}

	// 2. Get resource and verify current user is owner (not just shared access)
	voucher, err := s.voucherRepo.GetByID(ctx, voucherID)
	if err != nil {
		return err
	}
	if voucher.UserID == nil || *voucher.UserID != currentOwnerID {
		return errors.New("only owner can transfer")
	}

	// 3. Prevent self-transfer
	if newOwnerID == currentOwnerID {
		return errors.New("cannot transfer to yourself")
	}

	// 4. Update ownership
	voucher.UserID = &newOwnerID
	if err := s.voucherRepo.Update(ctx, voucher); err != nil {
		return err
	}

	// 5. Delete ALL shares (clean slate)
	if err := s.db.WithContext(ctx).Where("voucher_id = ?", voucherID).Delete(&models.VoucherShare{}).Error; err != nil {
		return err
	}

	// 6. Create notification for new owner
	var currentUser models.User
	if err := s.db.WithContext(ctx).Where("id = ?", currentOwnerID).First(&currentUser).Error; err == nil {
		// Best effort notification - don't fail the transfer if notification fails
		if err := s.notificationService.CreateTransferNotification(
			ctx,
			newOwnerID,
			currentOwnerID,
			currentUser.DisplayName(),
			"voucher",
			voucherID,
		); err != nil {
			slog.Warn("Failed to create transfer notification for voucher",
				"voucher_id", voucherID,
				"new_owner_id", newOwnerID,
				"error", err)
		}
	}

	return nil
}

// TransferGiftCardOwnership transfers ownership of a gift card to a new owner.
// - Validates new owner exists
// - Updates user_id
// - Deletes ALL existing shares (clean slate)
// - Returns error if current user is not owner
func (s *TransferService) TransferGiftCardOwnership(ctx context.Context, giftCardID, newOwnerID, currentOwnerID uuid.UUID) error {
	// 1. Validate new owner exists
	var newOwner models.User
	if err := s.db.WithContext(ctx).Where("id = ?", newOwnerID).First(&newOwner).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("new owner not found")
		}
		return err
	}

	// 2. Get resource and verify current user is owner (not just shared access)
	giftCard, err := s.giftCardRepo.GetByID(ctx, giftCardID)
	if err != nil {
		return err
	}
	if giftCard.UserID == nil || *giftCard.UserID != currentOwnerID {
		return errors.New("only owner can transfer")
	}

	// 3. Prevent self-transfer
	if newOwnerID == currentOwnerID {
		return errors.New("cannot transfer to yourself")
	}

	// 4. Update ownership
	giftCard.UserID = &newOwnerID
	if err := s.giftCardRepo.Update(ctx, giftCard); err != nil {
		return err
	}

	// 5. Delete ALL shares (clean slate)
	if err := s.db.WithContext(ctx).Where("gift_card_id = ?", giftCardID).Delete(&models.GiftCardShare{}).Error; err != nil {
		return err
	}

	// 6. Create notification for new owner
	var currentUser models.User
	if err := s.db.WithContext(ctx).Where("id = ?", currentOwnerID).First(&currentUser).Error; err == nil {
		// Best effort notification - don't fail the transfer if notification fails
		if err := s.notificationService.CreateTransferNotification(
			ctx,
			newOwnerID,
			currentOwnerID,
			currentUser.DisplayName(),
			"gift_card",
			giftCardID,
		); err != nil {
			slog.Warn("Failed to create transfer notification for gift card",
				"gift_card_id", giftCardID,
				"new_owner_id", newOwnerID,
				"error", err)
		}
	}

	return nil
}
