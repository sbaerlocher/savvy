package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// addCaseInsensitiveEmailIndex replaces the case-sensitive email index with a case-insensitive one
// Equivalent to: migrations/000006_case_insensitive_email_index.up.sql
func addCaseInsensitiveEmailIndex() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202601250006_case_insensitive_email_index",
		Migrate: func(tx *gorm.DB) error {
			// Drop the existing case-sensitive unique index
			if err := tx.Exec(`
				DROP INDEX IF EXISTS idx_users_email;
			`).Error; err != nil {
				return err
			}

			// Create case-insensitive unique index using LOWER()
			if err := tx.Exec(`
				CREATE UNIQUE INDEX idx_users_email_lower ON users (LOWER(email));
			`).Error; err != nil {
				return err
			}

			// Add comment to explain the index
			if err := tx.Exec(`
				COMMENT ON INDEX idx_users_email_lower IS 'Case-insensitive unique index on email to prevent duplicate emails with different cases (e.g., Test@Email.com and test@email.com)';
			`).Error; err != nil {
				return err
			}

			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			// Drop case-insensitive index
			if err := tx.Exec("DROP INDEX IF EXISTS idx_users_email_lower").Error; err != nil {
				return err
			}

			// Recreate original case-sensitive index
			if err := tx.Exec(`
				CREATE UNIQUE INDEX idx_users_email ON users (email);
			`).Error; err != nil {
				return err
			}

			return nil
		},
	}
}
