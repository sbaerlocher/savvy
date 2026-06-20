package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// normalizeEmails creates trigger to automatically lowercase emails
// Equivalent to: migrations/000004_normalize_emails.up.sql
func normalizeEmails() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202601230004_normalize_emails",
		Migrate: func(tx *gorm.DB) error {
			// Normalize all existing emails to lowercase
			if err := tx.Exec("UPDATE users SET email = LOWER(email)").Error; err != nil {
				return err
			}

			// Create trigger function to automatically lowercase emails on insert/update
			if err := createFunction(tx, `
				CREATE OR REPLACE FUNCTION enforce_lowercase_email()
				RETURNS TRIGGER AS $$
				BEGIN
					NEW.email = LOWER(TRIM(NEW.email));
					RETURN NEW;
				END;
				$$ LANGUAGE plpgsql;
			`); err != nil {
				return err
			}

			// Create trigger on users table
			if err := createTrigger(tx, "trigger_lowercase_email", "users",
				"BEFORE", "INSERT OR UPDATE", "enforce_lowercase_email"); err != nil {
				return err
			}

			// Add comment for documentation
			return addComment(tx, `
				COMMENT ON FUNCTION enforce_lowercase_email() IS 'Automatically converts email addresses to lowercase to ensure case-insensitive uniqueness';
			`)
		},
		Rollback: func(tx *gorm.DB) error {
			if err := dropTrigger(tx, "trigger_lowercase_email", "users"); err != nil {
				return err
			}
			return dropFunction(tx, "enforce_lowercase_email")
		},
	}
}
