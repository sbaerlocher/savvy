// Package repository defines data access interfaces.
package repository

import (
	"context"
	"savvy/internal/models"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GormUserRepository implements UserRepository using GORM.
type GormUserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository.
func NewUserRepository(db *gorm.DB) UserRepository {
	return &GormUserRepository{db: db}
}

// GetByID retrieves a user by their ID.
func (r *GormUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail normalizes email to lowercase for case-insensitive lookup.
func (r *GormUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))

	var user models.User
	if err := r.db.WithContext(ctx).Where("LOWER(email) = ?", normalizedEmail).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// Create creates a new user in the database.
func (r *GormUserRepository) Create(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// Update updates an existing user in the database.
func (r *GormUserRepository) Update(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// GetAll retrieves all users ordered by creation date (descending).
func (r *GormUserRepository) GetAll(ctx context.Context) ([]models.User, error) {
	var users []models.User
	if err := r.db.WithContext(ctx).Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// GetByIDs retrieves users by a list of IDs.
func (r *GormUserRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]models.User, error) {
	if len(ids) == 0 {
		return []models.User{}, nil
	}
	var users []models.User
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// SearchByIDs retrieves users matching the given IDs, optionally filtered by a search query.
func (r *GormUserRepository) SearchByIDs(ctx context.Context, ids []uuid.UUID, query string) ([]models.User, error) {
	if len(ids) == 0 {
		return []models.User{}, nil
	}
	q := r.db.WithContext(ctx).Where("id IN ?", ids)
	if query != "" {
		q = q.Where("email ILIKE ? OR first_name ILIKE ? OR last_name ILIKE ?",
			"%"+query+"%", "%"+query+"%", "%"+query+"%")
	}
	var users []models.User
	if err := q.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
