package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// fixNotificationPreferenceDefaults changes notification preference defaults from true to false.
// Email preferences are now only enabled when email is verified.
// Push preferences are now only enabled when a push subscription is registered.
// This migration updates existing users: disable email prefs for unverified users,
// disable push prefs for users without push subscriptions.
func fixNotificationPreferenceDefaults() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202603160030_fix_notification_preference_defaults",
		Migrate: func(tx *gorm.DB) error {
			// Change column defaults from true to false
			if err := tx.Exec(`
				ALTER TABLE users
				ALTER COLUMN push_notifications_enabled SET DEFAULT false,
				ALTER COLUMN email_notifications_enabled SET DEFAULT false,
				ALTER COLUMN push_reminders_enabled SET DEFAULT false,
				ALTER COLUMN push_sharing_enabled SET DEFAULT false,
				ALTER COLUMN email_reminders_enabled SET DEFAULT false,
				ALTER COLUMN email_sharing_enabled SET DEFAULT false;
			`).Error; err != nil {
				return err
			}

			// Disable email preferences for users who have NOT verified their email
			// Use IS NOT TRUE to also cover potential NULL values from legacy data
			if err := tx.Exec(`
				UPDATE users SET
					email_notifications_enabled = false,
					email_reminders_enabled = false,
					email_sharing_enabled = false
				WHERE email_verified IS NOT TRUE;
			`).Error; err != nil {
				return err
			}

			// Disable push preferences for users who have NO push subscriptions
			if err := tx.Exec(`
				UPDATE users SET
					push_notifications_enabled = false,
					push_reminders_enabled = false,
					push_sharing_enabled = false
				WHERE id NOT IN (
					SELECT DISTINCT user_id FROM push_subscriptions
				);
			`).Error; err != nil {
				return err
			}

			if err := tx.Exec(`COMMENT ON COLUMN users.email_notifications_enabled IS 'Global email channel — auto-enabled on email verification'`).Error; err != nil {
				return err
			}

			return tx.Exec(`COMMENT ON COLUMN users.push_notifications_enabled IS 'Global push channel — auto-enabled on first push subscription'`).Error
		},
		Rollback: func(tx *gorm.DB) error {
			// Restore original defaults
			if err := tx.Exec(`
				ALTER TABLE users
				ALTER COLUMN push_notifications_enabled SET DEFAULT true,
				ALTER COLUMN email_notifications_enabled SET DEFAULT true,
				ALTER COLUMN push_reminders_enabled SET DEFAULT true,
				ALTER COLUMN push_sharing_enabled SET DEFAULT true,
				ALTER COLUMN email_reminders_enabled SET DEFAULT true,
				ALTER COLUMN email_sharing_enabled SET DEFAULT true;
			`).Error; err != nil {
				return err
			}

			// Re-enable all preferences (original behavior)
			return tx.Exec(`
				UPDATE users SET
					push_notifications_enabled = true,
					email_notifications_enabled = true,
					push_reminders_enabled = true,
					push_sharing_enabled = true,
					email_reminders_enabled = true,
					email_sharing_enabled = true;
			`).Error
		},
	}
}
