package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"savvy/internal/mocks"
	"savvy/internal/models"
)

// ==================== Helper Functions ====================

func setupMerchantsTest() (*MerchantsHandler, *mocks.MockMerchantServiceInterface) {
	mockMerchantService := new(mocks.MockMerchantServiceInterface)
	handler := NewMerchantsHandler(mockMerchantService)
	return handler, mockMerchantService
}

// ==================== List Tests ====================

func TestMerchantsHandler_List_Success(t *testing.T) {
	handler, mockMerchantService := setupMerchantsTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/merchants", "")

	merchants := []models.Merchant{
		{ID: uuid.New(), Name: "IKEA", Color: "#0051BA"},
		{ID: uuid.New(), Name: "Amazon", Color: "#FF9900"},
	}
	mockMerchantService.On("GetAllMerchants", mock.Anything).Return(merchants, nil)

	err := handler.List(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string][]MerchantDTO
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Len(t, response["merchants"], 2)
	mockMerchantService.AssertExpectations(t)
}

func TestMerchantsHandler_List_ServiceError(t *testing.T) {
	handler, mockMerchantService := setupMerchantsTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/merchants", "")

	mockMerchantService.On("GetAllMerchants", mock.Anything).Return([]models.Merchant(nil), errors.New("database error"))

	err := handler.List(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockMerchantService.AssertExpectations(t)
}

// ==================== Search Tests ====================

func TestMerchantsHandler_Search_WithQuery(t *testing.T) {
	handler, mockMerchantService := setupMerchantsTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/merchants/search?q=ike", "")
	c.QueryParams().Set("q", "ike")

	merchants := []models.Merchant{
		{ID: uuid.New(), Name: "IKEA", Color: "#0051BA"},
	}
	mockMerchantService.On("SearchMerchants", mock.Anything, "ike").Return(merchants, nil)

	err := handler.Search(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string][]MerchantDTO
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Len(t, response["merchants"], 1)
	mockMerchantService.AssertExpectations(t)
}

func TestMerchantsHandler_Search_EmptyQuery(t *testing.T) {
	handler, mockMerchantService := setupMerchantsTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/merchants/search", "")

	merchants := []models.Merchant{
		{ID: uuid.New(), Name: "IKEA"},
	}
	mockMerchantService.On("GetAllMerchants", mock.Anything).Return(merchants, nil)

	err := handler.Search(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockMerchantService.AssertExpectations(t)
}

func TestMerchantsHandler_Search_ServiceError(t *testing.T) {
	handler, mockMerchantService := setupMerchantsTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/merchants/search?q=test", "")
	c.QueryParams().Set("q", "test")

	mockMerchantService.On("SearchMerchants", mock.Anything, "test").Return([]models.Merchant(nil), errors.New("database error"))

	err := handler.Search(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockMerchantService.AssertExpectations(t)
}

// ==================== Show Tests ====================

func TestMerchantsHandler_Show_Success(t *testing.T) {
	handler, mockMerchantService := setupMerchantsTest()
	merchantID := uuid.New()
	c, rec := createTestContext(http.MethodGet, "/api/v1/merchants/:id", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: merchantID.String()}})

	merchant := &models.Merchant{ID: merchantID, Name: "IKEA", Color: "#0051BA"}
	mockMerchantService.On("GetMerchantByID", mock.Anything, merchantID).Return(merchant, nil)

	err := handler.Show(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockMerchantService.AssertExpectations(t)
}

func TestMerchantsHandler_Show_InvalidID(t *testing.T) {
	handler, _ := setupMerchantsTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/merchants/:id", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "invalid-uuid"}})

	err := handler.Show(c)

	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestMerchantsHandler_Show_NotFound(t *testing.T) {
	handler, mockMerchantService := setupMerchantsTest()
	merchantID := uuid.New()
	c, rec := createTestContext(http.MethodGet, "/api/v1/merchants/:id", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: merchantID.String()}})

	mockMerchantService.On("GetMerchantByID", mock.Anything, merchantID).Return((*models.Merchant)(nil), errors.New("not found"))

	err := handler.Show(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockMerchantService.AssertExpectations(t)
}

// ==================== Create Tests (Admin) ====================

func TestMerchantsHandler_Create_Success(t *testing.T) {
	handler, mockMerchantService := setupMerchantsTest()
	body := `{"name":"New Merchant","color":"#FF0000"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/admin/merchants", body)

	mockMerchantService.On("CreateMerchant", mock.Anything, mock.AnythingOfType("*models.Merchant")).Return(nil)

	err := handler.Create(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	mockMerchantService.AssertExpectations(t)
}

func TestMerchantsHandler_Create_DefaultColor(t *testing.T) {
	handler, mockMerchantService := setupMerchantsTest()
	body := `{"name":"New Merchant"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/admin/merchants", body)

	mockMerchantService.On("CreateMerchant", mock.Anything, mock.MatchedBy(func(m *models.Merchant) bool {
		return m.Color == "#3B82F6"
	})).Return(nil)

	err := handler.Create(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	mockMerchantService.AssertExpectations(t)
}

func TestMerchantsHandler_Create_MissingName(t *testing.T) {
	handler, _ := setupMerchantsTest()
	body := `{"color":"#FF0000"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/admin/merchants", body)

	err := handler.Create(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestMerchantsHandler_Create_InvalidJSON(t *testing.T) {
	handler, _ := setupMerchantsTest()
	c, rec := createTestContext(http.MethodPost, "/api/v1/admin/merchants", "invalid-json")

	err := handler.Create(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestMerchantsHandler_Create_DuplicateName(t *testing.T) {
	handler, mockMerchantService := setupMerchantsTest()
	body := `{"name":"Existing Merchant"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/admin/merchants", body)

	mockMerchantService.On("CreateMerchant", mock.Anything, mock.AnythingOfType("*models.Merchant")).
		Return(errors.New("merchant with this name already exists"))

	err := handler.Create(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusConflict, rec.Code)
	mockMerchantService.AssertExpectations(t)
}

func TestMerchantsHandler_Create_ServiceError(t *testing.T) {
	handler, mockMerchantService := setupMerchantsTest()
	body := `{"name":"New Merchant"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/admin/merchants", body)

	mockMerchantService.On("CreateMerchant", mock.Anything, mock.AnythingOfType("*models.Merchant")).
		Return(errors.New("database error"))

	err := handler.Create(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockMerchantService.AssertExpectations(t)
}

// ==================== Update Tests (Admin) ====================

func TestMerchantsHandler_Update_Success(t *testing.T) {
	handler, mockMerchantService := setupMerchantsTest()
	merchantID := uuid.New()
	body := `{"name":"Updated Merchant"}`
	c, rec := createTestContext(http.MethodPatch, "/api/v1/admin/merchants/:id", body)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: merchantID.String()}})

	merchant := &models.Merchant{ID: merchantID, Name: "Old Merchant", Color: "#3B82F6"}
	mockMerchantService.On("GetMerchantByID", mock.Anything, merchantID).Return(merchant, nil)
	mockMerchantService.On("UpdateMerchant", mock.Anything, mock.AnythingOfType("*models.Merchant")).Return(nil)

	err := handler.Update(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockMerchantService.AssertExpectations(t)
}

func TestMerchantsHandler_Update_NotFound(t *testing.T) {
	handler, mockMerchantService := setupMerchantsTest()
	merchantID := uuid.New()
	body := `{"name":"Updated"}`
	c, rec := createTestContext(http.MethodPatch, "/api/v1/admin/merchants/:id", body)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: merchantID.String()}})

	mockMerchantService.On("GetMerchantByID", mock.Anything, merchantID).Return((*models.Merchant)(nil), errors.New("not found"))

	err := handler.Update(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockMerchantService.AssertExpectations(t)
}

func TestMerchantsHandler_Update_DuplicateName(t *testing.T) {
	handler, mockMerchantService := setupMerchantsTest()
	merchantID := uuid.New()
	body := `{"name":"Existing Name"}`
	c, rec := createTestContext(http.MethodPatch, "/api/v1/admin/merchants/:id", body)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: merchantID.String()}})

	merchant := &models.Merchant{ID: merchantID, Name: "Old Merchant"}
	mockMerchantService.On("GetMerchantByID", mock.Anything, merchantID).Return(merchant, nil)
	mockMerchantService.On("UpdateMerchant", mock.Anything, mock.AnythingOfType("*models.Merchant")).
		Return(errors.New("merchant with this name already exists"))

	err := handler.Update(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusConflict, rec.Code)
	mockMerchantService.AssertExpectations(t)
}

// ==================== Delete Tests (Admin) ====================

func TestMerchantsHandler_Delete_Success(t *testing.T) {
	handler, mockMerchantService := setupMerchantsTest()
	merchantID := uuid.New()
	c, rec := createTestContext(http.MethodDelete, "/api/v1/admin/merchants/:id", "")
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: merchantID.String()}})

	merchant := &models.Merchant{ID: merchantID, Name: "Test Merchant"}
	mockMerchantService.On("GetMerchantByID", mock.Anything, merchantID).Return(merchant, nil)
	mockMerchantService.On("DeleteMerchant", mock.Anything, merchantID).Return(nil)

	err := handler.Delete(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockMerchantService.AssertExpectations(t)
}

func TestMerchantsHandler_Delete_NotFound(t *testing.T) {
	handler, mockMerchantService := setupMerchantsTest()
	merchantID := uuid.New()
	c, rec := createTestContext(http.MethodDelete, "/api/v1/admin/merchants/:id", "")
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: merchantID.String()}})

	mockMerchantService.On("GetMerchantByID", mock.Anything, merchantID).Return((*models.Merchant)(nil), errors.New("not found"))

	err := handler.Delete(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockMerchantService.AssertExpectations(t)
}

func TestMerchantsHandler_Delete_ServiceError(t *testing.T) {
	handler, mockMerchantService := setupMerchantsTest()
	merchantID := uuid.New()
	c, rec := createTestContext(http.MethodDelete, "/api/v1/admin/merchants/:id", "")
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: merchantID.String()}})

	merchant := &models.Merchant{ID: merchantID, Name: "Test Merchant"}
	mockMerchantService.On("GetMerchantByID", mock.Anything, merchantID).Return(merchant, nil)
	mockMerchantService.On("DeleteMerchant", mock.Anything, merchantID).Return(errors.New("database error"))

	err := handler.Delete(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockMerchantService.AssertExpectations(t)
}
