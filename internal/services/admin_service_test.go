package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"savvy/internal/models"
	"savvy/internal/repository"
)

// TestAdminService_GetAllUsers tests retrieving all users
func TestAdminService_GetAllUsers(t *testing.T) {
	db := setupTestDB(t)
	service := NewAdminService(repository.NewUserRepository(db), repository.NewAuditLogRepository(db))
	ctx := context.Background()

	// Create test users
	user1 := &models.User{Email: "user1@example.com", PasswordHash: "hash1", FirstName: "Test", LastName: "User", Role: "user"}
	user2 := &models.User{Email: "user2@example.com", PasswordHash: "hash2", FirstName: "Test", LastName: "User", Role: "admin"}
	db.Create(user1)
	time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	db.Create(user2)

	// Test: Get all users
	users, err := service.GetAllUsers(ctx)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(users), 2)
	// Should be ordered by created_at DESC
	assert.Equal(t, user2.Email, users[0].Email)
}

// TestAdminService_GetUserByID tests retrieving a user by ID
func TestAdminService_GetUserByID(t *testing.T) {
	db := setupTestDB(t)
	service := NewAdminService(repository.NewUserRepository(db), repository.NewAuditLogRepository(db))
	ctx := context.Background()

	// Create test user
	user := &models.User{Email: "test@example.com", PasswordHash: "hash", FirstName: "Test", LastName: "User", Role: "user"}
	db.Create(user)

	// Test: Get user by ID
	result, err := service.GetUserByID(ctx, user.ID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, user.Email, result.Email)
}

// TestAdminService_GetUserByID_NotFound tests error when user doesn't exist
func TestAdminService_GetUserByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	service := NewAdminService(repository.NewUserRepository(db), repository.NewAuditLogRepository(db))
	ctx := context.Background()

	// Test: Non-existent user
	nonExistentID := uuid.New()
	result, err := service.GetUserByID(ctx, nonExistentID)

	assert.Error(t, err)
	assert.Nil(t, result)
}

// TestAdminService_UpdateUserRole tests updating a user's role
func TestAdminService_UpdateUserRole(t *testing.T) {
	db := setupTestDB(t)
	service := NewAdminService(repository.NewUserRepository(db), repository.NewAuditLogRepository(db))
	ctx := context.Background()

	// Create test user
	user := &models.User{Email: "test@example.com", PasswordHash: "hash", FirstName: "Test", LastName: "User", Role: "user"}
	db.Create(user)

	// Test: Update role to admin
	err := service.UpdateUserRole(ctx, user.ID, "admin")

	assert.NoError(t, err)

	// Verify change
	var updated models.User
	db.First(&updated, user.ID)
	assert.Equal(t, "admin", updated.Role)
}

// TestAdminService_UpdateUserRole_InvalidRole tests validation of invalid role
func TestAdminService_UpdateUserRole_InvalidRole(t *testing.T) {
	db := setupTestDB(t)
	service := NewAdminService(repository.NewUserRepository(db), repository.NewAuditLogRepository(db))
	ctx := context.Background()

	// Create test user
	user := &models.User{Email: "test@example.com", PasswordHash: "hash", FirstName: "Test", LastName: "User", Role: "user"}
	db.Create(user)

	// Test: Invalid role
	err := service.UpdateUserRole(ctx, user.ID, "superuser")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid role")
}

// TestAdminService_UpdateUserRole_UserNotFound tests error when user doesn't exist
func TestAdminService_UpdateUserRole_UserNotFound(t *testing.T) {
	db := setupTestDB(t)
	service := NewAdminService(repository.NewUserRepository(db), repository.NewAuditLogRepository(db))
	ctx := context.Background()

	// Test: Non-existent user
	nonExistentID := uuid.New()
	err := service.UpdateUserRole(ctx, nonExistentID, "admin")

	assert.Error(t, err)
}

// TestAdminService_UpdateUser tests updating complete user profile
func TestAdminService_UpdateUser(t *testing.T) {
	db := setupTestDB(t)
	service := NewAdminService(repository.NewUserRepository(db), repository.NewAuditLogRepository(db))
	ctx := context.Background()

	// Create test user
	user := &models.User{
		Email:        "old@example.com",
		FirstName:    "Old",
		LastName:     "Name",
		Role:         "user",
		PasswordHash: "hash",
	}
	db.Create(user)

	// Test: Update all fields
	err := service.UpdateUser(ctx, user.ID, "new@example.com", "New", "Name", "admin")

	assert.NoError(t, err)

	// Verify changes
	var updated models.User
	db.First(&updated, user.ID)
	assert.Equal(t, "new@example.com", updated.Email)
	assert.Equal(t, "New", updated.FirstName)
	assert.Equal(t, "Name", updated.LastName)
	assert.Equal(t, "admin", updated.Role)
}

// TestAdminService_UpdateUser_InvalidRole tests validation
func TestAdminService_UpdateUser_InvalidRole(t *testing.T) {
	db := setupTestDB(t)
	service := NewAdminService(repository.NewUserRepository(db), repository.NewAuditLogRepository(db))
	ctx := context.Background()

	// Create test user
	user := &models.User{Email: "test@example.com", PasswordHash: "hash", FirstName: "Test", LastName: "User", Role: "user"}
	db.Create(user)

	// Test: Invalid role
	err := service.UpdateUser(ctx, user.ID, "test@example.com", "First", "Last", "invalid")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid role")
}

// TestAdminService_CreateLocalUser tests creating a local auth user
func TestAdminService_CreateLocalUser(t *testing.T) {
	db := setupTestDB(t)
	service := NewAdminService(repository.NewUserRepository(db), repository.NewAuditLogRepository(db))
	ctx := context.Background()

	// Test: Create local user
	user := &models.User{
		Email:        "local@example.com",
		PasswordHash: "hash",
		AuthProvider: "local",
		Role:         "user",
	}

	err := service.CreateLocalUser(ctx, user)

	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, user.ID)
}

// TestAdminService_CreateLocalUser_NonLocalAuth tests validation
func TestAdminService_CreateLocalUser_NonLocalAuth(t *testing.T) {
	db := setupTestDB(t)
	service := NewAdminService(repository.NewUserRepository(db), repository.NewAuditLogRepository(db))
	ctx := context.Background()

	// Test: Try to create OAuth user
	user := &models.User{
		Email:        "oauth@example.com",
		AuthProvider: "google",
		Role:         "user",
	}

	err := service.CreateLocalUser(ctx, user)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "can only create local auth users")
}

// TestAdminService_GetAuditLogs tests retrieving audit logs with filters
func TestAdminService_GetAuditLogs(t *testing.T) {
	db := setupTestDB(t)
	service := NewAdminService(repository.NewUserRepository(db), repository.NewAuditLogRepository(db))
	ctx := context.Background()

	// Create test user
	user := &models.User{Email: "test@example.com", PasswordHash: "hash", FirstName: "Test", LastName: "User", Role: "admin"}
	db.Create(user)

	// Create test audit logs
	log1 := &models.AuditLog{
		UserID:       &user.ID,
		Action:       "delete",
		ResourceType: "cards",
		ResourceID:   uuid.New(),
		ResourceData: "{}",
	}
	log2 := &models.AuditLog{
		UserID:       &user.ID,
		Action:       "hard_delete",
		ResourceType: "vouchers",
		ResourceID:   uuid.New(),
		ResourceData: "{}",
	}
	db.Create(log1)
	time.Sleep(10 * time.Millisecond)
	db.Create(log2)

	// Test: Get all logs
	filters := repository.AuditLogFilters{
		Page:    1,
		PerPage: 10,
	}
	result, err := service.GetAuditLogs(ctx, filters)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.GreaterOrEqual(t, len(result.Logs), 2)
	assert.GreaterOrEqual(t, result.Total, int64(2))
}

// TestAdminService_GetAuditLogs_FilterByUser tests filtering by user
func TestAdminService_GetAuditLogs_FilterByUser(t *testing.T) {
	db := setupTestDB(t)
	service := NewAdminService(repository.NewUserRepository(db), repository.NewAuditLogRepository(db))
	ctx := context.Background()

	// Create test users
	user1 := &models.User{Email: "user1@example.com", PasswordHash: "hash", FirstName: "Test", LastName: "User", Role: "admin"}
	user2 := &models.User{Email: "user2@example.com", PasswordHash: "hash", FirstName: "Test", LastName: "User", Role: "admin"}
	db.Create(user1)
	db.Create(user2)

	// Create audit logs for different users
	log1 := &models.AuditLog{UserID: &user1.ID, Action: "delete", ResourceType: "cards", ResourceID: uuid.New(), ResourceData: "{}"}
	log2 := &models.AuditLog{UserID: &user2.ID, Action: "delete", ResourceType: "cards", ResourceID: uuid.New(), ResourceData: "{}"}
	db.Create(log1)
	db.Create(log2)

	// Test: Filter by user1
	filters := repository.AuditLogFilters{
		UserID:  &user1.ID,
		Page:    1,
		PerPage: 10,
	}
	result, err := service.GetAuditLogs(ctx, filters)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	// Should only contain user1's logs
	for _, log := range result.Logs {
		assert.Equal(t, user1.ID, *log.UserID)
	}
}

// TestAdminService_GetAuditLogs_FilterByResourceType tests filtering by resource type
func TestAdminService_GetAuditLogs_FilterByResourceType(t *testing.T) {
	db := setupTestDB(t)
	service := NewAdminService(repository.NewUserRepository(db), repository.NewAuditLogRepository(db))
	ctx := context.Background()

	// Create test user
	user := &models.User{Email: "test@example.com", PasswordHash: "hash", FirstName: "Test", LastName: "User", Role: "admin"}
	db.Create(user)

	// Create audit logs for different resource types
	log1 := &models.AuditLog{UserID: &user.ID, Action: "delete", ResourceType: "cards", ResourceID: uuid.New(), ResourceData: "{}"}
	log2 := &models.AuditLog{UserID: &user.ID, Action: "delete", ResourceType: "vouchers", ResourceID: uuid.New(), ResourceData: "{}"}
	db.Create(log1)
	db.Create(log2)

	// Test: Filter by cards
	filters := repository.AuditLogFilters{
		ResourceType: "cards",
		Page:         1,
		PerPage:      10,
	}
	result, err := service.GetAuditLogs(ctx, filters)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	// Should only contain cards logs
	for _, log := range result.Logs {
		assert.Equal(t, "cards", log.ResourceType)
	}
}

// TestAdminService_GetAuditLogs_FilterByAction tests filtering by action
func TestAdminService_GetAuditLogs_FilterByAction(t *testing.T) {
	db := setupTestDB(t)
	service := NewAdminService(repository.NewUserRepository(db), repository.NewAuditLogRepository(db))
	ctx := context.Background()

	// Create test user
	user := &models.User{Email: "test@example.com", PasswordHash: "hash", FirstName: "Test", LastName: "User", Role: "admin"}
	db.Create(user)

	// Create audit logs with different actions
	log1 := &models.AuditLog{UserID: &user.ID, Action: "delete", ResourceType: "cards", ResourceID: uuid.New(), ResourceData: "{}"}
	log2 := &models.AuditLog{UserID: &user.ID, Action: "restore", ResourceType: "cards", ResourceID: uuid.New(), ResourceData: "{}"}
	db.Create(log1)
	db.Create(log2)

	// Test: Filter by delete action
	filters := repository.AuditLogFilters{
		Action:  "delete",
		Page:    1,
		PerPage: 10,
	}
	result, err := service.GetAuditLogs(ctx, filters)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	// Should only contain delete logs
	for _, log := range result.Logs {
		assert.Equal(t, "delete", log.Action)
	}
}

// TestAdminService_GetAuditLogs_Pagination tests pagination
func TestAdminService_GetAuditLogs_Pagination(t *testing.T) {
	db := setupTestDB(t)
	service := NewAdminService(repository.NewUserRepository(db), repository.NewAuditLogRepository(db))
	ctx := context.Background()

	// Create test user
	user := &models.User{Email: "test@example.com", PasswordHash: "hash", FirstName: "Test", LastName: "User", Role: "admin"}
	db.Create(user)

	// Create multiple audit logs
	for i := 0; i < 15; i++ {
		log := &models.AuditLog{
			UserID:       &user.ID,
			Action:       "delete",
			ResourceType: "cards",
			ResourceID:   uuid.New(),
			ResourceData: "{}",
		}
		db.Create(log)
		time.Sleep(1 * time.Millisecond)
	}

	// Test: Get first page
	filters := repository.AuditLogFilters{
		Page:    1,
		PerPage: 10,
	}
	result, err := service.GetAuditLogs(ctx, filters)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.LessOrEqual(t, len(result.Logs), 10)
	assert.GreaterOrEqual(t, result.Total, int64(15))

	// Test: Get second page
	filters.Page = 2
	result2, err := service.GetAuditLogs(ctx, filters)

	assert.NoError(t, err)
	assert.NotNil(t, result2)
	assert.GreaterOrEqual(t, len(result2.Logs), 5)
}

// TestAdminService_CreateAuditLog tests creating audit log entry
func TestAdminService_CreateAuditLog(t *testing.T) {
	db := setupTestDB(t)
	service := NewAdminService(repository.NewUserRepository(db), repository.NewAuditLogRepository(db))
	ctx := context.Background()

	// Create test user
	user := &models.User{Email: "test@example.com", PasswordHash: "hash", FirstName: "Test", LastName: "User", Role: "admin"}
	db.Create(user)

	// Test: Create audit log
	resourceID := uuid.New()
	log := &models.AuditLog{
		UserID:       &user.ID,
		Action:       "delete",
		ResourceType: "cards",
		ResourceID:   resourceID,
		ResourceData: "{}",
	}

	err := service.CreateAuditLog(ctx, log)

	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, log.ID)

	// Verify it was created
	var saved models.AuditLog
	db.First(&saved, log.ID)
	assert.Equal(t, resourceID, saved.ResourceID)
}

// TestAdminService_RestoreResource_Card tests restoring a deleted card
func TestAdminService_RestoreResource_Card(t *testing.T) {
	db := setupTestDB(t)
	service := NewAdminService(repository.NewUserRepository(db), repository.NewAuditLogRepository(db))
	ctx := context.Background()

	// Create user first (required for foreign key)
	user := &models.User{Email: "cardowner@test.com", PasswordHash: "hash", FirstName: "Test", LastName: "User", Role: "user"}
	db.Create(user)

	// Create and soft-delete a card
	card := &models.Card{
		UserID:       &user.ID,
		CardNumber:   "12345",
		MerchantName: "Test",
	}
	result := db.Create(card)
	if result.Error != nil {
		t.Fatalf("Failed to create card: %v", result.Error)
	}
	assert.NotEqual(t, uuid.Nil, card.ID, "Card ID should be set after creation")

	db.Delete(card) // Soft delete

	// Verify it's deleted (requires Unscoped to find soft-deleted records)
	var check models.Card
	err := db.First(&check, card.ID).Error
	assert.Error(t, err) // Should not be found without Unscoped

	// Test: Restore card
	err = service.RestoreResource(ctx, "cards", card.ID)

	assert.NoError(t, err)

	// Verify it's restored
	err = db.First(&check, card.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "12345", check.CardNumber)
}

// TestAdminService_RestoreResource_NotDeleted tests error when resource is not deleted
func TestAdminService_RestoreResource_NotDeleted(t *testing.T) {
	db := setupTestDB(t)
	service := NewAdminService(repository.NewUserRepository(db), repository.NewAuditLogRepository(db))
	ctx := context.Background()

	// Create user first (required for foreign key)
	user := &models.User{Email: "cardowner2@test.com", PasswordHash: "hash", FirstName: "Test", LastName: "User", Role: "user"}
	db.Create(user)

	// Create active card (not deleted)
	card := &models.Card{
		UserID:       &user.ID,
		CardNumber:   "12345",
		MerchantName: "Test",
	}
	result := db.Create(card)
	if result.Error != nil {
		t.Fatalf("Failed to create card: %v", result.Error)
	}
	assert.NotEqual(t, uuid.Nil, card.ID)

	// Test: Try to restore active resource
	err := service.RestoreResource(ctx, "cards", card.ID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "resource is not deleted")
}

// TestAdminService_RestoreResource_UnsupportedType tests error for unsupported resource type
func TestAdminService_RestoreResource_UnsupportedType(t *testing.T) {
	db := setupTestDB(t)
	service := NewAdminService(repository.NewUserRepository(db), repository.NewAuditLogRepository(db))
	ctx := context.Background()

	// Test: Unsupported resource type
	err := service.RestoreResource(ctx, "unsupported", uuid.New())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported resource type")
}

// TestAdminService_RestoreResource_Voucher tests restoring a deleted voucher
func TestAdminService_RestoreResource_Voucher(t *testing.T) {
	db := setupTestDB(t)
	service := NewAdminService(repository.NewUserRepository(db), repository.NewAuditLogRepository(db))
	ctx := context.Background()

	// Create user first (required for foreign key)
	user := &models.User{Email: "voucherowner@test.com", PasswordHash: "hash", FirstName: "Test", LastName: "User", Role: "user"}
	db.Create(user)

	// Create and soft-delete a voucher
	voucher := &models.Voucher{
		UserID:         &user.ID,
		Code:           "TEST123",
		MerchantName:   "Test",
		ValidFrom:      time.Now(),
		ValidUntil:     time.Now().Add(24 * time.Hour),
		UsageLimitType: "single_use",
	}
	result := db.Create(voucher)
	if result.Error != nil {
		t.Fatalf("Failed to create voucher: %v", result.Error)
	}
	assert.NotEqual(t, uuid.Nil, voucher.ID)

	db.Delete(voucher) // Soft delete

	// Test: Restore voucher
	err := service.RestoreResource(ctx, "vouchers", voucher.ID)

	assert.NoError(t, err)

	// Verify it's restored
	var check models.Voucher
	err = db.First(&check, voucher.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "TEST123", check.Code)
}

// TestAdminService_ValidateImpersonation tests impersonation validation
func TestAdminService_ValidateImpersonation(t *testing.T) {
	db := setupTestDB(t)
	service := NewAdminService(repository.NewUserRepository(db), repository.NewAuditLogRepository(db))
	ctx := context.Background()

	// Create admin and regular user
	admin := &models.User{Email: "admin@example.com", PasswordHash: "hash", FirstName: "Test", LastName: "User", Role: "admin"}
	user := &models.User{Email: "user@example.com", PasswordHash: "hash", FirstName: "Test", LastName: "User", Role: "user"}
	db.Create(admin)
	db.Create(user)

	// Test: Valid impersonation
	err := service.ValidateImpersonation(ctx, admin.ID, user.ID)

	assert.NoError(t, err)
}

// TestAdminService_ValidateImpersonation_NonAdmin tests error when impersonator is not admin
func TestAdminService_ValidateImpersonation_NonAdmin(t *testing.T) {
	db := setupTestDB(t)
	service := NewAdminService(repository.NewUserRepository(db), repository.NewAuditLogRepository(db))
	ctx := context.Background()

	// Create two regular users
	user1 := &models.User{Email: "user1@example.com", PasswordHash: "hash", FirstName: "Test", LastName: "User", Role: "user"}
	user2 := &models.User{Email: "user2@example.com", PasswordHash: "hash", FirstName: "Test", LastName: "User", Role: "user"}
	db.Create(user1)
	db.Create(user2)

	// Test: Non-admin trying to impersonate
	err := service.ValidateImpersonation(ctx, user1.ID, user2.ID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "only admins can impersonate")
}

// TestAdminService_ValidateImpersonation_SelfImpersonation tests error for self-impersonation
func TestAdminService_ValidateImpersonation_SelfImpersonation(t *testing.T) {
	db := setupTestDB(t)
	service := NewAdminService(repository.NewUserRepository(db), repository.NewAuditLogRepository(db))
	ctx := context.Background()

	// Create admin
	admin := &models.User{Email: "admin@example.com", PasswordHash: "hash", FirstName: "Test", LastName: "User", Role: "admin"}
	db.Create(admin)

	// Test: Admin trying to impersonate themselves
	err := service.ValidateImpersonation(ctx, admin.ID, admin.ID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot impersonate yourself")
}

// TestAdminService_ValidateImpersonation_AdminNotFound tests error when admin doesn't exist
func TestAdminService_ValidateImpersonation_AdminNotFound(t *testing.T) {
	db := setupTestDB(t)
	service := NewAdminService(repository.NewUserRepository(db), repository.NewAuditLogRepository(db))
	ctx := context.Background()

	// Create target user
	user := &models.User{Email: "user@example.com", PasswordHash: "hash", FirstName: "Test", LastName: "User", Role: "user"}
	db.Create(user)

	// Test: Non-existent admin
	nonExistentID := uuid.New()
	err := service.ValidateImpersonation(ctx, nonExistentID, user.ID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "admin not found")
}

// TestAdminService_ValidateImpersonation_TargetNotFound tests error when target user doesn't exist
func TestAdminService_ValidateImpersonation_TargetNotFound(t *testing.T) {
	db := setupTestDB(t)
	service := NewAdminService(repository.NewUserRepository(db), repository.NewAuditLogRepository(db))
	ctx := context.Background()

	// Create admin
	admin := &models.User{Email: "admin@example.com", PasswordHash: "hash", FirstName: "Test", LastName: "User", Role: "admin"}
	db.Create(admin)

	// Test: Non-existent target user
	nonExistentID := uuid.New()
	err := service.ValidateImpersonation(ctx, admin.ID, nonExistentID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "target user not found")
}

// TestAdminService_StartImpersonation tests creating impersonation start audit log
func TestAdminService_StartImpersonation(t *testing.T) {
	db := setupTestDB(t)
	service := NewAdminService(repository.NewUserRepository(db), repository.NewAuditLogRepository(db))
	ctx := context.Background()

	// Create admin and target user
	admin := &models.User{Email: "admin@example.com", PasswordHash: "hash", FirstName: "Test", LastName: "User", Role: "admin"}
	user := &models.User{Email: "user@example.com", PasswordHash: "hash", FirstName: "Test", LastName: "User", Role: "user"}
	db.Create(admin)
	db.Create(user)

	// Test: Start impersonation
	resourceData := map[string]interface{}{
		"target_user_email": user.Email,
		"target_user_role":  user.Role,
	}
	err := service.StartImpersonation(ctx, admin.ID, user.ID, resourceData)

	assert.NoError(t, err)

	// Verify audit log was created
	var log models.AuditLog
	err = db.Where("user_id = ? AND action = ? AND resource_id = ?",
		admin.ID, "impersonate_start", user.ID).First(&log).Error
	assert.NoError(t, err)
	assert.Equal(t, "users", log.ResourceType)
	assert.Contains(t, log.ResourceData, "target_user_email")
}

// TestAdminService_StopImpersonation tests creating impersonation stop audit log
func TestAdminService_StopImpersonation(t *testing.T) {
	db := setupTestDB(t)
	service := NewAdminService(repository.NewUserRepository(db), repository.NewAuditLogRepository(db))
	ctx := context.Background()

	// Create admin and target user
	admin := &models.User{Email: "admin@example.com", PasswordHash: "hash", FirstName: "Test", LastName: "User", Role: "admin"}
	user := &models.User{Email: "user@example.com", PasswordHash: "hash", FirstName: "Test", LastName: "User", Role: "user"}
	db.Create(admin)
	db.Create(user)

	// Test: Stop impersonation
	resourceData := map[string]interface{}{
		"target_user_email": user.Email,
		"target_user_role":  user.Role,
	}
	err := service.StopImpersonation(ctx, admin.ID, user.ID, resourceData)

	assert.NoError(t, err)

	// Verify audit log was created
	var log models.AuditLog
	err = db.Where("user_id = ? AND action = ? AND resource_id = ?",
		admin.ID, "impersonate_stop", user.ID).First(&log).Error
	assert.NoError(t, err)
	assert.Equal(t, "users", log.ResourceType)
	assert.Contains(t, log.ResourceData, "target_user_email")
}
