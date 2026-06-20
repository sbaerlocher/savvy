<script lang="ts">
	import type { ProfileDTO } from '$lib/api';
	import { authApi, exportApi, profileApi } from '$lib/api';
	import { authStore } from '$lib/stores/auth';
	import { configStore } from '$lib/stores/config';
	import { languageStore, t, type Language } from '$lib/stores/i18n';
	import { toastStore } from '$lib/stores/toast';
	import { get } from 'svelte/store';

	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);

	interface Props {
		profile: ProfileDTO;
	}

	let { profile }: Props = $props();

	// Email verification state
	let isSendingVerification = $state(false);

	// Export state
	let isExporting = $state(false);

	// Delete account state
	let showDeleteModal = $state(false);
	let deleteConfirmation = $state('');
	let deletePassword = $state('');
	let isDeleting = $state(false);

	// Language
	const languages: { code: Language; name: string }[] = [
		{ code: 'de', name: 'Deutsch' },
		{ code: 'en', name: 'English' },
		{ code: 'fr', name: 'Français' }
	];

	function changeLanguage(lang: Language) {
		languageStore.setLanguage(lang);
	}

	function formatDate(dateStr: string): string {
		try {
			return new Date(dateStr).toLocaleDateString(undefined, {
				year: 'numeric',
				month: 'long',
				day: 'numeric'
			});
		} catch {
			return dateStr;
		}
	}

	async function handleExport() {
		isExporting = true;
		try {
			const { blob, filename } = await exportApi.download();
			const url = URL.createObjectURL(blob);
			const a = document.createElement('a');
			a.href = url;
			a.download = filename;
			a.click();
			URL.revokeObjectURL(url);
			toastStore.success(tr('settings.export.success'));
		} catch {
			toastStore.error(tr('settings.export.error'));
		} finally {
			isExporting = false;
		}
	}

	async function handleSendVerification() {
		isSendingVerification = true;

		try {
			await authApi.requestVerification();
			toastStore.success(tr('settings.emailVerification.sent'));
		} catch {
			toastStore.error(tr('settings.emailVerification.sentError'));
		} finally {
			isSendingVerification = false;
		}
	}

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

<div class="space-y-6">
	<!-- Account Info + Email Verification -->
	<div
		class="bg-white rounded-lg shadow-lg p-6 overflow-hidden"
		style="border-left: 6px solid #06b6d4"
	>
		<h3 class="text-lg font-semibold text-gray-900 mb-4">
			{tr('settings.account.title')}
		</h3>

		<dl class="space-y-3">
			<div class="flex justify-between">
				<dt class="text-sm text-gray-500">
					{tr('settings.account.memberSince')}
				</dt>
				<dd class="text-sm text-gray-900">
					{formatDate(profile.created_at)}
				</dd>
			</div>
			<div class="flex justify-between">
				<dt class="text-sm text-gray-500">
					{tr('settings.account.authProvider')}
				</dt>
				<dd class="text-sm text-gray-900">
					{#if profile.auth_provider === 'local'}
						{tr('settings.account.providerLocal')}
					{:else}
						{tr('settings.account.providerOAuth')}
					{/if}
				</dd>
			</div>
		</dl>

		<!-- Language -->
		<div class="mt-4 pt-4 border-t border-gray-100">
			<h4 class="text-sm font-medium text-gray-700 mb-2">
				{tr('aria.selectLanguage')}
			</h4>
			<div class="flex gap-2">
				{#each languages as lang (lang.code)}
					<button
						type="button"
						onclick={() => changeLanguage(lang.code)}
						class="px-3 py-1.5 text-sm rounded-md transition-colors {$languageStore ===
						lang.code
							? 'bg-cyan-50 text-cyan-600 font-semibold border border-cyan-200'
							: 'text-gray-600 hover:bg-gray-50 border border-gray-200'}"
					>
						{lang.name}
					</button>
				{/each}
			</div>
		</div>

		<!-- Email Verification -->
		<div class="mt-4 pt-4 border-t border-gray-100">
			<h4 class="text-sm font-medium text-gray-700 mb-2">
				{tr('settings.emailVerification.title')}
			</h4>
			<div class="flex items-center gap-2 flex-wrap">
				<span class="text-sm text-gray-600 break-all">{profile.email}</span>
				{#if profile.email_verified}
					<span
						class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800"
					>
						<svg class="w-3 h-3 mr-1" fill="currentColor" viewBox="0 0 20 20">
							<path
								fill-rule="evenodd"
								d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
								clip-rule="evenodd"
							/>
						</svg>
						{tr('settings.emailVerification.verified')}
					</span>
				{:else}
					<span
						class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-yellow-100 text-yellow-800"
					>
						{tr('settings.emailVerification.notVerified')}
					</span>
				{/if}
			</div>

			{#if !profile.email_verified}
				{#if $configStore.smtp_enabled}
					<button
						type="button"
						onclick={handleSendVerification}
						disabled={isSendingVerification}
						class="btn btn-ghost text-sm w-full mt-3"
					>
						{#if isSendingVerification}
							<span class="relative inline-flex h-3 w-3 mr-2"
								><span
									class="animate-ping absolute inline-flex h-full w-full rounded-full bg-cyan-400 opacity-75"
								></span><span
									class="relative inline-flex rounded-full h-3 w-3 bg-cyan-500"
								></span></span
							>
							{tr('settings.emailVerification.sending')}
						{:else}
							{tr('settings.emailVerification.verifyButton')}
						{/if}
					</button>
				{:else}
					<span class="text-xs text-gray-400 mt-2 block"
						>{tr('settings.emailVerification.smtpDisabled')}</span
					>
				{/if}
			{/if}
		</div>
	</div>

	<!-- Data Export -->
	<div
		class="bg-white rounded-lg shadow-lg p-6 overflow-hidden"
		style="border-left: 6px solid #06b6d4"
	>
		<h3 class="text-lg font-semibold text-gray-900 mb-2">
			{tr('settings.export.title')}
		</h3>
		<p class="text-sm text-gray-600 mb-4">
			{tr('settings.export.description')}
		</p>
		<button
			type="button"
			onclick={handleExport}
			disabled={isExporting}
			class="btn btn-ghost w-full"
		>
			{#if isExporting}
				<span class="relative inline-flex h-3 w-3 mr-2"
					><span
						class="animate-ping absolute inline-flex h-full w-full rounded-full bg-cyan-400 opacity-75"
					></span><span
						class="relative inline-flex rounded-full h-3 w-3 bg-cyan-500"
					></span></span
				>
				{tr('settings.export.downloading')}
			{:else}
				<svg
					class="w-4 h-4 mr-2"
					fill="none"
					stroke="currentColor"
					viewBox="0 0 24 24"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
					/>
				</svg>
				{tr('settings.export.button')}
			{/if}
		</button>
	</div>

	<!-- Danger Zone -->
	<div class="bg-white rounded-lg shadow-lg p-6 border-2 border-red-200">
		<h3 class="text-lg font-semibold text-red-600 mb-4">
			{tr('settings.dangerZone.title')}
		</h3>

		<div class="space-y-3">
			<div>
				<h4 class="text-sm font-medium text-gray-900">
					{tr('settings.dangerZone.deleteAccount')}
				</h4>
				<p class="text-sm text-gray-500 mt-1">
					{tr('settings.dangerZone.deleteDescription')}
				</p>
			</div>
			<button
				type="button"
				onclick={() => (showDeleteModal = true)}
				class="btn btn-ghost text-red-600 border-red-300 hover:bg-red-50"
			>
				{tr('settings.dangerZone.deleteButton')}
			</button>
		</div>
	</div>
</div>

{#if showDeleteModal}
	<div
		class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4"
		role="dialog"
		aria-modal="true"
		aria-labelledby="delete-modal-title"
		tabindex="-1"
		onclick={handleBackdropClick}
		onkeydown={handleModalKeydown}
	>
		<div class="bg-white rounded-lg shadow-xl max-w-md w-full p-6">
			<h3
				id="delete-modal-title"
				class="text-lg font-semibold text-red-600 mb-2"
			>
				{tr('settings.dangerZone.deleteConfirmTitle')}
			</h3>
			<p class="text-sm text-gray-600 mb-4">
				{tr('settings.dangerZone.deleteConfirmMessage')}
			</p>

			<form onsubmit={handleDeleteAccount} class="space-y-4">
				<div>
					<label
						for="deleteConfirmation"
						class="block text-sm font-medium text-gray-700 mb-1"
					>
						{tr('settings.dangerZone.deleteConfirmPlaceholder')}
					</label>
					<input
						id="deleteConfirmation"
						type="text"
						bind:value={deleteConfirmation}
						disabled={isDeleting}
						autocomplete="off"
						class="input"
						placeholder={tr('settings.dangerZone.deleteConfirmWord')}
					/>
				</div>

				{#if profile.auth_provider === 'local'}
					<div>
						<label
							for="deletePassword"
							class="block text-sm font-medium text-gray-700 mb-1"
						>
							{tr('settings.dangerZone.passwordRequired')}
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
						{tr('common.cancel')}
					</button>
					<button
						type="submit"
						disabled={isDeleting ||
							deleteConfirmation !== 'DELETE' ||
							(profile.auth_provider === 'local' && !deletePassword)}
						class="flex-1 px-4 py-2 bg-red-600 text-white rounded-md hover:bg-red-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors text-sm font-medium"
					>
						{#if isDeleting}
							<span class="relative inline-flex h-3 w-3 mr-2"
								><span
									class="animate-ping absolute inline-flex h-full w-full rounded-full bg-red-400 opacity-75"
								></span><span
									class="relative inline-flex rounded-full h-3 w-3 bg-red-500"
								></span></span
							>
							{tr('settings.dangerZone.deleting')}
						{:else}
							{tr('settings.dangerZone.deleteButton')}
						{/if}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}
