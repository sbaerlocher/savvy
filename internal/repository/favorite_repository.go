// Package repository contains data access interfaces and implementations.
package repository

import (
	"context"
	"savvy/internal/models"

	"github.com/google/uuid"
)

// FavoriteRepository defines the interface for favorite data access.
type FavoriteRepository interface {
	// Create creates a new favorite (or restores soft-deleted).
	Create(ctx context.Context, favorite *models.UserFavorite) error

	// GetByUserAndResource retrieves a favorite by user and resource (including soft-deleted).
	GetByUserAndResource(ctx context.Context, userID uuid.UUID, resourceType string, resourceID uuid.UUID) (*models.UserFavorite, error)

	// GetByUser retrieves all favorites for a user.
	GetByUser(ctx context.Context, userID uuid.UUID) ([]models.UserFavorite, error)

	// Delete soft-deletes a favorite.
	Delete(ctx context.Context, favorite *models.UserFavorite) error

	// Restore restores a soft-deleted favorite.
	Restore(ctx context.Context, favorite *models.UserFavorite) error

	// IsFavorite checks if a resource is favorited by user.
	IsFavorite(ctx context.Context, userID uuid.UUID, resourceType string, resourceID uuid.UUID) (bool, error)
}
