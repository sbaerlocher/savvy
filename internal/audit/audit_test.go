package audit

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"savvy/internal/models"
)

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) *gorm.DB {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://savvy:savvy_dev_password@localhost:5432/savvy?sslmode=disable" // #nosec G101 -- test credentials
	}

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		t.Skipf("Skipping test: PostgreSQL not available: %v", err)
		return nil
	}

	// Limit connection pool to prevent deadlocks with parallel test packages
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get underlying DB: %v", err)
	}
	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetMaxIdleConns(2)
	t.Cleanup(func() { _ = sqlDB.Close() })

	// Auto-migrate audit logs
	err = db.AutoMigrate(&models.User{}, &models.AuditLog{}, &models.Card{}, &models.Merchant{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	// Use targeted DELETEs (not TRUNCATE) to avoid deadlocks with
	// parallel test packages that share the same database.
	db.Exec(`DO $$
	BEGIN
		DELETE FROM audit_logs WHERE user_id IN (SELECT id FROM users WHERE email LIKE '%@example.com') OR user_id IS NULL;
		DELETE FROM cards WHERE user_id IN (SELECT id FROM users WHERE email LIKE '%@example.com') OR user_id IS NULL;
		DELETE FROM merchants WHERE name LIKE 'Test%';
		DELETE FROM users WHERE email LIKE '%@example.com';
	END $$`)

	return db
}

func TestLogDeletion(t *testing.T) {
	db := setupTestDB(t)

	// Create test user
	user := &models.User{
		Email:        "test@example.com",
		PasswordHash: "hashed",
		FirstName:    "Test",
		LastName:     "User",
	}
	db.Create(user)

	// Test data to be logged
	resourceID := uuid.New()
	resourceData := map[string]interface{}{
		"id":            resourceID,
		"merchant_name": "Test Merchant",
		"card_number":   "1234567890",
	}

	// Log deletion
	err := LogDeletion(db, &user.ID, "cards", resourceID, resourceData, "192.168.1.1", "Mozilla/5.0")
	require.NoError(t, err)

	// Verify audit log was created
	var auditLog models.AuditLog
	err = db.Where("resource_id = ?", resourceID).First(&auditLog).Error
	require.NoError(t, err)

	assert.Equal(t, user.ID, *auditLog.UserID)
	assert.Equal(t, "delete", auditLog.Action)
	assert.Equal(t, "cards", auditLog.ResourceType)
	assert.Equal(t, resourceID, auditLog.ResourceID)
	assert.Equal(t, "192.168.1.1", auditLog.IPAddress)
	assert.Equal(t, "Mozilla/5.0", auditLog.UserAgent)

	// Verify resource data is valid JSON
	var parsedData map[string]interface{}
	err = json.Unmarshal([]byte(auditLog.ResourceData), &parsedData)
	require.NoError(t, err)
	assert.Equal(t, "Test Merchant", parsedData["merchant_name"])
}

func TestLogDeletion_NilUserID(t *testing.T) {
	db := setupTestDB(t)

	resourceID := uuid.New()
	resourceData := map[string]interface{}{
		"id": resourceID,
	}

	// Log deletion without user ID (system deletion)
	err := LogDeletion(db, nil, "cards", resourceID, resourceData, "192.168.1.1", "System")
	require.NoError(t, err)

	// Verify audit log was created
	var auditLog models.AuditLog
	err = db.Where("resource_id = ?", resourceID).First(&auditLog).Error
	require.NoError(t, err)

	assert.Nil(t, auditLog.UserID)
	assert.Equal(t, "delete", auditLog.Action)
}

func TestLogDeletion_InvalidJSON(t *testing.T) {
	db := setupTestDB(t)

	// Try to log with un-marshalable data
	type InvalidStruct struct {
		Channel chan int // channels cannot be marshaled to JSON
	}

	resourceID := uuid.New()
	resourceData := InvalidStruct{
		Channel: make(chan int),
	}

	err := LogDeletion(db, nil, "cards", resourceID, resourceData, "192.168.1.1", "System")
	assert.Error(t, err) // Should fail due to JSON marshaling error
}

func TestLogDeletionFromContext(t *testing.T) {
	db := setupTestDB(t)

	// Create test user
	user := &models.User{
		Email:        "context@example.com",
		PasswordHash: "hashed",
		FirstName:    "Context",
		LastName:     "User",
	}
	db.Create(user)

	// Create Echo context
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/cards/123", nil)
	req.Header.Set("User-Agent", "TestAgent/1.0")
	req.Header.Set("X-Forwarded-For", "203.0.113.1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("current_user", user)

	// Test data
	resourceID := uuid.New()
	resourceData := map[string]string{
		"test": "data",
	}

	// Log deletion from context
	err := LogDeletionFromContext(c, db, "cards", resourceID, resourceData)
	require.NoError(t, err)

	// Verify audit log
	var auditLog models.AuditLog
	err = db.Where("resource_id = ?", resourceID).First(&auditLog).Error
	require.NoError(t, err)

	assert.Equal(t, user.ID, *auditLog.UserID)
	assert.Equal(t, "delete", auditLog.Action)
	assert.Equal(t, "TestAgent/1.0", auditLog.UserAgent)
}

func TestLogDeletionFromContext_NoUser(t *testing.T) {
	db := setupTestDB(t)

	// Create Echo context without user
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/cards/123", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	resourceID := uuid.New()
	resourceData := map[string]string{"test": "data"}

	// Should still work without user
	err := LogDeletionFromContext(c, db, "cards", resourceID, resourceData)
	require.NoError(t, err)

	var auditLog models.AuditLog
	err = db.Where("resource_id = ?", resourceID).First(&auditLog).Error
	require.NoError(t, err)

	assert.Nil(t, auditLog.UserID)
}

func TestLogUpdate(t *testing.T) {
	db := setupTestDB(t)

	user := &models.User{
		Email:        "update@example.com",
		PasswordHash: "hashed",
		FirstName:    "Update",
		LastName:     "User",
	}
	db.Create(user)

	resourceID := uuid.New()
	resourceData := map[string]interface{}{
		"id":            resourceID,
		"merchant_name": "Updated Merchant",
	}

	err := LogUpdate(db, &user.ID, "cards", resourceID, resourceData, "192.168.1.1", "Mozilla/5.0")
	require.NoError(t, err)

	var auditLog models.AuditLog
	err = db.Where("resource_id = ?", resourceID).First(&auditLog).Error
	require.NoError(t, err)

	assert.Equal(t, "update", auditLog.Action)
	assert.Equal(t, "cards", auditLog.ResourceType)
}

func TestLogUpdateFromContext(t *testing.T) {
	db := setupTestDB(t)

	user := &models.User{
		Email:        "updatectx@example.com",
		PasswordHash: "hashed",
		FirstName:    "Update",
		LastName:     "Context",
	}
	db.Create(user)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPatch, "/cards/123", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("current_user", user)

	resourceID := uuid.New()
	resourceData := map[string]string{"test": "update"}

	err := LogUpdateFromContext(c, db, "cards", resourceID, resourceData)
	require.NoError(t, err)

	var auditLog models.AuditLog
	err = db.Where("resource_id = ?", resourceID).First(&auditLog).Error
	require.NoError(t, err)

	assert.Equal(t, "update", auditLog.Action)
}

func TestLogTransfer(t *testing.T) {
	db := setupTestDB(t)

	user := &models.User{
		Email:        "transfer@example.com",
		PasswordHash: "hashed",
		FirstName:    "Transfer",
		LastName:     "User",
	}
	db.Create(user)

	resourceID := uuid.New()
	resourceData := map[string]interface{}{
		"old_owner": user.ID,
		"new_owner": uuid.New(),
	}

	err := LogTransfer(db, &user.ID, "cards", resourceID, resourceData, "192.168.1.1", "Mozilla/5.0")
	require.NoError(t, err)

	var auditLog models.AuditLog
	err = db.Where("resource_id = ?", resourceID).First(&auditLog).Error
	require.NoError(t, err)

	assert.Equal(t, "transfer", auditLog.Action)
	assert.Equal(t, "cards", auditLog.ResourceType)
}

func TestAddUserIDToContext(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	ctx = AddUserIDToContext(ctx, userID)

	// Verify value was stored
	value := ctx.Value(userIDKey)
	require.NotNil(t, value)
	assert.Equal(t, userID, value.(uuid.UUID))
}

func TestAddIPAddressToContext(t *testing.T) {
	ctx := context.Background()
	ipAddress := "192.168.1.1"

	ctx = AddIPAddressToContext(ctx, ipAddress)

	value := ctx.Value(ipAddressKey)
	require.NotNil(t, value)
	assert.Equal(t, ipAddress, value.(string))
}

func TestAddUserAgentToContext(t *testing.T) {
	ctx := context.Background()
	userAgent := "TestAgent/1.0"

	ctx = AddUserAgentToContext(ctx, userAgent)

	value := ctx.Value(userAgentKey)
	require.NotNil(t, value)
	assert.Equal(t, userAgent, value.(string))
}

func TestAddAuditContextToContext(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	ipAddress := "192.168.1.1"
	userAgent := "TestAgent/1.0"

	ctx = AddAuditContextToContext(ctx, userID, ipAddress, userAgent)

	// Verify all values were stored
	assert.Equal(t, userID, ctx.Value(userIDKey).(uuid.UUID))
	assert.Equal(t, ipAddress, ctx.Value(ipAddressKey).(string))
	assert.Equal(t, userAgent, ctx.Value(userAgentKey).(string))
}

func TestExtractAuditInfo(t *testing.T) {
	t.Run("Extract from populated context", func(t *testing.T) {
		ctx := context.Background()
		expectedIP := "192.168.1.1"
		expectedUA := "TestAgent/1.0"

		ctx = AddIPAddressToContext(ctx, expectedIP)
		ctx = AddUserAgentToContext(ctx, expectedUA)

		ipAddress, userAgent := ExtractAuditInfo(ctx)

		assert.Equal(t, expectedIP, ipAddress)
		assert.Equal(t, expectedUA, userAgent)
	})

	t.Run("Extract from empty context", func(t *testing.T) {
		ctx := context.Background()

		ipAddress, userAgent := ExtractAuditInfo(ctx)

		assert.Empty(t, ipAddress)
		assert.Empty(t, userAgent)
	})

	t.Run("Extract from nil context", func(t *testing.T) {
		ipAddress, userAgent := ExtractAuditInfo(context.Background())

		assert.Empty(t, ipAddress)
		assert.Empty(t, userAgent)
	})
}

func TestSetupAuditHooks(t *testing.T) {
	db := setupTestDB(t)

	// Setup audit hooks
	err := SetupAuditHooks(db)
	require.NoError(t, err)

	// Create user and card
	user := &models.User{
		Email:        "hook@example.com",
		PasswordHash: "hashed",
		FirstName:    "Hook",
		LastName:     "Test",
	}
	db.Create(user)

	// Create merchant first
	merchant := &models.Merchant{
		Name: "Test Merchant",
	}
	db.Create(merchant)

	card := &models.Card{
		UserID:       &user.ID,
		MerchantID:   &merchant.ID,
		MerchantName: "Test Merchant",
		CardNumber:   "1234567890",
		Program:      "Test Program",
		BarcodeType:  "CODE128",
		Status:       "active",
	}
	db.Create(card)

	// Add audit context
	ctx := context.Background()
	ctx = AddAuditContextToContext(ctx, user.ID, "192.168.1.1", "TestAgent")

	// Delete card with context
	err = db.WithContext(ctx).Delete(card).Error
	require.NoError(t, err)

	// Wait a moment for async operations
	time.Sleep(100 * time.Millisecond)

	// Verify audit log was automatically created
	var auditLog models.AuditLog
	err = db.Where("resource_id = ? AND action = ?", card.ID, "delete").First(&auditLog).Error
	require.NoError(t, err)

	assert.Equal(t, user.ID, *auditLog.UserID)
	assert.Equal(t, "cards", auditLog.ResourceType)
	assert.Equal(t, "192.168.1.1", auditLog.IPAddress)
	assert.Equal(t, "TestAgent", auditLog.UserAgent)
}

func TestBeforeDeleteHook_SkipsAuditLogs(t *testing.T) {
	db := setupTestDB(t)

	err := SetupAuditHooks(db)
	require.NoError(t, err)

	// Create an audit log
	auditLog := &models.AuditLog{
		Action:       "test",
		ResourceType: "test",
		ResourceID:   uuid.New(),
		ResourceData: `{"test": "data"}`,
		IPAddress:    "127.0.0.1",
		UserAgent:    "Test",
	}
	db.Create(auditLog)

	// Get initial count
	initialCount := int64(0)
	db.Model(&models.AuditLog{}).Count(&initialCount)
	assert.Equal(t, int64(1), initialCount)

	// Delete the audit log - should not create another audit log
	err = db.Where("id = ?", auditLog.ID).Delete(&models.AuditLog{}).Error
	require.NoError(t, err)

	// Count should be 0 (deleted), not 1 (if it had created an audit log for itself)
	finalCount := int64(0)
	db.Model(&models.AuditLog{}).Count(&finalCount)
	assert.Equal(t, int64(0), finalCount)
}

func TestBeforeDeleteHook_WithoutContext(t *testing.T) {
	db := setupTestDB(t)

	err := SetupAuditHooks(db)
	require.NoError(t, err)

	// Create card without user context
	merchant := &models.Merchant{
		Name: "Test Merchant",
	}
	db.Create(merchant)

	card := &models.Card{
		MerchantID:   &merchant.ID,
		MerchantName: "Test Merchant",
		CardNumber:   "1234567890",
		Program:      "Test Program",
		BarcodeType:  "CODE128",
		Status:       "active",
	}
	db.Create(card)

	// Delete without context
	err = db.Delete(card).Error
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	// Audit log should still be created, just without user info
	var auditLog models.AuditLog
	err = db.Where("resource_id = ?", card.ID).First(&auditLog).Error
	require.NoError(t, err)

	assert.Nil(t, auditLog.UserID)
	assert.Equal(t, "cards", auditLog.ResourceType)
}
