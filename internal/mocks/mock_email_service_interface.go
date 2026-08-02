package mocks

import (
	"context"
	"savvy/internal/email"

	"github.com/stretchr/testify/mock"
)

// MockEmailServiceInterface is a mock for email.ServiceInterface.
type MockEmailServiceInterface struct {
	mock.Mock
}

func (m *MockEmailServiceInterface) SendPasswordReset(ctx context.Context, toEmail, toName, resetURL, expiresIn, language string) error {
	return m.Called(ctx, toEmail, toName, resetURL, expiresIn, language).Error(0)
}

func (m *MockEmailServiceInterface) SendEmailVerification(ctx context.Context, toEmail, toName, verifyURL, language string) error {
	return m.Called(ctx, toEmail, toName, verifyURL, language).Error(0)
}

func (m *MockEmailServiceInterface) SendAccountDeletionConfirmation(ctx context.Context, toEmail, toName, language string) error {
	return m.Called(ctx, toEmail, toName, language).Error(0)
}

func (m *MockEmailServiceInterface) SendExpiryReminder(ctx context.Context, toEmail, toName string, data email.ExpiryReminderData, unsubscribeURL, language string) error {
	return m.Called(ctx, toEmail, toName, data, unsubscribeURL, language).Error(0)
}

func (m *MockEmailServiceInterface) SendValidityStart(ctx context.Context, toEmail, toName string, data email.ValidityStartData, unsubscribeURL, language string) error {
	return m.Called(ctx, toEmail, toName, data, unsubscribeURL, language).Error(0)
}

func (m *MockEmailServiceInterface) SendShareNotification(ctx context.Context, toEmail, toName, fromName, resourceType, merchantName, description string, amount float64, currency, resourceURL, unsubscribeURL, language string) error {
	return m.Called(ctx, toEmail, toName, fromName, resourceType, merchantName, description, amount, currency, resourceURL, unsubscribeURL, language).Error(0)
}

func (m *MockEmailServiceInterface) SendTransferNotification(ctx context.Context, toEmail, toName, fromName, resourceType, merchantName, description string, amount float64, currency, resourceURL, unsubscribeURL, language string) error {
	return m.Called(ctx, toEmail, toName, fromName, resourceType, merchantName, description, amount, currency, resourceURL, unsubscribeURL, language).Error(0)
}

func (m *MockEmailServiceInterface) SendTestEmail(ctx context.Context, toEmail, toName, language string) error {
	return m.Called(ctx, toEmail, toName, language).Error(0)
}

func (m *MockEmailServiceInterface) CheckConnection(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}
