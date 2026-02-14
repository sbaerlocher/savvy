// Package services contains business logic.
package services

import (
	"context"
	"fmt"
	"savvy/internal/models"
	"savvy/internal/repository"
	"strings"
	"time"

	"github.com/google/uuid"
)

// SessionDTO represents a session for API responses.
type SessionDTO struct {
	ID           string `json:"id"`
	IPAddress    string `json:"ip_address"`
	UserAgent    string `json:"user_agent"`
	DeviceInfo   string `json:"device_info"`
	BrowserInfo  string `json:"browser_info"`
	IsCurrent    bool   `json:"is_current"`
	CreatedAt    string `json:"created_at"`
	LastActiveAt string `json:"last_active_at"`
}

// SessionServiceInterface defines session management operations.
type SessionServiceInterface interface {
	ListUserSessions(ctx context.Context, userID uuid.UUID, currentTokenHash string) ([]SessionDTO, error)
	RevokeSession(ctx context.Context, userID, sessionID uuid.UUID) error
	RevokeOtherSessions(ctx context.Context, userID uuid.UUID, currentTokenHash string) (int64, error)
	RevokeAllSessions(ctx context.Context, userID uuid.UUID) (int64, error)
	CleanupExpired(ctx context.Context) (int64, error)
}

// SessionService implements SessionServiceInterface.
type SessionService struct {
	repo repository.SessionRepository
}

// NewSessionService creates a new session service.
func NewSessionService(repo repository.SessionRepository) SessionServiceInterface {
	return &SessionService{repo: repo}
}

// ListUserSessions returns all active sessions for a user.
func (s *SessionService) ListUserSessions(ctx context.Context, userID uuid.UUID, currentTokenHash string) ([]SessionDTO, error) {
	sessions, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}

	dtos := make([]SessionDTO, 0, len(sessions))
	for _, sess := range sessions {
		dtos = append(dtos, sessionToDTO(sess, currentTokenHash))
	}
	return dtos, nil
}

// RevokeSession deletes a specific session.
func (s *SessionService) RevokeSession(ctx context.Context, userID, sessionID uuid.UUID) error {
	if err := s.repo.DeleteByID(ctx, sessionID, userID); err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	return nil
}

// RevokeOtherSessions deletes all sessions for a user except the current one.
func (s *SessionService) RevokeOtherSessions(ctx context.Context, userID uuid.UUID, currentTokenHash string) (int64, error) {
	count, err := s.repo.DeleteAllByUserIDExcept(ctx, userID, currentTokenHash)
	if err != nil {
		return 0, fmt.Errorf("revoke other sessions: %w", err)
	}
	return count, nil
}

// RevokeAllSessions deletes all sessions for a user.
func (s *SessionService) RevokeAllSessions(ctx context.Context, userID uuid.UUID) (int64, error) {
	count, err := s.repo.DeleteAllByUserID(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("revoke all sessions: %w", err)
	}
	return count, nil
}

// CleanupExpired removes all expired sessions.
func (s *SessionService) CleanupExpired(ctx context.Context) (int64, error) {
	count, err := s.repo.DeleteExpired(ctx)
	if err != nil {
		return 0, fmt.Errorf("cleanup expired sessions: %w", err)
	}
	return count, nil
}

// sessionToDTO converts a session model to a DTO.
func sessionToDTO(sess models.Session, currentTokenHash string) SessionDTO {
	return SessionDTO{
		ID:           sess.ID.String(),
		IPAddress:    sess.IPAddress,
		UserAgent:    sess.UserAgent,
		DeviceInfo:   parseDeviceInfo(sess.UserAgent),
		BrowserInfo:  parseBrowserInfo(sess.UserAgent),
		IsCurrent:    sess.TokenHash == currentTokenHash,
		CreatedAt:    sess.CreatedAt.Format(time.RFC3339),
		LastActiveAt: sess.LastActiveAt.Format(time.RFC3339),
	}
}

// parseDeviceInfo extracts OS/device information from a user-agent string.
func parseDeviceInfo(ua string) string {
	ua = strings.ToLower(ua)

	switch {
	case strings.Contains(ua, "iphone"):
		return "iPhone"
	case strings.Contains(ua, "ipad"):
		return "iPad"
	case strings.Contains(ua, "android") && strings.Contains(ua, "mobile"):
		return "Android Phone"
	case strings.Contains(ua, "android"):
		return "Android Tablet"
	case strings.Contains(ua, "macintosh") || strings.Contains(ua, "mac os"):
		return "macOS"
	case strings.Contains(ua, "windows"):
		return "Windows"
	case strings.Contains(ua, "linux"):
		return "Linux"
	case strings.Contains(ua, "cros"):
		return "ChromeOS"
	default:
		return "Unknown"
	}
}

// parseBrowserInfo extracts browser name from a user-agent string.
func parseBrowserInfo(ua string) string {
	uaLower := strings.ToLower(ua)

	// Order matters: check more specific browsers first
	switch {
	case strings.Contains(uaLower, "edg/") || strings.Contains(uaLower, "edge/"):
		return "Edge"
	case strings.Contains(uaLower, "opr/") || strings.Contains(uaLower, "opera"):
		return "Opera"
	case strings.Contains(uaLower, "brave"):
		return "Brave"
	case strings.Contains(uaLower, "vivaldi"):
		return "Vivaldi"
	case strings.Contains(uaLower, "firefox/"):
		return "Firefox"
	case strings.Contains(uaLower, "safari/") && !strings.Contains(uaLower, "chrome/"):
		return "Safari"
	case strings.Contains(uaLower, "chrome/"):
		return "Chrome"
	default:
		return "Unknown"
	}
}
