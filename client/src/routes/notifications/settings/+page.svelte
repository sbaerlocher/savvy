<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { profileApi, type ProfileDTO } from '$lib/api';
	import NotificationsSection from '$lib/components/settings/NotificationsSection.svelte';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import { authStore } from '$lib/stores/auth';
	import { t } from '$lib/stores/i18n';
	import { toastStore } from '$lib/stores/toast';
	import { logger } from '$lib/utils/logger';
	import { onMount } from 'svelte';
	import { get } from 'svelte/store';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	const pageLogger = logger.child('NotificationSettingsPage');
	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);

	let profile = $state<ProfileDTO | null>(null);
	let isLoadingProfile = $state(true);

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

<svelte:head>
	<title>{tr('settings.notifications.title')} - {tr('common.appName')}</title>
</svelte:head>

<div class="px-4 max-w-7xl mx-auto">
	<PageHeader title={tr('settings.notifications.title')} />

	{#if isLoadingProfile}
		<LoadingSpinner />
	{:else if profile}
		<div class="w-full lg:max-w-2xl">
			<NotificationsSection {profile} onProfileUpdated={handleProfileUpdated} />
		</div>
	{/if}
</div>
