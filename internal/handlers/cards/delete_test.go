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
)

func TestDeleteHandler_Owner(t *testing.T) {
	// Setup
	e := echo.New()
	cardID := uuid.New()
	req := httptest.NewRequest(http.MethodDelete, "/cards/"+cardID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cardID.String())

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
		CanView:   true,
		CanEdit:   true,
		CanDelete: true,
		IsOwner:   true,
	}
	mockAuthz.On("CheckCardAccess", mock.Anything, userID, cardID).Return(perms, nil)

	// Mock card service
	mockCardService := new(MockCardService)
	mockCardService.On("DeleteCard", mock.Anything, cardID).Return(nil)

	// Create handler with mocks
	handler := &Handler{
		authzService: mockAuthz,
		cardService:  mockCardService,
	}

	// Execute
	err := handler.Delete(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/cards", rec.Header().Get("Location"))
	assert.Equal(t, "/cards", rec.Header().Get("HX-Redirect"))
	mockAuthz.AssertExpectations(t)
	mockCardService.AssertExpectations(t)
}

func TestDeleteHandler_Forbidden(t *testing.T) {
	// Setup
	e := echo.New()
	cardID := uuid.New()
	req := httptest.NewRequest(http.MethodDelete, "/cards/"+cardID.String(), nil)
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

	// Mock authz service - deny delete
	mockAuthz := new(MockAuthzService)
	perms := &services.ResourcePermissions{
		CanView:   true,
		CanEdit:   false,
		CanDelete: false, // Not allowed to delete
		IsOwner:   false,
	}
	mockAuthz.On("CheckCardAccess", mock.Anything, userID, cardID).Return(perms, nil)

	// Create handler with mock
	handler := &Handler{
		authzService: mockAuthz,
	}

	// Execute
	err := handler.Delete(c)

	// Assert - Should redirect
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/cards", rec.Header().Get("Location"))
	mockAuthz.AssertExpectations(t)
}

func TestDeleteHandler_InvalidID(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/cards/invalid-id", nil)
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
	err := handler.Delete(c)

	// Assert - Should redirect
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/cards", rec.Header().Get("Location"))
}

func TestDeleteHandler_ServiceError(t *testing.T) {
	// Setup
	e := echo.New()
	cardID := uuid.New()
	req := httptest.NewRequest(http.MethodDelete, "/cards/"+cardID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cardID.String())

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
		CanView:   true,
		CanEdit:   true,
		CanDelete: true,
		IsOwner:   true,
	}
	mockAuthz.On("CheckCardAccess", mock.Anything, userID, cardID).Return(perms, nil)

	// Mock card service - return error
	mockCardService := new(MockCardService)
	mockCardService.On("DeleteCard", mock.Anything, cardID).Return(errors.New("delete failed"))

	// Create handler with mocks
	handler := &Handler{
		authzService: mockAuthz,
		cardService:  mockCardService,
	}

	// Execute
	err := handler.Delete(c)

	// Assert - Should redirect on error
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/cards", rec.Header().Get("Location"))
	mockAuthz.AssertExpectations(t)
	mockCardService.AssertExpectations(t)
}
