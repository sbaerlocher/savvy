package cards

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"savvy/internal/models"
)

func TestIndexHandler_Success(t *testing.T) {
	// Setup
	e := echo.New()
	userID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/cards", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Setup user context
	user := &models.User{
		ID:    userID,
		Email: "test@example.com",
	}
	c.Set("current_user", user)
	setupI18nContext(c)

	// Mock card service
	mockCardService := new(MockCardService)
	cards := []models.Card{
		{
			ID:           uuid.New(),
			UserID:       &userID,
			CardNumber:   "1234567890",
			MerchantName: "Test Merchant 1",
		},
		{
			ID:           uuid.New(),
			UserID:       &userID,
			CardNumber:   "0987654321",
			MerchantName: "Test Merchant 2",
		},
	}
	mockCardService.On("GetUserCards", mock.Anything, userID).Return(cards, nil)

	// Create handler with mock
	handler := &Handler{
		cardService: mockCardService,
	}

	// Execute
	err := handler.Index(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockCardService.AssertExpectations(t)
}

func TestIndexHandler_EmptyList(t *testing.T) {
	// Setup
	e := echo.New()
	userID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/cards", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Setup user context
	user := &models.User{
		ID:    userID,
		Email: "test@example.com",
	}
	c.Set("current_user", user)
	setupI18nContext(c)

	// Mock card service - return empty list
	mockCardService := new(MockCardService)
	emptyCards := []models.Card{}
	mockCardService.On("GetUserCards", mock.Anything, userID).Return(emptyCards, nil)

	// Create handler with mock
	handler := &Handler{
		cardService: mockCardService,
	}

	// Execute
	err := handler.Index(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockCardService.AssertExpectations(t)
}

func TestIndexHandler_ServiceError(t *testing.T) {
	// Setup
	e := echo.New()
	userID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/cards", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Setup user context
	user := &models.User{
		ID:    userID,
		Email: "test@example.com",
	}
	c.Set("current_user", user)
	setupI18nContext(c)

	// Mock card service - return error
	mockCardService := new(MockCardService)
	mockCardService.On("GetUserCards", mock.Anything, userID).Return(nil, errors.New("database error"))

	// Create handler with mock
	handler := &Handler{
		cardService: mockCardService,
	}

	// Execute
	err := handler.Index(c)

	// Assert - error should be returned
	assert.Error(t, err)
	mockCardService.AssertExpectations(t)
}

func TestIndexHandler_WithImpersonation(t *testing.T) {
	// Setup
	e := echo.New()
	userID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/cards", nil)
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

	// Mock card service
	mockCardService := new(MockCardService)
	cards := []models.Card{
		{
			ID:           uuid.New(),
			UserID:       &userID,
			CardNumber:   "1234567890",
			MerchantName: "Test Merchant",
		},
	}
	mockCardService.On("GetUserCards", mock.Anything, userID).Return(cards, nil)

	// Create handler with mock
	handler := &Handler{
		cardService: mockCardService,
	}

	// Execute
	err := handler.Index(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockCardService.AssertExpectations(t)
}
