package migrations

import (
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// addEmailVerificationAndTokens adds email verification columns to users and creates email_tokens table
// Migration 000019 - 2026-02-19
func addEmailVerificationAndTokens() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202602190019_add_email_verification_and_tokens",
		Migrate: func(tx *gorm.DB) error {
			// Add email_verified column with default false
			if err := tx.Exec(`
				ALTER TABLE users
				ADD COLUMN IF NOT EXISTS email_verified BOOLEAN NOT NULL DEFAULT false;
			`).Error; err != nil {
				return err
			}

			// Add email_verified_at column (nullable timestamp)
			if err := tx.Exec(`
				ALTER TABLE users
				ADD COLUMN IF NOT EXISTS email_verified_at TIMESTAMP WITH TIME ZONE;
			`).Error; err != nil {
				return err
			}

			// Mark all existing users as verified (they registered before verification was required)
			if err := tx.Exec(`
				UPDATE users SET email_verified = true, email_verified_at = CURRENT_TIMESTAMP
				WHERE email_verified = false;
			`).Error; err != nil {
				return err
			}

			// Create email_tokens table
			type EmailToken struct {
				ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
				UserID    uuid.UUID  `gorm:"type:uuid;not null;index:idx_email_tokens_user_type"`
				TokenHash string     `gorm:"type:varchar(64);not null;uniqueIndex:idx_email_tokens_token_hash"`
				TokenType string     `gorm:"type:varchar(50);not null;index:idx_email_tokens_user_type"`
				ExpiresAt time.Time  `gorm:"type:timestamp with time zone;not null"`
				UsedAt    *time.Time `gorm:"type:timestamp with time zone"`
				CreatedAt time.Time  `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP"`
			}

			if err := tx.AutoMigrate(&EmailToken{}); err != nil {
				return err
			}

			// Add expires_at index for cleanup queries
			if err := tx.Exec(`
				CREATE INDEX IF NOT EXISTS idx_email_tokens_expires_at
				ON email_tokens (expires_at);
			`).Error; err != nil {
				return err
			}

			// Add foreign key constraint for user_id
			if err := tx.Exec(`
				DO $$
				BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM pg_constraint WHERE conname = 'fk_email_tokens_user'
					) THEN
						ALTER TABLE email_tokens
						ADD CONSTRAINT fk_email_tokens_user
						FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
					END IF;
				END $$;
			`).Error; err != nil {
				return err
			}

			// Add column comments
			if err := tx.Exec(`COMMENT ON COLUMN users.email_verified IS 'Whether the user email address has been verified'`).Error; err != nil {
				return err
			}
			if err := tx.Exec(`COMMENT ON COLUMN users.email_verified_at IS 'Timestamp when email was verified'`).Error; err != nil {
				return err
			}
			if err := tx.Exec(`COMMENT ON TABLE email_tokens IS 'Tokens for email verification and password reset'`).Error; err != nil {
				return err
			}
			if err := tx.Exec(`COMMENT ON COLUMN email_tokens.token_hash IS 'SHA-256 hash of the token (plain token is sent to user)'`).Error; err != nil {
				return err
			}
			if err := tx.Exec(`COMMENT ON COLUMN email_tokens.token_type IS 'Token type: email_verification or password_reset'`).Error; err != nil {
				return err
			}

			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			// Drop email_tokens table
			if err := tx.Exec("DROP TABLE IF EXISTS email_tokens CASCADE").Error; err != nil {
				return err
			}

			// Drop columns from users
			if err := tx.Exec("ALTER TABLE users DROP COLUMN IF EXISTS email_verified_at").Error; err != nil {
				return err
			}
			return tx.Exec("ALTER TABLE users DROP COLUMN IF EXISTS email_verified").Error
		},
	}
}
