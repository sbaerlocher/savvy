package vouchers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/text/language"
	savvyi18n "savvy/internal/i18n"
	"savvy/internal/models"
	"savvy/internal/services"
)

func init() {
	// Initialize i18n bundle for tests
	savvyi18n.Bundle = i18n.NewBundle(language.German)
	savvyi18n.Bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
}

// MockVoucherService is a manual mock for VoucherServiceInterface
type MockVoucherService struct {
	mock.Mock
}

func (m *MockVoucherService) CreateVoucher(ctx context.Context, voucher *models.Voucher) error {
	args := m.Called(ctx, voucher)
	return args.Error(0)
}

func (m *MockVoucherService) GetVoucher(ctx context.Context, id uuid.UUID) (*models.Voucher, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Voucher), args.Error(1)
}

func (m *MockVoucherService) GetUserVouchers(ctx context.Context, userID uuid.UUID) ([]models.Voucher, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Voucher), args.Error(1)
}

func (m *MockVoucherService) UpdateVoucher(ctx context.Context, voucher *models.Voucher) error {
	args := m.Called(ctx, voucher)
	return args.Error(0)
}

func (m *MockVoucherService) DeleteVoucher(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockVoucherService) CountUserVouchers(ctx context.Context, userID uuid.UUID) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

// MockAuthzService is a manual mock for AuthzServiceInterface
type MockAuthzService struct {
	mock.Mock
}

func (m *MockAuthzService) CheckCardAccess(ctx context.Context, userID, cardID uuid.UUID) (*services.ResourcePermissions, error) {
	args := m.Called(ctx, userID, cardID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.ResourcePermissions), args.Error(1)
}

func (m *MockAuthzService) CheckVoucherAccess(ctx context.Context, userID, voucherID uuid.UUID) (*services.ResourcePermissions, error) {
	args := m.Called(ctx, userID, voucherID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.ResourcePermissions), args.Error(1)
}

func (m *MockAuthzService) CheckGiftCardAccess(ctx context.Context, userID, giftCardID uuid.UUID) (*services.ResourcePermissions, error) {
	args := m.Called(ctx, userID, giftCardID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.ResourcePermissions), args.Error(1)
}

// MockMerchantService is a manual mock for MerchantServiceInterface
type MockMerchantService struct {
	mock.Mock
}

func (m *MockMerchantService) CreateMerchant(ctx context.Context, merchant *models.Merchant) error {
	args := m.Called(ctx, merchant)
	return args.Error(0)
}

func (m *MockMerchantService) GetMerchantByID(ctx context.Context, id uuid.UUID) (*models.Merchant, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Merchant), args.Error(1)
}

func (m *MockMerchantService) GetMerchantByName(ctx context.Context, name string) (*models.Merchant, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Merchant), args.Error(1)
}

func (m *MockMerchantService) GetAllMerchants(ctx context.Context) ([]models.Merchant, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Merchant), args.Error(1)
}

func (m *MockMerchantService) SearchMerchants(ctx context.Context, query string) ([]models.Merchant, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Merchant), args.Error(1)
}

func (m *MockMerchantService) UpdateMerchant(ctx context.Context, merchant *models.Merchant) error {
	args := m.Called(ctx, merchant)
	return args.Error(0)
}

func (m *MockMerchantService) DeleteMerchant(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockMerchantService) GetMerchantCount(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

// MockFavoriteService is a manual mock for FavoriteServiceInterface
type MockFavoriteService struct {
	mock.Mock
}

func (m *MockFavoriteService) ToggleFavorite(ctx context.Context, userID uuid.UUID, resourceType string, resourceID uuid.UUID) error {
	args := m.Called(ctx, userID, resourceType, resourceID)
	return args.Error(0)
}

func (m *MockFavoriteService) GetUserFavorites(ctx context.Context, userID uuid.UUID) ([]models.UserFavorite, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.UserFavorite), args.Error(1)
}

func (m *MockFavoriteService) IsFavorite(ctx context.Context, userID uuid.UUID, resourceType string, resourceID uuid.UUID) (bool, error) {
	args := m.Called(ctx, userID, resourceType, resourceID)
	return args.Bool(0), args.Error(1)
}

func (m *MockFavoriteService) GetFavoriteCards(ctx context.Context, userID uuid.UUID) ([]models.Card, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Card), args.Error(1)
}

func (m *MockFavoriteService) GetFavoriteVouchers(ctx context.Context, userID uuid.UUID) ([]models.Voucher, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Voucher), args.Error(1)
}

func (m *MockFavoriteService) GetFavoriteGiftCards(ctx context.Context, userID uuid.UUID) ([]models.GiftCard, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.GiftCard), args.Error(1)
}

// MockShareService is a manual mock for ShareServiceInterface
type MockShareService struct {
	mock.Mock
}

func (m *MockShareService) ShareCard(ctx context.Context, cardID, sharedWithID uuid.UUID, canEdit, canDelete bool) error {
	args := m.Called(ctx, cardID, sharedWithID, canEdit, canDelete)
	return args.Error(0)
}

func (m *MockShareService) ShareVoucher(ctx context.Context, voucherID, sharedWithID uuid.UUID) error {
	args := m.Called(ctx, voucherID, sharedWithID)
	return args.Error(0)
}

func (m *MockShareService) ShareGiftCard(ctx context.Context, giftCardID, sharedWithID uuid.UUID, canEdit, canDelete, canEditTransactions bool) error {
	args := m.Called(ctx, giftCardID, sharedWithID, canEdit, canDelete, canEditTransactions)
	return args.Error(0)
}

func (m *MockShareService) UpdateCardShare(ctx context.Context, shareID uuid.UUID, canEdit, canDelete bool) error {
	args := m.Called(ctx, shareID, canEdit, canDelete)
	return args.Error(0)
}

func (m *MockShareService) UpdateGiftCardShare(ctx context.Context, shareID uuid.UUID, canEdit, canDelete, canEditTransactions bool) error {
	args := m.Called(ctx, shareID, canEdit, canDelete, canEditTransactions)
	return args.Error(0)
}

func (m *MockShareService) RevokeCardShare(ctx context.Context, shareID uuid.UUID) error {
	args := m.Called(ctx, shareID)
	return args.Error(0)
}

func (m *MockShareService) RevokeVoucherShare(ctx context.Context, shareID uuid.UUID) error {
	args := m.Called(ctx, shareID)
	return args.Error(0)
}

func (m *MockShareService) RevokeGiftCardShare(ctx context.Context, shareID uuid.UUID) error {
	args := m.Called(ctx, shareID)
	return args.Error(0)
}

func (m *MockShareService) GetCardShares(ctx context.Context, cardID uuid.UUID) ([]models.CardShare, error) {
	args := m.Called(ctx, cardID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.CardShare), args.Error(1)
}

func (m *MockShareService) GetVoucherShares(ctx context.Context, voucherID uuid.UUID) ([]models.VoucherShare, error) {
	args := m.Called(ctx, voucherID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.VoucherShare), args.Error(1)
}

func (m *MockShareService) GetGiftCardShares(ctx context.Context, giftCardID uuid.UUID) ([]models.GiftCardShare, error) {
	args := m.Called(ctx, giftCardID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.GiftCardShare), args.Error(1)
}

func (m *MockShareService) HasCardAccess(ctx context.Context, cardID, userID uuid.UUID) (bool, error) {
	args := m.Called(ctx, cardID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockShareService) HasVoucherAccess(ctx context.Context, voucherID, userID uuid.UUID) (bool, error) {
	args := m.Called(ctx, voucherID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockShareService) HasGiftCardAccess(ctx context.Context, giftCardID, userID uuid.UUID) (bool, error) {
	args := m.Called(ctx, giftCardID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockShareService) CreateCardShare(ctx context.Context, cardID, sharedWithID uuid.UUID, canEdit, canDelete bool) error {
	args := m.Called(ctx, cardID, sharedWithID, canEdit, canDelete)
	return args.Error(0)
}

func (m *MockShareService) CreateVoucherShare(ctx context.Context, voucherID, sharedWithID uuid.UUID) error {
	args := m.Called(ctx, voucherID, sharedWithID)
	return args.Error(0)
}

func (m *MockShareService) CreateGiftCardShare(ctx context.Context, giftCardID, sharedWithID uuid.UUID, canEdit, canDelete, canEditTransactions bool) error {
	args := m.Called(ctx, giftCardID, sharedWithID, canEdit, canDelete, canEditTransactions)
	return args.Error(0)
}

func (m *MockShareService) GetSharedUsers(ctx context.Context, userID uuid.UUID, searchQuery string) ([]models.User, error) {
	args := m.Called(ctx, userID, searchQuery)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.User), args.Error(1)
}

// ============================================================================
// TEST HELPERS
// ============================================================================

// setupI18nContext sets up i18n localizer in the request context for testing
func setupI18nContext(c echo.Context) {
	localizer := savvyi18n.NewLocalizer("de")
	ctx := savvyi18n.SetLocalizer(c.Request().Context(), localizer)
	c.SetRequest(c.Request().WithContext(ctx))
}

// ============================================================================
// TESTS
// ============================================================================

func TestShowHandler_Owner(t *testing.T) {
	// Setup
	e := echo.New()
	mockVoucherService := new(MockVoucherService)
	mockAuthzService := new(MockAuthzService)
	mockShareService := new(MockShareService)
	mockMerchantService := new(MockMerchantService)
	mockFavoriteService := new(MockFavoriteService)

	handler := &Handler{
		voucherService:  mockVoucherService,
		authzService:    mockAuthzService,
		shareService:    mockShareService,
		merchantService: mockMerchantService,
		favoriteService: mockFavoriteService,
	}

	// Test data
	userID := uuid.New()
	voucherID := uuid.New()
	user := &models.User{
		ID:    userID,
		Email: "test@example.com",
	}
	voucher := &models.Voucher{
		ID:     voucherID,
		UserID: &userID,
		Code:   "TEST123",
	}
	perms := &services.ResourcePermissions{
		IsOwner:   true,
		CanView:   true,
		CanEdit:   true,
		CanDelete: true,
	}
	merchants := []models.Merchant{
		{ID: uuid.New(), Name: "Test Merchant"},
	}
	shares := []models.VoucherShare{}

	// Mock expectations
	mockAuthzService.On("CheckVoucherAccess", mock.Anything, userID, voucherID).Return(perms, nil)
	mockVoucherService.On("GetVoucher", mock.Anything, voucherID).Return(voucher, nil)
	mockShareService.On("GetVoucherShares", mock.Anything, voucherID).Return(shares, nil)
	mockMerchantService.On("GetAllMerchants", mock.Anything).Return(merchants, nil)
	mockFavoriteService.On("IsFavorite", mock.Anything, userID, "voucher", voucherID).Return(false, nil)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/vouchers/"+voucherID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(voucherID.String())
	c.Set("current_user", user)
	c.Set("csrf", "test-token")
	setupI18nContext(c)

	// Execute
	err := handler.Show(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockAuthzService.AssertExpectations(t)
	mockVoucherService.AssertExpectations(t)
	mockShareService.AssertExpectations(t)
	mockMerchantService.AssertExpectations(t)
	mockFavoriteService.AssertExpectations(t)
}

func TestShowHandler_SharedUser(t *testing.T) {
	// Setup
	e := echo.New()
	mockVoucherService := new(MockVoucherService)
	mockAuthzService := new(MockAuthzService)
	mockMerchantService := new(MockMerchantService)
	mockFavoriteService := new(MockFavoriteService)

	handler := &Handler{
		voucherService:  mockVoucherService,
		authzService:    mockAuthzService,
		merchantService: mockMerchantService,
		favoriteService: mockFavoriteService,
		// shareService not needed for non-owner
	}

	// Test data
	userID := uuid.New()
	voucherID := uuid.New()
	ownerID := uuid.New()
	user := &models.User{
		ID:    userID,
		Email: "shared@example.com",
	}
	voucher := &models.Voucher{
		ID:     voucherID,
		UserID: &ownerID,
		Code:   "TEST123",
	}
	perms := &services.ResourcePermissions{
		IsOwner:   false,
		CanView:   true,
		CanEdit:   false, // Vouchers are always read-only for shared users
		CanDelete: false,
	}
	merchants := []models.Merchant{
		{ID: uuid.New(), Name: "Test Merchant"},
	}

	// Mock expectations
	mockAuthzService.On("CheckVoucherAccess", mock.Anything, userID, voucherID).Return(perms, nil)
	mockVoucherService.On("GetVoucher", mock.Anything, voucherID).Return(voucher, nil)
	mockMerchantService.On("GetAllMerchants", mock.Anything).Return(merchants, nil)
	mockFavoriteService.On("IsFavorite", mock.Anything, userID, "voucher", voucherID).Return(false, nil)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/vouchers/"+voucherID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(voucherID.String())
	c.Set("current_user", user)
	c.Set("csrf", "test-token")
	setupI18nContext(c)

	// Execute
	err := handler.Show(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockAuthzService.AssertExpectations(t)
	mockVoucherService.AssertExpectations(t)
	mockMerchantService.AssertExpectations(t)
	mockFavoriteService.AssertExpectations(t)
}

func TestShowHandler_Forbidden(t *testing.T) {
	// Setup
	e := echo.New()
	mockVoucherService := new(MockVoucherService)
	mockAuthzService := new(MockAuthzService)

	handler := &Handler{
		voucherService: mockVoucherService,
		authzService:   mockAuthzService,
		db:             nil,
	}

	// Test data
	userID := uuid.New()
	voucherID := uuid.New()
	user := &models.User{
		ID:    userID,
		Email: "unauthorized@example.com",
	}

	// Mock expectations - return forbidden error
	mockAuthzService.On("CheckVoucherAccess", mock.Anything, userID, voucherID).Return(nil, services.ErrForbidden)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/vouchers/"+voucherID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(voucherID.String())
	c.Set("current_user", user)
	setupI18nContext(c)

	// Execute
	err := handler.Show(c)

	// Assert - should redirect to /vouchers
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/vouchers", rec.Header().Get("Location"))
	mockAuthzService.AssertExpectations(t)
}

func TestShowHandler_VoucherNotFound(t *testing.T) {
	// Setup
	e := echo.New()
	mockVoucherService := new(MockVoucherService)
	mockAuthzService := new(MockAuthzService)

	handler := &Handler{
		voucherService: mockVoucherService,
		authzService:   mockAuthzService,
	}

	// Test data
	userID := uuid.New()
	voucherID := uuid.New()
	user := &models.User{
		ID:    userID,
		Email: "test@example.com",
	}
	perms := &services.ResourcePermissions{
		IsOwner:   true,
		CanView:   true,
		CanEdit:   true,
		CanDelete: true,
	}

	// Mock expectations
	mockAuthzService.On("CheckVoucherAccess", mock.Anything, userID, voucherID).Return(perms, nil)
	mockVoucherService.On("GetVoucher", mock.Anything, voucherID).Return(nil, errors.New("not found"))

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/vouchers/"+voucherID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(voucherID.String())
	c.Set("current_user", user)
	setupI18nContext(c)

	// Execute
	err := handler.Show(c)

	// Assert - should redirect to /vouchers
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/vouchers", rec.Header().Get("Location"))
	mockAuthzService.AssertExpectations(t)
	mockVoucherService.AssertExpectations(t)
}

func TestShowHandler_InvalidVoucherID(t *testing.T) {
	// Setup
	e := echo.New()
	mockVoucherService := new(MockVoucherService)
	mockAuthzService := new(MockAuthzService)

	handler := &Handler{
		voucherService: mockVoucherService,
		authzService:   mockAuthzService,
		db:             nil,
	}

	// Test data
	user := &models.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}

	// Create request with invalid UUID
	req := httptest.NewRequest(http.MethodGet, "/vouchers/invalid-uuid", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("invalid-uuid")
	c.Set("current_user", user)
	setupI18nContext(c)

	// Execute
	err := handler.Show(c)

	// Assert - should redirect to /vouchers
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/vouchers", rec.Header().Get("Location"))
	mockAuthzService.AssertNotCalled(t, "CheckVoucherAccess")
}
