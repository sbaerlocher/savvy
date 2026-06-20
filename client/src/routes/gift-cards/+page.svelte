<script lang="ts">
	import {
		ApiError,
		batchApi,
		giftCardsApi,
		translateBatchError
	} from '$lib/api';
	import Barcode from '$lib/components/Barcode.svelte';
	import BatchConfirmModal from '$lib/components/BatchConfirmModal.svelte';
	import BatchPanel from '$lib/components/BatchPanel.svelte';
	import BottomSheet from '$lib/components/BottomSheet.svelte';
	import ImportDialog from '$lib/components/ImportDialog.svelte';
	import { authStore } from '$lib/stores/auth';
	import { locale, t } from '$lib/stores/i18n';
	import { isOnline } from '$lib/stores/offline';
	import { toastStore } from '$lib/stores/toast';
	import type { GiftCardDTO } from '$lib/types/api';
	import { debounce } from '$lib/utils/debounce';
	import { formatCurrency } from '$lib/utils/currency';
	import { logger } from '$lib/utils/logger';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import { categoryColors } from '$lib/utils/category-colors';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { onMount } from 'svelte';
	import { SvelteSet } from 'svelte/reactivity';
	import { get } from 'svelte/store';

	// Svelte 5 compatible translation wrapper
	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);
	const pageLogger = logger.child('GiftCardsPage');
	const currentLocale = $derived($locale || 'de-DE');

	let giftCards = $state<GiftCardDTO[]>([]);
	let isLoading = $state(true);
	let searchInput = $state(''); // User input (immediate)
	let search = $state(''); // Debounced search value (used for filtering)
	let sortBy = $state('name-asc');
	let ownerFilter = $state('all');
	let merchantFilter = $state('all');
	let statusFilter = $state('active'); // Default: show only active gift cards
	let favoritesOnly = $state(false);
	let showFilterMenu = $state(false);

	// Batch selection state
	let selectMode = $state(false);
	let selectedIds = new SvelteSet<string>();
	let batchAction = $state<'delete' | 'share' | 'transfer'>('delete');
	let showBatchModal = $state(false);
	let batchLoading = $state(false);
	let showImportDialog = $state(false);

	const selectedCount = $derived(selectedIds.size);

	function toggleSelectMode() {
		selectMode = !selectMode;
		if (!selectMode) {
			selectedIds.clear();
		} else {
			showFilterMenu = false;
		}
	}

	function toggleSelection(id: string) {
		if (selectedIds.has(id)) {
			selectedIds.delete(id);
		} else {
			selectedIds.add(id);
		}
	}

	function selectAll() {
		selectedIds.clear();
		for (const gc of filteredGiftCards) {
			selectedIds.add(gc.id);
		}
	}

	function deselectAll() {
		selectedIds.clear();
	}

	function openBatchModal(action: 'delete' | 'share' | 'transfer') {
		batchAction = action;
		showBatchModal = true;
	}

	async function executeBatchAction(
		email: string,
		permissions: {
			canEdit: boolean;
			canDelete: boolean;
			canEditTransactions: boolean;
		}
	) {
		batchLoading = true;
		const ids = [...selectedIds];
		try {
			let result;
			if (batchAction === 'delete') {
				result = await batchApi.deleteGiftCards(ids);
			} else if (batchAction === 'share') {
				result = await batchApi.shareGiftCards({
					ids,
					email,
					can_edit: permissions.canEdit,
					can_delete: permissions.canDelete,
					can_edit_transactions: permissions.canEditTransactions
				});
			} else {
				result = await batchApi.transferGiftCards(ids, email);
			}

			const failed = result.failed || [];
			if (failed.length === 0) {
				const key =
					batchAction === 'delete'
						? 'batch.deleteSuccess'
						: batchAction === 'share'
							? 'batch.shareSuccess'
							: 'batch.transferSuccess';
				toastStore.success($t(key, { count: result.success_count }));
			} else if (result.success_count > 0) {
				const reason = translateBatchError(failed[0]?.error || '', $t);
				toastStore.success(
					reason
						? $t('batch.partialSuccessWithReason', {
								success: result.success_count,
								total: ids.length,
								reason
							})
						: $t('batch.partialSuccess', {
								success: result.success_count,
								total: ids.length
							})
				);
			} else {
				const reason = translateBatchError(failed[0]?.error || '', $t);
				toastStore.error(
					reason ? $t('batch.errorWithReason', { reason }) : $t('batch.error')
				);
			}

			showBatchModal = false;
			selectMode = false;
			selectedIds.clear();
			await loadGiftCards();
		} catch (err) {
			const msg =
				err instanceof ApiError
					? translateBatchError(err.message, $t)
					: $t('batch.error');
			toastStore.error(msg);
		} finally {
			batchLoading = false;
		}
	}

	async function handleBatchExport() {
		try {
			const { blob, filename } = await batchApi.exportGiftCards([
				...selectedIds
			]);
			const url = URL.createObjectURL(blob);
			const a = document.createElement('a');
			a.href = url;
			a.download = filename;
			document.body.appendChild(a);
			a.click();
			document.body.removeChild(a);
			URL.revokeObjectURL(url);
			toastStore.success(
				$t('batch.exportSuccess', { count: selectedIds.size })
			);
		} catch (err) {
			pageLogger.error('Batch export failed', { error: err });
			toastStore.error($t('batch.exportError'));
		}
	}

	// Debounced search update (300ms delay)
	const updateSearch = debounce((value: string) => {
		search = value;
	}, 300);

	const isOffline = $derived(!$isOnline);
	const hasActiveFilters = $derived(
		search !== '' ||
			ownerFilter !== 'all' ||
			merchantFilter !== 'all' ||
			statusFilter !== 'active' ||
			favoritesOnly ||
			sortBy !== 'name-asc'
	);
	const currentUserId = $derived($authStore.user?.id);
	const uniqueMerchants = $derived(() => {
		const merchants = giftCards
			.map((gc) => gc.merchant)
			.filter((m, i, arr) => m && arr.findIndex((x) => x?.id === m.id) === i)
			.sort((a, b) => (a?.name || '').localeCompare(b?.name || ''));
		return merchants;
	});

	// Reactive filtered gift cards - automatically recalculates when dependencies change
	const filteredGiftCards = $derived.by(() => {
		let result = giftCards;

		// Search filter
		if (search) {
			const query = search.toLowerCase();
			result = result.filter(
				(giftCard) =>
					giftCard.merchant?.name.toLowerCase().includes(query) ||
					giftCard.card_number.toLowerCase().includes(query) ||
					giftCard.notes?.toLowerCase().includes(query)
			);
		}

		// Owner filter
		if (ownerFilter === 'mine') {
			result = result.filter(
				(giftCard) => giftCard.owner?.id === currentUserId
			);
		}

		// Merchant filter
		if (merchantFilter !== 'all') {
			result = result.filter(
				(giftCard) => giftCard.merchant?.id === merchantFilter
			);
		}

		// Favorites filter
		if (favoritesOnly) {
			result = result.filter((giftCard) => giftCard.is_favorite);
		}

		// Status filter
		if (statusFilter !== 'all') {
			result = result.filter(
				(giftCard) => getComputedStatus(giftCard) === statusFilter
			);
		}

		// Sort
		result = [...result].sort((a, b) => {
			switch (sortBy) {
				case 'name-asc':
					return (a.merchant?.name || '').localeCompare(b.merchant?.name || '');
				case 'name-desc':
					return (b.merchant?.name || '').localeCompare(a.merchant?.name || '');
				case 'newest':
					return (
						new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
					);
				case 'oldest':
					return (
						new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
					);
				case 'balance-high':
					return b.current_balance - a.current_balance;
				case 'balance-low':
					return a.current_balance - b.current_balance;
				default:
					return 0;
			}
		});

		return result;
	});

	const sharedSelectedCount = $derived(
		filteredGiftCards.filter(
			(gc) =>
				selectedIds.has(gc.id) && gc.owner && gc.owner.id !== currentUserId
		).length
	);
	// Gift cards have permissions in list response, so check can_delete precisely
	const hasNonDeletableShared = $derived(
		filteredGiftCards.some(
			(gc) =>
				selectedIds.has(gc.id) &&
				gc.owner &&
				gc.owner.id !== currentUserId &&
				!gc.permissions?.can_delete
		)
	);

	onMount(async () => {
		loadFilters();
		await loadGiftCards();
	});

	async function loadGiftCards() {
		isLoading = true;
		try {
			// Phase 1: Load first 50 items immediately (fast initial display)
			const initialResponse = await giftCardsApi.list(1, 50);
			giftCards = initialResponse.gift_cards;
			isLoading = false; // User sees first 50 items immediately

			// Phase 2: Progressive background loading if there are more items
			if (initialResponse.pagination && initialResponse.pagination.total > 50) {
				const totalPages = initialResponse.pagination.total_pages;
				pageLogger.info(
					`[Progressive Loading] Loading ${totalPages - 1} more pages in background`
				);

				// Load remaining pages in background (100 items per page)
				for (let page = 2; page <= totalPages; page++) {
					try {
						const response = await giftCardsApi.list(page, 100);
						// Append new items to existing gift cards
						giftCards = [...giftCards, ...response.gift_cards];
						pageLogger.debug(
							`[Progressive Loading] Loaded page ${page}/${totalPages}, total gift cards: ${giftCards.length}`
						);
					} catch (err) {
						pageLogger.warn(
							`[Progressive Loading] Failed to load page ${page}:`,
							err
						);
						// Continue loading other pages even if one fails
					}
				}

				pageLogger.info(
					`[Progressive Loading] Complete. Total gift cards: ${giftCards.length}`
				);
			}
		} catch (err) {
			pageLogger.error(
				'[loadGiftCards] Failed to load initial gift cards:',
				err
			);
			toastStore.error(tr('giftCards.loadError'));
			isLoading = false;
		}
	}

	// Update search when searchInput changes (debounced)
	$effect(() => {
		updateSearch(searchInput);
	});

	// Validation helpers for localStorage filters (SVL-SEC-002)
	const VALID_SORT_VALUES = [
		'name-asc',
		'name-desc',
		'newest',
		'oldest',
		'balance-high',
		'balance-low'
	] as const;
	const VALID_OWNER_VALUES = ['all', 'mine'] as const;
	const VALID_STATUS_VALUES = ['all', 'active', 'expired', 'depleted'] as const;

	function isValidSortBy(
		value: unknown
	): value is (typeof VALID_SORT_VALUES)[number] {
		return (
			typeof value === 'string' &&
			VALID_SORT_VALUES.includes(value as (typeof VALID_SORT_VALUES)[number])
		);
	}

	function isValidOwnerFilter(
		value: unknown
	): value is (typeof VALID_OWNER_VALUES)[number] {
		return (
			typeof value === 'string' &&
			VALID_OWNER_VALUES.includes(value as (typeof VALID_OWNER_VALUES)[number])
		);
	}

	function isValidStatusFilter(
		value: unknown
	): value is (typeof VALID_STATUS_VALUES)[number] {
		return (
			typeof value === 'string' &&
			VALID_STATUS_VALUES.includes(
				value as (typeof VALID_STATUS_VALUES)[number]
			)
		);
	}

	function isValidSearch(value: unknown): boolean {
		return typeof value === 'string' && value.length <= 200;
	}

	function loadFilters() {
		try {
			const saved = localStorage.getItem('savvy_gift_cards_filters');
			if (saved) {
				const filters = JSON.parse(saved);

				// Validate and apply filters with defaults
				const savedSearch = isValidSearch(filters.search) ? filters.search : '';
				searchInput = savedSearch;
				search = savedSearch;

				sortBy = isValidSortBy(filters.sortBy) ? filters.sortBy : 'name-asc';
				ownerFilter = isValidOwnerFilter(filters.ownerFilter)
					? filters.ownerFilter
					: 'all';
				merchantFilter = filters.merchantFilter || 'all'; // merchantFilter is dynamic (merchant IDs)
				favoritesOnly = filters.favoritesOnly === true;
				statusFilter = isValidStatusFilter(filters.statusFilter)
					? filters.statusFilter
					: 'active';

				// Log if validation failed (indicates potential tampering)
				if (
					!isValidSearch(filters.search) ||
					!isValidSortBy(filters.sortBy) ||
					!isValidOwnerFilter(filters.ownerFilter) ||
					!isValidStatusFilter(filters.statusFilter)
				) {
					pageLogger.warn(
						'Invalid filter data in localStorage, using defaults'
					);
				}
			}
		} catch (e) {
			pageLogger.error('Failed to load filters:', e);
			// Clear invalid localStorage on parse error
			localStorage.removeItem('savvy_gift_cards_filters');
		}
	}

	// Compute status based on balance and expiration date
	function getComputedStatus(giftCard: GiftCardDTO): string {
		if (giftCard.current_balance === 0) return 'depleted';
		if (giftCard.expires_at && new Date(giftCard.expires_at) < new Date())
			return 'expired';
		return 'active';
	}

	function getStatusBadge(status: string): { class: string; text: string } {
		switch (status) {
			case 'expired':
				return {
					class: 'bg-red-100 text-red-800',
					text: tr('giftCards.status.expired')
				};
			case 'depleted':
				return {
					class: 'bg-orange-100 text-orange-800',
					text: tr('giftCards.status.depleted')
				};
			default:
				return { class: '', text: '' };
		}
	}

	// Save filters to localStorage when they change
	$effect(() => {
		try {
			const filters = {
				search,
				sortBy,
				ownerFilter,
				merchantFilter,
				favoritesOnly,
				statusFilter
			};
			localStorage.setItem('savvy_gift_cards_filters', JSON.stringify(filters));
		} catch (e) {
			pageLogger.error('Failed to save filters:', e);
		}
	});
</script>

<svelte:head>
	<title>{tr('giftCards.title')} - {tr('common.appName')}</title>
</svelte:head>

<BatchConfirmModal
	action={batchAction}
	count={selectedCount}
	isOpen={showBatchModal}
	isLoading={batchLoading}
	onConfirm={executeBatchAction}
	onCancel={() => (showBatchModal = false)}
	showTransactionPermission={true}
/>

<ImportDialog
	isOpen={showImportDialog}
	onClose={() => (showImportDialog = false)}
	onImported={loadGiftCards}
	defaultResourceType="gift-cards"
/>

<div class="px-4 max-w-7xl mx-auto" class:pb-40={selectMode}>
	<div class="mb-8">
		<div class="flex items-center gap-3">
			<div class="w-2 h-8 rounded-full {categoryColors.giftCards.accent}"></div>
			<h1 class="text-3xl font-bold text-gray-900">{tr('giftCards.title')}</h1>
		</div>
	</div>

	{#if giftCards.length > 0}
		<!-- Search bar and action buttons -->
		<div class="flex flex-col sm:flex-row gap-3 mb-6">
			<!-- Search Bar -->
			<div class="flex-1">
				<input
					type="text"
					bind:value={searchInput}
					placeholder={tr('common.search')}
					class="input bg-white"
				/>
			</div>

			<!-- Action Buttons (Desktop) -->
			<div class="hidden sm:flex gap-3">
				<!-- Select Mode Button -->
				<button
					type="button"
					onclick={toggleSelectMode}
					disabled={isOffline}
					class="btn btn-ghost {selectMode
						? 'ring-2 ring-cyan-500 border-cyan-500'
						: ''}"
					title={tr('batch.selectMode')}
					aria-label={tr('batch.selectMode')}
					aria-pressed={selectMode}
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
							d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"
						></path>
					</svg>
				</button>
				<!-- Filter Button -->
				<button
					type="button"
					onclick={(e) => {
						e.stopPropagation();
						showFilterMenu = !showFilterMenu;
					}}
					class="relative btn btn-ghost"
					title={tr('common.filter')}
					aria-label={tr('common.filterGiftCards')}
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

				<!-- Import Button (desktop only) -->
				<button
					type="button"
					onclick={() => (showImportDialog = true)}
					disabled={isOffline}
					class="hidden sm:inline-flex btn btn-ghost whitespace-nowrap items-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
					aria-label={tr('settings.import.title')}
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
							d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12"
						/>
					</svg>
				</button>

				<!-- New Gift Card Button -->
				<a
					href={resolve('/gift-cards/new')}
					data-testid="new-gift-card-btn-desktop"
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
					{tr('giftCards.newGiftCard')}
				</a>
			</div>

			<!-- Action Buttons (Mobile) -->
			<div class="flex sm:hidden gap-3">
				<!-- Select Mode Button (Mobile) -->
				<button
					type="button"
					onclick={toggleSelectMode}
					disabled={isOffline}
					class="flex-1 btn btn-ghost {selectMode
						? 'ring-2 ring-cyan-500 border-cyan-500'
						: ''}"
					aria-label={tr('batch.selectMode')}
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
							d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"
						></path>
					</svg>
				</button>
				<!-- Filter Button (Mobile) -->
				<button
					type="button"
					onclick={(e) => {
						pageLogger.debug('Mobile filter button clicked');
						e.stopPropagation();
						showFilterMenu = !showFilterMenu;
						pageLogger.debug('showFilterMenu is now:', showFilterMenu);
					}}
					class="relative flex-1 btn btn-ghost"
					aria-label={tr('common.filterGiftCards')}
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

				<!-- New Gift Card Button (Mobile) -->
				<a
					href={resolve('/gift-cards/new')}
					data-testid="new-gift-card-btn-mobile"
					onclick={(e) => {
						if (isOffline) e.preventDefault();
					}}
					class="btn btn-sm btn-primary flex-[2] text-center flex items-center justify-center gap-2 {isOffline
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
					<span>{tr('common.new')}</span>
				</a>
			</div>
		</div>
	{:else if !isLoading}
		<div class="inline-block mb-6">
			<a
				href={resolve('/gift-cards/new')}
				data-testid="new-gift-card-btn-header"
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
				{tr('giftCards.newGiftCard')}
			</a>
		</div>
	{/if}

	{#if isLoading}
		<LoadingSpinner />
	{:else if giftCards.length === 0}
		<div class="bg-gray-50 rounded-lg p-12 text-center">
			<p class="text-gray-600 text-lg mb-4">{tr('giftCards.noGiftCards')}</p>
			<div class="inline-block">
				<a
					href={resolve('/gift-cards/new')}
					data-testid="new-gift-card-btn-empty"
					onclick={(e) => {
						if (isOffline) e.preventDefault();
					}}
					class="btn btn-text flex items-center gap-2 {isOffline
						? 'opacity-50 cursor-not-allowed pointer-events-none'
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
					{tr('giftCards.noGiftCardsHint')}
				</a>
			</div>
		</div>
	{:else if filteredGiftCards.length === 0}
		<div class="bg-gray-50 rounded-lg p-12 text-center">
			<p class="text-gray-600 text-lg mb-4">
				{tr('giftCards.noGiftCardsFound')}
			</p>
			{#if hasActiveFilters}
				<button
					type="button"
					onclick={() => {
						search = '';
						sortBy = 'name-asc';
						ownerFilter = 'all';
						merchantFilter = 'all';
						favoritesOnly = false;
						statusFilter = 'active';
					}}
					class="btn btn-ghost"
				>
					{tr('common.resetFilters')}
				</button>
			{/if}
		</div>
	{:else}
		<!-- Grid with optional Side-Panel -->
		<div
			class="grid grid-cols-1 {showFilterMenu || selectMode
				? 'lg:grid-cols-3'
				: ''} gap-6"
		>
			<!-- Gift Cards Grid (2/3 when filter or batch panel is open on desktop) -->
			<div class={showFilterMenu || selectMode ? 'lg:col-span-2' : ''}>
				<div
					class="grid grid-cols-1 md:grid-cols-2 {showFilterMenu || selectMode
						? ''
						: 'lg:grid-cols-3'} gap-6"
				>
					{#each filteredGiftCards as giftCard (giftCard.id)}
						{@const computedStatus = getComputedStatus(giftCard)}
						<div
							class="block bg-white rounded-lg shadow-lg hover:shadow-xl transition overflow-hidden relative {selectMode &&
							selectedIds.has(giftCard.id)
								? 'ring-2 ring-cyan-500'
								: ''}"
							style="border-left: 6px solid {giftCard.merchant?.color ||
								'#6B7280'}"
							role="button"
							tabindex="0"
							data-owner={giftCard.owner && giftCard.owner.id !== currentUserId
								? 'shared'
								: 'owned'}
							onclick={() => {
								if (selectMode) {
									toggleSelection(giftCard.id);
								} else {
									goto(resolve(`/gift-cards/${giftCard.id}`));
								}
							}}
							onkeydown={(e) => {
								if (e.key === 'Enter' || e.key === ' ') {
									e.preventDefault();
									if (selectMode) {
										toggleSelection(giftCard.id);
									} else {
										goto(resolve(`/gift-cards/${giftCard.id}`));
									}
								}
							}}
						>
							<div
								class="p-6 flex flex-col h-full {computedStatus !== 'active'
									? 'opacity-50 grayscale'
									: ''}"
							>
								<!-- Status Overlay -->
								{#if computedStatus !== 'active'}
									{@const badge = getStatusBadge(computedStatus)}
									<div
										class="absolute inset-0 flex items-center justify-center z-10 pointer-events-none"
									>
										<span
											class="px-4 py-1.5 text-sm font-semibold rounded-full {badge.class} shadow-sm"
										>
											{badge.text}
										</span>
									</div>
								{/if}

								<!-- Balance + Merchant/Owner -->
								<div
									class="grid grid-cols-[auto_1fr] gap-x-4 items-center mb-4"
								>
									<p
										class="text-2xl font-bold row-span-2"
										style="color: {giftCard.merchant?.color || '#10B981'};"
									>
										{formatCurrency(
											giftCard.current_balance,
											giftCard.currency,
											$locale
										)}
									</p>
									<p class="text-sm text-gray-500 truncate text-right">
										{giftCard.merchant?.name || tr('giftCards.title')}
									</p>
									<p class="text-xs text-gray-400 text-right">
										{#if giftCard.owner && giftCard.owner.id !== currentUserId}
											{tr('giftCards.sharedBy', {
												name: giftCard.owner.first_name || giftCard.owner.email
											})}
										{:else if giftCard.shared_with_count > 0}
											{tr('giftCards.sharedWithCount', {
												count: String(giftCard.shared_with_count)
											})}
										{:else}
											{tr('giftCards.sharedWithNone')}
										{/if}
									</p>
								</div>

								<!-- Barcode -->
								<div class="mt-auto">
									<div
										class="bg-gray-50 rounded-lg p-4 border border-gray-200 h-[120px] flex flex-col justify-center"
									>
										<div class="flex justify-center mb-2">
											<Barcode
												value={giftCard.card_number}
												type={giftCard.barcode_type || 'CODE128'}
												height={64}
												maxHeight={64}
											/>
										</div>
										<p
											class="text-center text-xs text-gray-600 font-mono break-all"
										>
											{giftCard.card_number}
										</p>
									</div>

									<!-- Footer -->
									<p class="text-xs text-gray-500 truncate mt-3">
										{#if giftCard.expires_at}
											{tr('giftCards.expiresAtLabel')}
											{new Date(
												giftCard.expires_at.split('T')[0]
											).toLocaleDateString(currentLocale)}
										{:else if giftCard.notes}
											{giftCard.notes}
										{:else}
											&nbsp;
										{/if}
									</p>
								</div>
							</div>
						</div>
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
										{tr('common.results', { count: filteredGiftCards.length })}
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
							<!-- Owner -->
							<div class="pb-4">
								<label
									class="block text-xs font-medium text-gray-500 uppercase tracking-wider mb-2"
									for="owner-filter-desktop-giftcards"
								>
									{tr('giftCards.owner')}
								</label>
								<select
									id="owner-filter-desktop-giftcards"
									bind:value={ownerFilter}
									class="w-full bg-gray-50 border border-gray-200 rounded-lg px-3 py-2 text-sm text-gray-700 focus:ring-2 focus:ring-cyan-500 focus:border-cyan-500 transition-colors"
								>
									<option value="all">{tr('giftCards.allGiftCards')}</option>
									<option value="mine">{tr('giftCards.myGiftCards')}</option>
								</select>
							</div>

							<div class="border-t border-gray-100"></div>

							<!-- Merchant -->
							<div class="py-4">
								<label
									class="block text-xs font-medium text-gray-500 uppercase tracking-wider mb-2"
									for="merchant-filter-desktop-giftcards"
								>
									{tr('giftCards.merchant')}
								</label>
								<select
									id="merchant-filter-desktop-giftcards"
									bind:value={merchantFilter}
									class="w-full bg-gray-50 border border-gray-200 rounded-lg px-3 py-2 text-sm text-gray-700 focus:ring-2 focus:ring-cyan-500 focus:border-cyan-500 transition-colors"
								>
									<option value="all">{tr('giftCards.allMerchants')}</option>
									{#each uniqueMerchants() as merchant (merchant?.id)}
										{#if merchant}
											<option value={merchant.id}>{merchant.name}</option>
										{/if}
									{/each}
								</select>
							</div>

							<div class="border-t border-gray-100"></div>

							<!-- Favorites Toggle -->
							<div class="py-4">
								<button
									type="button"
									role="switch"
									aria-checked={favoritesOnly}
									onclick={() => (favoritesOnly = !favoritesOnly)}
									class="flex items-center justify-between w-full cursor-pointer group"
								>
									<div class="flex items-center gap-2">
										<svg
											class="w-4 h-4 transition-colors {favoritesOnly
												? 'text-amber-500'
												: 'text-gray-400 group-hover:text-amber-400'}"
											fill={favoritesOnly ? 'currentColor' : 'none'}
											stroke="currentColor"
											viewBox="0 0 24 24"
										>
											<path
												stroke-linecap="round"
												stroke-linejoin="round"
												stroke-width="2"
												d="M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z"
											></path>
										</svg>
										<span class="text-sm font-medium text-gray-700"
											>{tr('common.favoritesOnly')}</span
										>
									</div>
									<div
										class="relative w-9 h-5 rounded-full transition-colors {favoritesOnly
											? 'bg-cyan-600'
											: 'bg-gray-200'}"
									>
										<div
											class="absolute top-0.5 left-0.5 w-4 h-4 rounded-full bg-white shadow-sm transition-transform {favoritesOnly
												? 'translate-x-4'
												: 'translate-x-0'}"
										></div>
									</div>
								</button>
							</div>

							<div class="border-t border-gray-100"></div>

							<!-- Status -->
							<div class="py-4">
								<label
									class="block text-xs font-medium text-gray-500 uppercase tracking-wider mb-2"
									for="status-filter-desktop-giftcards"
								>
									{tr('giftCards.statusLabel')}
								</label>
								<select
									id="status-filter-desktop-giftcards"
									bind:value={statusFilter}
									class="w-full bg-gray-50 border border-gray-200 rounded-lg px-3 py-2 text-sm text-gray-700 focus:ring-2 focus:ring-cyan-500 focus:border-cyan-500 transition-colors"
								>
									<option value="all">{tr('giftCards.allStatuses')}</option>
									<option value="active">{tr('giftCards.status.active')}</option
									>
									<option value="expired"
										>{tr('giftCards.status.expired')}</option
									>
									<option value="depleted"
										>{tr('giftCards.status.depleted')}</option
									>
								</select>
							</div>

							<div class="border-t border-gray-100"></div>

							<!-- Sort -->
							<div class="pt-4">
								<label
									class="block text-xs font-medium text-gray-500 uppercase tracking-wider mb-2"
									for="sort-filter-desktop-giftcards"
								>
									{tr('giftCards.sortBy')}
								</label>
								<select
									id="sort-filter-desktop-giftcards"
									bind:value={sortBy}
									class="w-full bg-gray-50 border border-gray-200 rounded-lg px-3 py-2 text-sm text-gray-700 focus:ring-2 focus:ring-cyan-500 focus:border-cyan-500 transition-colors"
								>
									<option value="name-asc">{tr('giftCards.sortNameAsc')}</option
									>
									<option value="name-desc"
										>{tr('giftCards.sortNameDesc')}</option
									>
									<option value="newest">{tr('giftCards.sortNewest')}</option>
									<option value="oldest">{tr('giftCards.sortOldest')}</option>
									<option value="balance-high"
										>{tr('giftCards.sortBalanceHigh')}</option
									>
									<option value="balance-low"
										>{tr('giftCards.sortBalanceLow')}</option
									>
								</select>
							</div>

							{#if hasActiveFilters}
								<div class="border-t border-gray-100 mt-4"></div>
								<div class="pt-4">
									<button
										type="button"
										onclick={() => {
											search = '';
											searchInput = '';
											sortBy = 'name-asc';
											ownerFilter = 'all';
											merchantFilter = 'all';
											favoritesOnly = false;
											statusFilter = 'active';
										}}
										class="w-full text-sm text-gray-500 hover:text-gray-700 py-2 rounded-lg hover:bg-gray-50 transition-colors flex items-center justify-center gap-1.5"
									>
										<svg
											class="w-3.5 h-3.5"
											fill="none"
											stroke="currentColor"
											viewBox="0 0 24 24"
										>
											<path
												stroke-linecap="round"
												stroke-linejoin="round"
												stroke-width="2"
												d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
											></path>
										</svg>
										{tr('common.resetFilters')}
									</button>
								</div>
							{/if}
						</div>
					</div>
				</div>
			{/if}

			<!-- Batch Actions Side-Panel (Desktop only) -->
			{#if selectMode}
				<BatchPanel
					{selectedCount}
					totalCount={filteredGiftCards.length}
					{sharedSelectedCount}
					{hasNonDeletableShared}
					onSelectAll={selectAll}
					onDeselectAll={deselectAll}
					onDelete={() => openBatchModal('delete')}
					onShare={() => openBatchModal('share')}
					onTransfer={() => openBatchModal('transfer')}
					onExport={handleBatchExport}
					onCancel={toggleSelectMode}
				/>
			{/if}
		</div>

		<!-- Mobile Filter Bottom Sheet -->
		<BottomSheet
			open={showFilterMenu}
			onClose={() => (showFilterMenu = false)}
			ariaLabel={tr('common.filterGiftCards')}
		>
			<!-- Header -->
			<div class="px-6 pb-3 flex items-center justify-between">
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
						class="text-xs text-gray-500 bg-gray-100 px-2.5 py-1 rounded-full tabular-nums"
					>
						{tr('common.results', { count: filteredGiftCards.length })}
					</span>
					<button
						type="button"
						onclick={() => (showFilterMenu = false)}
						class="text-gray-400 hover:text-gray-600 transition-colors"
						aria-label={tr('common.closeFilters')}
					>
						<svg
							class="w-5 h-5"
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

			<div class="border-t border-gray-100"></div>

			<div class="px-6 pt-4">
				<!-- Owner -->
				<div class="pb-4">
					<label
						class="block text-xs font-medium text-gray-500 uppercase tracking-wider mb-2"
						for="owner-filter-mobile-giftcards"
					>
						{tr('giftCards.owner')}
					</label>
					<select
						id="owner-filter-mobile-giftcards"
						bind:value={ownerFilter}
						class="w-full bg-gray-50 border border-gray-200 rounded-lg px-3 py-2.5 text-sm text-gray-700 focus:ring-2 focus:ring-cyan-500 focus:border-cyan-500"
					>
						<option value="all">{tr('giftCards.allGiftCards')}</option>
						<option value="mine">{tr('giftCards.myGiftCards')}</option>
					</select>
				</div>

				<div class="border-t border-gray-100"></div>

				<!-- Merchant -->
				<div class="py-4">
					<label
						class="block text-xs font-medium text-gray-500 uppercase tracking-wider mb-2"
						for="merchant-filter-mobile-giftcards"
					>
						{tr('giftCards.merchant')}
					</label>
					<select
						id="merchant-filter-mobile-giftcards"
						bind:value={merchantFilter}
						class="w-full bg-gray-50 border border-gray-200 rounded-lg px-3 py-2.5 text-sm text-gray-700 focus:ring-2 focus:ring-cyan-500 focus:border-cyan-500"
					>
						<option value="all">{tr('giftCards.allMerchants')}</option>
						{#each uniqueMerchants() as merchant (merchant?.id)}
							{#if merchant}
								<option value={merchant.id}>{merchant.name}</option>
							{/if}
						{/each}
					</select>
				</div>

				<div class="border-t border-gray-100"></div>

				<!-- Favorites Toggle -->
				<div class="py-4">
					<button
						type="button"
						role="switch"
						aria-checked={favoritesOnly}
						onclick={() => (favoritesOnly = !favoritesOnly)}
						class="flex items-center justify-between w-full cursor-pointer group"
					>
						<div class="flex items-center gap-2">
							<svg
								class="w-4 h-4 transition-colors {favoritesOnly
									? 'text-amber-500'
									: 'text-gray-400 group-hover:text-amber-400'}"
								fill={favoritesOnly ? 'currentColor' : 'none'}
								stroke="currentColor"
								viewBox="0 0 24 24"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z"
								></path>
							</svg>
							<span class="text-sm font-medium text-gray-700"
								>{tr('common.favoritesOnly')}</span
							>
						</div>
						<div
							class="relative w-9 h-5 rounded-full transition-colors {favoritesOnly
								? 'bg-cyan-600'
								: 'bg-gray-200'}"
						>
							<div
								class="absolute top-0.5 left-0.5 w-4 h-4 rounded-full bg-white shadow-sm transition-transform {favoritesOnly
									? 'translate-x-4'
									: 'translate-x-0'}"
							></div>
						</div>
					</button>
				</div>

				<div class="border-t border-gray-100"></div>

				<!-- Status -->
				<div class="py-4">
					<label
						class="block text-xs font-medium text-gray-500 uppercase tracking-wider mb-2"
						for="status-filter-mobile-giftcards"
					>
						{tr('giftCards.statusLabel')}
					</label>
					<select
						id="status-filter-mobile-giftcards"
						bind:value={statusFilter}
						class="w-full bg-gray-50 border border-gray-200 rounded-lg px-3 py-2.5 text-sm text-gray-700 focus:ring-2 focus:ring-cyan-500 focus:border-cyan-500"
					>
						<option value="all">{tr('giftCards.allStatuses')}</option>
						<option value="active">{tr('giftCards.status.active')}</option>
						<option value="expired">{tr('giftCards.status.expired')}</option>
						<option value="depleted">{tr('giftCards.status.depleted')}</option>
					</select>
				</div>

				<div class="border-t border-gray-100"></div>

				<!-- Sort -->
				<div class="py-4">
					<label
						class="block text-xs font-medium text-gray-500 uppercase tracking-wider mb-2"
						for="sort-filter-mobile-giftcards"
					>
						{tr('giftCards.sortBy')}
					</label>
					<select
						id="sort-filter-mobile-giftcards"
						bind:value={sortBy}
						class="w-full bg-gray-50 border border-gray-200 rounded-lg px-3 py-2.5 text-sm text-gray-700 focus:ring-2 focus:ring-cyan-500 focus:border-cyan-500"
					>
						<option value="name-asc">{tr('giftCards.sortNameAsc')}</option>
						<option value="name-desc">{tr('giftCards.sortNameDesc')}</option>
						<option value="newest">{tr('giftCards.sortNewest')}</option>
						<option value="oldest">{tr('giftCards.sortOldest')}</option>
						<option value="balance-high"
							>{tr('giftCards.sortBalanceHigh')}</option
						>
						<option value="balance-low">{tr('giftCards.sortBalanceLow')}</option
						>
					</select>
				</div>

				{#if hasActiveFilters}
					<div class="border-t border-gray-100"></div>
					<div class="py-3">
						<button
							type="button"
							onclick={() => {
								search = '';
								searchInput = '';
								sortBy = 'name-asc';
								ownerFilter = 'all';
								merchantFilter = 'all';
								favoritesOnly = false;
								statusFilter = 'active';
							}}
							class="w-full text-sm text-gray-500 hover:text-gray-700 py-2 rounded-lg hover:bg-gray-50 transition-colors flex items-center justify-center gap-1.5"
						>
							<svg
								class="w-3.5 h-3.5"
								fill="none"
								stroke="currentColor"
								viewBox="0 0 24 24"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
								></path>
							</svg>
							{tr('common.resetFilters')}
						</button>
					</div>
				{/if}

				<div class="pb-2 pt-1">
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
	{/if}
</div>
