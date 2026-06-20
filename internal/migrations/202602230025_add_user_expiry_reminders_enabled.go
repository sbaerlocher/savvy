package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// addUserExpiryRemindersEnabled adds a preference field for opting out of expiry reminders
// Migration 000025 - 2026-02-23
func addUserExpiryRemindersEnabled() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202602230025_add_user_expiry_reminders_enabled",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec(`
				ALTER TABLE users
				ADD COLUMN IF NOT EXISTS expiry_reminders_enabled BOOLEAN NOT NULL DEFAULT true;
			`).Error; err != nil {
				return err
			}

			return tx.Exec(`
				COMMENT ON COLUMN users.expiry_reminders_enabled IS 'Whether the user wants to receive expiry reminders for vouchers and gift cards';
			`).Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Exec("ALTER TABLE users DROP COLUMN IF EXISTS expiry_reminders_enabled").Error
		},
	}
}
