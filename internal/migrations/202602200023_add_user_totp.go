package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// addUserTOTP creates the user_totps table for 2FA
func addUserTOTP() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202602200023_add_user_totp",
		Migrate: func(tx *gorm.DB) error {
			// Create user_totps table
			if err := tx.Exec(`
				CREATE TABLE IF NOT EXISTS user_totps (
					id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					user_id UUID NOT NULL,
					secret TEXT NOT NULL,
					backup_codes TEXT NOT NULL,
					enabled BOOLEAN NOT NULL DEFAULT false,
					verified BOOLEAN NOT NULL DEFAULT false,
					enabled_at TIMESTAMP WITH TIME ZONE,
					created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
				)
			`).Error; err != nil {
				return err
			}

			// Add unique index on user_id
			if err := tx.Exec(`
				DO $$
				BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM pg_indexes WHERE indexname = 'idx_user_totps_user_id'
					) THEN
						CREATE UNIQUE INDEX idx_user_totps_user_id ON user_totps(user_id);
					END IF;
				END $$;
			`).Error; err != nil {
				return err
			}

			// Add foreign key constraint
			if err := tx.Exec(`
				DO $$
				BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM pg_constraint WHERE conname = 'fk_user_totps_user'
					) THEN
						ALTER TABLE user_totps
						ADD CONSTRAINT fk_user_totps_user
						FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
					END IF;
				END $$;
			`).Error; err != nil {
				return err
			}

			return tx.Exec(`COMMENT ON TABLE user_totps IS 'Stores TOTP two-factor authentication configuration per user'`).Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Exec("DROP TABLE IF EXISTS user_totps CASCADE").Error
		},
	}
}
