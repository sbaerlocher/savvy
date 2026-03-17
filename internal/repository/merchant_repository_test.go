package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"savvy/internal/models"
	"savvy/internal/testutil"
)

// setupTestDB returns a transaction-isolated test database.
// Every test gets its own transaction that is rolled back automatically.
func setupTestDB(t *testing.T) *gorm.DB {
	return testutil.NewTestDB(t)
}

// createTestUser creates a test user for foreign key relationships.
// No cleanup needed — the transaction rollback handles everything.
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


	merchant2 := &models.Merchant{
		Name:  "Test Exact Name 2",
		Color: "#222222",
	}
	db.Create(merchant2)


	// Get by exact name should return first one only
	found, err := repo.GetByName(ctx, "Test Exact Name")
	assert.NoError(t, err)
	assert.Equal(t, merchant1.ID, found.ID)
	assert.Equal(t, "Test Exact Name", found.Name)
	assert.NotEqual(t, merchant2.ID, found.ID)
}
