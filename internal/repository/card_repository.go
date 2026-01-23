// Package repository defines data access interfaces.
package repository

import (
	"context"
	"savvy/internal/models"

	"github.com/google/uuid"
)

// CardRepository defines the interface for card data access.
type CardRepository interface {
	// Create creates a new card
	Create(ctx context.Context, card *models.Card) error

	// GetByID retrieves a card by ID with optional preloads
	GetByID(ctx context.Context, id uuid.UUID, preloads ...string) (*models.Card, error)

	// GetByUserID retrieves all cards for a user
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.Card, error)

	// GetSharedWithUser retrieves cards shared with a user
	GetSharedWithUser(ctx context.Context, userID uuid.UUID) ([]models.Card, error)

	// Update updates a card
	Update(ctx context.Context, card *models.Card) error

	// Delete soft-deletes a card
	Delete(ctx context.Context, id uuid.UUID) error

	// Count counts cards for a user
	Count(ctx context.Context, userID uuid.UUID) (int64, error)
}
