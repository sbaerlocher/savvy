// Package repository defines data access interfaces.
package repository

import (
	"context"
	"savvy/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TOTPRepository defines the interface for TOTP data access.
type TOTPRepository interface {
	// GetByUserID retrieves TOTP config for a user
	GetByUserID(ctx context.Context, userID uuid.UUID) (*models.UserTOTP, error)

	// Create creates a new TOTP config
	Create(ctx context.Context, totp *models.UserTOTP) error

	// Update updates a TOTP config
	Update(ctx context.Context, totp *models.UserTOTP) error

	// Delete deletes a TOTP config
	Delete(ctx context.Context, userID uuid.UUID) error
}

// GormTOTPRepository implements TOTPRepository using GORM.
type GormTOTPRepository struct {
	db *gorm.DB
}

// NewTOTPRepository creates a new GORM-based TOTP repository.
func NewTOTPRepository(db *gorm.DB) TOTPRepository {
	return &GormTOTPRepository{db: db}
}

// GetByUserID retrieves TOTP config for a user.
func (r *GormTOTPRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.UserTOTP, error) {
	var totp models.UserTOTP
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&totp).Error
	if err != nil {
		return nil, err
	}
	return &totp, nil
}

// Create creates a new TOTP config.
func (r *GormTOTPRepository) Create(ctx context.Context, totp *models.UserTOTP) error {
	return r.db.WithContext(ctx).Create(totp).Error
}

// Update updates a TOTP config.
func (r *GormTOTPRepository) Update(ctx context.Context, totp *models.UserTOTP) error {
	return r.db.WithContext(ctx).Save(totp).Error
}

// Delete deletes a TOTP config.
func (r *GormTOTPRepository) Delete(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&models.UserTOTP{}).Error
}
