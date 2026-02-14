// Package repository defines data access interfaces.
package repository

import (
	"context"
	"savvy/internal/models"

	"github.com/google/uuid"
)

// UserRepository defines the interface for user data access.
type UserRepository interface {
	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)

	// GetByEmail retrieves a user by email (case-insensitive)
	GetByEmail(ctx context.Context, email string) (*models.User, error)

	// Create creates a new user
	Create(ctx context.Context, user *models.User) error

	// Update updates a user
	Update(ctx context.Context, user *models.User) error

	// GetAll retrieves all users ordered by creation date (descending)
	GetAll(ctx context.Context) ([]models.User, error)

	// GetByIDs retrieves users by a list of IDs
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]models.User, error)

	// SearchByIDs retrieves users matching the given IDs, optionally filtered by a search query
	SearchByIDs(ctx context.Context, ids []uuid.UUID, query string) ([]models.User, error)
}
