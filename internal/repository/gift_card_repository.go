// Package repository defines data access interfaces.
package repository

import (
	"context"
	"savvy/internal/models"

	"github.com/google/uuid"
)

// GiftCardRepository defines the interface for gift card data access.
type GiftCardRepository interface {
	// Create creates a new gift card
	Create(ctx context.Context, giftCard *models.GiftCard) error

	// GetByID retrieves a gift card by ID with optional preloads
	GetByID(ctx context.Context, id uuid.UUID, preloads ...string) (*models.GiftCard, error)

	// GetByUserID retrieves all gift cards for a user
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.GiftCard, error)

	// GetSharedWithUser retrieves gift cards shared with a user
	GetSharedWithUser(ctx context.Context, userID uuid.UUID) ([]models.GiftCard, error)

	// Update updates a gift card
	Update(ctx context.Context, giftCard *models.GiftCard) error

	// Delete soft-deletes a gift card
	Delete(ctx context.Context, id uuid.UUID) error

	// Count counts gift cards for a user
	Count(ctx context.Context, userID uuid.UUID) (int64, error)

	// GetTotalBalance calculates total balance across all user's gift cards
	GetTotalBalance(ctx context.Context, userID uuid.UUID) (float64, error)

	// CreateTransaction creates a new transaction for a gift card
	CreateTransaction(ctx context.Context, transaction *models.GiftCardTransaction) error

	// GetTransaction retrieves a transaction by ID, validating it belongs to the gift card
	GetTransaction(ctx context.Context, transactionID, giftCardID uuid.UUID) (*models.GiftCardTransaction, error)

	// DeleteTransaction deletes a transaction by ID
	DeleteTransaction(ctx context.Context, transactionID uuid.UUID) error
}
