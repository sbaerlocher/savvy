// Package repository defines data access interfaces.
package repository

import (
	"context"
	"errors"
	"fmt"
	"savvy/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ErrOwnershipChanged indicates a concurrent ownership change was detected.
var ErrOwnershipChanged = errors.New("ownership changed concurrently")

// GormTransferRepository implements TransferRepository using GORM transactions.
type GormTransferRepository struct {
	db *gorm.DB
}

// NewTransferRepository creates a new transfer repository.
func NewTransferRepository(db *gorm.DB) TransferRepository {
	return &GormTransferRepository{db: db}
}

// TransferCardOwnership atomically updates the card owner and removes all existing shares.
// Uses conditional UPDATE to prevent TOCTOU race conditions.
func (r *GormTransferRepository) TransferCardOwnership(ctx context.Context, card *models.Card, newOwnerID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Atomically update only if current owner still matches
		result := tx.Model(&models.Card{}).
			Where("id = ? AND user_id = ?", card.ID, card.UserID).
			Update("user_id", newOwnerID)
		if result.Error != nil {
			return fmt.Errorf("update card ownership: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return ErrOwnershipChanged
		}
		if err := tx.Where("card_id = ?", card.ID).Delete(&models.CardShare{}).Error; err != nil {
			return fmt.Errorf("delete card shares: %w", err)
		}
		return nil
	})
}

// TransferVoucherOwnership atomically updates the voucher owner and removes all existing shares.
// Uses conditional UPDATE to prevent TOCTOU race conditions.
func (r *GormTransferRepository) TransferVoucherOwnership(ctx context.Context, voucher *models.Voucher, newOwnerID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&models.Voucher{}).
			Where("id = ? AND user_id = ?", voucher.ID, voucher.UserID).
			Update("user_id", newOwnerID)
		if result.Error != nil {
			return fmt.Errorf("update voucher ownership: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return ErrOwnershipChanged
		}
		if err := tx.Where("voucher_id = ?", voucher.ID).Delete(&models.VoucherShare{}).Error; err != nil {
			return fmt.Errorf("delete voucher shares: %w", err)
		}
		return nil
	})
}

// TransferGiftCardOwnership atomically updates the gift card owner and removes all existing shares.
// Uses conditional UPDATE to prevent TOCTOU race conditions.
func (r *GormTransferRepository) TransferGiftCardOwnership(ctx context.Context, giftCard *models.GiftCard, newOwnerID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&models.GiftCard{}).
			Where("id = ? AND user_id = ?", giftCard.ID, giftCard.UserID).
			Update("user_id", newOwnerID)
		if result.Error != nil {
			return fmt.Errorf("update gift card ownership: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return ErrOwnershipChanged
		}
		if err := tx.Where("gift_card_id = ?", giftCard.ID).Delete(&models.GiftCardShare{}).Error; err != nil {
			return fmt.Errorf("delete gift card shares: %w", err)
		}
		return nil
	})
}
