// Package repository contains data access implementations.
package repository

import (
	"context"
	"savvy/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GormFavoriteRepository is a GORM implementation of FavoriteRepository.
type GormFavoriteRepository struct {
	db *gorm.DB
}

// NewFavoriteRepository creates a new favorite repository.
func NewFavoriteRepository(db *gorm.DB) FavoriteRepository {
	return &GormFavoriteRepository{db: db}
}

func (r *GormFavoriteRepository) Create(ctx context.Context, favorite *models.UserFavorite) error {
	return r.db.WithContext(ctx).Create(favorite).Error
}

// GetByUserAndResource includes soft-deleted favorites for toggle functionality.
func (r *GormFavoriteRepository) GetByUserAndResource(ctx context.Context, userID uuid.UUID, resourceType string, resourceID uuid.UUID) (*models.UserFavorite, error) {
	var favorite models.UserFavorite
	err := r.db.WithContext(ctx).
		Unscoped().
		Where("user_id = ? AND resource_type = ? AND resource_id = ?", userID, resourceType, resourceID).
		First(&favorite).Error
	if err != nil {
		return nil, err
	}
	return &favorite, nil
}

func (r *GormFavoriteRepository) GetByUser(ctx context.Context, userID uuid.UUID) ([]models.UserFavorite, error) {
	var favorites []models.UserFavorite
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&favorites).Error
	return favorites, err
}

func (r *GormFavoriteRepository) Delete(ctx context.Context, favorite *models.UserFavorite) error {
	return r.db.WithContext(ctx).Delete(favorite).Error
}

func (r *GormFavoriteRepository) Restore(ctx context.Context, favorite *models.UserFavorite) error {
	return r.db.WithContext(ctx).
		Unscoped().
		Model(favorite).
		Update("deleted_at", nil).Error
}

func (r *GormFavoriteRepository) IsFavorite(ctx context.Context, userID uuid.UUID, resourceType string, resourceID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.UserFavorite{}).
		Where("user_id = ? AND resource_type = ? AND resource_id = ?", userID, resourceType, resourceID).
		Count(&count).Error
	return count > 0, err
}
