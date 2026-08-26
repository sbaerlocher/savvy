<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { t } from '$lib/stores/i18n';
	import { toastStore } from '$lib/stores/toast';
	import { logger } from '$lib/utils/logger';
	import { adminApi } from '$lib/api';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import { platform } from '$lib/utils/platform';
	import { ICON_WARNING } from '$lib/icons';

	const pageLogger = logger.child('SystemHealthPage');

	type HealthStatus = 'ready' | 'degraded' | 'not_ready';

	interface CheckResult {
		status: 'healthy' | 'unhealthy' | 'not_configured';
		enabled: boolean;
		latency_ms?: number;
		error?: string;
	}

	interface SystemHealth {
		status: HealthStatus;
		timestamp: string;
		checks: {
			database: CheckResult;
			smtp: CheckResult;
			oauth: CheckResult;
			vapid: CheckResult;
			totp_encryption: CheckResult;
		};
	}

	let health = $state<SystemHealth | null>(null);
	let isLoading = $state(true);
	let isSendingEmail = $state(false);
	let isSendingPush = $state(false);
	let expandedService = $state<string | null>(null);

	function toggleService(service: string) {
		expandedService = expandedService === service ? null : service;
	}
	let autoRefresh = $state(true);
	let refreshInterval: ReturnType<typeof setInterval> | null = null;
	let lastRefresh = $state<Date | null>(null);

	// Auto-refresh every 30 seconds
	const REFRESH_INTERVAL = 30000;

	onMount(async () => {
		pageLogger.debug('System Health page loaded');
		await loadHealthStatus();
		startAutoRefresh();
	});

	onDestroy(() => {
		stopAutoRefresh();
	});

	async function loadHealthStatus() {
		isLoading = true;
		try {
			const response = await fetch('/api/v1/admin/system-health');
			if (response.ok) {
				health = await response.json();
				lastRefresh = new Date();
			} else {
				toastStore.error($t('admin.systemHealth.loadError'));
			}
		} catch (error) {
			pageLogger.error('Failed to load health status', { error });
			toastStore.error($t('admin.systemHealth.loadError'));
		} finally {
			isLoading = false;
		}
	}

	async function sendTestEmail() {
		if (isSendingEmail) return;

		isSendingEmail = true;
		try {
			const data = await adminApi.sendTestEmail();
			toastStore.success(
				data.message || $t('admin.systemHealth.sendTestEmail')
			);
		} catch (error) {
			pageLogger.error('Failed to send test email', { error });
			toastStore.error($t('admin.systemHealth.sendTestEmail'));
		} finally {
			isSendingEmail = false;
		}
	}

	async function sendTestPush() {
		if (isSendingPush) return;

		isSendingPush = true;
		try {
			const data = await adminApi.sendTestPush();
			toastStore.success(data.message || $t('admin.systemHealth.sendTestPush'));
		} catch (error) {
			pageLogger.error('Failed to send test push', { error });
			toastStore.error($t('admin.systemHealth.sendTestPush'));
		} finally {
			isSendingPush = false;
		}
	}

	function startAutoRefresh() {
		if (!autoRefresh) return;
		refreshInterval = setInterval(loadHealthStatus, REFRESH_INTERVAL);
	}

	function stopAutoRefresh() {
		if (refreshInterval) {
			clearInterval(refreshInterval);
			refreshInterval = null;
		}
	}

	function toggleAutoRefresh() {
		autoRefresh = !autoRefresh;
		if (autoRefresh) {
			startAutoRefresh();
		} else {
			stopAutoRefresh();
		}
	}

	function getStatusColor(check: CheckResult): string {
		if (!check.enabled) return 'text-text-faint';
		return check.status === 'healthy' ? 'text-success-600' : 'text-danger-600';
	}

	function getStatusLabel(status: string): string {
		switch (status) {
			case 'healthy':
				return $t('admin.systemHealth.healthy');
			case 'unhealthy':
				return $t('admin.systemHealth.unhealthy');
			case 'not_configured':
				return $t('admin.systemHealth.notConfigured');
			default:
				return status;
		}
	}

	// Desktop mockup (screen-AdminDesktop, "System-Health · Tabelle") renders the
	// services from one list into a four-column grid; the native platforms keep
	// the expandable card list. `platform` is a module constant, so a plain
	// const, not $derived.
	const DESKTOP = platform === 'other';

	const ICON_DB =
		'M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4';
	const ICON_MAIL =
		'M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z';
	const ICON_KEY =
		'M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z';
	const ICON_BELL_HEALTH =
		'M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9';
	const ICON_LOCK_HEALTH =
		'M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z';
	const ICON_SEND = 'M12 19l9 2-9-18-9 18 9-2zm0 0v-8';

	type ServiceRow = {
		key: keyof SystemHealth['checks'];
		icon: string;
		name: string;
		detail: string;
		check: CheckResult;
		test: 'email' | 'push' | null;
	};

	// Detail column text per service — mirrors the per-service copy the card list
	// shows when expanded.
	function serviceDetail(
		key: keyof SystemHealth['checks'],
		check: CheckResult
	): string {
		if (check.error) return check.error;
		switch (key) {
			case 'database':
				if (check.latency_ms !== undefined)
					return $t('admin.systemHealth.latency').replace(
						'{ms}',
						check.latency_ms.toString()
					);
				if (!check.enabled) return $t('admin.systemHealth.notConfiguredDesc');
				return $t('admin.systemHealth.criticalService');
			case 'smtp':
				return check.enabled ? '' : $t('admin.systemHealth.emailDisabled');
			case 'oauth':
				return check.enabled
					? $t('admin.systemHealth.oauthConnectivity')
					: $t('admin.systemHealth.oauthDisabled');
			case 'vapid':
				return check.enabled ? '' : $t('admin.systemHealth.pushDisabled');
			case 'totp_encryption':
				return check.enabled
					? $t('admin.systemHealth.encryptionValid')
					: $t('admin.systemHealth.totpDisabled');
		}
	}

	const serviceRows = $derived<ServiceRow[]>(
		health
			? (
					[
						['database', ICON_DB, $t('admin.systemHealth.database'), null],
						['smtp', ICON_MAIL, $t('admin.systemHealth.smtp'), 'email'],
						['oauth', ICON_KEY, $t('admin.systemHealth.oauth'), null],
						[
							'vapid',
							ICON_BELL_HEALTH,
							$t('admin.systemHealth.pushNotifications'),
							'push'
						],
						[
							'totp_encryption',
							ICON_LOCK_HEALTH,
							$t('admin.systemHealth.totp'),
							null
						]
					] as const
				).map(([key, icon, name, test]) => ({
					key,
					icon,
					name,
					test,
					check: health!.checks[key],
					detail: serviceDetail(key, health!.checks[key])
				}))
			: []
	);

	// Status badge tone — same three-way split the card list uses.
	function badgeTone(check: CheckResult): string {
		if (!check.enabled) return 'bg-border-soft text-text-strong';
		return check.status === 'healthy'
			? 'bg-success-100 text-success-800'
			: 'bg-danger-100 text-danger-800';
	}

	function statusDot(check: CheckResult): string {
		if (!check.enabled) return 'bg-text-faint';
		return check.status === 'healthy' ? 'bg-success' : 'bg-danger-600';
	}

	function getTimeAgo(date: Date | null): string {
		if (!date) return '-';
		const seconds = Math.floor((new Date().getTime() - date.getTime()) / 1000);
		if (seconds < 60)
			return $t('admin.systemHealth.secondsAgo').replace(
				'{n}',
				seconds.toString()
			);
		const minutes = Math.floor(seconds / 60);
		if (minutes < 60)
			return $t('admin.systemHealth.minutesAgo').replace(
				'{n}',
				minutes.toString()
			);
		const hours = Math.floor(minutes / 60);
		return $t('admin.systemHealth.hoursAgo').replace('{n}', hours.toString());
	}
</script>

<svelte:head>
	<title>{$t('nav.adminSystemHealth')} - {$t('common.appName')}</title>
</svelte:head>

<div class="px-4 pb-20 md:pb-4">
	{#if DESKTOP}
		<!-- Desktop mockup: title, status pill, auto-refresh control and the
		     service table all live inside one elevated panel. -->
		<div
			class="overflow-hidden rounded-4xl border border-border bg-surface shadow-panel"
		>
			<div class="flex items-end justify-between px-7.5 pt-6 pb-4.5">
				<div>
					<h1 class="text-heading text-text">
						{$t('nav.adminSystemHealth')}
					</h1>
					<p class="mt-0.5 text-label font-normal text-text-subtle">
						{$t('admin.systemHealth.subtitle')}
					</p>
				</div>

				<div class="flex items-center gap-2.5">
					{#if health}
						{@const tone =
							health.status === 'ready'
								? 'bg-success-50 border-success-200 text-success-800'
								: health.status === 'degraded'
									? 'bg-warning-50 border-warning-200 text-warning-800'
									: 'bg-danger-50 border-danger-200 text-danger-800'}
						{@const iconStroke =
							health.status === 'ready'
								? 'text-success-600'
								: health.status === 'degraded'
									? 'text-warning-600'
									: 'text-danger-600'}
						<span
							class="inline-flex h-10 items-center gap-2 rounded-lg border px-3.5 text-label {tone}"
						>
							<svg
								class="h-4.25 w-4.25 shrink-0 {iconStroke}"
								fill="none"
								stroke="currentColor"
								stroke-width="2"
								viewBox="0 0 24 24"
								aria-hidden="true"
							>
								{#if health.status === 'ready'}
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
									/>
								{:else if health.status === 'degraded'}
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										d={ICON_WARNING}
									/>
								{:else}
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z"
									/>
								{/if}
							</svg>
							{health.status === 'ready'
								? $t('admin.systemHealth.ready')
								: health.status === 'degraded'
									? $t('admin.systemHealth.degraded')
									: $t('admin.systemHealth.notReady')}
						</span>
					{/if}

					<button
						type="button"
						onclick={toggleAutoRefresh}
						class="inline-flex h-10 items-center gap-2 rounded-lg border-2 bg-surface px-3.5 text-body-sm font-semibold transition-colors {autoRefresh
							? 'border-accent text-accent-hover'
							: 'border-border-field text-text-ink2'}"
					>
						<svg
							class="h-4 w-4 shrink-0"
							fill="none"
							stroke="currentColor"
							stroke-width="2"
							viewBox="0 0 24 24"
							aria-hidden="true"
						>
							{#if autoRefresh}
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
								/>
							{:else}
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									d="M10 9v6m4-6v6m7-3a9 9 0 11-18 0 9 9 0 0118 0z"
								/>
							{/if}
						</svg>
						{autoRefresh
							? $t('admin.systemHealth.autoRefreshInterval').replace(
									'{s}',
									(REFRESH_INTERVAL / 1000).toString()
								)
							: $t('admin.systemHealth.paused')}
					</button>

					<button
						type="button"
						onclick={loadHealthStatus}
						disabled={isLoading}
						class="font-mono text-mono-sm text-text-faint transition-colors hover:text-text-muted disabled:cursor-not-allowed disabled:opacity-50"
						title={$t('admin.systemHealth.refreshNow')}
					>
						{isLoading
							? $t('admin.systemHealth.refreshing')
							: $t('admin.systemHealth.updatedAgo').replace(
									'{ago}',
									getTimeAgo(lastRefresh)
								)}
					</button>
				</div>
			</div>

			{#if isLoading && !health}
				<div class="border-t border-border-soft"><LoadingSpinner /></div>
			{:else if health}
				<div class="border-t border-border-soft">
					<div
						class="grid grid-cols-[1.4fr_2.4fr_1.4fr_1fr] border-b border-border-soft bg-surface-1 px-7.5 py-3"
					>
						<span class="text-section-eyebrow uppercase text-text-subtle"
							>{$t('admin.systemHealth.service')}</span
						>
						<span class="text-section-eyebrow uppercase text-text-subtle"
							>{$t('admin.systemHealth.details')}</span
						>
						<span class="text-section-eyebrow uppercase text-text-subtle"
							>{$t('admin.users.actions')}</span
						>
						<span
							class="text-section-eyebrow text-right uppercase text-text-subtle"
							>{$t('admin.systemHealth.statusLabel')}</span
						>
					</div>

					{#each serviceRows as row (row.key)}
						<div
							data-testid="health-service-row"
							class="grid grid-cols-[1.4fr_2.4fr_1.4fr_1fr] items-center border-b border-border-soft px-7.5 py-3.75 transition-colors hover:bg-surface-1"
						>
							<span class="flex items-center gap-3">
								<span
									class="flex h-8.5 w-8.5 flex-none items-center justify-center rounded-md border border-border-soft bg-surface-2"
								>
									<svg
										class="h-4.25 w-4.25 {getStatusColor(row.check)}"
										fill="none"
										stroke="currentColor"
										stroke-width="2"
										viewBox="0 0 24 24"
										aria-hidden="true"
									>
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											d={row.icon}
										/>
									</svg>
								</span>
								<span class="text-body font-semibold text-text">{row.name}</span
								>
							</span>

							<span
								class="text-label font-normal {row.check.status === 'unhealthy'
									? 'text-danger-600'
									: 'text-text-muted'}">{row.detail}</span
							>

							<span>
								{#if row.test && row.check.enabled}
									{@const busy =
										row.test === 'email' ? isSendingEmail : isSendingPush}
									<button
										type="button"
										onclick={row.test === 'email'
											? sendTestEmail
											: sendTestPush}
										disabled={busy}
										class="inline-flex h-8.5 items-center gap-1.5 rounded-md border border-border-field bg-surface px-3.25 text-body-sm font-semibold text-text transition-colors hover:bg-surface-1 disabled:cursor-not-allowed disabled:opacity-50"
									>
										{#if busy}
											{$t('admin.systemHealth.testing')}
										{:else}
											<svg
												class="h-3.75 w-3.75"
												fill="none"
												stroke="currentColor"
												stroke-width="2"
												viewBox="0 0 24 24"
												aria-hidden="true"
											>
												<path
													stroke-linecap="round"
													stroke-linejoin="round"
													d={ICON_SEND}
												/>
											</svg>
											{row.test === 'email'
												? $t('admin.systemHealth.sendTestEmail')
												: $t('admin.systemHealth.sendTestPush')}
										{/if}
									</button>
								{/if}
							</span>

							<span class="flex justify-end">
								<span
									class="inline-flex flex-none items-center gap-1.5 rounded-full px-2.5 py-0.75 text-eyebrow whitespace-nowrap {badgeTone(
										row.check
									)}"
								>
									<span class="h-1.5 w-1.5 rounded-full {statusDot(row.check)}"
									></span>
									{getStatusLabel(row.check.status)}
								</span>
							</span>
						</div>
					{/each}
				</div>
			{:else}
				<div class="border-t border-border-soft px-7.5 py-6">
					<div class="flex items-center gap-3 text-danger-800">
						<svg
							class="h-6 w-6 text-danger-600"
							fill="none"
							stroke="currentColor"
							stroke-width="2"
							viewBox="0 0 24 24"
							aria-hidden="true"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
							/>
						</svg>
						<span>{$t('admin.systemHealth.loadErrorRetry')}</span>
					</div>
				</div>
			{/if}
		</div>
	{:else}
		<!-- Header -->
		<div class="mb-8">
			<h1 class="text-3xl font-bold text-text">
				{$t('nav.adminSystemHealth')}
			</h1>
		</div>

		<!-- Controls + Overall Status -->
		<div class="flex flex-col sm:flex-row gap-3 mb-6">
			<!-- Overall Status (left, where search bar is on other pages) -->
			<div
				class="sm:flex-1 flex items-center gap-3 control px-4 bg-white border border-border-field rounded-md"
			>
				{#if isLoading && !health}
					<span class="text-sm text-text-subtle">{$t('common.loading')}</span>
				{:else if health}
					{#if health.status === 'ready'}
						<svg
							class="w-5 h-5 text-success-600 shrink-0"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
							/>
						</svg>
						<span class="text-sm font-medium text-success-600"
							>{$t('admin.systemHealth.ready')}</span
						>
					{:else if health.status === 'degraded'}
						<svg
							class="w-5 h-5 text-warning-600 shrink-0"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
							/>
						</svg>
						<span class="text-sm font-medium text-warning-600"
							>{$t('admin.systemHealth.degraded')}</span
						>
					{:else}
						<svg
							class="w-5 h-5 text-danger-600 shrink-0"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z"
							/>
						</svg>
						<span class="text-sm font-medium text-danger-600"
							>{$t('admin.systemHealth.notReady')}</span
						>
					{/if}
					{#if lastRefresh}
						<span class="text-xs text-text-faint ml-auto"
							>{getTimeAgo(lastRefresh)}</span
						>
					{/if}
				{:else}
					<span class="text-sm text-text-faint">—</span>
				{/if}
			</div>

			<!-- Action Buttons (right, full-width on mobile) -->
			<div class="grid grid-cols-3 sm:flex gap-3 w-full sm:w-auto control">
				<button
					class="col-span-1 sm:flex-none flex items-center justify-center gap-2 h-full px-4 bg-white border rounded-md hover:bg-surface-1 transition-colors {autoRefresh
						? 'ring-2 ring-accent border-accent'
						: 'border-border-field'}"
					onclick={toggleAutoRefresh}
					title={autoRefresh
						? $t('admin.systemHealth.autoRefresh')
						: $t('admin.systemHealth.paused')}
				>
					{#if autoRefresh}
						<svg
							class="w-5 h-5 text-text-muted"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
							/>
						</svg>
					{:else}
						<svg
							class="w-5 h-5 text-text-muted"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M10 9v6m4-6v6m7-3a9 9 0 11-18 0 9 9 0 0118 0z"
							/>
						</svg>
					{/if}
				</button>
				<button
					class="col-span-2 sm:flex-none btn btn-ghost h-full"
					onclick={loadHealthStatus}
					disabled={isLoading}
				>
					{#if isLoading}
						{$t('admin.systemHealth.refreshing')}
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
								d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
							/>
						</svg>
						{$t('admin.systemHealth.refreshNow')}
					{/if}
				</button>
			</div>
		</div>

		{#if isLoading && !health}
			<LoadingSpinner />
		{:else if health}
			<!-- Mobile: Card List -->
			<div class="md:hidden bg-white shadow rounded-lg divide-y divide-border">
				<!-- Database -->
				<button
					class="w-full px-4 py-3 flex items-center gap-3 text-left"
					onclick={() => toggleService('database')}
				>
					<div class={getStatusColor(health.checks.database)}>
						<svg
							class="w-5 h-5"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4"
							/>
						</svg>
					</div>
					<span class="text-sm font-medium text-text flex-1"
						>{$t('admin.systemHealth.database')}</span
					>
					<span
						class="px-2 py-1 text-xs font-medium rounded-full {health.checks
							.database.enabled
							? health.checks.database.status === 'healthy'
								? 'bg-success-100 text-success-800'
								: 'bg-danger-100 text-danger-800'
							: 'bg-border-soft text-text-strong'}"
					>
						{getStatusLabel(health.checks.database.status)}
					</span>
					<svg
						class="w-4 h-4 text-text-faint transition-transform {expandedService ===
						'database'
							? 'rotate-180'
							: ''}"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M19 9l-7 7-7-7"
						/>
					</svg>
				</button>
				{#if expandedService === 'database'}
					<div class="px-4 py-3 bg-surface-1 text-sm text-text-muted">
						{#if health.checks.database.error}
							<span class="text-danger-600">{health.checks.database.error}</span
							>
						{:else if health.checks.database.latency_ms !== undefined}
							{$t('admin.systemHealth.latency').replace(
								'{ms}',
								health.checks.database.latency_ms.toString()
							)}
						{:else if !health.checks.database.enabled}
							{$t('admin.systemHealth.notConfiguredDesc')}
						{:else}
							{$t('admin.systemHealth.criticalService')}
						{/if}
					</div>
				{/if}

				<!-- SMTP -->
				<button
					class="w-full px-4 py-3 flex items-center gap-3 text-left"
					onclick={() => toggleService('smtp')}
				>
					<div class={getStatusColor(health.checks.smtp)}>
						<svg
							class="w-5 h-5"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
							/>
						</svg>
					</div>
					<span class="text-sm font-medium text-text flex-1"
						>{$t('admin.systemHealth.smtp')}</span
					>
					<span
						class="px-2 py-1 text-xs font-medium rounded-full {health.checks
							.smtp.enabled
							? health.checks.smtp.status === 'healthy'
								? 'bg-success-100 text-success-800'
								: 'bg-danger-100 text-danger-800'
							: 'bg-border-soft text-text-strong'}"
					>
						{getStatusLabel(health.checks.smtp.status)}
					</span>
					<svg
						class="w-4 h-4 text-text-faint transition-transform {expandedService ===
						'smtp'
							? 'rotate-180'
							: ''}"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M19 9l-7 7-7-7"
						/>
					</svg>
				</button>
				{#if expandedService === 'smtp'}
					<div class="px-4 py-3 bg-surface-1 space-y-3">
						<div class="text-sm text-text-muted">
							{#if health.checks.smtp.error}
								<span class="text-danger-600">{health.checks.smtp.error}</span>
							{:else if !health.checks.smtp.enabled}
								{$t('admin.systemHealth.emailDisabled')}
							{/if}
						</div>
						{#if health.checks.smtp.enabled}
							<button
								class="btn btn-sm btn-ghost w-full"
								onclick={sendTestEmail}
								disabled={isSendingEmail}
							>
								{#if isSendingEmail}
									{$t('admin.systemHealth.testing')}
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
											d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8"
										/>
									</svg>
									{$t('admin.systemHealth.sendTestEmail')}
								{/if}
							</button>
						{/if}
					</div>
				{/if}

				<!-- OAuth -->
				<button
					class="w-full px-4 py-3 flex items-center gap-3 text-left"
					onclick={() => toggleService('oauth')}
				>
					<div class={getStatusColor(health.checks.oauth)}>
						<svg
							class="w-5 h-5"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z"
							/>
						</svg>
					</div>
					<span class="text-sm font-medium text-text flex-1"
						>{$t('admin.systemHealth.oauth')}</span
					>
					<span
						class="px-2 py-1 text-xs font-medium rounded-full {health.checks
							.oauth.enabled
							? health.checks.oauth.status === 'healthy'
								? 'bg-success-100 text-success-800'
								: 'bg-danger-100 text-danger-800'
							: 'bg-border-soft text-text-strong'}"
					>
						{getStatusLabel(health.checks.oauth.status)}
					</span>
					<svg
						class="w-4 h-4 text-text-faint transition-transform {expandedService ===
						'oauth'
							? 'rotate-180'
							: ''}"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M19 9l-7 7-7-7"
						/>
					</svg>
				</button>
				{#if expandedService === 'oauth'}
					<div class="px-4 py-3 bg-surface-1 text-sm text-text-muted">
						{#if health.checks.oauth.error}
							<span class="text-danger-600">{health.checks.oauth.error}</span>
						{:else if health.checks.oauth.enabled}
							{$t('admin.systemHealth.oauthConnectivity')}
						{:else}
							{$t('admin.systemHealth.oauthDisabled')}
						{/if}
					</div>
				{/if}

				<!-- VAPID (Push) -->
				<button
					class="w-full px-4 py-3 flex items-center gap-3 text-left"
					onclick={() => toggleService('vapid')}
				>
					<div class={getStatusColor(health.checks.vapid)}>
						<svg
							class="w-5 h-5"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"
							/>
						</svg>
					</div>
					<span class="text-sm font-medium text-text flex-1"
						>{$t('admin.systemHealth.pushNotifications')}</span
					>
					<span
						class="px-2 py-1 text-xs font-medium rounded-full {health.checks
							.vapid.enabled
							? health.checks.vapid.status === 'healthy'
								? 'bg-success-100 text-success-800'
								: 'bg-danger-100 text-danger-800'
							: 'bg-border-soft text-text-strong'}"
					>
						{getStatusLabel(health.checks.vapid.status)}
					</span>
					<svg
						class="w-4 h-4 text-text-faint transition-transform {expandedService ===
						'vapid'
							? 'rotate-180'
							: ''}"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M19 9l-7 7-7-7"
						/>
					</svg>
				</button>
				{#if expandedService === 'vapid'}
					<div class="px-4 py-3 bg-surface-1 space-y-3">
						<div class="text-sm text-text-muted">
							{#if health.checks.vapid.error}
								<span class="text-danger-600">{health.checks.vapid.error}</span>
							{:else if !health.checks.vapid.enabled}
								{$t('admin.systemHealth.pushDisabled')}
							{/if}
						</div>
						{#if health.checks.vapid.enabled}
							<button
								class="btn btn-sm btn-ghost w-full"
								onclick={sendTestPush}
								disabled={isSendingPush}
							>
								{#if isSendingPush}
									{$t('admin.systemHealth.testing')}
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
											d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8"
										/>
									</svg>
									{$t('admin.systemHealth.sendTestPush')}
								{/if}
							</button>
						{/if}
					</div>
				{/if}

				<!-- TOTP (2FA) -->
				<button
					class="w-full px-4 py-3 flex items-center gap-3 text-left"
					onclick={() => toggleService('totp')}
				>
					<div class={getStatusColor(health.checks.totp_encryption)}>
						<svg
							class="w-5 h-5"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"
							/>
						</svg>
					</div>
					<span class="text-sm font-medium text-text flex-1"
						>{$t('admin.systemHealth.totp')}</span
					>
					<span
						class="px-2 py-1 text-xs font-medium rounded-full {health.checks
							.totp_encryption.enabled
							? health.checks.totp_encryption.status === 'healthy'
								? 'bg-success-100 text-success-800'
								: 'bg-danger-100 text-danger-800'
							: 'bg-border-soft text-text-strong'}"
					>
						{getStatusLabel(health.checks.totp_encryption.status)}
					</span>
					<svg
						class="w-4 h-4 text-text-faint transition-transform {expandedService ===
						'totp'
							? 'rotate-180'
							: ''}"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M19 9l-7 7-7-7"
						/>
					</svg>
				</button>
				{#if expandedService === 'totp'}
					<div class="px-4 py-3 bg-surface-1 text-sm text-text-muted">
						{#if health.checks.totp_encryption.error}
							<span class="text-danger-600"
								>{health.checks.totp_encryption.error}</span
							>
						{:else if health.checks.totp_encryption.enabled}
							{$t('admin.systemHealth.encryptionValid')}
						{:else}
							{$t('admin.systemHealth.totpDisabled')}
						{/if}
					</div>
				{/if}
			</div>

			<!-- Desktop: Table -->
			<div class="hidden md:block bg-white shadow rounded-lg overflow-hidden">
				<table class="min-w-full divide-y divide-border">
					<thead class="bg-surface-1">
						<tr>
							<th
								class="px-6 py-3 text-left text-xs font-medium text-text-subtle uppercase tracking-wider"
							>
								{$t('admin.systemHealth.service')}
							</th>
							<th
								class="px-6 py-3 text-left text-xs font-medium text-text-subtle uppercase tracking-wider"
							>
								{$t('admin.systemHealth.details')}
							</th>
							<th
								class="px-6 py-3 text-right text-xs font-medium text-text-subtle uppercase tracking-wider"
							>
								{$t('admin.users.actions')}
							</th>
							<th
								class="px-6 py-3 text-right text-xs font-medium text-text-subtle uppercase tracking-wider"
							>
								{$t('admin.systemHealth.statusLabel')}
							</th>
						</tr>
					</thead>
					<tbody class="bg-white divide-y divide-border">
						<!-- Database -->
						<tr class="hover:bg-surface-1 transition-colors">
							<td class="px-6 py-4 whitespace-nowrap">
								<div class="flex items-center gap-3">
									<div class={getStatusColor(health.checks.database)}>
										<svg
											class="w-5 h-5"
											fill="none"
											stroke="currentColor"
											viewBox="0 0 24 24"
										>
											<path
												stroke-linecap="round"
												stroke-linejoin="round"
												stroke-width="2"
												d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4"
											/>
										</svg>
									</div>
									<span class="text-sm font-medium text-text"
										>{$t('admin.systemHealth.database')}</span
									>
								</div>
							</td>
							<td class="px-6 py-4 whitespace-nowrap text-sm text-text-subtle">
								{#if health.checks.database.error}
									<span class="text-danger-600"
										>{health.checks.database.error}</span
									>
								{:else if health.checks.database.latency_ms !== undefined}
									{$t('admin.systemHealth.latency').replace(
										'{ms}',
										health.checks.database.latency_ms.toString()
									)}
								{:else if !health.checks.database.enabled}
									{$t('admin.systemHealth.notConfiguredDesc')}
								{:else}
									{$t('admin.systemHealth.criticalService')}
								{/if}
							</td>
							<td class="px-6 py-4 whitespace-nowrap text-right"></td>
							<td class="px-6 py-4 whitespace-nowrap text-right">
								<span
									class="px-2 py-1 text-xs font-medium rounded-full {health
										.checks.database.enabled
										? health.checks.database.status === 'healthy'
											? 'bg-success-100 text-success-800'
											: 'bg-danger-100 text-danger-800'
										: 'bg-border-soft text-text-strong'}"
								>
									{getStatusLabel(health.checks.database.status)}
								</span>
							</td>
						</tr>

						<!-- SMTP -->
						<tr class="hover:bg-surface-1 transition-colors">
							<td class="px-6 py-4 whitespace-nowrap">
								<div class="flex items-center gap-3">
									<div class={getStatusColor(health.checks.smtp)}>
										<svg
											class="w-5 h-5"
											fill="none"
											stroke="currentColor"
											viewBox="0 0 24 24"
										>
											<path
												stroke-linecap="round"
												stroke-linejoin="round"
												stroke-width="2"
												d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
											/>
										</svg>
									</div>
									<span class="text-sm font-medium text-text"
										>{$t('admin.systemHealth.smtp')}</span
									>
								</div>
							</td>
							<td class="px-6 py-4 whitespace-nowrap text-sm text-text-subtle">
								{#if health.checks.smtp.error}
									<span class="text-danger-600">{health.checks.smtp.error}</span
									>
								{:else if !health.checks.smtp.enabled}
									{$t('admin.systemHealth.emailDisabled')}
								{/if}
							</td>
							<td class="px-6 py-4 whitespace-nowrap text-right">
								{#if health.checks.smtp.enabled}
									<button
										class="btn btn-sm btn-ghost"
										onclick={sendTestEmail}
										disabled={isSendingEmail}
									>
										{#if isSendingEmail}
											{$t('admin.systemHealth.testing')}
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
													d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8"
												/>
											</svg>
											{$t('admin.systemHealth.sendTestEmail')}
										{/if}
									</button>
								{/if}
							</td>
							<td class="px-6 py-4 whitespace-nowrap text-right">
								<span
									class="px-2 py-1 text-xs font-medium rounded-full {health
										.checks.smtp.enabled
										? health.checks.smtp.status === 'healthy'
											? 'bg-success-100 text-success-800'
											: 'bg-danger-100 text-danger-800'
										: 'bg-border-soft text-text-strong'}"
								>
									{getStatusLabel(health.checks.smtp.status)}
								</span>
							</td>
						</tr>

						<!-- OAuth -->
						<tr class="hover:bg-surface-1 transition-colors">
							<td class="px-6 py-4 whitespace-nowrap">
								<div class="flex items-center gap-3">
									<div class={getStatusColor(health.checks.oauth)}>
										<svg
											class="w-5 h-5"
											fill="none"
											stroke="currentColor"
											viewBox="0 0 24 24"
										>
											<path
												stroke-linecap="round"
												stroke-linejoin="round"
												stroke-width="2"
												d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z"
											/>
										</svg>
									</div>
									<span class="text-sm font-medium text-text"
										>{$t('admin.systemHealth.oauth')}</span
									>
								</div>
							</td>
							<td class="px-6 py-4 whitespace-nowrap text-sm text-text-subtle">
								{#if health.checks.oauth.error}
									<span class="text-danger-600"
										>{health.checks.oauth.error}</span
									>
								{:else if health.checks.oauth.enabled}
									{$t('admin.systemHealth.oauthConnectivity')}
								{:else}
									{$t('admin.systemHealth.oauthDisabled')}
								{/if}
							</td>
							<td class="px-6 py-4 whitespace-nowrap text-right"></td>
							<td class="px-6 py-4 whitespace-nowrap text-right">
								<span
									class="px-2 py-1 text-xs font-medium rounded-full {health
										.checks.oauth.enabled
										? health.checks.oauth.status === 'healthy'
											? 'bg-success-100 text-success-800'
											: 'bg-danger-100 text-danger-800'
										: 'bg-border-soft text-text-strong'}"
								>
									{getStatusLabel(health.checks.oauth.status)}
								</span>
							</td>
						</tr>

						<!-- VAPID (Push) -->
						<tr class="hover:bg-surface-1 transition-colors">
							<td class="px-6 py-4 whitespace-nowrap">
								<div class="flex items-center gap-3">
									<div class={getStatusColor(health.checks.vapid)}>
										<svg
											class="w-5 h-5"
											fill="none"
											stroke="currentColor"
											viewBox="0 0 24 24"
										>
											<path
												stroke-linecap="round"
												stroke-linejoin="round"
												stroke-width="2"
												d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"
											/>
										</svg>
									</div>
									<span class="text-sm font-medium text-text"
										>{$t('admin.systemHealth.pushNotifications')}</span
									>
								</div>
							</td>
							<td class="px-6 py-4 whitespace-nowrap text-sm text-text-subtle">
								{#if health.checks.vapid.error}
									<span class="text-danger-600"
										>{health.checks.vapid.error}</span
									>
								{:else if !health.checks.vapid.enabled}
									{$t('admin.systemHealth.pushDisabled')}
								{/if}
							</td>
							<td class="px-6 py-4 whitespace-nowrap text-right">
								{#if health.checks.vapid.enabled}
									<button
										class="btn btn-sm btn-ghost"
										onclick={sendTestPush}
										disabled={isSendingPush}
									>
										{#if isSendingPush}
											{$t('admin.systemHealth.testing')}
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
													d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8"
												/>
											</svg>
											{$t('admin.systemHealth.sendTestPush')}
										{/if}
									</button>
								{/if}
							</td>
							<td class="px-6 py-4 whitespace-nowrap text-right">
								<span
									class="px-2 py-1 text-xs font-medium rounded-full {health
										.checks.vapid.enabled
										? health.checks.vapid.status === 'healthy'
											? 'bg-success-100 text-success-800'
											: 'bg-danger-100 text-danger-800'
										: 'bg-border-soft text-text-strong'}"
								>
									{getStatusLabel(health.checks.vapid.status)}
								</span>
							</td>
						</tr>

						<!-- TOTP (2FA) -->
						<tr class="hover:bg-surface-1 transition-colors">
							<td class="px-6 py-4 whitespace-nowrap">
								<div class="flex items-center gap-3">
									<div class={getStatusColor(health.checks.totp_encryption)}>
										<svg
											class="w-5 h-5"
											fill="none"
											stroke="currentColor"
											viewBox="0 0 24 24"
										>
											<path
												stroke-linecap="round"
												stroke-linejoin="round"
												stroke-width="2"
												d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"
											/>
										</svg>
									</div>
									<span class="text-sm font-medium text-text"
										>{$t('admin.systemHealth.totp')}</span
									>
								</div>
							</td>
							<td class="px-6 py-4 whitespace-nowrap text-sm text-text-subtle">
								{#if health.checks.totp_encryption.error}
									<span class="text-danger-600"
										>{health.checks.totp_encryption.error}</span
									>
								{:else if health.checks.totp_encryption.enabled}
									{$t('admin.systemHealth.encryptionValid')}
								{:else}
									{$t('admin.systemHealth.totpDisabled')}
								{/if}
							</td>
							<td class="px-6 py-4 whitespace-nowrap text-right"></td>
							<td class="px-6 py-4 whitespace-nowrap text-right">
								<span
									class="px-2 py-1 text-xs font-medium rounded-full {health
										.checks.totp_encryption.enabled
										? health.checks.totp_encryption.status === 'healthy'
											? 'bg-success-100 text-success-800'
											: 'bg-danger-100 text-danger-800'
										: 'bg-border-soft text-text-strong'}"
								>
									{getStatusLabel(health.checks.totp_encryption.status)}
								</span>
							</td>
						</tr>
					</tbody>
				</table>
			</div>
		{:else}
			<!-- Error State -->
			<div class="bg-danger-50 border border-danger-200 rounded-lg p-4">
				<div class="flex items-center gap-3">
					<svg
						class="w-6 h-6 text-danger-600"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
						/>
					</svg>
					<span class="text-danger-800"
						>{$t('admin.systemHealth.loadErrorRetry')}</span
					>
				</div>
			</div>
		{/if}
	{/if}
</div>
