package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"savvy/internal/mocks"
	"savvy/internal/models"
)

// ==================== Helper Functions ====================

func setupPushTest() (*PushHandler, *mocks.MockPushServiceInterface) {
	mockPushService := new(mocks.MockPushServiceInterface)
	handler := NewPushHandler(mockPushService)
	return handler, mockPushService
}

// ==================== Subscribe Tests ====================

func TestPushHandler_Subscribe_Success(t *testing.T) {
	handler, mockPushService := setupPushTest()
	body := `{"endpoint":"https://push.example.com/sub/123","keys":{"p256dh":"key1","auth":"key2"}}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/push/subscribe", body)
	user := createTestUser()
	c.Set("current_user", user)

	mockPushService.On("Subscribe", mock.Anything, user.ID, "https://push.example.com/sub/123", "key1", "key2", "").Return(nil)

	err := handler.Subscribe(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "subscribed", response["status"])
	mockPushService.AssertExpectations(t)
}

func TestPushHandler_Subscribe_Unauthorized(t *testing.T) {
	handler, _ := setupPushTest()
	body := `{"endpoint":"https://push.example.com/sub/123","keys":{"p256dh":"key1","auth":"key2"}}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/push/subscribe", body)
	// No user set

	err := handler.Subscribe(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestPushHandler_Subscribe_MissingFields(t *testing.T) {
	handler, _ := setupPushTest()
	body := `{"endpoint":"","keys":{"p256dh":"","auth":""}}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/push/subscribe", body)
	user := createTestUser()
	c.Set("current_user", user)

	err := handler.Subscribe(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestPushHandler_Subscribe_InvalidJSON(t *testing.T) {
	handler, _ := setupPushTest()
	c, rec := createTestContext(http.MethodPost, "/api/v1/push/subscribe", "invalid-json")
	user := createTestUser()
	c.Set("current_user", user)

	err := handler.Subscribe(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestPushHandler_Subscribe_ServiceError(t *testing.T) {
	handler, mockPushService := setupPushTest()
	body := `{"endpoint":"https://push.example.com/sub/123","keys":{"p256dh":"key1","auth":"key2"}}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/push/subscribe", body)
	user := createTestUser()
	c.Set("current_user", user)

	mockPushService.On("Subscribe", mock.Anything, user.ID, "https://push.example.com/sub/123", "key1", "key2", "").
		Return(errors.New("database error"))

	err := handler.Subscribe(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockPushService.AssertExpectations(t)
}

// ==================== Unsubscribe Tests ====================

func TestPushHandler_Unsubscribe_Success(t *testing.T) {
	handler, mockPushService := setupPushTest()
	body := `{"endpoint":"https://push.example.com/sub/123"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/push/unsubscribe", body)
	user := createTestUser()
	c.Set("current_user", user)

	mockPushService.On("Unsubscribe", mock.Anything, "https://push.example.com/sub/123").Return(nil)

	err := handler.Unsubscribe(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "unsubscribed", response["status"])
	mockPushService.AssertExpectations(t)
}

func TestPushHandler_Unsubscribe_Unauthorized(t *testing.T) {
	handler, _ := setupPushTest()
	body := `{"endpoint":"https://push.example.com/sub/123"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/push/unsubscribe", body)

	err := handler.Unsubscribe(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestPushHandler_Unsubscribe_MissingEndpoint(t *testing.T) {
	handler, _ := setupPushTest()
	body := `{"endpoint":""}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/push/unsubscribe", body)
	user := createTestUser()
	c.Set("current_user", user)

	err := handler.Unsubscribe(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestPushHandler_Unsubscribe_ServiceError(t *testing.T) {
	handler, mockPushService := setupPushTest()
	body := `{"endpoint":"https://push.example.com/sub/123"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/push/unsubscribe", body)
	user := createTestUser()
	c.Set("current_user", user)

	mockPushService.On("Unsubscribe", mock.Anything, "https://push.example.com/sub/123").
		Return(errors.New("database error"))

	err := handler.Unsubscribe(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockPushService.AssertExpectations(t)
}

// ==================== GetVAPIDKey Tests ====================

func TestPushHandler_GetVAPIDKey_Success(t *testing.T) {
	handler, mockPushService := setupPushTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/push/vapid-key", "")

	mockPushService.On("GetVAPIDPublicKey").Return("test-vapid-public-key")

	err := handler.GetVAPIDKey(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "test-vapid-public-key", response["public_key"])
	mockPushService.AssertExpectations(t)
}

// Ensure the _ import is used (models.User is referenced via createTestUser)
var _ = (*models.User)(nil)
