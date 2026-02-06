package vouchers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	savvyi18n "savvy/internal/i18n"
	"savvy/internal/models"
)

func TestNewHandler_GetForm(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/vouchers/new", nil)
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
	mockMerchantService := new(MockMerchantService)
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
	req := httptest.NewRequest(http.MethodGet, "/vouchers/new", nil)
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
	mockMerchantService := new(MockMerchantService)
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

func TestNewHandler_ServiceError(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/vouchers/new", nil)
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
	mockMerchantService := new(MockMerchantService)
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
