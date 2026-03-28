package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// initTestSessionStore initializes a session store for CSRF tests.
func initTestSessionStore() {
	setupTestAuth() // Uses CookieStore for tests (no DB needed)
}

// extractCSRFToken extracts the CSRF token from response cookies.
func extractCSRFToken(rec *httptest.ResponseRecorder) string {
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == CSRFCookieName {
			return cookie.Value
		}
	}
	return ""
}

// extractSessionCookie extracts the session cookie from response.
// Returns the last match, which mirrors browser behavior when multiple
// Set-Cookie headers exist (e.g. after RegenerateSession: delete old + create new).
func extractSessionCookie(rec *httptest.ResponseRecorder) string {
	var result string
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == "session" {
			result = cookie.Name + "=" + cookie.Value
		}
	}
	return result
}

func TestCSRFApiMiddleware_GET_SetsToken(t *testing.T) {
	initTestSessionStore()

	e := echo.New()
	e.Use(CSRFApiMiddleware)
	e.GET("/test", func(c *echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	// Check CSRF cookie is set
	var csrfCookie *http.Cookie
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == CSRFCookieName {
			csrfCookie = cookie
			break
		}
	}

	require.NotNil(t, csrfCookie)
	assert.NotEmpty(t, csrfCookie.Value)
	assert.Equal(t, "/", csrfCookie.Path)
	assert.False(t, csrfCookie.HttpOnly) // Must be readable by JS
	assert.Equal(t, http.SameSiteLaxMode, csrfCookie.SameSite)
	assert.Equal(t, 86400, csrfCookie.MaxAge)
}

func TestCSRFApiMiddleware_POST_ValidToken(t *testing.T) {
	initTestSessionStore()

	e := echo.New()
	e.Use(CSRFApiMiddleware)
	e.GET("/test", func(c *echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})
	e.POST("/test", func(c *echo.Context) error {
		return c.String(http.StatusOK, "Created")
	})

	// First GET request to establish session and get CSRF token
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec1 := httptest.NewRecorder()
	e.ServeHTTP(rec1, req1)

	csrfToken := extractCSRFToken(rec1)
	sessionCookie := extractSessionCookie(rec1)
	require.NotEmpty(t, csrfToken)
	require.NotEmpty(t, sessionCookie)

	// POST request with session cookie and CSRF header
	req2 := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader("data"))
	req2.Header.Set("Cookie", sessionCookie+"; "+CSRFCookieName+"="+csrfToken)
	req2.Header.Set(CSRFHeaderName, csrfToken)
	rec2 := httptest.NewRecorder()

	e.ServeHTTP(rec2, req2)

	assert.Equal(t, http.StatusOK, rec2.Code)
}

func TestCSRFApiMiddleware_POST_MissingToken(t *testing.T) {
	initTestSessionStore()

	e := echo.New()
	e.Use(CSRFApiMiddleware)
	e.POST("/test", func(c *echo.Context) error {
		return c.String(http.StatusOK, "Created")
	})

	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader("data"))
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Contains(t, rec.Body.String(), "csrf_token_mismatch")
	assert.Contains(t, rec.Body.String(), "CSRF token validation failed")
}

func TestCSRFApiMiddleware_POST_InvalidToken(t *testing.T) {
	initTestSessionStore()

	e := echo.New()
	e.Use(CSRFApiMiddleware)
	e.GET("/test", func(c *echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})
	e.POST("/test", func(c *echo.Context) error {
		return c.String(http.StatusOK, "Created")
	})

	// GET to establish session
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec1 := httptest.NewRecorder()
	e.ServeHTTP(rec1, req1)

	sessionCookie := extractSessionCookie(rec1)
	require.NotEmpty(t, sessionCookie)

	// POST with wrong CSRF token (session has a different one)
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader("data"))
	req.Header.Set("Cookie", sessionCookie)
	req.Header.Set(CSRFHeaderName, "invalid-token")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Contains(t, rec.Body.String(), "csrf_token_mismatch")
}

func TestCSRFApiMiddleware_PUT_ValidToken(t *testing.T) {
	initTestSessionStore()

	e := echo.New()
	e.Use(CSRFApiMiddleware)
	e.GET("/test", func(c *echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})
	e.PUT("/test", func(c *echo.Context) error {
		return c.String(http.StatusOK, "Updated")
	})

	// GET to get token
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec1 := httptest.NewRecorder()
	e.ServeHTTP(rec1, req1)

	csrfToken := extractCSRFToken(rec1)
	sessionCookie := extractSessionCookie(rec1)

	// PUT request with token
	req2 := httptest.NewRequest(http.MethodPut, "/test", strings.NewReader("data"))
	req2.Header.Set("Cookie", sessionCookie+"; "+CSRFCookieName+"="+csrfToken)
	req2.Header.Set(CSRFHeaderName, csrfToken)
	rec2 := httptest.NewRecorder()

	e.ServeHTTP(rec2, req2)

	assert.Equal(t, http.StatusOK, rec2.Code)
}

func TestCSRFApiMiddleware_PATCH_ValidToken(t *testing.T) {
	initTestSessionStore()

	e := echo.New()
	e.Use(CSRFApiMiddleware)
	e.GET("/test", func(c *echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})
	e.PATCH("/test", func(c *echo.Context) error {
		return c.String(http.StatusOK, "Patched")
	})

	// GET to get token
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec1 := httptest.NewRecorder()
	e.ServeHTTP(rec1, req1)

	csrfToken := extractCSRFToken(rec1)
	sessionCookie := extractSessionCookie(rec1)

	// PATCH request with token
	req2 := httptest.NewRequest(http.MethodPatch, "/test", strings.NewReader("data"))
	req2.Header.Set("Cookie", sessionCookie+"; "+CSRFCookieName+"="+csrfToken)
	req2.Header.Set(CSRFHeaderName, csrfToken)
	rec2 := httptest.NewRecorder()

	e.ServeHTTP(rec2, req2)

	assert.Equal(t, http.StatusOK, rec2.Code)
}

func TestCSRFApiMiddleware_DELETE_ValidToken(t *testing.T) {
	initTestSessionStore()

	e := echo.New()
	e.Use(CSRFApiMiddleware)
	e.GET("/test", func(c *echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})
	e.DELETE("/test", func(c *echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	// GET to get token
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec1 := httptest.NewRecorder()
	e.ServeHTTP(rec1, req1)

	csrfToken := extractCSRFToken(rec1)
	sessionCookie := extractSessionCookie(rec1)

	// DELETE request with token
	req2 := httptest.NewRequest(http.MethodDelete, "/test", nil)
	req2.Header.Set("Cookie", sessionCookie+"; "+CSRFCookieName+"="+csrfToken)
	req2.Header.Set(CSRFHeaderName, csrfToken)
	rec2 := httptest.NewRecorder()

	e.ServeHTTP(rec2, req2)

	assert.Equal(t, http.StatusNoContent, rec2.Code)
}

func TestCSRFApiMiddleware_DELETE_MissingToken(t *testing.T) {
	initTestSessionStore()

	e := echo.New()
	e.Use(CSRFApiMiddleware)
	e.DELETE("/test", func(c *echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodDelete, "/test", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Contains(t, rec.Body.String(), "csrf_token_mismatch")
}

func TestCSRFApiMiddleware_SecureCookie_HTTPS(t *testing.T) {
	initTestSessionStore()

	e := echo.New()
	e.Use(CSRFApiMiddleware)
	e.GET("/test", func(c *echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	var csrfCookie *http.Cookie
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == CSRFCookieName {
			csrfCookie = cookie
			break
		}
	}

	require.NotNil(t, csrfCookie)
	assert.True(t, csrfCookie.Secure)
}

func TestCSRFApiMiddleware_InsecureCookie_HTTP(t *testing.T) {
	initTestSessionStore()

	e := echo.New()
	e.Use(CSRFApiMiddleware)
	e.GET("/test", func(c *echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	var csrfCookie *http.Cookie
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == CSRFCookieName {
			csrfCookie = cookie
			break
		}
	}

	require.NotNil(t, csrfCookie)
	assert.False(t, csrfCookie.Secure)
}

func TestCSRFApiMiddleware_ReuseExistingToken(t *testing.T) {
	initTestSessionStore()

	e := echo.New()
	e.Use(CSRFApiMiddleware)
	e.GET("/test", func(c *echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	// First request — establishes session with CSRF token
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec1 := httptest.NewRecorder()
	e.ServeHTTP(rec1, req1)

	token1 := extractCSRFToken(rec1)
	sessionCookie := extractSessionCookie(rec1)
	require.NotEmpty(t, token1)
	require.NotEmpty(t, sessionCookie)

	// Second request with same session — token should be reused
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.Header.Set("Cookie", sessionCookie)
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req2)

	token2 := extractCSRFToken(rec2)

	// Token should be reused from session
	assert.Equal(t, token1, token2)
}

func TestCSRFApiMiddleware_CookieAloneNotSufficient(t *testing.T) {
	initTestSessionStore()

	e := echo.New()
	e.Use(CSRFApiMiddleware)
	e.GET("/test", func(c *echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})
	e.POST("/test", func(c *echo.Context) error {
		return c.String(http.StatusOK, "Created")
	})

	// Attacker injects a cookie with their own token (subdomain injection)
	// but doesn't have the session, so the session token will be different
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader("data"))
	req.Header.Set("Cookie", CSRFCookieName+"=attacker-injected-token")
	req.Header.Set(CSRFHeaderName, "attacker-injected-token")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	// Should fail because session has a different (or no) token
	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Contains(t, rec.Body.String(), "csrf_token_mismatch")
}

func TestGenerateCSRFToken(t *testing.T) {
	// Generate multiple tokens
	tokens := make(map[string]bool)
	for i := 0; i < 100; i++ {
		token, err := generateCSRFToken()
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.Greater(t, len(token), 32) // Base64 encoded 32 bytes should be longer
		tokens[token] = true
	}

	// All tokens should be unique
	assert.Equal(t, 100, len(tokens))
}

func TestGetOrGenerateSessionCSRFToken_CreatesNewToken(t *testing.T) {
	initTestSessionStore()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	token, err := getOrGenerateSessionCSRFToken(c)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Greater(t, len(token), 32)

	// Verify token is stored in session
	sess, _ := GetSession(c)
	sessionToken, ok := sess.Values[csrfSessionKey].(string)
	assert.True(t, ok)
	assert.Equal(t, token, sessionToken)
}

func TestCSRFTokenRotatesOnSessionRegeneration(t *testing.T) {
	initTestSessionStore()

	e := echo.New()
	e.Use(CSRFApiMiddleware)
	e.GET("/test", func(c *echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})
	e.POST("/test", func(c *echo.Context) error {
		return c.String(http.StatusOK, "Created")
	})

	// 1. GET to establish session + CSRF token
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec1 := httptest.NewRecorder()
	e.ServeHTTP(rec1, req1)

	oldToken := extractCSRFToken(rec1)
	oldSessionCookie := extractSessionCookie(rec1)
	require.NotEmpty(t, oldToken)
	require.NotEmpty(t, oldSessionCookie)

	// 2. Verify old token works for POST
	reqPost := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader("data"))
	reqPost.Header.Set("Cookie", oldSessionCookie+"; "+CSRFCookieName+"="+oldToken)
	reqPost.Header.Set(CSRFHeaderName, oldToken)
	recPost := httptest.NewRecorder()
	e.ServeHTTP(recPost, reqPost)
	assert.Equal(t, http.StatusOK, recPost.Code, "old token should work before regeneration")

	// 3. Simulate login: regenerate session (this destroys the old session incl. CSRF token)
	reqLogin := httptest.NewRequest(http.MethodGet, "/test", nil)
	reqLogin.Header.Set("Cookie", oldSessionCookie)
	recLogin := httptest.NewRecorder()
	cLogin := e.NewContext(reqLogin, recLogin)

	newSession, err := RegenerateSession(cLogin)
	require.NoError(t, err)
	newSession.Values["user_id"] = "test-user"
	err = SaveSession(cLogin, newSession)
	require.NoError(t, err)

	newSessionCookie := extractSessionCookie(recLogin)
	require.NotEmpty(t, newSessionCookie)

	// 4. GET with new session — should get a NEW CSRF token
	reqGet2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	reqGet2.Header.Set("Cookie", newSessionCookie)
	recGet2 := httptest.NewRecorder()
	e.ServeHTTP(recGet2, reqGet2)

	newToken := extractCSRFToken(recGet2)
	require.NotEmpty(t, newToken)
	assert.NotEqual(t, oldToken, newToken, "CSRF token must change after session regeneration")

	// 5. Old token must be rejected with new session
	reqOld := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader("data"))
	newSessionCookie2 := extractSessionCookie(recGet2)
	reqOld.Header.Set("Cookie", newSessionCookie2+"; "+CSRFCookieName+"="+oldToken)
	reqOld.Header.Set(CSRFHeaderName, oldToken)
	recOld := httptest.NewRecorder()
	e.ServeHTTP(recOld, reqOld)
	assert.Equal(t, http.StatusForbidden, recOld.Code, "old CSRF token must be rejected after session regeneration")

	// 6. New token must work with new session
	reqNew := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader("data"))
	reqNew.Header.Set("Cookie", newSessionCookie2+"; "+CSRFCookieName+"="+newToken)
	reqNew.Header.Set(CSRFHeaderName, newToken)
	recNew := httptest.NewRecorder()
	e.ServeHTTP(recNew, reqNew)
	assert.Equal(t, http.StatusOK, recNew.Code, "new CSRF token must work after session regeneration")
}

func TestGetOrGenerateSessionCSRFToken_ReusesExistingToken(t *testing.T) {
	initTestSessionStore()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Pre-populate session with a token
	sess, _ := GetSession(c)
	existingToken := "existing-session-csrf-token"
	sess.Values[csrfSessionKey] = existingToken
	_ = SaveSession(c, sess)

	token, err := getOrGenerateSessionCSRFToken(c)
	require.NoError(t, err)
	assert.Equal(t, existingToken, token)
}
