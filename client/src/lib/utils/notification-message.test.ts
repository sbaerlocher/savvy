import { describe, it, expect } from 'vitest';
import { reminderMessageKey } from './notification-message';
import type { NotificationDTO } from '$lib/types/api';

function notification(
	type: string,
	metadata: Record<string, unknown>
): NotificationDTO {
	return { type, metadata } as unknown as NotificationDTO;
}

describe('reminderMessageKey', () => {
	it('renders the absolute date for expiry', () => {
		const result = reminderMessageKey(
			notification('expiry_reminder', {
				merchant_name: 'Coop',
				days_left: 3,
				expires_at: '26. Februar 2026'
			}),
			'voucher'
		);
		expect(result?.key).toBe('notifications.expiryReminderOn');
		expect(result?.params.expires_at).toBe('26. Februar 2026');
	});

	it('falls back to the day-zero wording when the date is missing', () => {
		const result = reminderMessageKey(
			notification('expiry_reminder', { merchant_name: 'Coop', days_left: 0 }),
			'voucher'
		);
		expect(result?.key).toBe('notifications.expiryReminderToday');
		expect(result?.params).not.toHaveProperty('days');
	});

	it('falls back to the singular wording on the last day', () => {
		const result = reminderMessageKey(
			notification('expiry_reminder', { merchant_name: 'Coop', days_left: 1 }),
			'voucher'
		);
		expect(result?.key).toBe('notifications.expiryReminderTomorrow');
		expect(result?.params).not.toHaveProperty('days');
	});

	it('falls back to the plural wording for more than one day', () => {
		const result = reminderMessageKey(
			notification('expiry_reminder', { merchant_name: 'Coop', days_left: 3 }),
			'voucher'
		);
		expect(result?.key).toBe('notifications.expiryReminder');
		expect(result?.params.days).toBe('3');
	});

	it('renders the absolute date for validity start', () => {
		const result = reminderMessageKey(
			notification('validity_start', {
				merchant_name: 'Coop',
				valid_from: '1. Januar 2026'
			}),
			'voucher'
		);
		expect(result?.key).toBe('notifications.validityStart');
		expect(result?.params.valid_from).toBe('1. Januar 2026');
	});

	it('falls back to dateless wording when valid_from is missing', () => {
		const result = reminderMessageKey(
			notification('validity_start', { merchant_name: 'Coop' }),
			'voucher'
		);
		expect(result?.key).toBe('notifications.validityStartSoon');
	});

	it('returns null for non-reminder notifications', () => {
		expect(
			reminderMessageKey(notification('share_received', {}), 'voucher')
		).toBeNull();
	});
});
