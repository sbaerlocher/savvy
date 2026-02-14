import { writable, get } from 'svelte/store';
import { pushApi } from '$lib/api/push';
import { logger } from '$lib/utils/logger';

const pushLogger = logger.child('PushStore');

interface PushState {
	isSupported: boolean;
	isSubscribed: boolean;
	permission: NotificationPermission | 'default';
	isLoading: boolean;
}

function createPushStore() {
	const { subscribe, set, update } = writable<PushState>({
		isSupported: false,
		isSubscribed: false,
		permission: 'default',
		isLoading: false
	});

	async function getRegistration(): Promise<ServiceWorkerRegistration | null> {
		if (!('serviceWorker' in navigator)) return null;
		return navigator.serviceWorker.ready;
	}

	return {
		subscribe,

		async init(): Promise<void> {
			const isSupported =
				'serviceWorker' in navigator &&
				'PushManager' in window &&
				'Notification' in window;

			if (!isSupported) {
				set({
					isSupported: false,
					isSubscribed: false,
					permission: 'default',
					isLoading: false
				});
				return;
			}

			const permission = Notification.permission;
			let isSubscribed = false;

			try {
				const registration = await getRegistration();
				if (registration) {
					const sub = await registration.pushManager.getSubscription();
					isSubscribed = sub !== null;
				}
			} catch (error) {
				pushLogger.warn('Failed to check push subscription', { error });
			}

			set({ isSupported, isSubscribed, permission, isLoading: false });
		},

		async enablePush(): Promise<boolean> {
			update((s) => ({ ...s, isLoading: true }));

			try {
				// Get VAPID key from server
				const { public_key } = await pushApi.getVAPIDKey();
				if (!public_key) {
					pushLogger.warn('No VAPID key available');
					update((s) => ({ ...s, isLoading: false }));
					return false;
				}

				// Request notification permission
				const permission = await Notification.requestPermission();
				if (permission !== 'granted') {
					update((s) => ({ ...s, permission, isLoading: false }));
					return false;
				}

				// Subscribe to push
				const registration = await getRegistration();
				if (!registration) {
					update((s) => ({ ...s, isLoading: false }));
					return false;
				}

				const applicationServerKey = urlBase64ToUint8Array(public_key)
					.buffer as ArrayBuffer;
				const subscription = await registration.pushManager.subscribe({
					userVisibleOnly: true,
					applicationServerKey
				});

				// Send subscription to server
				await pushApi.subscribe(subscription);

				update((s) => ({
					...s,
					isSubscribed: true,
					permission: 'granted',
					isLoading: false
				}));
				pushLogger.info('Push notifications enabled');
				return true;
			} catch (error) {
				pushLogger.error('Failed to enable push notifications', { error });
				update((s) => ({ ...s, isLoading: false }));
				return false;
			}
		},

		async disablePush(): Promise<boolean> {
			update((s) => ({ ...s, isLoading: true }));

			try {
				const registration = await getRegistration();
				if (registration) {
					const subscription = await registration.pushManager.getSubscription();
					if (subscription) {
						await pushApi.unsubscribe(subscription.endpoint);
						await subscription.unsubscribe();
					}
				}

				update((s) => ({ ...s, isSubscribed: false, isLoading: false }));
				pushLogger.info('Push notifications disabled');
				return true;
			} catch (error) {
				pushLogger.error('Failed to disable push notifications', { error });
				update((s) => ({ ...s, isLoading: false }));
				return false;
			}
		},

		async toggle(): Promise<boolean> {
			const state = get({ subscribe });
			if (state.isSubscribed) {
				return this.disablePush();
			}
			return this.enablePush();
		}
	};
}

// Convert VAPID key from base64url to Uint8Array
function urlBase64ToUint8Array(base64String: string): Uint8Array {
	const padding = '='.repeat((4 - (base64String.length % 4)) % 4);
	const base64 = (base64String + padding).replace(/-/g, '+').replace(/_/g, '/');

	const rawData = window.atob(base64);
	const outputArray = new Uint8Array(rawData.length);

	for (let i = 0; i < rawData.length; ++i) {
		outputArray[i] = rawData.charCodeAt(i);
	}
	return outputArray;
}

export const pushStore = createPushStore();
