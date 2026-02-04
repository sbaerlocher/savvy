// Package repository defines data access interfaces.
package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Entity is a constraint for types that can be used with BaseRepository.
// All repository entities must have an ID field of type uuid.UUID.
type Entity interface {
	any
}

// ShareConfig defines the configuration for GetSharedWithUser queries.
type ShareConfig struct {
	// ShareTableName is the name of the share table (e.g., "card_shares")
	ShareTableName string
	// ResourceIDColumn is the column name for the resource ID (e.g., "card_id")
	ResourceIDColumn string
	// TableName is the name of the main table (e.g., "cards")
	TableName string
}

// BaseRepository provides generic CRUD operations for entities.
// T is the entity type (e.g., models.Card, models.Voucher).
type BaseRepository[T Entity] struct {
	db          *gorm.DB
	shareConfig *ShareConfig
}

// NewBaseRepository creates a new base repository.
func NewBaseRepository[T Entity](db *gorm.DB, shareConfig *ShareConfig) *BaseRepository[T] {
	return &BaseRepository[T]{
		db:          db,
		shareConfig: shareConfig,
	}
}

// Create creates a new entity.
func (r *BaseRepository[T]) Create(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

// GetByID retrieves an entity by ID with optional preloads.
func (r *BaseRepository[T]) GetByID(ctx context.Context, id uuid.UUID, preloads ...string) (*T, error) {
	var entity T
	query := r.db.WithContext(ctx)

	for _, preload := range preloads {
		query = query.Preload(preload)
	}

	if err := query.First(&entity, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return &entity, nil
}

// GetByUserID retrieves all entities for a user.
func (r *BaseRepository[T]) GetByUserID(ctx context.Context, userID uuid.UUID) ([]T, error) {
	var entities []T
	err := r.db.WithContext(ctx).
		Preload("Merchant").
		Preload("User").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&entities).Error

	return entities, err
}

// GetSharedWithUser retrieves entities shared with a user (only active shares).
// Requires shareConfig to be set.
func (r *BaseRepository[T]) GetSharedWithUser(ctx context.Context, userID uuid.UUID) ([]T, error) {
	if r.shareConfig == nil {
		return nil, gorm.ErrInvalidData
	}

	var entities []T
	err := r.db.WithContext(ctx).
		Preload("Merchant").
		Preload("User").
		Joins("INNER JOIN "+r.shareConfig.ShareTableName+" ON "+
			r.shareConfig.ShareTableName+"."+r.shareConfig.ResourceIDColumn+" = "+
			r.shareConfig.TableName+".id").
		Where(r.shareConfig.ShareTableName+".shared_with_id = ? AND "+
			r.shareConfig.ShareTableName+".deleted_at IS NULL", userID).
		Order(r.shareConfig.TableName + ".created_at DESC").
		Find(&entities).Error

	return entities, err
}

// Update updates an entity.
func (r *BaseRepository[T]) Update(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Save(entity).Error
}

// Delete soft-deletes an entity.
func (r *BaseRepository[T]) Delete(ctx context.Context, id uuid.UUID) error {
	var entity T
	return r.db.WithContext(ctx).Delete(&entity, "id = ?", id).Error
}

// Count counts entities for a user.
func (r *BaseRepository[T]) Count(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	var entity T
	err := r.db.WithContext(ctx).
		Model(&entity).
		Where("user_id = ?", userID).
		Count(&count).Error

	return count, err
}
