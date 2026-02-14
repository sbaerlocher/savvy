// Package repository defines data access interfaces.
package repository

import (
	"context"
	"savvy/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EmailTokenRepository defines the interface for email token data access.
type EmailTokenRepository interface {
	// Create creates a new email token
	Create(ctx context.Context, token *models.EmailToken) error

	// GetByTokenHash retrieves a token by its SHA-256 hash
	GetByTokenHash(ctx context.Context, tokenHash string) (*models.EmailToken, error)

	// MarkUsed marks a token as used
	MarkUsed(ctx context.Context, id uuid.UUID) error

	// DeleteByUserAndType deletes all unused tokens for a user and type
	DeleteByUserAndType(ctx context.Context, userID uuid.UUID, tokenType string) error

	// DeleteExpiredTokens removes all expired tokens from the database
	DeleteExpiredTokens(ctx context.Context) error
}

// GormEmailTokenRepository is the GORM implementation of EmailTokenRepository.
type GormEmailTokenRepository struct {
	db *gorm.DB
}

// NewEmailTokenRepository creates a new GORM-based email token repository.
func NewEmailTokenRepository(db *gorm.DB) EmailTokenRepository {
	return &GormEmailTokenRepository{db: db}
}

// Create creates a new email token.
func (r *GormEmailTokenRepository) Create(ctx context.Context, token *models.EmailToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

// GetByTokenHash retrieves a token by its SHA-256 hash.
func (r *GormEmailTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*models.EmailToken, error) {
	var token models.EmailToken
	err := r.db.WithContext(ctx).Where("token_hash = ?", tokenHash).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// MarkUsed marks a token as used by setting used_at to now.
func (r *GormEmailTokenRepository) MarkUsed(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.EmailToken{}).
		Where("id = ?", id).
		Update("used_at", gorm.Expr("CURRENT_TIMESTAMP")).Error
}

// DeleteByUserAndType deletes all unused tokens for a specific user and type.
func (r *GormEmailTokenRepository) DeleteByUserAndType(ctx context.Context, userID uuid.UUID, tokenType string) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND token_type = ? AND used_at IS NULL", userID, tokenType).
		Delete(&models.EmailToken{}).Error
}

// DeleteExpiredTokens removes all expired tokens.
func (r *GormEmailTokenRepository) DeleteExpiredTokens(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < CURRENT_TIMESTAMP").
		Delete(&models.EmailToken{}).Error
}
