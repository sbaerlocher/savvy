package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"savvy/internal/models"
)

func TestNotificationRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewNotificationRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db)
	notification := &models.Notification{
		UserID:       userID,
		Type:         models.NotificationTypeShareReceived,
		ResourceType: "card",
		ResourceID:   uuid.New(),
		Metadata: models.NotificationMetadata{
			"from_user_name": "Test User",
			"from_user_id":   uuid.New().String(),
		},
	}

	err := repo.Create(ctx, notification)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, notification.ID)

	// Verify it was created
	var found models.Notification
	err = db.First(&found, "id = ?", notification.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, userID, found.UserID)
	assert.Equal(t, models.NotificationTypeShareReceived, found.Type)
	assert.Equal(t, "card", found.ResourceType)

	// Cleanup
	db.Exec("DELETE FROM notifications WHERE id = ?", notification.ID)
}

func TestNotificationRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewNotificationRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db)
	notification := &models.Notification{
		UserID:       userID,
		Type:         models.NotificationTypeTransferReceived,
		ResourceType: "voucher",
		ResourceID:   uuid.New(),
		Metadata: models.NotificationMetadata{
			"from_user_name": "Sender User",
		},
	}
	db.Create(notification)
	defer db.Exec("DELETE FROM notifications WHERE id = ?", notification.ID)

	found, err := repo.GetByID(ctx, notification.ID)
	assert.NoError(t, err)
	assert.Equal(t, notification.ID, found.ID)
	assert.Equal(t, userID, found.UserID)
	assert.Equal(t, models.NotificationTypeTransferReceived, found.Type)
}

func TestNotificationRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewNotificationRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, uuid.New())
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestNotificationRepository_GetByUserID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewNotificationRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db)
	notifications := []models.Notification{
		{
			UserID:       userID,
			Type:         models.NotificationTypeShareReceived,
			ResourceType: "card",
			ResourceID:   uuid.New(),
			Metadata:     models.NotificationMetadata{"from_user_name": "User1"},
		},
		{
			UserID:       userID,
			Type:         models.NotificationTypeTransferReceived,
			ResourceType: "gift_card",
			ResourceID:   uuid.New(),
			Metadata:     models.NotificationMetadata{"from_user_name": "User2"},
		},
	}
	for i := range notifications {
		db.Create(&notifications[i])
		defer db.Exec("DELETE FROM notifications WHERE id = ?", notifications[i].ID)
	}

	// Test with limit and offset
	found, err := repo.GetByUserID(ctx, userID, 10, 0)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(found), 2)

	// Test pagination with offset
	found, err = repo.GetByUserID(ctx, userID, 1, 0)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(found))

	// Test empty result for non-existent user
	found, err = repo.GetByUserID(ctx, uuid.New(), 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(found))
}

func TestNotificationRepository_GetUnreadCount(t *testing.T) {
	db := setupTestDB(t)
	repo := NewNotificationRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db)

	// Create some read and unread notifications
	notifications := []models.Notification{
		{
			UserID:       userID,
			Type:         models.NotificationTypeShareReceived,
			ResourceType: "card",
			ResourceID:   uuid.New(),
			IsRead:       false,
			Metadata:     models.NotificationMetadata{},
		},
		{
			UserID:       userID,
			Type:         models.NotificationTypeShareReceived,
			ResourceType: "voucher",
			ResourceID:   uuid.New(),
			IsRead:       false,
			Metadata:     models.NotificationMetadata{},
		},
		{
			UserID:       userID,
			Type:         models.NotificationTypeTransferReceived,
			ResourceType: "gift_card",
			ResourceID:   uuid.New(),
			IsRead:       true,
			Metadata:     models.NotificationMetadata{},
		},
	}
	for i := range notifications {
		db.Create(&notifications[i])
		defer db.Exec("DELETE FROM notifications WHERE id = ?", notifications[i].ID)
	}

	// Count unread notifications
	count, err := repo.GetUnreadCount(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)

	// Test zero count for non-existent user
	count, err = repo.GetUnreadCount(ctx, uuid.New())
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestNotificationRepository_MarkAsRead(t *testing.T) {
	db := setupTestDB(t)
	repo := NewNotificationRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db)
	notification := &models.Notification{
		UserID:       userID,
		Type:         models.NotificationTypeShareReceived,
		ResourceType: "card",
		ResourceID:   uuid.New(),
		IsRead:       false,
		Metadata:     models.NotificationMetadata{},
	}
	db.Create(notification)
	defer db.Exec("DELETE FROM notifications WHERE id = ?", notification.ID)

	// Mark as read
	err := repo.MarkAsRead(ctx, userID, notification.ID)
	assert.NoError(t, err)

	// Verify it's marked as read
	var found models.Notification
	db.First(&found, "id = ?", notification.ID)
	assert.True(t, found.IsRead)
	assert.NotNil(t, found.ReadAt)
}

func TestNotificationRepository_MarkAllAsRead(t *testing.T) {
	db := setupTestDB(t)
	repo := NewNotificationRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db)

	// Create multiple unread notifications
	notifications := []models.Notification{
		{
			UserID:       userID,
			Type:         models.NotificationTypeShareReceived,
			ResourceType: "card",
			ResourceID:   uuid.New(),
			IsRead:       false,
			Metadata:     models.NotificationMetadata{},
		},
		{
			UserID:       userID,
			Type:         models.NotificationTypeShareReceived,
			ResourceType: "voucher",
			ResourceID:   uuid.New(),
			IsRead:       false,
			Metadata:     models.NotificationMetadata{},
		},
	}
	for i := range notifications {
		db.Create(&notifications[i])
		defer db.Exec("DELETE FROM notifications WHERE id = ?", notifications[i].ID)
	}

	// Mark all as read
	err := repo.MarkAllAsRead(ctx, userID)
	assert.NoError(t, err)

	// Verify all are marked as read
	var found []models.Notification
	db.Where("user_id = ?", userID).Find(&found)
	for _, n := range found {
		if n.ID == notifications[0].ID || n.ID == notifications[1].ID {
			assert.True(t, n.IsRead)
			assert.NotNil(t, n.ReadAt)
		}
	}
}

func TestNotificationRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewNotificationRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db)
	notification := &models.Notification{
		UserID:       userID,
		Type:         models.NotificationTypeShareReceived,
		ResourceType: "card",
		ResourceID:   uuid.New(),
		Metadata:     models.NotificationMetadata{},
	}
	db.Create(notification)

	// Delete it
	err := repo.Delete(ctx, userID, notification.ID)
	assert.NoError(t, err)

	// Verify it's soft deleted
	var found models.Notification
	err = db.First(&found, "id = ?", notification.ID).Error
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)

	// Verify it exists with Unscoped
	err = db.Unscoped().First(&found, "id = ?", notification.ID).Error
	assert.NoError(t, err)
	assert.NotNil(t, found.DeletedAt)
}
