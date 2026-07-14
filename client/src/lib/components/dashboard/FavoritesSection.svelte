<script lang="ts">
	import { t, locale } from '$lib/stores/i18n';
	import { authStore } from '$lib/stores/auth';
	import { isOnline } from '$lib/stores/offline';
	import { resolve } from '$app/paths';
	import ResourceTile from '$lib/components/ui/ResourceTile.svelte';
	import {
		cardToTileModel,
		voucherToTileModel,
		giftCardToTileModel
	} from '$lib/utils/tile-model';
	import { isVoucherValid, isGiftCardActive } from '$lib/utils/resource-status';
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

	// Dashboard is the checkout quick-access, so only surface usable favorites:
	// hide expired/inactive vouchers and expired/depleted gift cards (matching
	// the default status filter in WalletView). Cards have no expiry/balance.
	const cardTiles = $derived(
		data.recent_cards.map((c) =>
			cardToTileModel(c, currentUserId, currentLocale)
		)
	);
	const voucherTiles = $derived(
		data.recent_vouchers
			.filter(isVoucherValid)
			.map((v) => voucherToTileModel(v, currentUserId, currentLocale))
	);
	const giftCardTiles = $derived(
		data.recent_gift_cards
			.filter(isGiftCardActive)
			.map((g) => giftCardToTileModel(g, currentUserId, currentLocale))
	);

	const isEmpty = $derived(
		cardTiles.length === 0 &&
			voucherTiles.length === 0 &&
			giftCardTiles.length === 0
	);
</script>

<div data-testid="favorites-section">
	<div data-testid="favorites-list">
		{#if isEmpty}
			<div class="rounded-xl border border-border/80 bg-white py-8 text-center">
				<p class="text-sm text-text-subtle">{$t('dashboard.noActivity')}</p>
				<p class="mt-1 mb-4 text-xs text-text-faint">
					{$t('dashboard.noActivityHint')}
				</p>
				<a
					href={resolve('/cards/new')}
					onclick={(e) => {
						if (!$isOnline) e.preventDefault();
					}}
					class="inline-flex items-center gap-1 text-sm font-medium text-accent transition hover:text-accent-800 {!$isOnline
						? 'pointer-events-none opacity-50'
						: ''}"
				>
					{$t('dashboard.getStarted')} →
				</a>
			</div>
		{:else}
			<div class="grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-3">
				{#each cardTiles as model (model.id)}
					<ResourceTile {model} showBarcode compact {onShowBarcode} />
				{/each}
				{#each voucherTiles as model (model.id)}
					<ResourceTile {model} showBarcode compact {onShowBarcode} />
				{/each}
				{#each giftCardTiles as model (model.id)}
					<ResourceTile {model} showBarcode compact {onShowBarcode} />
				{/each}
			</div>
		{/if}
	</div>
</div>
