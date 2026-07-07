package services

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
	"savvy/internal/email"
	"savvy/internal/models"
	"savvy/internal/repository"
)

// ==================== Mock Definitions ====================

type mockCardRepo struct{ mock.Mock }

func (m *mockCardRepo) Create(ctx context.Context, card *models.Card) error {
	return m.Called(ctx, card).Error(0)
}
func (m *mockCardRepo) GetByID(ctx context.Context, id uuid.UUID, _ ...string) (*models.Card, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Card), args.Error(1)
}
func (m *mockCardRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.Card, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.Card), args.Error(1)
}
func (m *mockCardRepo) GetSharedWithUser(ctx context.Context, userID uuid.UUID) ([]models.Card, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.Card), args.Error(1)
}
func (m *mockCardRepo) Update(ctx context.Context, card *models.Card) error {
	return m.Called(ctx, card).Error(0)
}
func (m *mockCardRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockCardRepo) Count(ctx context.Context, userID uuid.UUID) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockCardRepo) GetAllForUserPaginated(_ context.Context, _ uuid.UUID, _ repository.PaginationParams) (*repository.PaginatedResult[models.Card], error) {
	return nil, nil
}
func (m *mockCardRepo) FindByCardNumber(_ context.Context, _ string, _ uuid.UUID) (*models.Card, error) {
	return nil, nil
}
func (m *mockCardRepo) FindSharedByCardNumber(_ context.Context, _ string, _ *uuid.UUID, _ uuid.UUID) (*models.Card, error) {
	return nil, nil
}
func (m *mockCardRepo) FindDeletedByCardNumber(_ context.Context, _ string, _ uuid.UUID) (*models.Card, error) {
	return nil, nil
}
func (m *mockCardRepo) RestoreByID(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
	return nil
}
func (m *mockCardRepo) Search(_ context.Context, _ uuid.UUID, _ string) ([]models.Card, error) {
	return nil, nil
}

var _ repository.CardRepository = (*mockCardRepo)(nil)

type mockVoucherRepo struct{ mock.Mock }

func (m *mockVoucherRepo) Create(ctx context.Context, v *models.Voucher) error {
	return m.Called(ctx, v).Error(0)
}
func (m *mockVoucherRepo) GetByID(ctx context.Context, id uuid.UUID, _ ...string) (*models.Voucher, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Voucher), args.Error(1)
}
func (m *mockVoucherRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.Voucher, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.Voucher), args.Error(1)
}
func (m *mockVoucherRepo) GetSharedWithUser(ctx context.Context, userID uuid.UUID) ([]models.Voucher, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.Voucher), args.Error(1)
}
func (m *mockVoucherRepo) Update(ctx context.Context, v *models.Voucher) error {
	return m.Called(ctx, v).Error(0)
}
func (m *mockVoucherRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockVoucherRepo) Count(ctx context.Context, userID uuid.UUID) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockVoucherRepo) GetAllForUserPaginated(_ context.Context, _ uuid.UUID, _ repository.PaginationParams) (*repository.PaginatedResult[models.Voucher], error) {
	return nil, nil
}
func (m *mockVoucherRepo) GetExpiringVouchers(_ context.Context, _ int) ([]models.Voucher, error) {
	return nil, nil
}
func (m *mockVoucherRepo) GetVouchersStartingTomorrow(_ context.Context) ([]models.Voucher, error) {
	return nil, nil
}
func (m *mockVoucherRepo) FindByVoucherCode(_ context.Context, _ string, _ uuid.UUID) (*models.Voucher, error) {
	return nil, nil
}
func (m *mockVoucherRepo) FindDeletedByCode(_ context.Context, _ string, _ uuid.UUID) (*models.Voucher, error) {
	return nil, nil
}
func (m *mockVoucherRepo) RestoreByID(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
	return nil
}
func (m *mockVoucherRepo) Search(_ context.Context, _ uuid.UUID, _ string) ([]models.Voucher, error) {
	return nil, nil
}

var _ repository.VoucherRepository = (*mockVoucherRepo)(nil)

type mockGiftCardRepo struct{ mock.Mock }

func (m *mockGiftCardRepo) Create(ctx context.Context, gc *models.GiftCard) error {
	return m.Called(ctx, gc).Error(0)
}
func (m *mockGiftCardRepo) GetByID(ctx context.Context, id uuid.UUID, _ ...string) (*models.GiftCard, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GiftCard), args.Error(1)
}
func (m *mockGiftCardRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.GiftCard, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.GiftCard), args.Error(1)
}
func (m *mockGiftCardRepo) GetSharedWithUser(ctx context.Context, userID uuid.UUID) ([]models.GiftCard, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.GiftCard), args.Error(1)
}
func (m *mockGiftCardRepo) Update(ctx context.Context, gc *models.GiftCard) error {
	return m.Called(ctx, gc).Error(0)
}
func (m *mockGiftCardRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockGiftCardRepo) Count(ctx context.Context, userID uuid.UUID) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockGiftCardRepo) GetTotalBalance(ctx context.Context, userID uuid.UUID) (float64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(float64), args.Error(1)
}
func (m *mockGiftCardRepo) CreateTransaction(ctx context.Context, tx *models.GiftCardTransaction) error {
	return m.Called(ctx, tx).Error(0)
}
func (m *mockGiftCardRepo) GetTransaction(ctx context.Context, txID, gcID uuid.UUID) (*models.GiftCardTransaction, error) {
	args := m.Called(ctx, txID, gcID)
	return args.Get(0).(*models.GiftCardTransaction), args.Error(1)
}
func (m *mockGiftCardRepo) DeleteTransaction(ctx context.Context, txID uuid.UUID) error {
	return m.Called(ctx, txID).Error(0)
}
func (m *mockGiftCardRepo) GetAllForUserPaginated(_ context.Context, _ uuid.UUID, _ repository.PaginationParams) (*repository.PaginatedResult[models.GiftCard], error) {
	return nil, nil
}
func (m *mockGiftCardRepo) GetExpiringGiftCards(_ context.Context, _ int) ([]models.GiftCard, error) {
	return nil, nil
}
func (m *mockGiftCardRepo) FindByCardNumber(_ context.Context, _ string, _ uuid.UUID) (*models.GiftCard, error) {
	return nil, nil
}
func (m *mockGiftCardRepo) FindDeletedByCardNumber(_ context.Context, _ string, _ uuid.UUID) (*models.GiftCard, error) {
	return nil, nil
}
func (m *mockGiftCardRepo) RestoreByID(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
	return nil
}
func (m *mockGiftCardRepo) Search(_ context.Context, _ uuid.UUID, _ string) ([]models.GiftCard, error) {
	return nil, nil
}

var _ repository.GiftCardRepository = (*mockGiftCardRepo)(nil)

type mockCardShareRepo struct{ mock.Mock }

func (m *mockCardShareRepo) Create(ctx context.Context, s *models.CardShare) error {
	return m.Called(ctx, s).Error(0)
}
func (m *mockCardShareRepo) GetByCardAndUser(ctx context.Context, cardID, userID uuid.UUID) (*models.CardShare, error) {
	args := m.Called(ctx, cardID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CardShare), args.Error(1)
}
func (m *mockCardShareRepo) GetByCardID(ctx context.Context, cardID uuid.UUID) ([]models.CardShare, error) {
	args := m.Called(ctx, cardID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.CardShare), args.Error(1)
}
func (m *mockCardShareRepo) Update(ctx context.Context, s *models.CardShare) error {
	return m.Called(ctx, s).Error(0)
}
func (m *mockCardShareRepo) DeleteByCardAndUser(ctx context.Context, cardID, userID uuid.UUID) error {
	return m.Called(ctx, cardID, userID).Error(0)
}
func (m *mockCardShareRepo) DeleteByCardID(ctx context.Context, cardID uuid.UUID) error {
	return m.Called(ctx, cardID).Error(0)
}
func (m *mockCardShareRepo) CountByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockCardShareRepo) CountByCardIDs(ctx context.Context, cardIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
	args := m.Called(ctx, cardIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[uuid.UUID]int64), args.Error(1)
}
func (m *mockCardShareRepo) GetSharedUserIDs(ctx context.Context, ownerID uuid.UUID) ([]uuid.UUID, error) {
	args := m.Called(ctx, ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]uuid.UUID), args.Error(1)
}

var _ repository.CardShareRepository = (*mockCardShareRepo)(nil)

type mockVoucherShareRepo struct{ mock.Mock }

func (m *mockVoucherShareRepo) Create(ctx context.Context, s *models.VoucherShare) error {
	return m.Called(ctx, s).Error(0)
}
func (m *mockVoucherShareRepo) GetByVoucherAndUser(ctx context.Context, voucherID, userID uuid.UUID) (*models.VoucherShare, error) {
	args := m.Called(ctx, voucherID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.VoucherShare), args.Error(1)
}
func (m *mockVoucherShareRepo) GetByVoucherID(ctx context.Context, voucherID uuid.UUID) ([]models.VoucherShare, error) {
	args := m.Called(ctx, voucherID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.VoucherShare), args.Error(1)
}
func (m *mockVoucherShareRepo) Update(ctx context.Context, s *models.VoucherShare) error {
	return m.Called(ctx, s).Error(0)
}
func (m *mockVoucherShareRepo) DeleteByVoucherAndUser(ctx context.Context, voucherID, userID uuid.UUID) error {
	return m.Called(ctx, voucherID, userID).Error(0)
}
func (m *mockVoucherShareRepo) DeleteByVoucherID(ctx context.Context, voucherID uuid.UUID) error {
	return m.Called(ctx, voucherID).Error(0)
}
func (m *mockVoucherShareRepo) CountByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockVoucherShareRepo) CountByVoucherIDs(ctx context.Context, voucherIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
	args := m.Called(ctx, voucherIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[uuid.UUID]int64), args.Error(1)
}
func (m *mockVoucherShareRepo) GetSharedUserIDs(ctx context.Context, ownerID uuid.UUID) ([]uuid.UUID, error) {
	args := m.Called(ctx, ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]uuid.UUID), args.Error(1)
}

var _ repository.VoucherShareRepository = (*mockVoucherShareRepo)(nil)

type mockGiftCardShareRepo struct{ mock.Mock }

func (m *mockGiftCardShareRepo) Create(ctx context.Context, s *models.GiftCardShare) error {
	return m.Called(ctx, s).Error(0)
}
func (m *mockGiftCardShareRepo) GetByGiftCardAndUser(ctx context.Context, gcID, userID uuid.UUID) (*models.GiftCardShare, error) {
	args := m.Called(ctx, gcID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GiftCardShare), args.Error(1)
}
func (m *mockGiftCardShareRepo) GetByGiftCardID(ctx context.Context, gcID uuid.UUID) ([]models.GiftCardShare, error) {
	args := m.Called(ctx, gcID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.GiftCardShare), args.Error(1)
}
func (m *mockGiftCardShareRepo) Update(ctx context.Context, s *models.GiftCardShare) error {
	return m.Called(ctx, s).Error(0)
}
func (m *mockGiftCardShareRepo) DeleteByGiftCardAndUser(ctx context.Context, gcID, userID uuid.UUID) error {
	return m.Called(ctx, gcID, userID).Error(0)
}
func (m *mockGiftCardShareRepo) DeleteByGiftCardID(ctx context.Context, gcID uuid.UUID) error {
	return m.Called(ctx, gcID).Error(0)
}
func (m *mockGiftCardShareRepo) CountByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockGiftCardShareRepo) CountByGiftCardIDs(ctx context.Context, giftCardIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
	args := m.Called(ctx, giftCardIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[uuid.UUID]int64), args.Error(1)
}
func (m *mockGiftCardShareRepo) GetSharedUserIDs(ctx context.Context, ownerID uuid.UUID) ([]uuid.UUID, error) {
	args := m.Called(ctx, ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]uuid.UUID), args.Error(1)
}

var _ repository.GiftCardShareRepository = (*mockGiftCardShareRepo)(nil)

type mockUserRepo struct{ mock.Mock }

func (m *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}
func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}
func (m *mockUserRepo) Create(ctx context.Context, user *models.User) error {
	return m.Called(ctx, user).Error(0)
}
func (m *mockUserRepo) Update(ctx context.Context, user *models.User) error {
	return m.Called(ctx, user).Error(0)
}
func (m *mockUserRepo) GetAll(ctx context.Context) ([]models.User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.User), args.Error(1)
}
func (m *mockUserRepo) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]models.User, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]models.User), args.Error(1)
}
func (m *mockUserRepo) SearchByIDs(ctx context.Context, ids []uuid.UUID, query string) ([]models.User, error) {
	args := m.Called(ctx, ids, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.User), args.Error(1)
}

var _ repository.UserRepository = (*mockUserRepo)(nil)

type mockAuditLogRepo struct{ mock.Mock }

func (m *mockAuditLogRepo) Create(ctx context.Context, log *models.AuditLog) error {
	return m.Called(ctx, log).Error(0)
}
func (m *mockAuditLogRepo) GetFiltered(ctx context.Context, filters repository.AuditLogFilters) ([]models.AuditLog, int64, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).([]models.AuditLog), args.Get(1).(int64), args.Error(2)
}
func (m *mockAuditLogRepo) RestoreResource(ctx context.Context, resourceType string, resourceID uuid.UUID) error {
	return m.Called(ctx, resourceType, resourceID).Error(0)
}

var _ repository.AuditLogRepository = (*mockAuditLogRepo)(nil)

type mockNotifService struct{ mock.Mock }

func (m *mockNotifService) CreateShareNotification(ctx context.Context, recipientID, fromUserID uuid.UUID, fromUserName, resourceType string, resourceID uuid.UUID, permissions map[string]bool) error {
	return m.Called(ctx, recipientID, fromUserID, fromUserName, resourceType, resourceID, permissions).Error(0)
}
func (m *mockNotifService) CreateTransferNotification(ctx context.Context, recipientID, fromUserID uuid.UUID, fromUserName, resourceType string, resourceID uuid.UUID) error {
	return m.Called(ctx, recipientID, fromUserID, fromUserName, resourceType, resourceID).Error(0)
}
func (m *mockNotifService) GetUserNotifications(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Notification, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0).([]models.Notification), args.Error(1)
}
func (m *mockNotifService) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockNotifService) MarkAsRead(ctx context.Context, userID, notifID uuid.UUID) error {
	return m.Called(ctx, userID, notifID).Error(0)
}
func (m *mockNotifService) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	return m.Called(ctx, userID).Error(0)
}
func (m *mockNotifService) DeleteNotification(ctx context.Context, userID, notifID uuid.UUID) error {
	return m.Called(ctx, userID, notifID).Error(0)
}
func (m *mockNotifService) ArchiveOldRead(ctx context.Context, olderThanDays int) (int64, error) {
	args := m.Called(ctx, olderThanDays)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockNotifService) SetPushService(_ PushServiceInterface) {}
func (m *mockNotifService) SetEmailService(_ email.ServiceInterface, _ EmailTokenServiceInterface, _ string) {
}

var _ NotificationServiceInterface = (*mockNotifService)(nil)

// ==================== Test Setup ====================

type shareTestDeps struct {
	cardRepo         *mockCardRepo
	voucherRepo      *mockVoucherRepo
	giftCardRepo     *mockGiftCardRepo
	cardShareRepo    *mockCardShareRepo
	voucherShareRepo *mockVoucherShareRepo
	gcShareRepo      *mockGiftCardShareRepo
	userRepo         *mockUserRepo
	auditRepo        *mockAuditLogRepo
	notifService     *mockNotifService
	service          ShareServiceInterface
}

func setupShareService() *shareTestDeps {
	d := &shareTestDeps{
		cardRepo:         new(mockCardRepo),
		voucherRepo:      new(mockVoucherRepo),
		giftCardRepo:     new(mockGiftCardRepo),
		cardShareRepo:    new(mockCardShareRepo),
		voucherShareRepo: new(mockVoucherShareRepo),
		gcShareRepo:      new(mockGiftCardShareRepo),
		userRepo:         new(mockUserRepo),
		auditRepo:        new(mockAuditLogRepo),
		notifService:     new(mockNotifService),
	}
	d.service = NewShareService(
		nil, // db: nil uses mock repo fallback path for unit tests
		d.cardRepo, d.voucherRepo, d.giftCardRepo,
		d.cardShareRepo, d.voucherShareRepo, d.gcShareRepo,
		d.userRepo, d.auditRepo, d.notifService,
	)
	return d
}

func ptrUUID(id uuid.UUID) *uuid.UUID { return &id }

// ==================== CreateCardShare Tests ====================

func TestCreateCardShare_Success(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	ownerID := uuid.New()
	cardID := uuid.New()
	sharedWithID := uuid.New()

	card := &models.Card{ID: cardID, UserID: ptrUUID(ownerID)}
	owner := &models.User{ID: ownerID, FirstName: "Alice", LastName: "Owner"}

	d.cardRepo.On("GetByID", ctx, cardID).Return(card, nil)
	d.cardShareRepo.On("Create", ctx, mock.AnythingOfType("*models.CardShare")).Return(nil)
	d.userRepo.On("GetByID", ctx, ownerID).Return(owner, nil)
	d.notifService.On("CreateShareNotification", ctx, sharedWithID, ownerID, "Alice Owner", "card", cardID, map[string]bool{"can_edit": true, "can_delete": false}).Return(nil)

	err := d.service.CreateCardShare(ctx, ownerID, cardID, sharedWithID, true, false)
	assert.NoError(t, err)
	d.cardShareRepo.AssertCalled(t, "Create", ctx, mock.AnythingOfType("*models.CardShare"))
}

func TestCreateCardShare_NilCardID(t *testing.T) {
	d := setupShareService()
	err := d.service.CreateCardShare(context.Background(), uuid.New(), uuid.Nil, uuid.New(), false, false)
	assert.EqualError(t, err, "card ID is required")
}

func TestCreateCardShare_NilSharedWithID(t *testing.T) {
	d := setupShareService()
	err := d.service.CreateCardShare(context.Background(), uuid.New(), uuid.New(), uuid.Nil, false, false)
	assert.EqualError(t, err, "shared with user ID is required")
}

func TestCreateCardShare_CardNotFound(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	cardID := uuid.New()
	d.cardRepo.On("GetByID", ctx, cardID).Return(nil, gorm.ErrRecordNotFound)

	err := d.service.CreateCardShare(ctx, uuid.New(), cardID, uuid.New(), false, false)
	assert.EqualError(t, err, "card not found")
}

func TestCreateCardShare_RepoError(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	cardID := uuid.New()
	dbErr := errors.New("db connection failed")
	d.cardRepo.On("GetByID", ctx, cardID).Return(nil, dbErr)

	err := d.service.CreateCardShare(ctx, uuid.New(), cardID, uuid.New(), false, false)
	assert.ErrorIs(t, err, dbErr)
}

func TestCreateCardShare_NotOwner(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	ownerID := uuid.New()
	callerID := uuid.New()
	cardID := uuid.New()

	card := &models.Card{ID: cardID, UserID: ptrUUID(ownerID)}
	d.cardRepo.On("GetByID", ctx, cardID).Return(card, nil)

	err := d.service.CreateCardShare(ctx, callerID, cardID, uuid.New(), false, false)
	assert.ErrorIs(t, err, ErrNotOwner)
}

func TestCreateCardShare_NilOwner(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	cardID := uuid.New()

	card := &models.Card{ID: cardID, UserID: nil}
	d.cardRepo.On("GetByID", ctx, cardID).Return(card, nil)

	err := d.service.CreateCardShare(ctx, uuid.New(), cardID, uuid.New(), false, false)
	assert.ErrorIs(t, err, ErrNotOwner)
}

func TestCreateCardShare_CannotShareWithOwner(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	ownerID := uuid.New()
	cardID := uuid.New()

	card := &models.Card{ID: cardID, UserID: ptrUUID(ownerID)}
	d.cardRepo.On("GetByID", ctx, cardID).Return(card, nil)

	err := d.service.CreateCardShare(ctx, ownerID, cardID, ownerID, false, false)
	assert.EqualError(t, err, "cannot share card with its owner")
}

func TestCreateCardShare_ShareRepoError(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	ownerID := uuid.New()
	cardID := uuid.New()
	sharedWithID := uuid.New()

	card := &models.Card{ID: cardID, UserID: ptrUUID(ownerID)}
	d.cardRepo.On("GetByID", ctx, cardID).Return(card, nil)
	d.cardShareRepo.On("Create", ctx, mock.AnythingOfType("*models.CardShare")).Return(errors.New("duplicate key"))

	err := d.service.CreateCardShare(ctx, ownerID, cardID, sharedWithID, false, false)
	assert.EqualError(t, err, "duplicate key")
}

func TestCreateCardShare_NotificationFailureDoesNotBlock(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	ownerID := uuid.New()
	cardID := uuid.New()
	sharedWithID := uuid.New()

	card := &models.Card{ID: cardID, UserID: ptrUUID(ownerID)}
	owner := &models.User{ID: ownerID, FirstName: "Alice", LastName: "Owner"}

	d.cardRepo.On("GetByID", ctx, cardID).Return(card, nil)
	d.cardShareRepo.On("Create", ctx, mock.AnythingOfType("*models.CardShare")).Return(nil)
	d.userRepo.On("GetByID", ctx, ownerID).Return(owner, nil)
	d.notifService.On("CreateShareNotification", ctx, sharedWithID, ownerID, "Alice Owner", "card", cardID, mock.Anything).Return(errors.New("notif failed"))

	err := d.service.CreateCardShare(ctx, ownerID, cardID, sharedWithID, true, false)
	assert.NoError(t, err)
}

func TestCreateCardShare_OwnerUserNotFoundDoesNotBlock(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	ownerID := uuid.New()
	cardID := uuid.New()
	sharedWithID := uuid.New()

	card := &models.Card{ID: cardID, UserID: ptrUUID(ownerID)}

	d.cardRepo.On("GetByID", ctx, cardID).Return(card, nil)
	d.cardShareRepo.On("Create", ctx, mock.AnythingOfType("*models.CardShare")).Return(nil)
	d.userRepo.On("GetByID", ctx, ownerID).Return(nil, errors.New("user not found"))

	err := d.service.CreateCardShare(ctx, ownerID, cardID, sharedWithID, true, false)
	assert.NoError(t, err)
}

// ==================== CreateVoucherShare Tests ====================

func TestCreateVoucherShare_Success(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	ownerID := uuid.New()
	voucherID := uuid.New()
	sharedWithID := uuid.New()

	voucher := &models.Voucher{ID: voucherID, UserID: ptrUUID(ownerID)}
	owner := &models.User{ID: ownerID, FirstName: "Bob", LastName: "Doe"}

	d.voucherRepo.On("GetByID", ctx, voucherID).Return(voucher, nil)
	d.voucherShareRepo.On("Create", ctx, mock.AnythingOfType("*models.VoucherShare")).Return(nil)
	d.userRepo.On("GetByID", ctx, ownerID).Return(owner, nil)
	d.notifService.On("CreateShareNotification", ctx, sharedWithID, ownerID, "Bob Doe", "voucher", voucherID, mock.Anything).Return(nil)

	err := d.service.CreateVoucherShare(ctx, ownerID, voucherID, sharedWithID)
	assert.NoError(t, err)
}

func TestCreateVoucherShare_NilVoucherID(t *testing.T) {
	d := setupShareService()
	err := d.service.CreateVoucherShare(context.Background(), uuid.New(), uuid.Nil, uuid.New())
	assert.EqualError(t, err, "voucher ID is required")
}

func TestCreateVoucherShare_NilSharedWithID(t *testing.T) {
	d := setupShareService()
	err := d.service.CreateVoucherShare(context.Background(), uuid.New(), uuid.New(), uuid.Nil)
	assert.EqualError(t, err, "shared with user ID is required")
}

func TestCreateVoucherShare_NotFound(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	vID := uuid.New()
	d.voucherRepo.On("GetByID", ctx, vID).Return(nil, gorm.ErrRecordNotFound)

	err := d.service.CreateVoucherShare(ctx, uuid.New(), vID, uuid.New())
	assert.EqualError(t, err, "voucher not found")
}

func TestCreateVoucherShare_RepoError(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	vID := uuid.New()
	dbErr := errors.New("timeout")
	d.voucherRepo.On("GetByID", ctx, vID).Return(nil, dbErr)

	err := d.service.CreateVoucherShare(ctx, uuid.New(), vID, uuid.New())
	assert.ErrorIs(t, err, dbErr)
}

func TestCreateVoucherShare_NotOwner(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	ownerID := uuid.New()
	callerID := uuid.New()
	vID := uuid.New()

	voucher := &models.Voucher{ID: vID, UserID: ptrUUID(ownerID)}
	d.voucherRepo.On("GetByID", ctx, vID).Return(voucher, nil)

	err := d.service.CreateVoucherShare(ctx, callerID, vID, uuid.New())
	assert.ErrorIs(t, err, ErrNotOwner)
}

func TestCreateVoucherShare_NilOwner(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	vID := uuid.New()

	voucher := &models.Voucher{ID: vID, UserID: nil}
	d.voucherRepo.On("GetByID", ctx, vID).Return(voucher, nil)

	err := d.service.CreateVoucherShare(ctx, uuid.New(), vID, uuid.New())
	assert.ErrorIs(t, err, ErrNotOwner)
}

func TestCreateVoucherShare_CannotShareWithOwner(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	ownerID := uuid.New()
	vID := uuid.New()

	voucher := &models.Voucher{ID: vID, UserID: ptrUUID(ownerID)}
	d.voucherRepo.On("GetByID", ctx, vID).Return(voucher, nil)

	err := d.service.CreateVoucherShare(ctx, ownerID, vID, ownerID)
	assert.EqualError(t, err, "cannot share voucher with its owner")
}

func TestCreateVoucherShare_ShareRepoError(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	ownerID := uuid.New()
	vID := uuid.New()

	voucher := &models.Voucher{ID: vID, UserID: ptrUUID(ownerID)}
	d.voucherRepo.On("GetByID", ctx, vID).Return(voucher, nil)
	d.voucherShareRepo.On("Create", ctx, mock.AnythingOfType("*models.VoucherShare")).Return(errors.New("duplicate"))

	err := d.service.CreateVoucherShare(ctx, ownerID, vID, uuid.New())
	assert.EqualError(t, err, "duplicate")
}

func TestCreateVoucherShare_NotificationFailure(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	ownerID := uuid.New()
	vID := uuid.New()
	sharedWithID := uuid.New()

	voucher := &models.Voucher{ID: vID, UserID: ptrUUID(ownerID)}
	owner := &models.User{ID: ownerID, FirstName: "Bob", LastName: "D"}

	d.voucherRepo.On("GetByID", ctx, vID).Return(voucher, nil)
	d.voucherShareRepo.On("Create", ctx, mock.AnythingOfType("*models.VoucherShare")).Return(nil)
	d.userRepo.On("GetByID", ctx, ownerID).Return(owner, nil)
	d.notifService.On("CreateShareNotification", ctx, sharedWithID, ownerID, mock.Anything, "voucher", vID, mock.Anything).Return(errors.New("fail"))

	err := d.service.CreateVoucherShare(ctx, ownerID, vID, sharedWithID)
	assert.NoError(t, err)
}

// ==================== CreateGiftCardShare Tests ====================

func TestCreateGiftCardShare_Success(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	ownerID := uuid.New()
	gcID := uuid.New()
	sharedWithID := uuid.New()

	gc := &models.GiftCard{ID: gcID, UserID: ptrUUID(ownerID)}
	owner := &models.User{ID: ownerID, FirstName: "Carol", LastName: "Test"}

	d.giftCardRepo.On("GetByID", ctx, gcID).Return(gc, nil)
	d.gcShareRepo.On("Create", ctx, mock.AnythingOfType("*models.GiftCardShare")).Return(nil)
	d.userRepo.On("GetByID", ctx, ownerID).Return(owner, nil)
	d.notifService.On("CreateShareNotification", ctx, sharedWithID, ownerID, "Carol Test", "gift_card", gcID, mock.Anything).Return(nil)

	err := d.service.CreateGiftCardShare(ctx, ownerID, gcID, sharedWithID, true, true, true)
	assert.NoError(t, err)
}

func TestCreateGiftCardShare_NilGiftCardID(t *testing.T) {
	d := setupShareService()
	err := d.service.CreateGiftCardShare(context.Background(), uuid.New(), uuid.Nil, uuid.New(), false, false, false)
	assert.EqualError(t, err, "gift card ID is required")
}

func TestCreateGiftCardShare_NilSharedWithID(t *testing.T) {
	d := setupShareService()
	err := d.service.CreateGiftCardShare(context.Background(), uuid.New(), uuid.New(), uuid.Nil, false, false, false)
	assert.EqualError(t, err, "shared with user ID is required")
}

func TestCreateGiftCardShare_NotFound(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	gcID := uuid.New()
	d.giftCardRepo.On("GetByID", ctx, gcID).Return(nil, gorm.ErrRecordNotFound)

	err := d.service.CreateGiftCardShare(ctx, uuid.New(), gcID, uuid.New(), false, false, false)
	assert.EqualError(t, err, "gift card not found")
}

func TestCreateGiftCardShare_RepoError(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	gcID := uuid.New()
	dbErr := errors.New("db error")
	d.giftCardRepo.On("GetByID", ctx, gcID).Return(nil, dbErr)

	err := d.service.CreateGiftCardShare(ctx, uuid.New(), gcID, uuid.New(), false, false, false)
	assert.ErrorIs(t, err, dbErr)
}

func TestCreateGiftCardShare_NotOwner(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	ownerID := uuid.New()
	callerID := uuid.New()
	gcID := uuid.New()

	gc := &models.GiftCard{ID: gcID, UserID: ptrUUID(ownerID)}
	d.giftCardRepo.On("GetByID", ctx, gcID).Return(gc, nil)

	err := d.service.CreateGiftCardShare(ctx, callerID, gcID, uuid.New(), false, false, false)
	assert.ErrorIs(t, err, ErrNotOwner)
}

func TestCreateGiftCardShare_NilOwner(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	gcID := uuid.New()

	gc := &models.GiftCard{ID: gcID, UserID: nil}
	d.giftCardRepo.On("GetByID", ctx, gcID).Return(gc, nil)

	err := d.service.CreateGiftCardShare(ctx, uuid.New(), gcID, uuid.New(), false, false, false)
	assert.ErrorIs(t, err, ErrNotOwner)
}

func TestCreateGiftCardShare_CannotShareWithOwner(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	ownerID := uuid.New()
	gcID := uuid.New()

	gc := &models.GiftCard{ID: gcID, UserID: ptrUUID(ownerID)}
	d.giftCardRepo.On("GetByID", ctx, gcID).Return(gc, nil)

	err := d.service.CreateGiftCardShare(ctx, ownerID, gcID, ownerID, false, false, false)
	assert.EqualError(t, err, "cannot share gift card with its owner")
}

func TestCreateGiftCardShare_ShareRepoError(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	ownerID := uuid.New()
	gcID := uuid.New()

	gc := &models.GiftCard{ID: gcID, UserID: ptrUUID(ownerID)}
	d.giftCardRepo.On("GetByID", ctx, gcID).Return(gc, nil)
	d.gcShareRepo.On("Create", ctx, mock.AnythingOfType("*models.GiftCardShare")).Return(errors.New("dup"))

	err := d.service.CreateGiftCardShare(ctx, ownerID, gcID, uuid.New(), false, false, false)
	assert.EqualError(t, err, "dup")
}

func TestCreateGiftCardShare_OwnerUserNotFound(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	ownerID := uuid.New()
	gcID := uuid.New()

	gc := &models.GiftCard{ID: gcID, UserID: ptrUUID(ownerID)}
	d.giftCardRepo.On("GetByID", ctx, gcID).Return(gc, nil)
	d.gcShareRepo.On("Create", ctx, mock.AnythingOfType("*models.GiftCardShare")).Return(nil)
	d.userRepo.On("GetByID", ctx, ownerID).Return(nil, errors.New("not found"))

	err := d.service.CreateGiftCardShare(ctx, ownerID, gcID, uuid.New(), false, false, false)
	assert.NoError(t, err)
}

// ==================== GetShares Tests ====================

func TestGetCardShares_Success(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	cardID := uuid.New()
	shares := []models.CardShare{{ID: uuid.New(), CardID: cardID}}

	d.cardShareRepo.On("GetByCardID", ctx, cardID).Return(shares, nil)

	result, err := d.service.GetCardShares(ctx, cardID)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestGetVoucherShares_Success(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	vID := uuid.New()
	shares := []models.VoucherShare{{ID: uuid.New(), VoucherID: vID}}

	d.voucherShareRepo.On("GetByVoucherID", ctx, vID).Return(shares, nil)

	result, err := d.service.GetVoucherShares(ctx, vID)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestGetGiftCardShares_Success(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	gcID := uuid.New()
	shares := []models.GiftCardShare{{ID: uuid.New(), GiftCardID: gcID}}

	d.gcShareRepo.On("GetByGiftCardID", ctx, gcID).Return(shares, nil)

	result, err := d.service.GetGiftCardShares(ctx, gcID)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

// ==================== GetSharedUsers Tests ====================

func TestGetSharedUsers_Success(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	userID := uuid.New()
	user1 := uuid.New()
	user2 := uuid.New()

	d.cardShareRepo.On("GetSharedUserIDs", ctx, userID).Return([]uuid.UUID{user1}, nil)
	d.voucherShareRepo.On("GetSharedUserIDs", ctx, userID).Return([]uuid.UUID{user2}, nil)
	d.gcShareRepo.On("GetSharedUserIDs", ctx, userID).Return([]uuid.UUID{user1}, nil) // duplicate
	d.userRepo.On("SearchByIDs", ctx, mock.Anything, "query").Return([]models.User{
		{ID: user1}, {ID: user2},
	}, nil)

	result, err := d.service.GetSharedUsers(ctx, userID, "query")
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestGetSharedUsers_NoShares(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	userID := uuid.New()

	d.cardShareRepo.On("GetSharedUserIDs", ctx, userID).Return([]uuid.UUID(nil), nil)
	d.voucherShareRepo.On("GetSharedUserIDs", ctx, userID).Return([]uuid.UUID(nil), nil)
	d.gcShareRepo.On("GetSharedUserIDs", ctx, userID).Return([]uuid.UUID(nil), nil)

	result, err := d.service.GetSharedUsers(ctx, userID, "")
	assert.NoError(t, err)
	assert.Empty(t, result)
}

// ==================== DeleteCardShare Tests ====================

func TestDeleteCardShare_Success(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	ownerID := uuid.New()
	cardID := uuid.New()
	sharedWithID := uuid.New()

	card := &models.Card{ID: cardID, UserID: ptrUUID(ownerID)}
	d.cardRepo.On("GetByID", ctx, cardID).Return(card, nil)
	d.cardShareRepo.On("DeleteByCardAndUser", ctx, cardID, sharedWithID).Return(nil)

	err := d.service.DeleteCardShare(ctx, ownerID, cardID, sharedWithID)
	assert.NoError(t, err)
}

func TestDeleteCardShare_NotOwner(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	ownerID := uuid.New()
	callerID := uuid.New()
	cardID := uuid.New()

	card := &models.Card{ID: cardID, UserID: ptrUUID(ownerID)}
	d.cardRepo.On("GetByID", ctx, cardID).Return(card, nil)

	err := d.service.DeleteCardShare(ctx, callerID, cardID, uuid.New())
	assert.ErrorIs(t, err, ErrNotOwner)
}

func TestDeleteCardShare_NilOwner(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	cardID := uuid.New()

	card := &models.Card{ID: cardID, UserID: nil}
	d.cardRepo.On("GetByID", ctx, cardID).Return(card, nil)

	err := d.service.DeleteCardShare(ctx, uuid.New(), cardID, uuid.New())
	assert.ErrorIs(t, err, ErrNotOwner)
}

func TestDeleteCardShare_NotFound(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	cardID := uuid.New()
	d.cardRepo.On("GetByID", ctx, cardID).Return(nil, gorm.ErrRecordNotFound)

	err := d.service.DeleteCardShare(ctx, uuid.New(), cardID, uuid.New())
	assert.EqualError(t, err, "card not found")
}

func TestDeleteCardShare_RepoError(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	cardID := uuid.New()
	dbErr := errors.New("db error")
	d.cardRepo.On("GetByID", ctx, cardID).Return(nil, dbErr)

	err := d.service.DeleteCardShare(ctx, uuid.New(), cardID, uuid.New())
	assert.ErrorIs(t, err, dbErr)
}

// ==================== DeleteVoucherShare Tests ====================

func TestDeleteVoucherShare_Success(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	ownerID := uuid.New()
	vID := uuid.New()
	sharedWithID := uuid.New()

	voucher := &models.Voucher{ID: vID, UserID: ptrUUID(ownerID)}
	d.voucherRepo.On("GetByID", ctx, vID).Return(voucher, nil)
	d.voucherShareRepo.On("DeleteByVoucherAndUser", ctx, vID, sharedWithID).Return(nil)

	err := d.service.DeleteVoucherShare(ctx, ownerID, vID, sharedWithID)
	assert.NoError(t, err)
}

func TestDeleteVoucherShare_NotOwner(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	ownerID := uuid.New()
	vID := uuid.New()

	voucher := &models.Voucher{ID: vID, UserID: ptrUUID(ownerID)}
	d.voucherRepo.On("GetByID", ctx, vID).Return(voucher, nil)

	err := d.service.DeleteVoucherShare(ctx, uuid.New(), vID, uuid.New())
	assert.ErrorIs(t, err, ErrNotOwner)
}

func TestDeleteVoucherShare_NotFound(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	vID := uuid.New()
	d.voucherRepo.On("GetByID", ctx, vID).Return(nil, gorm.ErrRecordNotFound)

	err := d.service.DeleteVoucherShare(ctx, uuid.New(), vID, uuid.New())
	assert.EqualError(t, err, "voucher not found")
}

func TestDeleteVoucherShare_RepoError(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	vID := uuid.New()
	dbErr := errors.New("db err")
	d.voucherRepo.On("GetByID", ctx, vID).Return(nil, dbErr)

	err := d.service.DeleteVoucherShare(ctx, uuid.New(), vID, uuid.New())
	assert.ErrorIs(t, err, dbErr)
}

// ==================== DeleteGiftCardShare Tests ====================

func TestDeleteGiftCardShare_Success(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	ownerID := uuid.New()
	gcID := uuid.New()
	sharedWithID := uuid.New()

	gc := &models.GiftCard{ID: gcID, UserID: ptrUUID(ownerID)}
	d.giftCardRepo.On("GetByID", ctx, gcID).Return(gc, nil)
	d.gcShareRepo.On("DeleteByGiftCardAndUser", ctx, gcID, sharedWithID).Return(nil)

	err := d.service.DeleteGiftCardShare(ctx, ownerID, gcID, sharedWithID)
	assert.NoError(t, err)
}

func TestDeleteGiftCardShare_NotOwner(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	ownerID := uuid.New()
	gcID := uuid.New()

	gc := &models.GiftCard{ID: gcID, UserID: ptrUUID(ownerID)}
	d.giftCardRepo.On("GetByID", ctx, gcID).Return(gc, nil)

	err := d.service.DeleteGiftCardShare(ctx, uuid.New(), gcID, uuid.New())
	assert.ErrorIs(t, err, ErrNotOwner)
}

func TestDeleteGiftCardShare_NotFound(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	gcID := uuid.New()
	d.giftCardRepo.On("GetByID", ctx, gcID).Return(nil, gorm.ErrRecordNotFound)

	err := d.service.DeleteGiftCardShare(ctx, uuid.New(), gcID, uuid.New())
	assert.EqualError(t, err, "gift card not found")
}

func TestDeleteGiftCardShare_RepoError(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	gcID := uuid.New()
	dbErr := errors.New("db err")
	d.giftCardRepo.On("GetByID", ctx, gcID).Return(nil, dbErr)

	err := d.service.DeleteGiftCardShare(ctx, uuid.New(), gcID, uuid.New())
	assert.ErrorIs(t, err, dbErr)
}

// ==================== UpdateCardShare Tests ====================

func TestUpdateCardShare_Success(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	ownerID := uuid.New()
	cardID := uuid.New()
	sharedWithID := uuid.New()
	shareID := uuid.New()

	share := &models.CardShare{ID: shareID, CardID: cardID, SharedWithID: sharedWithID}
	card := &models.Card{ID: cardID, UserID: ptrUUID(ownerID)}

	d.cardShareRepo.On("GetByCardAndUser", ctx, cardID, sharedWithID).Return(share, nil)
	d.cardRepo.On("GetByID", ctx, cardID).Return(card, nil)
	d.auditRepo.On("Create", ctx, mock.AnythingOfType("*models.AuditLog")).Return(nil)
	d.cardShareRepo.On("Update", ctx, mock.AnythingOfType("*models.CardShare")).Return(nil)

	err := d.service.UpdateCardShare(ctx, ownerID, cardID, sharedWithID, true, true)
	assert.NoError(t, err)
}

func TestUpdateCardShare_NilCardID(t *testing.T) {
	d := setupShareService()
	err := d.service.UpdateCardShare(context.Background(), uuid.New(), uuid.Nil, uuid.New(), false, false)
	assert.EqualError(t, err, "card ID is required")
}

func TestUpdateCardShare_NilSharedWithID(t *testing.T) {
	d := setupShareService()
	err := d.service.UpdateCardShare(context.Background(), uuid.New(), uuid.New(), uuid.Nil, false, false)
	assert.EqualError(t, err, "shared with user ID is required")
}

func TestUpdateCardShare_ShareNotFound(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	cardID := uuid.New()
	sharedWithID := uuid.New()

	d.cardShareRepo.On("GetByCardAndUser", ctx, cardID, sharedWithID).Return(nil, gorm.ErrRecordNotFound)

	err := d.service.UpdateCardShare(ctx, uuid.New(), cardID, sharedWithID, false, false)
	assert.EqualError(t, err, "share not found")
}

func TestUpdateCardShare_ShareRepoError(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	cardID := uuid.New()
	sharedWithID := uuid.New()
	dbErr := errors.New("db err")

	d.cardShareRepo.On("GetByCardAndUser", ctx, cardID, sharedWithID).Return(nil, dbErr)

	err := d.service.UpdateCardShare(ctx, uuid.New(), cardID, sharedWithID, false, false)
	assert.ErrorIs(t, err, dbErr)
}

func TestUpdateCardShare_CardRepoError(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	cardID := uuid.New()
	sharedWithID := uuid.New()
	dbErr := errors.New("card repo err")

	share := &models.CardShare{ID: uuid.New(), CardID: cardID, SharedWithID: sharedWithID}
	d.cardShareRepo.On("GetByCardAndUser", ctx, cardID, sharedWithID).Return(share, nil)
	d.cardRepo.On("GetByID", ctx, cardID).Return(nil, dbErr)

	err := d.service.UpdateCardShare(ctx, uuid.New(), cardID, sharedWithID, false, false)
	assert.ErrorIs(t, err, dbErr)
}

func TestUpdateCardShare_NotOwner(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	ownerID := uuid.New()
	callerID := uuid.New()
	cardID := uuid.New()
	sharedWithID := uuid.New()

	share := &models.CardShare{ID: uuid.New(), CardID: cardID, SharedWithID: sharedWithID}
	card := &models.Card{ID: cardID, UserID: ptrUUID(ownerID)}

	d.cardShareRepo.On("GetByCardAndUser", ctx, cardID, sharedWithID).Return(share, nil)
	d.cardRepo.On("GetByID", ctx, cardID).Return(card, nil)

	err := d.service.UpdateCardShare(ctx, callerID, cardID, sharedWithID, false, false)
	assert.ErrorIs(t, err, ErrNotOwner)
}

func TestUpdateCardShare_NilOwner(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	cardID := uuid.New()
	sharedWithID := uuid.New()

	share := &models.CardShare{ID: uuid.New(), CardID: cardID, SharedWithID: sharedWithID}
	card := &models.Card{ID: cardID, UserID: nil}

	d.cardShareRepo.On("GetByCardAndUser", ctx, cardID, sharedWithID).Return(share, nil)
	d.cardRepo.On("GetByID", ctx, cardID).Return(card, nil)

	err := d.service.UpdateCardShare(ctx, uuid.New(), cardID, sharedWithID, false, false)
	assert.ErrorIs(t, err, ErrNotOwner)
}

func TestUpdateCardShare_AuditLogFailureDoesNotBlock(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	ownerID := uuid.New()
	cardID := uuid.New()
	sharedWithID := uuid.New()

	share := &models.CardShare{ID: uuid.New(), CardID: cardID, SharedWithID: sharedWithID}
	card := &models.Card{ID: cardID, UserID: ptrUUID(ownerID)}

	d.cardShareRepo.On("GetByCardAndUser", ctx, cardID, sharedWithID).Return(share, nil)
	d.cardRepo.On("GetByID", ctx, cardID).Return(card, nil)
	d.auditRepo.On("Create", ctx, mock.AnythingOfType("*models.AuditLog")).Return(errors.New("audit fail"))
	d.cardShareRepo.On("Update", ctx, mock.AnythingOfType("*models.CardShare")).Return(nil)

	err := d.service.UpdateCardShare(ctx, ownerID, cardID, sharedWithID, true, false)
	assert.NoError(t, err)
}

func TestUpdateCardShare_UpdateRepoError(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	ownerID := uuid.New()
	cardID := uuid.New()
	sharedWithID := uuid.New()

	share := &models.CardShare{ID: uuid.New(), CardID: cardID, SharedWithID: sharedWithID}
	card := &models.Card{ID: cardID, UserID: ptrUUID(ownerID)}

	d.cardShareRepo.On("GetByCardAndUser", ctx, cardID, sharedWithID).Return(share, nil)
	d.cardRepo.On("GetByID", ctx, cardID).Return(card, nil)
	d.auditRepo.On("Create", ctx, mock.AnythingOfType("*models.AuditLog")).Return(nil)
	d.cardShareRepo.On("Update", ctx, mock.AnythingOfType("*models.CardShare")).Return(errors.New("update fail"))

	err := d.service.UpdateCardShare(ctx, ownerID, cardID, sharedWithID, true, false)
	assert.EqualError(t, err, "update fail")
}

// ==================== UpdateGiftCardShare Tests ====================

func TestUpdateGiftCardShare_Success(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	ownerID := uuid.New()
	gcID := uuid.New()
	sharedWithID := uuid.New()
	shareID := uuid.New()

	share := &models.GiftCardShare{ID: shareID, GiftCardID: gcID, SharedWithID: sharedWithID}
	gc := &models.GiftCard{ID: gcID, UserID: ptrUUID(ownerID)}

	d.gcShareRepo.On("GetByGiftCardAndUser", ctx, gcID, sharedWithID).Return(share, nil)
	d.giftCardRepo.On("GetByID", ctx, gcID).Return(gc, nil)
	d.auditRepo.On("Create", ctx, mock.AnythingOfType("*models.AuditLog")).Return(nil)
	d.gcShareRepo.On("Update", ctx, mock.AnythingOfType("*models.GiftCardShare")).Return(nil)

	err := d.service.UpdateGiftCardShare(ctx, ownerID, gcID, sharedWithID, true, false, true)
	assert.NoError(t, err)
}

func TestUpdateGiftCardShare_NilGiftCardID(t *testing.T) {
	d := setupShareService()
	err := d.service.UpdateGiftCardShare(context.Background(), uuid.New(), uuid.Nil, uuid.New(), false, false, false)
	assert.EqualError(t, err, "gift card ID is required")
}

func TestUpdateGiftCardShare_NilSharedWithID(t *testing.T) {
	d := setupShareService()
	err := d.service.UpdateGiftCardShare(context.Background(), uuid.New(), uuid.New(), uuid.Nil, false, false, false)
	assert.EqualError(t, err, "shared with user ID is required")
}

func TestUpdateGiftCardShare_ShareNotFound(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	gcID := uuid.New()
	sharedWithID := uuid.New()

	d.gcShareRepo.On("GetByGiftCardAndUser", ctx, gcID, sharedWithID).Return(nil, gorm.ErrRecordNotFound)

	err := d.service.UpdateGiftCardShare(ctx, uuid.New(), gcID, sharedWithID, false, false, false)
	assert.EqualError(t, err, "share not found")
}

func TestUpdateGiftCardShare_ShareRepoError(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	gcID := uuid.New()
	sharedWithID := uuid.New()
	dbErr := errors.New("db err")

	d.gcShareRepo.On("GetByGiftCardAndUser", ctx, gcID, sharedWithID).Return(nil, dbErr)

	err := d.service.UpdateGiftCardShare(ctx, uuid.New(), gcID, sharedWithID, false, false, false)
	assert.ErrorIs(t, err, dbErr)
}

func TestUpdateGiftCardShare_GiftCardRepoError(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	gcID := uuid.New()
	sharedWithID := uuid.New()
	dbErr := errors.New("gc repo err")

	share := &models.GiftCardShare{ID: uuid.New(), GiftCardID: gcID, SharedWithID: sharedWithID}
	d.gcShareRepo.On("GetByGiftCardAndUser", ctx, gcID, sharedWithID).Return(share, nil)
	d.giftCardRepo.On("GetByID", ctx, gcID).Return(nil, dbErr)

	err := d.service.UpdateGiftCardShare(ctx, uuid.New(), gcID, sharedWithID, false, false, false)
	assert.ErrorIs(t, err, dbErr)
}

func TestUpdateGiftCardShare_NotOwner(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	ownerID := uuid.New()
	callerID := uuid.New()
	gcID := uuid.New()
	sharedWithID := uuid.New()

	share := &models.GiftCardShare{ID: uuid.New(), GiftCardID: gcID, SharedWithID: sharedWithID}
	gc := &models.GiftCard{ID: gcID, UserID: ptrUUID(ownerID)}

	d.gcShareRepo.On("GetByGiftCardAndUser", ctx, gcID, sharedWithID).Return(share, nil)
	d.giftCardRepo.On("GetByID", ctx, gcID).Return(gc, nil)

	err := d.service.UpdateGiftCardShare(ctx, callerID, gcID, sharedWithID, false, false, false)
	assert.ErrorIs(t, err, ErrNotOwner)
}

func TestUpdateGiftCardShare_NilOwner(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	gcID := uuid.New()
	sharedWithID := uuid.New()

	share := &models.GiftCardShare{ID: uuid.New(), GiftCardID: gcID, SharedWithID: sharedWithID}
	gc := &models.GiftCard{ID: gcID, UserID: nil}

	d.gcShareRepo.On("GetByGiftCardAndUser", ctx, gcID, sharedWithID).Return(share, nil)
	d.giftCardRepo.On("GetByID", ctx, gcID).Return(gc, nil)

	err := d.service.UpdateGiftCardShare(ctx, uuid.New(), gcID, sharedWithID, false, false, false)
	assert.ErrorIs(t, err, ErrNotOwner)
}

func TestUpdateGiftCardShare_AuditLogFailureDoesNotBlock(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	ownerID := uuid.New()
	gcID := uuid.New()
	sharedWithID := uuid.New()

	share := &models.GiftCardShare{ID: uuid.New(), GiftCardID: gcID, SharedWithID: sharedWithID}
	gc := &models.GiftCard{ID: gcID, UserID: ptrUUID(ownerID)}

	d.gcShareRepo.On("GetByGiftCardAndUser", ctx, gcID, sharedWithID).Return(share, nil)
	d.giftCardRepo.On("GetByID", ctx, gcID).Return(gc, nil)
	d.auditRepo.On("Create", ctx, mock.AnythingOfType("*models.AuditLog")).Return(errors.New("audit fail"))
	d.gcShareRepo.On("Update", ctx, mock.AnythingOfType("*models.GiftCardShare")).Return(nil)

	err := d.service.UpdateGiftCardShare(ctx, ownerID, gcID, sharedWithID, true, false, true)
	assert.NoError(t, err)
}

func TestUpdateGiftCardShare_UpdateRepoError(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	ownerID := uuid.New()
	gcID := uuid.New()
	sharedWithID := uuid.New()

	share := &models.GiftCardShare{ID: uuid.New(), GiftCardID: gcID, SharedWithID: sharedWithID}
	gc := &models.GiftCard{ID: gcID, UserID: ptrUUID(ownerID)}

	d.gcShareRepo.On("GetByGiftCardAndUser", ctx, gcID, sharedWithID).Return(share, nil)
	d.giftCardRepo.On("GetByID", ctx, gcID).Return(gc, nil)
	d.auditRepo.On("Create", ctx, mock.AnythingOfType("*models.AuditLog")).Return(nil)
	d.gcShareRepo.On("Update", ctx, mock.AnythingOfType("*models.GiftCardShare")).Return(errors.New("update fail"))

	err := d.service.UpdateGiftCardShare(ctx, ownerID, gcID, sharedWithID, true, false, true)
	assert.EqualError(t, err, "update fail")
}

// ==================== GetCardShareCounts Tests ====================

func TestShareService_GetCardShareCounts_Success(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()

	id1 := uuid.New()
	id2 := uuid.New()
	cardIDs := []uuid.UUID{id1, id2}

	expected := map[uuid.UUID]int64{id1: 3, id2: 1}
	d.cardShareRepo.On("CountByCardIDs", ctx, cardIDs).Return(expected, nil)

	result, err := d.service.GetCardShareCounts(ctx, cardIDs)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	d.cardShareRepo.AssertCalled(t, "CountByCardIDs", ctx, cardIDs)
}

func TestShareService_GetCardShareCounts_Error(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()

	cardIDs := []uuid.UUID{uuid.New()}
	dbErr := errors.New("db error")
	d.cardShareRepo.On("CountByCardIDs", ctx, cardIDs).Return(nil, dbErr)

	result, err := d.service.GetCardShareCounts(ctx, cardIDs)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, dbErr)
}

// ==================== GetVoucherShareCounts Tests ====================

func TestShareService_GetVoucherShareCounts_Success(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()

	id1 := uuid.New()
	id2 := uuid.New()
	voucherIDs := []uuid.UUID{id1, id2}

	expected := map[uuid.UUID]int64{id1: 2, id2: 5}
	d.voucherShareRepo.On("CountByVoucherIDs", ctx, voucherIDs).Return(expected, nil)

	result, err := d.service.GetVoucherShareCounts(ctx, voucherIDs)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	d.voucherShareRepo.AssertCalled(t, "CountByVoucherIDs", ctx, voucherIDs)
}

func TestShareService_GetVoucherShareCounts_Error(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()

	voucherIDs := []uuid.UUID{uuid.New()}
	dbErr := errors.New("connection refused")
	d.voucherShareRepo.On("CountByVoucherIDs", ctx, voucherIDs).Return(nil, dbErr)

	result, err := d.service.GetVoucherShareCounts(ctx, voucherIDs)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, dbErr)
}

// ==================== GetGiftCardShareCounts Tests ====================

func TestShareService_GetGiftCardShareCounts_Success(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()

	id1 := uuid.New()
	id2 := uuid.New()
	gcIDs := []uuid.UUID{id1, id2}

	expected := map[uuid.UUID]int64{id1: 0, id2: 4}
	d.gcShareRepo.On("CountByGiftCardIDs", ctx, gcIDs).Return(expected, nil)

	result, err := d.service.GetGiftCardShareCounts(ctx, gcIDs)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	d.gcShareRepo.AssertCalled(t, "CountByGiftCardIDs", ctx, gcIDs)
}

func TestShareService_GetGiftCardShareCounts_Error(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()

	gcIDs := []uuid.UUID{uuid.New()}
	dbErr := errors.New("timeout")
	d.gcShareRepo.On("CountByGiftCardIDs", ctx, gcIDs).Return(nil, dbErr)

	result, err := d.service.GetGiftCardShareCounts(ctx, gcIDs)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, dbErr)
}

// ==================== isDuplicateShareError Tests ====================

func TestIsDuplicateShareError_True(t *testing.T) {
	err := errors.New("ERROR: duplicate key value violates unique constraint \"unique_active_card_share\" (SQLSTATE 23505)")
	assert.True(t, isDuplicateShareError(err))
}

func TestIsDuplicateShareError_TrueMinimalMatch(t *testing.T) {
	err := errors.New("unique_active constraint violated")
	assert.True(t, isDuplicateShareError(err))
}

func TestIsDuplicateShareError_FalseNonDuplicate(t *testing.T) {
	err := errors.New("connection refused")
	assert.False(t, isDuplicateShareError(err))
}

func TestIsDuplicateShareError_FalseNil(t *testing.T) {
	assert.False(t, isDuplicateShareError(nil))
}

// ==================== CreateCardShare - Duplicate Share ====================

func TestCreateCardShare_DuplicateShare(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	ownerID := uuid.New()
	cardID := uuid.New()
	sharedWithID := uuid.New()

	card := &models.Card{ID: cardID, UserID: ptrUUID(ownerID)}
	d.cardRepo.On("GetByID", ctx, cardID).Return(card, nil)
	d.cardShareRepo.On("Create", ctx, mock.AnythingOfType("*models.CardShare")).
		Return(errors.New("unique_active constraint violated"))

	err := d.service.CreateCardShare(ctx, ownerID, cardID, sharedWithID, false, false)
	assert.ErrorIs(t, err, ErrAlreadyShared)
}

// ==================== CreateVoucherShare - Duplicate Share ====================

func TestCreateVoucherShare_DuplicateShare(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	ownerID := uuid.New()
	vID := uuid.New()
	sharedWithID := uuid.New()

	voucher := &models.Voucher{ID: vID, UserID: ptrUUID(ownerID)}
	d.voucherRepo.On("GetByID", ctx, vID).Return(voucher, nil)
	d.voucherShareRepo.On("Create", ctx, mock.AnythingOfType("*models.VoucherShare")).
		Return(errors.New("unique_active constraint violated"))

	err := d.service.CreateVoucherShare(ctx, ownerID, vID, sharedWithID)
	assert.ErrorIs(t, err, ErrAlreadyShared)
}

// ==================== CreateGiftCardShare - Duplicate Share ====================

func TestCreateGiftCardShare_DuplicateShare(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	ownerID := uuid.New()
	gcID := uuid.New()
	sharedWithID := uuid.New()

	gc := &models.GiftCard{ID: gcID, UserID: ptrUUID(ownerID)}
	d.giftCardRepo.On("GetByID", ctx, gcID).Return(gc, nil)
	d.gcShareRepo.On("Create", ctx, mock.AnythingOfType("*models.GiftCardShare")).
		Return(errors.New("unique_active constraint violated"))

	err := d.service.CreateGiftCardShare(ctx, ownerID, gcID, sharedWithID, false, false, false)
	assert.ErrorIs(t, err, ErrAlreadyShared)
}

// ==================== GetSharedUsers - Repo Errors ====================

func TestGetSharedUsers_RepoErrorsSilentlyIgnored(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	userID := uuid.New()
	user1 := uuid.New()

	// Card share repo returns error, voucher and gift card return users
	d.cardShareRepo.On("GetSharedUserIDs", ctx, userID).Return(nil, errors.New("card repo error"))
	d.voucherShareRepo.On("GetSharedUserIDs", ctx, userID).Return([]uuid.UUID{user1}, nil)
	d.gcShareRepo.On("GetSharedUserIDs", ctx, userID).Return([]uuid.UUID(nil), nil)
	d.userRepo.On("SearchByIDs", ctx, mock.Anything, "").Return([]models.User{
		{ID: user1},
	}, nil)

	result, err := d.service.GetSharedUsers(ctx, userID, "")
	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestGetSharedUsers_AllRepoErrors(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	userID := uuid.New()

	// All share repos return errors - results in empty ID list
	d.cardShareRepo.On("GetSharedUserIDs", ctx, userID).Return(nil, errors.New("err1"))
	d.voucherShareRepo.On("GetSharedUserIDs", ctx, userID).Return(nil, errors.New("err2"))
	d.gcShareRepo.On("GetSharedUserIDs", ctx, userID).Return(nil, errors.New("err3"))

	result, err := d.service.GetSharedUsers(ctx, userID, "")
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestGetSharedUsers_SearchByIDsError(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	userID := uuid.New()
	user1 := uuid.New()

	d.cardShareRepo.On("GetSharedUserIDs", ctx, userID).Return([]uuid.UUID{user1}, nil)
	d.voucherShareRepo.On("GetSharedUserIDs", ctx, userID).Return([]uuid.UUID(nil), nil)
	d.gcShareRepo.On("GetSharedUserIDs", ctx, userID).Return([]uuid.UUID(nil), nil)
	d.userRepo.On("SearchByIDs", ctx, mock.Anything, "query").Return(nil, errors.New("search error"))

	result, err := d.service.GetSharedUsers(ctx, userID, "query")
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestGetSharedUsers_DeduplicatesUserIDs(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	userID := uuid.New()
	sharedUser := uuid.New()

	// Same user appears in all three repos
	d.cardShareRepo.On("GetSharedUserIDs", ctx, userID).Return([]uuid.UUID{sharedUser}, nil)
	d.voucherShareRepo.On("GetSharedUserIDs", ctx, userID).Return([]uuid.UUID{sharedUser}, nil)
	d.gcShareRepo.On("GetSharedUserIDs", ctx, userID).Return([]uuid.UUID{sharedUser}, nil)
	d.userRepo.On("SearchByIDs", ctx, []uuid.UUID{sharedUser}, "").Return([]models.User{
		{ID: sharedUser},
	}, nil)

	result, err := d.service.GetSharedUsers(ctx, userID, "")
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	// Verify SearchByIDs was called with exactly 1 ID (deduplicated)
	d.userRepo.AssertCalled(t, "SearchByIDs", ctx, []uuid.UUID{sharedUser}, "")
}

// ==================== DeleteCardShare - Delete Repo Error ====================

func TestDeleteCardShare_DeleteRepoError(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	ownerID := uuid.New()
	cardID := uuid.New()
	sharedWithID := uuid.New()

	card := &models.Card{ID: cardID, UserID: ptrUUID(ownerID)}
	d.cardRepo.On("GetByID", ctx, cardID).Return(card, nil)
	d.cardShareRepo.On("DeleteByCardAndUser", ctx, cardID, sharedWithID).Return(errors.New("delete failed"))

	err := d.service.DeleteCardShare(ctx, ownerID, cardID, sharedWithID)
	assert.EqualError(t, err, "delete failed")
}

// ==================== DeleteVoucherShare - Additional Tests ====================

func TestDeleteVoucherShare_NilOwner(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	vID := uuid.New()

	voucher := &models.Voucher{ID: vID, UserID: nil}
	d.voucherRepo.On("GetByID", ctx, vID).Return(voucher, nil)

	err := d.service.DeleteVoucherShare(ctx, uuid.New(), vID, uuid.New())
	assert.ErrorIs(t, err, ErrNotOwner)
}

func TestDeleteVoucherShare_DeleteRepoError(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	ownerID := uuid.New()
	vID := uuid.New()
	sharedWithID := uuid.New()

	voucher := &models.Voucher{ID: vID, UserID: ptrUUID(ownerID)}
	d.voucherRepo.On("GetByID", ctx, vID).Return(voucher, nil)
	d.voucherShareRepo.On("DeleteByVoucherAndUser", ctx, vID, sharedWithID).Return(errors.New("delete failed"))

	err := d.service.DeleteVoucherShare(ctx, ownerID, vID, sharedWithID)
	assert.EqualError(t, err, "delete failed")
}

// ==================== DeleteGiftCardShare - Additional Tests ====================

func TestDeleteGiftCardShare_NilOwner(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	gcID := uuid.New()

	gc := &models.GiftCard{ID: gcID, UserID: nil}
	d.giftCardRepo.On("GetByID", ctx, gcID).Return(gc, nil)

	err := d.service.DeleteGiftCardShare(ctx, uuid.New(), gcID, uuid.New())
	assert.ErrorIs(t, err, ErrNotOwner)
}

func TestDeleteGiftCardShare_DeleteRepoError(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	ownerID := uuid.New()
	gcID := uuid.New()
	sharedWithID := uuid.New()

	gc := &models.GiftCard{ID: gcID, UserID: ptrUUID(ownerID)}
	d.giftCardRepo.On("GetByID", ctx, gcID).Return(gc, nil)
	d.gcShareRepo.On("DeleteByGiftCardAndUser", ctx, gcID, sharedWithID).Return(errors.New("delete failed"))

	err := d.service.DeleteGiftCardShare(ctx, ownerID, gcID, sharedWithID)
	assert.EqualError(t, err, "delete failed")
}

// ==================== GetCardShares / GetVoucherShares / GetGiftCardShares - Error Paths ====================

func TestGetCardShares_Error(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	cardID := uuid.New()

	d.cardShareRepo.On("GetByCardID", ctx, cardID).Return(nil, errors.New("db error"))

	result, err := d.service.GetCardShares(ctx, cardID)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestGetVoucherShares_Error(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	vID := uuid.New()

	d.voucherShareRepo.On("GetByVoucherID", ctx, vID).Return(nil, errors.New("db error"))

	result, err := d.service.GetVoucherShares(ctx, vID)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestGetGiftCardShares_Error(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	gcID := uuid.New()

	d.gcShareRepo.On("GetByGiftCardID", ctx, gcID).Return(nil, errors.New("db error"))

	result, err := d.service.GetGiftCardShares(ctx, gcID)
	assert.Error(t, err)
	assert.Nil(t, result)
}

// ==================== DeleteAllShares (bulk revoke) Tests ====================

func TestDeleteAllCardShares_Success(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	ownerID := uuid.New()
	cardID := uuid.New()

	card := &models.Card{ID: cardID, UserID: ptrUUID(ownerID)}
	d.cardRepo.On("GetByID", ctx, cardID).Return(card, nil)
	d.cardShareRepo.On("DeleteByCardID", ctx, cardID).Return(nil)

	err := d.service.DeleteAllCardShares(ctx, ownerID, cardID)
	assert.NoError(t, err)
	d.cardShareRepo.AssertExpectations(t)
}

func TestDeleteAllCardShares_NotOwner(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	ownerID := uuid.New()
	cardID := uuid.New()

	card := &models.Card{ID: cardID, UserID: ptrUUID(ownerID)}
	d.cardRepo.On("GetByID", ctx, cardID).Return(card, nil)

	err := d.service.DeleteAllCardShares(ctx, uuid.New(), cardID)
	assert.ErrorIs(t, err, ErrNotOwner)
	d.cardShareRepo.AssertNotCalled(t, "DeleteByCardID", ctx, cardID)
}

func TestDeleteAllCardShares_NotFound(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	cardID := uuid.New()
	d.cardRepo.On("GetByID", ctx, cardID).Return(nil, gorm.ErrRecordNotFound)

	err := d.service.DeleteAllCardShares(ctx, uuid.New(), cardID)
	assert.EqualError(t, err, "card not found")
}

func TestDeleteAllVoucherShares_Success(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	ownerID := uuid.New()
	vID := uuid.New()

	voucher := &models.Voucher{ID: vID, UserID: ptrUUID(ownerID)}
	d.voucherRepo.On("GetByID", ctx, vID).Return(voucher, nil)
	d.voucherShareRepo.On("DeleteByVoucherID", ctx, vID).Return(nil)

	err := d.service.DeleteAllVoucherShares(ctx, ownerID, vID)
	assert.NoError(t, err)
	d.voucherShareRepo.AssertExpectations(t)
}

func TestDeleteAllVoucherShares_NotOwner(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	ownerID := uuid.New()
	vID := uuid.New()

	voucher := &models.Voucher{ID: vID, UserID: ptrUUID(ownerID)}
	d.voucherRepo.On("GetByID", ctx, vID).Return(voucher, nil)

	err := d.service.DeleteAllVoucherShares(ctx, uuid.New(), vID)
	assert.ErrorIs(t, err, ErrNotOwner)
}

func TestDeleteAllGiftCardShares_Success(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	ownerID := uuid.New()
	gcID := uuid.New()

	giftCard := &models.GiftCard{ID: gcID, UserID: ptrUUID(ownerID)}
	d.giftCardRepo.On("GetByID", ctx, gcID).Return(giftCard, nil)
	d.gcShareRepo.On("DeleteByGiftCardID", ctx, gcID).Return(nil)

	err := d.service.DeleteAllGiftCardShares(ctx, ownerID, gcID)
	assert.NoError(t, err)
	d.gcShareRepo.AssertExpectations(t)
}

func TestDeleteAllGiftCardShares_NotOwner(t *testing.T) {
	d := setupShareService()
	ctx := context.Background()
	ownerID := uuid.New()
	gcID := uuid.New()

	giftCard := &models.GiftCard{ID: gcID, UserID: ptrUUID(ownerID)}
	d.giftCardRepo.On("GetByID", ctx, gcID).Return(giftCard, nil)

	err := d.service.DeleteAllGiftCardShares(ctx, uuid.New(), gcID)
	assert.ErrorIs(t, err, ErrNotOwner)
}
