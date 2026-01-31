package cards

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	savvyi18n "savvy/internal/i18n"
	"savvy/internal/models"
)

// MockMerchantServiceNew for testing
type MockMerchantServiceNew struct {
	mock.Mock
}

func (m *MockMerchantServiceNew) GetAllMerchants(ctx context.Context) ([]models.Merchant, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.Merchant), args.Error(1)
}

func (m *MockMerchantServiceNew) GetMerchantByID(ctx context.Context, id uuid.UUID) (*models.Merchant, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Merchant), args.Error(1)
}

func (m *MockMerchantServiceNew) CreateMerchant(ctx context.Context, merchant *models.Merchant) error {
	args := m.Called(ctx, merchant)
	return args.Error(0)
}

func (m *MockMerchantServiceNew) UpdateMerchant(ctx context.Context, merchant *models.Merchant) error {
	args := m.Called(ctx, merchant)
	return args.Error(0)
}

func (m *MockMerchantServiceNew) DeleteMerchant(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockMerchantServiceNew) SearchMerchants(ctx context.Context, query string) ([]models.Merchant, error) {
	args := m.Called(ctx, query)
	return args.Get(0).([]models.Merchant), args.Error(1)
}

func (m *MockMerchantServiceNew) GetMerchantCount(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func TestNewHandler_GetForm(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/cards/new", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Setup i18n context
	localizer := savvyi18n.NewLocalizer("de")
	ctx := savvyi18n.SetLocalizer(c.Request().Context(), localizer)
	c.SetRequest(c.Request().WithContext(ctx))

	// Setup user context
	user := &models.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}
	c.Set("current_user", user)
	c.Set("csrf", "test-csrf-token")

	// Mock merchant service
	mockMerchantService := new(MockMerchantServiceNew)
	merchants := []models.Merchant{
		{ID: uuid.New(), Name: "Test Merchant"},
	}
	mockMerchantService.On("GetAllMerchants", mock.Anything).Return(merchants, nil)

	// Create handler with mock
	handler := &Handler{
		merchantService: mockMerchantService,
	}

	// Execute
	err := handler.New(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockMerchantService.AssertExpectations(t)
}

func TestNewHandler_WithoutCSRF(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/cards/new", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Setup i18n context
	localizer := savvyi18n.NewLocalizer("de")
	ctx := savvyi18n.SetLocalizer(c.Request().Context(), localizer)
	c.SetRequest(c.Request().WithContext(ctx))

	// Setup user context (no CSRF)
	user := &models.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}
	c.Set("current_user", user)

	// Mock merchant service
	mockMerchantService := new(MockMerchantServiceNew)
	merchants := []models.Merchant{}
	mockMerchantService.On("GetAllMerchants", mock.Anything).Return(merchants, nil)

	// Create handler with mock
	handler := &Handler{
		merchantService: mockMerchantService,
	}

	// Execute
	err := handler.New(c)

	// Assert - Should still work without CSRF (just empty string)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockMerchantService.AssertExpectations(t)
}

func TestNewHandler_MerchantServiceError(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/cards/new", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Setup i18n context
	localizer := savvyi18n.NewLocalizer("de")
	ctx := savvyi18n.SetLocalizer(c.Request().Context(), localizer)
	c.SetRequest(c.Request().WithContext(ctx))

	// Setup user context
	user := &models.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}
	c.Set("current_user", user)
	c.Set("csrf", "test-csrf-token")

	// Mock merchant service with error
	mockMerchantService := new(MockMerchantServiceNew)
	mockMerchantService.On("GetAllMerchants", mock.Anything).Return([]models.Merchant{}, errors.New("db error"))

	// Create handler with mock
	handler := &Handler{
		merchantService: mockMerchantService,
	}

	// Execute
	err := handler.New(c)

	// Assert - Should handle error gracefully (empty merchants list)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockMerchantService.AssertExpectations(t)
}
