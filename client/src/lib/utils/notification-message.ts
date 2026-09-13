import type { NotificationDTO } from '$lib/types/api';

/**
 * Picks the translation key and params for a reminder notification body.
 *
 * Reminder bodies are rendered in two places (the notification list and the
 * panel), so the key choice lives here to keep both in sync.
 *
 * Both reminders prefer the absolute date the server put in the metadata.
 * A relative wording is written once and then stays put: a notification is
 * only archived after NOTIFICATION_ARCHIVE_AFTER_DAYS, so "expires in 7 days"
 * would still claim that a week later, next to a "7 days ago" timestamp. The
 * relative forms remain as a fallback for notifications created before the
 * client rendered the date, and carry a day-0 and day-1 case of their own so
 * they never read "in 1 days".
 *
 * `resourceLabel` differs per call site (the panel uses the article form), so
 * it is passed in rather than derived here.
 */
export function reminderMessageKey(
	notification: NotificationDTO,
	resourceLabel: string
): { key: string; params: Record<string, string> } | null {
	const merchant = (notification.metadata.merchant_name as string) || '';

	if (notification.type === 'expiry_reminder') {
		const expiresAt = metadataDate(notification, 'expires_at');
		if (expiresAt) {
			return {
				key: 'notifications.expiryReminderOn',
				params: { resource: resourceLabel, merchant, expires_at: expiresAt }
			};
		}

		const daysLeft = Number(notification.metadata.days_left);
		if (daysLeft === 0) {
			return {
				key: 'notifications.expiryReminderToday',
				params: { resource: resourceLabel, merchant }
			};
		}
		if (daysLeft === 1) {
			return {
				key: 'notifications.expiryReminderTomorrow',
				params: { resource: resourceLabel, merchant }
			};
		}
		return {
			key: 'notifications.expiryReminder',
			params: {
				resource: resourceLabel,
				merchant,
				days: String(notification.metadata.days_left)
			}
		};
	}

	if (notification.type === 'validity_start') {
		const validFrom = metadataDate(notification, 'valid_from');
		return validFrom
			? {
					key: 'notifications.validityStart',
					params: { merchant, valid_from: validFrom }
				}
			: { key: 'notifications.validityStartSoon', params: { merchant } };
	}

	return null;
}

/** Reads a preformatted date string the server stored in the metadata. */
function metadataDate(
	notification: NotificationDTO,
	key: string
): string | null {
	const value = notification.metadata[key];
	return typeof value === 'string' && value !== '' ? value : null;
}
