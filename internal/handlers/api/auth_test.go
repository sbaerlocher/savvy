package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"savvy/internal/middleware"
	"savvy/internal/mocks"
	"savvy/internal/models"
	"savvy/internal/services"
)

// ==================== Helper Functions ====================

func setupAuthTest() (*AuthHandler, *mocks.MockUserServiceInterface) {
	// Initialize session store for auth tests using CookieStore (no DB needed)
	cookieStore := sessions.NewCookieStore([]byte("test-secret-key-32chars-minimum"))
	cookieStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: true,
		Secure:   false,
		SameSite: 2,
	}
	middleware.Store = cookieStore

	mockUserService := new(mocks.MockUserServiceInterface)
	handler := NewAuthHandler(mockUserService, nil, nil, "http://localhost:5173")
	return handler, mockUserService
}

func createTestUserWithPassword() *models.User {
	userID := uuid.New()
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	return &models.User{
		ID:           userID,
		Email:        "test@example.com",
		PasswordHash: string(hash),
		FirstName:    "Test",
		LastName:     "User",
		Role:         "user",
		AuthProvider: "local",
	}
}

// ==================== Login Tests ====================

func TestAuthHandler_Login_Success(t *testing.T) {
	handler, mockUserService := setupAuthTest()
	user := createTestUserWithPassword()
	body := `{"email":"test@example.com","password":"password123"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/login", body)

	mockUserService.On("GetUserByEmail", mock.Anything, "test@example.com").Return(user, nil)

	err := handler.Login(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NotNil(t, response["user"])
	mockUserService.AssertExpectations(t)
}

func TestAuthHandler_Login_InvalidRequest(t *testing.T) {
	handler, _ := setupAuthTest()
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/login", "invalid-json")

	err := handler.Login(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "invalid_request", response.Error)
}

func TestAuthHandler_Login_MissingFields(t *testing.T) {
	handler, _ := setupAuthTest()
	body := `{"email":"test@example.com"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/login", body)

	err := handler.Login(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "missing_fields", response.Error)
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	handler, mockUserService := setupAuthTest()
	user := createTestUserWithPassword()
	body := `{"email":"test@example.com","password":"wrongpassword"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/login", body)

	mockUserService.On("GetUserByEmail", mock.Anything, "test@example.com").Return(user, nil)
	mockUserService.On("UpdateUser", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)

	err := handler.Login(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "invalid_credentials", response.Error)
	mockUserService.AssertExpectations(t)
}

func TestAuthHandler_Login_UserNotFound(t *testing.T) {
	handler, mockUserService := setupAuthTest()
	body := `{"email":"notfound@example.com","password":"password123"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/login", body)

	mockUserService.On("GetUserByEmail", mock.Anything, "notfound@example.com").Return((*models.User)(nil), errors.New("not found"))

	err := handler.Login(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "invalid_credentials", response.Error)
	mockUserService.AssertExpectations(t)
}

func TestAuthHandler_Login_AccountLocked(t *testing.T) {
	handler, mockUserService := setupAuthTest()
	user := createTestUserWithPassword()
	lockUntil := time.Now().Add(15 * time.Minute)
	user.LockedUntil = &lockUntil
	body := `{"email":"test@example.com","password":"password123"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/login", body)

	mockUserService.On("GetUserByEmail", mock.Anything, "test@example.com").Return(user, nil)

	err := handler.Login(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "account_locked", response.Error)
	mockUserService.AssertExpectations(t)
}

func TestAuthHandler_Login_EmailNormalization(t *testing.T) {
	handler, mockUserService := setupAuthTest()
	user := createTestUserWithPassword()
	body := `{"email":"  TEST@EXAMPLE.COM  ","password":"password123"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/login", body)

	// Email should be normalized to lowercase and trimmed
	mockUserService.On("GetUserByEmail", mock.Anything, "test@example.com").Return(user, nil)

	err := handler.Login(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockUserService.AssertExpectations(t)
}

func TestAuthHandler_Login_ResetsFailedAttempts(t *testing.T) {
	handler, mockUserService := setupAuthTest()
	user := createTestUserWithPassword()
	user.FailedLoginAttempts = 3
	body := `{"email":"test@example.com","password":"password123"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/login", body)

	mockUserService.On("GetUserByEmail", mock.Anything, "test@example.com").Return(user, nil)
	mockUserService.On("UpdateUser", mock.Anything, mock.MatchedBy(func(u *models.User) bool {
		return u.FailedLoginAttempts == 0
	})).Return(nil)

	err := handler.Login(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockUserService.AssertExpectations(t)
}

// ==================== Register Tests ====================

func TestAuthHandler_Register_Success(t *testing.T) {
	handler, mockUserService := setupAuthTest()
	body := `{"email":"newuser@example.com","password":"SecurePass123!","first_name":"New","last_name":"User"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/register", body)

	mockUserService.On("GetUserByEmail", mock.Anything, "newuser@example.com").Return((*models.User)(nil), errors.New("not found"))
	mockUserService.On("CreateUser", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)

	err := handler.Register(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var response map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NotNil(t, response["user"])
	mockUserService.AssertExpectations(t)
}

func TestAuthHandler_Register_InvalidRequest(t *testing.T) {
	handler, _ := setupAuthTest()
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/register", "invalid-json")

	err := handler.Register(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAuthHandler_Register_MissingFields(t *testing.T) {
	handler, _ := setupAuthTest()
	body := `{"email":"newuser@example.com"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/register", body)

	err := handler.Register(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "missing_fields", response.Error)
}

func TestAuthHandler_Register_WeakPassword(t *testing.T) {
	handler, _ := setupAuthTest()
	body := `{"email":"newuser@example.com","password":"weak","first_name":"New","last_name":"User"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/register", body)

	err := handler.Register(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "weak_password", response.Error)
}

func TestAuthHandler_Register_EmailExists(t *testing.T) {
	handler, mockUserService := setupAuthTest()
	existingUser := createTestUser()
	body := `{"email":"test@example.com","password":"SecurePass123!","first_name":"New","last_name":"User"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/register", body)

	mockUserService.On("GetUserByEmail", mock.Anything, "test@example.com").Return(existingUser, nil)

	err := handler.Register(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusConflict, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "registration_failed", response.Error)
	mockUserService.AssertExpectations(t)
}

func TestAuthHandler_Register_EmailNormalization(t *testing.T) {
	handler, mockUserService := setupAuthTest()
	body := `{"email":"  NEWUSER@EXAMPLE.COM  ","password":"SecurePass123!","first_name":"New","last_name":"User"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/register", body)

	// Email should be normalized to lowercase and trimmed
	mockUserService.On("GetUserByEmail", mock.Anything, "newuser@example.com").Return((*models.User)(nil), errors.New("not found"))
	mockUserService.On("CreateUser", mock.Anything, mock.MatchedBy(func(u *models.User) bool {
		return u.Email == "newuser@example.com"
	})).Return(nil)

	err := handler.Register(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	mockUserService.AssertExpectations(t)
}

func TestAuthHandler_Register_PasswordComplexity_MinLength(t *testing.T) {
	handler, _ := setupAuthTest()
	body := `{"email":"newuser@example.com","password":"Short1!","first_name":"New","last_name":"User"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/register", body)

	err := handler.Register(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "weak_password", response.Error)
	assert.Contains(t, response.Message, "at least 12 characters")
}

func TestAuthHandler_Register_PasswordComplexity_Insufficient(t *testing.T) {
	handler, _ := setupAuthTest()
	body := `{"email":"newuser@example.com","password":"alllowercase","first_name":"New","last_name":"User"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/register", body)

	err := handler.Register(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "weak_password", response.Error)
	assert.Contains(t, response.Message, "at least 3 of")
}

// ==================== Logout Tests ====================

func TestAuthHandler_Logout_Success(t *testing.T) {
	handler, _ := setupAuthTest()
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/logout", "")

	err := handler.Logout(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "Logged out", response["message"])
}

func TestAuthHandler_Logout_NoSession(t *testing.T) {
	handler, _ := setupAuthTest()
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/logout", "")

	err := handler.Logout(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "Logged out", response["message"])
}

// ==================== Me Tests ====================

func TestAuthHandler_Me_Success(t *testing.T) {
	handler, _ := setupAuthTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/auth/me", "")
	user := createTestUser()
	c.Set("current_user", user)
	c.Set("is_impersonating", false)

	err := handler.Me(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	userData := response["user"].(map[string]interface{})
	assert.Equal(t, user.ID.String(), userData["id"])
	assert.Equal(t, user.Email, userData["email"])
	assert.Equal(t, user.FirstName, userData["first_name"])
	assert.Equal(t, user.LastName, userData["last_name"])
	assert.Equal(t, false, userData["is_admin"])
	assert.Equal(t, false, userData["is_impersonating"])
}

func TestAuthHandler_Me_Admin(t *testing.T) {
	handler, _ := setupAuthTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/auth/me", "")
	user := createTestUser()
	user.Role = "admin"
	c.Set("current_user", user)

	err := handler.Me(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	userData := response["user"].(map[string]interface{})
	assert.Equal(t, true, userData["is_admin"])
}

func TestAuthHandler_Me_Impersonating(t *testing.T) {
	handler, _ := setupAuthTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/auth/me", "")
	user := createTestUser()
	c.Set("current_user", user)
	c.Set("is_impersonating", true)

	err := handler.Me(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	userData := response["user"].(map[string]interface{})
	assert.Equal(t, true, userData["is_impersonating"])
}

func TestAuthHandler_Me_NotAuthenticated(t *testing.T) {
	handler, _ := setupAuthTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/auth/me", "")
	// No user set in context

	err := handler.Me(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "unauthorized", response.Error)
}

// ==================== Password Validation Tests ====================

func TestValidatePassword_TooShort(t *testing.T) {
	errCode, err := validatePassword("Short1!")
	assert.Error(t, err)
	assert.Equal(t, "weak_password", errCode)
	assert.Contains(t, err.Error(), "at least 12 characters")
}

func TestValidatePassword_LackComplexity(t *testing.T) {
	errCode, err := validatePassword("onlylowercase")
	assert.Error(t, err)
	assert.Equal(t, "weak_password", errCode)
	assert.Contains(t, err.Error(), "at least 3 of")
}

func TestValidatePassword_Valid_ThreeTypes(t *testing.T) {
	errCode, err := validatePassword("Lowercase123") // uppercase, lowercase, digit
	assert.NoError(t, err)
	assert.Equal(t, "", errCode)
}

func TestValidatePassword_Valid_FourTypes(t *testing.T) {
	errCode, err := validatePassword("LowerCase123!") // all four types
	assert.NoError(t, err)
	assert.Equal(t, "", errCode)
}

func TestValidatePassword_Valid_MinLength(t *testing.T) {
	errCode, err := validatePassword("Lowercase12!") // exactly 12 characters
	assert.NoError(t, err)
	assert.Equal(t, "", errCode)
}

func TestValidatePassword_TooLong(t *testing.T) {
	// 73 characters exceeds bcrypt's 72-byte limit
	longPassword := "Abc123!Abc123!Abc123!Abc123!Abc123!Abc123!Abc123!Abc123!Abc123!Abc123!xyz"
	errCode, err := validatePassword(longPassword)
	assert.Error(t, err)
	assert.Equal(t, "password_too_long", errCode)
}

func TestAuthHandler_Login_PasswordTooLong(t *testing.T) {
	handler, _ := setupAuthTest()
	longPassword := "Abc123!Abc123!Abc123!Abc123!Abc123!Abc123!Abc123!Abc123!Abc123!Abc123!xyz"
	body := `{"email":"test@example.com","password":"` + longPassword + `"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/login", body)

	err := handler.Login(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "password_too_long", response.Error)
}

// ==================== Full Auth Handler Helper ====================

func setupAuthTestFull() (*AuthHandler, *mocks.MockUserServiceInterface, *mocks.MockEmailTokenServiceInterface, *mocks.MockEmailServiceInterface) {
	cookieStore := sessions.NewCookieStore([]byte("test-secret-key-32chars-minimum"))
	cookieStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: true,
		Secure:   false,
		SameSite: 2,
	}
	middleware.Store = cookieStore

	mockUserService := new(mocks.MockUserServiceInterface)
	mockEmailTokenService := new(mocks.MockEmailTokenServiceInterface)
	mockEmailService := new(mocks.MockEmailServiceInterface)
	handler := NewAuthHandler(mockUserService, mockEmailTokenService, mockEmailService, "http://localhost:5173")
	return handler, mockUserService, mockEmailTokenService, mockEmailService
}

// ==================== SetTOTPService / SetSessionService Tests ====================

func TestAuthHandler_SetTOTPService(t *testing.T) {
	handler, _, _, _ := setupAuthTestFull()

	assert.Nil(t, handler.totpService)

	mockTOTP := new(MockTOTPServiceInterface)
	handler.SetTOTPService(mockTOTP)

	assert.NotNil(t, handler.totpService)
	assert.Equal(t, mockTOTP, handler.totpService)
}

func TestAuthHandler_SetSessionService(t *testing.T) {
	handler, _, _, _ := setupAuthTestFull()

	assert.Nil(t, handler.sessionService)

	mockSession := new(mocks.MockSessionServiceInterface)
	handler.SetSessionService(mockSession)

	assert.NotNil(t, handler.sessionService)
	assert.Equal(t, mockSession, handler.sessionService)
}

// ==================== RequestVerification Tests ====================

func TestAuthHandler_RequestVerification_Success(t *testing.T) {
	handler, _, mockEmailTokenService, mockEmailService := setupAuthTestFull()
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/request-verification", "")
	user := createTestUser()
	user.EmailVerified = false
	user.AuthProvider = "local"
	user.Language = "en"
	c.Set("current_user", user)

	mockEmailTokenService.On("CreateVerificationToken", mock.Anything, user.ID).Return("test-token-123", nil)
	mockEmailService.On("SendEmailVerification", mock.Anything, user.Email, user.FirstName, "http://localhost:5173/verify-email?token=test-token-123", "en").Return(nil)

	err := handler.RequestVerification(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "Verification email sent", response["message"])
	mockEmailTokenService.AssertExpectations(t)
	mockEmailService.AssertExpectations(t)
}

func TestAuthHandler_RequestVerification_AlreadyVerified(t *testing.T) {
	handler, _, _, _ := setupAuthTestFull()
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/request-verification", "")
	user := createTestUser()
	user.EmailVerified = true
	c.Set("current_user", user)

	err := handler.RequestVerification(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "Email already verified", response["message"])
	assert.Equal(t, true, response["email_verified"])
}

func TestAuthHandler_RequestVerification_NotAuthenticated(t *testing.T) {
	handler, _, _, _ := setupAuthTestFull()
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/request-verification", "")
	// No user set in context

	err := handler.RequestVerification(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "unauthorized", response.Error)
}

func TestAuthHandler_RequestVerification_TokenCreationFailure(t *testing.T) {
	handler, _, mockEmailTokenService, _ := setupAuthTestFull()
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/request-verification", "")
	user := createTestUser()
	user.EmailVerified = false
	c.Set("current_user", user)

	mockEmailTokenService.On("CreateVerificationToken", mock.Anything, user.ID).Return("", errors.New("db error"))

	err := handler.RequestVerification(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "server_error", response.Error)
	mockEmailTokenService.AssertExpectations(t)
}

func TestAuthHandler_RequestVerification_EmailSendFailure(t *testing.T) {
	handler, _, mockEmailTokenService, mockEmailService := setupAuthTestFull()
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/request-verification", "")
	user := createTestUser()
	user.EmailVerified = false
	user.Language = "de"
	c.Set("current_user", user)

	mockEmailTokenService.On("CreateVerificationToken", mock.Anything, user.ID).Return("test-token", nil)
	mockEmailService.On("SendEmailVerification", mock.Anything, user.Email, user.FirstName, "http://localhost:5173/verify-email?token=test-token", "de").Return(errors.New("smtp error"))

	err := handler.RequestVerification(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "server_error", response.Error)
	mockEmailTokenService.AssertExpectations(t)
	mockEmailService.AssertExpectations(t)
}

func TestAuthHandler_RequestVerification_EmptyFirstName_UsesEmail(t *testing.T) {
	handler, _, mockEmailTokenService, mockEmailService := setupAuthTestFull()
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/request-verification", "")
	user := createTestUser()
	user.EmailVerified = false
	user.FirstName = ""
	user.Language = "en"
	c.Set("current_user", user)

	mockEmailTokenService.On("CreateVerificationToken", mock.Anything, user.ID).Return("test-token", nil)
	// When FirstName is empty, displayName falls back to email
	mockEmailService.On("SendEmailVerification", mock.Anything, user.Email, user.Email, "http://localhost:5173/verify-email?token=test-token", "en").Return(nil)

	err := handler.RequestVerification(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockEmailService.AssertExpectations(t)
}

// ==================== VerifyEmail Tests ====================

func TestAuthHandler_VerifyEmail_Success(t *testing.T) {
	handler, _, mockEmailTokenService, _ := setupAuthTestFull()
	body := `{"token":"valid-verification-token"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/verify-email", body)

	mockEmailTokenService.On("VerifyEmail", mock.Anything, "valid-verification-token").Return(nil)

	err := handler.VerifyEmail(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "Email verified successfully", response["message"])
	assert.Equal(t, true, response["email_verified"])
	mockEmailTokenService.AssertExpectations(t)
}

func TestAuthHandler_VerifyEmail_InvalidRequest(t *testing.T) {
	handler, _, _, _ := setupAuthTestFull()
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/verify-email", "invalid-json")

	err := handler.VerifyEmail(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "invalid_request", response.Error)
}

func TestAuthHandler_VerifyEmail_MissingToken(t *testing.T) {
	handler, _, _, _ := setupAuthTestFull()
	body := `{"token":""}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/verify-email", body)

	err := handler.VerifyEmail(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "missing_fields", response.Error)
}

func TestAuthHandler_VerifyEmail_TokenNotFound(t *testing.T) {
	handler, _, mockEmailTokenService, _ := setupAuthTestFull()
	body := `{"token":"nonexistent-token"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/verify-email", body)

	mockEmailTokenService.On("VerifyEmail", mock.Anything, "nonexistent-token").Return(services.ErrTokenNotFound)

	err := handler.VerifyEmail(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "invalid_token", response.Error)
	mockEmailTokenService.AssertExpectations(t)
}

func TestAuthHandler_VerifyEmail_TokenExpired(t *testing.T) {
	handler, _, mockEmailTokenService, _ := setupAuthTestFull()
	body := `{"token":"expired-token"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/verify-email", body)

	mockEmailTokenService.On("VerifyEmail", mock.Anything, "expired-token").Return(services.ErrTokenExpired)

	err := handler.VerifyEmail(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "token_expired", response.Error)
	mockEmailTokenService.AssertExpectations(t)
}

func TestAuthHandler_VerifyEmail_TokenUsed(t *testing.T) {
	handler, _, mockEmailTokenService, _ := setupAuthTestFull()
	body := `{"token":"used-token"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/verify-email", body)

	mockEmailTokenService.On("VerifyEmail", mock.Anything, "used-token").Return(services.ErrTokenUsed)

	err := handler.VerifyEmail(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "token_used", response.Error)
	mockEmailTokenService.AssertExpectations(t)
}

func TestAuthHandler_VerifyEmail_InternalError(t *testing.T) {
	handler, _, mockEmailTokenService, _ := setupAuthTestFull()
	body := `{"token":"some-token"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/verify-email", body)

	mockEmailTokenService.On("VerifyEmail", mock.Anything, "some-token").Return(errors.New("unexpected db error"))

	err := handler.VerifyEmail(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "server_error", response.Error)
	mockEmailTokenService.AssertExpectations(t)
}

// ==================== UnsubscribeNotifications Tests ====================

func TestAuthHandler_UnsubscribeNotifications_Success(t *testing.T) {
	handler, _, mockEmailTokenService, _ := setupAuthTestFull()
	body := `{"token":"valid-unsub-token"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/unsubscribe-notifications", body)

	mockEmailTokenService.On("UnsubscribeNotifications", mock.Anything, "valid-unsub-token").Return(nil)

	err := handler.UnsubscribeNotifications(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "Successfully unsubscribed from notification emails", response["message"])
	assert.Equal(t, false, response["email_sharing_enabled"])
	mockEmailTokenService.AssertExpectations(t)
}

func TestAuthHandler_UnsubscribeNotifications_InvalidRequest(t *testing.T) {
	handler, _, _, _ := setupAuthTestFull()
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/unsubscribe-notifications", "invalid-json")

	err := handler.UnsubscribeNotifications(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "invalid_request", response.Error)
}

func TestAuthHandler_UnsubscribeNotifications_MissingToken(t *testing.T) {
	handler, _, _, _ := setupAuthTestFull()
	body := `{"token":""}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/unsubscribe-notifications", body)

	err := handler.UnsubscribeNotifications(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "missing_fields", response.Error)
}

func TestAuthHandler_UnsubscribeNotifications_TokenNotFound(t *testing.T) {
	handler, _, mockEmailTokenService, _ := setupAuthTestFull()
	body := `{"token":"bad-token"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/unsubscribe-notifications", body)

	mockEmailTokenService.On("UnsubscribeNotifications", mock.Anything, "bad-token").Return(services.ErrTokenNotFound)

	err := handler.UnsubscribeNotifications(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "invalid_token", response.Error)
	mockEmailTokenService.AssertExpectations(t)
}

func TestAuthHandler_UnsubscribeNotifications_TokenExpired(t *testing.T) {
	handler, _, mockEmailTokenService, _ := setupAuthTestFull()
	body := `{"token":"expired-token"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/unsubscribe-notifications", body)

	mockEmailTokenService.On("UnsubscribeNotifications", mock.Anything, "expired-token").Return(services.ErrTokenExpired)

	err := handler.UnsubscribeNotifications(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "token_expired", response.Error)
	mockEmailTokenService.AssertExpectations(t)
}

func TestAuthHandler_UnsubscribeNotifications_TokenUsed(t *testing.T) {
	handler, _, mockEmailTokenService, _ := setupAuthTestFull()
	body := `{"token":"used-token"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/unsubscribe-notifications", body)

	mockEmailTokenService.On("UnsubscribeNotifications", mock.Anything, "used-token").Return(services.ErrTokenUsed)

	err := handler.UnsubscribeNotifications(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "token_used", response.Error)
	mockEmailTokenService.AssertExpectations(t)
}

func TestAuthHandler_UnsubscribeNotifications_InternalError(t *testing.T) {
	handler, _, mockEmailTokenService, _ := setupAuthTestFull()
	body := `{"token":"some-token"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/unsubscribe-notifications", body)

	mockEmailTokenService.On("UnsubscribeNotifications", mock.Anything, "some-token").Return(errors.New("db error"))

	err := handler.UnsubscribeNotifications(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "server_error", response.Error)
	mockEmailTokenService.AssertExpectations(t)
}

// ==================== UnsubscribeReminders Tests ====================

func TestAuthHandler_UnsubscribeReminders_Success(t *testing.T) {
	handler, _, mockEmailTokenService, _ := setupAuthTestFull()
	body := `{"token":"valid-unsub-token"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/unsubscribe-reminders", body)

	mockEmailTokenService.On("UnsubscribeReminders", mock.Anything, "valid-unsub-token").Return(nil)

	err := handler.UnsubscribeReminders(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "Successfully unsubscribed from expiry reminder emails", response["message"])
	assert.Equal(t, false, response["email_reminders_enabled"])
	mockEmailTokenService.AssertExpectations(t)
}

func TestAuthHandler_UnsubscribeReminders_InvalidRequest(t *testing.T) {
	handler, _, _, _ := setupAuthTestFull()
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/unsubscribe-reminders", "invalid-json")

	err := handler.UnsubscribeReminders(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "invalid_request", response.Error)
}

func TestAuthHandler_UnsubscribeReminders_MissingToken(t *testing.T) {
	handler, _, _, _ := setupAuthTestFull()
	body := `{"token":""}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/unsubscribe-reminders", body)

	err := handler.UnsubscribeReminders(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "missing_fields", response.Error)
}

func TestAuthHandler_UnsubscribeReminders_TokenNotFound(t *testing.T) {
	handler, _, mockEmailTokenService, _ := setupAuthTestFull()
	body := `{"token":"bad-token"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/unsubscribe-reminders", body)

	mockEmailTokenService.On("UnsubscribeReminders", mock.Anything, "bad-token").Return(services.ErrTokenNotFound)

	err := handler.UnsubscribeReminders(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "invalid_token", response.Error)
	mockEmailTokenService.AssertExpectations(t)
}

func TestAuthHandler_UnsubscribeReminders_TokenExpired(t *testing.T) {
	handler, _, mockEmailTokenService, _ := setupAuthTestFull()
	body := `{"token":"expired-token"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/unsubscribe-reminders", body)

	mockEmailTokenService.On("UnsubscribeReminders", mock.Anything, "expired-token").Return(services.ErrTokenExpired)

	err := handler.UnsubscribeReminders(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "token_expired", response.Error)
	mockEmailTokenService.AssertExpectations(t)
}

func TestAuthHandler_UnsubscribeReminders_TokenUsed(t *testing.T) {
	handler, _, mockEmailTokenService, _ := setupAuthTestFull()
	body := `{"token":"used-token"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/unsubscribe-reminders", body)

	mockEmailTokenService.On("UnsubscribeReminders", mock.Anything, "used-token").Return(services.ErrTokenUsed)

	err := handler.UnsubscribeReminders(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "token_used", response.Error)
	mockEmailTokenService.AssertExpectations(t)
}

func TestAuthHandler_UnsubscribeReminders_InternalError(t *testing.T) {
	handler, _, mockEmailTokenService, _ := setupAuthTestFull()
	body := `{"token":"some-token"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/unsubscribe-reminders", body)

	mockEmailTokenService.On("UnsubscribeReminders", mock.Anything, "some-token").Return(errors.New("db error"))

	err := handler.UnsubscribeReminders(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "server_error", response.Error)
	mockEmailTokenService.AssertExpectations(t)
}

// ==================== RequestPasswordReset Tests ====================

func TestAuthHandler_RequestPasswordReset_Success(t *testing.T) {
	handler, mockUserService, mockEmailTokenService, mockEmailService := setupAuthTestFull()
	body := `{"email":"test@example.com"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/forgot-password", body)

	user := createTestUser()
	user.AuthProvider = "local"
	user.Language = "en"

	mockUserService.On("GetUserByEmail", mock.Anything, "test@example.com").Return(user, nil)
	mockEmailTokenService.On("CreatePasswordResetToken", mock.Anything, user.ID).Return("reset-token-123", nil)
	mockEmailService.On("SendPasswordReset", mock.Anything, user.Email, user.FirstName, "http://localhost:5173/reset-password?token=reset-token-123", "1 hour", "en").Return(nil)

	err := handler.RequestPasswordReset(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Contains(t, response["message"], "If an account exists")
	mockUserService.AssertExpectations(t)
	mockEmailTokenService.AssertExpectations(t)
	mockEmailService.AssertExpectations(t)
}

func TestAuthHandler_RequestPasswordReset_InvalidRequest(t *testing.T) {
	handler, _, _, _ := setupAuthTestFull()
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/forgot-password", "invalid-json")

	err := handler.RequestPasswordReset(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "invalid_request", response.Error)
}

func TestAuthHandler_RequestPasswordReset_MissingEmail(t *testing.T) {
	handler, _, _, _ := setupAuthTestFull()
	body := `{"email":""}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/forgot-password", body)

	err := handler.RequestPasswordReset(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "missing_fields", response.Error)
}

func TestAuthHandler_RequestPasswordReset_UserNotFound(t *testing.T) {
	handler, mockUserService, _, _ := setupAuthTestFull()
	body := `{"email":"nonexistent@example.com"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/forgot-password", body)

	mockUserService.On("GetUserByEmail", mock.Anything, "nonexistent@example.com").Return((*models.User)(nil), errors.New("not found"))

	err := handler.RequestPasswordReset(c)

	assert.NoError(t, err)
	// Always returns 200 to prevent email enumeration
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Contains(t, response["message"], "If an account exists")
	mockUserService.AssertExpectations(t)
}

func TestAuthHandler_RequestPasswordReset_OAuthUser(t *testing.T) {
	handler, mockUserService, _, _ := setupAuthTestFull()
	body := `{"email":"oauth@example.com"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/forgot-password", body)

	user := createTestUser()
	user.Email = "oauth@example.com"
	user.AuthProvider = "oauth"

	mockUserService.On("GetUserByEmail", mock.Anything, "oauth@example.com").Return(user, nil)

	err := handler.RequestPasswordReset(c)

	assert.NoError(t, err)
	// Returns 200 without creating token or sending email
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Contains(t, response["message"], "If an account exists")
	mockUserService.AssertExpectations(t)
}

func TestAuthHandler_RequestPasswordReset_TokenCreationFailure(t *testing.T) {
	handler, mockUserService, mockEmailTokenService, _ := setupAuthTestFull()
	body := `{"email":"test@example.com"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/forgot-password", body)

	user := createTestUser()
	user.AuthProvider = "local"

	mockUserService.On("GetUserByEmail", mock.Anything, "test@example.com").Return(user, nil)
	mockEmailTokenService.On("CreatePasswordResetToken", mock.Anything, user.ID).Return("", errors.New("db error"))

	err := handler.RequestPasswordReset(c)

	assert.NoError(t, err)
	// Still returns 200 to prevent enumeration
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Contains(t, response["message"], "If an account exists")
	mockUserService.AssertExpectations(t)
	mockEmailTokenService.AssertExpectations(t)
}

func TestAuthHandler_RequestPasswordReset_EmailSendFailure(t *testing.T) {
	handler, mockUserService, mockEmailTokenService, mockEmailService := setupAuthTestFull()
	body := `{"email":"test@example.com"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/forgot-password", body)

	user := createTestUser()
	user.AuthProvider = "local"
	user.Language = "de"

	mockUserService.On("GetUserByEmail", mock.Anything, "test@example.com").Return(user, nil)
	mockEmailTokenService.On("CreatePasswordResetToken", mock.Anything, user.ID).Return("reset-token", nil)
	mockEmailService.On("SendPasswordReset", mock.Anything, user.Email, user.FirstName, "http://localhost:5173/reset-password?token=reset-token", "1 hour", "de").Return(errors.New("smtp error"))

	err := handler.RequestPasswordReset(c)

	assert.NoError(t, err)
	// Still returns 200 even when email fails
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Contains(t, response["message"], "If an account exists")
	mockUserService.AssertExpectations(t)
	mockEmailTokenService.AssertExpectations(t)
	mockEmailService.AssertExpectations(t)
}

func TestAuthHandler_RequestPasswordReset_EmailNormalization(t *testing.T) {
	handler, mockUserService, mockEmailTokenService, mockEmailService := setupAuthTestFull()
	body := `{"email":"  TEST@EXAMPLE.COM  "}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/forgot-password", body)

	user := createTestUser()
	user.AuthProvider = "local"
	user.Language = "en"

	// Email should be normalized to lowercase and trimmed
	mockUserService.On("GetUserByEmail", mock.Anything, "test@example.com").Return(user, nil)
	mockEmailTokenService.On("CreatePasswordResetToken", mock.Anything, user.ID).Return("reset-token", nil)
	mockEmailService.On("SendPasswordReset", mock.Anything, user.Email, user.FirstName, mock.Anything, "1 hour", "en").Return(nil)

	err := handler.RequestPasswordReset(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockUserService.AssertExpectations(t)
}

func TestAuthHandler_RequestPasswordReset_EmptyFirstName_UsesEmail(t *testing.T) {
	handler, mockUserService, mockEmailTokenService, mockEmailService := setupAuthTestFull()
	body := `{"email":"test@example.com"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/forgot-password", body)

	user := createTestUser()
	user.AuthProvider = "local"
	user.FirstName = ""
	user.Language = "en"

	mockUserService.On("GetUserByEmail", mock.Anything, "test@example.com").Return(user, nil)
	mockEmailTokenService.On("CreatePasswordResetToken", mock.Anything, user.ID).Return("reset-token", nil)
	// When FirstName is empty, displayName falls back to email
	mockEmailService.On("SendPasswordReset", mock.Anything, user.Email, user.Email, mock.Anything, "1 hour", "en").Return(nil)

	err := handler.RequestPasswordReset(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockEmailService.AssertExpectations(t)
}

// ==================== ResetPassword Tests ====================

func TestAuthHandler_ResetPassword_Success(t *testing.T) {
	handler, mockUserService, mockEmailTokenService, _ := setupAuthTestFull()
	body := `{"token":"valid-reset-token","password":"NewSecurePass123!"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/reset-password", body)

	resetUser := &models.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		FirstName:    "Test",
		LastName:     "User",
		Role:         "user",
		AuthProvider: "local",
	}

	mockEmailTokenService.On("ConsumePasswordResetToken", mock.Anything, "valid-reset-token").Return(resetUser, nil)
	mockUserService.On("UpdateUser", mock.Anything, mock.MatchedBy(func(u *models.User) bool {
		// Password hash should be set, lockout reset, and PasswordChangedAt set
		return u.PasswordHash != "" &&
			u.FailedLoginAttempts == 0 &&
			u.LockedUntil == nil &&
			u.PasswordChangedAt != nil
	})).Return(nil)

	err := handler.ResetPassword(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Contains(t, response["message"], "Password has been reset successfully")
	mockEmailTokenService.AssertExpectations(t)
	mockUserService.AssertExpectations(t)
}

func TestAuthHandler_ResetPassword_InvalidRequest(t *testing.T) {
	handler, _, _, _ := setupAuthTestFull()
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/reset-password", "invalid-json")

	err := handler.ResetPassword(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "invalid_request", response.Error)
}

func TestAuthHandler_ResetPassword_MissingToken(t *testing.T) {
	handler, _, _, _ := setupAuthTestFull()
	body := `{"token":"","password":"NewSecurePass123!"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/reset-password", body)

	err := handler.ResetPassword(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "missing_fields", response.Error)
}

func TestAuthHandler_ResetPassword_MissingPassword(t *testing.T) {
	handler, _, _, _ := setupAuthTestFull()
	body := `{"token":"valid-token","password":""}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/reset-password", body)

	err := handler.ResetPassword(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "missing_fields", response.Error)
}

func TestAuthHandler_ResetPassword_WeakPassword(t *testing.T) {
	handler, _, _, _ := setupAuthTestFull()
	body := `{"token":"valid-token","password":"weak"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/reset-password", body)

	err := handler.ResetPassword(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "weak_password", response.Error)
}

func TestAuthHandler_ResetPassword_TokenNotFound(t *testing.T) {
	handler, _, mockEmailTokenService, _ := setupAuthTestFull()
	body := `{"token":"bad-token","password":"NewSecurePass123!"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/reset-password", body)

	mockEmailTokenService.On("ConsumePasswordResetToken", mock.Anything, "bad-token").Return(nil, services.ErrTokenNotFound)

	err := handler.ResetPassword(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "invalid_token", response.Error)
	mockEmailTokenService.AssertExpectations(t)
}

func TestAuthHandler_ResetPassword_TokenExpired(t *testing.T) {
	handler, _, mockEmailTokenService, _ := setupAuthTestFull()
	body := `{"token":"expired-token","password":"NewSecurePass123!"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/reset-password", body)

	mockEmailTokenService.On("ConsumePasswordResetToken", mock.Anything, "expired-token").Return(nil, services.ErrTokenExpired)

	err := handler.ResetPassword(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "token_expired", response.Error)
	mockEmailTokenService.AssertExpectations(t)
}

func TestAuthHandler_ResetPassword_TokenUsed(t *testing.T) {
	handler, _, mockEmailTokenService, _ := setupAuthTestFull()
	body := `{"token":"used-token","password":"NewSecurePass123!"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/reset-password", body)

	mockEmailTokenService.On("ConsumePasswordResetToken", mock.Anything, "used-token").Return(nil, services.ErrTokenUsed)

	err := handler.ResetPassword(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "token_used", response.Error)
	mockEmailTokenService.AssertExpectations(t)
}

func TestAuthHandler_ResetPassword_ConsumeTokenInternalError(t *testing.T) {
	handler, _, mockEmailTokenService, _ := setupAuthTestFull()
	body := `{"token":"some-token","password":"NewSecurePass123!"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/reset-password", body)

	mockEmailTokenService.On("ConsumePasswordResetToken", mock.Anything, "some-token").Return(nil, errors.New("unexpected error"))

	err := handler.ResetPassword(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "server_error", response.Error)
	mockEmailTokenService.AssertExpectations(t)
}

func TestAuthHandler_ResetPassword_UpdateUserFailure(t *testing.T) {
	handler, mockUserService, mockEmailTokenService, _ := setupAuthTestFull()
	body := `{"token":"valid-token","password":"NewSecurePass123!"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/reset-password", body)

	resetUser := &models.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		AuthProvider: "local",
	}

	mockEmailTokenService.On("ConsumePasswordResetToken", mock.Anything, "valid-token").Return(resetUser, nil)
	mockUserService.On("UpdateUser", mock.Anything, mock.AnythingOfType("*models.User")).Return(errors.New("db error"))

	err := handler.ResetPassword(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "server_error", response.Error)
	mockEmailTokenService.AssertExpectations(t)
	mockUserService.AssertExpectations(t)
}

func TestAuthHandler_ResetPassword_WithSessionService_RevokesAllSessions(t *testing.T) {
	handler, mockUserService, mockEmailTokenService, _ := setupAuthTestFull()
	mockSessionService := new(mocks.MockSessionServiceInterface)
	handler.SetSessionService(mockSessionService)

	body := `{"token":"valid-token","password":"NewSecurePass123!"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/reset-password", body)

	resetUser := &models.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		AuthProvider: "local",
	}

	mockEmailTokenService.On("ConsumePasswordResetToken", mock.Anything, "valid-token").Return(resetUser, nil)
	mockUserService.On("UpdateUser", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)
	mockSessionService.On("RevokeAllSessions", mock.Anything, resetUser.ID).Return(int64(3), nil)

	err := handler.ResetPassword(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Contains(t, response["message"], "Password has been reset successfully")
	mockSessionService.AssertExpectations(t)
}

func TestAuthHandler_ResetPassword_WithSessionService_RevokeError(t *testing.T) {
	handler, mockUserService, mockEmailTokenService, _ := setupAuthTestFull()
	mockSessionService := new(mocks.MockSessionServiceInterface)
	handler.SetSessionService(mockSessionService)

	body := `{"token":"valid-token","password":"NewSecurePass123!"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/reset-password", body)

	resetUser := &models.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		AuthProvider: "local",
	}

	mockEmailTokenService.On("ConsumePasswordResetToken", mock.Anything, "valid-token").Return(resetUser, nil)
	mockUserService.On("UpdateUser", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)
	// Session revocation error should not prevent success response
	mockSessionService.On("RevokeAllSessions", mock.Anything, resetUser.ID).Return(int64(0), errors.New("session cleanup error"))

	err := handler.ResetPassword(c)

	assert.NoError(t, err)
	// Should still return 200 even if session revocation fails
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Contains(t, response["message"], "Password has been reset successfully")
	mockSessionService.AssertExpectations(t)
}

func TestAuthHandler_ResetPassword_WithoutSessionService(t *testing.T) {
	handler, mockUserService, mockEmailTokenService, _ := setupAuthTestFull()
	// Do NOT set session service - should still work

	body := `{"token":"valid-token","password":"NewSecurePass123!"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/reset-password", body)

	resetUser := &models.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		AuthProvider: "local",
	}

	mockEmailTokenService.On("ConsumePasswordResetToken", mock.Anything, "valid-token").Return(resetUser, nil)
	mockUserService.On("UpdateUser", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)

	err := handler.ResetPassword(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Contains(t, response["message"], "Password has been reset successfully")
	mockEmailTokenService.AssertExpectations(t)
	mockUserService.AssertExpectations(t)
}

func TestAuthHandler_ResetPassword_PasswordComplexity_NoUppercase(t *testing.T) {
	handler, _, _, _ := setupAuthTestFull()
	body := `{"token":"valid-token","password":"alllowercase1"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/reset-password", body)

	err := handler.ResetPassword(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "weak_password", response.Error)
	assert.Contains(t, response.Message, "at least 3 of")
}
