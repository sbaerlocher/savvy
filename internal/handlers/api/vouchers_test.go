package api //nolint:revive // "api" is a meaningful package name for API handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"savvy/internal/mocks"
	"savvy/internal/models"
	"savvy/internal/services"
)

// ==================== Helper Functions ====================

func setupVouchersTest() (*VouchersHandler, *mocks.MockVoucherServiceInterface, *mocks.MockAuthzServiceInterface, *mocks.MockMerchantServiceInterface, *mocks.MockUserServiceInterface, *mocks.MockFavoriteServiceInterface, *mocks.MockShareServiceInterface, *mocks.MockTransferServiceInterface) {
	mockVoucherService := new(mocks.MockVoucherServiceInterface)
	mockAuthzService := new(mocks.MockAuthzServiceInterface)
	mockMerchantService := new(mocks.MockMerchantServiceInterface)
	mockUserService := new(mocks.MockUserServiceInterface)
	mockFavoriteService := new(mocks.MockFavoriteServiceInterface)
	mockShareService := new(mocks.MockShareServiceInterface)
	mockTransferService := new(mocks.MockTransferServiceInterface)

	handler := NewVouchersHandler(
		mockVoucherService,
		mockAuthzService,
		mockMerchantService,
		mockUserService,
		mockFavoriteService,
		mockShareService,
		mockTransferService,
	)

	return handler, mockVoucherService, mockAuthzService, mockMerchantService, mockUserService, mockFavoriteService, mockShareService, mockTransferService
}

func createTestVoucher() *models.Voucher {
	voucherID := uuid.New()
	userID := uuid.New()
	merchantID := uuid.New()
	validFrom := time.Now()
	validUntil := time.Now().Add(30 * 24 * time.Hour)
	return &models.Voucher{
		ID:                voucherID,
		UserID:            &userID,
		MerchantID:        &merchantID,
		MerchantName:      "Test Merchant",
		Code:              "VOUCHER123",
		Type:              "percentage",
		Value:             10.0,
		MinPurchaseAmount: 50.0,
		ValidFrom:         validFrom,
		ValidUntil:        validUntil,
		UsageLimitType:    "single_use",
		BarcodeType:       "CODE128",
	}
}

// ==================== List Tests ====================

func TestVouchersHandler_List_Success(t *testing.T) {
	handler, mockVoucherService, _, _, _, mockFavoriteService, mockShareService, _ := setupVouchersTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/vouchers", "")
	user := createTestUser()
	c.Set("current_user", user)

	vouchers := []models.Voucher{*createTestVoucher()}
	mockVoucherService.On("GetUserVouchers", mock.Anything, user.ID).Return(vouchers, nil)
	mockFavoriteService.On("IsFavorite", mock.Anything, user.ID, "voucher", vouchers[0].ID).Return(false, nil)
	mockShareService.On("GetVoucherShareCounts", mock.Anything, mock.Anything).Return(map[uuid.UUID]int64{}, nil)

	err := handler.List(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response VoucherListResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Len(t, response.Vouchers, 1)
	mockVoucherService.AssertExpectations(t)
	mockFavoriteService.AssertExpectations(t)
}

func TestVouchersHandler_List_ServiceError(t *testing.T) {
	handler, mockVoucherService, _, _, _, _, _, _ := setupVouchersTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/vouchers", "")
	user := createTestUser()
	c.Set("current_user", user)

	mockVoucherService.On("GetUserVouchers", mock.Anything, user.ID).Return([]models.Voucher(nil), errors.New("database error"))

	err := handler.List(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockVoucherService.AssertExpectations(t)
}

// ==================== Show Tests ====================

func TestVouchersHandler_Show_Success(t *testing.T) {
	handler, mockVoucherService, mockAuthzService, _, _, mockFavoriteService, mockShareService, _ := setupVouchersTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/vouchers/:id", "")
	user := createTestUser()
	voucher := createTestVoucher()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: voucher.ID.String()}})

	perms := &services.ResourcePermissions{
		CanView:   true,
		CanEdit:   true,
		CanDelete: true,
		IsOwner:   true,
	}

	mockAuthzService.On("CheckVoucherAccess", mock.Anything, user.ID, voucher.ID).Return(perms, nil)
	mockVoucherService.On("GetVoucher", mock.Anything, voucher.ID).Return(voucher, nil)
	mockFavoriteService.On("IsFavorite", mock.Anything, user.ID, "voucher", voucher.ID).Return(false, nil)
	mockShareService.On("GetVoucherShares", mock.Anything, voucher.ID).Return([]models.VoucherShare{}, nil)

	err := handler.Show(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response VoucherDetailResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, voucher.ID.String(), response.Voucher.ID)
	assert.True(t, response.Permissions.IsOwner)
	mockAuthzService.AssertExpectations(t)
	mockVoucherService.AssertExpectations(t)
	mockFavoriteService.AssertExpectations(t)
	mockShareService.AssertExpectations(t)
}

func TestVouchersHandler_Show_InvalidID(t *testing.T) {
	handler, _, _, _, _, _, _, _ := setupVouchersTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/vouchers/:id", "")
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "invalid-uuid"}})

	err := handler.Show(c)

	// parseResourceID writes JSON response AND returns error
	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestVouchersHandler_Show_Forbidden(t *testing.T) {
	handler, _, mockAuthzService, _, _, _, _, _ := setupVouchersTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/vouchers/:id", "")
	user := createTestUser()
	voucherID := uuid.New()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: voucherID.String()}})

	mockAuthzService.On("CheckVoucherAccess", mock.Anything, user.ID, voucherID).Return((*services.ResourcePermissions)(nil), errors.New("forbidden"))

	err := handler.Show(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockAuthzService.AssertExpectations(t)
}

func TestVouchersHandler_Show_NotFound(t *testing.T) {
	handler, mockVoucherService, mockAuthzService, _, _, _, _, _ := setupVouchersTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/vouchers/:id", "")
	user := createTestUser()
	voucherID := uuid.New()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: voucherID.String()}})

	perms := &services.ResourcePermissions{CanView: true}
	mockAuthzService.On("CheckVoucherAccess", mock.Anything, user.ID, voucherID).Return(perms, nil)
	mockVoucherService.On("GetVoucher", mock.Anything, voucherID).Return((*models.Voucher)(nil), errors.New("not found"))

	err := handler.Show(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockAuthzService.AssertExpectations(t)
	mockVoucherService.AssertExpectations(t)
}

// ==================== Create Tests ====================

func TestVouchersHandler_Create_Success(t *testing.T) {
	handler, mockVoucherService, _, mockMerchantService, _, _, _, _ := setupVouchersTest()
	merchantID := uuid.New()
	validFrom := time.Now().Format(time.RFC3339)
	validUntil := time.Now().Add(30 * 24 * time.Hour).Format(time.RFC3339)
	body := `{"merchant_id":"` + merchantID.String() + `","code":"VOUCHER123","type":"percentage","value":10,"valid_from":"` + validFrom + `","valid_until":"` + validUntil + `"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/vouchers", body)
	user := createTestUser()
	c.Set("current_user", user)

	merchant := &models.Merchant{ID: merchantID, Name: "Test Merchant"}
	mockMerchantService.On("GetMerchantByID", mock.Anything, merchantID).Return(merchant, nil)
	mockVoucherService.On("CheckDuplicate", mock.Anything, "VOUCHER123", user.ID, (*uuid.UUID)(nil)).Return((*models.Voucher)(nil), nil)
	mockVoucherService.On("FindDeletedDuplicate", mock.Anything, "VOUCHER123", user.ID).Return((*models.Voucher)(nil), nil)
	mockVoucherService.On("CreateVoucher", mock.Anything, mock.AnythingOfType("*models.Voucher")).Return(nil)
	mockVoucherService.On("GetVoucher", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(createTestVoucher(), nil)

	err := handler.Create(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	mockMerchantService.AssertExpectations(t)
	mockVoucherService.AssertExpectations(t)
}

func TestVouchersHandler_Create_InvalidRequest(t *testing.T) {
	handler, _, _, _, _, _, _, _ := setupVouchersTest()
	c, rec := createTestContext(http.MethodPost, "/api/v1/vouchers", "invalid-json")
	user := createTestUser()
	c.Set("current_user", user)

	err := handler.Create(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestVouchersHandler_Create_MissingFields(t *testing.T) {
	handler, _, _, _, _, _, _, _ := setupVouchersTest()
	body := `{"code":"VOUCHER123"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/vouchers", body)
	user := createTestUser()
	c.Set("current_user", user)

	err := handler.Create(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestVouchersHandler_Create_InvalidDateFormat(t *testing.T) {
	handler, _, _, _, _, _, _, _ := setupVouchersTest()
	body := `{"code":"VOUCHER123","type":"percentage","value":10,"valid_from":"invalid-date","valid_until":"2025-12-31"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/vouchers", body)
	user := createTestUser()
	c.Set("current_user", user)

	err := handler.Create(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestVouchersHandler_Create_WithNewMerchant(t *testing.T) {
	handler, mockVoucherService, _, mockMerchantService, _, _, _, _ := setupVouchersTest()
	validFrom := time.Now().Format(time.RFC3339)
	validUntil := time.Now().Add(30 * 24 * time.Hour).Format(time.RFC3339)
	body := `{"new_merchant_name":"New Merchant","code":"VOUCHER123","type":"percentage","value":10,"valid_from":"` + validFrom + `","valid_until":"` + validUntil + `"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/vouchers", body)
	user := createTestUser()
	c.Set("current_user", user)

	mockMerchantService.On("CreateMerchant", mock.Anything, mock.AnythingOfType("*models.Merchant")).Return(nil)
	mockVoucherService.On("CheckDuplicate", mock.Anything, "VOUCHER123", user.ID, (*uuid.UUID)(nil)).Return((*models.Voucher)(nil), nil)
	mockVoucherService.On("FindDeletedDuplicate", mock.Anything, "VOUCHER123", user.ID).Return((*models.Voucher)(nil), nil)
	mockVoucherService.On("CreateVoucher", mock.Anything, mock.AnythingOfType("*models.Voucher")).Return(nil)
	mockVoucherService.On("GetVoucher", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(createTestVoucher(), nil)

	err := handler.Create(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	mockMerchantService.AssertExpectations(t)
	mockVoucherService.AssertExpectations(t)
}

// ==================== Update Tests ====================

func TestVouchersHandler_Update_Success(t *testing.T) {
	handler, mockVoucherService, mockAuthzService, _, _, mockFavoriteService, mockShareService, _ := setupVouchersTest()
	voucher := createTestVoucher()
	body := `{"code":"NEWCODE456"}`
	c, rec := createTestContext(http.MethodPut, "/api/v1/vouchers/:id", body)
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: voucher.ID.String()}})

	perms := &services.ResourcePermissions{
		CanView:   true,
		CanEdit:   true,
		CanDelete: true,
		IsOwner:   true,
	}

	mockAuthzService.On("CheckVoucherAccess", mock.Anything, user.ID, voucher.ID).Return(perms, nil)
	mockVoucherService.On("GetVoucher", mock.Anything, voucher.ID).Return(voucher, nil).Twice()
	mockVoucherService.On("CheckDuplicate", mock.Anything, "NEWCODE456", user.ID, &voucher.ID).Return((*models.Voucher)(nil), nil)
	mockVoucherService.On("UpdateVoucher", mock.Anything, mock.AnythingOfType("*models.Voucher")).Return(nil)
	mockFavoriteService.On("IsFavorite", mock.Anything, user.ID, "voucher", voucher.ID).Return(false, nil)
	mockShareService.On("GetVoucherShares", mock.Anything, voucher.ID).Return([]models.VoucherShare{}, nil)

	err := handler.Update(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockAuthzService.AssertExpectations(t)
	mockVoucherService.AssertExpectations(t)
	mockFavoriteService.AssertExpectations(t)
	mockShareService.AssertExpectations(t)
}

func TestVouchersHandler_Update_Forbidden(t *testing.T) {
	handler, _, mockAuthzService, _, _, _, _, _ := setupVouchersTest()
	voucherID := uuid.New()
	c, rec := createTestContext(http.MethodPut, "/api/v1/vouchers/:id", `{}`)
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: voucherID.String()}})

	perms := &services.ResourcePermissions{CanView: true, CanEdit: false}
	mockAuthzService.On("CheckVoucherAccess", mock.Anything, user.ID, voucherID).Return(perms, nil)

	err := handler.Update(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockAuthzService.AssertExpectations(t)
}

func TestVouchersHandler_Update_NotFound(t *testing.T) {
	handler, mockVoucherService, mockAuthzService, _, _, _, _, _ := setupVouchersTest()
	voucherID := uuid.New()
	c, rec := createTestContext(http.MethodPut, "/api/v1/vouchers/:id", `{}`)
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: voucherID.String()}})

	perms := &services.ResourcePermissions{CanEdit: true}
	mockAuthzService.On("CheckVoucherAccess", mock.Anything, user.ID, voucherID).Return(perms, nil)
	mockVoucherService.On("GetVoucher", mock.Anything, voucherID).Return((*models.Voucher)(nil), errors.New("not found"))

	err := handler.Update(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockAuthzService.AssertExpectations(t)
	mockVoucherService.AssertExpectations(t)
}

// ==================== Delete Tests ====================

func TestVouchersHandler_Delete_Success(t *testing.T) {
	handler, mockVoucherService, mockAuthzService, _, _, _, _, _ := setupVouchersTest()
	voucherID := uuid.New()
	c, rec := createTestContext(http.MethodDelete, "/api/v1/vouchers/:id", "")
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: voucherID.String()}})

	perms := &services.ResourcePermissions{
		CanView:   true,
		CanDelete: true,
	}

	mockAuthzService.On("CheckVoucherAccess", mock.Anything, user.ID, voucherID).Return(perms, nil)
	mockVoucherService.On("DeleteVoucher", mock.Anything, voucherID).Return(nil)

	err := handler.Delete(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockAuthzService.AssertExpectations(t)
	mockVoucherService.AssertExpectations(t)
}

// ==================== ToggleFavorite Tests ====================

func TestVouchersHandler_ToggleFavorite_Success(t *testing.T) {
	handler, _, mockAuthzService, _, _, mockFavoriteService, _, _ := setupVouchersTest()
	voucherID := uuid.New()
	c, rec := createTestContext(http.MethodPost, "/api/v1/vouchers/:id/favorite", "")
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: voucherID.String()}})

	perms := &services.ResourcePermissions{CanView: true}
	mockAuthzService.On("CheckVoucherAccess", mock.Anything, user.ID, voucherID).Return(perms, nil)
	mockFavoriteService.On("ToggleFavorite", mock.Anything, user.ID, "voucher", voucherID).Return(nil)
	mockFavoriteService.On("IsFavorite", mock.Anything, user.ID, "voucher", voucherID).Return(true, nil)

	err := handler.ToggleFavorite(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]bool
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.True(t, response["is_favorite"])
	mockAuthzService.AssertExpectations(t)
	mockFavoriteService.AssertExpectations(t)
}

// ==================== CreateShare Tests ====================

func TestVouchersHandler_CreateShare_Success(t *testing.T) {
	handler, _, mockAuthzService, _, mockUserService, _, mockShareService, _ := setupVouchersTest()
	voucherID := uuid.New()
	sharedUserID := uuid.New()
	body := `{"emails":["shared@example.com"]}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/vouchers/:id/share", body)
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: voucherID.String()}})

	perms := &services.ResourcePermissions{IsOwner: true}
	sharedUser := &models.User{ID: sharedUserID, Email: "shared@example.com"}

	mockAuthzService.On("CheckVoucherAccess", mock.Anything, user.ID, voucherID).Return(perms, nil)
	mockUserService.On("GetUserByEmail", mock.Anything, "shared@example.com").Return(sharedUser, nil)
	mockShareService.On("CreateVoucherShare", mock.Anything, mock.Anything, voucherID, sharedUserID).Return(nil)
	mockShareService.On("GetVoucherShares", mock.Anything, voucherID).Return([]models.VoucherShare{}, nil)

	err := handler.CreateShare(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	var resp ShareCreateResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, 1, resp.SuccessCount)
	assert.Empty(t, resp.Failed)
	mockAuthzService.AssertExpectations(t)
	mockUserService.AssertExpectations(t)
	mockShareService.AssertExpectations(t)
}

func TestVouchersHandler_CreateShare_NotOwner(t *testing.T) {
	handler, _, mockAuthzService, _, _, _, _, _ := setupVouchersTest()
	voucherID := uuid.New()
	body := `{"email":"shared@example.com"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/vouchers/:id/share", body)
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: voucherID.String()}})

	perms := &services.ResourcePermissions{IsOwner: false}
	mockAuthzService.On("CheckVoucherAccess", mock.Anything, user.ID, voucherID).Return(perms, nil)

	err := handler.CreateShare(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockAuthzService.AssertExpectations(t)
}

func TestVouchersHandler_CreateShare_SelfShare(t *testing.T) {
	handler, _, mockAuthzService, _, mockUserService, _, mockShareService, _ := setupVouchersTest()
	voucherID := uuid.New()
	body := `{"emails":["test@example.com"]}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/vouchers/:id/share", body)
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: voucherID.String()}})

	perms := &services.ResourcePermissions{IsOwner: true}
	mockAuthzService.On("CheckVoucherAccess", mock.Anything, user.ID, voucherID).Return(perms, nil)
	mockUserService.On("GetUserByEmail", mock.Anything, "test@example.com").Return(user, nil)
	mockShareService.On("GetVoucherShares", mock.Anything, voucherID).Return([]models.VoucherShare{}, nil)

	err := handler.CreateShare(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	var resp ShareCreateResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, 0, resp.SuccessCount)
	assert.Len(t, resp.Failed, 1)
	mockAuthzService.AssertExpectations(t)
	mockUserService.AssertExpectations(t)
}

// ==================== DeleteShare Tests ====================

func TestVouchersHandler_DeleteShare_Success(t *testing.T) {
	handler, _, mockAuthzService, _, _, _, mockShareService, _ := setupVouchersTest()
	voucherID := uuid.New()
	sharedWithID := uuid.New()
	c, rec := createTestContext(http.MethodDelete, "/api/v1/vouchers/:id/share/:sharedWithID", "")
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: voucherID.String()}, {Name: "sharedWithID", Value: sharedWithID.String()}})

	perms := &services.ResourcePermissions{IsOwner: true}
	mockAuthzService.On("CheckVoucherAccess", mock.Anything, user.ID, voucherID).Return(perms, nil)
	mockShareService.On("DeleteVoucherShare", mock.Anything, mock.Anything, voucherID, sharedWithID).Return(nil)

	err := handler.DeleteShare(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockAuthzService.AssertExpectations(t)
	mockShareService.AssertExpectations(t)
}

// ==================== DeleteAllShares Tests ====================

func TestVouchersHandler_DeleteAllShares_Success(t *testing.T) {
	handler, _, mockAuthzService, _, _, _, mockShareService, _ := setupVouchersTest()
	voucherID := uuid.New()
	c, rec := createTestContext(http.MethodDelete, "/api/v1/vouchers/:id/shares", "")
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: voucherID.String()}})

	perms := &services.ResourcePermissions{IsOwner: true}
	mockAuthzService.On("CheckVoucherAccess", mock.Anything, user.ID, voucherID).Return(perms, nil)
	mockShareService.On("DeleteAllVoucherShares", mock.Anything, mock.Anything, voucherID).Return(nil)

	err := handler.DeleteAllShares(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockAuthzService.AssertExpectations(t)
	mockShareService.AssertExpectations(t)
}

func TestVouchersHandler_DeleteAllShares_NotOwner(t *testing.T) {
	handler, _, mockAuthzService, _, _, _, mockShareService, _ := setupVouchersTest()
	voucherID := uuid.New()
	c, rec := createTestContext(http.MethodDelete, "/api/v1/vouchers/:id/shares", "")
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: voucherID.String()}})

	perms := &services.ResourcePermissions{IsOwner: false}
	mockAuthzService.On("CheckVoucherAccess", mock.Anything, user.ID, voucherID).Return(perms, nil)

	err := handler.DeleteAllShares(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockAuthzService.AssertExpectations(t)
	mockShareService.AssertNotCalled(t, "DeleteAllVoucherShares", mock.Anything, mock.Anything, voucherID)
}

// ==================== Transfer Tests ====================

func TestVouchersHandler_Transfer_Success(t *testing.T) {
	handler, _, mockAuthzService, _, mockUserService, _, _, mockTransferService := setupVouchersTest()
	voucherID := uuid.New()
	newOwnerID := uuid.New()
	body := `{"new_owner_email":"newowner@example.com"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/vouchers/:id/transfer", body)
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: voucherID.String()}})

	perms := &services.ResourcePermissions{IsOwner: true}
	newOwner := &models.User{ID: newOwnerID, Email: "newowner@example.com"}

	mockAuthzService.On("CheckVoucherAccess", mock.Anything, user.ID, voucherID).Return(perms, nil)
	mockUserService.On("GetUserByEmail", mock.Anything, "newowner@example.com").Return(newOwner, nil)
	mockTransferService.On("TransferVoucherOwnership", mock.Anything, voucherID, newOwnerID, user.ID).Return(nil)

	err := handler.Transfer(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockAuthzService.AssertExpectations(t)
	mockUserService.AssertExpectations(t)
	mockTransferService.AssertExpectations(t)
}

func TestVouchersHandler_Transfer_NotOwner(t *testing.T) {
	handler, _, mockAuthzService, _, _, _, _, _ := setupVouchersTest()
	voucherID := uuid.New()
	c, rec := createTestContext(http.MethodPost, "/api/v1/vouchers/:id/transfer", `{"new_owner_email":"test@example.com"}`)
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: voucherID.String()}})

	perms := &services.ResourcePermissions{IsOwner: false}
	mockAuthzService.On("CheckVoucherAccess", mock.Anything, user.ID, voucherID).Return(perms, nil)

	err := handler.Transfer(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockAuthzService.AssertExpectations(t)
}

// ==================== Error Message Sanitization Tests ====================

func TestVouchersHandler_CreateShare_ServiceError_NoLeakedDetails(t *testing.T) {
	handler, _, mockAuthzService, _, mockUserService, _, mockShareService, _ := setupVouchersTest()
	voucherID := uuid.New()
	sharedUserID := uuid.New()
	body := `{"emails":["shared@example.com"]}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/vouchers/:id/share", body)
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: voucherID.String()}})

	perms := &services.ResourcePermissions{IsOwner: true}
	sharedUser := &models.User{ID: sharedUserID, Email: "shared@example.com"}

	mockAuthzService.On("CheckVoucherAccess", mock.Anything, user.ID, voucherID).Return(perms, nil)
	mockUserService.On("GetUserByEmail", mock.Anything, "shared@example.com").Return(sharedUser, nil)
	// Simulate a GORM error with internal DB details
	mockShareService.On("CreateVoucherShare", mock.Anything, mock.Anything, voucherID, sharedUserID).
		Return(errors.New("ERROR: duplicate key value violates unique constraint \"voucher_shares_pkey\" (SQLSTATE 23505)"))
	mockShareService.On("GetVoucherShares", mock.Anything, voucherID).Return([]models.VoucherShare{}, nil)

	err := handler.CreateShare(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)

	// Failed entries carry a generic reason, never the raw DB error.
	body2 := rec.Body.String()
	assert.NotContains(t, body2, "duplicate key")
	assert.NotContains(t, body2, "voucher_shares_pkey")
	assert.NotContains(t, body2, "SQLSTATE")
	var resp ShareCreateResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, 0, resp.SuccessCount)
	assert.Len(t, resp.Failed, 1)
	assert.Equal(t, "share failed", resp.Failed[0].Error)
	mockShareService.AssertExpectations(t)
}

// ==================== floatOrDefault Tests ====================

func TestFloatOrDefault_Nil(t *testing.T) {
	result := floatOrDefault(nil, 42.0)
	assert.Equal(t, 42.0, result)
}

func TestFloatOrDefault_WithValue(t *testing.T) {
	val := 99.5
	result := floatOrDefault(&val, 42.0)
	assert.Equal(t, 99.5, result)
}

func TestFloatOrDefault_Zero(t *testing.T) {
	val := 0.0
	result := floatOrDefault(&val, 42.0)
	assert.Equal(t, 0.0, result)
}

// ==================== checkVoucherDuplicate Tests ====================

func TestCheckVoucherDuplicate_NilCode(t *testing.T) {
	result := checkVoucherDuplicate(nil, nil, nil, uuid.New(), nil)
	assert.Nil(t, result)
}

func TestCheckVoucherDuplicate_EmptyCode(t *testing.T) {
	empty := ""
	e := echo.New()
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	c := e.NewContext(req, nil)
	result := checkVoucherDuplicate(c, nil, &empty, uuid.New(), nil)
	assert.Nil(t, result)
}

func TestCheckVoucherDuplicate_NoDuplicate(t *testing.T) {
	mockSvc := new(mocks.MockVoucherServiceInterface)
	code := "ABC123"
	userID := uuid.New()
	e := echo.New()
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	c := e.NewContext(req, nil)

	mockSvc.On("CheckDuplicate", mock.Anything, code, userID, (*uuid.UUID)(nil)).Return((*models.Voucher)(nil), nil)

	result := checkVoucherDuplicate(c, mockSvc, &code, userID, nil)
	assert.Nil(t, result)
	mockSvc.AssertExpectations(t)
}

func TestCheckVoucherDuplicate_Found(t *testing.T) {
	mockSvc := new(mocks.MockVoucherServiceInterface)
	code := "ABC123"
	userID := uuid.New()
	dupID := uuid.New()
	e := echo.New()
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	c := e.NewContext(req, nil)

	dup := &models.Voucher{Code: "ABC123"}
	dup.ID = dupID
	dup.MerchantName = "TestMerchant"
	mockSvc.On("CheckDuplicate", mock.Anything, code, userID, (*uuid.UUID)(nil)).Return(dup, nil)

	result := checkVoucherDuplicate(c, mockSvc, &code, userID, nil)
	assert.NotNil(t, result)
	assert.True(t, result.HasDuplicate)
	assert.Equal(t, "TestMerchant", result.MerchantName)
	assert.Equal(t, "ABC123", result.ResourceNumber)
	mockSvc.AssertExpectations(t)
}

func TestCheckVoucherDuplicate_ServiceError(t *testing.T) {
	mockSvc := new(mocks.MockVoucherServiceInterface)
	code := "ABC123"
	userID := uuid.New()
	e := echo.New()
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	c := e.NewContext(req, nil)

	mockSvc.On("CheckDuplicate", mock.Anything, code, userID, (*uuid.UUID)(nil)).Return((*models.Voucher)(nil), errors.New("db error"))

	result := checkVoucherDuplicate(c, mockSvc, &code, userID, nil)
	assert.Nil(t, result)
	mockSvc.AssertExpectations(t)
}

// ==================== Restore Tests ====================

func TestVouchersHandler_Create_DeletedDuplicate_Returns409WithDeletedFlag(t *testing.T) {
	handler, mockVoucherService, _, mockMerchantService, _, _, _, _ := setupVouchersTest()
	merchantID := uuid.New()
	validFrom := time.Now().Format(time.RFC3339)
	validUntil := time.Now().Add(30 * 24 * time.Hour).Format(time.RFC3339)
	body := `{"merchant_id":"` + merchantID.String() + `","code":"VOUCHER123","type":"percentage","value":10,"valid_from":"` + validFrom + `","valid_until":"` + validUntil + `"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/vouchers", body)
	user := createTestUser()
	c.Set("current_user", user)

	deletedVoucher := createTestVoucher()
	deletedVoucher.ID = uuid.New()
	deletedVoucher.UserID = &user.ID
	deletedVoucher.Code = "VOUCHER123"
	deletedVoucher.MerchantName = "Test Merchant"

	merchant := &models.Merchant{ID: merchantID, Name: "Test Merchant"}
	mockMerchantService.On("GetMerchantByID", mock.Anything, merchantID).Return(merchant, nil)
	// No active duplicate
	mockVoucherService.On("CheckDuplicate", mock.Anything, "VOUCHER123", user.ID, (*uuid.UUID)(nil)).Return((*models.Voucher)(nil), nil)
	// Soft-deleted duplicate found
	mockVoucherService.On("FindDeletedDuplicate", mock.Anything, "VOUCHER123", user.ID).Return(deletedVoucher, nil)

	err := handler.Create(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusConflict, rec.Code)

	var response DuplicateErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "duplicate_barcode", response.Error)
	assert.NotNil(t, response.Duplicate)
	assert.True(t, response.Duplicate.Deleted, "expected duplicate.deleted == true")
	assert.Equal(t, deletedVoucher.ID.String(), response.Duplicate.ExistingID)
	mockMerchantService.AssertExpectations(t)
	mockVoucherService.AssertExpectations(t)
}

func TestVouchersHandler_Restore_Success(t *testing.T) {
	handler, mockVoucherService, _, _, _, mockFavoriteService, _, _ := setupVouchersTest()
	restoredVoucher := createTestVoucher()
	c, rec := createTestContext(http.MethodPost, "/api/v1/vouchers/:id/restore", "")
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: restoredVoucher.ID.String()}})

	mockVoucherService.On("RestoreVoucher", mock.Anything, restoredVoucher.ID, user.ID).Return(restoredVoucher, nil)
	mockFavoriteService.On("IsFavorite", mock.Anything, user.ID, "voucher", restoredVoucher.ID).Return(false, nil)

	err := handler.Restore(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response VoucherDetailResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, restoredVoucher.ID.String(), response.Voucher.ID)
	mockVoucherService.AssertExpectations(t)
	mockFavoriteService.AssertExpectations(t)
}

func TestVouchersHandler_Restore_NotFound(t *testing.T) {
	handler, mockVoucherService, _, _, _, _, _, _ := setupVouchersTest()
	voucherID := uuid.New()
	c, rec := createTestContext(http.MethodPost, "/api/v1/vouchers/:id/restore", "")
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: voucherID.String()}})

	mockVoucherService.On("RestoreVoucher", mock.Anything, voucherID, user.ID).Return((*models.Voucher)(nil), nil)

	err := handler.Restore(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockVoucherService.AssertExpectations(t)
}

func TestVouchersHandler_Restore_InvalidID(t *testing.T) {
	handler, _, _, _, _, _, _, _ := setupVouchersTest()
	c, rec := createTestContext(http.MethodPost, "/api/v1/vouchers/:id/restore", "")
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "invalid-uuid"}})

	err := handler.Restore(c)

	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestVouchersHandler_Restore_ServiceError(t *testing.T) {
	handler, mockVoucherService, _, _, _, _, _, _ := setupVouchersTest()
	voucherID := uuid.New()
	c, rec := createTestContext(http.MethodPost, "/api/v1/vouchers/:id/restore", "")
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: voucherID.String()}})

	mockVoucherService.On("RestoreVoucher", mock.Anything, voucherID, user.ID).Return((*models.Voucher)(nil), errors.New("db error"))

	err := handler.Restore(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockVoucherService.AssertExpectations(t)
}
