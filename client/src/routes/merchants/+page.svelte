<script lang="ts">
	import { resolve } from '$app/paths';
	import { cardsApi, vouchersApi, giftCardsApi, merchantsApi } from '$lib/api';
	import BottomSheet from '$lib/components/BottomSheet.svelte';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import { authStore } from '$lib/stores/auth';
	import { t } from '$lib/stores/i18n';
	import { isOnline } from '$lib/stores/offline';
	import { toastStore } from '$lib/stores/toast';
	import {
		deriveMerchantOverview,
		type MerchantOverview
	} from '$lib/utils/merchant-aggregator';
	import { onMount } from 'svelte';
	import { get } from 'svelte/store';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import MerchantFilters from '$lib/components/MerchantFilters.svelte';
	import { ICON_CLOSE, ICON_FILTER_LINES, ICON_SEARCH } from '$lib/icons';
	import { platform } from '$lib/utils/platform';

	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);

	// Android renders this screen as Material 3 (mockup screen-MerchantsAndroid):
	// a one-column list of tonal cards, a single outlined "Filter" chip instead
	// of the search+filter+new row, and search moved into the filter sheet.
	// `platform` is a module constant, so a plain const, not $derived.
	const IS_ANDROID = platform === 'android';
	// iOS renders its own chrome for this screen (mockup screen-MerchantsIOS): a
	// count eyebrow, a right-aligned glass "Filter" pill next to an admin "+",
	// one stacked column of bordered cards and a glass empty state.
	const IOS = platform === 'ios';

	// Desktop renders its own chrome (mockup screen-MerchantsDesktop): a count
	// eyebrow above the title, a single filter icon button, and search moved into
	// the filter side panel.
	const IS_DESKTOP = platform === 'other';

	const isAdmin = $derived($authStore.user?.is_admin || false);
	const isOffline = $derived(!$isOnline);

	let merchants = $state<MerchantOverview[]>([]);
	let isLoading = $state(true);
	let searchInput = $state('');
	let sortBy = $state('name-asc');
	let typeFilter = $state('all');
	let statusFilter = $state('all');
	let showFilterMenu = $state(false);

	const hasActiveFilters = $derived(
		sortBy !== 'name-asc' || typeFilter !== 'all' || statusFilter !== 'all'
	);

	// Only the narrowing filters — `hasActiveFilters` also covers `sortBy`, which
	// reorders without dropping rows and is restored from localStorage, so it
	// would leave the eyebrow claiming "filtered" on every later visit.
	const isNarrowed = $derived(
		typeFilter !== 'all' || statusFilter !== 'all' || !!searchInput.trim()
	);

	// The side panel is the desktop form of the filter UI; the native layouts get
	// the bottom sheet from the same `showFilterMenu` flag. Gating on the platform
	// keeps the panel out of the DOM there — a CSS-only hide would leave a second
	// copy of the search field behind.
	const panelOpen = $derived(showFilterMenu && IS_DESKTOP);

	const filteredMerchants = $derived.by(() => {
		let result = visibleMerchants;

		// Search filter
		if (searchInput.trim()) {
			const q = searchInput.trim().toLowerCase();
			result = result.filter((m) => m.name.toLowerCase().includes(q));
		}

		// Type filter
		if (typeFilter !== 'all') {
			result = result.filter((m) => {
				switch (typeFilter) {
					case 'cards':
						return m.cards_count + m.cards_inactive_count > 0;
					case 'vouchers':
						return m.vouchers_count + m.vouchers_inactive_count > 0;
					case 'gift-cards':
						return m.gift_cards_count + m.gift_cards_inactive_count > 0;
					default:
						return true;
				}
			});
		}

		// Status filter
		if (statusFilter === 'active') {
			result = result.filter(
				(m) => m.cards_count + m.vouchers_count + m.gift_cards_count > 0
			);
		} else if (statusFilter === 'inactive') {
			result = result.filter(
				(m) =>
					m.cards_inactive_count +
						m.vouchers_inactive_count +
						m.gift_cards_inactive_count >
					0
			);
		}

		// Sort
		result = [...result].sort((a, b) => {
			const totalA =
				a.cards_count +
				a.cards_inactive_count +
				a.vouchers_count +
				a.vouchers_inactive_count +
				a.gift_cards_count +
				a.gift_cards_inactive_count;
			const totalB =
				b.cards_count +
				b.cards_inactive_count +
				b.vouchers_count +
				b.vouchers_inactive_count +
				b.gift_cards_count +
				b.gift_cards_inactive_count;
			const activeA = a.cards_count + a.vouchers_count + a.gift_cards_count;
			const activeB = b.cards_count + b.vouchers_count + b.gift_cards_count;
			const inactiveA =
				a.cards_inactive_count +
				a.vouchers_inactive_count +
				a.gift_cards_inactive_count;
			const inactiveB =
				b.cards_inactive_count +
				b.vouchers_inactive_count +
				b.gift_cards_inactive_count;
			switch (sortBy) {
				case 'name-asc':
					return a.name.localeCompare(b.name);
				case 'name-desc':
					return b.name.localeCompare(a.name);
				case 'items-desc':
					return totalB - totalA;
				case 'active-desc':
					return activeB - activeA;
				case 'inactive-desc':
					return inactiveB - inactiveA;
				case 'balance-desc':
					return b.total_gift_card_balance - a.total_gift_card_balance;
				default:
					return 0;
			}
		});

		return result;
	});

	async function loadMerchants() {
		isLoading = true;
		try {
			const [cardsRes, vouchersRes, giftCardsRes, merchantsRes] =
				await Promise.all([
					cardsApi.list().catch(() => ({ cards: [] })),
					vouchersApi.list().catch(() => ({ vouchers: [] })),
					giftCardsApi.list().catch(() => ({ gift_cards: [] })),
					merchantsApi.list().catch(() => ({ merchants: [] }))
				]);

			merchants = deriveMerchantOverview(
				cardsRes.cards,
				vouchersRes.vouchers,
				giftCardsRes.gift_cards,
				merchantsRes.merchants
			);
		} catch {
			toastStore.error(tr('merchantOverview.loadError'));
		} finally {
			isLoading = false;
		}
	}

	onMount(() => {
		loadFilters();
		loadMerchants();
	});

	function getTotalItems(m: MerchantOverview): number {
		return (
			m.cards_count +
			m.cards_inactive_count +
			m.vouchers_count +
			m.vouchers_inactive_count +
			m.gift_cards_count +
			m.gift_cards_inactive_count
		);
	}

	function formatBalance(balance: number): string {
		return balance.toFixed(2);
	}

	function loadFilters() {
		try {
			const saved = localStorage.getItem('savvy_merchants_filters');
			if (saved) {
				const filters = JSON.parse(saved);
				sortBy = filters.sortBy || 'name-asc';
				typeFilter = filters.typeFilter || 'all';
				statusFilter = filters.statusFilter || 'all';
			}
		} catch {
			// ignore
		}
	}

	function saveFilters() {
		try {
			localStorage.setItem(
				'savvy_merchants_filters',
				JSON.stringify({ sortBy, typeFilter, statusFilter })
			);
		} catch {
			// ignore
		}
	}

	$effect(() => {
		// Reference filter values so the effect re-runs on any change
		void [sortBy, typeFilter, statusFilter];
		saveFilters();
	});

	function resetFilters() {
		sortBy = 'name-asc';
		typeFilter = 'all';
		statusFilter = 'all';
		searchInput = '';
	}

	const merchantsWithCards = $derived(
		merchants.filter((m) => m.cards_count + m.cards_inactive_count > 0).length
	);
	const merchantsWithVouchers = $derived(
		merchants.filter((m) => m.vouchers_count + m.vouchers_inactive_count > 0)
			.length
	);
	const merchantsWithGiftCards = $derived(
		merchants.filter(
			(m) => m.gift_cards_count + m.gift_cards_inactive_count > 0
		).length
	);

	// Chrome row eyebrow, mirroring the mockup boards: a plain merchant count, and
	// "shown of total · filtered" once a filter narrows the list.
	const visibleMerchants = $derived(
		isAdmin ? merchants : merchants.filter((m) => getTotalItems(m) > 0)
	);
	const chromeEyebrow = $derived(
		isNarrowed
			? tr('dashboard.entriesFiltered', {
					shown: filteredMerchants.length,
					total: visibleMerchants.length
				})
			: `${visibleMerchants.length} ${tr('merchantOverview.title')}`
	);

	const listSortOptions = $derived([
		{ value: 'name-asc', label: tr('merchantOverview.sortNameAsc') },
		{ value: 'name-desc', label: tr('merchantOverview.sortNameDesc') },
		{ value: 'items-desc', label: tr('merchantOverview.sortItemsDesc') },
		{ value: 'active-desc', label: tr('merchantOverview.sortActiveDesc') },
		{ value: 'inactive-desc', label: tr('merchantOverview.sortInactiveDesc') },
		{ value: 'balance-desc', label: tr('merchantOverview.sortBalanceDesc') }
	]);

	// Empty-state card. The phone carries a native card — a flat M3 one on
	// Android, a glass grouped-inset one on iOS (mockups); from `sm` up every
	// platform falls back to the plain panel.
	const EMPTY_CARD_CLASS = IS_ANDROID
		? 'max-sm:rounded-m3-lg max-sm:bg-m3-card max-sm:flex max-sm:flex-col max-sm:items-center max-sm:gap-2 max-sm:px-6.5 max-sm:py-11.5 bg-surface-1 rounded-lg p-12 text-center'
		: IOS
			? 'max-sm:liquid-glass-card max-sm:rounded-[var(--radius-inset)] max-sm:flex max-sm:flex-col max-sm:items-center max-sm:gap-2.5 max-sm:px-6.5 max-sm:py-11 bg-surface-1 rounded-lg p-12 text-center'
			: 'bg-surface-1 rounded-lg p-12 text-center';
	const EMPTY_TITLE_CLASS =
		IS_ANDROID || IOS
			? 'text-text-muted max-sm:mb-0 max-sm:mt-1.5 max-sm:text-subheading max-sm:text-text text-lg mb-4'
			: 'text-text-muted text-lg mb-4';

	const listStatusOptions = $derived([
		{ value: 'all', label: tr('merchantOverview.filterStatusAll') },
		{ value: 'active', label: tr('merchantOverview.filterStatusActive') },
		{ value: 'inactive', label: tr('merchantOverview.filterStatusInactive') }
	]);
</script>

<svelte:head>
	<title>{tr('merchantOverview.title')} - {tr('common.appName')}</title>
</svelte:head>

<!-- iOS admin shortcut to merchant creation, in the same glass pill shape as
     the Filter control next to it. Disabled offline like every other mutation. -->
{#snippet iosAdminAddPill()}
	<a
		href={resolve('/admin/merchants/new')}
		onclick={(e) => {
			if (isOffline) e.preventDefault();
		}}
		aria-label={tr('merchantOverview.addMerchant')}
		title={tr('merchantOverview.addMerchant')}
		class="liquid-glass-card inline-flex h-10 w-10 items-center justify-center rounded-[var(--radius-lg)] text-accent-700 {isOffline
			? 'pointer-events-none opacity-50'
			: ''}"
	>
		<svg
			class="h-5 w-5"
			fill="none"
			stroke="currentColor"
			stroke-width="2.1"
			viewBox="0 0 24 24"
			aria-hidden="true"
		>
			<path stroke-linecap="round" d="M12 5v14M5 12h14" />
		</svg>
	</a>
{/snippet}

<!-- Storefront glyph above the empty-state copy (Android mockup). -->
{#snippet emptyIcon()}
	<svg
		class="h-10.5 w-10.5 text-text-placeholder"
		fill="none"
		stroke="currentColor"
		stroke-width="1.5"
		viewBox="0 0 24 24"
		aria-hidden="true"
	>
		<path
			stroke-linecap="round"
			stroke-linejoin="round"
			d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4"
		/>
	</svg>
{/snippet}

<!-- Desktop chrome row action: the lone filter button beside the title (mockup
     board 1A/1B). Search and the type/sort/status groups live in the panel it
     opens, so this is the whole toolbar on desktop. -->
{#snippet desktopFilterButton()}
	<button
		type="button"
		onclick={(e: MouseEvent) => {
			e.stopPropagation();
			showFilterMenu = !showFilterMenu;
		}}
		class="control relative flex items-center justify-center rounded-lg border bg-white px-4 transition-colors hover:bg-surface-1 {hasActiveFilters
			? 'border-accent text-accent-hover ring-2 ring-accent'
			: 'border-border-field text-text-muted'}"
		title={tr('common.filter')}
		aria-label={tr('common.filter')}
		aria-expanded={showFilterMenu}
	>
		<svg
			class="h-5 w-5"
			fill="none"
			stroke="currentColor"
			stroke-width="2"
			viewBox="0 0 24 24"
			aria-hidden="true"
		>
			<path
				stroke-linecap="round"
				stroke-linejoin="round"
				d={ICON_FILTER_LINES}
			/>
		</svg>
		{#if hasActiveFilters}
			<span
				class="absolute -top-1 -right-1 h-3 w-3 rounded-full border-2 border-paper bg-accent"
			></span>
		{/if}
	</button>
{/snippet}

<!-- Search field for the desktop filter panel (mockup board 1B). Desktop has no
     header search — this is the only one, so the panel renders at every desktop
     width and the bottom sheet stays closed there. -->
{#snippet searchField()}
	<label
		class="mb-4.5 flex items-center gap-2.25 control rounded-md border border-border-field bg-white px-3.5"
	>
		<svg
			class="h-4 w-4 shrink-0 text-text-faint"
			fill="none"
			stroke="currentColor"
			stroke-width="2"
			viewBox="0 0 24 24"
			aria-hidden="true"
		>
			<path stroke-linecap="round" stroke-linejoin="round" d={ICON_SEARCH} />
		</svg>
		<input
			type="search"
			data-testid="merchant-search"
			bind:value={searchInput}
			placeholder={tr('common.search')}
			aria-label={tr('common.search')}
			class="min-w-0 flex-1 bg-transparent text-body text-text placeholder:text-text-placeholder focus:outline-none"
		/>
	</label>
{/snippet}

<!-- One count chip on a merchant card: outlined for active counts, filled-flat
     for the inactive/expired ones. Android steps the chip up to the M3 size
     from the mockup; the other platforms keep the smaller pill. -->
{#snippet countChip(count: number, label: string, active: boolean)}
	<span
		class="inline-flex items-center rounded-full {IS_ANDROID || IOS
			? 'max-sm:text-chip max-sm:px-2.75 max-sm:py-1 px-2.5 py-0.5 text-xs font-medium'
			: 'px-2.5 py-0.5 text-xs font-medium'} {active
			? 'bg-surface-1 border border-border text-text-muted'
			: 'bg-border-soft text-text-faint'}"
	>
		{count}
		{label}
	</span>
{/snippet}

<div class="px-4 pb-20 md:pb-4">
	<!-- Header. Android carries the count as an eyebrow (mockup); the mobile
	     header actions stay on their default for every platform — on iOS that row
	     is the only create entry point on this screen. -->
	<PageHeader
		title={tr('merchantOverview.title')}
		eyebrow={IS_ANDROID
			? `${filteredMerchants.length} ${tr('nav.merchants')}`
			: IS_DESKTOP || IOS
				? chromeEyebrow
				: undefined}
		mobileActions={!IS_DESKTOP}
		actions={IS_DESKTOP && merchants.length > 0
			? desktopFilterButton
			: undefined}
	/>

	{#if IOS && merchants.length > 0}
		<!-- iOS phone chrome: a right-aligned glass Filter pill. Search, type,
		     status and sort all live inside the filter sheet (mockup). Admins get
		     an extra "+" pill — the header glyph opens the global TypeChoiceDialog,
		     which creates resources, not merchants, so this is the only route to
		     /admin/merchants/new on a phone. An empty list has nothing to filter,
		     so the row drops out there and the empty branch carries the "+" alone;
		     without the guard both rows would draw the same glyph twice. -->
		<div class="mb-3.5 flex items-center justify-end gap-2 sm:hidden">
			{#if isAdmin}
				{@render iosAdminAddPill()}
			{/if}
			<button
				type="button"
				data-testid="merchant-filter-pill"
				onclick={(e: MouseEvent) => {
					e.stopPropagation();
					showFilterMenu = !showFilterMenu;
				}}
				aria-pressed={hasActiveFilters}
				aria-expanded={showFilterMenu}
				class="liquid-glass-card relative inline-flex h-10 items-center gap-1.75 rounded-[var(--radius-lg)] px-3.5 text-body font-semibold text-text-muted"
			>
				<svg
					class="h-4.5 w-4.5 shrink-0"
					fill="none"
					stroke="currentColor"
					stroke-width="1.9"
					viewBox="0 0 24 24"
					aria-hidden="true"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						d={ICON_FILTER_LINES}
					/>
				</svg>
				{tr('common.filter')}
				{#if hasActiveFilters}
					<span
						class="absolute -top-0.75 -right-0.75 h-2 w-2 rounded-full border border-surface bg-accent"
					></span>
				{/if}
			</button>
		</div>
	{/if}

	{#if IS_ANDROID}
		<!-- M3 chip row: one content-sized outlined "Filter" chip. Search lives in
		     the sheet, and "New" is the nav FAB (mockup). -->
		<div class="mb-3.5 flex items-center gap-2 sm:hidden">
			<button
				type="button"
				data-testid="merchant-filter-chip"
				onclick={(e: MouseEvent) => {
					e.stopPropagation();
					showFilterMenu = !showFilterMenu;
				}}
				aria-pressed={hasActiveFilters}
				aria-expanded={showFilterMenu}
				class="relative inline-flex h-8 items-center gap-1.5 rounded-m3-sm px-3.5 text-label whitespace-nowrap transition-colors {hasActiveFilters
					? 'bg-m3-secondary-container text-m3-on-secondary-container'
					: 'border border-border-chip text-text-muted'}"
			>
				<svg
					class="h-4.5 w-4.5 shrink-0"
					fill="none"
					stroke="currentColor"
					stroke-width="1.9"
					viewBox="0 0 24 24"
					aria-hidden="true"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						d={ICON_FILTER_LINES}
					/>
				</svg>
				{tr('common.filter')}
			</button>
		</div>
	{/if}

	{#if merchants.length > 0 && !IS_DESKTOP}
		<!-- Search + Filter + New Button. Android replaces the phone-width row with
		     the M3 chip above, so it starts at `sm` there. Desktop drops the row
		     entirely: the chrome row carries the filter button and search moved
		     into the panel (mockup). -->
		<div
			class="flex flex-col sm:flex-row gap-3 mb-6 {IS_ANDROID || IOS
				? 'max-sm:hidden'
				: ''}"
		>
			<!-- Search Bar -->
			<div class="flex-1">
				<input
					type="search"
					data-testid="merchant-search"
					bind:value={searchInput}
					placeholder={tr('common.search')}
					class="w-full px-4 py-2 bg-white border border-border-field rounded-md focus:ring-accent focus:border-accent"
				/>
			</div>

			<!-- Desktop: Filter + New Button -->
			<div class="hidden sm:flex gap-3">
				<button
					type="button"
					onclick={(e: MouseEvent) => {
						e.stopPropagation();
						showFilterMenu = !showFilterMenu;
					}}
					class="flex items-center justify-center gap-2 control px-4 bg-white border border-border-field rounded-md hover:bg-surface-1 transition-colors relative"
					title={tr('common.filter')}
					aria-label={tr('common.filter')}
					aria-expanded={showFilterMenu}
				>
					<svg
						class="w-5 h-5 text-text-muted"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z"
						/>
					</svg>
					{#if hasActiveFilters}
						<span
							class="absolute -top-1 -right-1 w-3 h-3 bg-accent rounded-full"
						></span>
					{/if}
				</button>
				{#if isAdmin}
					<a
						href={resolve('/admin/merchants/new')}
						onclick={(e) => {
							if (isOffline) e.preventDefault();
						}}
						class="btn btn-primary whitespace-nowrap flex items-center gap-2 {isOffline
							? 'opacity-50 cursor-not-allowed pointer-events-none blur-[0.5px]'
							: ''}"
					>
						{#if isOffline}
							<svg
								class="w-4 h-4"
								fill="none"
								stroke="currentColor"
								viewBox="0 0 24 24"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"
								></path>
							</svg>
						{:else}
							<span>+</span>
						{/if}
						{tr('merchantOverview.addMerchant')}
					</a>
				{/if}
			</div>

			<!-- Mobile: Filter + New Button (eigene Zeile) -->
			<div class="flex sm:hidden gap-3">
				<button
					type="button"
					onclick={(e: MouseEvent) => {
						e.stopPropagation();
						showFilterMenu = !showFilterMenu;
					}}
					class="flex-1 flex items-center justify-center control bg-white border border-border-field rounded-md hover:bg-surface-1 transition-colors relative"
					aria-label={tr('common.filter')}
					aria-expanded={showFilterMenu}
				>
					<svg
						class="w-5 h-5 text-text-muted"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z"
						/>
					</svg>
					{#if hasActiveFilters}
						<span
							class="absolute -top-1 -right-1 w-3 h-3 bg-accent rounded-full"
						></span>
					{/if}
				</button>
				{#if isAdmin}
					<a
						href={resolve('/admin/merchants/new')}
						onclick={(e) => {
							if (isOffline) e.preventDefault();
						}}
						class="flex-1 btn btn-primary flex items-center justify-center gap-2 {isOffline
							? 'opacity-50 cursor-not-allowed pointer-events-none blur-[0.5px]'
							: ''}"
					>
						{#if isOffline}
							<svg
								class="w-4 h-4"
								fill="none"
								stroke="currentColor"
								viewBox="0 0 24 24"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"
								></path>
							</svg>
						{:else}
							<span>+</span>
						{/if}
						{tr('merchantOverview.addMerchant')}
					</a>
				{/if}
			</div>
		</div>
	{:else if !isLoading && isAdmin && !IS_DESKTOP}
		{#if IOS}
			<!-- Empty list, admin: the "+" pill is the only path to merchant
			     creation on a phone. `platform` is width-independent — an iPad is
			     `ios` at every width — so the labelled button below still covers
			     `sm` and up. -->
			<div class="mb-3.5 flex justify-end sm:hidden">
				{@render iosAdminAddPill()}
			</div>
		{/if}
		<div class="inline-block mb-6 {IS_ANDROID || IOS ? 'max-sm:hidden' : ''}">
			<a
				href={resolve('/admin/merchants/new')}
				onclick={(e) => {
					if (isOffline) e.preventDefault();
				}}
				class="btn btn-primary whitespace-nowrap flex items-center gap-2 {isOffline
					? 'opacity-50 cursor-not-allowed pointer-events-none blur-[0.5px]'
					: ''}"
			>
				{#if isOffline}
					<svg
						class="w-4 h-4"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"
						></path>
					</svg>
				{:else}
					<span>+</span>
				{/if}
				{tr('merchantOverview.addMerchant')}
			</a>
		</div>
	{/if}

	<!-- One grid wrapper around every state, with the panel rendered once outside
	     the branch chain. Svelte does not share DOM across if-branches, so a panel
	     rendered per branch would remount its search input whenever a keystroke
	     crossed the zero-results boundary — dropping focus mid-word. -->
	<div class="grid grid-cols-1 {panelOpen ? 'lg:grid-cols-3' : ''} gap-6">
		<div class={panelOpen ? 'lg:col-span-2' : ''}>
			<!-- Loading -->
			{#if isLoading}
				<LoadingSpinner />
			{:else if filteredMerchants.length === 0 && (searchInput || hasActiveFilters)}
				<!-- No results with filters -->
				<div class={EMPTY_CARD_CLASS}>
					{#if IS_ANDROID || IOS}
						{@render emptyIcon()}
					{/if}
					<p class={EMPTY_TITLE_CLASS}>{tr('search.no_results')}</p>
					<button type="button" onclick={resetFilters} class="btn btn-ghost">
						{tr('common.resetFilters')}
					</button>
				</div>
			{:else if filteredMerchants.length === 0}
				<!-- Empty State -->
				<div class={EMPTY_CARD_CLASS}>
					{#if IS_ANDROID || IOS}
						{@render emptyIcon()}
					{/if}
					<p class={EMPTY_TITLE_CLASS}>
						{tr('merchantOverview.noMerchants')}
					</p>
					<p
						class="text-text-faint text-sm {IS_ANDROID || IOS
							? 'max-sm:mt-0 max-sm:max-w-62.5 mt-1'
							: 'mt-1'}"
					>
						{tr('merchantOverview.noMerchantsHint')}
					</p>
				</div>
			{:else}
				<div
					class="grid grid-cols-1 sm:grid-cols-2 {panelOpen
						? ''
						: 'lg:grid-cols-3'} {IS_ANDROID || IOS ? 'max-sm:gap-3' : ''} gap-4"
				>
					{#each filteredMerchants as merchant (merchant.id)}
						<!-- Native card: the merchant colour is a bar inside the clipped
						     corner radius, not a left border on the box — flat and tonal on
						     Android, bordered on white for iOS (mockups). -->
						<a
							href={resolve(`/merchants/${merchant.id}`)}
							class="overflow-hidden group h-full flex flex-col {IS_ANDROID
								? 'relative max-sm:rounded-m3-lg max-sm:bg-m3-card sm:rounded-lg sm:bg-white sm:shadow sm:hover:shadow-md sm:transition-shadow'
								: IS_DESKTOP
									? 'relative bg-white rounded-lg shadow-card transition-shadow'
									: IOS
										? 'relative max-sm:rounded-[var(--radius-2xl)] max-sm:border max-sm:border-border max-sm:bg-surface max-sm:shadow-sm max-sm:transition-transform max-sm:active:scale-[0.99] sm:rounded-lg sm:bg-white sm:shadow sm:hover:shadow-md sm:transition-shadow'
										: 'bg-white rounded-lg shadow hover:shadow-md transition-shadow'}"
							style={IS_ANDROID || IS_DESKTOP || IOS
								? undefined
								: `border-left: 6px solid ${merchant.color}`}
						>
							{#if IS_ANDROID || IS_DESKTOP || IOS}
								<span
									class="absolute inset-y-0 left-0 w-1.5"
									style="background-color: {merchant.color}"
								></span>
							{/if}
							<div
								class="flex flex-col flex-1 {IS_ANDROID || IOS
									? 'max-sm:gap-2.25 max-sm:pt-3.75 max-sm:pr-4 max-sm:pb-3.25 max-sm:pl-5 p-4'
									: IS_DESKTOP
										? 'pt-4.25 pr-4.5 pb-3.75 pl-5.5'
										: 'p-4'}"
							>
								<!-- Merchant name -->
								<h2
									class="text-text truncate {IS_DESKTOP
										? 'text-heading'
										: 'text-lg font-semibold'} {IS_ANDROID
										? 'max-sm:text-heading'
										: ''} {IOS ? 'max-sm:-tracking-[0.01em]' : ''}"
								>
									{merchant.name}
								</h2>

								<!-- Active item counts -->
								{#if merchant.cards_count > 0 || merchant.vouchers_count > 0 || merchant.gift_cards_count > 0}
									<div
										class="flex flex-wrap {IS_ANDROID || IOS
											? 'max-sm:mt-0 max-sm:gap-1.5 mt-3 gap-2'
											: 'mt-3 gap-2'}"
									>
										{#if merchant.cards_count > 0}
											{@render countChip(
												merchant.cards_count,
												tr('merchantOverview.cards'),
												true
											)}
										{/if}
										{#if merchant.vouchers_count > 0}
											{@render countChip(
												merchant.vouchers_count,
												tr('merchantOverview.vouchers'),
												true
											)}
										{/if}
										{#if merchant.gift_cards_count > 0}
											{@render countChip(
												merchant.gift_cards_count,
												tr('merchantOverview.giftCards'),
												true
											)}
										{/if}
									</div>
								{/if}

								<!-- Inactive/expired item counts (greyed out) -->
								{#if merchant.cards_inactive_count > 0 || merchant.vouchers_inactive_count > 0 || merchant.gift_cards_inactive_count > 0}
									<div
										class="flex flex-wrap {IS_ANDROID || IOS
											? 'max-sm:mt-0 max-sm:gap-1.5 mt-1.5 gap-2'
											: 'mt-1.5 gap-2'}"
									>
										{#if merchant.cards_inactive_count > 0}
											{@render countChip(
												merchant.cards_inactive_count,
												tr('merchantOverview.cards'),
												false
											)}
										{/if}
										{#if merchant.vouchers_inactive_count > 0}
											{@render countChip(
												merchant.vouchers_inactive_count,
												tr('merchantOverview.vouchers'),
												false
											)}
										{/if}
										{#if merchant.gift_cards_inactive_count > 0}
											{@render countChip(
												merchant.gift_cards_inactive_count,
												tr('merchantOverview.giftCards'),
												false
											)}
										{/if}
									</div>
								{/if}

								<!-- Balance + total items (pushed to bottom) -->
								<div
									class="mt-auto flex items-center justify-between text-text-subtle {IS_ANDROID ||
									IOS
										? 'max-sm:pt-1.25 max-sm:text-chip max-sm:font-normal pt-3 text-sm'
										: 'pt-3 text-sm'}"
								>
									<span
										>{getTotalItems(merchant)}
										{tr('merchantOverview.items')}</span
									>
									{#if merchant.total_gift_card_balance > 0}
										<span
											class="font-semibold text-text {IS_ANDROID || IOS
												? 'max-sm:text-label max-sm:tabular-nums'
												: ''}"
										>
											CHF {formatBalance(merchant.total_gift_card_balance)}
										</span>
									{/if}
								</div>
							</div>
						</a>
					{/each}
				</div>
			{/if}
		</div>

		<!-- Filter Side-Panel (desktop only). Shown at every desktop width, not
		     `hidden lg:block`: it holds the only search field on desktop, so hiding
		     it below `lg` would leave no way to search — the bottom sheet is gated
		     off there. Above `lg` it takes its own grid column. -->
		{#if panelOpen}
			{@render filterSidePanel()}
		{/if}
	</div>
</div>

<!-- Desktop filter side panel, rendered from a single call site outside the
     state chain so its search input keeps its identity — and its focus — while
     the list swaps between results, no-results and empty. -->
{#snippet filterSidePanel()}
	<div class="lg:col-span-1">
		<div class="bg-white rounded-xl shadow-card sticky top-4 overflow-hidden">
			<!-- Header -->
			<div class="px-5 py-4 bg-surface-1 border-b border-border-soft">
				<div class="flex items-center justify-between">
					<div class="flex items-center gap-2">
						<svg
							class="w-4 h-4 text-text-subtle"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z"
							/>
						</svg>
						<h3 class="text-sm font-semibold text-text">
							{tr('common.filter')}
						</h3>
					</div>
					<div class="flex items-center gap-2.5">
						<span
							class="text-xs text-text-subtle bg-white px-2.5 py-1 rounded-full border border-border tabular-nums"
						>
							{tr('common.results', { count: filteredMerchants.length })}
						</span>
						<button
							type="button"
							onclick={() => (showFilterMenu = false)}
							class="text-text-faint hover:text-text-muted transition-colors"
							aria-label={tr('common.closeFilters')}
						>
							<svg
								class="w-4 h-4"
								fill="none"
								stroke="currentColor"
								viewBox="0 0 24 24"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d={ICON_CLOSE}
								></path>
							</svg>
						</button>
					</div>
				</div>
			</div>

			<div class="p-5">
				{@render searchField()}
				<MerchantFilters
					bind:typeFilter
					bind:statusFilter
					bind:sortBy
					sortOptions={listSortOptions}
					statusOptions={listStatusOptions}
					cardsCount={merchantsWithCards}
					vouchersCount={merchantsWithVouchers}
					giftCardsCount={merchantsWithGiftCards}
					{hasActiveFilters}
					onReset={resetFilters}
					idPrefix="merchants-desktop"
				/>
			</div>
		</div>
	</div>
{/snippet}

<!-- Mobile Filter Bottom Sheet -->
<BottomSheet
	open={showFilterMenu && !IS_DESKTOP}
	onClose={() => (showFilterMenu = false)}
	maxHeight="80vh"
	ariaLabel={tr('common.filter')}
	tonalAndroid
>
	<div class={IS_ANDROID ? 'px-5.5 pb-5.5' : IOS ? 'px-4 pt-1 pb-4' : 'p-6'}>
		<div
			class="flex items-center justify-between {IS_ANDROID
				? 'pt-1.5 pb-3.5'
				: IOS
					? 'px-1 pb-3'
					: 'mb-4'}"
		>
			<!-- Both native sheet titles carry the filter glyph next to the label
			     (mockups). -->
			<h3
				class="flex items-center gap-2.25 text-text {IS_ANDROID || IOS
					? 'text-heading'
					: 'text-lg font-semibold'}"
			>
				{#if IS_ANDROID || IOS}
					<svg
						class="h-4.25 w-4.25 shrink-0 text-text-subtle"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
						viewBox="0 0 24 24"
						aria-hidden="true"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							d={ICON_FILTER_LINES}
						/>
					</svg>
				{/if}
				{tr('common.filter')}
			</h3>
			{#if IOS}
				<!-- iOS closes the sheet with a "Done" text action, not a glyph
				     (mockup); the scrim tap and Escape still dismiss it. -->
				<button
					type="button"
					onclick={() => (showFilterMenu = false)}
					class="text-[length:var(--text-code)] font-semibold text-accent-700 transition-opacity active:opacity-60"
				>
					{tr('common.done')}
				</button>
			{:else}
				<button
					type="button"
					onclick={() => (showFilterMenu = false)}
					class="text-text-faint hover:text-text-muted transition-colors"
					aria-label={tr('common.close')}
				>
					<svg
						class={IS_ANDROID ? 'h-5.25 w-5.25' : 'w-6 h-6'}
						fill="none"
						stroke="currentColor"
						stroke-width="2"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							d={ICON_CLOSE}
						/>
					</svg>
				</button>
			{/if}
		</div>

		{#if IOS}
			<!-- iOS moves the merchant search into the sheet as a plain bordered
			     field above the grouped-inset cards (mockup). -->
			<label
				class="mb-3 flex h-11 items-center gap-2.25 rounded-[var(--radius-lg)] border border-border-field bg-surface px-3.5"
			>
				<svg
					class="h-4.25 w-4.25 shrink-0 text-text-faint"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
					viewBox="0 0 24 24"
					aria-hidden="true"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						d={ICON_SEARCH}
					/>
				</svg>
				<input
					type="search"
					data-testid="merchant-search-sheet-ios"
					bind:value={searchInput}
					placeholder={tr('common.search')}
					aria-label={tr('common.search')}
					class="min-w-0 flex-1 bg-transparent text-[length:var(--text-code)] text-text placeholder:text-text-placeholder focus:outline-none"
				/>
			</label>
		{/if}

		{#if IS_ANDROID}
			<!-- Android moves the merchant search off the list and into the sheet as
			     an outlined M3 text field (mockup). -->
			<div
				class="mb-4.5 flex h-12 items-center gap-2.25 rounded-m3-xs border border-border-field px-3.5"
			>
				<svg
					class="h-4.5 w-4.5 shrink-0 text-text-faint"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
					viewBox="0 0 24 24"
					aria-hidden="true"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						d={ICON_SEARCH}
					/>
				</svg>
				<input
					type="search"
					data-testid="merchant-search-sheet"
					bind:value={searchInput}
					placeholder={tr('common.search')}
					aria-label={tr('common.search')}
					class="min-w-0 flex-1 bg-transparent text-text placeholder:text-text-placeholder focus:outline-none"
				/>
			</div>
		{/if}

		<div class={IS_ANDROID || IOS ? '' : 'px-6 pt-4'}>
			<MerchantFilters
				bind:typeFilter
				bind:statusFilter
				bind:sortBy
				sortOptions={listSortOptions}
				statusOptions={listStatusOptions}
				cardsCount={merchantsWithCards}
				vouchersCount={merchantsWithVouchers}
				giftCardsCount={merchantsWithGiftCards}
				{hasActiveFilters}
				onReset={resetFilters}
				showTypeLabel={IS_ANDROID}
				iosFlatGroups
				idPrefix="merchants-mobile"
			/>
		</div>

		{#if !IOS}
			<!-- iOS confirms with the header's "Done" text action instead (mockup). -->
			<div class={IS_ANDROID ? 'pt-4.5' : 'px-6 pb-4 pt-2'}>
				<button
					type="button"
					onclick={() => (showFilterMenu = false)}
					class={IS_ANDROID
						? 'w-full rounded-m3-full bg-accent px-4 py-3.25 text-label text-white shadow-[var(--shadow-accent)] transition-colors'
						: 'w-full btn btn-primary'}
				>
					{tr('common.done')}
				</button>
			</div>
		{/if}
	</div>
</BottomSheet>
