<script lang="ts">
	import { onMount } from 'svelte';
	import { authStore } from '$lib/stores/auth';
	import { isOnline } from '$lib/stores/offline';
	import { t } from '$lib/stores/i18n';
	import { dashboardApi } from '$lib/api';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { logger } from '$lib/utils/logger';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import DashboardHeader from '$lib/components/dashboard/DashboardHeader.svelte';
	import FavoritesSection from '$lib/components/dashboard/FavoritesSection.svelte';
	import QuickActions from '$lib/components/dashboard/QuickActions.svelte';
	import DashboardTips from '$lib/components/dashboard/DashboardTips.svelte';
	import BarcodeModal, {
		type BarcodeModalItem
	} from '$lib/components/dashboard/BarcodeModal.svelte';
	import type { DashboardResponse } from '$lib/types/api';

	const pageLogger = logger.child('DashboardPage');

	let data = $state<DashboardResponse | null>(null);
	let isLoading = $state(true);
	let isRefreshing = $state(false);
	let error = $state<string | null>(null);

	let barcodeModalItem = $state<BarcodeModalItem | null>(null);

	onMount(async () => {
		if (!$authStore.isAuthenticated) {
			goto(resolve('/login'));
			return;
		}

		await loadDashboard();
	});

	async function loadDashboard() {
		isLoading = true;
		error = null;
		try {
			// Phase 1: Show cached data immediately
			const cached = await dashboardApi.getCached();
			if (cached) {
				data = cached;
				isLoading = false;
				isRefreshing = true;
			}

			// Phase 2: Fetch fresh data from network
			if (navigator.onLine) {
				try {
					data = await dashboardApi.get();
				} catch (err) {
					if (!cached) {
						error = $t('common.error');
						pageLogger.error('Failed to load dashboard', { error: err });
					}
				}
			} else if (!cached) {
				error = $t('common.error');
			}
		} catch (err) {
			error = $t('common.error');
			pageLogger.error('Failed to load dashboard', { error: err });
		} finally {
			isLoading = false;
			isRefreshing = false;
		}
	}

	const isOffline = $derived(!$isOnline);
</script>

<svelte:head>
	<title>{$t('dashboard.title')} - {$t('common.appName')}</title>
</svelte:head>

<div class="px-4 max-w-7xl mx-auto">
	{#if isLoading}
		<LoadingSpinner />
	{:else if error}
		<div class="bg-white rounded-lg shadow-md p-6 text-center">
			<p class="text-red-600 mb-4">{error}</p>
			<button onclick={loadDashboard} class="btn btn-primary"
				>{$t('common.retry')}</button
			>
		</div>
	{:else if data}
		<DashboardHeader stats={data.stats} {isRefreshing} />

		<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
			<!-- Left column: Favorites (2/3 width) -->
			<div class="lg:col-span-2 space-y-6">
				<FavoritesSection
					{data}
					onShowBarcode={(item) => (barcodeModalItem = item)}
				/>
			</div>

			<!-- Right column: Quick Actions + Tips (1/3 width) -->
			<div class="lg:col-span-1 space-y-6">
				<QuickActions {isOffline} />
				<DashboardTips />
			</div>
		</div>
	{/if}
</div>

<BarcodeModal
	item={barcodeModalItem}
	onClose={() => (barcodeModalItem = null)}
/>
