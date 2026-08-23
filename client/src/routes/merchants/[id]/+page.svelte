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
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { onMount } from 'svelte';
	import { get } from 'svelte/store';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import { ICON_PENCIL, ICON_SEARCH } from '$lib/icons';
	import { platform } from '$lib/utils/platform';

	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);

	// Android renders this screen as Material 3 (mockup screen-MerchantsAndroid,
	// board 2): a back row above the merchant name, the edit action as a tonal
	// chip, a tonal search field and a visible type chip row.
	// `platform` is a module constant, so a plain const, not $derived.
	const IS_ANDROID = platform === 'android';

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

	// Android draws the edit action as a tonal M3 assist chip on the phone and
	// falls back to the shared grey button from `sm` up (mockup).
	const EDIT_BUTTON =
		'btn btn-xs btn-gray whitespace-nowrap flex items-center gap-1.5';
	const EDIT_ACTION_CLASS = IS_ANDROID
		? `max-sm:inline-flex max-sm:h-8 max-sm:shrink-0 max-sm:items-center max-sm:gap-1.25 max-sm:rounded-m3-full max-sm:border-none max-sm:bg-m3-card max-sm:px-3.5 max-sm:text-label max-sm:text-text-ink2 ${EDIT_BUTTON}`
		: EDIT_BUTTON;

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
		typeFilterPlacement={IS_ANDROID ? 'top' : 'batch-header'}
		filterShowAll={IS_ANDROID}
		androidBarcodeInTypeRow
		barcodeButtonVariant="icon"
		maxWidth={false}
	>
		{#snippet header()}
			<!-- Header. Android puts a back row above the merchant name and carries
			     the edit action as a tonal M3 chip (mockup). -->
			{#if IS_ANDROID}
				<div
					class="-mx-4 mb-1 flex items-center gap-1.5 py-2 pr-2 pl-3 sm:hidden"
				>
					<button
						type="button"
						onclick={() => goto(resolve('/merchants'))}
						aria-label={tr('common.back')}
						class="text-text hover:bg-surface-1 inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-m3-full transition-colors"
					>
						<svg
							class="h-5.5 w-5.5"
							fill="none"
							stroke="currentColor"
							stroke-width="2"
							stroke-linecap="round"
							stroke-linejoin="round"
							viewBox="0 0 24 24"
						>
							<path d="M19 12H5M11 6l-6 6 6 6" />
						</svg>
					</button>
					<span class="text-subheading font-medium text-text-muted"
						>{tr('merchantOverview.title')}</span
					>
				</div>
			{/if}
			<div class={IS_ANDROID ? 'max-sm:mb-4 mb-8' : 'mb-8'}>
				<div
					class="flex items-center justify-between {IS_ANDROID
						? 'max-sm:gap-3'
						: ''}"
				>
					<div
						class="flex min-w-0 items-center gap-3 {IS_ANDROID
							? 'max-sm:flex-1'
							: ''}"
					>
						<div
							class="w-2 h-8 shrink-0 rounded-full"
							style="background-color: {merchant?.color}"
						></div>
						<h1
							class="text-text truncate {IS_ANDROID
								? 'max-sm:text-screen-title text-3xl font-bold'
								: 'text-3xl font-bold'}"
						>
							{merchant?.name}
						</h1>
					</div>
					{#if isAdmin && merchant}
						<a
							href={resolve(`/admin/merchants/${merchant.id}/edit`)}
							class={EDIT_ACTION_CLASS}
						>
							{#if IS_ANDROID}
								<svg
									class="hidden h-3.25 w-3.25 shrink-0 max-sm:block"
									fill="none"
									stroke="currentColor"
									stroke-width="2"
									stroke-linecap="round"
									stroke-linejoin="round"
									viewBox="0 0 24 24"
									aria-hidden="true"
								>
									<path d={ICON_PENCIL} />
								</svg>
							{/if}
							{tr('common.edit')}
						</a>
					{/if}
				</div>
			</div>
		{/snippet}

		{#snippet emptyIcon()}
			<svg
				class="w-16 h-16 mx-auto text-text-placeholder mb-4 {IS_ANDROID
					? 'max-sm:mx-0 max-sm:mb-0 max-sm:h-10.5 max-sm:w-10.5'
					: ''}"
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
			<!-- Merchant detail: always-visible plain search input. Android draws it
			     as the tonal M3 field from the mockup instead. -->
			{#if IS_ANDROID}
				<div
					class="mb-3 flex h-11 items-center gap-2 rounded-m3-md bg-m3-card px-3.5 sm:hidden"
				>
					<svg
						class="h-4.25 w-4.25 shrink-0 text-text-faint"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
						stroke-linecap="round"
						stroke-linejoin="round"
						viewBox="0 0 24 24"
						aria-hidden="true"
					>
						<path d={ICON_SEARCH} />
					</svg>
					<input
						type="search"
						bind:value={filters.searchInput}
						placeholder={tr('common.search')}
						aria-label={tr('common.search')}
						class="min-w-0 flex-1 bg-transparent text-body text-text placeholder:text-text-placeholder focus:outline-none"
					/>
				</div>
			{/if}
			<div class={IS_ANDROID ? 'mb-6 max-sm:hidden' : 'mb-6'}>
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
