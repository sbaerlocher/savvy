package shares

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"savvy/internal/models"
	"savvy/internal/services"
)

// ==================== Mock Services ====================

type mockAuthzService struct {
	services.AuthzServiceInterface
	cardResult    *services.ResourcePermissions
	cardErr       error
	voucherResult *services.ResourcePermissions
	voucherErr    error
	gcResult      *services.ResourcePermissions
	gcErr         error
}

func (m *mockAuthzService) CheckCardAccess(_ context.Context, _, _ uuid.UUID) (*services.ResourcePermissions, error) {
	return m.cardResult, m.cardErr
}

func (m *mockAuthzService) CheckVoucherAccess(_ context.Context, _, _ uuid.UUID) (*services.ResourcePermissions, error) {
	return m.voucherResult, m.voucherErr
}

func (m *mockAuthzService) CheckGiftCardAccess(_ context.Context, _, _ uuid.UUID) (*services.ResourcePermissions, error) {
	return m.gcResult, m.gcErr
}

type mockUserSvc struct {
	services.UserServiceInterface
	users map[string]*models.User
	err   error
}

func (m *mockUserSvc) GetUserByEmail(_ context.Context, email string) (*models.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	if u, ok := m.users[email]; ok {
		return u, nil
	}
	return nil, errors.New("record not found")
}

type mockShareSvc struct {
	services.ShareServiceInterface

	// Return values
	cardShares    []models.CardShare
	voucherShares []models.VoucherShare
	gcShares      []models.GiftCardShare
	err           error

	// Captured call arguments
	lastCreateCardCall    *createCardShareCall
	lastCreateVoucherCall *createVoucherShareCall
	lastCreateGCCall      *createGCShareCall
	lastUpdateCardCall    *updateCardShareCall
	lastUpdateGCCall      *updateGCShareCall
	lastDeleteCardCall    *deleteShareCall
	lastDeleteVoucherCall *deleteShareCall
	lastDeleteGCCall      *deleteShareCall
}

type createCardShareCall struct {
	callerUserID, cardID, sharedWithID uuid.UUID
	canEdit, canDelete                 bool
}
type createVoucherShareCall struct {
	callerUserID, voucherID, sharedWithID uuid.UUID
}
type createGCShareCall struct {
	callerUserID, giftCardID, sharedWithID  uuid.UUID
	canEdit, canDelete, canEditTransactions bool
}
type updateCardShareCall struct {
	callerUserID, cardID, sharedWithID uuid.UUID
	canEdit, canDelete                 bool
}
type updateGCShareCall struct {
	callerUserID, giftCardID, sharedWithID  uuid.UUID
	canEdit, canDelete, canEditTransactions bool
}
type deleteShareCall struct {
	callerUserID, resourceID, sharedWithID uuid.UUID
}

func (m *mockShareSvc) GetCardShares(_ context.Context, _ uuid.UUID) ([]models.CardShare, error) {
	return m.cardShares, m.err
}

func (m *mockShareSvc) GetVoucherShares(_ context.Context, _ uuid.UUID) ([]models.VoucherShare, error) {
	return m.voucherShares, m.err
}

func (m *mockShareSvc) GetGiftCardShares(_ context.Context, _ uuid.UUID) ([]models.GiftCardShare, error) {
	return m.gcShares, m.err
}

func (m *mockShareSvc) CreateCardShare(_ context.Context, callerUserID, cardID, sharedWithID uuid.UUID, canEdit, canDelete bool) error {
	m.lastCreateCardCall = &createCardShareCall{callerUserID, cardID, sharedWithID, canEdit, canDelete}
	return m.err
}

func (m *mockShareSvc) CreateVoucherShare(_ context.Context, callerUserID, voucherID, sharedWithID uuid.UUID) error {
	m.lastCreateVoucherCall = &createVoucherShareCall{callerUserID, voucherID, sharedWithID}
	return m.err
}

func (m *mockShareSvc) CreateGiftCardShare(_ context.Context, callerUserID, giftCardID, sharedWithID uuid.UUID, canEdit, canDelete, canEditTransactions bool) error {
	m.lastCreateGCCall = &createGCShareCall{callerUserID, giftCardID, sharedWithID, canEdit, canDelete, canEditTransactions}
	return m.err
}

func (m *mockShareSvc) UpdateCardShare(_ context.Context, callerUserID, cardID, sharedWithID uuid.UUID, canEdit, canDelete bool) error {
	m.lastUpdateCardCall = &updateCardShareCall{callerUserID, cardID, sharedWithID, canEdit, canDelete}
	return m.err
}

func (m *mockShareSvc) UpdateGiftCardShare(_ context.Context, callerUserID, giftCardID, sharedWithID uuid.UUID, canEdit, canDelete, canEditTransactions bool) error {
	m.lastUpdateGCCall = &updateGCShareCall{callerUserID, giftCardID, sharedWithID, canEdit, canDelete, canEditTransactions}
	return m.err
}

func (m *mockShareSvc) DeleteCardShare(_ context.Context, callerUserID, cardID, sharedWithID uuid.UUID) error {
	m.lastDeleteCardCall = &deleteShareCall{callerUserID, cardID, sharedWithID}
	return m.err
}

func (m *mockShareSvc) DeleteVoucherShare(_ context.Context, callerUserID, voucherID, sharedWithID uuid.UUID) error {
	m.lastDeleteVoucherCall = &deleteShareCall{callerUserID, voucherID, sharedWithID}
	return m.err
}

func (m *mockShareSvc) DeleteGiftCardShare(_ context.Context, callerUserID, giftCardID, sharedWithID uuid.UUID) error {
	m.lastDeleteGCCall = &deleteShareCall{callerUserID, giftCardID, sharedWithID}
	return m.err
}

// ==================== Capability Tests ====================

func TestCardAdapter_Capabilities(t *testing.T) {
	adapter := NewCardShareAdapter(nil, nil, nil)
	assert.Equal(t, "cards", adapter.ResourceType())
	assert.Equal(t, "Card", adapter.ResourceName())
	assert.True(t, adapter.SupportsEdit())
	assert.False(t, adapter.HasTransactionPermission())
}

func TestVoucherAdapter_Capabilities(t *testing.T) {
	adapter := NewVoucherShareAdapter(nil, nil, nil)
	assert.Equal(t, "vouchers", adapter.ResourceType())
	assert.Equal(t, "Voucher", adapter.ResourceName())
	assert.False(t, adapter.SupportsEdit())
	assert.False(t, adapter.HasTransactionPermission())
}

func TestGiftCardAdapter_Capabilities(t *testing.T) {
	adapter := NewGiftCardShareAdapter(nil, nil, nil)
	assert.Equal(t, "gift_cards", adapter.ResourceType())
	assert.Equal(t, "Gift Card", adapter.ResourceName())
	assert.True(t, adapter.SupportsEdit())
	assert.True(t, adapter.HasTransactionPermission())
}

func TestVoucherAdapter_UpdateNotSupported(t *testing.T) {
	adapter := NewVoucherShareAdapter(nil, nil, nil)
	err := adapter.UpdateShare(context.Background(), UpdateShareRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not supported")
}

// ==================== CheckOwnership Tests ====================

func TestCardAdapter_CheckOwnership_IsOwner(t *testing.T) {
	mock := &mockAuthzService{cardResult: &services.ResourcePermissions{IsOwner: true}}
	adapter := NewCardShareAdapter(nil, mock, nil)

	isOwner, err := adapter.CheckOwnership(context.Background(), uuid.New(), uuid.New())
	assert.NoError(t, err)
	assert.True(t, isOwner)
}

func TestCardAdapter_CheckOwnership_NotOwner(t *testing.T) {
	mock := &mockAuthzService{cardResult: &services.ResourcePermissions{IsOwner: false}}
	adapter := NewCardShareAdapter(nil, mock, nil)

	isOwner, err := adapter.CheckOwnership(context.Background(), uuid.New(), uuid.New())
	assert.NoError(t, err)
	assert.False(t, isOwner)
}

func TestCardAdapter_CheckOwnership_Forbidden(t *testing.T) {
	mock := &mockAuthzService{cardErr: services.ErrForbidden}
	adapter := NewCardShareAdapter(nil, mock, nil)

	isOwner, err := adapter.CheckOwnership(context.Background(), uuid.New(), uuid.New())
	assert.NoError(t, err)
	assert.False(t, isOwner)
}

func TestCardAdapter_CheckOwnership_GenericError(t *testing.T) {
	genericErr := errors.New("database connection failed")
	mock := &mockAuthzService{cardErr: genericErr}
	adapter := NewCardShareAdapter(nil, mock, nil)

	isOwner, err := adapter.CheckOwnership(context.Background(), uuid.New(), uuid.New())
	assert.Error(t, err)
	assert.False(t, isOwner)
	assert.Equal(t, genericErr, err)
}

func TestVoucherAdapter_CheckOwnership_IsOwner(t *testing.T) {
	mock := &mockAuthzService{voucherResult: &services.ResourcePermissions{IsOwner: true}}
	adapter := NewVoucherShareAdapter(nil, mock, nil)

	isOwner, err := adapter.CheckOwnership(context.Background(), uuid.New(), uuid.New())
	assert.NoError(t, err)
	assert.True(t, isOwner)
}

func TestVoucherAdapter_CheckOwnership_Forbidden(t *testing.T) {
	mock := &mockAuthzService{voucherErr: services.ErrForbidden}
	adapter := NewVoucherShareAdapter(nil, mock, nil)

	isOwner, err := adapter.CheckOwnership(context.Background(), uuid.New(), uuid.New())
	assert.NoError(t, err)
	assert.False(t, isOwner)
}

func TestVoucherAdapter_CheckOwnership_GenericError(t *testing.T) {
	genericErr := errors.New("database connection failed")
	mock := &mockAuthzService{voucherErr: genericErr}
	adapter := NewVoucherShareAdapter(nil, mock, nil)

	isOwner, err := adapter.CheckOwnership(context.Background(), uuid.New(), uuid.New())
	assert.Error(t, err)
	assert.False(t, isOwner)
	assert.Equal(t, genericErr, err)
}

func TestGiftCardAdapter_CheckOwnership_IsOwner(t *testing.T) {
	mock := &mockAuthzService{gcResult: &services.ResourcePermissions{IsOwner: true}}
	adapter := NewGiftCardShareAdapter(nil, mock, nil)

	isOwner, err := adapter.CheckOwnership(context.Background(), uuid.New(), uuid.New())
	assert.NoError(t, err)
	assert.True(t, isOwner)
}

func TestGiftCardAdapter_CheckOwnership_Forbidden(t *testing.T) {
	mock := &mockAuthzService{gcErr: services.ErrForbidden}
	adapter := NewGiftCardShareAdapter(nil, mock, nil)

	isOwner, err := adapter.CheckOwnership(context.Background(), uuid.New(), uuid.New())
	assert.NoError(t, err)
	assert.False(t, isOwner)
}

func TestGiftCardAdapter_CheckOwnership_GenericError(t *testing.T) {
	genericErr := errors.New("database connection failed")
	mock := &mockAuthzService{gcErr: genericErr}
	adapter := NewGiftCardShareAdapter(nil, mock, nil)

	isOwner, err := adapter.CheckOwnership(context.Background(), uuid.New(), uuid.New())
	assert.Error(t, err)
	assert.False(t, isOwner)
	assert.Equal(t, genericErr, err)
}

// ==================== Card Adapter: ListShares ====================

func TestCardAdapter_ListShares_Success(t *testing.T) {
	ctx := context.Background()
	cardID := uuid.New()
	sharedUserID := uuid.New()
	shareID := uuid.New()
	now := time.Now()

	sharedUser := &models.User{ID: sharedUserID, Email: "shared@example.com", FirstName: "Shared", LastName: "User"}
	shareSvc := &mockShareSvc{
		cardShares: []models.CardShare{
			{ID: shareID, CardID: cardID, SharedWithID: sharedUserID, SharedWithUser: sharedUser, CanEdit: true, CanDelete: false},
		},
	}
	shareSvc.cardShares[0].CreatedAt = now

	adapter := NewCardShareAdapter(shareSvc, nil, nil)
	views, err := adapter.ListShares(ctx, cardID)
	assert.NoError(t, err)
	assert.Len(t, views, 1)
	assert.Equal(t, shareID, views[0].ID)
	assert.Equal(t, cardID, views[0].ResourceID)
	assert.True(t, views[0].CanEdit)
	assert.False(t, views[0].CanDelete)
	assert.Equal(t, sharedUser, views[0].SharedWith)
}

func TestCardAdapter_ListShares_Empty(t *testing.T) {
	shareSvc := &mockShareSvc{cardShares: []models.CardShare{}}
	adapter := NewCardShareAdapter(shareSvc, nil, nil)

	views, err := adapter.ListShares(context.Background(), uuid.New())
	assert.NoError(t, err)
	assert.Empty(t, views)
}

func TestCardAdapter_ListShares_Error(t *testing.T) {
	shareSvc := &mockShareSvc{err: errors.New("db error")}
	adapter := NewCardShareAdapter(shareSvc, nil, nil)

	views, err := adapter.ListShares(context.Background(), uuid.New())
	assert.Error(t, err)
	assert.Nil(t, views)
}

// ==================== Card Adapter: CreateShare ====================

func TestCardAdapter_CreateShare_Success(t *testing.T) {
	ctx := context.Background()
	ownerID := uuid.New()
	cardID := uuid.New()
	sharedUser := &models.User{ID: uuid.New(), Email: "shared@example.com"}

	userSvc := &mockUserSvc{users: map[string]*models.User{sharedUser.Email: sharedUser}}
	shareSvc := &mockShareSvc{}
	adapter := NewCardShareAdapter(shareSvc, nil, userSvc)

	err := adapter.CreateShare(ctx, CreateShareRequest{
		UserID:          ownerID,
		ResourceID:      cardID,
		SharedWithEmail: sharedUser.Email,
		CanEdit:         true,
		CanDelete:       false,
	})
	assert.NoError(t, err)
	assert.NotNil(t, shareSvc.lastCreateCardCall)
	assert.Equal(t, ownerID, shareSvc.lastCreateCardCall.callerUserID)
	assert.Equal(t, cardID, shareSvc.lastCreateCardCall.cardID)
	assert.Equal(t, sharedUser.ID, shareSvc.lastCreateCardCall.sharedWithID)
	assert.True(t, shareSvc.lastCreateCardCall.canEdit)
	assert.False(t, shareSvc.lastCreateCardCall.canDelete)
}

func TestCardAdapter_CreateShare_UserNotFound(t *testing.T) {
	userSvc := &mockUserSvc{users: map[string]*models.User{}}
	adapter := NewCardShareAdapter(nil, nil, userSvc)

	err := adapter.CreateShare(context.Background(), CreateShareRequest{
		SharedWithEmail: "nonexistent@example.com",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

func TestCardAdapter_CreateShare_ServiceError(t *testing.T) {
	sharedUser := &models.User{ID: uuid.New(), Email: "shared@example.com"}
	userSvc := &mockUserSvc{users: map[string]*models.User{sharedUser.Email: sharedUser}}
	shareSvc := &mockShareSvc{err: errors.New("cannot share card with its owner")}
	adapter := NewCardShareAdapter(shareSvc, nil, userSvc)

	err := adapter.CreateShare(context.Background(), CreateShareRequest{
		UserID:          uuid.New(),
		ResourceID:      uuid.New(),
		SharedWithEmail: sharedUser.Email,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot share card with its owner")
}

// ==================== Card Adapter: UpdateShare ====================

func TestCardAdapter_UpdateShare_Success(t *testing.T) {
	ctx := context.Background()
	callerID := uuid.New()
	cardID := uuid.New()
	sharedWithID := uuid.New()

	shareSvc := &mockShareSvc{}
	adapter := NewCardShareAdapter(shareSvc, nil, nil)

	err := adapter.UpdateShare(ctx, UpdateShareRequest{
		CallerUserID: callerID,
		SharedWithID: sharedWithID,
		ResourceID:   cardID,
		CanEdit:      true,
		CanDelete:    true,
	})
	assert.NoError(t, err)
	assert.NotNil(t, shareSvc.lastUpdateCardCall)
	assert.Equal(t, callerID, shareSvc.lastUpdateCardCall.callerUserID)
	assert.Equal(t, cardID, shareSvc.lastUpdateCardCall.cardID)
	assert.Equal(t, sharedWithID, shareSvc.lastUpdateCardCall.sharedWithID)
	assert.True(t, shareSvc.lastUpdateCardCall.canEdit)
	assert.True(t, shareSvc.lastUpdateCardCall.canDelete)
}

func TestCardAdapter_UpdateShare_Error(t *testing.T) {
	shareSvc := &mockShareSvc{err: errors.New("share not found")}
	adapter := NewCardShareAdapter(shareSvc, nil, nil)

	err := adapter.UpdateShare(context.Background(), UpdateShareRequest{
		CallerUserID: uuid.New(),
		SharedWithID: uuid.New(),
		ResourceID:   uuid.New(),
	})
	assert.Error(t, err)
}

// ==================== Card Adapter: DeleteShare ====================

func TestCardAdapter_DeleteShare_Success(t *testing.T) {
	ctx := context.Background()
	callerID := uuid.New()
	sharedWithID := uuid.New()
	cardID := uuid.New()

	shareSvc := &mockShareSvc{}
	adapter := NewCardShareAdapter(shareSvc, nil, nil)

	err := adapter.DeleteShare(ctx, callerID, sharedWithID, cardID)
	assert.NoError(t, err)
	assert.NotNil(t, shareSvc.lastDeleteCardCall)
	assert.Equal(t, callerID, shareSvc.lastDeleteCardCall.callerUserID)
	assert.Equal(t, cardID, shareSvc.lastDeleteCardCall.resourceID)
	assert.Equal(t, sharedWithID, shareSvc.lastDeleteCardCall.sharedWithID)
}

func TestCardAdapter_DeleteShare_Error(t *testing.T) {
	shareSvc := &mockShareSvc{err: services.ErrNotOwner}
	adapter := NewCardShareAdapter(shareSvc, nil, nil)

	err := adapter.DeleteShare(context.Background(), uuid.New(), uuid.New(), uuid.New())
	assert.Error(t, err)
	assert.Equal(t, services.ErrNotOwner, err)
}

// ==================== Voucher Adapter: ListShares ====================

func TestVoucherAdapter_ListShares_Success(t *testing.T) {
	ctx := context.Background()
	voucherID := uuid.New()
	sharedUserID := uuid.New()
	shareID := uuid.New()

	sharedUser := &models.User{ID: sharedUserID, Email: "shared@example.com"}
	shareSvc := &mockShareSvc{
		voucherShares: []models.VoucherShare{
			{ID: shareID, VoucherID: voucherID, SharedWithID: sharedUserID, SharedWithUser: sharedUser},
		},
	}

	adapter := NewVoucherShareAdapter(shareSvc, nil, nil)
	views, err := adapter.ListShares(ctx, voucherID)
	assert.NoError(t, err)
	assert.Len(t, views, 1)
	assert.Equal(t, shareID, views[0].ID)
	assert.Equal(t, voucherID, views[0].ResourceID)
	assert.False(t, views[0].CanEdit, "voucher shares are always read-only")
	assert.False(t, views[0].CanDelete, "voucher shares are always read-only")
}

func TestVoucherAdapter_ListShares_Error(t *testing.T) {
	shareSvc := &mockShareSvc{err: errors.New("db error")}
	adapter := NewVoucherShareAdapter(shareSvc, nil, nil)

	views, err := adapter.ListShares(context.Background(), uuid.New())
	assert.Error(t, err)
	assert.Nil(t, views)
}

// ==================== Voucher Adapter: CreateShare ====================

func TestVoucherAdapter_CreateShare_Success(t *testing.T) {
	ctx := context.Background()
	ownerID := uuid.New()
	voucherID := uuid.New()
	sharedUser := &models.User{ID: uuid.New(), Email: "shared@example.com"}

	userSvc := &mockUserSvc{users: map[string]*models.User{sharedUser.Email: sharedUser}}
	shareSvc := &mockShareSvc{}
	adapter := NewVoucherShareAdapter(shareSvc, nil, userSvc)

	err := adapter.CreateShare(ctx, CreateShareRequest{
		UserID:          ownerID,
		ResourceID:      voucherID,
		SharedWithEmail: sharedUser.Email,
	})
	assert.NoError(t, err)
	assert.NotNil(t, shareSvc.lastCreateVoucherCall)
	assert.Equal(t, ownerID, shareSvc.lastCreateVoucherCall.callerUserID)
	assert.Equal(t, voucherID, shareSvc.lastCreateVoucherCall.voucherID)
	assert.Equal(t, sharedUser.ID, shareSvc.lastCreateVoucherCall.sharedWithID)
}

func TestVoucherAdapter_CreateShare_UserNotFound(t *testing.T) {
	userSvc := &mockUserSvc{users: map[string]*models.User{}}
	adapter := NewVoucherShareAdapter(nil, nil, userSvc)

	err := adapter.CreateShare(context.Background(), CreateShareRequest{
		SharedWithEmail: "nonexistent@example.com",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

func TestVoucherAdapter_CreateShare_ServiceError(t *testing.T) {
	sharedUser := &models.User{ID: uuid.New(), Email: "shared@example.com"}
	userSvc := &mockUserSvc{users: map[string]*models.User{sharedUser.Email: sharedUser}}
	shareSvc := &mockShareSvc{err: errors.New("cannot share voucher with its owner")}
	adapter := NewVoucherShareAdapter(shareSvc, nil, userSvc)

	err := adapter.CreateShare(context.Background(), CreateShareRequest{
		UserID:          uuid.New(),
		ResourceID:      uuid.New(),
		SharedWithEmail: sharedUser.Email,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot share voucher with its owner")
}

// ==================== Voucher Adapter: DeleteShare ====================

func TestVoucherAdapter_DeleteShare_Success(t *testing.T) {
	ctx := context.Background()
	callerID := uuid.New()
	sharedWithID := uuid.New()
	voucherID := uuid.New()

	shareSvc := &mockShareSvc{}
	adapter := NewVoucherShareAdapter(shareSvc, nil, nil)

	err := adapter.DeleteShare(ctx, callerID, sharedWithID, voucherID)
	assert.NoError(t, err)
	assert.NotNil(t, shareSvc.lastDeleteVoucherCall)
	assert.Equal(t, callerID, shareSvc.lastDeleteVoucherCall.callerUserID)
	assert.Equal(t, voucherID, shareSvc.lastDeleteVoucherCall.resourceID)
	assert.Equal(t, sharedWithID, shareSvc.lastDeleteVoucherCall.sharedWithID)
}

func TestVoucherAdapter_DeleteShare_Error(t *testing.T) {
	shareSvc := &mockShareSvc{err: services.ErrNotOwner}
	adapter := NewVoucherShareAdapter(shareSvc, nil, nil)

	err := adapter.DeleteShare(context.Background(), uuid.New(), uuid.New(), uuid.New())
	assert.Error(t, err)
	assert.Equal(t, services.ErrNotOwner, err)
}

// ==================== Gift Card Adapter: ListShares ====================

func TestGiftCardAdapter_ListShares_Success(t *testing.T) {
	ctx := context.Background()
	gcID := uuid.New()
	sharedUserID := uuid.New()
	shareID := uuid.New()

	sharedUser := &models.User{ID: sharedUserID, Email: "shared@example.com"}
	shareSvc := &mockShareSvc{
		gcShares: []models.GiftCardShare{
			{ID: shareID, GiftCardID: gcID, SharedWithID: sharedUserID, SharedWithUser: sharedUser,
				CanEdit: true, CanDelete: false, CanEditTransactions: true},
		},
	}

	adapter := NewGiftCardShareAdapter(shareSvc, nil, nil)
	views, err := adapter.ListShares(ctx, gcID)
	assert.NoError(t, err)
	assert.Len(t, views, 1)
	assert.Equal(t, shareID, views[0].ID)
	assert.Equal(t, gcID, views[0].ResourceID)
	assert.True(t, views[0].CanEdit)
	assert.False(t, views[0].CanDelete)
	assert.True(t, views[0].CanEditTransactions)
	assert.Equal(t, sharedUser, views[0].SharedWith)
}

func TestGiftCardAdapter_ListShares_Error(t *testing.T) {
	shareSvc := &mockShareSvc{err: errors.New("db error")}
	adapter := NewGiftCardShareAdapter(shareSvc, nil, nil)

	views, err := adapter.ListShares(context.Background(), uuid.New())
	assert.Error(t, err)
	assert.Nil(t, views)
}

// ==================== Gift Card Adapter: CreateShare ====================

func TestGiftCardAdapter_CreateShare_Success(t *testing.T) {
	ctx := context.Background()
	ownerID := uuid.New()
	gcID := uuid.New()
	sharedUser := &models.User{ID: uuid.New(), Email: "shared@example.com"}

	userSvc := &mockUserSvc{users: map[string]*models.User{sharedUser.Email: sharedUser}}
	shareSvc := &mockShareSvc{}
	adapter := NewGiftCardShareAdapter(shareSvc, nil, userSvc)

	err := adapter.CreateShare(ctx, CreateShareRequest{
		UserID:              ownerID,
		ResourceID:          gcID,
		SharedWithEmail:     sharedUser.Email,
		CanEdit:             true,
		CanDelete:           false,
		CanEditTransactions: true,
	})
	assert.NoError(t, err)
	assert.NotNil(t, shareSvc.lastCreateGCCall)
	assert.Equal(t, ownerID, shareSvc.lastCreateGCCall.callerUserID)
	assert.Equal(t, gcID, shareSvc.lastCreateGCCall.giftCardID)
	assert.Equal(t, sharedUser.ID, shareSvc.lastCreateGCCall.sharedWithID)
	assert.True(t, shareSvc.lastCreateGCCall.canEdit)
	assert.False(t, shareSvc.lastCreateGCCall.canDelete)
	assert.True(t, shareSvc.lastCreateGCCall.canEditTransactions)
}

func TestGiftCardAdapter_CreateShare_UserNotFound(t *testing.T) {
	userSvc := &mockUserSvc{users: map[string]*models.User{}}
	adapter := NewGiftCardShareAdapter(nil, nil, userSvc)

	err := adapter.CreateShare(context.Background(), CreateShareRequest{
		SharedWithEmail: "nonexistent@example.com",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

func TestGiftCardAdapter_CreateShare_ServiceError(t *testing.T) {
	sharedUser := &models.User{ID: uuid.New(), Email: "shared@example.com"}
	userSvc := &mockUserSvc{users: map[string]*models.User{sharedUser.Email: sharedUser}}
	shareSvc := &mockShareSvc{err: errors.New("cannot share gift card with its owner")}
	adapter := NewGiftCardShareAdapter(shareSvc, nil, userSvc)

	err := adapter.CreateShare(context.Background(), CreateShareRequest{
		UserID:          uuid.New(),
		ResourceID:      uuid.New(),
		SharedWithEmail: sharedUser.Email,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot share gift card with its owner")
}

// ==================== Gift Card Adapter: UpdateShare ====================

func TestGiftCardAdapter_UpdateShare_Success(t *testing.T) {
	ctx := context.Background()
	callerID := uuid.New()
	gcID := uuid.New()
	sharedWithID := uuid.New()

	shareSvc := &mockShareSvc{}
	adapter := NewGiftCardShareAdapter(shareSvc, nil, nil)

	err := adapter.UpdateShare(ctx, UpdateShareRequest{
		CallerUserID:        callerID,
		SharedWithID:        sharedWithID,
		ResourceID:          gcID,
		CanEdit:             true,
		CanDelete:           true,
		CanEditTransactions: true,
	})
	assert.NoError(t, err)
	assert.NotNil(t, shareSvc.lastUpdateGCCall)
	assert.Equal(t, callerID, shareSvc.lastUpdateGCCall.callerUserID)
	assert.Equal(t, gcID, shareSvc.lastUpdateGCCall.giftCardID)
	assert.Equal(t, sharedWithID, shareSvc.lastUpdateGCCall.sharedWithID)
	assert.True(t, shareSvc.lastUpdateGCCall.canEdit)
	assert.True(t, shareSvc.lastUpdateGCCall.canDelete)
	assert.True(t, shareSvc.lastUpdateGCCall.canEditTransactions)
}

func TestGiftCardAdapter_UpdateShare_Error(t *testing.T) {
	shareSvc := &mockShareSvc{err: errors.New("share not found")}
	adapter := NewGiftCardShareAdapter(shareSvc, nil, nil)

	err := adapter.UpdateShare(context.Background(), UpdateShareRequest{
		CallerUserID: uuid.New(),
		SharedWithID: uuid.New(),
		ResourceID:   uuid.New(),
	})
	assert.Error(t, err)
}

// ==================== Gift Card Adapter: DeleteShare ====================

func TestGiftCardAdapter_DeleteShare_Success(t *testing.T) {
	ctx := context.Background()
	callerID := uuid.New()
	sharedWithID := uuid.New()
	gcID := uuid.New()

	shareSvc := &mockShareSvc{}
	adapter := NewGiftCardShareAdapter(shareSvc, nil, nil)

	err := adapter.DeleteShare(ctx, callerID, sharedWithID, gcID)
	assert.NoError(t, err)
	assert.NotNil(t, shareSvc.lastDeleteGCCall)
	assert.Equal(t, callerID, shareSvc.lastDeleteGCCall.callerUserID)
	assert.Equal(t, gcID, shareSvc.lastDeleteGCCall.resourceID)
	assert.Equal(t, sharedWithID, shareSvc.lastDeleteGCCall.sharedWithID)
}

func TestGiftCardAdapter_DeleteShare_Error(t *testing.T) {
	shareSvc := &mockShareSvc{err: services.ErrNotOwner}
	adapter := NewGiftCardShareAdapter(shareSvc, nil, nil)

	err := adapter.DeleteShare(context.Background(), uuid.New(), uuid.New(), uuid.New())
	assert.Error(t, err)
	assert.Equal(t, services.ErrNotOwner, err)
}
