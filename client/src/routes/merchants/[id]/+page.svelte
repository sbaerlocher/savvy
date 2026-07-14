<script lang="ts">
	import { cardsApi, giftCardsApi, merchantsApi, vouchersApi } from '$lib/api';
	import WalletView, {
		type WalletFilterState
	} from '$lib/components/ui/WalletView.svelte';
	import { authStore } from '$lib/stores/auth';
	import { locale, t } from '$lib/stores/i18n';
	import { isOnline } from '$lib/stores/offline';
	import { toastStore } from '$lib/stores/toast';
	import type {
		CardDTO,
		GiftCardDTO,
		MerchantDTO,
		VoucherDTO
	} from '$lib/types/api';
	import { extractMerchantFromItems } from '$lib/utils/merchant-aggregator';
	import { page } from '$app/stores';
	import { resolve } from '$app/paths';
	import { onMount } from 'svelte';
	import { get } from 'svelte/store';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);

	const isAdmin = $derived($authStore.user?.is_admin || false);
	const currentUserId = $derived($authStore.user?.id);
	const currentLocale = $derived($locale || 'de-DE');
	const isOffline = $derived(!$isOnline);

	let merchant = $state<MerchantDTO | null>(null);
	let cards = $state<CardDTO[]>([]);
	let vouchers = $state<VoucherDTO[]>([]);
	let giftCards = $state<GiftCardDTO[]>([]);
	let isLoading = $state(true);

	// Merchant-detail keeps its own filter scope (must NOT share the wallet's
	// module-level walletFilters). Inline $state, not a factory — a $state-returning
	// factory in a .svelte.ts module breaks the SSR build.
	const filters = $state<WalletFilterState>({
		searchInput: '',
		typeFilter: 'all',
		statusFilter: 'active',
		sortBy: 'newest',
		ownerFilter: 'all',
		favoritesOnly: false,
		expiringFilter: 'all',
		scrollY: 0
	});

	const enabledTypes = { cards: true, vouchers: true, giftCards: true };

	const merchantId = $derived(get(page).params.id ?? '');

	async function loadData() {
		isLoading = true;
		try {
			const [merchantRes, cardsRes, vouchersRes, giftCardsRes] =
				await Promise.all([
					merchantsApi.get(merchantId).catch(() => null),
					cardsApi.list().catch(() => ({ cards: [] })),
					vouchersApi.list().catch(() => ({ vouchers: [] })),
					giftCardsApi.list().catch(() => ({ gift_cards: [] }))
				]);

			// Filter by merchant_id
			cards = cardsRes.cards.filter((c) => c.merchant_id === merchantId);
			vouchers = vouchersRes.vouchers.filter(
				(v) => v.merchant_id === merchantId
			);
			giftCards = giftCardsRes.gift_cards.filter(
				(g) => g.merchant_id === merchantId
			);

			// Use merchant from API, or extract from cached items as fallback
			merchant =
				merchantRes?.merchant ??
				extractMerchantFromItems(merchantId, cards, vouchers, giftCards);
		} catch {
			toastStore.error(tr('merchantOverview.detail.loadError'));
		} finally {
			isLoading = false;
		}
	}

	onMount(() => {
		loadData();
	});
</script>

<svelte:head>
	<title
		>{merchant?.name ?? tr('merchantOverview.detail.title')} - {tr(
			'common.appName'
		)}</title
	>
</svelte:head>

{#if isLoading}
	<div class="px-4 pb-20 md:pb-4">
		<LoadingSpinner />
	</div>
{:else if merchant}
	<WalletView
		{filters}
		{cards}
		{vouchers}
		{giftCards}
		{enabledTypes}
		{currentUserId}
		{currentLocale}
		{isOffline}
		onReload={loadData}
		barcodeStorageKey="savvy_merchant_detail_show_barcodes"
		idPrefix="detail"
		matchMerchantName={false}
		typeFilterPlacement="batch-header"
		barcodeButtonVariant="icon"
		maxWidth={false}
	>
		{#snippet header()}
			<!-- Header -->
			<div class="mb-8">
				<div class="flex items-center justify-between">
					<div class="flex items-center gap-3">
						<div
							class="w-2 h-8 rounded-full"
							style="background-color: {merchant?.color}"
						></div>
						<h1 class="text-3xl font-bold text-text">{merchant?.name}</h1>
					</div>
					{#if isAdmin && merchant}
						<a
							href={resolve(`/admin/merchants/${merchant.id}/edit`)}
							class="btn btn-xs btn-gray whitespace-nowrap flex items-center gap-1.5"
						>
							{tr('common.edit')}
						</a>
					{/if}
				</div>
			</div>
		{/snippet}

		{#snippet emptyIcon()}
			<svg
				class="w-16 h-16 mx-auto text-text-placeholder mb-4"
				fill="none"
				stroke="currentColor"
				viewBox="0 0 24 24"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="1.5"
					d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4"
				/>
			</svg>
		{/snippet}

		{#snippet searchField()}
			<!-- Merchant detail: always-visible plain search input. -->
			<div class="mb-6">
				<input
					type="search"
					bind:value={filters.searchInput}
					placeholder={tr('common.search')}
					class="w-full px-4 py-2 bg-white border border-border-field rounded-md focus:ring-accent focus:border-accent"
				/>
			</div>
		{/snippet}
	</WalletView>
{/if}
