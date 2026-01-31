package cards

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"savvy/internal/models"
	"savvy/internal/services"
	savvyi18n "savvy/internal/i18n"
)

func TestEditHandler_Owner(t *testing.T) {
	// Setup
	e := echo.New()
	cardID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/cards/"+cardID.String()+"/edit", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cardID.String())

	// Setup i18n context
	localizer := savvyi18n.NewLocalizer("de")
	ctx := savvyi18n.SetLocalizer(c.Request().Context(), localizer)
	c.SetRequest(c.Request().WithContext(ctx))

	// Setup user context
	userID := uuid.New()
	user := &models.User{
		ID:    userID,
		Email: "owner@example.com",
	}
	c.Set("current_user", user)
	c.Set("csrf", "test-csrf-token")

	// Mock authz service
	mockAuthz := new(MockAuthzService)
	perms := &services.ResourcePermissions{
		CanView:   true,
		CanEdit:   true,
		CanDelete: true,
		IsOwner:   true,
	}
	mockAuthz.On("CheckCardAccess", mock.Anything, userID, cardID).Return(perms, nil)

	// Mock card service
	mockCardService := new(MockCardService)
	card := &models.Card{
		ID:         cardID,
		UserID:     &userID,
		CardNumber: "1234",
	}
	mockCardService.On("GetCard", mock.Anything, cardID).Return(card, nil)

	// Mock merchant service
	mockMerchantService := new(MockMerchantServiceNew)
	merchants := []models.Merchant{
		{ID: uuid.New(), Name: "Test Merchant"},
	}
	mockMerchantService.On("GetAllMerchants", mock.Anything).Return(merchants, nil)

	// Create handler with mocks
	handler := &Handler{
		authzService:    mockAuthz,
		cardService:     mockCardService,
		merchantService: mockMerchantService,
	}

	// Execute
	err := handler.Edit(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockAuthz.AssertExpectations(t)
	mockCardService.AssertExpectations(t)
	mockMerchantService.AssertExpectations(t)
}

func TestEditHandler_Forbidden(t *testing.T) {
	// Setup
	e := echo.New()
	cardID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/cards/"+cardID.String()+"/edit", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cardID.String())

	// Setup user context
	userID := uuid.New()
	user := &models.User{
		ID:    userID,
		Email: "notowner@example.com",
	}
	c.Set("current_user", user)

	// Mock authz service - deny edit
	mockAuthz := new(MockAuthzService)
	perms := &services.ResourcePermissions{
		CanView:   true,
		CanEdit:   false, // Not allowed to edit
		CanDelete: false,
		IsOwner:   false,
	}
	mockAuthz.On("CheckCardAccess", mock.Anything, userID, cardID).Return(perms, nil)

	// Create handler with mock
	handler := &Handler{
		authzService: mockAuthz,
	}

	// Execute
	err := handler.Edit(c)

	// Assert - Should redirect
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/cards", rec.Header().Get("Location"))
	mockAuthz.AssertExpectations(t)
}

func TestEditHandler_InvalidID(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/cards/invalid-id/edit", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("invalid-id")

	// Setup user context
	user := &models.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}
	c.Set("current_user", user)

	// Create handler
	handler := &Handler{}

	// Execute
	err := handler.Edit(c)

	// Assert - Should redirect
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/cards", rec.Header().Get("Location"))
}

func TestEditHandler_CardNotFound(t *testing.T) {
	// Setup
	e := echo.New()
	cardID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/cards/"+cardID.String()+"/edit", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cardID.String())

	// Setup user context
	userID := uuid.New()
	user := &models.User{
		ID:    userID,
		Email: "test@example.com",
	}
	c.Set("current_user", user)

	// Mock authz service
	mockAuthz := new(MockAuthzService)
	perms := &services.ResourcePermissions{
		CanView:   true,
		CanEdit:   true,
		CanDelete: true,
		IsOwner:   true,
	}
	mockAuthz.On("CheckCardAccess", mock.Anything, userID, cardID).Return(perms, nil)

	// Mock card service - return not found error
	mockCardService := new(MockCardService)
	mockCardService.On("GetCard", mock.Anything, cardID).Return(nil, errors.New("not found"))

	// Create handler with mocks
	handler := &Handler{
		authzService: mockAuthz,
		cardService:  mockCardService,
	}

	// Execute
	err := handler.Edit(c)

	// Assert - Should redirect
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/cards", rec.Header().Get("Location"))
	mockAuthz.AssertExpectations(t)
	mockCardService.AssertExpectations(t)
}
