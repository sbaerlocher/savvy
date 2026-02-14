package mocks

import (
	"context"
	"savvy/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockEmailTokenServiceInterface is a mock for services.EmailTokenServiceInterface.
type MockEmailTokenServiceInterface struct {
	mock.Mock
}

func (m *MockEmailTokenServiceInterface) CreateVerificationToken(ctx context.Context, userID uuid.UUID) (string, error) {
	args := m.Called(ctx, userID)
	return args.String(0), args.Error(1)
}

func (m *MockEmailTokenServiceInterface) VerifyEmail(ctx context.Context, token string) error {
	return m.Called(ctx, token).Error(0)
}

func (m *MockEmailTokenServiceInterface) CreatePasswordResetToken(ctx context.Context, userID uuid.UUID) (string, error) {
	args := m.Called(ctx, userID)
	return args.String(0), args.Error(1)
}

func (m *MockEmailTokenServiceInterface) ConsumePasswordResetToken(ctx context.Context, token string) (*models.User, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockEmailTokenServiceInterface) CreateUnsubscribeToken(ctx context.Context, userID uuid.UUID) (string, error) {
	args := m.Called(ctx, userID)
	return args.String(0), args.Error(1)
}

func (m *MockEmailTokenServiceInterface) UnsubscribeNotifications(ctx context.Context, token string) error {
	return m.Called(ctx, token).Error(0)
}

func (m *MockEmailTokenServiceInterface) CreateUnsubscribeReminderToken(ctx context.Context, userID uuid.UUID) (string, error) {
	args := m.Called(ctx, userID)
	return args.String(0), args.Error(1)
}

func (m *MockEmailTokenServiceInterface) UnsubscribeReminders(ctx context.Context, token string) error {
	return m.Called(ctx, token).Error(0)
}

func (m *MockEmailTokenServiceInterface) CleanupExpiredTokens(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}
