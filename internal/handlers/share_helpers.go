// Package handlers contains HTTP request handlers for the savvy system.
package handlers

import (
	"context"
	"net/http"
	"savvy/internal/audit"
	"savvy/internal/database"

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

// updateShareWithContext updates a share record with context
func updateShareWithContext(ctx context.Context, share any) error {
	return database.DB.WithContext(ctx).Save(share).Error
}
