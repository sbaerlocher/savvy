package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// addNotificationsSoftDelete adds soft delete support to notifications table
// Migration 000016 - 2026-02-06
func addNotificationsSoftDelete() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202602060016_add_notifications_soft_delete",
		Migrate: func(tx *gorm.DB) error {
			// Add deleted_at column with index
			if err := tx.Exec(`
				ALTER TABLE notifications
				ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE;
			`).Error; err != nil {
				return err
			}

			// Add index on deleted_at for soft delete queries
			if err := tx.Exec(`
				CREATE INDEX IF NOT EXISTS idx_notifications_deleted_at
				ON notifications (deleted_at);
			`).Error; err != nil {
				return err
			}

			// Add column comment
			if err := tx.Exec(`
				COMMENT ON COLUMN notifications.deleted_at IS 'Soft delete timestamp - NULL means not deleted';
			`).Error; err != nil {
				return err
			}

			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			// Drop index and column
			if err := tx.Exec("DROP INDEX IF EXISTS idx_notifications_deleted_at").Error; err != nil {
				return err
			}
			return tx.Exec("ALTER TABLE notifications DROP COLUMN IF EXISTS deleted_at").Error
		},
	}
}
