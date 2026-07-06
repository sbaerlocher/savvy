<script lang="ts">
	import { t } from '$lib/stores/i18n';
	import { authStore } from '$lib/stores/auth';
	import { resolve } from '$app/paths';
	import type { DashboardResponse } from '$lib/types/api';

	let {
		stats,
		isRefreshing
	}: {
		stats: DashboardResponse['stats'];
		isRefreshing: boolean;
	} = $props();
</script>

<div class="mb-8">
	<div class="flex items-center gap-3">
		<h1 class="text-3xl font-bold text-gray-900">
			{$t('dashboard.welcome', { name: $authStore.user?.first_name || '' })}
		</h1>
		{#if isRefreshing}
			<span class="text-xs text-gray-400 animate-pulse"
				>{$t('common.refreshing')}</span
			>
		{/if}
	</div>
	<div
		class="mt-2 hidden lg:flex flex-wrap items-center gap-x-4 gap-y-1 text-sm text-gray-600"
	>
		<a
			href={resolve('/cards')}
			data-testid="dashboard-stat-cards"
			class="hover:text-cyan-600 transition"
		>
			<span class="font-semibold text-gray-900">{stats.cards_count}</span>
			{$t('dashboard.stats.cards')}
		</a>
		<span class="text-gray-300">|</span>
		<a
			href={resolve('/vouchers')}
			data-testid="dashboard-stat-vouchers"
			class="hover:text-green-600 transition"
		>
			<span class="font-semibold text-gray-900">{stats.vouchers_count}</span>
			{$t('dashboard.stats.vouchers')}
		</a>
		<span class="text-gray-300">|</span>
		<a
			href={resolve('/gift-cards')}
			data-testid="dashboard-stat-gift-cards"
			class="hover:text-red-600 transition"
		>
			<span class="font-semibold text-gray-900">{stats.gift_cards_count}</span>
			{$t('dashboard.stats.giftCards')}
		</a>
		{#if stats.total_balance > 0}
			<span class="text-gray-300">|</span>
			<span>
				<span class="font-semibold text-gray-900"
					>CHF {Math.round(stats.total_balance)}</span
				>
				{$t('dashboard.stats.totalBalance')}
			</span>
		{/if}
	</div>
</div>
