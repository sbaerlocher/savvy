<script lang="ts">
	/* global __APP_VERSION__ -- compile-time constant injected by Vite `define` (see vite.config.ts, typed in src/app.d.ts) */
	import { page } from '$app/stores';
	import { resolve } from '$app/paths';
	import { goto } from '$app/navigation';
	import MobileNav from '$lib/components/MobileNav.svelte';
	import OfflineIndicator from '$lib/components/OfflineIndicator.svelte';
	import Toast from '$lib/components/Toast.svelte';
	import NotificationPanel from '$lib/components/NotificationPanel.svelte';
	import TypeChoiceDialog from '$lib/components/TypeChoiceDialog.svelte';
	import { authStore } from '$lib/stores/auth';
	import { configStore } from '$lib/stores/config';
	import { t } from '$lib/stores/i18n';
	import { pwaStore } from '$lib/stores/pwa';
	import { toastStore } from '$lib/stores/toast';
	import { showOfflineBanner } from '$lib/stores/offline';
	import { browser } from '$app/environment';
	import { onMount, type Snippet } from 'svelte';
	import { logger } from '$lib/utils/logger';
	import { platform } from '$lib/utils/platform';
	import '../app.css';
	import type { LayoutData } from './$types';

	const layoutLogger = logger.child('Layout');

	let { data, children }: { data: LayoutData; children?: Snippet } = $props();

	const showMobileNav = $derived(
		$authStore.isAuthenticated &&
			!$page.url.pathname.startsWith('/login') &&
			!$page.url.pathname.startsWith('/register')
	);

	// Track preload state to prevent duplicate runs
	let preloadStarted = false;

	// Reactive: Trigger preload whenever data.user becomes available
	// This handles both initial page load AND client-side navigation after login
	// (onMount only fires once, so it misses goto('/dashboard') after login)
	$effect(() => {
		if (data.user) {
			authStore.setUser(data.user);

			if (!preloadStarted && browser && navigator.onLine) {
				preloadStarted = true;
				import('$lib/utils/preload').then(({ preloadEverything }) => {
					preloadEverything()
						.then(() => {
							layoutLogger.info(
								'Offline data preloaded! You can now use the app offline.'
							);
						})
						.catch((error) => {
							layoutLogger.warn(
								'Preload failed (some data may not be available offline):',
								error
							);
						});
				});
			}
		} else if (browser) {
			// For public routes or after logout
			preloadStarted = false;
			// Only check auth if localStorage says we were previously logged in
			// (validates session is still valid). Skip if never logged in to avoid 401.
			const cached = authStore.getCachedAuth();
			if (navigator.onLine && cached?.isAuthenticated) {
				authStore.checkAuth();
			}
		}
	});

	onMount(async () => {
		// Load app config
		await configStore.load();

		// Show app version in console
		layoutLogger.info(`Savvy v${__APP_VERSION__}`);
		layoutLogger.info(
			`PWA Auto-Update: ${$pwaStore.autoUpdateEnabled ? 'Enabled' : 'Disabled'}`
		);

		// Register Service Worker manually (injectRegister: "inline" doesn't work with adapter-static)
		if ('serviceWorker' in navigator) {
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
					layoutLogger.debug(
						`Service Worker not available (${swCheck.status}) - skipping registration`
					);
					return;
				}
			} catch {
				layoutLogger.debug(
					'Service Worker not reachable - skipping registration'
				);
				return;
			}

			try {
				const registration = await navigator.serviceWorker.register(swUrl, {
					type: import.meta.env.DEV ? 'module' : 'classic'
				});
				layoutLogger.info(`Service Worker registered: ${swUrl}`);
				pwaStore.setRegistration(registration);

				// Check if SW is installing (first time install)
				if (registration.installing) {
					layoutLogger.info(
						'Service Worker installing (first time) - precache will run'
					);
					registration.installing.addEventListener('statechange', (e) => {
						const target = e.target as ServiceWorker;
						if (target.state === 'activated') {
							layoutLogger.info(
								'Service Worker activated - precache completed'
							);
						}
					});
				}

				// Check for updates
				registration.update();

				// Notify user when new SW takes control (skipWaiting is automatic)
				// Skip in dev mode — Vite HMR constantly regenerates the SW, causing reload loops
				if (!import.meta.env.DEV) {
					let refreshing = false;
					navigator.serviceWorker.addEventListener('controllerchange', () => {
						if (refreshing) return;
						refreshing = true;
						layoutLogger.info(
							'New Service Worker activated - update available'
						);
						toastStore.info($t('pwa.updateAvailable'));
					});
				}
			} catch (error) {
				layoutLogger.error('Service Worker registration failed:', error);
				toastStore.warning($t('pwa.offlineUnavailable'));
			}
		}
	});

	async function handleLogout() {
		showUserMenu = false;
		await authStore.logout();
		window.location.href = '/login';
	}

	let showUserMenu = $state(false);
	let showAdminMenu = $state(false);
	let showNewDialog = $state(false);
	let desktopSearch = $state('');

	// Desktop search submits into Wallet (single global search entry).
	function submitDesktopSearch(e: SubmitEvent) {
		e.preventDefault();
		const q = desktopSearch.trim();
		const query = q ? `?search=${encodeURIComponent(q)}` : '?search=1';
		/* eslint-disable-next-line svelte/no-navigation-without-resolve -- base is resolve()d; the rest is a query string */
		goto(`${resolve('/wallet')}${query}`);
	}

	async function handleStopImpersonation() {
		layoutLogger.debug('Stop impersonation button clicked');
		try {
			await authStore.stopImpersonation();
			// Redirect happens in authStore.stopImpersonation
		} catch (err: unknown) {
			layoutLogger.error('Failed to stop impersonation', { error: err });
			const message =
				err instanceof Error
					? err.message
					: $t('admin.impersonate_info.stop_failed');
			toastStore.error(message || $t('admin.impersonate_info.stop_failed'));
		}
	}
</script>

<!-- Offline Indicator (über der Navigation) -->
<OfflineIndicator />

<!-- Desktop Navigation -->
{#if $authStore.isAuthenticated && !$page.url.pathname.startsWith('/login') && !$page.url.pathname.startsWith('/register')}
	<nav
		class="bg-white border-b border-gray-200 transition-all duration-300 ease-out"
		class:mt-16={$showOfflineBanner}
		class:sm:mt-12={$showOfflineBanner}
	>
		<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
			<div class="flex justify-between h-16">
				<div class="flex items-center">
					<div class="flex-shrink-0">
						<a href={resolve('/dashboard')} class="flex items-center space-x-3">
							<img src="/logo.png" alt="Savvy Logo" class="h-12 sm:h-16" />
							<span class="hidden sm:inline text-2xl font-bold text-cyan-600"
								>{$t('common.appName')}</span
							>
						</a>
					</div>

					<!-- App shell: three places — Start · Wallet · Profile -->
					<div class="hidden sm:ml-6 sm:flex sm:space-x-8">
						<a
							href={resolve('/dashboard')}
							data-testid="nav-start-desktop"
							class="border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700 inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium transition-colors"
							class:border-cyan-600={$page.url.pathname.startsWith(
								'/dashboard'
							)}
							class:text-gray-900={$page.url.pathname.startsWith('/dashboard')}
						>
							{$t('nav.start')}
						</a>
						<a
							href={resolve('/wallet')}
							data-testid="nav-wallet-desktop"
							class="border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700 inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium transition-colors"
							class:border-cyan-600={$page.url.pathname.startsWith('/wallet')}
							class:text-gray-900={$page.url.pathname.startsWith('/wallet')}
						>
							{$t('nav.wallet')}
						</a>
						<a
							href={resolve('/profile')}
							data-testid="nav-profile-desktop"
							class="border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700 inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium transition-colors"
							class:border-cyan-600={$page.url.pathname.startsWith('/profile')}
							class:text-gray-900={$page.url.pathname.startsWith('/profile')}
						>
							{$t('nav.profile')}
						</a>
					</div>
				</div>

				<div class="flex items-center space-x-4">
					<!-- Desktop: inline search + New (single global entries) -->
					<form
						onsubmit={submitDesktopSearch}
						class="hidden sm:flex items-center"
						role="search"
					>
						<input
							type="search"
							bind:value={desktopSearch}
							placeholder={$t('common.search')}
							aria-label={$t('common.search')}
							class="w-48 px-3 py-1.5 text-sm bg-white border border-gray-300 rounded-md focus:ring-cyan-500 focus:border-cyan-500"
						/>
					</form>
					<button
						type="button"
						onclick={() => (showNewDialog = true)}
						data-testid="nav-new-desktop"
						class="hidden sm:inline-flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium bg-cyan-600 text-white rounded-md hover:bg-cyan-700 transition-colors"
					>
						<svg
							class="w-4 h-4"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M12 4v16m8-8H4"
							/>
						</svg>
						{$t('common.new')}
					</button>

					<!-- Mobile: iOS shows header "+" (New); Android shows search (New = FAB) -->
					{#if platform === 'ios'}
						<button
							type="button"
							onclick={() => (showNewDialog = true)}
							data-testid="nav-new-mobile"
							aria-label={$t('common.new')}
							class="sm:hidden inline-flex items-center justify-center text-cyan-600 p-1"
						>
							<svg
								class="w-6 h-6"
								fill="none"
								stroke="currentColor"
								viewBox="0 0 24 24"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M12 4v16m8-8H4"
								/>
							</svg>
						</button>
					{:else}
						<!-- eslint-disable svelte/no-navigation-without-resolve -- base is resolve()d; ?search is a query string -->
						<a
							href={resolve('/wallet') + '?search=1'}
							data-testid="nav-search-mobile"
							aria-label={$t('common.search')}
							class="sm:hidden inline-flex items-center justify-center text-gray-600 p-1"
						>
							<!-- eslint-enable svelte/no-navigation-without-resolve -->
							<svg
								class="w-6 h-6"
								fill="none"
								stroke="currentColor"
								viewBox="0 0 24 24"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
								/>
							</svg>
						</a>
					{/if}

					<!-- Impersonation Indicator -->
					{#if $authStore.user?.is_impersonating}
						<div
							class="hidden sm:flex items-center bg-yellow-50 border border-yellow-300 rounded-lg px-3 py-2 gap-2"
						>
							<!-- Admin Icon -->
							<svg
								class="w-5 h-5 text-yellow-700"
								fill="none"
								stroke="currentColor"
								viewBox="0 0 24 24"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"
								></path>
							</svg>

							<!-- Arrow -->
							<svg
								class="w-4 h-4 text-yellow-600"
								fill="none"
								stroke="currentColor"
								viewBox="0 0 24 24"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M13 7l5 5m0 0l-5 5m5-5H6"
								></path>
							</svg>

							<!-- User Icon + Email -->
							<div class="flex items-center gap-1.5">
								<svg
									class="w-5 h-5 text-yellow-700"
									fill="none"
									stroke="currentColor"
									viewBox="0 0 24 24"
								>
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										stroke-width="2"
										d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"
									></path>
								</svg>
								<span class="text-sm text-yellow-800 font-medium">
									{$authStore.user.email}
								</span>
							</div>

							<!-- Stop Button -->
							<button
								type="button"
								onclick={(e) => {
									layoutLogger.debug('Button onclick handler triggered');
									e.stopPropagation();
									layoutLogger.debug('About to call handleStopImpersonation');
									handleStopImpersonation();
								}}
								class="ml-2 text-xs bg-yellow-600 hover:bg-yellow-700 text-white px-3 py-1 rounded font-medium transition-colors"
							>
								{$t('admin.stop_impersonating')}
							</button>
						</div>
					{/if}

					<!-- Admin Dropdown (only for admins, hidden during impersonation; desktop only — mockup mobile header has no admin controls) -->
					{#if !$authStore.user?.is_impersonating}
						{#if data.user?.is_admin && !$authStore.user?.is_impersonating}
							<div class="relative hidden sm:block">
								<button
									type="button"
									onclick={(e) => {
										e.stopPropagation();
										showAdminMenu = !showAdminMenu;
									}}
									class="inline-flex text-purple-600 hover:text-purple-700 p-2 rounded-md hover:bg-purple-50 transition-colors"
									title={$t('nav.admin')}
									aria-expanded={showAdminMenu}
								>
									<svg
										class="w-6 h-6"
										fill="none"
										stroke="currentColor"
										viewBox="0 0 24 24"
									>
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											stroke-width="2"
											d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"
										></path>
									</svg>
								</button>

								{#if showAdminMenu}
									<div
										role="menu"
										tabindex="-1"
										class="absolute right-0 mt-2 w-56 rounded-lg shadow-xl py-2 z-50 {platform ===
										'ios'
											? 'bg-white/70 backdrop-blur-xl backdrop-saturate-150 border border-white/30'
											: 'bg-white border border-gray-200'}"
										onclick={(e) => e.stopPropagation()}
										onkeydown={(e) => {
											if (e.key === 'Escape') showAdminMenu = false;
										}}
									>
										<div
											class="px-4 py-3 text-sm border-b {platform === 'ios'
												? 'border-white/30'
												: 'border-gray-200'}"
										>
											<div class="font-semibold text-purple-700">
												{$t('nav.admin')}
											</div>
										</div>

										<a
											href={resolve('/admin/users')}
											onclick={() => (showAdminMenu = false)}
											class="flex items-center w-full px-4 py-3 text-sm text-gray-700 hover:bg-gray-50 transition-colors"
										>
											<svg
												class="w-5 h-5 mr-3"
												fill="none"
												stroke="currentColor"
												viewBox="0 0 24 24"
											>
												<path
													stroke-linecap="round"
													stroke-linejoin="round"
													stroke-width="2"
													d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z"
												></path>
											</svg>
											{$t('nav.adminUsers')}
										</a>

										<a
											href={resolve('/admin/audit-log')}
											onclick={() => (showAdminMenu = false)}
											class="flex items-center w-full px-4 py-3 text-sm text-gray-700 hover:bg-gray-50 transition-colors"
										>
											<svg
												class="w-5 h-5 mr-3"
												fill="none"
												stroke="currentColor"
												viewBox="0 0 24 24"
											>
												<path
													stroke-linecap="round"
													stroke-linejoin="round"
													stroke-width="2"
													d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01"
												></path>
											</svg>
											{$t('nav.adminAuditLog')}
										</a>

										<a
											href={resolve('/admin/system-health')}
											onclick={() => (showAdminMenu = false)}
											class="flex items-center w-full px-4 py-3 text-sm text-gray-700 hover:bg-gray-50 transition-colors"
										>
											<svg
												class="w-5 h-5 mr-3"
												fill="none"
												stroke="currentColor"
												viewBox="0 0 24 24"
											>
												<path
													stroke-linecap="round"
													stroke-linejoin="round"
													stroke-width="2"
													d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
												></path>
											</svg>
											{$t('nav.adminSystemHealth')}
										</a>

										{#if $configStore.is_development}
											<a
												href={resolve('/admin/email-templates')}
												onclick={() => (showAdminMenu = false)}
												class="flex items-center w-full px-4 py-3 text-sm text-gray-700 hover:bg-gray-50 transition-colors"
											>
												<svg
													class="w-5 h-5 mr-3"
													fill="none"
													stroke="currentColor"
													viewBox="0 0 24 24"
												>
													<path
														stroke-linecap="round"
														stroke-linejoin="round"
														stroke-width="2"
														d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
													></path>
												</svg>
												{$t('nav.adminEmailTemplates')}
											</a>
										{/if}
									</div>
								{/if}
							</div>
						{/if}

						<!-- Notifications (hidden during impersonation) -->
						<!-- Mobile: bell sits before the "+" to match the mockup order (bell · +); desktop keeps its natural position. -->
						{#if !$authStore.user?.is_impersonating}
							<div class="relative order-first sm:order-none">
								<NotificationPanel />
							</div>
						{/if}
					{/if}

					<!-- User Menu (hidden during impersonation; desktop only — Profile is a bottom-nav place on mobile, so no redundant user icon in the mobile header) -->
					{#if !$authStore.user?.is_impersonating}
						<div class="relative hidden sm:block">
							<button
								type="button"
								onclick={(e) => {
									e.stopPropagation();
									showUserMenu = !showUserMenu;
								}}
								class="flex items-center text-sm text-gray-500 hover:text-gray-700 transition-colors"
								aria-label={$t('aria.openUserMenu')}
								aria-expanded={showUserMenu}
							>
								<svg
									class="w-6 h-6"
									fill="none"
									stroke="currentColor"
									viewBox="0 0 24 24"
								>
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										stroke-width="2"
										d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"
									></path>
								</svg>
							</button>

							{#if showUserMenu}
								<div
									role="menu"
									tabindex="-1"
									class="absolute right-0 mt-2 w-56 rounded-lg shadow-xl py-2 z-50 {platform ===
									'ios'
										? 'bg-white/70 backdrop-blur-xl backdrop-saturate-150 border border-white/30'
										: 'bg-white border border-gray-200'}"
									onclick={(e) => e.stopPropagation()}
									onkeydown={(e) => {
										if (e.key === 'Escape') showUserMenu = false;
									}}
								>
									{#if $authStore.user}
										<div
											class="px-4 py-3 text-sm border-b {platform === 'ios'
												? 'border-white/30'
												: 'border-gray-200'}"
										>
											{#if $authStore.user.first_name || $authStore.user.last_name}
												<div class="font-semibold text-gray-900 mb-1">
													{$authStore.user.first_name || ''}
													{$authStore.user.last_name || ''}
												</div>
											{/if}
											<div class="text-xs text-gray-600 break-all">
												{$authStore.user.email}
											</div>
										</div>
									{/if}

									<a
										href={resolve('/profile')}
										onclick={() => (showUserMenu = false)}
										class="flex items-center w-full px-4 py-3 text-sm text-gray-700 hover:bg-gray-50 transition-colors"
									>
										<svg
											class="w-5 h-5 mr-3"
											fill="none"
											stroke="currentColor"
											viewBox="0 0 24 24"
										>
											<path
												stroke-linecap="round"
												stroke-linejoin="round"
												stroke-width="2"
												d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"
											></path>
										</svg>
										{$t('nav.profile')}
									</a>

									<a
										href={resolve('/security')}
										onclick={() => (showUserMenu = false)}
										class="flex items-center w-full px-4 py-3 text-sm text-gray-700 hover:bg-gray-50 transition-colors"
									>
										<svg
											class="w-5 h-5 mr-3"
											fill="none"
											stroke="currentColor"
											viewBox="0 0 24 24"
										>
											<path
												stroke-linecap="round"
												stroke-linejoin="round"
												stroke-width="2"
												d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"
											></path>
										</svg>
										{$t('nav.security')}
									</a>

									<a
										href={resolve('/notifications')}
										onclick={() => (showUserMenu = false)}
										class="flex items-center w-full px-4 py-3 text-sm text-gray-700 hover:bg-gray-50 transition-colors"
									>
										<svg
											class="w-5 h-5 mr-3"
											fill="none"
											stroke="currentColor"
											viewBox="0 0 24 24"
										>
											<path
												stroke-linecap="round"
												stroke-linejoin="round"
												stroke-width="2"
												d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"
											></path>
										</svg>
										{$t('nav.notifications')}
									</a>

									<button
										type="button"
										onclick={handleLogout}
										class="flex items-center w-full px-4 py-3 text-sm text-red-600 hover:bg-gray-50 transition-colors"
									>
										<svg
											class="w-5 h-5 mr-3"
											fill="none"
											stroke="currentColor"
											viewBox="0 0 24 24"
										>
											<path
												stroke-linecap="round"
												stroke-linejoin="round"
												stroke-width="2"
												d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"
											></path>
										</svg>
										{$t('nav.logout')}
									</button>
								</div>
							{/if}
						</div>
					{/if}
				</div>
			</div>
		</div>
	</nav>
{/if}

<main
	class="max-w-7xl mx-auto pt-14 pb-6 px-4 sm:px-6 lg:px-8"
	class:pt-4={!$authStore.isAuthenticated}
	class:main-with-mobile-nav={showMobileNav && platform !== 'ios'}
	class:main-with-mobile-nav-floating={showMobileNav && platform === 'ios'}
>
	{@render children?.()}
</main>

<footer
	class="bg-gradient-to-b from-gray-50 to-gray-100 border-t border-gray-200 mt-12 hidden sm:block"
>
	<div class="max-w-7xl mx-auto py-12 px-4 sm:px-6 lg:px-8">
		<div class="grid grid-cols-1 md:grid-cols-4 gap-8">
			<!-- App Info -->
			<div class="space-y-4">
				<h3 class="text-lg font-bold text-gray-900">{$t('common.appName')}</h3>
				<p class="text-sm text-gray-600 leading-relaxed">
					{$t('footer.description')}
				</p>
				<div class="flex space-x-3">
					<a
						href="https://github.com/sbaerlocher/savvy"
						target="_blank"
						rel="noopener noreferrer"
						class="text-gray-500 hover:text-gray-700 transition-colors"
						aria-label={$t('aria.githubLink')}
					>
						<svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
							<path
								fill-rule="evenodd"
								d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z"
								clip-rule="evenodd"
							></path>
						</svg>
					</a>
				</div>
			</div>

			<!-- Quick Links -->
			<div class="space-y-4">
				<h3
					class="text-sm font-semibold text-gray-900 uppercase tracking-wider"
				>
					{$t('footer.quickLinks')}
				</h3>
				<ul class="space-y-2">
					<li>
						<a
							href={resolve('/dashboard')}
							class="text-sm text-gray-600 hover:text-cyan-600 transition-colors"
						>
							{$t('nav.start')}
						</a>
					</li>
					<li>
						<a
							href={resolve('/wallet')}
							class="text-sm text-gray-600 hover:text-cyan-600 transition-colors"
						>
							{$t('nav.wallet')}
						</a>
					</li>
					<li>
						<a
							href={resolve('/profile')}
							class="text-sm text-gray-600 hover:text-cyan-600 transition-colors"
						>
							{$t('nav.profile')}
						</a>
					</li>
				</ul>
			</div>

			<!-- Resources -->
			<div class="space-y-4">
				<h3
					class="text-sm font-semibold text-gray-900 uppercase tracking-wider"
				>
					{$t('footer.resources')}
				</h3>
				<ul class="space-y-2">
					<li>
						<a
							href="https://github.com/sbaerlocher/savvy#readme"
							target="_blank"
							rel="noopener noreferrer"
							class="text-sm text-gray-600 hover:text-cyan-600 transition-colors"
						>
							{$t('footer.documentation')}
						</a>
					</li>
					<li>
						<a
							href="https://github.com/sbaerlocher/savvy/issues"
							target="_blank"
							rel="noopener noreferrer"
							class="text-sm text-gray-600 hover:text-cyan-600 transition-colors"
						>
							{$t('footer.reportBug')}
						</a>
					</li>
					<li>
						<a
							href="https://github.com/sbaerlocher/savvy/releases"
							target="_blank"
							rel="noopener noreferrer"
							class="text-sm text-gray-600 hover:text-cyan-600 transition-colors"
						>
							{$t('footer.changelog')}
						</a>
					</li>
				</ul>
			</div>

			<!-- Tech Stack -->
			<div class="space-y-4">
				<h3
					class="text-sm font-semibold text-gray-900 uppercase tracking-wider"
				>
					{$t('footer.builtWith')}
				</h3>
				<ul class="space-y-2">
					<li>
						<a
							href="https://svelte.dev"
							target="_blank"
							rel="noopener noreferrer"
							class="text-sm text-gray-600 hover:text-cyan-600 transition-colors"
						>
							SvelteKit
						</a>
					</li>
					<li>
						<a
							href="https://echo.labstack.com"
							target="_blank"
							rel="noopener noreferrer"
							class="text-sm text-gray-600 hover:text-cyan-600 transition-colors"
						>
							Echo Framework
						</a>
					</li>
					<li>
						<a
							href="https://tailwindcss.com"
							target="_blank"
							rel="noopener noreferrer"
							class="text-sm text-gray-600 hover:text-cyan-600 transition-colors"
						>
							Tailwind CSS
						</a>
					</li>
					<li>
						<a
							href="https://gorm.io"
							target="_blank"
							rel="noopener noreferrer"
							class="text-sm text-gray-600 hover:text-cyan-600 transition-colors"
						>
							GORM
						</a>
					</li>
				</ul>
			</div>
		</div>

		<!-- Bottom Bar -->
		<div class="mt-8 pt-8 border-t border-gray-300">
			<div
				class="flex flex-col md:flex-row justify-between items-center space-y-4 md:space-y-0"
			>
				<div
					class="flex flex-col md:flex-row items-center gap-3 text-sm text-gray-500"
				>
					<p>{$t('footer.copyright')}</p>
					<span class="hidden md:inline">•</span>
					<p class="flex items-center gap-1.5 font-mono">
						<svg
							class="w-4 h-4"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z"
							></path>
						</svg>
						<span>v{__APP_VERSION__}</span>
					</p>
				</div>
				<div class="flex items-center space-x-2 text-sm text-gray-500">
					<span>{$t('footer.developedWith')}</span>
					<svg
						class="w-4 h-4 text-red-500"
						fill="currentColor"
						viewBox="0 0 20 20"
					>
						<path
							fill-rule="evenodd"
							d="M3.172 5.172a4 4 0 015.656 0L10 6.343l1.172-1.171a4 4 0 115.656 5.656L10 17.657l-6.828-6.829a4 4 0 010-5.656z"
							clip-rule="evenodd"
						></path>
					</svg>
					<span>{$t('footer.inSwitzerland')}</span>
				</div>
			</div>
		</div>
	</div>
</footer>

<Toast />

{#if showMobileNav}
	<MobileNav onNew={() => (showNewDialog = true)} />
{/if}

<!-- Global type-choice ("New") dialog, triggered from desktop button, iOS header +, Android FAB -->
<TypeChoiceDialog
	bind:open={showNewDialog}
	onClose={() => (showNewDialog = false)}
/>

<svelte:window
	onclick={() => {
		if (showUserMenu) showUserMenu = false;
		if (showAdminMenu) showAdminMenu = false;
	}}
/>
