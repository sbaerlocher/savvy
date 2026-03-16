package services

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"savvy/internal/models"
	"savvy/internal/repository"
)

// ==================== Mock Push Subscription Repository ====================

type MockPushSubscriptionRepo struct {
	mock.Mock
}

var _ repository.PushSubscriptionRepository = (*MockPushSubscriptionRepo)(nil)

func (m *MockPushSubscriptionRepo) Create(ctx context.Context, sub *models.PushSubscription) error {
	return m.Called(ctx, sub).Error(0)
}

func (m *MockPushSubscriptionRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.PushSubscription, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.PushSubscription), args.Error(1)
}

func (m *MockPushSubscriptionRepo) DeleteByEndpoint(ctx context.Context, endpoint string) error {
	return m.Called(ctx, endpoint).Error(0)
}

// ==================== Mock User Repository (for Push) ====================

type MockUserRepo struct {
	mock.Mock
}

var _ repository.UserRepository = (*MockUserRepo)(nil)

func (m *MockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepo) Create(ctx context.Context, user *models.User) error {
	return m.Called(ctx, user).Error(0)
}

func (m *MockUserRepo) Update(ctx context.Context, user *models.User) error {
	return m.Called(ctx, user).Error(0)
}

func (m *MockUserRepo) GetAll(ctx context.Context) ([]models.User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockUserRepo) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]models.User, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockUserRepo) SearchByIDs(ctx context.Context, ids []uuid.UUID, query string) ([]models.User, error) {
	args := m.Called(ctx, ids, query)
	return args.Get(0).([]models.User), args.Error(1)
}

// ==================== PushService Tests ====================

func TestPushService_IsEnabled_WithKeys(t *testing.T) {
	repo := new(MockPushSubscriptionRepo)
	svc := NewPushService(repo, nil, "public-key", "private-key", "mailto:test@example.com")

	assert.True(t, svc.IsEnabled())
}

func TestPushService_IsEnabled_WithoutKeys(t *testing.T) {
	repo := new(MockPushSubscriptionRepo)
	svc := NewPushService(repo, nil, "", "", "")

	assert.False(t, svc.IsEnabled())
}

func TestPushService_IsEnabled_PartialKeys(t *testing.T) {
	repo := new(MockPushSubscriptionRepo)

	svc1 := NewPushService(repo, nil, "public", "", "")
	assert.False(t, svc1.IsEnabled())

	svc2 := NewPushService(repo, nil, "", "private", "")
	assert.False(t, svc2.IsEnabled())

	svc3 := NewPushService(repo, nil, "public", "private", "")
	assert.False(t, svc3.IsEnabled())
}

func TestPushService_GetVAPIDPublicKey(t *testing.T) {
	repo := new(MockPushSubscriptionRepo)
	svc := NewPushService(repo, nil, "my-public-key", "private", "mailto:test@example.com")

	assert.Equal(t, "my-public-key", svc.GetVAPIDPublicKey())
}

func TestPushService_Subscribe_Success_FirstSubscription(t *testing.T) {
	repo := new(MockPushSubscriptionRepo)
	userRepo := new(MockUserRepo)
	svc := NewPushService(repo, userRepo, "pub", "priv", "mailto:test@example.com")
	ctx := context.Background()
	userID := uuid.New()

	// First subscription: no existing subs
	repo.On("GetByUserID", ctx, userID).Return([]models.PushSubscription{}, nil)
	repo.On("Create", ctx, mock.AnythingOfType("*models.PushSubscription")).Return(nil)

	// Should enable push preferences
	user := &models.User{ID: userID, PushNotificationsEnabled: false}
	userRepo.On("GetByID", ctx, userID).Return(user, nil)
	userRepo.On("Update", ctx, mock.MatchedBy(func(u *models.User) bool {
		return u.PushNotificationsEnabled && u.PushRemindersEnabled && u.PushSharingEnabled
	})).Return(nil)

	err := svc.Subscribe(ctx, userID, "https://push.example.com/sub1", "p256dh-key", "auth-key", "Mozilla/5.0")

	assert.NoError(t, err)
	repo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestPushService_Subscribe_Success_ExistingSubscription(t *testing.T) {
	repo := new(MockPushSubscriptionRepo)
	userRepo := new(MockUserRepo)
	svc := NewPushService(repo, userRepo, "pub", "priv", "mailto:test@example.com")
	ctx := context.Background()
	userID := uuid.New()

	// Existing subscription: should NOT update preferences
	repo.On("GetByUserID", ctx, userID).Return([]models.PushSubscription{{UserID: userID}}, nil)
	repo.On("Create", ctx, mock.AnythingOfType("*models.PushSubscription")).Return(nil)

	err := svc.Subscribe(ctx, userID, "https://push.example.com/sub2", "p256dh-key", "auth-key", "Mozilla/5.0")

	assert.NoError(t, err)
	repo.AssertExpectations(t)
	userRepo.AssertNotCalled(t, "GetByID")
}

func TestPushService_Subscribe_NilUserRepo(t *testing.T) {
	repo := new(MockPushSubscriptionRepo)
	svc := NewPushService(repo, nil, "pub", "priv", "mailto:test@example.com")
	ctx := context.Background()
	userID := uuid.New()

	repo.On("GetByUserID", ctx, userID).Return([]models.PushSubscription{}, nil)
	repo.On("Create", ctx, mock.AnythingOfType("*models.PushSubscription")).Return(nil)

	err := svc.Subscribe(ctx, userID, "https://push.example.com/sub1", "p256dh-key", "auth-key", "Mozilla/5.0")

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestPushService_Unsubscribe_Success(t *testing.T) {
	repo := new(MockPushSubscriptionRepo)
	svc := NewPushService(repo, nil, "pub", "priv", "mailto:test@example.com")
	ctx := context.Background()

	repo.On("DeleteByEndpoint", ctx, "https://push.example.com/sub1").Return(nil)

	err := svc.Unsubscribe(ctx, "https://push.example.com/sub1")

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestPushService_SendPushToUser_NotEnabled(t *testing.T) {
	repo := new(MockPushSubscriptionRepo)
	svc := NewPushService(repo, nil, "", "", "") // Not enabled
	ctx := context.Background()
	userID := uuid.New()

	err := svc.SendPushToUser(ctx, userID, "Title", "Body", "/url")

	assert.NoError(t, err)
	// Repo should not be called at all
	repo.AssertNotCalled(t, "GetByUserID")
}

func TestPushService_SendPushToUser_NoSubscriptions(t *testing.T) {
	repo := new(MockPushSubscriptionRepo)
	svc := NewPushService(repo, nil, "pub", "priv", "mailto:test@example.com")
	ctx := context.Background()
	userID := uuid.New()

	repo.On("GetByUserID", ctx, userID).Return([]models.PushSubscription{}, nil)

	err := svc.SendPushToUser(ctx, userID, "Title", "Body", "/url")

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestPushService_SendPushToUser_RepoError(t *testing.T) {
	repo := new(MockPushSubscriptionRepo)
	svc := NewPushService(repo, nil, "pub", "priv", "mailto:test@example.com")
	ctx := context.Background()
	userID := uuid.New()

	repo.On("GetByUserID", ctx, userID).Return(nil, assert.AnError)

	err := svc.SendPushToUser(ctx, userID, "Title", "Body", "/url")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get push subscriptions")
	repo.AssertExpectations(t)
}

// ==================== Utility Function Tests ====================

func TestToBase64URL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"abc+def/ghi=", "abc-def_ghi"},
		{"no+changes=", "no-changes"},
		{"already-safe_key", "already-safe_key"},
		{"a+b/c==", "a-b_c"},
		{"", ""},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, toBase64URL(tt.input))
	}
}

func TestTruncateEndpoint(t *testing.T) {
	// Short endpoint (<=60 chars)
	short := "https://push.example.com/sub/123"
	assert.Equal(t, short, truncateEndpoint(short))

	// Exactly 60 chars
	exact60 := "https://push.example.com/subscription/abcdefghij1234567890ab"
	assert.Equal(t, exact60, truncateEndpoint(exact60))

	// Long endpoint (>60 chars)
	long := "https://fcm.googleapis.com/fcm/send/very-long-subscription-id-that-exceeds-sixty-characters-in-length"
	result := truncateEndpoint(long)
	assert.Contains(t, result, "...")
	assert.Len(t, result, 53) // 30 + 3 + 20
}

func TestPushService_IsEnabled(t *testing.T) {
	repo := new(MockPushSubscriptionRepo)

	// Enabled when all VAPID params are set
	svc := NewPushService(repo, nil, "pub", "priv", "mailto:test@example.com")
	assert.True(t, svc.IsEnabled())

	// Disabled when public key is empty
	svc2 := NewPushService(repo, nil, "", "priv", "mailto:test@example.com")
	assert.False(t, svc2.IsEnabled())
}

func TestPushService_GetVAPIDPublicKey_Value(t *testing.T) {
	repo := new(MockPushSubscriptionRepo)
	svc := NewPushService(repo, nil, "my-public-key", "priv", "mailto:test@example.com")
	assert.Equal(t, "my-public-key", svc.GetVAPIDPublicKey())
}

func TestPushService_SendTestPush_NotEnabled(t *testing.T) {
	repo := new(MockPushSubscriptionRepo)
	svc := NewPushService(repo, nil, "", "", "")
	ctx := context.Background()

	err := svc.SendTestPush(ctx, uuid.New())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "push notifications not enabled")
}
