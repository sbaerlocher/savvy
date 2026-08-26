<script lang="ts">
	import type { ProfileDTO } from '$lib/api';
	import { authApi, exportApi, profileApi } from '$lib/api';
	import { authStore } from '$lib/stores/auth';
	import { configStore } from '$lib/stores/config';
	import { languageStore, t, type Language } from '$lib/stores/i18n';
	import { toastStore } from '$lib/stores/toast';
	import { get } from 'svelte/store';
	import DeleteAccountCard from './DeleteAccountCard.svelte';

	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);

	interface Props {
		profile: ProfileDTO;
		onProfileUpdated: (profile: ProfileDTO) => void;
	}

	let { profile, onProfileUpdated }: Props = $props();

	let firstName = $state('');
	let lastName = $state('');
	let isSavingProfile = $state(false);
	let isExporting = $state(false);
	let isSendingVerification = $state(false);

	const isLocal = $derived(profile.auth_provider === 'local');

	$effect(() => {
		if (!isSavingProfile) {
			firstName = profile.first_name || '';
			lastName = profile.last_name || '';
		}
	});

	const languageNames: Record<Language, string> = {
		de: 'Deutsch',
		en: 'English',
		fr: 'Français'
	};

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

<div class="flex flex-col items-start gap-6 lg:flex-row">
	<div class="flex min-w-0 flex-1 flex-col gap-5">
		<!-- Personal details -->
		<div class="rounded-xl border border-border bg-white p-6">
			<h3 class="mb-4 text-heading font-semibold text-text">
				{$t('settings.profile.title')}
			</h3>

			{#if !isLocal}
				<p class="mb-4 text-body text-text-subtle">
					{$t('settings.profile.oauthNote')}
				</p>
			{/if}

			<form onsubmit={handleSaveProfile}>
				<div class="mb-4 grid grid-cols-2 gap-4">
					<div>
						<label
							for="firstName"
							class="mb-1.5 block text-label font-medium text-text-ink2"
						>
							{$t('settings.profile.firstName')}
						</label>
						<input
							id="firstName"
							type="text"
							bind:value={firstName}
							disabled={isSavingProfile || !isLocal}
							class="input"
							class:bg-surface-1={!isLocal}
							class:text-text-subtle={!isLocal}
							class:cursor-not-allowed={!isLocal}
						/>
					</div>
					<div>
						<label
							for="lastName"
							class="mb-1.5 block text-label font-medium text-text-ink2"
						>
							{$t('settings.profile.lastName')}
						</label>
						<input
							id="lastName"
							type="text"
							bind:value={lastName}
							disabled={isSavingProfile || !isLocal}
							class="input"
							class:bg-surface-1={!isLocal}
							class:text-text-subtle={!isLocal}
							class:cursor-not-allowed={!isLocal}
						/>
					</div>
				</div>

				<!-- Email row carries the verification state inline, the way the
				     mockup shows it, instead of a separate account-card block. -->
				<div class="mb-5">
					<label
						for="email"
						class="mb-1.5 block text-label font-medium text-text-ink2"
					>
						{$t('settings.profile.email')}
					</label>
					<!-- The mockup shows the verification badge inside the field. The
					     email stays a real disabled input for semantics; the badge is a
					     flex sibling rather than an overlay, so it reserves its own
					     width in every language instead of relying on fixed padding. -->
					<div
						class="input flex h-11 items-center gap-3 bg-surface-1 py-0 pr-3.5"
					>
						<input
							id="email"
							type="email"
							value={profile.email}
							disabled
							class="min-w-0 flex-1 cursor-not-allowed border-0 bg-transparent p-0 text-text-muted focus:outline-none"
						/>
						<span class="flex flex-none items-center">
							{#if profile.email_verified}
								<span
									class="inline-flex shrink-0 items-center gap-1 rounded-full bg-success-100 px-2 py-0.5 text-eyebrow font-semibold text-success-800"
								>
									{$t('settings.emailVerification.verified')}
								</span>
							{:else}
								<span
									class="inline-flex shrink-0 items-center gap-1 rounded-full bg-warning-100 px-2 py-0.5 text-eyebrow font-semibold text-warning-800"
								>
									{$t('settings.emailVerification.notVerified')}
								</span>
							{/if}
						</span>
					</div>

					{#if !profile.email_verified}
						{#if $configStore.smtp_enabled}
							<button
								type="button"
								onclick={handleSendVerification}
								disabled={isSendingVerification}
								class="btn btn-ghost btn-sm mt-2"
							>
								{isSendingVerification
									? $t('settings.emailVerification.sending')
									: $t('settings.emailVerification.verifyButton')}
							</button>
						{:else}
							<span class="mt-2 block text-body-sm text-text-faint">
								{$t('settings.emailVerification.smtpDisabled')}
							</span>
						{/if}
					{/if}
				</div>

				{#if isLocal}
					<button
						type="submit"
						disabled={isSavingProfile}
						class="btn btn-primary"
					>
						{isSavingProfile
							? $t('settings.profile.saving')
							: $t('settings.profile.saveButton')}
					</button>
				{/if}
			</form>
		</div>

		<!-- Data export -->
		<div class="rounded-xl border border-border bg-white p-6">
			<h3 class="mb-1.5 text-heading font-semibold text-text">
				{$t('settings.export.title')}
			</h3>
			<p class="mb-4 max-w-lg text-body text-text-muted">
				{$t('settings.export.description')}
			</p>
			<button
				type="button"
				onclick={handleExport}
				disabled={isExporting}
				class="btn btn-ghost"
			>
				{#if isExporting}
					{$t('settings.export.downloading')}
				{:else}
					<svg
						class="mr-2 h-4 w-4"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
						aria-hidden="true"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
						/>
					</svg>
					{$t('settings.export.button')}
				{/if}
			</button>
		</div>
	</div>

	<!-- Sidebar: account facts + danger zone -->
	<div class="flex w-full flex-col gap-5 lg:w-90 lg:flex-none">
		<div class="rounded-xl border border-border bg-white p-6">
			<h3 class="mb-4 text-subheading font-semibold text-text">
				{$t('settings.account.title')}
			</h3>
			<dl class="flex flex-col gap-3">
				<div class="flex justify-between gap-3">
					<dt class="text-label font-normal text-text-subtle">
						{$t('settings.account.memberSince')}
					</dt>
					<dd class="text-label font-normal text-text">
						{formatDate(profile.created_at)}
					</dd>
				</div>
				<div class="flex justify-between gap-3">
					<dt class="text-label font-normal text-text-subtle">
						{$t('settings.account.authProvider')}
					</dt>
					<dd class="text-label font-normal text-text">
						{isLocal
							? $t('settings.account.providerLocal')
							: $t('settings.account.providerOAuth')}
					</dd>
				</div>
				<div class="flex items-center justify-between gap-3">
					<dt class="text-label font-normal text-text-subtle">
						{$t('aria.selectLanguage')}
					</dt>
					<dd>
						<label class="sr-only" for="settingsLanguage">
							{$t('aria.selectLanguage')}
						</label>
						<select
							id="settingsLanguage"
							value={$languageStore}
							onchange={(e) =>
								languageStore.setLanguage(e.currentTarget.value as Language)}
							class="cursor-pointer rounded-md border border-border bg-white px-2 py-1 text-label font-normal text-text"
						>
							{#each Object.entries(languageNames) as [code, name] (code)}
								<option value={code}>{name}</option>
							{/each}
						</select>
					</dd>
				</div>
			</dl>
		</div>

		<DeleteAccountCard {profile} />
	</div>
</div>
