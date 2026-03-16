package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"savvy/internal/models"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Mock SessionRepository ---

// mockSessionRepo implements repository.SessionRepository for testing PGStore
// without a real database.
type mockSessionRepo struct {
	// sessions stores sessions by token hash for FindByTokenHash lookups.
	sessions map[string]*models.Session

	// created collects sessions passed to Create for verification.
	created []*models.Session

	// updated collects sessions passed to Update for verification.
	updated []*models.Session

	// deletedTokenHashes collects token hashes passed to DeleteByTokenHash.
	deletedTokenHashes []string

	// findErr forces FindByTokenHash to return this error.
	findErr error

	// createErr forces Create to return this error.
	createErr error

	// updateErr forces Update to return this error.
	updateErr error

	// deleteErr forces DeleteByTokenHash to return this error.
	deleteErr error
}

func newMockSessionRepo() *mockSessionRepo {
	return &mockSessionRepo{
		sessions: make(map[string]*models.Session),
	}
}

func (m *mockSessionRepo) FindByTokenHash(_ context.Context, tokenHash string) (*models.Session, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	sess, ok := m.sessions[tokenHash]
	if !ok {
		return nil, errors.New("session not found")
	}
	return sess, nil
}

func (m *mockSessionRepo) Create(_ context.Context, session *models.Session) error {
	if m.createErr != nil {
		return m.createErr
	}
	if session.ID == uuid.Nil {
		session.ID = uuid.New()
	}
	m.created = append(m.created, session)
	m.sessions[session.TokenHash] = session
	return nil
}

func (m *mockSessionRepo) Update(_ context.Context, session *models.Session) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.updated = append(m.updated, session)
	m.sessions[session.TokenHash] = session
	return nil
}

func (m *mockSessionRepo) DeleteByTokenHash(_ context.Context, tokenHash string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	m.deletedTokenHashes = append(m.deletedTokenHashes, tokenHash)
	delete(m.sessions, tokenHash)
	return nil
}

// Unused session management methods — required by the interface but not exercised by PGStore.

func (m *mockSessionRepo) GetByUserID(_ context.Context, _ uuid.UUID) ([]models.Session, error) {
	return nil, nil
}

func (m *mockSessionRepo) DeleteByID(_ context.Context, _, _ uuid.UUID) error {
	return nil
}

func (m *mockSessionRepo) DeleteAllByUserIDExcept(_ context.Context, _ uuid.UUID, _ string) (int64, error) {
	return 0, nil
}

func (m *mockSessionRepo) DeleteAllByUserID(_ context.Context, _ uuid.UUID) (int64, error) {
	return 0, nil
}

func (m *mockSessionRepo) DeleteExpired(_ context.Context) (int64, error) {
	return 0, nil
}

func (m *mockSessionRepo) CountActive(_ context.Context) (int64, error) {
	return 0, nil
}

// --- Helper: create a PGStore for testing ---

func newTestPGStore() (*PGStore, *mockSessionRepo) {
	repo := newMockSessionRepo()
	store := NewPGStore(repo, 3600) // 1 hour max age
	return store, repo
}

// --- NewPGStore Constructor Tests ---

func TestNewPGStore_DefaultOptions(t *testing.T) {
	repo := newMockSessionRepo()
	store := NewPGStore(repo, 7200)

	assert.Equal(t, 7200, store.maxAge)
	assert.NotNil(t, store.Options)
	assert.Equal(t, "/", store.Options.Path)
	assert.Equal(t, 7200, store.Options.MaxAge)
	assert.True(t, store.Options.HttpOnly)
	assert.False(t, store.Options.Secure)
	assert.Equal(t, http.SameSiteLaxMode, store.Options.SameSite)
}

// --- New Session Tests ---

func TestNew_NoCookie_ReturnsNewEmptySession(t *testing.T) {
	store, _ := newTestPGStore()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	session, err := store.New(req, "session")

	require.NoError(t, err)
	assert.True(t, session.IsNew)
	assert.Empty(t, session.Values)
	assert.Equal(t, "session", session.Name())
}

func TestNew_CookieExistsButNotInDB_ReturnsNewEmptySession(t *testing.T) {
	store, repo := newTestPGStore()
	repo.findErr = errors.New("session not found")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "nonexistent-token"})

	session, err := store.New(req, "session")

	require.NoError(t, err)
	assert.True(t, session.IsNew)
	assert.Empty(t, session.Values)
}

func TestNew_CookieExistsAndFoundInDB_LoadsSessionData(t *testing.T) {
	store, repo := newTestPGStore()

	// Prepare session data in "DB"
	rawToken := "test-token-abc123"
	tokenHash := hashToken(rawToken)
	sessionID := uuid.New()
	now := time.Now()

	// Encode some values
	originalValues := map[interface{}]interface{}{
		"user_id": "user-123",
		"role":    "admin",
	}
	encodedData, err := gobEncode(originalValues)
	require.NoError(t, err)

	repo.sessions[tokenHash] = &models.Session{
		ID:           sessionID,
		TokenHash:    tokenHash,
		Data:         encodedData,
		LastActiveAt: now,
		ExpiresAt:    now.Add(time.Hour),
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: rawToken})

	session, err := store.New(req, "session")

	require.NoError(t, err)
	assert.False(t, session.IsNew)
	assert.Equal(t, "user-123", session.Values["user_id"])
	assert.Equal(t, "admin", session.Values["role"])
	// Internal metadata should be stashed
	assert.Equal(t, tokenHash, session.Values[sessionTokenHashKey])
	assert.Equal(t, sessionID.String(), session.Values[sessionDBIDKey])
	assert.Equal(t, now.Unix(), session.Values[sessionLastActiveKey])
}

func TestNew_CorruptedGobData_ReturnsNewEmptySession(t *testing.T) {
	store, repo := newTestPGStore()

	rawToken := "corrupt-token"
	tokenHash := hashToken(rawToken)

	repo.sessions[tokenHash] = &models.Session{
		ID:           uuid.New(),
		TokenHash:    tokenHash,
		Data:         []byte("this-is-not-valid-gob-data"),
		LastActiveAt: time.Now(),
		ExpiresAt:    time.Now().Add(time.Hour),
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: rawToken})

	session, err := store.New(req, "session")

	require.NoError(t, err)
	assert.True(t, session.IsNew)
	// Should not have loaded any internal metadata
	assert.Nil(t, session.Values[sessionTokenHashKey])
}

func TestNew_EmptyCookieValue_ReturnsNewSession(t *testing.T) {
	store, _ := newTestPGStore()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: ""})

	session, err := store.New(req, "session")

	require.NoError(t, err)
	assert.True(t, session.IsNew)
}

// --- Save Session Tests ---

func TestSave_NewSession_GeneratesTokenAndCreates(t *testing.T) {
	store, repo := newTestPGStore()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	session, err := store.New(req, "session")
	require.NoError(t, err)
	session.Values["greeting"] = "hello"

	err = store.Save(req, rec, session)
	require.NoError(t, err)

	// Session should have been created in repo
	require.Len(t, repo.created, 1)
	created := repo.created[0]
	assert.NotEmpty(t, created.TokenHash)
	assert.NotNil(t, created.Data)
	assert.False(t, created.ExpiresAt.IsZero())
	assert.False(t, created.CreatedAt.IsZero())
	assert.False(t, created.LastActiveAt.IsZero())

	// Session should no longer be new
	assert.False(t, session.IsNew)

	// Internal metadata should be stashed in session values
	assert.NotEmpty(t, session.Values[sessionTokenHashKey])
	assert.NotEmpty(t, session.Values[sessionDBIDKey])

	// A cookie should be set in the response
	cookies := rec.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "session" {
			sessionCookie = c
			break
		}
	}
	require.NotNil(t, sessionCookie)
	assert.NotEmpty(t, sessionCookie.Value)
	// Cookie value should be 128-char hex (64 bytes)
	assert.Len(t, sessionCookie.Value, 128)
}

func TestSave_ExistingSession_UpdatesRepo(t *testing.T) {
	store, repo := newTestPGStore()

	dbID := uuid.New()
	tokenHash := "existing-hash-abc"

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	req.Header.Set("User-Agent", "TestBrowser/1.0")
	rec := httptest.NewRecorder()

	session, err := store.New(req, "session")
	require.NoError(t, err)
	session.IsNew = false
	session.Values[sessionTokenHashKey] = tokenHash
	session.Values[sessionDBIDKey] = dbID.String()
	session.Values["test_key"] = "test_value"

	err = store.Save(req, rec, session)
	require.NoError(t, err)

	// Should have called Update, not Create
	assert.Empty(t, repo.created)
	require.Len(t, repo.updated, 1)
	updated := repo.updated[0]
	assert.Equal(t, dbID, updated.ID)
	assert.Equal(t, tokenHash, updated.TokenHash)
	assert.Equal(t, "10.0.0.1", updated.IPAddress)
	assert.Equal(t, "TestBrowser/1.0", updated.UserAgent)

	// Internal keys should be stripped from serialized data
	decoded := make(map[interface{}]interface{})
	err = gobDecode(updated.Data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, "test_value", decoded["test_key"])
	assert.Nil(t, decoded[sessionTokenHashKey])
	assert.Nil(t, decoded[sessionDBIDKey])
	assert.Nil(t, decoded[sessionLastActiveKey])
}

func TestSave_ExistingSession_RefreshesCookie(t *testing.T) {
	store, _ := newTestPGStore()

	dbID := uuid.New()
	tokenHash := "existing-hash-abc"
	rawToken := "raw-token-value"

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	req.Header.Set("User-Agent", "TestBrowser/1.0")
	req.AddCookie(&http.Cookie{Name: "session", Value: rawToken})
	rec := httptest.NewRecorder()

	session, err := store.New(req, "session")
	require.NoError(t, err)
	session.IsNew = false
	session.Values[sessionTokenHashKey] = tokenHash
	session.Values[sessionDBIDKey] = dbID.String()

	err = store.Save(req, rec, session)
	require.NoError(t, err)

	// Cookie should be refreshed with the same value and full MaxAge
	cookies := rec.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "session" {
			sessionCookie = c
			break
		}
	}
	require.NotNil(t, sessionCookie, "cookie should be refreshed on existing-session save")
	assert.Equal(t, rawToken, sessionCookie.Value)
	assert.Equal(t, 3600, sessionCookie.MaxAge)
}

func TestSave_DeleteSession_CallsDeleteByTokenHash(t *testing.T) {
	store, repo := newTestPGStore()

	tokenHash := "to-delete-hash"

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	session, err := store.New(req, "session")
	require.NoError(t, err)
	session.Values[sessionTokenHashKey] = tokenHash
	session.Options.MaxAge = -1 // trigger deletion

	err = store.Save(req, rec, session)
	require.NoError(t, err)

	// Should have called DeleteByTokenHash
	require.Len(t, repo.deletedTokenHashes, 1)
	assert.Equal(t, tokenHash, repo.deletedTokenHashes[0])

	// Cookie should be expired
	cookies := rec.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "session" {
			sessionCookie = c
			break
		}
	}
	require.NotNil(t, sessionCookie)
	assert.True(t, sessionCookie.MaxAge < 0)
}

func TestSave_DeleteSession_NoTokenHash_StillSetsExpiredCookie(t *testing.T) {
	store, repo := newTestPGStore()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	session, err := store.New(req, "session")
	require.NoError(t, err)
	// No tokenHash set — nothing to delete from DB
	session.Options.MaxAge = -1

	err = store.Save(req, rec, session)
	require.NoError(t, err)

	// Should NOT have called delete (no token hash)
	assert.Empty(t, repo.deletedTokenHashes)

	// Cookie should still be expired
	cookies := rec.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "session" {
			sessionCookie = c
			break
		}
	}
	require.NotNil(t, sessionCookie)
	assert.True(t, sessionCookie.MaxAge < 0)
}

func TestSave_ExistingSession_MissingMetadata_ReturnsError(t *testing.T) {
	store, _ := newTestPGStore()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	session, err := store.New(req, "session")
	require.NoError(t, err)
	session.IsNew = false
	// Deliberately omit sessionTokenHashKey and sessionDBIDKey

	err = store.Save(req, rec, session)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session missing internal metadata")
}

func TestSave_ExistingSession_InvalidDBID_ReturnsError(t *testing.T) {
	store, _ := newTestPGStore()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	session, err := store.New(req, "session")
	require.NoError(t, err)
	session.IsNew = false
	session.Values[sessionTokenHashKey] = "some-hash"
	session.Values[sessionDBIDKey] = "not-a-valid-uuid"

	err = store.Save(req, rec, session)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid session DB ID")
}

func TestSave_NewSession_WithUserID_PassesUserIDToRepo(t *testing.T) {
	store, repo := newTestPGStore()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	session, err := store.New(req, "session")
	require.NoError(t, err)

	testUserID := uuid.New().String()
	session.Values[SessionKeyUserID] = testUserID

	err = store.Save(req, rec, session)
	require.NoError(t, err)

	require.Len(t, repo.created, 1)
	created := repo.created[0]
	require.NotNil(t, created.UserID)
	assert.Equal(t, testUserID, created.UserID.String())
}

func TestSave_NewSession_WithoutUserID_NilUserID(t *testing.T) {
	store, repo := newTestPGStore()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	session, err := store.New(req, "session")
	require.NoError(t, err)
	// No user_id set

	err = store.Save(req, rec, session)
	require.NoError(t, err)

	require.Len(t, repo.created, 1)
	assert.Nil(t, repo.created[0].UserID)
}

func TestSave_CreateError_ReturnsError(t *testing.T) {
	store, repo := newTestPGStore()
	repo.createErr = errors.New("database connection lost")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	session, err := store.New(req, "session")
	require.NoError(t, err)

	err = store.Save(req, rec, session)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "create session")
}

func TestSave_UpdateError_ReturnsError(t *testing.T) {
	store, repo := newTestPGStore()
	repo.updateErr = errors.New("update failed")

	dbID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	session, err := store.New(req, "session")
	require.NoError(t, err)
	session.IsNew = false
	session.Values[sessionTokenHashKey] = "some-hash"
	session.Values[sessionDBIDKey] = dbID.String()

	err = store.Save(req, rec, session)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update session")
}

// --- Get Session Test ---

func TestGet_DelegatesToGetRegistry(t *testing.T) {
	store, _ := newTestPGStore()

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	session, err := store.Get(req, "session")
	require.NoError(t, err)
	assert.NotNil(t, session)
	assert.True(t, session.IsNew)

	// Calling Get again with the same request should return the same session
	session2, err := store.Get(req, "session")
	require.NoError(t, err)
	// gorilla/sessions.GetRegistry returns the same session for the same request+name
	assert.Same(t, session, session2)
}

// --- Helper Function Tests ---

func TestHashToken_ConsistentSHA256(t *testing.T) {
	token := "my-test-token"
	expected := sha256.Sum256([]byte(token))
	expectedHex := hex.EncodeToString(expected[:])

	result := hashToken(token)

	assert.Equal(t, expectedHex, result)
	assert.Len(t, result, 64) // SHA-256 = 32 bytes = 64 hex chars
}

func TestHashToken_DifferentInputsDifferentHashes(t *testing.T) {
	hash1 := hashToken("token-a")
	hash2 := hashToken("token-b")
	assert.NotEqual(t, hash1, hash2)
}

func TestHashToken_SameInputSameHash(t *testing.T) {
	hash1 := hashToken("same-token")
	hash2 := hashToken("same-token")
	assert.Equal(t, hash1, hash2)
}

func TestGenerateToken_Produces128CharHex(t *testing.T) {
	token, err := generateToken()
	require.NoError(t, err)

	// 64 bytes -> 128 hex characters
	assert.Len(t, token, 128)

	// Should be valid hex
	_, err = hex.DecodeString(token)
	assert.NoError(t, err)
}

func TestGenerateToken_ProducesUniqueTokens(t *testing.T) {
	token1, err := generateToken()
	require.NoError(t, err)
	token2, err := generateToken()
	require.NoError(t, err)

	assert.NotEqual(t, token1, token2)
}

func TestGobEncodeDecode_RoundTrip(t *testing.T) {
	original := map[interface{}]interface{}{
		"string_key": "string_value",
		"int_key":    42,
		"bool_key":   true,
	}

	encoded, err := gobEncode(original)
	require.NoError(t, err)
	assert.NotEmpty(t, encoded)

	decoded := make(map[interface{}]interface{})
	err = gobDecode(encoded, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "string_value", decoded["string_key"])
	assert.Equal(t, 42, decoded["int_key"])
	assert.Equal(t, true, decoded["bool_key"])
}

func TestGobEncodeDecode_EmptyMap(t *testing.T) {
	original := map[interface{}]interface{}{}

	encoded, err := gobEncode(original)
	require.NoError(t, err)

	decoded := make(map[interface{}]interface{})
	err = gobDecode(encoded, &decoded)
	require.NoError(t, err)

	assert.Empty(t, decoded)
}

func TestGobDecode_InvalidData_ReturnsError(t *testing.T) {
	decoded := make(map[interface{}]interface{})
	err := gobDecode([]byte("invalid gob data"), &decoded)
	assert.Error(t, err)
}

func TestGobDecode_UsesLimitReader(t *testing.T) {
	// Verify that gobDecode uses io.LimitReader by checking the constant
	assert.Equal(t, int64(1<<20), int64(maxSessionDataSize), "maxSessionDataSize should be 1 MB")

	// Encoding a normal map should work fine (well under 1 MB)
	original := map[interface{}]interface{}{"key": "value"}
	encoded, err := gobEncode(original)
	require.NoError(t, err)

	decoded := make(map[interface{}]interface{})
	err = gobDecode(encoded, &decoded)
	require.NoError(t, err)
	assert.Equal(t, "value", decoded["key"])
}

// --- extractIP Tests ---

func TestExtractIP_XForwardedFor_SingleIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.50")

	ip := extractIP(req)
	assert.Equal(t, "203.0.113.50", ip)
}

func TestExtractIP_XForwardedFor_MultipleIPs(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.50, 70.41.3.18, 150.172.238.178")

	ip := extractIP(req)
	assert.Equal(t, "203.0.113.50", ip)
}

func TestExtractIP_XRealIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Real-IP", "198.51.100.1")

	ip := extractIP(req)
	assert.Equal(t, "198.51.100.1", ip)
}

func TestExtractIP_XForwardedFor_TakesPrecedenceOverXRealIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.50")
	req.Header.Set("X-Real-IP", "198.51.100.1")

	ip := extractIP(req)
	assert.Equal(t, "203.0.113.50", ip)
}

func TestExtractIP_FallbackToRemoteAddr_StripsPort(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1:54321"

	ip := extractIP(req)
	assert.Equal(t, "192.168.1.1", ip)
}

func TestExtractIP_RemoteAddr_NoPort(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1"

	ip := extractIP(req)
	assert.Equal(t, "192.168.1.1", ip)
}

func TestExtractIP_IPv6_RemoteAddr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "[::1]:8080"

	ip := extractIP(req)
	// The implementation strips from last colon, so for [::1]:8080 it strips :8080
	assert.Equal(t, "[::1]", ip)
}

// --- setCookie Tests ---

func TestSetCookie_UsesOptionsSecureFlag(t *testing.T) {
	rec := httptest.NewRecorder()

	opts := &sessions.Options{
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	setCookie(rec, "session", "test-token", opts)

	cookies := rec.Result().Cookies()
	require.Len(t, cookies, 1)
	assert.True(t, cookies[0].Secure)
}

// --- GetMaxAge Test ---

func TestGetMaxAge_ReturnsConfiguredMaxAge(t *testing.T) {
	tests := []struct {
		name   string
		maxAge int
	}{
		{"1 hour", 3600},
		{"1 day", 86400},
		{"7 days", 604800},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockSessionRepo()
			store := NewPGStore(repo, tt.maxAge)
			assert.Equal(t, tt.maxAge, store.GetMaxAge())
		})
	}
}

// --- Integration-Style Tests ---

func TestSave_StripsInternalKeysFromSerializedData(t *testing.T) {
	store, repo := newTestPGStore()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	session, err := store.New(req, "session")
	require.NoError(t, err)

	session.Values["user_data"] = "important"
	session.Values[sessionTokenHashKey] = "should-be-stripped"
	session.Values[sessionDBIDKey] = "should-be-stripped"
	session.Values[sessionLastActiveKey] = int64(1234567890)

	err = store.Save(req, rec, session)
	require.NoError(t, err)

	require.Len(t, repo.created, 1)

	// Decode the stored data and verify internal keys are NOT present
	decoded := make(map[interface{}]interface{})
	err = gobDecode(repo.created[0].Data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "important", decoded["user_data"])
	assert.Nil(t, decoded[sessionTokenHashKey])
	assert.Nil(t, decoded[sessionDBIDKey])
	assert.Nil(t, decoded[sessionLastActiveKey])
}

func TestNew_Then_Save_Then_Load_RoundTrip(t *testing.T) {
	store, _ := newTestPGStore()

	// 1. Create and save a new session
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	rec1 := httptest.NewRecorder()

	session1, err := store.New(req1, "session")
	require.NoError(t, err)
	session1.Values["user_id"] = "user-abc"
	session1.Values["theme"] = "dark"

	err = store.Save(req1, rec1, session1)
	require.NoError(t, err)

	// 2. Extract cookie from response
	cookies := rec1.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "session" {
			sessionCookie = c
			break
		}
	}
	require.NotNil(t, sessionCookie)

	// 3. Load session with cookie
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.AddCookie(sessionCookie)

	session2, err := store.New(req2, "session")
	require.NoError(t, err)

	assert.False(t, session2.IsNew)
	assert.Equal(t, "user-abc", session2.Values["user_id"])
	assert.Equal(t, "dark", session2.Values["theme"])
}

func TestSave_ExpiresAtIsSetCorrectly(t *testing.T) {
	maxAge := 7200 // 2 hours
	repo := newMockSessionRepo()
	store := NewPGStore(repo, maxAge)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	session, err := store.New(req, "session")
	require.NoError(t, err)

	before := time.Now()
	err = store.Save(req, rec, session)
	require.NoError(t, err)
	after := time.Now()

	require.Len(t, repo.created, 1)
	expiresAt := repo.created[0].ExpiresAt

	// ExpiresAt should be approximately now + maxAge seconds
	expectedMin := before.Add(time.Duration(maxAge) * time.Second)
	expectedMax := after.Add(time.Duration(maxAge) * time.Second)
	assert.True(t, expiresAt.After(expectedMin) || expiresAt.Equal(expectedMin),
		"ExpiresAt %v should be >= %v", expiresAt, expectedMin)
	assert.True(t, expiresAt.Before(expectedMax) || expiresAt.Equal(expectedMax),
		"ExpiresAt %v should be <= %v", expiresAt, expectedMax)
}

func TestExtractIP_XForwardedFor_WithSpaces(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.50, 70.41.3.18")

	ip := extractIP(req)
	// The implementation scans for comma without trimming spaces
	assert.Equal(t, "203.0.113.50", ip)
}

func TestNew_SessionOptionsAreCopied(t *testing.T) {
	store, _ := newTestPGStore()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	session, err := store.New(req, "session")
	require.NoError(t, err)

	// Modifying session options should NOT affect the store's default options
	session.Options.MaxAge = 999
	assert.Equal(t, 3600, store.Options.MaxAge)
}

func TestSave_DeleteSession_RepoError_StillSetsCookie(t *testing.T) {
	store, repo := newTestPGStore()
	repo.deleteErr = errors.New("delete failed")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	session, err := store.New(req, "session")
	require.NoError(t, err)
	session.Values[sessionTokenHashKey] = "some-hash"
	session.Options.MaxAge = -1

	// Save should still succeed (delete error is logged but not returned)
	err = store.Save(req, rec, session)
	require.NoError(t, err)

	// Cookie should still be expired
	cookies := rec.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "session" {
			found = true
			assert.True(t, c.MaxAge < 0)
			break
		}
	}
	assert.True(t, found, "session cookie should be set even if DB delete fails")
}

func TestSetCookie_SetsCorrectAttributes(t *testing.T) {
	rec := httptest.NewRecorder() //nolint:bodyclose // no request body

	opts := &sessions.Options{
		Path:     "/app",
		MaxAge:   7200,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	setCookie(rec, "session", "my-token", opts)

	cookies := rec.Result().Cookies()
	require.Len(t, cookies, 1)
	cookie := cookies[0]

	assert.Equal(t, "session", cookie.Name)
	assert.Equal(t, "my-token", cookie.Value)
	assert.Equal(t, "/app", cookie.Path)
	assert.Equal(t, 7200, cookie.MaxAge)
	assert.True(t, cookie.HttpOnly)
	assert.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
}

// --- Edge Cases ---

func TestSave_NewSession_IPAndUserAgentCaptured(t *testing.T) {
	store, repo := newTestPGStore()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "10.20.30.40")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Test)")
	rec := httptest.NewRecorder()

	session, err := store.New(req, "session")
	require.NoError(t, err)

	err = store.Save(req, rec, session)
	require.NoError(t, err)

	require.Len(t, repo.created, 1)
	assert.Equal(t, "10.20.30.40", repo.created[0].IPAddress)
	assert.Equal(t, "Mozilla/5.0 (Test)", repo.created[0].UserAgent)
}

func TestSave_NewSession_InvalidUserID_IgnoredGracefully(t *testing.T) {
	store, repo := newTestPGStore()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	session, err := store.New(req, "session")
	require.NoError(t, err)
	session.Values[SessionKeyUserID] = "not-a-uuid" // invalid UUID

	err = store.Save(req, rec, session)
	require.NoError(t, err)

	require.Len(t, repo.created, 1)
	assert.Nil(t, repo.created[0].UserID, "invalid UUID should result in nil UserID")
}

func TestGenerateToken_AllCharsAreHex(t *testing.T) {
	token, err := generateToken()
	require.NoError(t, err)

	for _, c := range token {
		assert.True(t, strings.ContainsRune("0123456789abcdef", c),
			"token char %c is not a valid hex character", c)
	}
}
