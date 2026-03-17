package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"savvy/internal/models"
)

func TestUserRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &models.User{
		Email:        "test-create@example.com",
		PasswordHash: "hashed_password_123",
		FirstName:    "Test",
		LastName:     "Create",
	}

	err := repo.Create(ctx, user)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, user.ID)

	// Verify it was created
	var found models.User
	err = db.First(&found, "id = ?", user.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "test-create@example.com", found.Email)
	assert.Equal(t, "Test", found.FirstName)
	assert.Equal(t, "Create", found.LastName)

}

func TestUserRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &models.User{
		Email:        "test-getbyid@example.com",
		PasswordHash: "hashed_password",
		FirstName:    "GetByID",
		LastName:     "User",
	}
	db.Create(user)

	found, err := repo.GetByID(ctx, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
	assert.Equal(t, "test-getbyid@example.com", found.Email)
	assert.Equal(t, "GetByID", found.FirstName)
	assert.Equal(t, "User", found.LastName)
}

func TestUserRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, uuid.New())
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &models.User{
		Email:        "test-getbyemail@example.com",
		PasswordHash: "hashed_password",
		FirstName:    "GetByEmail",
		LastName:     "User",
	}
	db.Create(user)

	found, err := repo.GetByEmail(ctx, "test-getbyemail@example.com")
	assert.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
	assert.Equal(t, "test-getbyemail@example.com", found.Email)
}

func TestUserRepository_GetByEmail_CaseInsensitive(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &models.User{
		Email:        "test-case@example.com",
		PasswordHash: "hashed_password",
		FirstName:    "Case",
		LastName:     "User",
	}
	db.Create(user)

	// Test uppercase
	found, err := repo.GetByEmail(ctx, "TEST-CASE@EXAMPLE.COM")
	assert.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)

	// Test mixed case
	found, err = repo.GetByEmail(ctx, "Test-Case@Example.Com")
	assert.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
}

func TestUserRepository_GetByEmail_WithWhitespace(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &models.User{
		Email:        "test-whitespace@example.com",
		PasswordHash: "hashed_password",
		FirstName:    "Whitespace",
		LastName:     "User",
	}
	db.Create(user)

	// Test with leading/trailing whitespace
	found, err := repo.GetByEmail(ctx, "  test-whitespace@example.com  ")
	assert.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
}

func TestUserRepository_GetByEmail_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	_, err := repo.GetByEmail(ctx, "nonexistent@example.com")
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestUserRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &models.User{
		Email:        "test-update@example.com",
		PasswordHash: "original_password",
		FirstName:    "Original",
		LastName:     "Name",
	}
	db.Create(user)

	// Update user
	user.FirstName = "Updated"
	user.LastName = "User"
	user.PasswordHash = "updated_password"
	err := repo.Update(ctx, user)
	assert.NoError(t, err)

	// Verify update
	var found models.User
	db.First(&found, "id = ?", user.ID)
	assert.Equal(t, "Updated", found.FirstName)
	assert.Equal(t, "User", found.LastName)
	assert.Equal(t, "updated_password", found.PasswordHash)
	assert.Equal(t, "test-update@example.com", found.Email)
}

func TestUserRepository_Update_EmailChange(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &models.User{
		Email:        "test-original@example.com",
		PasswordHash: "hashed_password",
		FirstName:    "Test",
		LastName:     "User",
	}
	db.Create(user)

	// Update email
	user.Email = "test-updated@example.com"
	err := repo.Update(ctx, user)
	assert.NoError(t, err)

	// Verify email was updated
	var found models.User
	db.First(&found, "id = ?", user.ID)
	assert.Equal(t, "test-updated@example.com", found.Email)
}
