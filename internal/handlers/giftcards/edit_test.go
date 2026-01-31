package giftcards

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
	giftCardID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/gift-cards/"+giftCardID.String()+"/edit", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(giftCardID.String())

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
		CanView:             true,
		CanEdit:             true,
		CanDelete:           true,
		CanEditTransactions: true,
		IsOwner:             true,
	}
	mockAuthz.On("CheckGiftCardAccess", mock.Anything, userID, giftCardID).Return(perms, nil)

	// Mock gift card service
	mockGiftCardService := new(MockGiftCardService)
	giftCard := &models.GiftCard{
		ID:         giftCardID,
		UserID:     &userID,
		CardNumber: "GC123",
	}
	mockGiftCardService.On("GetGiftCard", mock.Anything, giftCardID).Return(giftCard, nil)

	// Mock merchant service
	mockMerchantService := new(MockMerchantService)
	merchants := []models.Merchant{
		{ID: uuid.New(), Name: "Test Merchant"},
	}
	mockMerchantService.On("GetAllMerchants", mock.Anything).Return(merchants, nil)

	// Create handler with mocks
	handler := &Handler{
		authzService:    mockAuthz,
		giftCardService: mockGiftCardService,
		merchantService: mockMerchantService,
	}

	// Execute
	err := handler.Edit(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockAuthz.AssertExpectations(t)
	mockGiftCardService.AssertExpectations(t)
	mockMerchantService.AssertExpectations(t)
}

func TestEditHandler_Forbidden(t *testing.T) {
	// Setup
	e := echo.New()
	giftCardID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/gift-cards/"+giftCardID.String()+"/edit", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(giftCardID.String())

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
		CanView:             true,
		CanEdit:             false, // Not allowed to edit
		CanDelete:           false,
		CanEditTransactions: false,
		IsOwner:             false,
	}
	mockAuthz.On("CheckGiftCardAccess", mock.Anything, userID, giftCardID).Return(perms, nil)

	// Create handler with mock
	handler := &Handler{
		authzService:    mockAuthz,
		giftCardService: nil,
		db:              nil,
	}

	// Execute
	err := handler.Edit(c)

	// Assert - Should redirect
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/gift-cards", rec.Header().Get("Location"))
	mockAuthz.AssertExpectations(t)
}

func TestEditHandler_InvalidID(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/gift-cards/invalid-id/edit", nil)
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
	handler := &Handler{
		authzService:    nil,
		giftCardService: nil,
		db:              nil,
	}

	// Execute
	err := handler.Edit(c)

	// Assert - Should redirect
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/gift-cards", rec.Header().Get("Location"))
}

func TestEditHandler_GiftCardNotFound(t *testing.T) {
	// Setup
	e := echo.New()
	giftCardID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/gift-cards/"+giftCardID.String()+"/edit", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(giftCardID.String())

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
		CanView:             true,
		CanEdit:             true,
		CanDelete:           true,
		CanEditTransactions: true,
		IsOwner:             true,
	}
	mockAuthz.On("CheckGiftCardAccess", mock.Anything, userID, giftCardID).Return(perms, nil)

	// Mock gift card service - return not found error
	mockGiftCardService := new(MockGiftCardService)
	mockGiftCardService.On("GetGiftCard", mock.Anything, giftCardID).Return(nil, errors.New("not found"))

	// Create handler with mocks
	handler := &Handler{
		authzService:    mockAuthz,
		giftCardService: mockGiftCardService,
	}

	// Execute
	err := handler.Edit(c)

	// Assert - Should redirect
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/gift-cards", rec.Header().Get("Location"))
	mockAuthz.AssertExpectations(t)
	mockGiftCardService.AssertExpectations(t)
}
