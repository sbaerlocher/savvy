<script lang="ts">
	import { t, locale } from '$lib/stores/i18n';
	import { authStore } from '$lib/stores/auth';
	import { resolve } from '$app/paths';
	import ResourceTile from '$lib/components/ui/ResourceTile.svelte';
	import {
		cardToTileModel,
		voucherToTileModel,
		giftCardToTileModel
	} from '$lib/utils/tile-model';
	import type { DashboardResponse } from '$lib/types/api';
	import type { BarcodeModalItem } from './BarcodeModal.svelte';

	let {
		data,
		onShowBarcode
	}: {
		data: DashboardResponse;
		onShowBarcode: (item: BarcodeModalItem) => void;
	} = $props();

	const currentUserId = $derived($authStore.user?.id);
	const currentLocale = $derived($locale || 'de');

	const cardTiles = $derived(
		data.recent_cards
			.slice(0, 3)
			.map((c) => cardToTileModel(c, currentUserId, currentLocale))
	);
	const voucherTiles = $derived(
		data.recent_vouchers
			.slice(0, 3)
			.map((v) => voucherToTileModel(v, currentUserId, currentLocale))
	);
	const giftCardTiles = $derived(
		data.recent_gift_cards
			.slice(0, 3)
			.map((g) => giftCardToTileModel(g, currentUserId, currentLocale))
	);
</script>

<div class="bg-white rounded-lg shadow-md p-6" data-testid="favorites-section">
	<h2 class="text-xl font-semibold text-gray-900 mb-4">
		{#if data.has_favorites}
			{$t('dashboard.favorites')}
		{:else}
			{$t('dashboard.recentlyAdded')}
		{/if}
	</h2>
	<div data-testid="favorites-list">
		{#if data.recent_cards.length === 0 && data.recent_vouchers.length === 0 && data.recent_gift_cards.length === 0}
			<div class="text-center py-8">
				<p class="text-gray-500 text-sm">
					{$t('dashboard.noActivity')}
				</p>
				<p class="text-gray-400 text-xs mt-1 mb-4">
					{$t('dashboard.noActivityHint')}
				</p>
				<a
					href={resolve('/cards/new')}
					class="inline-flex items-center gap-1 text-sm font-medium text-cyan-600 hover:text-cyan-800 transition"
				>
					{$t('dashboard.getStarted')} →
				</a>
			</div>
		{:else}
			<div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
				<!-- Cards -->
				{#if data.recent_cards.length > 0 && (!data.has_favorites || data.has_card_favorites)}
					{#each cardTiles as model (model.id)}
						<ResourceTile {model} showBarcode compact {onShowBarcode} />
					{/each}
				{/if}

				<!-- Vouchers -->
				{#if data.recent_vouchers.length > 0 && (!data.has_favorites || data.has_voucher_favorites)}
					{#each voucherTiles as model (model.id)}
						<ResourceTile {model} showBarcode compact {onShowBarcode} />
					{/each}
				{/if}

				<!-- Gift Cards -->
				{#if data.recent_gift_cards.length > 0 && (!data.has_favorites || data.has_gift_card_favorites)}
					{#each giftCardTiles as model (model.id)}
						<ResourceTile {model} showBarcode compact {onShowBarcode} />
					{/each}
				{/if}
			</div>
		{/if}
	</div>
</div>
