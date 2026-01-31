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
	"savvy/internal/services"
)

func TestUpdateHandler_Owner(t *testing.T) {
	// Setup
	e := echo.New()
	userID := uuid.New()
	cardID := uuid.New()
	merchantID := uuid.New()

	// Form data
	formData := url.Values{}
	formData.Set("merchant_id", merchantID.String())
	formData.Set("card_number", "9999999999")
	formData.Set("program", "Updated Program")
	formData.Set("barcode_type", "QR")
	formData.Set("status", "active")
	formData.Set("notes", "Updated notes")

	req := httptest.NewRequest(http.MethodPost, "/cards/"+cardID.String(), strings.NewReader(formData.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cardID.String())

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
	mockCardService := new(MockCardService)
	mockMerchantService := new(MockMerchantService)

	perms := &services.ResourcePermissions{
		CanView:   true,
		CanEdit:   true,
		CanDelete: true,
		IsOwner:   true,
	}
	mockAuthzService.On("CheckCardAccess", mock.Anything, userID, cardID).Return(perms, nil)

	card := &models.Card{
		ID:         cardID,
		UserID:     &userID,
		CardNumber: "1234567890",
	}
	mockCardService.On("GetCard", mock.Anything, cardID).Return(card, nil)

	merchant := &models.Merchant{
		ID:   merchantID,
		Name: "Updated Merchant",
	}
	mockMerchantService.On("GetMerchantByID", mock.Anything, merchantID).Return(merchant, nil)

	mockCardService.On("UpdateCard", mock.Anything, mock.AnythingOfType("*models.Card")).Return(nil)

	// Create handler with mocks
	handler := &Handler{
		authzService:    mockAuthzService,
		cardService:     mockCardService,
		merchantService: mockMerchantService,
		db:              nil, // Not needed for this test (audit log uses database.DB directly)
	}

	// Execute
	err := handler.Update(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/cards/"+cardID.String(), rec.Header().Get("Location"))
	mockAuthzService.AssertExpectations(t)
	mockCardService.AssertExpectations(t)
	mockMerchantService.AssertExpectations(t)
}

func TestUpdateHandler_Forbidden(t *testing.T) {
	// Setup
	e := echo.New()
	userID := uuid.New()
	cardID := uuid.New()

	// Form data
	formData := url.Values{}
	formData.Set("card_number", "9999999999")

	req := httptest.NewRequest(http.MethodPost, "/cards/"+cardID.String(), strings.NewReader(formData.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cardID.String())

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
		CanView:   true,
		CanEdit:   false, // Not allowed to edit
		CanDelete: false,
		IsOwner:   false,
	}
	mockAuthzService.On("CheckCardAccess", mock.Anything, userID, cardID).Return(perms, nil)

	// Create handler with mock
	handler := &Handler{
		authzService: mockAuthzService,
	}

	// Execute
	err := handler.Update(c)

	// Assert - should redirect
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/cards", rec.Header().Get("Location"))
	mockAuthzService.AssertExpectations(t)
}

func TestUpdateHandler_ValidationError(t *testing.T) {
	// Setup
	e := echo.New()
	userID := uuid.New()
	cardID := uuid.New()

	// Invalid form data (empty card_number)
	formData := url.Values{}
	formData.Set("card_number", "")

	req := httptest.NewRequest(http.MethodPost, "/cards/"+cardID.String(), strings.NewReader(formData.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cardID.String())

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
	mockCardService := new(MockCardService)

	perms := &services.ResourcePermissions{
		CanView:   true,
		CanEdit:   true,
		CanDelete: true,
		IsOwner:   true,
	}
	mockAuthzService.On("CheckCardAccess", mock.Anything, userID, cardID).Return(perms, nil)

	card := &models.Card{
		ID:         cardID,
		UserID:     &userID,
		CardNumber: "1234567890",
	}
	mockCardService.On("GetCard", mock.Anything, cardID).Return(card, nil)

	// Update fails with validation error
	mockCardService.On("UpdateCard", mock.Anything, mock.AnythingOfType("*models.Card")).Return(errors.New("validation error"))

	// Create handler with mocks
	handler := &Handler{
		authzService: mockAuthzService,
		cardService:  mockCardService,
	}

	// Execute
	err := handler.Update(c)

	// Assert - should redirect back to edit form
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/cards/"+cardID.String()+"/edit", rec.Header().Get("Location"))
	mockAuthzService.AssertExpectations(t)
	mockCardService.AssertExpectations(t)
}

func TestUpdateHandler_ServiceError(t *testing.T) {
	// Setup
	e := echo.New()
	userID := uuid.New()
	cardID := uuid.New()

	// Valid form data
	formData := url.Values{}
	formData.Set("card_number", "9999999999")
	formData.Set("program", "Test")

	req := httptest.NewRequest(http.MethodPost, "/cards/"+cardID.String(), strings.NewReader(formData.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cardID.String())

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
	mockCardService := new(MockCardService)

	perms := &services.ResourcePermissions{
		CanView:   true,
		CanEdit:   true,
		CanDelete: true,
		IsOwner:   true,
	}
	mockAuthzService.On("CheckCardAccess", mock.Anything, userID, cardID).Return(perms, nil)

	card := &models.Card{
		ID:         cardID,
		UserID:     &userID,
		CardNumber: "1234567890",
	}
	mockCardService.On("GetCard", mock.Anything, cardID).Return(card, nil)

	// Update fails with service error
	mockCardService.On("UpdateCard", mock.Anything, mock.AnythingOfType("*models.Card")).Return(errors.New("database error"))

	// Create handler with mocks
	handler := &Handler{
		authzService: mockAuthzService,
		cardService:  mockCardService,
	}

	// Execute
	err := handler.Update(c)

	// Assert - should redirect back to edit form
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/cards/"+cardID.String()+"/edit", rec.Header().Get("Location"))
	mockAuthzService.AssertExpectations(t)
	mockCardService.AssertExpectations(t)
}

func TestUpdateHandler_NotFound(t *testing.T) {
	// Setup
	e := echo.New()
	userID := uuid.New()
	cardID := uuid.New()

	// Form data
	formData := url.Values{}
	formData.Set("card_number", "9999999999")

	req := httptest.NewRequest(http.MethodPost, "/cards/"+cardID.String(), strings.NewReader(formData.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cardID.String())

	// Setup user context
	user := &models.User{
		ID:    userID,
		Email: "owner@example.com",
	}
	c.Set("current_user", user)
	setupI18nContext(c)

	// Mock services
	mockAuthzService := new(MockAuthzService)
	mockCardService := new(MockCardService)

	perms := &services.ResourcePermissions{
		CanView:   true,
		CanEdit:   true,
		CanDelete: true,
		IsOwner:   true,
	}
	mockAuthzService.On("CheckCardAccess", mock.Anything, userID, cardID).Return(perms, nil)

	// Card not found
	mockCardService.On("GetCard", mock.Anything, cardID).Return(nil, errors.New("not found"))

	// Create handler with mocks
	handler := &Handler{
		authzService: mockAuthzService,
		cardService:  mockCardService,
	}

	// Execute
	err := handler.Update(c)

	// Assert - should redirect to cards list
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/cards", rec.Header().Get("Location"))
	mockAuthzService.AssertExpectations(t)
	mockCardService.AssertExpectations(t)
}
