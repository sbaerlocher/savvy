package repository

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"savvy/internal/models"
)

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://savvy:savvy_dev_password@localhost:5432/savvy?sslmode=disable" // #nosec G101 -- test credential, not real
	}

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		t.Skipf("Skipping test: PostgreSQL not available: %v", err)
		return nil
	}

	// Auto-migrate models
	err = db.AutoMigrate(
		&models.User{},
		&models.Merchant{},
		&models.Card{},
		&models.CardShare{},
		&models.Voucher{},
		&models.VoucherShare{},
		&models.GiftCard{},
		&models.GiftCardShare{},
		&models.GiftCardTransaction{},
		&models.UserFavorite{},
		&models.AuditLog{},
		&models.Notification{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// Clean up test data in correct order (dependent records first)
	db.Exec("DELETE FROM user_favorites")
	db.Exec("DELETE FROM card_shares")
	db.Exec("DELETE FROM voucher_shares")
	db.Exec("DELETE FROM gift_card_shares")
	db.Exec("DELETE FROM gift_card_transactions")
	db.Exec("DELETE FROM audit_logs")
	db.Exec("DELETE FROM notifications")
	db.Exec("DELETE FROM cards WHERE card_number LIKE 'TEST%' OR card_number LIKE 'PRELOAD%'")
	db.Exec("DELETE FROM vouchers WHERE code LIKE 'TEST%'")
	db.Exec("DELETE FROM gift_cards WHERE card_number LIKE 'TEST%'")
	db.Exec("DELETE FROM gift_cards WHERE card_number LIKE 'GIFT%'")
	db.Exec("DELETE FROM gift_cards WHERE card_number LIKE 'GC%'")
	db.Exec("DELETE FROM merchants WHERE name LIKE 'Test%'")
	db.Exec("DELETE FROM users WHERE email LIKE 'test-%@example.com'")

	return db
}

// createTestUser creates a test user for foreign key relationships
func createTestUser(t *testing.T, db *gorm.DB) uuid.UUID {
	t.Helper()
	userID := uuid.New()
	user := &models.User{
		ID:           userID,
		Email:        "test-" + userID.String()[:8] + "@example.com",
		PasswordHash: "hashed",
		FirstName:    "Test",
		LastName:     "User",
	}
	err := db.Create(user).Error
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	t.Cleanup(func() {
		// Clean up dependent records first
		db.Exec("DELETE FROM user_favorites WHERE user_id = ?", userID)
		db.Exec("DELETE FROM card_shares WHERE shared_with_id = ?", userID)
		db.Exec("DELETE FROM voucher_shares WHERE shared_with_id = ?", userID)
		db.Exec("DELETE FROM gift_card_shares WHERE shared_with_id = ?", userID)
		db.Exec("DELETE FROM notifications WHERE user_id = ?", userID)
		db.Exec("DELETE FROM cards WHERE user_id = ?", userID)
		db.Exec("DELETE FROM vouchers WHERE user_id = ?", userID)
		db.Exec("DELETE FROM gift_cards WHERE user_id = ?", userID)
		db.Exec("DELETE FROM audit_logs WHERE user_id = ?", userID)
		// Now delete the user
		db.Exec("DELETE FROM users WHERE id = ?", userID)
	})
	return userID
}

func TestMerchantRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMerchantRepository(db)
	ctx := context.Background()

	merchant := &models.Merchant{
		Name:  "Test Merchant Create",
		Color: "#FF0000",
	}

	err := repo.Create(ctx, merchant)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, merchant.ID)

	// Verify it was created
	var found models.Merchant
	err = db.First(&found, "id = ?", merchant.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "Test Merchant Create", found.Name)
	assert.Equal(t, "#FF0000", found.Color)

	// Cleanup
	db.Exec("DELETE FROM merchants WHERE id = ?", merchant.ID)
}

func TestMerchantRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMerchantRepository(db)
	ctx := context.Background()

	// Create a test merchant
	merchant := &models.Merchant{
		Name:  "Test Merchant GetByID",
		Color: "#00FF00",
	}
	db.Create(merchant)
	defer db.Exec("DELETE FROM merchants WHERE id = ?", merchant.ID)

	// Retrieve it
	found, err := repo.GetByID(ctx, merchant.ID)
	assert.NoError(t, err)
	assert.Equal(t, merchant.ID, found.ID)
	assert.Equal(t, "Test Merchant GetByID", found.Name)
	assert.Equal(t, "#00FF00", found.Color)
}

func TestMerchantRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMerchantRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, uuid.New())
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestMerchantRepository_GetAll(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMerchantRepository(db)
	ctx := context.Background()

	// Create test merchants
	merchants := []models.Merchant{
		{Name: "Test Merchant All A", Color: "#FF0000"},
		{Name: "Test Merchant All B", Color: "#00FF00"},
	}
	for i := range merchants {
		db.Create(&merchants[i])
		defer db.Exec("DELETE FROM merchants WHERE id = ?", merchants[i].ID)
	}

	// Get all
	found, err := repo.GetAll(ctx)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(found), 2)
}

func TestMerchantRepository_Search(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMerchantRepository(db)
	ctx := context.Background()

	// Create unique test merchants
	merchants := []models.Merchant{
		{Name: "Test Search Apple", Color: "#FF0000"},
		{Name: "Test Search Amazon", Color: "#00FF00"},
	}
	for i := range merchants {
		db.Create(&merchants[i])
		defer db.Exec("DELETE FROM merchants WHERE id = ?", merchants[i].ID)
	}

	// Search for "Apple"
	found, err := repo.Search(ctx, "Apple")
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(found), 1)

	// Verify we found the right one
	foundApple := false
	for _, m := range found {
		if m.Name == "Test Search Apple" {
			foundApple = true
			break
		}
	}
	assert.True(t, foundApple, "Should find Test Search Apple")
}

func TestMerchantRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMerchantRepository(db)
	ctx := context.Background()

	// Create a merchant
	merchant := &models.Merchant{
		Name:  "Test Merchant Update Original",
		Color: "#FF0000",
	}
	db.Create(merchant)
	defer db.Exec("DELETE FROM merchants WHERE id = ?", merchant.ID)

	// Update it
	merchant.Name = "Test Merchant Update Modified"
	merchant.Color = "#00FF00"
	err := repo.Update(ctx, merchant)
	assert.NoError(t, err)

	// Verify update
	var found models.Merchant
	db.First(&found, "id = ?", merchant.ID)
	assert.Equal(t, "Test Merchant Update Modified", found.Name)
	assert.Equal(t, "#00FF00", found.Color)
}

func TestMerchantRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMerchantRepository(db)
	ctx := context.Background()

	// Create a merchant
	merchant := &models.Merchant{
		Name:  "Test Merchant Delete",
		Color: "#FF0000",
	}
	db.Create(merchant)

	// Delete it
	err := repo.Delete(ctx, merchant.ID)
	assert.NoError(t, err)

	// Verify it's deleted (soft delete, so check deleted_at)
	var found models.Merchant
	err = db.First(&found, "id = ?", merchant.ID).Error
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestMerchantRepository_Count(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMerchantRepository(db)
	ctx := context.Background()

	// Get initial count
	initialCount, err := repo.Count(ctx)
	assert.NoError(t, err)

	// Create test merchants
	merchants := []models.Merchant{
		{Name: "Test Merchant Count A", Color: "#FF0000"},
		{Name: "Test Merchant Count B", Color: "#00FF00"},
	}
	for i := range merchants {
		db.Create(&merchants[i])
		defer db.Exec("DELETE FROM merchants WHERE id = ?", merchants[i].ID)
	}

	// Count should increase by 2
	newCount, err := repo.Count(ctx)
	assert.NoError(t, err)
	assert.Equal(t, initialCount+2, newCount)
}

func TestMerchantRepository_GetByName(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMerchantRepository(db)
	ctx := context.Background()

	// Create a test merchant with unique name
	merchant := &models.Merchant{
		Name:  "Test GetByName Unique",
		Color: "#AABBCC",
	}
	db.Create(merchant)
	defer db.Exec("DELETE FROM merchants WHERE id = ?", merchant.ID)

	// Retrieve by name
	found, err := repo.GetByName(ctx, "Test GetByName Unique")
	assert.NoError(t, err)
	assert.Equal(t, merchant.ID, found.ID)
	assert.Equal(t, "Test GetByName Unique", found.Name)
	assert.Equal(t, "#AABBCC", found.Color)
}

func TestMerchantRepository_GetByName_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMerchantRepository(db)
	ctx := context.Background()

	// Try to get non-existent merchant
	_, err := repo.GetByName(ctx, "NonExistent Merchant Name XYZ123")
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestMerchantRepository_GetByName_Exact(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMerchantRepository(db)
	ctx := context.Background()

	// Create test merchants with similar names
	merchant1 := &models.Merchant{
		Name:  "Test Exact Name",
		Color: "#111111",
	}
	db.Create(merchant1)
	defer db.Exec("DELETE FROM merchants WHERE id = ?", merchant1.ID)

	merchant2 := &models.Merchant{
		Name:  "Test Exact Name 2",
		Color: "#222222",
	}
	db.Create(merchant2)
	defer db.Exec("DELETE FROM merchants WHERE id = ?", merchant2.ID)

	// Get by exact name should return first one only
	found, err := repo.GetByName(ctx, "Test Exact Name")
	assert.NoError(t, err)
	assert.Equal(t, merchant1.ID, found.ID)
	assert.Equal(t, "Test Exact Name", found.Name)
	assert.NotEqual(t, merchant2.ID, found.ID)
}
