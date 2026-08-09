<script lang="ts">
	import { cardsApi, giftCardsApi, vouchersApi } from '$lib/api';
	import WalletView from '$lib/components/ui/WalletView.svelte';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import { authStore } from '$lib/stores/auth';
	import { configStore } from '$lib/stores/config';
	import { locale, t } from '$lib/stores/i18n';
	import { isOnline } from '$lib/stores/offline';
	import { toastStore } from '$lib/stores/toast';
	import type { CardDTO, GiftCardDTO, VoucherDTO } from '$lib/types/api';
	import { page } from '$app/stores';
	import { onMount, tick } from 'svelte';
	import { beforeNavigate } from '$app/navigation';
	import { get } from 'svelte/store';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import { logger } from '$lib/utils/logger';
	import { platform } from '$lib/utils/platform';
	import { walletFilters } from '$lib/stores/walletFilters.svelte';

	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);
	const pageLogger = logger.child('WalletPage');

	const currentUserId = $derived($authStore.user?.id);
	const currentLocale = $derived($locale || 'de-DE');
	const isOffline = $derived(!$isOnline);

	// Feature toggles: only load/show enabled resource types.
	const showCards = $derived($configStore.features.cards);
	const showVouchers = $derived($configStore.features.vouchers);
	const showGiftCards = $derived($configStore.features.gift_cards);
	const enabledTypes = $derived({
		cards: showCards,
		vouchers: showVouchers,
		giftCards: showGiftCards
	});

	let cards = $state<CardDTO[]>([]);
	let vouchers = $state<VoucherDTO[]>([]);
	let giftCards = $state<GiftCardDTO[]>([]);
	let isLoading = $state(true);

	const totalItems = $derived(
		cards.length + vouchers.length + giftCards.length
	);

	// Apply ?type= query param on mount (from /cards /vouchers /gift-cards redirects).
	const validTypes = ['cards', 'vouchers', 'gift-cards'];

	// Offline-first progressive loader for one resource type:
	// Offline-first loader: fetch a first page for fast display, then the
	// remaining pages in the background. Page size stays constant so
	// total_pages lines up with the offset of every follow-up page.
	const PAGE_SIZE = 50;

	async function loadResource<T>(
		listFn: (
			page?: number,
			perPage?: number
		) => Promise<
			{
				pagination?: import('$lib/types/api').PaginationMeta;
			} & Record<string, unknown>
		>,
		key: string,
		assign: (items: T[]) => void,
		append: (items: T[]) => void,
		onFirstPage: () => void
	) {
		const initial = await listFn(1, PAGE_SIZE);
		assign((initial[key] as T[]) ?? []);
		onFirstPage();

		if (initial.pagination && initial.pagination.total > PAGE_SIZE) {
			const totalPages = initial.pagination.total_pages;
			for (let p = 2; p <= totalPages; p++) {
				try {
					const res = await listFn(p, PAGE_SIZE);
					append((res[key] as T[]) ?? []);
				} catch (err) {
					pageLogger.warn(`[Progressive] Failed page ${p} for ${key}:`, err);
				}
			}
		}
	}

	async function loadData() {
		isLoading = true;
		// Clear the spinner as soon as any enabled type has its first page, so
		// the "fast first page" is actually shown while the rest streams in.
		const revealFirstPage = () => {
			isLoading = false;
		};
		try {
			// Three parallel offline-first loads; each shows its first page fast.
			await Promise.all([
				showCards
					? loadResource<CardDTO>(
							cardsApi.list,
							'cards',
							(items) => (cards = items),
							(items) => (cards = [...cards, ...items]),
							revealFirstPage
						)
					: Promise.resolve(),
				showVouchers
					? loadResource<VoucherDTO>(
							vouchersApi.list,
							'vouchers',
							(items) => (vouchers = items),
							(items) => (vouchers = [...vouchers, ...items]),
							revealFirstPage
						)
					: Promise.resolve(),
				showGiftCards
					? loadResource<GiftCardDTO>(
							giftCardsApi.list,
							'gift_cards',
							(items) => (giftCards = items),
							(items) => (giftCards = [...giftCards, ...items]),
							revealFirstPage
						)
					: Promise.resolve()
			]);
		} catch (err) {
			pageLogger.error('[loadData] Failed:', err);
			toastStore.error(tr('common.error'));
		} finally {
			isLoading = false;
		}
	}

	// Search field element, focused when arriving via ?search (global search entry).
	let searchEl = $state<HTMLInputElement | null>(null);
	let wantSearchFocus = $state(false);
	// Whether the search field is shown. There is no permanent search bar; it
	// only appears when entered via the ?search focus path, and stays until
	// cancelled.
	let searchOpen = $state(false);

	// The input only renders once loadData() finishes (isLoading → false), so we
	// focus reactively when the element binds rather than on a fixed timer.
	$effect(() => {
		if (wantSearchFocus && searchEl) {
			searchEl.focus();
			wantSearchFocus = false;
		}
	});

	function cancelSearch() {
		walletFilters.searchInput = '';
		searchOpen = false;
	}

	// Global search entry: ?search (1 = focus only, other = prefill query).
	// Reactive on the param so navigating to /wallet?search=… while already on
	// the page still applies — onMount alone would miss same-page navigations.
	const searchParam = $derived($page.url.searchParams.get('search'));
	$effect(() => {
		if (searchParam) {
			if (searchParam !== '1') walletFilters.searchInput = searchParam;
			searchOpen = true;
			wantSearchFocus = true;
		}
	});

	// Remember the scroll position when leaving, so returning from a detail page
	// restores it (SvelteKit's own restore fires before the async list renders).
	beforeNavigate(() => {
		walletFilters.scrollY = window.scrollY;
	});

	onMount(async () => {
		const t = get(page).url.searchParams.get('type');
		if (t && validTypes.includes(t)) {
			walletFilters.typeFilter = t;
			// Mirror the type-change effect: 'active' is invalid for vouchers.
			walletFilters.statusFilter = t === 'vouchers' ? 'valid' : 'active';
		}
		await loadData();
		// List is now in the DOM — restore the saved scroll position.
		if (walletFilters.scrollY > 0) {
			await tick();
			window.scrollTo(0, walletFilters.scrollY);
		}
	});
</script>

<svelte:head>
	<title>{tr('nav.wallet')} - {tr('common.appName')}</title>
</svelte:head>

{#if isLoading}
	<div class="px-4 max-w-7xl mx-auto pb-20 md:pb-4">
		<PageHeader
			eyebrow={`${totalItems} ${tr('dashboard.entries')}`}
			title={tr('nav.wallet')}
		/>
		<LoadingSpinner />
	</div>
{:else}
	<WalletView
		filters={walletFilters}
		{cards}
		{vouchers}
		{giftCards}
		{enabledTypes}
		{currentUserId}
		{currentLocale}
		{isOffline}
		onReload={loadData}
		barcodeStorageKey="savvy_wallet_show_barcodes"
		idPrefix="wallet"
		matchMerchantName={true}
		typeFilterPlacement="top"
		barcodeButtonVariant="label"
		filterShowAll={false}
		maxWidth={true}
	>
		{#snippet header()}
			<!-- Header: count above title (mockup "7 Einträge"). -->
			<PageHeader
				eyebrow={`${totalItems} ${tr('dashboard.entries')}`}
				title={tr('nav.wallet')}
			/>
		{/snippet}

		{#snippet selectHeader(selectedCount: number)}
			<!-- iOS select mode: the eyebrow counts the selection instead of the
			     total, and the bell / "new" actions step aside (mockup Phone 3). -->
			<PageHeader
				eyebrow={tr('batch.selected', { count: selectedCount })}
				title={tr('nav.wallet')}
				mobileActions={false}
			/>
		{/snippet}

		{#snippet searchField()}
			<!-- Search field: only shown when arriving via ?search focus path.
			     iOS puts search in the bottom-nav pill and Android in the header,
			     so only the desktop fallback shows this top field (the query still
			     filters via the ?search param on every platform). -->
			{#if searchOpen && platform === 'other'}
				<div class="mb-6 flex gap-2">
					<input
						type="search"
						bind:this={searchEl}
						bind:value={walletFilters.searchInput}
						placeholder={tr('common.search')}
						class="w-full flex-1 rounded-md border border-border-field bg-white px-4 py-2 focus:border-accent focus:ring-accent"
					/>
					<button
						type="button"
						onclick={cancelSearch}
						class="btn btn-ghost whitespace-nowrap"
					>
						{tr('common.cancel')}
					</button>
				</div>
			{/if}
		{/snippet}
	</WalletView>
{/if}
