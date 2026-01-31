package vouchers

import (
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
	voucherID := uuid.New()
	req := httptest.NewRequest(http.MethodDelete, "/vouchers/"+voucherID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(voucherID.String())

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
	mockAuthz.On("CheckVoucherAccess", mock.Anything, userID, voucherID).Return(perms, nil)

	// Mock voucher service
	mockVoucherService := new(MockVoucherService)
	mockVoucherService.On("DeleteVoucher", mock.Anything, voucherID).Return(nil)

	// Create handler with mocks
	handler := &Handler{
		authzService:   mockAuthz,
		voucherService: mockVoucherService,
		db:             nil,
	}

	// Execute
	err := handler.Delete(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/vouchers", rec.Header().Get("Location"))
	mockAuthz.AssertExpectations(t)
	mockVoucherService.AssertExpectations(t)
}

func TestDeleteHandler_Forbidden(t *testing.T) {
	// Setup
	e := echo.New()
	voucherID := uuid.New()
	req := httptest.NewRequest(http.MethodDelete, "/vouchers/"+voucherID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(voucherID.String())

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
	mockAuthz.On("CheckVoucherAccess", mock.Anything, userID, voucherID).Return(perms, nil)

	// Create handler with mock
	handler := &Handler{
		authzService:   mockAuthz,
		voucherService: nil,
		db:             nil,
	}

	// Execute
	err := handler.Delete(c)

	// Assert - Should redirect
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/vouchers", rec.Header().Get("Location"))
	mockAuthz.AssertExpectations(t)
}

func TestDeleteHandler_InvalidID(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/vouchers/invalid-id", nil)
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
		authzService:   nil,
		voucherService: nil,
		db:             nil,
	}

	// Execute
	err := handler.Delete(c)

	// Assert - Should redirect
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/vouchers", rec.Header().Get("Location"))
}

func TestDeleteHandler_ServiceError(t *testing.T) {
	// Setup
	e := echo.New()
	voucherID := uuid.New()
	req := httptest.NewRequest(http.MethodDelete, "/vouchers/"+voucherID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(voucherID.String())

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
	mockAuthz.On("CheckVoucherAccess", mock.Anything, userID, voucherID).Return(perms, nil)

	// Mock voucher service - simulate service error
	mockVoucherService := new(MockVoucherService)
	mockVoucherService.On("DeleteVoucher", mock.Anything, voucherID).Return(assert.AnError)

	// Create handler with mocks
	handler := &Handler{
		authzService:   mockAuthz,
		voucherService: mockVoucherService,
		db:             nil,
	}

	// Execute
	err := handler.Delete(c)

	// Assert - Should handle error (redirect on service error)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/vouchers", rec.Header().Get("Location"))
	mockAuthz.AssertExpectations(t)
	mockVoucherService.AssertExpectations(t)
}
