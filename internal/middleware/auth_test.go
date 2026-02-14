package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"savvy/internal/models"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserService is a mock implementation of UserService
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func setupTestAuth() {
	// Initialize test session store using CookieStore (no DB needed in tests)
	cookieStore := sessions.NewCookieStore([]byte("test-secret-key"))
	cookieStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
		Secure:   false,
		SameSite: 2,
	}
	Store = cookieStore
}

func TestSetCurrentUserWithService_ValidSession(t *testing.T) {
	setupTestAuth()

	// Create mock user service
	mockUserService := new(MockUserService)
	testUserID := uuid.New()
	testUser := &models.User{
		ID:        testUserID,
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		Role:      "user",
	}

	mockUserService.On("GetUserByID", mock.Anything, testUserID).Return(testUser, nil)

	// Create Echo instance
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Create session with user_id
	session, _ := Store.Get(req, "session")
	session.Values["user_id"] = testUserID.String()
	_ = session.Save(req, rec)

	// Update request with session cookie
	req.Header.Set("Cookie", rec.Header().Get("Set-Cookie"))

	// Create middleware
	handler := SetCurrentUserWithService(mockUserService)(func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	// Execute
	err := handler(c)
	assert.NoError(t, err)

	// Verify user was set
	currentUser := c.Get("current_user")
	assert.NotNil(t, currentUser)
	assert.Equal(t, testUser, currentUser)

	// Verify user in context
	ctxUser := c.Request().Context().Value(UserContextKey)
	assert.Equal(t, testUser, ctxUser)

	mockUserService.AssertExpectations(t)
}

func TestSetCurrentUserWithService_NoSession(t *testing.T) {
	setupTestAuth()

	mockUserService := new(MockUserService)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := SetCurrentUserWithService(mockUserService)(func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	err := handler(c)
	assert.NoError(t, err)

	// User should not be set
	currentUser := c.Get("current_user")
	assert.Nil(t, currentUser)

	// UserService should not be called
	mockUserService.AssertNotCalled(t, "GetUserByID")
}

func TestSetCurrentUserWithService_InvalidUserID(t *testing.T) {
	setupTestAuth()

	mockUserService := new(MockUserService)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Create session with invalid user_id
	session, _ := Store.Get(req, "session")
	session.Values["user_id"] = "not-a-uuid"
	_ = session.Save(req, rec)

	req.Header.Set("Cookie", rec.Header().Get("Set-Cookie"))

	handler := SetCurrentUserWithService(mockUserService)(func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	err := handler(c)
	assert.NoError(t, err)

	// User should not be set
	currentUser := c.Get("current_user")
	assert.Nil(t, currentUser)

	mockUserService.AssertNotCalled(t, "GetUserByID")
}

func TestSetCurrentUserWithService_UserNotFound(t *testing.T) {
	setupTestAuth()

	mockUserService := new(MockUserService)
	testUserID := uuid.New()

	mockUserService.On("GetUserByID", mock.Anything, testUserID).Return(nil, assert.AnError)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	session, _ := Store.Get(req, "session")
	session.Values["user_id"] = testUserID.String()
	_ = session.Save(req, rec)

	req.Header.Set("Cookie", rec.Header().Get("Set-Cookie"))

	handler := SetCurrentUserWithService(mockUserService)(func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	err := handler(c)
	assert.NoError(t, err)

	// User should not be set when GetUserByID fails
	currentUser := c.Get("current_user")
	assert.Nil(t, currentUser)

	mockUserService.AssertExpectations(t)
}

func TestSetCurrentUserWithService_Impersonation(t *testing.T) {
	setupTestAuth()

	mockUserService := new(MockUserService)
	testUserID := uuid.New()
	adminUserID := uuid.New()

	testUser := &models.User{
		ID:        testUserID,
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		Role:      "user",
	}

	mockUserService.On("GetUserByID", mock.Anything, testUserID).Return(testUser, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Create session with impersonation
	session, _ := Store.Get(req, "session")
	session.Values["user_id"] = testUserID.String()
	session.Values["impersonated_by"] = adminUserID.String()
	_ = session.Save(req, rec)

	req.Header.Set("Cookie", rec.Header().Get("Set-Cookie"))

	handler := SetCurrentUserWithService(mockUserService)(func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	err := handler(c)
	assert.NoError(t, err)

	// Verify impersonation flag is set
	isImpersonating := c.Get("is_impersonating")
	assert.Equal(t, true, isImpersonating)

	mockUserService.AssertExpectations(t)
}

func TestRequireAuth_Authenticated(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Set current user
	user := &models.User{
		ID:        uuid.New(),
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		Role:      "user",
	}
	c.Set("current_user", user)

	handler := RequireAuth(func(c echo.Context) error {
		return c.String(http.StatusOK, "Protected content")
	})

	err := handler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequireAuth_NotAuthenticatedHTMLRoute(t *testing.T) {
	setupTestAuth()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := RequireAuth(func(c echo.Context) error {
		return c.String(http.StatusOK, "Protected content")
	})

	err := handler(c)

	// Should redirect to login
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/login", rec.Header().Get("Location"))
}

func TestRequireAuth_NotAuthenticatedAPIRoute(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := RequireAuth(func(c echo.Context) error {
		return c.String(http.StatusOK, "Protected content")
	})

	err := handler(c)

	// Should return JSON 401
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "unauthorized")
	assert.Contains(t, rec.Body.String(), "Not authenticated")
}

func TestRequireAdmin_AdminUser(t *testing.T) {
	setupTestAuth()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Set admin user
	admin := &models.User{
		ID:        uuid.New(),
		Email:     "admin@example.com",
		FirstName: "Admin",
		LastName:  "User",
		Role:      "admin",
	}
	c.Set("current_user", admin)

	handler := RequireAdmin(func(c echo.Context) error {
		return c.String(http.StatusOK, "Admin content")
	})

	err := handler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequireAdmin_RegularUser(t *testing.T) {
	setupTestAuth()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Set regular user
	user := &models.User{
		ID:        uuid.New(),
		Email:     "user@example.com",
		FirstName: "Regular",
		LastName:  "User",
		Role:      "user",
	}
	c.Set("current_user", user)

	handler := RequireAdmin(func(c echo.Context) error {
		return c.String(http.StatusOK, "Admin content")
	})

	err := handler(c)

	// Should redirect to home
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/", rec.Header().Get("Location"))
}

func TestRequireAdmin_NotAuthenticated(t *testing.T) {
	setupTestAuth()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := RequireAdmin(func(c echo.Context) error {
		return c.String(http.StatusOK, "Admin content")
	})

	err := handler(c)

	// Should redirect to login
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/login", rec.Header().Get("Location"))
}

func TestRequireAdmin_InvalidUserType(t *testing.T) {
	setupTestAuth()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Set invalid user type
	c.Set("current_user", "not-a-user")

	handler := RequireAdmin(func(c echo.Context) error {
		return c.String(http.StatusOK, "Admin content")
	})

	err := handler(c)

	// Should redirect to home
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/", rec.Header().Get("Location"))
}

func TestRequireAuth_SessionSaveError(t *testing.T) {
	setupTestAuth()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// No user set, should redirect
	handler := RequireAuth(func(c echo.Context) error {
		return c.String(http.StatusOK, "Protected")
	})

	err := handler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
}

func TestRequireAuth_APIRouteVariations(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		isAPI    bool
		wantCode int
	}{
		{
			name:     "API route /api/v1/cards",
			path:     "/api/v1/cards",
			isAPI:    true,
			wantCode: http.StatusUnauthorized,
		},
		{
			name:     "API route /api/users",
			path:     "/api/users",
			isAPI:    true,
			wantCode: http.StatusUnauthorized,
		},
		{
			name:     "HTML route /cards",
			path:     "/cards",
			isAPI:    false,
			wantCode: http.StatusSeeOther,
		},
		{
			name:     "HTML route /about",
			path:     "/about",
			isAPI:    false,
			wantCode: http.StatusSeeOther,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTestAuth()

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			handler := RequireAuth(func(c echo.Context) error {
				return c.String(http.StatusOK, "OK")
			})

			err := handler(c)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantCode, rec.Code)
		})
	}
}
