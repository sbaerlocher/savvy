package giftcards

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"savvy/internal/models"
)

func TestCreateHandler_Success(t *testing.T) {
	// Setup
	e := echo.New()
	userID := uuid.New()
	merchantID := uuid.New()

	// Form data
	formData := url.Values{}
	formData.Set("merchant_id", merchantID.String())
	formData.Set("card_number", "GC1234567890")
	formData.Set("initial_balance", "100.00")
	formData.Set("currency", "CHF")
	formData.Set("pin", "1234")
	formData.Set("expires_at", "2027-12-31")
	formData.Set("barcode_type", "CODE128")
	formData.Set("notes", "Test gift card")

	req := httptest.NewRequest(http.MethodPost, "/gift-cards", strings.NewReader(formData.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Setup user context
	user := &models.User{
		ID:    userID,
		Email: "test@example.com",
	}
	c.Set("current_user", user)
	c.Set("csrf", "test-csrf-token")
	setupI18nContext(c)

	// Mock services
	mockGiftCardService := new(MockGiftCardService)
	mockMerchantService := new(MockMerchantService)

	merchant := &models.Merchant{
		ID:   merchantID,
		Name: "Test Merchant",
	}
	mockMerchantService.On("GetMerchantByID", mock.Anything, merchantID).Return(merchant, nil)
	mockGiftCardService.On("CreateGiftCard", mock.Anything, mock.AnythingOfType("*models.GiftCard")).Return(nil)

	// Create handler with mocks
	handler := &Handler{
		giftCardService: mockGiftCardService,
		merchantService: mockMerchantService,
	}

	// Execute
	err := handler.Create(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Contains(t, rec.Header().Get("Location"), "/gift-cards/")
	mockGiftCardService.AssertExpectations(t)
	mockMerchantService.AssertExpectations(t)
}

func TestCreateHandler_ValidationError_MissingCardNumber(t *testing.T) {
	// Setup
	e := echo.New()
	userID := uuid.New()

	// Form data without card_number
	formData := url.Values{}
	formData.Set("merchant_name", "Test Merchant")
	formData.Set("initial_balance", "50.00")
	formData.Set("currency", "CHF")

	req := httptest.NewRequest(http.MethodPost, "/gift-cards", strings.NewReader(formData.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Setup user context
	user := &models.User{
		ID:    userID,
		Email: "test@example.com",
	}
	c.Set("current_user", user)
	c.Set("csrf", "test-csrf-token")
	setupI18nContext(c)

	// Mock services
	mockGiftCardService := new(MockGiftCardService)
	// CreateGiftCard will be called but might fail validation
	mockGiftCardService.On("CreateGiftCard", mock.Anything, mock.AnythingOfType("*models.GiftCard")).Return(errors.New("validation error: card_number required"))

	// Create handler with mocks
	handler := &Handler{
		giftCardService: mockGiftCardService,
	}

	// Execute
	err := handler.Create(c)

	// Assert - should redirect back to new form with error
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Contains(t, rec.Header().Get("Location"), "/gift-cards/new")
	mockGiftCardService.AssertExpectations(t)
}

func TestCreateHandler_InvalidDate(t *testing.T) {
	// Setup
	e := echo.New()
	userID := uuid.New()

	// Form data with invalid expiration date
	formData := url.Values{}
	formData.Set("merchant_name", "Test Merchant")
	formData.Set("card_number", "GC123")
	formData.Set("initial_balance", "50.00")
	formData.Set("currency", "CHF")
	formData.Set("expires_at", "invalid-date")

	req := httptest.NewRequest(http.MethodPost, "/gift-cards", strings.NewReader(formData.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Setup user context
	user := &models.User{
		ID:    userID,
		Email: "test@example.com",
	}
	c.Set("current_user", user)
	c.Set("csrf", "test-csrf-token")
	setupI18nContext(c)

	// Create handler (no mocks needed - validation happens before service call)
	handler := &Handler{}

	// Execute
	err := handler.Create(c)

	// Assert - should redirect to new form with date error
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Contains(t, rec.Header().Get("Location"), "/gift-cards/new?error=invalid_date")
}

func TestCreateHandler_ServiceError(t *testing.T) {
	// Setup
	e := echo.New()
	userID := uuid.New()

	// Valid form data
	formData := url.Values{}
	formData.Set("merchant_name", "Test Merchant")
	formData.Set("card_number", "GC1234567890")
	formData.Set("initial_balance", "100.00")
	formData.Set("currency", "EUR")

	req := httptest.NewRequest(http.MethodPost, "/gift-cards", strings.NewReader(formData.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Setup user context
	user := &models.User{
		ID:    userID,
		Email: "test@example.com",
	}
	c.Set("current_user", user)
	c.Set("csrf", "test-csrf-token")
	setupI18nContext(c)

	// Mock services - return error
	mockGiftCardService := new(MockGiftCardService)
	mockGiftCardService.On("CreateGiftCard", mock.Anything, mock.AnythingOfType("*models.GiftCard")).Return(errors.New("database error"))

	// Create handler with mocks
	handler := &Handler{
		giftCardService: mockGiftCardService,
	}

	// Execute
	err := handler.Create(c)

	// Assert - should redirect to new form with error
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Contains(t, rec.Header().Get("Location"), "/gift-cards/new?error=")
	mockGiftCardService.AssertExpectations(t)
}

func TestCreateHandler_MerchantNotFound(t *testing.T) {
	// Setup
	e := echo.New()
	userID := uuid.New()
	merchantID := uuid.New()

	// Form data with invalid merchant_id
	formData := url.Values{}
	formData.Set("merchant_id", merchantID.String())
	formData.Set("card_number", "GC999")
	formData.Set("initial_balance", "75.00")

	req := httptest.NewRequest(http.MethodPost, "/gift-cards", strings.NewReader(formData.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Setup user context
	user := &models.User{
		ID:    userID,
		Email: "test@example.com",
	}
	c.Set("current_user", user)
	c.Set("csrf", "test-csrf-token")
	setupI18nContext(c)

	// Mock services
	mockGiftCardService := new(MockGiftCardService)
	mockMerchantService := new(MockMerchantService)

	// Merchant not found
	mockMerchantService.On("GetMerchantByID", mock.Anything, merchantID).Return(nil, errors.New("not found"))
	// Gift card creation should still succeed (without merchant name)
	mockGiftCardService.On("CreateGiftCard", mock.Anything, mock.AnythingOfType("*models.GiftCard")).Return(nil)

	// Create handler with mocks
	handler := &Handler{
		giftCardService: mockGiftCardService,
		merchantService: mockMerchantService,
	}

	// Execute
	err := handler.Create(c)

	// Assert - should still succeed (merchant is optional)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Contains(t, rec.Header().Get("Location"), "/gift-cards/")
	mockGiftCardService.AssertExpectations(t)
	mockMerchantService.AssertExpectations(t)
}

func TestCreateHandler_NewMerchantName(t *testing.T) {
	// Setup
	e := echo.New()
	userID := uuid.New()

	// Form data with new merchant name
	formData := url.Values{}
	formData.Set("merchant_id", "new")
	formData.Set("merchant_name", "Brand New Store")
	formData.Set("card_number", "GC555")
	formData.Set("initial_balance", "200.00")
	formData.Set("currency", "USD")

	req := httptest.NewRequest(http.MethodPost, "/gift-cards", strings.NewReader(formData.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Setup user context
	user := &models.User{
		ID:    userID,
		Email: "test@example.com",
	}
	c.Set("current_user", user)
	c.Set("csrf", "test-csrf-token")
	setupI18nContext(c)

	// Mock services
	mockGiftCardService := new(MockGiftCardService)
	mockGiftCardService.On("CreateGiftCard", mock.Anything, mock.MatchedBy(func(giftCard *models.GiftCard) bool {
		return giftCard.MerchantName == "Brand New Store" && giftCard.MerchantID == nil
	})).Return(nil)

	// Create handler with mocks
	handler := &Handler{
		giftCardService: mockGiftCardService,
	}

	// Execute
	err := handler.Create(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Contains(t, rec.Header().Get("Location"), "/gift-cards/")
	mockGiftCardService.AssertExpectations(t)
}
