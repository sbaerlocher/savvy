package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"savvy/internal/mocks"
	"savvy/internal/models"
)

// ==================== Helper Functions ====================

func setupProfileTest() (*ProfileHandler, *mocks.MockUserServiceInterface, *mocks.MockAccountServiceInterface) {
	mockUserService := new(mocks.MockUserServiceInterface)
	mockAccountService := new(mocks.MockAccountServiceInterface)
	handler := NewProfileHandler(mockUserService, mockAccountService)
	return handler, mockUserService, mockAccountService
}

func createTestLocalUser() *models.User {
	user := createTestUser()
	user.AuthProvider = "local"
	hash, _ := bcrypt.GenerateFromPassword([]byte("OldPassword123!"), 12)
	user.PasswordHash = string(hash)
	return user
}

// ==================== GetProfile Tests ====================

func TestProfileHandler_GetProfile_Success(t *testing.T) {
	handler, _, _ := setupProfileTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/profile", "")
	user := createTestUser()
	c.Set("current_user", user)

	err := handler.GetProfile(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, user.Email, response["profile"]["email"])
	assert.Equal(t, user.FirstName, response["profile"]["first_name"])
}

func TestProfileHandler_GetProfile_Unauthorized(t *testing.T) {
	handler, _, _ := setupProfileTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/profile", "")
	// No user set

	err := handler.GetProfile(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// ==================== UpdateProfile Tests ====================

func TestProfileHandler_UpdateProfile_Success(t *testing.T) {
	handler, mockUserService, _ := setupProfileTest()
	body := `{"first_name":"Updated","last_name":"Name","language":"de"}`
	c, rec := createTestContext(http.MethodPatch, "/api/v1/profile", body)
	user := createTestLocalUser()
	c.Set("current_user", user)

	mockUserService.On("UpdateUser", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)

	err := handler.UpdateProfile(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "Updated", response["profile"]["first_name"])
	assert.Equal(t, "Name", response["profile"]["last_name"])
	assert.Equal(t, "de", response["profile"]["language"])
	mockUserService.AssertExpectations(t)
}

func TestProfileHandler_UpdateProfile_Unauthorized(t *testing.T) {
	handler, _, _ := setupProfileTest()
	body := `{"first_name":"Updated"}`
	c, rec := createTestContext(http.MethodPatch, "/api/v1/profile", body)

	err := handler.UpdateProfile(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestProfileHandler_UpdateProfile_InvalidJSON(t *testing.T) {
	handler, _, _ := setupProfileTest()
	c, rec := createTestContext(http.MethodPatch, "/api/v1/profile", "invalid-json")
	user := createTestUser()
	c.Set("current_user", user)

	err := handler.UpdateProfile(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestProfileHandler_UpdateProfile_OAuthUserCannotChangeName(t *testing.T) {
	handler, mockUserService, _ := setupProfileTest()
	body := `{"first_name":"NewName"}`
	c, rec := createTestContext(http.MethodPatch, "/api/v1/profile", body)
	user := createTestUser()
	user.AuthProvider = "oauth"
	user.FirstName = "Original"
	c.Set("current_user", user)

	mockUserService.On("UpdateUser", mock.Anything, mock.MatchedBy(func(u *models.User) bool {
		return u.FirstName == "Original" // Name should NOT be changed for OAuth users
	})).Return(nil)

	err := handler.UpdateProfile(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockUserService.AssertExpectations(t)
}

func TestProfileHandler_UpdateProfile_InvalidLanguage(t *testing.T) {
	handler, mockUserService, _ := setupProfileTest()
	body := `{"language":"xx"}`
	c, rec := createTestContext(http.MethodPatch, "/api/v1/profile", body)
	user := createTestUser()
	user.Language = "en"
	c.Set("current_user", user)

	mockUserService.On("UpdateUser", mock.Anything, mock.MatchedBy(func(u *models.User) bool {
		return u.Language == "en" // Language should NOT be changed for invalid value
	})).Return(nil)

	err := handler.UpdateProfile(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockUserService.AssertExpectations(t)
}

func TestProfileHandler_UpdateProfile_NotificationPreferences(t *testing.T) {
	handler, mockUserService, _ := setupProfileTest()
	body := `{"push_notifications_enabled":false,"email_reminders_enabled":false}`
	c, rec := createTestContext(http.MethodPatch, "/api/v1/profile", body)
	user := createTestUser()
	user.PushNotificationsEnabled = true
	user.EmailRemindersEnabled = true
	c.Set("current_user", user)

	mockUserService.On("UpdateUser", mock.Anything, mock.MatchedBy(func(u *models.User) bool {
		return !u.PushNotificationsEnabled && !u.EmailRemindersEnabled
	})).Return(nil)

	err := handler.UpdateProfile(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockUserService.AssertExpectations(t)
}

func TestProfileHandler_UpdateProfile_ServiceError(t *testing.T) {
	handler, mockUserService, _ := setupProfileTest()
	body := `{"first_name":"Updated"}`
	c, rec := createTestContext(http.MethodPatch, "/api/v1/profile", body)
	user := createTestLocalUser()
	c.Set("current_user", user)

	mockUserService.On("UpdateUser", mock.Anything, mock.AnythingOfType("*models.User")).
		Return(errors.New("database error"))

	err := handler.UpdateProfile(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockUserService.AssertExpectations(t)
}

// ==================== ChangePassword Tests ====================

func TestProfileHandler_ChangePassword_Unauthorized(t *testing.T) {
	handler, _, _ := setupProfileTest()
	body := `{"current_password":"old","new_password":"new"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/profile/change-password", body)

	err := handler.ChangePassword(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestProfileHandler_ChangePassword_OAuthUser(t *testing.T) {
	handler, _, _ := setupProfileTest()
	body := `{"current_password":"old","new_password":"new"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/profile/change-password", body)
	user := createTestUser()
	user.AuthProvider = "oauth"
	c.Set("current_user", user)

	err := handler.ChangePassword(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "oauth_user", response.Error)
}

func TestProfileHandler_ChangePassword_MissingFields(t *testing.T) {
	handler, _, _ := setupProfileTest()
	body := `{"current_password":"","new_password":""}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/profile/change-password", body)
	user := createTestLocalUser()
	c.Set("current_user", user)

	err := handler.ChangePassword(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestProfileHandler_ChangePassword_WrongCurrentPassword(t *testing.T) {
	handler, _, _ := setupProfileTest()
	body := `{"current_password":"WrongPassword123!","new_password":"NewPassword123!"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/profile/change-password", body)
	user := createTestLocalUser()
	c.Set("current_user", user)

	err := handler.ChangePassword(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "invalid_password", response.Error)
}

func TestProfileHandler_ChangePassword_WeakNewPassword(t *testing.T) {
	handler, _, _ := setupProfileTest()
	body := `{"current_password":"OldPassword123!","new_password":"short"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/profile/change-password", body)
	user := createTestLocalUser()
	c.Set("current_user", user)

	err := handler.ChangePassword(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "weak_password", response.Error)
}

func TestProfileHandler_ChangePassword_Success(t *testing.T) {
	handler, mockUserService, _ := setupProfileTest()
	body := `{"current_password":"OldPassword123!","new_password":"NewSecurePass123!"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/profile/change-password", body)
	user := createTestLocalUser()
	c.Set("current_user", user)

	mockUserService.On("UpdateUser", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)

	err := handler.ChangePassword(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockUserService.AssertExpectations(t)
}

// ==================== DeleteAccount Tests ====================

func TestProfileHandler_DeleteAccount_Unauthorized(t *testing.T) {
	handler, _, _ := setupProfileTest()
	body := `{"password":"test","confirmation":"DELETE"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/profile/delete-account", body)

	err := handler.DeleteAccount(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestProfileHandler_DeleteAccount_MissingConfirmation(t *testing.T) {
	handler, _, _ := setupProfileTest()
	body := `{"password":"OldPassword123!","confirmation":"WRONG"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/profile/delete-account", body)
	user := createTestLocalUser()
	c.Set("current_user", user)

	err := handler.DeleteAccount(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "confirmation_required", response.Error)
}

func TestProfileHandler_DeleteAccount_LocalUser_MissingPassword(t *testing.T) {
	handler, _, _ := setupProfileTest()
	body := `{"password":"","confirmation":"DELETE"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/profile/delete-account", body)
	user := createTestLocalUser()
	c.Set("current_user", user)

	err := handler.DeleteAccount(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "password_required", response.Error)
}

func TestProfileHandler_DeleteAccount_WrongPassword(t *testing.T) {
	handler, _, _ := setupProfileTest()
	body := `{"password":"WrongPassword123!","confirmation":"DELETE"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/profile/delete-account", body)
	user := createTestLocalUser()
	c.Set("current_user", user)

	err := handler.DeleteAccount(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "invalid_password", response.Error)
}

func TestProfileHandler_DeleteAccount_Success(t *testing.T) {
	handler, _, mockAccountService := setupProfileTest()
	body := `{"password":"OldPassword123!","confirmation":"DELETE"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/profile/delete-account", body)
	user := createTestLocalUser()
	c.Set("current_user", user)

	mockAccountService.On("DeleteAccount", mock.Anything, user.ID).Return(nil)

	err := handler.DeleteAccount(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockAccountService.AssertExpectations(t)
}

func TestProfileHandler_DeleteAccount_OAuthUser_NoPasswordNeeded(t *testing.T) {
	handler, _, mockAccountService := setupProfileTest()
	body := `{"confirmation":"DELETE"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/profile/delete-account", body)
	user := createTestUser()
	user.AuthProvider = "oauth"
	c.Set("current_user", user)

	mockAccountService.On("DeleteAccount", mock.Anything, user.ID).Return(nil)

	err := handler.DeleteAccount(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockAccountService.AssertExpectations(t)
}

func TestProfileHandler_DeleteAccount_ServiceError(t *testing.T) {
	handler, _, mockAccountService := setupProfileTest()
	body := `{"password":"OldPassword123!","confirmation":"DELETE"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/profile/delete-account", body)
	user := createTestLocalUser()
	c.Set("current_user", user)

	mockAccountService.On("DeleteAccount", mock.Anything, user.ID).
		Return(errors.New("database error"))

	err := handler.DeleteAccount(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockAccountService.AssertExpectations(t)
}

// ==================== SetSessionService Tests ====================

func TestProfileHandler_SetSessionService(t *testing.T) {
	handler, _, _ := setupProfileTest()
	assert.Nil(t, handler.sessionService)

	mockSessionService := new(mocks.MockSessionServiceInterface)
	handler.SetSessionService(mockSessionService)
	assert.NotNil(t, handler.sessionService)
}
