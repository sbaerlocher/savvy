package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// addUserPasswordChangedAt adds a timestamp column to track when the password was last changed.
// Used to invalidate sessions created before the password change (M1/M2 security fix).
// Migration 000028 - 2026-02-28
func addUserPasswordChangedAt() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202602280028_add_user_password_changed_at",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec(`
				ALTER TABLE users
				ADD COLUMN IF NOT EXISTS password_changed_at TIMESTAMPTZ;
			`).Error; err != nil {
				return err
			}

			return tx.Exec(`
				COMMENT ON COLUMN users.password_changed_at IS 'When the password was last changed; sessions created before this timestamp are invalidated';
			`).Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Exec("ALTER TABLE users DROP COLUMN IF EXISTS password_changed_at").Error
		},
	}
}
