import { describe, expect, it, vi } from 'vitest';
import {
	handleNotificationClick,
	resolveTargetURL
} from './notification-click';

const ORIGIN = 'https://savvy.example';

describe('resolveTargetURL', () => {
	it('keeps a same-origin relative path', () => {
		expect(resolveTargetURL('/gift-cards', ORIGIN)).toBe('/gift-cards');
	});

	it('keeps a same-origin absolute URL', () => {
		expect(resolveTargetURL(`${ORIGIN}/cards`, ORIGIN)).toBe(`${ORIGIN}/cards`);
	});

	it('rejects a cross-origin URL', () => {
		expect(resolveTargetURL('https://evil.example/phish', ORIGIN)).toBe('/');
	});

	it('falls back to root for undefined', () => {
		expect(resolveTargetURL(undefined, ORIGIN)).toBe('/');
	});

	it('falls back to root for an unparseable URL', () => {
		expect(resolveTargetURL('http://[not a url', ORIGIN)).toBe('/');
	});
});

/** Builds a `clients` double plus handles to assert against. */
function makeClients(
	windows: Array<{ url?: string; navigate?: () => Promise<unknown> }> = []
) {
	const openWindow = vi.fn().mockResolvedValue(undefined);
	const clientObjects = windows.map((w) => ({
		url: w.url,
		focus: vi.fn().mockResolvedValue(undefined),
		navigate: w.navigate
			? vi.fn(w.navigate)
			: vi.fn().mockResolvedValue(undefined)
	}));

	return {
		clients: { matchAll: vi.fn().mockResolvedValue(clientObjects), openWindow },
		clientObjects,
		openWindow
	};
}

describe('handleNotificationClick', () => {
	it('navigates an existing window instead of only focusing it', async () => {
		const { clients, clientObjects, openWindow } = makeClients([
			{ url: `${ORIGIN}/dashboard` }
		]);

		await handleNotificationClick(clients, '/gift-cards', ORIGIN);

		expect(clientObjects[0].focus).toHaveBeenCalled();
		expect(clientObjects[0].navigate).toHaveBeenCalledWith('/gift-cards');
		expect(openWindow).not.toHaveBeenCalled();
	});

	it('opens a new window when no client exists', async () => {
		const { clients, openWindow } = makeClients([]);

		await handleNotificationClick(clients, '/gift-cards', ORIGIN);

		expect(openWindow).toHaveBeenCalledWith('/gift-cards');
	});

	it('falls back to openWindow when navigate() rejects', async () => {
		const { clients, openWindow } = makeClients([
			{
				url: `${ORIGIN}/dashboard`,
				navigate: () => Promise.reject(new Error('unsupported'))
			}
		]);

		await handleNotificationClick(clients, '/gift-cards', ORIGIN);

		expect(openWindow).toHaveBeenCalledWith('/gift-cards');
	});

	it('prefers a window already on the target path', async () => {
		const { clients, clientObjects } = makeClients([
			{ url: `${ORIGIN}/dashboard` },
			{ url: `${ORIGIN}/gift-cards` }
		]);

		await handleNotificationClick(clients, '/gift-cards', ORIGIN);

		expect(clientObjects[1].navigate).toHaveBeenCalledWith('/gift-cards');
		expect(clientObjects[0].navigate).not.toHaveBeenCalled();
	});

	it('navigates to root for a cross-origin payload', async () => {
		const { clients, clientObjects } = makeClients([
			{ url: `${ORIGIN}/dashboard` }
		]);

		await handleNotificationClick(
			clients,
			'https://evil.example/phish',
			ORIGIN
		);

		expect(clientObjects[0].navigate).toHaveBeenCalledWith('/');
	});
});
