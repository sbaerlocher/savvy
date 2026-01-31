package vouchers

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
	voucherID := uuid.New()
	merchantID := uuid.New()

	// Form data
	formData := url.Values{}
	formData.Set("merchant_id", merchantID.String())
	formData.Set("code", "UPDATED20")
	formData.Set("type", "percentage")
	formData.Set("value", "25.00")
	formData.Set("description", "Updated description")
	formData.Set("min_purchase_amount", "100.00")
	formData.Set("valid_from", "2026-02-01")
	formData.Set("valid_until", "2026-04-01")
	formData.Set("usage_limit_type", "unlimited")
	formData.Set("barcode_type", "QR")

	req := httptest.NewRequest(http.MethodPost, "/vouchers/"+voucherID.String(), strings.NewReader(formData.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(voucherID.String())

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
	mockVoucherService := new(MockVoucherService)
	mockMerchantService := new(MockMerchantService)

	perms := &services.ResourcePermissions{
		CanView:   true,
		CanEdit:   true,
		CanDelete: true,
		IsOwner:   true,
	}
	mockAuthzService.On("CheckVoucherAccess", mock.Anything, userID, voucherID).Return(perms, nil)

	validFrom := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	validUntil := time.Date(2026, 3, 1, 23, 59, 59, 0, time.UTC)
	voucher := &models.Voucher{
		ID:         voucherID,
		UserID:     &userID,
		Code:       "SAVE20",
		ValidFrom:  validFrom,
		ValidUntil: validUntil,
	}
	mockVoucherService.On("GetVoucher", mock.Anything, voucherID).Return(voucher, nil)

	merchant := &models.Merchant{
		ID:   merchantID,
		Name: "Updated Merchant",
	}
	mockMerchantService.On("GetMerchantByID", mock.Anything, merchantID).Return(merchant, nil)

	mockVoucherService.On("UpdateVoucher", mock.Anything, mock.AnythingOfType("*models.Voucher")).Return(nil)

	// Create handler with mocks
	handler := &Handler{
		authzService:    mockAuthzService,
		voucherService:  mockVoucherService,
		merchantService: mockMerchantService,
	}

	// Execute
	err := handler.Update(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/vouchers/"+voucherID.String(), rec.Header().Get("Location"))
	mockAuthzService.AssertExpectations(t)
	mockVoucherService.AssertExpectations(t)
	mockMerchantService.AssertExpectations(t)
}

func TestUpdateHandler_Forbidden(t *testing.T) {
	// Setup
	e := echo.New()
	userID := uuid.New()
	voucherID := uuid.New()

	// Form data
	formData := url.Values{}
	formData.Set("code", "UPDATED")

	req := httptest.NewRequest(http.MethodPost, "/vouchers/"+voucherID.String(), strings.NewReader(formData.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(voucherID.String())

	// Setup user context
	user := &models.User{
		ID:    userID,
		Email: "notowner@example.com",
	}
	c.Set("current_user", user)
	setupI18nContext(c)

	// Mock services - no edit permission (vouchers are read-only when shared)
	mockAuthzService := new(MockAuthzService)
	perms := &services.ResourcePermissions{
		CanView:   true,
		CanEdit:   false, // Not allowed to edit
		CanDelete: false,
		IsOwner:   false,
	}
	mockAuthzService.On("CheckVoucherAccess", mock.Anything, userID, voucherID).Return(perms, nil)

	// Create handler with mock
	handler := &Handler{
		authzService: mockAuthzService,
	}

	// Execute
	err := handler.Update(c)

	// Assert - should redirect
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/vouchers", rec.Header().Get("Location"))
	mockAuthzService.AssertExpectations(t)
}

func TestUpdateHandler_InvalidDateRange(t *testing.T) {
	// Setup
	e := echo.New()
	userID := uuid.New()
	voucherID := uuid.New()

	// Invalid form data (end date before start date)
	formData := url.Values{}
	formData.Set("code", "TEST")
	formData.Set("type", "fixed")
	formData.Set("value", "10.00")
	formData.Set("valid_from", "2026-03-01")
	formData.Set("valid_until", "2026-02-01") // Before valid_from

	req := httptest.NewRequest(http.MethodPost, "/vouchers/"+voucherID.String(), strings.NewReader(formData.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(voucherID.String())

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
	mockVoucherService := new(MockVoucherService)

	perms := &services.ResourcePermissions{
		CanView:   true,
		CanEdit:   true,
		CanDelete: true,
		IsOwner:   true,
	}
	mockAuthzService.On("CheckVoucherAccess", mock.Anything, userID, voucherID).Return(perms, nil)

	validFrom := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	validUntil := time.Date(2026, 3, 1, 23, 59, 59, 0, time.UTC)
	voucher := &models.Voucher{
		ID:         voucherID,
		UserID:     &userID,
		Code:       "SAVE20",
		ValidFrom:  validFrom,
		ValidUntil: validUntil,
	}
	mockVoucherService.On("GetVoucher", mock.Anything, voucherID).Return(voucher, nil)

	// Create handler with mocks
	handler := &Handler{
		authzService:   mockAuthzService,
		voucherService: mockVoucherService,
	}

	// Execute
	err := handler.Update(c)

	// Assert - should redirect back to edit form with error
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Contains(t, rec.Header().Get("Location"), "/vouchers/"+voucherID.String()+"/edit?error=invalid_date")
	mockAuthzService.AssertExpectations(t)
	mockVoucherService.AssertExpectations(t)
}

func TestUpdateHandler_ServiceError(t *testing.T) {
	// Setup
	e := echo.New()
	userID := uuid.New()
	voucherID := uuid.New()

	// Valid form data
	formData := url.Values{}
	formData.Set("code", "UPDATED")
	formData.Set("type", "fixed")
	formData.Set("value", "15.00")
	formData.Set("valid_from", "2026-02-01")
	formData.Set("valid_until", "2026-03-01")

	req := httptest.NewRequest(http.MethodPost, "/vouchers/"+voucherID.String(), strings.NewReader(formData.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(voucherID.String())

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
	mockVoucherService := new(MockVoucherService)

	perms := &services.ResourcePermissions{
		CanView:   true,
		CanEdit:   true,
		CanDelete: true,
		IsOwner:   true,
	}
	mockAuthzService.On("CheckVoucherAccess", mock.Anything, userID, voucherID).Return(perms, nil)

	validFrom := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	validUntil := time.Date(2026, 3, 1, 23, 59, 59, 0, time.UTC)
	voucher := &models.Voucher{
		ID:         voucherID,
		UserID:     &userID,
		Code:       "SAVE20",
		ValidFrom:  validFrom,
		ValidUntil: validUntil,
	}
	mockVoucherService.On("GetVoucher", mock.Anything, voucherID).Return(voucher, nil)

	// Update fails with service error
	mockVoucherService.On("UpdateVoucher", mock.Anything, mock.AnythingOfType("*models.Voucher")).Return(errors.New("database error"))

	// Create handler with mocks
	handler := &Handler{
		authzService:   mockAuthzService,
		voucherService: mockVoucherService,
	}

	// Execute
	err := handler.Update(c)

	// Assert - should redirect back to edit form
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/vouchers/"+voucherID.String()+"/edit", rec.Header().Get("Location"))
	mockAuthzService.AssertExpectations(t)
	mockVoucherService.AssertExpectations(t)
}

func TestUpdateHandler_NotFound(t *testing.T) {
	// Setup
	e := echo.New()
	userID := uuid.New()
	voucherID := uuid.New()

	// Form data
	formData := url.Values{}
	formData.Set("code", "NOTFOUND")

	req := httptest.NewRequest(http.MethodPost, "/vouchers/"+voucherID.String(), strings.NewReader(formData.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(voucherID.String())

	// Setup user context
	user := &models.User{
		ID:    userID,
		Email: "owner@example.com",
	}
	c.Set("current_user", user)
	setupI18nContext(c)

	// Mock services
	mockAuthzService := new(MockAuthzService)
	mockVoucherService := new(MockVoucherService)

	perms := &services.ResourcePermissions{
		CanView:   true,
		CanEdit:   true,
		CanDelete: true,
		IsOwner:   true,
	}
	mockAuthzService.On("CheckVoucherAccess", mock.Anything, userID, voucherID).Return(perms, nil)

	// Voucher not found
	mockVoucherService.On("GetVoucher", mock.Anything, voucherID).Return(nil, errors.New("not found"))

	// Create handler with mocks
	handler := &Handler{
		authzService:   mockAuthzService,
		voucherService: mockVoucherService,
	}

	// Execute
	err := handler.Update(c)

	// Assert - should redirect to vouchers list
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/vouchers", rec.Header().Get("Location"))
	mockAuthzService.AssertExpectations(t)
	mockVoucherService.AssertExpectations(t)
}
