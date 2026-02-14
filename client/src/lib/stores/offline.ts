/**
 * Offline State Store
 *
 * Single source of truth for online/offline status.
 * Uses navigator.onLine + window events for real-time tracking.
 */

import { writable, derived } from 'svelte/store';
import { browser } from '$app/environment';

export interface OfflineState {
	isOnline: boolean;
}

function createOfflineStore() {
	const { subscribe, update } = writable<OfflineState>({
		isOnline: browser ? navigator.onLine : true
	});

	if (browser) {
		window.addEventListener('online', () =>
			update((s) => ({ ...s, isOnline: true }))
		);
		window.addEventListener('offline', () =>
			update((s) => ({ ...s, isOnline: false }))
		);
	}

	return { subscribe };
}

export const offlineStore = createOfflineStore();

export const isOnline = derived(offlineStore, ($s) => $s.isOnline);
export const showOfflineBanner = derived(offlineStore, ($s) => !$s.isOnline);
