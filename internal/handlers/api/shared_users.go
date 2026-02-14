// Package api contains JSON API handlers for shared users.
//
//nolint:revive // "api" is a meaningful package name for API handlers
package api

import (
	"net/http"
	"savvy/internal/models"
	"savvy/internal/services"

	"github.com/labstack/echo/v4"
)

// SharedUsersHandler handles shared users API endpoints.
type SharedUsersHandler struct {
	shareService services.ShareServiceInterface
}

// NewSharedUsersHandler creates a new shared users API handler.
func NewSharedUsersHandler(shareService services.ShareServiceInterface) *SharedUsersHandler {
	return &SharedUsersHandler{
		shareService: shareService,
	}
}

// Search returns users that the current user has shared resources with
// GET /api/v1/shared-users?q=search
func (h *SharedUsersHandler) Search(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	searchQuery := c.QueryParam("q")

	users, err := h.shareService.GetSharedUsers(c.Request().Context(), user.ID, searchQuery)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to search users",
		})
	}

	// Convert to DTOs
	userDTOs := make([]UserDTO, len(users))
	for i, u := range users {
		userDTOs[i] = UserDTO{
			ID:        u.ID.String(),
			Email:     u.Email,
			FirstName: u.FirstName,
			LastName:  u.LastName,
			IsAdmin:   u.IsAdmin(),
		}
	}

	return c.JSON(http.StatusOK, map[string]any{
		"users": userDTOs,
	})
}
