// Package repository provides data access layer implementations.
package repository

import (
	"context"
	"fmt"
	"savvy/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SessionRepository defines the interface for session data access.
type SessionRepository interface {
	// Store operations (used by PGStore)
	FindByTokenHash(ctx context.Context, tokenHash string) (*models.Session, error)
	Create(ctx context.Context, session *models.Session) error
	Update(ctx context.Context, session *models.Session) error
	DeleteByTokenHash(ctx context.Context, tokenHash string) error

	// Session management operations (used by SessionService)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.Session, error)
	DeleteByID(ctx context.Context, sessionID, userID uuid.UUID) error
	DeleteAllByUserIDExcept(ctx context.Context, userID uuid.UUID, exceptTokenHash string) (int64, error)
	DeleteAllByUserID(ctx context.Context, userID uuid.UUID) (int64, error)

	// Cleanup + Metrics
	DeleteExpired(ctx context.Context) (int64, error)
	CountActive(ctx context.Context) (int64, error)
}

// sessionRepository implements SessionRepository.
type sessionRepository struct {
	db *gorm.DB
}

// NewSessionRepository creates a new session repository.
func NewSessionRepository(db *gorm.DB) SessionRepository {
	return &sessionRepository{db: db}
}

// FindByTokenHash retrieves a non-expired session by its token hash.
func (r *sessionRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*models.Session, error) {
	var session models.Session
	err := r.db.WithContext(ctx).
		Where("token_hash = ? AND expires_at > ?", tokenHash, time.Now()).
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// Create creates a new session.
func (r *sessionRepository) Create(ctx context.Context, session *models.Session) error {
	return r.db.WithContext(ctx).Create(session).Error
}

// Update updates an existing session.
func (r *sessionRepository) Update(ctx context.Context, session *models.Session) error {
	return r.db.WithContext(ctx).Save(session).Error
}

// DeleteByTokenHash deletes a session by its token hash.
func (r *sessionRepository) DeleteByTokenHash(ctx context.Context, tokenHash string) error {
	return r.db.WithContext(ctx).Where("token_hash = ?", tokenHash).Delete(&models.Session{}).Error
}

// GetByUserID retrieves all active sessions for a user, ordered by last active time.
func (r *sessionRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.Session, error) {
	var sessions []models.Session
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND expires_at > ?", userID, time.Now()).
		Order("last_active_at DESC").
		Find(&sessions).Error
	return sessions, err
}

// DeleteByID deletes a specific session by ID, scoped to a user for security.
func (r *sessionRepository) DeleteByID(ctx context.Context, sessionID, userID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", sessionID, userID).
		Delete(&models.Session{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// DeleteAllByUserIDExcept deletes all sessions for a user except the one with the given token hash.
// exceptTokenHash must not be empty to prevent accidental deletion of all sessions.
func (r *sessionRepository) DeleteAllByUserIDExcept(ctx context.Context, userID uuid.UUID, exceptTokenHash string) (int64, error) {
	if exceptTokenHash == "" {
		return 0, fmt.Errorf("exceptTokenHash must not be empty")
	}
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND token_hash != ?", userID, exceptTokenHash).
		Delete(&models.Session{})
	return result.RowsAffected, result.Error
}

// DeleteAllByUserID deletes all sessions for a user.
func (r *sessionRepository) DeleteAllByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&models.Session{})
	return result.RowsAffected, result.Error
}

// DeleteExpired removes all expired sessions.
func (r *sessionRepository) DeleteExpired(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("expires_at <= ?", time.Now()).
		Delete(&models.Session{})
	return result.RowsAffected, result.Error
}

// CountActive returns the number of active (non-expired) sessions.
func (r *sessionRepository) CountActive(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Session{}).
		Where("expires_at > ?", time.Now()).
		Count(&count).Error
	return count, err
}
