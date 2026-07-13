<script lang="ts">
	import {
		ApiError,
		batchApi,
		cardsApi,
		giftCardsApi,
		merchantsApi,
		translateBatchError,
		vouchersApi
	} from '$lib/api';
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
	import MerchantFilters from '$lib/components/MerchantFilters.svelte';
	import TypeFilterButtons from '$lib/components/TypeFilterButtons.svelte';
	import { SvelteSet } from 'svelte/reactivity';

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

	let searchInput = $state('');
	let typeFilter = $state('all');
	let statusFilter = $state('all');
	let sortBy = $state('newest');
	let ownerFilter = $state('all');
	let favoritesOnly = $state(false);
	let expiringFilter = $state('all');
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

	const hasActiveFilters = $derived(
		typeFilter !== 'all' ||
			statusFilter !== 'all' ||
			sortBy !== 'newest' ||
			ownerFilter !== 'all' ||
			favoritesOnly ||
			expiringFilter !== 'all'
	);

	// Show status filter only when vouchers or gift cards are visible
	const showStatusFilter = $derived(typeFilter !== 'cards');

	const detailSortOptions = $derived.by(() => {
		const opts = [
			{ value: 'newest', label: tr('merchantOverview.detail.sortNewest') },
			{ value: 'oldest', label: tr('merchantOverview.detail.sortOldest') }
		];
		if (typeFilter !== 'cards') {
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
		if (typeFilter === 'vouchers') {
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
		if (typeFilter === 'gift-cards') {
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

	const merchantId = $derived(get(page).params.id ?? '');

	// Helper: check if item expires within N days
	function expiresWithinDays(
		dateStr: string | undefined,
		days: number
	): boolean {
		if (!dateStr) return false;
		const expiry = new Date(dateStr);
		const now = new Date();
		const diffMs = expiry.getTime() - now.getTime();
		return diffMs > 0 && diffMs <= days * 24 * 60 * 60 * 1000;
	}

	// Shared filter: ownership, favorites, expiring
	function applyCommonFilters<
		T extends { owner?: { id?: string }; is_favorite: boolean }
	>(items: T[], getExpiryDate: (item: T) => string | undefined): T[] {
		let result = items;
		if (ownerFilter === 'mine') {
			result = result.filter(
				(item) => !item.owner || item.owner.id === currentUserId
			);
		} else if (ownerFilter === 'shared') {
			result = result.filter(
				(item) => item.owner && item.owner.id !== currentUserId
			);
		}
		if (favoritesOnly) {
			result = result.filter((item) => item.is_favorite);
		}
		if (expiringFilter === '7') {
			result = result.filter((item) =>
				expiresWithinDays(getExpiryDate(item), 7)
			);
		} else if (expiringFilter === '30') {
			result = result.filter((item) =>
				expiresWithinDays(getExpiryDate(item), 30)
			);
		}
		return result;
	}

	// Sort helper
	function sortItems<T>(
		items: T[],
		getDate: (item: T) => string,
		getValue: (item: T) => number,
		getExpiry: (item: T) => string | undefined
	): T[] {
		return [...items].sort((a, b) => {
			switch (sortBy) {
				case 'newest':
					return (
						new Date(getDate(b)).getTime() - new Date(getDate(a)).getTime()
					);
				case 'oldest':
					return (
						new Date(getDate(a)).getTime() - new Date(getDate(b)).getTime()
					);
				case 'value-desc':
					return getValue(b) - getValue(a);
				case 'value-asc':
					return getValue(a) - getValue(b);
				case 'expiry-asc': {
					const ea = getExpiry(a);
					const eb = getExpiry(b);
					if (!ea && !eb) return 0;
					if (!ea) return 1;
					if (!eb) return -1;
					return new Date(ea).getTime() - new Date(eb).getTime();
				}
				default:
					return 0;
			}
		});
	}

	// Filtered items
	const filteredCards = $derived.by(() => {
		if (typeFilter === 'vouchers' || typeFilter === 'gift-cards') return [];
		let result = cards;
		// Expiring filter does not apply to cards (no expiry) - skip if set
		if (expiringFilter !== 'all') return [];
		if (ownerFilter === 'mine') {
			result = result.filter((c) => !c.owner || c.owner.id === currentUserId);
		} else if (ownerFilter === 'shared') {
			result = result.filter((c) => c.owner && c.owner.id !== currentUserId);
		}
		if (favoritesOnly) {
			result = result.filter((c) => c.is_favorite);
		}
		if (searchInput.trim()) {
			const q = searchInput.trim().toLowerCase();
			result = result.filter(
				(c) =>
					c.card_number.toLowerCase().includes(q) ||
					c.program?.toLowerCase().includes(q) ||
					c.notes?.toLowerCase().includes(q)
			);
		}
		return sortItems(
			result,
			(c) => c.created_at,
			() => 0,
			() => undefined
		);
	});

	const filteredVouchers = $derived.by(() => {
		if (typeFilter === 'cards' || typeFilter === 'gift-cards') return [];
		let result = vouchers;
		if (typeFilter === 'all') {
			if (statusFilter === 'active') {
				result = result.filter((v) => v.status === 'valid');
			} else if (statusFilter === 'inactive') {
				result = result.filter((v) => v.status !== 'valid');
			}
		} else {
			if (statusFilter === 'valid') {
				result = result.filter((v) => v.status === 'valid');
			} else if (statusFilter === 'expired') {
				result = result.filter((v) => v.status === 'expired');
			} else if (statusFilter === 'inactive') {
				result = result.filter(
					(v) => v.status === 'inactive' || v.status === 'used'
				);
			}
		}
		result = applyCommonFilters(result, (v) => v.valid_until);
		if (searchInput.trim()) {
			const q = searchInput.trim().toLowerCase();
			result = result.filter(
				(v) =>
					v.code.toLowerCase().includes(q) ||
					v.description?.toLowerCase().includes(q)
			);
		}
		return sortItems(
			result,
			(v) => v.created_at,
			(v) => v.value,
			(v) => v.valid_until
		);
	});

	const filteredGiftCards = $derived.by(() => {
		if (typeFilter === 'cards' || typeFilter === 'vouchers') return [];
		let result = giftCards;
		if (typeFilter === 'all') {
			if (statusFilter === 'active') {
				result = result.filter((g) => getComputedStatus(g) === 'active');
			} else if (statusFilter === 'inactive') {
				result = result.filter((g) => getComputedStatus(g) !== 'active');
			}
		} else {
			if (statusFilter === 'active') {
				result = result.filter((g) => getComputedStatus(g) === 'active');
			} else if (statusFilter === 'expired') {
				result = result.filter((g) => getComputedStatus(g) === 'expired');
			} else if (statusFilter === 'depleted') {
				result = result.filter((g) => getComputedStatus(g) === 'depleted');
			}
		}
		result = applyCommonFilters(result, (g) => g.expires_at);
		if (searchInput.trim()) {
			const q = searchInput.trim().toLowerCase();
			result = result.filter(
				(g) =>
					g.card_number.toLowerCase().includes(q) ||
					g.notes?.toLowerCase().includes(q)
			);
		}
		return sortItems(
			result,
			(g) => g.created_at,
			(g) => g.current_balance,
			(g) => g.expires_at
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
		if (typeFilter === 'cards') return filteredCards;
		if (typeFilter === 'vouchers') return filteredVouchers;
		if (typeFilter === 'gift-cards') return filteredGiftCards;
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

	// Reset selection and status filter when type filter changes
	let lastTypeFilter = 'all';
	$effect(() => {
		if (typeFilter !== lastTypeFilter) {
			if (selectMode) {
				selectedIds.clear();
			}
			// Reset status filter when type changes to avoid invalid combinations
			statusFilter = 'all';
			// Reset sort/expiring if switching to cards (no value/expiry)
			if (typeFilter === 'cards') {
				if (['value-desc', 'value-asc', 'expiry-asc'].includes(sortBy)) {
					sortBy = 'newest';
				}
				expiringFilter = 'all';
			}
			lastTypeFilter = typeFilter;
		}
	});

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
		showBarcodes =
			localStorage.getItem('savvy_merchant_detail_show_barcodes') === 'true';
		loadData();
	});

	function toggleBarcodes() {
		showBarcodes = !showBarcodes;
		localStorage.setItem(
			'savvy_merchant_detail_show_barcodes',
			String(showBarcodes)
		);
	}

	function resetFilters() {
		typeFilter = 'all';
		statusFilter = 'all';
		sortBy = 'newest';
		ownerFilter = 'all';
		favoritesOnly = false;
		expiringFilter = 'all';
		searchInput = '';
	}

	// Batch functions
	function toggleSelectMode() {
		selectMode = !selectMode;
		if (!selectMode) {
			selectedIds.clear();
			typeFilter = 'all';
		} else {
			showFilterMenu = false;
			// Force a specific type filter when entering select mode
			if (typeFilter === 'all') {
				if (cards.length > 0) typeFilter = 'cards';
				else if (vouchers.length > 0) typeFilter = 'vouchers';
				else if (giftCards.length > 0) typeFilter = 'gift-cards';
			}
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
		const ids = [...selectedIds];
		try {
			let result;
			if (batchAction === 'delete') {
				if (typeFilter === 'cards') result = await batchApi.deleteCards(ids);
				else if (typeFilter === 'vouchers')
					result = await batchApi.deleteVouchers(ids);
				else result = await batchApi.deleteGiftCards(ids);
			} else if (batchAction === 'share') {
				const req = {
					ids,
					email,
					can_edit: permissions.canEdit,
					can_delete: permissions.canDelete,
					...(typeFilter === 'gift-cards'
						? { can_edit_transactions: permissions.canEditTransactions }
						: {})
				};
				if (typeFilter === 'cards') result = await batchApi.shareCards(req);
				else if (typeFilter === 'vouchers')
					result = await batchApi.shareVouchers(req);
				else result = await batchApi.shareGiftCards(req);
			} else {
				if (typeFilter === 'cards')
					result = await batchApi.transferCards(ids, email);
				else if (typeFilter === 'vouchers')
					result = await batchApi.transferVouchers(ids, email);
				else result = await batchApi.transferGiftCards(ids, email);
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
			await loadData();
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
			let result;
			const ids = [...selectedIds];
			if (typeFilter === 'cards') result = await batchApi.exportCards(ids);
			else if (typeFilter === 'vouchers')
				result = await batchApi.exportVouchers(ids);
			else result = await batchApi.exportGiftCards(ids);

			const url = URL.createObjectURL(result.blob);
			const a = document.createElement('a');
			a.href = url;
			a.download = result.filename;
			a.click();
			URL.revokeObjectURL(url);
			toastStore.success(
				$t('batch.exportSuccess', { count: selectedIds.size })
			);
		} catch {
			toastStore.error($t('batch.exportError'));
		}
	}

	// Gift card computed status
	function getComputedStatus(giftCard: GiftCardDTO): string {
		if (giftCard.current_balance === 0) return 'depleted';
		if (giftCard.expires_at && new Date(giftCard.expires_at) < new Date())
			return 'expired';
		return 'active';
	}
</script>

<svelte:head>
	<title
		>{merchant?.name ?? tr('merchantOverview.detail.title')} - {tr(
			'common.appName'
		)}</title
	>
</svelte:head>

<BatchConfirmModal
	action={batchAction}
	count={selectedCount}
	isOpen={showBatchModal}
	isLoading={batchLoading}
	onConfirm={executeBatchAction}
	onCancel={() => (showBatchModal = false)}
	hidePermissions={typeFilter === 'vouchers'}
	showTransactionPermission={typeFilter === 'gift-cards'}
/>

<ImportDialog
	isOpen={showImportDialog}
	onClose={() => (showImportDialog = false)}
	onImported={loadData}
/>

<BarcodeModal
	item={barcodeModalItem}
	onClose={() => (barcodeModalItem = null)}
/>

<div class="px-4 pb-20 md:pb-4" class:pb-40={selectMode}>
	{#if isLoading}
		<LoadingSpinner />
	{:else if merchant}
		<!-- Header -->
		<div class="mb-8">
			<div class="flex items-center justify-between">
				<div class="flex items-center gap-3">
					<div
						class="w-2 h-8 rounded-full"
						style="background-color: {merchant.color}"
					></div>
					<h1 class="text-3xl font-bold text-text">{merchant.name}</h1>
				</div>
				{#if isAdmin}
					<a
						href={resolve(`/admin/merchants/${merchant.id}/edit`)}
						class="btn btn-xs btn-gray whitespace-nowrap flex items-center gap-1.5"
					>
						{tr('common.edit')}
					</a>
				{/if}
			</div>
		</div>

		{#if totalItems === 0}
			<div class="bg-surface-1 rounded-lg p-12 text-center">
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
				<p class="text-text-muted text-lg mb-4">
					{tr('merchantOverview.detail.noItems')}
				</p>
				<p class="text-text-faint text-sm">
					{tr('merchantOverview.detail.noItemsHint')}
				</p>
			</div>
		{:else}
			<!-- Search + Action Buttons -->
			<div class="flex flex-col sm:flex-row gap-3 mb-6">
				<!-- Search Bar -->
				<div class="flex-1">
					<input
						type="search"
						bind:value={searchInput}
						placeholder={tr('common.search')}
						class="w-full px-4 py-2 bg-white border border-border-field rounded-md focus:ring-accent focus:border-accent"
					/>
				</div>

				<!-- Action Buttons (Desktop) -->
				<div class="hidden sm:flex gap-3">
					<!-- Select Mode Button -->
					<button
						type="button"
						onclick={toggleSelectMode}
						disabled={isOffline}
						class="flex items-center justify-center gap-2 h-[42px] px-4 bg-white border rounded-md hover:bg-surface-1 transition-colors disabled:opacity-50 disabled:cursor-not-allowed {selectMode
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
					<!-- Barcode Toggle Button -->
					<button
						type="button"
						onclick={toggleBarcodes}
						class="btn btn-ghost {showBarcodes
							? 'ring-2 ring-accent border-accent'
							: ''}"
						title={showBarcodes
							? tr('barcodeToggle.hide')
							: tr('barcodeToggle.show')}
						aria-label={showBarcodes
							? tr('barcodeToggle.hide')
							: tr('barcodeToggle.show')}
						aria-pressed={showBarcodes}
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
								d="M4 5h1v14H4V5zm3 0h1v14H7V5zm3 0h2v14h-2V5zm4 0h1v14h-1V5zm3 0h2v14h-2V5z"
							></path>
						</svg>
					</button>
					<!-- Filter Button -->
					<button
						type="button"
						onclick={(e: MouseEvent) => {
							e.stopPropagation();
							showFilterMenu = !showFilterMenu;
						}}
						class="flex items-center justify-center gap-2 h-[42px] px-4 bg-white border border-border-field rounded-md hover:bg-surface-1 transition-colors relative"
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
						class="flex-1 flex items-center justify-center h-[42px] bg-white border border-border-field rounded-md hover:bg-surface-1 transition-colors disabled:opacity-50 disabled:cursor-not-allowed {selectMode
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
					<!-- Barcode Toggle Button (Mobile) -->
					<button
						type="button"
						onclick={toggleBarcodes}
						class="flex-1 btn btn-ghost {showBarcodes
							? 'ring-2 ring-accent border-accent'
							: ''}"
						aria-label={showBarcodes
							? tr('barcodeToggle.hide')
							: tr('barcodeToggle.show')}
						aria-pressed={showBarcodes}
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
								d="M4 5h1v14H4V5zm3 0h1v14H7V5zm3 0h2v14h-2V5zm4 0h1v14h-1V5zm3 0h2v14h-2V5z"
							></path>
						</svg>
					</button>
					<!-- Filter Button (Mobile) -->
					<button
						type="button"
						onclick={(e: MouseEvent) => {
							e.stopPropagation();
							showFilterMenu = !showFilterMenu;
						}}
						class="flex-1 flex items-center justify-center h-[42px] bg-white border border-border-field rounded-md hover:bg-surface-1 transition-colors relative"
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
				</div>
			</div>

			{#if totalFiltered === 0 && (searchInput || hasActiveFilters)}
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
							class="grid grid-cols-1 md:grid-cols-2 {showFilterMenu ||
							selectMode
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
										bind:typeFilter
										bind:statusFilter
										bind:sortBy
										bind:ownerFilter
										bind:favoritesOnly
										bind:expiringFilter
										sortOptions={detailSortOptions}
										statusOptions={detailStatusOptions}
										cardsCount={cards.length}
										vouchersCount={vouchers.length}
										giftCardsCount={giftCards.length}
										{showStatusFilter}
										ownerOptions={detailOwnerOptions}
										expiringOptions={detailExpiringOptions}
										showExpiringFilter={typeFilter !== 'cards'}
										{hasActiveFilters}
										onReset={resetFilters}
										idPrefix="detail-desktop"
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
						>
							{#snippet headerExtra()}
								<TypeFilterButtons
									bind:typeFilter
									cardsCount={cards.length}
									vouchersCount={vouchers.length}
									giftCardsCount={giftCards.length}
									allowToggle={false}
								/>
							{/snippet}
						</BatchPanel>
					{/if}
				</div>
			{/if}
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
	<div class="p-6">
		<div class="flex items-center justify-between mb-4">
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

		<div class="px-6 pt-4">
			<MerchantFilters
				bind:typeFilter
				bind:statusFilter
				bind:sortBy
				bind:ownerFilter
				bind:favoritesOnly
				bind:expiringFilter
				sortOptions={detailSortOptions}
				statusOptions={detailStatusOptions}
				cardsCount={cards.length}
				vouchersCount={vouchers.length}
				giftCardsCount={giftCards.length}
				{showStatusFilter}
				ownerOptions={detailOwnerOptions}
				expiringOptions={detailExpiringOptions}
				showExpiringFilter={typeFilter !== 'cards'}
				{hasActiveFilters}
				onReset={resetFilters}
				idPrefix="detail-mobile"
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
