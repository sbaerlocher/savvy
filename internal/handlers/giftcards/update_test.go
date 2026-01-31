package giftcards

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"savvy/internal/models"
	"savvy/internal/services"
)

func TestUpdateHandler_Owner(t *testing.T) {
	// Setup
	e := echo.New()
	userID := uuid.New()
	giftCardID := uuid.New()
	merchantID := uuid.New()

	// Form data
	formData := url.Values{}
	formData.Set("merchant_id", merchantID.String())
	formData.Set("card_number", "GC9999999999")
	formData.Set("initial_balance", "150.00")
	formData.Set("currency", "EUR")
	formData.Set("pin", "9999")
	formData.Set("expires_at", "2028-12-31")
	formData.Set("status", "active")
	formData.Set("barcode_type", "QR")
	formData.Set("notes", "Updated notes")

	req := httptest.NewRequest(http.MethodPost, "/gift-cards/"+giftCardID.String(), strings.NewReader(formData.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(giftCardID.String())

	// Setup user context
	user := &models.User{
		ID:    userID,
		Email: "owner@example.com",
	}
	c.Set("current_user", user)
	c.Set("csrf", "test-csrf-token")
	setupI18nContext(c)

	// Mock services
	mockAuthzService := new(MockAuthzService)
	mockGiftCardService := new(MockGiftCardService)
	mockMerchantService := new(MockMerchantService)

	perms := &services.ResourcePermissions{
		CanView:             true,
		CanEdit:             true,
		CanDelete:           true,
		CanEditTransactions: true,
		IsOwner:             true,
	}
	mockAuthzService.On("CheckGiftCardAccess", mock.Anything, userID, giftCardID).Return(perms, nil)

	expiresAt := time.Date(2027, 12, 31, 23, 59, 59, 0, time.UTC)
	giftCard := &models.GiftCard{
		ID:             giftCardID,
		UserID:         &userID,
		CardNumber:     "GC1234567890",
		InitialBalance: 100.00,
		Currency:       "CHF",
		ExpiresAt:      &expiresAt,
	}
	mockGiftCardService.On("GetGiftCard", mock.Anything, giftCardID).Return(giftCard, nil)

	merchant := &models.Merchant{
		ID:   merchantID,
		Name: "Updated Merchant",
	}
	mockMerchantService.On("GetMerchantByID", mock.Anything, merchantID).Return(merchant, nil)

	mockGiftCardService.On("UpdateGiftCard", mock.Anything, mock.AnythingOfType("*models.GiftCard")).Return(nil)

	// Create handler with mocks
	handler := &Handler{
		authzService:    mockAuthzService,
		giftCardService: mockGiftCardService,
		merchantService: mockMerchantService,
	}

	// Execute
	err := handler.Update(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/gift-cards/"+giftCardID.String(), rec.Header().Get("Location"))
	mockAuthzService.AssertExpectations(t)
	mockGiftCardService.AssertExpectations(t)
	mockMerchantService.AssertExpectations(t)
}

func TestUpdateHandler_Forbidden(t *testing.T) {
	// Setup
	e := echo.New()
	userID := uuid.New()
	giftCardID := uuid.New()

	// Form data
	formData := url.Values{}
	formData.Set("card_number", "GC9999999999")

	req := httptest.NewRequest(http.MethodPost, "/gift-cards/"+giftCardID.String(), strings.NewReader(formData.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(giftCardID.String())

	// Setup user context
	user := &models.User{
		ID:    userID,
		Email: "notowner@example.com",
	}
	c.Set("current_user", user)
	setupI18nContext(c)

	// Mock services - no edit permission
	mockAuthzService := new(MockAuthzService)
	perms := &services.ResourcePermissions{
		CanView:             true,
		CanEdit:             false, // Not allowed to edit
		CanDelete:           false,
		CanEditTransactions: false,
		IsOwner:             false,
	}
	mockAuthzService.On("CheckGiftCardAccess", mock.Anything, userID, giftCardID).Return(perms, nil)

	// Create handler with mock
	handler := &Handler{
		authzService: mockAuthzService,
	}

	// Execute
	err := handler.Update(c)

	// Assert - should redirect
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/gift-cards", rec.Header().Get("Location"))
	mockAuthzService.AssertExpectations(t)
}

func TestUpdateHandler_InvalidDate(t *testing.T) {
	// Setup
	e := echo.New()
	userID := uuid.New()
	giftCardID := uuid.New()

	// Invalid form data (invalid expiration date)
	formData := url.Values{}
	formData.Set("card_number", "GC123")
	formData.Set("initial_balance", "50.00")
	formData.Set("expires_at", "not-a-date")

	req := httptest.NewRequest(http.MethodPost, "/gift-cards/"+giftCardID.String(), strings.NewReader(formData.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(giftCardID.String())

	// Setup user context
	user := &models.User{
		ID:    userID,
		Email: "owner@example.com",
	}
	c.Set("current_user", user)
	c.Set("csrf", "test-csrf-token")
	setupI18nContext(c)

	// Mock services
	mockAuthzService := new(MockAuthzService)
	mockGiftCardService := new(MockGiftCardService)

	perms := &services.ResourcePermissions{
		CanView:             true,
		CanEdit:             true,
		CanDelete:           true,
		CanEditTransactions: true,
		IsOwner:             true,
	}
	mockAuthzService.On("CheckGiftCardAccess", mock.Anything, userID, giftCardID).Return(perms, nil)

	expiresAt := time.Date(2027, 12, 31, 23, 59, 59, 0, time.UTC)
	giftCard := &models.GiftCard{
		ID:             giftCardID,
		UserID:         &userID,
		CardNumber:     "GC123",
		InitialBalance: 50.00,
		ExpiresAt:      &expiresAt,
	}
	mockGiftCardService.On("GetGiftCard", mock.Anything, giftCardID).Return(giftCard, nil)

	// Create handler with mocks
	handler := &Handler{
		authzService:    mockAuthzService,
		giftCardService: mockGiftCardService,
	}

	// Execute
	err := handler.Update(c)

	// Assert - should redirect back to edit form with error
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Contains(t, rec.Header().Get("Location"), "/gift-cards/"+giftCardID.String()+"/edit?error=invalid_date")
	mockAuthzService.AssertExpectations(t)
	mockGiftCardService.AssertExpectations(t)
}

func TestUpdateHandler_ServiceError(t *testing.T) {
	// Setup
	e := echo.New()
	userID := uuid.New()
	giftCardID := uuid.New()

	// Valid form data
	formData := url.Values{}
	formData.Set("card_number", "GC9999")
	formData.Set("initial_balance", "99.99")

	req := httptest.NewRequest(http.MethodPost, "/gift-cards/"+giftCardID.String(), strings.NewReader(formData.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(giftCardID.String())

	// Setup user context
	user := &models.User{
		ID:    userID,
		Email: "owner@example.com",
	}
	c.Set("current_user", user)
	c.Set("csrf", "test-csrf-token")
	setupI18nContext(c)

	// Mock services
	mockAuthzService := new(MockAuthzService)
	mockGiftCardService := new(MockGiftCardService)

	perms := &services.ResourcePermissions{
		CanView:             true,
		CanEdit:             true,
		CanDelete:           true,
		CanEditTransactions: true,
		IsOwner:             true,
	}
	mockAuthzService.On("CheckGiftCardAccess", mock.Anything, userID, giftCardID).Return(perms, nil)

	giftCard := &models.GiftCard{
		ID:             giftCardID,
		UserID:         &userID,
		CardNumber:     "GC123",
		InitialBalance: 50.00,
	}
	mockGiftCardService.On("GetGiftCard", mock.Anything, giftCardID).Return(giftCard, nil)

	// Update fails with service error
	mockGiftCardService.On("UpdateGiftCard", mock.Anything, mock.AnythingOfType("*models.GiftCard")).Return(errors.New("database error"))

	// Create handler with mocks
	handler := &Handler{
		authzService:    mockAuthzService,
		giftCardService: mockGiftCardService,
	}

	// Execute
	err := handler.Update(c)

	// Assert - should redirect back to edit form
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/gift-cards/"+giftCardID.String()+"/edit", rec.Header().Get("Location"))
	mockAuthzService.AssertExpectations(t)
	mockGiftCardService.AssertExpectations(t)
}

func TestUpdateHandler_NotFound(t *testing.T) {
	// Setup
	e := echo.New()
	userID := uuid.New()
	giftCardID := uuid.New()

	// Form data
	formData := url.Values{}
	formData.Set("card_number", "NOTFOUND")

	req := httptest.NewRequest(http.MethodPost, "/gift-cards/"+giftCardID.String(), strings.NewReader(formData.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(giftCardID.String())

	// Setup user context
	user := &models.User{
		ID:    userID,
		Email: "owner@example.com",
	}
	c.Set("current_user", user)
	setupI18nContext(c)

	// Mock services
	mockAuthzService := new(MockAuthzService)
	mockGiftCardService := new(MockGiftCardService)

	perms := &services.ResourcePermissions{
		CanView:             true,
		CanEdit:             true,
		CanDelete:           true,
		CanEditTransactions: true,
		IsOwner:             true,
	}
	mockAuthzService.On("CheckGiftCardAccess", mock.Anything, userID, giftCardID).Return(perms, nil)

	// Gift card not found
	mockGiftCardService.On("GetGiftCard", mock.Anything, giftCardID).Return(nil, errors.New("not found"))

	// Create handler with mocks
	handler := &Handler{
		authzService:    mockAuthzService,
		giftCardService: mockGiftCardService,
	}

	// Execute
	err := handler.Update(c)

	// Assert - should redirect to gift cards list
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/gift-cards", rec.Header().Get("Location"))
	mockAuthzService.AssertExpectations(t)
	mockGiftCardService.AssertExpectations(t)
}
