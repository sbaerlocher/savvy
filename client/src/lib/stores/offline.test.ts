import { describe, it, expect, beforeEach, vi } from 'vitest';
import { get } from 'svelte/store';
import { offlineStore, isOnline, showOfflineBanner } from './offline';

describe('offlineStore', () => {
	beforeEach(() => {
		vi.clearAllMocks();

		// Reset navigator.onLine to true
		Object.defineProperty(navigator, 'onLine', {
			writable: true,
			value: true
		});
	});

	describe('Online/Offline Detection', () => {
		it('should initialize with online status', () => {
			const state = get(offlineStore);
			expect(state.isOnline).toBe(true);
		});

		it('should update when going offline', () => {
			window.dispatchEvent(new Event('offline'));

			const state = get(offlineStore);
			expect(state.isOnline).toBe(false);
		});

		it('should update when going online', () => {
			// First go offline
			window.dispatchEvent(new Event('offline'));
			expect(get(offlineStore).isOnline).toBe(false);

			// Then go online
			window.dispatchEvent(new Event('online'));
			expect(get(offlineStore).isOnline).toBe(true);
		});
	});

	describe('Derived Stores', () => {
		it('isOnline should reflect current online status', () => {
			expect(get(isOnline)).toBe(true);

			window.dispatchEvent(new Event('offline'));
			expect(get(isOnline)).toBe(false);

			window.dispatchEvent(new Event('online'));
			expect(get(isOnline)).toBe(true);
		});

		it('showOfflineBanner should be inverse of isOnline', () => {
			expect(get(showOfflineBanner)).toBe(false);

			window.dispatchEvent(new Event('offline'));
			expect(get(showOfflineBanner)).toBe(true);

			window.dispatchEvent(new Event('online'));
			expect(get(showOfflineBanner)).toBe(false);
		});
	});
});
