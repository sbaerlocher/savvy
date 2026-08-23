<script lang="ts">
	/* global __APP_VERSION__ -- compile-time constant injected by Vite `define` (see vite.config.ts, typed in src/app.d.ts) */
	import { page } from '$app/stores';
	import MobileNav from '$lib/components/MobileNav.svelte';
	import OfflineIndicator from '$lib/components/OfflineIndicator.svelte';
	import Toast from '$lib/components/Toast.svelte';
	import TypeChoiceDialog from '$lib/components/TypeChoiceDialog.svelte';
	import DesktopNav from '$lib/components/shell/DesktopNav.svelte';
	import AppFooter from '$lib/components/shell/AppFooter.svelte';
	import { authStore } from '$lib/stores/auth';
	import { configStore } from '$lib/stores/config';
	import { pwaStore } from '$lib/stores/pwa';
	import { registerServiceWorker } from '$lib/pwa/register-sw';
	import { showNewDialog } from '$lib/stores/newDialog';
	import { selectModeActive } from '$lib/stores/selectMode';
	import { browser } from '$app/environment';
	import { onMount, type Snippet } from 'svelte';
	import { logger } from '$lib/utils/logger';
	import { platform } from '$lib/utils/platform';
	import '../app.css';
	import type { LayoutData } from './$types';

	const layoutLogger = logger.child('Layout');

	let { data, children }: { data: LayoutData; children?: Snippet } = $props();

	// Resource detail is a sub-screen on both native platforms: the Android
	// mockup (screen-ResourceDetailAndroid) keeps only the top app bar and the
	// edit FAB, and no iOS detail frame (screen-ResourceDetailIOS) shows the
	// floating nav. The list routes themselves keep it, so this matches
	// "/<kind>/<id>" but neither "/<kind>" nor "/<kind>/new".
	const NATIVE_DETAIL_ROUTE = $derived(
		platform !== 'other' &&
			/^\/(cards|vouchers|gift-cards)\/(?!new$)[^/]+$/.test($page.url.pathname)
	);

	// iOS select mode: the floating batch bar sits in the nav's own slot
	// (mockup screen-WalletIOS, Phone 3), so the nav steps aside while it is up.
	// Android/desktop keep their nav — their batch bar stacks above it.
	const showMobileNav = $derived(
		$authStore.isAuthenticated &&
			!$page.url.pathname.startsWith('/login') &&
			!$page.url.pathname.startsWith('/register') &&
			// Native select mode rearranges the bottom chrome: Android swaps the nav
			// bar and FAB for the M3 contextual top app bar plus batch bottom bar,
			// iOS lets the floating batch bar take the nav's slot (wallet mockups).
			!(platform !== 'other' && $selectModeActive) &&
			!NATIVE_DETAIL_ROUTE
	);

	// Android auth screens are full-bleed centered cards (mockups
	// screen-AuthAndroid / screen-PasswordResetAndroid): the shell's page
	// padding would push the card off-center, so it steps aside for those
	// routes only. The footer goes with it — it is `hidden sm:block`, so on an
	// Android tablet it would stack below a full-viewport-height main and turn
	// the centered card into a scrolling page.
	const ANDROID_FULL_BLEED = $derived(
		platform === 'android' &&
			['/login', '/register', '/forgot-password', '/reset-password'].some(
				(route) => $page.url.pathname.startsWith(route)
			)
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
		await registerServiceWorker();
	});
</script>

<!-- Offline Indicator (über der Navigation) -->
<OfflineIndicator />

<!-- Desktop Navigation -->
{#if $authStore.isAuthenticated && !$page.url.pathname.startsWith('/login') && !$page.url.pathname.startsWith('/register')}
	<DesktopNav user={data.user} />
{/if}

<main
	class={ANDROID_FULL_BLEED
		? ''
		: 'max-w-7xl mx-auto pt-4 pb-6 px-4 sm:px-6 lg:px-8'}
	class:main-with-mobile-nav={showMobileNav &&
		platform !== 'ios' &&
		platform !== 'android'}
	class:main-with-mobile-nav-android={showMobileNav && platform === 'android'}
	class:main-with-mobile-nav-floating={showMobileNav && platform === 'ios'}
>
	{@render children?.()}
</main>

{#if !ANDROID_FULL_BLEED}
	<AppFooter />
{/if}

<Toast />

{#if showMobileNav}
	<MobileNav onNew={() => ($showNewDialog = true)} />
{/if}

<!-- Global type-choice ("New") dialog, triggered from desktop button, iOS header +, Android FAB -->
<TypeChoiceDialog
	bind:open={$showNewDialog}
	onClose={() => ($showNewDialog = false)}
/>
