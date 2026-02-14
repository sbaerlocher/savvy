<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { t } from '$lib/stores/i18n';
	import { toastStore } from '$lib/stores/toast';
	import { logger } from '$lib/utils/logger';
	import { adminApi } from '$lib/api';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

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
		if (!check.enabled) return 'text-gray-400';
		return check.status === 'healthy' ? 'text-green-600' : 'text-red-600';
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

	function formatTimestamp(timestamp: string): string {
		const date = new Date(timestamp);
		return date.toLocaleTimeString();
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
	<!-- Header -->
	<div class="mb-8">
		<h1 class="text-3xl font-bold text-gray-900">
			{$t('nav.adminSystemHealth')}
		</h1>
	</div>

	<!-- Controls + Overall Status -->
	<div class="flex flex-col sm:flex-row gap-3 mb-6">
		<!-- Overall Status (left, where search bar is on other pages) -->
		<div
			class="sm:flex-1 flex items-center gap-3 h-[42px] px-4 bg-white border border-gray-300 rounded-md"
		>
			{#if isLoading && !health}
				<span class="text-sm text-gray-500">{$t('common.loading')}</span>
			{:else if health}
				{#if health.status === 'ready'}
					<svg
						class="w-5 h-5 text-green-600 shrink-0"
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
					<span class="text-sm font-medium text-green-600"
						>{$t('admin.systemHealth.ready')}</span
					>
				{:else if health.status === 'degraded'}
					<svg
						class="w-5 h-5 text-yellow-600 shrink-0"
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
					<span class="text-sm font-medium text-yellow-600"
						>{$t('admin.systemHealth.degraded')}</span
					>
				{:else}
					<svg
						class="w-5 h-5 text-red-600 shrink-0"
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
					<span class="text-sm font-medium text-red-600"
						>{$t('admin.systemHealth.notReady')}</span
					>
				{/if}
				{#if lastRefresh}
					<span class="text-xs text-gray-400 ml-auto"
						>{getTimeAgo(lastRefresh)}</span
					>
				{/if}
			{:else}
				<span class="text-sm text-gray-400">—</span>
			{/if}
		</div>

		<!-- Action Buttons (right, full-width on mobile) -->
		<div class="grid grid-cols-3 sm:flex gap-3 w-full sm:w-auto h-[42px]">
			<button
				class="col-span-1 sm:flex-none flex items-center justify-center gap-2 h-full px-4 bg-white border rounded-md hover:bg-gray-50 transition-colors {autoRefresh
					? 'ring-2 ring-cyan-500 border-cyan-500'
					: 'border-gray-300'}"
				onclick={toggleAutoRefresh}
				title={autoRefresh
					? $t('admin.systemHealth.autoRefresh')
					: $t('admin.systemHealth.paused')}
			>
				{#if autoRefresh}
					<svg
						class="w-5 h-5 text-gray-600"
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
						class="w-5 h-5 text-gray-600"
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
		<div class="md:hidden bg-white shadow rounded-lg divide-y divide-gray-200">
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
				<span class="text-sm font-medium text-gray-900 flex-1"
					>{$t('admin.systemHealth.database')}</span
				>
				<span
					class="px-2 py-1 text-xs font-medium rounded-full {health.checks
						.database.enabled
						? health.checks.database.status === 'healthy'
							? 'bg-green-100 text-green-800'
							: 'bg-red-100 text-red-800'
						: 'bg-gray-100 text-gray-800'}"
				>
					{getStatusLabel(health.checks.database.status)}
				</span>
				<svg
					class="w-4 h-4 text-gray-400 transition-transform {expandedService ===
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
				<div class="px-4 py-3 bg-gray-50 text-sm text-gray-600">
					{#if health.checks.database.error}
						<span class="text-red-600">{health.checks.database.error}</span>
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
				<span class="text-sm font-medium text-gray-900 flex-1"
					>{$t('admin.systemHealth.smtp')}</span
				>
				<span
					class="px-2 py-1 text-xs font-medium rounded-full {health.checks.smtp
						.enabled
						? health.checks.smtp.status === 'healthy'
							? 'bg-green-100 text-green-800'
							: 'bg-red-100 text-red-800'
						: 'bg-gray-100 text-gray-800'}"
				>
					{getStatusLabel(health.checks.smtp.status)}
				</span>
				<svg
					class="w-4 h-4 text-gray-400 transition-transform {expandedService ===
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
				<div class="px-4 py-3 bg-gray-50 space-y-3">
					<div class="text-sm text-gray-600">
						{#if health.checks.smtp.error}
							<span class="text-red-600">{health.checks.smtp.error}</span>
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
				<span class="text-sm font-medium text-gray-900 flex-1"
					>{$t('admin.systemHealth.oauth')}</span
				>
				<span
					class="px-2 py-1 text-xs font-medium rounded-full {health.checks.oauth
						.enabled
						? health.checks.oauth.status === 'healthy'
							? 'bg-green-100 text-green-800'
							: 'bg-red-100 text-red-800'
						: 'bg-gray-100 text-gray-800'}"
				>
					{getStatusLabel(health.checks.oauth.status)}
				</span>
				<svg
					class="w-4 h-4 text-gray-400 transition-transform {expandedService ===
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
				<div class="px-4 py-3 bg-gray-50 text-sm text-gray-600">
					{#if health.checks.oauth.error}
						<span class="text-red-600">{health.checks.oauth.error}</span>
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
				<span class="text-sm font-medium text-gray-900 flex-1"
					>{$t('admin.systemHealth.pushNotifications')}</span
				>
				<span
					class="px-2 py-1 text-xs font-medium rounded-full {health.checks.vapid
						.enabled
						? health.checks.vapid.status === 'healthy'
							? 'bg-green-100 text-green-800'
							: 'bg-red-100 text-red-800'
						: 'bg-gray-100 text-gray-800'}"
				>
					{getStatusLabel(health.checks.vapid.status)}
				</span>
				<svg
					class="w-4 h-4 text-gray-400 transition-transform {expandedService ===
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
				<div class="px-4 py-3 bg-gray-50 space-y-3">
					<div class="text-sm text-gray-600">
						{#if health.checks.vapid.error}
							<span class="text-red-600">{health.checks.vapid.error}</span>
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
				<span class="text-sm font-medium text-gray-900 flex-1"
					>{$t('admin.systemHealth.totp')}</span
				>
				<span
					class="px-2 py-1 text-xs font-medium rounded-full {health.checks
						.totp_encryption.enabled
						? health.checks.totp_encryption.status === 'healthy'
							? 'bg-green-100 text-green-800'
							: 'bg-red-100 text-red-800'
						: 'bg-gray-100 text-gray-800'}"
				>
					{getStatusLabel(health.checks.totp_encryption.status)}
				</span>
				<svg
					class="w-4 h-4 text-gray-400 transition-transform {expandedService ===
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
				<div class="px-4 py-3 bg-gray-50 text-sm text-gray-600">
					{#if health.checks.totp_encryption.error}
						<span class="text-red-600"
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
			<table class="min-w-full divide-y divide-gray-200">
				<thead class="bg-gray-50">
					<tr>
						<th
							class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
						>
							{$t('admin.systemHealth.service')}
						</th>
						<th
							class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
						>
							{$t('admin.systemHealth.details')}
						</th>
						<th
							class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider"
						>
							{$t('admin.users.actions')}
						</th>
						<th
							class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider"
						>
							{$t('admin.systemHealth.statusLabel')}
						</th>
					</tr>
				</thead>
				<tbody class="bg-white divide-y divide-gray-200">
					<!-- Database -->
					<tr class="hover:bg-gray-50 transition-colors">
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
								<span class="text-sm font-medium text-gray-900"
									>{$t('admin.systemHealth.database')}</span
								>
							</div>
						</td>
						<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
							{#if health.checks.database.error}
								<span class="text-red-600">{health.checks.database.error}</span>
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
								class="px-2 py-1 text-xs font-medium rounded-full {health.checks
									.database.enabled
									? health.checks.database.status === 'healthy'
										? 'bg-green-100 text-green-800'
										: 'bg-red-100 text-red-800'
									: 'bg-gray-100 text-gray-800'}"
							>
								{getStatusLabel(health.checks.database.status)}
							</span>
						</td>
					</tr>

					<!-- SMTP -->
					<tr class="hover:bg-gray-50 transition-colors">
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
								<span class="text-sm font-medium text-gray-900"
									>{$t('admin.systemHealth.smtp')}</span
								>
							</div>
						</td>
						<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
							{#if health.checks.smtp.error}
								<span class="text-red-600">{health.checks.smtp.error}</span>
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
								class="px-2 py-1 text-xs font-medium rounded-full {health.checks
									.smtp.enabled
									? health.checks.smtp.status === 'healthy'
										? 'bg-green-100 text-green-800'
										: 'bg-red-100 text-red-800'
									: 'bg-gray-100 text-gray-800'}"
							>
								{getStatusLabel(health.checks.smtp.status)}
							</span>
						</td>
					</tr>

					<!-- OAuth -->
					<tr class="hover:bg-gray-50 transition-colors">
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
								<span class="text-sm font-medium text-gray-900"
									>{$t('admin.systemHealth.oauth')}</span
								>
							</div>
						</td>
						<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
							{#if health.checks.oauth.error}
								<span class="text-red-600">{health.checks.oauth.error}</span>
							{:else if health.checks.oauth.enabled}
								{$t('admin.systemHealth.oauthConnectivity')}
							{:else}
								{$t('admin.systemHealth.oauthDisabled')}
							{/if}
						</td>
						<td class="px-6 py-4 whitespace-nowrap text-right"></td>
						<td class="px-6 py-4 whitespace-nowrap text-right">
							<span
								class="px-2 py-1 text-xs font-medium rounded-full {health.checks
									.oauth.enabled
									? health.checks.oauth.status === 'healthy'
										? 'bg-green-100 text-green-800'
										: 'bg-red-100 text-red-800'
									: 'bg-gray-100 text-gray-800'}"
							>
								{getStatusLabel(health.checks.oauth.status)}
							</span>
						</td>
					</tr>

					<!-- VAPID (Push) -->
					<tr class="hover:bg-gray-50 transition-colors">
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
								<span class="text-sm font-medium text-gray-900"
									>{$t('admin.systemHealth.pushNotifications')}</span
								>
							</div>
						</td>
						<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
							{#if health.checks.vapid.error}
								<span class="text-red-600">{health.checks.vapid.error}</span>
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
								class="px-2 py-1 text-xs font-medium rounded-full {health.checks
									.vapid.enabled
									? health.checks.vapid.status === 'healthy'
										? 'bg-green-100 text-green-800'
										: 'bg-red-100 text-red-800'
									: 'bg-gray-100 text-gray-800'}"
							>
								{getStatusLabel(health.checks.vapid.status)}
							</span>
						</td>
					</tr>

					<!-- TOTP (2FA) -->
					<tr class="hover:bg-gray-50 transition-colors">
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
								<span class="text-sm font-medium text-gray-900"
									>{$t('admin.systemHealth.totp')}</span
								>
							</div>
						</td>
						<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
							{#if health.checks.totp_encryption.error}
								<span class="text-red-600"
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
								class="px-2 py-1 text-xs font-medium rounded-full {health.checks
									.totp_encryption.enabled
									? health.checks.totp_encryption.status === 'healthy'
										? 'bg-green-100 text-green-800'
										: 'bg-red-100 text-red-800'
									: 'bg-gray-100 text-gray-800'}"
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
		<div class="bg-red-50 border border-red-200 rounded-lg p-4">
			<div class="flex items-center gap-3">
				<svg
					class="w-6 h-6 text-red-600"
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
				<span class="text-red-800"
					>{$t('admin.systemHealth.loadErrorRetry')}</span
				>
			</div>
		</div>
	{/if}
</div>
