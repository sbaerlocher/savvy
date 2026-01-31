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
		dbURL = "postgres://savvy:savvy_dev_password@localhost:5432/savvy?sslmode=disable"
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
		&models.Voucher{},
		&models.GiftCard{},
		&models.UserFavorite{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// Clean up test data
	db.Exec("DELETE FROM merchants WHERE name LIKE 'Test%'")
	db.Exec("DELETE FROM cards WHERE card_number LIKE 'TEST%'")
	db.Exec("DELETE FROM vouchers WHERE code LIKE 'TEST%'")
	db.Exec("DELETE FROM gift_cards WHERE card_number LIKE 'TEST%'")
	db.Exec("DELETE FROM gift_cards WHERE card_number LIKE 'GIFT%'")

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
	}
	db.Create(user)
	t.Cleanup(func() {
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
