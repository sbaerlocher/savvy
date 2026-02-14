package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNotification_BeforeCreate(t *testing.T) {
	notification := &Notification{}

	// ID should be Nil before BeforeCreate
	assert.Equal(t, uuid.Nil, notification.ID)

	err := notification.BeforeCreate(nil)

	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, notification.ID)
}

func TestNotification_BeforeCreate_ExistingID(t *testing.T) {
	existingID := uuid.New()
	notification := &Notification{
		ID: existingID,
	}

	err := notification.BeforeCreate(nil)

	assert.NoError(t, err)
	assert.Equal(t, existingID, notification.ID) // ID should not change
}

func TestNotification_MarkAsRead(t *testing.T) {
	notification := &Notification{
		IsRead: false,
		ReadAt: nil,
	}

	notification.MarkAsRead()

	assert.True(t, notification.IsRead)
	assert.NotNil(t, notification.ReadAt)
	assert.WithinDuration(t, time.Now(), *notification.ReadAt, time.Second)
}

func TestNotification_GetFromUserID_WithMetadata(t *testing.T) {
	notification := &Notification{
		Metadata: NotificationMetadata{
			"from_user_id": "123e4567-e89b-12d3-a456-426614174000",
		},
	}

	userID := notification.GetFromUserID()

	assert.Equal(t, "123e4567-e89b-12d3-a456-426614174000", userID)
}

func TestNotification_GetFromUserID_NoMetadata(t *testing.T) {
	notification := &Notification{
		Metadata: NotificationMetadata{},
	}

	userID := notification.GetFromUserID()

	assert.Equal(t, "", userID)
}

func TestNotification_GetFromUserName_WithMetadata(t *testing.T) {
	notification := &Notification{
		Metadata: NotificationMetadata{
			"from_user_name": "John Doe",
		},
	}

	userName := notification.GetFromUserName()

	assert.Equal(t, "John Doe", userName)
}

func TestNotification_GetFromUserName_NoMetadata(t *testing.T) {
	notification := &Notification{
		Metadata: NotificationMetadata{},
	}

	userName := notification.GetFromUserName()

	assert.Equal(t, "Unknown User", userName)
}

func TestNotification_GetPermissions_WithMetadata(t *testing.T) {
	notification := &Notification{
		Metadata: NotificationMetadata{
			"permissions": map[string]interface{}{
				"can_edit":   true,
				"can_delete": false,
			},
		},
	}

	permissions := notification.GetPermissions()

	assert.NotNil(t, permissions)
	assert.Equal(t, true, permissions["can_edit"])
	assert.Equal(t, false, permissions["can_delete"])
}

func TestNotification_GetPermissions_NoMetadata(t *testing.T) {
	notification := &Notification{
		Metadata: NotificationMetadata{},
	}

	permissions := notification.GetPermissions()

	assert.Nil(t, permissions)
}

func TestNotification_IsShareNotification_True(t *testing.T) {
	notification := &Notification{
		Type: NotificationTypeShareReceived,
	}

	assert.True(t, notification.IsShareNotification())
}

func TestNotification_IsShareNotification_False(t *testing.T) {
	notification := &Notification{
		Type: NotificationTypeTransferReceived,
	}

	assert.False(t, notification.IsShareNotification())
}

func TestNotification_IsTransferNotification_True(t *testing.T) {
	notification := &Notification{
		Type: NotificationTypeTransferReceived,
	}

	assert.True(t, notification.IsTransferNotification())
}

func TestNotification_IsTransferNotification_False(t *testing.T) {
	notification := &Notification{
		Type: NotificationTypeShareReceived,
	}

	assert.False(t, notification.IsTransferNotification())
}

func TestNotificationMetadata_Value_Nil(t *testing.T) {
	var metadata NotificationMetadata

	value, err := metadata.Value()

	assert.NoError(t, err)
	assert.Equal(t, "{}", value)
}

func TestNotificationMetadata_Value_WithData(t *testing.T) {
	metadata := NotificationMetadata{
		"key1": "value1",
		"key2": 123,
	}

	value, err := metadata.Value()

	assert.NoError(t, err)

	// Unmarshal to verify JSON is valid
	var result map[string]interface{}
	err = json.Unmarshal(value.([]byte), &result)
	assert.NoError(t, err)
	assert.Equal(t, "value1", result["key1"])
	assert.Equal(t, float64(123), result["key2"]) // JSON numbers are float64
}

func TestNotificationMetadata_Scan_Nil(t *testing.T) {
	var metadata NotificationMetadata

	err := metadata.Scan(nil)

	assert.NoError(t, err)
	assert.NotNil(t, metadata)
	assert.Len(t, metadata, 0)
}

func TestNotificationMetadata_Scan_ValidJSON(t *testing.T) {
	var metadata NotificationMetadata
	jsonData := []byte(`{"key1": "value1", "key2": 123}`)

	err := metadata.Scan(jsonData)

	assert.NoError(t, err)
	assert.Equal(t, "value1", metadata["key1"])
	assert.Equal(t, float64(123), metadata["key2"]) // JSON numbers are float64
}

func TestNotificationMetadata_Scan_InvalidType(t *testing.T) {
	var metadata NotificationMetadata

	err := metadata.Scan("not a byte slice")

	assert.NoError(t, err)
	assert.NotNil(t, metadata)
	assert.Len(t, metadata, 0)
}

func TestNotificationMetadata_Scan_EmptyJSON(t *testing.T) {
	var metadata NotificationMetadata
	jsonData := []byte(`{}`)

	err := metadata.Scan(jsonData)

	assert.NoError(t, err)
	assert.NotNil(t, metadata)
	assert.Len(t, metadata, 0)
}

func TestNotificationType_ShareReceived(t *testing.T) {
	assert.Equal(t, NotificationType("share_received"), NotificationTypeShareReceived)
}

func TestNotificationType_TransferReceived(t *testing.T) {
	assert.Equal(t, NotificationType("transfer_received"), NotificationTypeTransferReceived)
}
