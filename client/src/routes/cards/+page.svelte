<script lang="ts">
	import { ApiError, batchApi, cardsApi, translateBatchError } from '$lib/api';
	import Barcode from '$lib/components/Barcode.svelte';
	import BatchConfirmModal from '$lib/components/BatchConfirmModal.svelte';
	import BatchPanel from '$lib/components/BatchPanel.svelte';
	import BottomSheet from '$lib/components/BottomSheet.svelte';
	import ImportDialog from '$lib/components/ImportDialog.svelte';
	import { authStore } from '$lib/stores/auth';
	import { t } from '$lib/stores/i18n';
	import { isOnline } from '$lib/stores/offline';
	import { toastStore } from '$lib/stores/toast';
	import type { CardDTO } from '$lib/types/api';
	import { debounce } from '$lib/utils/debounce';
	import { logger } from '$lib/utils/logger';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import { categoryColors } from '$lib/utils/category-colors';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { get } from 'svelte/store';

	// Svelte 5 compatible translation wrapper
	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);
	const pageLogger = logger.child('CardsPage');

	let cards = $state<CardDTO[]>([]);
	let isLoading = $state(true);
	let isLoadingMore = $state(false); // Progressive background loading indicator
	let searchInput = $state(''); // User input (immediate)
	let search = $state(''); // Debounced search value (used for filtering)
	let sortBy = $state('name-asc');
	let ownerFilter = $state('all');
	let merchantFilter = $state('all');
	let statusFilter = $state('active');
	let favoritesOnly = $state(false);
	let showFilterMenu = $state(false);
	let dialogRef = $state<HTMLDivElement | undefined>(undefined);
	let filterApplied = $state(false);

	// Batch selection state
	let selectMode = $state(false);
	let selectedIds = $state<Set<string>>(new Set());
	let batchAction = $state<'delete' | 'share' | 'transfer'>('delete');
	let showBatchModal = $state(false);
	let batchLoading = $state(false);
	let showImportDialog = $state(false);

	const selectedCount = $derived(selectedIds.size);

	function toggleSelectMode() {
		selectMode = !selectMode;
		if (!selectMode) {
			selectedIds = new Set();
		} else {
			showFilterMenu = false;
		}
	}

	function toggleSelection(id: string) {
		const next = new Set(selectedIds);
		if (next.has(id)) {
			next.delete(id);
		} else {
			next.add(id);
		}
		selectedIds = next;
	}

	function selectAll() {
		selectedIds = new Set(filteredCards.map((c) => c.id));
	}

	function deselectAll() {
		selectedIds = new Set();
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
				result = await batchApi.deleteCards(ids);
			} else if (batchAction === 'share') {
				result = await batchApi.shareCards({
					ids,
					email,
					can_edit: permissions.canEdit,
					can_delete: permissions.canDelete
				});
			} else {
				result = await batchApi.transferCards(ids, email);
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
			selectedIds = new Set();
			await loadCards();
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
			const { blob, filename } = await batchApi.exportCards([...selectedIds]);
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
		const merchants = cards
			.map((c) => c.merchant)
			.filter((m, i, arr) => m && arr.findIndex((x) => x?.id === m.id) === i)
			.sort((a, b) => (a?.name || '').localeCompare(b?.name || ''));
		return merchants;
	});

	// Reactive filtering and sorting using $derived.by for optimal performance
	const filteredCards = $derived.by(() => {
		let result = cards;

		// Search filter
		if (search) {
			const query = search.toLowerCase();
			result = result.filter(
				(card) =>
					card.merchant?.name.toLowerCase().includes(query) ||
					card.card_number.toLowerCase().includes(query) ||
					card.notes?.toLowerCase().includes(query)
			);
		}

		// Owner filter
		if (ownerFilter === 'mine') {
			result = result.filter((card) => card.owner?.id === currentUserId);
		}

		// Merchant filter
		if (merchantFilter !== 'all') {
			result = result.filter((card) => card.merchant?.id === merchantFilter);
		}

		// Favorites filter
		if (favoritesOnly) {
			result = result.filter((card) => card.is_favorite);
		}

		// Status filter
		if (statusFilter === 'active') {
			result = result.filter((card) => card.status === 'active');
		} else if (statusFilter === 'inactive') {
			result = result.filter((card) => card.status !== 'active');
		}

		// Sort
		return [...result].sort((a, b) => {
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
				default:
					return 0;
			}
		});
	});

	const sharedSelectedCount = $derived(
		filteredCards.filter(
			(c) => selectedIds.has(c.id) && c.owner && c.owner.id !== currentUserId
		).length
	);
	// Cards list doesn't include permissions, so conservatively disable delete for all shared items
	const hasNonDeletableShared = $derived(sharedSelectedCount > 0);

	onMount(async () => {
		loadFilters();
		await loadCards();
	});

	async function loadCards() {
		isLoading = true;
		try {
			// Phase 1: Load first 50 items immediately (fast initial display)
			const initialResponse = await cardsApi.list(1, 50);
			cards = initialResponse.cards;
			isLoading = false; // User sees first 50 items immediately

			// Phase 2: Progressive background loading if there are more items
			if (initialResponse.pagination && initialResponse.pagination.total > 50) {
				const totalPages = initialResponse.pagination.total_pages;
				pageLogger.info(
					`[Progressive Loading] Loading ${totalPages - 1} more pages in background`
				);

				isLoadingMore = true;

				// Load remaining pages in background (100 items per page)
				for (let page = 2; page <= totalPages; page++) {
					try {
						const response = await cardsApi.list(page, 100);
						// Append new items to existing cards
						cards = [...cards, ...response.cards];
						pageLogger.debug(
							`[Progressive Loading] Loaded page ${page}/${totalPages}, total cards: ${cards.length}`
						);
					} catch (err) {
						pageLogger.warn(
							`[Progressive Loading] Failed to load page ${page}:`,
							err
						);
						// Continue loading other pages even if one fails
					}
				}

				isLoadingMore = false;
				pageLogger.info(
					`[Progressive Loading] Complete. Total cards: ${cards.length}`
				);
			}
		} catch (err) {
			pageLogger.error('[loadCards] Failed to load initial cards:', err);
			toastStore.error($t('common.error'));
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
		'oldest'
	] as const;
	const VALID_OWNER_VALUES = ['all', 'mine'] as const;
	const VALID_STATUS_VALUES = ['active', 'inactive', 'all-status'] as const;

	function isValidSortBy(
		value: unknown
	): value is (typeof VALID_SORT_VALUES)[number] {
		return (
			typeof value === 'string' && VALID_SORT_VALUES.includes(value as any)
		);
	}

	function isValidOwnerFilter(
		value: unknown
	): value is (typeof VALID_OWNER_VALUES)[number] {
		return (
			typeof value === 'string' && VALID_OWNER_VALUES.includes(value as any)
		);
	}

	function isValidStatusFilter(
		value: unknown
	): value is (typeof VALID_STATUS_VALUES)[number] {
		return (
			typeof value === 'string' && VALID_STATUS_VALUES.includes(value as any)
		);
	}

	function isValidSearch(value: unknown): boolean {
		return typeof value === 'string' && value.length <= 200;
	}

	function loadFilters() {
		try {
			const saved = localStorage.getItem('savvy_cards_filters');
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
				merchantFilter = filters.merchantFilter || 'all';
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
			localStorage.removeItem('savvy_cards_filters');
		}
	}

	// Save filters to localStorage whenever they change
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
			localStorage.setItem('savvy_cards_filters', JSON.stringify(filters));
		} catch (e) {
			pageLogger.error('Failed to save filters:', e);
		}
	});

	// Focus trap for filter modal (SVL-A11Y-001)
	$effect(() => {
		if (showFilterMenu && dialogRef) {
			const focusableElements = dialogRef.querySelectorAll<HTMLElement>(
				'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
			);
			const firstElement = focusableElements[0];
			const lastElement = focusableElements[focusableElements.length - 1];

			// Focus first element
			firstElement?.focus();

			const handleKeydown = (e: KeyboardEvent) => {
				if (e.key === 'Tab') {
					if (e.shiftKey && document.activeElement === firstElement) {
						e.preventDefault();
						lastElement?.focus();
					} else if (!e.shiftKey && document.activeElement === lastElement) {
						e.preventDefault();
						firstElement?.focus();
					}
				}
			};

			document.addEventListener('keydown', handleKeydown);
			return () => document.removeEventListener('keydown', handleKeydown);
		}
	});

	async function toggleFavorite(cardId: string) {
		try {
			const response = await cardsApi.toggleFavorite(cardId);
			cards = cards.map((c) =>
				c.id === cardId ? { ...c, is_favorite: response.is_favorite } : c
			);
		} catch (err) {
			toastStore.error($t('common.error'));
		}
	}

	function getStatusBadge(status: string): { class: string; text: string } {
		switch (status) {
			case 'inactive':
				return {
					class: 'bg-gray-100 text-gray-800',
					text: $t('cards.status.inactive')
				};
			case 'expired':
				return {
					class: 'bg-red-100 text-red-800',
					text: $t('cards.status.expired')
				};
			case 'lost':
				return {
					class: 'bg-orange-100 text-orange-800',
					text: $t('cards.status.lost')
				};
			case 'blocked':
				return {
					class: 'bg-red-100 text-red-800',
					text: $t('cards.status.blocked')
				};
			default:
				return { class: '', text: '' };
		}
	}
</script>

<svelte:head>
	<title>{tr('cards.title')} - {tr('common.appName')}</title>
</svelte:head>

<BatchConfirmModal
	action={batchAction}
	count={selectedCount}
	isOpen={showBatchModal}
	isLoading={batchLoading}
	onConfirm={executeBatchAction}
	onCancel={() => (showBatchModal = false)}
/>

<ImportDialog
	isOpen={showImportDialog}
	onClose={() => (showImportDialog = false)}
	onImported={loadCards}
	defaultResourceType="cards"
/>

<div class="px-4 max-w-7xl mx-auto" class:pb-40={selectMode}>
	<div class="mb-8">
		<div class="flex items-center gap-3">
			<div class="w-2 h-8 rounded-full {categoryColors.cards.accent}"></div>
			<h1 class="text-3xl font-bold text-gray-900">{tr('cards.title')}</h1>
		</div>
	</div>

	{#if cards.length > 0}
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
						{
							pageLogger.debug('Toggling filter, current:', showFilterMenu);
							showFilterMenu = !showFilterMenu;
							pageLogger.debug('New state:', showFilterMenu);
						}
					}}
					class="relative btn btn-ghost"
					title={tr('common.filter')}
					aria-label={tr('common.filterCards')}
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

				<!-- New Card Button -->
				<a
					href="/cards/new"
					data-testid="new-card-btn-desktop"
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
					{tr('cards.newCard')}
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
						e.stopPropagation();
						{
							pageLogger.debug('Toggling filter, current:', showFilterMenu);
							showFilterMenu = !showFilterMenu;
							pageLogger.debug('New state:', showFilterMenu);
						}
					}}
					class="relative flex-1 btn btn-ghost"
					aria-label={tr('common.filterCards')}
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

				<!-- New Card Button (Mobile) -->
				<a
					href="/cards/new"
					data-testid="new-card-btn-mobile"
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
				href="/cards/new"
				data-testid="new-card-btn-header"
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
				{tr('cards.newCard')}
			</a>
		</div>
	{/if}

	{#if isLoading}
		<LoadingSpinner />
	{:else if cards.length === 0}
		<div class="bg-gray-50 rounded-lg p-12 text-center">
			<p class="text-gray-600 text-lg mb-4">{tr('cards.noCards')}</p>
			<div class="inline-block">
				<a
					href="/cards/new"
					data-testid="new-card-btn-empty"
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
					{tr('cards.noCardsHint')}
				</a>
			</div>
		</div>
	{:else if filteredCards.length === 0}
		<div class="bg-gray-50 rounded-lg p-12 text-center">
			<p class="text-gray-600 text-lg mb-4">{tr('cards.noCardsFound')}</p>
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
			<!-- Cards Grid (2/3 when filter or batch panel is open on desktop) -->
			<div class={showFilterMenu || selectMode ? 'lg:col-span-2' : ''}>
				<div
					class="grid grid-cols-1 md:grid-cols-2 {showFilterMenu || selectMode
						? ''
						: 'lg:grid-cols-3'} gap-6"
				>
					{#each filteredCards as card (card.id)}
						<div
							class="block bg-white rounded-lg shadow-lg hover:shadow-xl transition overflow-hidden relative {selectMode &&
							selectedIds.has(card.id)
								? 'ring-2 ring-cyan-500'
								: ''}"
							style="border-left: 6px solid {card.merchant?.color || '#6B7280'}"
							role="button"
							tabindex="0"
							data-owner={card.owner && card.owner.id !== currentUserId
								? 'shared'
								: 'owned'}
							onclick={() => {
								if (selectMode) {
									toggleSelection(card.id);
								} else {
									goto(`/cards/${card.id}`);
								}
							}}
							onkeydown={(e) => {
								if (e.key === 'Enter' || e.key === ' ') {
									e.preventDefault();
									if (selectMode) {
										toggleSelection(card.id);
									} else {
										goto(`/cards/${card.id}`);
									}
								}
							}}
						>
							<div
								class="p-6 flex flex-col h-full {card.status !== 'active'
									? 'opacity-50 grayscale'
									: ''}"
							>
								<!-- Status Overlay -->
								{#if card.status !== 'active'}
									{@const badge = getStatusBadge(card.status || '')}
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

								<!-- Merchant + Owner -->
								<div class="flex items-start justify-end gap-2 mb-4">
									<div class="text-right min-w-0">
										<p class="text-sm text-gray-500 truncate">
											{card.merchant?.name || $t('cards.title')}
										</p>
										<p class="text-xs text-gray-400">
											{#if card.owner && card.owner.id !== currentUserId}
												{tr('cards.sharedBy', {
													name: card.owner.first_name || card.owner.email
												})}
											{:else if card.shared_with_count > 0}
												{tr('cards.sharedWithCount', {
													count: String(card.shared_with_count)
												})}
											{:else}
												{tr('cards.sharedWithNone')}
											{/if}
										</p>
									</div>
								</div>

								<!-- Barcode -->
								<div class="mt-auto">
									<div
										class="bg-gray-50 rounded-lg p-4 border border-gray-200 h-[120px] flex flex-col justify-center"
									>
										<div class="flex justify-center mb-2">
											<Barcode
												value={card.card_number}
												type={card.barcode_type || 'CODE128'}
												height={64}
												maxHeight={64}
											/>
										</div>
										<p
											class="text-center text-xs text-gray-600 font-mono break-all"
										>
											{card.card_number}
										</p>
									</div>

									<!-- Footer: Program -->
									<p class="text-xs text-gray-500 truncate mt-3">
										{#if card.program}
											{card.program}
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
										{tr('common.results', { count: filteredCards.length })}
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
									for="owner-filter-desktop"
								>
									{tr('cards.owner')}
								</label>
								<select
									id="owner-filter-desktop"
									bind:value={ownerFilter}
									class="w-full bg-gray-50 border border-gray-200 rounded-lg px-3 py-2 text-sm text-gray-700 focus:ring-2 focus:ring-cyan-500 focus:border-cyan-500 transition-colors"
								>
									<option value="all">{tr('cards.allCards')}</option>
									<option value="mine">{tr('cards.myCards')}</option>
								</select>
							</div>

							<div class="border-t border-gray-100"></div>

							<!-- Merchant -->
							<div class="py-4">
								<label
									class="block text-xs font-medium text-gray-500 uppercase tracking-wider mb-2"
									for="merchant-filter-desktop"
								>
									{tr('cards.merchant')}
								</label>
								<select
									id="merchant-filter-desktop"
									bind:value={merchantFilter}
									class="w-full bg-gray-50 border border-gray-200 rounded-lg px-3 py-2 text-sm text-gray-700 focus:ring-2 focus:ring-cyan-500 focus:border-cyan-500 transition-colors"
								>
									<option value="all">{tr('cards.allMerchants')}</option>
									{#each uniqueMerchants() as merchant}
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
									for="status-filter-desktop"
								>
									{tr('cards.statusLabel')}
								</label>
								<select
									id="status-filter-desktop"
									bind:value={statusFilter}
									class="w-full bg-gray-50 border border-gray-200 rounded-lg px-3 py-2 text-sm text-gray-700 focus:ring-2 focus:ring-cyan-500 focus:border-cyan-500 transition-colors"
								>
									<option value="active">{tr('cards.activeOnly')}</option>
									<option value="all-status">{tr('cards.allStatuses')}</option>
									<option value="inactive">{tr('cards.inactiveOnly')}</option>
								</select>
							</div>

							<div class="border-t border-gray-100"></div>

							<!-- Sort -->
							<div class="pt-4">
								<label
									class="block text-xs font-medium text-gray-500 uppercase tracking-wider mb-2"
									for="sort-filter-desktop"
								>
									{tr('cards.sortBy')}
								</label>
								<select
									id="sort-filter-desktop"
									bind:value={sortBy}
									class="w-full bg-gray-50 border border-gray-200 rounded-lg px-3 py-2 text-sm text-gray-700 focus:ring-2 focus:ring-cyan-500 focus:border-cyan-500 transition-colors"
								>
									<option value="name-asc">{tr('cards.sortNameAsc')}</option>
									<option value="name-desc">{tr('cards.sortNameDesc')}</option>
									<option value="newest">{tr('cards.sortNewest')}</option>
									<option value="oldest">{tr('cards.sortOldest')}</option>
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
					totalCount={filteredCards.length}
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
			ariaLabel={tr('common.filter')}
			bind:dialogRef
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
					<h3
						id="filter-dialog-title"
						class="text-sm font-semibold text-gray-900"
					>
						{tr('common.filter')}
					</h3>
				</div>
				<div class="flex items-center gap-2.5">
					<span
						class="text-xs text-gray-500 bg-gray-100 px-2.5 py-1 rounded-full tabular-nums"
					>
						{tr('common.results', { count: filteredCards.length })}
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
						for="owner-filter-mobile"
					>
						{tr('cards.owner')}
					</label>
					<select
						id="owner-filter-mobile"
						bind:value={ownerFilter}
						class="w-full bg-gray-50 border border-gray-200 rounded-lg px-3 py-2.5 text-sm text-gray-700 focus:ring-2 focus:ring-cyan-500 focus:border-cyan-500"
					>
						<option value="all">{tr('cards.allCards')}</option>
						<option value="mine">{tr('cards.myCards')}</option>
					</select>
				</div>

				<div class="border-t border-gray-100"></div>

				<!-- Merchant -->
				<div class="py-4">
					<label
						class="block text-xs font-medium text-gray-500 uppercase tracking-wider mb-2"
						for="merchant-filter-mobile"
					>
						{tr('cards.merchant')}
					</label>
					<select
						id="merchant-filter-mobile"
						bind:value={merchantFilter}
						class="w-full bg-gray-50 border border-gray-200 rounded-lg px-3 py-2.5 text-sm text-gray-700 focus:ring-2 focus:ring-cyan-500 focus:border-cyan-500"
					>
						<option value="all">{tr('cards.allMerchants')}</option>
						{#each uniqueMerchants() as merchant}
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
						for="status-filter-mobile"
					>
						{tr('cards.statusLabel')}
					</label>
					<select
						id="status-filter-mobile"
						bind:value={statusFilter}
						class="w-full bg-gray-50 border border-gray-200 rounded-lg px-3 py-2.5 text-sm text-gray-700 focus:ring-2 focus:ring-cyan-500 focus:border-cyan-500"
					>
						<option value="active">{tr('cards.activeOnly')}</option>
						<option value="all-status">{tr('cards.allStatuses')}</option>
						<option value="inactive">{tr('cards.inactiveOnly')}</option>
					</select>
				</div>

				<div class="border-t border-gray-100"></div>

				<!-- Sort -->
				<div class="py-4">
					<label
						class="block text-xs font-medium text-gray-500 uppercase tracking-wider mb-2"
						for="sort-filter-mobile"
					>
						{tr('cards.sortBy')}
					</label>
					<select
						id="sort-filter-mobile"
						bind:value={sortBy}
						class="w-full bg-gray-50 border border-gray-200 rounded-lg px-3 py-2.5 text-sm text-gray-700 focus:ring-2 focus:ring-cyan-500 focus:border-cyan-500"
					>
						<option value="name-asc">{tr('cards.sortNameAsc')}</option>
						<option value="name-desc">{tr('cards.sortNameDesc')}</option>
						<option value="newest">{tr('cards.sortNewest')}</option>
						<option value="oldest">{tr('cards.sortOldest')}</option>
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
						onclick={() => {
							showFilterMenu = false;
							filterApplied = true;
							setTimeout(() => {
								filterApplied = false;
							}, 1000);
						}}
						class="w-full btn btn-primary"
					>
						{tr('common.done')}
					</button>
				</div>
			</div>
		</BottomSheet>

		<!-- Live Region for Filter Announcements (SVL-A11Y-001) -->
		<div role="status" aria-live="polite" class="sr-only">
			{#if filterApplied}
				{filteredCards.length}
				{filteredCards.length === 1 ? tr('cards.card') : tr('cards.cards')}
				{tr('common.found')}
			{/if}
		</div>
	{/if}
</div>
