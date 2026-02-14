<script lang="ts">
	import { goto } from '$app/navigation';
	import { profileApi, type ProfileDTO } from '$lib/api';
	import ProfileSection from '$lib/components/settings/ProfileSection.svelte';
	import SecuritySection from '$lib/components/settings/SecuritySection.svelte';
	import { authStore } from '$lib/stores/auth';
	import { t } from '$lib/stores/i18n';
	import { toastStore } from '$lib/stores/toast';
	import { logger } from '$lib/utils/logger';
	import { onMount } from 'svelte';
	import { get } from 'svelte/store';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	const pageLogger = logger.child('ProfilePage');
	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);

	let profile = $state<ProfileDTO | null>(null);
	let isLoadingProfile = $state(true);

	onMount(async () => {
		if (!$authStore.isAuthenticated) {
			goto('/login');
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
	<title>{tr('profile.title')} - {tr('common.appName')}</title>
</svelte:head>

<div class="px-4 max-w-7xl mx-auto">
	<div class="mb-8">
		<h1 class="text-3xl font-bold text-gray-900">{tr('profile.title')}</h1>
	</div>

	{#if isLoadingProfile}
		<LoadingSpinner />
	{:else if profile}
		<div class="flex flex-col lg:flex-row gap-6 items-start">
			<div class="w-full lg:w-2/3">
				<ProfileSection {profile} onProfileUpdated={handleProfileUpdated} />
			</div>
			<div class="w-full lg:w-1/3">
				<SecuritySection {profile} />
			</div>
		</div>
	{/if}
</div>
