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
	import {
		isCardActive,
		isVoucherValid,
		isGiftCardActive
	} from '$lib/utils/resource-status';
	import { platform } from '$lib/utils/platform';

	// Desktop uses the wallet's tile chrome 1:1 (same padding, same barcode
	// height); the compact variant stays a native-dashboard treatment.
	const IS_DESKTOP = platform === 'other';
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
	// hide non-active cards (inactive/expired/lost/blocked), expired/inactive
	// vouchers and expired/depleted gift cards (matching the default status
	// filter in WalletView).
	const cardTiles = $derived(
		data.recent_cards
			.filter(isCardActive)
			.map((c) => cardToTileModel(c, currentUserId, currentLocale))
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

	// Empty-state chrome per platform (dashboard mockups): iOS = bordered card,
	// Android = borderless M3 surface with large radius, desktop = bordered
	// card, roomier.
	const emptyStateClass =
		platform === 'android'
			? 'rounded-[var(--radius-m3-lg)] bg-m3-card px-6 py-8.5'
			: platform === 'ios'
				? 'rounded-xl border border-border bg-white px-card py-8.5'
				: 'rounded-2xl border border-border bg-white px-6 py-11';
	// Empty-state type per platform (Android mockup: 15px title, 12.5px hint,
	// accent-700 CTA; iOS mockup: 14px title, 12.5px hint, accent CTA; desktop
	// mockup: 16px title — nearest token is --text-subheading at 15.5px — and a
	// 13px hint).
	const emptyTitleClass =
		platform === 'android' || platform === 'other'
			? 'text-subheading text-text'
			: 'text-sm font-semibold text-text';
	const emptyHintClass =
		platform === 'android'
			? 'mt-1 mb-4 text-body-sm text-text-faint'
			: platform === 'ios'
				? 'mt-1.5 mb-4 text-body-sm text-text-faint'
				: 'mt-1.5 mb-4 text-label text-text-faint';
	const emptyCtaClass =
		platform === 'android'
			? 'text-accent-700 hover:text-accent-800'
			: 'text-accent hover:text-accent-800';
</script>

<div data-testid="favorites-section">
	<div data-testid="favorites-list">
		{#if isEmpty}
			<div class="{emptyStateClass} text-center">
				<p class={emptyTitleClass}>
					{$t('dashboard.noActivity')}
				</p>
				<p class={emptyHintClass}>
					{$t('dashboard.noActivityHint')}
				</p>
				<a
					href={resolve('/cards/new')}
					onclick={(e) => {
						if (!$isOnline) e.preventDefault();
					}}
					class="inline-flex items-center gap-1 text-sm font-semibold transition {emptyCtaClass} {!$isOnline
						? 'pointer-events-none opacity-50'
						: ''}"
				>
					{$t('dashboard.getStarted')} →
				</a>
			</div>
		{:else}
			<!-- Desktop mockup: 18px grid gap; mobile keeps the tighter 12px. -->
			<div
				class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:gap-4.5 xl:grid-cols-3"
			>
				{#each cardTiles as model (model.id)}
					<ResourceTile
						{model}
						showBarcode
						compact={!IS_DESKTOP}
						{onShowBarcode}
					/>
				{/each}
				{#each voucherTiles as model (model.id)}
					<ResourceTile
						{model}
						showBarcode
						compact={!IS_DESKTOP}
						{onShowBarcode}
					/>
				{/each}
				{#each giftCardTiles as model (model.id)}
					<ResourceTile
						{model}
						showBarcode
						compact={!IS_DESKTOP}
						{onShowBarcode}
					/>
				{/each}
			</div>
		{/if}
	</div>
</div>
