// Package repository defines data access interfaces.
//
//nolint:dupl // Wrapper methods required for interface compliance with Go generics
package repository

import (
	"context"
	"savvy/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GormCardRepository implements CardRepository using GORM.
type GormCardRepository struct {
	*BaseRepository[models.Card]
}

// NewCardRepository creates a new card repository.
func NewCardRepository(db *gorm.DB) CardRepository {
	return &GormCardRepository{
		BaseRepository: NewBaseRepository[models.Card](db, &ShareConfig{
			ShareTableName:   "card_shares",
			ResourceIDColumn: "card_id",
			TableName:        "cards",
		}),
	}
}

// Create creates a new card.
func (r *GormCardRepository) Create(ctx context.Context, card *models.Card) error {
	return r.BaseRepository.Create(ctx, card)
}

// GetByID retrieves a card by ID with optional preloads.
func (r *GormCardRepository) GetByID(ctx context.Context, id uuid.UUID, preloads ...string) (*models.Card, error) {
	return r.BaseRepository.GetByID(ctx, id, preloads...)
}

// GetByUserID retrieves all cards for a user.
func (r *GormCardRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.Card, error) {
	return r.BaseRepository.GetByUserID(ctx, userID)
}

// GetSharedWithUser retrieves cards shared with a user.
func (r *GormCardRepository) GetSharedWithUser(ctx context.Context, userID uuid.UUID) ([]models.Card, error) {
	return r.BaseRepository.GetSharedWithUser(ctx, userID)
}

// Update updates a card.
func (r *GormCardRepository) Update(ctx context.Context, card *models.Card) error {
	return r.BaseRepository.Update(ctx, card)
}

// Delete soft-deletes a card.
func (r *GormCardRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.BaseRepository.Delete(ctx, id)
}

// Count counts cards for a user.
func (r *GormCardRepository) Count(ctx context.Context, userID uuid.UUID) (int64, error) {
	return r.BaseRepository.Count(ctx, userID)
}
