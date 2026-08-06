package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// addNotificationEmailDelivery turns the notifications table into an outbox by
// giving every row an email delivery state. Until now the email was sent inline
// while the row was created and a send failure was only logged, so a failed mail
// was never retried.
//
// The default is 'skipped', not 'pending': existing rows predate the outbox and
// their mails were already sent (or already lost). With 'pending' the first
// dispatcher run would mail out the entire notification history.
//
// Migration 000033 - 2026-08-06
func addNotificationEmailDelivery() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202608060033_add_notification_email_delivery",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec(`
				ALTER TABLE notifications
				ADD COLUMN IF NOT EXISTS email_status VARCHAR(20) NOT NULL DEFAULT 'skipped';
			`).Error; err != nil {
				return err
			}

			if err := tx.Exec(`
				ALTER TABLE notifications
				ADD COLUMN IF NOT EXISTS email_attempts INTEGER NOT NULL DEFAULT 0;
			`).Error; err != nil {
				return err
			}

			if err := tx.Exec(`
				ALTER TABLE notifications
				ADD COLUMN IF NOT EXISTS email_last_error TEXT;
			`).Error; err != nil {
				return err
			}

			// Partial index: the dispatcher only ever scans for pending rows, and
			// those are a tiny slice of the table. A full index would carry every
			// delivered notification for no benefit.
			if err := tx.Exec(`
				CREATE INDEX IF NOT EXISTS idx_notifications_email_pending
				ON notifications (created_at)
				WHERE email_status = 'pending';
			`).Error; err != nil {
				return err
			}

			// Stale-claim recovery scans sending rows by age.
			if err := tx.Exec(`
				CREATE INDEX IF NOT EXISTS idx_notifications_email_sending
				ON notifications (updated_at)
				WHERE email_status = 'sending';
			`).Error; err != nil {
				return err
			}

			return tx.Exec(`
				COMMENT ON COLUMN notifications.email_status IS 'Email delivery state: pending, sending, sent, failed, skipped';
			`).Error
		},
		Rollback: func(tx *gorm.DB) error {
			if err := tx.Exec("DROP INDEX IF EXISTS idx_notifications_email_sending").Error; err != nil {
				return err
			}
			if err := tx.Exec("DROP INDEX IF EXISTS idx_notifications_email_pending").Error; err != nil {
				return err
			}
			if err := tx.Exec("ALTER TABLE notifications DROP COLUMN IF EXISTS email_last_error").Error; err != nil {
				return err
			}
			if err := tx.Exec("ALTER TABLE notifications DROP COLUMN IF EXISTS email_attempts").Error; err != nil {
				return err
			}
			return tx.Exec("ALTER TABLE notifications DROP COLUMN IF EXISTS email_status").Error
		},
	}
}
