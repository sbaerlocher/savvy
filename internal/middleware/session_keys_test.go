package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Typed Getter Tests ---

func TestGetSessionUserID_Present(t *testing.T) {
	session := sessions.NewSession(nil, "session")
	session.Values[SessionKeyUserID] = "abc-123"
	assert.Equal(t, "abc-123", GetSessionUserID(session))
}

func TestGetSessionUserID_Missing(t *testing.T) {
	session := sessions.NewSession(nil, "session")
	assert.Equal(t, "", GetSessionUserID(session))
}

func TestGetSessionUserID_WrongType(t *testing.T) {
	session := sessions.NewSession(nil, "session")
	session.Values[SessionKeyUserID] = 42
	assert.Equal(t, "", GetSessionUserID(session))
}

func TestGetSessionCreatedAt_Present(t *testing.T) {
	session := sessions.NewSession(nil, "session")
	session.Values[SessionKeySessionCreatedAt] = int64(1700000000)
	assert.Equal(t, int64(1700000000), GetSessionCreatedAt(session))
}

func TestGetSessionCreatedAt_Missing(t *testing.T) {
	session := sessions.NewSession(nil, "session")
	assert.Equal(t, int64(0), GetSessionCreatedAt(session))
}

func TestGetSession2FAPendingUserID_Present(t *testing.T) {
	session := sessions.NewSession(nil, "session")
	session.Values[SessionKey2FAPendingUserID] = "user-456"
	assert.Equal(t, "user-456", GetSession2FAPendingUserID(session))
}

func TestGetSession2FAPendingUserID_Missing(t *testing.T) {
	session := sessions.NewSession(nil, "session")
	assert.Equal(t, "", GetSession2FAPendingUserID(session))
}

func TestGetSession2FAPendingCreatedAt_Present(t *testing.T) {
	session := sessions.NewSession(nil, "session")
	session.Values[SessionKey2FAPendingCreatedAt] = int64(1700000000)
	assert.Equal(t, int64(1700000000), GetSession2FAPendingCreatedAt(session))
}

func TestGetSession2FAPendingCreatedAt_Missing(t *testing.T) {
	session := sessions.NewSession(nil, "session")
	assert.Equal(t, int64(0), GetSession2FAPendingCreatedAt(session))
}

func TestGetSessionOAuthState_Present(t *testing.T) {
	session := sessions.NewSession(nil, "session")
	session.Values[SessionKeyOAuthState] = "random-state-123"
	assert.Equal(t, "random-state-123", GetSessionOAuthState(session))
}

func TestGetSessionOAuthState_Missing(t *testing.T) {
	session := sessions.NewSession(nil, "session")
	assert.Equal(t, "", GetSessionOAuthState(session))
}

func TestGetSessionOriginalUserID_Present(t *testing.T) {
	session := sessions.NewSession(nil, "session")
	session.Values[SessionKeyOriginalUserID] = "admin-789"
	assert.Equal(t, "admin-789", GetSessionOriginalUserID(session))
}

func TestGetSessionOriginalUserID_Missing(t *testing.T) {
	session := sessions.NewSession(nil, "session")
	assert.Equal(t, "", GetSessionOriginalUserID(session))
}

func TestGetSessionImpersonatedBy_Present(t *testing.T) {
	session := sessions.NewSession(nil, "session")
	session.Values[SessionKeyImpersonatedBy] = "admin-789"
	assert.Equal(t, "admin-789", GetSessionImpersonatedBy(session))
}

func TestGetSessionImpersonatedBy_Missing(t *testing.T) {
	session := sessions.NewSession(nil, "session")
	assert.Equal(t, "", GetSessionImpersonatedBy(session))
}

// --- Composite Helper Tests ---

func TestCreateUserSession_SetsUserIDAndTimestamp(t *testing.T) {
	setupTestAuth()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	before := currentUnix()
	sess, err := CreateUserSession(c, "user-abc")
	after := currentUnix()

	require.NoError(t, err)
	assert.Equal(t, "user-abc", GetSessionUserID(sess))

	ts := GetSessionCreatedAt(sess)
	assert.GreaterOrEqual(t, ts, before)
	assert.LessOrEqual(t, ts, after)

	// Should not have 2FA or impersonation keys
	assert.Equal(t, "", GetSession2FAPendingUserID(sess))
	assert.Equal(t, "", GetSessionOriginalUserID(sess))
}

func TestCreate2FAPendingSession_SetsCorrectKeys(t *testing.T) {
	setupTestAuth()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	before := currentUnix()
	sess, err := Create2FAPendingSession(c, "pending-user")
	after := currentUnix()

	require.NoError(t, err)
	assert.Equal(t, "pending-user", GetSession2FAPendingUserID(sess))

	ts := GetSession2FAPendingCreatedAt(sess)
	assert.GreaterOrEqual(t, ts, before)
	assert.LessOrEqual(t, ts, after)

	// Should NOT have user_id (not yet authenticated)
	assert.Equal(t, "", GetSessionUserID(sess))
}

func TestDestroySession_ClearsUserValues(t *testing.T) {
	setupTestAuth()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Create a session with user data
	sess, err := GetSession(c)
	require.NoError(t, err)
	sess.Values[SessionKeyUserID] = "user-123"
	sess.Values[SessionKeySessionCreatedAt] = int64(1700000000)
	err = sess.Save(req, rec)
	require.NoError(t, err)

	// Destroy it
	err = DestroySession(c)
	require.NoError(t, err)

	// Verify session was marked for deletion
	sess2, err := GetSession(c)
	require.NoError(t, err)
	assert.Equal(t, -1, sess2.Options.MaxAge)
}

func TestDestroySession_NoExistingSession(t *testing.T) {
	setupTestAuth()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Destroy with no prior session should not error
	err := DestroySession(c)
	assert.NoError(t, err)
}

func TestCreateImpersonationSession_SetsAllKeys(t *testing.T) {
	setupTestAuth()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	sess, err := CreateImpersonationSession(c, "target-user", "admin-user")
	require.NoError(t, err)

	assert.Equal(t, "target-user", GetSessionUserID(sess))
	assert.Equal(t, "admin-user", GetSessionOriginalUserID(sess))
	assert.Equal(t, "admin-user", GetSessionImpersonatedBy(sess))
}

func TestStopImpersonationSession_RestoresAdmin(t *testing.T) {
	setupTestAuth()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	sess, err := StopImpersonationSession(c, "admin-user")
	require.NoError(t, err)

	assert.Equal(t, "admin-user", GetSessionUserID(sess))
	assert.Equal(t, "", GetSessionOriginalUserID(sess))
	assert.Equal(t, "", GetSessionImpersonatedBy(sess))
}

// currentUnix returns the current Unix timestamp for time-window assertions.
func currentUnix() int64 {
	return time.Now().Unix()
}
