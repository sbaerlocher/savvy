package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"savvy/internal/mocks"
	"savvy/internal/services"
)

// ==================== Helper Functions ====================

func setupImportTest() (*ImportHandler, *mocks.MockImportServiceInterface) {
	mockImportService := new(mocks.MockImportServiceInterface)
	handler := NewImportHandler(mockImportService)
	return handler, mockImportService
}

// ==================== ImportJSON Tests ====================

func TestImportHandler_ImportJSON_Success(t *testing.T) {
	handler, mockImportService := setupImportTest()
	body := `{"cards":[{"merchant_name":"IKEA","card_number":"123"}],"vouchers":[],"gift_cards":[]}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/import/json", body)
	user := createTestUser()
	c.Set("current_user", user)

	result := &services.ImportResult{
		CardsImported:    1,
		VouchersImported: 0,
	}
	mockImportService.On("ImportJSON", mock.Anything, user.ID, mock.AnythingOfType("*services.ExportData")).Return(result, nil)

	err := handler.ImportJSON(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockImportService.AssertExpectations(t)
}

func TestImportHandler_ImportJSON_Unauthorized(t *testing.T) {
	handler, _ := setupImportTest()
	body := `{"cards":[],"vouchers":[],"gift_cards":[]}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/import/json", body)
	// No user set

	err := handler.ImportJSON(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestImportHandler_ImportJSON_InvalidJSON(t *testing.T) {
	handler, _ := setupImportTest()
	c, rec := createTestContext(http.MethodPost, "/api/v1/import/json", "invalid-json")
	user := createTestUser()
	c.Set("current_user", user)

	err := handler.ImportJSON(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestImportHandler_ImportJSON_ServiceError(t *testing.T) {
	handler, mockImportService := setupImportTest()
	body := `{"cards":[{"merchant_name":"IKEA","card_number":"123"}],"vouchers":[],"gift_cards":[]}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/import/json", body)
	user := createTestUser()
	c.Set("current_user", user)

	mockImportService.On("ImportJSON", mock.Anything, user.ID, mock.AnythingOfType("*services.ExportData")).
		Return((*services.ImportResult)(nil), errors.New("import failed"))

	err := handler.ImportJSON(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockImportService.AssertExpectations(t)
}

// ==================== PreviewJSON Tests ====================

func TestImportHandler_PreviewJSON_Success(t *testing.T) {
	handler, mockImportService := setupImportTest()
	body := `{"cards":[{"merchant_name":"IKEA","card_number":"123"}],"vouchers":[],"gift_cards":[]}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/import/json/preview", body)
	user := createTestUser()
	c.Set("current_user", user)

	preview := &services.ImportPreview{
		Cards:     1,
		Vouchers:  0,
		GiftCards: 0,
	}
	mockImportService.On("PreviewJSON", mock.Anything, mock.AnythingOfType("*services.ExportData")).Return(preview, nil)

	err := handler.PreviewJSON(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response services.ImportPreview
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, 1, response.Cards)
	mockImportService.AssertExpectations(t)
}

func TestImportHandler_PreviewJSON_Unauthorized(t *testing.T) {
	handler, _ := setupImportTest()
	body := `{"cards":[],"vouchers":[],"gift_cards":[]}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/import/json/preview", body)

	err := handler.PreviewJSON(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestImportHandler_PreviewJSON_ServiceError(t *testing.T) {
	handler, mockImportService := setupImportTest()
	body := `{"cards":[{"merchant_name":"IKEA","card_number":"123"}],"vouchers":[],"gift_cards":[]}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/import/json/preview", body)
	user := createTestUser()
	c.Set("current_user", user)

	mockImportService.On("PreviewJSON", mock.Anything, mock.AnythingOfType("*services.ExportData")).
		Return((*services.ImportPreview)(nil), errors.New("preview failed"))

	err := handler.PreviewJSON(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockImportService.AssertExpectations(t)
}

// ==================== CSV Import Helper ====================

func createCSVTestContext(url, filename, csvContent string) (echo.Context, *httptest.ResponseRecorder) {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", filename)
	_, _ = part.Write([]byte(csvContent))
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, url, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(req, rec)
	return c, rec
}

// ==================== ImportCardsCSV Tests ====================

func TestImportHandler_ImportCardsCSV_Success(t *testing.T) {
	handler, mockImportService := setupImportTest()
	csvContent := "merchant_name,card_number\nIKEA,123456"
	c, rec := createCSVTestContext("/api/v1/import/csv/cards", "cards.csv", csvContent)
	user := createTestUser()
	c.Set("current_user", user)

	result := &services.ImportResult{CardsImported: 1}
	mockImportService.On("ImportCardsCSV", mock.Anything, user.ID, mock.Anything).Return(result, nil)

	err := handler.ImportCardsCSV(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockImportService.AssertExpectations(t)
}

func TestImportHandler_ImportCardsCSV_Unauthorized(t *testing.T) {
	handler, _ := setupImportTest()
	csvContent := "merchant_name,card_number\nIKEA,123456"
	c, rec := createCSVTestContext("/api/v1/import/csv/cards", "cards.csv", csvContent)

	err := handler.ImportCardsCSV(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestImportHandler_ImportCardsCSV_InvalidExtension(t *testing.T) {
	handler, _ := setupImportTest()
	c, rec := createCSVTestContext("/api/v1/import/csv/cards", "cards.json", "not csv")
	user := createTestUser()
	c.Set("current_user", user)

	err := handler.ImportCardsCSV(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "bad_request", response.Error)
	assert.Contains(t, response.Message, "CSV")
}

func TestImportHandler_ImportCardsCSV_ServiceError(t *testing.T) {
	handler, mockImportService := setupImportTest()
	csvContent := "merchant_name,card_number\nIKEA,123456"
	c, rec := createCSVTestContext("/api/v1/import/csv/cards", "cards.csv", csvContent)
	user := createTestUser()
	c.Set("current_user", user)

	mockImportService.On("ImportCardsCSV", mock.Anything, user.ID, mock.Anything).
		Return((*services.ImportResult)(nil), errors.New("invalid CSV format"))

	err := handler.ImportCardsCSV(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "import_failed", response.Error)
	mockImportService.AssertExpectations(t)
}

// ==================== ImportVouchersCSV Tests ====================

func TestImportHandler_ImportVouchersCSV_Success(t *testing.T) {
	handler, mockImportService := setupImportTest()
	csvContent := "merchant_name,code,type,value\nIKEA,ABC123,percentage,10"
	c, rec := createCSVTestContext("/api/v1/import/csv/vouchers", "vouchers.csv", csvContent)
	user := createTestUser()
	c.Set("current_user", user)

	result := &services.ImportResult{VouchersImported: 1}
	mockImportService.On("ImportVouchersCSV", mock.Anything, user.ID, mock.Anything).Return(result, nil)

	err := handler.ImportVouchersCSV(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockImportService.AssertExpectations(t)
}

// ==================== ImportGiftCardsCSV Tests ====================

func TestImportHandler_ImportGiftCardsCSV_Success(t *testing.T) {
	handler, mockImportService := setupImportTest()
	csvContent := "merchant_name,card_number,initial_balance,currency\nIKEA,GC123,100.00,CHF"
	c, rec := createCSVTestContext("/api/v1/import/csv/gift-cards", "gift-cards.csv", csvContent)
	user := createTestUser()
	c.Set("current_user", user)

	result := &services.ImportResult{GiftCardsImported: 1}
	mockImportService.On("ImportGiftCardsCSV", mock.Anything, user.ID, mock.Anything).Return(result, nil)

	err := handler.ImportGiftCardsCSV(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockImportService.AssertExpectations(t)
}

func TestImportHandler_ImportCardsCSV_NoFile(t *testing.T) {
	handler, _ := setupImportTest()
	// Create request without file upload
	c, rec := createTestContext(http.MethodPost, "/api/v1/import/csv/cards", "")
	user := createTestUser()
	c.Set("current_user", user)

	err := handler.ImportCardsCSV(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "bad_request", response.Error)
	assert.Contains(t, response.Message, "No file")
}
