<script lang="ts">
	import {
		ApiError,
		batchApi,
		cardsApi,
		giftCardsApi,
		translateBatchError,
		vouchersApi
	} from '$lib/api';
	import BatchConfirmModal from '$lib/components/BatchConfirmModal.svelte';
	import BatchPanel from '$lib/components/BatchPanel.svelte';
	import BottomSheet from '$lib/components/BottomSheet.svelte';
	import ImportDialog from '$lib/components/ImportDialog.svelte';
	import ResourceTile from '$lib/components/ui/ResourceTile.svelte';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import BarcodeModal, {
		type BarcodeModalItem
	} from '$lib/components/dashboard/BarcodeModal.svelte';
	import {
		cardToTileModel,
		voucherToTileModel,
		giftCardToTileModel
	} from '$lib/utils/tile-model';
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
	import MerchantFilters from '$lib/components/MerchantFilters.svelte';
	import TypeFilterButtons from '$lib/components/TypeFilterButtons.svelte';
	import { logger } from '$lib/utils/logger';
	import { SvelteSet } from 'svelte/reactivity';
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

	let cards = $state<CardDTO[]>([]);
	let vouchers = $state<VoucherDTO[]>([]);
	let giftCards = $state<GiftCardDTO[]>([]);
	let isLoading = $state(true);

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

	// Apply ?type= query param on mount (from /cards /vouchers /gift-cards redirects).
	const validTypes = ['cards', 'vouchers', 'gift-cards'];

	const hasActiveFilters = $derived(
		walletFilters.typeFilter !== 'all' ||
			walletFilters.statusFilter !== 'all' ||
			walletFilters.sortBy !== 'newest' ||
			walletFilters.ownerFilter !== 'all' ||
			walletFilters.favoritesOnly ||
			walletFilters.expiringFilter !== 'all'
	);

	// Show status filter only when vouchers or gift cards are visible
	const showStatusFilter = $derived(walletFilters.typeFilter !== 'cards');

	const detailSortOptions = $derived.by(() => {
		const opts = [
			{ value: 'newest', label: tr('merchantOverview.detail.sortNewest') },
			{ value: 'oldest', label: tr('merchantOverview.detail.sortOldest') }
		];
		if (walletFilters.typeFilter !== 'cards') {
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
		if (walletFilters.typeFilter === 'vouchers') {
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
		if (walletFilters.typeFilter === 'gift-cards') {
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
		if (walletFilters.ownerFilter === 'mine') {
			result = result.filter(
				(item) => !item.owner || item.owner.id === currentUserId
			);
		} else if (walletFilters.ownerFilter === 'shared') {
			result = result.filter(
				(item) => item.owner && item.owner.id !== currentUserId
			);
		}
		if (walletFilters.favoritesOnly) {
			result = result.filter((item) => item.is_favorite);
		}
		if (walletFilters.expiringFilter === '7') {
			result = result.filter((item) =>
				expiresWithinDays(getExpiryDate(item), 7)
			);
		} else if (walletFilters.expiringFilter === '30') {
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
			switch (walletFilters.sortBy) {
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
		if (!showCards) return [];
		if (
			walletFilters.typeFilter === 'vouchers' ||
			walletFilters.typeFilter === 'gift-cards'
		)
			return [];
		let result = cards;
		// Expiring filter does not apply to cards (no expiry) - skip if set
		if (walletFilters.expiringFilter !== 'all') return [];
		if (walletFilters.ownerFilter === 'mine') {
			result = result.filter((c) => !c.owner || c.owner.id === currentUserId);
		} else if (walletFilters.ownerFilter === 'shared') {
			result = result.filter((c) => c.owner && c.owner.id !== currentUserId);
		}
		if (walletFilters.favoritesOnly) {
			result = result.filter((c) => c.is_favorite);
		}
		if (walletFilters.searchInput.trim()) {
			const q = walletFilters.searchInput.trim().toLowerCase();
			result = result.filter(
				(c) =>
					c.merchant?.name.toLowerCase().includes(q) ||
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
		if (!showVouchers) return [];
		if (
			walletFilters.typeFilter === 'cards' ||
			walletFilters.typeFilter === 'gift-cards'
		)
			return [];
		let result = vouchers;
		if (walletFilters.typeFilter === 'all') {
			if (walletFilters.statusFilter === 'active') {
				result = result.filter((v) => v.status === 'valid');
			} else if (walletFilters.statusFilter === 'inactive') {
				result = result.filter((v) => v.status !== 'valid');
			}
		} else {
			if (walletFilters.statusFilter === 'valid') {
				result = result.filter((v) => v.status === 'valid');
			} else if (walletFilters.statusFilter === 'expired') {
				result = result.filter((v) => v.status === 'expired');
			} else if (walletFilters.statusFilter === 'inactive') {
				result = result.filter(
					(v) => v.status === 'inactive' || v.status === 'used'
				);
			}
		}
		result = applyCommonFilters(result, (v) => v.valid_until);
		if (walletFilters.searchInput.trim()) {
			const q = walletFilters.searchInput.trim().toLowerCase();
			result = result.filter(
				(v) =>
					v.merchant?.name.toLowerCase().includes(q) ||
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
		if (!showGiftCards) return [];
		if (
			walletFilters.typeFilter === 'cards' ||
			walletFilters.typeFilter === 'vouchers'
		)
			return [];
		let result = giftCards;
		if (walletFilters.typeFilter === 'all') {
			if (walletFilters.statusFilter === 'active') {
				result = result.filter((g) => getComputedStatus(g) === 'active');
			} else if (walletFilters.statusFilter === 'inactive') {
				result = result.filter((g) => getComputedStatus(g) !== 'active');
			}
		} else {
			if (walletFilters.statusFilter === 'active') {
				result = result.filter((g) => getComputedStatus(g) === 'active');
			} else if (walletFilters.statusFilter === 'expired') {
				result = result.filter((g) => getComputedStatus(g) === 'expired');
			} else if (walletFilters.statusFilter === 'depleted') {
				result = result.filter((g) => getComputedStatus(g) === 'depleted');
			}
		}
		result = applyCommonFilters(result, (g) => g.expires_at);
		if (walletFilters.searchInput.trim()) {
			const q = walletFilters.searchInput.trim().toLowerCase();
			result = result.filter(
				(g) =>
					g.merchant?.name.toLowerCase().includes(q) ||
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
		if (walletFilters.typeFilter === 'cards') return filteredCards;
		if (walletFilters.typeFilter === 'vouchers') return filteredVouchers;
		if (walletFilters.typeFilter === 'gift-cards') return filteredGiftCards;
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
		if (walletFilters.typeFilter !== lastTypeFilter) {
			if (selectMode) {
				selectedIds.clear();
			}
			// Reset status filter when type changes to avoid invalid combinations
			walletFilters.statusFilter = 'all';
			// Reset sort/expiring if switching to cards (no value/expiry)
			if (walletFilters.typeFilter === 'cards') {
				if (
					['value-desc', 'value-asc', 'expiry-asc'].includes(
						walletFilters.sortBy
					)
				) {
					walletFilters.sortBy = 'newest';
				}
				walletFilters.expiringFilter = 'all';
			}
			lastTypeFilter = walletFilters.typeFilter;
		}
	});

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
		showBarcodes =
			localStorage.getItem('savvy_wallet_show_barcodes') === 'true';
		const t = get(page).url.searchParams.get('type');
		if (t && validTypes.includes(t)) {
			walletFilters.typeFilter = t;
			lastTypeFilter = t;
		}
		await loadData();
		// List is now in the DOM — restore the saved scroll position.
		if (walletFilters.scrollY > 0) {
			await tick();
			window.scrollTo(0, walletFilters.scrollY);
		}
	});

	function toggleBarcodes() {
		showBarcodes = !showBarcodes;
		localStorage.setItem('savvy_wallet_show_barcodes', String(showBarcodes));
	}

	function resetFilters() {
		walletFilters.typeFilter = 'all';
		walletFilters.statusFilter = 'all';
		walletFilters.sortBy = 'newest';
		walletFilters.ownerFilter = 'all';
		walletFilters.favoritesOnly = false;
		walletFilters.expiringFilter = 'all';
		walletFilters.searchInput = '';
	}

	// Batch functions
	function toggleSelectMode() {
		selectMode = !selectMode;
		if (!selectMode) {
			selectedIds.clear();
			walletFilters.typeFilter = 'all';
		} else {
			showFilterMenu = false;
			// Batch endpoints are per-type; force a concrete type when entering select mode.
			if (walletFilters.typeFilter === 'all') {
				if (cards.length > 0) walletFilters.typeFilter = 'cards';
				else if (vouchers.length > 0) walletFilters.typeFilter = 'vouchers';
				else if (giftCards.length > 0) walletFilters.typeFilter = 'gift-cards';
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
				if (walletFilters.typeFilter === 'cards')
					result = await batchApi.deleteCards(ids);
				else if (walletFilters.typeFilter === 'vouchers')
					result = await batchApi.deleteVouchers(ids);
				else result = await batchApi.deleteGiftCards(ids);
			} else if (batchAction === 'share') {
				const req = {
					ids,
					email,
					can_edit: permissions.canEdit,
					can_delete: permissions.canDelete,
					...(walletFilters.typeFilter === 'gift-cards'
						? { can_edit_transactions: permissions.canEditTransactions }
						: {})
				};
				if (walletFilters.typeFilter === 'cards')
					result = await batchApi.shareCards(req);
				else if (walletFilters.typeFilter === 'vouchers')
					result = await batchApi.shareVouchers(req);
				else result = await batchApi.shareGiftCards(req);
			} else {
				if (walletFilters.typeFilter === 'cards')
					result = await batchApi.transferCards(ids, email);
				else if (walletFilters.typeFilter === 'vouchers')
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
			if (walletFilters.typeFilter === 'cards')
				result = await batchApi.exportCards(ids);
			else if (walletFilters.typeFilter === 'vouchers')
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
	<title>{tr('nav.wallet')} - {tr('common.appName')}</title>
</svelte:head>

<BatchConfirmModal
	action={batchAction}
	count={selectedCount}
	isOpen={showBatchModal}
	isLoading={batchLoading}
	onConfirm={executeBatchAction}
	onCancel={() => (showBatchModal = false)}
	hidePermissions={walletFilters.typeFilter === 'vouchers'}
	showTransactionPermission={walletFilters.typeFilter === 'gift-cards'}
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

<div class="px-4 max-w-7xl mx-auto pb-20 md:pb-4" class:pb-40={selectMode}>
	<!-- Header: count above title (mockup "7 Einträge"). -->
	<PageHeader
		eyebrow={`${totalItems} ${tr('dashboard.entries')}`}
		title={tr('nav.wallet')}
	/>

	{#if isLoading}
		<LoadingSpinner />
	{:else if totalItems === 0}
		<div class="bg-surface-1 rounded-lg p-12 text-center">
			<p class="text-text-muted text-lg mb-4">
				{tr('merchantOverview.detail.noItems')}
			</p>
			<p class="text-text-faint text-sm">
				{tr('merchantOverview.detail.noItemsHint')}
			</p>
		</div>
	{:else}
		<!-- Type filter: always visible (All · Cards · Vouchers · Gift). -->
		<div class="mb-4">
			<TypeFilterButtons
				bind:typeFilter={walletFilters.typeFilter}
				cardsCount={cards.length}
				vouchersCount={vouchers.length}
				giftCardsCount={giftCards.length}
				showAll={false}
			/>
		</div>

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

		<!-- WalletToolbar: ¼ Select · ¼ Filter · ½ Barcode-Toggle -->
		<div class="flex flex-col sm:flex-row gap-3 mb-6">
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
				<!-- Barcode Toggle Button (½ width emphasis) -->
				<button
					type="button"
					onclick={toggleBarcodes}
					class="flex items-center justify-center gap-2 h-[42px] px-6 bg-white border rounded-md hover:bg-surface-1 transition-colors {showBarcodes
						? 'ring-2 ring-accent border-accent'
						: 'border-border-field'}"
					aria-label={showBarcodes
						? tr('barcodeToggle.hide')
						: tr('barcodeToggle.show')}
					aria-pressed={showBarcodes}
				>
					<span class="text-sm font-medium text-text-muted whitespace-nowrap">
						{tr('barcodeToggle.label')}
					</span>
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
				<!-- Barcode Toggle Button (Mobile, ½ width) -->
				<button
					type="button"
					onclick={toggleBarcodes}
					class="flex-[2] flex items-center justify-center gap-2 h-[42px] bg-white border rounded-md hover:bg-surface-1 transition-colors {showBarcodes
						? 'ring-2 ring-accent border-accent'
						: 'border-border-field'}"
					aria-label={showBarcodes
						? tr('barcodeToggle.hide')
						: tr('barcodeToggle.show')}
					aria-pressed={showBarcodes}
				>
					<span class="text-sm font-medium text-text-muted whitespace-nowrap">
						{tr('barcodeToggle.label')}
					</span>
				</button>
			</div>
		</div>

		{#if totalFiltered === 0 && (walletFilters.searchInput || hasActiveFilters)}
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
									bind:typeFilter={walletFilters.typeFilter}
									bind:statusFilter={walletFilters.statusFilter}
									bind:sortBy={walletFilters.sortBy}
									bind:ownerFilter={walletFilters.ownerFilter}
									bind:favoritesOnly={walletFilters.favoritesOnly}
									bind:expiringFilter={walletFilters.expiringFilter}
									sortOptions={detailSortOptions}
									statusOptions={detailStatusOptions}
									cardsCount={cards.length}
									vouchersCount={vouchers.length}
									giftCardsCount={giftCards.length}
									{showStatusFilter}
									ownerOptions={detailOwnerOptions}
									expiringOptions={detailExpiringOptions}
									showExpiringFilter={walletFilters.typeFilter !== 'cards'}
									{hasActiveFilters}
									onReset={resetFilters}
									showAll={false}
									idPrefix="wallet-desktop"
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
				bind:typeFilter={walletFilters.typeFilter}
				bind:statusFilter={walletFilters.statusFilter}
				bind:sortBy={walletFilters.sortBy}
				bind:ownerFilter={walletFilters.ownerFilter}
				bind:favoritesOnly={walletFilters.favoritesOnly}
				bind:expiringFilter={walletFilters.expiringFilter}
				sortOptions={detailSortOptions}
				statusOptions={detailStatusOptions}
				cardsCount={cards.length}
				vouchersCount={vouchers.length}
				giftCardsCount={giftCards.length}
				{showStatusFilter}
				ownerOptions={detailOwnerOptions}
				expiringOptions={detailExpiringOptions}
				showExpiringFilter={walletFilters.typeFilter !== 'cards'}
				{hasActiveFilters}
				onReset={resetFilters}
				showAll={false}
				idPrefix="wallet-mobile"
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
