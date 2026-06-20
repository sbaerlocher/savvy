package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// addPerChannelNotificationPreferences adds 6 per-channel notification preference columns
// (2 channel toggles + 4 category toggles) and migrates old category preferences.
//
// Channels: push_notifications_enabled, email_notifications_enabled
// Categories per channel: {push,email}_expiry_enabled (expiry + validity start),
//
//	{push,email}_share_enabled (share + transfer)
//
// Data migration: copies expiry_reminders_enabled → email_reminders_enabled,
//
//	share_emails_enabled → email_sharing_enabled
//
// Migration 000027 - 2026-02-28
func addPerChannelNotificationPreferences() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202602280027_add_per_channel_notification_preferences",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec(`
				ALTER TABLE users
				ADD COLUMN IF NOT EXISTS push_notifications_enabled BOOLEAN NOT NULL DEFAULT true,
				ADD COLUMN IF NOT EXISTS email_notifications_enabled BOOLEAN NOT NULL DEFAULT true,
				ADD COLUMN IF NOT EXISTS push_reminders_enabled BOOLEAN NOT NULL DEFAULT true,
				ADD COLUMN IF NOT EXISTS push_sharing_enabled BOOLEAN NOT NULL DEFAULT true,
				ADD COLUMN IF NOT EXISTS email_reminders_enabled BOOLEAN NOT NULL DEFAULT true,
				ADD COLUMN IF NOT EXISTS email_sharing_enabled BOOLEAN NOT NULL DEFAULT true;
			`).Error; err != nil {
				return err
			}

			// Migrate old category preferences to new per-channel email fields
			if err := tx.Exec(`
				UPDATE users
				SET email_reminders_enabled = expiry_reminders_enabled,
				    email_sharing_enabled = share_emails_enabled;
			`).Error; err != nil {
				return err
			}

			// Add column comments
			for _, stmt := range []string{
				"COMMENT ON COLUMN users.push_notifications_enabled IS 'Global toggle for push notifications'",
				"COMMENT ON COLUMN users.email_notifications_enabled IS 'Global toggle for email notifications'",
				"COMMENT ON COLUMN users.push_reminders_enabled IS 'Push notifications for expiry reminders and validity start'",
				"COMMENT ON COLUMN users.push_sharing_enabled IS 'Push notifications for share and transfer events'",
				"COMMENT ON COLUMN users.email_reminders_enabled IS 'Email notifications for expiry reminders and validity start'",
				"COMMENT ON COLUMN users.email_sharing_enabled IS 'Email notifications for share and transfer events'",
			} {
				if err := tx.Exec(stmt).Error; err != nil {
					return err
				}
			}

			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Exec(`
				ALTER TABLE users
				DROP COLUMN IF EXISTS push_notifications_enabled,
				DROP COLUMN IF EXISTS email_notifications_enabled,
				DROP COLUMN IF EXISTS push_reminders_enabled,
				DROP COLUMN IF EXISTS push_sharing_enabled,
				DROP COLUMN IF EXISTS email_reminders_enabled,
				DROP COLUMN IF EXISTS email_sharing_enabled;
			`).Error
		},
	}
}
