package api

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"savvy/internal/mocks"
	"savvy/internal/services"
)

// ==================== Helper Functions ====================

func setupExportTest() (*ExportHandler, *mocks.MockExportServiceInterface) {
	mockExportService := new(mocks.MockExportServiceInterface)
	handler := NewExportHandler(mockExportService)
	return handler, mockExportService
}

// ==================== ExportData Tests ====================

func TestExportHandler_ExportData_Success(t *testing.T) {
	handler, mockExportService := setupExportTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/export", "")
	user := createTestUser()
	c.Set("current_user", user)

	exportData := &services.ExportData{
		Cards:    []services.ExportCard{{MerchantName: "IKEA", CardNumber: "123"}},
		Vouchers: []services.ExportVoucher{},
	}
	mockExportService.On("ExportUserData", mock.Anything, user.ID).Return(exportData, nil)

	err := handler.ExportData(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Disposition"), "savvy-export-")
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	mockExportService.AssertExpectations(t)
}

func TestExportHandler_ExportData_Unauthorized(t *testing.T) {
	handler, _ := setupExportTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/export", "")
	// No user set

	err := handler.ExportData(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestExportHandler_ExportData_ServiceError(t *testing.T) {
	handler, mockExportService := setupExportTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/export", "")
	user := createTestUser()
	c.Set("current_user", user)

	mockExportService.On("ExportUserData", mock.Anything, user.ID).
		Return((*services.ExportData)(nil), errors.New("database error"))

	err := handler.ExportData(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockExportService.AssertExpectations(t)
}
