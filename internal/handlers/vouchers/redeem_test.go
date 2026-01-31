package vouchers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"savvy/internal/models"
	"savvy/internal/services"
)

func TestRedeem_Success(t *testing.T) {
	e := echo.New()
	voucherID := uuid.New()
	req := httptest.NewRequest(http.MethodPost, "/vouchers/"+voucherID.String()+"/redeem", nil)
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
	}
	mockAuthz.On("CheckVoucherAccess", mock.Anything, userID, voucherID).Return(perms, nil)

	mockVoucherService := new(MockVoucherService)
	voucher := &models.Voucher{
		ID:             voucherID,
		UserID:         &userID,
		Code:           "VOUCHER123",
		ValidFrom:      time.Now().Add(-24 * time.Hour),
		ValidUntil:     time.Now().Add(24 * time.Hour),
		UsageLimitType: "unlimited",
		UsedCount:      0,
	}
	mockVoucherService.On("GetVoucher", mock.Anything, voucherID).Return(voucher, nil)
	mockVoucherService.On("UpdateVoucher", mock.Anything, mock.AnythingOfType("*models.Voucher")).Return(nil)

	handler := &Handler{
		authzService:   mockAuthz,
		voucherService: mockVoucherService,
	}

	err := handler.Redeem(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Contains(t, rec.Header().Get("Location"), "/vouchers/"+voucherID.String())
	assert.Contains(t, rec.Header().Get("Location"), "success=redeemed")
	mockAuthz.AssertExpectations(t)
	mockVoucherService.AssertExpectations(t)
}

func TestRedeem_InvalidID(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/vouchers/invalid-id/redeem", nil)
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

	err := handler.Redeem(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/vouchers", rec.Header().Get("Location"))
}

func TestRedeem_Unauthorized(t *testing.T) {
	e := echo.New()
	voucherID := uuid.New()
	req := httptest.NewRequest(http.MethodPost, "/vouchers/"+voucherID.String()+"/redeem", nil)
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
	mockAuthz.On("CheckVoucherAccess", mock.Anything, userID, voucherID).Return(nil, errors.New("unauthorized"))

	handler := &Handler{
		authzService: mockAuthz,
	}

	err := handler.Redeem(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/vouchers", rec.Header().Get("Location"))
	mockAuthz.AssertExpectations(t)
}

func TestRedeem_VoucherNotFound(t *testing.T) {
	e := echo.New()
	voucherID := uuid.New()
	req := httptest.NewRequest(http.MethodPost, "/vouchers/"+voucherID.String()+"/redeem", nil)
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
	}
	mockAuthz.On("CheckVoucherAccess", mock.Anything, userID, voucherID).Return(perms, nil)

	mockVoucherService := new(MockVoucherService)
	mockVoucherService.On("GetVoucher", mock.Anything, voucherID).Return(nil, errors.New("not found"))

	handler := &Handler{
		authzService:   mockAuthz,
		voucherService: mockVoucherService,
	}

	err := handler.Redeem(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/vouchers", rec.Header().Get("Location"))
	mockAuthz.AssertExpectations(t)
	mockVoucherService.AssertExpectations(t)
}

func TestRedeem_ExpiredVoucher(t *testing.T) {
	e := echo.New()
	voucherID := uuid.New()
	req := httptest.NewRequest(http.MethodPost, "/vouchers/"+voucherID.String()+"/redeem", nil)
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
	}
	mockAuthz.On("CheckVoucherAccess", mock.Anything, userID, voucherID).Return(perms, nil)

	mockVoucherService := new(MockVoucherService)
	voucher := &models.Voucher{
		ID:             voucherID,
		UserID:         &userID,
		Code:           "EXPIRED123",
		ValidFrom:      time.Now().Add(-48 * time.Hour),
		ValidUntil:     time.Now().Add(-24 * time.Hour), // Expired
		UsageLimitType: "unlimited",
		UsedCount:      0,
	}
	mockVoucherService.On("GetVoucher", mock.Anything, voucherID).Return(voucher, nil)

	handler := &Handler{
		authzService:   mockAuthz,
		voucherService: mockVoucherService,
	}

	err := handler.Redeem(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Contains(t, rec.Header().Get("Location"), "/vouchers/"+voucherID.String())
	assert.Contains(t, rec.Header().Get("Location"), "error=cannot_redeem")
	mockAuthz.AssertExpectations(t)
	mockVoucherService.AssertExpectations(t)
}

func TestRedeem_UpdateError(t *testing.T) {
	e := echo.New()
	voucherID := uuid.New()
	req := httptest.NewRequest(http.MethodPost, "/vouchers/"+voucherID.String()+"/redeem", nil)
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
	}
	mockAuthz.On("CheckVoucherAccess", mock.Anything, userID, voucherID).Return(perms, nil)

	mockVoucherService := new(MockVoucherService)
	voucher := &models.Voucher{
		ID:             voucherID,
		UserID:         &userID,
		Code:           "VOUCHER123",
		ValidFrom:      time.Now().Add(-24 * time.Hour),
		ValidUntil:     time.Now().Add(24 * time.Hour),
		UsageLimitType: "unlimited",
		UsedCount:      0,
	}
	mockVoucherService.On("GetVoucher", mock.Anything, voucherID).Return(voucher, nil)
	mockVoucherService.On("UpdateVoucher", mock.Anything, mock.AnythingOfType("*models.Voucher")).Return(errors.New("database error"))

	handler := &Handler{
		authzService:   mockAuthz,
		voucherService: mockVoucherService,
	}

	err := handler.Redeem(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Contains(t, rec.Header().Get("Location"), "/vouchers/"+voucherID.String())
	assert.Contains(t, rec.Header().Get("Location"), "error=database_error")
	mockAuthz.AssertExpectations(t)
	mockVoucherService.AssertExpectations(t)
}
