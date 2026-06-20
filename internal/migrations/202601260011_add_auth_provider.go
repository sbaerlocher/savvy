package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// addAuthProvider adds the auth_provider column to users table to distinguish OAuth from local users
// Migration 000011 - 2026-01-26
func addAuthProvider() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202601260011_add_auth_provider",
		Migrate: func(tx *gorm.DB) error {
			// Add auth_provider column with default 'local'
			if err := tx.Exec(`
				ALTER TABLE users
				ADD COLUMN IF NOT EXISTS auth_provider VARCHAR(50) NOT NULL DEFAULT 'local';
			`).Error; err != nil {
				return err
			}

			// Add index for auth_provider queries
			if err := tx.Exec(`
				CREATE INDEX IF NOT EXISTS idx_users_auth_provider
				ON users (auth_provider);
			`).Error; err != nil {
				return err
			}

			// Add column comment
			if err := tx.Exec(`
				COMMENT ON COLUMN users.auth_provider IS 'Authentication provider: "local" for username/password, "oauth" for OAuth/OIDC';
			`).Error; err != nil {
				return err
			}

			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			// Drop index
			if err := tx.Exec("DROP INDEX IF EXISTS idx_users_auth_provider").Error; err != nil {
				return err
			}

			// Drop column
			if err := tx.Exec("ALTER TABLE users DROP COLUMN IF EXISTS auth_provider").Error; err != nil {
				return err
			}

			return nil
		},
	}
}
