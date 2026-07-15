import { resolve } from '$app/paths';

/**
 * Maps a resource type/kind to its resolved detail-route path.
 *
 * Uses the typed `resolve('/cards/[id]', { id })` overload so SvelteKit's
 * typed-route check stays intact. Unknown types fall back to `/dashboard`
 * (only reachable from callers with a loosely-typed `string`, e.g.
 * notification payloads).
 */
export function resourceDetailPath(type: string, id: string): string {
	switch (type) {
		case 'card':
			return resolve('/cards/[id]', { id });
		case 'voucher':
			return resolve('/vouchers/[id]', { id });
		case 'gift_card':
			return resolve('/gift-cards/[id]', { id });
		default:
			return resolve('/dashboard');
	}
}
