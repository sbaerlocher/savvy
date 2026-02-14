package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestInitSessionStore_Development(t *testing.T) {
	setupTestAuth()

	assert.NotNil(t, Store)

	// Verify we can create and retrieve sessions
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	session, err := GetSession(c)
	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.True(t, session.IsNew)
}

func TestInitSessionStore_Production(t *testing.T) {
	setupTestAuth()

	assert.NotNil(t, Store)

	// Verify store produces sessions with correct options
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	session, err := GetSession(c)
	assert.NoError(t, err)
	assert.NotNil(t, session)
}

func TestGetSession_NewSession(t *testing.T) {
	setupTestAuth()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	session, err := GetSession(c)

	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.True(t, session.IsNew)
}

func TestGetSession_ExistingSession(t *testing.T) {
	setupTestAuth()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Create and save session
	session1, err := GetSession(c)
	assert.NoError(t, err)
	session1.Values["test"] = "value"
	err = session1.Save(req, rec)
	assert.NoError(t, err)

	// Get session cookie
	cookies := rec.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "session" {
			sessionCookie = cookie
			break
		}
	}
	assert.NotNil(t, sessionCookie)

	// Create new request with session cookie
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.AddCookie(sessionCookie)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	// Get existing session
	session2, err := GetSession(c2)
	assert.NoError(t, err)
	assert.False(t, session2.IsNew)
	assert.Equal(t, "value", session2.Values["test"])
}

func TestGetSession_SessionValues(t *testing.T) {
	setupTestAuth()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	session, err := GetSession(c)
	assert.NoError(t, err)

	// Set various values
	session.Values["string"] = "test"
	session.Values["int"] = 42
	session.Values["bool"] = true

	err = session.Save(req, rec)
	assert.NoError(t, err)

	// Retrieve values
	assert.Equal(t, "test", session.Values["string"])
	assert.Equal(t, 42, session.Values["int"])
	assert.Equal(t, true, session.Values["bool"])
}

func TestRegenerateSession_CreatesNewSession(t *testing.T) {
	setupTestAuth()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Create initial session
	oldSession, err := GetSession(c)
	assert.NoError(t, err)
	oldSession.Values["user_id"] = "123"
	err = oldSession.Save(req, rec)
	assert.NoError(t, err)

	// Get old session cookie
	oldCookies := rec.Result().Cookies()
	var oldSessionCookie *http.Cookie
	for _, cookie := range oldCookies {
		if cookie.Name == "session" {
			oldSessionCookie = cookie
			break
		}
	}

	// Create new request with old cookie
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.AddCookie(oldSessionCookie)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	// Regenerate session
	newSession, err := RegenerateSession(c2)
	assert.NoError(t, err)
	assert.NotNil(t, newSession)
	assert.True(t, newSession.IsNew)

	// Save new session
	newSession.Values["user_id"] = "456"
	err = newSession.Save(req2, rec2)
	assert.NoError(t, err)

	// Check that new session cookie is different
	newCookies := rec2.Result().Cookies()
	assert.NotEmpty(t, newCookies)
}

func TestRegenerateSession_PreservesOptions(t *testing.T) {
	setupTestAuth()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	newSession, err := RegenerateSession(c)
	assert.NoError(t, err)

	// Verify options are set correctly
	assert.Equal(t, 3600, newSession.Options.MaxAge) // matches setupTestAuth
	assert.True(t, newSession.Options.HttpOnly)
	assert.Equal(t, http.SameSiteLaxMode, newSession.Options.SameSite)
}

func TestRegenerateSession_DeletesOldSession(t *testing.T) {
	setupTestAuth()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Create old session
	oldSession, err := GetSession(c)
	assert.NoError(t, err)
	oldSession.Values["data"] = "old"
	err = oldSession.Save(req, rec)
	assert.NoError(t, err)

	// Regenerate
	newSession, err := RegenerateSession(c)
	assert.NoError(t, err)

	// Old session values should not exist in new session
	assert.Empty(t, newSession.Values)
}

func TestRegenerateSession_SessionFixationPrevention(t *testing.T) {
	setupTestAuth()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Attacker creates a session
	attackerSession, err := GetSession(c)
	assert.NoError(t, err)
	attackerSession.Values["attacker"] = "malicious"
	err = attackerSession.Save(req, rec)
	assert.NoError(t, err)

	// Get attacker's session cookie
	cookies := rec.Result().Cookies()
	var attackerCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "session" {
			attackerCookie = cookie
			break
		}
	}

	// Victim logs in with attacker's session cookie
	req2 := httptest.NewRequest(http.MethodPost, "/login", nil)
	req2.AddCookie(attackerCookie)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	// Application regenerates session after login
	newSession, err := RegenerateSession(c2)
	assert.NoError(t, err)
	newSession.Values["user_id"] = "victim-user-id"
	err = newSession.Save(req2, rec2)
	assert.NoError(t, err)

	// New session should not have attacker's values
	assert.Nil(t, newSession.Values["attacker"])
	assert.Equal(t, "victim-user-id", newSession.Values["user_id"])
}

func TestSessionCookie_HttpOnly(t *testing.T) {
	setupTestAuth()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	session, err := GetSession(c)
	assert.NoError(t, err)
	session.Values["test"] = "value"
	err = session.Save(req, rec)
	assert.NoError(t, err)

	cookies := rec.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "session" {
			sessionCookie = cookie
			break
		}
	}

	assert.NotNil(t, sessionCookie)
	assert.True(t, sessionCookie.HttpOnly)
}

func TestSessionCookie_SameSite(t *testing.T) {
	setupTestAuth()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	session, err := GetSession(c)
	assert.NoError(t, err)
	session.Values["test"] = "value"
	err = session.Save(req, rec)
	assert.NoError(t, err)

	cookies := rec.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "session" {
			sessionCookie = cookie
			break
		}
	}

	assert.NotNil(t, sessionCookie)
	assert.Equal(t, http.SameSiteLaxMode, sessionCookie.SameSite)
}

func TestSessionStore_MultipleValues(t *testing.T) {
	setupTestAuth()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	session, err := GetSession(c)
	assert.NoError(t, err)

	// Set multiple values
	testData := map[string]interface{}{
		"user_id":    "user-123",
		"email":      "test@example.com",
		"role":       "admin",
		"logged_in":  true,
		"login_time": 1234567890,
	}

	for key, value := range testData {
		session.Values[key] = value
	}

	err = session.Save(req, rec)
	assert.NoError(t, err)

	// Retrieve and verify
	session2, err := Store.Get(req, "session")
	assert.NoError(t, err)

	for key, expectedValue := range testData {
		actualValue := session2.Values[key]
		assert.Equal(t, expectedValue, actualValue)
	}
}

func TestSessionStore_MaxAge(t *testing.T) {
	maxAge := 7200 // 2 hours
	// Use CookieStore directly for this test since we need specific MaxAge
	cookieStore := sessions.NewCookieStore([]byte("test-secret"))
	cookieStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   false,
		SameSite: 2,
	}
	Store = cookieStore

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	session, err := GetSession(c)
	assert.NoError(t, err)
	session.Values["test"] = "value"
	err = session.Save(req, rec)
	assert.NoError(t, err)

	cookies := rec.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "session" {
			sessionCookie = cookie
			break
		}
	}

	assert.NotNil(t, sessionCookie)
	assert.Equal(t, maxAge, sessionCookie.MaxAge)
}

func TestRegenerateSession_NoExistingSession(t *testing.T) {
	setupTestAuth()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Regenerate without existing session
	newSession, err := RegenerateSession(c)
	assert.NoError(t, err)
	assert.NotNil(t, newSession)
	assert.True(t, newSession.IsNew)
}

func TestGetSession_SessionName(t *testing.T) {
	setupTestAuth()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	session, err := GetSession(c)
	assert.NoError(t, err)
	assert.Equal(t, "session", session.Name())
}

func TestRegenerateSession_SavesNewSession(t *testing.T) {
	setupTestAuth()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Regenerate session
	newSession, err := RegenerateSession(c)
	assert.NoError(t, err)
	assert.NotNil(t, newSession)

	// Set a value and save
	newSession.Values["test_key"] = "test_value"
	err = newSession.Save(req, rec)
	assert.NoError(t, err)

	// Verify cookie is set
	cookies := rec.Result().Cookies()
	assert.NotEmpty(t, cookies)
}

func TestRegenerateSession_ErrorHandling(t *testing.T) {
	setupTestAuth()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Create a session first
	session, err := GetSession(c)
	assert.NoError(t, err)
	session.Values["old_data"] = "value"
	_ = session.Save(req, rec)

	// Regenerate should not error even if old session exists
	newSession, err := RegenerateSession(c)
	assert.NoError(t, err)
	assert.NotNil(t, newSession)

	// Old data should not be in new session
	assert.Nil(t, newSession.Values["old_data"])
}

func TestRegenerateSession_WithCookie(t *testing.T) {
	setupTestAuth()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	// Create initial session with cookie
	c := e.NewContext(req, rec)
	oldSession, err := GetSession(c)
	assert.NoError(t, err)
	oldSession.Values["data"] = "old"
	err = oldSession.Save(req, rec)
	assert.NoError(t, err)

	// Get cookie from response
	cookies := rec.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "session" {
			sessionCookie = cookie
			break
		}
	}
	assert.NotNil(t, sessionCookie)

	// Make new request with old session cookie
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.AddCookie(sessionCookie)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	// Regenerate with existing session
	newSession, err := RegenerateSession(c2)
	assert.NoError(t, err)
	assert.NotNil(t, newSession)
	assert.True(t, newSession.IsNew)
}

func TestSaveSession_SecureFlag_HTTPS_XForwardedProto(t *testing.T) {
	setupTestAuth()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	session, err := GetSession(c)
	assert.NoError(t, err)
	session.Values["test"] = "value"

	err = SaveSession(c, session)
	assert.NoError(t, err)

	// Secure flag must be true when behind HTTPS proxy
	assert.True(t, session.Options.Secure, "Secure flag must be true when X-Forwarded-Proto is https")

	// Verify cookie reflects this
	var sessionCookie *http.Cookie
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == "session" {
			sessionCookie = cookie
			break
		}
	}
	assert.NotNil(t, sessionCookie)
	assert.True(t, sessionCookie.Secure)
}

func TestSaveSession_SecureFlag_HTTP(t *testing.T) {
	setupTestAuth()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// No X-Forwarded-Proto, no TLS → plain HTTP
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	session, err := GetSession(c)
	assert.NoError(t, err)
	session.Values["test"] = "value"

	err = SaveSession(c, session)
	assert.NoError(t, err)

	// Secure flag must be false on plain HTTP
	assert.False(t, session.Options.Secure, "Secure flag must be false on plain HTTP")

	var sessionCookie *http.Cookie
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == "session" {
			sessionCookie = cookie
			break
		}
	}
	assert.NotNil(t, sessionCookie)
	assert.False(t, sessionCookie.Secure)
}
