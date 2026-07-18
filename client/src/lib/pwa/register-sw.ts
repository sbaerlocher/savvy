import { get } from 'svelte/store';
import { t } from '$lib/stores/i18n';
import { pwaStore } from '$lib/stores/pwa';
import { toastStore } from '$lib/stores/toast';
import { logger } from '$lib/utils/logger';

const swLogger = logger.child('Layout');

// The controllerchange listener is global, not per-registration: attach it once
// per page load, otherwise re-registering stacks duplicate "update available"
// toasts.
let controllerChangeBound = false;

/**
 * Registers the service worker manually.
 *
 * injectRegister: "inline" doesn't work with adapter-static, so registration is
 * done imperatively here. May run more than once per page load (re-registration
 * from the profile page), so i18n is read via get(t) rather than the reactive $t.
 *
 * Returns true when a service worker is registered, false when registration was
 * skipped (no browser support, SW file unreachable) or failed — callers use this
 * to avoid reporting success for a registration that never happened.
 */
export async function registerServiceWorker(): Promise<boolean> {
	if (!('serviceWorker' in navigator)) {
		return false;
	}

	// SvelteKit generates service-worker.js from src/service-worker.ts
	const swUrl = import.meta.env.DEV
		? '/dev-sw.js?dev-sw'
		: '/service-worker.js';

	// Verify the SW file actually exists before registering.
	// Prevents zombie registrations when accessing the Go dev server (port 8080)
	// which doesn't serve static assets — only the Vite dev server (5173) or
	// production builds have the SW file.
	try {
		const swCheck = await fetch(swUrl, { method: 'HEAD' });
		if (!swCheck.ok) {
			swLogger.debug(
				`Service Worker not available (${swCheck.status}) - skipping registration`
			);
			return false;
		}
	} catch {
		swLogger.debug('Service Worker not reachable - skipping registration');
		return false;
	}

	try {
		const registration = await navigator.serviceWorker.register(swUrl, {
			type: import.meta.env.DEV ? 'module' : 'classic'
		});
		swLogger.info(`Service Worker registered: ${swUrl}`);
		pwaStore.setRegistration(registration);

		// Check if SW is installing (first time install)
		if (registration.installing) {
			swLogger.info(
				'Service Worker installing (first time) - precache will run'
			);
			registration.installing.addEventListener('statechange', (e) => {
				const target = e.target as ServiceWorker;
				if (target.state === 'activated') {
					swLogger.info('Service Worker activated - precache completed');
				}
			});
		}

		// Check for updates
		registration.update();

		// Notify user when new SW takes control (skipWaiting is automatic)
		// Skip in dev mode — Vite HMR constantly regenerates the SW, causing reload loops
		if (!import.meta.env.DEV && !controllerChangeBound) {
			controllerChangeBound = true;
			let refreshing = false;
			navigator.serviceWorker.addEventListener('controllerchange', () => {
				if (refreshing) return;
				refreshing = true;
				swLogger.info('New Service Worker activated - update available');
				toastStore.info(get(t)('pwa.updateAvailable'));
			});
		}

		return true;
	} catch (error) {
		swLogger.error('Service Worker registration failed:', error);
		toastStore.warning(get(t)('pwa.offlineUnavailable'));
		return false;
	}
}
