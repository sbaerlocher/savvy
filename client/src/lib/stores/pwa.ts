import { writable, get } from 'svelte/store';

interface PWAState {
	needsRefresh: boolean;
	registration: ServiceWorkerRegistration | null;
	autoUpdateEnabled: boolean;
}

function createPWAStore() {
	const store = writable<PWAState>({
		needsRefresh: false,
		registration: null,
		autoUpdateEnabled: true // Enable auto-update by default
	});

	const { subscribe, update } = store;

	return {
		subscribe,

		setRegistration(registration: ServiceWorkerRegistration): void {
			update((state) => ({ ...state, registration }));
		},

		setNeedsRefresh(needsRefresh: boolean): void {
			update((state) => ({ ...state, needsRefresh }));
		},

		async updateServiceWorker(): Promise<void> {
			// With skipWaiting() in the SW install handler, updates are automatic.
			// This method triggers a manual check and reload if a new SW is waiting.
			const state = get(store);
			const currentRegistration = state.registration;

			if (currentRegistration) {
				await currentRegistration.update();
			}

			// The controllerchange listener in +layout.svelte handles the reload
			update((state) => ({ ...state, needsRefresh: false }));
		},

		/**
		 * Unregisters every service worker for this origin and registers a fresh
		 * one. Unlike updateServiceWorker() (an update check), this recovers a
		 * stuck/zombie registration — e.g. inside an installed desktop shortcut.
		 * Caches and IndexedDB are left untouched; reset.html covers the nuclear
		 * variant.
		 */
		async reregisterServiceWorker(): Promise<void> {
			if (!('serviceWorker' in navigator)) {
				return;
			}

			const registrations = await navigator.serviceWorker.getRegistrations();
			await Promise.all(registrations.map((r) => r.unregister()));

			update((state) => ({ ...state, registration: null }));

			const { registerServiceWorker } = await import('$lib/pwa/register-sw');
			await registerServiceWorker();
		},

		setAutoUpdate(enabled: boolean): void {
			update((state) => ({ ...state, autoUpdateEnabled: enabled }));
		}
	};
}

export const pwaStore = createPWAStore();
