<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { profileApi, type ProfileDTO } from '$lib/api';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import AdminHubSection from './AdminHubSection.svelte';
	import AndroidProfileSection from './AndroidProfileSection.svelte';
	import AndroidSecuritySection from './AndroidSecuritySection.svelte';
	import M3SettingsRow from './M3SettingsRow.svelte';
	import NotificationsSection from './NotificationsSection.svelte';
	import { authStore } from '$lib/stores/auth';
	import { t } from '$lib/stores/i18n';
	import { pwaStore } from '$lib/stores/pwa';
	import { toastStore } from '$lib/stores/toast';
	import { logger } from '$lib/utils/logger';
	import { onMount } from 'svelte';
	import { get } from 'svelte/store';

	/**
	 * The Android settings screen (mockup screen-SettingsAndroid): one M3
	 * scroll with profile, security and notifications, plus the admin hub for
	 * admins (mockup screen-AdminAndroid, frame "Admin-Einstieg · Profil").
	 * Mounted from /profile — Android's account destination — inside a
	 * `width="full"` PageShell, which also renders the title row.
	 */
	const pageLogger = logger.child('AndroidSettingsScreen');
	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);

	let profile = $state<ProfileDTO | null>(null);
	let isLoadingProfile = $state(true);
	// Service-worker recovery: /profile is the app's only entry point for it
	// (same as the iOS and desktop branches), so the Android screen carries a
	// row for it too.
	let swSupported = $state(false);
	let isReregistering = $state(false);

	const ICON_REFRESH =
		'M4 4v5h5M20 20v-5h-5M5.6 9a7 7 0 0111.6-2.6L20 9M4 15l2.8 2.6A7 7 0 0018.4 15';
	const ICON_LOGOUT =
		'M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1';

	const showAdminHub = $derived($authStore.user?.is_admin ?? false);

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

	// Below `sm` DesktopNav's logout does not exist, so /profile is the only
	// way to sign out — same full reload as the other platform branches.
	async function handleLogout() {
		await authStore.logout();
		window.location.href = '/login';
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

<div class="pb-10">
	{#if isLoadingProfile}
		<LoadingSpinner />
	{:else if profile}
		{#if showAdminHub}
			<!-- Admin entry, admins only; the hub links stay the existing /admin
			     sub-routes. List rows full-width like the sections below. -->
			<div class="pb-3">
				<AdminHubSection />
			</div>
		{/if}

		<AndroidProfileSection {profile} onProfileUpdated={handleProfileUpdated} />

		<div class="mx-6 my-3 h-px bg-border-soft"></div>

		<AndroidSecuritySection {profile} />

		<div class="mx-6 mt-3.5 mb-3 h-px bg-border-soft"></div>

		<NotificationsSection
			sectionHeader
			{profile}
			onProfileUpdated={handleProfileUpdated}
		/>

		{#if swSupported}
			<div class="mx-6 mt-3.5 mb-3 h-px bg-border-soft"></div>

			<M3SettingsRow
				icon={ICON_REFRESH}
				title={tr('pwa.reregisterTitle')}
				subtitle={isReregistering
					? tr('pwa.reregistering')
					: tr('pwa.reregisterDesc')}
				onclick={handleReregister}
				disabled={isReregistering}
			/>
		{/if}

		<div class="mx-6 mt-3.5 mb-3 h-px bg-border-soft"></div>

		<M3SettingsRow
			icon={ICON_LOGOUT}
			title={tr('nav.logout')}
			danger
			onclick={handleLogout}
		/>
	{/if}
</div>
