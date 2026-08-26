<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { profileApi, type ProfileDTO } from '$lib/api';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import IOSSettingsScreen from '$lib/components/settings/IOSSettingsScreen.svelte';
	import ProfileSection from '$lib/components/settings/ProfileSection.svelte';
	import SecuritySection from '$lib/components/settings/SecuritySection.svelte';
	import { authStore } from '$lib/stores/auth';
	import { t } from '$lib/stores/i18n';
	import { pwaStore } from '$lib/stores/pwa';
	import { toastStore } from '$lib/stores/toast';
	import { logger } from '$lib/utils/logger';
	import { platform } from '$lib/utils/platform';
	import { onMount } from 'svelte';
	import { get } from 'svelte/store';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	const pageLogger = logger.child('ProfilePage');
	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);

	// iOS renders the whole settings screen from screen-SettingsIOS (profile,
	// security, sessions and notifications as grouped-inset sections) instead of
	// the two-column card layout. `platform` is a module constant, so this is a
	// plain const, not $derived.
	const IS_IOS = platform === 'ios';

	let profile = $state<ProfileDTO | null>(null);
	let isLoadingProfile = $state(true);
	let isReregistering = $state(false);
	let swSupported = $state(false);

	// The only other logout lives in DesktopNav's user menu, which is
	// `hidden sm:block` — so below `sm` there was no way to sign out at all.
	// Profile is the bottom nav's account destination, so the action belongs
	// here. Same full reload as the desktop one, to drop all in-memory state.
	async function handleLogout() {
		await authStore.logout();
		window.location.href = '/login';
	}

	onMount(async () => {
		swSupported = 'serviceWorker' in navigator;

		if (!$authStore.isAuthenticated) {
			goto(resolve('/login'));
			return;
		}

		try {
			const response = await profileApi.get();
			profile = response.profile;
		} catch (error) {
			pageLogger.error('Failed to load profile', { error });
			toastStore.error(tr('common.error'));
		} finally {
			isLoadingProfile = false;
		}
	});

	function handleProfileUpdated(updatedProfile: ProfileDTO) {
		profile = updatedProfile;
		authStore.checkAuth();
	}

	async function handleReregister() {
		isReregistering = true;

		try {
			const registered = await pwaStore.reregisterServiceWorker();
			if (registered) {
				toastStore.success(tr('pwa.reregisterSuccess'));
			} else {
				toastStore.error(tr('pwa.reregisterError'));
			}
		} catch (error) {
			pageLogger.error('Service Worker re-registration failed', { error });
			toastStore.error(tr('pwa.reregisterError'));
		} finally {
			isReregistering = false;
		}
	}
</script>

<svelte:head>
	<title>{tr('profile.title')} - {tr('common.appName')}</title>
</svelte:head>

{#if IS_IOS}
	<!-- Horizontal padding comes from the layout's own px-4, same as the other
	     iOS screens; a second inset here would double it. -->
	<div>
		{#if isLoadingProfile}
			<LoadingSpinner />
		{:else if profile}
			<IOSSettingsScreen {profile} onProfileUpdated={handleProfileUpdated} />
		{/if}
	</div>
{:else}
	<div class="px-4 max-w-7xl mx-auto">
		<PageHeader title={tr('profile.title')} />

		{#if isLoadingProfile}
			<LoadingSpinner />
		{:else if profile}
			<div class="flex flex-col lg:flex-row gap-6 items-start">
				<div class="w-full lg:w-2/3">
					<ProfileSection {profile} onProfileUpdated={handleProfileUpdated} />
				</div>
				<div class="w-full lg:w-1/3 space-y-6">
					<SecuritySection {profile} />

					{#if swSupported}
						<div
							class="overflow-hidden rounded-xl border border-border bg-white p-6"
						>
							<h3 class="text-lg font-semibold text-text mb-2">
								{tr('pwa.reregisterTitle')}
							</h3>
							<p class="text-sm text-text-muted mb-4">
								{tr('pwa.reregisterDesc')}
							</p>
							<button
								type="button"
								onclick={handleReregister}
								disabled={isReregistering}
								class="btn btn-ghost w-full"
							>
								{isReregistering
									? tr('pwa.reregistering')
									: tr('pwa.reregisterButton')}
							</button>
						</div>
					{/if}

					<!-- Sign out. Only below `sm`: wider viewports already carry this in
					     DesktopNav's user menu, and two logout buttons on one screen would
					     be redundant. -->
					<div class="sm:hidden">
						<button
							type="button"
							onclick={handleLogout}
							class="flex w-full items-center justify-center gap-2 rounded-xl border border-border bg-white px-4 py-3 text-sm font-medium text-danger-600 transition-colors hover:bg-surface-1"
						>
							<svg
								class="h-5 w-5"
								fill="none"
								stroke="currentColor"
								viewBox="0 0 24 24"
								aria-hidden="true"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"
								/>
							</svg>
							{tr('nav.logout')}
						</button>
					</div>
				</div>
			</div>
		{/if}
	</div>
{/if}
