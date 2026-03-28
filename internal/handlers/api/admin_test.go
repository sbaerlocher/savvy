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
	"savvy/internal/repository"
	"savvy/internal/services"
)

// ==================== Helper Functions ====================

func setupAdminTest() (*AdminHandler, *mocks.MockAdminServiceInterface, *mocks.MockUserServiceInterface, *mocks.MockHealthCheckServiceInterface, *mocks.MockEmailServiceInterface, *mocks.MockPushServiceInterface) {
	mockAdminService := new(mocks.MockAdminServiceInterface)
	mockUserService := new(mocks.MockUserServiceInterface)
	mockHealthService := new(mocks.MockHealthCheckServiceInterface)
	mockEmailSvc := new(mocks.MockEmailServiceInterface)
	mockPushSvc := new(mocks.MockPushServiceInterface)
	handler := NewAdminHandler(mockAdminService, mockUserService, mockHealthService, mockEmailSvc, mockPushSvc)
	return handler, mockAdminService, mockUserService, mockHealthService, mockEmailSvc, mockPushSvc
}

// ==================== ListUsers Tests ====================

func TestAdminHandler_ListUsers_Success(t *testing.T) {
	handler, mockAdminService, _, _, _, _ := setupAdminTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/admin/users", "")

	users := []models.User{
		{
			ID:           uuid.New(),
			Email:        "user1@example.com",
			FirstName:    "User",
			LastName:     "One",
			Role:         "user",
			AuthProvider: "local",
		},
		{
			ID:           uuid.New(),
			Email:        "admin@example.com",
			FirstName:    "Admin",
			LastName:     "User",
			Role:         "admin",
			AuthProvider: "local",
		},
	}

	mockAdminService.On("GetAllUsers", mock.Anything).Return(users, nil)

	err := handler.ListUsers(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response AdminUserListResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Len(t, response.Users, 2)
	assert.Equal(t, "user1@example.com", response.Users[0].Email)
	mockAdminService.AssertExpectations(t)
}

func TestAdminHandler_ListUsers_ServiceError(t *testing.T) {
	handler, mockAdminService, _, _, _, _ := setupAdminTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/admin/users", "")

	mockAdminService.On("GetAllUsers", mock.Anything).Return([]models.User(nil), errors.New("database error"))

	err := handler.ListUsers(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "server_error", response.Error)
	mockAdminService.AssertExpectations(t)
}

// ==================== GetUser Tests ====================

func TestAdminHandler_GetUser_Success(t *testing.T) {
	handler, mockAdminService, _, _, _, _ := setupAdminTest()
	userID := uuid.New()
	c, rec := createTestContext(http.MethodGet, "/api/v1/admin/users/"+userID.String(), "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: userID.String()}})

	user := &models.User{
		ID:           userID,
		Email:        "test@example.com",
		FirstName:    "Test",
		LastName:     "User",
		Role:         "user",
		AuthProvider: "local",
	}

	mockAdminService.On("GetUserByID", mock.Anything, userID).Return(user, nil)

	err := handler.GetUser(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NotNil(t, response["user"])
	mockAdminService.AssertExpectations(t)
}

func TestAdminHandler_GetUser_InvalidID(t *testing.T) {
	handler, _, _, _, _, _ := setupAdminTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/admin/users/invalid", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "invalid"}})

	err := handler.GetUser(c)

	// parseResourceID writes JSON response AND returns error
	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "invalid_id", response.Error)
}

func TestAdminHandler_GetUser_NotFound(t *testing.T) {
	handler, mockAdminService, _, _, _, _ := setupAdminTest()
	userID := uuid.New()
	c, rec := createTestContext(http.MethodGet, "/api/v1/admin/users/"+userID.String(), "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: userID.String()}})

	mockAdminService.On("GetUserByID", mock.Anything, userID).Return((*models.User)(nil), errors.New("not found"))

	err := handler.GetUser(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "user_not_found", response.Error)
	mockAdminService.AssertExpectations(t)
}

// ==================== CreateUser Tests ====================

func TestAdminHandler_CreateUser_Success(t *testing.T) {
	handler, mockAdminService, mockUserService, _, _, _ := setupAdminTest()

	requestBody := `{
		"email": "newuser@example.com",
		"password": "SecurePass123!",
		"first_name": "New",
		"last_name": "User",
		"role": "user"
	}`

	c, rec := createTestContext(http.MethodPost, "/api/v1/admin/users", requestBody)

	newUser := &models.User{
		ID:           uuid.New(),
		Email:        "newuser@example.com",
		FirstName:    "New",
		LastName:     "User",
		Role:         "user",
		AuthProvider: "local",
	}

	mockUserService.On("GetUserByEmail", mock.Anything, "newuser@example.com").Return((*models.User)(nil), errors.New("not found"))
	mockAdminService.On("CreateLocalUser", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil).Run(func(args mock.Arguments) {
		user := args.Get(1).(*models.User)
		user.ID = newUser.ID
	})

	err := handler.CreateUser(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var response map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NotNil(t, response["user"])
	mockAdminService.AssertExpectations(t)
	mockUserService.AssertExpectations(t)
}

func TestAdminHandler_CreateUser_MissingFields(t *testing.T) {
	handler, _, _, _, _, _ := setupAdminTest()

	requestBody := `{
		"email": "newuser@example.com"
	}`

	c, rec := createTestContext(http.MethodPost, "/api/v1/admin/users", requestBody)

	err := handler.CreateUser(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "missing_fields", response.Error)
}

func TestAdminHandler_CreateUser_EmailAlreadyExists(t *testing.T) {
	handler, _, mockUserService, _, _, _ := setupAdminTest()

	requestBody := `{
		"email": "existing@example.com",
		"password": "SecurePass123!",
		"first_name": "Test",
		"last_name": "User",
		"role": "user"
	}`

	c, rec := createTestContext(http.MethodPost, "/api/v1/admin/users", requestBody)

	existingUser := &models.User{
		ID:    uuid.New(),
		Email: "existing@example.com",
	}

	mockUserService.On("GetUserByEmail", mock.Anything, "existing@example.com").Return(existingUser, nil)

	err := handler.CreateUser(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusConflict, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "email_exists", response.Error)
	mockUserService.AssertExpectations(t)
}

// ==================== Error Message Sanitization Tests ====================

func TestAdminHandler_StartImpersonation_UnknownError_NoLeakedDetails(t *testing.T) {
	handler, mockAdminService, _, _, _, _ := setupAdminTest()
	targetUserID := uuid.New()
	c, rec := createTestContext(http.MethodPost, "/api/v1/admin/users/:id/impersonate", "")
	admin := &models.User{
		ID:    uuid.New(),
		Email: "admin@example.com",
		Role:  "admin",
	}
	c.Set("current_user", admin)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: targetUserID.String()}})

	// Simulate a GORM/DB error leaking through the service layer
	mockAdminService.On("ValidateImpersonation", mock.Anything, admin.ID, targetUserID).
		Return(errors.New("ERROR: relation \"users\" does not exist (SQLSTATE 42P01)"))

	err := handler.StartImpersonation(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "validation_failed", response.Error)
	// Must return generic message, not the DB error
	assert.Equal(t, "Impersonation validation failed", response.Message)
	assert.NotContains(t, response.Message, "relation")
	assert.NotContains(t, response.Message, "users")
	assert.NotContains(t, response.Message, "SQLSTATE")
	mockAdminService.AssertExpectations(t)
}

func TestAdminHandler_StartImpersonation_KnownErrors_SafeMessages(t *testing.T) {
	tests := []struct {
		name           string
		serviceError   string
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:           "Only admins can impersonate",
			serviceError:   "only admins can impersonate",
			expectedStatus: http.StatusForbidden,
			expectedMsg:    "Only admins can impersonate users",
		},
		{
			name:           "Cannot impersonate yourself",
			serviceError:   "cannot impersonate yourself",
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "Cannot impersonate yourself",
		},
		{
			name:           "Target user not found",
			serviceError:   "target user not found",
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "Target user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockAdminService, _, _, _, _ := setupAdminTest()
			targetUserID := uuid.New()
			c, rec := createTestContext(http.MethodPost, "/api/v1/admin/users/:id/impersonate", "")
			admin := &models.User{
				ID:    uuid.New(),
				Email: "admin@example.com",
				Role:  "admin",
			}
			c.Set("current_user", admin)
			c.SetPathValues(echo.PathValues{{Name: "id", Value: targetUserID.String()}})

			mockAdminService.On("ValidateImpersonation", mock.Anything, admin.ID, targetUserID).
				Return(errors.New(tt.serviceError))

			err := handler.StartImpersonation(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			var response ErrorResponse
			_ = json.Unmarshal(rec.Body.Bytes(), &response)
			assert.Equal(t, "validation_failed", response.Error)
			assert.Equal(t, tt.expectedMsg, response.Message)
			mockAdminService.AssertExpectations(t)
		})
	}
}

// ==================== UpdateUser Tests ====================

func TestAdminHandler_UpdateUser_Success(t *testing.T) {
	handler, mockAdminService, mockUserService, _, _, _ := setupAdminTest()
	userID := uuid.New()

	requestBody := `{
		"email": "updated@example.com",
		"first_name": "Updated",
		"last_name": "Name",
		"role": "admin"
	}`

	c, rec := createTestContext(http.MethodPatch, "/api/v1/admin/users/"+userID.String(), requestBody)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: userID.String()}})

	existingUser := &models.User{
		ID:           userID,
		Email:        "original@example.com",
		FirstName:    "Original",
		LastName:     "User",
		Role:         "user",
		AuthProvider: "local",
	}

	mockAdminService.On("GetUserByID", mock.Anything, userID).Return(existingUser, nil)
	// Email changed, so email conflict check is performed - no conflict found
	mockUserService.On("GetUserByEmail", mock.Anything, "updated@example.com").Return((*models.User)(nil), errors.New("not found"))
	mockAdminService.On("UpdateUser", mock.Anything, userID, "updated@example.com", "Updated", "Name", "admin").Return(nil)

	err := handler.UpdateUser(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "User updated successfully", response["message"])
	mockAdminService.AssertExpectations(t)
	mockUserService.AssertExpectations(t)
}

func TestAdminHandler_UpdateUser_PartialUpdate_OnlyEmail(t *testing.T) {
	handler, mockAdminService, mockUserService, _, _, _ := setupAdminTest()
	userID := uuid.New()

	requestBody := `{
		"email": "newemail@example.com"
	}`

	c, rec := createTestContext(http.MethodPatch, "/api/v1/admin/users/"+userID.String(), requestBody)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: userID.String()}})

	existingUser := &models.User{
		ID:           userID,
		Email:        "old@example.com",
		FirstName:    "Keep",
		LastName:     "Same",
		Role:         "user",
		AuthProvider: "local",
	}

	mockAdminService.On("GetUserByID", mock.Anything, userID).Return(existingUser, nil)
	// Email changed, so conflict check happens - no conflict found
	mockUserService.On("GetUserByEmail", mock.Anything, "newemail@example.com").Return((*models.User)(nil), errors.New("not found"))
	// Existing first_name, last_name and role should be preserved
	mockAdminService.On("UpdateUser", mock.Anything, userID, "newemail@example.com", "Keep", "Same", "user").Return(nil)

	err := handler.UpdateUser(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockAdminService.AssertExpectations(t)
	mockUserService.AssertExpectations(t)
}

func TestAdminHandler_UpdateUser_PartialUpdate_OnlyName(t *testing.T) {
	handler, mockAdminService, _, _, _, _ := setupAdminTest()
	userID := uuid.New()

	requestBody := `{
		"first_name": "NewFirst",
		"last_name": "NewLast"
	}`

	c, rec := createTestContext(http.MethodPatch, "/api/v1/admin/users/"+userID.String(), requestBody)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: userID.String()}})

	existingUser := &models.User{
		ID:           userID,
		Email:        "keep@example.com",
		FirstName:    "OldFirst",
		LastName:     "OldLast",
		Role:         "admin",
		AuthProvider: "local",
	}

	mockAdminService.On("GetUserByID", mock.Anything, userID).Return(existingUser, nil)
	// Email stays the same, so no email conflict check. Role also stays.
	mockAdminService.On("UpdateUser", mock.Anything, userID, "keep@example.com", "NewFirst", "NewLast", "admin").Return(nil)

	err := handler.UpdateUser(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockAdminService.AssertExpectations(t)
}

func TestAdminHandler_UpdateUser_InvalidID(t *testing.T) {
	handler, _, _, _, _, _ := setupAdminTest()
	c, rec := createTestContext(http.MethodPatch, "/api/v1/admin/users/invalid", `{"first_name":"Test"}`)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "invalid"}})

	err := handler.UpdateUser(c)

	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "invalid_id", response.Error)
}

func TestAdminHandler_UpdateUser_UserNotFound(t *testing.T) {
	handler, mockAdminService, _, _, _, _ := setupAdminTest()
	userID := uuid.New()

	c, rec := createTestContext(http.MethodPatch, "/api/v1/admin/users/"+userID.String(), `{"first_name":"Test"}`)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: userID.String()}})

	mockAdminService.On("GetUserByID", mock.Anything, userID).Return((*models.User)(nil), errors.New("not found"))

	err := handler.UpdateUser(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "user_not_found", response.Error)
	mockAdminService.AssertExpectations(t)
}

func TestAdminHandler_UpdateUser_InvalidRole(t *testing.T) {
	handler, mockAdminService, _, _, _, _ := setupAdminTest()
	userID := uuid.New()

	requestBody := `{
		"role": "superadmin"
	}`

	c, rec := createTestContext(http.MethodPatch, "/api/v1/admin/users/"+userID.String(), requestBody)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: userID.String()}})

	existingUser := &models.User{
		ID:           userID,
		Email:        "test@example.com",
		FirstName:    "Test",
		LastName:     "User",
		Role:         "user",
		AuthProvider: "local",
	}

	mockAdminService.On("GetUserByID", mock.Anything, userID).Return(existingUser, nil)

	err := handler.UpdateUser(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "invalid_role", response.Error)
	assert.Equal(t, "Role must be 'user' or 'admin'", response.Message)
	mockAdminService.AssertExpectations(t)
}

func TestAdminHandler_UpdateUser_EmailConflict(t *testing.T) {
	handler, mockAdminService, mockUserService, _, _, _ := setupAdminTest()
	userID := uuid.New()
	otherUserID := uuid.New()

	requestBody := `{
		"email": "taken@example.com"
	}`

	c, rec := createTestContext(http.MethodPatch, "/api/v1/admin/users/"+userID.String(), requestBody)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: userID.String()}})

	existingUser := &models.User{
		ID:           userID,
		Email:        "original@example.com",
		FirstName:    "Test",
		LastName:     "User",
		Role:         "user",
		AuthProvider: "local",
	}

	conflictingUser := &models.User{
		ID:    otherUserID,
		Email: "taken@example.com",
	}

	mockAdminService.On("GetUserByID", mock.Anything, userID).Return(existingUser, nil)
	mockUserService.On("GetUserByEmail", mock.Anything, "taken@example.com").Return(conflictingUser, nil)

	err := handler.UpdateUser(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusConflict, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "email_exists", response.Error)
	assert.Equal(t, "Another user with this email already exists", response.Message)
	mockAdminService.AssertExpectations(t)
	mockUserService.AssertExpectations(t)
}

func TestAdminHandler_UpdateUser_EmailChangedToSameUser(t *testing.T) {
	// When email lookup returns the same user (same ID), it should not be treated as a conflict
	handler, mockAdminService, mockUserService, _, _, _ := setupAdminTest()
	userID := uuid.New()

	requestBody := `{
		"email": "changed@example.com"
	}`

	c, rec := createTestContext(http.MethodPatch, "/api/v1/admin/users/"+userID.String(), requestBody)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: userID.String()}})

	existingUser := &models.User{
		ID:           userID,
		Email:        "original@example.com",
		FirstName:    "Test",
		LastName:     "User",
		Role:         "user",
		AuthProvider: "local",
	}

	// GetUserByEmail returns the same user (same ID) - not a conflict
	sameUser := &models.User{
		ID:    userID,
		Email: "changed@example.com",
	}

	mockAdminService.On("GetUserByID", mock.Anything, userID).Return(existingUser, nil)
	mockUserService.On("GetUserByEmail", mock.Anything, "changed@example.com").Return(sameUser, nil)
	mockAdminService.On("UpdateUser", mock.Anything, userID, "changed@example.com", "Test", "User", "user").Return(nil)

	err := handler.UpdateUser(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockAdminService.AssertExpectations(t)
	mockUserService.AssertExpectations(t)
}

func TestAdminHandler_UpdateUser_ServiceError(t *testing.T) {
	handler, mockAdminService, _, _, _, _ := setupAdminTest()
	userID := uuid.New()

	requestBody := `{
		"first_name": "Updated"
	}`

	c, rec := createTestContext(http.MethodPatch, "/api/v1/admin/users/"+userID.String(), requestBody)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: userID.String()}})

	existingUser := &models.User{
		ID:           userID,
		Email:        "test@example.com",
		FirstName:    "Test",
		LastName:     "User",
		Role:         "user",
		AuthProvider: "local",
	}

	mockAdminService.On("GetUserByID", mock.Anything, userID).Return(existingUser, nil)
	mockAdminService.On("UpdateUser", mock.Anything, userID, "test@example.com", "Updated", "User", "user").Return(errors.New("database error"))

	err := handler.UpdateUser(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "server_error", response.Error)
	assert.Equal(t, "Failed to update user", response.Message)
	mockAdminService.AssertExpectations(t)
}

func TestAdminHandler_UpdateUser_EmptyFieldsAfterUpdate(t *testing.T) {
	// Setting first_name to "" results in missing_fields error
	handler, mockAdminService, _, _, _, _ := setupAdminTest()
	userID := uuid.New()

	requestBody := `{
		"first_name": ""
	}`

	c, rec := createTestContext(http.MethodPatch, "/api/v1/admin/users/"+userID.String(), requestBody)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: userID.String()}})

	existingUser := &models.User{
		ID:           userID,
		Email:        "test@example.com",
		FirstName:    "Test",
		LastName:     "User",
		Role:         "user",
		AuthProvider: "local",
	}

	mockAdminService.On("GetUserByID", mock.Anything, userID).Return(existingUser, nil)

	err := handler.UpdateUser(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "missing_fields", response.Error)
	mockAdminService.AssertExpectations(t)
}

func TestAdminHandler_UpdateUser_EmailNormalization(t *testing.T) {
	handler, mockAdminService, mockUserService, _, _, _ := setupAdminTest()
	userID := uuid.New()

	requestBody := `{
		"email": "  UPPER@Example.COM  "
	}`

	c, rec := createTestContext(http.MethodPatch, "/api/v1/admin/users/"+userID.String(), requestBody)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: userID.String()}})

	existingUser := &models.User{
		ID:           userID,
		Email:        "old@example.com",
		FirstName:    "Test",
		LastName:     "User",
		Role:         "user",
		AuthProvider: "local",
	}

	mockAdminService.On("GetUserByID", mock.Anything, userID).Return(existingUser, nil)
	// Email changed, so conflict check happens with normalized email
	mockUserService.On("GetUserByEmail", mock.Anything, "upper@example.com").Return((*models.User)(nil), errors.New("not found"))
	// Email should be normalized to lowercase and trimmed
	mockAdminService.On("UpdateUser", mock.Anything, userID, "upper@example.com", "Test", "User", "user").Return(nil)

	err := handler.UpdateUser(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockAdminService.AssertExpectations(t)
	mockUserService.AssertExpectations(t)
}

// ==================== GetAuditLogs Tests ====================

func TestAdminHandler_GetAuditLogs_Success(t *testing.T) {
	handler, mockAdminService, _, _, _, _ := setupAdminTest()

	c, rec := createTestContext(http.MethodGet, "/api/v1/admin/audit-log?page=1&per_page=10", "")

	logID := uuid.New()
	resourceID := uuid.New()
	result := &services.AuditLogResult{
		Logs: []models.AuditLog{
			{
				ID:           logID,
				Action:       "delete",
				ResourceType: "cards",
				ResourceID:   resourceID,
				ResourceData: `{"name":"test"}`,
				IPAddress:    "127.0.0.1",
				UserAgent:    "TestAgent",
				CreatedAt:    time.Now(),
			},
		},
		Total: 1,
	}

	expectedFilters := repository.AuditLogFilters{
		Page:    1,
		PerPage: 10,
	}

	mockAdminService.On("GetAuditLogs", mock.Anything, expectedFilters).Return(result, nil)

	err := handler.GetAuditLogs(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response AuditLogListResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Len(t, response.Logs, 1)
	assert.Equal(t, int64(1), response.Total)
	assert.Equal(t, 1, response.Page)
	assert.Equal(t, 10, response.PerPage)
	assert.Equal(t, 1, response.TotalPages)
	assert.Equal(t, "delete", response.Logs[0].Action)
	assert.Equal(t, "cards", response.Logs[0].ResourceType)
	mockAdminService.AssertExpectations(t)
}

func TestAdminHandler_GetAuditLogs_DefaultPagination(t *testing.T) {
	handler, mockAdminService, _, _, _, _ := setupAdminTest()

	// No page or per_page params - defaults should apply (page=1, per_page=20)
	c, rec := createTestContext(http.MethodGet, "/api/v1/admin/audit-log", "")

	result := &services.AuditLogResult{
		Logs:  []models.AuditLog{},
		Total: 0,
	}

	expectedFilters := repository.AuditLogFilters{
		Page:    1,
		PerPage: 20,
	}

	mockAdminService.On("GetAuditLogs", mock.Anything, expectedFilters).Return(result, nil)

	err := handler.GetAuditLogs(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response AuditLogListResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, 1, response.Page)
	assert.Equal(t, 20, response.PerPage)
	assert.Equal(t, 0, response.TotalPages)
	mockAdminService.AssertExpectations(t)
}

func TestAdminHandler_GetAuditLogs_WithUserIDFilter(t *testing.T) {
	handler, mockAdminService, _, _, _, _ := setupAdminTest()

	filterUserID := uuid.New()
	c, rec := createTestContext(http.MethodGet, "/api/v1/admin/audit-log?page=1&per_page=10&user_id="+filterUserID.String(), "")

	result := &services.AuditLogResult{
		Logs:  []models.AuditLog{},
		Total: 0,
	}

	expectedFilters := repository.AuditLogFilters{
		UserID:  &filterUserID,
		Page:    1,
		PerPage: 10,
	}

	mockAdminService.On("GetAuditLogs", mock.Anything, expectedFilters).Return(result, nil)

	err := handler.GetAuditLogs(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockAdminService.AssertExpectations(t)
}

func TestAdminHandler_GetAuditLogs_InvalidUserIDFilter(t *testing.T) {
	handler, _, _, _, _, _ := setupAdminTest()

	c, rec := createTestContext(http.MethodGet, "/api/v1/admin/audit-log?user_id=not-a-uuid", "")

	err := handler.GetAuditLogs(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "invalid_user_id", response.Error)
	assert.Equal(t, "Invalid user ID format", response.Message)
}

func TestAdminHandler_GetAuditLogs_WithAllFilters(t *testing.T) {
	handler, mockAdminService, _, _, _, _ := setupAdminTest()

	filterUserID := uuid.New()
	c, rec := createTestContext(http.MethodGet,
		"/api/v1/admin/audit-log?page=2&per_page=5&user_id="+filterUserID.String()+
			"&resource_type=cards&action=delete&date_from=2026-01-01&date_to=2026-12-31&search=test", "")

	result := &services.AuditLogResult{
		Logs:  []models.AuditLog{},
		Total: 15,
	}

	expectedFilters := repository.AuditLogFilters{
		UserID:       &filterUserID,
		ResourceType: "cards",
		Action:       "delete",
		DateFrom:     "2026-01-01",
		DateTo:       "2026-12-31",
		SearchQuery:  "test",
		Page:         2,
		PerPage:      5,
	}

	mockAdminService.On("GetAuditLogs", mock.Anything, expectedFilters).Return(result, nil)

	err := handler.GetAuditLogs(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response AuditLogListResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, int64(15), response.Total)
	assert.Equal(t, 2, response.Page)
	assert.Equal(t, 5, response.PerPage)
	assert.Equal(t, 3, response.TotalPages) // ceil(15/5) = 3
	mockAdminService.AssertExpectations(t)
}

func TestAdminHandler_GetAuditLogs_ServiceError(t *testing.T) {
	handler, mockAdminService, _, _, _, _ := setupAdminTest()

	c, rec := createTestContext(http.MethodGet, "/api/v1/admin/audit-log?page=1&per_page=10", "")

	expectedFilters := repository.AuditLogFilters{
		Page:    1,
		PerPage: 10,
	}

	mockAdminService.On("GetAuditLogs", mock.Anything, expectedFilters).Return((*services.AuditLogResult)(nil), errors.New("database error"))

	err := handler.GetAuditLogs(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "server_error", response.Error)
	assert.Equal(t, "Failed to load audit logs", response.Message)
	mockAdminService.AssertExpectations(t)
}

func TestAdminHandler_GetAuditLogs_PerPageOverMax(t *testing.T) {
	handler, mockAdminService, _, _, _, _ := setupAdminTest()

	// per_page > 100 should be reset to 20
	c, rec := createTestContext(http.MethodGet, "/api/v1/admin/audit-log?page=1&per_page=200", "")

	result := &services.AuditLogResult{
		Logs:  []models.AuditLog{},
		Total: 0,
	}

	expectedFilters := repository.AuditLogFilters{
		Page:    1,
		PerPage: 20,
	}

	mockAdminService.On("GetAuditLogs", mock.Anything, expectedFilters).Return(result, nil)

	err := handler.GetAuditLogs(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response AuditLogListResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, 20, response.PerPage)
	mockAdminService.AssertExpectations(t)
}

// ==================== RestoreResource Tests ====================

func TestAdminHandler_RestoreResource_Success(t *testing.T) {
	handler, mockAdminService, _, _, _, _ := setupAdminTest()
	resourceID := uuid.New()

	c, rec := createTestContext(http.MethodPost, "/api/v1/admin/restore/cards/"+resourceID.String(), "")
	c.SetPathValues(echo.PathValues{{Name: "resource_type", Value: "cards"}, {Name: "resource_id", Value: resourceID.String()}})

	mockAdminService.On("RestoreResource", mock.Anything, "cards", resourceID).Return(nil)

	err := handler.RestoreResource(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "Resource restored successfully", response["message"])
	mockAdminService.AssertExpectations(t)
}

func TestAdminHandler_RestoreResource_AllValidTypes(t *testing.T) {
	validTypes := []string{
		"cards", "card_shares", "vouchers", "voucher_shares",
		"gift_cards", "gift_card_shares", "gift_card_transactions", "merchants",
	}

	for _, resourceType := range validTypes {
		t.Run(resourceType, func(t *testing.T) {
			handler, mockAdminService, _, _, _, _ := setupAdminTest()
			resourceID := uuid.New()

			c, rec := createTestContext(http.MethodPost, "/api/v1/admin/restore/"+resourceType+"/"+resourceID.String(), "")
			c.SetPathValues(echo.PathValues{{Name: "resource_type", Value: resourceType}, {Name: "resource_id", Value: resourceID.String()}})

			mockAdminService.On("RestoreResource", mock.Anything, resourceType, resourceID).Return(nil)

			err := handler.RestoreResource(c)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)
			mockAdminService.AssertExpectations(t)
		})
	}
}

func TestAdminHandler_RestoreResource_InvalidResourceType(t *testing.T) {
	handler, _, _, _, _, _ := setupAdminTest()
	resourceID := uuid.New()

	c, rec := createTestContext(http.MethodPost, "/api/v1/admin/restore/users/"+resourceID.String(), "")
	c.SetPathValues(echo.PathValues{{Name: "resource_type", Value: "users"}, {Name: "resource_id", Value: resourceID.String()}})

	err := handler.RestoreResource(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "invalid_resource_type", response.Error)
	assert.Equal(t, "Invalid resource type", response.Message)
}

func TestAdminHandler_RestoreResource_InvalidResourceID(t *testing.T) {
	handler, _, _, _, _, _ := setupAdminTest()

	c, rec := createTestContext(http.MethodPost, "/api/v1/admin/restore/cards/not-a-uuid", "")
	c.SetPathValues(echo.PathValues{{Name: "resource_type", Value: "cards"}, {Name: "resource_id", Value: "not-a-uuid"}})

	err := handler.RestoreResource(c)

	// parseUUIDParam writes JSON response AND returns error
	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "invalid_resource_id", response.Error)
}

func TestAdminHandler_RestoreResource_ResourceNotDeleted(t *testing.T) {
	handler, mockAdminService, _, _, _, _ := setupAdminTest()
	resourceID := uuid.New()

	c, rec := createTestContext(http.MethodPost, "/api/v1/admin/restore/vouchers/"+resourceID.String(), "")
	c.SetPathValues(echo.PathValues{{Name: "resource_type", Value: "vouchers"}, {Name: "resource_id", Value: resourceID.String()}})

	mockAdminService.On("RestoreResource", mock.Anything, "vouchers", resourceID).Return(errors.New("resource is not deleted"))

	err := handler.RestoreResource(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "resource_not_deleted", response.Error)
	assert.Equal(t, "Resource is not deleted", response.Message)
	mockAdminService.AssertExpectations(t)
}

func TestAdminHandler_RestoreResource_ServiceError(t *testing.T) {
	handler, mockAdminService, _, _, _, _ := setupAdminTest()
	resourceID := uuid.New()

	c, rec := createTestContext(http.MethodPost, "/api/v1/admin/restore/gift_cards/"+resourceID.String(), "")
	c.SetPathValues(echo.PathValues{{Name: "resource_type", Value: "gift_cards"}, {Name: "resource_id", Value: resourceID.String()}})

	mockAdminService.On("RestoreResource", mock.Anything, "gift_cards", resourceID).Return(errors.New("database error"))

	err := handler.RestoreResource(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "server_error", response.Error)
	assert.Equal(t, "Failed to restore resource", response.Message)
	mockAdminService.AssertExpectations(t)
}

// ==================== GetSystemHealth Tests ====================

func TestAdminHandler_GetSystemHealth_Success(t *testing.T) {
	handler, _, _, mockHealthService, _, _ := setupAdminTest()

	c, rec := createTestContext(http.MethodGet, "/api/v1/admin/system-health", "")

	latency := int64(5)
	report := &services.ReadinessReport{
		Status:    "ready",
		Timestamp: time.Now(),
		Checks: map[string]services.CheckResult{
			"database": {
				Status:    "healthy",
				Enabled:   true,
				LatencyMs: &latency,
			},
			"smtp": {
				Status:  "not_configured",
				Enabled: false,
			},
		},
	}

	mockHealthService.On("CheckReadiness", mock.Anything).Return(report, nil)

	err := handler.GetSystemHealth(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response services.ReadinessReport
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "ready", response.Status)
	assert.Contains(t, response.Checks, "database")
	assert.Equal(t, "healthy", response.Checks["database"].Status)
	assert.Contains(t, response.Checks, "smtp")
	assert.Equal(t, "not_configured", response.Checks["smtp"].Status)
	mockHealthService.AssertExpectations(t)
}

func TestAdminHandler_GetSystemHealth_ServiceError(t *testing.T) {
	handler, _, _, mockHealthService, _, _ := setupAdminTest()

	c, rec := createTestContext(http.MethodGet, "/api/v1/admin/system-health", "")

	mockHealthService.On("CheckReadiness", mock.Anything).Return((*services.ReadinessReport)(nil), errors.New("health check failed"))

	err := handler.GetSystemHealth(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "server_error", response.Error)
	assert.Equal(t, "Failed to check system health", response.Message)
	mockHealthService.AssertExpectations(t)
}

// ==================== SendTestEmail Tests ====================

func TestAdminHandler_SendTestEmail_Success(t *testing.T) {
	handler, _, _, _, mockEmailSvc, _ := setupAdminTest()

	c, rec := createTestContext(http.MethodPost, "/api/v1/admin/test-email", "")
	admin := &models.User{
		ID:        uuid.New(),
		Email:     "admin@example.com",
		FirstName: "Admin",
		LastName:  "User",
		Language:  "en",
	}
	c.Set("current_user", admin)

	mockEmailSvc.On("SendTestEmail", mock.Anything, "admin@example.com", "Admin", "en").Return(nil)

	err := handler.SendTestEmail(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "Test email sent successfully! Check your inbox.", response["message"])
	mockEmailSvc.AssertExpectations(t)
}

func TestAdminHandler_SendTestEmail_Unauthorized(t *testing.T) {
	handler, _, _, _, _, _ := setupAdminTest()

	c, rec := createTestContext(http.MethodPost, "/api/v1/admin/test-email", "")
	// No current_user set

	err := handler.SendTestEmail(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "unauthorized", response.Error)
}

func TestAdminHandler_SendTestEmail_InvalidUserType(t *testing.T) {
	handler, _, _, _, _, _ := setupAdminTest()

	c, rec := createTestContext(http.MethodPost, "/api/v1/admin/test-email", "")
	c.Set("current_user", "not-a-user-struct")

	err := handler.SendTestEmail(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "server_error", response.Error)
}

func TestAdminHandler_SendTestEmail_SMTPError(t *testing.T) {
	handler, _, _, _, mockEmailSvc, _ := setupAdminTest()

	c, rec := createTestContext(http.MethodPost, "/api/v1/admin/test-email", "")
	admin := &models.User{
		ID:        uuid.New(),
		Email:     "admin@example.com",
		FirstName: "Admin",
		Language:  "de",
	}
	c.Set("current_user", admin)

	mockEmailSvc.On("SendTestEmail", mock.Anything, "admin@example.com", "Admin", "de").Return(errors.New("SMTP connection refused"))

	err := handler.SendTestEmail(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "smtp_error", response.Error)
	assert.Contains(t, response.Message, "Failed to send test email")
	mockEmailSvc.AssertExpectations(t)
}

// ==================== SendTestPush Tests ====================

func TestAdminHandler_SendTestPush_Success(t *testing.T) {
	handler, _, _, _, _, mockPushSvc := setupAdminTest()

	c, rec := createTestContext(http.MethodPost, "/api/v1/admin/test-push", "")
	admin := &models.User{
		ID:        uuid.New(),
		Email:     "admin@example.com",
		FirstName: "Admin",
	}
	c.Set("current_user", admin)

	mockPushSvc.On("SendTestPush", mock.Anything, admin.ID).Return(nil)

	err := handler.SendTestPush(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "Test push notification sent! Check your browser notifications.", response["message"])
	mockPushSvc.AssertExpectations(t)
}

func TestAdminHandler_SendTestPush_Unauthorized(t *testing.T) {
	handler, _, _, _, _, _ := setupAdminTest()

	c, rec := createTestContext(http.MethodPost, "/api/v1/admin/test-push", "")
	// No current_user set

	err := handler.SendTestPush(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "unauthorized", response.Error)
}

func TestAdminHandler_SendTestPush_InvalidUserType(t *testing.T) {
	handler, _, _, _, _, _ := setupAdminTest()

	c, rec := createTestContext(http.MethodPost, "/api/v1/admin/test-push", "")
	c.Set("current_user", "not-a-user-struct")

	err := handler.SendTestPush(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "server_error", response.Error)
}

func TestAdminHandler_SendTestPush_PushNotConfigured(t *testing.T) {
	// Create handler with nil pushService
	mockAdminService := new(mocks.MockAdminServiceInterface)
	mockUserService := new(mocks.MockUserServiceInterface)
	mockHealthService := new(mocks.MockHealthCheckServiceInterface)
	mockEmailSvc := new(mocks.MockEmailServiceInterface)
	handler := NewAdminHandler(mockAdminService, mockUserService, mockHealthService, mockEmailSvc, nil)

	c, rec := createTestContext(http.MethodPost, "/api/v1/admin/test-push", "")
	admin := &models.User{
		ID:        uuid.New(),
		Email:     "admin@example.com",
		FirstName: "Admin",
	}
	c.Set("current_user", admin)

	err := handler.SendTestPush(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "push_not_configured", response.Error)
	assert.Equal(t, "Push notification service is not configured", response.Message)
}

func TestAdminHandler_SendTestPush_PushError(t *testing.T) {
	handler, _, _, _, _, mockPushSvc := setupAdminTest()

	c, rec := createTestContext(http.MethodPost, "/api/v1/admin/test-push", "")
	admin := &models.User{
		ID:        uuid.New(),
		Email:     "admin@example.com",
		FirstName: "Admin",
	}
	c.Set("current_user", admin)

	mockPushSvc.On("SendTestPush", mock.Anything, admin.ID).Return(errors.New("push delivery failed"))

	err := handler.SendTestPush(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "push_error", response.Error)
	assert.Contains(t, response.Message, "Failed to send test push notification")
	mockPushSvc.AssertExpectations(t)
}

// ==================== SendPreviewEmail Tests ====================

func TestAdminHandler_SendPreviewEmail_TestEmail(t *testing.T) {
	handler, _, _, _, mockEmailSvc, _ := setupAdminTest()

	c, rec := createTestContext(http.MethodPost, "/api/v1/admin/preview-email", `{"template":"test_email","language":"en"}`)
	admin := &models.User{
		ID:        uuid.New(),
		Email:     "admin@example.com",
		FirstName: "Admin",
		Language:  "de",
	}
	c.Set("current_user", admin)

	mockEmailSvc.On("SendTestEmail", mock.Anything, "admin@example.com", "Admin", "en").Return(nil)

	err := handler.SendPreviewEmail(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "Preview email sent successfully! Check your inbox.", response["message"])
	mockEmailSvc.AssertExpectations(t)
}

func TestAdminHandler_SendPreviewEmail_PasswordReset(t *testing.T) {
	handler, _, _, _, mockEmailSvc, _ := setupAdminTest()

	c, rec := createTestContext(http.MethodPost, "/api/v1/admin/preview-email", `{"template":"password_reset","language":"fr"}`)
	admin := &models.User{
		ID:        uuid.New(),
		Email:     "admin@example.com",
		FirstName: "Admin",
		Language:  "de",
	}
	c.Set("current_user", admin)

	mockEmailSvc.On("SendPasswordReset", mock.Anything, "admin@example.com", "Admin",
		"https://example.com/reset-password?token=sample-preview-token", "1 hour", "fr").Return(nil)

	err := handler.SendPreviewEmail(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockEmailSvc.AssertExpectations(t)
}

func TestAdminHandler_SendPreviewEmail_EmailVerification(t *testing.T) {
	handler, _, _, _, mockEmailSvc, _ := setupAdminTest()

	c, rec := createTestContext(http.MethodPost, "/api/v1/admin/preview-email", `{"template":"email_verification","language":"de"}`)
	admin := &models.User{
		ID:        uuid.New(),
		Email:     "admin@example.com",
		FirstName: "Admin",
		Language:  "de",
	}
	c.Set("current_user", admin)

	mockEmailSvc.On("SendEmailVerification", mock.Anything, "admin@example.com", "Admin",
		"https://example.com/verify-email?token=sample-preview-token", "de").Return(nil)

	err := handler.SendPreviewEmail(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockEmailSvc.AssertExpectations(t)
}

func TestAdminHandler_SendPreviewEmail_AccountDeleted(t *testing.T) {
	handler, _, _, _, mockEmailSvc, _ := setupAdminTest()

	c, rec := createTestContext(http.MethodPost, "/api/v1/admin/preview-email", `{"template":"account_deleted","language":"en"}`)
	admin := &models.User{
		ID:        uuid.New(),
		Email:     "admin@example.com",
		FirstName: "Admin",
		Language:  "de",
	}
	c.Set("current_user", admin)

	mockEmailSvc.On("SendAccountDeletionConfirmation", mock.Anything, "admin@example.com", "Admin", "en").Return(nil)

	err := handler.SendPreviewEmail(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockEmailSvc.AssertExpectations(t)
}

func TestAdminHandler_SendPreviewEmail_ExpiryReminder(t *testing.T) {
	handler, _, _, _, mockEmailSvc, _ := setupAdminTest()

	c, rec := createTestContext(http.MethodPost, "/api/v1/admin/preview-email", `{"template":"expiry_reminder","language":"en"}`)
	admin := &models.User{
		ID:        uuid.New(),
		Email:     "admin@example.com",
		FirstName: "Admin",
		Language:  "de",
	}
	c.Set("current_user", admin)

	mockEmailSvc.On("SendExpiryReminder", mock.Anything, "admin@example.com", "Admin", mock.AnythingOfType("email.ExpiryReminderData"),
		"https://example.com/unsubscribe?token=sample-preview-token&type=reminders", "en").Return(nil)

	err := handler.SendPreviewEmail(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockEmailSvc.AssertExpectations(t)
}

func TestAdminHandler_SendPreviewEmail_ShareNotification(t *testing.T) {
	handler, _, _, _, mockEmailSvc, _ := setupAdminTest()

	c, rec := createTestContext(http.MethodPost, "/api/v1/admin/preview-email", `{"template":"share_notification","language":"en"}`)
	admin := &models.User{
		ID:        uuid.New(),
		Email:     "admin@example.com",
		FirstName: "Admin",
		Language:  "de",
	}
	c.Set("current_user", admin)

	mockEmailSvc.On("SendShareNotification", mock.Anything, "admin@example.com", "Admin", "Max Muster", "voucher",
		"https://example.com/vouchers/sample-uuid",
		"https://example.com/unsubscribe?token=sample-preview-token&type=notifications", "en").Return(nil)

	err := handler.SendPreviewEmail(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockEmailSvc.AssertExpectations(t)
}

func TestAdminHandler_SendPreviewEmail_TransferNotification(t *testing.T) {
	handler, _, _, _, mockEmailSvc, _ := setupAdminTest()

	c, rec := createTestContext(http.MethodPost, "/api/v1/admin/preview-email", `{"template":"transfer_notification","language":"en"}`)
	admin := &models.User{
		ID:        uuid.New(),
		Email:     "admin@example.com",
		FirstName: "Admin",
		Language:  "de",
	}
	c.Set("current_user", admin)

	mockEmailSvc.On("SendTransferNotification", mock.Anything, "admin@example.com", "Admin", "Max Muster", "gift_card",
		"https://example.com/gift-cards/sample-uuid",
		"https://example.com/unsubscribe?token=sample-preview-token&type=notifications", "en").Return(nil)

	err := handler.SendPreviewEmail(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockEmailSvc.AssertExpectations(t)
}

func TestAdminHandler_SendPreviewEmail_ValidityStart(t *testing.T) {
	handler, _, _, _, mockEmailSvc, _ := setupAdminTest()

	c, rec := createTestContext(http.MethodPost, "/api/v1/admin/preview-email", `{"template":"validity_start","language":"en"}`)
	admin := &models.User{
		ID:        uuid.New(),
		Email:     "admin@example.com",
		FirstName: "Admin",
		Language:  "de",
	}
	c.Set("current_user", admin)

	mockEmailSvc.On("SendValidityStart", mock.Anything, "admin@example.com", "Admin", mock.AnythingOfType("email.ValidityStartData"),
		"https://example.com/unsubscribe?token=sample-preview-token&type=reminders", "en").Return(nil)

	err := handler.SendPreviewEmail(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockEmailSvc.AssertExpectations(t)
}

func TestAdminHandler_SendPreviewEmail_UnknownTemplate(t *testing.T) {
	handler, _, _, _, _, _ := setupAdminTest()

	c, rec := createTestContext(http.MethodPost, "/api/v1/admin/preview-email", `{"template":"nonexistent","language":"en"}`)
	admin := &models.User{
		ID:        uuid.New(),
		Email:     "admin@example.com",
		FirstName: "Admin",
		Language:  "de",
	}
	c.Set("current_user", admin)

	err := handler.SendPreviewEmail(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "invalid_template", response.Error)
	assert.Contains(t, response.Message, "nonexistent")
}

func TestAdminHandler_SendPreviewEmail_Unauthorized(t *testing.T) {
	handler, _, _, _, _, _ := setupAdminTest()

	c, rec := createTestContext(http.MethodPost, "/api/v1/admin/preview-email", `{"template":"test_email"}`)
	// No current_user set

	err := handler.SendPreviewEmail(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "unauthorized", response.Error)
}

func TestAdminHandler_SendPreviewEmail_InvalidUserType(t *testing.T) {
	handler, _, _, _, _, _ := setupAdminTest()

	c, rec := createTestContext(http.MethodPost, "/api/v1/admin/preview-email", `{"template":"test_email"}`)
	c.Set("current_user", "not-a-user-struct")

	err := handler.SendPreviewEmail(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "server_error", response.Error)
}

func TestAdminHandler_SendPreviewEmail_SMTPError(t *testing.T) {
	handler, _, _, _, mockEmailSvc, _ := setupAdminTest()

	c, rec := createTestContext(http.MethodPost, "/api/v1/admin/preview-email", `{"template":"test_email","language":"en"}`)
	admin := &models.User{
		ID:        uuid.New(),
		Email:     "admin@example.com",
		FirstName: "Admin",
		Language:  "de",
	}
	c.Set("current_user", admin)

	mockEmailSvc.On("SendTestEmail", mock.Anything, "admin@example.com", "Admin", "en").Return(errors.New("SMTP error"))

	err := handler.SendPreviewEmail(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "smtp_error", response.Error)
	assert.Contains(t, response.Message, "Failed to send preview email")
	mockEmailSvc.AssertExpectations(t)
}

func TestAdminHandler_SendPreviewEmail_DefaultLanguageFromUser(t *testing.T) {
	// When no language is specified in the request, uses user's language
	handler, _, _, _, mockEmailSvc, _ := setupAdminTest()

	c, rec := createTestContext(http.MethodPost, "/api/v1/admin/preview-email", `{"template":"test_email"}`)
	admin := &models.User{
		ID:        uuid.New(),
		Email:     "admin@example.com",
		FirstName: "Admin",
		Language:  "fr",
	}
	c.Set("current_user", admin)

	// Language should fall back to user's language "fr"
	mockEmailSvc.On("SendTestEmail", mock.Anything, "admin@example.com", "Admin", "fr").Return(nil)

	err := handler.SendPreviewEmail(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockEmailSvc.AssertExpectations(t)
}
