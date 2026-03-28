// Package api contains JSON API handlers for the SvelteKit frontend.
package api

//
//nolint:revive // "api" is a meaningful package name for API handlers

import (
	"log/slog"
	"net/http"
	"savvy/internal/middleware"
	"savvy/internal/models"
	"savvy/internal/services"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

// SessionsHandler handles session management API endpoints.
type SessionsHandler struct {
	sessionService services.SessionServiceInterface
}

// NewSessionsHandler creates a new sessions API handler.
func NewSessionsHandler(sessionService services.SessionServiceInterface) *SessionsHandler {
	return &SessionsHandler{
		sessionService: sessionService,
	}
}

// List returns all active sessions for the authenticated user.
// GET /api/v1/profile/sessions
func (h *SessionsHandler) List(c echo.Context) error {
	user, ok := c.Get("current_user").(*models.User)
	if !ok || user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "Not authenticated",
		})
	}

	currentTokenHash := middleware.GetCurrentSessionTokenHash(c)

	sessions, err := h.sessionService.ListUserSessions(c.Request().Context(), user.ID, currentTokenHash)
	if err != nil {
		slog.Error("Failed to list sessions", "user_id", user.ID, "error", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to load sessions",
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"sessions": sessions,
	})
}

// Revoke deletes a specific session.
// DELETE /api/v1/profile/sessions/:id
func (h *SessionsHandler) Revoke(c echo.Context) error {
	user, ok := c.Get("current_user").(*models.User)
	if !ok || user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "Not authenticated",
		})
	}

	sessionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid session ID",
		})
	}

	if err := h.sessionService.RevokeSession(c.Request().Context(), user.ID, sessionID); err != nil {
		slog.Error("Failed to revoke session", "user_id", user.ID, "session_id", sessionID, "error", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to revoke session",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Session revoked",
	})
}

// RevokeOthers revokes all sessions except the current one.
// POST /api/v1/profile/sessions/revoke-others
func (h *SessionsHandler) RevokeOthers(c echo.Context) error {
	user, ok := c.Get("current_user").(*models.User)
	if !ok || user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "Not authenticated",
		})
	}

	currentTokenHash := middleware.GetCurrentSessionTokenHash(c)
	if currentTokenHash == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "no_session",
			Message: "No active session found",
		})
	}

	count, err := h.sessionService.RevokeOtherSessions(c.Request().Context(), user.ID, currentTokenHash)
	if err != nil {
		slog.Error("Failed to revoke other sessions", "user_id", user.ID, "error", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to revoke other sessions",
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"message":       "Other sessions revoked",
		"revoked_count": count,
	})
}
