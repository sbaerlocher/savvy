package migrations

import (
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// addNotifications creates the notifications table for in-app notifications
// Migration 000015 - 2026-02-06
func addNotifications() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202602060015_add_notifications",
		Migrate: func(tx *gorm.DB) error {
			// Define Notification struct for migration
			type Notification struct {
				ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
				UserID       uuid.UUID  `gorm:"type:uuid;not null;index:idx_notifications_user_id"`
				Type         string     `gorm:"type:varchar(50);not null;index:idx_notifications_type"`
				ResourceType string     `gorm:"type:varchar(50);not null"`
				ResourceID   uuid.UUID  `gorm:"type:uuid;not null;index:idx_notifications_resource"`
				Metadata     string     `gorm:"type:jsonb;default:'{}'"`
				IsRead       bool       `gorm:"default:false;index:idx_notifications_is_read"`
				ReadAt       *time.Time `gorm:"type:timestamp with time zone"`
				CreatedAt    time.Time  `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP;index:idx_notifications_created_at"`
				UpdatedAt    time.Time  `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP"`
			}

			// Create table
			if err := tx.AutoMigrate(&Notification{}); err != nil {
				return err
			}

			// Add foreign key constraint for user_id (idempotent)
			if err := tx.Exec(`
				DO $$
				BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM pg_constraint WHERE conname = 'fk_notifications_user'
					) THEN
						ALTER TABLE notifications
						ADD CONSTRAINT fk_notifications_user
						FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
					END IF;
				END $$;
			`).Error; err != nil {
				return err
			}

			// Add composite index for unread notifications query (most common query)
			if err := tx.Exec(`
				CREATE INDEX IF NOT EXISTS idx_notifications_user_unread
				ON notifications (user_id, created_at DESC)
				WHERE is_read = FALSE;
			`).Error; err != nil {
				return err
			}

			// Add composite index for resource lookups
			if err := tx.Exec(`
				CREATE INDEX IF NOT EXISTS idx_notifications_resource_lookup
				ON notifications (resource_type, resource_id, created_at DESC);
			`).Error; err != nil {
				return err
			}

			// Add table comment
			if err := tx.Exec(`
				COMMENT ON TABLE notifications IS 'In-app notifications for share and transfer events';
			`).Error; err != nil {
				return err
			}

			// Add column comments (each must be a separate statement)
			if err := tx.Exec("COMMENT ON COLUMN notifications.type IS 'Notification type: share_received, transfer_received'").Error; err != nil {
				return err
			}
			if err := tx.Exec("COMMENT ON COLUMN notifications.resource_type IS 'Type of resource: card, voucher, gift_card'").Error; err != nil {
				return err
			}
			if err := tx.Exec("COMMENT ON COLUMN notifications.resource_id IS 'UUID of the resource'").Error; err != nil {
				return err
			}
			if err := tx.Exec("COMMENT ON COLUMN notifications.metadata IS 'JSONB metadata: from_user_id, from_user_name, permissions, etc.'").Error; err != nil {
				return err
			}

			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Exec("DROP TABLE IF EXISTS notifications CASCADE").Error
		},
	}
}
