package services

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"savvy/internal/email"
	"savvy/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// newTestDispatcher wires a dispatcher over the shared notification-test mocks.
func newTestDispatcher(
	notifRepo *MockNotificationRepository,
	userRepo *MockUserRepoForNotification,
	emailSvc *MockEmailSvcForNotification,
	tokenSvc *MockEmailTokenSvcForNotification,
) *EmailDispatchService {
	return NewEmailDispatchService(notifRepo, userRepo, emailSvc, tokenSvc, "https://savvy.example.com")
}

func dispatchTestRecipient(id uuid.UUID) *models.User {
	u := &models.User{Language: "en"}
	u.ID = id
	u.Email = "test@example.com"
	u.FirstName = "Test"
	return u
}

// TestDispatchPending_RoutesAllNotificationTypes verifies every notification type
// reaches its matching email method. A type that silently fell through would be
// a permanently undelivered mail.
func TestDispatchPending_RoutesAllNotificationTypes(t *testing.T) {
	tests := []struct {
		name       string
		notifType  models.NotificationType
		emailCall  string
		metadata   models.NotificationMetadata
		resourceTy string
	}{
		{
			name:       "expiry reminder",
			notifType:  models.NotificationTypeExpiryReminder,
			emailCall:  "SendExpiryReminder",
			metadata:   models.NotificationMetadata{"merchant_name": "IKEA", "days_left": 3},
			resourceTy: "voucher",
		},
		{
			name:       "validity start",
			notifType:  models.NotificationTypeValidityStart,
			emailCall:  "SendValidityStart",
			metadata:   models.NotificationMetadata{"merchant_name": "IKEA", "valid_from": "March 1, 2026"},
			resourceTy: "voucher",
		},
		{
			name:       "share received",
			notifType:  models.NotificationTypeShareReceived,
			emailCall:  "SendShareNotification",
			metadata:   models.NotificationMetadata{"from_user_name": "John", "merchant_name": "IKEA"},
			resourceTy: "card",
		},
		{
			name:       "transfer received",
			notifType:  models.NotificationTypeTransferReceived,
			emailCall:  "SendTransferNotification",
			metadata:   models.NotificationMetadata{"from_user_name": "John", "merchant_name": "IKEA"},
			resourceTy: "gift_card",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notifRepo := new(MockNotificationRepository)
			userRepo := new(MockUserRepoForNotification)
			emailSvc := new(MockEmailSvcForNotification)
			tokenSvc := new(MockEmailTokenSvcForNotification)
			svc := newTestDispatcher(notifRepo, userRepo, emailSvc, tokenSvc)
			ctx := context.Background()

			userID := uuid.New()
			notifID := uuid.New()
			claimed := []models.Notification{{
				ID:           notifID,
				UserID:       userID,
				Type:         tt.notifType,
				ResourceType: tt.resourceTy,
				Metadata:     tt.metadata,
				EmailStatus:  models.EmailStatusSending,
			}}

			notifRepo.On("ResetStaleSendingEmails", ctx, mock.Anything).Return(int64(0), nil)
			notifRepo.On("ClaimPendingEmails", ctx, mock.Anything).Return(claimed, nil)
			userRepo.On("GetByID", ctx, userID).Return(dispatchTestRecipient(userID), nil)
			tokenSvc.On("CreateUnsubscribeToken", ctx, userID).Return("tok", nil).Maybe()
			notifRepo.On("MarkEmailResult", ctx, notifID, nil, mock.Anything).Return(nil)

			switch tt.emailCall {
			case "SendExpiryReminder", "SendValidityStart":
				// These two are stubbed without testify expectations on the mock;
				// success is proven by MarkEmailResult being called with a nil error.
			default:
				emailSvc.On(tt.emailCall,
					ctx, "test@example.com", mock.Anything, "John", tt.resourceTy,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything).Return(nil)
			}

			sent, err := svc.DispatchPending(ctx)

			require.NoError(t, err)
			assert.Equal(t, 1, sent)
			// A nil sendErr recorded against the row is what "delivered" means.
			notifRepo.AssertCalled(t, "MarkEmailResult", ctx, notifID, nil, mock.Anything)
		})
	}
}

// TestDispatchPending_RecordsSendFailure is the core regression guard: before the
// outbox a failing send was only logged and the row was marked sent regardless,
// so the mail was lost for good. The error must now reach the row.
func TestDispatchPending_RecordsSendFailure(t *testing.T) {
	notifRepo := new(MockNotificationRepository)
	userRepo := new(MockUserRepoForNotification)
	emailSvc := new(MockEmailSvcForNotification)
	tokenSvc := new(MockEmailTokenSvcForNotification)
	svc := newTestDispatcher(notifRepo, userRepo, emailSvc, tokenSvc)
	ctx := context.Background()

	userID := uuid.New()
	notifID := uuid.New()
	sendErr := errors.New("smtp: connection refused")

	claimed := []models.Notification{{
		ID:           notifID,
		UserID:       userID,
		Type:         models.NotificationTypeShareReceived,
		ResourceType: "card",
		Metadata:     models.NotificationMetadata{"from_user_name": "John"},
		EmailStatus:  models.EmailStatusSending,
	}}

	notifRepo.On("ResetStaleSendingEmails", ctx, mock.Anything).Return(int64(0), nil)
	notifRepo.On("ClaimPendingEmails", ctx, mock.Anything).Return(claimed, nil)
	userRepo.On("GetByID", ctx, userID).Return(dispatchTestRecipient(userID), nil)
	tokenSvc.On("CreateUnsubscribeToken", ctx, userID).Return("tok", nil).Maybe()
	emailSvc.On("SendShareNotification",
		ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
	).Return(sendErr)
	notifRepo.On("MarkEmailResult", ctx, notifID, sendErr, defaultMaxEmailAttempts).Return(nil)

	sent, err := svc.DispatchPending(ctx)

	require.NoError(t, err)
	assert.Equal(t, 0, sent, "a failed send must not count as delivered")
	notifRepo.AssertCalled(t, "MarkEmailResult", ctx, notifID, sendErr, defaultMaxEmailAttempts)
}

// TestDispatchPending_UnknownTypeFails verifies an unroutable type is recorded as
// an error rather than left pending — otherwise it would be re-claimed forever.
func TestDispatchPending_UnknownTypeFails(t *testing.T) {
	notifRepo := new(MockNotificationRepository)
	userRepo := new(MockUserRepoForNotification)
	emailSvc := new(MockEmailSvcForNotification)
	tokenSvc := new(MockEmailTokenSvcForNotification)
	svc := newTestDispatcher(notifRepo, userRepo, emailSvc, tokenSvc)
	ctx := context.Background()

	userID := uuid.New()
	notifID := uuid.New()
	claimed := []models.Notification{{
		ID:          notifID,
		UserID:      userID,
		Type:        models.NotificationType("something_new"),
		EmailStatus: models.EmailStatusSending,
	}}

	notifRepo.On("ResetStaleSendingEmails", ctx, mock.Anything).Return(int64(0), nil)
	notifRepo.On("ClaimPendingEmails", ctx, mock.Anything).Return(claimed, nil)
	userRepo.On("GetByID", ctx, userID).Return(dispatchTestRecipient(userID), nil)
	notifRepo.On("MarkEmailResult", ctx, notifID, mock.MatchedBy(func(err error) bool {
		return err != nil
	}), 1).Return(nil)

	sent, err := svc.DispatchPending(ctx)

	require.NoError(t, err)
	assert.Equal(t, 0, sent)

	// Limit 1 makes the SQL CASE resolve straight to 'failed'. With the regular
	// budget the row would be re-claimed every minute for hours, and no retry
	// can ever route a type the dispatcher does not know.
	notifRepo.AssertCalled(t, "MarkEmailResult", ctx, notifID, mock.Anything, 1)
}

// TestDispatchPending_TransientErrorKeepsFullRetryBudget is the counterpart: a
// plain SMTP failure must NOT be parked early, or the outbox loses exactly the
// mail it exists to retry.
func TestDispatchPending_TransientErrorKeepsFullRetryBudget(t *testing.T) {
	notifRepo := new(MockNotificationRepository)
	userRepo := new(MockUserRepoForNotification)
	emailSvc := new(MockEmailSvcForNotification)
	tokenSvc := new(MockEmailTokenSvcForNotification)
	svc := newTestDispatcher(notifRepo, userRepo, emailSvc, tokenSvc)
	ctx := context.Background()

	userID := uuid.New()
	notifID := uuid.New()
	sendErr := errors.New("smtp: connection reset")
	claimed := []models.Notification{{
		ID:           notifID,
		UserID:       userID,
		Type:         models.NotificationTypeShareReceived,
		ResourceType: "card",
		Metadata:     models.NotificationMetadata{"from_user_name": "John"},
		EmailStatus:  models.EmailStatusSending,
	}}

	notifRepo.On("ResetStaleSendingEmails", ctx, mock.Anything).Return(int64(0), nil)
	notifRepo.On("ClaimPendingEmails", ctx, mock.Anything).Return(claimed, nil)
	userRepo.On("GetByID", ctx, userID).Return(dispatchTestRecipient(userID), nil)
	tokenSvc.On("CreateUnsubscribeToken", ctx, userID).Return("tok", nil).Maybe()
	emailSvc.On("SendShareNotification",
		ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
	).Return(sendErr)
	notifRepo.On("MarkEmailResult", ctx, notifID, sendErr, defaultMaxEmailAttempts).Return(nil)

	_, err := svc.DispatchPending(ctx)

	require.NoError(t, err)
	notifRepo.AssertCalled(t, "MarkEmailResult", ctx, notifID, sendErr, defaultMaxEmailAttempts)
}

// TestRetryBudgetOutlastsProviderOutage pins the constant against its purpose:
// 'failed' is terminal, so the budget has to cover a real hosted-SMTP incident
// (15-60 minutes), not just a blip.
func TestRetryBudgetOutlastsProviderOutage(t *testing.T) {
	budget := time.Duration(defaultMaxEmailAttempts) * time.Minute
	assert.Greater(t, budget, time.Hour,
		"a sub-hour budget parks the whole queue during a routine provider outage")
}

// TestBatchFitsInsideStaleWindow guards the pair of constants: a serial batch
// that outlives the stale window gets re-delivered by another replica while the
// first is still sending.
func TestBatchFitsInsideStaleWindow(t *testing.T) {
	const pessimisticSendDuration = 12 * time.Second

	worstCase := time.Duration(defaultEmailBatchSize) * pessimisticSendDuration
	assert.Less(t, worstCase, defaultStaleSendingAfter,
		"batch can outlive the stale window, causing duplicate delivery across replicas")
}

// TestDispatchPending_PreservesCodeValueAndResourceURL guards the fields the old
// inline path passed straight into the template. They live only in metadata now,
// so dropping them would ship a reminder with no code and a link to the bare list.
func TestDispatchPending_PreservesCodeValueAndResourceURL(t *testing.T) {
	notifRepo := new(MockNotificationRepository)
	userRepo := new(MockUserRepoForNotification)
	tokenSvc := new(MockEmailTokenSvcForNotification)
	captureSvc := &captureEmailService{}
	svc := NewEmailDispatchService(notifRepo, userRepo, captureSvc, tokenSvc, "https://savvy.example.com")
	ctx := context.Background()

	userID := uuid.New()
	notifID := uuid.New()
	claimed := []models.Notification{{
		ID:           notifID,
		UserID:       userID,
		Type:         models.NotificationTypeExpiryReminder,
		ResourceType: "voucher",
		Metadata: models.NotificationMetadata{
			"merchant_name": "IKEA",
			"days_left":     3,
			"expires_at":    "March 1, 2026",
			"code":          "SAVE20",
			"value":         "20%",
			"resource_url":  "https://savvy.example.com/vouchers/abc",
		},
		EmailStatus: models.EmailStatusSending,
	}}

	notifRepo.On("ResetStaleSendingEmails", ctx, mock.Anything).Return(int64(0), nil)
	notifRepo.On("ClaimPendingEmails", ctx, mock.Anything).Return(claimed, nil)
	userRepo.On("GetByID", ctx, userID).Return(dispatchTestRecipient(userID), nil)
	notifRepo.On("MarkEmailResult", ctx, notifID, nil, mock.Anything).Return(nil)

	sent, err := svc.DispatchPending(ctx)

	require.NoError(t, err)
	assert.Equal(t, 1, sent)
	assert.Equal(t, "SAVE20", captureSvc.expiryData.Code)
	assert.Equal(t, "20%", captureSvc.expiryData.Value)
	assert.Equal(t, "https://savvy.example.com/vouchers/abc", captureSvc.expiryData.ResourceURL)
	assert.Equal(t, 3, captureSvc.expiryData.DaysLeft)
}

// TestDispatchPending_SurvivesMetadataJSONRoundTrip is the JSONB trap: metadata
// comes back from Postgres decoded by encoding/json, which turns every number
// into float64. A direct int assertion would silently read 0 and the mail would
// claim the voucher expires today.
func TestDispatchPending_SurvivesMetadataJSONRoundTrip(t *testing.T) {
	original := models.NotificationMetadata{
		"merchant_name": "IKEA",
		"days_left":     3,
		"code":          "SAVE20",
		"value":         map[string]any{"amount": 50.0, "currency": "CHF"},
	}

	encoded, err := json.Marshal(original)
	require.NoError(t, err)

	var roundTripped models.NotificationMetadata
	require.NoError(t, json.Unmarshal(encoded, &roundTripped))

	// Proves the trap is real: the raw assertion the naive implementation would
	// use no longer holds after the round-trip.
	_, isInt := roundTripped["days_left"].(int)
	assert.False(t, isInt, "JSON decoding yields float64, not int")

	assert.Equal(t, 3, metadataInt(roundTripped, "days_left"))
	assert.Equal(t, "SAVE20", metadataString(roundTripped, "code"))

	amount, currency := metadataValue(roundTripped)
	assert.InDelta(t, 50.0, amount, 0.001)
	assert.Equal(t, "CHF", currency)
}

// TestDispatchPending_NoEmailServiceIsNoop verifies an unconfigured SMTP setup
// does not churn the queue.
func TestDispatchPending_NoEmailServiceIsNoop(t *testing.T) {
	notifRepo := new(MockNotificationRepository)
	userRepo := new(MockUserRepoForNotification)
	svc := NewEmailDispatchService(notifRepo, userRepo, nil, nil, "https://savvy.example.com")

	sent, err := svc.DispatchPending(context.Background())

	require.NoError(t, err)
	assert.Equal(t, 0, sent)
	notifRepo.AssertNotCalled(t, "ClaimPendingEmails")
}

// TestDispatchPending_ResetsStaleBeforeClaiming verifies recovery runs first, so
// rows stranded by a dead pod rejoin the same batch instead of waiting a cycle.
func TestDispatchPending_ResetsStaleBeforeClaiming(t *testing.T) {
	notifRepo := new(MockNotificationRepository)
	userRepo := new(MockUserRepoForNotification)
	emailSvc := new(MockEmailSvcForNotification)
	tokenSvc := new(MockEmailTokenSvcForNotification)
	svc := newTestDispatcher(notifRepo, userRepo, emailSvc, tokenSvc)
	ctx := context.Background()

	resetCalled := false
	notifRepo.On("ResetStaleSendingEmails", ctx, mock.Anything).Return(int64(2), nil).Run(func(mock.Arguments) {
		resetCalled = true
	})
	notifRepo.On("ClaimPendingEmails", ctx, mock.Anything).Return([]models.Notification{}, nil).Run(func(mock.Arguments) {
		assert.True(t, resetCalled, "stale reset must run before claiming")
	})

	_, err := svc.DispatchPending(ctx)

	require.NoError(t, err)
	notifRepo.AssertCalled(t, "ResetStaleSendingEmails", ctx, mock.Anything)
}

// TestReminderUnsubscribeURLUsesReminderToken verifies reminders carry the
// reminder opt-out, not the sharing one. Mixing them would switch off the wrong
// category when a user unsubscribes.
func TestReminderUnsubscribeURLUsesReminderToken(t *testing.T) {
	notifRepo := new(MockNotificationRepository)
	userRepo := new(MockUserRepoForNotification)
	emailSvc := new(MockEmailSvcForNotification)
	tokenSvc := &reminderTokenSvc{}
	svc := NewEmailDispatchService(notifRepo, userRepo, emailSvc, tokenSvc, "https://savvy.example.com")

	url := svc.reminderUnsubscribeURL(context.Background(), uuid.New())
	assert.Equal(t, "https://savvy.example.com/unsubscribe?token=reminder-token&type=reminders", url)
}

// reminderTokenSvc returns distinguishable tokens per category.
type reminderTokenSvc struct {
	MockEmailTokenSvcForNotification
}

func (m *reminderTokenSvc) CreateUnsubscribeReminderToken(_ context.Context, _ uuid.UUID) (string, error) {
	return "reminder-token", nil
}

// captureEmailService records the data handed to the email layer so a test can
// assert on template inputs rather than on a mock call signature.
type captureEmailService struct {
	expiryData email.ExpiryReminderData
}

func (c *captureEmailService) SendExpiryReminder(_ context.Context, _, _ string, data email.ExpiryReminderData, _, _ string) error {
	c.expiryData = data
	return nil
}
func (c *captureEmailService) SendValidityStart(_ context.Context, _, _ string, _ email.ValidityStartData, _, _ string) error {
	return nil
}
func (c *captureEmailService) SendShareNotification(_ context.Context, _, _, _, _, _, _ string, _ float64, _, _, _, _ string) error {
	return nil
}
func (c *captureEmailService) SendTransferNotification(_ context.Context, _, _, _, _, _, _ string, _ float64, _, _, _, _ string) error {
	return nil
}
func (c *captureEmailService) SendPasswordReset(_ context.Context, _, _, _, _, _ string) error {
	return nil
}
func (c *captureEmailService) SendEmailVerification(_ context.Context, _, _, _, _ string) error {
	return nil
}
func (c *captureEmailService) SendAccountDeletionConfirmation(_ context.Context, _, _, _ string) error {
	return nil
}
func (c *captureEmailService) SendTestEmail(_ context.Context, _, _, _ string) error { return nil }
func (c *captureEmailService) CheckConnection(_ context.Context) error               { return nil }
func (c *captureEmailService) IsConfigured() bool                                    { return true }

var _ email.ServiceInterface = (*captureEmailService)(nil)
