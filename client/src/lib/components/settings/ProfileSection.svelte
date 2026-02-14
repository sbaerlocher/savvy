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

	// Profile form state
	let firstName = $state(profile.first_name || '');
	let lastName = $state(profile.last_name || '');
	let isSavingProfile = $state(false);

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
			<h3 class="text-lg font-semibold text-gray-900 mb-4">
				{tr('settings.profile.title')}
			</h3>

			{#if profile.auth_provider !== 'local'}
				<p class="text-sm text-gray-500 mb-4">
					{tr('settings.profile.oauthNote')}
				</p>
			{/if}

			<form onsubmit={handleSaveProfile} class="space-y-4">
				<div>
					<label
						for="firstName"
						class="block text-sm font-medium text-gray-700 mb-1"
					>
						{tr('settings.profile.firstName')}
					</label>
					<input
						id="firstName"
						type="text"
						bind:value={firstName}
						disabled={isSavingProfile || profile.auth_provider !== 'local'}
						class="input"
						class:bg-gray-50={profile.auth_provider !== 'local'}
						class:text-gray-500={profile.auth_provider !== 'local'}
						class:cursor-not-allowed={profile.auth_provider !== 'local'}
					/>
				</div>

				<div>
					<label
						for="lastName"
						class="block text-sm font-medium text-gray-700 mb-1"
					>
						{tr('settings.profile.lastName')}
					</label>
					<input
						id="lastName"
						type="text"
						bind:value={lastName}
						disabled={isSavingProfile || profile.auth_provider !== 'local'}
						class="input"
						class:bg-gray-50={profile.auth_provider !== 'local'}
						class:text-gray-500={profile.auth_provider !== 'local'}
						class:cursor-not-allowed={profile.auth_provider !== 'local'}
					/>
				</div>

				<div>
					<label
						for="email"
						class="block text-sm font-medium text-gray-700 mb-1"
					>
						{tr('settings.profile.email')}
					</label>
					<input
						id="email"
						type="email"
						value={profile.email}
						disabled
						class="input bg-gray-50 text-gray-500 cursor-not-allowed"
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
										class="animate-ping absolute inline-flex h-full w-full rounded-full bg-cyan-400 opacity-75"
									></span><span
										class="relative inline-flex rounded-full h-3 w-3 bg-cyan-500"
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
