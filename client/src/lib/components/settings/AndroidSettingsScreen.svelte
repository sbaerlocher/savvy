<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { profileApi, type ProfileDTO } from '$lib/api';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import AdminHubSection from './AdminHubSection.svelte';
	import AndroidProfileSection from './AndroidProfileSection.svelte';
	import AndroidSecuritySection from './AndroidSecuritySection.svelte';
	import NotificationsSection from './NotificationsSection.svelte';
	import { authStore } from '$lib/stores/auth';
	import { t } from '$lib/stores/i18n';
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

	const showAdminHub = $derived($authStore.user?.is_admin ?? false);

	onMount(async () => {
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
	{/if}
</div>
