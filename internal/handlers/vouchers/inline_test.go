package vouchers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	savvyi18n "savvy/internal/i18n"
	"savvy/internal/models"
	"savvy/internal/services"
)

func TestEditInline_Success(t *testing.T) {
	e := echo.New()
	voucherID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/vouchers/"+voucherID.String()+"/edit-inline", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(voucherID.String())

	// Setup i18n context
	localizer := savvyi18n.NewLocalizer("de")
	ctx := savvyi18n.SetLocalizer(c.Request().Context(), localizer)
	c.SetRequest(c.Request().WithContext(ctx))

	userID := uuid.New()
	user := &models.User{
		ID:    userID,
		Email: "test@example.com",
	}
	c.Set("current_user", user)
	c.Set("csrf", "test-csrf-token")

	mockAuthz := new(MockAuthzService)
	perms := &services.ResourcePermissions{
		CanView: true,
		CanEdit: true,
		IsOwner: true,
	}
	mockAuthz.On("CheckVoucherAccess", mock.Anything, userID, voucherID).Return(perms, nil)

	mockVoucherService := new(MockVoucherService)
	voucher := &models.Voucher{
		ID:         voucherID,
		UserID:     &userID,
		Code:       "VOUCHER123",
		ValidFrom:  time.Now(),
		ValidUntil: time.Now().Add(24 * time.Hour),
	}
	mockVoucherService.On("GetVoucher", mock.Anything, voucherID).Return(voucher, nil)

	mockMerchantService := new(MockMerchantService)
	merchants := []models.Merchant{
		{ID: uuid.New(), Name: "Test Merchant"},
	}
	mockMerchantService.On("GetAllMerchants", mock.Anything).Return(merchants, nil)

	handler := &Handler{
		authzService:    mockAuthz,
		voucherService:  mockVoucherService,
		merchantService: mockMerchantService,
	}

	err := handler.EditInline(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockAuthz.AssertExpectations(t)
	mockVoucherService.AssertExpectations(t)
	mockMerchantService.AssertExpectations(t)
}

func TestEditInline_Forbidden(t *testing.T) {
	e := echo.New()
	voucherID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/vouchers/"+voucherID.String()+"/edit-inline", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(voucherID.String())

	userID := uuid.New()
	user := &models.User{
		ID:    userID,
		Email: "test@example.com",
	}
	c.Set("current_user", user)

	mockAuthz := new(MockAuthzService)
	perms := &services.ResourcePermissions{
		CanView: true,
		CanEdit: false, // Not allowed to edit
		IsOwner: false,
	}
	mockAuthz.On("CheckVoucherAccess", mock.Anything, userID, voucherID).Return(perms, nil)

	handler := &Handler{
		authzService: mockAuthz,
	}

	err := handler.EditInline(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockAuthz.AssertExpectations(t)
}

func TestEditInline_InvalidID(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/vouchers/invalid-id/edit-inline", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("invalid-id")

	user := &models.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}
	c.Set("current_user", user)

	handler := &Handler{}

	err := handler.EditInline(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCancelEdit_Success(t *testing.T) {
	e := echo.New()
	voucherID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/vouchers/"+voucherID.String()+"/cancel-edit", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(voucherID.String())

	// Setup i18n context
	localizer := savvyi18n.NewLocalizer("de")
	ctx := savvyi18n.SetLocalizer(c.Request().Context(), localizer)
	c.SetRequest(c.Request().WithContext(ctx))

	userID := uuid.New()
	user := &models.User{
		ID:    userID,
		Email: "test@example.com",
	}
	c.Set("current_user", user)
	c.Set("csrf", "test-csrf-token")

	mockAuthz := new(MockAuthzService)
	perms := &services.ResourcePermissions{
		CanView: true,
		CanEdit: true,
		IsOwner: true,
	}
	mockAuthz.On("CheckVoucherAccess", mock.Anything, userID, voucherID).Return(perms, nil)

	mockVoucherService := new(MockVoucherService)
	voucher := &models.Voucher{
		ID:         voucherID,
		UserID:     &userID,
		Code:       "VOUCHER123",
		ValidFrom:  time.Now(),
		ValidUntil: time.Now().Add(24 * time.Hour),
	}
	mockVoucherService.On("GetVoucher", mock.Anything, voucherID).Return(voucher, nil)

	mockFavoriteService := new(MockFavoriteService)
	mockFavoriteService.On("IsFavorite", mock.Anything, userID, "voucher", voucherID).Return(false, nil)

	handler := &Handler{
		authzService:    mockAuthz,
		voucherService:  mockVoucherService,
		favoriteService: mockFavoriteService,
	}

	err := handler.CancelEdit(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockAuthz.AssertExpectations(t)
	mockVoucherService.AssertExpectations(t)
	mockFavoriteService.AssertExpectations(t)
}

func TestCancelEdit_InvalidID(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/vouchers/invalid-id/cancel-edit", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("invalid-id")

	user := &models.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}
	c.Set("current_user", user)

	handler := &Handler{}

	err := handler.CancelEdit(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUpdateInline_Success(t *testing.T) {
	e := echo.New()
	voucherID := uuid.New()
	merchantID := uuid.New()

	validFrom := time.Now().Format("2006-01-02")
	validUntil := time.Now().Add(30 * 24 * time.Hour).Format("2006-01-02")

	formData := url.Values{}
	formData.Set("code", "NEWCODE")
	formData.Set("type", "percentage")
	formData.Set("value", "20.00")
	formData.Set("description", "20% off")
	formData.Set("min_purchase_amount", "50.00")
	formData.Set("valid_from", validFrom)
	formData.Set("valid_until", validUntil)
	formData.Set("usage_limit_type", "unlimited")
	formData.Set("barcode_type", "CODE128")
	formData.Set("merchant_id", merchantID.String())

	req := httptest.NewRequest(http.MethodPost, "/vouchers/"+voucherID.String()+"/update-inline", strings.NewReader(formData.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(voucherID.String())

	// Setup i18n context
	localizer := savvyi18n.NewLocalizer("de")
	ctx := savvyi18n.SetLocalizer(c.Request().Context(), localizer)
	c.SetRequest(c.Request().WithContext(ctx))

	userID := uuid.New()
	user := &models.User{
		ID:    userID,
		Email: "test@example.com",
	}
	c.Set("current_user", user)
	c.Set("csrf", "test-csrf-token")

	mockAuthz := new(MockAuthzService)
	perms := &services.ResourcePermissions{
		CanView: true,
		CanEdit: true,
		IsOwner: true,
	}
	mockAuthz.On("CheckVoucherAccess", mock.Anything, userID, voucherID).Return(perms, nil)

	mockVoucherService := new(MockVoucherService)
	voucher := &models.Voucher{
		ID:         voucherID,
		UserID:     &userID,
		Code:       "OLDCODE",
		ValidFrom:  time.Now(),
		ValidUntil: time.Now().Add(24 * time.Hour),
	}
	mockVoucherService.On("GetVoucher", mock.Anything, voucherID).Return(voucher, nil).Twice()
	mockVoucherService.On("UpdateVoucher", mock.Anything, mock.AnythingOfType("*models.Voucher")).Return(nil)

	mockMerchantService := new(MockMerchantService)
	merchant := &models.Merchant{
		ID:   merchantID,
		Name: "Test Merchant",
	}
	mockMerchantService.On("GetMerchantByID", mock.Anything, merchantID).Return(merchant, nil)

	mockFavoriteService := new(MockFavoriteService)
	mockFavoriteService.On("IsFavorite", mock.Anything, userID, "voucher", voucherID).Return(false, nil)

	handler := &Handler{
		authzService:    mockAuthz,
		voucherService:  mockVoucherService,
		merchantService: mockMerchantService,
		favoriteService: mockFavoriteService,
	}

	err := handler.UpdateInline(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockAuthz.AssertExpectations(t)
	mockVoucherService.AssertExpectations(t)
	mockMerchantService.AssertExpectations(t)
	mockFavoriteService.AssertExpectations(t)
}

func TestUpdateInline_Forbidden(t *testing.T) {
	e := echo.New()
	voucherID := uuid.New()

	formData := url.Values{}
	formData.Set("code", "NEWCODE")

	req := httptest.NewRequest(http.MethodPost, "/vouchers/"+voucherID.String()+"/update-inline", strings.NewReader(formData.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(voucherID.String())

	userID := uuid.New()
	user := &models.User{
		ID:    userID,
		Email: "test@example.com",
	}
	c.Set("current_user", user)

	mockAuthz := new(MockAuthzService)
	perms := &services.ResourcePermissions{
		CanView: true,
		CanEdit: false, // Not allowed to edit
		IsOwner: false,
	}
	mockAuthz.On("CheckVoucherAccess", mock.Anything, userID, voucherID).Return(perms, nil)

	handler := &Handler{
		authzService: mockAuthz,
	}

	err := handler.UpdateInline(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockAuthz.AssertExpectations(t)
}

func TestUpdateInline_InvalidDateRange(t *testing.T) {
	e := echo.New()
	voucherID := uuid.New()

	formData := url.Values{}
	formData.Set("code", "NEWCODE")
	formData.Set("valid_from", "2024-12-31")
	formData.Set("valid_until", "2024-01-01") // Invalid: end before start

	req := httptest.NewRequest(http.MethodPost, "/vouchers/"+voucherID.String()+"/update-inline", strings.NewReader(formData.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(voucherID.String())

	userID := uuid.New()
	user := &models.User{
		ID:    userID,
		Email: "test@example.com",
	}
	c.Set("current_user", user)

	mockAuthz := new(MockAuthzService)
	perms := &services.ResourcePermissions{
		CanView: true,
		CanEdit: true,
		IsOwner: true,
	}
	mockAuthz.On("CheckVoucherAccess", mock.Anything, userID, voucherID).Return(perms, nil)

	mockVoucherService := new(MockVoucherService)
	voucher := &models.Voucher{
		ID:     voucherID,
		UserID: &userID,
		Code:   "OLDCODE",
	}
	mockVoucherService.On("GetVoucher", mock.Anything, voucherID).Return(voucher, nil)

	handler := &Handler{
		authzService:   mockAuthz,
		voucherService: mockVoucherService,
	}

	err := handler.UpdateInline(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockAuthz.AssertExpectations(t)
	mockVoucherService.AssertExpectations(t)
}
