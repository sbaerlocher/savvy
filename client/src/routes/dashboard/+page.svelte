<script lang="ts">
	import { authStore } from '$lib/stores/auth';
	import { t } from '$lib/stores/i18n';
	import { platform } from '$lib/utils/platform';
	import PageShell from '$lib/components/layout/PageShell.svelte';
	import Section from './Section.svelte';
	import type { DashboardResponse } from '$lib/types/api';

	const IS_DESKTOP = platform === 'other';

	const firstName = $derived($authStore.user?.first_name || '');
	// Refresh hint next to the eyebrow: iOS mockup uses the 11px eyebrow step,
	// Android 12px.
	const refreshHintClass =
		platform === 'ios' ? 'text-eyebrow font-medium' : 'text-xs font-medium';

	// Bound from the Section: the hint sits in the title row, the loading
	// lives in the Section.
	let isRefreshing = $state(false);

	// Stat tiles beside the title row (desktop mockup), fed by the section's
	// dashboard load through the binding below.
	let desktopStats = $state<DashboardResponse['stats'] | null>(null);
	const desktopEntries = $derived(
		desktopStats
			? desktopStats.cards_count +
					desktopStats.vouchers_count +
					desktopStats.gift_cards_count
			: 0
	);

	// Desktop stat tile chrome: the shared title-action look (42px, rounded-lg,
	// field border on white) without its hover — the tiles are read-only. The
	// taller two-line tiles live in the section with the native layout.
	const STAT_TILE =
		'control flex items-center gap-2.5 rounded-lg border border-border-field bg-white px-4';
	const STAT_VALUE =
		'font-mono text-lg font-semibold tabular-nums text-text-strong';
	const STAT_LABEL = 'text-label font-normal text-text-subtle';
</script>

<svelte:head>
	<title>{$t('dashboard.title')} - {$t('common.appName')}</title>
</svelte:head>

<!-- Stats beside the title row (desktop mockup). Native renders its stats
     inside the section, after the favourites. -->
{#snippet desktopStatTiles()}
	<!-- Rendered empty until the section's load delivers the stats: the aside
	     prop must stay CONSTANT — flipping it between undefined and a snippet
	     makes PageShell switch layout branches, which remounts the section and
	     loops its load. -->
	{#if desktopStats}
		<div class="flex items-center gap-2.5">
			<div data-testid="dashboard-stat-balance" class={STAT_TILE}>
				<span class={STAT_VALUE}>
					CHF {Math.round(desktopStats?.total_balance ?? 0)}
				</span>
				<span class={STAT_LABEL}>{$t('dashboard.totalBalanceShort')}</span>
			</div>
			<div data-testid="dashboard-stat-entries" class={STAT_TILE}>
				<span class={STAT_VALUE}>{desktopEntries}</span>
				<span class={STAT_LABEL}>{$t('dashboard.entries')}</span>
			</div>
		</div>
	{/if}
{/snippet}

<PageShell
	eyebrow={$t('dashboard.greeting', { name: firstName })}
	title={$t('dashboard.yourFavorites')}
	aside={IS_DESKTOP ? desktopStatTiles : undefined}
>
	<!-- Both native mockups put the refresh indicator inline next to the
	     greeting eyebrow; desktop keeps it in the actions slot. -->
	{#snippet eyebrowAside()}
		{#if isRefreshing && platform !== 'other'}
			<span class="animate-pulse {refreshHintClass} text-text-faint"
				>{$t('common.refreshing')}</span
			>
		{/if}
	{/snippet}
	{#snippet actions()}
		{#if isRefreshing && platform === 'other'}
			<span class="animate-pulse text-xs text-text-faint"
				>{$t('common.refreshing')}</span
			>
		{/if}
	{/snippet}
	<Section bind:isRefreshing bind:stats={desktopStats} />
</PageShell>
