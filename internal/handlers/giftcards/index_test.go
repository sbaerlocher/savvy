package giftcards

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

	req := httptest.NewRequest(http.MethodGet, "/gift-cards", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Setup user context
	user := &models.User{
		ID:    userID,
		Email: "test@example.com",
	}
	c.Set("current_user", user)
	setupI18nContext(c)

	// Mock gift card service
	mockGiftCardService := new(MockGiftCardService)
	expiresAt := time.Date(2027, 12, 31, 23, 59, 59, 0, time.UTC)
	giftCards := []models.GiftCard{
		{
			ID:             uuid.New(),
			UserID:         &userID,
			CardNumber:     "GC1234567890",
			MerchantName:   "Test Merchant 1",
			InitialBalance: 100.00,
			Currency:       "CHF",
			ExpiresAt:      &expiresAt,
		},
		{
			ID:             uuid.New(),
			UserID:         &userID,
			CardNumber:     "GC0987654321",
			MerchantName:   "Test Merchant 2",
			InitialBalance: 50.00,
			Currency:       "EUR",
			ExpiresAt:      &expiresAt,
		},
	}
	mockGiftCardService.On("GetUserGiftCards", mock.Anything, userID).Return(giftCards, nil)

	// Create handler with mock
	handler := &Handler{
		giftCardService: mockGiftCardService,
	}

	// Execute
	err := handler.Index(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockGiftCardService.AssertExpectations(t)
}

func TestIndexHandler_EmptyList(t *testing.T) {
	// Setup
	e := echo.New()
	userID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/gift-cards", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Setup user context
	user := &models.User{
		ID:    userID,
		Email: "test@example.com",
	}
	c.Set("current_user", user)
	setupI18nContext(c)

	// Mock gift card service - return empty list
	mockGiftCardService := new(MockGiftCardService)
	emptyGiftCards := []models.GiftCard{}
	mockGiftCardService.On("GetUserGiftCards", mock.Anything, userID).Return(emptyGiftCards, nil)

	// Create handler with mock
	handler := &Handler{
		giftCardService: mockGiftCardService,
	}

	// Execute
	err := handler.Index(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockGiftCardService.AssertExpectations(t)
}

func TestIndexHandler_ServiceError(t *testing.T) {
	// Setup
	e := echo.New()
	userID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/gift-cards", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Setup user context
	user := &models.User{
		ID:    userID,
		Email: "test@example.com",
	}
	c.Set("current_user", user)
	setupI18nContext(c)

	// Mock gift card service - return error
	mockGiftCardService := new(MockGiftCardService)
	mockGiftCardService.On("GetUserGiftCards", mock.Anything, userID).Return(nil, errors.New("database error"))

	// Create handler with mock
	handler := &Handler{
		giftCardService: mockGiftCardService,
	}

	// Execute
	err := handler.Index(c)

	// Assert - error should be returned
	assert.Error(t, err)
	mockGiftCardService.AssertExpectations(t)
}

func TestIndexHandler_WithImpersonation(t *testing.T) {
	// Setup
	e := echo.New()
	userID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/gift-cards", nil)
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

	// Mock gift card service
	mockGiftCardService := new(MockGiftCardService)
	expiresAt := time.Date(2028, 6, 30, 23, 59, 59, 0, time.UTC)
	giftCards := []models.GiftCard{
		{
			ID:             uuid.New(),
			UserID:         &userID,
			CardNumber:     "GC555",
			MerchantName:   "Test Store",
			InitialBalance: 75.00,
			Currency:       "USD",
			ExpiresAt:      &expiresAt,
		},
	}
	mockGiftCardService.On("GetUserGiftCards", mock.Anything, userID).Return(giftCards, nil)

	// Create handler with mock
	handler := &Handler{
		giftCardService: mockGiftCardService,
	}

	// Execute
	err := handler.Index(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockGiftCardService.AssertExpectations(t)
}
