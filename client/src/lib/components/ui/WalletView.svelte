<script lang="ts" module>
	// Shared filter shape for the wallet-style list screens (wallet + merchant
	// detail). The wallet passes the module-level `walletFilters` $state; merchant
	// detail declares its own inline $state instance so the two scopes stay
	// separate. Either way it's a $state proxy mutated in place — so `filters` is a
	// plain (non-bindable) prop here.
	export interface WalletFilterState {
		searchInput: string;
		typeFilter: string;
		statusFilter: string;
		sortBy: string;
		ownerFilter: string;
		favoritesOnly: boolean;
		expiringFilter: string;
		scrollY: number;
	}
</script>

<script lang="ts">
	import type { Snippet } from 'svelte';
	import { ApiError, batchApi, translateBatchError } from '$lib/api';
	import BatchConfirmModal from '$lib/components/BatchConfirmModal.svelte';
	import BatchPanel from '$lib/components/BatchPanel.svelte';
	import BottomSheet from '$lib/components/BottomSheet.svelte';
	import ImportDialog from '$lib/components/ImportDialog.svelte';
	import ResourceTile from '$lib/components/ui/ResourceTile.svelte';
	import BarcodeModal, {
		type BarcodeModalItem
	} from '$lib/components/dashboard/BarcodeModal.svelte';
	import {
		cardToTileModel,
		voucherToTileModel,
		giftCardToTileModel
	} from '$lib/utils/tile-model';
	import { t } from '$lib/stores/i18n';
	import { toastStore } from '$lib/stores/toast';
	import type {
		BatchResponse,
		CardDTO,
		GiftCardDTO,
		VoucherDTO
	} from '$lib/types/api';
	import { get } from 'svelte/store';
	import { onMount, untrack } from 'svelte';
	import MerchantFilters from '$lib/components/MerchantFilters.svelte';
	import TypeFilterButtons from '$lib/components/TypeFilterButtons.svelte';
	import { SvelteSet } from 'svelte/reactivity';
	import { getGiftCardStatus } from '$lib/utils/resource-status';
	import {
		applyCommonFilters,
		searchMerchant,
		sortItems
	} from '$lib/wallet/filter';

	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);

	interface Props {
		filters: WalletFilterState;
		cards: CardDTO[];
		vouchers: VoucherDTO[];
		giftCards: GiftCardDTO[];
		enabledTypes: { cards: boolean; vouchers: boolean; giftCards: boolean };
		currentUserId: string | undefined;
		currentLocale: string;
		isOffline: boolean;
		onReload: () => void | Promise<void>;
		barcodeStorageKey: string;
		idPrefix: string;
		matchMerchantName?: boolean;
		typeFilterPlacement?: 'top' | 'batch-header';
		barcodeButtonVariant?: 'label' | 'icon';
		filterShowAll?: boolean;
		maxWidth?: boolean;
		header?: Snippet;
		emptyIcon?: Snippet;
		searchField?: Snippet;
	}

	let {
		filters,
		cards,
		vouchers,
		giftCards,
		enabledTypes,
		currentUserId,
		currentLocale,
		isOffline,
		onReload,
		barcodeStorageKey,
		idPrefix,
		matchMerchantName = true,
		typeFilterPlacement = 'top',
		barcodeButtonVariant = 'label',
		filterShowAll = false,
		maxWidth = true,
		header,
		emptyIcon,
		searchField
	}: Props = $props();

	let showFilterMenu = $state(false);

	// Batch selection state
	let selectMode = $state(false);
	let selectedIds = new SvelteSet<string>();
	let batchAction = $state<'delete' | 'share' | 'transfer'>('delete');
	let showBatchModal = $state(false);
	let batchLoading = $state(false);
	let showImportDialog = $state(false);

	// Barcode visibility toggle (per-list, localStorage-persisted, default off).
	let showBarcodes = $state(false);
	let barcodeModalItem = $state<BarcodeModalItem | null>(null);

	// 'active' (and 'valid' for vouchers) is the default status, not a
	// user-applied filter — so it does not count toward hasActiveFilters.
	const isDefaultStatus = $derived(
		filters.statusFilter === 'active' ||
			(filters.typeFilter === 'vouchers' && filters.statusFilter === 'valid')
	);
	const hasActiveFilters = $derived(
		filters.typeFilter !== 'all' ||
			!isDefaultStatus ||
			filters.sortBy !== 'newest' ||
			filters.ownerFilter !== 'all' ||
			filters.favoritesOnly ||
			filters.expiringFilter !== 'all'
	);

	// Status filter applies to all types (cards use active/inactive too)
	const showStatusFilter = true;

	const detailSortOptions = $derived.by(() => {
		const opts = [
			{ value: 'newest', label: tr('merchantOverview.detail.sortNewest') },
			{ value: 'oldest', label: tr('merchantOverview.detail.sortOldest') }
		];
		if (filters.typeFilter !== 'cards') {
			opts.push(
				{
					value: 'value-desc',
					label: tr('merchantOverview.detail.sortValueDesc')
				},
				{
					value: 'value-asc',
					label: tr('merchantOverview.detail.sortValueAsc')
				},
				{
					value: 'expiry-asc',
					label: tr('merchantOverview.detail.sortExpiryAsc')
				}
			);
		}
		return opts;
	});

	const detailStatusOptions = $derived.by(() => {
		if (filters.typeFilter === 'vouchers') {
			return [
				{ value: 'all', label: tr('merchantOverview.detail.statusAll') },
				{ value: 'valid', label: tr('vouchers.status.valid') },
				{
					value: 'expired',
					label: tr('merchantOverview.detail.statusExpired')
				},
				{
					value: 'inactive',
					label: tr('merchantOverview.detail.statusInactive')
				}
			];
		}
		if (filters.typeFilter === 'gift-cards') {
			return [
				{ value: 'all', label: tr('merchantOverview.detail.statusAll') },
				{ value: 'active', label: tr('giftCards.status.active') },
				{
					value: 'expired',
					label: tr('merchantOverview.detail.statusExpired')
				},
				{ value: 'depleted', label: tr('giftCards.status.depleted') }
			];
		}
		return [
			{ value: 'all', label: tr('merchantOverview.detail.statusAll') },
			{ value: 'active', label: tr('merchantOverview.filterStatusActive') },
			{ value: 'inactive', label: tr('merchantOverview.filterStatusInactive') }
		];
	});

	const detailOwnerOptions = $derived([
		{ value: 'all', label: tr('merchantOverview.detail.ownerAll') },
		{ value: 'mine', label: tr('merchantOverview.detail.ownerMine') },
		{ value: 'shared', label: tr('merchantOverview.detail.ownerShared') }
	]);

	const detailExpiringOptions = $derived([
		{ value: 'all', label: tr('merchantOverview.detail.expiringAll') },
		{ value: '7', label: tr('merchantOverview.detail.expiring7') },
		{ value: '30', label: tr('merchantOverview.detail.expiring30') }
	]);

	// Filtered items
	const filteredCards = $derived.by(() => {
		if (!enabledTypes.cards) return [];
		if (
			filters.typeFilter === 'vouchers' ||
			filters.typeFilter === 'gift-cards'
		)
			return [];
		let result = cards;
		// Expiring filter does not apply to cards (no expiry) - skip if set
		if (filters.expiringFilter !== 'all') return [];
		// Cards have no expiry, only active/inactive status
		if (filters.statusFilter === 'active') {
			result = result.filter((c) => c.status !== 'inactive');
		} else if (filters.statusFilter === 'inactive') {
			result = result.filter((c) => c.status === 'inactive');
		}
		if (filters.ownerFilter === 'mine') {
			result = result.filter((c) => !c.owner || c.owner.id === currentUserId);
		} else if (filters.ownerFilter === 'shared') {
			result = result.filter((c) => c.owner && c.owner.id !== currentUserId);
		}
		if (filters.favoritesOnly) {
			result = result.filter((c) => c.is_favorite);
		}
		if (filters.searchInput.trim()) {
			const q = filters.searchInput.trim().toLowerCase();
			result = result.filter(
				(c) =>
					searchMerchant(c.merchant?.name, q, matchMerchantName) ||
					c.card_number.toLowerCase().includes(q) ||
					c.program?.toLowerCase().includes(q) ||
					c.notes?.toLowerCase().includes(q)
			);
		}
		return sortItems(
			result,
			(c) => c.created_at,
			() => 0,
			() => undefined,
			filters.sortBy
		);
	});

	const filteredVouchers = $derived.by(() => {
		if (!enabledTypes.vouchers) return [];
		if (filters.typeFilter === 'cards' || filters.typeFilter === 'gift-cards')
			return [];
		let result = vouchers;
		if (filters.typeFilter === 'all') {
			if (filters.statusFilter === 'active') {
				result = result.filter((v) => v.status === 'valid');
			} else if (filters.statusFilter === 'inactive') {
				result = result.filter((v) => v.status !== 'valid');
			}
		} else {
			if (filters.statusFilter === 'valid') {
				result = result.filter((v) => v.status === 'valid');
			} else if (filters.statusFilter === 'expired') {
				result = result.filter((v) => v.status === 'expired');
			} else if (filters.statusFilter === 'inactive') {
				result = result.filter(
					(v) => v.status === 'inactive' || v.status === 'used'
				);
			}
		}
		result = applyCommonFilters(
			result,
			(v) => v.valid_until,
			filters.ownerFilter,
			filters.favoritesOnly,
			filters.expiringFilter,
			currentUserId
		);
		if (filters.searchInput.trim()) {
			const q = filters.searchInput.trim().toLowerCase();
			result = result.filter(
				(v) =>
					searchMerchant(v.merchant?.name, q, matchMerchantName) ||
					v.code.toLowerCase().includes(q) ||
					v.description?.toLowerCase().includes(q)
			);
		}
		return sortItems(
			result,
			(v) => v.created_at,
			(v) => v.value,
			(v) => v.valid_until,
			filters.sortBy
		);
	});

	const filteredGiftCards = $derived.by(() => {
		if (!enabledTypes.giftCards) return [];
		if (filters.typeFilter === 'cards' || filters.typeFilter === 'vouchers')
			return [];
		let result = giftCards;
		if (filters.typeFilter === 'all') {
			if (filters.statusFilter === 'active') {
				result = result.filter((g) => getGiftCardStatus(g) === 'active');
			} else if (filters.statusFilter === 'inactive') {
				result = result.filter((g) => getGiftCardStatus(g) !== 'active');
			}
		} else {
			if (filters.statusFilter === 'active') {
				result = result.filter((g) => getGiftCardStatus(g) === 'active');
			} else if (filters.statusFilter === 'expired') {
				result = result.filter((g) => getGiftCardStatus(g) === 'expired');
			} else if (filters.statusFilter === 'depleted') {
				result = result.filter((g) => getGiftCardStatus(g) === 'depleted');
			}
		}
		result = applyCommonFilters(
			result,
			(g) => g.expires_at,
			filters.ownerFilter,
			filters.favoritesOnly,
			filters.expiringFilter,
			currentUserId
		);
		if (filters.searchInput.trim()) {
			const q = filters.searchInput.trim().toLowerCase();
			result = result.filter(
				(g) =>
					searchMerchant(g.merchant?.name, q, matchMerchantName) ||
					g.card_number.toLowerCase().includes(q) ||
					g.notes?.toLowerCase().includes(q)
			);
		}
		return sortItems(
			result,
			(g) => g.created_at,
			(g) => g.current_balance,
			(g) => g.expires_at,
			filters.sortBy
		);
	});

	const cardTiles = $derived(
		filteredCards.map((c) => cardToTileModel(c, currentUserId, currentLocale))
	);
	const voucherTiles = $derived(
		filteredVouchers.map((v) =>
			voucherToTileModel(v, currentUserId, currentLocale)
		)
	);
	const giftCardTiles = $derived(
		filteredGiftCards.map((g) =>
			giftCardToTileModel(g, currentUserId, currentLocale)
		)
	);

	const totalFiltered = $derived(
		filteredCards.length + filteredVouchers.length + filteredGiftCards.length
	);
	const totalItems = $derived(
		cards.length + vouchers.length + giftCards.length
	);

	// Batch derived values
	const selectedCount = $derived(selectedIds.size);
	const currentFilteredItems = $derived.by(() => {
		if (filters.typeFilter === 'cards') return filteredCards;
		if (filters.typeFilter === 'vouchers') return filteredVouchers;
		if (filters.typeFilter === 'gift-cards') return filteredGiftCards;
		return [...filteredCards, ...filteredVouchers, ...filteredGiftCards];
	});
	const sharedSelectedCount = $derived.by(() => {
		const items = currentFilteredItems;
		return items.filter((item) => {
			if (!selectedIds.has(item.id)) return false;
			const owner = 'owner' in item ? item.owner : undefined;
			return owner && owner.id !== currentUserId;
		}).length;
	});
	const hasNonDeletableShared = $derived(sharedSelectedCount > 0);

	// Share modal permission hints, derived from the actual selection (not the
	// type filter), since a selection can span categories.
	const selectionHasGiftCards = $derived(
		filteredGiftCards.some((g) => selectedIds.has(g.id))
	);
	const selectionOnlyVouchers = $derived(
		selectedIds.size > 0 &&
			filteredVouchers.some((v) => selectedIds.has(v.id)) &&
			!filteredCards.some((c) => selectedIds.has(c.id)) &&
			!selectionHasGiftCards
	);

	// Reset selection and status filter when type filter changes. Seeded from
	// the initial filter (untracked so the read is not a reactive dependency).
	let lastTypeFilter = untrack(() => filters.typeFilter);
	$effect(() => {
		if (filters.typeFilter !== lastTypeFilter) {
			if (selectMode) {
				selectedIds.clear();
			}
			// Reset status filter to the active-equivalent for the new type.
			// Vouchers use 'valid'; cards/gift-cards/all use 'active'.
			filters.statusFilter =
				filters.typeFilter === 'vouchers' ? 'valid' : 'active';
			// Reset sort/expiring if switching to cards (no value/expiry)
			if (filters.typeFilter === 'cards') {
				if (
					['value-desc', 'value-asc', 'expiry-asc'].includes(filters.sortBy)
				) {
					filters.sortBy = 'newest';
				}
				filters.expiringFilter = 'all';
			}
			lastTypeFilter = filters.typeFilter;
		}
	});

	// Barcode toggle: load persisted preference once mounted. localStorage is
	// browser-only (must not run during SSR) and this is a one-shot read, not a
	// reactive effect (which would re-read on dep change and clobber the toggle).
	onMount(() => {
		showBarcodes = localStorage.getItem(barcodeStorageKey) === 'true';
	});

	function toggleBarcodes() {
		showBarcodes = !showBarcodes;
		localStorage.setItem(barcodeStorageKey, String(showBarcodes));
	}

	function resetFilters() {
		filters.typeFilter = 'all';
		filters.statusFilter = 'active';
		filters.sortBy = 'newest';
		filters.ownerFilter = 'all';
		filters.favoritesOnly = false;
		filters.expiringFilter = 'all';
		filters.searchInput = '';
	}

	// Batch functions
	function toggleSelectMode() {
		selectMode = !selectMode;
		if (!selectMode) {
			selectedIds.clear();
		} else {
			showFilterMenu = false;
			// Selection works across all categories; batch actions fan out to the
			// per-type endpoints, so no forced type switch on enter.
		}
	}

	// Group the current selection by resource type. Batch endpoints are per-type,
	// so a cross-category selection is dispatched as one call per non-empty group.
	function groupSelectedByType() {
		const sel = selectedIds;
		return {
			cards: filteredCards.filter((c) => sel.has(c.id)).map((c) => c.id),
			vouchers: filteredVouchers.filter((v) => sel.has(v.id)).map((v) => v.id),
			giftCards: filteredGiftCards.filter((g) => sel.has(g.id)).map((g) => g.id)
		};
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
		for (const item of currentFilteredItems) {
			selectedIds.add(item.id);
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
		const groups = groupSelectedByType();
		try {
			// One call per non-empty type group; merge into a single result so the
			// toast/aggregation below stays type-agnostic.
			const calls: Promise<BatchResponse>[] = [];
			if (batchAction === 'delete') {
				if (groups.cards.length) calls.push(batchApi.deleteCards(groups.cards));
				if (groups.vouchers.length)
					calls.push(batchApi.deleteVouchers(groups.vouchers));
				if (groups.giftCards.length)
					calls.push(batchApi.deleteGiftCards(groups.giftCards));
			} else if (batchAction === 'share') {
				const base = {
					email,
					can_edit: permissions.canEdit,
					can_delete: permissions.canDelete
				};
				if (groups.cards.length)
					calls.push(batchApi.shareCards({ ...base, ids: groups.cards }));
				// Vouchers are read-only shares; backend ignores edit flags.
				if (groups.vouchers.length)
					calls.push(batchApi.shareVouchers({ ...base, ids: groups.vouchers }));
				if (groups.giftCards.length)
					calls.push(
						batchApi.shareGiftCards({
							...base,
							ids: groups.giftCards,
							can_edit_transactions: permissions.canEditTransactions
						})
					);
			} else {
				if (groups.cards.length)
					calls.push(batchApi.transferCards(groups.cards, email));
				if (groups.vouchers.length)
					calls.push(batchApi.transferVouchers(groups.vouchers, email));
				if (groups.giftCards.length)
					calls.push(batchApi.transferGiftCards(groups.giftCards, email));
			}

			const ids = [...groups.cards, ...groups.vouchers, ...groups.giftCards];
			// allSettled, not all: a rejected type-group must not discard the
			// groups that already committed server-side. Fulfilled results are
			// aggregated; a rejected group is reported as a failed item so the
			// UI still reloads and clears the selection below.
			const settled = await Promise.allSettled(calls);
			const result = {
				success_count: settled.reduce(
					(n, s) => n + (s.status === 'fulfilled' ? s.value.success_count : 0),
					0
				),
				failed: settled.flatMap((s) =>
					s.status === 'fulfilled'
						? s.value.failed || []
						: [
								{
									id: '',
									error:
										s.reason instanceof ApiError
											? s.reason.message
											: 'batch.error'
								}
							]
				)
			};

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
			await onReload();
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
			const groups = groupSelectedByType();
			const count = selectedIds.size;
			// One export file per non-empty type group (endpoints are per-type).
			const exports: { blob: Blob; filename: string }[] = [];
			if (groups.cards.length)
				exports.push(await batchApi.exportCards(groups.cards));
			if (groups.vouchers.length)
				exports.push(await batchApi.exportVouchers(groups.vouchers));
			if (groups.giftCards.length)
				exports.push(await batchApi.exportGiftCards(groups.giftCards));

			for (const result of exports) {
				const url = URL.createObjectURL(result.blob);
				const a = document.createElement('a');
				a.href = url;
				a.download = result.filename;
				a.click();
				URL.revokeObjectURL(url);
			}
			toastStore.success($t('batch.exportSuccess', { count }));
		} catch {
			toastStore.error($t('batch.exportError'));
		}
	}
</script>

{#snippet batchTypeFilter()}
	<TypeFilterButtons
		bind:typeFilter={filters.typeFilter}
		cardsCount={cards.length}
		vouchersCount={vouchers.length}
		giftCardsCount={giftCards.length}
		allowToggle={false}
	/>
{/snippet}

<BatchConfirmModal
	action={batchAction}
	count={selectedCount}
	isOpen={showBatchModal}
	isLoading={batchLoading}
	onConfirm={executeBatchAction}
	onCancel={() => (showBatchModal = false)}
	hidePermissions={selectionOnlyVouchers}
	showTransactionPermission={selectionHasGiftCards}
/>

<ImportDialog
	isOpen={showImportDialog}
	onClose={() => (showImportDialog = false)}
	onImported={onReload}
/>

<BarcodeModal
	item={barcodeModalItem}
	onClose={() => (barcodeModalItem = null)}
/>

<div
	class="px-4 {maxWidth ? 'max-w-7xl mx-auto' : ''} pb-20 md:pb-4"
	class:pb-40={selectMode}
>
	{@render header?.()}

	{#if totalItems === 0}
		<div class="bg-surface-1 rounded-lg p-12 text-center">
			{@render emptyIcon?.()}
			<p class="text-text-muted text-lg mb-4">
				{tr('merchantOverview.detail.noItems')}
			</p>
			<p class="text-text-faint text-sm">
				{tr('merchantOverview.detail.noItemsHint')}
			</p>
		</div>
	{:else}
		<!-- Type filter (top placement): always visible. -->
		{#if typeFilterPlacement === 'top'}
			<div class="mb-4">
				<TypeFilterButtons
					bind:typeFilter={filters.typeFilter}
					cardsCount={cards.length}
					vouchersCount={vouchers.length}
					giftCardsCount={giftCards.length}
					showAll={filterShowAll}
					allowToggle={!selectMode}
				/>
			</div>
		{/if}

		{@render searchField?.()}

		<!-- WalletToolbar: Select · Filter · Barcode-Toggle · Import -->
		<div class="flex flex-col sm:flex-row gap-3 mb-6">
			<!-- Action Buttons (Desktop) -->
			<div class="hidden sm:flex gap-3">
				<!-- Select Mode Button -->
				<button
					type="button"
					onclick={toggleSelectMode}
					disabled={isOffline}
					class="flex items-center justify-center gap-2 control px-4 bg-white border rounded-md hover:bg-surface-1 transition-colors disabled:opacity-50 disabled:cursor-not-allowed {selectMode
						? 'ring-2 ring-accent border-accent'
						: 'border-border-field'}"
					title={tr('batch.selectMode')}
					aria-label={tr('batch.selectMode')}
					aria-pressed={selectMode}
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
							d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"
						/>
					</svg>
				</button>
				<!-- Filter Button -->
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
				<!-- Barcode Toggle Button -->
				<button
					type="button"
					onclick={toggleBarcodes}
					class="flex items-center justify-center gap-2 control {barcodeButtonVariant ===
					'label'
						? 'px-6'
						: 'px-4'} bg-white border rounded-md hover:bg-surface-1 transition-colors {showBarcodes
						? 'ring-2 ring-accent border-accent'
						: 'border-border-field'}"
					title={showBarcodes
						? tr('barcodeToggle.hide')
						: tr('barcodeToggle.show')}
					aria-label={showBarcodes
						? tr('barcodeToggle.hide')
						: tr('barcodeToggle.show')}
					aria-pressed={showBarcodes}
				>
					{#if barcodeButtonVariant === 'label'}
						<span class="text-sm font-medium text-text-muted whitespace-nowrap">
							{tr('barcodeToggle.label')}
						</span>
					{:else}
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
								d="M4 5h1v14H4V5zm3 0h1v14H7V5zm3 0h2v14h-2V5zm4 0h1v14h-1V5zm3 0h2v14h-2V5z"
							></path>
						</svg>
					{/if}
				</button>
				<!-- Import Button -->
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
			</div>

			<!-- Action Buttons (Mobile) -->
			<div class="flex sm:hidden gap-3">
				<!-- Select Mode Button (Mobile) -->
				<button
					type="button"
					onclick={toggleSelectMode}
					disabled={isOffline}
					class="flex-1 flex items-center justify-center control bg-white border border-border-field rounded-md hover:bg-surface-1 transition-colors disabled:opacity-50 disabled:cursor-not-allowed {selectMode
						? 'ring-2 ring-accent border-accent'
						: ''}"
					aria-label={tr('batch.selectMode')}
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
							d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"
						/>
					</svg>
				</button>
				<!-- Filter Button (Mobile) -->
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
				<!-- Barcode Toggle Button (Mobile) -->
				<button
					type="button"
					onclick={toggleBarcodes}
					class="flex-[2] flex items-center justify-center gap-2 control bg-white border rounded-md hover:bg-surface-1 transition-colors {showBarcodes
						? 'ring-2 ring-accent border-accent'
						: 'border-border-field'}"
					aria-label={showBarcodes
						? tr('barcodeToggle.hide')
						: tr('barcodeToggle.show')}
					aria-pressed={showBarcodes}
				>
					{#if barcodeButtonVariant === 'label'}
						<span class="text-sm font-medium text-text-muted whitespace-nowrap">
							{tr('barcodeToggle.label')}
						</span>
					{:else}
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
								d="M4 5h1v14H4V5zm3 0h1v14H7V5zm3 0h2v14h-2V5zm4 0h1v14h-1V5zm3 0h2v14h-2V5z"
							></path>
						</svg>
					{/if}
				</button>
			</div>
		</div>

		{#if totalFiltered === 0 && (filters.searchInput || hasActiveFilters)}
			<!-- No results with filters -->
			<div class="bg-surface-1 rounded-lg p-12 text-center">
				<p class="text-text-muted text-lg mb-4">{tr('search.no_results')}</p>
				{#if hasActiveFilters}
					<button type="button" onclick={resetFilters} class="btn btn-ghost">
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
				<!-- Items (mixed grid, sorted: cards → vouchers → gift cards) -->
				<div class={showFilterMenu || selectMode ? 'lg:col-span-2' : ''}>
					<div
						class="grid grid-cols-1 md:grid-cols-2 {showFilterMenu || selectMode
							? ''
							: 'lg:grid-cols-3'} gap-6"
					>
						<!-- Cards -->
						{#each cardTiles as model (model.id)}
							<ResourceTile
								{model}
								showBarcode={showBarcodes}
								{selectMode}
								selected={selectedIds.has(model.id)}
								onSelect={toggleSelection}
								onShowBarcode={(item) => (barcodeModalItem = item)}
							/>
						{/each}

						<!-- Vouchers -->
						{#each voucherTiles as model (model.id)}
							<ResourceTile
								{model}
								showBarcode={showBarcodes}
								{selectMode}
								selected={selectedIds.has(model.id)}
								onSelect={toggleSelection}
								onShowBarcode={(item) => (barcodeModalItem = item)}
							/>
						{/each}

						<!-- Gift Cards -->
						{#each giftCardTiles as model (model.id)}
							<ResourceTile
								{model}
								showBarcode={showBarcodes}
								{selectMode}
								selected={selectedIds.has(model.id)}
								onSelect={toggleSelection}
								onShowBarcode={(item) => (barcodeModalItem = item)}
							/>
						{/each}
					</div>
				</div>

				<!-- Filter Side-Panel (Desktop only) -->
				{#if showFilterMenu && !selectMode}
					<div class="hidden lg:block lg:col-span-1">
						<div
							class="bg-white rounded-xl shadow-lg sticky top-4 overflow-hidden"
						>
							<!-- Header -->
							<div
								class="px-5 py-4 bg-surface-1/80 border-b border-border-soft"
							>
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
											{tr('common.results', { count: totalFiltered })}
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
													d="M6 18L18 6M6 6l12 12"
												></path>
											</svg>
										</button>
									</div>
								</div>
							</div>

							<div class="p-5">
								<MerchantFilters
									bind:typeFilter={filters.typeFilter}
									bind:statusFilter={filters.statusFilter}
									bind:sortBy={filters.sortBy}
									bind:ownerFilter={filters.ownerFilter}
									bind:favoritesOnly={filters.favoritesOnly}
									bind:expiringFilter={filters.expiringFilter}
									sortOptions={detailSortOptions}
									statusOptions={detailStatusOptions}
									cardsCount={cards.length}
									vouchersCount={vouchers.length}
									giftCardsCount={giftCards.length}
									{showStatusFilter}
									ownerOptions={detailOwnerOptions}
									expiringOptions={detailExpiringOptions}
									showExpiringFilter={filters.typeFilter !== 'cards'}
									{hasActiveFilters}
									onReset={resetFilters}
									showAll={filterShowAll}
									idPrefix="{idPrefix}-desktop"
								/>
							</div>
						</div>
					</div>
				{/if}

				<!-- Batch Actions Side-Panel (Desktop only) -->
				{#if selectMode}
					<BatchPanel
						{selectedCount}
						totalCount={currentFilteredItems.length}
						{sharedSelectedCount}
						{hasNonDeletableShared}
						onSelectAll={selectAll}
						onDeselectAll={deselectAll}
						onDelete={() => openBatchModal('delete')}
						onShare={() => openBatchModal('share')}
						onTransfer={() => openBatchModal('transfer')}
						onExport={handleBatchExport}
						onCancel={toggleSelectMode}
						headerExtra={typeFilterPlacement === 'batch-header'
							? batchTypeFilter
							: undefined}
					/>
				{/if}
			</div>
		{/if}
	{/if}
</div>

<!-- Mobile Filter Bottom Sheet -->
<BottomSheet
	open={showFilterMenu && !selectMode}
	onClose={() => (showFilterMenu = false)}
	maxHeight="80vh"
	ariaLabel={tr('common.filter')}
>
	<div class="px-4 pb-4 pt-1">
		<div class="mb-3 flex items-center justify-between">
			<h3 class="text-lg font-semibold text-text">
				{tr('common.filter')}
			</h3>
			<button
				type="button"
				onclick={() => (showFilterMenu = false)}
				class="text-text-faint hover:text-text-muted transition-colors"
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

		<div class="pt-1">
			<MerchantFilters
				bind:typeFilter={filters.typeFilter}
				bind:statusFilter={filters.statusFilter}
				bind:sortBy={filters.sortBy}
				bind:ownerFilter={filters.ownerFilter}
				bind:favoritesOnly={filters.favoritesOnly}
				bind:expiringFilter={filters.expiringFilter}
				sortOptions={detailSortOptions}
				statusOptions={detailStatusOptions}
				cardsCount={cards.length}
				vouchersCount={vouchers.length}
				giftCardsCount={giftCards.length}
				{showStatusFilter}
				ownerOptions={detailOwnerOptions}
				expiringOptions={detailExpiringOptions}
				showExpiringFilter={filters.typeFilter !== 'cards'}
				{hasActiveFilters}
				onReset={resetFilters}
				showAll={filterShowAll}
				idPrefix="{idPrefix}-mobile"
			/>
		</div>

		<div class="px-6 pb-6 pt-2">
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
