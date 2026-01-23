// Package repository defines data access interfaces.
package repository

import (
	"context"
	"savvy/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GormCardRepository implements CardRepository using GORM.
type GormCardRepository struct {
	db *gorm.DB
}

// NewCardRepository creates a new card repository.
func NewCardRepository(db *gorm.DB) CardRepository {
	return &GormCardRepository{db: db}
}

// Create creates a new card.
func (r *GormCardRepository) Create(ctx context.Context, card *models.Card) error {
	return r.db.WithContext(ctx).Create(card).Error
}

// GetByID retrieves a card by ID with optional preloads.
func (r *GormCardRepository) GetByID(ctx context.Context, id uuid.UUID, preloads ...string) (*models.Card, error) {
	var card models.Card
	query := r.db.WithContext(ctx)

	for _, preload := range preloads {
		query = query.Preload(preload)
	}

	if err := query.First(&card, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return &card, nil
}

// GetByUserID retrieves all cards for a user.
func (r *GormCardRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.Card, error) {
	var cards []models.Card
	err := r.db.WithContext(ctx).
		Preload("Merchant").
		Preload("User").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&cards).Error

	return cards, err
}

// GetSharedWithUser retrieves cards shared with a user.
func (r *GormCardRepository) GetSharedWithUser(ctx context.Context, userID uuid.UUID) ([]models.Card, error) {
	var cards []models.Card
	err := r.db.WithContext(ctx).
		Preload("Merchant").
		Preload("User").
		Joins("INNER JOIN card_shares ON card_shares.card_id = cards.id").
		Where("card_shares.shared_with_id = ?", userID).
		Order("cards.created_at DESC").
		Find(&cards).Error

	return cards, err
}

// Update updates a card.
func (r *GormCardRepository) Update(ctx context.Context, card *models.Card) error {
	return r.db.WithContext(ctx).Save(card).Error
}

// Delete soft-deletes a card.
func (r *GormCardRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Card{}, "id = ?", id).Error
}

// Count counts cards for a user.
func (r *GormCardRepository) Count(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Card{}).
		Where("user_id = ?", userID).
		Count(&count).Error

	return count, err
}
