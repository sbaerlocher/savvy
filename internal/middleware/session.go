// Package middleware contains Echo middleware for authentication, sessions, and observability.
package middleware

import (
	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
)

// Store is the global session store used for cookie-based sessions
var Store *sessions.CookieStore

// InitSessionStore initializes the session store with the given secret and production flag
func InitSessionStore(secret string, isProduction bool) {
	Store = sessions.NewCookieStore([]byte(secret))
	Store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		Secure:   isProduction, // true in production (HTTPS), false in development
		SameSite: 2,            // Lax
	}
}

// GetSession retrieves the session for the current request
func GetSession(c echo.Context) (*sessions.Session, error) {
	return Store.Get(c.Request(), "session")
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
	if err := oldSession.Save(c.Request(), c.Response()); err != nil {
		return nil, err
	}

	// Create a NEW session with a fresh ID
	// Using Store.New() instead of Store.Get() forces a new session ID
	newSession := sessions.NewSession(Store, "session")
	newSession.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		Secure:   Store.Options.Secure, // Inherit from store config
		SameSite: 2,                    // Lax
	}
	newSession.IsNew = true

	return newSession, nil
}
