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

// Create creates a new card in the database.
func (r *GormCardRepository) Create(ctx context.Context, card *models.Card) error {
	return r.BaseRepository.Create(ctx, card)
}

// GetByID retrieves a card by its ID with optional preloads.
func (r *GormCardRepository) GetByID(ctx context.Context, id uuid.UUID, preloads ...string) (*models.Card, error) {
	return r.BaseRepository.GetByID(ctx, id, preloads...)
}

// GetByUserID retrieves all cards owned by a specific user.
func (r *GormCardRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.Card, error) {
	return r.BaseRepository.GetByUserID(ctx, userID)
}

// GetSharedWithUser retrieves all cards shared with a specific user.
func (r *GormCardRepository) GetSharedWithUser(ctx context.Context, userID uuid.UUID) ([]models.Card, error) {
	return r.BaseRepository.GetSharedWithUser(ctx, userID)
}

// Update updates an existing card in the database.
func (r *GormCardRepository) Update(ctx context.Context, card *models.Card) error {
	return r.BaseRepository.Update(ctx, card)
}

// Delete soft-deletes a card from the database.
func (r *GormCardRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.BaseRepository.Delete(ctx, id)
}

// Count returns the total number of cards owned by a user.
func (r *GormCardRepository) Count(ctx context.Context, userID uuid.UUID) (int64, error) {
	return r.BaseRepository.Count(ctx, userID)
}

// GetAllForUserPaginated retrieves all cards (owned + shared) with pagination.
func (r *GormCardRepository) GetAllForUserPaginated(ctx context.Context, userID uuid.UUID, params PaginationParams) (*PaginatedResult[models.Card], error) {
	return r.BaseRepository.GetAllForUserPaginated(ctx, userID, params)
}

// FindByCardNumber finds a card by card number for a specific user.
func (r *GormCardRepository) FindByCardNumber(ctx context.Context, cardNumber string, userID uuid.UUID) (*models.Card, error) {
	var card models.Card
	err := r.db.WithContext(ctx).
		Where("card_number = ? AND user_id = ?", cardNumber, userID).
		First(&card).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // No duplicate found
		}
		return nil, err
	}

	return &card, nil
}

// FindDeletedByCardNumber finds a soft-deleted card by card number for a specific user.
func (r *GormCardRepository) FindDeletedByCardNumber(ctx context.Context, cardNumber string, userID uuid.UUID) (*models.Card, error) {
	var card models.Card
	err := r.db.WithContext(ctx).
		Unscoped().
		Where("card_number = ? AND user_id = ? AND deleted_at IS NOT NULL", cardNumber, userID).
		First(&card).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &card, nil
}

// RestoreByID clears deleted_at for a soft-deleted card owned by the user.
func (r *GormCardRepository) RestoreByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Unscoped().
		Model(&models.Card{}).
		Where("id = ? AND user_id = ? AND deleted_at IS NOT NULL", id, userID).
		Update("deleted_at", nil).Error
}

// Search searches cards by query (merchant name, card number, program, notes).
