package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"savvy/internal/models"
)

func TestFavoriteRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewFavoriteRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db)
	resourceID := uuid.New()
	favorite := &models.UserFavorite{
		UserID:       userID,
		ResourceType: "card",
		ResourceID:   resourceID,
	}

	err := repo.Create(ctx, favorite)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, favorite.ID)

	db.Exec("DELETE FROM user_favorites WHERE id = ?", favorite.ID)
}

func TestFavoriteRepository_GetByUser(t *testing.T) {
	db := setupTestDB(t)
	repo := NewFavoriteRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db)
	favorites := []models.UserFavorite{
		{UserID: userID, ResourceType: "card", ResourceID: uuid.New()},
		{UserID: userID, ResourceType: "voucher", ResourceID: uuid.New()},
	}
	for i := range favorites {
		db.Create(&favorites[i])
		defer db.Exec("DELETE FROM user_favorites WHERE id = ?", favorites[i].ID)
	}

	found, err := repo.GetByUser(ctx, userID)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(found), 2)
}

func TestFavoriteRepository_GetByUserAndResource(t *testing.T) {
	db := setupTestDB(t)
	repo := NewFavoriteRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db)
	resourceID := uuid.New()
	favorite := &models.UserFavorite{
		UserID:       userID,
		ResourceType: "card",
		ResourceID:   resourceID,
	}
	db.Create(favorite)
	defer db.Exec("DELETE FROM user_favorites WHERE id = ?", favorite.ID)

	found, err := repo.GetByUserAndResource(ctx, userID, "card", resourceID)
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, userID, found.UserID)
	assert.Equal(t, resourceID, found.ResourceID)
}

func TestFavoriteRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewFavoriteRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db)
	resourceID := uuid.New()
	favorite := &models.UserFavorite{
		UserID:       userID,
		ResourceType: "card",
		ResourceID:   resourceID,
	}
	db.Create(favorite)

	err := repo.Delete(ctx, favorite)
	assert.NoError(t, err)
}

func TestFavoriteRepository_Restore(t *testing.T) {
	db := setupTestDB(t)
	repo := NewFavoriteRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db)
	resourceID := uuid.New()
	favorite := &models.UserFavorite{
		UserID:       userID,
		ResourceType: "card",
		ResourceID:   resourceID,
	}
	db.Create(favorite)
	defer db.Exec("DELETE FROM user_favorites WHERE id = ?", favorite.ID)

	// Soft delete it first
	db.Delete(favorite)

	// Restore it
	err := repo.Restore(ctx, favorite)
	assert.NoError(t, err)

	// Verify it's restored
	var found models.UserFavorite
	err = db.First(&found, "id = ?", favorite.ID).Error
	assert.NoError(t, err)
}

func TestFavoriteRepository_IsFavorite(t *testing.T) {
	db := setupTestDB(t)
	repo := NewFavoriteRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db)
	resourceID := uuid.New()
	favorite := &models.UserFavorite{
		UserID:       userID,
		ResourceType: "card",
		ResourceID:   resourceID,
	}
	db.Create(favorite)
	defer db.Exec("DELETE FROM user_favorites WHERE id = ?", favorite.ID)

	isFav, err := repo.IsFavorite(ctx, userID, "card", resourceID)
	assert.NoError(t, err)
	assert.True(t, isFav)

	// Test non-existent
	isFav, err = repo.IsFavorite(ctx, uuid.New(), "card", uuid.New())
	assert.NoError(t, err)
	assert.False(t, isFav)
}
