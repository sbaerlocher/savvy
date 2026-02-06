package vouchers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	savvyi18n "savvy/internal/i18n"
	"savvy/internal/models"
	"savvy/internal/services"
)

func TestEditHandler_Owner(t *testing.T) {
	// Setup
	e := echo.New()
	voucherID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/vouchers/"+voucherID.String()+"/edit", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(voucherID.String())

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
	mockAuthz.On("CheckVoucherAccess", mock.Anything, userID, voucherID).Return(perms, nil)

	// Mock voucher service
	mockVoucherService := new(MockVoucherService)
	voucher := &models.Voucher{
		ID:     voucherID,
		UserID: &userID,
		Code:   "TEST123",
	}
	mockVoucherService.On("GetVoucher", mock.Anything, voucherID).Return(voucher, nil)

	// Mock merchant service
	mockMerchantService := new(MockMerchantService)
	merchants := []models.Merchant{
		{ID: uuid.New(), Name: "Test Merchant"},
	}
	mockMerchantService.On("GetAllMerchants", mock.Anything).Return(merchants, nil)

	// Create handler with mocks
	handler := &Handler{
		authzService:    mockAuthz,
		voucherService:  mockVoucherService,
		merchantService: mockMerchantService,
	}

	// Execute
	err := handler.Edit(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockAuthz.AssertExpectations(t)
	mockVoucherService.AssertExpectations(t)
	mockMerchantService.AssertExpectations(t)
}

func TestEditHandler_Forbidden(t *testing.T) {
	// Setup
	e := echo.New()
	voucherID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/vouchers/"+voucherID.String()+"/edit", nil)
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

	// Mock authz service - deny edit
	mockAuthz := new(MockAuthzService)
	perms := &services.ResourcePermissions{
		CanView:   true,
		CanEdit:   false, // Not allowed to edit
		CanDelete: false,
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
	err := handler.Edit(c)

	// Assert - Should redirect
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/vouchers", rec.Header().Get("Location"))
	mockAuthz.AssertExpectations(t)
}

func TestEditHandler_InvalidID(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/vouchers/invalid-id/edit", nil)
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
	err := handler.Edit(c)

	// Assert - Should redirect
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/vouchers", rec.Header().Get("Location"))
}

func TestEditHandler_VoucherNotFound(t *testing.T) {
	// Setup
	e := echo.New()
	voucherID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/vouchers/"+voucherID.String()+"/edit", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(voucherID.String())

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
	mockAuthz.On("CheckVoucherAccess", mock.Anything, userID, voucherID).Return(perms, nil)

	// Mock voucher service - return not found error
	mockVoucherService := new(MockVoucherService)
	mockVoucherService.On("GetVoucher", mock.Anything, voucherID).Return(nil, errors.New("not found"))

	// Create handler with mocks
	handler := &Handler{
		authzService:   mockAuthz,
		voucherService: mockVoucherService,
	}

	// Execute
	err := handler.Edit(c)

	// Assert - Should redirect
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/vouchers", rec.Header().Get("Location"))
	mockAuthz.AssertExpectations(t)
	mockVoucherService.AssertExpectations(t)
}
