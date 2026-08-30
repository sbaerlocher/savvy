<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { profileApi, type ProfileDTO } from '$lib/api';
	import NotificationsSection from '$lib/components/settings/NotificationsSection.svelte';
	import { authStore } from '$lib/stores/auth';
	import { t } from '$lib/stores/i18n';
	import { toastStore } from '$lib/stores/toast';
	import { logger } from '$lib/utils/logger';
	import { platform } from '$lib/utils/platform';
	import { onMount } from 'svelte';
	import { get } from 'svelte/store';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

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

{#if isLoadingProfile}
	<LoadingSpinner />
{:else if profile}
	{#if IS_DESKTOP}
		<!-- Same two-column split as the profile and security tabs: cards in the
		     wide left column, a w-90 sidebar on the right. -->
		<div class="flex flex-col items-start gap-6 lg:flex-row">
			<div class="flex min-w-0 flex-1 flex-col gap-5">
				<NotificationsSection
					{profile}
					onProfileUpdated={handleProfileUpdated}
				/>
			</div>
			<div class="flex w-full flex-col gap-5 lg:w-90 lg:flex-none">
				<div class="rounded-xl border border-border bg-white p-6">
					<h3 class="mb-1.5 text-subheading font-semibold text-text">
						{$t('settings.notifications.hintTitle')}
					</h3>
					<p class="text-body-sm text-text-muted">
						{$t('settings.notifications.ios.subcategoryHint')}
					</p>
				</div>
			</div>
		</div>
	{:else}
		<div class="w-full">
			<NotificationsSection {profile} onProfileUpdated={handleProfileUpdated} />
		</div>
	{/if}
{/if}
