// Package middleware contains Echo middleware for authentication, sessions, and observability.
package middleware

import (
	"fmt"
	"time"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
)

// Session key constants for user-facing session values.
// These replace magic strings scattered across handlers.
//
// Internal PGStore keys (__pgstore_*) are defined in pgstore.go
// and must never be used outside the middleware package.
const (
	// SessionKeyUserID stores the authenticated user's UUID (string).
	SessionKeyUserID = "user_id"

	// SessionKeySessionCreatedAt stores the session creation time (int64, Unix timestamp).
	// Used for stale session detection after password changes.
	SessionKeySessionCreatedAt = "session_created_at"

	// SessionKeyOAuthState stores the OAuth CSRF state parameter (string).
	SessionKeyOAuthState = "oauth_state"

	// SessionKeyOAuthLogin indicates this session was created via OAuth (bool).
	SessionKeyOAuthLogin = "oauth_login"

	// SessionKey2FAPendingUserID stores the user ID during 2FA challenge (string UUID).
	SessionKey2FAPendingUserID = "2fa_pending_user_id"

	// SessionKey2FAPendingCreatedAt stores when the 2FA challenge started (int64, Unix).
	SessionKey2FAPendingCreatedAt = "2fa_pending_created_at"

	// SessionKeyOriginalUserID stores the admin's real ID during impersonation (string UUID).
	SessionKeyOriginalUserID = "original_user_id"

	// SessionKeyImpersonatedBy stores who initiated the impersonation (string UUID).
	SessionKeyImpersonatedBy = "impersonated_by"
)

// --- Typed Getters ---

// GetSessionUserID returns the user_id from the session, or empty string if not set.
func GetSessionUserID(session *sessions.Session) string {
	if s, ok := session.Values[SessionKeyUserID].(string); ok {
		return s
	}
	return ""
}

// GetSessionCreatedAt returns the session creation Unix timestamp, or 0 if not set.
func GetSessionCreatedAt(session *sessions.Session) int64 {
	if ts, ok := session.Values[SessionKeySessionCreatedAt].(int64); ok {
		return ts
	}
	return 0
}

// GetSession2FAPendingUserID returns the 2FA pending user ID, or empty string if not set.
func GetSession2FAPendingUserID(session *sessions.Session) string {
	if s, ok := session.Values[SessionKey2FAPendingUserID].(string); ok {
		return s
	}
	return ""
}

// GetSession2FAPendingCreatedAt returns the 2FA pending creation Unix timestamp, or 0 if not set.
func GetSession2FAPendingCreatedAt(session *sessions.Session) int64 {
	if ts, ok := session.Values[SessionKey2FAPendingCreatedAt].(int64); ok {
		return ts
	}
	return 0
}

// GetSessionOAuthState returns the OAuth state string, or empty string if not set.
func GetSessionOAuthState(session *sessions.Session) string {
	if s, ok := session.Values[SessionKeyOAuthState].(string); ok {
		return s
	}
	return ""
}

// GetSessionOriginalUserID returns the original admin user ID during impersonation.
func GetSessionOriginalUserID(session *sessions.Session) string {
	if s, ok := session.Values[SessionKeyOriginalUserID].(string); ok {
		return s
	}
	return ""
}

// GetSessionImpersonatedBy returns who initiated the impersonation.
func GetSessionImpersonatedBy(session *sessions.Session) string {
	if s, ok := session.Values[SessionKeyImpersonatedBy].(string); ok {
		return s
	}
	return ""
}

// --- Composite Helpers ---

// CreateUserSession regenerates the session (for session fixation prevention),
// sets user_id and session_created_at, and saves. This is the single place
// where an authenticated user session is established.
//
// Used after: successful login, successful registration, successful 2FA challenge,
// successful OAuth callback.
func CreateUserSession(c echo.Context, userID string) (*sessions.Session, error) {
	newSession, err := RegenerateSession(c)
	if err != nil {
		return nil, fmt.Errorf("regenerate session: %w", err)
	}
	newSession.Values[SessionKeyUserID] = userID
	newSession.Values[SessionKeySessionCreatedAt] = time.Now().Unix()
	if err := SaveSession(c, newSession); err != nil {
		return nil, fmt.Errorf("save session: %w", err)
	}
	return newSession, nil
}

// Create2FAPendingSession regenerates the session and sets the 2FA pending state.
// Used when login succeeds but 2FA verification is still required.
func Create2FAPendingSession(c echo.Context, userID string) (*sessions.Session, error) {
	newSession, err := RegenerateSession(c)
	if err != nil {
		return nil, fmt.Errorf("regenerate session: %w", err)
	}
	newSession.Values[SessionKey2FAPendingUserID] = userID
	newSession.Values[SessionKey2FAPendingCreatedAt] = time.Now().Unix()
	if err := SaveSession(c, newSession); err != nil {
		return nil, fmt.Errorf("save session: %w", err)
	}
	return newSession, nil
}

// DestroySession safely invalidates the current session.
// It preserves PGStore internal metadata so the DB row can be properly deleted,
// then sets MaxAge = -1 to expire the cookie and trigger DB deletion.
//
// This replaces the unsafe pattern: session.Values = make(map[interface{}]interface{})
func DestroySession(c echo.Context) error {
	session, err := GetSession(c)
	if err != nil {
		return nil // No session to destroy
	}
	ClearSessionUserValues(session)
	session.Options.MaxAge = -1
	return SaveSession(c, session)
}

// CreateImpersonationSession regenerates the session and sets up impersonation state.
// The target user becomes the active user, the admin's ID is preserved for restoration.
func CreateImpersonationSession(c echo.Context, targetUserID, adminUserID string) (*sessions.Session, error) {
	newSession, err := RegenerateSession(c)
	if err != nil {
		return nil, fmt.Errorf("regenerate session: %w", err)
	}
	newSession.Values[SessionKeyUserID] = targetUserID
	newSession.Values[SessionKeyOriginalUserID] = adminUserID
	newSession.Values[SessionKeyImpersonatedBy] = adminUserID
	if err := SaveSession(c, newSession); err != nil {
		return nil, fmt.Errorf("save session: %w", err)
	}
	return newSession, nil
}

// StopImpersonationSession regenerates the session and restores the admin's own session.
func StopImpersonationSession(c echo.Context, adminUserID string) (*sessions.Session, error) {
	newSession, err := RegenerateSession(c)
	if err != nil {
		return nil, fmt.Errorf("regenerate session: %w", err)
	}
	newSession.Values[SessionKeyUserID] = adminUserID
	if err := SaveSession(c, newSession); err != nil {
		return nil, fmt.Errorf("save session: %w", err)
	}
	return newSession, nil
}
