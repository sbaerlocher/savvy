<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import {
		profileApi,
		sessionsApi,
		type ProfileDTO,
		type SessionDTO
	} from '$lib/api';
	import TwoFactorSettings from '$lib/components/TwoFactorSettings.svelte';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import { authStore } from '$lib/stores/auth';
	import { configStore } from '$lib/stores/config';
	import { t } from '$lib/stores/i18n';
	import { toastStore } from '$lib/stores/toast';
	import { logger } from '$lib/utils/logger';
	import { onMount } from 'svelte';
	import { get } from 'svelte/store';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	const pageLogger = logger.child('SecurityPage');
	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);

	let profile = $state<ProfileDTO | null>(null);
	let isLoadingProfile = $state(true);

	// Password state
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

	// Sessions state
	let sessions = $state<SessionDTO[]>([]);
	let isLoadingSessions = $state(true);
	let revokingSessionId = $state<string | null>(null);
	let isRevokingOthers = $state(false);

	onMount(async () => {
		if (!$authStore.isAuthenticated) {
			goto(resolve('/login'));
			return;
		}

		await Promise.all([loadProfile(), loadSessions()]);
	});

	async function loadProfile() {
		try {
			const response = await profileApi.get();
			profile = response.profile;
		} catch (error) {
			pageLogger.error('Failed to load profile', { error });
			toastStore.error(tr('common.error'));
		} finally {
			isLoadingProfile = false;
		}
	}

	async function loadSessions() {
		try {
			const response = await sessionsApi.list();
			sessions = response.sessions || [];
		} catch (error) {
			pageLogger.error('Failed to load sessions', { error });
		} finally {
			isLoadingSessions = false;
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
		} catch {
			toastStore.error(tr('settings.password.error'));
		} finally {
			isChangingPassword = false;
		}
	}

	function formatRelativeTime(dateStr: string): string {
		try {
			const date = new Date(dateStr);
			const now = new Date();
			const diffMs = now.getTime() - date.getTime();
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

<svelte:head>
	<title>{tr('nav.security')} - {tr('common.appName')}</title>
</svelte:head>

<div class="px-4 max-w-7xl mx-auto">
	<PageHeader title={tr('nav.security')} />

	{#if isLoadingProfile}
		<LoadingSpinner />
	{:else if profile}
		<div class="flex flex-col lg:flex-row gap-6 items-start">
			<!-- Password + 2FA (on mobile: first, on desktop: left side) -->
			<div class="w-full lg:w-2/3 space-y-6">
				<!-- Password Change -->
				{#if profile.auth_provider === 'local'}
					<div class="overflow-hidden rounded-xl border border-border bg-white">
						<div class="p-6">
							<h3 class="text-lg font-semibold text-text mb-4">
								{tr('settings.password.title')}
							</h3>

							<form onsubmit={handleChangePassword} class="space-y-4">
								<div>
									<label
										for="currentPassword"
										class="block text-sm font-medium text-text-ink2 mb-1"
									>
										{tr('settings.password.currentPassword')}
									</label>
									<input
										id="currentPassword"
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
										for="newPassword"
										class="block text-sm font-medium text-text-ink2 mb-1"
									>
										{tr('settings.password.newPassword')}
									</label>
									<input
										id="newPassword"
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
														class="h-1 flex-1 rounded-full transition-colors {passwordStrength.score >=
														i
															? passwordStrength.score <= 2
																? 'bg-red-400'
																: passwordStrength.score === 3
																	? 'bg-yellow-400'
																	: 'bg-green-400'
															: 'bg-border'}"
													></div>
												{/each}
											</div>
											{#if passwordStrength.label}
												<p class="text-xs text-red-500 mt-1">
													{passwordStrength.label}
												</p>
											{/if}
										</div>
									{/if}
								</div>

								<div>
									<label
										for="confirmPassword"
										class="block text-sm font-medium text-text-ink2 mb-1"
									>
										{tr('settings.password.confirmPassword')}
									</label>
									<input
										id="confirmPassword"
										type="password"
										autocomplete="new-password"
										required
										bind:value={confirmPassword}
										disabled={isChangingPassword}
										class="input"
									/>
								</div>

								<div class="pt-2">
									<button
										type="submit"
										disabled={isChangingPassword}
										class="btn btn-primary"
									>
										{#if isChangingPassword}
											<span class="relative inline-flex h-3 w-3 mr-2"
												><span
													class="animate-ping absolute inline-flex h-full w-full rounded-full bg-accent-400 opacity-75"
												></span><span
													class="relative inline-flex rounded-full h-3 w-3 bg-accent"
												></span></span
											>
											{tr('settings.password.changing')}
										{:else}
											{tr('settings.password.changeButton')}
										{/if}
									</button>
								</div>
							</form>
						</div>
					</div>
				{:else}
					<div class="overflow-hidden rounded-xl border border-border bg-white">
						<div class="p-6">
							<h3 class="text-lg font-semibold text-text mb-4">
								{tr('settings.password.title')}
							</h3>
							<p class="text-sm text-text-subtle">
								{tr('settings.password.oauthNote')}
							</p>
						</div>
					</div>
				{/if}

				<!-- 2FA: always shown so users with existing 2FA can manage it
				     (status/disable/backup codes) even when new enrollment is
				     disabled server-side. The component loads its own status and
				     only offers setup when enrollmentEnabled. -->
				{#if profile}
					<TwoFactorSettings
						authProvider={profile.auth_provider || 'local'}
						enrollmentEnabled={$configStore.two_factor_enabled}
					/>
				{/if}
			</div>

			<!-- Active Sessions (on mobile: second, on desktop: right side) -->
			<div class="w-full lg:w-1/3">
				<div
					class="overflow-hidden rounded-xl border border-border bg-white p-6"
				>
					<div class="flex items-center justify-between mb-4">
						<h3 class="text-lg font-semibold text-text">
							{tr('settings.sessions.title')}
						</h3>
						{#if sessions.length > 1}
							<button
								type="button"
								onclick={handleRevokeOthers}
								disabled={isRevokingOthers}
								class="text-xs text-accent hover:text-accent-hover font-medium disabled:opacity-50"
							>
								{#if isRevokingOthers}
									...
								{:else}
									{tr('settings.sessions.revokeOthers')}
								{/if}
							</button>
						{/if}
					</div>

					{#if isLoadingSessions}
						<div class="flex justify-center py-4">
							<span class="relative inline-flex h-4 w-4"
								><span
									class="animate-ping absolute inline-flex h-full w-full rounded-full bg-accent-400 opacity-75"
								></span><span
									class="relative inline-flex rounded-full h-4 w-4 bg-accent"
								></span></span
							>
						</div>
					{:else if sessions.length === 0}
						<p class="text-sm text-text-subtle">
							{tr('settings.sessions.noOtherSessions')}
						</p>
					{:else}
						<div class="space-y-3">
							{#each sessions as session (session.id)}
								<div
									class="flex items-start justify-between gap-3 p-3 rounded-lg {session.is_current
										? 'bg-accent-50 border border-accent-200'
										: 'bg-surface-1'}"
								>
									<div class="flex-1 min-w-0">
										<div class="flex items-center gap-2 flex-wrap">
											<span class="text-sm font-medium text-text">
												{session.browser_info ||
													tr('settings.sessions.unknownBrowser')}
											</span>
											<span class="text-xs text-text-subtle">
												{session.device_info ||
													tr('settings.sessions.unknownDevice')}
											</span>
											{#if session.is_current}
												<span
													class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800"
												>
													{tr('settings.sessions.current')}
												</span>
											{/if}
										</div>
										<div
											class="mt-1 flex items-center gap-3 text-xs text-text-subtle"
										>
											{#if session.ip_address}
												<span>{session.ip_address}</span>
											{/if}
											<span
												>{tr('settings.sessions.lastActive')}:
												{formatRelativeTime(session.last_active_at)}</span
											>
										</div>
									</div>
									{#if !session.is_current}
										<button
											type="button"
											onclick={() => handleRevokeSession(session.id)}
											disabled={revokingSessionId === session.id}
											class="flex-shrink-0 p-1.5 text-text-faint hover:text-red-500 transition-colors disabled:opacity-50"
											title={tr('settings.sessions.revoke')}
											aria-label={tr('settings.sessions.revoke')}
										>
											{#if revokingSessionId === session.id}
												<span class="relative inline-flex h-3 w-3"
													><span
														class="animate-ping absolute inline-flex h-full w-full rounded-full bg-accent-400 opacity-75"
													></span><span
														class="relative inline-flex rounded-full h-3 w-3 bg-accent"
													></span></span
												>
											{:else}
												<svg
													class="w-4 h-4"
													fill="none"
													stroke="currentColor"
													viewBox="0 0 24 24"
												>
													<path
														stroke-linecap="round"
														stroke-linejoin="round"
														stroke-width="2"
														d="M6 18L18 6M6 6l12 12"
													/>
												</svg>
											{/if}
										</button>
									{/if}
								</div>
							{/each}
						</div>
					{/if}
				</div>
			</div>
		</div>
	{/if}
</div>
