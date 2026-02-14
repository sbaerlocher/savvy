// Package repository defines data access interfaces.
package repository

import (
	"context"
	"fmt"
	"regexp"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// validIdentifier matches safe SQL identifiers (lowercase letters and underscores only).
var validIdentifier = regexp.MustCompile(`^[a-z_]+$`)

// Entity is a constraint for types that can be used with BaseRepository.
// All repository entities must have an ID field of type uuid.UUID.
type Entity interface {
	any
}

// ShareConfig defines the configuration for GetSharedWithUser queries.
// All field values MUST be hardcoded SQL identifiers (lowercase letters and underscores only).
type ShareConfig struct {
	// ShareTableName is the name of the share table (e.g., "card_shares")
	ShareTableName string
	// ResourceIDColumn is the column name for the resource ID (e.g., "card_id")
	ResourceIDColumn string
	// TableName is the name of the main table (e.g., "cards")
	TableName string
}

// Validate checks that all ShareConfig fields are non-empty and contain only safe SQL identifiers.
func (sc *ShareConfig) Validate() error {
	if sc.ShareTableName == "" {
		return fmt.Errorf("ShareConfig: ShareTableName must not be empty")
	}
	if sc.ResourceIDColumn == "" {
		return fmt.Errorf("ShareConfig: ResourceIDColumn must not be empty")
	}
	if sc.TableName == "" {
		return fmt.Errorf("ShareConfig: TableName must not be empty")
	}
	if !validIdentifier.MatchString(sc.ShareTableName) {
		return fmt.Errorf("ShareConfig: ShareTableName %q contains invalid characters", sc.ShareTableName)
	}
	if !validIdentifier.MatchString(sc.ResourceIDColumn) {
		return fmt.Errorf("ShareConfig: ResourceIDColumn %q contains invalid characters", sc.ResourceIDColumn)
	}
	if !validIdentifier.MatchString(sc.TableName) {
		return fmt.Errorf("ShareConfig: TableName %q contains invalid characters", sc.TableName)
	}
	return nil
}

// BaseRepository provides generic CRUD operations for entities.
// T is the entity type (e.g., models.Card, models.Voucher).
type BaseRepository[T Entity] struct {
	db          *gorm.DB
	shareConfig *ShareConfig
}

// NewBaseRepository creates a new base repository.
// Panics if shareConfig is provided but contains invalid values (fail-fast at initialization).
func NewBaseRepository[T Entity](db *gorm.DB, shareConfig *ShareConfig) *BaseRepository[T] {
	if shareConfig != nil {
		if err := shareConfig.Validate(); err != nil {
			panic(fmt.Sprintf("invalid share config: %v", err))
		}
	}
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
	// Load entity first to populate data for audit logging
	var entity T
	if err := r.db.WithContext(ctx).Preload("Merchant").First(&entity, "id = ?", id).Error; err != nil {
		return err
	}

	// Now delete with populated entity data
	return r.db.WithContext(ctx).Delete(&entity).Error
}

// GetAllForUserPaginated retrieves all entities (owned + shared) for a user with pagination.
// Uses a subquery approach to combine owned and shared items efficiently.
func (r *BaseRepository[T]) GetAllForUserPaginated(ctx context.Context, userID uuid.UUID, params PaginationParams) (*PaginatedResult[T], error) {
	params = NormalizePagination(params)

	// Build the base condition: owned OR shared with user
	baseQuery := r.db.WithContext(ctx).
		Preload("Merchant").
		Preload("User")

	if r.shareConfig != nil {
		baseQuery = baseQuery.Where(
			r.db.Where("user_id = ?", userID).Or(
				"id IN (SELECT "+r.shareConfig.ResourceIDColumn+" FROM "+r.shareConfig.ShareTableName+
					" WHERE shared_with_id = ? AND deleted_at IS NULL)", userID,
			),
		)
	} else {
		baseQuery = baseQuery.Where("user_id = ?", userID)
	}

	// Count total
	var total int64
	var countModel T
	countQuery := r.db.WithContext(ctx).Model(&countModel)
	if r.shareConfig != nil {
		countQuery = countQuery.Where(
			r.db.Where("user_id = ?", userID).Or(
				"id IN (SELECT "+r.shareConfig.ResourceIDColumn+" FROM "+r.shareConfig.ShareTableName+
					" WHERE shared_with_id = ? AND deleted_at IS NULL)", userID,
			),
		)
	} else {
		countQuery = countQuery.Where("user_id = ?", userID)
	}
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, err
	}

	// Fetch paginated items
	var items []T
	if err := ApplyPagination(baseQuery, params).
		Order("created_at DESC").
		Find(&items).Error; err != nil {
		return nil, err
	}

	return &PaginatedResult[T]{
		Items:      items,
		Total:      total,
		Page:       params.Page,
		PerPage:    params.PerPage,
		TotalPages: CalculateTotalPages(total, params.PerPage),
	}, nil
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
