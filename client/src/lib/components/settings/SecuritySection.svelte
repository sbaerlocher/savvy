<script lang="ts">
	import { ICON_BELL, ICON_LOCK } from '$lib/icons';
	import { resolve } from '$app/paths';
	import type { ProfileDTO } from '$lib/api';
	import { authApi, exportApi } from '$lib/api';
	import { configStore } from '$lib/stores/config';
	import { languageStore, t, type Language } from '$lib/stores/i18n';
	import { toastStore } from '$lib/stores/toast';
	import { get } from 'svelte/store';
	import DeleteAccountCard from './DeleteAccountCard.svelte';

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
</script>

<div class="space-y-6">
	<!-- Account Info + Email Verification -->
	<div class="overflow-hidden rounded-xl border border-border bg-white p-6">
		<h3 class="text-lg font-semibold text-text mb-4">
			{tr('settings.account.title')}
		</h3>

		<dl class="space-y-3">
			<div class="flex justify-between">
				<dt class="text-sm text-text-subtle">
					{tr('settings.account.memberSince')}
				</dt>
				<dd class="text-sm text-text">
					{formatDate(profile.created_at)}
				</dd>
			</div>
			<div class="flex justify-between">
				<dt class="text-sm text-text-subtle">
					{tr('settings.account.authProvider')}
				</dt>
				<dd class="text-sm text-text">
					{#if profile.auth_provider === 'local'}
						{tr('settings.account.providerLocal')}
					{:else}
						{tr('settings.account.providerOAuth')}
					{/if}
				</dd>
			</div>
		</dl>

		<!-- Language -->
		<div class="mt-4 pt-4 border-t border-border-soft">
			<h4 class="text-sm font-medium text-text-ink2 mb-2">
				{tr('aria.selectLanguage')}
			</h4>
			<div class="flex gap-2">
				{#each languages as lang (lang.code)}
					<button
						type="button"
						onclick={() => changeLanguage(lang.code)}
						class="px-3 py-1.5 text-sm rounded-md transition-colors {$languageStore ===
						lang.code
							? 'bg-accent-50 text-accent font-semibold border border-accent-200'
							: 'text-text-muted hover:bg-surface-1 border border-border'}"
					>
						{lang.name}
					</button>
				{/each}
			</div>
		</div>

		<!-- Email Verification -->
		<div class="mt-4 pt-4 border-t border-border-soft">
			<h4 class="text-sm font-medium text-text-ink2 mb-2">
				{tr('settings.emailVerification.title')}
			</h4>
			<div class="flex items-center gap-2 flex-wrap">
				<span class="text-sm text-text-muted break-all">{profile.email}</span>
				{#if profile.email_verified}
					<span
						class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-success-100 text-success-800"
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
						class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-warning-100 text-warning-800"
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
									class="animate-ping absolute inline-flex h-full w-full rounded-full bg-accent-400 opacity-75"
								></span><span
									class="relative inline-flex rounded-full h-3 w-3 bg-accent"
								></span></span
							>
							{tr('settings.emailVerification.sending')}
						{:else}
							{tr('settings.emailVerification.verifyButton')}
						{/if}
					</button>
				{:else}
					<span class="text-xs text-text-faint mt-2 block"
						>{tr('settings.emailVerification.smtpDisabled')}</span
					>
				{/if}
			{/if}
		</div>
	</div>

	<!-- Link to the dedicated security page (password, 2FA, sessions).
	     Placed right after the account card on all viewports. -->
	<a
		href={resolve('/security')}
		class="group flex items-center gap-4 rounded-xl border border-border bg-white p-6 transition hover:border-border-field"
	>
		<div
			class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-border-soft text-text-subtle"
		>
			<svg
				class="h-5 w-5"
				fill="none"
				stroke="currentColor"
				viewBox="0 0 24 24"
				aria-hidden="true"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d={ICON_LOCK}
				/>
			</svg>
		</div>
		<div class="min-w-0 flex-1">
			<p class="font-semibold text-text group-hover:text-accent-hover">
				{tr('profile.securityLink.title')}
			</p>
			<p class="text-sm text-text-subtle">
				{tr('profile.securityLink.description')}
			</p>
		</div>
		<svg
			class="h-5 w-5 shrink-0 text-text-faint"
			fill="none"
			stroke="currentColor"
			viewBox="0 0 24 24"
			aria-hidden="true"
		>
			<path
				stroke-linecap="round"
				stroke-linejoin="round"
				stroke-width="2"
				d="M9 5l7 7-7 7"
			/>
		</svg>
	</a>

	<!-- Link to the notification preferences page. -->
	<a
		href={resolve('/notifications/settings')}
		class="group flex items-center gap-4 rounded-xl border border-border bg-white p-6 transition hover:border-border-field"
	>
		<div
			class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-border-soft text-text-subtle"
		>
			<svg
				class="h-5 w-5"
				fill="none"
				stroke="currentColor"
				viewBox="0 0 24 24"
				aria-hidden="true"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d={ICON_BELL}
				/>
			</svg>
		</div>
		<div class="min-w-0 flex-1">
			<p class="font-semibold text-text group-hover:text-accent-hover">
				{tr('profile.notificationsLink.title')}
			</p>
			<p class="text-sm text-text-subtle">
				{tr('profile.notificationsLink.description')}
			</p>
		</div>
		<svg
			class="h-5 w-5 shrink-0 text-text-faint"
			fill="none"
			stroke="currentColor"
			viewBox="0 0 24 24"
			aria-hidden="true"
		>
			<path
				stroke-linecap="round"
				stroke-linejoin="round"
				stroke-width="2"
				d="M9 5l7 7-7 7"
			/>
		</svg>
	</a>

	<!-- Data Export -->
	<div class="overflow-hidden rounded-xl border border-border bg-white p-6">
		<h3 class="text-lg font-semibold text-text mb-2">
			{tr('settings.export.title')}
		</h3>
		<p class="text-sm text-text-muted mb-4">
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
						class="animate-ping absolute inline-flex h-full w-full rounded-full bg-accent-400 opacity-75"
					></span><span
						class="relative inline-flex rounded-full h-3 w-3 bg-accent"
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

	<DeleteAccountCard {profile} />
</div>
