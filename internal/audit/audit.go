// Package audit provides audit logging for database operations.
package audit

import (
	"context"
	"encoding/json"
	"reflect"
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
func LogDeletionFromContext(c *echo.Context, db *gorm.DB, resourceType string, resourceID uuid.UUID, resourceData interface{}) error {
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
func LogUpdateFromContext(c *echo.Context, db *gorm.DB, resourceType string, resourceID uuid.UUID, resourceData interface{}) error {
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

// auditedTables is the set of tables whose deletions are recorded in the audit
// log. It intentionally excludes notifications and user_favorites (churn, not
// worth auditing). Membership is keyed by GORM table name.
var auditedTables = map[string]bool{
	"cards":                  true,
	"card_shares":            true,
	"vouchers":               true,
	"voucher_shares":         true,
	"gift_cards":             true,
	"gift_card_shares":       true,
	"gift_card_transactions": true,
	"merchants":              true,
}

// beforeDeleteHook is called BEFORE any DELETE operation (to capture full data).
//
// The resource id and data are read from the rows the DELETE actually targets,
// NOT from db.Statement.Dest. Callers routinely delete via an empty struct plus
// a WHERE clause (e.g. Delete(&models.GiftCard{}, "id = ?", id) or
// Where("voucher_id = ?", id).Delete(&models.VoucherShare{})); in those cases
// Dest carries a zero-value ID, which previously wrote uuid.Nil into the audit
// row and made the record impossible to restore. Re-selecting the target rows
// captures the real ids for every delete shape, single-row and bulk alike.
func beforeDeleteHook(db *gorm.DB) {
	if db.Statement == nil || db.Statement.Table == "audit_logs" {
		return
	}

	resourceType := db.Statement.Table
	if !auditedTables[resourceType] {
		return
	}

	// Re-select the rows this DELETE targets. The delete condition comes from
	// one of two mutually exclusive shapes:
	//   1. Conditional delete: Where(...).Delete(&Model{}) or
	//      Delete(&Model{}, "id = ?", id) — the condition is in the WHERE
	//      clause and Dest is a zero-value struct. The WHERE alone reproduces
	//      the delete's targeting.
	//   2. Struct delete: db.Delete(&loadedStruct) — no explicit WHERE; the
	//      primary key lives in Dest, so Model(Dest) reproduces GORM's derived
	//      PK condition.
	// These are handled separately rather than combined: layering Model(Dest)
	// under an explicit WHERE would let one silently override the other. Every
	// delete call site in this codebase is exactly one shape (verified by
	// grep), so the mixed case (keyed struct AND explicit WHERE) does not
	// occur — but if one is ever added, we fail loud below instead of
	// silently auditing the wrong rows.
	// Unscoped so soft-delete conditions don't hide the rows we're about to
	// (soft-)delete. Rows read as generic maps to stay model-agnostic.
	sel := db.Session(&gorm.Session{NewDB: true, SkipHooks: true}).
		WithContext(db.Statement.Context).
		Unscoped().
		Table(resourceType)

	whereClause, hasWhere := db.Statement.Clauses["WHERE"]
	switch {
	case hasWhere:
		if destHasKey(db) {
			// Mixed shape: a keyed struct plus an explicit WHERE. The audit
			// re-select can only reproduce one of the two conditions, so it
			// might target different rows than the DELETE. Refuse loudly
			// rather than write a misleading audit entry.
			db.Logger.Error(db.Statement.Context,
				"audit: delete on %s has both a keyed struct and an explicit WHERE; skipping audit to avoid mis-targeting", resourceType)
			return
		}
		sel.Statement.Clauses["WHERE"] = whereClause
	case db.Statement.Dest != nil && destHasKey(db):
		sel = sel.Model(db.Statement.Dest)
	default:
		// Neither a keyed struct nor a WHERE clause: refuse rather than
		// audit-log an unbounded delete.
		return
	}

	var rows []map[string]interface{}
	if err := sel.Find(&rows).Error; err != nil {
		db.Logger.Error(db.Statement.Context, "audit: failed to load rows before delete: %v", err)
		return
	}
	if len(rows) == 0 {
		return
	}

	userID, ipAddress, userAgent := auditInfoFromContext(db.Statement.Context)

	sqlDB, err := db.DB()
	if err != nil {
		db.Logger.Error(db.Statement.Context, "audit: failed to get SQL DB: %v", err)
		return
	}

	// ponytail: one INSERT per deleted row. Bulk deletes in this codebase are
	// small (shares per resource, per-user GDPR wipe, batch cap 50), so N+1 is
	// fine; switch to a multi-row INSERT if a large-fan-out delete ever lands.
	for _, row := range rows {
		resourceID, ok := rowUUID(row["id"])
		if !ok {
			continue
		}
		dataJSON, _ := json.Marshal(row)

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

		if _, err := sqlDB.ExecContext(db.Statement.Context, query, args...); err != nil {
			db.Logger.Error(db.Statement.Context, "audit: failed to create audit log: %v", err)
		}
	}
}

// destHasKey reports whether the delete's Dest struct carries a non-zero
// primary key (an "ID" field). All audited models key on ID uuid.UUID, so a
// non-nil ID means GORM will derive a PK WHERE from the struct (shape 2). A
// zero ID means Dest is just a table marker and the real condition lives in an
// explicit WHERE clause (shape 1).
func destHasKey(db *gorm.DB) bool {
	if db.Statement.Dest == nil {
		return false
	}
	v := reflect.ValueOf(db.Statement.Dest)
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return false
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return false
	}
	idField := v.FieldByName("ID")
	if !idField.IsValid() {
		return false
	}
	id, ok := idField.Interface().(uuid.UUID)
	return ok && id != uuid.Nil
}

// rowUUID coerces a scanned "id" column into a uuid.UUID. The pgx/database-sql
// path may hand back a [16]byte, a string, or a uuid.UUID depending on driver.
func rowUUID(v interface{}) (uuid.UUID, bool) {
	switch id := v.(type) {
	case uuid.UUID:
		return id, id != uuid.Nil
	case [16]byte:
		u := uuid.UUID(id)
		return u, u != uuid.Nil
	case string:
		u, err := uuid.Parse(id)
		return u, err == nil && u != uuid.Nil
	case []byte:
		u, err := uuid.ParseBytes(id)
		return u, err == nil && u != uuid.Nil
	default:
		return uuid.Nil, false
	}
}

// auditInfoFromContext pulls the acting user, IP, and user agent stashed on the
// context by the request middleware.
func auditInfoFromContext(ctx context.Context) (userID *uuid.UUID, ipAddress, userAgent string) {
	if ctx == nil {
		return nil, "", ""
	}
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
	return userID, ipAddress, userAgent
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
