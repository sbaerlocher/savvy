// Package repository contains data access interfaces and implementations.
package repository

import (
	"context"
	"savvy/internal/models"

	"github.com/google/uuid"
)

// MerchantRepository defines the interface for merchant data access.
type MerchantRepository interface {
	// Create creates a new merchant.
	Create(ctx context.Context, merchant *models.Merchant) error

	// GetByID retrieves a merchant by ID.
	GetByID(ctx context.Context, id uuid.UUID) (*models.Merchant, error)

	// GetAll retrieves all merchants.
	GetAll(ctx context.Context) ([]models.Merchant, error)

	// Search searches merchants by name.
	Search(ctx context.Context, query string) ([]models.Merchant, error)

	// Update updates an existing merchant.
	Update(ctx context.Context, merchant *models.Merchant) error

	// Delete deletes a merchant by ID.
	Delete(ctx context.Context, id uuid.UUID) error

	// Count returns the total number of merchants.
	Count(ctx context.Context) (int64, error)
}
