package audit

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"savvy/internal/models"
	"savvy/internal/testutil"
)

func TestLogDeletion(t *testing.T) {
	db := testutil.NewTestDB(t)

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
	db := testutil.NewTestDB(t)

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
	db := testutil.NewTestDB(t)

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
	db := testutil.NewTestDB(t)

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
	db := testutil.NewTestDB(t)

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
	db := testutil.NewTestDB(t)

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
	db := testutil.NewTestDB(t)

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
	db := testutil.NewTestDB(t)

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
	db := testutil.NewTestDBDirect(t)

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

// TestBeforeDeleteHook_EmptyStructWithWhere is the regression guard for the
// bug where deleting via an empty struct + WHERE clause (the pattern used by
// gift cards, merchants, and all *_shares) wrote uuid.Nil as resource_id,
// making the audit entry impossible to restore. The hook must read the real id
// from the targeted row, not from the zero-value Dest struct.
func TestBeforeDeleteHook_EmptyStructWithWhere(t *testing.T) {
	db := testutil.NewTestDBDirect(t)
	require.NoError(t, SetupAuditHooks(db))

	merchant := &models.Merchant{Name: "Test Merchant"}
	db.Create(merchant)

	user := &models.User{Email: "gc@example.com", PasswordHash: "h", FirstName: "G", LastName: "C"}
	db.Create(user)

	giftCard := &models.GiftCard{
		UserID:         &user.ID,
		MerchantID:     &merchant.ID,
		MerchantName:   "Test Merchant",
		CardNumber:     "GC-123",
		BarcodeType:    "CODE128",
		Status:         "active",
		Currency:       "CHF",
		InitialBalance: 100,
	}
	db.Create(giftCard)

	ctx := AddAuditContextToContext(context.Background(), user.ID, "10.0.0.1", "TestAgent")

	// The bug reproduction: empty struct + WHERE (mirrors GiftCardRepository.Delete).
	err := db.WithContext(ctx).Delete(&models.GiftCard{}, "id = ?", giftCard.ID).Error
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	var auditLog models.AuditLog
	err = db.Where("resource_type = ? AND action = ?", "gift_cards", "delete").First(&auditLog).Error
	require.NoError(t, err)

	// The real id must be captured, not uuid.Nil.
	assert.Equal(t, giftCard.ID, auditLog.ResourceID)
	assert.NotEqual(t, uuid.Nil, auditLog.ResourceID)

	// Resource data must carry the real row, not a zero-value struct.
	var parsed map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(auditLog.ResourceData), &parsed))
	assert.Equal(t, "Test Merchant", parsed["merchant_name"])
}

// TestBeforeDeleteHook_BulkDelete verifies that a WHERE clause matching multiple
// rows produces one audit entry per row (with its own real id), not a single
// nil-id entry.
func TestBeforeDeleteHook_BulkDelete(t *testing.T) {
	db := testutil.NewTestDBDirect(t)
	require.NoError(t, SetupAuditHooks(db))

	owner := &models.User{Email: "owner@example.com", PasswordHash: "h", FirstName: "O", LastName: "W"}
	db.Create(owner)
	u1 := &models.User{Email: "u1@example.com", PasswordHash: "h", FirstName: "U", LastName: "1"}
	db.Create(u1)
	u2 := &models.User{Email: "u2@example.com", PasswordHash: "h", FirstName: "U", LastName: "2"}
	db.Create(u2)

	voucher := &models.Voucher{
		UserID:       &owner.ID,
		MerchantName: "M",
		Code:         "VCODE-1",
		Type:         "percentage",
		Value:        10,
		ValidFrom:    time.Now(),
		ValidUntil:   time.Now().Add(24 * time.Hour),
		BarcodeType:  "CODE128",
	}
	db.Create(voucher)

	s1 := &models.VoucherShare{VoucherID: voucher.ID, SharedWithID: u1.ID}
	s2 := &models.VoucherShare{VoucherID: voucher.ID, SharedWithID: u2.ID}
	db.Create(s1)
	db.Create(s2)

	ctx := AddAuditContextToContext(context.Background(), owner.ID, "10.0.0.1", "TestAgent")

	// Bulk delete both shares (mirrors TransferRepository share cleanup).
	err := db.WithContext(ctx).Where("voucher_id = ?", voucher.ID).Delete(&models.VoucherShare{}).Error
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	var logs []models.AuditLog
	err = db.Where("resource_type = ? AND action = ?", "voucher_shares", "delete").Find(&logs).Error
	require.NoError(t, err)

	// One audit entry per deleted share, each with a real id.
	require.Len(t, logs, 2)
	ids := map[uuid.UUID]bool{}
	for _, l := range logs {
		assert.NotEqual(t, uuid.Nil, l.ResourceID)
		ids[l.ResourceID] = true
	}
	assert.True(t, ids[s1.ID])
	assert.True(t, ids[s2.ID])
}

// TestBeforeDeleteHook_InsideTransaction guards against the re-select running on
// a different connection than the in-flight DELETE. The real transfer path
// deletes shares inside db.Transaction(...); the hook's re-select must see the
// (not-yet-deleted) rows on the transaction's own connection and the audit
// entry must survive the commit. If the re-select grabbed a fresh pool
// connection outside the tx, it would read zero rows and write no audit entry.
func TestBeforeDeleteHook_InsideTransaction(t *testing.T) {
	db := testutil.NewTestDBDirect(t)
	require.NoError(t, SetupAuditHooks(db))

	owner := &models.User{Email: "txowner@example.com", PasswordHash: "h", FirstName: "T", LastName: "X"}
	db.Create(owner)
	u1 := &models.User{Email: "txu1@example.com", PasswordHash: "h", FirstName: "U", LastName: "1"}
	db.Create(u1)

	voucher := &models.Voucher{
		UserID:       &owner.ID,
		MerchantName: "M",
		Code:         "TXCODE-1",
		Type:         "percentage",
		Value:        10,
		ValidFrom:    time.Now(),
		ValidUntil:   time.Now().Add(24 * time.Hour),
		BarcodeType:  "CODE128",
	}
	db.Create(voucher)
	share := &models.VoucherShare{VoucherID: voucher.ID, SharedWithID: u1.ID}
	db.Create(share)

	ctx := AddAuditContextToContext(context.Background(), owner.ID, "10.0.0.1", "TestAgent")

	// Delete inside an explicit transaction, mirroring TransferRepository.
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Where("voucher_id = ?", voucher.ID).Delete(&models.VoucherShare{}).Error
	})
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	var auditLog models.AuditLog
	err = db.Where("resource_type = ? AND resource_id = ?", "voucher_shares", share.ID).First(&auditLog).Error
	require.NoError(t, err)
	assert.Equal(t, share.ID, auditLog.ResourceID)
	assert.NotEqual(t, uuid.Nil, auditLog.ResourceID)
}

func TestBeforeDeleteHook_SkipsAuditLogs(t *testing.T) {
	db := testutil.NewTestDBDirect(t)

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
	db := testutil.NewTestDBDirect(t)

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
