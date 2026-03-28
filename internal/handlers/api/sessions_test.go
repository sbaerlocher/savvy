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
	"savvy/internal/services"
)

// ==================== Helper Functions ====================

func setupSessionsTest() (*SessionsHandler, *mocks.MockSessionServiceInterface) {
	mockSessionService := new(mocks.MockSessionServiceInterface)
	handler := NewSessionsHandler(mockSessionService)
	return handler, mockSessionService
}

// ==================== List Tests ====================

func TestSessionsHandler_List_Success(t *testing.T) {
	handler, mockSessionService := setupSessionsTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/profile/sessions", "")
	user := createTestUser()
	c.Set("current_user", user)

	// GetCurrentSessionTokenHash will return "" since no session store is configured in tests
	sessions := []services.SessionDTO{
		{
			ID:          uuid.New().String(),
			IPAddress:   "192.168.1.1",
			DeviceInfo:  "Desktop",
			BrowserInfo: "Chrome",
			IsCurrent:   true,
		},
	}
	mockSessionService.On("ListUserSessions", mock.Anything, user.ID, "").Return(sessions, nil)

	err := handler.List(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string][]services.SessionDTO
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Len(t, response["sessions"], 1)
	assert.True(t, response["sessions"][0].IsCurrent)
	mockSessionService.AssertExpectations(t)
}

func TestSessionsHandler_List_Unauthorized(t *testing.T) {
	handler, _ := setupSessionsTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/profile/sessions", "")
	// No user set

	err := handler.List(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestSessionsHandler_List_ServiceError(t *testing.T) {
	handler, mockSessionService := setupSessionsTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/profile/sessions", "")
	user := createTestUser()
	c.Set("current_user", user)

	mockSessionService.On("ListUserSessions", mock.Anything, user.ID, "").
		Return([]services.SessionDTO(nil), errors.New("database error"))

	err := handler.List(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSessionService.AssertExpectations(t)
}

// ==================== Revoke Tests ====================

func TestSessionsHandler_Revoke_Success(t *testing.T) {
	handler, mockSessionService := setupSessionsTest()
	sessionID := uuid.New()
	c, rec := createTestContext(http.MethodDelete, "/api/v1/profile/sessions/:id", "")
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: sessionID.String()}})

	mockSessionService.On("RevokeSession", mock.Anything, user.ID, sessionID).Return(nil)

	err := handler.Revoke(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "Session revoked", response["message"])
	mockSessionService.AssertExpectations(t)
}

func TestSessionsHandler_Revoke_Unauthorized(t *testing.T) {
	handler, _ := setupSessionsTest()
	c, rec := createTestContext(http.MethodDelete, "/api/v1/profile/sessions/:id", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: uuid.New().String()}})

	err := handler.Revoke(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestSessionsHandler_Revoke_InvalidID(t *testing.T) {
	handler, _ := setupSessionsTest()
	c, rec := createTestContext(http.MethodDelete, "/api/v1/profile/sessions/:id", "")
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "invalid-uuid"}})

	err := handler.Revoke(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestSessionsHandler_Revoke_ServiceError(t *testing.T) {
	handler, mockSessionService := setupSessionsTest()
	sessionID := uuid.New()
	c, rec := createTestContext(http.MethodDelete, "/api/v1/profile/sessions/:id", "")
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: sessionID.String()}})

	mockSessionService.On("RevokeSession", mock.Anything, user.ID, sessionID).
		Return(errors.New("not found"))

	err := handler.Revoke(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSessionService.AssertExpectations(t)
}

// ==================== RevokeOthers Tests ====================

func TestSessionsHandler_RevokeOthers_Unauthorized(t *testing.T) {
	handler, _ := setupSessionsTest()
	c, rec := createTestContext(http.MethodPost, "/api/v1/profile/sessions/revoke-others", "")

	err := handler.RevokeOthers(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestSessionsHandler_RevokeOthers_NoSession(t *testing.T) {
	handler, _ := setupSessionsTest()
	c, rec := createTestContext(http.MethodPost, "/api/v1/profile/sessions/revoke-others", "")
	user := createTestUser()
	c.Set("current_user", user)
	// GetCurrentSessionTokenHash will return "" since no session store configured

	err := handler.RevokeOthers(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
