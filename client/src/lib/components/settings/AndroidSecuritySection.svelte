<script lang="ts">
	import type { ProfileDTO, SessionDTO } from '$lib/api';
	import { authApi, profileApi, sessionsApi } from '$lib/api';
	import { configStore } from '$lib/stores/config';
	import { t } from '$lib/stores/i18n';
	import { toastStore } from '$lib/stores/toast';
	import { logger } from '$lib/utils/logger';
	import { onMount } from 'svelte';
	import { get } from 'svelte/store';
	import M3SettingsRow from './M3SettingsRow.svelte';
	import ToggleSwitch from './ToggleSwitch.svelte';
	import TwoFactorSettings from '$lib/components/TwoFactorSettings.svelte';

	const sectionLogger = logger.child('AndroidSecuritySection');
	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);

	interface Props {
		profile: ProfileDTO;
	}

	let { profile }: Props = $props();

	// Mockup glyph paths (screen-SettingsAndroid).
	const ICON_PADLOCK_BODY = 'M4 10h16v11H4z';
	const ICON_PADLOCK_SHACKLE = 'M8 10V7a4 4 0 018 0v3';
	const ICON_SHIELD = 'M12 3l7 3v6c0 4-3 7-7 9-4-2-7-5-7-9V6z';
	const ICON_SHIELD_CHECK = 'M9.5 12l1.8 1.8L15 10';
	const ICON_PHONE = 'M6 2h12v20H6z';
	const ICON_PHONE_SPEAKER = 'M11 18h2';
	const ICON_LAPTOP = 'M3 4h18v12H3z';
	const ICON_LAPTOP_BASE = 'M8 20h8M12 16v4';
	const ICON_CHEVRON_RIGHT = 'M9 5l7 7-7 7';
	const ICON_X = 'M6 6l12 12M18 6L6 18';

	// Password
	let showPasswordForm = $state(false);
	let currentPassword = $state('');
	let newPassword = $state('');
	let confirmPassword = $state('');
	let isChangingPassword = $state(false);

	// Password strength (mirrors backend: min 12 chars, 3/4 complexity)
	let passwordStrength = $derived.by(() => {
		if (!newPassword) return { score: 0, label: '' };

		const hasUpper = /[A-Z]/.test(newPassword);
		const hasLower = /[a-z]/.test(newPassword);
		const hasDigit = /[0-9]/.test(newPassword);
		const hasSpecial = /[^a-zA-Z0-9]/.test(newPassword);
		const complexity = [hasUpper, hasLower, hasDigit, hasSpecial].filter(
			Boolean
		).length;

		if (newPassword.length < 12)
			return { score: 1, label: tr('settings.password.tooShort') };
		if (complexity < 3)
			return { score: 2, label: tr('settings.password.tooWeak') };
		return { score: complexity >= 4 ? 4 : 3, label: '' };
	});

	// 2FA — the row shows a switch; enabling opens the existing setup flow,
	// disabling goes through TwoFactorSettings' own confirmation.
	let twoFactorEnabled = $state(false);
	let showTwoFactor = $state(false);

	// Sessions
	let sessions = $state<SessionDTO[]>([]);
	let revokingSessionId = $state<string | null>(null);
	let isRevokingOthers = $state(false);

	onMount(async () => {
		await Promise.all([loadSessions(), loadTwoFactorStatus()]);
	});

	async function loadSessions() {
		try {
			const response = await sessionsApi.list();
			sessions = response.sessions || [];
		} catch (error) {
			sectionLogger.error('Failed to load sessions', { error });
		}
	}

	async function loadTwoFactorStatus() {
		try {
			const status = await authApi.get2FAStatus();
			twoFactorEnabled = status.enabled;
		} catch (error) {
			sectionLogger.error('Failed to load 2FA status', { error });
		}
	}

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
			showPasswordForm = false;
		} catch {
			toastStore.error(tr('settings.password.error'));
		} finally {
			isChangingPassword = false;
		}
	}

	function formatRelativeTime(dateStr: string): string {
		try {
			const date = new Date(dateStr);
			const diffMs = Date.now() - date.getTime();
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
</script>

<h2 class="text-label px-6 pt-2.5 pb-2 text-accent">
	{tr('settings.sections.security')}
</h2>

{#if profile.auth_provider === 'local'}
	<M3SettingsRow
		icon="{ICON_PADLOCK_BODY} {ICON_PADLOCK_SHACKLE}"
		title={tr('settings.password.title')}
		subtitle={tr('settings.password.androidSubtitle')}
		onclick={() => (showPasswordForm = !showPasswordForm)}
	>
		{#snippet trailing()}
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
				<path d={ICON_CHEVRON_RIGHT} />
			</svg>
		{/snippet}
	</M3SettingsRow>

	{#if showPasswordForm}
		<form onsubmit={handleChangePassword} class="space-y-4 px-6 pt-1 pb-4">
			<div>
				<label
					for="androidCurrentPassword"
					class="mb-1 block text-label font-medium text-text-ink2"
				>
					{tr('settings.password.currentPassword')}
				</label>
				<input
					id="androidCurrentPassword"
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
					for="androidNewPassword"
					class="mb-1 block text-label font-medium text-text-ink2"
				>
					{tr('settings.password.newPassword')}
				</label>
				<input
					id="androidNewPassword"
					type="password"
					autocomplete="new-password"
					required
					bind:value={newPassword}
					disabled={isChangingPassword}
					class="input"
				/>
				{#if newPassword}
					<div class="mt-2">
						<div class="flex gap-1">
							{#each [1, 2, 3, 4] as i (i)}
								<div
									class="h-1 flex-1 rounded-m3-full transition-colors {passwordStrength.score >=
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
					</div>
				{/if}
			</div>

			<div>
				<label
					for="androidConfirmPassword"
					class="mb-1 block text-label font-medium text-text-ink2"
				>
					{tr('settings.password.confirmPassword')}
				</label>
				<input
					id="androidConfirmPassword"
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

<M3SettingsRow
	icon="{ICON_SHIELD} {ICON_SHIELD_CHECK}"
	title={tr('settings.twoFactor.title')}
	subtitle={tr('settings.twoFactor.androidSubtitle')}
>
	{#snippet trailing()}
		<ToggleSwitch
			bare
			checked={twoFactorEnabled}
			label={tr('settings.twoFactor.title')}
			onToggle={() => (showTwoFactor = !showTwoFactor)}
		/>
	{/snippet}
</M3SettingsRow>

<!-- The switch opens the shared 2FA flow (setup, verify, backup codes,
     disable); it owns the actual enable/disable calls. -->
{#if showTwoFactor}
	<div class="px-6 pt-1 pb-4">
		<TwoFactorSettings
			authProvider={profile.auth_provider || 'local'}
			enrollmentEnabled={$configStore.two_factor_enabled}
		/>
	</div>
{/if}

<h3 class="text-label px-6 pt-3.5 pb-1 text-text-ink2">
	{tr('settings.sessions.title')}
</h3>

{#each sessions as session (session.id)}
	<div
		class="flex w-full items-center gap-4 px-6 py-2.75 {session.is_current
			? 'bg-accent-50'
			: 'active:bg-ground-active'}"
	>
		<span
			class="flex h-10 w-10 shrink-0 items-center justify-center rounded-m3-full {session.is_current
				? 'bg-accent-100 text-accent-850'
				: 'bg-tile-tint text-text-subtle'}"
		>
			<svg
				class="h-4.5 w-4.5"
				viewBox="0 0 24 24"
				fill="none"
				stroke="currentColor"
				stroke-width="2"
				stroke-linecap="round"
				stroke-linejoin="round"
				aria-hidden="true"
			>
				{#if session.is_current}
					<path d={ICON_PHONE} />
					<path d={ICON_PHONE_SPEAKER} />
				{:else}
					<path d={ICON_LAPTOP} />
					<path d={ICON_LAPTOP_BASE} />
				{/if}
			</svg>
		</span>
		<div class="min-w-0 flex-1">
			<div class="flex items-center gap-1.75">
				<span class="text-body font-semibold text-text">
					{session.device_info || tr('settings.sessions.unknownDevice')} · {session.browser_info ||
						tr('settings.sessions.unknownBrowser')}
				</span>
				{#if session.is_current}
					<span
						class="text-tag rounded-m3-full bg-accent-100 px-2 py-0.25 text-accent-850"
					>
						{tr('settings.sessions.androidCurrent')}
					</span>
				{/if}
			</div>
			<div class="mt-0.5 font-mono text-body-sm text-text-muted">
				{#if session.ip_address}{session.ip_address} ·
				{/if}{formatRelativeTime(session.last_active_at)}
			</div>
		</div>
		{#if !session.is_current}
			<button
				type="button"
				onclick={() => handleRevokeSession(session.id)}
				disabled={revokingSessionId === session.id}
				title={tr('settings.sessions.revoke')}
				aria-label={tr('settings.sessions.revoke')}
				class="flex h-9.5 w-9.5 shrink-0 items-center justify-center rounded-m3-full text-text-subtle transition-colors active:bg-danger-50 active:text-danger-600 disabled:opacity-50"
			>
				<svg
					class="h-4 w-4"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="2.2"
					stroke-linecap="round"
					aria-hidden="true"
				>
					<path d={ICON_X} />
				</svg>
			</button>
		{/if}
	</div>
{:else}
	<p class="px-6 py-3 text-body text-text-subtle">
		{tr('settings.sessions.noOtherSessions')}
	</p>
{/each}

{#if sessions.length > 1}
	<div class="px-6 pt-2.5 pb-1">
		<button
			type="button"
			onclick={handleRevokeOthers}
			disabled={isRevokingOthers}
			class="text-label inline-flex items-center gap-2 rounded-m3-full border border-border-chip px-5 py-2.5 text-accent transition-colors active:bg-accent-50 disabled:opacity-50"
		>
			{tr('settings.sessions.revokeOthers')}
		</button>
	</div>
{/if}
