package api

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
)

// ==================== Helper Functions ====================

func setupNotificationsTest() (*NotificationsHandler, *mocks.MockNotificationServiceInterface) {
	mockNotificationService := new(mocks.MockNotificationServiceInterface)
	handler := NewNotificationsHandler(mockNotificationService)
	return handler, mockNotificationService
}

// ==================== MarkAsRead Tests ====================

func TestNotificationsHandler_MarkAsRead_Success(t *testing.T) {
	handler, mockService := setupNotificationsTest()
	notificationID := uuid.New()
	c, rec := createTestContext(http.MethodPost, "/api/v1/notifications/:id/read", "")
	user := createTestUser()
	c.Set("current_user", user)
	c.SetParamNames("id")
	c.SetParamValues(notificationID.String())

	mockService.On("MarkAsRead", mock.Anything, user.ID, notificationID).Return(nil)

	err := handler.MarkAsRead(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "Notification marked as read", response["message"])
	mockService.AssertExpectations(t)
}

func TestNotificationsHandler_MarkAsRead_NotFound_IDOR(t *testing.T) {
	handler, mockService := setupNotificationsTest()
	notificationID := uuid.New() // Notification belonging to another user
	c, _ := createTestContext(http.MethodPost, "/api/v1/notifications/:id/read", "")
	user := createTestUser()
	c.Set("current_user", user)
	c.SetParamNames("id")
	c.SetParamValues(notificationID.String())

	// Service returns error because notification doesn't belong to this user
	mockService.On("MarkAsRead", mock.Anything, user.ID, notificationID).Return(errors.New("record not found"))

	err := handler.MarkAsRead(c)

	// Should return 404, not 500 — prevents information leakage
	httpErr, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusNotFound, httpErr.Code)
	mockService.AssertExpectations(t)
}

func TestNotificationsHandler_MarkAsRead_InvalidID(t *testing.T) {
	handler, _ := setupNotificationsTest()
	c, _ := createTestContext(http.MethodPost, "/api/v1/notifications/:id/read", "")
	user := createTestUser()
	c.Set("current_user", user)
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")

	err := handler.MarkAsRead(c)

	assert.Error(t, err)
}

func TestNotificationsHandler_MarkAsRead_PassesAuthenticatedUserID(t *testing.T) {
	handler, mockService := setupNotificationsTest()
	notificationID := uuid.New()
	c, _ := createTestContext(http.MethodPost, "/api/v1/notifications/:id/read", "")

	// Specific user - the mock will verify exactly this user ID is passed
	user := &models.User{
		ID:    uuid.New(),
		Email: "attacker@example.com",
	}
	c.Set("current_user", user)
	c.SetParamNames("id")
	c.SetParamValues(notificationID.String())

	// Only expect call with THIS user's ID (not any other)
	mockService.On("MarkAsRead", mock.Anything, user.ID, notificationID).Return(nil)

	err := handler.MarkAsRead(c)
	assert.NoError(t, err)

	// Verify the mock was called with the correct user ID
	mockService.AssertCalled(t, "MarkAsRead", mock.Anything, user.ID, notificationID)
	mockService.AssertExpectations(t)
}

// ==================== Delete Tests ====================

func TestNotificationsHandler_Delete_Success(t *testing.T) {
	handler, mockService := setupNotificationsTest()
	notificationID := uuid.New()
	c, rec := createTestContext(http.MethodDelete, "/api/v1/notifications/:id", "")
	user := createTestUser()
	c.Set("current_user", user)
	c.SetParamNames("id")
	c.SetParamValues(notificationID.String())

	mockService.On("DeleteNotification", mock.Anything, user.ID, notificationID).Return(nil)

	err := handler.Delete(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "Notification deleted", response["message"])
	mockService.AssertExpectations(t)
}

func TestNotificationsHandler_Delete_NotFound_IDOR(t *testing.T) {
	handler, mockService := setupNotificationsTest()
	notificationID := uuid.New() // Notification belonging to another user
	c, _ := createTestContext(http.MethodDelete, "/api/v1/notifications/:id", "")
	user := createTestUser()
	c.Set("current_user", user)
	c.SetParamNames("id")
	c.SetParamValues(notificationID.String())

	// Service returns error because notification doesn't belong to this user
	mockService.On("DeleteNotification", mock.Anything, user.ID, notificationID).Return(errors.New("record not found"))

	err := handler.Delete(c)

	// Should return 404, not 500 — prevents information leakage
	httpErr, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusNotFound, httpErr.Code)
	mockService.AssertExpectations(t)
}

func TestNotificationsHandler_Delete_InvalidID(t *testing.T) {
	handler, _ := setupNotificationsTest()
	c, _ := createTestContext(http.MethodDelete, "/api/v1/notifications/:id", "")
	user := createTestUser()
	c.Set("current_user", user)
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")

	err := handler.Delete(c)

	assert.Error(t, err)
}

func TestNotificationsHandler_Delete_PassesAuthenticatedUserID(t *testing.T) {
	handler, mockService := setupNotificationsTest()
	notificationID := uuid.New()
	c, _ := createTestContext(http.MethodDelete, "/api/v1/notifications/:id", "")

	// Specific user - the mock will verify exactly this user ID is passed
	user := &models.User{
		ID:    uuid.New(),
		Email: "attacker@example.com",
	}
	c.Set("current_user", user)
	c.SetParamNames("id")
	c.SetParamValues(notificationID.String())

	mockService.On("DeleteNotification", mock.Anything, user.ID, notificationID).Return(nil)

	err := handler.Delete(c)
	assert.NoError(t, err)

	// Verify the mock was called with the correct user ID
	mockService.AssertCalled(t, "DeleteNotification", mock.Anything, user.ID, notificationID)
	mockService.AssertExpectations(t)
}

// ==================== List Tests ====================

func TestNotificationsHandler_List_Success(t *testing.T) {
	handler, mockService := setupNotificationsTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/notifications", "")
	user := createTestUser()
	c.Set("current_user", user)

	notifications := []models.Notification{
		{
			ID:           uuid.New(),
			UserID:       user.ID,
			Type:         models.NotificationTypeShareReceived,
			ResourceType: "card",
			ResourceID:   uuid.New(),
			Metadata:     models.NotificationMetadata{},
			IsRead:       false,
		},
	}
	mockService.On("GetUserNotifications", mock.Anything, user.ID, 50, 0).Return(notifications, nil)

	err := handler.List(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockService.AssertExpectations(t)
}

func TestNotificationsHandler_List_Error(t *testing.T) {
	handler, mockService := setupNotificationsTest()
	c, _ := createTestContext(http.MethodGet, "/api/v1/notifications", "")
	user := createTestUser()
	c.Set("current_user", user)

	mockService.On("GetUserNotifications", mock.Anything, user.ID, 50, 0).Return([]models.Notification(nil), errors.New("db error"))

	err := handler.List(c)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusInternalServerError, httpErr.Code)
	mockService.AssertExpectations(t)
}

// ==================== GetUnreadCount Tests ====================

func TestNotificationsHandler_GetUnreadCount_Success(t *testing.T) {
	handler, mockService := setupNotificationsTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/notifications/unread-count", "")
	user := createTestUser()
	c.Set("current_user", user)

	mockService.On("GetUnreadCount", mock.Anything, user.ID).Return(int64(5), nil)

	err := handler.GetUnreadCount(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]int64
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, int64(5), response["count"])
	mockService.AssertExpectations(t)
}

// ==================== MarkAllAsRead Tests ====================

func TestNotificationsHandler_MarkAllAsRead_Success(t *testing.T) {
	handler, mockService := setupNotificationsTest()
	c, rec := createTestContext(http.MethodPost, "/api/v1/notifications/read-all", "")
	user := createTestUser()
	c.Set("current_user", user)

	mockService.On("MarkAllAsRead", mock.Anything, user.ID).Return(nil)

	err := handler.MarkAllAsRead(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockService.AssertExpectations(t)
}

func TestNotificationsHandler_MarkAllAsRead_Error(t *testing.T) {
	handler, mockService := setupNotificationsTest()
	c, _ := createTestContext(http.MethodPost, "/api/v1/notifications/read-all", "")
	user := createTestUser()
	c.Set("current_user", user)

	mockService.On("MarkAllAsRead", mock.Anything, user.ID).Return(errors.New("db error"))

	err := handler.MarkAllAsRead(c)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusInternalServerError, httpErr.Code)
	mockService.AssertExpectations(t)
}

// ==================== toNotificationDTO Tests ====================

func TestToNotificationDTO_WithReadAt(t *testing.T) {
	readAt := time.Now()
	n := &models.Notification{
		Type:         "share",
		ResourceType: "card",
		ResourceID:   uuid.New(),
		Metadata:     map[string]interface{}{"key": "value"},
		IsRead:       true,
		ReadAt:       &readAt,
	}
	n.ID = uuid.New()
	n.CreatedAt = time.Now()

	dto := toNotificationDTO(n)

	assert.True(t, dto.IsRead)
	assert.NotNil(t, dto.ReadAt)
	assert.Equal(t, "share", dto.Type)
	assert.Equal(t, "card", dto.ResourceType)
}

func TestToNotificationDTO_WithoutReadAt(t *testing.T) {
	n := &models.Notification{
		Type:         "transfer",
		ResourceType: "voucher",
		ResourceID:   uuid.New(),
		Metadata:     map[string]interface{}{},
		IsRead:       false,
		ReadAt:       nil,
	}
	n.ID = uuid.New()
	n.CreatedAt = time.Now()

	dto := toNotificationDTO(n)

	assert.False(t, dto.IsRead)
	assert.Nil(t, dto.ReadAt)
	assert.Equal(t, "transfer", dto.Type)
}

// ==================== List Tests (additional branches) ====================

func TestNotificationsHandler_List_WithLimitParam(t *testing.T) {
	handler, mockService := setupNotificationsTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/notifications?limit=10", "")
	user := createTestUser()
	c.Set("current_user", user)

	mockService.On("GetUserNotifications", mock.Anything, user.ID, 10, 0).Return([]models.Notification{}, nil)

	err := handler.List(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockService.AssertExpectations(t)
}

func TestNotificationsHandler_List_InvalidLimitParam(t *testing.T) {
	handler, mockService := setupNotificationsTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/notifications?limit=abc", "")
	user := createTestUser()
	c.Set("current_user", user)

	// Invalid limit falls back to default (50)
	mockService.On("GetUserNotifications", mock.Anything, user.ID, 50, 0).Return([]models.Notification{}, nil)

	err := handler.List(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockService.AssertExpectations(t)
}
