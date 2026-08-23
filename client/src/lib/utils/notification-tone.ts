import type { NotificationDTO } from '$lib/types/api';

/**
 * Desktop notification rows render the type icon inside a tinted tile. Each
 * notification type maps to one tone, which supplies the tile background, the
 * icon ink and the unread-dot colour. Mirrors the tone table in the desktop
 * mockup (screen-NotificationsDesktop).
 */
export type NotificationTone = {
	/** Tailwind background utility for the icon tile. */
	tile: string;
	/** Tailwind text utility for the icon itself. */
	ink: string;
	/** Tailwind background utility for the unread dot. */
	dot: string;
};

const TONES = {
	accent: {
		tile: 'bg-accent-50',
		ink: 'text-accent-700',
		dot: 'bg-accent-700'
	},
	transfer: {
		tile: 'bg-transfer-50',
		ink: 'text-transfer-600',
		dot: 'bg-transfer-600'
	},
	warning: {
		tile: 'bg-warning-50',
		ink: 'text-warning-700',
		dot: 'bg-warning-700'
	},
	success: {
		tile: 'bg-success-50',
		ink: 'text-success-700',
		dot: 'bg-success-700'
	}
} as const satisfies Record<string, NotificationTone>;

/** Resolves the tone for a notification type; unknown types fall back to accent. */
export function notificationTone(
	type: NotificationDTO['type']
): NotificationTone {
	switch (type) {
		case 'transfer_received':
			return TONES.transfer;
		case 'expiry_reminder':
			return TONES.warning;
		case 'validity_start':
			return TONES.success;
		default:
			return TONES.accent;
	}
}
