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
	import { platform } from '$lib/utils/platform';
	import { onMount } from 'svelte';
	import { get } from 'svelte/store';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import SettingsTabs from '$lib/components/settings/SettingsTabs.svelte';

	// `platform` is a module constant, so a plain const, not $derived. Desktop
	// renders this route as the notifications tab of the merged settings page.
	const IS_DESKTOP = platform === 'other';

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
	{#if IS_DESKTOP}
		<div class="mb-5">
			<div class="text-label font-normal text-text-subtle">
				{$t('settings.sections.account')}
			</div>
			<h1 class="mt-0.5 text-title text-text">{$t('settings.title')}</h1>
		</div>
		<SettingsTabs active="notifications" />
	{:else}
		<PageHeader title={tr('settings.notifications.title')} />
	{/if}

	{#if isLoadingProfile}
		<LoadingSpinner />
	{:else if profile}
		<div class="w-full {IS_DESKTOP ? 'lg:max-w-160' : 'lg:max-w-2xl'}">
			<NotificationsSection
				{profile}
				onProfileUpdated={handleProfileUpdated}
				showTitle={!IS_DESKTOP}
			/>
			{#if IS_DESKTOP}
				<p class="mt-3 pl-0.5 text-body-sm text-text-faint">
					{$t('settings.notifications.ios.subcategoryHint')}
				</p>
			{/if}
		</div>
	{/if}
</div>
