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

		setAutoUpdate(enabled: boolean): void {
			update((state) => ({ ...state, autoUpdateEnabled: enabled }));
		}
	};
}

export const pwaStore = createPWAStore();
