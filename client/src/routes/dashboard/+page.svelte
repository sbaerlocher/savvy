<script lang="ts">
	import { onMount } from 'svelte';
	import { authStore } from '$lib/stores/auth';
	import { t } from '$lib/stores/i18n';
	import { dashboardApi } from '$lib/api';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { logger } from '$lib/utils/logger';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import FavoritesSection from '$lib/components/dashboard/FavoritesSection.svelte';
	import type { DashboardResponse } from '$lib/types/api';

	const pageLogger = logger.child('DashboardPage');

	let data = $state<DashboardResponse | null>(null);
	let isLoading = $state(true);
	let isRefreshing = $state(false);
	let error = $state<string | null>(null);

	const firstName = $derived($authStore.user?.first_name || '');
	// Total entries across all resource types (drives the "Einträge" stat).
	const entriesCount = $derived(
		data
			? data.stats.cards_count +
					data.stats.vouchers_count +
					data.stats.gift_cards_count
			: 0
	);

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
</script>

<svelte:head>
	<title>{$t('dashboard.title')} - {$t('common.appName')}</title>
</svelte:head>

<div class="mx-auto max-w-7xl px-4">
	{#if isLoading}
		<LoadingSpinner />
	{:else if error}
		<div class="rounded-xl border border-border/80 bg-white p-6 text-center">
			<p class="mb-4 text-red-600">{error}</p>
			<button onclick={loadDashboard} class="btn btn-primary"
				>{$t('common.retry')}</button
			>
		</div>
	{:else if data}
		<!-- Grid gives the two layouts from the same markup:
		     mobile (1 col) stacks via `order` — header → favorites → stats;
		     desktop (2 cols) puts the header top-left and stats top-right,
		     favorites spanning the full width below. -->
		<div class="grid grid-cols-1 gap-x-4 lg:grid-cols-[1fr_auto]">
			<div class="order-1 lg:col-start-1 lg:row-start-1">
				<PageHeader
					eyebrow={$t('dashboard.greeting', { name: firstName })}
					title={$t('dashboard.yourFavorites')}
				>
					{#snippet actions()}
						{#if isRefreshing}
							<span class="animate-pulse text-xs text-text-faint"
								>{$t('common.refreshing')}</span
							>
						{/if}
					{/snippet}
				</PageHeader>
			</div>

			<!-- Stats: after favorites on mobile (order-3), top-right on desktop. -->
			<div
				class="order-3 mt-6 grid grid-cols-2 gap-3 lg:order-none lg:col-start-2 lg:row-start-1 lg:mt-0"
			>
				<div
					data-testid="dashboard-stat-balance"
					class="rounded-xl border border-border/80 bg-white px-4 py-3 lg:min-w-32"
				>
					<p class="text-2xl font-bold tabular-nums text-text">
						CHF {Math.round(data.stats.total_balance)}
					</p>
					<p class="text-sm text-text-subtle">
						{$t('dashboard.totalBalanceShort')}
					</p>
				</div>
				<div
					data-testid="dashboard-stat-entries"
					class="rounded-xl border border-border/80 bg-white px-4 py-3 lg:min-w-32"
				>
					<p class="text-2xl font-bold tabular-nums text-text">
						{entriesCount}
					</p>
					<p class="text-sm text-text-subtle">{$t('dashboard.entries')}</p>
				</div>
			</div>

			<!-- At-checkout favorites: barcode always visible (register quick access) -->
			<section class="order-2 lg:col-span-2 lg:row-start-2">
				<h2
					class="mb-3 text-xs font-semibold uppercase tracking-wider text-text-subtle"
				>
					{$t('dashboard.atCheckout')}
				</h2>
				<FavoritesSection {data} />
			</section>
		</div>
	{/if}
</div>
