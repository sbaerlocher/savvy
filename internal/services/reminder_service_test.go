package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"savvy/internal/email"
	"savvy/internal/models"
	"savvy/internal/repository"
)

// ==================== Mock Reminder Repository ====================

type MockReminderRepo struct {
	mock.Mock
}

var _ repository.ReminderRepository = (*MockReminderRepo)(nil)

func (m *MockReminderRepo) HasBeenSent(ctx context.Context, userID uuid.UUID, resourceType string, resourceID uuid.UUID, daysBefore int) (bool, error) {
	args := m.Called(ctx, userID, resourceType, resourceID, daysBefore)
	return args.Bool(0), args.Error(1)
}

func (m *MockReminderRepo) MarkSent(ctx context.Context, reminder *models.ExpiryReminderSent) error {
	return m.Called(ctx, reminder).Error(0)
}

// ==================== Mock Voucher Repository (for Reminder) ====================

type mockVoucherRepoForReminder struct {
	mock.Mock
}

var _ repository.VoucherRepository = (*mockVoucherRepoForReminder)(nil)

func (m *mockVoucherRepoForReminder) GetExpiringVouchers(ctx context.Context, withinDays int) ([]models.Voucher, error) {
	args := m.Called(ctx, withinDays)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Voucher), args.Error(1)
}
func (m *mockVoucherRepoForReminder) GetVouchersStartingTomorrow(ctx context.Context) ([]models.Voucher, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Voucher), args.Error(1)
}

// Stubs for remaining VoucherRepository methods
func (m *mockVoucherRepoForReminder) Create(_ context.Context, _ *models.Voucher) error { return nil }
func (m *mockVoucherRepoForReminder) GetByID(_ context.Context, _ uuid.UUID, _ ...string) (*models.Voucher, error) {
	return nil, nil
}
func (m *mockVoucherRepoForReminder) GetByUserID(_ context.Context, _ uuid.UUID) ([]models.Voucher, error) {
	return nil, nil
}
func (m *mockVoucherRepoForReminder) GetSharedWithUser(_ context.Context, _ uuid.UUID) ([]models.Voucher, error) {
	return nil, nil
}
func (m *mockVoucherRepoForReminder) Update(_ context.Context, _ *models.Voucher) error { return nil }
func (m *mockVoucherRepoForReminder) Delete(_ context.Context, _ uuid.UUID) error       { return nil }
func (m *mockVoucherRepoForReminder) Count(_ context.Context, _ uuid.UUID) (int64, error) {
	return 0, nil
}
func (m *mockVoucherRepoForReminder) GetAllForUserPaginated(_ context.Context, _ uuid.UUID, _ repository.PaginationParams) (*repository.PaginatedResult[models.Voucher], error) {
	return nil, nil
}
func (m *mockVoucherRepoForReminder) FindByVoucherCode(_ context.Context, _ string, _ uuid.UUID) (*models.Voucher, error) {
	return nil, nil
}
func (m *mockVoucherRepoForReminder) FindDeletedByCode(_ context.Context, _ string, _ uuid.UUID) (*models.Voucher, error) {
	return nil, nil
}
func (m *mockVoucherRepoForReminder) RestoreByID(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
	return nil
}
func (m *mockVoucherRepoForReminder) Search(_ context.Context, _ uuid.UUID, _ string) ([]models.Voucher, error) {
	return nil, nil
}

// ==================== Mock Gift Card Repository (for Reminder) ====================

type mockGiftCardRepoForReminder struct {
	mock.Mock
}

var _ repository.GiftCardRepository = (*mockGiftCardRepoForReminder)(nil)

func (m *mockGiftCardRepoForReminder) GetExpiringGiftCards(ctx context.Context, withinDays int) ([]models.GiftCard, error) {
	args := m.Called(ctx, withinDays)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.GiftCard), args.Error(1)
}

// Stubs for remaining GiftCardRepository methods
func (m *mockGiftCardRepoForReminder) Create(_ context.Context, _ *models.GiftCard) error { return nil }
func (m *mockGiftCardRepoForReminder) GetByID(_ context.Context, _ uuid.UUID, _ ...string) (*models.GiftCard, error) {
	return nil, nil
}
func (m *mockGiftCardRepoForReminder) GetByUserID(_ context.Context, _ uuid.UUID) ([]models.GiftCard, error) {
	return nil, nil
}
func (m *mockGiftCardRepoForReminder) GetSharedWithUser(_ context.Context, _ uuid.UUID) ([]models.GiftCard, error) {
	return nil, nil
}
func (m *mockGiftCardRepoForReminder) Update(_ context.Context, _ *models.GiftCard) error { return nil }
func (m *mockGiftCardRepoForReminder) Delete(_ context.Context, _ uuid.UUID) error        { return nil }
func (m *mockGiftCardRepoForReminder) Count(_ context.Context, _ uuid.UUID) (int64, error) {
	return 0, nil
}
func (m *mockGiftCardRepoForReminder) GetTotalBalance(_ context.Context, _ uuid.UUID) (float64, error) {
	return 0, nil
}
func (m *mockGiftCardRepoForReminder) CreateTransaction(_ context.Context, _ *models.GiftCardTransaction) error {
	return nil
}
func (m *mockGiftCardRepoForReminder) GetTransaction(_ context.Context, _, _ uuid.UUID) (*models.GiftCardTransaction, error) {
	return nil, nil
}
func (m *mockGiftCardRepoForReminder) DeleteTransaction(_ context.Context, _ uuid.UUID) error {
	return nil
}
func (m *mockGiftCardRepoForReminder) GetAllForUserPaginated(_ context.Context, _ uuid.UUID, _ repository.PaginationParams) (*repository.PaginatedResult[models.GiftCard], error) {
	return nil, nil
}
func (m *mockGiftCardRepoForReminder) FindByCardNumber(_ context.Context, _ string, _ uuid.UUID) (*models.GiftCard, error) {
	return nil, nil
}
func (m *mockGiftCardRepoForReminder) FindDeletedByCardNumber(_ context.Context, _ string, _ uuid.UUID) (*models.GiftCard, error) {
	return nil, nil
}
func (m *mockGiftCardRepoForReminder) RestoreByID(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
	return nil
}
func (m *mockGiftCardRepoForReminder) Search(_ context.Context, _ uuid.UUID, _ string) ([]models.GiftCard, error) {
	return nil, nil
}

// ==================== Mock Notification Repository (for Reminder) ====================

type mockNotifRepoForReminder struct {
	mock.Mock
}

var _ repository.NotificationRepository = (*mockNotifRepoForReminder)(nil)

func (m *mockNotifRepoForReminder) Create(ctx context.Context, n *models.Notification) error {
	return m.Called(ctx, n).Error(0)
}

func (m *mockNotifRepoForReminder) GetByID(_ context.Context, _ uuid.UUID) (*models.Notification, error) {
	return nil, nil
}
func (m *mockNotifRepoForReminder) GetByUserID(_ context.Context, _ uuid.UUID, _, _ int) ([]models.Notification, error) {
	return nil, nil
}
func (m *mockNotifRepoForReminder) GetUnreadCount(_ context.Context, _ uuid.UUID) (int64, error) {
	return 0, nil
}
func (m *mockNotifRepoForReminder) MarkAsRead(_ context.Context, _, _ uuid.UUID) error { return nil }
func (m *mockNotifRepoForReminder) MarkAllAsRead(_ context.Context, _ uuid.UUID) error { return nil }
func (m *mockNotifRepoForReminder) Delete(_ context.Context, _, _ uuid.UUID) error     { return nil }
func (m *mockNotifRepoForReminder) ArchiveOldRead(_ context.Context, _ time.Time) (int64, error) {
	return 0, nil
}

// ==================== Mock Push Service (for Reminder) ====================

type mockPushSvcForReminder struct {
	mock.Mock
}

var _ PushServiceInterface = (*mockPushSvcForReminder)(nil)

func (m *mockPushSvcForReminder) SendPushToUser(ctx context.Context, userID uuid.UUID, title, body, url string) error {
	return m.Called(ctx, userID, title, body, url).Error(0)
}

func (m *mockPushSvcForReminder) Subscribe(_ context.Context, _ uuid.UUID, _, _, _, _ string) error {
	return nil
}
func (m *mockPushSvcForReminder) Unsubscribe(_ context.Context, _ string) error     { return nil }
func (m *mockPushSvcForReminder) GetVAPIDPublicKey() string                         { return "" }
func (m *mockPushSvcForReminder) IsEnabled() bool                                   { return true }
func (m *mockPushSvcForReminder) SendTestPush(_ context.Context, _ uuid.UUID) error { return nil }

// ==================== Mock Email Service (for Reminder) ====================

type mockEmailSvcForReminder struct {
	mock.Mock
}

func (m *mockEmailSvcForReminder) SendExpiryReminder(ctx context.Context, toEmail, toName string, data email.ExpiryReminderData, unsubscribeURL, language string) error {
	return m.Called(ctx, toEmail, toName, data, unsubscribeURL, language).Error(0)
}
func (m *mockEmailSvcForReminder) SendValidityStart(ctx context.Context, toEmail, toName string, data email.ValidityStartData, unsubscribeURL, language string) error {
	return m.Called(ctx, toEmail, toName, data, unsubscribeURL, language).Error(0)
}

func (m *mockEmailSvcForReminder) SendPasswordReset(_ context.Context, _, _, _, _, _ string) error {
	return nil
}
func (m *mockEmailSvcForReminder) SendEmailVerification(_ context.Context, _, _, _, _ string) error {
	return nil
}
func (m *mockEmailSvcForReminder) SendAccountDeletionConfirmation(_ context.Context, _, _, _ string) error {
	return nil
}
func (m *mockEmailSvcForReminder) CheckConnection(_ context.Context) error {
	return nil
}
func (m *mockEmailSvcForReminder) SendTestEmail(_ context.Context, _, _, _ string) error {
	return nil
}

func (m *mockEmailSvcForReminder) SendShareNotification(_ context.Context, _, _, _, _, _, _, _ string) error {
	return nil
}

func (m *mockEmailSvcForReminder) SendTransferNotification(_ context.Context, _, _, _, _, _, _, _ string) error {
	return nil
}

// ==================== Tests ====================

func setupReminderTest() (
	*ReminderService,
	*MockReminderRepo,
	*mockVoucherRepoForReminder,
	*mockGiftCardRepoForReminder,
	*mockNotifRepoForReminder,
	*mockPushSvcForReminder,
	*mockEmailSvcForReminder,
) {
	reminderRepo := new(MockReminderRepo)
	voucherRepo := new(mockVoucherRepoForReminder)
	giftCardRepo := new(mockGiftCardRepoForReminder)
	notifRepo := new(mockNotifRepoForReminder)
	pushSvc := new(mockPushSvcForReminder)
	emailSvc := new(mockEmailSvcForReminder)

	svc := NewReminderService(reminderRepo, voucherRepo, giftCardRepo, nil, nil, notifRepo, pushSvc, emailSvc, nil, []int{7, 3, 1}, time.UTC, "https://savvy.example.com")

	return svc.(*ReminderService), reminderRepo, voucherRepo, giftCardRepo, notifRepo, pushSvc, emailSvc
}

func TestReminderService_CheckAndSendReminders_Vouchers(t *testing.T) {
	svc, reminderRepo, voucherRepo, giftCardRepo, notifRepo, pushSvc, emailSvc := setupReminderTest()
	ctx := context.Background()

	userID := uuid.New()
	merchant := &models.Merchant{ID: uuid.New(), Name: "IKEA"}
	user := &models.User{ID: userID, Email: "test@example.com", FirstName: "Test", Language: "en", PushRemindersEnabled: true, EmailRemindersEnabled: true, PushNotificationsEnabled: true, EmailNotificationsEnabled: true}
	voucherID := uuid.New()

	expiringVoucher := models.Voucher{
		ID:         voucherID,
		UserID:     &userID,
		ValidUntil: time.Now().Add(3 * 24 * time.Hour),
		Merchant:   merchant,
		User:       user,
	}

	// For each daysBefore (7, 3, 1):
	// - GetExpiringVouchers called
	// - GetExpiringGiftCards called

	// Day 7: no expiring vouchers
	voucherRepo.On("GetExpiringVouchers", ctx, 7).Return([]models.Voucher{}, nil)
	giftCardRepo.On("GetExpiringGiftCards", ctx, 7).Return([]models.GiftCard{}, nil)

	// Day 3: one expiring voucher
	voucherRepo.On("GetExpiringVouchers", ctx, 3).Return([]models.Voucher{expiringVoucher}, nil)
	giftCardRepo.On("GetExpiringGiftCards", ctx, 3).Return([]models.GiftCard{}, nil)

	// Day 1: no expiring
	voucherRepo.On("GetExpiringVouchers", ctx, 1).Return([]models.Voucher{}, nil)
	giftCardRepo.On("GetExpiringGiftCards", ctx, 1).Return([]models.GiftCard{}, nil)

	reminderRepo.On("HasBeenSent", ctx, userID, "voucher", voucherID, 3).Return(false, nil)
	notifRepo.On("Create", ctx, mock.AnythingOfType("*models.Notification")).Return(nil)
	pushSvc.On("SendPushToUser", ctx, userID, mock.AnythingOfType("string"), mock.AnythingOfType("string"), "/vouchers").Return(nil)
	emailSvc.On("SendExpiryReminder", ctx, "test@example.com", "Test", mock.AnythingOfType("email.ExpiryReminderData"), "", "en").Return(nil)
	reminderRepo.On("MarkSent", ctx, mock.AnythingOfType("*models.ExpiryReminderSent")).Return(nil)
	voucherRepo.On("GetVouchersStartingTomorrow", ctx).Return([]models.Voucher{}, nil)

	err := svc.CheckAndSendReminders(ctx)

	assert.NoError(t, err)
	notifRepo.AssertExpectations(t)
	reminderRepo.AssertExpectations(t)
	pushSvc.AssertExpectations(t)
	emailSvc.AssertExpectations(t)
}

func TestReminderService_CheckAndSendReminders_GiftCards(t *testing.T) {
	svc, reminderRepo, voucherRepo, giftCardRepo, notifRepo, pushSvc, emailSvc := setupReminderTest()
	ctx := context.Background()

	userID := uuid.New()
	merchant := &models.Merchant{ID: uuid.New(), Name: "Amazon"}
	user := &models.User{ID: userID, Email: "test@example.com", FirstName: "Test", Language: "de", PushRemindersEnabled: true, EmailRemindersEnabled: true, PushNotificationsEnabled: true, EmailNotificationsEnabled: true}
	giftCardID := uuid.New()
	expiresAt := time.Now().Add(1 * 24 * time.Hour)

	expiringGC := models.GiftCard{
		ID:        giftCardID,
		UserID:    &userID,
		ExpiresAt: &expiresAt,
		Merchant:  merchant,
		User:      user,
	}

	// Day 7, 3: no expiring items
	for _, d := range []int{7, 3} {
		voucherRepo.On("GetExpiringVouchers", ctx, d).Return([]models.Voucher{}, nil)
		giftCardRepo.On("GetExpiringGiftCards", ctx, d).Return([]models.GiftCard{}, nil)
	}

	// Day 1: one expiring gift card
	voucherRepo.On("GetExpiringVouchers", ctx, 1).Return([]models.Voucher{}, nil)
	giftCardRepo.On("GetExpiringGiftCards", ctx, 1).Return([]models.GiftCard{expiringGC}, nil)

	reminderRepo.On("HasBeenSent", ctx, userID, "gift_card", giftCardID, 1).Return(false, nil)
	notifRepo.On("Create", ctx, mock.AnythingOfType("*models.Notification")).Return(nil)
	pushSvc.On("SendPushToUser", ctx, userID, mock.AnythingOfType("string"), mock.AnythingOfType("string"), "/gift_cards").Return(nil)
	emailSvc.On("SendExpiryReminder", ctx, "test@example.com", "Test", mock.AnythingOfType("email.ExpiryReminderData"), "", "de").Return(nil)
	reminderRepo.On("MarkSent", ctx, mock.AnythingOfType("*models.ExpiryReminderSent")).Return(nil)
	voucherRepo.On("GetVouchersStartingTomorrow", ctx).Return([]models.Voucher{}, nil)

	err := svc.CheckAndSendReminders(ctx)

	assert.NoError(t, err)
	notifRepo.AssertExpectations(t)
	reminderRepo.AssertExpectations(t)
}

func TestReminderService_CheckAndSendReminders_AlreadySent(t *testing.T) {
	svc, reminderRepo, voucherRepo, giftCardRepo, _, _, _ := setupReminderTest()
	ctx := context.Background()

	userID := uuid.New()
	voucherID := uuid.New()
	merchant := &models.Merchant{ID: uuid.New(), Name: "Test"}

	expiringVoucher := models.Voucher{
		ID:         voucherID,
		UserID:     &userID,
		ValidUntil: time.Now().Add(3 * 24 * time.Hour),
		Merchant:   merchant,
	}

	for _, d := range []int{7, 3, 1} {
		if d == 3 {
			voucherRepo.On("GetExpiringVouchers", ctx, d).Return([]models.Voucher{expiringVoucher}, nil)
		} else {
			voucherRepo.On("GetExpiringVouchers", ctx, d).Return([]models.Voucher{}, nil)
		}
		giftCardRepo.On("GetExpiringGiftCards", ctx, d).Return([]models.GiftCard{}, nil)
	}

	// Already sent → should skip
	reminderRepo.On("HasBeenSent", ctx, userID, "voucher", voucherID, 3).Return(true, nil)
	voucherRepo.On("GetVouchersStartingTomorrow", ctx).Return([]models.Voucher{}, nil)

	err := svc.CheckAndSendReminders(ctx)

	assert.NoError(t, err)
	// Notification should NOT have been created
	reminderRepo.AssertNotCalled(t, "MarkSent")
}

func TestReminderService_CheckAndSendReminders_NilUserID(t *testing.T) {
	svc, _, voucherRepo, giftCardRepo, _, _, _ := setupReminderTest()
	ctx := context.Background()

	// Voucher with nil UserID should be skipped
	expiringVoucher := models.Voucher{
		ID:         uuid.New(),
		UserID:     nil, // nil user
		ValidUntil: time.Now().Add(3 * 24 * time.Hour),
	}

	for _, d := range []int{7, 3, 1} {
		if d == 3 {
			voucherRepo.On("GetExpiringVouchers", ctx, d).Return([]models.Voucher{expiringVoucher}, nil)
		} else {
			voucherRepo.On("GetExpiringVouchers", ctx, d).Return([]models.Voucher{}, nil)
		}
		giftCardRepo.On("GetExpiringGiftCards", ctx, d).Return([]models.GiftCard{}, nil)
	}
	voucherRepo.On("GetVouchersStartingTomorrow", ctx).Return([]models.Voucher{}, nil)

	err := svc.CheckAndSendReminders(ctx)

	assert.NoError(t, err)
}

func TestReminderService_CheckAndSendReminders_NoExpiring(t *testing.T) {
	svc, _, voucherRepo, giftCardRepo, _, _, _ := setupReminderTest()
	ctx := context.Background()

	for _, d := range []int{7, 3, 1} {
		voucherRepo.On("GetExpiringVouchers", ctx, d).Return([]models.Voucher{}, nil)
		giftCardRepo.On("GetExpiringGiftCards", ctx, d).Return([]models.GiftCard{}, nil)
	}
	voucherRepo.On("GetVouchersStartingTomorrow", ctx).Return([]models.Voucher{}, nil)

	err := svc.CheckAndSendReminders(ctx)

	assert.NoError(t, err)
	voucherRepo.AssertExpectations(t)
	giftCardRepo.AssertExpectations(t)
}

func TestReminderService_NilPushService(t *testing.T) {
	reminderRepo := new(MockReminderRepo)
	voucherRepo := new(mockVoucherRepoForReminder)
	giftCardRepo := new(mockGiftCardRepoForReminder)
	notifRepo := new(mockNotifRepoForReminder)
	emailSvc := new(mockEmailSvcForReminder)

	// Create service with nil push service
	svc := NewReminderService(reminderRepo, voucherRepo, giftCardRepo, nil, nil, notifRepo, nil, emailSvc, nil, []int{3}, time.UTC, "https://savvy.example.com")
	ctx := context.Background()

	userID := uuid.New()
	merchant := &models.Merchant{ID: uuid.New(), Name: "Test"}
	user := &models.User{ID: userID, Email: "test@example.com", FirstName: "Test", Language: "en", PushRemindersEnabled: true, EmailRemindersEnabled: true, PushNotificationsEnabled: true, EmailNotificationsEnabled: true}
	voucherID := uuid.New()

	expiringVoucher := models.Voucher{
		ID:         voucherID,
		UserID:     &userID,
		ValidUntil: time.Now().Add(3 * 24 * time.Hour),
		Merchant:   merchant,
		User:       user,
	}

	voucherRepo.On("GetExpiringVouchers", ctx, 3).Return([]models.Voucher{expiringVoucher}, nil)
	giftCardRepo.On("GetExpiringGiftCards", ctx, 3).Return([]models.GiftCard{}, nil)
	reminderRepo.On("HasBeenSent", ctx, userID, "voucher", voucherID, 3).Return(false, nil)
	notifRepo.On("Create", ctx, mock.AnythingOfType("*models.Notification")).Return(nil)
	emailSvc.On("SendExpiryReminder", ctx, "test@example.com", "Test", mock.AnythingOfType("email.ExpiryReminderData"), "", "en").Return(nil)
	reminderRepo.On("MarkSent", ctx, mock.AnythingOfType("*models.ExpiryReminderSent")).Return(nil)
	voucherRepo.On("GetVouchersStartingTomorrow", ctx).Return([]models.Voucher{}, nil)

	err := svc.CheckAndSendReminders(ctx)

	assert.NoError(t, err)
	notifRepo.AssertExpectations(t)
	reminderRepo.AssertExpectations(t)
}

func TestReminderService_NilEmailService(t *testing.T) {
	reminderRepo := new(MockReminderRepo)
	voucherRepo := new(mockVoucherRepoForReminder)
	giftCardRepo := new(mockGiftCardRepoForReminder)
	notifRepo := new(mockNotifRepoForReminder)
	pushSvc := new(mockPushSvcForReminder)

	// Create service with nil email service
	svc := NewReminderService(reminderRepo, voucherRepo, giftCardRepo, nil, nil, notifRepo, pushSvc, nil, nil, []int{3}, time.UTC, "https://savvy.example.com")
	ctx := context.Background()

	userID := uuid.New()
	merchant := &models.Merchant{ID: uuid.New(), Name: "Test"}
	user := &models.User{ID: userID, Email: "test@example.com", FirstName: "Test", Language: "en", PushRemindersEnabled: true, EmailRemindersEnabled: true, PushNotificationsEnabled: true, EmailNotificationsEnabled: true}
	voucherID := uuid.New()

	expiringVoucher := models.Voucher{
		ID:         voucherID,
		UserID:     &userID,
		ValidUntil: time.Now().Add(3 * 24 * time.Hour),
		Merchant:   merchant,
		User:       user,
	}

	voucherRepo.On("GetExpiringVouchers", ctx, 3).Return([]models.Voucher{expiringVoucher}, nil)
	giftCardRepo.On("GetExpiringGiftCards", ctx, 3).Return([]models.GiftCard{}, nil)
	reminderRepo.On("HasBeenSent", ctx, userID, "voucher", voucherID, 3).Return(false, nil)
	notifRepo.On("Create", ctx, mock.AnythingOfType("*models.Notification")).Return(nil)
	pushSvc.On("SendPushToUser", ctx, userID, mock.AnythingOfType("string"), mock.AnythingOfType("string"), "/vouchers").Return(nil)
	reminderRepo.On("MarkSent", ctx, mock.AnythingOfType("*models.ExpiryReminderSent")).Return(nil)
	voucherRepo.On("GetVouchersStartingTomorrow", ctx).Return([]models.Voucher{}, nil)

	err := svc.CheckAndSendReminders(ctx)

	assert.NoError(t, err)
	notifRepo.AssertExpectations(t)
	reminderRepo.AssertExpectations(t)
}

func TestReminderService_CalculateDaysLeft(t *testing.T) {
	// Test that days left calculation uses UTC date extraction.
	// In production, dates are stored as end-of-day UTC (T23:59:59Z) from the frontend.
	// The function must extract the date from UTC to avoid timezone-shift issues.
	loc, err := time.LoadLocation("Europe/Zurich")
	assert.NoError(t, err)

	reminderRepo := new(MockReminderRepo)
	voucherRepo := new(mockVoucherRepoForReminder)
	giftCardRepo := new(mockGiftCardRepoForReminder)
	notifRepo := new(mockNotifRepoForReminder)
	pushSvc := new(mockPushSvcForReminder)
	emailSvc := new(mockEmailSvcForReminder)

	svc := NewReminderService(reminderRepo, voucherRepo, giftCardRepo, nil, nil, notifRepo, pushSvc, emailSvc, nil, []int{3}, loc, "https://savvy.example.com").(*ReminderService)

	// Create expiry dates in UTC to match production behavior.
	// The frontend stores dates as end-of-day UTC (T23:59:59Z),
	// so the UTC date is the canonical calendar date.
	now := time.Now().In(loc)

	// Helper: create an end-of-day UTC time for N days from now (local timezone reference)
	makeExpiryUTC := func(daysFromNow int) time.Time {
		target := now.AddDate(0, 0, daysFromNow)
		return time.Date(target.Year(), target.Month(), target.Day(), 23, 59, 59, 0, time.UTC)
	}

	tests := []struct {
		name       string
		expiryTime time.Time
		wantDays   int
	}{
		{
			name:       "expires tomorrow (end of day UTC)",
			expiryTime: makeExpiryUTC(1),
			wantDays:   1,
		},
		{
			name: "expires tomorrow (noon UTC)",
			expiryTime: func() time.Time {
				target := now.AddDate(0, 0, 1)
				return time.Date(target.Year(), target.Month(), target.Day(), 12, 0, 0, 0, time.UTC)
			}(),
			wantDays: 1,
		},
		{
			name:       "expires in 3 days (end of day UTC)",
			expiryTime: makeExpiryUTC(3),
			wantDays:   3,
		},
		{
			name: "expires in 3 days (midnight UTC)",
			expiryTime: func() time.Time {
				target := now.AddDate(0, 0, 3)
				return time.Date(target.Year(), target.Month(), target.Day(), 0, 0, 0, 0, time.UTC)
			}(),
			wantDays: 3,
		},
		{
			name:       "expires today (end of day UTC)",
			expiryTime: makeExpiryUTC(0),
			wantDays:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.calculateDaysLeft(tt.expiryTime)
			assert.Equal(t, tt.wantDays, got, "expected %d days, got %d", tt.wantDays, got)
		})
	}
}

func TestReminderService_NoDuplicateReminders(t *testing.T) {
	// Test that a voucher expiring in 3 days only triggers ONE reminder
	// even though it matches both daysBefore=7 and daysBefore=3
	reminderRepo := new(MockReminderRepo)
	voucherRepo := new(mockVoucherRepoForReminder)
	giftCardRepo := new(mockGiftCardRepoForReminder)
	notifRepo := new(mockNotifRepoForReminder)
	pushSvc := new(mockPushSvcForReminder)
	emailSvc := new(mockEmailSvcForReminder)

	svc := NewReminderService(reminderRepo, voucherRepo, giftCardRepo, nil, nil, notifRepo, pushSvc, emailSvc, nil, []int{7, 3, 1}, time.UTC, "https://savvy.example.com")
	ctx := context.Background()

	userID := uuid.New()
	merchant := &models.Merchant{ID: uuid.New(), Name: "IKEA"}
	user := &models.User{ID: userID, Email: "test@example.com", FirstName: "Test", Language: "en", PushRemindersEnabled: true, EmailRemindersEnabled: true, PushNotificationsEnabled: true, EmailNotificationsEnabled: true}
	voucherID := uuid.New()

	// Voucher expires in EXACTLY 3 days
	expiringVoucher := models.Voucher{
		ID:         voucherID,
		UserID:     &userID,
		ValidUntil: time.Now().Add(3 * 24 * time.Hour),
		Merchant:   merchant,
		User:       user,
	}

	// GetExpiringVouchers will be called 3 times (for days 7, 3, 1)
	// It returns the voucher for days=7 and days=3 (since 3 <= 7 and 3 <= 3)
	voucherRepo.On("GetExpiringVouchers", mock.Anything, 7).Return([]models.Voucher{expiringVoucher}, nil)
	voucherRepo.On("GetExpiringVouchers", mock.Anything, 3).Return([]models.Voucher{expiringVoucher}, nil)
	voucherRepo.On("GetExpiringVouchers", mock.Anything, 1).Return([]models.Voucher{}, nil)

	giftCardRepo.On("GetExpiringGiftCards", mock.Anything, mock.Anything).Return([]models.GiftCard{}, nil)

	// HasBeenSent should only be called ONCE (for daysBefore=3)
	// because daysBefore=7 is skipped (daysLeft=3 != daysBefore=7)
	reminderRepo.On("HasBeenSent", mock.Anything, userID, "voucher", voucherID, 3).Return(false, nil)
	reminderRepo.On("MarkSent", mock.Anything, mock.MatchedBy(func(r *models.ExpiryReminderSent) bool {
		return r.UserID == userID && r.ResourceID == voucherID && r.DaysBefore == 3
	})).Return(nil)

	// Notification should be created ONCE
	notifRepo.On("Create", mock.Anything, mock.MatchedBy(func(n *models.Notification) bool {
		return n.UserID == userID && n.ResourceType == "voucher" && n.ResourceID == voucherID &&
			n.Metadata["days_left"] == 3 // Should show 3 days left
	})).Return(nil).Once() // IMPORTANT: .Once() ensures it's called exactly once

	pushSvc.On("SendPushToUser", mock.Anything, userID, mock.AnythingOfType("string"), mock.AnythingOfType("string"), "/vouchers").Return(nil)
	emailSvc.On("SendExpiryReminder", mock.Anything, "test@example.com", "Test", mock.AnythingOfType("email.ExpiryReminderData"), "", "en").Return(nil)
	voucherRepo.On("GetVouchersStartingTomorrow", mock.Anything).Return([]models.Voucher{}, nil)

	err := svc.CheckAndSendReminders(ctx)

	assert.NoError(t, err)
	notifRepo.AssertExpectations(t) // This will fail if Create was called more than once
	reminderRepo.AssertExpectations(t)
}

func TestReminderService_ValidityStart_HappyPath(t *testing.T) {
	svc, reminderRepo, voucherRepo, giftCardRepo, notifRepo, pushSvc, emailSvc := setupReminderTest()
	ctx := context.Background()

	userID := uuid.New()
	merchant := &models.Merchant{ID: uuid.New(), Name: "IKEA"}
	user := &models.User{ID: userID, Email: "test@example.com", FirstName: "Test", Language: "en", PushRemindersEnabled: true, EmailRemindersEnabled: true, PushNotificationsEnabled: true, EmailNotificationsEnabled: true}
	voucherID := uuid.New()

	tomorrow := time.Now().AddDate(0, 0, 1)
	startingVoucher := models.Voucher{
		ID:        voucherID,
		UserID:    &userID,
		ValidFrom: tomorrow,
		Merchant:  merchant,
		User:      user,
		Code:      "TESTCODE",
		Type:      "percentage",
		Value:     20,
	}

	// Expiry reminder loop: no expiring items
	for _, d := range []int{7, 3, 1} {
		voucherRepo.On("GetExpiringVouchers", ctx, d).Return([]models.Voucher{}, nil)
		giftCardRepo.On("GetExpiringGiftCards", ctx, d).Return([]models.GiftCard{}, nil)
	}

	// Validity start: one voucher starting tomorrow
	voucherRepo.On("GetVouchersStartingTomorrow", ctx).Return([]models.Voucher{startingVoucher}, nil)
	reminderRepo.On("HasBeenSent", ctx, userID, "voucher_start", voucherID, 1).Return(false, nil)
	notifRepo.On("Create", ctx, mock.MatchedBy(func(n *models.Notification) bool {
		return n.UserID == userID &&
			n.Type == models.NotificationTypeValidityStart &&
			n.ResourceType == "voucher" &&
			n.ResourceID == voucherID &&
			n.Metadata["merchant_name"] == "IKEA"
	})).Return(nil)
	pushSvc.On("SendPushToUser", ctx, userID, mock.AnythingOfType("string"), mock.AnythingOfType("string"), "/vouchers").Return(nil)
	emailSvc.On("SendValidityStart", ctx, "test@example.com", "Test", mock.AnythingOfType("email.ValidityStartData"), "", "en").Return(nil)
	reminderRepo.On("MarkSent", ctx, mock.MatchedBy(func(r *models.ExpiryReminderSent) bool {
		return r.UserID == userID && r.ResourceType == "voucher_start" && r.ResourceID == voucherID && r.DaysBefore == 1
	})).Return(nil)

	err := svc.CheckAndSendReminders(ctx)

	assert.NoError(t, err)
	notifRepo.AssertExpectations(t)
	reminderRepo.AssertExpectations(t)
	pushSvc.AssertExpectations(t)
	emailSvc.AssertExpectations(t)
}

func TestReminderService_ValidityStart_AlreadySent(t *testing.T) {
	svc, reminderRepo, voucherRepo, giftCardRepo, _, _, _ := setupReminderTest()
	ctx := context.Background()

	userID := uuid.New()
	voucherID := uuid.New()
	merchant := &models.Merchant{ID: uuid.New(), Name: "Test"}
	user := &models.User{ID: userID, Email: "test@example.com", FirstName: "Test", Language: "en", PushRemindersEnabled: true, EmailRemindersEnabled: true, PushNotificationsEnabled: true, EmailNotificationsEnabled: true}

	tomorrow := time.Now().AddDate(0, 0, 1)
	startingVoucher := models.Voucher{
		ID:        voucherID,
		UserID:    &userID,
		ValidFrom: tomorrow,
		Merchant:  merchant,
		User:      user,
	}

	for _, d := range []int{7, 3, 1} {
		voucherRepo.On("GetExpiringVouchers", ctx, d).Return([]models.Voucher{}, nil)
		giftCardRepo.On("GetExpiringGiftCards", ctx, d).Return([]models.GiftCard{}, nil)
	}

	voucherRepo.On("GetVouchersStartingTomorrow", ctx).Return([]models.Voucher{startingVoucher}, nil)
	reminderRepo.On("HasBeenSent", ctx, userID, "voucher_start", voucherID, 1).Return(true, nil)

	err := svc.CheckAndSendReminders(ctx)

	assert.NoError(t, err)
	// MarkSent should NOT be called for validity start (already sent)
	reminderRepo.AssertNotCalled(t, "MarkSent")
}

func TestReminderService_ValidityStart_OptedOut(t *testing.T) {
	svc, reminderRepo, voucherRepo, giftCardRepo, notifRepo, pushSvc, emailSvc := setupReminderTest()
	ctx := context.Background()

	userID := uuid.New()
	merchant := &models.Merchant{ID: uuid.New(), Name: "Test"}
	user := &models.User{ID: userID, Email: "test@example.com", FirstName: "Test", Language: "en", PushRemindersEnabled: false, EmailRemindersEnabled: false, PushNotificationsEnabled: true, EmailNotificationsEnabled: true}
	voucherID := uuid.New()

	tomorrow := time.Now().AddDate(0, 0, 1)
	startingVoucher := models.Voucher{
		ID:        voucherID,
		UserID:    &userID,
		ValidFrom: tomorrow,
		Merchant:  merchant,
		User:      user,
	}

	for _, d := range []int{7, 3, 1} {
		voucherRepo.On("GetExpiringVouchers", ctx, d).Return([]models.Voucher{}, nil)
		giftCardRepo.On("GetExpiringGiftCards", ctx, d).Return([]models.GiftCard{}, nil)
	}

	voucherRepo.On("GetVouchersStartingTomorrow", ctx).Return([]models.Voucher{startingVoucher}, nil)
	reminderRepo.On("HasBeenSent", ctx, userID, "voucher_start", voucherID, 1).Return(false, nil)
	// In-app notification is always created
	notifRepo.On("Create", ctx, mock.AnythingOfType("*models.Notification")).Return(nil)
	reminderRepo.On("MarkSent", ctx, mock.AnythingOfType("*models.ExpiryReminderSent")).Return(nil)

	err := svc.CheckAndSendReminders(ctx)

	assert.NoError(t, err)
	notifRepo.AssertExpectations(t)
	reminderRepo.AssertExpectations(t)
	// Push and email should NOT be called (PushRemindersEnabled=false, EmailRemindersEnabled=false)
	pushSvc.AssertNotCalled(t, "SendPushToUser")
	emailSvc.AssertNotCalled(t, "SendValidityStart")
}

// ==================== Channel Preference Tests ====================

func TestReminderService_PushDisabled(t *testing.T) {
	svc, reminderRepo, voucherRepo, giftCardRepo, notifRepo, pushSvc, emailSvc := setupReminderTest()
	ctx := context.Background()

	userID := uuid.New()
	merchant := &models.Merchant{ID: uuid.New(), Name: "IKEA"}
	user := &models.User{ID: userID, Email: "test@example.com", FirstName: "Test", Language: "en", PushRemindersEnabled: true, EmailRemindersEnabled: true, PushNotificationsEnabled: false, EmailNotificationsEnabled: true}
	voucherID := uuid.New()

	expiringVoucher := models.Voucher{
		ID:         voucherID,
		UserID:     &userID,
		ValidUntil: time.Now().Add(3 * 24 * time.Hour),
		Merchant:   merchant,
		User:       user,
	}

	voucherRepo.On("GetExpiringVouchers", ctx, 7).Return([]models.Voucher{}, nil)
	voucherRepo.On("GetExpiringVouchers", ctx, 3).Return([]models.Voucher{expiringVoucher}, nil)
	voucherRepo.On("GetExpiringVouchers", ctx, 1).Return([]models.Voucher{}, nil)
	for _, d := range []int{7, 3, 1} {
		giftCardRepo.On("GetExpiringGiftCards", ctx, d).Return([]models.GiftCard{}, nil)
	}
	voucherRepo.On("GetVouchersStartingTomorrow", ctx).Return([]models.Voucher{}, nil)

	reminderRepo.On("HasBeenSent", ctx, userID, "voucher", voucherID, 3).Return(false, nil)
	notifRepo.On("Create", ctx, mock.AnythingOfType("*models.Notification")).Return(nil)
	emailSvc.On("SendExpiryReminder", ctx, "test@example.com", "Test", mock.AnythingOfType("email.ExpiryReminderData"), "", "en").Return(nil)
	reminderRepo.On("MarkSent", ctx, mock.AnythingOfType("*models.ExpiryReminderSent")).Return(nil)

	err := svc.CheckAndSendReminders(ctx)

	assert.NoError(t, err)
	// In-app and email should be sent
	notifRepo.AssertExpectations(t)
	emailSvc.AssertExpectations(t)
	reminderRepo.AssertExpectations(t)
	// Push should NOT be called
	pushSvc.AssertNotCalled(t, "SendPushToUser")
}

func TestReminderService_EmailDisabled(t *testing.T) {
	svc, reminderRepo, voucherRepo, giftCardRepo, notifRepo, pushSvc, emailSvc := setupReminderTest()
	ctx := context.Background()

	userID := uuid.New()
	merchant := &models.Merchant{ID: uuid.New(), Name: "IKEA"}
	user := &models.User{ID: userID, Email: "test@example.com", FirstName: "Test", Language: "en", PushRemindersEnabled: true, EmailRemindersEnabled: true, PushNotificationsEnabled: true, EmailNotificationsEnabled: false}
	voucherID := uuid.New()

	expiringVoucher := models.Voucher{
		ID:         voucherID,
		UserID:     &userID,
		ValidUntil: time.Now().Add(3 * 24 * time.Hour),
		Merchant:   merchant,
		User:       user,
	}

	voucherRepo.On("GetExpiringVouchers", ctx, 7).Return([]models.Voucher{}, nil)
	voucherRepo.On("GetExpiringVouchers", ctx, 3).Return([]models.Voucher{expiringVoucher}, nil)
	voucherRepo.On("GetExpiringVouchers", ctx, 1).Return([]models.Voucher{}, nil)
	for _, d := range []int{7, 3, 1} {
		giftCardRepo.On("GetExpiringGiftCards", ctx, d).Return([]models.GiftCard{}, nil)
	}
	voucherRepo.On("GetVouchersStartingTomorrow", ctx).Return([]models.Voucher{}, nil)

	reminderRepo.On("HasBeenSent", ctx, userID, "voucher", voucherID, 3).Return(false, nil)
	notifRepo.On("Create", ctx, mock.AnythingOfType("*models.Notification")).Return(nil)
	pushSvc.On("SendPushToUser", ctx, userID, mock.AnythingOfType("string"), mock.AnythingOfType("string"), "/vouchers").Return(nil)
	reminderRepo.On("MarkSent", ctx, mock.AnythingOfType("*models.ExpiryReminderSent")).Return(nil)

	err := svc.CheckAndSendReminders(ctx)

	assert.NoError(t, err)
	// In-app and push should be sent
	notifRepo.AssertExpectations(t)
	pushSvc.AssertExpectations(t)
	reminderRepo.AssertExpectations(t)
	// Email should NOT be called
	emailSvc.AssertNotCalled(t, "SendExpiryReminder")
}

func TestReminderService_BothChannelsDisabled(t *testing.T) {
	svc, reminderRepo, voucherRepo, giftCardRepo, notifRepo, pushSvc, emailSvc := setupReminderTest()
	ctx := context.Background()

	userID := uuid.New()
	merchant := &models.Merchant{ID: uuid.New(), Name: "IKEA"}
	user := &models.User{ID: userID, Email: "test@example.com", FirstName: "Test", Language: "en", PushRemindersEnabled: true, EmailRemindersEnabled: true, PushNotificationsEnabled: false, EmailNotificationsEnabled: false}
	voucherID := uuid.New()

	expiringVoucher := models.Voucher{
		ID:         voucherID,
		UserID:     &userID,
		ValidUntil: time.Now().Add(3 * 24 * time.Hour),
		Merchant:   merchant,
		User:       user,
	}

	voucherRepo.On("GetExpiringVouchers", ctx, 7).Return([]models.Voucher{}, nil)
	voucherRepo.On("GetExpiringVouchers", ctx, 3).Return([]models.Voucher{expiringVoucher}, nil)
	voucherRepo.On("GetExpiringVouchers", ctx, 1).Return([]models.Voucher{}, nil)
	for _, d := range []int{7, 3, 1} {
		giftCardRepo.On("GetExpiringGiftCards", ctx, d).Return([]models.GiftCard{}, nil)
	}
	voucherRepo.On("GetVouchersStartingTomorrow", ctx).Return([]models.Voucher{}, nil)

	reminderRepo.On("HasBeenSent", ctx, userID, "voucher", voucherID, 3).Return(false, nil)
	// In-app notification is ALWAYS created
	notifRepo.On("Create", ctx, mock.AnythingOfType("*models.Notification")).Return(nil)
	reminderRepo.On("MarkSent", ctx, mock.AnythingOfType("*models.ExpiryReminderSent")).Return(nil)

	err := svc.CheckAndSendReminders(ctx)

	assert.NoError(t, err)
	// Only in-app should be created
	notifRepo.AssertExpectations(t)
	reminderRepo.AssertExpectations(t)
	// Neither push nor email should be called
	pushSvc.AssertNotCalled(t, "SendPushToUser")
	emailSvc.AssertNotCalled(t, "SendExpiryReminder")
}

// ==================== Mock Email Token Service (for Reminder) ====================

type mockEmailTokenSvcForReminder struct {
	mock.Mock
}

var _ EmailTokenServiceInterface = (*mockEmailTokenSvcForReminder)(nil)

func (m *mockEmailTokenSvcForReminder) CreateVerificationToken(ctx context.Context, userID uuid.UUID) (string, error) {
	args := m.Called(ctx, userID)
	return args.String(0), args.Error(1)
}
func (m *mockEmailTokenSvcForReminder) VerifyEmail(_ context.Context, _ string) error { return nil }
func (m *mockEmailTokenSvcForReminder) CreatePasswordResetToken(_ context.Context, _ uuid.UUID) (string, error) {
	return "", nil
}
func (m *mockEmailTokenSvcForReminder) ConsumePasswordResetToken(_ context.Context, _ string) (*models.User, error) {
	return nil, nil
}
func (m *mockEmailTokenSvcForReminder) CreateUnsubscribeToken(_ context.Context, _ uuid.UUID) (string, error) {
	return "", nil
}
func (m *mockEmailTokenSvcForReminder) UnsubscribeNotifications(_ context.Context, _ string) error {
	return nil
}
func (m *mockEmailTokenSvcForReminder) CreateUnsubscribeReminderToken(ctx context.Context, userID uuid.UUID) (string, error) {
	args := m.Called(ctx, userID)
	return args.String(0), args.Error(1)
}
func (m *mockEmailTokenSvcForReminder) UnsubscribeReminders(_ context.Context, _ string) error {
	return nil
}
func (m *mockEmailTokenSvcForReminder) CleanupExpiredTokens(_ context.Context) error {
	return nil
}

// ==================== Mock Voucher Share Repository (for Reminder) ====================

type mockVoucherShareRepoForReminder struct {
	mock.Mock
}

var _ repository.VoucherShareRepository = (*mockVoucherShareRepoForReminder)(nil)

func (m *mockVoucherShareRepoForReminder) GetByVoucherID(ctx context.Context, voucherID uuid.UUID) ([]models.VoucherShare, error) {
	args := m.Called(ctx, voucherID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.VoucherShare), args.Error(1)
}
func (m *mockVoucherShareRepoForReminder) Create(_ context.Context, _ *models.VoucherShare) error {
	return nil
}
func (m *mockVoucherShareRepoForReminder) GetByVoucherAndUser(_ context.Context, _, _ uuid.UUID) (*models.VoucherShare, error) {
	return nil, nil
}
func (m *mockVoucherShareRepoForReminder) Update(_ context.Context, _ *models.VoucherShare) error {
	return nil
}
func (m *mockVoucherShareRepoForReminder) DeleteByVoucherAndUser(_ context.Context, _, _ uuid.UUID) error {
	return nil
}
func (m *mockVoucherShareRepoForReminder) DeleteByVoucherID(_ context.Context, _ uuid.UUID) error {
	return nil
}
func (m *mockVoucherShareRepoForReminder) CountByUser(_ context.Context, _ uuid.UUID) (int64, error) {
	return 0, nil
}
func (m *mockVoucherShareRepoForReminder) CountByVoucherIDs(_ context.Context, _ []uuid.UUID) (map[uuid.UUID]int64, error) {
	return nil, nil
}
func (m *mockVoucherShareRepoForReminder) GetSharedUserIDs(_ context.Context, _ uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}

// ==================== Mock Gift Card Share Repository (for Reminder) ====================

type mockGiftCardShareRepoForReminder struct {
	mock.Mock
}

var _ repository.GiftCardShareRepository = (*mockGiftCardShareRepoForReminder)(nil)

func (m *mockGiftCardShareRepoForReminder) GetByGiftCardID(ctx context.Context, giftCardID uuid.UUID) ([]models.GiftCardShare, error) {
	args := m.Called(ctx, giftCardID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.GiftCardShare), args.Error(1)
}
func (m *mockGiftCardShareRepoForReminder) Create(_ context.Context, _ *models.GiftCardShare) error {
	return nil
}
func (m *mockGiftCardShareRepoForReminder) GetByGiftCardAndUser(_ context.Context, _, _ uuid.UUID) (*models.GiftCardShare, error) {
	return nil, nil
}
func (m *mockGiftCardShareRepoForReminder) Update(_ context.Context, _ *models.GiftCardShare) error {
	return nil
}
func (m *mockGiftCardShareRepoForReminder) DeleteByGiftCardAndUser(_ context.Context, _, _ uuid.UUID) error {
	return nil
}
func (m *mockGiftCardShareRepoForReminder) DeleteByGiftCardID(_ context.Context, _ uuid.UUID) error {
	return nil
}
func (m *mockGiftCardShareRepoForReminder) CountByUser(_ context.Context, _ uuid.UUID) (int64, error) {
	return 0, nil
}
func (m *mockGiftCardShareRepoForReminder) CountByGiftCardIDs(_ context.Context, _ []uuid.UUID) (map[uuid.UUID]int64, error) {
	return nil, nil
}
func (m *mockGiftCardShareRepoForReminder) GetSharedUserIDs(_ context.Context, _ uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}

// ==================== Helper: create ReminderService with direct struct access ====================

func newTestReminderService(
	reminderRepo repository.ReminderRepository,
	voucherRepo repository.VoucherRepository,
	giftCardRepo repository.GiftCardRepository,
	voucherShareRepo repository.VoucherShareRepository,
	giftCardShareRepo repository.GiftCardShareRepository,
	notifRepo repository.NotificationRepository,
	pushSvc PushServiceInterface,
	emailSvc email.ServiceInterface,
	emailTokenSvc EmailTokenServiceInterface,
	daysBefore []int,
	location *time.Location,
	frontendURL string,
) *ReminderService {
	return NewReminderService(
		reminderRepo, voucherRepo, giftCardRepo,
		voucherShareRepo, giftCardShareRepo,
		notifRepo, pushSvc, emailSvc, emailTokenSvc,
		daysBefore, location, frontendURL,
	).(*ReminderService)
}

// ==================== frenchMonth Tests ====================

func TestReminderService_FrenchMonth(t *testing.T) {
	expected := map[time.Month]string{
		time.January:   "janvier",
		time.February:  "février",
		time.March:     "mars",
		time.April:     "avril",
		time.May:       "mai",
		time.June:      "juin",
		time.July:      "juillet",
		time.August:    "août",
		time.September: "septembre",
		time.October:   "octobre",
		time.November:  "novembre",
		time.December:  "décembre",
	}

	for month, want := range expected {
		t.Run(month.String(), func(t *testing.T) {
			got := frenchMonth(month)
			assert.Equal(t, want, got)
		})
	}
}

// ==================== germanMonth Tests ====================

func TestReminderService_GermanMonth(t *testing.T) {
	expected := map[time.Month]string{
		time.January:   "Januar",
		time.February:  "Februar",
		time.March:     "März",
		time.April:     "April",
		time.May:       "Mai",
		time.June:      "Juni",
		time.July:      "Juli",
		time.August:    "August",
		time.September: "September",
		time.October:   "Oktober",
		time.November:  "November",
		time.December:  "Dezember",
	}

	for month, want := range expected {
		t.Run(month.String(), func(t *testing.T) {
			got := germanMonth(month)
			assert.Equal(t, want, got)
		})
	}
}

// ==================== formatCurrency Tests ====================

func TestReminderService_FormatCurrency(t *testing.T) {
	svc := newTestReminderService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, time.UTC, "")

	t.Run("empty currency defaults to CHF", func(t *testing.T) {
		result := svc.formatCurrency(50.00, "")
		assert.Equal(t, "CHF 50.00", result)
	})

	t.Run("with EUR currency", func(t *testing.T) {
		result := svc.formatCurrency(99.99, "EUR")
		assert.Equal(t, "EUR 99.99", result)
	})

	t.Run("with USD currency", func(t *testing.T) {
		result := svc.formatCurrency(25.50, "USD")
		assert.Equal(t, "USD 25.50", result)
	})

	t.Run("zero amount", func(t *testing.T) {
		result := svc.formatCurrency(0, "CHF")
		assert.Equal(t, "CHF 0.00", result)
	})

	t.Run("large amount", func(t *testing.T) {
		result := svc.formatCurrency(1234.56, "GBP")
		assert.Equal(t, "GBP 1234.56", result)
	})
}

// ==================== formatVoucherValue Tests ====================

func TestReminderService_FormatVoucherValue(t *testing.T) {
	svc := newTestReminderService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, time.UTC, "")

	t.Run("percentage type", func(t *testing.T) {
		v := &models.Voucher{Type: "percentage", Value: 20}
		result := svc.formatVoucherValue(v)
		assert.Equal(t, "20%", result)
	})

	t.Run("percentage type fractional", func(t *testing.T) {
		v := &models.Voucher{Type: "percentage", Value: 15.5}
		result := svc.formatVoucherValue(v)
		assert.Equal(t, "16%", result) // %.0f rounds 15.5 to 16
	})

	t.Run("fixed_amount type with currency", func(t *testing.T) {
		v := &models.Voucher{Type: "fixed_amount", Value: 50.00, Currency: "EUR"}
		result := svc.formatVoucherValue(v)
		assert.Equal(t, "EUR 50.00", result)
	})

	t.Run("fixed_amount type empty currency defaults to CHF", func(t *testing.T) {
		v := &models.Voucher{Type: "fixed_amount", Value: 100.00, Currency: ""}
		result := svc.formatVoucherValue(v)
		assert.Equal(t, "CHF 100.00", result)
	})

	t.Run("bonus_points type", func(t *testing.T) {
		v := &models.Voucher{Type: "bonus_points", Value: 222}
		result := svc.formatVoucherValue(v)
		assert.Equal(t, "+222", result)
	})

	t.Run("bonus_points type fractional", func(t *testing.T) {
		v := &models.Voucher{Type: "bonus_points", Value: 99.5}
		result := svc.formatVoucherValue(v)
		assert.Equal(t, "+100", result) // %.0f rounds 99.5 to 100
	})

	t.Run("free type returns localized label", func(t *testing.T) {
		// Without i18n.Init the bundle falls back to the message ID; the point
		// of the test is that free resolves via i18n (locale-driven) and is not
		// the hardcoded German string nor the empty default.
		v := &models.Voucher{Type: "free", Value: 0}
		result := svc.formatVoucherValue(v)
		assert.NotEmpty(t, result)
		assert.Equal(t, "voucher.type.free", result)
	})

	t.Run("default/unknown type returns empty", func(t *testing.T) {
		v := &models.Voucher{Type: "points_multiplier", Value: 2}
		result := svc.formatVoucherValue(v)
		assert.Equal(t, "", result)
	})

	t.Run("empty type returns empty", func(t *testing.T) {
		v := &models.Voucher{Type: "", Value: 10}
		result := svc.formatVoucherValue(v)
		assert.Equal(t, "", result)
	})
}

// ==================== buildResourceURL Tests ====================

func TestReminderService_BuildResourceURL(t *testing.T) {
	t.Run("with valid frontend URL", func(t *testing.T) {
		svc := newTestReminderService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, time.UTC, "https://savvy.example.com")
		resourceID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
		result := svc.buildResourceURL("vouchers", resourceID)
		assert.Equal(t, "https://savvy.example.com/vouchers/550e8400-e29b-41d4-a716-446655440000", result)
	})

	t.Run("with empty frontend URL returns empty", func(t *testing.T) {
		svc := newTestReminderService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, time.UTC, "")
		resourceID := uuid.New()
		result := svc.buildResourceURL("vouchers", resourceID)
		assert.Equal(t, "", result)
	})

	t.Run("gift-cards path", func(t *testing.T) {
		svc := newTestReminderService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, time.UTC, "https://savvy.example.com")
		resourceID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
		result := svc.buildResourceURL("gift-cards", resourceID)
		assert.Equal(t, "https://savvy.example.com/gift-cards/550e8400-e29b-41d4-a716-446655440000", result)
	})

	t.Run("trailing slash in frontend URL is stripped", func(t *testing.T) {
		svc := newTestReminderService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, time.UTC, "https://savvy.example.com/")
		resourceID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
		result := svc.buildResourceURL("vouchers", resourceID)
		assert.Equal(t, "https://savvy.example.com/vouchers/550e8400-e29b-41d4-a716-446655440000", result)
	})
}

// ==================== formatDate Tests ====================

func TestReminderService_FormatDate(t *testing.T) {
	svc := newTestReminderService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, time.UTC, "")

	// Use a fixed date: March 15, 2026 at 23:59:59 UTC (typical end-of-day UTC storage)
	testDate := time.Date(2026, time.March, 15, 23, 59, 59, 0, time.UTC)

	t.Run("German language", func(t *testing.T) {
		user := &models.User{Language: "de"}
		result := svc.formatDate(testDate, user)
		assert.Equal(t, "15. März 2026", result)
	})

	t.Run("French language", func(t *testing.T) {
		user := &models.User{Language: "fr"}
		result := svc.formatDate(testDate, user)
		assert.Equal(t, "15 mars 2026", result)
	})

	t.Run("English language (default)", func(t *testing.T) {
		user := &models.User{Language: "en"}
		result := svc.formatDate(testDate, user)
		assert.Equal(t, "March 15, 2026", result)
	})

	t.Run("empty language defaults to English format", func(t *testing.T) {
		user := &models.User{Language: ""}
		result := svc.formatDate(testDate, user)
		assert.Equal(t, "March 15, 2026", result)
	})

	t.Run("nil user defaults to English format", func(t *testing.T) {
		result := svc.formatDate(testDate, nil)
		assert.Equal(t, "March 15, 2026", result)
	})

	t.Run("unknown language defaults to English format", func(t *testing.T) {
		user := &models.User{Language: "it"}
		result := svc.formatDate(testDate, user)
		assert.Equal(t, "March 15, 2026", result)
	})

	t.Run("January date in German", func(t *testing.T) {
		janDate := time.Date(2026, time.January, 1, 23, 59, 59, 0, time.UTC)
		user := &models.User{Language: "de"}
		result := svc.formatDate(janDate, user)
		assert.Equal(t, "1. Januar 2026", result)
	})

	t.Run("December date in French", func(t *testing.T) {
		decDate := time.Date(2026, time.December, 31, 23, 59, 59, 0, time.UTC)
		user := &models.User{Language: "fr"}
		result := svc.formatDate(decDate, user)
		assert.Equal(t, "31 décembre 2026", result)
	})
}

// ==================== generateUnsubscribeURL Tests ====================

func TestReminderService_GenerateUnsubscribeURL(t *testing.T) {
	t.Run("nil emailTokenService returns empty string", func(t *testing.T) {
		svc := newTestReminderService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, time.UTC, "https://savvy.example.com")
		ctx := context.Background()
		userID := uuid.New()
		result := svc.generateUnsubscribeURL(ctx, userID)
		assert.Equal(t, "", result)
	})

	t.Run("successful token generation", func(t *testing.T) {
		tokenSvc := new(mockEmailTokenSvcForReminder)
		svc := newTestReminderService(nil, nil, nil, nil, nil, nil, nil, nil, tokenSvc, nil, time.UTC, "https://savvy.example.com")
		ctx := context.Background()
		userID := uuid.New()

		tokenSvc.On("CreateUnsubscribeReminderToken", ctx, userID).Return("test-token-123", nil)

		result := svc.generateUnsubscribeURL(ctx, userID)
		assert.Equal(t, "https://savvy.example.com/unsubscribe?token=test-token-123&type=reminders", result)
		tokenSvc.AssertExpectations(t)
	})

	t.Run("token creation error returns empty string", func(t *testing.T) {
		tokenSvc := new(mockEmailTokenSvcForReminder)
		svc := newTestReminderService(nil, nil, nil, nil, nil, nil, nil, nil, tokenSvc, nil, time.UTC, "https://savvy.example.com")
		ctx := context.Background()
		userID := uuid.New()

		tokenSvc.On("CreateUnsubscribeReminderToken", ctx, userID).Return("", fmt.Errorf("database error"))

		result := svc.generateUnsubscribeURL(ctx, userID)
		assert.Equal(t, "", result)
		tokenSvc.AssertExpectations(t)
	})
}

// ==================== checkVouchers Tests ====================

func TestReminderService_CheckVouchers_RepoError(t *testing.T) {
	reminderRepo := new(MockReminderRepo)
	voucherRepo := new(mockVoucherRepoForReminder)
	giftCardRepo := new(mockGiftCardRepoForReminder)
	notifRepo := new(mockNotifRepoForReminder)

	svc := newTestReminderService(reminderRepo, voucherRepo, giftCardRepo, nil, nil, notifRepo, nil, nil, nil, []int{3}, time.UTC, "")
	ctx := context.Background()

	// Simulate repo error for vouchers
	voucherRepo.On("GetExpiringVouchers", ctx, 3).Return(nil, fmt.Errorf("database connection failed"))
	giftCardRepo.On("GetExpiringGiftCards", ctx, 3).Return([]models.GiftCard{}, nil)
	voucherRepo.On("GetVouchersStartingTomorrow", ctx).Return([]models.Voucher{}, nil)

	// CheckAndSendReminders does not propagate errors (logs them), so it returns nil
	err := svc.CheckAndSendReminders(ctx)
	assert.NoError(t, err)
	voucherRepo.AssertExpectations(t)
}

func TestReminderService_CheckVouchers_NilUserID(t *testing.T) {
	reminderRepo := new(MockReminderRepo)
	voucherRepo := new(mockVoucherRepoForReminder)
	giftCardRepo := new(mockGiftCardRepoForReminder)
	notifRepo := new(mockNotifRepoForReminder)

	svc := newTestReminderService(reminderRepo, voucherRepo, giftCardRepo, nil, nil, notifRepo, nil, nil, nil, []int{3}, time.UTC, "")
	ctx := context.Background()

	voucherWithNilUser := models.Voucher{
		ID:         uuid.New(),
		UserID:     nil,
		ValidUntil: time.Now().Add(3 * 24 * time.Hour),
	}

	voucherRepo.On("GetExpiringVouchers", ctx, 3).Return([]models.Voucher{voucherWithNilUser}, nil)
	giftCardRepo.On("GetExpiringGiftCards", ctx, 3).Return([]models.GiftCard{}, nil)
	voucherRepo.On("GetVouchersStartingTomorrow", ctx).Return([]models.Voucher{}, nil)

	err := svc.CheckAndSendReminders(ctx)
	assert.NoError(t, err)
	// HasBeenSent should NOT be called because nil UserID vouchers are skipped
	reminderRepo.AssertNotCalled(t, "HasBeenSent")
}

func TestReminderService_CheckVouchers_NilMerchant(t *testing.T) {
	reminderRepo := new(MockReminderRepo)
	voucherRepo := new(mockVoucherRepoForReminder)
	giftCardRepo := new(mockGiftCardRepoForReminder)
	notifRepo := new(mockNotifRepoForReminder)

	userID := uuid.New()
	voucherID := uuid.New()
	user := &models.User{ID: userID, Email: "test@example.com", FirstName: "Test", Language: "en", PushNotificationsEnabled: false, EmailNotificationsEnabled: false}

	svc := newTestReminderService(reminderRepo, voucherRepo, giftCardRepo, nil, nil, notifRepo, nil, nil, nil, []int{3}, time.UTC, "")
	ctx := context.Background()

	voucherNoMerchant := models.Voucher{
		ID:         voucherID,
		UserID:     &userID,
		Merchant:   nil, // nil merchant should fall back to "Unknown"
		User:       user,
		ValidUntil: time.Now().Add(3 * 24 * time.Hour),
	}

	voucherRepo.On("GetExpiringVouchers", ctx, 3).Return([]models.Voucher{voucherNoMerchant}, nil)
	giftCardRepo.On("GetExpiringGiftCards", ctx, 3).Return([]models.GiftCard{}, nil)
	voucherRepo.On("GetVouchersStartingTomorrow", ctx).Return([]models.Voucher{}, nil)

	reminderRepo.On("HasBeenSent", ctx, userID, "voucher", voucherID, 3).Return(false, nil)
	notifRepo.On("Create", ctx, mock.MatchedBy(func(n *models.Notification) bool {
		return n.Metadata["merchant_name"] == "Unknown"
	})).Return(nil)
	reminderRepo.On("MarkSent", ctx, mock.AnythingOfType("*models.ExpiryReminderSent")).Return(nil)

	err := svc.CheckAndSendReminders(ctx)
	assert.NoError(t, err)
	notifRepo.AssertExpectations(t)
}

// ==================== checkGiftCards Tests ====================

func TestReminderService_CheckGiftCards_RepoError(t *testing.T) {
	reminderRepo := new(MockReminderRepo)
	voucherRepo := new(mockVoucherRepoForReminder)
	giftCardRepo := new(mockGiftCardRepoForReminder)
	notifRepo := new(mockNotifRepoForReminder)

	svc := newTestReminderService(reminderRepo, voucherRepo, giftCardRepo, nil, nil, notifRepo, nil, nil, nil, []int{3}, time.UTC, "")
	ctx := context.Background()

	voucherRepo.On("GetExpiringVouchers", ctx, 3).Return([]models.Voucher{}, nil)
	giftCardRepo.On("GetExpiringGiftCards", ctx, 3).Return(nil, fmt.Errorf("database connection failed"))
	voucherRepo.On("GetVouchersStartingTomorrow", ctx).Return([]models.Voucher{}, nil)

	err := svc.CheckAndSendReminders(ctx)
	assert.NoError(t, err)
	giftCardRepo.AssertExpectations(t)
}

func TestReminderService_CheckGiftCards_NilUserID(t *testing.T) {
	reminderRepo := new(MockReminderRepo)
	voucherRepo := new(mockVoucherRepoForReminder)
	giftCardRepo := new(mockGiftCardRepoForReminder)
	notifRepo := new(mockNotifRepoForReminder)

	svc := newTestReminderService(reminderRepo, voucherRepo, giftCardRepo, nil, nil, notifRepo, nil, nil, nil, []int{3}, time.UTC, "")
	ctx := context.Background()

	expiresAt := time.Now().Add(3 * 24 * time.Hour)
	gcNilUser := models.GiftCard{
		ID:        uuid.New(),
		UserID:    nil,
		ExpiresAt: &expiresAt,
		Merchant:  &models.Merchant{ID: uuid.New(), Name: "Test"},
	}

	voucherRepo.On("GetExpiringVouchers", ctx, 3).Return([]models.Voucher{}, nil)
	giftCardRepo.On("GetExpiringGiftCards", ctx, 3).Return([]models.GiftCard{gcNilUser}, nil)
	voucherRepo.On("GetVouchersStartingTomorrow", ctx).Return([]models.Voucher{}, nil)

	err := svc.CheckAndSendReminders(ctx)
	assert.NoError(t, err)
	// HasBeenSent should NOT be called because nil UserID gift cards are skipped
	reminderRepo.AssertNotCalled(t, "HasBeenSent")
}

func TestReminderService_CheckGiftCards_NilExpiresAt(t *testing.T) {
	reminderRepo := new(MockReminderRepo)
	voucherRepo := new(mockVoucherRepoForReminder)
	giftCardRepo := new(mockGiftCardRepoForReminder)
	notifRepo := new(mockNotifRepoForReminder)

	userID := uuid.New()
	svc := newTestReminderService(reminderRepo, voucherRepo, giftCardRepo, nil, nil, notifRepo, nil, nil, nil, []int{3}, time.UTC, "")
	ctx := context.Background()

	gcNilExpiry := models.GiftCard{
		ID:        uuid.New(),
		UserID:    &userID,
		ExpiresAt: nil, // nil ExpiresAt should be skipped
		Merchant:  &models.Merchant{ID: uuid.New(), Name: "Test"},
	}

	voucherRepo.On("GetExpiringVouchers", ctx, 3).Return([]models.Voucher{}, nil)
	giftCardRepo.On("GetExpiringGiftCards", ctx, 3).Return([]models.GiftCard{gcNilExpiry}, nil)
	voucherRepo.On("GetVouchersStartingTomorrow", ctx).Return([]models.Voucher{}, nil)

	err := svc.CheckAndSendReminders(ctx)
	assert.NoError(t, err)
	reminderRepo.AssertNotCalled(t, "HasBeenSent")
}

func TestReminderService_CheckGiftCards_NilMerchant(t *testing.T) {
	reminderRepo := new(MockReminderRepo)
	voucherRepo := new(mockVoucherRepoForReminder)
	giftCardRepo := new(mockGiftCardRepoForReminder)
	notifRepo := new(mockNotifRepoForReminder)

	userID := uuid.New()
	gcID := uuid.New()
	expiresAt := time.Now().Add(3 * 24 * time.Hour)
	user := &models.User{ID: userID, Email: "test@example.com", FirstName: "Test", Language: "en", PushNotificationsEnabled: false, EmailNotificationsEnabled: false}

	svc := newTestReminderService(reminderRepo, voucherRepo, giftCardRepo, nil, nil, notifRepo, nil, nil, nil, []int{3}, time.UTC, "")
	ctx := context.Background()

	gcNoMerchant := models.GiftCard{
		ID:        gcID,
		UserID:    &userID,
		ExpiresAt: &expiresAt,
		Merchant:  nil, // nil merchant should fall back to "Unknown"
		User:      user,
	}

	voucherRepo.On("GetExpiringVouchers", ctx, 3).Return([]models.Voucher{}, nil)
	giftCardRepo.On("GetExpiringGiftCards", ctx, 3).Return([]models.GiftCard{gcNoMerchant}, nil)
	voucherRepo.On("GetVouchersStartingTomorrow", ctx).Return([]models.Voucher{}, nil)

	reminderRepo.On("HasBeenSent", ctx, userID, "gift_card", gcID, 3).Return(false, nil)
	notifRepo.On("Create", ctx, mock.MatchedBy(func(n *models.Notification) bool {
		return n.Metadata["merchant_name"] == "Unknown"
	})).Return(nil)
	reminderRepo.On("MarkSent", ctx, mock.AnythingOfType("*models.ExpiryReminderSent")).Return(nil)

	err := svc.CheckAndSendReminders(ctx)
	assert.NoError(t, err)
	notifRepo.AssertExpectations(t)
}

// ==================== sendValidityStartPush Tests ====================

func TestReminderService_SendValidityStartPush_NilPushService(t *testing.T) {
	svc := newTestReminderService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, time.UTC, "")
	ctx := context.Background()
	userID := uuid.New()

	// Should not panic with nil pushService
	svc.sendValidityStartPush(ctx, userID, "IKEA", "en")
}

func TestReminderService_SendValidityStartPush_Success(t *testing.T) {
	pushSvc := new(mockPushSvcForReminder)
	svc := newTestReminderService(nil, nil, nil, nil, nil, nil, pushSvc, nil, nil, nil, time.UTC, "")
	ctx := context.Background()
	userID := uuid.New()

	pushSvc.On("SendPushToUser", ctx, userID, mock.AnythingOfType("string"), mock.AnythingOfType("string"), "/vouchers").Return(nil)

	svc.sendValidityStartPush(ctx, userID, "IKEA", "en")
	pushSvc.AssertExpectations(t)
}

func TestReminderService_SendValidityStartPush_Error(t *testing.T) {
	pushSvc := new(mockPushSvcForReminder)
	svc := newTestReminderService(nil, nil, nil, nil, nil, nil, pushSvc, nil, nil, nil, time.UTC, "")
	ctx := context.Background()
	userID := uuid.New()

	pushSvc.On("SendPushToUser", ctx, userID, mock.AnythingOfType("string"), mock.AnythingOfType("string"), "/vouchers").Return(fmt.Errorf("push failed"))

	// Should not panic, just logs warning
	svc.sendValidityStartPush(ctx, userID, "IKEA", "en")
	pushSvc.AssertExpectations(t)
}

// ==================== sendValidityStartEmail Tests ====================

func TestReminderService_SendValidityStartEmail_NilEmailService(t *testing.T) {
	svc := newTestReminderService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, time.UTC, "")
	ctx := context.Background()
	user := &models.User{Email: "test@example.com", FirstName: "Test"}
	data := email.ValidityStartData{MerchantName: "IKEA"}

	// Should not panic with nil emailService
	svc.sendValidityStartEmail(ctx, user, data)
}

func TestReminderService_SendValidityStartEmail_NilUser(t *testing.T) {
	emailSvc := new(mockEmailSvcForReminder)
	svc := newTestReminderService(nil, nil, nil, nil, nil, nil, nil, emailSvc, nil, nil, time.UTC, "")
	ctx := context.Background()
	data := email.ValidityStartData{MerchantName: "IKEA"}

	// Should not panic with nil user
	svc.sendValidityStartEmail(ctx, nil, data)
	// SendValidityStart should NOT be called
	emailSvc.AssertNotCalled(t, "SendValidityStart")
}

func TestReminderService_SendValidityStartEmail_Success(t *testing.T) {
	emailSvc := new(mockEmailSvcForReminder)
	svc := newTestReminderService(nil, nil, nil, nil, nil, nil, nil, emailSvc, nil, nil, time.UTC, "https://savvy.example.com")
	ctx := context.Background()
	userID := uuid.New()
	user := &models.User{ID: userID, Email: "test@example.com", FirstName: "Test", Language: "de"}
	data := email.ValidityStartData{MerchantName: "IKEA"}

	emailSvc.On("SendValidityStart", ctx, "test@example.com", "Test", data, "", "de").Return(nil)

	svc.sendValidityStartEmail(ctx, user, data)
	emailSvc.AssertExpectations(t)
}

func TestReminderService_SendValidityStartEmail_EmptyFirstName(t *testing.T) {
	emailSvc := new(mockEmailSvcForReminder)
	svc := newTestReminderService(nil, nil, nil, nil, nil, nil, nil, emailSvc, nil, nil, time.UTC, "")
	ctx := context.Background()
	userID := uuid.New()
	user := &models.User{ID: userID, Email: "test@example.com", FirstName: "", Language: "en"}
	data := email.ValidityStartData{MerchantName: "IKEA"}

	// When FirstName is empty, should use Email as name
	emailSvc.On("SendValidityStart", ctx, "test@example.com", "test@example.com", data, "", "en").Return(nil)

	svc.sendValidityStartEmail(ctx, user, data)
	emailSvc.AssertExpectations(t)
}

func TestReminderService_SendValidityStartEmail_EmptyLanguage(t *testing.T) {
	emailSvc := new(mockEmailSvcForReminder)
	svc := newTestReminderService(nil, nil, nil, nil, nil, nil, nil, emailSvc, nil, nil, time.UTC, "")
	ctx := context.Background()
	userID := uuid.New()
	user := &models.User{ID: userID, Email: "test@example.com", FirstName: "Test", Language: ""}
	data := email.ValidityStartData{MerchantName: "IKEA"}

	// When Language is empty, should default to "en"
	emailSvc.On("SendValidityStart", ctx, "test@example.com", "Test", data, "", "en").Return(nil)

	svc.sendValidityStartEmail(ctx, user, data)
	emailSvc.AssertExpectations(t)
}

func TestReminderService_SendValidityStartEmail_Error(t *testing.T) {
	emailSvc := new(mockEmailSvcForReminder)
	svc := newTestReminderService(nil, nil, nil, nil, nil, nil, nil, emailSvc, nil, nil, time.UTC, "")
	ctx := context.Background()
	userID := uuid.New()
	user := &models.User{ID: userID, Email: "test@example.com", FirstName: "Test", Language: "en"}
	data := email.ValidityStartData{MerchantName: "IKEA"}

	emailSvc.On("SendValidityStart", ctx, "test@example.com", "Test", data, "", "en").Return(fmt.Errorf("smtp error"))

	// Should not panic, just logs warning
	svc.sendValidityStartEmail(ctx, user, data)
	emailSvc.AssertExpectations(t)
}

// ==================== sendEmail Tests (expiry reminder) ====================

func TestReminderService_SendEmail_NilEmailService(t *testing.T) {
	svc := newTestReminderService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, time.UTC, "")
	ctx := context.Background()
	user := &models.User{Email: "test@example.com"}
	data := email.ExpiryReminderData{}

	// Should not panic
	svc.sendEmail(ctx, user, data)
}

func TestReminderService_SendEmail_NilUser(t *testing.T) {
	emailSvc := new(mockEmailSvcForReminder)
	svc := newTestReminderService(nil, nil, nil, nil, nil, nil, nil, emailSvc, nil, nil, time.UTC, "")
	ctx := context.Background()
	data := email.ExpiryReminderData{}

	// Should not panic
	svc.sendEmail(ctx, nil, data)
	emailSvc.AssertNotCalled(t, "SendExpiryReminder")
}

func TestReminderService_SendEmail_EmptyFirstNameUsesEmail(t *testing.T) {
	emailSvc := new(mockEmailSvcForReminder)
	svc := newTestReminderService(nil, nil, nil, nil, nil, nil, nil, emailSvc, nil, nil, time.UTC, "")
	ctx := context.Background()
	userID := uuid.New()
	user := &models.User{ID: userID, Email: "test@example.com", FirstName: "", Language: "en"}
	data := email.ExpiryReminderData{MerchantName: "Test"}

	emailSvc.On("SendExpiryReminder", ctx, "test@example.com", "test@example.com", data, "", "en").Return(nil)

	svc.sendEmail(ctx, user, data)
	emailSvc.AssertExpectations(t)
}

func TestReminderService_SendEmail_EmptyLanguageDefaultsToEn(t *testing.T) {
	emailSvc := new(mockEmailSvcForReminder)
	svc := newTestReminderService(nil, nil, nil, nil, nil, nil, nil, emailSvc, nil, nil, time.UTC, "")
	ctx := context.Background()
	userID := uuid.New()
	user := &models.User{ID: userID, Email: "test@example.com", FirstName: "Test", Language: ""}
	data := email.ExpiryReminderData{MerchantName: "Test"}

	emailSvc.On("SendExpiryReminder", ctx, "test@example.com", "Test", data, "", "en").Return(nil)

	svc.sendEmail(ctx, user, data)
	emailSvc.AssertExpectations(t)
}

// ==================== sendPush Tests (expiry reminder) ====================

func TestReminderService_SendPush_NilPushService(t *testing.T) {
	svc := newTestReminderService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, time.UTC, "")
	ctx := context.Background()
	userID := uuid.New()

	// Should not panic
	svc.sendPush(ctx, userID, "IKEA", "voucher", 3, "en")
}

func TestReminderService_SendPush_OneDayLeft(t *testing.T) {
	pushSvc := new(mockPushSvcForReminder)
	svc := newTestReminderService(nil, nil, nil, nil, nil, nil, pushSvc, nil, nil, nil, time.UTC, "")
	ctx := context.Background()
	userID := uuid.New()

	// daysLeft == 1 triggers the "tomorrow" message body
	pushSvc.On("SendPushToUser", ctx, userID, mock.AnythingOfType("string"), mock.AnythingOfType("string"), "/vouchers").Return(nil)

	svc.sendPush(ctx, userID, "IKEA", "voucher", 1, "en")
	pushSvc.AssertExpectations(t)
}

func TestReminderService_SendPush_MultipleDays(t *testing.T) {
	pushSvc := new(mockPushSvcForReminder)
	svc := newTestReminderService(nil, nil, nil, nil, nil, nil, pushSvc, nil, nil, nil, time.UTC, "")
	ctx := context.Background()
	userID := uuid.New()

	// daysLeft > 1 triggers the general days message body
	pushSvc.On("SendPushToUser", ctx, userID, mock.AnythingOfType("string"), mock.AnythingOfType("string"), "/gift_cards").Return(nil)

	svc.sendPush(ctx, userID, "Amazon", "gift_card", 7, "de")
	pushSvc.AssertExpectations(t)
}

func TestReminderService_SendPush_Error(t *testing.T) {
	pushSvc := new(mockPushSvcForReminder)
	svc := newTestReminderService(nil, nil, nil, nil, nil, nil, pushSvc, nil, nil, nil, time.UTC, "")
	ctx := context.Background()
	userID := uuid.New()

	pushSvc.On("SendPushToUser", ctx, userID, mock.AnythingOfType("string"), mock.AnythingOfType("string"), "/vouchers").Return(fmt.Errorf("push error"))

	// Should not panic
	svc.sendPush(ctx, userID, "IKEA", "voucher", 3, "en")
	pushSvc.AssertExpectations(t)
}

// ==================== sendReminderToUser Tests ====================

func TestReminderService_SendReminderToUser_HasBeenSentError(t *testing.T) {
	reminderRepo := new(MockReminderRepo)
	notifRepo := new(mockNotifRepoForReminder)

	svc := newTestReminderService(reminderRepo, nil, nil, nil, nil, notifRepo, nil, nil, nil, nil, time.UTC, "")
	ctx := context.Background()
	userID := uuid.New()
	resourceID := uuid.New()

	reminderRepo.On("HasBeenSent", ctx, userID, "voucher", resourceID, 3).Return(false, fmt.Errorf("db error"))

	result := svc.sendReminderToUser(ctx, userID, nil, "voucher", resourceID, 3, 3, "IKEA", email.ExpiryReminderData{})
	assert.False(t, result)
	// Notification should NOT be created
	notifRepo.AssertNotCalled(t, "Create")
}

func TestReminderService_SendReminderToUser_NotifCreateError(t *testing.T) {
	reminderRepo := new(MockReminderRepo)
	notifRepo := new(mockNotifRepoForReminder)

	svc := newTestReminderService(reminderRepo, nil, nil, nil, nil, notifRepo, nil, nil, nil, nil, time.UTC, "")
	ctx := context.Background()
	userID := uuid.New()
	resourceID := uuid.New()

	reminderRepo.On("HasBeenSent", ctx, userID, "voucher", resourceID, 3).Return(false, nil)
	notifRepo.On("Create", ctx, mock.AnythingOfType("*models.Notification")).Return(fmt.Errorf("db error"))

	result := svc.sendReminderToUser(ctx, userID, nil, "voucher", resourceID, 3, 3, "IKEA", email.ExpiryReminderData{})
	assert.False(t, result)
	// MarkSent should NOT be called
	reminderRepo.AssertNotCalled(t, "MarkSent")
}

func TestReminderService_SendReminderToUser_NilUser(t *testing.T) {
	reminderRepo := new(MockReminderRepo)
	notifRepo := new(mockNotifRepoForReminder)
	pushSvc := new(mockPushSvcForReminder)
	emailSvc := new(mockEmailSvcForReminder)

	svc := newTestReminderService(reminderRepo, nil, nil, nil, nil, notifRepo, pushSvc, emailSvc, nil, nil, time.UTC, "")
	ctx := context.Background()
	userID := uuid.New()
	resourceID := uuid.New()

	reminderRepo.On("HasBeenSent", ctx, userID, "voucher", resourceID, 3).Return(false, nil)
	notifRepo.On("Create", ctx, mock.AnythingOfType("*models.Notification")).Return(nil)
	// When user is nil, push and email are still sent (nil user falls through the preference check)
	pushSvc.On("SendPushToUser", ctx, userID, mock.AnythingOfType("string"), mock.AnythingOfType("string"), "/vouchers").Return(nil)
	// Email won't be sent because sendEmail checks for nil user
	reminderRepo.On("MarkSent", ctx, mock.AnythingOfType("*models.ExpiryReminderSent")).Return(nil)

	result := svc.sendReminderToUser(ctx, userID, nil, "voucher", resourceID, 3, 3, "IKEA", email.ExpiryReminderData{})
	assert.True(t, result)
	reminderRepo.AssertExpectations(t)
	notifRepo.AssertExpectations(t)
	pushSvc.AssertExpectations(t)
	// sendEmail returns early for nil user
	emailSvc.AssertNotCalled(t, "SendExpiryReminder")
}

// ==================== checkVoucherValidityStart Tests ====================

func TestReminderService_CheckVoucherValidityStart_RepoError(t *testing.T) {
	reminderRepo := new(MockReminderRepo)
	voucherRepo := new(mockVoucherRepoForReminder)
	giftCardRepo := new(mockGiftCardRepoForReminder)
	notifRepo := new(mockNotifRepoForReminder)

	svc := newTestReminderService(reminderRepo, voucherRepo, giftCardRepo, nil, nil, notifRepo, nil, nil, nil, []int{3}, time.UTC, "")
	ctx := context.Background()

	voucherRepo.On("GetExpiringVouchers", ctx, 3).Return([]models.Voucher{}, nil)
	giftCardRepo.On("GetExpiringGiftCards", ctx, 3).Return([]models.GiftCard{}, nil)
	voucherRepo.On("GetVouchersStartingTomorrow", ctx).Return(nil, fmt.Errorf("database error"))

	err := svc.CheckAndSendReminders(ctx)
	assert.NoError(t, err) // Errors are logged, not propagated
	voucherRepo.AssertExpectations(t)
}

func TestReminderService_CheckVoucherValidityStart_NilUserID(t *testing.T) {
	reminderRepo := new(MockReminderRepo)
	voucherRepo := new(mockVoucherRepoForReminder)
	giftCardRepo := new(mockGiftCardRepoForReminder)
	notifRepo := new(mockNotifRepoForReminder)

	svc := newTestReminderService(reminderRepo, voucherRepo, giftCardRepo, nil, nil, notifRepo, nil, nil, nil, []int{3}, time.UTC, "")
	ctx := context.Background()

	tomorrow := time.Now().AddDate(0, 0, 1)
	voucherNilUser := models.Voucher{
		ID:        uuid.New(),
		UserID:    nil,
		ValidFrom: tomorrow,
		Merchant:  &models.Merchant{ID: uuid.New(), Name: "Test"},
	}

	voucherRepo.On("GetExpiringVouchers", ctx, 3).Return([]models.Voucher{}, nil)
	giftCardRepo.On("GetExpiringGiftCards", ctx, 3).Return([]models.GiftCard{}, nil)
	voucherRepo.On("GetVouchersStartingTomorrow", ctx).Return([]models.Voucher{voucherNilUser}, nil)

	err := svc.CheckAndSendReminders(ctx)
	assert.NoError(t, err)
	// HasBeenSent should NOT be called
	reminderRepo.AssertNotCalled(t, "HasBeenSent")
}

// ==================== Share Recipient Reminder Tests ====================

func TestReminderService_VoucherShareRecipients(t *testing.T) {
	reminderRepo := new(MockReminderRepo)
	voucherRepo := new(mockVoucherRepoForReminder)
	giftCardRepo := new(mockGiftCardRepoForReminder)
	notifRepo := new(mockNotifRepoForReminder)
	pushSvc := new(mockPushSvcForReminder)
	emailSvc := new(mockEmailSvcForReminder)
	voucherShareRepo := new(mockVoucherShareRepoForReminder)

	svc := newTestReminderService(reminderRepo, voucherRepo, giftCardRepo, voucherShareRepo, nil, notifRepo, pushSvc, emailSvc, nil, []int{3}, time.UTC, "https://savvy.example.com")
	ctx := context.Background()

	ownerID := uuid.New()
	sharedWithID := uuid.New()
	voucherID := uuid.New()
	merchant := &models.Merchant{ID: uuid.New(), Name: "IKEA"}
	ownerUser := &models.User{ID: ownerID, Email: "owner@example.com", FirstName: "Owner", Language: "en", PushRemindersEnabled: true, EmailRemindersEnabled: true, PushNotificationsEnabled: true, EmailNotificationsEnabled: true}
	sharedUser := &models.User{ID: sharedWithID, Email: "shared@example.com", FirstName: "Shared", Language: "de", PushRemindersEnabled: true, EmailRemindersEnabled: true, PushNotificationsEnabled: true, EmailNotificationsEnabled: true}

	expiringVoucher := models.Voucher{
		ID:         voucherID,
		UserID:     &ownerID,
		ValidUntil: time.Now().Add(3 * 24 * time.Hour),
		Merchant:   merchant,
		User:       ownerUser,
	}

	voucherRepo.On("GetExpiringVouchers", ctx, 3).Return([]models.Voucher{expiringVoucher}, nil)
	giftCardRepo.On("GetExpiringGiftCards", ctx, 3).Return([]models.GiftCard{}, nil)
	voucherRepo.On("GetVouchersStartingTomorrow", ctx).Return([]models.Voucher{}, nil)

	// Owner reminder
	reminderRepo.On("HasBeenSent", ctx, ownerID, "voucher", voucherID, 3).Return(false, nil)
	notifRepo.On("Create", ctx, mock.MatchedBy(func(n *models.Notification) bool {
		return n.UserID == ownerID
	})).Return(nil)
	pushSvc.On("SendPushToUser", ctx, ownerID, mock.AnythingOfType("string"), mock.AnythingOfType("string"), "/vouchers").Return(nil)
	emailSvc.On("SendExpiryReminder", ctx, "owner@example.com", "Owner", mock.AnythingOfType("email.ExpiryReminderData"), "", "en").Return(nil)
	reminderRepo.On("MarkSent", ctx, mock.MatchedBy(func(r *models.ExpiryReminderSent) bool {
		return r.UserID == ownerID
	})).Return(nil)

	// Shared user reminder
	shares := []models.VoucherShare{
		{
			VoucherID:      voucherID,
			SharedWithID:   sharedWithID,
			SharedWithUser: sharedUser,
		},
	}
	voucherShareRepo.On("GetByVoucherID", ctx, voucherID).Return(shares, nil)

	reminderRepo.On("HasBeenSent", ctx, sharedWithID, "voucher", voucherID, 3).Return(false, nil)
	notifRepo.On("Create", ctx, mock.MatchedBy(func(n *models.Notification) bool {
		return n.UserID == sharedWithID
	})).Return(nil)
	pushSvc.On("SendPushToUser", ctx, sharedWithID, mock.AnythingOfType("string"), mock.AnythingOfType("string"), "/vouchers").Return(nil)
	emailSvc.On("SendExpiryReminder", ctx, "shared@example.com", "Shared", mock.AnythingOfType("email.ExpiryReminderData"), "", "de").Return(nil)
	reminderRepo.On("MarkSent", ctx, mock.MatchedBy(func(r *models.ExpiryReminderSent) bool {
		return r.UserID == sharedWithID
	})).Return(nil)

	err := svc.CheckAndSendReminders(ctx)
	assert.NoError(t, err)
	notifRepo.AssertExpectations(t)
	reminderRepo.AssertExpectations(t)
	voucherShareRepo.AssertExpectations(t)
}

func TestReminderService_GiftCardShareRecipients(t *testing.T) {
	reminderRepo := new(MockReminderRepo)
	voucherRepo := new(mockVoucherRepoForReminder)
	giftCardRepo := new(mockGiftCardRepoForReminder)
	notifRepo := new(mockNotifRepoForReminder)
	pushSvc := new(mockPushSvcForReminder)
	emailSvc := new(mockEmailSvcForReminder)
	giftCardShareRepo := new(mockGiftCardShareRepoForReminder)

	svc := newTestReminderService(reminderRepo, voucherRepo, giftCardRepo, nil, giftCardShareRepo, notifRepo, pushSvc, emailSvc, nil, []int{1}, time.UTC, "https://savvy.example.com")
	ctx := context.Background()

	ownerID := uuid.New()
	sharedWithID := uuid.New()
	gcID := uuid.New()
	expiresAt := time.Now().Add(1 * 24 * time.Hour)
	merchant := &models.Merchant{ID: uuid.New(), Name: "Amazon"}
	ownerUser := &models.User{ID: ownerID, Email: "owner@example.com", FirstName: "Owner", Language: "en", PushRemindersEnabled: true, EmailRemindersEnabled: true, PushNotificationsEnabled: true, EmailNotificationsEnabled: true}
	sharedUser := &models.User{ID: sharedWithID, Email: "shared@example.com", FirstName: "Shared", Language: "fr", PushRemindersEnabled: true, EmailRemindersEnabled: true, PushNotificationsEnabled: true, EmailNotificationsEnabled: true}

	expiringGC := models.GiftCard{
		ID:        gcID,
		UserID:    &ownerID,
		ExpiresAt: &expiresAt,
		Merchant:  merchant,
		User:      ownerUser,
	}

	voucherRepo.On("GetExpiringVouchers", ctx, 1).Return([]models.Voucher{}, nil)
	giftCardRepo.On("GetExpiringGiftCards", ctx, 1).Return([]models.GiftCard{expiringGC}, nil)
	voucherRepo.On("GetVouchersStartingTomorrow", ctx).Return([]models.Voucher{}, nil)

	// Owner reminder
	reminderRepo.On("HasBeenSent", ctx, ownerID, "gift_card", gcID, 1).Return(false, nil)
	notifRepo.On("Create", ctx, mock.MatchedBy(func(n *models.Notification) bool {
		return n.UserID == ownerID
	})).Return(nil)
	pushSvc.On("SendPushToUser", ctx, ownerID, mock.AnythingOfType("string"), mock.AnythingOfType("string"), "/gift_cards").Return(nil)
	emailSvc.On("SendExpiryReminder", ctx, "owner@example.com", "Owner", mock.AnythingOfType("email.ExpiryReminderData"), "", "en").Return(nil)
	reminderRepo.On("MarkSent", ctx, mock.MatchedBy(func(r *models.ExpiryReminderSent) bool {
		return r.UserID == ownerID
	})).Return(nil)

	// Shared user reminder
	shares := []models.GiftCardShare{
		{
			GiftCardID:     gcID,
			SharedWithID:   sharedWithID,
			SharedWithUser: sharedUser,
		},
	}
	giftCardShareRepo.On("GetByGiftCardID", ctx, gcID).Return(shares, nil)

	reminderRepo.On("HasBeenSent", ctx, sharedWithID, "gift_card", gcID, 1).Return(false, nil)
	notifRepo.On("Create", ctx, mock.MatchedBy(func(n *models.Notification) bool {
		return n.UserID == sharedWithID
	})).Return(nil)
	pushSvc.On("SendPushToUser", ctx, sharedWithID, mock.AnythingOfType("string"), mock.AnythingOfType("string"), "/gift_cards").Return(nil)
	emailSvc.On("SendExpiryReminder", ctx, "shared@example.com", "Shared", mock.AnythingOfType("email.ExpiryReminderData"), "", "fr").Return(nil)
	reminderRepo.On("MarkSent", ctx, mock.MatchedBy(func(r *models.ExpiryReminderSent) bool {
		return r.UserID == sharedWithID
	})).Return(nil)

	err := svc.CheckAndSendReminders(ctx)
	assert.NoError(t, err)
	notifRepo.AssertExpectations(t)
	reminderRepo.AssertExpectations(t)
	giftCardShareRepo.AssertExpectations(t)
}

// ==================== NewReminderService Tests ====================

func TestReminderService_NewReminderService_NilLocation(t *testing.T) {
	// When location is nil, it should default to UTC
	svc := NewReminderService(nil, nil, nil, nil, nil, nil, nil, nil, nil, []int{3}, nil, "https://savvy.example.com/")
	rs := svc.(*ReminderService)
	assert.Equal(t, time.UTC, rs.location)
	// Trailing slash should be stripped
	assert.Equal(t, "https://savvy.example.com", rs.frontendURL)
}

// ==================== Validity Start Share Recipients ====================

func TestReminderService_ValidityStart_ShareRecipients(t *testing.T) {
	reminderRepo := new(MockReminderRepo)
	voucherRepo := new(mockVoucherRepoForReminder)
	giftCardRepo := new(mockGiftCardRepoForReminder)
	notifRepo := new(mockNotifRepoForReminder)
	voucherShareRepo := new(mockVoucherShareRepoForReminder)

	svc := newTestReminderService(reminderRepo, voucherRepo, giftCardRepo, voucherShareRepo, nil, notifRepo, nil, nil, nil, []int{3}, time.UTC, "")
	ctx := context.Background()

	ownerID := uuid.New()
	sharedWithID := uuid.New()
	voucherID := uuid.New()
	merchant := &models.Merchant{ID: uuid.New(), Name: "IKEA"}
	ownerUser := &models.User{ID: ownerID, Email: "owner@example.com", FirstName: "Owner", Language: "en", PushNotificationsEnabled: false, EmailNotificationsEnabled: false}
	sharedUser := &models.User{ID: sharedWithID, Email: "shared@example.com", FirstName: "Shared", Language: "de", PushNotificationsEnabled: false, EmailNotificationsEnabled: false}

	tomorrow := time.Now().AddDate(0, 0, 1)
	startingVoucher := models.Voucher{
		ID:        voucherID,
		UserID:    &ownerID,
		ValidFrom: tomorrow,
		Merchant:  merchant,
		User:      ownerUser,
	}

	voucherRepo.On("GetExpiringVouchers", ctx, 3).Return([]models.Voucher{}, nil)
	giftCardRepo.On("GetExpiringGiftCards", ctx, 3).Return([]models.GiftCard{}, nil)
	voucherRepo.On("GetVouchersStartingTomorrow", ctx).Return([]models.Voucher{startingVoucher}, nil)

	// Owner
	reminderRepo.On("HasBeenSent", ctx, ownerID, "voucher_start", voucherID, 1).Return(false, nil)
	notifRepo.On("Create", ctx, mock.MatchedBy(func(n *models.Notification) bool {
		return n.UserID == ownerID && n.Type == models.NotificationTypeValidityStart
	})).Return(nil)
	reminderRepo.On("MarkSent", ctx, mock.MatchedBy(func(r *models.ExpiryReminderSent) bool {
		return r.UserID == ownerID && r.ResourceType == "voucher_start"
	})).Return(nil)

	// Shared user
	shares := []models.VoucherShare{
		{
			VoucherID:      voucherID,
			SharedWithID:   sharedWithID,
			SharedWithUser: sharedUser,
		},
	}
	voucherShareRepo.On("GetByVoucherID", ctx, voucherID).Return(shares, nil)

	reminderRepo.On("HasBeenSent", ctx, sharedWithID, "voucher_start", voucherID, 1).Return(false, nil)
	notifRepo.On("Create", ctx, mock.MatchedBy(func(n *models.Notification) bool {
		return n.UserID == sharedWithID && n.Type == models.NotificationTypeValidityStart
	})).Return(nil)
	reminderRepo.On("MarkSent", ctx, mock.MatchedBy(func(r *models.ExpiryReminderSent) bool {
		return r.UserID == sharedWithID && r.ResourceType == "voucher_start"
	})).Return(nil)

	err := svc.CheckAndSendReminders(ctx)
	assert.NoError(t, err)
	notifRepo.AssertExpectations(t)
	reminderRepo.AssertExpectations(t)
	voucherShareRepo.AssertExpectations(t)
}

func TestReminderService_ValidityStart_ShareRepoError(t *testing.T) {
	reminderRepo := new(MockReminderRepo)
	voucherRepo := new(mockVoucherRepoForReminder)
	giftCardRepo := new(mockGiftCardRepoForReminder)
	notifRepo := new(mockNotifRepoForReminder)
	voucherShareRepo := new(mockVoucherShareRepoForReminder)

	svc := newTestReminderService(reminderRepo, voucherRepo, giftCardRepo, voucherShareRepo, nil, notifRepo, nil, nil, nil, []int{3}, time.UTC, "")
	ctx := context.Background()

	ownerID := uuid.New()
	voucherID := uuid.New()
	merchant := &models.Merchant{ID: uuid.New(), Name: "IKEA"}
	ownerUser := &models.User{ID: ownerID, Email: "owner@example.com", FirstName: "Owner", Language: "en", PushNotificationsEnabled: false, EmailNotificationsEnabled: false}

	tomorrow := time.Now().AddDate(0, 0, 1)
	startingVoucher := models.Voucher{
		ID:        voucherID,
		UserID:    &ownerID,
		ValidFrom: tomorrow,
		Merchant:  merchant,
		User:      ownerUser,
	}

	voucherRepo.On("GetExpiringVouchers", ctx, 3).Return([]models.Voucher{}, nil)
	giftCardRepo.On("GetExpiringGiftCards", ctx, 3).Return([]models.GiftCard{}, nil)
	voucherRepo.On("GetVouchersStartingTomorrow", ctx).Return([]models.Voucher{startingVoucher}, nil)

	reminderRepo.On("HasBeenSent", ctx, ownerID, "voucher_start", voucherID, 1).Return(false, nil)
	notifRepo.On("Create", ctx, mock.AnythingOfType("*models.Notification")).Return(nil)
	reminderRepo.On("MarkSent", ctx, mock.AnythingOfType("*models.ExpiryReminderSent")).Return(nil)

	// Share repo returns error - should be logged but not fail
	voucherShareRepo.On("GetByVoucherID", ctx, voucherID).Return(nil, fmt.Errorf("share repo error"))

	err := svc.CheckAndSendReminders(ctx)
	assert.NoError(t, err)
	voucherShareRepo.AssertExpectations(t)
}

// ==================== Voucher Share Repo Error Tests ====================

func TestReminderService_VoucherShareRepoError(t *testing.T) {
	reminderRepo := new(MockReminderRepo)
	voucherRepo := new(mockVoucherRepoForReminder)
	giftCardRepo := new(mockGiftCardRepoForReminder)
	notifRepo := new(mockNotifRepoForReminder)
	voucherShareRepo := new(mockVoucherShareRepoForReminder)

	svc := newTestReminderService(reminderRepo, voucherRepo, giftCardRepo, voucherShareRepo, nil, notifRepo, nil, nil, nil, []int{3}, time.UTC, "")
	ctx := context.Background()

	ownerID := uuid.New()
	voucherID := uuid.New()
	merchant := &models.Merchant{ID: uuid.New(), Name: "IKEA"}
	ownerUser := &models.User{ID: ownerID, Email: "owner@example.com", FirstName: "Owner", Language: "en", PushNotificationsEnabled: false, EmailNotificationsEnabled: false}

	expiringVoucher := models.Voucher{
		ID:         voucherID,
		UserID:     &ownerID,
		ValidUntil: time.Now().Add(3 * 24 * time.Hour),
		Merchant:   merchant,
		User:       ownerUser,
	}

	voucherRepo.On("GetExpiringVouchers", ctx, 3).Return([]models.Voucher{expiringVoucher}, nil)
	giftCardRepo.On("GetExpiringGiftCards", ctx, 3).Return([]models.GiftCard{}, nil)
	voucherRepo.On("GetVouchersStartingTomorrow", ctx).Return([]models.Voucher{}, nil)

	reminderRepo.On("HasBeenSent", ctx, ownerID, "voucher", voucherID, 3).Return(false, nil)
	notifRepo.On("Create", ctx, mock.AnythingOfType("*models.Notification")).Return(nil)
	reminderRepo.On("MarkSent", ctx, mock.AnythingOfType("*models.ExpiryReminderSent")).Return(nil)

	// Share repo returns error - should be logged but not break
	voucherShareRepo.On("GetByVoucherID", ctx, voucherID).Return(nil, fmt.Errorf("share repo error"))

	err := svc.CheckAndSendReminders(ctx)
	assert.NoError(t, err)
	voucherShareRepo.AssertExpectations(t)
}

func TestReminderService_GiftCardShareRepoError(t *testing.T) {
	reminderRepo := new(MockReminderRepo)
	voucherRepo := new(mockVoucherRepoForReminder)
	giftCardRepo := new(mockGiftCardRepoForReminder)
	notifRepo := new(mockNotifRepoForReminder)
	giftCardShareRepo := new(mockGiftCardShareRepoForReminder)

	svc := newTestReminderService(reminderRepo, voucherRepo, giftCardRepo, nil, giftCardShareRepo, notifRepo, nil, nil, nil, []int{1}, time.UTC, "")
	ctx := context.Background()

	ownerID := uuid.New()
	gcID := uuid.New()
	expiresAt := time.Now().Add(1 * 24 * time.Hour)
	merchant := &models.Merchant{ID: uuid.New(), Name: "Amazon"}
	ownerUser := &models.User{ID: ownerID, Email: "owner@example.com", FirstName: "Owner", Language: "en", PushNotificationsEnabled: false, EmailNotificationsEnabled: false}

	expiringGC := models.GiftCard{
		ID:        gcID,
		UserID:    &ownerID,
		ExpiresAt: &expiresAt,
		Merchant:  merchant,
		User:      ownerUser,
	}

	voucherRepo.On("GetExpiringVouchers", ctx, 1).Return([]models.Voucher{}, nil)
	giftCardRepo.On("GetExpiringGiftCards", ctx, 1).Return([]models.GiftCard{expiringGC}, nil)
	voucherRepo.On("GetVouchersStartingTomorrow", ctx).Return([]models.Voucher{}, nil)

	reminderRepo.On("HasBeenSent", ctx, ownerID, "gift_card", gcID, 1).Return(false, nil)
	notifRepo.On("Create", ctx, mock.AnythingOfType("*models.Notification")).Return(nil)
	reminderRepo.On("MarkSent", ctx, mock.AnythingOfType("*models.ExpiryReminderSent")).Return(nil)

	// Share repo returns error - should be logged but not break
	giftCardShareRepo.On("GetByGiftCardID", ctx, gcID).Return(nil, fmt.Errorf("share repo error"))

	err := svc.CheckAndSendReminders(ctx)
	assert.NoError(t, err)
	giftCardShareRepo.AssertExpectations(t)
}

// ==================== Share With Nil SharedWithUser Tests ====================

func TestReminderService_VoucherShare_NilSharedWithUser(t *testing.T) {
	reminderRepo := new(MockReminderRepo)
	voucherRepo := new(mockVoucherRepoForReminder)
	giftCardRepo := new(mockGiftCardRepoForReminder)
	notifRepo := new(mockNotifRepoForReminder)
	voucherShareRepo := new(mockVoucherShareRepoForReminder)

	svc := newTestReminderService(reminderRepo, voucherRepo, giftCardRepo, voucherShareRepo, nil, notifRepo, nil, nil, nil, []int{3}, time.UTC, "")
	ctx := context.Background()

	ownerID := uuid.New()
	voucherID := uuid.New()
	merchant := &models.Merchant{ID: uuid.New(), Name: "IKEA"}
	ownerUser := &models.User{ID: ownerID, Email: "owner@example.com", FirstName: "Owner", Language: "en", PushNotificationsEnabled: false, EmailNotificationsEnabled: false}

	expiringVoucher := models.Voucher{
		ID:         voucherID,
		UserID:     &ownerID,
		ValidUntil: time.Now().Add(3 * 24 * time.Hour),
		Merchant:   merchant,
		User:       ownerUser,
	}

	voucherRepo.On("GetExpiringVouchers", ctx, 3).Return([]models.Voucher{expiringVoucher}, nil)
	giftCardRepo.On("GetExpiringGiftCards", ctx, 3).Return([]models.GiftCard{}, nil)
	voucherRepo.On("GetVouchersStartingTomorrow", ctx).Return([]models.Voucher{}, nil)

	reminderRepo.On("HasBeenSent", ctx, ownerID, "voucher", voucherID, 3).Return(false, nil)
	notifRepo.On("Create", ctx, mock.AnythingOfType("*models.Notification")).Return(nil)
	reminderRepo.On("MarkSent", ctx, mock.AnythingOfType("*models.ExpiryReminderSent")).Return(nil)

	// Share with nil SharedWithUser should be skipped
	shares := []models.VoucherShare{
		{
			VoucherID:      voucherID,
			SharedWithID:   uuid.New(),
			SharedWithUser: nil, // nil user
		},
	}
	voucherShareRepo.On("GetByVoucherID", ctx, voucherID).Return(shares, nil)

	err := svc.CheckAndSendReminders(ctx)
	assert.NoError(t, err)
	// Only owner reminder should be sent, not the shared user with nil user
	voucherShareRepo.AssertExpectations(t)
}

func TestReminderService_GiftCardShare_NilSharedWithUser(t *testing.T) {
	reminderRepo := new(MockReminderRepo)
	voucherRepo := new(mockVoucherRepoForReminder)
	giftCardRepo := new(mockGiftCardRepoForReminder)
	notifRepo := new(mockNotifRepoForReminder)
	giftCardShareRepo := new(mockGiftCardShareRepoForReminder)

	svc := newTestReminderService(reminderRepo, voucherRepo, giftCardRepo, nil, giftCardShareRepo, notifRepo, nil, nil, nil, []int{1}, time.UTC, "")
	ctx := context.Background()

	ownerID := uuid.New()
	gcID := uuid.New()
	expiresAt := time.Now().Add(1 * 24 * time.Hour)
	merchant := &models.Merchant{ID: uuid.New(), Name: "Amazon"}
	ownerUser := &models.User{ID: ownerID, Email: "owner@example.com", FirstName: "Owner", Language: "en", PushNotificationsEnabled: false, EmailNotificationsEnabled: false}

	expiringGC := models.GiftCard{
		ID:        gcID,
		UserID:    &ownerID,
		ExpiresAt: &expiresAt,
		Merchant:  merchant,
		User:      ownerUser,
	}

	voucherRepo.On("GetExpiringVouchers", ctx, 1).Return([]models.Voucher{}, nil)
	giftCardRepo.On("GetExpiringGiftCards", ctx, 1).Return([]models.GiftCard{expiringGC}, nil)
	voucherRepo.On("GetVouchersStartingTomorrow", ctx).Return([]models.Voucher{}, nil)

	reminderRepo.On("HasBeenSent", ctx, ownerID, "gift_card", gcID, 1).Return(false, nil)
	notifRepo.On("Create", ctx, mock.AnythingOfType("*models.Notification")).Return(nil)
	reminderRepo.On("MarkSent", ctx, mock.AnythingOfType("*models.ExpiryReminderSent")).Return(nil)

	// Share with nil SharedWithUser should be skipped
	shares := []models.GiftCardShare{
		{
			GiftCardID:     gcID,
			SharedWithID:   uuid.New(),
			SharedWithUser: nil,
		},
	}
	giftCardShareRepo.On("GetByGiftCardID", ctx, gcID).Return(shares, nil)

	err := svc.CheckAndSendReminders(ctx)
	assert.NoError(t, err)
	giftCardShareRepo.AssertExpectations(t)
}
