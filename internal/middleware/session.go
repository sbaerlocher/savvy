// Package middleware contains Echo middleware for authentication, sessions, and observability.
package middleware

import (
	"errors"
	"savvy/internal/repository"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v5"
)

// Store is the global session store (PGStore backed by PostgreSQL)
var Store sessions.Store

// InitSessionStore initializes the PostgreSQL-backed session store.
func InitSessionStore(repo repository.SessionRepository, maxAge int) {
	Store = NewPGStore(repo, maxAge)
}

// GetSession retrieves the session for the current request.
// Returns an error if the session store has not been initialized.
func GetSession(c echo.Context) (*sessions.Session, error) {
	if Store == nil {
		return nil, errors.New("session store not initialized")
	}
	return Store.Get(c.Request(), "session")
}

// SaveSession saves a session with the Secure flag set dynamically based on actual HTTPS status
// This allows the same code to work in both HTTP (E2E/dev) and HTTPS (production) environments
func SaveSession(c echo.Context, session *sessions.Session) error {
	// Set Secure flag based on actual connection (same logic as CSRF middleware)
	// - True if direct TLS connection OR behind HTTPS reverse proxy
	// - False for plain HTTP (localhost, E2E tests)
	session.Options.Secure = c.Request().TLS != nil || c.Request().Header.Get("X-Forwarded-Proto") == "https"
	return session.Save(c.Request(), c.Response())
}

// pgstoreInternalKeys is the set of internal session keys that must be
// preserved across session clearing. Using an explicit allowlist instead
// of a fragile "__" prefix convention.
var pgstoreInternalKeys = map[string]bool{
	sessionTokenHashKey:  true,
	sessionDBIDKey:       true,
	sessionLastActiveKey: true,
	sessionRawTokenKey:   true,
}

// ClearSessionUserValues removes all user-facing session values while
// preserving internal PGStore metadata (token hash, DB ID, last active).
// Use this instead of manually deleting all session.Values entries.
func ClearSessionUserValues(session *sessions.Session) {
	for key := range session.Values {
		if keyStr, ok := key.(string); ok && pgstoreInternalKeys[keyStr] {
			continue // preserve internal PGStore keys
		}
		delete(session.Values, key)
	}
}

// RegenerateSession invalidates the old session and creates a new one with a new ID.
// This prevents session fixation attacks by ensuring a fresh session ID after authentication.
// Returns the new session.
func RegenerateSession(c echo.Context) (*sessions.Session, error) {
	// Get the old session
	oldSession, err := Store.Get(c.Request(), "session")
	if err != nil {
		return nil, err
	}

	// Mark old session for deletion (MaxAge = -1 deletes the cookie)
	oldSession.Options.MaxAge = -1
	if err := SaveSession(c, oldSession); err != nil {
		return nil, err
	}

	// Create a NEW session with a fresh ID
	// Using Store.New() instead of Store.Get() forces a new session ID
	newSession := sessions.NewSession(Store, "session")

	// Get MaxAge from store configuration
	maxAge := 604800 // default 7 days
	if pgStore, ok := Store.(*PGStore); ok {
		maxAge = pgStore.GetMaxAge()
	} else if cookieStore, ok := Store.(*sessions.CookieStore); ok {
		maxAge = cookieStore.Options.MaxAge
	}

	newSession.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   false, // Will be set dynamically in SaveSession
		SameSite: 2,     // Lax
	}
	newSession.IsNew = true

	return newSession, nil
}
