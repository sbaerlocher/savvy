// Package middleware contains Echo middleware for authentication, sessions, and observability.
package middleware

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"savvy/internal/models"
	"savvy/internal/repository"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
)

const (
	// maxSessionDataSize limits the size of session data to decode (1 MB).
	// Defense-in-depth: prevents excessive memory allocation from corrupted data.
	maxSessionDataSize = 1 << 20
)

const (
	// sessionTokenLength is the number of random bytes for session tokens (64 bytes = 512 bits).
	sessionTokenLength = 64

	// sessionTokenHashKey is a private key used to stash the token hash in session values.
	// This is stripped before persisting to DB.
	sessionTokenHashKey = "__pgstore_token_hash" //nolint:gosec // G101: not a credential, just a map key name

	// sessionDBIDKey is a private key used to stash the DB session ID in session values.
	sessionDBIDKey = "__pgstore_db_id"

	// sessionLastActiveKey is a private key used to stash the last active timestamp in session values.
	// This enables throttled LastActiveAt updates in the auth middleware.
	sessionLastActiveKey = "__pgstore_last_active_at"

	// sessionRawTokenKey stashes the raw (unhashed) token in memory so that
	// subsequent Save calls on the same request can refresh the cookie with the
	// correct value. Without this, the sliding-session refresh reads the stale
	// cookie from the original request, which breaks after RegenerateSession.
	sessionRawTokenKey = "__pgstore_raw_token"

	// lastActiveThrottle is the minimum interval between LastActiveAt updates.
	lastActiveThrottle = 60 * time.Second
)

// PGStore implements gorilla/sessions.Store using PostgreSQL.
type PGStore struct {
	repo    repository.SessionRepository
	maxAge  int // session max age in seconds
	Options *sessions.Options
}

// NewPGStore creates a new PostgreSQL session store.
func NewPGStore(repo repository.SessionRepository, maxAge int) *PGStore {
	return &PGStore{
		repo:   repo,
		maxAge: maxAge,
		Options: &sessions.Options{
			Path:     "/",
			MaxAge:   maxAge,
			HttpOnly: true,
			Secure:   false, // Set dynamically by SaveSession based on TLS/X-Forwarded-Proto
			SameSite: http.SameSiteLaxMode,
		},
	}
}

// Get retrieves a session by name from the request cookie.
// If the cookie doesn't exist or the session is not found in DB, returns a new empty session.
func (s *PGStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	return sessions.GetRegistry(r).Get(s, name)
}

// New creates a new session or loads an existing one from the database.
func (s *PGStore) New(r *http.Request, name string) (*sessions.Session, error) {
	session := sessions.NewSession(s, name)
	opts := *s.Options
	session.Options = &opts
	session.IsNew = true

	// Try to load from cookie
	cookie, err := r.Cookie(name)
	if err != nil || cookie.Value == "" {
		return session, nil
	}

	// Hash the cookie value to look up in DB
	tokenHash := hashToken(cookie.Value)

	// Look up session in DB
	ctx := r.Context()
	dbSession, err := s.repo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		// Session not found or expired — return a new empty session
		return session, nil
	}

	// Deserialize session data
	if err := gobDecode(dbSession.Data, &session.Values); err != nil {
		slog.Error("Failed to decode session data", "error", err) //nolint:gosec // session_id is internal UUID
		return session, nil
	}

	// Stash internal metadata for Save
	session.Values[sessionTokenHashKey] = tokenHash
	session.Values[sessionDBIDKey] = dbSession.ID.String()
	session.Values[sessionLastActiveKey] = dbSession.LastActiveAt.Unix()
	session.IsNew = false

	return session, nil
}

// Save persists a session to the database and sets the session cookie.
func (s *PGStore) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	ctx := r.Context()

	// Handle session deletion (MaxAge < 0)
	if session.Options.MaxAge < 0 {
		if tokenHash, ok := session.Values[sessionTokenHashKey].(string); ok && tokenHash != "" {
			if err := s.repo.DeleteByTokenHash(ctx, tokenHash); err != nil {
				slog.Error("Failed to delete session from DB", "error", err)
			}
		}
		// Set cookie to expire
		setCookie(w, session.Name(), "", session.Options)
		return nil
	}

	// Extract IP and User-Agent from request
	ipAddress := extractIP(r)
	userAgent := r.UserAgent()

	// Extract user_id from session values (if present and not a 2FA pending session)
	var userID *uuid.UUID
	if userIDStr, ok := session.Values[SessionKeyUserID].(string); ok && userIDStr != "" {
		if id, err := uuid.Parse(userIDStr); err == nil {
			userID = &id
		}
	}

	// Prepare session data for serialization (strip internal keys)
	valuesToEncode := make(map[interface{}]interface{})
	for k, v := range session.Values {
		if k == sessionTokenHashKey || k == sessionDBIDKey || k == sessionLastActiveKey || k == sessionRawTokenKey {
			continue
		}
		valuesToEncode[k] = v
	}

	data, err := gobEncode(valuesToEncode)
	if err != nil {
		return fmt.Errorf("encode session data: %w", err)
	}

	now := time.Now()

	if session.IsNew {
		// Generate new session token
		rawToken, err := generateToken()
		if err != nil {
			return fmt.Errorf("generate session token: %w", err)
		}
		tokenHash := hashToken(rawToken)

		dbSession := &models.Session{
			UserID:       userID,
			TokenHash:    tokenHash,
			Data:         data,
			IPAddress:    ipAddress,
			UserAgent:    userAgent,
			CreatedAt:    now,
			LastActiveAt: now,
			ExpiresAt:    now.Add(time.Duration(s.maxAge) * time.Second),
		}

		if err := s.repo.Create(ctx, dbSession); err != nil {
			return fmt.Errorf("create session: %w", err)
		}

		// Stash metadata for future Save calls
		session.Values[sessionTokenHashKey] = tokenHash
		session.Values[sessionDBIDKey] = dbSession.ID.String()
		session.Values[sessionRawTokenKey] = rawToken
		session.IsNew = false

		// Set cookie with raw token
		setCookie(w, session.Name(), rawToken, session.Options)
		return nil
	}

	// Update existing session
	tokenHash, _ := session.Values[sessionTokenHashKey].(string)
	dbIDStr, _ := session.Values[sessionDBIDKey].(string)
	if tokenHash == "" || dbIDStr == "" {
		return fmt.Errorf("session missing internal metadata")
	}

	dbID, err := uuid.Parse(dbIDStr)
	if err != nil {
		return fmt.Errorf("invalid session DB ID: %w", err)
	}

	dbSession := &models.Session{
		ID:           dbID,
		UserID:       userID,
		TokenHash:    tokenHash,
		Data:         data,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		LastActiveAt: now,
		ExpiresAt:    now.Add(time.Duration(s.maxAge) * time.Second),
	}

	if err := s.repo.Update(ctx, dbSession); err != nil {
		return fmt.Errorf("update session: %w", err)
	}

	// Refresh cookie MaxAge to implement sliding sessions.
	// Without this, the browser deletes the cookie after the original MaxAge
	// even though the DB session keeps extending via ExpiresAt updates.
	//
	// Prefer the in-memory raw token (set during session creation in the same request)
	// over the request cookie. After RegenerateSession, the request cookie still holds
	// the old (deleted) token, so reading r.Cookie() would overwrite the new cookie
	// with a stale value, causing a 401 on the next request.
	if rawToken, ok := session.Values[sessionRawTokenKey].(string); ok && rawToken != "" {
		setCookie(w, session.Name(), rawToken, session.Options)
	} else if cookie, err := r.Cookie(session.Name()); err == nil && cookie.Value != "" {
		setCookie(w, session.Name(), cookie.Value, session.Options)
	} else {
		slog.Debug("sliding session: no cookie present on request, skipping refresh", "session_name", session.Name())
	}

	return nil
}

// GetMaxAge returns the configured max age for the store.
func (s *PGStore) GetMaxAge() int {
	return s.maxAge
}

// generateToken creates a cryptographically secure random token.
func generateToken() (string, error) {
	b := make([]byte, sessionTokenLength)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("read random bytes: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// hashToken creates a SHA-256 hash of the raw token.
func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// gobEncode serializes session values using gob encoding.
func gobEncode(values map[interface{}]interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(values); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// gobDecode deserializes session values from gob encoding.
// Uses io.LimitReader as defense-in-depth against oversized session data.
func gobDecode(data []byte, dst *map[interface{}]interface{}) error {
	reader := io.LimitReader(bytes.NewReader(data), maxSessionDataSize)
	dec := gob.NewDecoder(reader)
	return dec.Decode(dst)
}

// extractIP extracts the client IP address from the request.
// It checks X-Forwarded-For and X-Real-IP headers for reverse proxy setups.
//
// TRUST MODEL: This function trusts proxy headers unconditionally.
// This is acceptable because:
//   - Savvy runs behind Traefik which sets X-Forwarded-For / X-Real-IP
//   - The IP is only used for session metadata display, not for security
//     decisions (rate limiting uses Echo's c.RealIP() via middleware)
//   - Direct exposure without a reverse proxy is not a supported deployment
func extractIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs: client, proxy1, proxy2
		// Take the first (leftmost) entry — the original client IP as set by Traefik
		for i := 0; i < len(xff); i++ {
			if xff[i] == ',' {
				return xff[:i]
			}
		}
		return xff
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	// Fallback to remote address (strip port)
	addr := r.RemoteAddr
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			return addr[:i]
		}
	}
	return addr
}

// setCookie sets the session cookie on the response.
// The Secure flag is derived from session.Options (set dynamically by SaveSession).
func setCookie(w http.ResponseWriter, name, value string, options *sessions.Options) {
	cookie := sessions.NewCookie(name, value, options)
	http.SetCookie(w, cookie)
}

// GetCurrentSessionTokenHash extracts the current session's token hash from the echo context.
func GetCurrentSessionTokenHash(c interface{ Request() *http.Request }) string {
	sess, err := Store.Get(c.Request(), "session")
	if err != nil {
		return ""
	}
	if hash, ok := sess.Values[sessionTokenHashKey].(string); ok {
		return hash
	}
	return ""
}
