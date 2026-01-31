package giftcards

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

// MockGiftCardService is a manual mock for GiftCardServiceInterface
type MockGiftCardService struct {
	mock.Mock
}

func (m *MockGiftCardService) CreateGiftCard(ctx context.Context, giftCard *models.GiftCard) error {
	args := m.Called(ctx, giftCard)
	return args.Error(0)
}

func (m *MockGiftCardService) GetGiftCard(ctx context.Context, id uuid.UUID) (*models.GiftCard, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GiftCard), args.Error(1)
}

func (m *MockGiftCardService) GetUserGiftCards(ctx context.Context, userID uuid.UUID) ([]models.GiftCard, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.GiftCard), args.Error(1)
}

func (m *MockGiftCardService) UpdateGiftCard(ctx context.Context, giftCard *models.GiftCard) error {
	args := m.Called(ctx, giftCard)
	return args.Error(0)
}

func (m *MockGiftCardService) DeleteGiftCard(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockGiftCardService) CountUserGiftCards(ctx context.Context, userID uuid.UUID) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockGiftCardService) GetTotalBalance(ctx context.Context, userID uuid.UUID) (float64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockGiftCardService) GetCurrentBalance(ctx context.Context, giftCardID uuid.UUID) (float64, error) {
	args := m.Called(ctx, giftCardID)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockGiftCardService) CanUserAccessGiftCard(ctx context.Context, giftCardID, userID uuid.UUID) (bool, error) {
	args := m.Called(ctx, giftCardID, userID)
	return args.Bool(0), args.Error(1)
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
	mockGiftCardService := new(MockGiftCardService)
	mockAuthzService := new(MockAuthzService)
	mockShareService := new(MockShareService)
	mockMerchantService := new(MockMerchantService)
	mockFavoriteService := new(MockFavoriteService)

	handler := &Handler{
		giftCardService: mockGiftCardService,
		authzService:    mockAuthzService,
		shareService:    mockShareService,
		merchantService: mockMerchantService,
		favoriteService: mockFavoriteService,
	}

	// Test data
	userID := uuid.New()
	giftCardID := uuid.New()
	user := &models.User{
		ID:    userID,
		Email: "test@example.com",
	}
	giftCard := &models.GiftCard{
		ID:         giftCardID,
		UserID:     &userID,
		CardNumber: "GC123",
	}
	perms := &services.ResourcePermissions{
		IsOwner:             true,
		CanView:             true,
		CanEdit:             true,
		CanDelete:           true,
		CanEditTransactions: true,
	}
	merchants := []models.Merchant{
		{ID: uuid.New(), Name: "Test Merchant"},
	}
	shares := []models.GiftCardShare{}

	// Mock expectations
	mockAuthzService.On("CheckGiftCardAccess", mock.Anything, userID, giftCardID).Return(perms, nil)
	mockGiftCardService.On("GetGiftCard", mock.Anything, giftCardID).Return(giftCard, nil)
	mockShareService.On("GetGiftCardShares", mock.Anything, giftCardID).Return(shares, nil)
	mockMerchantService.On("GetAllMerchants", mock.Anything).Return(merchants, nil)
	mockFavoriteService.On("IsFavorite", mock.Anything, userID, "gift_card", giftCardID).Return(false, nil)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/gift-cards/"+giftCardID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(giftCardID.String())
	c.Set("current_user", user)
	c.Set("csrf", "test-token")
	setupI18nContext(c)

	// Execute
	err := handler.Show(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockAuthzService.AssertExpectations(t)
	mockGiftCardService.AssertExpectations(t)
	mockShareService.AssertExpectations(t)
	mockMerchantService.AssertExpectations(t)
	mockFavoriteService.AssertExpectations(t)
}

func TestShowHandler_SharedUser(t *testing.T) {
	// Setup
	e := echo.New()
	mockGiftCardService := new(MockGiftCardService)
	mockAuthzService := new(MockAuthzService)
	mockMerchantService := new(MockMerchantService)
	mockFavoriteService := new(MockFavoriteService)

	handler := &Handler{
		giftCardService: mockGiftCardService,
		authzService:    mockAuthzService,
		merchantService: mockMerchantService,
		favoriteService: mockFavoriteService,
		// shareService not needed for non-owner
	}

	// Test data
	userID := uuid.New()
	giftCardID := uuid.New()
	ownerID := uuid.New()
	user := &models.User{
		ID:    userID,
		Email: "shared@example.com",
	}
	giftCard := &models.GiftCard{
		ID:         giftCardID,
		UserID:     &ownerID,
		CardNumber: "GC123",
	}
	perms := &services.ResourcePermissions{
		IsOwner:             false,
		CanView:             true,
		CanEdit:             true,
		CanDelete:           false,
		CanEditTransactions: true,
	}
	merchants := []models.Merchant{
		{ID: uuid.New(), Name: "Test Merchant"},
	}

	// Mock expectations
	mockAuthzService.On("CheckGiftCardAccess", mock.Anything, userID, giftCardID).Return(perms, nil)
	mockGiftCardService.On("GetGiftCard", mock.Anything, giftCardID).Return(giftCard, nil)
	mockMerchantService.On("GetAllMerchants", mock.Anything).Return(merchants, nil)
	mockFavoriteService.On("IsFavorite", mock.Anything, userID, "gift_card", giftCardID).Return(false, nil)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/gift-cards/"+giftCardID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(giftCardID.String())
	c.Set("current_user", user)
	c.Set("csrf", "test-token")
	setupI18nContext(c)

	// Execute
	err := handler.Show(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockAuthzService.AssertExpectations(t)
	mockGiftCardService.AssertExpectations(t)
	mockMerchantService.AssertExpectations(t)
	mockFavoriteService.AssertExpectations(t)
}

func TestShowHandler_Forbidden(t *testing.T) {
	// Setup
	e := echo.New()
	mockGiftCardService := new(MockGiftCardService)
	mockAuthzService := new(MockAuthzService)

	handler := &Handler{
		giftCardService: mockGiftCardService,
		authzService:    mockAuthzService,
		db:              nil,
	}

	// Test data
	userID := uuid.New()
	giftCardID := uuid.New()
	user := &models.User{
		ID:    userID,
		Email: "unauthorized@example.com",
	}

	// Mock expectations - return forbidden error
	mockAuthzService.On("CheckGiftCardAccess", mock.Anything, userID, giftCardID).Return(nil, services.ErrForbidden)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/gift-cards/"+giftCardID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(giftCardID.String())
	c.Set("current_user", user)
	setupI18nContext(c)

	// Execute
	err := handler.Show(c)

	// Assert - should redirect to /gift-cards
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/gift-cards", rec.Header().Get("Location"))
	mockAuthzService.AssertExpectations(t)
}

func TestShowHandler_GiftCardNotFound(t *testing.T) {
	// Setup
	e := echo.New()
	mockGiftCardService := new(MockGiftCardService)
	mockAuthzService := new(MockAuthzService)

	handler := &Handler{
		giftCardService: mockGiftCardService,
		authzService:    mockAuthzService,
	}

	// Test data
	userID := uuid.New()
	giftCardID := uuid.New()
	user := &models.User{
		ID:    userID,
		Email: "test@example.com",
	}
	perms := &services.ResourcePermissions{
		IsOwner:             true,
		CanView:             true,
		CanEdit:             true,
		CanDelete:           true,
		CanEditTransactions: true,
	}

	// Mock expectations
	mockAuthzService.On("CheckGiftCardAccess", mock.Anything, userID, giftCardID).Return(perms, nil)
	mockGiftCardService.On("GetGiftCard", mock.Anything, giftCardID).Return(nil, errors.New("not found"))

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/gift-cards/"+giftCardID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(giftCardID.String())
	c.Set("current_user", user)
	setupI18nContext(c)

	// Execute
	err := handler.Show(c)

	// Assert - should redirect to /gift-cards
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/gift-cards", rec.Header().Get("Location"))
	mockAuthzService.AssertExpectations(t)
	mockGiftCardService.AssertExpectations(t)
}

func TestShowHandler_InvalidGiftCardID(t *testing.T) {
	// Setup
	e := echo.New()
	mockGiftCardService := new(MockGiftCardService)
	mockAuthzService := new(MockAuthzService)

	handler := &Handler{
		giftCardService: mockGiftCardService,
		authzService:    mockAuthzService,
		db:              nil,
	}

	// Test data
	user := &models.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}

	// Create request with invalid UUID
	req := httptest.NewRequest(http.MethodGet, "/gift-cards/invalid-uuid", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("invalid-uuid")
	c.Set("current_user", user)
	setupI18nContext(c)

	// Execute
	err := handler.Show(c)

	// Assert - should redirect to /gift-cards
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/gift-cards", rec.Header().Get("Location"))
	mockAuthzService.AssertNotCalled(t, "CheckGiftCardAccess")
}
