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

// GetByID retrieves a user by ID.
func (r *GormUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail retrieves a user by email (case-insensitive).
func (r *GormUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	// Normalize email to lowercase for case-insensitive search
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))

	var user models.User
	if err := r.db.WithContext(ctx).Where("LOWER(email) = ?", normalizedEmail).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// Create creates a new user.
func (r *GormUserRepository) Create(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// Update updates a user.
func (r *GormUserRepository) Update(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}
