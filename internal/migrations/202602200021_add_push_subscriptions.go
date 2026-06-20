package migrations

import (
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// addPushSubscriptions creates the push_subscriptions table for Web Push notifications
// Migration 000021 - 2026-02-20
func addPushSubscriptions() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202602200021_add_push_subscriptions",
		Migrate: func(tx *gorm.DB) error {
			type PushSubscription struct {
				ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
				UserID    uuid.UUID `gorm:"type:uuid;not null;index:idx_push_subscriptions_user_id"`
				Endpoint  string    `gorm:"type:text;not null;uniqueIndex"`
				P256dhKey string    `gorm:"type:text;not null"`
				AuthKey   string    `gorm:"type:text;not null"` // #nosec G117 -- struct field name, not a hardcoded secret
				UserAgent string    `gorm:"type:text"`
				CreatedAt time.Time `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP"`
				UpdatedAt time.Time `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP"`
			}

			if err := tx.AutoMigrate(&PushSubscription{}); err != nil {
				return err
			}

			// Add foreign key constraint
			if err := tx.Exec(`
				DO $$
				BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM pg_constraint WHERE conname = 'fk_push_subscriptions_user'
					) THEN
						ALTER TABLE push_subscriptions
						ADD CONSTRAINT fk_push_subscriptions_user
						FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
					END IF;
				END $$;
			`).Error; err != nil {
				return err
			}

			// Add table comment
			return tx.Exec(`COMMENT ON TABLE push_subscriptions IS 'Web Push API subscriptions for real-time browser notifications'`).Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Exec("DROP TABLE IF EXISTS push_subscriptions CASCADE").Error
		},
	}
}
