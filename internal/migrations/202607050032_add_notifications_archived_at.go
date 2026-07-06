package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// addNotificationsArchivedAt adds an archived_at column so read notifications
// can be auto-archived out of the main list without being deleted.
// Migration 000032 - 2026-07-05
func addNotificationsArchivedAt() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202607050032_add_notifications_archived_at",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec(`
				ALTER TABLE notifications
				ADD COLUMN IF NOT EXISTS archived_at TIMESTAMP WITH TIME ZONE;
			`).Error; err != nil {
				return err
			}

			if err := tx.Exec(`
				CREATE INDEX IF NOT EXISTS idx_notifications_archived_at
				ON notifications (archived_at);
			`).Error; err != nil {
				return err
			}

			return tx.Exec(`
				COMMENT ON COLUMN notifications.archived_at IS 'Archive timestamp - NULL means active in the main list';
			`).Error
		},
		Rollback: func(tx *gorm.DB) error {
			if err := tx.Exec("DROP INDEX IF EXISTS idx_notifications_archived_at").Error; err != nil {
				return err
			}
			return tx.Exec("ALTER TABLE notifications DROP COLUMN IF EXISTS archived_at").Error
		},
	}
}
