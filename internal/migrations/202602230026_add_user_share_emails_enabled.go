package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// addUserShareEmailsEnabled adds a preference field for opting out of share/transfer email notifications
// Migration 000026 - 2026-02-23
func addUserShareEmailsEnabled() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202602230026_add_user_share_emails_enabled",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec(`
				ALTER TABLE users
				ADD COLUMN IF NOT EXISTS share_emails_enabled BOOLEAN NOT NULL DEFAULT true;
			`).Error; err != nil {
				return err
			}

			return tx.Exec(`
				COMMENT ON COLUMN users.share_emails_enabled IS 'Whether the user wants to receive email notifications for shares and transfers';
			`).Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Exec("ALTER TABLE users DROP COLUMN IF EXISTS share_emails_enabled").Error
		},
	}
}
