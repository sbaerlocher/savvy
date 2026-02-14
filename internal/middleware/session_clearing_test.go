package middleware

import (
	"testing"

	"github.com/gorilla/sessions"
	"github.com/stretchr/testify/assert"
)

// TestClearSessionUserValues_PreservesInternalKeys verifies that the explicit
// allowlist in pgstoreInternalKeys keeps PGStore metadata intact while
// removing all user-facing values.
func TestClearSessionUserValues_PreservesInternalKeys(t *testing.T) {
	session := sessions.NewSession(nil, "session")

	// Set internal PGStore keys
	session.Values[sessionTokenHashKey] = "abc123hash"
	session.Values[sessionDBIDKey] = "db-uuid-456"
	session.Values[sessionLastActiveKey] = int64(1700000000)

	// Set user-facing keys
	session.Values[SessionKeyUserID] = "user-123"
	session.Values[SessionKeySessionCreatedAt] = int64(1700000000)
	session.Values["custom_key"] = "custom_value"

	ClearSessionUserValues(session)

	// Internal keys must be preserved
	assert.Equal(t, "abc123hash", session.Values[sessionTokenHashKey])
	assert.Equal(t, "db-uuid-456", session.Values[sessionDBIDKey])
	assert.Equal(t, int64(1700000000), session.Values[sessionLastActiveKey])

	// User keys must be removed
	_, hasUserID := session.Values[SessionKeyUserID]
	assert.False(t, hasUserID, "user_id should be removed")

	_, hasCreatedAt := session.Values[SessionKeySessionCreatedAt]
	assert.False(t, hasCreatedAt, "session_created_at should be removed")

	_, hasCustom := session.Values["custom_key"]
	assert.False(t, hasCustom, "custom_key should be removed")
}

// TestClearSessionUserValues_EmptySession verifies no panic on an empty session.
func TestClearSessionUserValues_EmptySession(t *testing.T) {
	session := sessions.NewSession(nil, "session")

	// Should not panic
	ClearSessionUserValues(session)
	assert.Empty(t, session.Values)
}

// TestClearSessionUserValues_OnlyInternalKeys verifies nothing is removed when
// only internal keys are present.
func TestClearSessionUserValues_OnlyInternalKeys(t *testing.T) {
	session := sessions.NewSession(nil, "session")
	session.Values[sessionTokenHashKey] = "hash"
	session.Values[sessionDBIDKey] = "id"
	session.Values[sessionLastActiveKey] = int64(100)

	ClearSessionUserValues(session)

	assert.Len(t, session.Values, 3)
}

// TestClearSessionUserValues_NonStringKeys verifies that non-string keys
// (e.g. integers used by some session stores) are removed.
func TestClearSessionUserValues_NonStringKeys(t *testing.T) {
	session := sessions.NewSession(nil, "session")
	session.Values[sessionTokenHashKey] = "hash"
	session.Values[42] = "integer key"

	ClearSessionUserValues(session)

	assert.Equal(t, "hash", session.Values[sessionTokenHashKey])
	_, hasIntKey := session.Values[42]
	assert.False(t, hasIntKey, "non-string key should be removed")
}

// TestClearSessionUserValues_KeysWithDoubleUnderscoreNotPreserved verifies
// that the allowlist approach does NOT blindly preserve all "__" prefixed keys.
// Only the three explicit keys in pgstoreInternalKeys are preserved.
func TestClearSessionUserValues_KeysWithDoubleUnderscoreNotPreserved(t *testing.T) {
	session := sessions.NewSession(nil, "session")

	// Real internal keys
	session.Values[sessionTokenHashKey] = "real"

	// Fake "__" prefixed key that is NOT in the allowlist
	session.Values["__fake_internal_key"] = "should be removed"

	ClearSessionUserValues(session)

	assert.Equal(t, "real", session.Values[sessionTokenHashKey])

	_, hasFake := session.Values["__fake_internal_key"]
	assert.False(t, hasFake, "unknown __ key should NOT be preserved by allowlist")
}
