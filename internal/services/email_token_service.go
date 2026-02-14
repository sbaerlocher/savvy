// Package services contains business logic.
package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"savvy/internal/models"
	"savvy/internal/repository"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Token expiry durations
const (
	VerificationTokenExpiry = 24 * time.Hour
	PasswordResetExpiry     = 1 * time.Hour
	UnsubscribeTokenExpiry  = 7 * 24 * time.Hour // 7 days
)

// Errors for email token operations
var (
	ErrTokenNotFound = errors.New("token not found")
	ErrTokenExpired  = errors.New("token has expired")
	ErrTokenUsed     = errors.New("token has already been used")
)

// EmailTokenServiceInterface defines the contract for email token operations.
type EmailTokenServiceInterface interface {
	// CreateVerificationToken generates and stores a new email verification token.
	// Returns the plain token to be sent to the user.
	CreateVerificationToken(ctx context.Context, userID uuid.UUID) (string, error)

	// VerifyEmail validates the token and marks the user's email as verified.
	VerifyEmail(ctx context.Context, token string) error

	// CreatePasswordResetToken generates and stores a new password reset token.
	// Returns the plain token to be sent to the user.
	CreatePasswordResetToken(ctx context.Context, userID uuid.UUID) (string, error)

	// ConsumePasswordResetToken validates and consumes the token, returning the user.
	ConsumePasswordResetToken(ctx context.Context, token string) (*models.User, error)

	// CreateUnsubscribeToken generates a token for one-click email unsubscribe (share/transfer).
	// Returns the plain token to be included in notification emails.
	CreateUnsubscribeToken(ctx context.Context, userID uuid.UUID) (string, error)

	// UnsubscribeNotifications validates the token and disables share/transfer email notifications.
	UnsubscribeNotifications(ctx context.Context, token string) error

	// CreateUnsubscribeReminderToken generates a token for one-click expiry reminder unsubscribe.
	// Returns the plain token to be included in reminder emails.
	CreateUnsubscribeReminderToken(ctx context.Context, userID uuid.UUID) (string, error)

	// UnsubscribeReminders validates the token and disables expiry reminder email notifications.
	UnsubscribeReminders(ctx context.Context, token string) error

	// CleanupExpiredTokens removes all expired tokens from the database.
	CleanupExpiredTokens(ctx context.Context) error
}

// EmailTokenService handles email token operations.
type EmailTokenService struct {
	tokenRepo repository.EmailTokenRepository
	userRepo  repository.UserRepository
}

// NewEmailTokenService creates a new email token service.
func NewEmailTokenService(tokenRepo repository.EmailTokenRepository, userRepo repository.UserRepository) *EmailTokenService {
	return &EmailTokenService{
		tokenRepo: tokenRepo,
		userRepo:  userRepo,
	}
}

// CreateVerificationToken generates a verification token for the user.
func (s *EmailTokenService) CreateVerificationToken(ctx context.Context, userID uuid.UUID) (string, error) {
	// Delete any existing unused verification tokens for this user
	if err := s.tokenRepo.DeleteByUserAndType(ctx, userID, models.TokenTypeEmailVerification); err != nil {
		return "", fmt.Errorf("failed to delete existing tokens: %w", err)
	}

	return s.createToken(ctx, userID, models.TokenTypeEmailVerification, VerificationTokenExpiry)
}

// VerifyEmail validates the token and marks the user's email as verified.
func (s *EmailTokenService) VerifyEmail(ctx context.Context, token string) error {
	tokenHash := hashToken(token)

	emailToken, err := s.tokenRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTokenNotFound
		}
		return fmt.Errorf("failed to look up token: %w", err)
	}

	if emailToken.TokenType != models.TokenTypeEmailVerification {
		return ErrTokenNotFound
	}

	if emailToken.IsUsed() {
		return ErrTokenUsed
	}

	if emailToken.IsExpired() {
		return ErrTokenExpired
	}

	// Mark token as used
	if err := s.tokenRepo.MarkUsed(ctx, emailToken.ID); err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}

	// Update user's email verification status
	user, err := s.userRepo.GetByID(ctx, emailToken.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	now := time.Now()
	user.EmailVerified = true
	user.EmailVerifiedAt = &now
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// CreatePasswordResetToken generates a password reset token for the user.
func (s *EmailTokenService) CreatePasswordResetToken(ctx context.Context, userID uuid.UUID) (string, error) {
	// Delete any existing unused reset tokens for this user
	if err := s.tokenRepo.DeleteByUserAndType(ctx, userID, models.TokenTypePasswordReset); err != nil {
		return "", fmt.Errorf("failed to delete existing tokens: %w", err)
	}

	return s.createToken(ctx, userID, models.TokenTypePasswordReset, PasswordResetExpiry)
}

// ConsumePasswordResetToken validates the token, marks it as used, and returns the user.
func (s *EmailTokenService) ConsumePasswordResetToken(ctx context.Context, token string) (*models.User, error) {
	tokenHash := hashToken(token)

	emailToken, err := s.tokenRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTokenNotFound
		}
		return nil, fmt.Errorf("failed to look up token: %w", err)
	}

	if emailToken.TokenType != models.TokenTypePasswordReset {
		return nil, ErrTokenNotFound
	}

	if emailToken.IsUsed() {
		return nil, ErrTokenUsed
	}

	if emailToken.IsExpired() {
		return nil, ErrTokenExpired
	}

	// Mark token as used
	if err := s.tokenRepo.MarkUsed(ctx, emailToken.ID); err != nil {
		return nil, fmt.Errorf("failed to mark token as used: %w", err)
	}

	// Get the user
	user, err := s.userRepo.GetByID(ctx, emailToken.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// CreateUnsubscribeToken generates a token for one-click email unsubscribe.
// Unlike other token types, this does NOT delete existing tokens (multiple emails = multiple valid tokens).
func (s *EmailTokenService) CreateUnsubscribeToken(ctx context.Context, userID uuid.UUID) (string, error) {
	return s.createToken(ctx, userID, models.TokenTypeUnsubscribeNotification, UnsubscribeTokenExpiry)
}

// UnsubscribeNotifications validates the token and disables share/transfer email notifications.
func (s *EmailTokenService) UnsubscribeNotifications(ctx context.Context, token string) error {
	tokenHash := hashToken(token)

	emailToken, err := s.tokenRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTokenNotFound
		}
		return fmt.Errorf("failed to look up token: %w", err)
	}

	if emailToken.TokenType != models.TokenTypeUnsubscribeNotification {
		return ErrTokenNotFound
	}

	if emailToken.IsUsed() {
		return ErrTokenUsed
	}

	if emailToken.IsExpired() {
		return ErrTokenExpired
	}

	// Mark token as used
	if err := s.tokenRepo.MarkUsed(ctx, emailToken.ID); err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}

	// Disable share/transfer email notifications for the user
	user, err := s.userRepo.GetByID(ctx, emailToken.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	user.EmailSharingEnabled = false
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// CreateUnsubscribeReminderToken generates a token for one-click expiry reminder unsubscribe.
// Unlike other token types, this does NOT delete existing tokens (multiple emails = multiple valid tokens).
func (s *EmailTokenService) CreateUnsubscribeReminderToken(ctx context.Context, userID uuid.UUID) (string, error) {
	return s.createToken(ctx, userID, models.TokenTypeUnsubscribeReminders, UnsubscribeTokenExpiry)
}

// UnsubscribeReminders validates the token and disables expiry reminder email notifications.
func (s *EmailTokenService) UnsubscribeReminders(ctx context.Context, token string) error {
	tokenHash := hashToken(token)

	emailToken, err := s.tokenRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTokenNotFound
		}
		return fmt.Errorf("failed to look up token: %w", err)
	}

	if emailToken.TokenType != models.TokenTypeUnsubscribeReminders {
		return ErrTokenNotFound
	}

	if emailToken.IsUsed() {
		return ErrTokenUsed
	}

	if emailToken.IsExpired() {
		return ErrTokenExpired
	}

	// Mark token as used
	if err := s.tokenRepo.MarkUsed(ctx, emailToken.ID); err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}

	// Disable expiry reminder email notifications for the user
	user, err := s.userRepo.GetByID(ctx, emailToken.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	user.EmailRemindersEnabled = false
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// CleanupExpiredTokens removes all expired tokens.
func (s *EmailTokenService) CleanupExpiredTokens(ctx context.Context) error {
	return s.tokenRepo.DeleteExpiredTokens(ctx)
}

// createToken generates a cryptographically secure token, stores its hash, and returns the plain token.
func (s *EmailTokenService) createToken(ctx context.Context, userID uuid.UUID, tokenType string, expiry time.Duration) (string, error) {
	// Generate 32 random bytes
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}

	// Encode as hex string (64 chars)
	plainToken := hex.EncodeToString(randomBytes)

	// Hash the token for storage
	tokenHash := hashToken(plainToken)

	emailToken := &models.EmailToken{
		UserID:    userID,
		TokenHash: tokenHash,
		TokenType: tokenType,
		ExpiresAt: time.Now().Add(expiry),
	}

	if err := s.tokenRepo.Create(ctx, emailToken); err != nil {
		return "", fmt.Errorf("failed to store token: %w", err)
	}

	return plainToken, nil
}

// hashToken computes the SHA-256 hash of a plain token.
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
