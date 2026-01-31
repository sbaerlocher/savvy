// Package audit provides audit logging for database operations.
package audit

import (
	"context"
	"encoding/json"
	"savvy/internal/models"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"gorm.io/gorm"
)

// LogDeletion creates an audit log entry for a deletion operation
func LogDeletion(db *gorm.DB, userID *uuid.UUID, resourceType string, resourceID uuid.UUID, resourceData interface{}, ipAddress, userAgent string) error {
	// Serialize resource data to JSON
	dataJSON, err := json.Marshal(resourceData)
	if err != nil {
		return err
	}

	auditLog := models.AuditLog{
		UserID:       userID,
		Action:       "delete",
		ResourceType: resourceType,
		ResourceID:   resourceID,
		ResourceData: string(dataJSON),
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
	}

	return db.Create(&auditLog).Error
}

// LogDeletionFromContext is a convenience function that extracts user info from Echo context
func LogDeletionFromContext(c echo.Context, db *gorm.DB, resourceType string, resourceID uuid.UUID, resourceData interface{}) error {
	var userID *uuid.UUID
	if user, ok := c.Get("current_user").(*models.User); ok && user != nil {
		userID = &user.ID
	}

	ipAddress := c.RealIP()
	userAgent := c.Request().UserAgent()

	return LogDeletion(db, userID, resourceType, resourceID, resourceData, ipAddress, userAgent)
}

// LogUpdate creates an audit log entry for an update operation
func LogUpdate(db *gorm.DB, userID *uuid.UUID, resourceType string, resourceID uuid.UUID, resourceData interface{}, ipAddress, userAgent string) error {
	// Serialize resource data to JSON
	dataJSON, err := json.Marshal(resourceData)
	if err != nil {
		return err
	}

	auditLog := models.AuditLog{
		UserID:       userID,
		Action:       "update",
		ResourceType: resourceType,
		ResourceID:   resourceID,
		ResourceData: string(dataJSON),
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
	}

	return db.Create(&auditLog).Error
}

// LogUpdateFromContext is a convenience function that extracts user info from Echo context
func LogUpdateFromContext(c echo.Context, db *gorm.DB, resourceType string, resourceID uuid.UUID, resourceData interface{}) error {
	var userID *uuid.UUID
	if user, ok := c.Get("current_user").(*models.User); ok && user != nil {
		userID = &user.ID
	}

	ipAddress := c.RealIP()
	userAgent := c.Request().UserAgent()

	return LogUpdate(db, userID, resourceType, resourceID, resourceData, ipAddress, userAgent)
}

// SetupAuditHooks registers GORM callbacks for automatic audit logging
func SetupAuditHooks(db *gorm.DB) error {
	// Register AfterDelete callback for all models
	return db.Callback().Delete().After("gorm:after_delete").Register("audit:after_delete", afterDeleteHook)
}

// afterDeleteHook is called after any DELETE operation
func afterDeleteHook(db *gorm.DB) {
	// Skip if this is the audit_logs table itself
	if db.Statement.Table == "audit_logs" {
		return
	}

	// Get the deleted record from the statement
	if db.Statement.Dest == nil {
		return
	}

	// Extract resource info
	resourceType := db.Statement.Table

	// Try to get the ID from the deleted record
	var resourceID uuid.UUID
	switch v := db.Statement.Dest.(type) {
	case *models.Card:
		resourceID = v.ID
	case *models.CardShare:
		resourceID = v.ID
	case *models.Voucher:
		resourceID = v.ID
	case *models.VoucherShare:
		resourceID = v.ID
	case *models.GiftCard:
		resourceID = v.ID
	case *models.GiftCardShare:
		resourceID = v.ID
	case *models.GiftCardTransaction:
		resourceID = v.ID
	case *models.Merchant:
		resourceID = v.ID
	default:
		// Unknown type, skip audit
		return
	}

	// Try to get user ID from context
	var userID *uuid.UUID
	ctx := db.Statement.Context
	if ctx != nil {
		if val := ctx.Value(userIDKey); val != nil {
			if uid, ok := val.(uuid.UUID); ok {
				userID = &uid
			}
		}
	}

	// Create audit log (without triggering another hook)
	dataJSON, _ := json.Marshal(db.Statement.Dest)

	// Get underlying SQL DB connection
	sqlDB, err := db.DB()
	if err != nil {
		db.Logger.Error(db.Statement.Context, "Failed to get SQL DB: %v", err)
		return
	}

	// Insert directly using database/sql to avoid GORM type issues
	var query string
	var args []interface{}

	if userID != nil {
		query = `INSERT INTO audit_logs (user_id, action, resource_type, resource_id, resource_data, created_at)
		         VALUES ($1, $2, $3, $4, $5, NOW())`
		args = []interface{}{userID, "delete", resourceType, resourceID, dataJSON}
	} else {
		query = `INSERT INTO audit_logs (action, resource_type, resource_id, resource_data, created_at)
		         VALUES ($1, $2, $3, $4, NOW())`
		args = []interface{}{"delete", resourceType, resourceID, dataJSON}
	}

	_, err = sqlDB.ExecContext(db.Statement.Context, query, args...)
	if err != nil {
		db.Logger.Error(db.Statement.Context, "Failed to create audit log: %v", err)
	}
}

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const userIDKey contextKey = "current_user_id"

// AddUserIDToContext adds user ID to GORM context for audit logging
func AddUserIDToContext(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}
