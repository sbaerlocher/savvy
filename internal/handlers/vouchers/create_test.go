package vouchers

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
	formData.Set("code", "SAVE20")
	formData.Set("type", "percentage")
	formData.Set("value", "20.00")
	formData.Set("description", "20% discount")
	formData.Set("min_purchase_amount", "50.00")
	formData.Set("valid_from", "2026-02-01")
	formData.Set("valid_until", "2026-03-01")
	formData.Set("usage_limit_type", "unlimited")
	formData.Set("barcode_type", "QR")

	req := httptest.NewRequest(http.MethodPost, "/vouchers", strings.NewReader(formData.Encode()))
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
	mockVoucherService := new(MockVoucherService)
	mockMerchantService := new(MockMerchantService)

	merchant := &models.Merchant{
		ID:   merchantID,
		Name: "Test Merchant",
	}
	mockMerchantService.On("GetMerchantByID", mock.Anything, merchantID).Return(merchant, nil)
	mockVoucherService.On("CreateVoucher", mock.Anything, mock.AnythingOfType("*models.Voucher")).Return(nil)

	// Create handler with mocks
	handler := &Handler{
		voucherService:  mockVoucherService,
		merchantService: mockMerchantService,
	}

	// Execute
	err := handler.Create(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Contains(t, rec.Header().Get("Location"), "/vouchers/")
	mockVoucherService.AssertExpectations(t)
	mockMerchantService.AssertExpectations(t)
}

func TestCreateHandler_ValidationError_MissingCode(t *testing.T) {
	// Setup
	e := echo.New()
	userID := uuid.New()

	// Form data without code
	formData := url.Values{}
	formData.Set("merchant_name", "Test Merchant")
	formData.Set("type", "percentage")
	formData.Set("value", "10.00")
	formData.Set("valid_from", "2026-02-01")
	formData.Set("valid_until", "2026-03-01")

	req := httptest.NewRequest(http.MethodPost, "/vouchers", strings.NewReader(formData.Encode()))
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
	mockVoucherService := new(MockVoucherService)
	// CreateVoucher will be called but might fail validation
	mockVoucherService.On("CreateVoucher", mock.Anything, mock.AnythingOfType("*models.Voucher")).Return(errors.New("validation error: code required"))

	// Create handler with mocks
	handler := &Handler{
		voucherService: mockVoucherService,
	}

	// Execute
	err := handler.Create(c)

	// Assert - should redirect back to new form with error
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Contains(t, rec.Header().Get("Location"), "/vouchers/new")
	mockVoucherService.AssertExpectations(t)
}

func TestCreateHandler_InvalidDateRange(t *testing.T) {
	// Setup
	e := echo.New()
	userID := uuid.New()

	// Form data with invalid date range (end before start)
	formData := url.Values{}
	formData.Set("merchant_name", "Test Merchant")
	formData.Set("code", "TEST")
	formData.Set("type", "fixed")
	formData.Set("value", "10.00")
	formData.Set("valid_from", "2026-03-01")
	formData.Set("valid_until", "2026-02-01") // Before valid_from

	req := httptest.NewRequest(http.MethodPost, "/vouchers", strings.NewReader(formData.Encode()))
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
	assert.Contains(t, rec.Header().Get("Location"), "/vouchers/new?error=invalid_date")
}

func TestCreateHandler_ServiceError(t *testing.T) {
	// Setup
	e := echo.New()
	userID := uuid.New()

	// Valid form data
	formData := url.Values{}
	formData.Set("merchant_name", "Test Merchant")
	formData.Set("code", "SAVE20")
	formData.Set("type", "percentage")
	formData.Set("value", "20.00")
	formData.Set("valid_from", "2026-02-01")
	formData.Set("valid_until", "2026-03-01")
	formData.Set("usage_limit_type", "unlimited")

	req := httptest.NewRequest(http.MethodPost, "/vouchers", strings.NewReader(formData.Encode()))
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
	mockVoucherService := new(MockVoucherService)
	mockVoucherService.On("CreateVoucher", mock.Anything, mock.AnythingOfType("*models.Voucher")).Return(errors.New("database error"))

	// Create handler with mocks
	handler := &Handler{
		voucherService: mockVoucherService,
	}

	// Execute
	err := handler.Create(c)

	// Assert - should redirect to new form with error
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Contains(t, rec.Header().Get("Location"), "/vouchers/new?error=")
	mockVoucherService.AssertExpectations(t)
}

func TestCreateHandler_NewMerchantName(t *testing.T) {
	// Setup
	e := echo.New()
	userID := uuid.New()

	// Form data with new merchant name
	formData := url.Values{}
	formData.Set("merchant_id", "new")
	formData.Set("merchant_name", "Brand New Merchant")
	formData.Set("code", "WELCOME10")
	formData.Set("type", "fixed")
	formData.Set("value", "10.00")
	formData.Set("valid_from", "2026-02-01")
	formData.Set("valid_until", "2026-12-31")
	formData.Set("usage_limit_type", "single_use")

	req := httptest.NewRequest(http.MethodPost, "/vouchers", strings.NewReader(formData.Encode()))
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
	mockVoucherService := new(MockVoucherService)
	mockVoucherService.On("CreateVoucher", mock.Anything, mock.MatchedBy(func(voucher *models.Voucher) bool {
		return voucher.MerchantName == "Brand New Merchant" && voucher.MerchantID == nil
	})).Return(nil)

	// Create handler with mocks
	handler := &Handler{
		voucherService: mockVoucherService,
	}

	// Execute
	err := handler.Create(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Contains(t, rec.Header().Get("Location"), "/vouchers/")
	mockVoucherService.AssertExpectations(t)
}
