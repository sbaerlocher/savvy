<script lang="ts">
	import type { ProfileDTO } from '$lib/api';
	import { exportApi, profileApi } from '$lib/api';
	import { authStore } from '$lib/stores/auth';
	import { t } from '$lib/stores/i18n';
	import { toastStore } from '$lib/stores/toast';
	import { get } from 'svelte/store';
	import M3SettingsRow from './M3SettingsRow.svelte';

	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);

	interface Props {
		profile: ProfileDTO;
		onProfileUpdated: (profile: ProfileDTO) => void;
	}

	let { profile, onProfileUpdated }: Props = $props();

	// Leading glyphs are the mockup's own paths (screen-SettingsAndroid), which
	// differ from the shared $lib/icons set.
	const ICON_PENCIL_SQUARE =
		'M11 4h-5a2 2 0 00-2 2v12a2 2 0 002 2h12a2 2 0 002-2v-5';
	const ICON_PENCIL_TIP = 'M18.5 2.5a2.1 2.1 0 013 3L12 15l-4 1 1-4z';
	const ICON_ENVELOPE = 'M3 7l9 6 9-6';
	const ICON_ENVELOPE_BOX = 'M3 5h18v14H3z';
	const ICON_EXPORT_DOC =
		'M12 10v6m0 0l-3-3m3 3l3-3M4 17V5a2 2 0 012-2h6l6 6v8a2 2 0 01-2 2H6a2 2 0 01-2-2z';
	const ICON_TRASH_CAN =
		'M3 6h18M8 6V4a2 2 0 012-2h4a2 2 0 012 2v2m2 0v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6';

	let isEditingName = $state(false);
	let firstName = $state('');
	let lastName = $state('');
	let isSavingProfile = $state(false);

	let isExporting = $state(false);

	let showDeleteModal = $state(false);
	let deleteConfirmation = $state('');
	let deletePassword = $state('');
	let isDeleting = $state(false);

	const fullName = $derived(
		[profile.first_name, profile.last_name].filter(Boolean).join(' ')
	);
	const initials = $derived(
		[profile.first_name, profile.last_name]
			.filter(Boolean)
			.map((part) => part[0]?.toUpperCase() ?? '')
			.join('') || profile.email[0]?.toUpperCase()
	);
	const memberSince = $derived(formatMonthYear(profile.created_at));

	function formatMonthYear(dateStr: string): string {
		try {
			return new Date(dateStr).toLocaleDateString(undefined, {
				year: 'numeric',
				month: 'long'
			});
		} catch {
			return dateStr;
		}
	}

	// Toggles like the password row. Re-seeding on every tap would silently
	// discard what the user just typed when they tap the row again to collapse.
	function toggleNameEditor() {
		if (isEditingName) {
			isEditingName = false;
			return;
		}
		firstName = profile.first_name || '';
		lastName = profile.last_name || '';
		isEditingName = true;
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
			isEditingName = false;
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
		if (e.target === e.currentTarget) closeDeleteModal();
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

<h2 class="text-label px-6 pt-2.5 pb-2 text-accent">
	{tr('settings.profile.title')}
</h2>

<!-- Name: the mockup shows a read row with an edit affordance; tapping it
     reveals the existing first/last-name form in place. -->
<M3SettingsRow
	avatar={initials}
	title={tr('settings.profile.name')}
	subtitle={fullName || profile.email}
	onclick={profile.auth_provider === 'local' ? toggleNameEditor : undefined}
>
	{#snippet trailing()}
		{#if profile.auth_provider === 'local'}
			<svg
				class="h-5 w-5 shrink-0 text-text-subtle"
				viewBox="0 0 24 24"
				fill="none"
				stroke="currentColor"
				stroke-width="2"
				stroke-linecap="round"
				stroke-linejoin="round"
				aria-hidden="true"
			>
				<path d={ICON_PENCIL_SQUARE} />
				<path d={ICON_PENCIL_TIP} />
			</svg>
		{/if}
	{/snippet}
</M3SettingsRow>

{#if isEditingName}
	<!-- m3-filled-form restyles the shared .input/.label/.btn classes into M3
	     filled text fields and the right-aligned text + filled-pill button pair
	     (same treatment as the Android resource edit forms). -->
	<div class="m3-filled-form">
		<form onsubmit={handleSaveProfile} class="space-y-4 px-6 pt-1 pb-4">
			<div>
				<label for="firstName" class="label">
					{tr('settings.profile.firstName')}
				</label>
				<input
					id="firstName"
					type="text"
					bind:value={firstName}
					disabled={isSavingProfile}
					class="input"
				/>
			</div>
			<div>
				<label for="lastName" class="label">
					{tr('settings.profile.lastName')}
				</label>
				<input
					id="lastName"
					type="text"
					bind:value={lastName}
					disabled={isSavingProfile}
					class="input"
				/>
			</div>
			<div class="flex gap-2">
				<button
					type="button"
					onclick={() => (isEditingName = false)}
					disabled={isSavingProfile}
					class="btn btn-ghost"
				>
					{tr('common.cancel')}
				</button>
				<button
					type="submit"
					disabled={isSavingProfile}
					class="btn btn-primary"
				>
					{isSavingProfile
						? tr('settings.profile.saving')
						: tr('settings.profile.saveButton')}
				</button>
			</div>
		</form>
	</div>
{/if}

<!-- Email + verification state -->
<M3SettingsRow
	icon="{ICON_ENVELOPE_BOX} {ICON_ENVELOPE}"
	title={profile.email}
	subtitle={tr('settings.account.memberSinceValue', { date: memberSince })}
>
	{#snippet trailing()}
		{#if profile.email_verified}
			<span
				class="text-eyebrow inline-flex shrink-0 items-center gap-0.75 rounded-m3-full bg-success-100 px-2.25 py-0.75 text-success-800"
			>
				{tr('settings.emailVerification.verified')}
			</span>
		{:else}
			<span
				class="text-eyebrow inline-flex shrink-0 items-center gap-0.75 rounded-m3-full bg-warning-100 px-2.25 py-0.75 text-warning-800"
			>
				{tr('settings.emailVerification.notVerified')}
			</span>
		{/if}
	{/snippet}
</M3SettingsRow>

<M3SettingsRow
	icon={ICON_EXPORT_DOC}
	title={tr('settings.export.button')}
	subtitle={isExporting
		? tr('settings.export.downloading')
		: tr('settings.export.androidSubtitle')}
	onclick={handleExport}
	disabled={isExporting}
/>

<M3SettingsRow
	icon={ICON_TRASH_CAN}
	title={tr('settings.dangerZone.deleteAccount')}
	subtitle={tr('settings.dangerZone.androidSubtitle')}
	danger
	onclick={() => (showDeleteModal = true)}
/>

{#if showDeleteModal}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-scrim p-4"
		role="dialog"
		aria-modal="true"
		aria-labelledby="android-delete-modal-title"
		tabindex="-1"
		onclick={handleBackdropClick}
		onkeydown={handleModalKeydown}
	>
		<div class="w-full max-w-md rounded-m3-xl bg-m3-surface-container-high p-6">
			<h3
				id="android-delete-modal-title"
				class="mb-2 text-heading text-danger-600"
			>
				{tr('settings.dangerZone.deleteConfirmTitle')}
			</h3>
			<p class="mb-4 text-body text-text-muted">
				{tr('settings.dangerZone.deleteConfirmMessage')}
			</p>

			<!-- Same M3 treatment as the name editor: filled text fields, then an
			     M3-dialog button row — right-aligned text button + filled pill. -->
			<div class="m3-filled-form">
				<form onsubmit={handleDeleteAccount} class="space-y-4">
					<div>
						<label for="androidDeleteConfirmation" class="label">
							{tr('settings.dangerZone.deleteConfirmPlaceholder')}
						</label>
						<input
							id="androidDeleteConfirmation"
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
							<label for="androidDeletePassword" class="label">
								{tr('settings.dangerZone.passwordRequired')}
							</label>
							<input
								id="androidDeletePassword"
								type="password"
								bind:value={deletePassword}
								disabled={isDeleting}
								autocomplete="current-password"
								class="input"
							/>
						</div>
					{/if}

					<div class="flex items-center justify-end gap-2 pt-2">
						<button
							type="button"
							onclick={closeDeleteModal}
							disabled={isDeleting}
							class="text-label rounded-m3-full h-10 px-3 font-semibold text-text-ink2 transition-colors hover:bg-surface-1 disabled:opacity-50"
						>
							{tr('common.cancel')}
						</button>
						<button
							type="submit"
							disabled={isDeleting ||
								deleteConfirmation !== 'DELETE' ||
								(profile.auth_provider === 'local' && !deletePassword)}
							class="text-label rounded-m3-full h-10 bg-danger-600 px-6 font-semibold text-on-accent transition-colors hover:bg-danger-700 disabled:cursor-not-allowed disabled:opacity-50"
						>
							{isDeleting
								? tr('settings.dangerZone.deleting')
								: tr('settings.dangerZone.deleteButton')}
						</button>
					</div>
				</form>
			</div>
		</div>
	</div>
{/if}
