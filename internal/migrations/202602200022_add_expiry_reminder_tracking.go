package migrations

import (
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// addExpiryReminderTracking creates the expiry_reminders_sent table for tracking sent reminders
// Migration 000022 - 2026-02-20
func addExpiryReminderTracking() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202602200022_add_expiry_reminder_tracking",
		Migrate: func(tx *gorm.DB) error {
			type ExpiryReminderSent struct {
				ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
				UserID       uuid.UUID `gorm:"type:uuid;not null;index:idx_expiry_reminders_user"`
				ResourceType string    `gorm:"type:varchar(50);not null"`
				ResourceID   uuid.UUID `gorm:"type:uuid;not null"`
				DaysBefore   int       `gorm:"not null"`
				SentAt       time.Time `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP"`
			}

			if err := tx.AutoMigrate(&ExpiryReminderSent{}); err != nil {
				return err
			}

			// Add unique constraint to prevent duplicate reminders
			if err := tx.Exec(`
				DO $$
				BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM pg_constraint WHERE conname = 'expiry_reminders_sent_unique'
					) THEN
						ALTER TABLE expiry_reminder_sents
						ADD CONSTRAINT expiry_reminders_sent_unique
						UNIQUE (user_id, resource_type, resource_id, days_before);
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
						SELECT 1 FROM pg_constraint WHERE conname = 'fk_expiry_reminders_user'
					) THEN
						ALTER TABLE expiry_reminder_sents
						ADD CONSTRAINT fk_expiry_reminders_user
						FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
					END IF;
				END $$;
			`).Error; err != nil {
				return err
			}

			// Add table comment
			return tx.Exec(`COMMENT ON TABLE expiry_reminder_sents IS 'Tracks which expiry reminders have been sent to prevent duplicates'`).Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Exec("DROP TABLE IF EXISTS expiry_reminder_sents CASCADE").Error
		},
	}
}
