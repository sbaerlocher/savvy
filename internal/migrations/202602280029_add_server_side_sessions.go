package migrations

import (
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// addServerSideSessions creates the sessions table for server-side session management.
// Replaces CookieStore with PostgreSQL-backed sessions enabling session listing and revocation.
func addServerSideSessions() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202602280029_add_server_side_sessions",
		Migrate: func(tx *gorm.DB) error {
			type Session struct {
				ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
				UserID       *uuid.UUID `gorm:"type:uuid;index:idx_sessions_user_id"`
				TokenHash    string     `gorm:"type:text;not null;uniqueIndex:idx_sessions_token_hash"`
				Data         []byte     `gorm:"type:bytea;not null"`
				IPAddress    string     `gorm:"type:text;not null;default:''"`
				UserAgent    string     `gorm:"type:text;not null;default:''"`
				CreatedAt    time.Time  `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP"`
				LastActiveAt time.Time  `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP"`
				ExpiresAt    time.Time  `gorm:"type:timestamp with time zone;not null;index:idx_sessions_expires_at"`
			}

			if err := tx.AutoMigrate(&Session{}); err != nil {
				return err
			}

			// Add foreign key constraint (nullable - some sessions may not have a user yet)
			if err := tx.Exec(`
				DO $$
				BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM pg_constraint WHERE conname = 'fk_sessions_user'
					) THEN
						ALTER TABLE sessions
						ADD CONSTRAINT fk_sessions_user
						FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
					END IF;
				END $$;
			`).Error; err != nil {
				return err
			}

			return tx.Exec(`COMMENT ON TABLE sessions IS 'Server-side sessions for authentication, enabling session listing and revocation'`).Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Exec("DROP TABLE IF EXISTS sessions CASCADE").Error
		},
	}
}
