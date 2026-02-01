// Package repository defines data access interfaces.
package repository

import (
	"context"
	"savvy/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GormGiftCardRepository implements GiftCardRepository using GORM.
type GormGiftCardRepository struct {
	db *gorm.DB
}

// NewGiftCardRepository creates a new gift card repository.
func NewGiftCardRepository(db *gorm.DB) GiftCardRepository {
	return &GormGiftCardRepository{db: db}
}

// Create creates a new gift card.
func (r *GormGiftCardRepository) Create(ctx context.Context, giftCard *models.GiftCard) error {
	return r.db.WithContext(ctx).Create(giftCard).Error
}

// GetByID retrieves a gift card by ID with optional preloads.
func (r *GormGiftCardRepository) GetByID(ctx context.Context, id uuid.UUID, preloads ...string) (*models.GiftCard, error) {
	var giftCard models.GiftCard
	query := r.db.WithContext(ctx)

	for _, preload := range preloads {
		query = query.Preload(preload)
	}

	if err := query.First(&giftCard, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return &giftCard, nil
}

// GetByUserID retrieves all gift cards for a user.
func (r *GormGiftCardRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.GiftCard, error) {
	var giftCards []models.GiftCard
	err := r.db.WithContext(ctx).
		Preload("Merchant").
		Preload("User").
		Preload("Transactions", func(db *gorm.DB) *gorm.DB {
			return db.Order("transaction_date DESC")
		}).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&giftCards).Error

	return giftCards, err
}

// GetSharedWithUser retrieves gift cards shared with a user.
func (r *GormGiftCardRepository) GetSharedWithUser(ctx context.Context, userID uuid.UUID) ([]models.GiftCard, error) {
	var giftCards []models.GiftCard
	err := r.db.WithContext(ctx).
		Preload("Merchant").
		Preload("User").
		Preload("Transactions", func(db *gorm.DB) *gorm.DB {
			return db.Order("transaction_date DESC")
		}).
		Joins("INNER JOIN gift_card_shares ON gift_card_shares.gift_card_id = gift_cards.id").
		Where("gift_card_shares.shared_with_id = ?", userID).
		Order("gift_cards.created_at DESC").
		Find(&giftCards).Error

	return giftCards, err
}

// Update updates a gift card.
func (r *GormGiftCardRepository) Update(ctx context.Context, giftCard *models.GiftCard) error {
	return r.db.WithContext(ctx).Save(giftCard).Error
}

// Delete soft-deletes a gift card.
func (r *GormGiftCardRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.GiftCard{}, "id = ?", id).Error
}

// Count counts gift cards for a user.
func (r *GormGiftCardRepository) Count(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.GiftCard{}).
		Where("user_id = ?", userID).
		Count(&count).Error

	return count, err
}

// GetTotalBalance calculates total balance across all user's gift cards.
func (r *GormGiftCardRepository) GetTotalBalance(ctx context.Context, userID uuid.UUID) (float64, error) {
	giftCards, err := r.GetByUserID(ctx, userID)
	if err != nil {
		return 0, err
	}

	var totalBalance float64
	for _, gc := range giftCards {
		totalBalance += gc.CurrentBalance
	}

	return totalBalance, nil
}

// CreateTransaction creates a new transaction for a gift card.
func (r *GormGiftCardRepository) CreateTransaction(ctx context.Context, transaction *models.GiftCardTransaction) error {
	return r.db.WithContext(ctx).Create(transaction).Error
}

// GetTransaction retrieves a transaction by ID, validating it belongs to the gift card.
func (r *GormGiftCardRepository) GetTransaction(ctx context.Context, transactionID, giftCardID uuid.UUID) (*models.GiftCardTransaction, error) {
	var transaction models.GiftCardTransaction
	err := r.db.WithContext(ctx).
		Where("id = ? AND gift_card_id = ?", transactionID, giftCardID).
		First(&transaction).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

// DeleteTransaction deletes a transaction by ID.
func (r *GormGiftCardRepository) DeleteTransaction(ctx context.Context, transactionID uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.GiftCardTransaction{}, "id = ?", transactionID).Error
}
