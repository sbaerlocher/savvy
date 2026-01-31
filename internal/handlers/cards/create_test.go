package cards

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
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
	formData.Set("card_number", "1234567890")
	formData.Set("program", "Test Program")
	formData.Set("barcode_type", "CODE128")
	formData.Set("notes", "Test notes")

	req := httptest.NewRequest(http.MethodPost, "/cards", strings.NewReader(formData.Encode()))
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
	mockCardService := new(MockCardService)
	mockMerchantService := new(MockMerchantService)

	merchant := &models.Merchant{
		ID:   merchantID,
		Name: "Test Merchant",
	}
	mockMerchantService.On("GetMerchantByID", mock.Anything, merchantID).Return(merchant, nil)
	mockCardService.On("CreateCard", mock.Anything, mock.AnythingOfType("*models.Card")).Return(nil)

	// Create handler with mocks
	handler := &Handler{
		cardService:     mockCardService,
		merchantService: mockMerchantService,
	}

	// Execute
	err := handler.Create(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Contains(t, rec.Header().Get("Location"), "/cards/")
	mockCardService.AssertExpectations(t)
	mockMerchantService.AssertExpectations(t)
}

func TestCreateHandler_ValidationError_MissingCardNumber(t *testing.T) {
	// Setup
	e := echo.New()
	userID := uuid.New()

	// Form data without card_number
	formData := url.Values{}
	formData.Set("merchant_name", "Test Merchant")
	formData.Set("program", "Test Program")

	req := httptest.NewRequest(http.MethodPost, "/cards", strings.NewReader(formData.Encode()))
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
	mockCardService := new(MockCardService)
	// CreateCard will be called but might fail validation
	mockCardService.On("CreateCard", mock.Anything, mock.AnythingOfType("*models.Card")).Return(errors.New("validation error: card_number required"))

	// Create handler with mocks
	handler := &Handler{
		cardService: mockCardService,
	}

	// Execute
	err := handler.Create(c)

	// Assert - should redirect back to new form with error
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Contains(t, rec.Header().Get("Location"), "/cards/new")
	mockCardService.AssertExpectations(t)
}

func TestCreateHandler_ServiceError(t *testing.T) {
	// Setup
	e := echo.New()
	userID := uuid.New()

	// Valid form data
	formData := url.Values{}
	formData.Set("merchant_name", "Test Merchant")
	formData.Set("card_number", "1234567890")
	formData.Set("program", "Test Program")
	formData.Set("barcode_type", "CODE128")

	req := httptest.NewRequest(http.MethodPost, "/cards", strings.NewReader(formData.Encode()))
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
	mockCardService := new(MockCardService)
	mockCardService.On("CreateCard", mock.Anything, mock.AnythingOfType("*models.Card")).Return(errors.New("database error"))

	// Create handler with mocks
	handler := &Handler{
		cardService: mockCardService,
	}

	// Execute
	err := handler.Create(c)

	// Assert - should redirect to new form with error
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Contains(t, rec.Header().Get("Location"), "/cards/new?error=")
	mockCardService.AssertExpectations(t)
}

func TestCreateHandler_MerchantNotFound(t *testing.T) {
	// Setup
	e := echo.New()
	userID := uuid.New()
	merchantID := uuid.New()

	// Form data with invalid merchant_id
	formData := url.Values{}
	formData.Set("merchant_id", merchantID.String())
	formData.Set("card_number", "1234567890")
	formData.Set("program", "Test Program")

	req := httptest.NewRequest(http.MethodPost, "/cards", strings.NewReader(formData.Encode()))
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
	mockCardService := new(MockCardService)
	mockMerchantService := new(MockMerchantService)

	// Merchant not found
	mockMerchantService.On("GetMerchantByID", mock.Anything, merchantID).Return(nil, errors.New("not found"))
	// Card creation should still succeed (without merchant name)
	mockCardService.On("CreateCard", mock.Anything, mock.AnythingOfType("*models.Card")).Return(nil)

	// Create handler with mocks
	handler := &Handler{
		cardService:     mockCardService,
		merchantService: mockMerchantService,
	}

	// Execute
	err := handler.Create(c)

	// Assert - should still succeed (merchant is optional)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Contains(t, rec.Header().Get("Location"), "/cards/")
	mockCardService.AssertExpectations(t)
	mockMerchantService.AssertExpectations(t)
}

func TestCreateHandler_NewMerchantName(t *testing.T) {
	// Setup
	e := echo.New()
	userID := uuid.New()

	// Form data with new merchant name
	formData := url.Values{}
	formData.Set("merchant_id", "new")
	formData.Set("merchant_name", "Brand New Merchant")
	formData.Set("card_number", "9876543210")
	formData.Set("program", "Loyalty Program")

	req := httptest.NewRequest(http.MethodPost, "/cards", strings.NewReader(formData.Encode()))
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
	mockCardService := new(MockCardService)
	mockCardService.On("CreateCard", mock.Anything, mock.MatchedBy(func(card *models.Card) bool {
		return card.MerchantName == "Brand New Merchant" && card.MerchantID == nil
	})).Return(nil)

	// Create handler with mocks
	handler := &Handler{
		cardService: mockCardService,
	}

	// Execute
	err := handler.Create(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Contains(t, rec.Header().Get("Location"), "/cards/")
	mockCardService.AssertExpectations(t)
}
