<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/stores';
	import { profileApi, type ProfileDTO } from '$lib/api';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import AndroidProfileSection from '$lib/components/settings/AndroidProfileSection.svelte';
	import AndroidSecuritySection from '$lib/components/settings/AndroidSecuritySection.svelte';
	import NotificationsSection from '$lib/components/settings/NotificationsSection.svelte';
	import { authStore } from '$lib/stores/auth';
	import { t } from '$lib/stores/i18n';
	import { toastStore } from '$lib/stores/toast';
	import { logger } from '$lib/utils/logger';
	import { platform } from '$lib/utils/platform';
	import { onMount } from 'svelte';
	import { get } from 'svelte/store';

	const pageLogger = logger.child('SettingsPage');
	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);

	// Android renders the single settings screen the mockup shows (profile,
	// security and notifications as one scroll). iOS and desktop keep their
	// dedicated routes, so this stays the legacy redirect there. `platform` is
	// a module constant, so a plain const, not $derived.
	const IS_ANDROID = platform === 'android';

	let profile = $state<ProfileDTO | null>(null);
	let isLoadingProfile = $state(true);

	onMount(async () => {
		if (!IS_ANDROID) {
			// Legacy redirect: old tab URLs → the dedicated pages
			const urlTab = $page.url.searchParams.get('tab');
			if (urlTab === 'profile') {
				goto(resolve('/profile'), { replaceState: true });
				return;
			}
			if (urlTab === 'security') {
				goto(resolve('/security'), { replaceState: true });
				return;
			}
			goto(resolve('/notifications'), { replaceState: true });
			return;
		}

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
	{#if IS_ANDROID}
		<title>{tr('settings.title')} - {tr('common.appName')}</title>
	{/if}
</svelte:head>

{#if IS_ANDROID}
	<!-- M3 top app bar: 46px back circle next to the title, no page padding —
	     the list items carry their own 24px inset (mockup). -->
	<div class="-mx-4 pb-10">
		<div class="flex items-center gap-2 p-2 pb-3.5">
			<button
				type="button"
				onclick={() => history.back()}
				aria-label={tr('common.back')}
				class="flex h-11.5 w-11.5 shrink-0 items-center justify-center rounded-m3-full text-text-ink2 transition-colors active:bg-ground-active"
			>
				<svg
					class="h-5.5 w-5.5"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
					stroke-linecap="round"
					stroke-linejoin="round"
					aria-hidden="true"
				>
					<path d="M15 5l-7 7 7 7" />
				</svg>
			</button>
			<h1 class="text-heading font-semibold text-text">
				{tr('settings.title')}
			</h1>
		</div>

		{#if isLoadingProfile}
			<LoadingSpinner />
		{:else if profile}
			<AndroidProfileSection
				{profile}
				onProfileUpdated={handleProfileUpdated}
			/>

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
{/if}
