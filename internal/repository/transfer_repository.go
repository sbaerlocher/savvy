// Package repository defines data access interfaces.
package repository

import (
	"context"
	"savvy/internal/models"

	"github.com/google/uuid"
)

// TransferRepository defines data access operations for atomic ownership transfers.
type TransferRepository interface {
	TransferCardOwnership(ctx context.Context, card *models.Card, newOwnerID uuid.UUID) error
	TransferVoucherOwnership(ctx context.Context, voucher *models.Voucher, newOwnerID uuid.UUID) error
	TransferGiftCardOwnership(ctx context.Context, giftCard *models.GiftCard, newOwnerID uuid.UUID) error
}
