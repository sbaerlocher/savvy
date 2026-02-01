// Package handlers contains HTTP request handlers for the savvy system.
package handlers

import (
	"net/http"
	"savvy/internal/models"
	"savvy/internal/services"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// SharedUserResponse represents a user that has been shared with
type SharedUserResponse struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
}

// SharedUsersHandler handles shared users operations
type SharedUsersHandler struct {
	shareService services.ShareServiceInterface
}

// NewSharedUsersHandler creates a new shared users handler
func NewSharedUsersHandler(shareService services.ShareServiceInterface) *SharedUsersHandler {
	return &SharedUsersHandler{
		shareService: shareService,
	}
}

// Autocomplete returns users the current user has previously shared any resource with
// This endpoint is used for autocomplete in share forms
// GET /api/shared-users?q=search_query
func (h *SharedUsersHandler) Autocomplete(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	searchQuery := c.QueryParam("q")

	sharedUsers, err := h.shareService.GetSharedUsers(c.Request().Context(), user.ID, searchQuery)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch shared users",
		})
	}

	// Convert to response format
	response := make([]SharedUserResponse, len(sharedUsers))
	for i, u := range sharedUsers {
		response[i] = SharedUserResponse{
			ID:    u.ID,
			Name:  u.DisplayName(),
			Email: u.Email,
		}
	}

	return c.JSON(http.StatusOK, response)
}
