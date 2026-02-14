// Package repository defines data access interfaces.
package repository

import (
	"context"
	"savvy/internal/models"

	"github.com/google/uuid"
)

// CardShareRepository defines data access operations for card shares.
type CardShareRepository interface {
	Create(ctx context.Context, share *models.CardShare) error
	GetByCardAndUser(ctx context.Context, cardID, userID uuid.UUID) (*models.CardShare, error)
	GetByCardID(ctx context.Context, cardID uuid.UUID) ([]models.CardShare, error)
	Update(ctx context.Context, share *models.CardShare) error
	DeleteByCardAndUser(ctx context.Context, cardID, userID uuid.UUID) error
	DeleteByCardID(ctx context.Context, cardID uuid.UUID) error
	CountByUser(ctx context.Context, userID uuid.UUID) (int64, error)
	CountByCardIDs(ctx context.Context, cardIDs []uuid.UUID) (map[uuid.UUID]int64, error)
	GetSharedUserIDs(ctx context.Context, ownerID uuid.UUID) ([]uuid.UUID, error)
}
