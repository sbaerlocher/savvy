<script lang="ts">
	import { onMount } from 'svelte';
	import { authStore } from '$lib/stores/auth';
	import { t } from '$lib/stores/i18n';
	import { dashboardApi } from '$lib/api';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { logger } from '$lib/utils/logger';
	import { platform } from '$lib/utils/platform';
	import Skeleton from '$lib/components/ui/Skeleton.svelte';
	import SectionLabel from '$lib/components/ui/SectionLabel.svelte';
	import FavoritesSection from '$lib/components/dashboard/FavoritesSection.svelte';
	import BarcodeModal, {
		type BarcodeModalItem
	} from '$lib/components/dashboard/BarcodeModal.svelte';
	import type { DashboardResponse } from '$lib/types/api';

	// Bound up to the page: the refresh hint renders in the PageShell title row
	// (eyebrowAside/actions) and the stats feed the title-row tiles on desktop,
	// while the loading itself lives here.
	let {
		// eslint-disable-next-line no-useless-assignment -- write-only bindable, read by the page
		isRefreshing = $bindable(false),
		// eslint-disable-next-line no-useless-assignment -- write-only bindable, read by the page
		stats = $bindable(null)
	}: {
		isRefreshing?: boolean;
		stats?: DashboardResponse['stats'] | null;
	} = $props();

	const IS_DESKTOP = platform === 'other';

	$effect(() => {
		stats = data?.stats ?? null;
	});

	const pageLogger = logger.child('DashboardPage');

	let data = $state<DashboardResponse | null>(null);
	let isLoading = $state(true);
	let error = $state<string | null>(null);

	let barcodeModalItem = $state<BarcodeModalItem | null>(null);

	// Stat tile chrome per platform (dashboard mockups): iOS = bordered card,
	// Android = borderless M3 surface with larger radius, desktop = bordered
	// card with shadow.
	const statTileClass =
		platform === 'android'
			? 'rounded-[var(--radius-m3-lg)] bg-m3-card px-4.5 py-4'
			: platform === 'ios'
				? 'rounded-xl border border-border bg-white px-4 py-3'
				: 'rounded-xl border border-border bg-white px-5 py-4 shadow-card lg:min-w-36';
	// Stat value: iOS mockup uses the 23px mono step, others 24px semibold.
	const statValueClass =
		platform === 'ios'
			? 'font-mono text-stat tabular-nums text-text-strong'
			: 'font-mono text-2xl font-semibold tabular-nums text-text-strong';
	// Stat label: Android mockup uses 12.5px (--text-body-sm), iOS and desktop
	// 13px regular (--text-label carries a 600 companion weight, hence the
	// override).
	const statLabelClass =
		platform === 'android'
			? 'mt-1 text-body-sm text-text-subtle'
			: platform === 'ios'
				? 'mt-1 text-label font-normal text-text-subtle'
				: 'mt-1.5 text-label font-normal text-text-subtle';
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

{#if error}
	<div class="rounded-xl border border-border/80 bg-white p-6 text-center">
		<p class="mb-4 text-danger-600">{error}</p>
		<button onclick={loadDashboard} class="btn btn-primary"
			>{$t('common.retry')}</button
		>
	</div>
{:else if isLoading}
	<!-- Block skeleton mirroring the loaded layout: content first, then the
	     stats grid, matching the mobile order below.
	     Replaces the generic logo LoadingSpinner. -->
	<Skeleton class="h-40 w-full" />
	{#if !IS_DESKTOP}
		<div class="mt-6 grid grid-cols-2 gap-3">
			<Skeleton class="h-20 lg:min-w-36" />
			<Skeleton class="h-20 lg:min-w-36" />
		</div>
	{/if}
{:else if data}
	<!-- At-checkout favorites: barcode always visible (register quick access).
	     The kicker is native-only — on desktop the content area starts flush
	     under the title row, like every other page. -->
	<section>
		{#if !IS_DESKTOP}
			<!-- Android: same type step as the wallet's filter chips. -->
			<SectionLabel size={platform === 'android' ? 'label' : 'eyebrow'}>
				{$t('dashboard.atCheckout')}
			</SectionLabel>
		{/if}
		<FavoritesSection
			{data}
			onShowBarcode={(item) => (barcodeModalItem = item)}
		/>
	</section>
	<!-- Stats: after favorites on the phones; on desktop they live in the
	     title row (page-level tiles fed through the bound `stats`). -->
	{#if !IS_DESKTOP}
		<div class="mt-6">
			<div class="grid grid-cols-2 gap-3">
				<div data-testid="dashboard-stat-balance" class={statTileClass}>
					<p class={statValueClass}>
						<!-- `?? 0` only narrows the type: this renders inside the
					     `{:else if data}` branch, but TS cannot always see that. -->
						CHF {Math.round(data?.stats.total_balance ?? 0)}
					</p>
					<p class={statLabelClass}>
						{$t('dashboard.totalBalanceShort')}
					</p>
				</div>
				<div data-testid="dashboard-stat-entries" class={statTileClass}>
					<p class={statValueClass}>
						{entriesCount}
					</p>
					<p class={statLabelClass}>{$t('dashboard.entries')}</p>
				</div>
			</div>
		</div>
	{/if}
{/if}

<BarcodeModal
	item={barcodeModalItem}
	onClose={() => (barcodeModalItem = null)}
/>
