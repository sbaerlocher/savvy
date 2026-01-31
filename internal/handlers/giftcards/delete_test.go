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
)

func TestDeleteHandler_Owner(t *testing.T) {
	// Setup
	e := echo.New()
	giftCardID := uuid.New()
	req := httptest.NewRequest(http.MethodDelete, "/gift-cards/"+giftCardID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(giftCardID.String())

	// Setup user context
	userID := uuid.New()
	user := &models.User{
		ID:    userID,
		Email: "owner@example.com",
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

	// Mock gift card service
	mockGiftCardService := new(MockGiftCardService)
	mockGiftCardService.On("DeleteGiftCard", mock.Anything, giftCardID).Return(nil)

	// Create handler with mocks
	handler := &Handler{
		authzService:    mockAuthz,
		giftCardService: mockGiftCardService,
	}

	// Execute
	err := handler.Delete(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/gift-cards", rec.Header().Get("Location"))
	mockAuthz.AssertExpectations(t)
	mockGiftCardService.AssertExpectations(t)
}

func TestDeleteHandler_Forbidden(t *testing.T) {
	// Setup
	e := echo.New()
	giftCardID := uuid.New()
	req := httptest.NewRequest(http.MethodDelete, "/gift-cards/"+giftCardID.String(), nil)
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

	// Mock authz service - deny delete
	mockAuthz := new(MockAuthzService)
	perms := &services.ResourcePermissions{
		CanView:             true,
		CanEdit:             false,
		CanDelete:           false, // Not allowed to delete
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
	err := handler.Delete(c)

	// Assert - Should redirect
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/gift-cards", rec.Header().Get("Location"))
	mockAuthz.AssertExpectations(t)
}

func TestDeleteHandler_InvalidID(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/gift-cards/invalid-id", nil)
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
	err := handler.Delete(c)

	// Assert - Should redirect
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/gift-cards", rec.Header().Get("Location"))
}

func TestDeleteHandler_ServiceError(t *testing.T) {
	// Setup
	e := echo.New()
	giftCardID := uuid.New()
	req := httptest.NewRequest(http.MethodDelete, "/gift-cards/"+giftCardID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(giftCardID.String())

	// Setup user context
	userID := uuid.New()
	user := &models.User{
		ID:    userID,
		Email: "owner@example.com",
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

	// Mock gift card service with error
	mockGiftCardService := new(MockGiftCardService)
	mockGiftCardService.On("DeleteGiftCard", mock.Anything, giftCardID).Return(errors.New("db error"))

	// Create handler with mocks
	handler := &Handler{
		authzService:    mockAuthz,
		giftCardService: mockGiftCardService,
	}

	// Execute
	err := handler.Delete(c)

	// Assert - Should handle error gracefully (redirect)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/gift-cards", rec.Header().Get("Location"))
	mockAuthz.AssertExpectations(t)
	mockGiftCardService.AssertExpectations(t)
}
