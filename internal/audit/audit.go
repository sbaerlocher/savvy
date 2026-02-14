// Package audit provides audit logging for database operations.
package audit

import (
	"context"
	"encoding/json"
	"savvy/internal/models"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
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

// LogTransfer creates an audit log entry for an ownership transfer operation
func LogTransfer(db *gorm.DB, userID *uuid.UUID, resourceType string, resourceID uuid.UUID, resourceData interface{}, ipAddress, userAgent string) error {
	// Serialize resource data to JSON
	dataJSON, err := json.Marshal(resourceData)
	if err != nil {
		return err
	}

	auditLog := models.AuditLog{
		UserID:       userID,
		Action:       "transfer",
		ResourceType: resourceType,
		ResourceID:   resourceID,
		ResourceData: string(dataJSON),
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
	}

	return db.Create(&auditLog).Error
}

// LogTransferFromContext is a convenience function that extracts audit info from context
func LogTransferFromContext(ctx context.Context, db *gorm.DB, userID *uuid.UUID, resourceType string, resourceID uuid.UUID, resourceData interface{}) error {
	var ipAddress string
	var userAgent string

	// Extract IP address and user agent from context
	if ctx != nil {
		if val := ctx.Value(ipAddressKey); val != nil {
			if ip, ok := val.(string); ok {
				ipAddress = ip
			}
		}
		if val := ctx.Value(userAgentKey); val != nil {
			if ua, ok := val.(string); ok {
				userAgent = ua
			}
		}
	}

	return LogTransfer(db, userID, resourceType, resourceID, resourceData, ipAddress, userAgent)
}

// SetupAuditHooks registers GORM callbacks for automatic audit logging
func SetupAuditHooks(db *gorm.DB) error {
	// Register BeforeDelete callback to capture data BEFORE it's deleted
	return db.Callback().Delete().Before("gorm:before_delete").Register("audit:before_delete", beforeDeleteHook)
}

// beforeDeleteHook is called BEFORE any DELETE operation (to capture full data)
func beforeDeleteHook(db *gorm.DB) {
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
		// Unknown type, skip audit (intentionally excludes Notification and UserFavorite)
		return
	}

	// Try to get user ID, IP address, and user agent from context
	var userID *uuid.UUID
	var ipAddress string
	var userAgent string

	ctx := db.Statement.Context
	if ctx != nil {
		if val := ctx.Value(userIDKey); val != nil {
			if uid, ok := val.(uuid.UUID); ok {
				userID = &uid
			}
		}
		if val := ctx.Value(ipAddressKey); val != nil {
			if ip, ok := val.(string); ok {
				ipAddress = ip
			}
		}
		if val := ctx.Value(userAgentKey); val != nil {
			if ua, ok := val.(string); ok {
				userAgent = ua
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
		query = `INSERT INTO audit_logs (user_id, action, resource_type, resource_id, resource_data, ip_address, user_agent, created_at)
		         VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())`
		args = []interface{}{userID, "delete", resourceType, resourceID, dataJSON, ipAddress, userAgent}
	} else {
		query = `INSERT INTO audit_logs (action, resource_type, resource_id, resource_data, ip_address, user_agent, created_at)
		         VALUES ($1, $2, $3, $4, $5, $6, NOW())`
		args = []interface{}{"delete", resourceType, resourceID, dataJSON, ipAddress, userAgent}
	}

	_, err = sqlDB.ExecContext(db.Statement.Context, query, args...)
	if err != nil {
		db.Logger.Error(db.Statement.Context, "Failed to create audit log: %v", err)
	}
}

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	userIDKey    contextKey = "current_user_id"
	ipAddressKey contextKey = "ip_address"
	userAgentKey contextKey = "user_agent"
)

// AddUserIDToContext adds user ID to GORM context for audit logging
func AddUserIDToContext(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// AddIPAddressToContext adds IP address to GORM context for audit logging
func AddIPAddressToContext(ctx context.Context, ipAddress string) context.Context {
	return context.WithValue(ctx, ipAddressKey, ipAddress)
}

// AddUserAgentToContext adds user agent to GORM context for audit logging
func AddUserAgentToContext(ctx context.Context, userAgent string) context.Context {
	return context.WithValue(ctx, userAgentKey, userAgent)
}

// AddAuditContextToContext adds all audit information to GORM context
func AddAuditContextToContext(ctx context.Context, userID uuid.UUID, ipAddress, userAgent string) context.Context {
	ctx = AddUserIDToContext(ctx, userID)
	ctx = AddIPAddressToContext(ctx, ipAddress)
	ctx = AddUserAgentToContext(ctx, userAgent)
	return ctx
}

// ExtractAuditInfo extracts IP address and user agent from context
func ExtractAuditInfo(ctx context.Context) (ipAddress, userAgent string) {
	if ctx != nil {
		if val := ctx.Value(ipAddressKey); val != nil {
			if ip, ok := val.(string); ok {
				ipAddress = ip
			}
		}
		if val := ctx.Value(userAgentKey); val != nil {
			if ua, ok := val.(string); ok {
				userAgent = ua
			}
		}
	}
	return ipAddress, userAgent
}
