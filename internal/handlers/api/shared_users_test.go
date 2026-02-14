package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"savvy/internal/mocks"
	"savvy/internal/models"
)

// ==================== Helper Functions ====================

func setupSharedUsersTest() (*SharedUsersHandler, *mocks.MockShareServiceInterface) {
	mockShareService := new(mocks.MockShareServiceInterface)
	handler := NewSharedUsersHandler(mockShareService)
	return handler, mockShareService
}

// ==================== Search Tests ====================

func TestSharedUsersHandler_Search_WithQuery(t *testing.T) {
	handler, mockShareService := setupSharedUsersTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/shared-users?q=john", "")
	user := createTestUser()
	c.Set("current_user", user)
	c.QueryParams().Set("q", "john")

	sharedUsers := []models.User{
		{ID: uuid.New(), Email: "john@example.com", FirstName: "John", LastName: "Doe", Role: "user"},
	}
	mockShareService.On("GetSharedUsers", mock.Anything, user.ID, "john").Return(sharedUsers, nil)

	err := handler.Search(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string][]UserDTO
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Len(t, response["users"], 1)
	assert.Equal(t, "john@example.com", response["users"][0].Email)
	mockShareService.AssertExpectations(t)
}

func TestSharedUsersHandler_Search_EmptyQuery(t *testing.T) {
	handler, mockShareService := setupSharedUsersTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/shared-users", "")
	user := createTestUser()
	c.Set("current_user", user)

	sharedUsers := []models.User{
		{ID: uuid.New(), Email: "user1@example.com", FirstName: "User", LastName: "One", Role: "user"},
		{ID: uuid.New(), Email: "user2@example.com", FirstName: "User", LastName: "Two", Role: "user"},
	}
	mockShareService.On("GetSharedUsers", mock.Anything, user.ID, "").Return(sharedUsers, nil)

	err := handler.Search(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string][]UserDTO
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Len(t, response["users"], 2)
	mockShareService.AssertExpectations(t)
}

func TestSharedUsersHandler_Search_ServiceError(t *testing.T) {
	handler, mockShareService := setupSharedUsersTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/shared-users", "")
	user := createTestUser()
	c.Set("current_user", user)

	mockShareService.On("GetSharedUsers", mock.Anything, user.ID, "").
		Return([]models.User(nil), errors.New("database error"))

	err := handler.Search(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockShareService.AssertExpectations(t)
}

func TestSharedUsersHandler_Search_EmptyResult(t *testing.T) {
	handler, mockShareService := setupSharedUsersTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/shared-users?q=nonexistent", "")
	user := createTestUser()
	c.Set("current_user", user)
	c.QueryParams().Set("q", "nonexistent")

	mockShareService.On("GetSharedUsers", mock.Anything, user.ID, "nonexistent").Return([]models.User{}, nil)

	err := handler.Search(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string][]UserDTO
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Len(t, response["users"], 0)
	mockShareService.AssertExpectations(t)
}
