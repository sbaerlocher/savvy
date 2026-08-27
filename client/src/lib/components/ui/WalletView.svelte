<script lang="ts" module>
	import {
		ICON_BARCODE_TOGGLE,
		ICON_CHECK,
		ICON_CLIPBOARD_CHECK,
		ICON_CLOSE,
		ICON_FUNNEL,
		ICON_LINES
	} from '$lib/icons';
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
		giftCardToTileModel,
		type TileModel
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
	import { onDestroy, onMount, untrack } from 'svelte';
	import MerchantFilters from '$lib/components/MerchantFilters.svelte';
	import TypeFilterButtons from '$lib/components/TypeFilterButtons.svelte';
	import { SvelteSet } from 'svelte/reactivity';
	import { getGiftCardStatus } from '$lib/utils/resource-status';
	import {
		applyCommonFilters,
		matchesCardStatus,
		searchMerchant,
		sortItems
	} from '$lib/wallet/filter';
	import { platform } from '$lib/utils/platform';
	import { selectModeActive } from '$lib/stores/selectMode';

	// Each native platform renders its own chrome for this screen: Android an
	// outlined chip toolbar with a contextual top app bar in select mode, iOS a
	// segmented type filter with icon-only controls and a floating batch bar.
	// `platform` is a module constant, so plain consts, not $derived.
	const IS_ANDROID = platform === 'android';
	const IOS = platform === 'ios';
	// Desktop renders its own chrome for this screen: a single row above the grid
	// carrying count + title on the left and the actions on the right, no type
	// chip row (type lives in the filter panel).
	const IS_DESKTOP = platform === 'other';

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
		/** Android only: pull the barcode toggle out of the chip row and dock it
		 *  as a square icon button at the end of the type-chip row (mockup
		 *  screen-MerchantsAndroid, board 2). The wallet keeps the labelled chip. */
		androidBarcodeInTypeRow?: boolean;
		/** iOS phone toolbar shape. 'wallet' is the three-control row (select ·
		 *  filter · barcode). 'detail' is the merchant-detail row: the type pills
		 *  with their counts on their own row, with barcode, filter and select as
		 *  round buttons below (mockup screen-MerchantsIOS, Board 2). */
		iosToolbarVariant?: 'wallet' | 'detail';
		filterShowAll?: boolean;
		maxWidth?: boolean;
		/** Desktop mockup chrome: one row above the grid carrying count + title on
		 *  the left and the toolbar actions on the right, and no type chip row
		 *  (type is picked in the filter panel). Only the wallet screen opts in;
		 *  merchant detail keeps its own header snippet and the chip row. */
		desktopChrome?: boolean;
		/** Title for the desktop chrome row (ignored without `desktopChrome`). */
		chromeTitle?: string;
		/** Put the type chip row and the toolbar actions on one line, chips left
		 *  and actions right (mockup screen-MerchantsDesktop board 2). Merchant
		 *  detail opts in on desktop; the wallet has no chip row at all there. */
		inlineTypeToolbar?: boolean;
		header?: Snippet;
		/** Header variant for iOS select mode: the eyebrow counts the selection
		 *  instead of the total (mockup screen-WalletIOS, Phone 3). Falls back to
		 *  `header` when a call site does not provide one. */
		selectHeader?: Snippet<[number]>;
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
		androidBarcodeInTypeRow = false,
		iosToolbarVariant = 'wallet',
		filterShowAll = false,
		maxWidth = true,
		desktopChrome = false,
		chromeTitle = '',
		inlineTypeToolbar = false,
		header,
		selectHeader,
		emptyIcon,
		searchField
	}: Props = $props();

	// iOS wallet chrome (mockup screen-WalletIOS): three toolbar controls on the
	// 40px chip height — select and filter are square icon-only buttons, the
	// barcode toggle keeps its label and takes the rest of the row.
	// No background here: the active barcode state paints bg-accent, and a
	// base bg-surface would tie on specificity and win by source order.
	const IOS_TOOLBAR_BUTTON =
		'flex h-10 min-w-0 items-center justify-center gap-1.5 rounded-lg border text-body-sm font-semibold transition-colors';
	// Glyphs from the mockup's ui-WalletToolbar (checkbox, funnel, bar field).
	const ICON_SELECT_BOX = 'M9 12.5l2 2 4-4.5';
	const ICON_FUNNEL_SOLID = 'M4 5h16l-6 7v6l-4 2v-8z';
	const ICON_BARCODE_BARS = 'M4 6v12M7 6v12M10.5 6v12M14 6v12M17 6v12M20 6v12';

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

	// Desktop chrome is on only where the caller opted in AND the platform is
	// desktop. `platform` is a module constant, but the opt-in is a prop — so
	// $derived.
	const DESKTOP_CHROME = $derived(desktopChrome && IS_DESKTOP);
	// Opt-in is a prop, the platform is a module constant — so $derived.
	const BARCODE_IN_TYPE_ROW = $derived(androidBarcodeInTypeRow && IS_ANDROID);
	// The iOS merchant-detail chrome: its own toolbar shape and empty state.
	// Gated on the prop so the iOS wallet, which shares this component, is
	// untouched.
	const IOS_DETAIL = $derived(IOS && iosToolbarVariant === 'detail');

	// Same shape: the merged chip + toolbar row is a desktop-only treatment the
	// caller opts into.
	const INLINE_TYPE_TOOLBAR = $derived(inlineTypeToolbar && IS_DESKTOP);

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
		// Cards have no expiry, only a manual status (active/inactive/expired/…)
		result = result.filter((c) =>
			matchesCardStatus(c.status, filters.statusFilter)
		);
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
	const allFilteredSelected = $derived(
		currentFilteredItems.length > 0 &&
			selectedCount === currentFilteredItems.length
	);

	// Android replaces the chip row and toolbar with the contextual top app bar
	// while selecting — but only below `sm`, where that bar renders. Wider
	// viewports (a phone in landscape already is) keep the desktop chrome,
	// which holds the only other exit from select mode.
	const ANDROID_SELECT_HIDDEN = $derived(
		IS_ANDROID && selectMode ? 'max-sm:hidden' : ''
	);

	// Chrome row eyebrow/title, mirroring the three mockup boards: plain count,
	// "shown of total · filtered" once a filter narrows the list, and the
	// selection count while selecting.
	const chromeEyebrow = $derived(
		selectMode
			? tr('dashboard.selectionActive')
			: hasActiveFilters || filters.searchInput.trim()
				? tr('dashboard.entriesFiltered', {
						shown: totalFiltered,
						total: totalItems
					})
				: `${totalItems} ${tr('dashboard.entries')}`
	);
	const chromeHeading = $derived(
		selectMode ? tr('batch.selected', { count: selectedCount }) : chromeTitle
	);

	// Active type label on the filter button (mockup board B shows the button
	// carrying the picked type, not just the funnel glyph).
	const activeTypeLabel = $derived(
		filters.typeFilter === 'cards'
			? tr('merchantOverview.filterCards')
			: filters.typeFilter === 'vouchers'
				? tr('merchantOverview.filterVouchers')
				: filters.typeFilter === 'gift-cards'
					? tr('merchantOverview.filterGiftCards')
					: ''
	);

	// Share modal permission hints, derived from the actual selection (not the
	// type filter), since a selection can span categories.
	const selectionHasGiftCards = $derived(
		filteredGiftCards.some((g) => selectedIds.has(g.id))
	);
	const selectionHasVouchers = $derived(
		filteredVouchers.some((v) => selectedIds.has(v.id))
	);
	// The batch endpoints cap a request at 50 items and every type group is its
	// own request, so the number that can actually hit the cap is the largest
	// group — not the total selection.
	const largestSelectedGroup = $derived(
		Math.max(
			filteredCards.filter((c) => selectedIds.has(c.id)).length,
			filteredVouchers.filter((v) => selectedIds.has(v.id)).length,
			filteredGiftCards.filter((g) => selectedIds.has(g.id)).length
		)
	);
	const selectionOnlyVouchers = $derived(
		selectedIds.size > 0 &&
			selectionHasVouchers &&
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

	// Mirror select mode into the layout so the native platforms can swap their
	// navigation chrome: Android the whole bar, iOS the floating batch bar that
	// takes the nav's slot.
	$effect(() => {
		selectModeActive.set(selectMode);
	});

	// Leaving the screen mid selection would otherwise leave the nav bar hidden.
	onDestroy(() => selectModeActive.set(false));

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

<!-- Toolbar icon (select/filter buttons, desktop + mobile share the same 5x5 glyph). -->
{#snippet toolbarIcon(d: string, active = false)}
	<svg
		class="w-5 h-5 {active ? 'text-accent-hover' : 'text-text-muted'}"
		fill="none"
		stroke="currentColor"
		viewBox="0 0 24 24"
	>
		<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" {d} />
	</svg>
{/snippet}

<!-- Active-filter dot on the filter button (desktop + mobile). -->
{#snippet activeFilterDot()}
	{#if hasActiveFilters}
		<span class="absolute -top-1 -right-1 w-3 h-3 bg-accent rounded-full"
		></span>
	{/if}
{/snippet}

<!-- Android M3 chip toolbar: Select · Filter · Barcode as outlined chips that
     fill tonally once active (wallet mockup). Sized to content, not stretched. -->
{#snippet androidChip(
	label: string,
	d: string,
	isActive: boolean,
	onclick: () => void,
	disabled = false,
	showDot = false,
	expanded: boolean | undefined = undefined
)}
	<button
		type="button"
		{onclick}
		{disabled}
		aria-pressed={isActive}
		aria-expanded={expanded}
		class="relative inline-flex h-8 items-center gap-1.5 rounded-m3-sm px-3.5 text-label whitespace-nowrap transition-colors disabled:cursor-not-allowed disabled:opacity-50 {isActive
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
			<path stroke-linecap="round" stroke-linejoin="round" {d} />
		</svg>
		{label}
		{#if showDot}
			<span
				class="absolute -top-0.75 -right-0.75 h-2 w-2 rounded-m3-full border-2 border-paper bg-accent"
			></span>
		{/if}
	</button>
{/snippet}

<!-- Barcode-toggle button body: label variant vs inline barcode glyph (desktop + mobile). -->
{#snippet barcodeButtonContent()}
	{#if barcodeButtonVariant === 'label'}
		<span class="text-sm font-medium text-text-muted whitespace-nowrap">
			{tr('barcodeToggle.label')}
		</span>
	{:else}
		{@render toolbarIcon(ICON_BARCODE_TOGGLE)}
	{/if}
{/snippet}

<!-- iOS wallet toolbar (mockup ui-WalletToolbar): select · filter · barcode on
     one 40px row. Select and filter are icon-only (label lives in aria-label /
     title); the barcode toggle keeps its label and fills with the accent when
     on. -->
{#snippet iosDetailToolbar()}
	<!-- iOS merchant-detail toolbar (mockup screen-MerchantsIOS, Board 2): the
	     type pills with their counts, then the round controls. The mockup draws
	     only the barcode toggle, but the filter sheet still carries status, owner,
	     expiry, favourites and sort, and batch mode has no other phone entry
	     point — so both keep a button, on their own row so the pills stay
	     readable when a merchant carries all three types. -->
	<div class="mb-3 sm:hidden">
		<div class="scrollbar-none -mx-screen overflow-x-auto px-screen">
			<TypeFilterButtons
				bind:typeFilter={filters.typeFilter}
				cardsCount={cards.length}
				vouchersCount={vouchers.length}
				giftCardsCount={giftCards.length}
				showAll
				allowToggle={false}
				variant="count-pill"
			/>
		</div>
		<div class="mt-2 flex items-center justify-end gap-2">
			{@render iosDetailButton(
				toggleBarcodes,
				showBarcodes,
				tr('barcodeToggle.label'),
				iosBarcodeIcon
			)}
			{@render iosDetailButton(
				() => (showFilterMenu = !showFilterMenu),
				hasActiveFilters,
				tr('common.filter'),
				iosFunnelIcon
			)}
			{@render iosDetailButton(
				toggleSelectMode,
				selectMode,
				tr('batch.selectMode'),
				iosSelectIcon,
				isOffline
			)}
		</div>
	</div>
{/snippet}

{#snippet iosBarcodeIcon()}
	<path stroke-linecap="round" stroke-linejoin="round" d={ICON_BARCODE_BARS} />
{/snippet}

{#snippet iosFunnelIcon()}
	<path stroke-linecap="round" stroke-linejoin="round" d={ICON_FUNNEL_SOLID} />
{/snippet}

{#snippet iosSelectIcon()}
	<rect x="4" y="4" width="16" height="16" rx="3" />
	<path stroke-linecap="round" stroke-linejoin="round" d={ICON_SELECT_BOX} />
{/snippet}

{#snippet iosDetailButton(
	onclick: () => void,
	active: boolean,
	label: string,
	icon: Snippet,
	disabled = false
)}
	<button
		type="button"
		{onclick}
		{disabled}
		aria-pressed={active}
		aria-label={label}
		title={label}
		class="inline-flex h-8.5 w-8.5 shrink-0 items-center justify-center rounded-full transition-colors disabled:opacity-50 {active
			? 'bg-accent text-on-accent'
			: 'liquid-glass-card text-text-muted'}"
	>
		<svg
			class="h-4.25 w-4.25"
			fill="none"
			stroke="currentColor"
			stroke-width="1.9"
			viewBox="0 0 24 24"
		>
			{@render icon()}
		</svg>
	</button>
{/snippet}

{#snippet iosToolbar()}
	<div class="mb-3 flex gap-2 sm:hidden">
		<button
			type="button"
			onclick={toggleSelectMode}
			disabled={isOffline}
			aria-pressed={selectMode}
			aria-label={tr('batch.selectMode')}
			title={tr('batch.selectMode')}
			class="{IOS_TOOLBAR_BUTTON} w-13 shrink-0 disabled:opacity-50 {selectMode
				? 'border-accent text-accent-700'
				: 'border-border-chip bg-surface text-text-muted'}"
		>
			<svg
				class="h-5 w-5 shrink-0"
				fill="none"
				stroke="currentColor"
				stroke-width="1.9"
				viewBox="0 0 24 24"
			>
				<rect x="4" y="4" width="16" height="16" rx="3" />
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					d={ICON_SELECT_BOX}
				/>
			</svg>
		</button>

		<button
			type="button"
			onclick={(e: MouseEvent) => {
				e.stopPropagation();
				showFilterMenu = !showFilterMenu;
			}}
			aria-expanded={showFilterMenu}
			aria-label={tr('common.filter')}
			title={tr('common.filter')}
			class="{IOS_TOOLBAR_BUTTON} relative w-13 shrink-0 {hasActiveFilters
				? 'border-accent text-accent-700'
				: 'border-border-chip bg-surface text-text-muted'}"
		>
			<svg
				class="h-5 w-5 shrink-0"
				fill="none"
				stroke="currentColor"
				stroke-width="1.9"
				viewBox="0 0 24 24"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					d={ICON_FUNNEL_SOLID}
				/>
			</svg>
			{#if hasActiveFilters}
				<span
					class="absolute -top-0.75 -right-0.75 h-2 w-2 rounded-full border border-surface bg-accent"
				></span>
			{/if}
		</button>

		<button
			type="button"
			onclick={toggleBarcodes}
			aria-pressed={showBarcodes}
			class="{IOS_TOOLBAR_BUTTON} flex-1 {showBarcodes
				? 'border-accent bg-accent text-white'
				: 'border-border-chip bg-surface text-text-muted'}"
		>
			<svg
				class="h-3.75 w-3.75 shrink-0"
				fill="none"
				stroke="currentColor"
				stroke-width="1.9"
				viewBox="0 0 24 24"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					d={ICON_BARCODE_BARS}
				/>
			</svg>
			{tr('barcodeToggle.label')}
		</button>
	</div>
{/snippet}

<!-- Desktop toolbar: Import · Select · Filter · Barcodes. The mockup gives the
     filter button the picked type as its label and the barcode button an icon
     next to its label. Import leads the row rather than trailing it as in the
     mockup: the trailing slot changes width with the filter label and select
     state, which made the row shift under the cursor. -->
{#snippet importButton()}
	<button
		type="button"
		onclick={() => (showImportDialog = true)}
		disabled={isOffline}
		class="hidden sm:inline-flex btn btn-ghost whitespace-nowrap items-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed {DESKTOP_CHROME
			? 'control rounded-lg'
			: ''}"
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
		{#if DESKTOP_CHROME}
			<span class="text-label">{tr('settings.import.title')}</span>
		{/if}
	</button>
{/snippet}

{#snippet desktopActions()}
	<!-- Import leads the row only in the desktop chrome; elsewhere it keeps its
	     original trailing slot and `btn` sizing. -->
	{#if DESKTOP_CHROME}
		{@render importButton()}
	{/if}
	<!-- Select Mode Button -->
	<button
		type="button"
		onclick={toggleSelectMode}
		disabled={isOffline}
		class="flex items-center justify-center gap-2 control px-4 bg-white border hover:bg-surface-1 transition-colors disabled:opacity-50 disabled:cursor-not-allowed {DESKTOP_CHROME
			? 'rounded-lg'
			: 'rounded-md'} {selectMode
			? `ring-2 ring-accent border-accent${DESKTOP_CHROME ? ' text-accent-hover' : ''}`
			: 'border-border-field'}"
		title={tr('batch.selectMode')}
		aria-label={tr('batch.selectMode')}
		aria-pressed={selectMode}
	>
		{@render toolbarIcon(ICON_CLIPBOARD_CHECK, selectMode && DESKTOP_CHROME)}
	</button>
	<!-- Filter Button -->
	<button
		type="button"
		onclick={(e: MouseEvent) => {
			e.stopPropagation();
			showFilterMenu = !showFilterMenu;
		}}
		class="flex items-center justify-center gap-2 control px-4 bg-white border hover:bg-surface-1 transition-colors relative {DESKTOP_CHROME
			? 'rounded-lg'
			: 'rounded-md'} {hasActiveFilters && DESKTOP_CHROME
			? 'ring-2 ring-accent border-accent text-accent-hover'
			: 'border-border-field'}"
		title={tr('common.filter')}
		aria-label={DESKTOP_CHROME && activeTypeLabel
			? `${tr('common.filter')}: ${activeTypeLabel}`
			: tr('common.filter')}
		aria-expanded={showFilterMenu}
	>
		{@render toolbarIcon(ICON_FUNNEL, hasActiveFilters && DESKTOP_CHROME)}
		{#if DESKTOP_CHROME && activeTypeLabel}
			<span class="text-label whitespace-nowrap">{activeTypeLabel}</span>
		{/if}
		{@render activeFilterDot()}
	</button>
	<!-- Barcode Toggle Button -->
	<button
		type="button"
		onclick={toggleBarcodes}
		class="flex items-center justify-center gap-2 control {barcodeButtonVariant ===
		'label'
			? 'px-6'
			: 'px-4'} bg-white border hover:bg-surface-1 transition-colors {DESKTOP_CHROME
			? 'rounded-lg'
			: 'rounded-md'} {showBarcodes
			? 'ring-2 ring-accent border-accent'
			: 'border-border-field'}"
		title={showBarcodes ? tr('barcodeToggle.hide') : tr('barcodeToggle.show')}
		aria-label={showBarcodes
			? tr('barcodeToggle.hide')
			: tr('barcodeToggle.show')}
		aria-pressed={showBarcodes}
	>
		{#if DESKTOP_CHROME && barcodeButtonVariant === 'label'}
			<!-- Mockup pairs the barcode glyph with the label on desktop. -->
			{@render toolbarIcon(ICON_BARCODE_TOGGLE)}
		{/if}
		{@render barcodeButtonContent()}
	</button>
	{#if !DESKTOP_CHROME}
		{@render importButton()}
	{/if}
{/snippet}

<!-- Type chip row. Shared by the standalone row and the merged desktop row so
     the two can never drift in which types they offer. -->
{#snippet typeChipRow()}
	<TypeFilterButtons
		bind:typeFilter={filters.typeFilter}
		cardsCount={cards.length}
		vouchersCount={vouchers.length}
		giftCardsCount={giftCards.length}
		showAll={filterShowAll}
		allowToggle={!selectMode}
	/>
{/snippet}

<!-- One tile grid; the three resource lists render identical ResourceTiles. -->
{#snippet tileGrid(tiles: TileModel[])}
	{#each tiles as model (model.id)}
		<ResourceTile
			{model}
			showBarcode={showBarcodes}
			{selectMode}
			selected={selectedIds.has(model.id)}
			onSelect={toggleSelection}
			onShowBarcode={(item) => (barcodeModalItem = item)}
		/>
	{/each}
{/snippet}

<!-- MerchantFilters panel; desktop and mobile differ only by the idPrefix suffix
     and, on the Android sheet, by who owns the reset action. -->
{#snippet merchantFilterPanel(panelIdPrefix: string, hideReset = false)}
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
		allowTypeToggle={!selectMode}
		idPrefix={panelIdPrefix}
		{hideReset}
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
	mixedWithVouchers={selectionHasVouchers && !selectionOnlyVouchers}
	largestGroupCount={largestSelectedGroup}
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
	class="{IOS ? 'px-screen sm:px-4' : 'px-4'} {maxWidth
		? 'max-w-7xl mx-auto'
		: ''} pb-20 md:pb-4"
	class:pb-40={selectMode}
>
	{#if IS_ANDROID && selectMode}
		<!-- M3 contextual top app bar: replaces the page header while a selection
		     is active (wallet mockup). Fixed so it stays put while the list
		     scrolls under it, like the platform bar it mirrors. -->
		<!-- A region, not a toolbar: the bar mixes the selection count (a text
		     node) with its two controls, and both buttons are labelled already.
		     Below both Modal layers (backdrop z-55/z-70, panel z-60/z-80) so the
		     batch confirm dialog still covers it. -->
		<div
			class="fixed top-0 right-0 left-0 z-50 flex h-14 items-center justify-between bg-m3-secondary-container pr-3 pl-2 text-m3-on-secondary-container sm:hidden"
			role="region"
			aria-label={tr('batch.selectMode')}
		>
			<div class="flex items-center gap-2.5">
				<button
					type="button"
					onclick={toggleSelectMode}
					aria-label={tr('batch.exitSelectMode')}
					class="inline-flex h-10 w-10 items-center justify-center rounded-full"
				>
					<svg
						class="h-5.5 w-5.5"
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
				<!-- Mockup sets 17px/600 here; no type step matches, --text-subheading
				     (15.5px, 600) is the nearest without mono letter-spacing. -->
				<span class="text-subheading tabular-nums">
					{tr('batch.selected', { count: selectedCount })}
				</span>
			</div>
			<button
				type="button"
				onclick={allFilteredSelected ? deselectAll : selectAll}
				aria-label={allFilteredSelected
					? tr('batch.deselectAll')
					: tr('batch.selectAll')}
				class="inline-flex h-10 w-10 items-center justify-center rounded-full"
			>
				<svg
					class="h-5.25 w-5.25"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
					viewBox="0 0 24 24"
				>
					<path stroke-linecap="round" stroke-linejoin="round" d={ICON_LINES} />
				</svg>
			</button>
		</div>
		<!-- Spacer so the list starts below the fixed bar. -->
		<div class="h-14 sm:hidden"></div>
		<div class="hidden sm:block">{@render header?.()}</div>
	{:else if IOS && selectMode}
		<!-- iOS select-mode header row (mockup screen-WalletIOS, Phone 3):
		     select-all · count · Done, above the page title. The floating bar
		     below carries only the actions. -->
		<div class="mb-2.5 flex items-center justify-between sm:hidden">
			<button
				type="button"
				onclick={allFilteredSelected ? deselectAll : selectAll}
				class="text-[length:var(--text-code)] text-accent-700 transition-opacity active:opacity-60"
			>
				{allFilteredSelected ? tr('batch.deselectAll') : tr('batch.selectAll')}
			</button>
			<span
				class="text-[length:var(--text-label)] font-semibold tabular-nums text-text-subtle"
			>
				{selectedCount} / {currentFilteredItems.length}
			</span>
			<button
				type="button"
				onclick={toggleSelectMode}
				class="text-[length:var(--text-code)] font-bold text-accent-700 transition-opacity active:opacity-60"
			>
				{tr('common.done')}
			</button>
		</div>
		{#if selectHeader}
			{@render selectHeader(selectedCount)}
		{:else}
			{@render header?.()}
		{/if}
	{:else if !DESKTOP_CHROME}
		{@render header?.()}
	{/if}

	{#if DESKTOP_CHROME}
		<!-- Desktop chrome row (mockup): count + title left, actions right, one
		     line, directly above the grid. It replaces the `header` snippet
		     outright rather than hiding it per breakpoint — two headings in the
		     DOM would give the page two `h1`s, the hidden one first. Rendered
		     above the empty-state branch so an empty wallet still has a title. -->
		<!-- `flex-wrap`: the row is the only toolbar at every width now, and its
		     labelled buttons do not shrink — below `sm` they need a second line
		     rather than overflowing the page. -->
		<div class="mb-6 flex flex-wrap items-center gap-4">
			<div class="min-w-0 flex-1">
				<p class="text-eyebrow text-text-subtle uppercase">{chromeEyebrow}</p>
				<h1 class="mt-0.5 text-title text-text">{chromeHeading}</h1>
			</div>
			{#if totalItems > 0}
				<div class="flex items-center gap-3">
					{@render desktopActions()}
				</div>
			{/if}
		</div>
	{/if}

	{#if totalItems === 0}
		<!-- Android draws the empty state as a flat M3 card with the glyph above
		     the copy (mockup screen-MerchantsAndroid), the iOS merchant detail as
		     a glass grouped-inset card (mockup screen-MerchantsIOS, Board 2);
		     every other screen keeps the plain panel, the iOS wallet included. -->
		<div
			class="text-center {IOS_DETAIL
				? 'liquid-glass-card mt-2 flex flex-col items-center gap-2.5 rounded-[var(--radius-inset)] px-6.5 py-11 max-sm:mx-0 sm:mt-0 sm:block sm:rounded-lg sm:p-12'
				: 'bg-surface-1 rounded-lg p-12'} {IS_ANDROID
				? 'max-sm:flex max-sm:flex-col max-sm:items-center max-sm:gap-2 max-sm:rounded-m3-lg max-sm:bg-m3-card max-sm:px-6.5 max-sm:py-11.5'
				: ''}"
		>
			{@render emptyIcon?.()}
			<p
				class="text-text-muted {IOS_DETAIL
					? 'mt-1 text-subheading text-text sm:mb-4 sm:text-lg sm:text-text-muted'
					: 'text-lg mb-4'} {IS_ANDROID
					? 'max-sm:mt-1.5 max-sm:mb-0 max-sm:text-subheading'
					: ''}"
			>
				{tr('merchantOverview.detail.noItems')}
			</p>
			<p
				class="{IOS_DETAIL
					? 'max-w-62.5 text-body text-text-muted sm:max-w-none sm:text-sm sm:text-text-faint'
					: 'text-text-faint text-sm'} {IS_ANDROID ? 'max-sm:max-w-62.5' : ''}"
			>
				{tr('merchantOverview.detail.noItemsHint')}
			</p>
		</div>
	{:else}
		<!-- Android draws the search field above the type chips (mockup
		     screen-MerchantsAndroid, board 2); the other platforms keep the chip
		     row first. -->
		{#if IS_ANDROID}
			{@render searchField?.()}
		{/if}
		<!-- Type filter (top placement). The desktop mockup has no chip row — the
		     type is picked in the filter panel — so it drops out entirely there. -->
		{#if typeFilterPlacement === 'top' && !DESKTOP_CHROME && !INLINE_TYPE_TOOLBAR}
			<div
				class="mb-4 flex items-center {BARCODE_IN_TYPE_ROW
					? 'gap-2.5 sm:gap-0'
					: ''} {ANDROID_SELECT_HIDDEN}"
			>
				<div class="min-w-0 flex-1">
					{@render typeChipRow()}
				</div>
				{#if BARCODE_IN_TYPE_ROW}
					<!-- Square M3 icon button closing the chip row (mockup). Phone only:
					     from `sm` up the desktop action row renders its own barcode
					     toggle, and two controls for one state would collide. -->
					<button
						type="button"
						onclick={toggleBarcodes}
						aria-pressed={showBarcodes}
						aria-label={showBarcodes
							? tr('barcodeToggle.hide')
							: tr('barcodeToggle.show')}
						class="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-m3-sm transition-colors sm:hidden {showBarcodes
							? 'bg-m3-secondary-container text-m3-on-secondary-container'
							: 'border border-border-chip text-text-muted'}"
					>
						<svg
							class="h-4.25 w-4.25"
							fill="none"
							stroke="currentColor"
							stroke-width="1.9"
							stroke-linecap="round"
							viewBox="0 0 24 24"
							aria-hidden="true"
						>
							<path d={ICON_BARCODE_BARS} />
						</svg>
					</button>
				{/if}
			</div>
		{/if}

		{#if !IS_ANDROID}
			{@render searchField?.()}
		{/if}

		{#if INLINE_TYPE_TOOLBAR}
			<!-- Merged row (mockup board 2): chips left, toolbar actions right. -->
			<div class="mb-6 flex flex-wrap items-center gap-4">
				<div class="min-w-0 flex-1">
					{@render typeChipRow()}
				</div>
				<!-- Actions only from `sm` up: below that the mobile toolbar row
				     further down carries them, and rendering both would duplicate
				     every control. -->
				<div class="hidden items-center gap-3 sm:flex">
					{@render desktopActions()}
				</div>
			</div>
		{/if}

		<!-- WalletToolbar: Select · Filter · Barcode-Toggle · Import. Android hides
		     the block in select mode below `sm`, where the contextual top app bar
		     replaces it; on desktop the actions live in the chrome row instead. -->
		{#if IOS}
			{#if iosToolbarVariant === 'detail'}
				{@render iosDetailToolbar()}
			{:else}
				{@render iosToolbar()}
			{/if}
		{/if}
		{#if !DESKTOP_CHROME}
			<!-- iOS carries its own phone toolbar above, so this one starts at `sm`
			     there. `max-sm:hidden` rather than a hidden/flex pair: it does not
			     rely on the emission order inside the display-utility group. -->
			<div
				class="flex flex-col sm:flex-row gap-3 mb-6 {ANDROID_SELECT_HIDDEN} {IOS
					? 'max-sm:hidden'
					: ''} {INLINE_TYPE_TOOLBAR ? 'sm:hidden' : ''}"
			>
				<!-- Action Buttons (Desktop). Suppressed where the merged chip row
				     above already carries them, or they would render twice. -->
				{#if !INLINE_TYPE_TOOLBAR}
					<div class="hidden sm:flex gap-3">
						{@render desktopActions()}
					</div>
				{/if}

				<!-- Action Buttons (Mobile) -->
				{#if IS_ANDROID}
					<!-- M3 chip row: content-sized, left-aligned (mockup). -->
					<div class="flex gap-2 sm:hidden">
						{@render androidChip(
							tr('batch.selectMode'),
							ICON_CLIPBOARD_CHECK,
							selectMode,
							toggleSelectMode,
							isOffline
						)}
						<!-- Mockup draws the active filter chip with both the check glyph
						     and the dot badge; keep both, but keep the disclosure
						     semantics of the button this chip replaces. -->
						{@render androidChip(
							tr('common.filter'),
							hasActiveFilters ? ICON_CHECK : ICON_FUNNEL,
							hasActiveFilters,
							() => (showFilterMenu = !showFilterMenu),
							false,
							hasActiveFilters,
							showFilterMenu
						)}
						{#if !BARCODE_IN_TYPE_ROW}
							{@render androidChip(
								tr('barcodeToggle.label'),
								ICON_BARCODE_BARS,
								showBarcodes,
								toggleBarcodes
							)}
						{/if}
					</div>
				{:else}
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
							{@render toolbarIcon(ICON_CLIPBOARD_CHECK)}
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
							{@render toolbarIcon(ICON_FUNNEL)}
							{@render activeFilterDot()}
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
							{@render barcodeButtonContent()}
						</button>
					</div>
				{/if}
			</div>
		{/if}

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
							: 'lg:grid-cols-3'} {IS_ANDROID || IOS
							? 'gap-3 md:gap-6'
							: 'gap-6'}"
					>
						<!-- Cards -->
						{@render tileGrid(cardTiles)}

						<!-- Vouchers -->
						{@render tileGrid(voucherTiles)}

						<!-- Gift Cards -->
						{@render tileGrid(giftCardTiles)}
					</div>
				</div>

				<!-- Filter Side-Panel (Desktop only) -->
				{#if showFilterMenu && !selectMode}
					<div class="hidden lg:block lg:col-span-1">
						<div
							class="bg-white rounded-xl shadow-card sticky top-4 overflow-hidden"
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
												d={ICON_FUNNEL}
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
													d={ICON_CLOSE}
												></path>
											</svg>
										</button>
									</div>
								</div>
							</div>

							<div class="p-5">
								{@render merchantFilterPanel(`${idPrefix}-desktop`)}
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
	tonalAndroid
>
	<div class="{IS_ANDROID ? 'px-5' : 'px-4'} pb-4 pt-1">
		<div class="mb-3 flex items-center justify-between">
			<h3
				class="{IS_ANDROID
					? 'text-heading'
					: 'text-lg font-semibold'} text-text"
			>
				{tr('common.filter')}
			</h3>
			{#if IS_ANDROID}
				<!-- M3 sheet header: text reset action instead of a close glyph
				     (mockup); scrim tap and Esc still dismiss. Gated on active
				     filters like the in-panel button it replaces, so it never
				     offers a no-op reset. -->
				{#if hasActiveFilters}
					<button
						type="button"
						onclick={resetFilters}
						class="text-label text-accent-hover transition-colors"
					>
						{tr('common.resetFilters')}
					</button>
				{/if}
			{:else}
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
							d={ICON_CLOSE}
						/>
					</svg>
				</button>
			{/if}
		</div>

		<div class="pt-1">
			{@render merchantFilterPanel(`${idPrefix}-mobile`, IS_ANDROID)}
		</div>

		<div class={IS_ANDROID ? 'pt-4' : 'px-6 pb-6 pt-2'}>
			<button
				type="button"
				onclick={() => (showFilterMenu = false)}
				class={IS_ANDROID
					? 'w-full rounded-m3-full bg-accent px-4 py-3.5 text-body font-semibold text-white shadow-[var(--shadow-accent)] transition-colors'
					: 'w-full btn btn-primary'}
			>
				{IS_ANDROID
					? tr('common.showResults', { count: totalFiltered })
					: tr('common.done')}
			</button>
		</div>
	</div>
</BottomSheet>
