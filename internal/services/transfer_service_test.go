package services

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"savvy/internal/email"
	"savvy/internal/models"
	"savvy/internal/repository"
)

// ==================== Mock Definitions (prefixed with "transfer" to avoid conflicts) ====================

// --- transferMockCardRepo ---

type transferMockCardRepo struct{ mock.Mock }

func (m *transferMockCardRepo) Create(ctx context.Context, card *models.Card) error {
	return m.Called(ctx, card).Error(0)
}
func (m *transferMockCardRepo) GetByID(ctx context.Context, id uuid.UUID, _ ...string) (*models.Card, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Card), args.Error(1)
}
func (m *transferMockCardRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.Card, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.Card), args.Error(1)
}
func (m *transferMockCardRepo) GetSharedWithUser(ctx context.Context, userID uuid.UUID) ([]models.Card, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.Card), args.Error(1)
}
func (m *transferMockCardRepo) Update(ctx context.Context, card *models.Card) error {
	return m.Called(ctx, card).Error(0)
}
func (m *transferMockCardRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *transferMockCardRepo) Count(ctx context.Context, userID uuid.UUID) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}
func (m *transferMockCardRepo) GetAllForUserPaginated(_ context.Context, _ uuid.UUID, _ repository.PaginationParams) (*repository.PaginatedResult[models.Card], error) {
	return nil, nil
}
func (m *transferMockCardRepo) FindByCardNumber(_ context.Context, _ string, _ uuid.UUID) (*models.Card, error) {
	return nil, nil
}
func (m *transferMockCardRepo) FindSharedByCardNumber(_ context.Context, _ string, _ *uuid.UUID, _ uuid.UUID) (*models.Card, error) {
	return nil, nil
}
func (m *transferMockCardRepo) FindDeletedByCardNumber(_ context.Context, _ string, _ uuid.UUID) (*models.Card, error) {
	return nil, nil
}
func (m *transferMockCardRepo) RestoreByID(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
	return nil
}
func (m *transferMockCardRepo) Search(_ context.Context, _ uuid.UUID, _ string) ([]models.Card, error) {
	return nil, nil
}

var _ repository.CardRepository = (*transferMockCardRepo)(nil)

// --- transferMockVoucherRepo ---

type transferMockVoucherRepo struct{ mock.Mock }

func (m *transferMockVoucherRepo) Create(ctx context.Context, v *models.Voucher) error {
	return m.Called(ctx, v).Error(0)
}
func (m *transferMockVoucherRepo) GetByID(ctx context.Context, id uuid.UUID, _ ...string) (*models.Voucher, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Voucher), args.Error(1)
}
func (m *transferMockVoucherRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.Voucher, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.Voucher), args.Error(1)
}
func (m *transferMockVoucherRepo) GetSharedWithUser(ctx context.Context, userID uuid.UUID) ([]models.Voucher, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.Voucher), args.Error(1)
}
func (m *transferMockVoucherRepo) Update(ctx context.Context, v *models.Voucher) error {
	return m.Called(ctx, v).Error(0)
}
func (m *transferMockVoucherRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *transferMockVoucherRepo) Count(ctx context.Context, userID uuid.UUID) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}
func (m *transferMockVoucherRepo) GetAllForUserPaginated(_ context.Context, _ uuid.UUID, _ repository.PaginationParams) (*repository.PaginatedResult[models.Voucher], error) {
	return nil, nil
}
func (m *transferMockVoucherRepo) GetExpiringVouchers(_ context.Context, _ int) ([]models.Voucher, error) {
	return nil, nil
}
func (m *transferMockVoucherRepo) GetVouchersStartingTomorrow(_ context.Context) ([]models.Voucher, error) {
	return nil, nil
}
func (m *transferMockVoucherRepo) FindByVoucherCode(_ context.Context, _ string, _ uuid.UUID) (*models.Voucher, error) {
	return nil, nil
}
func (m *transferMockVoucherRepo) FindDeletedByCode(_ context.Context, _ string, _ uuid.UUID) (*models.Voucher, error) {
	return nil, nil
}
func (m *transferMockVoucherRepo) RestoreByID(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
	return nil
}
func (m *transferMockVoucherRepo) Search(_ context.Context, _ uuid.UUID, _ string) ([]models.Voucher, error) {
	return nil, nil
}

var _ repository.VoucherRepository = (*transferMockVoucherRepo)(nil)

// --- transferMockGiftCardRepo ---

type transferMockGiftCardRepo struct{ mock.Mock }

func (m *transferMockGiftCardRepo) Create(ctx context.Context, gc *models.GiftCard) error {
	return m.Called(ctx, gc).Error(0)
}
func (m *transferMockGiftCardRepo) GetByID(ctx context.Context, id uuid.UUID, _ ...string) (*models.GiftCard, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GiftCard), args.Error(1)
}
func (m *transferMockGiftCardRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.GiftCard, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.GiftCard), args.Error(1)
}
func (m *transferMockGiftCardRepo) GetSharedWithUser(ctx context.Context, userID uuid.UUID) ([]models.GiftCard, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.GiftCard), args.Error(1)
}
func (m *transferMockGiftCardRepo) Update(ctx context.Context, gc *models.GiftCard) error {
	return m.Called(ctx, gc).Error(0)
}
func (m *transferMockGiftCardRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *transferMockGiftCardRepo) Count(ctx context.Context, userID uuid.UUID) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}
func (m *transferMockGiftCardRepo) GetTotalBalance(ctx context.Context, userID uuid.UUID) (float64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(float64), args.Error(1)
}
func (m *transferMockGiftCardRepo) CreateTransaction(ctx context.Context, tx *models.GiftCardTransaction) error {
	return m.Called(ctx, tx).Error(0)
}
func (m *transferMockGiftCardRepo) GetTransaction(ctx context.Context, txID, gcID uuid.UUID) (*models.GiftCardTransaction, error) {
	args := m.Called(ctx, txID, gcID)
	return args.Get(0).(*models.GiftCardTransaction), args.Error(1)
}
func (m *transferMockGiftCardRepo) DeleteTransaction(ctx context.Context, txID uuid.UUID) error {
	return m.Called(ctx, txID).Error(0)
}
func (m *transferMockGiftCardRepo) GetAllForUserPaginated(_ context.Context, _ uuid.UUID, _ repository.PaginationParams) (*repository.PaginatedResult[models.GiftCard], error) {
	return nil, nil
}
func (m *transferMockGiftCardRepo) GetExpiringGiftCards(_ context.Context, _ int) ([]models.GiftCard, error) {
	return nil, nil
}
func (m *transferMockGiftCardRepo) FindByCardNumber(_ context.Context, _ string, _ uuid.UUID) (*models.GiftCard, error) {
	return nil, nil
}
func (m *transferMockGiftCardRepo) FindDeletedByCardNumber(_ context.Context, _ string, _ uuid.UUID) (*models.GiftCard, error) {
	return nil, nil
}
func (m *transferMockGiftCardRepo) RestoreByID(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
	return nil
}
func (m *transferMockGiftCardRepo) Search(_ context.Context, _ uuid.UUID, _ string) ([]models.GiftCard, error) {
	return nil, nil
}

var _ repository.GiftCardRepository = (*transferMockGiftCardRepo)(nil)

// --- transferMockUserRepo ---

type transferMockUserRepo struct{ mock.Mock }

func (m *transferMockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}
func (m *transferMockUserRepo) GetByEmail(ctx context.Context, e string) (*models.User, error) {
	args := m.Called(ctx, e)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}
func (m *transferMockUserRepo) Create(ctx context.Context, user *models.User) error {
	return m.Called(ctx, user).Error(0)
}
func (m *transferMockUserRepo) Update(ctx context.Context, user *models.User) error {
	return m.Called(ctx, user).Error(0)
}
func (m *transferMockUserRepo) GetAll(ctx context.Context) ([]models.User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.User), args.Error(1)
}
func (m *transferMockUserRepo) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]models.User, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]models.User), args.Error(1)
}
func (m *transferMockUserRepo) SearchByIDs(ctx context.Context, ids []uuid.UUID, query string) ([]models.User, error) {
	args := m.Called(ctx, ids, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.User), args.Error(1)
}

var _ repository.UserRepository = (*transferMockUserRepo)(nil)

// --- transferMockTransferRepo ---

type transferMockTransferRepo struct{ mock.Mock }

func (m *transferMockTransferRepo) TransferCardOwnership(ctx context.Context, card *models.Card, newOwnerID uuid.UUID) error {
	return m.Called(ctx, card, newOwnerID).Error(0)
}
func (m *transferMockTransferRepo) TransferVoucherOwnership(ctx context.Context, voucher *models.Voucher, newOwnerID uuid.UUID) error {
	return m.Called(ctx, voucher, newOwnerID).Error(0)
}
func (m *transferMockTransferRepo) TransferGiftCardOwnership(ctx context.Context, giftCard *models.GiftCard, newOwnerID uuid.UUID) error {
	return m.Called(ctx, giftCard, newOwnerID).Error(0)
}

var _ repository.TransferRepository = (*transferMockTransferRepo)(nil)

// --- transferMockAuditLogRepo ---

type transferMockAuditLogRepo struct{ mock.Mock }

func (m *transferMockAuditLogRepo) Create(ctx context.Context, log *models.AuditLog) error {
	return m.Called(ctx, log).Error(0)
}
func (m *transferMockAuditLogRepo) GetFiltered(ctx context.Context, filters repository.AuditLogFilters) ([]models.AuditLog, int64, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).([]models.AuditLog), args.Get(1).(int64), args.Error(2)
}
func (m *transferMockAuditLogRepo) RestoreResource(ctx context.Context, resourceType string, resourceID uuid.UUID) error {
	return m.Called(ctx, resourceType, resourceID).Error(0)
}

var _ repository.AuditLogRepository = (*transferMockAuditLogRepo)(nil)

// --- transferMockNotifService ---

type transferMockNotifService struct{ mock.Mock }

func (m *transferMockNotifService) CreateShareNotification(ctx context.Context, in ShareNotificationInput) error {
	return m.Called(ctx, in).Error(0)
}
func (m *transferMockNotifService) CreateTransferNotification(ctx context.Context, in TransferNotificationInput) error {
	return m.Called(ctx, in).Error(0)
}
func (m *transferMockNotifService) GetUserNotifications(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Notification, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0).([]models.Notification), args.Error(1)
}
func (m *transferMockNotifService) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}
func (m *transferMockNotifService) MarkAsRead(ctx context.Context, userID, notifID uuid.UUID) error {
	return m.Called(ctx, userID, notifID).Error(0)
}
func (m *transferMockNotifService) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	return m.Called(ctx, userID).Error(0)
}
func (m *transferMockNotifService) DeleteNotification(ctx context.Context, userID, notifID uuid.UUID) error {
	return m.Called(ctx, userID, notifID).Error(0)
}
func (m *transferMockNotifService) ArchiveOldRead(ctx context.Context, olderThanDays int) (int64, error) {
	args := m.Called(ctx, olderThanDays)
	return args.Get(0).(int64), args.Error(1)
}
func (m *transferMockNotifService) SetPushService(_ PushServiceInterface) {}
func (m *transferMockNotifService) SetEmailService(_ email.ServiceInterface, _ EmailTokenServiceInterface, _ string) {
}

var _ NotificationServiceInterface = (*transferMockNotifService)(nil)

// ==================== Test Setup ====================

type transferTestDeps struct {
	cardRepo     *transferMockCardRepo
	voucherRepo  *transferMockVoucherRepo
	giftCardRepo *transferMockGiftCardRepo
	userRepo     *transferMockUserRepo
	transferRepo *transferMockTransferRepo
	auditRepo    *transferMockAuditLogRepo
	notifService *transferMockNotifService
	service      TransferServiceInterface
}

func setupTransferService() *transferTestDeps {
	d := &transferTestDeps{
		cardRepo:     new(transferMockCardRepo),
		voucherRepo:  new(transferMockVoucherRepo),
		giftCardRepo: new(transferMockGiftCardRepo),
		userRepo:     new(transferMockUserRepo),
		transferRepo: new(transferMockTransferRepo),
		auditRepo:    new(transferMockAuditLogRepo),
		notifService: new(transferMockNotifService),
	}
	d.service = NewTransferService(
		d.cardRepo, d.voucherRepo, d.giftCardRepo,
		d.userRepo, d.transferRepo, d.auditRepo, d.notifService,
	)
	return d
}

func transferPtrUUID(id uuid.UUID) *uuid.UUID { return &id }

// ==================== TransferCardOwnership Tests ====================

func TestTransferCardOwnership_Success(t *testing.T) {
	d := setupTransferService()
	ctx := context.Background()
	ownerID := uuid.New()
	newOwnerID := uuid.New()
	cardID := uuid.New()

	card := &models.Card{ID: cardID, UserID: transferPtrUUID(ownerID)}
	newOwner := &models.User{ID: newOwnerID, Email: "new@example.com", FirstName: "New", LastName: "Owner"}
	currentOwner := &models.User{ID: ownerID, Email: "current@example.com", FirstName: "Current", LastName: "Owner"}

	// validateNewOwner: lookup new owner
	d.userRepo.On("GetByID", ctx, newOwnerID).Return(newOwner, nil)
	// GetByID for the card
	d.cardRepo.On("GetByID", ctx, cardID).Return(card, nil)
	// logTransferAudit: audit log creation
	d.auditRepo.On("Create", ctx, mock.AnythingOfType("*models.AuditLog")).Return(nil)
	// TransferCardOwnership in transfer repo
	d.transferRepo.On("TransferCardOwnership", ctx, card, newOwnerID).Return(nil)
	// sendTransferNotification: lookup current owner + create notification
	d.userRepo.On("GetByID", ctx, ownerID).Return(currentOwner, nil)
	transferMatch := mock.MatchedBy(func(in TransferNotificationInput) bool {
		return in.RecipientID == newOwnerID && in.FromUserID == ownerID && in.FromUserName == "Current Owner" &&
			in.ResourceType == "card" && in.ResourceID == cardID
	})
	d.notifService.On("CreateTransferNotification", ctx, transferMatch).Return(nil)

	err := d.service.TransferCardOwnership(ctx, cardID, newOwnerID, ownerID)
	assert.NoError(t, err)
	d.transferRepo.AssertCalled(t, "TransferCardOwnership", ctx, card, newOwnerID)
	d.auditRepo.AssertCalled(t, "Create", ctx, mock.AnythingOfType("*models.AuditLog"))
	d.notifService.AssertCalled(t, "CreateTransferNotification", ctx, transferMatch)
}

func TestTransferCardOwnership_NewOwnerNotFound(t *testing.T) {
	d := setupTransferService()
	ctx := context.Background()
	ownerID := uuid.New()
	newOwnerID := uuid.New()
	cardID := uuid.New()

	d.userRepo.On("GetByID", ctx, newOwnerID).Return(nil, errors.New("record not found"))

	err := d.service.TransferCardOwnership(ctx, cardID, newOwnerID, ownerID)
	assert.EqualError(t, err, "new owner not found")
	d.cardRepo.AssertNotCalled(t, "GetByID")
}

func TestTransferCardOwnership_TransferToSelf(t *testing.T) {
	d := setupTransferService()
	ctx := context.Background()
	ownerID := uuid.New()
	cardID := uuid.New()

	owner := &models.User{ID: ownerID, Email: "owner@example.com", FirstName: "Self", LastName: "User"}
	d.userRepo.On("GetByID", ctx, ownerID).Return(owner, nil)

	err := d.service.TransferCardOwnership(ctx, cardID, ownerID, ownerID)
	assert.EqualError(t, err, "cannot transfer to yourself")
	d.cardRepo.AssertNotCalled(t, "GetByID")
}

func TestTransferCardOwnership_CardNotFound(t *testing.T) {
	d := setupTransferService()
	ctx := context.Background()
	ownerID := uuid.New()
	newOwnerID := uuid.New()
	cardID := uuid.New()

	newOwner := &models.User{ID: newOwnerID, Email: "new@example.com", FirstName: "New", LastName: "Owner"}
	d.userRepo.On("GetByID", ctx, newOwnerID).Return(newOwner, nil)

	dbErr := errors.New("record not found")
	d.cardRepo.On("GetByID", ctx, cardID).Return(nil, dbErr)

	err := d.service.TransferCardOwnership(ctx, cardID, newOwnerID, ownerID)
	assert.ErrorIs(t, err, dbErr)
	d.transferRepo.AssertNotCalled(t, "TransferCardOwnership")
}

func TestTransferCardOwnership_NotOwner(t *testing.T) {
	d := setupTransferService()
	ctx := context.Background()
	ownerID := uuid.New()
	callerID := uuid.New()
	newOwnerID := uuid.New()
	cardID := uuid.New()

	card := &models.Card{ID: cardID, UserID: transferPtrUUID(ownerID)}
	newOwner := &models.User{ID: newOwnerID, Email: "new@example.com", FirstName: "New", LastName: "Owner"}
	d.userRepo.On("GetByID", ctx, newOwnerID).Return(newOwner, nil)
	d.cardRepo.On("GetByID", ctx, cardID).Return(card, nil)

	err := d.service.TransferCardOwnership(ctx, cardID, newOwnerID, callerID)
	assert.EqualError(t, err, "only owner can transfer")
	d.transferRepo.AssertNotCalled(t, "TransferCardOwnership")
}

func TestTransferCardOwnership_NilUserID(t *testing.T) {
	d := setupTransferService()
	ctx := context.Background()
	ownerID := uuid.New()
	newOwnerID := uuid.New()
	cardID := uuid.New()

	card := &models.Card{ID: cardID, UserID: nil}
	newOwner := &models.User{ID: newOwnerID, Email: "new@example.com", FirstName: "New", LastName: "Owner"}
	d.userRepo.On("GetByID", ctx, newOwnerID).Return(newOwner, nil)
	d.cardRepo.On("GetByID", ctx, cardID).Return(card, nil)

	err := d.service.TransferCardOwnership(ctx, cardID, newOwnerID, ownerID)
	assert.EqualError(t, err, "only owner can transfer")
}

func TestTransferCardOwnership_TransferRepoError(t *testing.T) {
	d := setupTransferService()
	ctx := context.Background()
	ownerID := uuid.New()
	newOwnerID := uuid.New()
	cardID := uuid.New()

	card := &models.Card{ID: cardID, UserID: transferPtrUUID(ownerID)}
	newOwner := &models.User{ID: newOwnerID, Email: "new@example.com", FirstName: "New", LastName: "Owner"}

	d.userRepo.On("GetByID", ctx, newOwnerID).Return(newOwner, nil)
	d.cardRepo.On("GetByID", ctx, cardID).Return(card, nil)

	transferErr := errors.New("db connection lost")
	d.transferRepo.On("TransferCardOwnership", ctx, card, newOwnerID).Return(transferErr)

	err := d.service.TransferCardOwnership(ctx, cardID, newOwnerID, ownerID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "transfer card ownership")
	assert.ErrorIs(t, err, transferErr)
	d.auditRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestTransferCardOwnership_AuditLogError_StillTransfers(t *testing.T) {
	d := setupTransferService()
	ctx := context.Background()
	ownerID := uuid.New()
	newOwnerID := uuid.New()
	cardID := uuid.New()

	card := &models.Card{ID: cardID, UserID: transferPtrUUID(ownerID)}
	newOwner := &models.User{ID: newOwnerID, Email: "new@example.com", FirstName: "New", LastName: "Owner"}
	currentOwner := &models.User{ID: ownerID, Email: "current@example.com", FirstName: "Current", LastName: "Owner"}

	d.userRepo.On("GetByID", ctx, newOwnerID).Return(newOwner, nil)
	d.cardRepo.On("GetByID", ctx, cardID).Return(card, nil)
	// Audit log fails, but transfer should still proceed (audit is best-effort logging)
	d.auditRepo.On("Create", ctx, mock.AnythingOfType("*models.AuditLog")).Return(errors.New("audit db error"))
	d.transferRepo.On("TransferCardOwnership", ctx, card, newOwnerID).Return(nil)
	d.userRepo.On("GetByID", ctx, ownerID).Return(currentOwner, nil)
	d.notifService.On("CreateTransferNotification", ctx, mock.MatchedBy(func(in TransferNotificationInput) bool {
		return in.RecipientID == newOwnerID && in.FromUserID == ownerID && in.ResourceType == "card" && in.ResourceID == cardID
	})).Return(nil)

	err := d.service.TransferCardOwnership(ctx, cardID, newOwnerID, ownerID)
	assert.NoError(t, err)
	d.transferRepo.AssertCalled(t, "TransferCardOwnership", ctx, card, newOwnerID)
}

func TestTransferCardOwnership_NotificationError_StillSucceeds(t *testing.T) {
	d := setupTransferService()
	ctx := context.Background()
	ownerID := uuid.New()
	newOwnerID := uuid.New()
	cardID := uuid.New()

	card := &models.Card{ID: cardID, UserID: transferPtrUUID(ownerID)}
	newOwner := &models.User{ID: newOwnerID, Email: "new@example.com", FirstName: "New", LastName: "Owner"}
	currentOwner := &models.User{ID: ownerID, Email: "current@example.com", FirstName: "Current", LastName: "Owner"}

	d.userRepo.On("GetByID", ctx, newOwnerID).Return(newOwner, nil)
	d.cardRepo.On("GetByID", ctx, cardID).Return(card, nil)
	d.auditRepo.On("Create", ctx, mock.AnythingOfType("*models.AuditLog")).Return(nil)
	d.transferRepo.On("TransferCardOwnership", ctx, card, newOwnerID).Return(nil)
	d.userRepo.On("GetByID", ctx, ownerID).Return(currentOwner, nil)
	// Notification fails, but transfer should still succeed (best-effort)
	d.notifService.On("CreateTransferNotification", ctx, mock.MatchedBy(func(in TransferNotificationInput) bool {
		return in.RecipientID == newOwnerID && in.FromUserID == ownerID && in.ResourceType == "card" && in.ResourceID == cardID
	})).Return(errors.New("notif error"))

	err := d.service.TransferCardOwnership(ctx, cardID, newOwnerID, ownerID)
	assert.NoError(t, err)
	d.transferRepo.AssertCalled(t, "TransferCardOwnership", ctx, card, newOwnerID)
}

func TestTransferCardOwnership_CurrentOwnerLookupFails_StillSucceeds(t *testing.T) {
	d := setupTransferService()
	ctx := context.Background()
	ownerID := uuid.New()
	newOwnerID := uuid.New()
	cardID := uuid.New()

	card := &models.Card{ID: cardID, UserID: transferPtrUUID(ownerID)}
	newOwner := &models.User{ID: newOwnerID, Email: "new@example.com", FirstName: "New", LastName: "Owner"}

	// First call: validate new owner (newOwnerID) -> success
	d.userRepo.On("GetByID", ctx, newOwnerID).Return(newOwner, nil)
	d.cardRepo.On("GetByID", ctx, cardID).Return(card, nil)
	d.auditRepo.On("Create", ctx, mock.AnythingOfType("*models.AuditLog")).Return(nil)
	d.transferRepo.On("TransferCardOwnership", ctx, card, newOwnerID).Return(nil)
	// Second call: sendTransferNotification looks up current owner (ownerID) -> fails
	d.userRepo.On("GetByID", ctx, ownerID).Return(nil, errors.New("user not found"))

	err := d.service.TransferCardOwnership(ctx, cardID, newOwnerID, ownerID)
	assert.NoError(t, err)
	d.transferRepo.AssertCalled(t, "TransferCardOwnership", ctx, card, newOwnerID)
	// Notification should NOT be called because current owner lookup failed
	d.notifService.AssertNotCalled(t, "CreateTransferNotification")
}

// ==================== TransferVoucherOwnership Tests ====================

func TestTransferVoucherOwnership_Success(t *testing.T) {
	d := setupTransferService()
	ctx := context.Background()
	ownerID := uuid.New()
	newOwnerID := uuid.New()
	voucherID := uuid.New()

	voucher := &models.Voucher{ID: voucherID, UserID: transferPtrUUID(ownerID)}
	newOwner := &models.User{ID: newOwnerID, Email: "new@example.com", FirstName: "New", LastName: "Owner"}
	currentOwner := &models.User{ID: ownerID, Email: "current@example.com", FirstName: "Current", LastName: "Owner"}

	d.userRepo.On("GetByID", ctx, newOwnerID).Return(newOwner, nil)
	d.voucherRepo.On("GetByID", ctx, voucherID).Return(voucher, nil)
	d.auditRepo.On("Create", ctx, mock.AnythingOfType("*models.AuditLog")).Return(nil)
	d.transferRepo.On("TransferVoucherOwnership", ctx, voucher, newOwnerID).Return(nil)
	d.userRepo.On("GetByID", ctx, ownerID).Return(currentOwner, nil)
	transferMatch := mock.MatchedBy(func(in TransferNotificationInput) bool {
		return in.RecipientID == newOwnerID && in.FromUserID == ownerID && in.FromUserName == "Current Owner" &&
			in.ResourceType == "voucher" && in.ResourceID == voucherID
	})
	d.notifService.On("CreateTransferNotification", ctx, transferMatch).Return(nil)

	err := d.service.TransferVoucherOwnership(ctx, voucherID, newOwnerID, ownerID)
	assert.NoError(t, err)
	d.transferRepo.AssertCalled(t, "TransferVoucherOwnership", ctx, voucher, newOwnerID)
	d.notifService.AssertCalled(t, "CreateTransferNotification", ctx, transferMatch)
}

func TestTransferVoucherOwnership_NotOwner(t *testing.T) {
	d := setupTransferService()
	ctx := context.Background()
	ownerID := uuid.New()
	callerID := uuid.New()
	newOwnerID := uuid.New()
	voucherID := uuid.New()

	voucher := &models.Voucher{ID: voucherID, UserID: transferPtrUUID(ownerID)}
	newOwner := &models.User{ID: newOwnerID, Email: "new@example.com", FirstName: "New", LastName: "Owner"}
	d.userRepo.On("GetByID", ctx, newOwnerID).Return(newOwner, nil)
	d.voucherRepo.On("GetByID", ctx, voucherID).Return(voucher, nil)

	err := d.service.TransferVoucherOwnership(ctx, voucherID, newOwnerID, callerID)
	assert.EqualError(t, err, "only owner can transfer")
	d.transferRepo.AssertNotCalled(t, "TransferVoucherOwnership")
}

func TestTransferVoucherOwnership_NilUserID(t *testing.T) {
	d := setupTransferService()
	ctx := context.Background()
	ownerID := uuid.New()
	newOwnerID := uuid.New()
	voucherID := uuid.New()

	voucher := &models.Voucher{ID: voucherID, UserID: nil}
	newOwner := &models.User{ID: newOwnerID, Email: "new@example.com", FirstName: "New", LastName: "Owner"}
	d.userRepo.On("GetByID", ctx, newOwnerID).Return(newOwner, nil)
	d.voucherRepo.On("GetByID", ctx, voucherID).Return(voucher, nil)

	err := d.service.TransferVoucherOwnership(ctx, voucherID, newOwnerID, ownerID)
	assert.EqualError(t, err, "only owner can transfer")
}

func TestTransferVoucherOwnership_VoucherNotFound(t *testing.T) {
	d := setupTransferService()
	ctx := context.Background()
	ownerID := uuid.New()
	newOwnerID := uuid.New()
	voucherID := uuid.New()

	newOwner := &models.User{ID: newOwnerID, Email: "new@example.com", FirstName: "New", LastName: "Owner"}
	d.userRepo.On("GetByID", ctx, newOwnerID).Return(newOwner, nil)

	dbErr := errors.New("record not found")
	d.voucherRepo.On("GetByID", ctx, voucherID).Return(nil, dbErr)

	err := d.service.TransferVoucherOwnership(ctx, voucherID, newOwnerID, ownerID)
	assert.ErrorIs(t, err, dbErr)
	d.transferRepo.AssertNotCalled(t, "TransferVoucherOwnership")
}

func TestTransferVoucherOwnership_TransferRepoError(t *testing.T) {
	d := setupTransferService()
	ctx := context.Background()
	ownerID := uuid.New()
	newOwnerID := uuid.New()
	voucherID := uuid.New()

	voucher := &models.Voucher{ID: voucherID, UserID: transferPtrUUID(ownerID)}
	newOwner := &models.User{ID: newOwnerID, Email: "new@example.com", FirstName: "New", LastName: "Owner"}

	d.userRepo.On("GetByID", ctx, newOwnerID).Return(newOwner, nil)
	d.voucherRepo.On("GetByID", ctx, voucherID).Return(voucher, nil)

	transferErr := errors.New("db error")
	d.transferRepo.On("TransferVoucherOwnership", ctx, voucher, newOwnerID).Return(transferErr)

	err := d.service.TransferVoucherOwnership(ctx, voucherID, newOwnerID, ownerID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "transfer voucher ownership")
	assert.ErrorIs(t, err, transferErr)
	d.auditRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestTransferVoucherOwnership_NewOwnerNotFound(t *testing.T) {
	d := setupTransferService()
	ctx := context.Background()
	ownerID := uuid.New()
	newOwnerID := uuid.New()
	voucherID := uuid.New()

	d.userRepo.On("GetByID", ctx, newOwnerID).Return(nil, errors.New("not found"))

	err := d.service.TransferVoucherOwnership(ctx, voucherID, newOwnerID, ownerID)
	assert.EqualError(t, err, "new owner not found")
}

func TestTransferVoucherOwnership_TransferToSelf(t *testing.T) {
	d := setupTransferService()
	ctx := context.Background()
	ownerID := uuid.New()
	voucherID := uuid.New()

	owner := &models.User{ID: ownerID, Email: "self@example.com", FirstName: "Self", LastName: "User"}
	d.userRepo.On("GetByID", ctx, ownerID).Return(owner, nil)

	err := d.service.TransferVoucherOwnership(ctx, voucherID, ownerID, ownerID)
	assert.EqualError(t, err, "cannot transfer to yourself")
}

// ==================== TransferGiftCardOwnership Tests ====================

func TestTransferGiftCardOwnership_Success(t *testing.T) {
	d := setupTransferService()
	ctx := context.Background()
	ownerID := uuid.New()
	newOwnerID := uuid.New()
	giftCardID := uuid.New()

	giftCard := &models.GiftCard{ID: giftCardID, UserID: transferPtrUUID(ownerID)}
	newOwner := &models.User{ID: newOwnerID, Email: "new@example.com", FirstName: "New", LastName: "Owner"}
	currentOwner := &models.User{ID: ownerID, Email: "current@example.com", FirstName: "Current", LastName: "Owner"}

	d.userRepo.On("GetByID", ctx, newOwnerID).Return(newOwner, nil)
	d.giftCardRepo.On("GetByID", ctx, giftCardID).Return(giftCard, nil)
	d.auditRepo.On("Create", ctx, mock.AnythingOfType("*models.AuditLog")).Return(nil)
	d.transferRepo.On("TransferGiftCardOwnership", ctx, giftCard, newOwnerID).Return(nil)
	d.userRepo.On("GetByID", ctx, ownerID).Return(currentOwner, nil)
	transferMatch := mock.MatchedBy(func(in TransferNotificationInput) bool {
		return in.RecipientID == newOwnerID && in.FromUserID == ownerID && in.FromUserName == "Current Owner" &&
			in.ResourceType == "gift_card" && in.ResourceID == giftCardID
	})
	d.notifService.On("CreateTransferNotification", ctx, transferMatch).Return(nil)

	err := d.service.TransferGiftCardOwnership(ctx, giftCardID, newOwnerID, ownerID)
	assert.NoError(t, err)
	d.transferRepo.AssertCalled(t, "TransferGiftCardOwnership", ctx, giftCard, newOwnerID)
	d.notifService.AssertCalled(t, "CreateTransferNotification", ctx, transferMatch)
}

func TestTransferGiftCardOwnership_NotOwner(t *testing.T) {
	d := setupTransferService()
	ctx := context.Background()
	ownerID := uuid.New()
	callerID := uuid.New()
	newOwnerID := uuid.New()
	giftCardID := uuid.New()

	giftCard := &models.GiftCard{ID: giftCardID, UserID: transferPtrUUID(ownerID)}
	newOwner := &models.User{ID: newOwnerID, Email: "new@example.com", FirstName: "New", LastName: "Owner"}
	d.userRepo.On("GetByID", ctx, newOwnerID).Return(newOwner, nil)
	d.giftCardRepo.On("GetByID", ctx, giftCardID).Return(giftCard, nil)

	err := d.service.TransferGiftCardOwnership(ctx, giftCardID, newOwnerID, callerID)
	assert.EqualError(t, err, "only owner can transfer")
	d.transferRepo.AssertNotCalled(t, "TransferGiftCardOwnership")
}

func TestTransferGiftCardOwnership_NilUserID(t *testing.T) {
	d := setupTransferService()
	ctx := context.Background()
	ownerID := uuid.New()
	newOwnerID := uuid.New()
	giftCardID := uuid.New()

	giftCard := &models.GiftCard{ID: giftCardID, UserID: nil}
	newOwner := &models.User{ID: newOwnerID, Email: "new@example.com", FirstName: "New", LastName: "Owner"}
	d.userRepo.On("GetByID", ctx, newOwnerID).Return(newOwner, nil)
	d.giftCardRepo.On("GetByID", ctx, giftCardID).Return(giftCard, nil)

	err := d.service.TransferGiftCardOwnership(ctx, giftCardID, newOwnerID, ownerID)
	assert.EqualError(t, err, "only owner can transfer")
}

func TestTransferGiftCardOwnership_GiftCardNotFound(t *testing.T) {
	d := setupTransferService()
	ctx := context.Background()
	ownerID := uuid.New()
	newOwnerID := uuid.New()
	giftCardID := uuid.New()

	newOwner := &models.User{ID: newOwnerID, Email: "new@example.com", FirstName: "New", LastName: "Owner"}
	d.userRepo.On("GetByID", ctx, newOwnerID).Return(newOwner, nil)

	dbErr := errors.New("record not found")
	d.giftCardRepo.On("GetByID", ctx, giftCardID).Return(nil, dbErr)

	err := d.service.TransferGiftCardOwnership(ctx, giftCardID, newOwnerID, ownerID)
	assert.ErrorIs(t, err, dbErr)
	d.transferRepo.AssertNotCalled(t, "TransferGiftCardOwnership")
}

func TestTransferGiftCardOwnership_TransferRepoError(t *testing.T) {
	d := setupTransferService()
	ctx := context.Background()
	ownerID := uuid.New()
	newOwnerID := uuid.New()
	giftCardID := uuid.New()

	giftCard := &models.GiftCard{ID: giftCardID, UserID: transferPtrUUID(ownerID)}
	newOwner := &models.User{ID: newOwnerID, Email: "new@example.com", FirstName: "New", LastName: "Owner"}

	d.userRepo.On("GetByID", ctx, newOwnerID).Return(newOwner, nil)
	d.giftCardRepo.On("GetByID", ctx, giftCardID).Return(giftCard, nil)

	transferErr := errors.New("db error")
	d.transferRepo.On("TransferGiftCardOwnership", ctx, giftCard, newOwnerID).Return(transferErr)

	err := d.service.TransferGiftCardOwnership(ctx, giftCardID, newOwnerID, ownerID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "transfer gift card ownership")
	assert.ErrorIs(t, err, transferErr)
	d.auditRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestTransferGiftCardOwnership_NewOwnerNotFound(t *testing.T) {
	d := setupTransferService()
	ctx := context.Background()
	ownerID := uuid.New()
	newOwnerID := uuid.New()
	giftCardID := uuid.New()

	d.userRepo.On("GetByID", ctx, newOwnerID).Return(nil, errors.New("not found"))

	err := d.service.TransferGiftCardOwnership(ctx, giftCardID, newOwnerID, ownerID)
	assert.EqualError(t, err, "new owner not found")
}

func TestTransferGiftCardOwnership_TransferToSelf(t *testing.T) {
	d := setupTransferService()
	ctx := context.Background()
	ownerID := uuid.New()
	giftCardID := uuid.New()

	owner := &models.User{ID: ownerID, Email: "self@example.com", FirstName: "Self", LastName: "User"}
	d.userRepo.On("GetByID", ctx, ownerID).Return(owner, nil)

	err := d.service.TransferGiftCardOwnership(ctx, giftCardID, ownerID, ownerID)
	assert.EqualError(t, err, "cannot transfer to yourself")
}

// ==================== Notes Privacy Tests ====================

// Same invariant as the share path: Card.Notes / GiftCard.Notes must never
// reach the notification Description, which ends up in the push body on the
// recipient's lockscreen. Voucher.Description is a real description and stays.
func TestTransferOwnership_NotesNeverReachNotificationDescription(t *testing.T) {
	const secret = "PIN 1234, back door code 9876"

	t.Run("card", func(t *testing.T) {
		d := setupTransferService()
		ctx := context.Background()
		ownerID, newOwnerID, cardID := uuid.New(), uuid.New(), uuid.New()

		card := &models.Card{ID: cardID, UserID: transferPtrUUID(ownerID), Notes: secret}
		newOwner := &models.User{ID: newOwnerID, Email: "new@example.com", FirstName: "New", LastName: "Owner"}
		currentOwner := &models.User{ID: ownerID, Email: "cur@example.com", FirstName: "Cur", LastName: "Owner"}

		d.userRepo.On("GetByID", ctx, newOwnerID).Return(newOwner, nil)
		d.cardRepo.On("GetByID", ctx, cardID).Return(card, nil)
		d.auditRepo.On("Create", ctx, mock.AnythingOfType("*models.AuditLog")).Return(nil)
		d.transferRepo.On("TransferCardOwnership", ctx, card, newOwnerID).Return(nil)
		d.userRepo.On("GetByID", ctx, ownerID).Return(currentOwner, nil)
		match := mock.MatchedBy(func(in TransferNotificationInput) bool { return in.Description == "" })
		d.notifService.On("CreateTransferNotification", ctx, match).Return(nil)

		assert.NoError(t, d.service.TransferCardOwnership(ctx, cardID, newOwnerID, ownerID))
		d.notifService.AssertCalled(t, "CreateTransferNotification", ctx, match)
	})

	t.Run("gift card", func(t *testing.T) {
		d := setupTransferService()
		ctx := context.Background()
		ownerID, newOwnerID, gcID := uuid.New(), uuid.New(), uuid.New()

		gc := &models.GiftCard{ID: gcID, UserID: transferPtrUUID(ownerID), Notes: secret, PIN: "4321"}
		newOwner := &models.User{ID: newOwnerID, Email: "new@example.com", FirstName: "New", LastName: "Owner"}
		currentOwner := &models.User{ID: ownerID, Email: "cur@example.com", FirstName: "Cur", LastName: "Owner"}

		d.userRepo.On("GetByID", ctx, newOwnerID).Return(newOwner, nil)
		d.giftCardRepo.On("GetByID", ctx, gcID).Return(gc, nil)
		d.auditRepo.On("Create", ctx, mock.AnythingOfType("*models.AuditLog")).Return(nil)
		d.transferRepo.On("TransferGiftCardOwnership", ctx, gc, newOwnerID).Return(nil)
		d.userRepo.On("GetByID", ctx, ownerID).Return(currentOwner, nil)
		match := mock.MatchedBy(func(in TransferNotificationInput) bool { return in.Description == "" })
		d.notifService.On("CreateTransferNotification", ctx, match).Return(nil)

		assert.NoError(t, d.service.TransferGiftCardOwnership(ctx, gcID, newOwnerID, ownerID))
		d.notifService.AssertCalled(t, "CreateTransferNotification", ctx, match)
	})

	t.Run("voucher keeps its description", func(t *testing.T) {
		d := setupTransferService()
		ctx := context.Background()
		ownerID, newOwnerID, voucherID := uuid.New(), uuid.New(), uuid.New()

		voucher := &models.Voucher{ID: voucherID, UserID: transferPtrUUID(ownerID), Description: "20% off"}
		newOwner := &models.User{ID: newOwnerID, Email: "new@example.com", FirstName: "New", LastName: "Owner"}
		currentOwner := &models.User{ID: ownerID, Email: "cur@example.com", FirstName: "Cur", LastName: "Owner"}

		d.userRepo.On("GetByID", ctx, newOwnerID).Return(newOwner, nil)
		d.voucherRepo.On("GetByID", ctx, voucherID).Return(voucher, nil)
		d.auditRepo.On("Create", ctx, mock.AnythingOfType("*models.AuditLog")).Return(nil)
		d.transferRepo.On("TransferVoucherOwnership", ctx, voucher, newOwnerID).Return(nil)
		d.userRepo.On("GetByID", ctx, ownerID).Return(currentOwner, nil)
		match := mock.MatchedBy(func(in TransferNotificationInput) bool { return in.Description == "20% off" })
		d.notifService.On("CreateTransferNotification", ctx, match).Return(nil)

		assert.NoError(t, d.service.TransferVoucherOwnership(ctx, voucherID, newOwnerID, ownerID))
		d.notifService.AssertCalled(t, "CreateTransferNotification", ctx, match)
	})
}
