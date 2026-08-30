<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import {
		authApi,
		exportApi,
		profileApi,
		sessionsApi,
		type ProfileDTO,
		type SessionDTO
	} from '$lib/api';
	import { ICON_CHEVRON_LEFT, ICON_SHIELD } from '$lib/icons';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import TwoFactorSettings from '$lib/components/TwoFactorSettings.svelte';
	import SectionLabel from '$lib/components/ui/SectionLabel.svelte';
	import AdminHubSection from './AdminHubSection.svelte';
	import NotificationsSection from './NotificationsSection.svelte';
	import { authStore } from '$lib/stores/auth';
	import { configStore } from '$lib/stores/config';
	import { languageStore, t, type Language } from '$lib/stores/i18n';
	import { toastStore } from '$lib/stores/toast';
	import { logger } from '$lib/utils/logger';
	import { pwaStore } from '$lib/stores/pwa';
	import { onMount } from 'svelte';
	import { get } from 'svelte/store';

	const pageLogger = logger.child('IOSSettingsScreen');
	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);

	interface Props {
		profile: ProfileDTO;
		onProfileUpdated: (profile: ProfileDTO) => void;
	}

	let { profile, onProfileUpdated }: Props = $props();

	// The mockup's Name / password / 2FA rows carry a chevron pointing at a
	// follow-up UI. The live client edits all three inline, so the chevron row
	// expands the existing form in place instead of routing away.
	let expanded = $state<'name' | 'password' | 'twoFactor' | null>(null);

	function toggleSection(section: 'name' | 'password' | 'twoFactor') {
		expanded = expanded === section ? null : section;
	}

	const isLocalAuth = $derived(profile.auth_provider === 'local');
	// Admin entry points live here on iOS (mockup screen-AdminIOS): the bottom
	// nav has no admin tab and DesktopNav's admin link is `hidden sm:block`,
	// so without this section a phone has no way into admin at all. Hidden while
	// impersonating, like the desktop link.
	const showAdminHub = $derived(
		$authStore.user?.is_admin === true && !$authStore.user?.is_impersonating
	);
	const displayName = $derived(
		[profile.first_name, profile.last_name].filter(Boolean).join(' ').trim()
	);

	// ---- Profile name -------------------------------------------------------
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
			expanded = null;
		} catch {
			toastStore.error(tr('settings.profile.error'));
		} finally {
			isSavingProfile = false;
		}
	}

	// ---- Password -----------------------------------------------------------
	let currentPassword = $state('');
	let newPassword = $state('');
	let confirmPassword = $state('');
	let isChangingPassword = $state(false);

	// Mirrors the backend rule: min 12 chars, 3 of 4 character classes.
	const passwordStrength = $derived.by(() => {
		if (!newPassword) return { score: 0, label: '' };

		const complexity = [
			/[A-Z]/.test(newPassword),
			/[a-z]/.test(newPassword),
			/[0-9]/.test(newPassword),
			/[^a-zA-Z0-9]/.test(newPassword)
		].filter(Boolean).length;

		if (newPassword.length < 12)
			return { score: 1, label: tr('settings.password.tooShort') };
		if (complexity < 3)
			return { score: 2, label: tr('settings.password.tooWeak') };
		return { score: complexity >= 4 ? 4 : 3, label: '' };
	});

	async function handleChangePassword(e: Event) {
		e.preventDefault();

		if (passwordStrength.score < 3) {
			toastStore.error(passwordStrength.label);
			return;
		}
		if (newPassword !== confirmPassword) {
			toastStore.error(tr('settings.password.mismatch'));
			return;
		}

		isChangingPassword = true;
		try {
			await profileApi.changePassword({
				current_password: currentPassword,
				new_password: newPassword
			});
			toastStore.success(tr('settings.password.success'));
			currentPassword = '';
			newPassword = '';
			confirmPassword = '';
			expanded = null;
		} catch {
			toastStore.error(tr('settings.password.error'));
		} finally {
			isChangingPassword = false;
		}
	}

	// ---- Sessions -----------------------------------------------------------
	let sessions = $state<SessionDTO[]>([]);
	let isLoadingSessions = $state(true);
	let isRevokingOthers = $state(false);
	let revokingSessionId = $state<string | null>(null);

	// Swipe-to-reveal on a session row, same 96px (w-24) action as the mockup.
	let swipedSessionId = $state<string | null>(null);
	let touchStartX = 0;
	let touchStartY = 0;
	let touchAxis: 'none' | 'horizontal' | 'vertical' = 'none';

	function handleTouchStart(e: TouchEvent) {
		touchStartX = e.touches[0].clientX;
		touchStartY = e.touches[0].clientY;
		touchAxis = 'none';
	}

	function handleTouchMove(e: TouchEvent, sessionId: string) {
		const dx = e.touches[0].clientX - touchStartX;
		const dy = e.touches[0].clientY - touchStartY;

		// Lock to one axis on first movement so a vertical scroll never opens
		// the action, and an open action never fights the scroll.
		if (touchAxis === 'none') {
			if (Math.abs(dx) < 6 && Math.abs(dy) < 6) return;
			touchAxis = Math.abs(dx) > Math.abs(dy) ? 'horizontal' : 'vertical';
		}
		if (touchAxis !== 'horizontal') return;

		if (dx < -24) {
			swipedSessionId = sessionId;
		} else if (dx > 24) {
			swipedSessionId = null;
		}
	}

	// A tap (no horizontal travel) on an open row closes it again, so the
	// destructive action does not stay exposed while the user scrolls on.
	function handleTouchEnd() {
		if (touchAxis === 'none') swipedSessionId = null;
	}

	onMount(async () => {
		swSupported = 'serviceWorker' in navigator;

		try {
			const response = await sessionsApi.list();
			sessions = response.sessions || [];
		} catch (error) {
			pageLogger.error('Failed to load sessions', { error });
		} finally {
			isLoadingSessions = false;
		}
	});

	async function handleRevokeSession(sessionId: string) {
		revokingSessionId = sessionId;
		try {
			await sessionsApi.revoke(sessionId);
			sessions = sessions.filter((s) => s.id !== sessionId);
			toastStore.success(tr('settings.sessions.revokeSuccess'));
		} catch {
			toastStore.error(tr('settings.sessions.revokeError'));
		} finally {
			revokingSessionId = null;
			swipedSessionId = null;
		}
	}

	async function handleRevokeOthers() {
		isRevokingOthers = true;
		try {
			await sessionsApi.revokeOthers();
			sessions = sessions.filter((s) => s.is_current);
			toastStore.success(tr('settings.sessions.revokeOthersSuccess'));
		} catch {
			toastStore.error(tr('settings.sessions.revokeError'));
		} finally {
			isRevokingOthers = false;
		}
	}

	function formatRelativeTime(dateStr: string): string {
		try {
			const diffMs = Date.now() - new Date(dateStr).getTime();
			const diffMinutes = Math.floor(diffMs / 60000);
			const diffHours = Math.floor(diffMs / 3600000);
			const diffDays = Math.floor(diffMs / 86400000);

			if (diffMinutes < 1) return tr('notifications.timeAgo.justNow');
			if (diffMinutes < 60)
				return tr('notifications.timeAgo.minutesAgo', { count: diffMinutes });
			if (diffHours < 24)
				return tr('notifications.timeAgo.hoursAgo', { count: diffHours });
			return tr('notifications.timeAgo.daysAgo', { count: diffDays });
		} catch {
			return dateStr;
		}
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

	// A session on a phone gets the phone glyph, everything else the desktop
	// one — matching the mockup's two row icons.
	function isMobileSession(session: SessionDTO): boolean {
		return /mobile|phone|android|iphone|ipad|tablet/i.test(
			session.device_info || ''
		);
	}

	// ---- Export -------------------------------------------------------------
	let isExporting = $state(false);

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

	// ---- Account deletion ---------------------------------------------------
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
		if (isLocalAuth && !deletePassword) return;

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

	// ---- Email verification -------------------------------------------------
	let isSendingVerification = $state(false);

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

	// ---- Service worker -----------------------------------------------------
	// iOS is where a wedged service worker bites hardest, so the recovery action
	// the other platforms have must exist here too.
	let swSupported = $state(false);
	let isReregistering = $state(false);

	async function handleReregister() {
		isReregistering = true;
		try {
			const registered = await pwaStore.reregisterServiceWorker();
			if (registered) {
				toastStore.success(tr('pwa.reregisterSuccess'));
			} else {
				toastStore.error(tr('pwa.reregisterError'));
			}
		} catch (error) {
			pageLogger.error('Service Worker re-registration failed', { error });
			toastStore.error(tr('pwa.reregisterError'));
		} finally {
			isReregistering = false;
		}
	}

	// ---- Language + sign out ------------------------------------------------
	const languages: { code: Language; name: string }[] = [
		{ code: 'de', name: 'Deutsch' },
		{ code: 'en', name: 'English' },
		{ code: 'fr', name: 'Français' }
	];

	async function handleLogout() {
		await authStore.logout();
		window.location.href = '/login';
	}

	function goBack() {
		if (history.length > 1) {
			history.back();
			return;
		}
		goto(resolve('/'));
	}
</script>

<!-- iOS settings screen (screen-SettingsIOS): one screen, grouped-inset
     sections for profile, security, sessions and notifications. -->
<div class="pb-6">
	<div class="flex items-start gap-2.75 pb-3 pt-2">
		<button
			type="button"
			onclick={goBack}
			aria-label={tr('common.back')}
			class="liquid-glass-surface inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-full text-accent-700"
		>
			<svg
				class="h-5 w-5"
				viewBox="0 0 24 24"
				fill="none"
				stroke="currentColor"
				stroke-width="2.3"
				stroke-linecap="round"
				stroke-linejoin="round"
			>
				<path d={ICON_CHEVRON_LEFT} />
			</svg>
		</button>
		<div class="min-w-0 flex-1">
			<p class="text-eyebrow uppercase text-text-subtle">
				{displayName || profile.email}{showAdminHub
					? ` · ${tr('admin.users.roleAdmin')}`
					: ''}
			</p>
			<h1 class="mt-0.5 text-screen-title text-text">
				{tr('settings.title')}
			</h1>
		</div>
		{#if showAdminHub}
			<!-- The shield marks the elevated session (mockup screen-AdminIOS). -->
			<span
				class="liquid-glass-surface mt-0.5 inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-full text-purple-600"
				title={tr('nav.admin')}
			>
				<svg
					class="h-5.25 w-5.25"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="2.1"
					stroke-linecap="round"
					stroke-linejoin="round"
					aria-hidden="true"
				>
					<path d={ICON_SHIELD} />
				</svg>
			</span>
		{/if}
	</div>

	<!-- ============ PROFILE ============ -->
	<SectionLabel inset>{tr('settings.profile.title')}</SectionLabel>
	<div class="mb-2 overflow-hidden rounded-inset bg-surface">
		<button
			type="button"
			onclick={() => toggleSection('name')}
			disabled={!isLocalAuth}
			aria-expanded={expanded === 'name'}
			class="flex w-full items-center justify-between gap-3 border-b border-border-soft px-4 py-3.5 text-left disabled:cursor-not-allowed"
		>
			<span class="min-w-0">
				<span
					class="block text-[length:var(--text-code)] font-normal text-text"
				>
					{tr('settings.profile.name')}
				</span>
				<!-- The row is inert for OAuth accounts; say why instead of just
				     swallowing the tap. -->
				{#if !isLocalAuth}
					<span class="mt-0.25 block text-body-sm text-text-subtle">
						{tr('settings.profile.oauthNote')}
					</span>
				{/if}
			</span>
			<span class="flex min-w-0 items-center gap-2">
				<span
					class="truncate text-[length:var(--text-code)] font-normal text-text-muted"
					>{displayName}</span
				>
				{#if isLocalAuth}
					<svg
						class="h-3.5 w-2 shrink-0 text-text-mono-faint"
						viewBox="0 0 8 14"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
						stroke-linecap="round"
						stroke-linejoin="round"
						aria-hidden="true"
					>
						<path d="M1 1l5.5 6L1 13" />
					</svg>
				{/if}
			</span>
		</button>

		{#if expanded === 'name'}
			<form
				onsubmit={handleSaveProfile}
				class="ios-glass-form space-y-3 border-b border-border-soft bg-surface-2 px-4 py-3.5"
			>
				<div>
					<label
						for="ios-firstName"
						class="mb-1 block text-body-sm text-text-ink2"
					>
						{tr('settings.profile.firstName')}
					</label>
					<input
						id="ios-firstName"
						type="text"
						bind:value={firstName}
						disabled={isSavingProfile}
						class="input"
					/>
				</div>
				<div>
					<label
						for="ios-lastName"
						class="mb-1 block text-body-sm text-text-ink2"
					>
						{tr('settings.profile.lastName')}
					</label>
					<input
						id="ios-lastName"
						type="text"
						bind:value={lastName}
						disabled={isSavingProfile}
						class="input"
					/>
				</div>
				<button
					type="submit"
					disabled={isSavingProfile}
					class="btn btn-primary w-full"
				>
					{isSavingProfile
						? tr('settings.profile.saving')
						: tr('settings.profile.saveButton')}
				</button>
			</form>
		{/if}

		<div class="flex items-center justify-between gap-3 px-4 py-3.5">
			<span
				class="shrink-0 text-[length:var(--text-code)] font-normal text-text"
			>
				{tr('settings.profile.email')}
			</span>
			<span class="flex min-w-0 items-center gap-2">
				<span
					class="truncate text-[length:var(--text-code)] font-normal text-text-muted"
					>{profile.email}</span
				>
				{#if profile.email_verified}
					<span
						class="inline-flex shrink-0 items-center gap-0.75 rounded-full bg-success-100 px-1.75 py-0.5 text-tag font-semibold text-success-800"
					>
						<svg
							class="h-2.5 w-2.5"
							viewBox="0 0 20 20"
							fill="currentColor"
							aria-hidden="true"
						>
							<path
								fill-rule="evenodd"
								d="M16.7 5.3a1 1 0 010 1.4l-8 8a1 1 0 01-1.4 0l-4-4a1 1 0 011.4-1.4L8 12.6l7.3-7.3a1 1 0 011.4 0z"
								clip-rule="evenodd"
							/>
						</svg>
						{tr('settings.emailVerification.verified')}
					</span>
				{:else}
					<span
						class="inline-flex shrink-0 items-center rounded-full bg-warning-100 px-1.75 py-0.5 text-tag font-semibold text-warning-800"
					>
						{tr('settings.emailVerification.notVerified')}
					</span>
				{/if}
			</span>
		</div>

		<!-- Without this the only way to verify is a link inside the mail the
		     user cannot request, so an unverified address would stay unverified
		     forever and never receive reminder or sharing mail. -->
		{#if !profile.email_verified && $configStore.smtp_enabled}
			<button
				type="button"
				onclick={handleSendVerification}
				disabled={isSendingVerification}
				class="flex w-full items-center justify-between gap-3 border-t border-border-soft px-4 py-3.5 text-left"
			>
				<span class="text-[length:var(--text-code)] font-normal text-accent">
					{isSendingVerification
						? tr('settings.emailVerification.sending')
						: tr('settings.emailVerification.verifyButton')}
				</span>
			</button>
		{/if}
	</div>
	<p class="px-1.5 pb-3.5 text-body-sm text-text-faint">
		{tr('settings.account.memberSince')}
		{formatDate(profile.created_at)} ·
		{isLocalAuth
			? tr('settings.account.providerLocal')
			: tr('settings.account.providerOAuth')}
	</p>

	<div class="mb-2 overflow-hidden rounded-inset bg-surface">
		<button
			type="button"
			onclick={handleExport}
			disabled={isExporting}
			class="relative flex w-full items-center gap-3 px-4 py-3.5 text-left"
		>
			<span
				class="absolute bottom-0 right-0 left-14.25 h-px bg-border-soft"
				aria-hidden="true"
			></span>
			<span
				class="flex h-7.25 w-7.25 shrink-0 items-center justify-center rounded-m3-sm bg-accent text-on-accent"
			>
				<svg
					class="h-4 w-4"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
					stroke-linecap="round"
					stroke-linejoin="round"
					aria-hidden="true"
				>
					<path
						d="M12 10v6m0 0l-3-3m3 3l3-3M4 17V5a2 2 0 012-2h6l6 6v8a2 2 0 01-2 2H6a2 2 0 01-2-2z"
					/>
				</svg>
			</span>
			<span class="flex-1 text-[length:var(--text-code)] font-normal text-text">
				{isExporting
					? tr('settings.export.downloading')
					: tr('settings.export.button')}
			</span>
			<svg
				class="h-3.5 w-2 shrink-0 text-text-mono-faint"
				viewBox="0 0 8 14"
				fill="none"
				stroke="currentColor"
				stroke-width="2"
				stroke-linecap="round"
				stroke-linejoin="round"
				aria-hidden="true"
			>
				<path d="M1 1l5.5 6L1 13" />
			</svg>
		</button>

		<button
			type="button"
			onclick={() => (showDeleteModal = true)}
			class="flex w-full items-center gap-3 px-4 py-3.5 text-left"
		>
			<span
				class="flex h-7.25 w-7.25 shrink-0 items-center justify-center rounded-m3-sm bg-danger-600 text-on-accent"
			>
				<svg
					class="h-4 w-4"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
					stroke-linecap="round"
					stroke-linejoin="round"
					aria-hidden="true"
				>
					<path
						d="M3 6h18M8 6V4a2 2 0 012-2h4a2 2 0 012 2v2m2 0v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6"
					/>
				</svg>
			</span>
			<span
				class="flex-1 text-[length:var(--text-code)] font-normal text-danger-600"
			>
				{tr('settings.dangerZone.deleteAccount')}
			</span>
		</button>
	</div>
	<p class="px-1.5 pb-6 text-body-sm text-text-faint">
		{tr('settings.dangerZone.deleteDescription')}
	</p>

	{#if showAdminHub}
		<!-- ============ ADMINISTRATION ============ -->
		<div class="pb-6">
			<AdminHubSection />
		</div>
	{/if}

	<!-- ============ SECURITY ============ -->
	<SectionLabel inset>{tr('nav.security')}</SectionLabel>
	<div class="mb-2 overflow-hidden rounded-inset bg-surface">
		{#if isLocalAuth}
			<button
				type="button"
				onclick={() => toggleSection('password')}
				aria-expanded={expanded === 'password'}
				class="relative flex w-full items-center gap-3 px-4 py-3.5 text-left"
			>
				<span
					class="absolute bottom-0 right-0 left-14.25 h-px bg-border-soft"
					aria-hidden="true"
				></span>
				<span
					class="flex h-7.25 w-7.25 shrink-0 items-center justify-center rounded-m3-sm bg-accent text-on-accent"
				>
					<svg
						class="h-3.75 w-3.75"
						viewBox="0 0 24 24"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
						stroke-linecap="round"
						stroke-linejoin="round"
						aria-hidden="true"
					>
						<rect x="4" y="10" width="16" height="11" rx="2.5" />
						<path d="M8 10V7a4 4 0 018 0v3" />
					</svg>
				</span>
				<span
					class="flex-1 text-[length:var(--text-code)] font-normal text-text"
				>
					{tr('settings.password.changeButton')}
				</span>
				<svg
					class="h-3.5 w-2 shrink-0 text-text-mono-faint"
					viewBox="0 0 8 14"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
					stroke-linecap="round"
					stroke-linejoin="round"
					aria-hidden="true"
				>
					<path d="M1 1l5.5 6L1 13" />
				</svg>
			</button>

			{#if expanded === 'password'}
				<form
					onsubmit={handleChangePassword}
					class="ios-glass-form space-y-3 border-b border-border-soft bg-surface-2 px-4 py-3.5"
				>
					<div>
						<label
							for="ios-currentPassword"
							class="mb-1 block text-body-sm text-text-ink2"
						>
							{tr('settings.password.currentPassword')}
						</label>
						<input
							id="ios-currentPassword"
							type="password"
							autocomplete="current-password"
							required
							bind:value={currentPassword}
							disabled={isChangingPassword}
							class="input"
						/>
					</div>
					<div>
						<label
							for="ios-newPassword"
							class="mb-1 block text-body-sm text-text-ink2"
						>
							{tr('settings.password.newPassword')}
						</label>
						<input
							id="ios-newPassword"
							type="password"
							autocomplete="new-password"
							required
							bind:value={newPassword}
							disabled={isChangingPassword}
							class="input"
						/>
						{#if newPassword}
							<div class="mt-2 flex gap-1">
								{#each [1, 2, 3, 4] as i (i)}
									<div
										class="h-1 flex-1 rounded-full transition-colors {passwordStrength.score >=
										i
											? passwordStrength.score <= 2
												? 'bg-danger-400'
												: passwordStrength.score === 3
													? 'bg-warning-400'
													: 'bg-success-400'
											: 'bg-border'}"
									></div>
								{/each}
							</div>
							{#if passwordStrength.label}
								<p class="mt-1 text-body-sm text-danger-500">
									{passwordStrength.label}
								</p>
							{/if}
						{/if}
					</div>
					<div>
						<label
							for="ios-confirmPassword"
							class="mb-1 block text-body-sm text-text-ink2"
						>
							{tr('settings.password.confirmPassword')}
						</label>
						<input
							id="ios-confirmPassword"
							type="password"
							autocomplete="new-password"
							required
							bind:value={confirmPassword}
							disabled={isChangingPassword}
							class="input"
						/>
					</div>
					<button
						type="submit"
						disabled={isChangingPassword}
						class="btn btn-primary w-full"
					>
						{isChangingPassword
							? tr('settings.password.changing')
							: tr('settings.password.changeButton')}
					</button>
				</form>
			{/if}
		{/if}

		<!-- 2FA. The mockup shows a switch, but enabling needs the QR/verify
		     flow, so the switch opens TwoFactorSettings, which owns the real
		     enabled state and every enable/disable step. -->
		<button
			type="button"
			onclick={() => toggleSection('twoFactor')}
			aria-expanded={expanded === 'twoFactor'}
			class="flex w-full items-center gap-3 px-4 py-3.25 text-left"
		>
			<span
				class="flex h-7.25 w-7.25 shrink-0 items-center justify-center rounded-m3-sm bg-accent text-on-accent"
			>
				<svg
					class="h-4 w-4"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
					stroke-linecap="round"
					stroke-linejoin="round"
					aria-hidden="true"
				>
					<path d="M12 3l7 3v6c0 4-3 7-7 9-4-2-7-5-7-9V6z" />
					<path d="M9.5 12l1.8 1.8L15 10" />
				</svg>
			</span>
			<span class="flex-1 text-[length:var(--text-code)] font-normal text-text">
				{tr('settings.twoFactor.title')}
			</span>
			<svg
				class="h-3.5 w-2 shrink-0 text-text-mono-faint"
				viewBox="0 0 8 14"
				fill="none"
				stroke="currentColor"
				stroke-width="2"
				stroke-linecap="round"
				stroke-linejoin="round"
				aria-hidden="true"
			>
				<path d="M1 1l5.5 6L1 13" />
			</svg>
		</button>

		{#if expanded === 'twoFactor'}
			<div class="border-t border-border-soft bg-surface-2 px-4 py-3.5">
				<TwoFactorSettings
					authProvider={profile.auth_provider || 'local'}
					enrollmentEnabled={$configStore.two_factor_enabled}
				/>
			</div>
		{/if}
	</div>
	<p class="px-1.5 pb-3.5 text-body-sm text-text-faint">
		{tr('settings.twoFactor.description')}
	</p>

	<!-- ============ ACTIVE SESSIONS ============ -->
	<div class="flex items-center justify-between px-1.5 pb-2">
		<span class="text-body-sm font-semibold uppercase text-text-subtle">
			{tr('settings.sessions.title')}
		</span>
		{#if sessions.length > 1}
			<button
				type="button"
				onclick={handleRevokeOthers}
				disabled={isRevokingOthers}
				class="text-label font-semibold text-accent disabled:opacity-60"
			>
				{tr('settings.sessions.revokeOthers')}
			</button>
		{/if}
	</div>
	<div class="mb-6 overflow-hidden rounded-inset bg-surface">
		{#if isLoadingSessions}
			<div class="px-4 py-3.5">
				<LoadingSpinner />
			</div>
		{:else if sessions.length === 0}
			<p
				class="px-4 py-3.5 text-[length:var(--text-code)] font-normal text-text-subtle"
			>
				{tr('settings.sessions.noOtherSessions')}
			</p>
		{:else}
			{#each sessions as session, i (session.id)}
				<div class="relative overflow-hidden">
					{#if !session.is_current}
						<!-- Concealed behind the row until swiped, so it leaves the tab
						     order while hidden; keyboard and screen-reader users revoke
						     via "Alle anderen abmelden" above. -->
						<button
							type="button"
							onclick={() => handleRevokeSession(session.id)}
							disabled={revokingSessionId === session.id}
							tabindex={swipedSessionId === session.id ? 0 : -1}
							aria-hidden={swipedSessionId !== session.id}
							class="absolute inset-y-0 right-0 flex w-24 items-center justify-center bg-danger-600 text-label font-semibold text-on-accent"
						>
							{tr('settings.sessions.revoke')}
						</button>
					{/if}
					<div
						class="relative flex items-start gap-3 bg-surface px-4 py-3.25 transition-transform duration-200 {swipedSessionId ===
						session.id
							? '-translate-x-24'
							: 'translate-x-0'}"
						ontouchstart={handleTouchStart}
						ontouchmove={(e) =>
							session.is_current ? undefined : handleTouchMove(e, session.id)}
						ontouchend={handleTouchEnd}
						role="group"
						aria-label={session.device_info ||
							tr('settings.sessions.unknownDevice')}
					>
						{#if i < sessions.length - 1}
							<span
								class="absolute bottom-0 right-0 left-10.25 h-px bg-border-soft"
								aria-hidden="true"
							></span>
						{/if}
						<span
							class="mt-0.25 flex h-7.25 w-7.25 shrink-0 items-center justify-center rounded-m3-sm text-on-accent {session.is_current
								? 'bg-accent'
								: 'bg-text-subtle'}"
						>
							{#if isMobileSession(session)}
								<svg
									class="h-3.75 w-3.75"
									viewBox="0 0 24 24"
									fill="none"
									stroke="currentColor"
									stroke-width="2"
									stroke-linecap="round"
									stroke-linejoin="round"
									aria-hidden="true"
								>
									<rect x="6" y="2" width="12" height="20" rx="2.5" />
									<path d="M11 18h2" />
								</svg>
							{:else}
								<svg
									class="h-3.75 w-3.75"
									viewBox="0 0 24 24"
									fill="none"
									stroke="currentColor"
									stroke-width="2"
									stroke-linecap="round"
									stroke-linejoin="round"
									aria-hidden="true"
								>
									<rect x="3" y="4" width="18" height="12" rx="2" />
									<path d="M8 20h8M12 16v4" />
								</svg>
							{/if}
						</span>
						<div class="min-w-0 flex-1">
							<div class="flex items-center gap-1.75">
								<span class="truncate text-subheading text-text">
									{session.device_info || tr('settings.sessions.unknownDevice')} ·
									{session.browser_info ||
										tr('settings.sessions.unknownBrowser')}
								</span>
								{#if session.is_current}
									<span
										class="shrink-0 rounded-full bg-success-100 px-1.75 py-0.25 text-tag font-semibold text-success-800"
									>
										{tr('settings.sessions.current')}
									</span>
								{/if}
							</div>
							<div class="mt-0.5 font-mono text-body-sm text-text-subtle">
								{#if session.ip_address}{session.ip_address} ·
								{/if}{formatRelativeTime(session.last_active_at)}
							</div>
						</div>
					</div>
				</div>
			{/each}
		{/if}
	</div>

	<!-- ============ NOTIFICATIONS ============ -->
	<NotificationsSection {profile} {onProfileUpdated} />

	<!-- Language and sign-out are not in the mockup, but removing them would
	     drop the only way to switch language or sign out on a phone. Kept in
	     the same grouped-inset chrome, below the mockup's own sections. -->
	<SectionLabel inset spaced>{tr('aria.selectLanguage')}</SectionLabel>
	<div class="mb-2 overflow-hidden rounded-inset bg-surface">
		{#each languages as lang, i (lang.code)}
			<button
				type="button"
				onclick={() => languageStore.setLanguage(lang.code)}
				class="relative flex w-full items-center justify-between gap-3 px-4 py-3.25 text-left"
			>
				{#if i < languages.length - 1}
					<span
						class="absolute bottom-0 right-0 left-4 h-px bg-border-soft"
						aria-hidden="true"
					></span>
				{/if}
				<span class="text-[length:var(--text-code)] font-normal text-text"
					>{lang.name}</span
				>
				{#if $languageStore === lang.code}
					<svg
						class="h-4 w-4 shrink-0 text-accent"
						viewBox="0 0 24 24"
						fill="none"
						stroke="currentColor"
						stroke-width="2.3"
						stroke-linecap="round"
						stroke-linejoin="round"
						aria-hidden="true"
					>
						<path d="M20 6L9 17l-5-5" />
					</svg>
				{/if}
			</button>
		{/each}
	</div>

	{#if swSupported}
		<div class="mt-6 overflow-hidden rounded-inset bg-surface">
			<button
				type="button"
				onclick={handleReregister}
				disabled={isReregistering}
				class="flex w-full items-center justify-between gap-3 px-4 py-3.5 text-left"
			>
				<span class="min-w-0">
					<span
						class="block text-[length:var(--text-code)] font-normal text-text"
					>
						{isReregistering
							? tr('pwa.reregistering')
							: tr('pwa.reregisterButton')}
					</span>
					<span class="mt-0.25 block text-body-sm text-text-subtle">
						{tr('pwa.reregisterDesc')}
					</span>
				</span>
			</button>
		</div>
	{/if}

	<div class="mt-6 overflow-hidden rounded-inset bg-surface">
		<button
			type="button"
			onclick={handleLogout}
			class="w-full px-4 py-3.5 text-center text-[length:var(--text-code)] font-semibold text-danger-600"
		>
			{tr('nav.logout')}
		</button>
	</div>
</div>

{#if showDeleteModal}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-scrim p-4"
		role="dialog"
		aria-modal="true"
		aria-labelledby="ios-delete-modal-title"
		tabindex="-1"
		onclick={handleBackdropClick}
		onkeydown={handleModalKeydown}
	>
		<div class="w-full max-w-md rounded-modal bg-surface p-6 shadow-modal">
			<h3 id="ios-delete-modal-title" class="mb-2 text-heading text-danger-600">
				{tr('settings.dangerZone.deleteConfirmTitle')}
			</h3>
			<p class="mb-4 text-body text-text-muted">
				{tr('settings.dangerZone.deleteConfirmMessage')}
			</p>

			<form onsubmit={handleDeleteAccount} class="ios-glass-form space-y-4">
				<div>
					<label
						for="ios-deleteConfirmation"
						class="mb-1 block text-body-sm text-text-ink2"
					>
						{tr('settings.dangerZone.deleteConfirmPlaceholder')}
					</label>
					<input
						id="ios-deleteConfirmation"
						type="text"
						bind:value={deleteConfirmation}
						disabled={isDeleting}
						autocomplete="off"
						class="input"
						placeholder={tr('settings.dangerZone.deleteConfirmWord')}
					/>
				</div>

				{#if isLocalAuth}
					<div>
						<label
							for="ios-deletePassword"
							class="mb-1 block text-body-sm text-text-ink2"
						>
							{tr('settings.dangerZone.passwordRequired')}
						</label>
						<input
							id="ios-deletePassword"
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
							(isLocalAuth && !deletePassword)}
						class="flex-1 rounded-full bg-danger-600 px-4 py-2 text-body font-semibold text-on-accent transition-colors hover:bg-danger-700 disabled:cursor-not-allowed disabled:opacity-50"
					>
						{isDeleting
							? tr('settings.dangerZone.deleting')
							: tr('settings.dangerZone.deleteButton')}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}
