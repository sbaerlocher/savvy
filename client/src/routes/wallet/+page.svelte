<script lang="ts">
	import { cardsApi, giftCardsApi, vouchersApi } from '$lib/api';
	import WalletView from '$lib/components/ui/WalletView.svelte';
	import PageShell from '$lib/components/layout/PageShell.svelte';
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
	import {
		ICON_BARCODE_TOGGLE,
		ICON_CLIPBOARD_CHECK,
		ICON_FUNNEL
	} from '$lib/icons';
	import { logger } from '$lib/utils/logger';
	import { platform } from '$lib/utils/platform';
	import { walletFilters } from '$lib/stores/walletFilters.svelte';

	// Skeleton phase: on desktop the shell owns the title row and the section
	// (WalletView) stays unmounted until it is re-attached in a later step.
	const IS_DESKTOP = platform === 'other';
	const IS_ANDROID = platform === 'android';

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

	// Count eyebrow above the title (mockup "39 Einträge") — the lists are
	// loaded by this page, so the total is available even while the section
	// stays unmounted.
	const totalItems = $derived(
		cards.length + vouchers.length + giftCards.length
	);

	// Desktop toolbar state (title-row actions), bound into WalletView.
	let headerEyebrow = $state('');
	let selecting = $state(false);
	let filterOpen = $state(false);
	let showImportDialog = $state(false);
	// Initialised by WalletView on mount (persisted under the storage key the
	// page passes down); the page toggle keeps the persistence in sync.
	let showBarcodes = $state(false);

	function toggleBarcodes() {
		showBarcodes = !showBarcodes;
		localStorage.setItem('savvy_wallet_show_barcodes', String(showBarcodes));
	}

	// Filter-button chrome, derived from the shared wallet filter store.
	const isDefaultStatus = $derived(
		walletFilters.statusFilter === 'active' ||
			(walletFilters.typeFilter === 'vouchers' &&
				walletFilters.statusFilter === 'valid')
	);
	const hasActiveFilters = $derived(
		walletFilters.typeFilter !== 'all' ||
			!isDefaultStatus ||
			walletFilters.sortBy !== 'newest' ||
			walletFilters.ownerFilter !== 'all' ||
			walletFilters.favoritesOnly ||
			walletFilters.expiringFilter !== 'all'
	);
	// The filter button carries the picked type as its label (mockup board B).
	const activeTypeLabel = $derived(
		walletFilters.typeFilter === 'cards'
			? tr('merchantOverview.filterCards')
			: walletFilters.typeFilter === 'vouchers'
				? tr('merchantOverview.filterVouchers')
				: walletFilters.typeFilter === 'gift-cards'
					? tr('merchantOverview.filterGiftCards')
					: ''
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

	// Global search entry: ?search (1 = open only, other = the query itself). The
	// input lives in the platform chrome — the desktop nav bar, the Android
	// header, the iOS bottom-nav pill — so the page only mirrors the param into
	// the filter state. Reactive on the param so navigating to /wallet?search=…
	// while already on the page still applies (onMount alone would miss it).
	const searchParam = $derived($page.url.searchParams.get('search'));
	$effect(() => {
		// '1' is the "empty box" signal every chrome sends, so it clears the query
		// rather than being ignored — otherwise the last search sticks forever.
		if (searchParam === '1') walletFilters.searchInput = '';
		else if (searchParam) walletFilters.searchInput = searchParam;
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

<!-- Desktop toolbar on the title row: Select · Filter · Barcodes · Import
     (mockup). All four drive WalletView through the bound state. -->
{#snippet toolbarIcon(d: string, active: boolean)}
	<svg
		class="w-5 h-5 {active ? 'text-accent-hover' : 'text-text-muted'}"
		fill="none"
		stroke="currentColor"
		viewBox="0 0 24 24"
	>
		<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" {d} />
	</svg>
{/snippet}

{#snippet desktopToolbar()}
	<button
		type="button"
		onclick={(e: MouseEvent) => {
			e.stopPropagation();
			filterOpen = !filterOpen;
		}}
		class="title-action relative {hasActiveFilters
			? 'ring-2 ring-accent border-accent text-accent-hover'
			: ''}"
		title={tr('common.filter')}
		aria-label={activeTypeLabel
			? `${tr('common.filter')}: ${activeTypeLabel}`
			: tr('common.filter')}
		aria-expanded={filterOpen}
	>
		{@render toolbarIcon(ICON_FUNNEL, hasActiveFilters)}
		{#if activeTypeLabel}
			<span class="text-label whitespace-nowrap">{activeTypeLabel}</span>
		{/if}
		{#if hasActiveFilters}
			<span class="absolute -top-1 -right-1 w-3 h-3 rounded-full bg-accent"
			></span>
		{/if}
	</button>
	<button
		type="button"
		onclick={() => (selecting = !selecting)}
		disabled={isOffline}
		class="title-action disabled:cursor-not-allowed disabled:opacity-50 {selecting
			? 'ring-2 ring-accent border-accent text-accent-hover'
			: ''}"
		title={tr('batch.selectMode')}
		aria-label={tr('batch.selectMode')}
		aria-pressed={selecting}
	>
		{@render toolbarIcon(ICON_CLIPBOARD_CHECK, selecting)}
	</button>
	<button
		type="button"
		onclick={toggleBarcodes}
		class="title-action px-6 {showBarcodes
			? 'ring-2 ring-accent border-accent'
			: ''}"
		title={showBarcodes ? tr('barcodeToggle.hide') : tr('barcodeToggle.show')}
		aria-label={showBarcodes
			? tr('barcodeToggle.hide')
			: tr('barcodeToggle.show')}
		aria-pressed={showBarcodes}
	>
		{@render toolbarIcon(ICON_BARCODE_TOGGLE, false)}
		<span class="text-sm font-medium whitespace-nowrap text-text-muted">
			{tr('barcodeToggle.label')}
		</span>
	</button>
	<button
		type="button"
		onclick={() => (showImportDialog = true)}
		disabled={isOffline}
		class="title-action whitespace-nowrap disabled:cursor-not-allowed disabled:opacity-50"
		aria-label={tr('settings.import.title')}
	>
		<svg
			class="w-4 h-4"
			fill="none"
			stroke="currentColor"
			viewBox="0 0 24 24"
			aria-hidden="true"
		>
			<path
				stroke-linecap="round"
				stroke-linejoin="round"
				stroke-width="2"
				d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12"
			/>
		</svg>
		<span class="text-label">{tr('settings.import.title')}</span>
	</button>
{/snippet}

{#if isLoading}
	<!-- No eyebrow while loading: the count is not known yet, and the previous
	     `${totalItems} entries` could only ever render "0" here. -->
	<PageShell title={tr('nav.wallet')}>
		<LoadingSpinner />
	</PageShell>
{:else}
	<!-- One title row for every platform, rendered by the shell: the count
	     eyebrow comes out of WalletView (bound), the desktop toolbar drives it
	     through the bound state below. Android in select mode hides the row
	     below `sm` — the fixed contextual app bar replaces it (mockup). -->
	<PageShell
		title={tr('nav.wallet')}
		eyebrow={headerEyebrow}
		mobileActions={!selecting}
		actions={IS_DESKTOP && totalItems > 0 ? desktopToolbar : undefined}
		headerClass={IS_ANDROID && selecting ? 'max-sm:hidden' : ''}
	>
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
			desktopChrome={true}
			bind:headerEyebrow
			bind:selectMode={selecting}
			bind:showFilterMenu={filterOpen}
			bind:showBarcodes
			bind:showImportDialog
		/>
	</PageShell>
{/if}
