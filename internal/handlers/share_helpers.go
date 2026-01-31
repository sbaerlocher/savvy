// Package handlers contains HTTP request handlers for the savvy system.
package handlers

import (
	"context"
	"fmt"
	"net/http"
	"savvy/internal/audit"
	"savvy/internal/database"
	"savvy/internal/models"
	"savvy/internal/services"

	"github.com/a-h/templ"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// deleteShareWithAudit deletes a share record with audit logging
func deleteShareWithAudit(c echo.Context, userID uuid.UUID, share any) error {
	ctx := audit.AddUserIDToContext(c.Request().Context(), userID)
	if err := database.DB.WithContext(ctx).Delete(share).Error; err != nil {
		return c.String(http.StatusInternalServerError, "Error deleting share")
	}
	return c.String(http.StatusOK, "")
}

// authzCheckFunc is a function that checks authorization for a resource
type authzCheckFunc func(ctx context.Context, userID, resourceID uuid.UUID) (*services.ResourcePermissions, error)

// handleDeleteShare is a generic handler for deleting shares
func handleDeleteShare(c echo.Context, authzCheck authzCheckFunc, shareModel any, resourceIDParam, resourceName string) error {
	user := c.Get("current_user").(*models.User)
	resourceID := c.Param(resourceIDParam)
	shareID := c.Param("share_id")

	resourceUUID, err := uuid.Parse(resourceID)
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprintf("Invalid %s ID", resourceName))
	}

	// Check authorization (only owners can delete shares)
	perms, err := authzCheck(c.Request().Context(), user.ID, resourceUUID)
	if err != nil {
		return c.String(http.StatusNotFound, fmt.Sprintf("%s not found", resourceName))
	}
	if !perms.IsOwner {
		return c.String(http.StatusForbidden, "Not authorized")
	}

	// Get share first (for audit logging)
	if err := database.DB.Where("id = ?", shareID).First(shareModel).Error; err != nil {
		return c.String(http.StatusNotFound, "Share not found")
	}

	// Delete share with user context for audit logging
	return deleteShareWithAudit(c, user.ID, shareModel)
}

// templateRenderFunc is a function that renders a share template
type templateRenderFunc func(ctx context.Context, csrfToken, resourceID string) templ.Component

// handleNewInlineShare is a generic handler for rendering inline share forms
func handleNewInlineShare(c echo.Context, authzCheck authzCheckFunc, templateFunc templateRenderFunc, resourceIDParam, resourceName string) error {
	user := c.Get("current_user").(*models.User)
	resourceID := c.Param(resourceIDParam)

	resourceUUID, err := uuid.Parse(resourceID)
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprintf("Invalid %s ID", resourceName))
	}

	// Check authorization (only owners can create shares)
	perms, err := authzCheck(c.Request().Context(), user.ID, resourceUUID)
	if err != nil {
		return c.String(http.StatusNotFound, fmt.Sprintf("%s not found", resourceName))
	}
	if !perms.IsOwner {
		return c.String(http.StatusForbidden, "Not authorized")
	}

	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = ""
	}
	component := templateFunc(c.Request().Context(), csrfToken, resourceID)
	return component.Render(c.Request().Context(), c.Response().Writer)
}

// shareWithPerms holds share data with permissions
type shareWithPerms struct {
	Share   any
	Perms   *services.ResourcePermissions
	CSRF    string
	ResID   string
	Context context.Context
}

// loadShareWithAuthz loads a share with authorization check
func loadShareWithAuthz(c echo.Context, authzCheck authzCheckFunc, shareModel any, resourceForeignKey, resourceName string, requireOwner bool) (*shareWithPerms, error) {
	user := c.Get("current_user").(*models.User)
	resourceID := c.Param("id")
	shareID := c.Param("share_id")

	resourceUUID, err := uuid.Parse(resourceID)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid %s ID", resourceName))
	}

	// Check authorization
	perms, err := authzCheck(c.Request().Context(), user.ID, resourceUUID)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("%s not found", resourceName))
	}
	if requireOwner && !perms.IsOwner {
		return nil, echo.NewHTTPError(http.StatusForbidden, "Not authorized")
	}

	// Get share with user
	query := fmt.Sprintf("id = ? AND %s = ?", resourceForeignKey)
	if err := database.DB.Where(query, shareID, resourceID).Preload("SharedWithUser").First(shareModel).Error; err != nil {
		return nil, echo.NewHTTPError(http.StatusNotFound, "Share not found")
	}

	return &shareWithPerms{
		Share:   shareModel,
		Perms:   perms,
		CSRF:    c.Get("csrf").(string),
		ResID:   resourceID,
		Context: c.Request().Context(),
	}, nil
}
