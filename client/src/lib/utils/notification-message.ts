import type { NotificationDTO } from '$lib/types/api';

/**
 * Picks the translation key and params for a reminder notification body.
 *
 * Reminder bodies are rendered in two places (the notification list and the
 * panel). The key choice lives here so both stay in sync: the expiry body
 * needs a singular form for the last day, and the validity-start body needs
 * the absolute date from the metadata rather than a relative "tomorrow" that
 * ages out of date while the notification sits in the list.
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
		const daysLeft = Number(notification.metadata.days_left);
		return daysLeft === 1
			? {
					key: 'notifications.expiryReminderTomorrow',
					params: { resource: resourceLabel, merchant }
				}
			: {
					key: 'notifications.expiryReminder',
					params: {
						resource: resourceLabel,
						merchant,
						days: String(notification.metadata.days_left)
					}
				};
	}

	if (notification.type === 'validity_start') {
		const validFrom = notification.metadata.valid_from;
		// Notifications created before the client rendered valid_from have no
		// date in their metadata; fall back to wording that needs none.
		return typeof validFrom === 'string' && validFrom !== ''
			? {
					key: 'notifications.validityStart',
					params: { merchant, valid_from: validFrom }
				}
			: { key: 'notifications.validityStartSoon', params: { merchant } };
	}

	return null;
}
