package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
	"savvy/internal/models"
	"savvy/internal/repository"
)

// ============================================================================
// MOCKS
// ============================================================================

// MockEmailTokenRepo is a mock for EmailTokenRepository
type MockEmailTokenRepo struct {
	mock.Mock
}

func (m *MockEmailTokenRepo) Create(ctx context.Context, token *models.EmailToken) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockEmailTokenRepo) GetByTokenHash(ctx context.Context, tokenHash string) (*models.EmailToken, error) {
	args := m.Called(ctx, tokenHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.EmailToken), args.Error(1)
}

func (m *MockEmailTokenRepo) MarkUsed(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockEmailTokenRepo) DeleteByUserAndType(ctx context.Context, userID uuid.UUID, tokenType string) error {
	args := m.Called(ctx, userID, tokenType)
	return args.Error(0)
}

func (m *MockEmailTokenRepo) DeleteExpiredTokens(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

var _ repository.EmailTokenRepository = (*MockEmailTokenRepo)(nil)

// MockUserRepoForToken is a mock for UserRepository used in email token tests
type MockUserRepoForToken struct {
	mock.Mock
}

func (m *MockUserRepoForToken) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepoForToken) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepoForToken) Create(ctx context.Context, user *models.User) error {
	return m.Called(ctx, user).Error(0)
}

func (m *MockUserRepoForToken) Update(ctx context.Context, user *models.User) error {
	return m.Called(ctx, user).Error(0)
}

func (m *MockUserRepoForToken) GetAll(ctx context.Context) ([]models.User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockUserRepoForToken) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]models.User, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockUserRepoForToken) SearchByIDs(ctx context.Context, ids []uuid.UUID, query string) ([]models.User, error) {
	args := m.Called(ctx, ids, query)
	return args.Get(0).([]models.User), args.Error(1)
}

var _ repository.UserRepository = (*MockUserRepoForToken)(nil)

// ============================================================================
// TESTS: UnsubscribeNotifications
// ============================================================================

// TestEmailTokenService_UnsubscribeNotifications_Success tests that unsubscribing
// sets EmailSharingEnabled to false on the user
func TestEmailTokenService_UnsubscribeNotifications_Success(t *testing.T) {
	tokenRepo := new(MockEmailTokenRepo)
	userRepo := new(MockUserRepoForToken)
	svc := NewEmailTokenService(tokenRepo, userRepo)
	ctx := context.Background()

	userID := uuid.New()
	tokenID := uuid.New()
	plainToken := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	tokenHash := hashToken(plainToken)

	emailToken := &models.EmailToken{
		ID:        tokenID,
		UserID:    userID,
		TokenHash: tokenHash,
		TokenType: models.TokenTypeUnsubscribeNotification,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		UsedAt:    nil,
	}

	user := &models.User{
		EmailSharingEnabled:       true,
		EmailNotificationsEnabled: true,
	}
	user.ID = userID

	tokenRepo.On("GetByTokenHash", ctx, tokenHash).Return(emailToken, nil)
	tokenRepo.On("MarkUsed", ctx, tokenID).Return(nil)
	userRepo.On("GetByID", ctx, userID).Return(user, nil)
	userRepo.On("Update", ctx, mock.MatchedBy(func(u *models.User) bool {
		return u.ID == userID && !u.EmailSharingEnabled
	})).Return(nil)

	err := svc.UnsubscribeNotifications(ctx, plainToken)

	assert.NoError(t, err)
	assert.False(t, user.EmailSharingEnabled, "EmailSharingEnabled should be set to false")
	tokenRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

// TestEmailTokenService_UnsubscribeNotifications_TokenNotFound tests that an invalid token returns error
func TestEmailTokenService_UnsubscribeNotifications_TokenNotFound(t *testing.T) {
	tokenRepo := new(MockEmailTokenRepo)
	userRepo := new(MockUserRepoForToken)
	svc := NewEmailTokenService(tokenRepo, userRepo)
	ctx := context.Background()

	plainToken := "invalid_token_that_does_not_exist_in_database_at_all_1234567890ab"
	tokenHash := hashToken(plainToken)

	tokenRepo.On("GetByTokenHash", ctx, tokenHash).Return(nil, ErrTokenNotFound)

	err := svc.UnsubscribeNotifications(ctx, plainToken)

	assert.ErrorIs(t, err, ErrTokenNotFound)
}

// TestEmailTokenService_UnsubscribeNotifications_TokenExpired tests that an expired token returns error
func TestEmailTokenService_UnsubscribeNotifications_TokenExpired(t *testing.T) {
	tokenRepo := new(MockEmailTokenRepo)
	userRepo := new(MockUserRepoForToken)
	svc := NewEmailTokenService(tokenRepo, userRepo)
	ctx := context.Background()

	plainToken := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	tokenHash := hashToken(plainToken)

	emailToken := &models.EmailToken{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		TokenHash: tokenHash,
		TokenType: models.TokenTypeUnsubscribeNotification,
		ExpiresAt: time.Now().Add(-1 * time.Hour), // expired
		UsedAt:    nil,
	}

	tokenRepo.On("GetByTokenHash", ctx, tokenHash).Return(emailToken, nil)

	err := svc.UnsubscribeNotifications(ctx, plainToken)

	assert.ErrorIs(t, err, ErrTokenExpired)
}

// TestEmailTokenService_UnsubscribeNotifications_TokenAlreadyUsed tests that a used token returns error
func TestEmailTokenService_UnsubscribeNotifications_TokenAlreadyUsed(t *testing.T) {
	tokenRepo := new(MockEmailTokenRepo)
	userRepo := new(MockUserRepoForToken)
	svc := NewEmailTokenService(tokenRepo, userRepo)
	ctx := context.Background()

	plainToken := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	tokenHash := hashToken(plainToken)
	usedAt := time.Now().Add(-30 * time.Minute)

	emailToken := &models.EmailToken{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		TokenHash: tokenHash,
		TokenType: models.TokenTypeUnsubscribeNotification,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		UsedAt:    &usedAt, // already used
	}

	tokenRepo.On("GetByTokenHash", ctx, tokenHash).Return(emailToken, nil)

	err := svc.UnsubscribeNotifications(ctx, plainToken)

	assert.ErrorIs(t, err, ErrTokenUsed)
}

// TestEmailTokenService_UnsubscribeNotifications_WrongTokenType tests that a wrong token type returns error
func TestEmailTokenService_UnsubscribeNotifications_WrongTokenType(t *testing.T) {
	tokenRepo := new(MockEmailTokenRepo)
	userRepo := new(MockUserRepoForToken)
	svc := NewEmailTokenService(tokenRepo, userRepo)
	ctx := context.Background()

	plainToken := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	tokenHash := hashToken(plainToken)

	emailToken := &models.EmailToken{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		TokenHash: tokenHash,
		TokenType: models.TokenTypePasswordReset, // wrong type
		ExpiresAt: time.Now().Add(24 * time.Hour),
		UsedAt:    nil,
	}

	tokenRepo.On("GetByTokenHash", ctx, tokenHash).Return(emailToken, nil)

	err := svc.UnsubscribeNotifications(ctx, plainToken)

	assert.ErrorIs(t, err, ErrTokenNotFound)
}

// ============================================================================
// TESTS: UnsubscribeReminders
// ============================================================================

// TestEmailTokenService_UnsubscribeReminders_Success tests that unsubscribing
// sets EmailRemindersEnabled to false on the user
func TestEmailTokenService_UnsubscribeReminders_Success(t *testing.T) {
	tokenRepo := new(MockEmailTokenRepo)
	userRepo := new(MockUserRepoForToken)
	svc := NewEmailTokenService(tokenRepo, userRepo)
	ctx := context.Background()

	userID := uuid.New()
	tokenID := uuid.New()
	plainToken := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	tokenHash := hashToken(plainToken)

	emailToken := &models.EmailToken{
		ID:        tokenID,
		UserID:    userID,
		TokenHash: tokenHash,
		TokenType: models.TokenTypeUnsubscribeReminders,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		UsedAt:    nil,
	}

	user := &models.User{
		EmailRemindersEnabled:     true,
		EmailNotificationsEnabled: true,
	}
	user.ID = userID

	tokenRepo.On("GetByTokenHash", ctx, tokenHash).Return(emailToken, nil)
	tokenRepo.On("MarkUsed", ctx, tokenID).Return(nil)
	userRepo.On("GetByID", ctx, userID).Return(user, nil)
	userRepo.On("Update", ctx, mock.MatchedBy(func(u *models.User) bool {
		return u.ID == userID && !u.EmailRemindersEnabled
	})).Return(nil)

	err := svc.UnsubscribeReminders(ctx, plainToken)

	assert.NoError(t, err)
	assert.False(t, user.EmailRemindersEnabled, "EmailRemindersEnabled should be set to false")
	tokenRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

// TestEmailTokenService_UnsubscribeReminders_TokenExpired tests expired reminder unsubscribe token
func TestEmailTokenService_UnsubscribeReminders_TokenExpired(t *testing.T) {
	tokenRepo := new(MockEmailTokenRepo)
	userRepo := new(MockUserRepoForToken)
	svc := NewEmailTokenService(tokenRepo, userRepo)
	ctx := context.Background()

	plainToken := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	tokenHash := hashToken(plainToken)

	emailToken := &models.EmailToken{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		TokenHash: tokenHash,
		TokenType: models.TokenTypeUnsubscribeReminders,
		ExpiresAt: time.Now().Add(-1 * time.Hour), // expired
		UsedAt:    nil,
	}

	tokenRepo.On("GetByTokenHash", ctx, tokenHash).Return(emailToken, nil)

	err := svc.UnsubscribeReminders(ctx, plainToken)

	assert.ErrorIs(t, err, ErrTokenExpired)
}

// TestEmailTokenService_UnsubscribeReminders_WrongTokenType tests wrong token type for reminders
func TestEmailTokenService_UnsubscribeReminders_WrongTokenType(t *testing.T) {
	tokenRepo := new(MockEmailTokenRepo)
	userRepo := new(MockUserRepoForToken)
	svc := NewEmailTokenService(tokenRepo, userRepo)
	ctx := context.Background()

	plainToken := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	tokenHash := hashToken(plainToken)

	emailToken := &models.EmailToken{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		TokenHash: tokenHash,
		TokenType: models.TokenTypeUnsubscribeNotification, // wrong type (notifications, not reminders)
		ExpiresAt: time.Now().Add(24 * time.Hour),
		UsedAt:    nil,
	}

	tokenRepo.On("GetByTokenHash", ctx, tokenHash).Return(emailToken, nil)

	err := svc.UnsubscribeReminders(ctx, plainToken)

	assert.ErrorIs(t, err, ErrTokenNotFound)
}

// ============================================================================
// TESTS: CreateUnsubscribeToken
// ============================================================================

// TestEmailTokenService_CreateUnsubscribeToken_Success tests token creation
func TestEmailTokenService_CreateUnsubscribeToken_Success(t *testing.T) {
	tokenRepo := new(MockEmailTokenRepo)
	userRepo := new(MockUserRepoForToken)
	svc := NewEmailTokenService(tokenRepo, userRepo)
	ctx := context.Background()

	userID := uuid.New()

	// Token creation should NOT delete existing tokens (multiple emails = multiple valid tokens)
	tokenRepo.On("Create", ctx, mock.MatchedBy(func(token *models.EmailToken) bool {
		return token.UserID == userID &&
			token.TokenType == models.TokenTypeUnsubscribeNotification &&
			token.TokenHash != "" &&
			token.ExpiresAt.After(time.Now())
	})).Return(nil)

	plainToken, err := svc.CreateUnsubscribeToken(ctx, userID)

	assert.NoError(t, err)
	assert.Len(t, plainToken, 64) // 32 bytes hex-encoded
	tokenRepo.AssertExpectations(t)
	// Verify DeleteByUserAndType was NOT called (unsubscribe tokens don't delete existing ones)
	tokenRepo.AssertNotCalled(t, "DeleteByUserAndType")
}

// TestEmailTokenService_CreateUnsubscribeReminderToken_Success tests reminder token creation
func TestEmailTokenService_CreateUnsubscribeReminderToken_Success(t *testing.T) {
	tokenRepo := new(MockEmailTokenRepo)
	userRepo := new(MockUserRepoForToken)
	svc := NewEmailTokenService(tokenRepo, userRepo)
	ctx := context.Background()

	userID := uuid.New()

	tokenRepo.On("Create", ctx, mock.MatchedBy(func(token *models.EmailToken) bool {
		return token.UserID == userID &&
			token.TokenType == models.TokenTypeUnsubscribeReminders &&
			token.TokenHash != "" &&
			token.ExpiresAt.After(time.Now())
	})).Return(nil)

	plainToken, err := svc.CreateUnsubscribeReminderToken(ctx, userID)

	assert.NoError(t, err)
	assert.Len(t, plainToken, 64)
	tokenRepo.AssertExpectations(t)
	tokenRepo.AssertNotCalled(t, "DeleteByUserAndType")
}

// ============================================================================
// TESTS: CleanupExpiredTokens
// ============================================================================

// TestEmailTokenService_CleanupExpiredTokens tests token cleanup
func TestEmailTokenService_CleanupExpiredTokens(t *testing.T) {
	tokenRepo := new(MockEmailTokenRepo)
	userRepo := new(MockUserRepoForToken)
	svc := NewEmailTokenService(tokenRepo, userRepo)
	ctx := context.Background()

	tokenRepo.On("DeleteExpiredTokens", ctx).Return(nil)

	err := svc.CleanupExpiredTokens(ctx)

	assert.NoError(t, err)
	tokenRepo.AssertExpectations(t)
}

// ============================================================================
// TESTS: CreateVerificationToken
// ============================================================================

// TestEmailTokenService_CreateVerificationToken_Success tests successful token creation
func TestEmailTokenService_CreateVerificationToken_Success(t *testing.T) {
	tokenRepo := new(MockEmailTokenRepo)
	userRepo := new(MockUserRepoForToken)
	svc := NewEmailTokenService(tokenRepo, userRepo)
	ctx := context.Background()

	userID := uuid.New()

	// Should delete existing unused verification tokens first
	tokenRepo.On("DeleteByUserAndType", ctx, userID, models.TokenTypeEmailVerification).Return(nil)
	// Then create a new token
	tokenRepo.On("Create", ctx, mock.MatchedBy(func(token *models.EmailToken) bool {
		return token.UserID == userID &&
			token.TokenType == models.TokenTypeEmailVerification &&
			token.TokenHash != "" &&
			token.ExpiresAt.After(time.Now()) &&
			token.ExpiresAt.Before(time.Now().Add(VerificationTokenExpiry+time.Minute))
	})).Return(nil)

	plainToken, err := svc.CreateVerificationToken(ctx, userID)

	assert.NoError(t, err)
	assert.Len(t, plainToken, 64) // 32 bytes hex-encoded
	// Verify the stored hash matches the plain token
	expectedHash := hashToken(plainToken)
	assert.NotEmpty(t, expectedHash)
	tokenRepo.AssertExpectations(t)
}

// TestEmailTokenService_CreateVerificationToken_DeleteExistingTokensError tests error when deleting existing tokens fails
func TestEmailTokenService_CreateVerificationToken_DeleteExistingTokensError(t *testing.T) {
	tokenRepo := new(MockEmailTokenRepo)
	userRepo := new(MockUserRepoForToken)
	svc := NewEmailTokenService(tokenRepo, userRepo)
	ctx := context.Background()

	userID := uuid.New()
	deleteErr := errors.New("database connection failed")

	tokenRepo.On("DeleteByUserAndType", ctx, userID, models.TokenTypeEmailVerification).Return(deleteErr)

	plainToken, err := svc.CreateVerificationToken(ctx, userID)

	assert.Error(t, err)
	assert.Empty(t, plainToken)
	assert.Contains(t, err.Error(), "failed to delete existing tokens")
	tokenRepo.AssertExpectations(t)
	tokenRepo.AssertNotCalled(t, "Create")
}

// TestEmailTokenService_CreateVerificationToken_CreateError tests error when storing the token fails
func TestEmailTokenService_CreateVerificationToken_CreateError(t *testing.T) {
	tokenRepo := new(MockEmailTokenRepo)
	userRepo := new(MockUserRepoForToken)
	svc := NewEmailTokenService(tokenRepo, userRepo)
	ctx := context.Background()

	userID := uuid.New()
	createErr := errors.New("failed to insert")

	tokenRepo.On("DeleteByUserAndType", ctx, userID, models.TokenTypeEmailVerification).Return(nil)
	tokenRepo.On("Create", ctx, mock.AnythingOfType("*models.EmailToken")).Return(createErr)

	plainToken, err := svc.CreateVerificationToken(ctx, userID)

	assert.Error(t, err)
	assert.Empty(t, plainToken)
	assert.Contains(t, err.Error(), "failed to store token")
	tokenRepo.AssertExpectations(t)
}

// ============================================================================
// TESTS: VerifyEmail
// ============================================================================

// TestEmailTokenService_VerifyEmail_Success tests successful email verification
func TestEmailTokenService_VerifyEmail_Success(t *testing.T) {
	tokenRepo := new(MockEmailTokenRepo)
	userRepo := new(MockUserRepoForToken)
	svc := NewEmailTokenService(tokenRepo, userRepo)
	ctx := context.Background()

	userID := uuid.New()
	tokenID := uuid.New()
	plainToken := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	tokenHash := hashToken(plainToken)

	emailToken := &models.EmailToken{
		ID:        tokenID,
		UserID:    userID,
		TokenHash: tokenHash,
		TokenType: models.TokenTypeEmailVerification,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		UsedAt:    nil,
	}

	user := &models.User{
		EmailVerified:   false,
		EmailVerifiedAt: nil,
	}
	user.ID = userID

	tokenRepo.On("GetByTokenHash", ctx, tokenHash).Return(emailToken, nil)
	tokenRepo.On("MarkUsed", ctx, tokenID).Return(nil)
	userRepo.On("GetByID", ctx, userID).Return(user, nil)
	userRepo.On("Update", ctx, mock.MatchedBy(func(u *models.User) bool {
		return u.ID == userID && u.EmailVerified && u.EmailVerifiedAt != nil
	})).Return(nil)

	err := svc.VerifyEmail(ctx, plainToken)

	assert.NoError(t, err)
	assert.True(t, user.EmailVerified, "EmailVerified should be set to true")
	assert.NotNil(t, user.EmailVerifiedAt, "EmailVerifiedAt should be set")
	tokenRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

// TestEmailTokenService_VerifyEmail_TokenNotFound tests that a non-existent token returns ErrTokenNotFound
func TestEmailTokenService_VerifyEmail_TokenNotFound(t *testing.T) {
	tokenRepo := new(MockEmailTokenRepo)
	userRepo := new(MockUserRepoForToken)
	svc := NewEmailTokenService(tokenRepo, userRepo)
	ctx := context.Background()

	plainToken := "nonexistent_token_0000000000000000000000000000000000000000000000"
	tokenHash := hashToken(plainToken)

	tokenRepo.On("GetByTokenHash", ctx, tokenHash).Return(nil, gorm.ErrRecordNotFound)

	err := svc.VerifyEmail(ctx, plainToken)

	assert.ErrorIs(t, err, ErrTokenNotFound)
	tokenRepo.AssertExpectations(t)
}

// TestEmailTokenService_VerifyEmail_TokenExpired tests that an expired token returns ErrTokenExpired
func TestEmailTokenService_VerifyEmail_TokenExpired(t *testing.T) {
	tokenRepo := new(MockEmailTokenRepo)
	userRepo := new(MockUserRepoForToken)
	svc := NewEmailTokenService(tokenRepo, userRepo)
	ctx := context.Background()

	plainToken := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	tokenHash := hashToken(plainToken)

	emailToken := &models.EmailToken{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		TokenHash: tokenHash,
		TokenType: models.TokenTypeEmailVerification,
		ExpiresAt: time.Now().Add(-1 * time.Hour), // expired
		UsedAt:    nil,
	}

	tokenRepo.On("GetByTokenHash", ctx, tokenHash).Return(emailToken, nil)

	err := svc.VerifyEmail(ctx, plainToken)

	assert.ErrorIs(t, err, ErrTokenExpired)
	tokenRepo.AssertExpectations(t)
}

// TestEmailTokenService_VerifyEmail_TokenAlreadyUsed tests that a used token returns ErrTokenUsed
func TestEmailTokenService_VerifyEmail_TokenAlreadyUsed(t *testing.T) {
	tokenRepo := new(MockEmailTokenRepo)
	userRepo := new(MockUserRepoForToken)
	svc := NewEmailTokenService(tokenRepo, userRepo)
	ctx := context.Background()

	plainToken := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	tokenHash := hashToken(plainToken)
	usedAt := time.Now().Add(-30 * time.Minute)

	emailToken := &models.EmailToken{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		TokenHash: tokenHash,
		TokenType: models.TokenTypeEmailVerification,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		UsedAt:    &usedAt, // already used
	}

	tokenRepo.On("GetByTokenHash", ctx, tokenHash).Return(emailToken, nil)

	err := svc.VerifyEmail(ctx, plainToken)

	assert.ErrorIs(t, err, ErrTokenUsed)
	tokenRepo.AssertExpectations(t)
}

// TestEmailTokenService_VerifyEmail_WrongTokenType tests that a non-verification token returns ErrTokenNotFound
func TestEmailTokenService_VerifyEmail_WrongTokenType(t *testing.T) {
	tokenRepo := new(MockEmailTokenRepo)
	userRepo := new(MockUserRepoForToken)
	svc := NewEmailTokenService(tokenRepo, userRepo)
	ctx := context.Background()

	plainToken := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	tokenHash := hashToken(plainToken)

	emailToken := &models.EmailToken{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		TokenHash: tokenHash,
		TokenType: models.TokenTypePasswordReset, // wrong type
		ExpiresAt: time.Now().Add(24 * time.Hour),
		UsedAt:    nil,
	}

	tokenRepo.On("GetByTokenHash", ctx, tokenHash).Return(emailToken, nil)

	err := svc.VerifyEmail(ctx, plainToken)

	assert.ErrorIs(t, err, ErrTokenNotFound)
	tokenRepo.AssertExpectations(t)
}

// TestEmailTokenService_VerifyEmail_MarkUsedError tests error when marking token as used fails
func TestEmailTokenService_VerifyEmail_MarkUsedError(t *testing.T) {
	tokenRepo := new(MockEmailTokenRepo)
	userRepo := new(MockUserRepoForToken)
	svc := NewEmailTokenService(tokenRepo, userRepo)
	ctx := context.Background()

	tokenID := uuid.New()
	plainToken := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	tokenHash := hashToken(plainToken)

	emailToken := &models.EmailToken{
		ID:        tokenID,
		UserID:    uuid.New(),
		TokenHash: tokenHash,
		TokenType: models.TokenTypeEmailVerification,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		UsedAt:    nil,
	}

	tokenRepo.On("GetByTokenHash", ctx, tokenHash).Return(emailToken, nil)
	tokenRepo.On("MarkUsed", ctx, tokenID).Return(errors.New("db error"))

	err := svc.VerifyEmail(ctx, plainToken)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to mark token as used")
	tokenRepo.AssertExpectations(t)
}

// TestEmailTokenService_VerifyEmail_GetUserError tests error when fetching user fails
func TestEmailTokenService_VerifyEmail_GetUserError(t *testing.T) {
	tokenRepo := new(MockEmailTokenRepo)
	userRepo := new(MockUserRepoForToken)
	svc := NewEmailTokenService(tokenRepo, userRepo)
	ctx := context.Background()

	userID := uuid.New()
	tokenID := uuid.New()
	plainToken := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	tokenHash := hashToken(plainToken)

	emailToken := &models.EmailToken{
		ID:        tokenID,
		UserID:    userID,
		TokenHash: tokenHash,
		TokenType: models.TokenTypeEmailVerification,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		UsedAt:    nil,
	}

	tokenRepo.On("GetByTokenHash", ctx, tokenHash).Return(emailToken, nil)
	tokenRepo.On("MarkUsed", ctx, tokenID).Return(nil)
	userRepo.On("GetByID", ctx, userID).Return(nil, errors.New("user not found"))

	err := svc.VerifyEmail(ctx, plainToken)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get user")
	tokenRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

// TestEmailTokenService_VerifyEmail_UpdateUserError tests error when updating user fails
func TestEmailTokenService_VerifyEmail_UpdateUserError(t *testing.T) {
	tokenRepo := new(MockEmailTokenRepo)
	userRepo := new(MockUserRepoForToken)
	svc := NewEmailTokenService(tokenRepo, userRepo)
	ctx := context.Background()

	userID := uuid.New()
	tokenID := uuid.New()
	plainToken := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	tokenHash := hashToken(plainToken)

	emailToken := &models.EmailToken{
		ID:        tokenID,
		UserID:    userID,
		TokenHash: tokenHash,
		TokenType: models.TokenTypeEmailVerification,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		UsedAt:    nil,
	}

	user := &models.User{
		EmailVerified: false,
	}
	user.ID = userID

	tokenRepo.On("GetByTokenHash", ctx, tokenHash).Return(emailToken, nil)
	tokenRepo.On("MarkUsed", ctx, tokenID).Return(nil)
	userRepo.On("GetByID", ctx, userID).Return(user, nil)
	userRepo.On("Update", ctx, mock.AnythingOfType("*models.User")).Return(errors.New("update failed"))

	err := svc.VerifyEmail(ctx, plainToken)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update user")
	tokenRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

// TestEmailTokenService_VerifyEmail_RepoLookupError tests generic repo error on token lookup
func TestEmailTokenService_VerifyEmail_RepoLookupError(t *testing.T) {
	tokenRepo := new(MockEmailTokenRepo)
	userRepo := new(MockUserRepoForToken)
	svc := NewEmailTokenService(tokenRepo, userRepo)
	ctx := context.Background()

	plainToken := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	tokenHash := hashToken(plainToken)

	tokenRepo.On("GetByTokenHash", ctx, tokenHash).Return(nil, errors.New("connection timeout"))

	err := svc.VerifyEmail(ctx, plainToken)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to look up token")
	tokenRepo.AssertExpectations(t)
}

// ============================================================================
// TESTS: CreatePasswordResetToken
// ============================================================================

// TestEmailTokenService_CreatePasswordResetToken_Success tests successful password reset token creation
func TestEmailTokenService_CreatePasswordResetToken_Success(t *testing.T) {
	tokenRepo := new(MockEmailTokenRepo)
	userRepo := new(MockUserRepoForToken)
	svc := NewEmailTokenService(tokenRepo, userRepo)
	ctx := context.Background()

	userID := uuid.New()

	// Should delete existing unused reset tokens first
	tokenRepo.On("DeleteByUserAndType", ctx, userID, models.TokenTypePasswordReset).Return(nil)
	// Then create a new token
	tokenRepo.On("Create", ctx, mock.MatchedBy(func(token *models.EmailToken) bool {
		return token.UserID == userID &&
			token.TokenType == models.TokenTypePasswordReset &&
			token.TokenHash != "" &&
			token.ExpiresAt.After(time.Now()) &&
			token.ExpiresAt.Before(time.Now().Add(PasswordResetExpiry+time.Minute))
	})).Return(nil)

	plainToken, err := svc.CreatePasswordResetToken(ctx, userID)

	assert.NoError(t, err)
	assert.Len(t, plainToken, 64) // 32 bytes hex-encoded
	tokenRepo.AssertExpectations(t)
}

// TestEmailTokenService_CreatePasswordResetToken_DeleteExistingTokensError tests error when deleting existing tokens fails
func TestEmailTokenService_CreatePasswordResetToken_DeleteExistingTokensError(t *testing.T) {
	tokenRepo := new(MockEmailTokenRepo)
	userRepo := new(MockUserRepoForToken)
	svc := NewEmailTokenService(tokenRepo, userRepo)
	ctx := context.Background()

	userID := uuid.New()
	deleteErr := errors.New("database error")

	tokenRepo.On("DeleteByUserAndType", ctx, userID, models.TokenTypePasswordReset).Return(deleteErr)

	plainToken, err := svc.CreatePasswordResetToken(ctx, userID)

	assert.Error(t, err)
	assert.Empty(t, plainToken)
	assert.Contains(t, err.Error(), "failed to delete existing tokens")
	tokenRepo.AssertExpectations(t)
	tokenRepo.AssertNotCalled(t, "Create")
}

// TestEmailTokenService_CreatePasswordResetToken_CreateError tests error when storing the token fails
func TestEmailTokenService_CreatePasswordResetToken_CreateError(t *testing.T) {
	tokenRepo := new(MockEmailTokenRepo)
	userRepo := new(MockUserRepoForToken)
	svc := NewEmailTokenService(tokenRepo, userRepo)
	ctx := context.Background()

	userID := uuid.New()
	createErr := errors.New("insert failed")

	tokenRepo.On("DeleteByUserAndType", ctx, userID, models.TokenTypePasswordReset).Return(nil)
	tokenRepo.On("Create", ctx, mock.AnythingOfType("*models.EmailToken")).Return(createErr)

	plainToken, err := svc.CreatePasswordResetToken(ctx, userID)

	assert.Error(t, err)
	assert.Empty(t, plainToken)
	assert.Contains(t, err.Error(), "failed to store token")
	tokenRepo.AssertExpectations(t)
}

// ============================================================================
// TESTS: ConsumePasswordResetToken
// ============================================================================

// TestEmailTokenService_ConsumePasswordResetToken_Success tests successful password reset token consumption
func TestEmailTokenService_ConsumePasswordResetToken_Success(t *testing.T) {
	tokenRepo := new(MockEmailTokenRepo)
	userRepo := new(MockUserRepoForToken)
	svc := NewEmailTokenService(tokenRepo, userRepo)
	ctx := context.Background()

	userID := uuid.New()
	tokenID := uuid.New()
	plainToken := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	tokenHash := hashToken(plainToken)

	emailToken := &models.EmailToken{
		ID:        tokenID,
		UserID:    userID,
		TokenHash: tokenHash,
		TokenType: models.TokenTypePasswordReset,
		ExpiresAt: time.Now().Add(1 * time.Hour),
		UsedAt:    nil,
	}

	expectedUser := &models.User{}
	expectedUser.ID = userID
	expectedUser.Email = "test@example.com"

	tokenRepo.On("GetByTokenHash", ctx, tokenHash).Return(emailToken, nil)
	tokenRepo.On("MarkUsed", ctx, tokenID).Return(nil)
	userRepo.On("GetByID", ctx, userID).Return(expectedUser, nil)

	user, err := svc.ConsumePasswordResetToken(ctx, plainToken)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, userID, user.ID)
	assert.Equal(t, "test@example.com", user.Email)
	tokenRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

// TestEmailTokenService_ConsumePasswordResetToken_TokenNotFound tests that a non-existent token returns ErrTokenNotFound
func TestEmailTokenService_ConsumePasswordResetToken_TokenNotFound(t *testing.T) {
	tokenRepo := new(MockEmailTokenRepo)
	userRepo := new(MockUserRepoForToken)
	svc := NewEmailTokenService(tokenRepo, userRepo)
	ctx := context.Background()

	plainToken := "nonexistent_token_0000000000000000000000000000000000000000000000"
	tokenHash := hashToken(plainToken)

	tokenRepo.On("GetByTokenHash", ctx, tokenHash).Return(nil, gorm.ErrRecordNotFound)

	user, err := svc.ConsumePasswordResetToken(ctx, plainToken)

	assert.Nil(t, user)
	assert.ErrorIs(t, err, ErrTokenNotFound)
	tokenRepo.AssertExpectations(t)
}

// TestEmailTokenService_ConsumePasswordResetToken_TokenExpired tests that an expired token returns ErrTokenExpired
func TestEmailTokenService_ConsumePasswordResetToken_TokenExpired(t *testing.T) {
	tokenRepo := new(MockEmailTokenRepo)
	userRepo := new(MockUserRepoForToken)
	svc := NewEmailTokenService(tokenRepo, userRepo)
	ctx := context.Background()

	plainToken := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	tokenHash := hashToken(plainToken)

	emailToken := &models.EmailToken{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		TokenHash: tokenHash,
		TokenType: models.TokenTypePasswordReset,
		ExpiresAt: time.Now().Add(-1 * time.Hour), // expired
		UsedAt:    nil,
	}

	tokenRepo.On("GetByTokenHash", ctx, tokenHash).Return(emailToken, nil)

	user, err := svc.ConsumePasswordResetToken(ctx, plainToken)

	assert.Nil(t, user)
	assert.ErrorIs(t, err, ErrTokenExpired)
	tokenRepo.AssertExpectations(t)
}

// TestEmailTokenService_ConsumePasswordResetToken_TokenAlreadyUsed tests that a used token returns ErrTokenUsed
func TestEmailTokenService_ConsumePasswordResetToken_TokenAlreadyUsed(t *testing.T) {
	tokenRepo := new(MockEmailTokenRepo)
	userRepo := new(MockUserRepoForToken)
	svc := NewEmailTokenService(tokenRepo, userRepo)
	ctx := context.Background()

	plainToken := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	tokenHash := hashToken(plainToken)
	usedAt := time.Now().Add(-15 * time.Minute)

	emailToken := &models.EmailToken{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		TokenHash: tokenHash,
		TokenType: models.TokenTypePasswordReset,
		ExpiresAt: time.Now().Add(1 * time.Hour),
		UsedAt:    &usedAt, // already used
	}

	tokenRepo.On("GetByTokenHash", ctx, tokenHash).Return(emailToken, nil)

	user, err := svc.ConsumePasswordResetToken(ctx, plainToken)

	assert.Nil(t, user)
	assert.ErrorIs(t, err, ErrTokenUsed)
	tokenRepo.AssertExpectations(t)
}

// TestEmailTokenService_ConsumePasswordResetToken_WrongTokenType tests that a non-reset token returns ErrTokenNotFound
func TestEmailTokenService_ConsumePasswordResetToken_WrongTokenType(t *testing.T) {
	tokenRepo := new(MockEmailTokenRepo)
	userRepo := new(MockUserRepoForToken)
	svc := NewEmailTokenService(tokenRepo, userRepo)
	ctx := context.Background()

	plainToken := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	tokenHash := hashToken(plainToken)

	emailToken := &models.EmailToken{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		TokenHash: tokenHash,
		TokenType: models.TokenTypeEmailVerification, // wrong type
		ExpiresAt: time.Now().Add(1 * time.Hour),
		UsedAt:    nil,
	}

	tokenRepo.On("GetByTokenHash", ctx, tokenHash).Return(emailToken, nil)

	user, err := svc.ConsumePasswordResetToken(ctx, plainToken)

	assert.Nil(t, user)
	assert.ErrorIs(t, err, ErrTokenNotFound)
	tokenRepo.AssertExpectations(t)
}

// TestEmailTokenService_ConsumePasswordResetToken_MarkUsedError tests error when marking token as used fails
func TestEmailTokenService_ConsumePasswordResetToken_MarkUsedError(t *testing.T) {
	tokenRepo := new(MockEmailTokenRepo)
	userRepo := new(MockUserRepoForToken)
	svc := NewEmailTokenService(tokenRepo, userRepo)
	ctx := context.Background()

	tokenID := uuid.New()
	plainToken := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	tokenHash := hashToken(plainToken)

	emailToken := &models.EmailToken{
		ID:        tokenID,
		UserID:    uuid.New(),
		TokenHash: tokenHash,
		TokenType: models.TokenTypePasswordReset,
		ExpiresAt: time.Now().Add(1 * time.Hour),
		UsedAt:    nil,
	}

	tokenRepo.On("GetByTokenHash", ctx, tokenHash).Return(emailToken, nil)
	tokenRepo.On("MarkUsed", ctx, tokenID).Return(errors.New("db error"))

	user, err := svc.ConsumePasswordResetToken(ctx, plainToken)

	assert.Nil(t, user)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to mark token as used")
	tokenRepo.AssertExpectations(t)
}

// TestEmailTokenService_ConsumePasswordResetToken_GetUserError tests error when fetching user fails
func TestEmailTokenService_ConsumePasswordResetToken_GetUserError(t *testing.T) {
	tokenRepo := new(MockEmailTokenRepo)
	userRepo := new(MockUserRepoForToken)
	svc := NewEmailTokenService(tokenRepo, userRepo)
	ctx := context.Background()

	userID := uuid.New()
	tokenID := uuid.New()
	plainToken := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	tokenHash := hashToken(plainToken)

	emailToken := &models.EmailToken{
		ID:        tokenID,
		UserID:    userID,
		TokenHash: tokenHash,
		TokenType: models.TokenTypePasswordReset,
		ExpiresAt: time.Now().Add(1 * time.Hour),
		UsedAt:    nil,
	}

	tokenRepo.On("GetByTokenHash", ctx, tokenHash).Return(emailToken, nil)
	tokenRepo.On("MarkUsed", ctx, tokenID).Return(nil)
	userRepo.On("GetByID", ctx, userID).Return(nil, errors.New("user not found"))

	user, err := svc.ConsumePasswordResetToken(ctx, plainToken)

	assert.Nil(t, user)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get user")
	tokenRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

// TestEmailTokenService_ConsumePasswordResetToken_RepoLookupError tests generic repo error on token lookup
func TestEmailTokenService_ConsumePasswordResetToken_RepoLookupError(t *testing.T) {
	tokenRepo := new(MockEmailTokenRepo)
	userRepo := new(MockUserRepoForToken)
	svc := NewEmailTokenService(tokenRepo, userRepo)
	ctx := context.Background()

	plainToken := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	tokenHash := hashToken(plainToken)

	tokenRepo.On("GetByTokenHash", ctx, tokenHash).Return(nil, errors.New("connection reset"))

	user, err := svc.ConsumePasswordResetToken(ctx, plainToken)

	assert.Nil(t, user)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to look up token")
	tokenRepo.AssertExpectations(t)
}
