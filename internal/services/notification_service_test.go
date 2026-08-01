package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"savvy/internal/email"
	"savvy/internal/models"
	"savvy/internal/repository"
)

// MockNotificationRepository is a mock for NotificationRepository
type MockNotificationRepository struct {
	mock.Mock
}

func (m *MockNotificationRepository) Create(ctx context.Context, notification *models.Notification) error {
	args := m.Called(ctx, notification)
	return args.Error(0)
}

func (m *MockNotificationRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Notification, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Notification), args.Error(1)
}

func (m *MockNotificationRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Notification, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Notification), args.Error(1)
}

func (m *MockNotificationRepository) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockNotificationRepository) MarkAsRead(ctx context.Context, userID, notificationID uuid.UUID) error {
	args := m.Called(ctx, userID, notificationID)
	return args.Error(0)
}

func (m *MockNotificationRepository) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockNotificationRepository) Delete(ctx context.Context, userID, notificationID uuid.UUID) error {
	args := m.Called(ctx, userID, notificationID)
	return args.Error(0)
}

func (m *MockNotificationRepository) ArchiveOldRead(ctx context.Context, cutoff time.Time) (int64, error) {
	args := m.Called(ctx, cutoff)
	return args.Get(0).(int64), args.Error(1)
}

// Ensure MockNotificationRepository implements NotificationRepository
var _ repository.NotificationRepository = (*MockNotificationRepository)(nil)

// ============================================================================
// TESTS
// ============================================================================

// TestNotificationService_CreateShareNotification tests creating a share notification
func TestNotificationService_CreateShareNotification(t *testing.T) {
	mockRepo := new(MockNotificationRepository)
	service := NewNotificationService(mockRepo, nil)
	ctx := context.Background()

	recipientID := uuid.New()
	fromUserID := uuid.New()
	resourceID := uuid.New()
	permissions := map[string]bool{
		"can_edit":   true,
		"can_delete": false,
	}

	mockRepo.On("Create", ctx, mock.MatchedBy(func(n *models.Notification) bool {
		return n.UserID == recipientID &&
			n.Type == models.NotificationTypeShareReceived &&
			n.ResourceType == "card" &&
			n.ResourceID == resourceID &&
			!n.IsRead
	})).Return(nil)

	// Test: Create share notification
	err := service.CreateShareNotification(ctx, recipientID, fromUserID, "John Doe", "card", resourceID, permissions)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// TestNotificationService_CreateShareNotification_NoPermissions tests without permissions
func TestNotificationService_CreateShareNotification_NoPermissions(t *testing.T) {
	mockRepo := new(MockNotificationRepository)
	service := NewNotificationService(mockRepo, nil)
	ctx := context.Background()

	recipientID := uuid.New()
	fromUserID := uuid.New()
	resourceID := uuid.New()

	mockRepo.On("Create", ctx, mock.MatchedBy(func(n *models.Notification) bool {
		return n.UserID == recipientID &&
			n.Type == models.NotificationTypeShareReceived &&
			n.ResourceType == "voucher" &&
			n.ResourceID == resourceID
	})).Return(nil)

	// Test: Create share notification without permissions
	err := service.CreateShareNotification(ctx, recipientID, fromUserID, "Jane Doe", "voucher", resourceID, nil)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// TestNotificationService_CreateShareNotification_Error tests handling of repository errors
func TestNotificationService_CreateShareNotification_Error(t *testing.T) {
	mockRepo := new(MockNotificationRepository)
	service := NewNotificationService(mockRepo, nil)
	ctx := context.Background()

	recipientID := uuid.New()
	fromUserID := uuid.New()
	resourceID := uuid.New()

	repoError := errors.New("database error")
	mockRepo.On("Create", ctx, mock.Anything).Return(repoError)

	// Test: Repository error is propagated
	err := service.CreateShareNotification(ctx, recipientID, fromUserID, "John", "card", resourceID, nil)

	assert.Error(t, err)
	assert.ErrorIs(t, err, repoError)
	mockRepo.AssertExpectations(t)
}

// TestNotificationService_CreateTransferNotification tests creating a transfer notification
func TestNotificationService_CreateTransferNotification(t *testing.T) {
	mockRepo := new(MockNotificationRepository)
	service := NewNotificationService(mockRepo, nil)
	ctx := context.Background()

	recipientID := uuid.New()
	fromUserID := uuid.New()
	resourceID := uuid.New()

	mockRepo.On("Create", ctx, mock.MatchedBy(func(n *models.Notification) bool {
		return n.UserID == recipientID &&
			n.Type == models.NotificationTypeTransferReceived &&
			n.ResourceType == "gift_card" &&
			n.ResourceID == resourceID &&
			!n.IsRead
	})).Return(nil)

	// Test: Create transfer notification
	err := service.CreateTransferNotification(ctx, recipientID, fromUserID, "Transfer User", "gift_card", resourceID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// TestNotificationService_CreateTransferNotification_Card tests transfer for card
func TestNotificationService_CreateTransferNotification_Card(t *testing.T) {
	mockRepo := new(MockNotificationRepository)
	service := NewNotificationService(mockRepo, nil)
	ctx := context.Background()

	recipientID := uuid.New()
	fromUserID := uuid.New()
	resourceID := uuid.New()

	mockRepo.On("Create", ctx, mock.MatchedBy(func(n *models.Notification) bool {
		return n.ResourceType == "card"
	})).Return(nil)

	// Test: Transfer notification for card
	err := service.CreateTransferNotification(ctx, recipientID, fromUserID, "Owner", "card", resourceID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// TestNotificationService_CreateTransferNotification_Error tests handling of errors
func TestNotificationService_CreateTransferNotification_Error(t *testing.T) {
	mockRepo := new(MockNotificationRepository)
	service := NewNotificationService(mockRepo, nil)
	ctx := context.Background()

	recipientID := uuid.New()
	fromUserID := uuid.New()
	resourceID := uuid.New()

	repoError := errors.New("create failed")
	mockRepo.On("Create", ctx, mock.Anything).Return(repoError)

	// Test: Error is propagated
	err := service.CreateTransferNotification(ctx, recipientID, fromUserID, "User", "voucher", resourceID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, repoError)
	mockRepo.AssertExpectations(t)
}

// TestNotificationService_GetUserNotifications tests retrieving notifications
func TestNotificationService_GetUserNotifications(t *testing.T) {
	mockRepo := new(MockNotificationRepository)
	service := NewNotificationService(mockRepo, nil)
	ctx := context.Background()

	userID := uuid.New()
	limit := 10
	offset := 0

	expectedNotifications := []models.Notification{
		{
			ID:           uuid.New(),
			UserID:       userID,
			Type:         models.NotificationTypeShareReceived,
			ResourceType: "card",
			ResourceID:   uuid.New(),
			IsRead:       false,
		},
		{
			ID:           uuid.New(),
			UserID:       userID,
			Type:         models.NotificationTypeTransferReceived,
			ResourceType: "voucher",
			ResourceID:   uuid.New(),
			IsRead:       true,
		},
	}

	mockRepo.On("GetByUserID", ctx, userID, limit, offset).Return(expectedNotifications, nil)

	// Test: Get user notifications
	notifications, err := service.GetUserNotifications(ctx, userID, limit, offset)

	assert.NoError(t, err)
	assert.Len(t, notifications, 2)
	assert.Equal(t, userID, notifications[0].UserID)
	mockRepo.AssertExpectations(t)
}

// TestNotificationService_GetUserNotifications_Empty tests empty result
func TestNotificationService_GetUserNotifications_Empty(t *testing.T) {
	mockRepo := new(MockNotificationRepository)
	service := NewNotificationService(mockRepo, nil)
	ctx := context.Background()

	userID := uuid.New()
	limit := 10
	offset := 0

	mockRepo.On("GetByUserID", ctx, userID, limit, offset).Return([]models.Notification{}, nil)

	// Test: Empty notifications
	notifications, err := service.GetUserNotifications(ctx, userID, limit, offset)

	assert.NoError(t, err)
	assert.Empty(t, notifications)
	mockRepo.AssertExpectations(t)
}

// TestNotificationService_GetUserNotifications_Pagination tests pagination
func TestNotificationService_GetUserNotifications_Pagination(t *testing.T) {
	mockRepo := new(MockNotificationRepository)
	service := NewNotificationService(mockRepo, nil)
	ctx := context.Background()

	userID := uuid.New()
	limit := 5
	offset := 10

	expectedNotifications := []models.Notification{
		{ID: uuid.New(), UserID: userID, Type: models.NotificationTypeShareReceived},
	}

	mockRepo.On("GetByUserID", ctx, userID, limit, offset).Return(expectedNotifications, nil)

	// Test: Pagination parameters are passed correctly
	notifications, err := service.GetUserNotifications(ctx, userID, limit, offset)

	assert.NoError(t, err)
	assert.Len(t, notifications, 1)
	mockRepo.AssertExpectations(t)
}

// TestNotificationService_GetUserNotifications_Error tests handling of errors
func TestNotificationService_GetUserNotifications_Error(t *testing.T) {
	mockRepo := new(MockNotificationRepository)
	service := NewNotificationService(mockRepo, nil)
	ctx := context.Background()

	userID := uuid.New()
	repoError := errors.New("query failed")

	mockRepo.On("GetByUserID", ctx, userID, 10, 0).Return(nil, repoError)

	// Test: Error is propagated
	notifications, err := service.GetUserNotifications(ctx, userID, 10, 0)

	assert.Error(t, err)
	assert.ErrorIs(t, err, repoError)
	assert.Nil(t, notifications)
	mockRepo.AssertExpectations(t)
}

// TestNotificationService_GetUnreadCount tests getting unread count
func TestNotificationService_GetUnreadCount(t *testing.T) {
	mockRepo := new(MockNotificationRepository)
	service := NewNotificationService(mockRepo, nil)
	ctx := context.Background()

	userID := uuid.New()
	expectedCount := int64(5)

	mockRepo.On("GetUnreadCount", ctx, userID).Return(expectedCount, nil)

	// Test: Get unread count
	count, err := service.GetUnreadCount(ctx, userID)

	assert.NoError(t, err)
	assert.Equal(t, expectedCount, count)
	mockRepo.AssertExpectations(t)
}

// TestNotificationService_GetUnreadCount_Zero tests zero unread notifications
func TestNotificationService_GetUnreadCount_Zero(t *testing.T) {
	mockRepo := new(MockNotificationRepository)
	service := NewNotificationService(mockRepo, nil)
	ctx := context.Background()

	userID := uuid.New()

	mockRepo.On("GetUnreadCount", ctx, userID).Return(int64(0), nil)

	// Test: Zero unread
	count, err := service.GetUnreadCount(ctx, userID)

	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
	mockRepo.AssertExpectations(t)
}

// TestNotificationService_GetUnreadCount_Error tests error handling
func TestNotificationService_GetUnreadCount_Error(t *testing.T) {
	mockRepo := new(MockNotificationRepository)
	service := NewNotificationService(mockRepo, nil)
	ctx := context.Background()

	userID := uuid.New()
	repoError := errors.New("count failed")

	mockRepo.On("GetUnreadCount", ctx, userID).Return(int64(0), repoError)

	// Test: Error is propagated
	count, err := service.GetUnreadCount(ctx, userID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, repoError)
	assert.Equal(t, int64(0), count)
	mockRepo.AssertExpectations(t)
}

// TestNotificationService_MarkAsRead tests marking a notification as read
func TestNotificationService_MarkAsRead(t *testing.T) {
	mockRepo := new(MockNotificationRepository)
	service := NewNotificationService(mockRepo, nil)
	ctx := context.Background()

	userID := uuid.New()
	notificationID := uuid.New()

	mockRepo.On("MarkAsRead", ctx, userID, notificationID).Return(nil)

	// Test: Mark as read
	err := service.MarkAsRead(ctx, userID, notificationID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// TestNotificationService_MarkAsRead_Error tests error handling
func TestNotificationService_MarkAsRead_Error(t *testing.T) {
	mockRepo := new(MockNotificationRepository)
	service := NewNotificationService(mockRepo, nil)
	ctx := context.Background()

	userID := uuid.New()
	notificationID := uuid.New()
	repoError := errors.New("update failed")

	mockRepo.On("MarkAsRead", ctx, userID, notificationID).Return(repoError)

	// Test: Error is propagated
	err := service.MarkAsRead(ctx, userID, notificationID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, repoError)
	mockRepo.AssertExpectations(t)
}

// TestNotificationService_MarkAllAsRead tests marking all notifications as read
func TestNotificationService_MarkAllAsRead(t *testing.T) {
	mockRepo := new(MockNotificationRepository)
	service := NewNotificationService(mockRepo, nil)
	ctx := context.Background()

	userID := uuid.New()

	mockRepo.On("MarkAllAsRead", ctx, userID).Return(nil)

	// Test: Mark all as read
	err := service.MarkAllAsRead(ctx, userID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// TestNotificationService_MarkAllAsRead_Error tests error handling
func TestNotificationService_MarkAllAsRead_Error(t *testing.T) {
	mockRepo := new(MockNotificationRepository)
	service := NewNotificationService(mockRepo, nil)
	ctx := context.Background()

	userID := uuid.New()
	repoError := errors.New("bulk update failed")

	mockRepo.On("MarkAllAsRead", ctx, userID).Return(repoError)

	// Test: Error is propagated
	err := service.MarkAllAsRead(ctx, userID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, repoError)
	mockRepo.AssertExpectations(t)
}

// TestNotificationService_DeleteNotification tests deleting a notification
func TestNotificationService_DeleteNotification(t *testing.T) {
	mockRepo := new(MockNotificationRepository)
	service := NewNotificationService(mockRepo, nil)
	ctx := context.Background()

	userID := uuid.New()
	notificationID := uuid.New()

	mockRepo.On("Delete", ctx, userID, notificationID).Return(nil)

	// Test: Delete notification
	err := service.DeleteNotification(ctx, userID, notificationID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// TestNotificationService_DeleteNotification_Error tests error handling
func TestNotificationService_DeleteNotification_Error(t *testing.T) {
	mockRepo := new(MockNotificationRepository)
	service := NewNotificationService(mockRepo, nil)
	ctx := context.Background()

	userID := uuid.New()
	notificationID := uuid.New()
	repoError := errors.New("delete failed")

	mockRepo.On("Delete", ctx, userID, notificationID).Return(repoError)

	// Test: Error is propagated
	err := service.DeleteNotification(ctx, userID, notificationID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, repoError)
	mockRepo.AssertExpectations(t)
}

// TestNotificationService_CreateShareNotification_MetadataStructure tests metadata structure
func TestNotificationService_CreateShareNotification_MetadataStructure(t *testing.T) {
	mockRepo := new(MockNotificationRepository)
	service := NewNotificationService(mockRepo, nil)
	ctx := context.Background()

	recipientID := uuid.New()
	fromUserID := uuid.New()
	resourceID := uuid.New()
	permissions := map[string]bool{
		"can_edit":              true,
		"can_delete":            false,
		"can_edit_transactions": true,
	}

	mockRepo.On("Create", ctx, mock.MatchedBy(func(n *models.Notification) bool {
		// Verify metadata contains expected fields
		fromUserIDStr, ok := n.Metadata["from_user_id"].(string)
		if !ok {
			return false
		}
		fromUserName, ok := n.Metadata["from_user_name"].(string)
		if !ok {
			return false
		}
		perms, ok := n.Metadata["permissions"].(map[string]bool)
		if !ok {
			return false
		}

		return fromUserIDStr == fromUserID.String() &&
			fromUserName == "Alice" &&
			perms["can_edit"] == true &&
			perms["can_delete"] == false &&
			perms["can_edit_transactions"] == true
	})).Return(nil)

	// Test: Metadata structure is correct
	err := service.CreateShareNotification(ctx, recipientID, fromUserID, "Alice", "gift_card", resourceID, permissions)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// TestNotificationService_CreateTransferNotification_MetadataStructure tests metadata for transfer
func TestNotificationService_CreateTransferNotification_MetadataStructure(t *testing.T) {
	mockRepo := new(MockNotificationRepository)
	service := NewNotificationService(mockRepo, nil)
	ctx := context.Background()

	recipientID := uuid.New()
	fromUserID := uuid.New()
	resourceID := uuid.New()

	mockRepo.On("Create", ctx, mock.MatchedBy(func(n *models.Notification) bool {
		// Verify metadata structure
		fromUserIDStr, ok := n.Metadata["from_user_id"].(string)
		if !ok {
			return false
		}
		fromUserName, ok := n.Metadata["from_user_name"].(string)
		if !ok {
			return false
		}

		return fromUserIDStr == fromUserID.String() &&
			fromUserName == "Bob"
	})).Return(nil)

	// Test: Transfer metadata structure
	err := service.CreateTransferNotification(ctx, recipientID, fromUserID, "Bob", "card", resourceID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// ============================================================================
// NOTIFICATION PREFERENCE GATING TESTS
// ============================================================================

// MockUserRepoForNotification is a mock for UserRepository used in notification tests
type MockUserRepoForNotification struct {
	mock.Mock
}

func (m *MockUserRepoForNotification) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepoForNotification) GetByEmail(ctx context.Context, e string) (*models.User, error) {
	args := m.Called(ctx, e)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepoForNotification) Create(ctx context.Context, user *models.User) error {
	return m.Called(ctx, user).Error(0)
}

func (m *MockUserRepoForNotification) Update(ctx context.Context, user *models.User) error {
	return m.Called(ctx, user).Error(0)
}

func (m *MockUserRepoForNotification) GetAll(ctx context.Context) ([]models.User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockUserRepoForNotification) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]models.User, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockUserRepoForNotification) SearchByIDs(ctx context.Context, ids []uuid.UUID, query string) ([]models.User, error) {
	args := m.Called(ctx, ids, query)
	return args.Get(0).([]models.User), args.Error(1)
}

var _ repository.UserRepository = (*MockUserRepoForNotification)(nil)

// MockPushSvcForNotification is a mock for PushServiceInterface
type MockPushSvcForNotification struct {
	mock.Mock
}

func (m *MockPushSvcForNotification) Subscribe(_ context.Context, _ uuid.UUID, _, _, _, _ string) error {
	return nil
}

func (m *MockPushSvcForNotification) Unsubscribe(_ context.Context, _ string) error {
	return nil
}

func (m *MockPushSvcForNotification) SendPushToUser(ctx context.Context, userID uuid.UUID, title, body, url string) error {
	args := m.Called(ctx, userID, title, body, url)
	return args.Error(0)
}

func (m *MockPushSvcForNotification) SendTestPush(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *MockPushSvcForNotification) GetVAPIDPublicKey() string { return "" }
func (m *MockPushSvcForNotification) IsEnabled() bool           { return true }

var _ PushServiceInterface = (*MockPushSvcForNotification)(nil)

// MockEmailSvcForNotification is a mock for email.ServiceInterface
type MockEmailSvcForNotification struct {
	mock.Mock
}

func (m *MockEmailSvcForNotification) SendShareNotification(ctx context.Context, toEmail, toName, fromName, resourceType, resourceURL, unsubscribeURL, language string) error {
	args := m.Called(ctx, toEmail, toName, fromName, resourceType, resourceURL, unsubscribeURL, language)
	return args.Error(0)
}

func (m *MockEmailSvcForNotification) SendTransferNotification(ctx context.Context, toEmail, toName, fromName, resourceType, resourceURL, unsubscribeURL, language string) error {
	args := m.Called(ctx, toEmail, toName, fromName, resourceType, resourceURL, unsubscribeURL, language)
	return args.Error(0)
}

func (m *MockEmailSvcForNotification) SendTestEmail(_ context.Context, _, _, _ string) error {
	return nil
}

func (m *MockEmailSvcForNotification) SendPasswordReset(_ context.Context, _, _, _, _, _ string) error {
	return nil
}

func (m *MockEmailSvcForNotification) SendEmailVerification(_ context.Context, _, _, _, _ string) error {
	return nil
}

func (m *MockEmailSvcForNotification) SendAccountDeletionConfirmation(_ context.Context, _, _, _ string) error {
	return nil
}

func (m *MockEmailSvcForNotification) SendExpiryReminder(_ context.Context, _, _ string, _ email.ExpiryReminderData, _, _ string) error {
	return nil
}

func (m *MockEmailSvcForNotification) SendValidityStart(_ context.Context, _, _ string, _ email.ValidityStartData, _, _ string) error {
	return nil
}

func (m *MockEmailSvcForNotification) CheckConnection(_ context.Context) error { return nil }
func (m *MockEmailSvcForNotification) IsConfigured() bool                      { return true }

var _ email.ServiceInterface = (*MockEmailSvcForNotification)(nil)

// MockEmailTokenSvcForNotification is a mock for EmailTokenServiceInterface
type MockEmailTokenSvcForNotification struct {
	mock.Mock
}

func (m *MockEmailTokenSvcForNotification) CreateVerificationToken(_ context.Context, _ uuid.UUID) (string, error) {
	return "", nil
}

func (m *MockEmailTokenSvcForNotification) VerifyEmail(_ context.Context, _ string) error {
	return nil
}

func (m *MockEmailTokenSvcForNotification) CreatePasswordResetToken(_ context.Context, _ uuid.UUID) (string, error) {
	return "", nil
}

func (m *MockEmailTokenSvcForNotification) ConsumePasswordResetToken(_ context.Context, _ string) (*models.User, error) {
	return nil, nil
}

func (m *MockEmailTokenSvcForNotification) CreateUnsubscribeToken(ctx context.Context, userID uuid.UUID) (string, error) {
	args := m.Called(ctx, userID)
	return args.String(0), args.Error(1)
}

func (m *MockEmailTokenSvcForNotification) UnsubscribeNotifications(_ context.Context, _ string) error {
	return nil
}

func (m *MockEmailTokenSvcForNotification) CreateUnsubscribeReminderToken(_ context.Context, _ uuid.UUID) (string, error) {
	return "", nil
}

func (m *MockEmailTokenSvcForNotification) UnsubscribeReminders(_ context.Context, _ string) error {
	return nil
}

func (m *MockEmailTokenSvcForNotification) CleanupExpiredTokens(_ context.Context) error {
	return nil
}

var _ EmailTokenServiceInterface = (*MockEmailTokenSvcForNotification)(nil)

// newTestNotificationService creates a NotificationService with all mocks wired up
func newTestNotificationService(
	notifRepo *MockNotificationRepository,
	userRepo *MockUserRepoForNotification,
	pushSvc *MockPushSvcForNotification,
	emailSvc *MockEmailSvcForNotification,
	emailTokenSvc *MockEmailTokenSvcForNotification,
) *NotificationService {
	svc := &NotificationService{
		repo:              notifRepo,
		userRepo:          userRepo,
		pushService:       pushSvc,
		emailService:      emailSvc,
		emailTokenService: emailTokenSvc,
		frontendURL:       "https://app.test",
	}
	return svc
}

// TestNotificationService_SharePushGating_PushSharingDisabled verifies that
// push notifications are NOT sent when PushSharingEnabled is false
func TestNotificationService_SharePushGating_PushSharingDisabled(t *testing.T) {
	notifRepo := new(MockNotificationRepository)
	userRepo := new(MockUserRepoForNotification)
	pushSvc := new(MockPushSvcForNotification)
	emailSvc := new(MockEmailSvcForNotification)
	emailTokenSvc := new(MockEmailTokenSvcForNotification)

	svc := newTestNotificationService(notifRepo, userRepo, pushSvc, emailSvc, emailTokenSvc)
	ctx := context.Background()

	recipientID := uuid.New()
	fromUserID := uuid.New()
	resourceID := uuid.New()

	recipient := &models.User{
		PushNotificationsEnabled:  true,
		PushSharingEnabled:        false, // disabled
		EmailNotificationsEnabled: false,
		EmailSharingEnabled:       false,
	}
	recipient.ID = recipientID
	recipient.Email = "test@example.com"

	notifRepo.On("Create", ctx, mock.Anything).Return(nil)
	userRepo.On("GetByID", ctx, recipientID).Return(recipient, nil)

	// Push should NOT be called because PushSharingEnabled is false
	// Email should NOT be called because EmailNotificationsEnabled is false

	err := svc.CreateShareNotification(ctx, recipientID, fromUserID, "John", "card", resourceID, nil)

	assert.NoError(t, err)
	pushSvc.AssertNotCalled(t, "SendPushToUser")
	emailSvc.AssertNotCalled(t, "SendShareNotification")
}

// TestNotificationService_SharePushGating_PushChannelDisabled verifies that
// push notifications are NOT sent when PushNotificationsEnabled is false
func TestNotificationService_SharePushGating_PushChannelDisabled(t *testing.T) {
	notifRepo := new(MockNotificationRepository)
	userRepo := new(MockUserRepoForNotification)
	pushSvc := new(MockPushSvcForNotification)
	emailSvc := new(MockEmailSvcForNotification)
	emailTokenSvc := new(MockEmailTokenSvcForNotification)

	svc := newTestNotificationService(notifRepo, userRepo, pushSvc, emailSvc, emailTokenSvc)
	ctx := context.Background()

	recipientID := uuid.New()
	fromUserID := uuid.New()
	resourceID := uuid.New()

	recipient := &models.User{
		PushNotificationsEnabled:  false, // channel disabled
		PushSharingEnabled:        true,
		EmailNotificationsEnabled: false,
		EmailSharingEnabled:       false,
	}
	recipient.ID = recipientID
	recipient.Email = "test@example.com"

	notifRepo.On("Create", ctx, mock.Anything).Return(nil)
	userRepo.On("GetByID", ctx, recipientID).Return(recipient, nil)

	err := svc.CreateShareNotification(ctx, recipientID, fromUserID, "John", "card", resourceID, nil)

	assert.NoError(t, err)
	pushSvc.AssertNotCalled(t, "SendPushToUser")
}

// TestNotificationService_SharePushGating_BothEnabled verifies that
// push notifications ARE sent when both PushNotificationsEnabled and PushSharingEnabled are true
func TestNotificationService_SharePushGating_BothEnabled(t *testing.T) {
	notifRepo := new(MockNotificationRepository)
	userRepo := new(MockUserRepoForNotification)
	pushSvc := new(MockPushSvcForNotification)
	emailSvc := new(MockEmailSvcForNotification)
	emailTokenSvc := new(MockEmailTokenSvcForNotification)

	svc := newTestNotificationService(notifRepo, userRepo, pushSvc, emailSvc, emailTokenSvc)
	ctx := context.Background()

	recipientID := uuid.New()
	fromUserID := uuid.New()
	resourceID := uuid.New()

	recipient := &models.User{
		PushNotificationsEnabled:  true,
		PushSharingEnabled:        true, // both enabled
		EmailNotificationsEnabled: false,
		EmailSharingEnabled:       false,
	}
	recipient.ID = recipientID
	recipient.Email = "test@example.com"

	notifRepo.On("Create", ctx, mock.Anything).Return(nil)
	userRepo.On("GetByID", ctx, recipientID).Return(recipient, nil)
	pushSvc.On("SendPushToUser", ctx, recipientID, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := svc.CreateShareNotification(ctx, recipientID, fromUserID, "John", "card", resourceID, nil)

	assert.NoError(t, err)
	pushSvc.AssertCalled(t, "SendPushToUser", ctx, recipientID, mock.Anything, mock.Anything, mock.Anything)
}

// TestNotificationService_ShareEmailGating_EmailSharingDisabled verifies that
// share emails are NOT sent when EmailSharingEnabled is false
func TestNotificationService_ShareEmailGating_EmailSharingDisabled(t *testing.T) {
	notifRepo := new(MockNotificationRepository)
	userRepo := new(MockUserRepoForNotification)
	pushSvc := new(MockPushSvcForNotification)
	emailSvc := new(MockEmailSvcForNotification)
	emailTokenSvc := new(MockEmailTokenSvcForNotification)

	svc := newTestNotificationService(notifRepo, userRepo, pushSvc, emailSvc, emailTokenSvc)
	ctx := context.Background()

	recipientID := uuid.New()
	fromUserID := uuid.New()
	resourceID := uuid.New()

	recipient := &models.User{
		PushNotificationsEnabled:  false,
		PushSharingEnabled:        false,
		EmailNotificationsEnabled: true,
		EmailSharingEnabled:       false, // category disabled
	}
	recipient.ID = recipientID
	recipient.Email = "test@example.com"

	notifRepo.On("Create", ctx, mock.Anything).Return(nil)
	userRepo.On("GetByID", ctx, recipientID).Return(recipient, nil)

	err := svc.CreateShareNotification(ctx, recipientID, fromUserID, "John", "card", resourceID, nil)

	assert.NoError(t, err)
	emailSvc.AssertNotCalled(t, "SendShareNotification")
}

// TestNotificationService_ShareEmailGating_EmailChannelDisabled verifies that
// share emails are NOT sent when EmailNotificationsEnabled is false
func TestNotificationService_ShareEmailGating_EmailChannelDisabled(t *testing.T) {
	notifRepo := new(MockNotificationRepository)
	userRepo := new(MockUserRepoForNotification)
	pushSvc := new(MockPushSvcForNotification)
	emailSvc := new(MockEmailSvcForNotification)
	emailTokenSvc := new(MockEmailTokenSvcForNotification)

	svc := newTestNotificationService(notifRepo, userRepo, pushSvc, emailSvc, emailTokenSvc)
	ctx := context.Background()

	recipientID := uuid.New()
	fromUserID := uuid.New()
	resourceID := uuid.New()

	recipient := &models.User{
		PushNotificationsEnabled:  false,
		PushSharingEnabled:        false,
		EmailNotificationsEnabled: false, // channel disabled
		EmailSharingEnabled:       true,
	}
	recipient.ID = recipientID
	recipient.Email = "test@example.com"

	notifRepo.On("Create", ctx, mock.Anything).Return(nil)
	userRepo.On("GetByID", ctx, recipientID).Return(recipient, nil)

	err := svc.CreateShareNotification(ctx, recipientID, fromUserID, "John", "card", resourceID, nil)

	assert.NoError(t, err)
	emailSvc.AssertNotCalled(t, "SendShareNotification")
}

// TestNotificationService_ShareEmailGating_BothEnabled verifies that
// share emails ARE sent when both EmailNotificationsEnabled and EmailSharingEnabled are true
func TestNotificationService_ShareEmailGating_BothEnabled(t *testing.T) {
	notifRepo := new(MockNotificationRepository)
	userRepo := new(MockUserRepoForNotification)
	pushSvc := new(MockPushSvcForNotification)
	emailSvc := new(MockEmailSvcForNotification)
	emailTokenSvc := new(MockEmailTokenSvcForNotification)

	svc := newTestNotificationService(notifRepo, userRepo, pushSvc, emailSvc, emailTokenSvc)
	ctx := context.Background()

	recipientID := uuid.New()
	fromUserID := uuid.New()
	resourceID := uuid.New()

	recipient := &models.User{
		PushNotificationsEnabled:  false,
		PushSharingEnabled:        false,
		EmailNotificationsEnabled: true,
		EmailSharingEnabled:       true, // both enabled
	}
	recipient.ID = recipientID
	recipient.Email = "test@example.com"
	recipient.FirstName = "Test"

	notifRepo.On("Create", ctx, mock.Anything).Return(nil)
	// GetByID is called twice: once for push gating (getRecipient), once for sendShareEmail
	userRepo.On("GetByID", ctx, recipientID).Return(recipient, nil)
	emailTokenSvc.On("CreateUnsubscribeToken", ctx, recipientID).Return("test-token", nil)
	emailSvc.On("SendShareNotification", ctx, recipient.Email, mock.Anything, "John", "card", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := svc.CreateShareNotification(ctx, recipientID, fromUserID, "John", "card", resourceID, nil)

	assert.NoError(t, err)
	emailSvc.AssertCalled(t, "SendShareNotification", ctx, recipient.Email, mock.Anything, "John", "card", mock.Anything, mock.Anything, mock.Anything)
}

// TestNotificationService_TransferPushGating_PushSharingDisabled verifies that
// transfer push notifications are NOT sent when PushSharingEnabled is false
func TestNotificationService_TransferPushGating_PushSharingDisabled(t *testing.T) {
	notifRepo := new(MockNotificationRepository)
	userRepo := new(MockUserRepoForNotification)
	pushSvc := new(MockPushSvcForNotification)
	emailSvc := new(MockEmailSvcForNotification)
	emailTokenSvc := new(MockEmailTokenSvcForNotification)

	svc := newTestNotificationService(notifRepo, userRepo, pushSvc, emailSvc, emailTokenSvc)
	ctx := context.Background()

	recipientID := uuid.New()
	fromUserID := uuid.New()
	resourceID := uuid.New()

	recipient := &models.User{
		PushNotificationsEnabled:  true,
		PushSharingEnabled:        false, // sharing disabled
		EmailNotificationsEnabled: false,
		EmailSharingEnabled:       false,
	}
	recipient.ID = recipientID
	recipient.Email = "test@example.com"

	notifRepo.On("Create", ctx, mock.Anything).Return(nil)
	userRepo.On("GetByID", ctx, recipientID).Return(recipient, nil)

	err := svc.CreateTransferNotification(ctx, recipientID, fromUserID, "John", "voucher", resourceID)

	assert.NoError(t, err)
	pushSvc.AssertNotCalled(t, "SendPushToUser")
}

// TestNotificationService_TransferEmailGating_EmailSharingDisabled verifies that
// transfer emails are NOT sent when EmailSharingEnabled is false
func TestNotificationService_TransferEmailGating_EmailSharingDisabled(t *testing.T) {
	notifRepo := new(MockNotificationRepository)
	userRepo := new(MockUserRepoForNotification)
	pushSvc := new(MockPushSvcForNotification)
	emailSvc := new(MockEmailSvcForNotification)
	emailTokenSvc := new(MockEmailTokenSvcForNotification)

	svc := newTestNotificationService(notifRepo, userRepo, pushSvc, emailSvc, emailTokenSvc)
	ctx := context.Background()

	recipientID := uuid.New()
	fromUserID := uuid.New()
	resourceID := uuid.New()

	recipient := &models.User{
		PushNotificationsEnabled:  false,
		PushSharingEnabled:        false,
		EmailNotificationsEnabled: true,
		EmailSharingEnabled:       false, // sharing disabled
	}
	recipient.ID = recipientID
	recipient.Email = "test@example.com"

	notifRepo.On("Create", ctx, mock.Anything).Return(nil)
	userRepo.On("GetByID", ctx, recipientID).Return(recipient, nil)

	err := svc.CreateTransferNotification(ctx, recipientID, fromUserID, "John", "gift_card", resourceID)

	assert.NoError(t, err)
	emailSvc.AssertNotCalled(t, "SendTransferNotification")
}

// TestNotificationService_TransferGating_AllEnabled verifies that both push and email
// are sent for transfer notifications when all preferences are enabled
func TestNotificationService_TransferGating_AllEnabled(t *testing.T) {
	notifRepo := new(MockNotificationRepository)
	userRepo := new(MockUserRepoForNotification)
	pushSvc := new(MockPushSvcForNotification)
	emailSvc := new(MockEmailSvcForNotification)
	emailTokenSvc := new(MockEmailTokenSvcForNotification)

	svc := newTestNotificationService(notifRepo, userRepo, pushSvc, emailSvc, emailTokenSvc)
	ctx := context.Background()

	recipientID := uuid.New()
	fromUserID := uuid.New()
	resourceID := uuid.New()

	recipient := &models.User{
		PushNotificationsEnabled:  true,
		PushSharingEnabled:        true,
		EmailNotificationsEnabled: true,
		EmailSharingEnabled:       true,
	}
	recipient.ID = recipientID
	recipient.Email = "test@example.com"
	recipient.FirstName = "Test"

	notifRepo.On("Create", ctx, mock.Anything).Return(nil)
	userRepo.On("GetByID", ctx, recipientID).Return(recipient, nil)
	pushSvc.On("SendPushToUser", ctx, recipientID, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	emailTokenSvc.On("CreateUnsubscribeToken", ctx, recipientID).Return("test-token", nil)
	emailSvc.On("SendTransferNotification", ctx, recipient.Email, mock.Anything, "John", "gift_card", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := svc.CreateTransferNotification(ctx, recipientID, fromUserID, "John", "gift_card", resourceID)

	assert.NoError(t, err)
	pushSvc.AssertCalled(t, "SendPushToUser", ctx, recipientID, mock.Anything, mock.Anything, mock.Anything)
	emailSvc.AssertCalled(t, "SendTransferNotification", ctx, recipient.Email, mock.Anything, "John", "gift_card", mock.Anything, mock.Anything, mock.Anything)
}

// ============================================================================
// SetPushService / SetEmailService Tests
// ============================================================================

func TestNotificationService_SetPushService(t *testing.T) {
	notifRepo := new(MockNotificationRepository)
	userRepo := new(MockUserRepoForNotification)
	svc := NewNotificationService(notifRepo, userRepo)

	pushSvc := new(MockPushSvcForNotification)
	svc.SetPushService(pushSvc)

	concrete := svc.(*NotificationService)
	assert.NotNil(t, concrete.pushService)
}

func TestNotificationService_SetEmailService(t *testing.T) {
	notifRepo := new(MockNotificationRepository)
	userRepo := new(MockUserRepoForNotification)
	svc := NewNotificationService(notifRepo, userRepo)

	emailSvc := new(MockEmailSvcForNotification)
	emailTokenSvc := new(MockEmailTokenSvcForNotification)
	svc.SetEmailService(emailSvc, emailTokenSvc, "https://example.com")

	concrete := svc.(*NotificationService)
	assert.NotNil(t, concrete.emailService)
	assert.NotNil(t, concrete.emailTokenService)
	assert.Equal(t, "https://example.com", concrete.frontendURL)
}

// ============================================================================
// sendPush disabled preferences Tests
// ============================================================================

func TestNotificationService_CreateShareNotification_PushDisabled(t *testing.T) {
	notifRepo := new(MockNotificationRepository)
	userRepo := new(MockUserRepoForNotification)
	svc := NewNotificationService(notifRepo, userRepo)

	pushSvc := new(MockPushSvcForNotification)
	svc.SetPushService(pushSvc)

	ctx := context.Background()
	recipientID := uuid.New()
	fromUserID := uuid.New()
	resourceID := uuid.New()

	recipient := &models.User{
		PushNotificationsEnabled: false, // Push globally disabled
		PushSharingEnabled:       true,
	}
	recipient.ID = recipientID

	notifRepo.On("Create", ctx, mock.Anything).Return(nil)
	userRepo.On("GetByID", ctx, recipientID).Return(recipient, nil)

	err := svc.CreateShareNotification(ctx, recipientID, fromUserID, "John", "card", resourceID, map[string]bool{"can_edit": true})

	assert.NoError(t, err)
	pushSvc.AssertNotCalled(t, "SendPushToUser")
}

func TestNotificationService_CreateShareNotification_PushSharingDisabled(t *testing.T) {
	notifRepo := new(MockNotificationRepository)
	userRepo := new(MockUserRepoForNotification)
	svc := NewNotificationService(notifRepo, userRepo)

	pushSvc := new(MockPushSvcForNotification)
	svc.SetPushService(pushSvc)

	ctx := context.Background()
	recipientID := uuid.New()
	fromUserID := uuid.New()
	resourceID := uuid.New()

	recipient := &models.User{
		PushNotificationsEnabled: true,
		PushSharingEnabled:       false, // Sharing push disabled
	}
	recipient.ID = recipientID

	notifRepo.On("Create", ctx, mock.Anything).Return(nil)
	userRepo.On("GetByID", ctx, recipientID).Return(recipient, nil)

	err := svc.CreateShareNotification(ctx, recipientID, fromUserID, "John", "card", resourceID, map[string]bool{"can_edit": true})

	assert.NoError(t, err)
	pushSvc.AssertNotCalled(t, "SendPushToUser")
}

// ============================================================================
// ADDITIONAL COVERAGE TESTS
// ============================================================================

// TestNotificationService_CreateTransferNotification_WithPushAndEmail tests the full
// transfer notification flow with push enabled and email disabled
func TestNotificationService_CreateTransferNotification_WithPushEnabled(t *testing.T) {
	notifRepo := new(MockNotificationRepository)
	userRepo := new(MockUserRepoForNotification)
	pushSvc := new(MockPushSvcForNotification)
	emailSvc := new(MockEmailSvcForNotification)
	emailTokenSvc := new(MockEmailTokenSvcForNotification)

	svc := newTestNotificationService(notifRepo, userRepo, pushSvc, emailSvc, emailTokenSvc)
	ctx := context.Background()

	recipientID := uuid.New()
	fromUserID := uuid.New()
	resourceID := uuid.New()

	recipient := &models.User{
		PushNotificationsEnabled:  true,
		PushSharingEnabled:        true,
		EmailNotificationsEnabled: false, // email disabled
		EmailSharingEnabled:       false,
	}
	recipient.ID = recipientID
	recipient.Email = "test@example.com"

	notifRepo.On("Create", ctx, mock.Anything).Return(nil)
	userRepo.On("GetByID", ctx, recipientID).Return(recipient, nil)
	pushSvc.On("SendPushToUser", ctx, recipientID, mock.Anything, mock.Anything, "/vouchers").Return(nil)

	err := svc.CreateTransferNotification(ctx, recipientID, fromUserID, "Alice", "voucher", resourceID)

	assert.NoError(t, err)
	pushSvc.AssertCalled(t, "SendPushToUser", ctx, recipientID, mock.Anything, mock.Anything, "/vouchers")
	emailSvc.AssertNotCalled(t, "SendTransferNotification")
}

// TestNotificationService_CreateTransferNotification_PushAndEmailDisabled tests transfer
// when both push and email are disabled
func TestNotificationService_CreateTransferNotification_PushAndEmailDisabled(t *testing.T) {
	notifRepo := new(MockNotificationRepository)
	userRepo := new(MockUserRepoForNotification)
	pushSvc := new(MockPushSvcForNotification)
	emailSvc := new(MockEmailSvcForNotification)
	emailTokenSvc := new(MockEmailTokenSvcForNotification)

	svc := newTestNotificationService(notifRepo, userRepo, pushSvc, emailSvc, emailTokenSvc)
	ctx := context.Background()

	recipientID := uuid.New()
	fromUserID := uuid.New()
	resourceID := uuid.New()

	recipient := &models.User{
		PushNotificationsEnabled:  false,
		PushSharingEnabled:        false,
		EmailNotificationsEnabled: false,
		EmailSharingEnabled:       false,
	}
	recipient.ID = recipientID
	recipient.Email = "test@example.com"

	notifRepo.On("Create", ctx, mock.Anything).Return(nil)
	userRepo.On("GetByID", ctx, recipientID).Return(recipient, nil)

	err := svc.CreateTransferNotification(ctx, recipientID, fromUserID, "Alice", "card", resourceID)

	assert.NoError(t, err)
	pushSvc.AssertNotCalled(t, "SendPushToUser")
	emailSvc.AssertNotCalled(t, "SendTransferNotification")
}

// TestNotificationService_SendPush_Error verifies that push errors are logged but do not
// propagate to the caller (best-effort delivery)
func TestNotificationService_SendPush_Error(t *testing.T) {
	notifRepo := new(MockNotificationRepository)
	userRepo := new(MockUserRepoForNotification)
	pushSvc := new(MockPushSvcForNotification)
	emailSvc := new(MockEmailSvcForNotification)
	emailTokenSvc := new(MockEmailTokenSvcForNotification)

	svc := newTestNotificationService(notifRepo, userRepo, pushSvc, emailSvc, emailTokenSvc)
	ctx := context.Background()

	recipientID := uuid.New()
	fromUserID := uuid.New()
	resourceID := uuid.New()

	recipient := &models.User{
		PushNotificationsEnabled:  true,
		PushSharingEnabled:        true,
		EmailNotificationsEnabled: false,
		EmailSharingEnabled:       false,
	}
	recipient.ID = recipientID
	recipient.Email = "test@example.com"

	notifRepo.On("Create", ctx, mock.Anything).Return(nil)
	userRepo.On("GetByID", ctx, recipientID).Return(recipient, nil)
	pushSvc.On("SendPushToUser", ctx, recipientID, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("push failed"))

	// Error should not propagate - notification creation still succeeds
	err := svc.CreateShareNotification(ctx, recipientID, fromUserID, "John", "card", resourceID, nil)

	assert.NoError(t, err)
	pushSvc.AssertCalled(t, "SendPushToUser", ctx, recipientID, mock.Anything, mock.Anything, mock.Anything)
}

// TestNotificationService_SendPush_NilPushService verifies that sendPush is a no-op
// when pushService is nil
func TestNotificationService_SendPush_NilPushService(t *testing.T) {
	notifRepo := new(MockNotificationRepository)
	service := NewNotificationService(notifRepo, nil)
	ctx := context.Background()

	recipientID := uuid.New()
	fromUserID := uuid.New()
	resourceID := uuid.New()

	notifRepo.On("Create", ctx, mock.Anything).Return(nil)

	// pushService is nil, should not panic
	err := service.CreateTransferNotification(ctx, recipientID, fromUserID, "John", "card", resourceID)

	assert.NoError(t, err)
}

// TestNotificationService_SendShareEmail_UserRepoError verifies that sendShareEmail
// handles userRepo.GetByID errors gracefully (best-effort)
func TestNotificationService_SendShareEmail_UserRepoError(t *testing.T) {
	notifRepo := new(MockNotificationRepository)
	userRepo := new(MockUserRepoForNotification)
	emailSvc := new(MockEmailSvcForNotification)
	emailTokenSvc := new(MockEmailTokenSvcForNotification)

	svc := &NotificationService{
		repo:              notifRepo,
		userRepo:          userRepo,
		emailService:      emailSvc,
		emailTokenService: emailTokenSvc,
		frontendURL:       "https://app.test",
	}
	ctx := context.Background()

	recipientID := uuid.New()
	fromUserID := uuid.New()
	resourceID := uuid.New()

	notifRepo.On("Create", ctx, mock.Anything).Return(nil)
	// getRecipient call returns nil (error)
	userRepo.On("GetByID", ctx, recipientID).Return(nil, errors.New("user not found"))

	// Email should not be called because userRepo.GetByID failed
	err := svc.CreateShareNotification(ctx, recipientID, fromUserID, "John", "card", resourceID, nil)

	assert.NoError(t, err)
	emailSvc.AssertNotCalled(t, "SendShareNotification")
}

// TestNotificationService_SendTransferEmail_UserRepoError verifies that sendTransferEmail
// handles userRepo.GetByID errors gracefully (best-effort)
func TestNotificationService_SendTransferEmail_UserRepoError(t *testing.T) {
	notifRepo := new(MockNotificationRepository)
	userRepo := new(MockUserRepoForNotification)
	emailSvc := new(MockEmailSvcForNotification)
	emailTokenSvc := new(MockEmailTokenSvcForNotification)

	svc := &NotificationService{
		repo:              notifRepo,
		userRepo:          userRepo,
		emailService:      emailSvc,
		emailTokenService: emailTokenSvc,
		frontendURL:       "https://app.test",
	}
	ctx := context.Background()

	recipientID := uuid.New()
	fromUserID := uuid.New()
	resourceID := uuid.New()

	notifRepo.On("Create", ctx, mock.Anything).Return(nil)
	userRepo.On("GetByID", ctx, recipientID).Return(nil, errors.New("user not found"))

	// Email should not be called because userRepo.GetByID failed
	err := svc.CreateTransferNotification(ctx, recipientID, fromUserID, "John", "voucher", resourceID)

	assert.NoError(t, err)
	emailSvc.AssertNotCalled(t, "SendTransferNotification")
}

// TestNotificationService_SendShareEmail_NilEmailService verifies sendShareEmail
// is a no-op when emailService is nil
func TestNotificationService_SendShareEmail_NilEmailService(t *testing.T) {
	notifRepo := new(MockNotificationRepository)
	userRepo := new(MockUserRepoForNotification)

	svc := &NotificationService{
		repo:     notifRepo,
		userRepo: userRepo,
		// emailService is nil
	}
	ctx := context.Background()

	recipientID := uuid.New()
	fromUserID := uuid.New()
	resourceID := uuid.New()

	notifRepo.On("Create", ctx, mock.Anything).Return(nil)
	userRepo.On("GetByID", ctx, recipientID).Return(nil, errors.New("not found"))

	err := svc.CreateShareNotification(ctx, recipientID, fromUserID, "John", "card", resourceID, nil)

	assert.NoError(t, err)
}

// TestNotificationService_GenerateUnsubscribeURL_NilTokenService verifies that
// generateUnsubscribeURL returns empty string when emailTokenService is nil
func TestNotificationService_GenerateUnsubscribeURL_NilTokenService(t *testing.T) {
	notifRepo := new(MockNotificationRepository)
	userRepo := new(MockUserRepoForNotification)
	emailSvc := new(MockEmailSvcForNotification)

	svc := &NotificationService{
		repo:         notifRepo,
		userRepo:     userRepo,
		emailService: emailSvc,
		frontendURL:  "https://app.test",
		// emailTokenService is nil
	}
	ctx := context.Background()

	recipientID := uuid.New()
	fromUserID := uuid.New()
	resourceID := uuid.New()

	recipient := &models.User{
		PushNotificationsEnabled:  false,
		PushSharingEnabled:        false,
		EmailNotificationsEnabled: true,
		EmailSharingEnabled:       true,
	}
	recipient.ID = recipientID
	recipient.Email = "test@example.com"
	recipient.FirstName = "Test"

	notifRepo.On("Create", ctx, mock.Anything).Return(nil)
	userRepo.On("GetByID", ctx, recipientID).Return(recipient, nil)
	// Unsubscribe URL will be "" because emailTokenService is nil
	emailSvc.On("SendShareNotification", ctx, recipient.Email, mock.Anything, "John", "card", mock.Anything, "", mock.Anything).Return(nil)

	err := svc.CreateShareNotification(ctx, recipientID, fromUserID, "John", "card", resourceID, nil)

	assert.NoError(t, err)
	emailSvc.AssertCalled(t, "SendShareNotification", ctx, recipient.Email, mock.Anything, "John", "card", mock.Anything, "", mock.Anything)
}

// TestNotificationService_GenerateUnsubscribeURL_TokenError verifies that
// generateUnsubscribeURL returns empty string when token creation fails
func TestNotificationService_GenerateUnsubscribeURL_TokenError(t *testing.T) {
	notifRepo := new(MockNotificationRepository)
	userRepo := new(MockUserRepoForNotification)
	emailSvc := new(MockEmailSvcForNotification)
	emailTokenSvc := new(MockEmailTokenSvcForNotification)

	svc := newTestNotificationService(notifRepo, userRepo, nil, emailSvc, emailTokenSvc)
	ctx := context.Background()

	recipientID := uuid.New()
	fromUserID := uuid.New()
	resourceID := uuid.New()

	recipient := &models.User{
		PushNotificationsEnabled:  false,
		PushSharingEnabled:        false,
		EmailNotificationsEnabled: true,
		EmailSharingEnabled:       true,
	}
	recipient.ID = recipientID
	recipient.Email = "test@example.com"
	recipient.FirstName = "Test"

	notifRepo.On("Create", ctx, mock.Anything).Return(nil)
	userRepo.On("GetByID", ctx, recipientID).Return(recipient, nil)
	emailTokenSvc.On("CreateUnsubscribeToken", ctx, recipientID).Return("", errors.New("token creation failed"))
	// Unsubscribe URL will be "" because token creation failed
	emailSvc.On("SendTransferNotification", ctx, recipient.Email, mock.Anything, "Bob", "gift_card", mock.Anything, "", mock.Anything).Return(nil)

	err := svc.CreateTransferNotification(ctx, recipientID, fromUserID, "Bob", "gift_card", resourceID)

	assert.NoError(t, err)
	emailSvc.AssertCalled(t, "SendTransferNotification", ctx, recipient.Email, mock.Anything, "Bob", "gift_card", mock.Anything, "", mock.Anything)
}

// TestNotificationService_SendShareEmail_SendError verifies that email send errors
// do not propagate (best-effort delivery)
func TestNotificationService_SendShareEmail_SendError(t *testing.T) {
	notifRepo := new(MockNotificationRepository)
	userRepo := new(MockUserRepoForNotification)
	emailSvc := new(MockEmailSvcForNotification)
	emailTokenSvc := new(MockEmailTokenSvcForNotification)

	svc := newTestNotificationService(notifRepo, userRepo, nil, emailSvc, emailTokenSvc)
	ctx := context.Background()

	recipientID := uuid.New()
	fromUserID := uuid.New()
	resourceID := uuid.New()

	recipient := &models.User{
		PushNotificationsEnabled:  false,
		PushSharingEnabled:        false,
		EmailNotificationsEnabled: true,
		EmailSharingEnabled:       true,
	}
	recipient.ID = recipientID
	recipient.Email = "test@example.com"
	recipient.FirstName = "Test"

	notifRepo.On("Create", ctx, mock.Anything).Return(nil)
	userRepo.On("GetByID", ctx, recipientID).Return(recipient, nil)
	emailTokenSvc.On("CreateUnsubscribeToken", ctx, recipientID).Return("tok", nil)
	emailSvc.On("SendShareNotification", ctx, recipient.Email, mock.Anything, "John", "voucher", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("smtp error"))

	// Error should not propagate
	err := svc.CreateShareNotification(ctx, recipientID, fromUserID, "John", "voucher", resourceID, nil)

	assert.NoError(t, err)
}

// TestNotificationService_SendTransferEmail_SendError verifies that transfer email
// send errors do not propagate (best-effort delivery)
func TestNotificationService_SendTransferEmail_SendError(t *testing.T) {
	notifRepo := new(MockNotificationRepository)
	userRepo := new(MockUserRepoForNotification)
	emailSvc := new(MockEmailSvcForNotification)
	emailTokenSvc := new(MockEmailTokenSvcForNotification)

	svc := newTestNotificationService(notifRepo, userRepo, nil, emailSvc, emailTokenSvc)
	ctx := context.Background()

	recipientID := uuid.New()
	fromUserID := uuid.New()
	resourceID := uuid.New()

	recipient := &models.User{
		PushNotificationsEnabled:  false,
		PushSharingEnabled:        false,
		EmailNotificationsEnabled: true,
		EmailSharingEnabled:       true,
	}
	recipient.ID = recipientID
	recipient.Email = "test@example.com"
	recipient.FirstName = "Test"

	notifRepo.On("Create", ctx, mock.Anything).Return(nil)
	userRepo.On("GetByID", ctx, recipientID).Return(recipient, nil)
	emailTokenSvc.On("CreateUnsubscribeToken", ctx, recipientID).Return("tok", nil)
	emailSvc.On("SendTransferNotification", ctx, recipient.Email, mock.Anything, "Bob", "card", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("smtp error"))

	// Error should not propagate
	err := svc.CreateTransferNotification(ctx, recipientID, fromUserID, "Bob", "card", resourceID)

	assert.NoError(t, err)
}

// TestNotificationService_GetRecipient_NilUserRepo verifies getRecipient returns nil
// when userRepo is nil
func TestNotificationService_GetRecipient_NilUserRepo(t *testing.T) {
	notifRepo := new(MockNotificationRepository)

	// userRepo is nil
	svc := &NotificationService{
		repo: notifRepo,
	}
	ctx := context.Background()

	recipientID := uuid.New()
	fromUserID := uuid.New()
	resourceID := uuid.New()

	notifRepo.On("Create", ctx, mock.Anything).Return(nil)

	// With nil userRepo, recipient lookup returns nil → push is still sent (nil check passes)
	err := svc.CreateTransferNotification(ctx, recipientID, fromUserID, "John", "card", resourceID)

	assert.NoError(t, err)
}

// TestNotificationService_CreateTransferNotification_VoucherResourceURL verifies the
// resourceURL construction for vouchers
func TestNotificationService_CreateTransferNotification_VoucherResourceURL(t *testing.T) {
	notifRepo := new(MockNotificationRepository)
	userRepo := new(MockUserRepoForNotification)
	pushSvc := new(MockPushSvcForNotification)

	svc := &NotificationService{
		repo:        notifRepo,
		userRepo:    userRepo,
		pushService: pushSvc,
	}
	ctx := context.Background()

	recipientID := uuid.New()
	fromUserID := uuid.New()
	resourceID := uuid.New()

	recipient := &models.User{
		PushNotificationsEnabled: true,
		PushSharingEnabled:       true,
	}
	recipient.ID = recipientID

	notifRepo.On("Create", ctx, mock.Anything).Return(nil)
	userRepo.On("GetByID", ctx, recipientID).Return(recipient, nil)
	// Verify the URL ends with "/vouchers"
	pushSvc.On("SendPushToUser", ctx, recipientID, mock.Anything, mock.Anything, "/vouchers").Return(nil)

	err := svc.CreateTransferNotification(ctx, recipientID, fromUserID, "User", "voucher", resourceID)

	assert.NoError(t, err)
	pushSvc.AssertCalled(t, "SendPushToUser", ctx, recipientID, mock.Anything, mock.Anything, "/vouchers")
}

// TestNotificationService_CreateShareNotification_GiftCardResourceURL verifies the
// resourceURL construction for gift cards
func TestNotificationService_CreateShareNotification_GiftCardResourceURL(t *testing.T) {
	notifRepo := new(MockNotificationRepository)
	userRepo := new(MockUserRepoForNotification)
	pushSvc := new(MockPushSvcForNotification)

	svc := &NotificationService{
		repo:        notifRepo,
		userRepo:    userRepo,
		pushService: pushSvc,
	}
	ctx := context.Background()

	recipientID := uuid.New()
	fromUserID := uuid.New()
	resourceID := uuid.New()

	recipient := &models.User{
		PushNotificationsEnabled: true,
		PushSharingEnabled:       true,
	}
	recipient.ID = recipientID

	notifRepo.On("Create", ctx, mock.Anything).Return(nil)
	userRepo.On("GetByID", ctx, recipientID).Return(recipient, nil)
	// Verify the URL is the hyphenated "/gift-cards" route that actually exists
	// in the client — "/gift_cards" would be served as the SPA shell (HTTP 200)
	// and render a white screen.
	pushSvc.On("SendPushToUser", ctx, recipientID, mock.Anything, mock.Anything, "/gift-cards").Return(nil)

	err := svc.CreateShareNotification(ctx, recipientID, fromUserID, "User", "gift_card", resourceID, nil)

	assert.NoError(t, err)
	pushSvc.AssertCalled(t, "SendPushToUser", ctx, recipientID, mock.Anything, mock.Anything, "/gift-cards")
}
