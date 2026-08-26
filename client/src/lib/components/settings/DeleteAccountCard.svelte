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
	}

	let { profile }: Props = $props();

	let showDeleteModal = $state(false);
	let deleteConfirmation = $state('');
	let deletePassword = $state('');
	let isDeleting = $state(false);

	function closeDeleteModal() {
		if (isDeleting) return;
		showDeleteModal = false;
		deleteConfirmation = '';
		deletePassword = '';
	}

	function handleModalKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') {
			closeDeleteModal();
			return;
		}

		// Focus trapping
		if (e.key === 'Tab') {
			const modal = e.currentTarget as HTMLElement;
			const focusable = modal.querySelectorAll<HTMLElement>(
				'input:not([disabled]), button:not([disabled]), [tabindex]:not([tabindex="-1"])'
			);
			if (focusable.length === 0) return;

			const first = focusable[0];
			const last = focusable[focusable.length - 1];

			if (e.shiftKey && document.activeElement === first) {
				e.preventDefault();
				last.focus();
			} else if (!e.shiftKey && document.activeElement === last) {
				e.preventDefault();
				first.focus();
			}
		}
	}

	function handleBackdropClick(e: MouseEvent) {
		if (e.target === e.currentTarget) {
			closeDeleteModal();
		}
	}

	async function handleDeleteAccount(e: Event) {
		e.preventDefault();

		if (deleteConfirmation !== 'DELETE') return;
		if (profile.auth_provider === 'local' && !deletePassword) return;

		isDeleting = true;

		try {
			await profileApi.deleteAccount({
				password: deletePassword,
				confirmation: deleteConfirmation
			});
			toastStore.success(tr('settings.dangerZone.deleteSuccess'));
			await authStore.logout();
			window.location.href = '/login';
		} catch {
			toastStore.error(tr('settings.dangerZone.deleteError'));
		} finally {
			isDeleting = false;
		}
	}
</script>

<div class="rounded-xl border-2 border-danger-200 bg-white p-6">
	<h3 class="mb-1.5 text-subheading font-semibold text-danger-600">
		{$t('settings.dangerZone.title')}
	</h3>
	<p class="mb-4 text-label font-normal text-text-muted">
		{$t('settings.dangerZone.deleteDescription')}
	</p>
	<button
		type="button"
		onclick={() => (showDeleteModal = true)}
		class="btn btn-ghost border-danger-300 text-danger-600 hover:bg-danger-50"
	>
		{$t('settings.dangerZone.deleteButton')}
	</button>
</div>

{#if showDeleteModal}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4"
		role="dialog"
		aria-modal="true"
		aria-labelledby="delete-modal-title"
		tabindex="-1"
		onclick={handleBackdropClick}
		onkeydown={handleModalKeydown}
	>
		<div class="w-full max-w-md rounded-lg bg-white p-6 shadow-xl">
			<h3
				id="delete-modal-title"
				class="mb-2 text-heading font-semibold text-danger-600"
			>
				{$t('settings.dangerZone.deleteConfirmTitle')}
			</h3>
			<p class="mb-4 text-body text-text-muted">
				{$t('settings.dangerZone.deleteConfirmMessage')}
			</p>

			<form onsubmit={handleDeleteAccount} class="space-y-4">
				<div>
					<label
						for="deleteConfirmation"
						class="mb-1 block text-body font-medium text-text-ink2"
					>
						{$t('settings.dangerZone.deleteConfirmPlaceholder')}
					</label>
					<input
						id="deleteConfirmation"
						type="text"
						bind:value={deleteConfirmation}
						disabled={isDeleting}
						autocomplete="off"
						class="input"
						placeholder={$t('settings.dangerZone.deleteConfirmWord')}
					/>
				</div>

				{#if profile.auth_provider === 'local'}
					<div>
						<label
							for="deletePassword"
							class="mb-1 block text-body font-medium text-text-ink2"
						>
							{$t('settings.dangerZone.passwordRequired')}
						</label>
						<input
							id="deletePassword"
							type="password"
							bind:value={deletePassword}
							disabled={isDeleting}
							autocomplete="current-password"
							class="input"
						/>
					</div>
				{/if}

				<div class="flex gap-3 pt-2">
					<button
						type="button"
						onclick={closeDeleteModal}
						disabled={isDeleting}
						class="btn btn-ghost flex-1"
					>
						{$t('common.cancel')}
					</button>
					<button
						type="submit"
						disabled={isDeleting ||
							deleteConfirmation !== 'DELETE' ||
							(profile.auth_provider === 'local' && !deletePassword)}
						class="flex-1 rounded-md bg-danger-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-danger-700 disabled:cursor-not-allowed disabled:opacity-50"
					>
						{isDeleting
							? $t('settings.dangerZone.deleting')
							: $t('settings.dangerZone.deleteButton')}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}
