package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// addUserLanguage adds the language column to users table for localized emails
// Migration 000020 - 2026-02-19
func addUserLanguage() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202602190020_add_user_language",
		Migrate: func(tx *gorm.DB) error {
			// Add language column with default 'de'
			if err := tx.Exec(`
				ALTER TABLE users
				ADD COLUMN IF NOT EXISTS language VARCHAR(5) NOT NULL DEFAULT 'de';
			`).Error; err != nil {
				return err
			}

			// Add column comment
			return tx.Exec(`
				COMMENT ON COLUMN users.language IS 'User preferred language for emails and UI (de, en, fr)';
			`).Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Exec("ALTER TABLE users DROP COLUMN IF EXISTS language").Error
		},
	}
}
