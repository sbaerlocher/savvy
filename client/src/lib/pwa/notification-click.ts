/**
 * Notification-click navigation for the service worker.
 *
 * Extracted from `service-worker.ts` so it can be unit tested: the service
 * worker module has top-level side effects and `self` references, which make
 * it non-importable from a test environment.
 */

/**
 * Resolves the raw URL from a push payload to a safe, same-origin target.
 *
 * Push payloads can be attacker-influenced, so a cross-origin or unparseable
 * URL falls back to the root path rather than being followed.
 */
export function resolveTargetURL(
	rawURL: string | undefined | null,
	origin: string
): string {
	if (!rawURL) {
		return '/';
	}

	try {
		const parsed = new URL(rawURL, origin);
		return parsed.origin === origin ? rawURL : '/';
	} catch {
		return '/';
	}
}

/** Minimal shape of the `WindowClient`s this handler works with. */
type NavigableClient = {
	url?: string;
	focus?: () => Promise<unknown> | unknown;
	navigate?: (url: string) => Promise<unknown>;
};

/** Minimal shape of `self.clients` used by {@link handleNotificationClick}. */
type ClientsLike = {
	matchAll: (options?: {
		type?: string;
		includeUncontrolled?: boolean;
	}) => Promise<NavigableClient[]>;
	openWindow: (url: string) => Promise<unknown>;
};

/**
 * Focuses an existing window and navigates it to `rawURL`, or opens a new one.
 *
 * The previous implementation only called `focus()` and then posted a
 * `{ type: 'NAVIGATE' }` message that no client ever listened for, so tapping a
 * notification while the app was already open surfaced the old view and dropped
 * the deep link — a white screen when the stale view could not render.
 *
 * `WindowClient.navigate()` requires Safari 16.4+, so a rejected or missing
 * `navigate` falls back to `openWindow()`.
 */
export async function handleNotificationClick(
	clients: ClientsLike,
	rawURL: string | undefined | null,
	origin: string
): Promise<void> {
	const url = resolveTargetURL(rawURL, origin);
	const clientList = await clients.matchAll({
		type: 'window',
		includeUncontrolled: true
	});

	// Prefer a client already on the target path, so an app that is already
	// showing the deep link is reused instead of navigated again.
	const target = new URL(url, origin).href;
	const client =
		clientList.find((c) => c.url && new URL(c.url, origin).href === target) ??
		clientList[0];

	if (client) {
		try {
			await client.focus?.();
			if (client.navigate) {
				await client.navigate(url);
				return;
			}
		} catch {
			// Fall through to openWindow below.
		}
	}

	await clients.openWindow(url);
}
