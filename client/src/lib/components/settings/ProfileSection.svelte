<script lang="ts">
	import type { ProfileDTO } from '$lib/api';
	import { profileApi } from '$lib/api';
	import { authStore } from '$lib/stores/auth';
	import { t } from '$lib/stores/i18n';
	import { toastStore } from '$lib/stores/toast';
	import { get } from 'svelte/store';

	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);

	interface Props {
		profile: ProfileDTO;
		onProfileUpdated: (profile: ProfileDTO) => void;
	}

	let { profile, onProfileUpdated }: Props = $props();

	// Profile form state — sync when profile prop changes
	let firstName = $state('');
	let lastName = $state('');

	let isSavingProfile = $state(false);

	$effect(() => {
		if (!isSavingProfile) {
			firstName = profile.first_name || '';
			lastName = profile.last_name || '';
		}
	});

	async function handleSaveProfile(e: Event) {
		e.preventDefault();
		isSavingProfile = true;

		try {
			const response = await profileApi.update({
				first_name: firstName,
				last_name: lastName
			});
			onProfileUpdated(response.profile);
			toastStore.success(tr('settings.profile.success'));
			await authStore.checkAuth();
		} catch {
			toastStore.error(tr('settings.profile.error'));
		} finally {
			isSavingProfile = false;
		}
	}
</script>

<div class="space-y-6">
	<!-- Profile Form -->
	<div
		class="bg-white rounded-lg shadow-lg overflow-hidden"
		style="border-left: 6px solid #06b6d4"
	>
		<div class="p-6">
			<h3 class="text-lg font-semibold text-text mb-4">
				{tr('settings.profile.title')}
			</h3>

			{#if profile.auth_provider !== 'local'}
				<p class="text-sm text-text-subtle mb-4">
					{tr('settings.profile.oauthNote')}
				</p>
			{/if}

			<form onsubmit={handleSaveProfile} class="space-y-4">
				<div>
					<label
						for="firstName"
						class="block text-sm font-medium text-text-ink2 mb-1"
					>
						{tr('settings.profile.firstName')}
					</label>
					<input
						id="firstName"
						type="text"
						bind:value={firstName}
						disabled={isSavingProfile || profile.auth_provider !== 'local'}
						class="input"
						class:bg-surface-1={profile.auth_provider !== 'local'}
						class:text-text-subtle={profile.auth_provider !== 'local'}
						class:cursor-not-allowed={profile.auth_provider !== 'local'}
					/>
				</div>

				<div>
					<label
						for="lastName"
						class="block text-sm font-medium text-text-ink2 mb-1"
					>
						{tr('settings.profile.lastName')}
					</label>
					<input
						id="lastName"
						type="text"
						bind:value={lastName}
						disabled={isSavingProfile || profile.auth_provider !== 'local'}
						class="input"
						class:bg-surface-1={profile.auth_provider !== 'local'}
						class:text-text-subtle={profile.auth_provider !== 'local'}
						class:cursor-not-allowed={profile.auth_provider !== 'local'}
					/>
				</div>

				<div>
					<label
						for="email"
						class="block text-sm font-medium text-text-ink2 mb-1"
					>
						{tr('settings.profile.email')}
					</label>
					<input
						id="email"
						type="email"
						value={profile.email}
						disabled
						class="input bg-surface-1 text-text-subtle cursor-not-allowed"
					/>
				</div>

				{#if profile.auth_provider === 'local'}
					<div class="pt-2">
						<button
							type="submit"
							disabled={isSavingProfile}
							class="btn btn-primary"
						>
							{#if isSavingProfile}
								<span class="relative inline-flex h-3 w-3 mr-2"
									><span
										class="animate-ping absolute inline-flex h-full w-full rounded-full bg-accent-400 opacity-75"
									></span><span
										class="relative inline-flex rounded-full h-3 w-3 bg-accent"
									></span></span
								>
								{tr('settings.profile.saving')}
							{:else}
								{tr('settings.profile.saveButton')}
							{/if}
						</button>
					</div>
				{/if}
			</form>
		</div>
	</div>
</div>
