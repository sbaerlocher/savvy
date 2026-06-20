package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// addUserSecurityColumns adds failed_login_attempts and locked_until columns
// for rate limiting and account lockout functionality
// Migration 000018 - 2026-02-09
func addUserSecurityColumns() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202602090018_add_user_security_columns",
		Migrate: func(tx *gorm.DB) error {
			// Add failed_login_attempts column with default 0
			if err := tx.Exec(`
				ALTER TABLE users
				ADD COLUMN IF NOT EXISTS failed_login_attempts INTEGER NOT NULL DEFAULT 0;
			`).Error; err != nil {
				return err
			}

			// Add locked_until column (nullable timestamp)
			if err := tx.Exec(`
				ALTER TABLE users
				ADD COLUMN IF NOT EXISTS locked_until TIMESTAMP WITH TIME ZONE;
			`).Error; err != nil {
				return err
			}

			// Add index on locked_until for efficient lockout queries
			if err := tx.Exec(`
				CREATE INDEX IF NOT EXISTS idx_users_locked_until
				ON users (locked_until);
			`).Error; err != nil {
				return err
			}

			// Add column comments
			if err := tx.Exec(`
				COMMENT ON COLUMN users.failed_login_attempts IS 'Number of consecutive failed login attempts (resets on successful login)';
			`).Error; err != nil {
				return err
			}

			if err := tx.Exec(`
				COMMENT ON COLUMN users.locked_until IS 'Account locked until this timestamp (NULL = not locked)';
			`).Error; err != nil {
				return err
			}

			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			// Drop index
			if err := tx.Exec("DROP INDEX IF EXISTS idx_users_locked_until").Error; err != nil {
				return err
			}

			// Drop columns
			if err := tx.Exec("ALTER TABLE users DROP COLUMN IF EXISTS locked_until").Error; err != nil {
				return err
			}

			if err := tx.Exec("ALTER TABLE users DROP COLUMN IF EXISTS failed_login_attempts").Error; err != nil {
				return err
			}

			return nil
		},
	}
}
