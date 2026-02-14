package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"savvy/internal/models"
)

// ============================================================================
// MOCK FOR SESSION REPOSITORY
// ============================================================================

type mockSessionRepo struct {
	mock.Mock
}

func (m *mockSessionRepo) FindByTokenHash(ctx context.Context, tokenHash string) (*models.Session, error) {
	args := m.Called(ctx, tokenHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Session), args.Error(1)
}

func (m *mockSessionRepo) Create(ctx context.Context, session *models.Session) error {
	return m.Called(ctx, session).Error(0)
}

func (m *mockSessionRepo) Update(ctx context.Context, session *models.Session) error {
	return m.Called(ctx, session).Error(0)
}

func (m *mockSessionRepo) DeleteByTokenHash(ctx context.Context, tokenHash string) error {
	return m.Called(ctx, tokenHash).Error(0)
}

func (m *mockSessionRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.Session, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Session), args.Error(1)
}

func (m *mockSessionRepo) DeleteByID(ctx context.Context, sessionID, userID uuid.UUID) error {
	return m.Called(ctx, sessionID, userID).Error(0)
}

func (m *mockSessionRepo) DeleteAllByUserIDExcept(ctx context.Context, userID uuid.UUID, exceptTokenHash string) (int64, error) {
	args := m.Called(ctx, userID, exceptTokenHash)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockSessionRepo) DeleteAllByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockSessionRepo) DeleteExpired(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockSessionRepo) CountActive(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

// ============================================================================
// ListUserSessions Tests
// ============================================================================

func TestSessionService_ListUserSessions_Success(t *testing.T) {
	repo := new(mockSessionRepo)
	svc := NewSessionService(repo)
	ctx := context.Background()
	userID := uuid.New()
	currentHash := "current-token-hash"

	sessions := []models.Session{
		{
			ID:           uuid.New(),
			TokenHash:    currentHash,
			IPAddress:    "1.2.3.4",
			UserAgent:    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) Chrome/120.0",
			CreatedAt:    time.Now().Add(-1 * time.Hour),
			LastActiveAt: time.Now(),
		},
		{
			ID:           uuid.New(),
			TokenHash:    "other-hash",
			IPAddress:    "5.6.7.8",
			UserAgent:    "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0) Safari/605.1",
			CreatedAt:    time.Now().Add(-24 * time.Hour),
			LastActiveAt: time.Now().Add(-2 * time.Hour),
		},
	}

	repo.On("GetByUserID", mock.Anything, userID).Return(sessions, nil)

	result, err := svc.ListUserSessions(ctx, userID, currentHash)

	require.NoError(t, err)
	require.Len(t, result, 2)

	// First session should be marked as current
	assert.True(t, result[0].IsCurrent)
	assert.False(t, result[1].IsCurrent)

	// Device/browser parsing
	assert.Equal(t, "macOS", result[0].DeviceInfo)
	assert.Equal(t, "Chrome", result[0].BrowserInfo)
	assert.Equal(t, "iPhone", result[1].DeviceInfo)
	assert.Equal(t, "Safari", result[1].BrowserInfo)
}

func TestSessionService_ListUserSessions_Error(t *testing.T) {
	repo := new(mockSessionRepo)
	svc := NewSessionService(repo)
	ctx := context.Background()
	userID := uuid.New()

	repo.On("GetByUserID", mock.Anything, userID).Return(nil, errors.New("db error"))

	result, err := svc.ListUserSessions(ctx, userID, "hash")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "list sessions")
}

// ============================================================================
// RevokeSession Tests
// ============================================================================

func TestSessionService_RevokeSession_Success(t *testing.T) {
	repo := new(mockSessionRepo)
	svc := NewSessionService(repo)
	ctx := context.Background()
	userID := uuid.New()
	sessionID := uuid.New()

	repo.On("DeleteByID", mock.Anything, sessionID, userID).Return(nil)

	err := svc.RevokeSession(ctx, userID, sessionID)
	assert.NoError(t, err)
}

func TestSessionService_RevokeSession_Error(t *testing.T) {
	repo := new(mockSessionRepo)
	svc := NewSessionService(repo)
	ctx := context.Background()
	userID := uuid.New()
	sessionID := uuid.New()

	repo.On("DeleteByID", mock.Anything, sessionID, userID).Return(errors.New("not found"))

	err := svc.RevokeSession(ctx, userID, sessionID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "revoke session")
}

// ============================================================================
// RevokeOtherSessions Tests
// ============================================================================

func TestSessionService_RevokeOtherSessions_Success(t *testing.T) {
	repo := new(mockSessionRepo)
	svc := NewSessionService(repo)
	ctx := context.Background()
	userID := uuid.New()
	currentHash := "keep-this-session"

	repo.On("DeleteAllByUserIDExcept", mock.Anything, userID, currentHash).Return(int64(3), nil)

	count, err := svc.RevokeOtherSessions(ctx, userID, currentHash)

	require.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

func TestSessionService_RevokeOtherSessions_RepoError(t *testing.T) {
	repo := new(mockSessionRepo)
	svc := NewSessionService(repo)
	ctx := context.Background()
	userID := uuid.New()

	repo.On("DeleteAllByUserIDExcept", mock.Anything, userID, "hash").
		Return(int64(0), errors.New("repo error"))

	count, err := svc.RevokeOtherSessions(ctx, userID, "hash")

	assert.Error(t, err)
	assert.Equal(t, int64(0), count)
	assert.Contains(t, err.Error(), "revoke other sessions")
}

// ============================================================================
// RevokeAllSessions Tests
// ============================================================================

func TestSessionService_RevokeAllSessions_Success(t *testing.T) {
	repo := new(mockSessionRepo)
	svc := NewSessionService(repo)
	ctx := context.Background()
	userID := uuid.New()

	repo.On("DeleteAllByUserID", mock.Anything, userID).Return(int64(5), nil)

	count, err := svc.RevokeAllSessions(ctx, userID)

	require.NoError(t, err)
	assert.Equal(t, int64(5), count)
}

func TestSessionService_RevokeAllSessions_Error(t *testing.T) {
	repo := new(mockSessionRepo)
	svc := NewSessionService(repo)
	ctx := context.Background()
	userID := uuid.New()

	repo.On("DeleteAllByUserID", mock.Anything, userID).Return(int64(0), errors.New("db error"))

	count, err := svc.RevokeAllSessions(ctx, userID)

	assert.Error(t, err)
	assert.Equal(t, int64(0), count)
	assert.Contains(t, err.Error(), "revoke all sessions")
}

func TestSessionService_RevokeAllSessions_ZeroDeleted(t *testing.T) {
	repo := new(mockSessionRepo)
	svc := NewSessionService(repo)
	ctx := context.Background()
	userID := uuid.New()

	repo.On("DeleteAllByUserID", mock.Anything, userID).Return(int64(0), nil)

	count, err := svc.RevokeAllSessions(ctx, userID)

	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

// ============================================================================
// CleanupExpired Tests
// ============================================================================

func TestSessionService_CleanupExpired_Success(t *testing.T) {
	repo := new(mockSessionRepo)
	svc := NewSessionService(repo)
	ctx := context.Background()

	repo.On("DeleteExpired", mock.Anything).Return(int64(10), nil)

	count, err := svc.CleanupExpired(ctx)

	require.NoError(t, err)
	assert.Equal(t, int64(10), count)
}

func TestSessionService_CleanupExpired_Error(t *testing.T) {
	repo := new(mockSessionRepo)
	svc := NewSessionService(repo)
	ctx := context.Background()

	repo.On("DeleteExpired", mock.Anything).Return(int64(0), errors.New("db error"))

	count, err := svc.CleanupExpired(ctx)

	assert.Error(t, err)
	assert.Equal(t, int64(0), count)
}

// ============================================================================
// ListUserSessions - Empty List
// ============================================================================

func TestSessionService_ListUserSessions_EmptyList(t *testing.T) {
	repo := new(mockSessionRepo)
	svc := NewSessionService(repo)
	ctx := context.Background()
	userID := uuid.New()

	repo.On("GetByUserID", mock.Anything, userID).Return([]models.Session{}, nil)

	result, err := svc.ListUserSessions(ctx, userID, "some-hash")

	require.NoError(t, err)
	assert.Empty(t, result)
	assert.NotNil(t, result)
}

// ============================================================================
// sessionToDTO Tests
// ============================================================================

func TestSessionToDTO_CurrentSession(t *testing.T) {
	sessionID := uuid.New()
	tokenHash := "matching-hash"
	createdAt := time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)
	lastActive := time.Date(2026, 3, 4, 15, 30, 0, 0, time.UTC)

	sess := models.Session{
		ID:           sessionID,
		TokenHash:    tokenHash,
		IPAddress:    "192.168.1.1",
		UserAgent:    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) Chrome/120.0",
		CreatedAt:    createdAt,
		LastActiveAt: lastActive,
	}

	dto := sessionToDTO(sess, tokenHash)

	assert.Equal(t, sessionID.String(), dto.ID)
	assert.Equal(t, "192.168.1.1", dto.IPAddress)
	assert.Equal(t, "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) Chrome/120.0", dto.UserAgent)
	assert.Equal(t, "macOS", dto.DeviceInfo)
	assert.Equal(t, "Chrome", dto.BrowserInfo)
	assert.True(t, dto.IsCurrent)
	assert.Equal(t, createdAt.Format(time.RFC3339), dto.CreatedAt)
	assert.Equal(t, lastActive.Format(time.RFC3339), dto.LastActiveAt)
}

func TestSessionToDTO_NotCurrentSession(t *testing.T) {
	sess := models.Session{
		ID:           uuid.New(),
		TokenHash:    "session-hash",
		IPAddress:    "10.0.0.1",
		UserAgent:    "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0) Safari/605.1",
		CreatedAt:    time.Now().Add(-24 * time.Hour),
		LastActiveAt: time.Now().Add(-1 * time.Hour),
	}

	dto := sessionToDTO(sess, "different-hash")

	assert.False(t, dto.IsCurrent)
	assert.Equal(t, "iPhone", dto.DeviceInfo)
	assert.Equal(t, "Safari", dto.BrowserInfo)
}

func TestSessionToDTO_EmptyUserAgent(t *testing.T) {
	sess := models.Session{
		ID:           uuid.New(),
		TokenHash:    "hash",
		IPAddress:    "",
		UserAgent:    "",
		CreatedAt:    time.Now(),
		LastActiveAt: time.Now(),
	}

	dto := sessionToDTO(sess, "hash")

	assert.True(t, dto.IsCurrent)
	assert.Equal(t, "Unknown", dto.DeviceInfo)
	assert.Equal(t, "Unknown", dto.BrowserInfo)
	assert.Empty(t, dto.IPAddress)
}

func TestSessionToDTO_WindowsFirefox(t *testing.T) {
	sess := models.Session{
		ID:           uuid.New(),
		TokenHash:    "win-hash",
		IPAddress:    "203.0.113.5",
		UserAgent:    "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
		CreatedAt:    time.Now(),
		LastActiveAt: time.Now(),
	}

	dto := sessionToDTO(sess, "other-hash")

	assert.False(t, dto.IsCurrent)
	assert.Equal(t, "Windows", dto.DeviceInfo)
	assert.Equal(t, "Firefox", dto.BrowserInfo)
}

func TestSessionToDTO_LinuxEdge(t *testing.T) {
	sess := models.Session{
		ID:           uuid.New(),
		TokenHash:    "linux-hash",
		IPAddress:    "172.16.0.1",
		UserAgent:    "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 Chrome/120.0 Edg/120.0",
		CreatedAt:    time.Now(),
		LastActiveAt: time.Now(),
	}

	dto := sessionToDTO(sess, "linux-hash")

	assert.True(t, dto.IsCurrent)
	assert.Equal(t, "Linux", dto.DeviceInfo)
	assert.Equal(t, "Edge", dto.BrowserInfo)
}

// ============================================================================
// parseDeviceInfo Tests
// ============================================================================

func TestParseDeviceInfo(t *testing.T) {
	tests := []struct {
		ua       string
		expected string
	}{
		{"Mozilla/5.0 (iPhone; CPU iPhone OS 17_0)", "iPhone"},
		{"Mozilla/5.0 (iPad; CPU OS 17_0)", "iPad"},
		{"Mozilla/5.0 (Linux; Android 14; Pixel 8) Mobile", "Android Phone"},
		{"Mozilla/5.0 (Linux; Android 14; SM-T500)", "Android Tablet"},
		{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)", "macOS"},
		{"Mozilla/5.0 (Windows NT 10.0; Win64; x64)", "Windows"},
		{"Mozilla/5.0 (X11; Linux x86_64)", "Linux"},
		{"Mozilla/5.0 (X11; CrOS x86_64 14541.0.0)", "ChromeOS"},
		{"SomeBot/1.0", "Unknown"},
		{"", "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, parseDeviceInfo(tt.ua))
		})
	}
}

// ============================================================================
// parseBrowserInfo Tests
// ============================================================================

func TestParseBrowserInfo(t *testing.T) {
	tests := []struct {
		ua       string
		expected string
	}{
		{"Mozilla/5.0 Edg/120.0", "Edge"},
		{"Mozilla/5.0 Edge/120.0", "Edge"},
		{"Mozilla/5.0 OPR/105.0", "Opera"},
		{"Mozilla/5.0 Opera/9.80", "Opera"},
		{"Mozilla/5.0 Brave Chrome/120.0", "Brave"},
		{"Mozilla/5.0 Vivaldi/6.4", "Vivaldi"},
		{"Mozilla/5.0 Firefox/121.0", "Firefox"},
		{"Mozilla/5.0 (Macintosh) AppleWebKit/605.1 Safari/605.1", "Safari"},
		{"Mozilla/5.0 Chrome/120.0.0.0 Safari/537.36", "Chrome"},
		{"SomeBot/1.0", "Unknown"},
		{"", "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, parseBrowserInfo(tt.ua))
		})
	}
}
