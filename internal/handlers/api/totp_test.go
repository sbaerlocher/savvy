package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"savvy/internal/middleware"
	"savvy/internal/models"
	"savvy/internal/services"
)

// ==================== Mock TOTP Service ====================

// MockTOTPServiceInterface is a mock implementation of services.TOTPServiceInterface.
type MockTOTPServiceInterface struct {
	mock.Mock
}

var _ services.TOTPServiceInterface = (*MockTOTPServiceInterface)(nil)

func (m *MockTOTPServiceInterface) GenerateSetup(ctx context.Context, userID uuid.UUID, email string) (*services.TOTPSetupResponse, error) {
	args := m.Called(ctx, userID, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.TOTPSetupResponse), args.Error(1)
}

func (m *MockTOTPServiceInterface) VerifyAndEnable(ctx context.Context, userID uuid.UUID, code string) error {
	args := m.Called(ctx, userID, code)
	return args.Error(0)
}

func (m *MockTOTPServiceInterface) Verify(ctx context.Context, userID uuid.UUID, code string) (bool, error) {
	args := m.Called(ctx, userID, code)
	return args.Bool(0), args.Error(1)
}

func (m *MockTOTPServiceInterface) VerifyBackupCode(ctx context.Context, userID uuid.UUID, code string) (bool, error) {
	args := m.Called(ctx, userID, code)
	return args.Bool(0), args.Error(1)
}

func (m *MockTOTPServiceInterface) Disable(ctx context.Context, userID uuid.UUID, code string) error {
	args := m.Called(ctx, userID, code)
	return args.Error(0)
}

func (m *MockTOTPServiceInterface) IsEnabled(ctx context.Context, userID uuid.UUID) (bool, error) {
	args := m.Called(ctx, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockTOTPServiceInterface) RegenerateBackupCodes(ctx context.Context, userID uuid.UUID, code string) ([]string, error) {
	args := m.Called(ctx, userID, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

// ==================== Test Helpers ====================

func setupTOTPTest() (*TOTPHandler, *MockTOTPServiceInterface) {
	mockService := new(MockTOTPServiceInterface)
	handler := NewTOTPHandler(mockService)
	return handler, mockService
}

func createLocalAuthUser() *models.User {
	user := createTestUser()
	user.AuthProvider = "local"
	return user
}

func createOAuthUser() *models.User {
	user := createTestUser()
	user.AuthProvider = "oauth"
	return user
}

// initSessionStore initializes the middleware session store for challenge tests.
func initSessionStore() {
	middleware.Store = sessions.NewCookieStore([]byte("test-secret-key-for-unit-tests!!"))
}

// ==================== Setup Tests ====================

func TestTOTPHandler_Setup_Success(t *testing.T) {
	handler, mockService := setupTOTPTest()
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/2fa/setup", "")
	user := createLocalAuthUser()
	c.Set("current_user", user)

	expectedSetup := &services.TOTPSetupResponse{ // #nosec G101 -- test data, not real credentials
		Secret:      "JBSWY3DPEHPK3PXP",
		QRCodeURL:   "otpauth://totp/Savvy:test@example.com?secret=JBSWY3DPEHPK3PXP&issuer=Savvy",
		BackupCodes: []string{"ABCD-1234", "EFGH-5678"},
	}
	mockService.On("GenerateSetup", mock.Anything, user.ID, user.Email).Return(expectedSetup, nil)

	err := handler.Setup(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response services.TOTPSetupResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, expectedSetup.Secret, response.Secret)
	assert.Equal(t, expectedSetup.QRCodeURL, response.QRCodeURL)
	assert.Len(t, response.BackupCodes, 2)
	mockService.AssertExpectations(t)
}

func TestTOTPHandler_Setup_Unauthorized(t *testing.T) {
	handler, _ := setupTOTPTest()
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/2fa/setup", "")
	// No user set in context

	err := handler.Setup(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "unauthorized", response.Error)
}

func TestTOTPHandler_Setup_OAuthUser(t *testing.T) {
	handler, _ := setupTOTPTest()
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/2fa/setup", "")
	user := createOAuthUser()
	c.Set("current_user", user)

	err := handler.Setup(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "oauth_user", response.Error)
	assert.Contains(t, response.Message, "OAuth")
}

func TestTOTPHandler_Setup_AlreadyEnabled(t *testing.T) {
	handler, mockService := setupTOTPTest()
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/2fa/setup", "")
	user := createLocalAuthUser()
	c.Set("current_user", user)

	mockService.On("GenerateSetup", mock.Anything, user.ID, user.Email).Return(nil, services.ErrTOTPAlreadyEnabled)

	err := handler.Setup(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusConflict, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "already_enabled", response.Error)
	mockService.AssertExpectations(t)
}

func TestTOTPHandler_Setup_ServiceError(t *testing.T) {
	handler, mockService := setupTOTPTest()
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/2fa/setup", "")
	user := createLocalAuthUser()
	c.Set("current_user", user)

	mockService.On("GenerateSetup", mock.Anything, user.ID, user.Email).Return(nil, errors.New("internal error"))

	err := handler.Setup(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "server_error", response.Error)
	mockService.AssertExpectations(t)
}

// ==================== Verify Tests ====================

func TestTOTPHandler_Verify_Success(t *testing.T) {
	handler, mockService := setupTOTPTest()
	body := `{"code":"123456"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/2fa/verify", body)
	user := createLocalAuthUser()
	c.Set("current_user", user)

	mockService.On("VerifyAndEnable", mock.Anything, user.ID, "123456").Return(nil)

	err := handler.Verify(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, true, response["enabled"])
	assert.Equal(t, "Two-factor authentication enabled successfully", response["message"])
	mockService.AssertExpectations(t)
}

func TestTOTPHandler_Verify_Unauthorized(t *testing.T) {
	handler, _ := setupTOTPTest()
	body := `{"code":"123456"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/2fa/verify", body)
	// No user set

	err := handler.Verify(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestTOTPHandler_Verify_MissingCode(t *testing.T) {
	handler, _ := setupTOTPTest()
	body := `{"code":""}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/2fa/verify", body)
	user := createLocalAuthUser()
	c.Set("current_user", user)

	err := handler.Verify(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "missing_fields", response.Error)
}

func TestTOTPHandler_Verify_EmptyBody(t *testing.T) {
	handler, _ := setupTOTPTest()
	body := `{}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/2fa/verify", body)
	user := createLocalAuthUser()
	c.Set("current_user", user)

	err := handler.Verify(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "missing_fields", response.Error)
}

func TestTOTPHandler_Verify_InvalidCode(t *testing.T) {
	handler, mockService := setupTOTPTest()
	body := `{"code":"000000"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/2fa/verify", body)
	user := createLocalAuthUser()
	c.Set("current_user", user)

	mockService.On("VerifyAndEnable", mock.Anything, user.ID, "000000").Return(services.ErrTOTPInvalidCode)

	err := handler.Verify(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "invalid_code", response.Error)
	mockService.AssertExpectations(t)
}

func TestTOTPHandler_Verify_NotSetup(t *testing.T) {
	handler, mockService := setupTOTPTest()
	body := `{"code":"123456"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/2fa/verify", body)
	user := createLocalAuthUser()
	c.Set("current_user", user)

	mockService.On("VerifyAndEnable", mock.Anything, user.ID, "123456").Return(services.ErrTOTPNotSetup)

	err := handler.Verify(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "not_setup", response.Error)
	mockService.AssertExpectations(t)
}

func TestTOTPHandler_Verify_AlreadyEnabled(t *testing.T) {
	handler, mockService := setupTOTPTest()
	body := `{"code":"123456"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/2fa/verify", body)
	user := createLocalAuthUser()
	c.Set("current_user", user)

	mockService.On("VerifyAndEnable", mock.Anything, user.ID, "123456").Return(services.ErrTOTPAlreadyEnabled)

	err := handler.Verify(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusConflict, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "already_enabled", response.Error)
	mockService.AssertExpectations(t)
}

func TestTOTPHandler_Verify_ServiceError(t *testing.T) {
	handler, mockService := setupTOTPTest()
	body := `{"code":"123456"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/2fa/verify", body)
	user := createLocalAuthUser()
	c.Set("current_user", user)

	mockService.On("VerifyAndEnable", mock.Anything, user.ID, "123456").Return(errors.New("database error"))

	err := handler.Verify(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "server_error", response.Error)
	mockService.AssertExpectations(t)
}

func TestTOTPHandler_Verify_InvalidRequestBody(t *testing.T) {
	handler, _ := setupTOTPTest()
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/2fa/verify", "invalid-json")
	user := createLocalAuthUser()
	c.Set("current_user", user)

	err := handler.Verify(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "invalid_request", response.Error)
}

// ==================== Challenge Tests ====================

func TestTOTPHandler_Challenge_Success(t *testing.T) {
	handler, mockService := setupTOTPTest()
	initSessionStore()

	body := `{"code":"123456"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/2fa/challenge", body)

	userID := uuid.New()

	// Set up session with pending 2FA user ID
	session, _ := middleware.Store.Get(c.Request(), "session")
	session.Values["2fa_pending_user_id"] = userID.String()
	session.Values["2fa_pending_created_at"] = time.Now().Unix()
	_ = session.Save(c.Request(), c.Response())

	mockService.On("Verify", mock.Anything, userID, "123456").Return(true, nil)

	err := handler.Challenge(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, true, response["authenticated"])
	assert.Equal(t, "Two-factor authentication verified", response["message"])
	mockService.AssertExpectations(t)
}

func TestTOTPHandler_Challenge_SuccessBackupCode(t *testing.T) {
	handler, mockService := setupTOTPTest()
	initSessionStore()

	body := `{"backup_code":"ABCD-1234"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/2fa/challenge", body)

	userID := uuid.New()

	session, _ := middleware.Store.Get(c.Request(), "session")
	session.Values["2fa_pending_user_id"] = userID.String()
	session.Values["2fa_pending_created_at"] = time.Now().Unix()
	_ = session.Save(c.Request(), c.Response())

	mockService.On("VerifyBackupCode", mock.Anything, userID, "ABCD-1234").Return(true, nil)

	err := handler.Challenge(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, true, response["authenticated"])
	mockService.AssertExpectations(t)
}

func TestTOTPHandler_Challenge_MissingCode(t *testing.T) {
	handler, _ := setupTOTPTest()
	initSessionStore()

	body := `{"code":"","backup_code":""}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/2fa/challenge", body)

	userID := uuid.New()

	session, _ := middleware.Store.Get(c.Request(), "session")
	session.Values["2fa_pending_user_id"] = userID.String()
	session.Values["2fa_pending_created_at"] = time.Now().Unix()
	_ = session.Save(c.Request(), c.Response())

	err := handler.Challenge(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "missing_fields", response.Error)
}

func TestTOTPHandler_Challenge_InvalidCode(t *testing.T) {
	handler, mockService := setupTOTPTest()
	initSessionStore()

	body := `{"code":"000000"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/2fa/challenge", body)

	userID := uuid.New()

	session, _ := middleware.Store.Get(c.Request(), "session")
	session.Values["2fa_pending_user_id"] = userID.String()
	session.Values["2fa_pending_created_at"] = time.Now().Unix()
	_ = session.Save(c.Request(), c.Response())

	mockService.On("Verify", mock.Anything, userID, "000000").Return(false, nil)

	err := handler.Challenge(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "invalid_code", response.Error)
	mockService.AssertExpectations(t)
}

func TestTOTPHandler_Challenge_NoSession(t *testing.T) {
	handler, _ := setupTOTPTest()
	// Do NOT initialize session store to trigger "session store not initialized" error
	middleware.Store = nil

	body := `{"code":"123456"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/2fa/challenge", body)

	err := handler.Challenge(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "unauthorized", response.Error)
}

func TestTOTPHandler_Challenge_NoPendingUserID(t *testing.T) {
	handler, _ := setupTOTPTest()
	initSessionStore()

	body := `{"code":"123456"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/2fa/challenge", body)

	// Session exists but has no 2fa_pending_user_id
	session, _ := middleware.Store.Get(c.Request(), "session")
	_ = session.Save(c.Request(), c.Response())

	err := handler.Challenge(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "unauthorized", response.Error)
}

func TestTOTPHandler_Challenge_InvalidPendingUserID(t *testing.T) {
	handler, _ := setupTOTPTest()
	initSessionStore()

	body := `{"code":"123456"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/2fa/challenge", body)

	// Set an invalid UUID in session
	session, _ := middleware.Store.Get(c.Request(), "session")
	session.Values["2fa_pending_user_id"] = "not-a-valid-uuid"
	session.Values["2fa_pending_created_at"] = time.Now().Unix()
	_ = session.Save(c.Request(), c.Response())

	err := handler.Challenge(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "invalid_session", response.Error)
}

func TestTOTPHandler_Challenge_ServiceError(t *testing.T) {
	handler, mockService := setupTOTPTest()
	initSessionStore()

	body := `{"code":"123456"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/2fa/challenge", body)

	userID := uuid.New()

	session, _ := middleware.Store.Get(c.Request(), "session")
	session.Values["2fa_pending_user_id"] = userID.String()
	session.Values["2fa_pending_created_at"] = time.Now().Unix()
	_ = session.Save(c.Request(), c.Response())

	mockService.On("Verify", mock.Anything, userID, "123456").Return(false, errors.New("database error"))

	err := handler.Challenge(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "server_error", response.Error)
	mockService.AssertExpectations(t)
}

func TestTOTPHandler_Challenge_InvalidRequestBody(t *testing.T) {
	handler, _ := setupTOTPTest()
	initSessionStore()

	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/2fa/challenge", "invalid-json")

	userID := uuid.New()

	session, _ := middleware.Store.Get(c.Request(), "session")
	session.Values["2fa_pending_user_id"] = userID.String()
	session.Values["2fa_pending_created_at"] = time.Now().Unix()
	_ = session.Save(c.Request(), c.Response())

	err := handler.Challenge(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "invalid_request", response.Error)
}

func TestTOTPHandler_Challenge_CodeTakesPrecedenceOverBackupCode(t *testing.T) {
	handler, mockService := setupTOTPTest()
	initSessionStore()

	// When both code and backup_code are provided, code should be used
	body := `{"code":"123456","backup_code":"ABCD-1234"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/2fa/challenge", body)

	userID := uuid.New()

	session, _ := middleware.Store.Get(c.Request(), "session")
	session.Values["2fa_pending_user_id"] = userID.String()
	session.Values["2fa_pending_created_at"] = time.Now().Unix()
	_ = session.Save(c.Request(), c.Response())

	// Verify should be called (not VerifyBackupCode) because code takes precedence
	mockService.On("Verify", mock.Anything, userID, "123456").Return(true, nil)

	err := handler.Challenge(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockService.AssertExpectations(t)
	mockService.AssertNotCalled(t, "VerifyBackupCode")
}

// ==================== Disable Tests ====================

func TestTOTPHandler_Disable_Success(t *testing.T) {
	handler, mockService := setupTOTPTest()
	body := `{"code":"123456"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/2fa/disable", body)
	user := createLocalAuthUser()
	c.Set("current_user", user)

	mockService.On("Disable", mock.Anything, user.ID, "123456").Return(nil)

	err := handler.Disable(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, false, response["enabled"])
	assert.Equal(t, "Two-factor authentication disabled", response["message"])
	mockService.AssertExpectations(t)
}

func TestTOTPHandler_Disable_Unauthorized(t *testing.T) {
	handler, _ := setupTOTPTest()
	body := `{"code":"123456"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/2fa/disable", body)
	// No user set

	err := handler.Disable(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestTOTPHandler_Disable_MissingCode(t *testing.T) {
	handler, _ := setupTOTPTest()
	body := `{"code":""}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/2fa/disable", body)
	user := createLocalAuthUser()
	c.Set("current_user", user)

	err := handler.Disable(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "missing_fields", response.Error)
	assert.Contains(t, response.Message, "required")
}

func TestTOTPHandler_Disable_EmptyBody(t *testing.T) {
	handler, _ := setupTOTPTest()
	body := `{}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/2fa/disable", body)
	user := createLocalAuthUser()
	c.Set("current_user", user)

	err := handler.Disable(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "missing_fields", response.Error)
}

func TestTOTPHandler_Disable_NotEnabled(t *testing.T) {
	handler, mockService := setupTOTPTest()
	body := `{"code":"123456"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/2fa/disable", body)
	user := createLocalAuthUser()
	c.Set("current_user", user)

	mockService.On("Disable", mock.Anything, user.ID, "123456").Return(services.ErrTOTPNotEnabled)

	err := handler.Disable(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "not_enabled", response.Error)
	mockService.AssertExpectations(t)
}

func TestTOTPHandler_Disable_InvalidCode(t *testing.T) {
	handler, mockService := setupTOTPTest()
	body := `{"code":"000000"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/2fa/disable", body)
	user := createLocalAuthUser()
	c.Set("current_user", user)

	mockService.On("Disable", mock.Anything, user.ID, "000000").Return(services.ErrTOTPInvalidCode)

	err := handler.Disable(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "invalid_code", response.Error)
	mockService.AssertExpectations(t)
}

func TestTOTPHandler_Disable_ServiceError(t *testing.T) {
	handler, mockService := setupTOTPTest()
	body := `{"code":"123456"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/2fa/disable", body)
	user := createLocalAuthUser()
	c.Set("current_user", user)

	mockService.On("Disable", mock.Anything, user.ID, "123456").Return(errors.New("database error"))

	err := handler.Disable(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "server_error", response.Error)
	mockService.AssertExpectations(t)
}

func TestTOTPHandler_Disable_InvalidRequestBody(t *testing.T) {
	handler, _ := setupTOTPTest()
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/2fa/disable", "invalid-json")
	user := createLocalAuthUser()
	c.Set("current_user", user)

	err := handler.Disable(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "invalid_request", response.Error)
}

// ==================== RegenerateBackupCodes Tests ====================

func TestTOTPHandler_RegenerateBackupCodes_Success(t *testing.T) {
	handler, mockService := setupTOTPTest()
	body := `{"code":"123456"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/2fa/backup-codes", body)
	user := createLocalAuthUser()
	c.Set("current_user", user)

	expectedCodes := []string{"AAAA-1111", "BBBB-2222", "CCCC-3333"}
	mockService.On("RegenerateBackupCodes", mock.Anything, user.ID, "123456").Return(expectedCodes, nil)

	err := handler.RegenerateBackupCodes(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string][]string
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Len(t, response["backup_codes"], 3)
	assert.Equal(t, "AAAA-1111", response["backup_codes"][0])
	assert.Equal(t, "BBBB-2222", response["backup_codes"][1])
	assert.Equal(t, "CCCC-3333", response["backup_codes"][2])
	mockService.AssertExpectations(t)
}

func TestTOTPHandler_RegenerateBackupCodes_Unauthorized(t *testing.T) {
	handler, _ := setupTOTPTest()
	body := `{"code":"123456"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/2fa/backup-codes", body)
	// No user set

	err := handler.RegenerateBackupCodes(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestTOTPHandler_RegenerateBackupCodes_MissingCode(t *testing.T) {
	handler, _ := setupTOTPTest()
	body := `{"code":""}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/2fa/backup-codes", body)
	user := createLocalAuthUser()
	c.Set("current_user", user)

	err := handler.RegenerateBackupCodes(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "missing_fields", response.Error)
}

func TestTOTPHandler_RegenerateBackupCodes_EmptyBody(t *testing.T) {
	handler, _ := setupTOTPTest()
	body := `{}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/2fa/backup-codes", body)
	user := createLocalAuthUser()
	c.Set("current_user", user)

	err := handler.RegenerateBackupCodes(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "missing_fields", response.Error)
}

func TestTOTPHandler_RegenerateBackupCodes_NotEnabled(t *testing.T) {
	handler, mockService := setupTOTPTest()
	body := `{"code":"123456"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/2fa/backup-codes", body)
	user := createLocalAuthUser()
	c.Set("current_user", user)

	mockService.On("RegenerateBackupCodes", mock.Anything, user.ID, "123456").Return(nil, services.ErrTOTPNotEnabled)

	err := handler.RegenerateBackupCodes(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "not_enabled", response.Error)
	mockService.AssertExpectations(t)
}

func TestTOTPHandler_RegenerateBackupCodes_InvalidCode(t *testing.T) {
	handler, mockService := setupTOTPTest()
	body := `{"code":"000000"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/2fa/backup-codes", body)
	user := createLocalAuthUser()
	c.Set("current_user", user)

	mockService.On("RegenerateBackupCodes", mock.Anything, user.ID, "000000").Return(nil, services.ErrTOTPInvalidCode)

	err := handler.RegenerateBackupCodes(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "invalid_code", response.Error)
	mockService.AssertExpectations(t)
}

func TestTOTPHandler_RegenerateBackupCodes_ServiceError(t *testing.T) {
	handler, mockService := setupTOTPTest()
	body := `{"code":"123456"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/2fa/backup-codes", body)
	user := createLocalAuthUser()
	c.Set("current_user", user)

	mockService.On("RegenerateBackupCodes", mock.Anything, user.ID, "123456").Return(nil, errors.New("database error"))

	err := handler.RegenerateBackupCodes(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "server_error", response.Error)
	mockService.AssertExpectations(t)
}

func TestTOTPHandler_RegenerateBackupCodes_InvalidRequestBody(t *testing.T) {
	handler, _ := setupTOTPTest()
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/2fa/backup-codes", "invalid-json")
	user := createLocalAuthUser()
	c.Set("current_user", user)

	err := handler.RegenerateBackupCodes(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "invalid_request", response.Error)
}

// ==================== Status Tests ====================

func TestTOTPHandler_Status_Success_Enabled(t *testing.T) {
	handler, mockService := setupTOTPTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/auth/2fa/status", "")
	user := createLocalAuthUser()
	c.Set("current_user", user)

	mockService.On("IsEnabled", mock.Anything, user.ID).Return(true, nil)

	err := handler.Status(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, true, response["enabled"])
	assert.Equal(t, true, response["is_local_auth"])
	mockService.AssertExpectations(t)
}

func TestTOTPHandler_Status_Success_Disabled(t *testing.T) {
	handler, mockService := setupTOTPTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/auth/2fa/status", "")
	user := createLocalAuthUser()
	c.Set("current_user", user)

	mockService.On("IsEnabled", mock.Anything, user.ID).Return(false, nil)

	err := handler.Status(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, false, response["enabled"])
	assert.Equal(t, true, response["is_local_auth"])
	mockService.AssertExpectations(t)
}

func TestTOTPHandler_Status_Unauthorized(t *testing.T) {
	handler, _ := setupTOTPTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/auth/2fa/status", "")
	// No user set

	err := handler.Status(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "unauthorized", response.Error)
}

func TestTOTPHandler_Status_LocalAuth(t *testing.T) {
	handler, mockService := setupTOTPTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/auth/2fa/status", "")
	user := createLocalAuthUser()
	c.Set("current_user", user)

	mockService.On("IsEnabled", mock.Anything, user.ID).Return(false, nil)

	err := handler.Status(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, true, response["is_local_auth"])
	mockService.AssertExpectations(t)
}

func TestTOTPHandler_Status_OAuthUser(t *testing.T) {
	handler, mockService := setupTOTPTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/auth/2fa/status", "")
	user := createOAuthUser()
	c.Set("current_user", user)

	mockService.On("IsEnabled", mock.Anything, user.ID).Return(false, nil)

	err := handler.Status(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, false, response["enabled"])
	assert.Equal(t, false, response["is_local_auth"])
	mockService.AssertExpectations(t)
}

func TestTOTPHandler_Status_ServiceError(t *testing.T) {
	handler, mockService := setupTOTPTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/auth/2fa/status", "")
	user := createLocalAuthUser()
	c.Set("current_user", user)

	mockService.On("IsEnabled", mock.Anything, user.ID).Return(false, errors.New("database error"))

	err := handler.Status(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "server_error", response.Error)
	mockService.AssertExpectations(t)
}

func TestTOTPHandler_Status_NilUserInContext(t *testing.T) {
	handler, _ := setupTOTPTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/auth/2fa/status", "")
	// Set a non-User type in context to trigger the type assertion failure
	c.Set("current_user", "not-a-user")

	err := handler.Status(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
