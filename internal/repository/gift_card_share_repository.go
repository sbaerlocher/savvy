// Package repository defines data access interfaces.
package repository

import (
	"context"
	"savvy/internal/models"

	"github.com/google/uuid"
)

// GiftCardShareRepository defines data access operations for gift card shares.
type GiftCardShareRepository interface {
	Create(ctx context.Context, share *models.GiftCardShare) error
	GetByGiftCardAndUser(ctx context.Context, giftCardID, userID uuid.UUID) (*models.GiftCardShare, error)
	GetByGiftCardID(ctx context.Context, giftCardID uuid.UUID) ([]models.GiftCardShare, error)
	Update(ctx context.Context, share *models.GiftCardShare) error
	DeleteByGiftCardAndUser(ctx context.Context, giftCardID, userID uuid.UUID) error
	DeleteByGiftCardID(ctx context.Context, giftCardID uuid.UUID) error
	CountByUser(ctx context.Context, userID uuid.UUID) (int64, error)
	CountByGiftCardIDs(ctx context.Context, giftCardIDs []uuid.UUID) (map[uuid.UUID]int64, error)
	GetSharedUserIDs(ctx context.Context, ownerID uuid.UUID) ([]uuid.UUID, error)
}
