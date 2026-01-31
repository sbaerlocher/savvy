package vouchers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"savvy/internal/models"
)

func TestIndexHandler_Success(t *testing.T) {
	// Setup
	e := echo.New()
	userID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/vouchers", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Setup user context
	user := &models.User{
		ID:    userID,
		Email: "test@example.com",
	}
	c.Set("current_user", user)
	setupI18nContext(c)

	// Mock voucher service
	mockVoucherService := new(MockVoucherService)
	vouchers := []models.Voucher{
		{
			ID:           uuid.New(),
			UserID:       &userID,
			Code:         "SAVE20",
			MerchantName: "Test Merchant 1",
			Type:         "percentage",
			Value:        20.0,
			ValidFrom:    time.Now().Add(-24 * time.Hour),
			ValidUntil:   time.Now().Add(30 * 24 * time.Hour),
		},
		{
			ID:           uuid.New(),
			UserID:       &userID,
			Code:         "WELCOME10",
			MerchantName: "Test Merchant 2",
			Type:         "fixed",
			Value:        10.0,
			ValidFrom:    time.Now().Add(-24 * time.Hour),
			ValidUntil:   time.Now().Add(60 * 24 * time.Hour),
		},
	}
	mockVoucherService.On("GetUserVouchers", mock.Anything, userID).Return(vouchers, nil)

	// Create handler with mock
	handler := &Handler{
		voucherService: mockVoucherService,
	}

	// Execute
	err := handler.Index(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockVoucherService.AssertExpectations(t)
}

func TestIndexHandler_EmptyList(t *testing.T) {
	// Setup
	e := echo.New()
	userID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/vouchers", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Setup user context
	user := &models.User{
		ID:    userID,
		Email: "test@example.com",
	}
	c.Set("current_user", user)
	setupI18nContext(c)

	// Mock voucher service - return empty list
	mockVoucherService := new(MockVoucherService)
	emptyVouchers := []models.Voucher{}
	mockVoucherService.On("GetUserVouchers", mock.Anything, userID).Return(emptyVouchers, nil)

	// Create handler with mock
	handler := &Handler{
		voucherService: mockVoucherService,
	}

	// Execute
	err := handler.Index(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockVoucherService.AssertExpectations(t)
}

func TestIndexHandler_ServiceError(t *testing.T) {
	// Setup
	e := echo.New()
	userID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/vouchers", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Setup user context
	user := &models.User{
		ID:    userID,
		Email: "test@example.com",
	}
	c.Set("current_user", user)
	setupI18nContext(c)

	// Mock voucher service - return error
	mockVoucherService := new(MockVoucherService)
	mockVoucherService.On("GetUserVouchers", mock.Anything, userID).Return(nil, errors.New("database error"))

	// Create handler with mock
	handler := &Handler{
		voucherService: mockVoucherService,
	}

	// Execute
	err := handler.Index(c)

	// Assert - error should be returned
	assert.Error(t, err)
	mockVoucherService.AssertExpectations(t)
}

func TestIndexHandler_WithImpersonation(t *testing.T) {
	// Setup
	e := echo.New()
	userID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/vouchers", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Setup user context with impersonation flag
	user := &models.User{
		ID:    userID,
		Email: "impersonated@example.com",
	}
	c.Set("current_user", user)
	c.Set("is_impersonating", true)
	setupI18nContext(c)

	// Mock voucher service
	mockVoucherService := new(MockVoucherService)
	vouchers := []models.Voucher{
		{
			ID:           uuid.New(),
			UserID:       &userID,
			Code:         "DISCOUNT",
			MerchantName: "Test Merchant",
			Type:         "percentage",
			Value:        15.0,
			ValidFrom:    time.Now().Add(-24 * time.Hour),
			ValidUntil:   time.Now().Add(30 * 24 * time.Hour),
		},
	}
	mockVoucherService.On("GetUserVouchers", mock.Anything, userID).Return(vouchers, nil)

	// Create handler with mock
	handler := &Handler{
		voucherService: mockVoucherService,
	}

	// Execute
	err := handler.Index(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockVoucherService.AssertExpectations(t)
}
