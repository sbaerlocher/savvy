<script lang="ts">
	import { resolve } from '$app/paths';
	import { cardsApi, vouchersApi, giftCardsApi, merchantsApi } from '$lib/api';
	import BottomSheet from '$lib/components/BottomSheet.svelte';
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
	import { categoryColors } from '$lib/utils/category-colors';

	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);

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

	const filteredMerchants = $derived.by(() => {
		let result = merchants;

		// Hide merchants with 0 items for non-admin users
		if (!isAdmin) {
			result = result.filter(
				(m) =>
					m.cards_count +
						m.cards_inactive_count +
						m.vouchers_count +
						m.vouchers_inactive_count +
						m.gift_cards_count +
						m.gift_cards_inactive_count >
					0
			);
		}

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

	const listSortOptions = $derived([
		{ value: 'name-asc', label: tr('merchantOverview.sortNameAsc') },
		{ value: 'name-desc', label: tr('merchantOverview.sortNameDesc') },
		{ value: 'items-desc', label: tr('merchantOverview.sortItemsDesc') },
		{ value: 'active-desc', label: tr('merchantOverview.sortActiveDesc') },
		{ value: 'inactive-desc', label: tr('merchantOverview.sortInactiveDesc') },
		{ value: 'balance-desc', label: tr('merchantOverview.sortBalanceDesc') }
	]);

	const listStatusOptions = $derived([
		{ value: 'all', label: tr('merchantOverview.filterStatusAll') },
		{ value: 'active', label: tr('merchantOverview.filterStatusActive') },
		{ value: 'inactive', label: tr('merchantOverview.filterStatusInactive') }
	]);
</script>

<svelte:head>
	<title>{tr('merchantOverview.title')} - {tr('common.appName')}</title>
</svelte:head>

<div class="px-4 pb-20 md:pb-4">
	<!-- Header -->
	<div class="mb-8">
		<div class="flex items-center gap-3">
			<div class="w-2 h-8 rounded-full bg-cyan-500"></div>
			<h1 class="text-3xl font-bold text-gray-900">
				{tr('merchantOverview.title')}
			</h1>
		</div>
	</div>

	{#if merchants.length > 0}
		<!-- Search + Filter + New Button -->
		<div class="flex flex-col sm:flex-row gap-3 mb-6">
			<!-- Search Bar -->
			<div class="flex-1">
				<input
					type="search"
					data-testid="merchant-search"
					bind:value={searchInput}
					placeholder={tr('common.search')}
					class="w-full px-4 py-2 bg-white border border-gray-300 rounded-md focus:ring-cyan-500 focus:border-cyan-500"
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
					class="flex items-center justify-center gap-2 h-[42px] px-4 bg-white border border-gray-300 rounded-md hover:bg-gray-50 transition-colors relative"
					title={tr('common.filter')}
					aria-label={tr('common.filter')}
					aria-expanded={showFilterMenu}
				>
					<svg
						class="w-5 h-5 text-gray-600"
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
							class="absolute -top-1 -right-1 w-3 h-3 bg-cyan-600 rounded-full"
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
					class="flex-1 flex items-center justify-center h-[42px] bg-white border border-gray-300 rounded-md hover:bg-gray-50 transition-colors relative"
					aria-label={tr('common.filter')}
					aria-expanded={showFilterMenu}
				>
					<svg
						class="w-5 h-5 text-gray-600"
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
							class="absolute -top-1 -right-1 w-3 h-3 bg-cyan-600 rounded-full"
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
	{:else if !isLoading && isAdmin}
		<div class="inline-block mb-6">
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

	<!-- Loading -->
	{#if isLoading}
		<LoadingSpinner />
	{:else if filteredMerchants.length === 0 && (searchInput || hasActiveFilters)}
		<!-- No results with filters -->
		<div class="bg-gray-50 rounded-lg p-12 text-center">
			<p class="text-gray-600 text-lg mb-4">{tr('search.no_results')}</p>
			<button type="button" onclick={resetFilters} class="btn btn-ghost">
				{tr('common.resetFilters')}
			</button>
		</div>
	{:else if filteredMerchants.length === 0}
		<!-- Empty State -->
		<div class="bg-gray-50 rounded-lg p-12 text-center">
			<p class="text-gray-600 text-lg mb-4">
				{tr('merchantOverview.noMerchants')}
			</p>
			<p class="text-gray-400 text-sm mt-1">
				{tr('merchantOverview.noMerchantsHint')}
			</p>
		</div>
	{:else}
		<!-- Grid with optional Side-Panel -->
		<div
			class="grid grid-cols-1 {showFilterMenu ? 'lg:grid-cols-3' : ''} gap-6"
		>
			<!-- Merchant Grid -->
			<div class={showFilterMenu ? 'lg:col-span-2' : ''}>
				<div
					class="grid grid-cols-1 sm:grid-cols-2 {showFilterMenu
						? ''
						: 'lg:grid-cols-3'} gap-4"
				>
					{#each filteredMerchants as merchant (merchant.id)}
						<a
							href={resolve(`/merchants/${merchant.id}`)}
							class="bg-white rounded-lg shadow hover:shadow-md transition-shadow overflow-hidden group h-full flex flex-col"
							style="border-left: 6px solid {merchant.color}"
						>
							<div class="p-4 flex flex-col flex-1">
								<!-- Merchant name -->
								<h2 class="text-lg font-semibold text-gray-900 truncate">
									{merchant.name}
								</h2>

								<!-- Active item counts -->
								{#if merchant.cards_count > 0 || merchant.vouchers_count > 0 || merchant.gift_cards_count > 0}
									<div class="mt-3 flex flex-wrap gap-2">
										{#if merchant.cards_count > 0}
											<span
												class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium {categoryColors
													.cards.badge}"
											>
												{merchant.cards_count}
												{tr('merchantOverview.cards')}
											</span>
										{/if}
										{#if merchant.vouchers_count > 0}
											<span
												class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium {categoryColors
													.vouchers.badge}"
											>
												{merchant.vouchers_count}
												{tr('merchantOverview.vouchers')}
											</span>
										{/if}
										{#if merchant.gift_cards_count > 0}
											<span
												class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium {categoryColors
													.giftCards.badge}"
											>
												{merchant.gift_cards_count}
												{tr('merchantOverview.giftCards')}
											</span>
										{/if}
									</div>
								{/if}

								<!-- Inactive/expired item counts (greyed out) -->
								{#if merchant.cards_inactive_count > 0 || merchant.vouchers_inactive_count > 0 || merchant.gift_cards_inactive_count > 0}
									<div class="mt-1.5 flex flex-wrap gap-2">
										{#if merchant.cards_inactive_count > 0}
											<span
												class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-400"
											>
												{merchant.cards_inactive_count}
												{tr('merchantOverview.cards')}
											</span>
										{/if}
										{#if merchant.vouchers_inactive_count > 0}
											<span
												class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-400"
											>
												{merchant.vouchers_inactive_count}
												{tr('merchantOverview.vouchers')}
											</span>
										{/if}
										{#if merchant.gift_cards_inactive_count > 0}
											<span
												class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-400"
											>
												{merchant.gift_cards_inactive_count}
												{tr('merchantOverview.giftCards')}
											</span>
										{/if}
									</div>
								{/if}

								<!-- Balance + total items (pushed to bottom) -->
								<div
									class="mt-auto pt-3 flex items-center justify-between text-sm text-gray-500"
								>
									<span
										>{getTotalItems(merchant)}
										{tr('merchantOverview.items')}</span
									>
									{#if merchant.total_gift_card_balance > 0}
										<span class="font-semibold text-gray-900">
											CHF {formatBalance(merchant.total_gift_card_balance)}
										</span>
									{/if}
								</div>
							</div>
						</a>
					{/each}
				</div>
			</div>

			<!-- Filter Side-Panel (Desktop only, redesigned) -->
			{#if showFilterMenu}
				<div class="hidden lg:block lg:col-span-1">
					<div
						class="bg-white rounded-xl shadow-lg sticky top-4 overflow-hidden"
					>
						<!-- Header -->
						<div class="px-5 py-4 bg-gray-50/80 border-b border-gray-100">
							<div class="flex items-center justify-between">
								<div class="flex items-center gap-2">
									<svg
										class="w-4 h-4 text-gray-500"
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
									<h3 class="text-sm font-semibold text-gray-900">
										{tr('common.filter')}
									</h3>
								</div>
								<div class="flex items-center gap-2.5">
									<span
										class="text-xs text-gray-500 bg-white px-2.5 py-1 rounded-full border border-gray-200 tabular-nums"
									>
										{tr('common.results', { count: filteredMerchants.length })}
									</span>
									<button
										type="button"
										onclick={() => (showFilterMenu = false)}
										class="text-gray-400 hover:text-gray-600 transition-colors"
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
												d="M6 18L18 6M6 6l12 12"
											></path>
										</svg>
									</button>
								</div>
							</div>
						</div>

						<div class="p-5">
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
			{/if}
		</div>
	{/if}
</div>

<!-- Mobile Filter Bottom Sheet -->
<BottomSheet
	open={showFilterMenu}
	onClose={() => (showFilterMenu = false)}
	maxHeight="80vh"
	ariaLabel={tr('common.filter')}
>
	<div class="p-6">
		<div class="flex items-center justify-between mb-4">
			<h3 class="text-lg font-semibold text-gray-900">
				{tr('common.filter')}
			</h3>
			<button
				type="button"
				onclick={() => (showFilterMenu = false)}
				class="text-gray-400 hover:text-gray-600 transition-colors"
				aria-label={tr('common.close')}
			>
				<svg
					class="w-6 h-6"
					fill="none"
					stroke="currentColor"
					viewBox="0 0 24 24"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M6 18L18 6M6 6l12 12"
					/>
				</svg>
			</button>
		</div>

		<div class="px-6 pt-4">
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
				idPrefix="merchants-mobile"
			/>
		</div>

		<div class="px-6 pb-4 pt-2">
			<button
				type="button"
				onclick={() => (showFilterMenu = false)}
				class="w-full btn btn-primary"
			>
				{tr('common.done')}
			</button>
		</div>
	</div>
</BottomSheet>
