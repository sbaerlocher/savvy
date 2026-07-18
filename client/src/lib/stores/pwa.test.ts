import { beforeEach, describe, expect, it, vi } from 'vitest';
import { get } from 'svelte/store';

const registerServiceWorker = vi.fn();
vi.mock('$lib/pwa/register-sw', () => ({
	registerServiceWorker: () => registerServiceWorker()
}));

import { pwaStore } from './pwa';

describe('pwaStore.reregisterServiceWorker', () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	it('unregisters every registration, clears state and registers again', async () => {
		const unregisterA = vi.fn().mockResolvedValue(true);
		const unregisterB = vi.fn().mockResolvedValue(true);

		pwaStore.setRegistration({
			unregister: unregisterA
		} as unknown as ServiceWorkerRegistration);

		vi.stubGlobal('navigator', {
			serviceWorker: {
				getRegistrations: vi
					.fn()
					.mockResolvedValue([
						{ unregister: unregisterA },
						{ unregister: unregisterB }
					])
			}
		});

		registerServiceWorker.mockResolvedValue(true);

		await expect(pwaStore.reregisterServiceWorker()).resolves.toBe(true);

		expect(unregisterA).toHaveBeenCalledOnce();
		expect(unregisterB).toHaveBeenCalledOnce();
		expect(get(pwaStore).registration).toBeNull();
		expect(registerServiceWorker).toHaveBeenCalledOnce();
	});

	it('reports failure when no fresh worker could be registered', async () => {
		vi.stubGlobal('navigator', {
			serviceWorker: {
				getRegistrations: vi.fn().mockResolvedValue([])
			}
		});
		registerServiceWorker.mockResolvedValue(false);

		await expect(pwaStore.reregisterServiceWorker()).resolves.toBe(false);
	});

	it('is a no-op without service worker support', async () => {
		vi.stubGlobal('navigator', {});

		await expect(pwaStore.reregisterServiceWorker()).resolves.toBe(false);

		expect(registerServiceWorker).not.toHaveBeenCalled();
	});
});
