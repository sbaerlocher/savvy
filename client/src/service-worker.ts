/// <reference types="@sveltejs/kit" />
/// <reference no-default-lib="true"/>
/// <reference lib="esnext" />
/// <reference lib="webworker" />

import { precacheAndRoute, cleanupOutdatedCaches } from 'workbox-precaching';
import { NavigationRoute, registerRoute } from 'workbox-routing';
import { CacheFirst, StaleWhileRevalidate } from 'workbox-strategies';
import { ExpirationPlugin } from 'workbox-expiration';
import { CacheableResponsePlugin } from 'workbox-cacheable-response';

// Type definitions for Workbox
declare let self: ServiceWorkerGlobalScope & {
	__WB_MANIFEST: Array<{ url: string; revision: string | null }>;
};

// ========================================================================
// PRECACHE SETUP - Static Assets
// ========================================================================

// Precache all static assets (manifest injected by @vite-pwa/sveltekit injectManifest)
precacheAndRoute(self.__WB_MANIFEST);

// Cleanup outdated caches from previous versions
cleanupOutdatedCaches();

// ========================================================================
// NAVIGATION HANDLING - SPA Fallback (StaleWhileRevalidate)
// ========================================================================

// SPA navigation: serve index.html for all navigation requests.
// Uses StaleWhileRevalidate to serve the cached shell instantly while
// fetching an updated version in the background. Since this is an SPA,
// index.html is just a shell — the real content comes from API calls.
// The updated shell will be available on the next navigation.
registerRoute(
	new NavigationRoute(
		new StaleWhileRevalidate({
			cacheName: 'navigation-pages',
			plugins: [
				new CacheableResponsePlugin({ statuses: [200] }),
				{
					// Hard fallback: if StaleWhileRevalidate yields neither cache
					// nor network (e.g. cold Android homescreen-shortcut start
					// before the shell is cached), serve the app shell from the
					// navigation cache — any URL, or `/` — so the launch renders
					// instead of a white screen.
					handlerDidError: async () => {
						const cache = await caches.open('navigation-pages');
						return (
							(await cache.match('/')) ||
							(await cache.match('/index.html')) ||
							(await caches.match('/')) ||
							Response.error()
						);
					}
				}
			]
		}),
		{
			denylist: [
				/^\/api/, // Never handle API routes as navigation
				/^\/auth/,
				/^\/reset/,
				/^\/health/,
				/^\/ready/,
				/^\/metrics/
			]
		}
	)
);

// ========================================================================
// RUNTIME CACHING - JS/CSS/WASM Chunks (safety net for offline navigation)
// ========================================================================
// Precaching covers production builds. Runtime caching adds:
// - Dev mode support (where __WB_MANIFEST is empty)
// - Safety net for lazy-loaded or missed chunks in production

// Immutable SvelteKit chunks (content-hashed → safe to cache permanently)
registerRoute(
	({ url }) => url.pathname.startsWith('/_app/immutable/'),
	new CacheFirst({
		cacheName: 'svelte-immutable-chunks',
		plugins: [
			new CacheableResponsePlugin({ statuses: [0, 200] }),
			new ExpirationPlugin({
				maxEntries: 200,
				maxAgeSeconds: 365 * 24 * 60 * 60
			})
		]
	})
);

// Other SvelteKit app assets (e.g. version.json, non-immutable files)
registerRoute(
	({ url }) =>
		url.pathname.startsWith('/_app/') &&
		!url.pathname.startsWith('/_app/immutable/'),
	new StaleWhileRevalidate({
		cacheName: 'svelte-app-assets',
		plugins: [
			new CacheableResponsePlugin({ statuses: [0, 200] }),
			new ExpirationPlugin({ maxEntries: 50, maxAgeSeconds: 24 * 60 * 60 })
		]
	})
);

// Dev mode: Cache Vite module requests (/@fs/, /node_modules/.vite/, /src/)
// These are unbundled ES modules that Vite serves in development
registerRoute(
	({ url, request }) =>
		(request.destination === 'script' ||
			request.destination === 'style' ||
			url.pathname.endsWith('.js') ||
			url.pathname.endsWith('.ts') ||
			url.pathname.endsWith('.css')) &&
		(url.pathname.startsWith('/@') ||
			url.pathname.startsWith('/node_modules/') ||
			url.pathname.startsWith('/src/')),
	new StaleWhileRevalidate({
		cacheName: 'vite-dev-modules',
		plugins: [
			new CacheableResponsePlugin({ statuses: [0, 200] }),
			new ExpirationPlugin({ maxEntries: 500, maxAgeSeconds: 24 * 60 * 60 })
		]
	})
);

// ========================================================================
// DEBUG LOGGING - Fetch handler to trace cache hits/misses
// ========================================================================

self.addEventListener('fetch', (event) => {
	const url = new URL(event.request.url);

	// Only log relevant requests (skip noise like chrome-extension, data URIs)
	if (!url.protocol.startsWith('http')) return;

	// Skip API requests (handled by app layer, not SW)
	if (url.pathname.startsWith('/api/')) return;

	// Log navigation requests
	if (event.request.mode === 'navigate') {
		console.log(`[SW] Navigate: ${url.pathname}`);
		return;
	}

	// Log script/style requests (the ones that matter for offline navigation)
	if (
		event.request.destination === 'script' ||
		event.request.destination === 'style' ||
		url.pathname.endsWith('.js') ||
		url.pathname.endsWith('.ts') ||
		url.pathname.endsWith('.css')
	) {
		// Check if this request will be handled by a runtime cache
		const isImmutable = url.pathname.startsWith('/_app/immutable/');
		const isAppAsset = url.pathname.startsWith('/_app/') && !isImmutable;
		const isViteDev =
			url.pathname.startsWith('/@') ||
			url.pathname.startsWith('/node_modules/') ||
			url.pathname.startsWith('/src/');

		if (isImmutable || isAppAsset || isViteDev) {
			const cacheName = isImmutable
				? 'svelte-immutable-chunks'
				: isAppAsset
					? 'svelte-app-assets'
					: 'vite-dev-modules';
			// Async check — log whether it was a cache hit
			caches.open(cacheName).then((cache) =>
				cache.match(event.request).then((cached) => {
					const status = cached ? 'HIT' : 'MISS';
					const short =
						url.pathname.length > 60
							? '...' + url.pathname.slice(-57)
							: url.pathname;
					console.log(`[SW] ${status} [${cacheName}] ${short}`);
				})
			);
		}
	}
});

// ========================================================================
// SERVICE WORKER LIFECYCLE EVENTS
// ========================================================================

self.addEventListener('install', () => {
	console.log('[SW] Installing Service Worker...');
	// Note: Cannot reference __WB_MANIFEST here (injectManifest allows only one match)
	// Skip waiting immediately so security/bug fixes reach users without delay
	self.skipWaiting();
});

self.addEventListener('activate', (event) => {
	console.log('[SW] Activating Service Worker...');
	// Log cache contents for debugging
	caches.keys().then((names) => {
		console.log(`[SW] Active caches: ${names.join(', ') || '(none)'}`);
		names.forEach((name) => {
			caches.open(name).then((cache) =>
				cache.keys().then((keys) => {
					console.log(`[SW] Cache "${name}": ${keys.length} entries`);
				})
			);
		});
	});
	// Take control of all clients immediately
	event.waitUntil(self.clients.claim());
});

// ========================================================================
// MESSAGE HANDLING - Communication with App
// ========================================================================

self.addEventListener('message', (event) => {
	if (event.data && event.data.type === 'SKIP_WAITING') {
		console.log(
			'[SW] Received SKIP_WAITING message - activating new Service Worker'
		);
		self.skipWaiting().then(() => {
			console.log(
				'[SW] skipWaiting() completed - new Service Worker will activate'
			);
		});
	}
});

// ========================================================================
// PUSH NOTIFICATION HANDLING
// ========================================================================

self.addEventListener('push', (event) => {
	if (!event.data) return;

	try {
		const data = event.data.json();
		const options: NotificationOptions = {
			body: data.body || '',
			icon: data.icon || '/favicon.png',
			badge: '/favicon.png',
			data: { url: data.url || '/' }
		};

		event.waitUntil(
			self.registration.showNotification(data.title || 'Savvy', options)
		);
	} catch {
		// Fallback for plain text push
		event.waitUntil(
			self.registration.showNotification('Savvy', {
				body: event.data.text(),
				icon: '/favicon.png'
			})
		);
	}
});

self.addEventListener('notificationclick', (event) => {
	event.notification.close();

	let url = event.notification.data?.url || '/';

	// Validate URL is same-origin to prevent phishing via compromised push payloads
	try {
		const parsed = new URL(url, self.location.origin);
		if (parsed.origin !== self.location.origin) {
			url = '/';
		}
	} catch {
		url = '/';
	}

	event.waitUntil(
		self.clients
			.matchAll({ type: 'window', includeUncontrolled: true })
			.then((clientList) => {
				// Focus existing window if available
				for (const client of clientList) {
					if ('focus' in client) {
						client.focus();
						client.postMessage({ type: 'NAVIGATE', url });
						return;
					}
				}
				// Open new window
				return self.clients.openWindow(url);
			})
	);
});

console.log(
	'[SW] Service Worker loaded - precache + runtime caching + push + debug logging'
);
