package middleware

import (
	"net/http"
	"net/http/httptest"
	"savvy/internal/models"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestRequireImpersonationOrAdmin_AdminUser(t *testing.T) {
	setupTestAuth()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/admin/merchants", nil)
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

	handler := RequireImpersonationOrAdmin(func(c echo.Context) error {
		return c.String(http.StatusOK, "Access granted")
	})

	err := handler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "Access granted", rec.Body.String())
}

func TestRequireImpersonationOrAdmin_RegularUserNotImpersonating(t *testing.T) {
	setupTestAuth()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/admin/merchants", nil)
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

	handler := RequireImpersonationOrAdmin(func(c echo.Context) error {
		return c.String(http.StatusOK, "Access granted")
	})

	err := handler(c)

	// Should return 403 Forbidden
	httpError, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusForbidden, httpError.Code)
	assert.Contains(t, httpError.Message, "Access denied")
}

func TestRequireImpersonationOrAdmin_RegularUserImpersonating(t *testing.T) {
	setupTestAuth()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/admin/merchants", nil)
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

	// Create session with impersonation flag
	session, _ := Store.Get(req, "session")
	session.Values["original_user_id"] = uuid.New().String()
	session.Values["original_user_is_admin"] = true
	_ = session.Save(req, rec)

	req.Header.Set("Cookie", rec.Header().Get("Set-Cookie"))

	handler := RequireImpersonationOrAdmin(func(c echo.Context) error {
		return c.String(http.StatusOK, "Access granted")
	})

	err := handler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "Access granted", rec.Body.String())
}

func TestRequireImpersonationOrAdmin_NoUser(t *testing.T) {
	setupTestAuth()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/admin/merchants", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// No user set in context

	handler := RequireImpersonationOrAdmin(func(c echo.Context) error {
		return c.String(http.StatusOK, "Access granted")
	})

	err := handler(c)

	// Should return 401 Unauthorized
	httpError, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, httpError.Code)
	assert.Contains(t, httpError.Message, "Authentication required")
}

func TestRequireImpersonationOrAdmin_InvalidUserType(t *testing.T) {
	setupTestAuth()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/admin/merchants", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Set invalid user type
	c.Set("current_user", "not-a-user-object")

	handler := RequireImpersonationOrAdmin(func(c echo.Context) error {
		return c.String(http.StatusOK, "Access granted")
	})

	err := handler(c)

	// Should return 401 Unauthorized
	httpError, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, httpError.Code)
}

func TestRequireImpersonationOrAdmin_SessionError(t *testing.T) {
	// Save original store
	originalStore := Store

	// Don't initialize Store to cause session error
	Store = nil

	// Use defer to always restore Store
	defer func() {
		Store = originalStore
	}()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/admin/merchants", nil)
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

	handler := RequireImpersonationOrAdmin(func(c echo.Context) error {
		return c.String(http.StatusOK, "Access granted")
	})

	// With nil Store guard, this returns a 500 error instead of panicking
	err := handler(c)
	httpError, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusInternalServerError, httpError.Code)
}

func TestRequireImpersonationOrAdmin_MultipleScenarios(t *testing.T) {
	tests := []struct {
		name             string
		user             *models.User
		hasImpersonation bool
		expectedCode     int
		expectedError    bool
	}{
		{
			name: "Admin without impersonation",
			user: &models.User{
				ID:        uuid.New(),
				Email:     "admin@example.com",
				FirstName: "Admin",
				LastName:  "User",
				Role:      "admin",
			},
			hasImpersonation: false,
			expectedCode:     http.StatusOK,
			expectedError:    false,
		},
		{
			name: "Admin with impersonation",
			user: &models.User{
				ID:        uuid.New(),
				Email:     "admin@example.com",
				FirstName: "Admin",
				LastName:  "User",
				Role:      "admin",
			},
			hasImpersonation: true,
			expectedCode:     http.StatusOK,
			expectedError:    false,
		},
		{
			name: "Regular user with impersonation",
			user: &models.User{
				ID:        uuid.New(),
				Email:     "user@example.com",
				FirstName: "Regular",
				LastName:  "User",
				Role:      "user",
			},
			hasImpersonation: true,
			expectedCode:     http.StatusOK,
			expectedError:    false,
		},
		{
			name: "Regular user without impersonation",
			user: &models.User{
				ID:        uuid.New(),
				Email:     "user@example.com",
				FirstName: "Regular",
				LastName:  "User",
				Role:      "user",
			},
			hasImpersonation: false,
			expectedCode:     http.StatusForbidden,
			expectedError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTestAuth()

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/admin/merchants", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			c.Set("current_user", tt.user)

			if tt.hasImpersonation {
				session, _ := Store.Get(req, "session")
				session.Values["original_user_id"] = uuid.New().String()
				session.Values["original_user_is_admin"] = true
				_ = session.Save(req, rec)
				req.Header.Set("Cookie", rec.Header().Get("Set-Cookie"))
			}

			handler := RequireImpersonationOrAdmin(func(c echo.Context) error {
				return c.String(http.StatusOK, "Access granted")
			})

			err := handler(c)

			if tt.expectedError {
				httpError, ok := err.(*echo.HTTPError)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, httpError.Code)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCode, rec.Code)
			}
		})
	}
}

func TestRequireImpersonationOrAdmin_NilOriginalUserID(t *testing.T) {
	setupTestAuth()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/admin/merchants", nil)
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

	// Create session but with nil original_user_id
	session, _ := Store.Get(req, "session")
	session.Values["original_user_id"] = nil
	_ = session.Save(req, rec)
	req.Header.Set("Cookie", rec.Header().Get("Set-Cookie"))

	handler := RequireImpersonationOrAdmin(func(c echo.Context) error {
		return c.String(http.StatusOK, "Access granted")
	})

	err := handler(c)

	// Should deny access (nil is not a valid impersonation marker)
	httpError, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusForbidden, httpError.Code)
}
