<script lang="ts">
	import { ICON_LOCK } from '$lib/icons';
	import type { Snippet } from 'svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { get } from 'svelte/store';
	import { authStore } from '$lib/stores/auth';
	import { t, locale } from '$lib/stores/i18n';
	import { offlineDB } from '$lib/stores/offline-db';
	import { toastStore } from '$lib/stores/toast';
	import { formatCurrency } from '$lib/utils/currency';
	import { platform } from '$lib/utils/platform';
	import { cardsApi, vouchersApi, giftCardsApi } from '$lib/api';
	import { resourceDetailPath } from '$lib/resource/routes';
	import DuplicateWarningBanner from '$lib/components/DuplicateWarningBanner.svelte';
	import BarcodeDisplay from '$lib/components/BarcodeDisplay.svelte';
	import ConfirmModal from '$lib/components/ConfirmModal.svelte';
	import TransferBox from '$lib/components/TransferBox.svelte';
	import EmailAutocomplete from '$lib/components/EmailAutocomplete.svelte';
	import SharePermissions from '$lib/components/SharePermissions.svelte';
	import ShareListItem from '$lib/components/ShareListItem.svelte';
	import ResourceActions from '$lib/components/ui/ResourceActions.svelte';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import { MERCHANT_DEFAULT_COLOR } from '$lib/utils/merchant-color';
	import {
		formatShareResult,
		shareResponseFromError
	} from '$lib/utils/share-result';
	import type {
		CardDTO,
		VoucherDTO,
		GiftCardDTO,
		ShareDTO,
		ShareCreateRequest,
		ShareCreateResponse
	} from '$lib/types/api';

	type Kind = 'card' | 'voucher' | 'gift_card';
	type ResourceDTO = CardDTO | VoucherDTO | GiftCardDTO;

	interface Props {
		kind: Kind;
		resource: ResourceDTO | null;
		shares: ShareDTO[];
		isOffline: boolean;
		/** 'editable' shows edit/delete permission checkboxes; 'readonly' (voucher). */
		shareMode?: 'editable' | 'readonly';
		/** Populate the route's bound edit fields; ResourceDetail then enters edit mode. */
		onStartEdit: () => void | Promise<void>;
		/**
		 * Per-kind edit form (field sets are disjoint → kept as a slot).
		 * Receives `cancel` (wire to the form's onCancel) and `close` (call from
		 * the route's saveEdit success path to leave edit mode).
		 */
		edit: Snippet<[{ cancel: () => void; close: () => void }]>;
		/** Gift-card only: balance + transactions ledger. */
		ledger?: Snippet;
	}

	let {
		kind,
		resource = $bindable(),
		shares = $bindable(),
		isOffline,
		shareMode = 'editable',
		onStartEdit,
		edit,
		ledger
	}: Props = $props();

	let isEditing = $state(false);

	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);

	// --- Kind config -------------------------------------------------------
	// Per-kind i18n key map + behaviour flags. Keys copied verbatim from the
	// three source routes — several are deliberately borrowed cross-namespace
	// (see the giftCards.* borrows below); do not "clean up".
	interface KindConfig {
		listPath: string;
		accentFallback: string;
		activeSentinel: string; // status value that counts as active/not-dimmed
		notesPresent: boolean;
		sharePermissions: boolean; // show SharePermissions in share create form
		showEditTransactions: boolean; // gift-card only
		favoriteAdd: string;
		favoriteRemove: string;
		deleteApi: (id: string) => Promise<unknown>;
		favoriteApi: (id: string) => Promise<{ is_favorite: boolean }>;
		offlineDelete: (id: string) => Promise<void>;
		i18n: Record<string, string>;
	}

	const CONFIG: Record<Kind, KindConfig> = {
		card: {
			listPath: '/cards',
			accentFallback: MERCHANT_DEFAULT_COLOR,
			activeSentinel: 'active',
			notesPresent: true,
			sharePermissions: true,
			showEditTransactions: false,
			favoriteAdd: 'common.addToFavorites',
			favoriteRemove: 'common.removeFromFavorites',
			deleteApi: (id) => cardsApi.delete(id),
			favoriteApi: (id) => cardsApi.toggleFavorite(id),
			offlineDelete: (id) => offlineDB.deleteCard(id),
			i18n: {
				titleFallback: 'common.card',
				sharedBy: 'cards.sharedBy',
				deleteButton: 'cards.deleteButton',
				deleteConfirm: 'cards.deleteConfirm',
				deleteConfirmMessage: 'cards.deleteConfirmMessage',
				deleteSuccess: 'cards.deleteSuccess',
				deleteError: 'cards.deleteError',
				shareTitle: 'common.share',
				shareAddButton: 'common.add',
				shareUserEmail: 'cards.sharing.userEmail',
				shareHint: 'giftCards.sharing.userMustBeRegistered',
				shareNow: 'giftCards.sharing.shareNow',
				shareError: 'cards.sharing.shareError',
				updateSuccess: 'cards.sharing.updateSuccess',
				updateError: 'cards.sharing.updateError',
				removeSuccess: 'cards.sharing.removeSuccess',
				removeError: 'cards.sharing.removeError',
				removeConfirm: 'cards.sharing.removeConfirm',
				removeConfirmMessage: 'cards.sharing.removeConfirmMessage',
				revokeAll: 'cards.sharing.revokeAll',
				revokeAllConfirm: 'cards.sharing.revokeAllConfirm',
				revokeAllConfirmMessage: 'cards.sharing.revokeAllConfirmMessage',
				revokeAllSuccess: 'cards.sharing.revokeAllSuccess',
				revokeAllError: 'cards.sharing.revokeAllError',
				notSharedYet: 'giftCards.sharing.notSharedYet',
				canEdit: 'cards.sharing.canEdit',
				canEditDesc: 'cards.sharing.canEditDesc',
				canDelete: 'cards.sharing.canDelete',
				canDeleteDesc: 'cards.sharing.canDeleteDesc',
				whatIsShared: 'cards.sharing.whatIsShared',
				transferButton: 'cards.transfer.button',
				transferTransferButton: 'cards.transfer.transferButton',
				transferWarning: 'cards.transfer.warning',
				transferWarningDetails: 'giftCards.transfer.warningDetails',
				transferEmailLabel: 'cards.transfer.newOwnerEmail',
				transferEmailHint: 'giftCards.sharing.userMustBeRegistered',
				transferWhatHappens: 'cards.transfer.whatHappens',
				transferConfirmTitle: 'cards.transfer.confirmTitle',
				transferConfirmMessage: 'cards.transfer.confirmMessage',
				transferSuccess: 'cards.transfer.success',
				transferError: 'cards.transfer.error'
			}
		},
		voucher: {
			listPath: '/vouchers',
			accentFallback: MERCHANT_DEFAULT_COLOR,
			activeSentinel: 'valid',
			notesPresent: false,
			sharePermissions: false,
			showEditTransactions: false,
			favoriteAdd: 'common.addToFavorites',
			favoriteRemove: 'common.removeFromFavorites',
			deleteApi: (id) => vouchersApi.delete(id),
			favoriteApi: (id) => vouchersApi.toggleFavorite(id),
			offlineDelete: (id) => offlineDB.deleteVoucher(id),
			i18n: {
				titleFallback: 'vouchers.title',
				sharedBy: 'vouchers.sharedBy',
				deleteButton: 'vouchers.deleteButton',
				deleteConfirm: 'vouchers.deleteConfirm',
				deleteConfirmMessage: 'vouchers.deleteConfirmMessage',
				deleteSuccess: 'vouchers.deleteSuccess',
				deleteError: 'vouchers.deleteError',
				shareTitle: 'common.share',
				shareAddButton: 'common.add',
				shareUserEmail: 'vouchers.sharing.userEmail',
				shareHint: 'vouchers.sharing.hint',
				shareNow: 'vouchers.sharing.shareNow',
				shareError: 'vouchers.sharing.shareError',
				updateSuccess: 'vouchers.sharing.updateSuccess',
				updateError: 'vouchers.sharing.updateError',
				removeSuccess: 'vouchers.sharing.removeSuccess',
				removeError: 'vouchers.sharing.removeError',
				removeConfirm: 'vouchers.sharing.removeConfirm',
				removeConfirmMessage: 'vouchers.sharing.removeConfirmMessage',
				revokeAll: 'vouchers.sharing.revokeAll',
				revokeAllConfirm: 'vouchers.sharing.revokeAllConfirm',
				revokeAllConfirmMessage: 'vouchers.sharing.revokeAllConfirmMessage',
				revokeAllSuccess: 'vouchers.sharing.revokeAllSuccess',
				revokeAllError: 'vouchers.sharing.revokeAllError',
				notSharedYet: 'vouchers.sharing.notSharedYet',
				whatIsShared: 'vouchers.sharing.whatIsShared',
				readOnlyNote: 'vouchers.sharing.readOnlyNote',
				manage: 'common.manage',
				removeShare: 'vouchers.sharing.removeShare',
				alwaysReadOnly: 'vouchers.sharing.alwaysReadOnly',
				canOnlyRemove: 'vouchers.sharing.canOnlyRemove',
				sharedCode: 'vouchers.sharing.sharedCode',
				sharedDetails: 'vouchers.sharing.sharedDetails',
				sharedDescription: 'vouchers.sharing.sharedDescription',
				transferButton: 'vouchers.transfer.button',
				transferTransferButton: 'vouchers.transfer.transferButton',
				transferWarning: 'vouchers.transfer.warning',
				transferWarningDetails: 'vouchers.transfer.warningDetail',
				transferEmailLabel: 'vouchers.transfer.newOwner',
				transferEmailHint: 'giftCards.sharing.userMustBeRegistered',
				transferWhatHappens: 'vouchers.transfer.whatHappens',
				transferConfirmTitle: 'vouchers.transfer.confirmTitle',
				transferConfirmMessage: 'vouchers.transfer.confirmMessage',
				transferSuccess: 'vouchers.transfer.success',
				transferError: 'vouchers.transfer.error'
			}
		},
		gift_card: {
			listPath: '/gift-cards',
			accentFallback: MERCHANT_DEFAULT_COLOR,
			activeSentinel: 'active',
			notesPresent: true,
			sharePermissions: true,
			showEditTransactions: true,
			favoriteAdd: 'common.addFavorite',
			favoriteRemove: 'common.removeFavorite',
			deleteApi: (id) => giftCardsApi.delete(id),
			favoriteApi: (id) => giftCardsApi.toggleFavorite(id),
			offlineDelete: (id) => offlineDB.deleteGiftCard(id),
			i18n: {
				titleFallback: 'giftCards.title',
				sharedBy: 'giftCards.sharedBy',
				notFound: 'giftCards.notFound',
				backToList: 'giftCards.backToList',
				deleteButton: 'giftCards.deleteButton',
				deleteConfirm: 'giftCards.deleteConfirm',
				deleteConfirmMessage: 'giftCards.deleteConfirmMessage',
				deleteSuccess: 'giftCards.deleteSuccess',
				deleteError: 'giftCards.deleteError',
				shareTitle: 'giftCards.sharing.title',
				shareAddButton: 'giftCards.sharing.addButton',
				shareUserEmail: 'giftCards.sharing.userEmail',
				shareHint: 'giftCards.sharing.userMustBeRegistered',
				shareNow: 'giftCards.sharing.shareNow',
				shareError: 'giftCards.sharing.shareError',
				updateSuccess: 'giftCards.sharing.updateSuccess',
				updateError: 'giftCards.sharing.updateError',
				removeSuccess: 'giftCards.sharing.removeSuccess',
				removeError: 'giftCards.sharing.removeError',
				removeConfirm: 'giftCards.sharing.removeConfirm',
				removeConfirmMessage: 'giftCards.sharing.removeConfirmMessage',
				revokeAll: 'giftCards.sharing.revokeAll',
				revokeAllConfirm: 'giftCards.sharing.revokeAllConfirm',
				revokeAllConfirmMessage: 'giftCards.sharing.revokeAllConfirmMessage',
				revokeAllSuccess: 'giftCards.sharing.revokeAllSuccess',
				revokeAllError: 'giftCards.sharing.revokeAllError',
				notSharedYet: 'giftCards.sharing.notSharedYet',
				canEdit: 'giftCards.sharing.canEdit',
				canEditDesc: 'giftCards.sharing.canEditDesc',
				canDelete: 'giftCards.sharing.canDelete',
				canDeleteDesc: 'giftCards.sharing.canDeleteDesc',
				canManageTransactions: 'giftCards.sharing.canManageTransactions',
				canManageTransactionsDesc:
					'giftCards.sharing.canManageTransactionsDesc',
				whatIsShared: 'giftCards.sharing.whatIsShared',
				transferTitle: 'giftCards.transfer.title',
				transferTransferButton: 'giftCards.transfer.transferButton',
				transferWarning: 'giftCards.transfer.warning',
				transferWarningDetails: 'giftCards.transfer.warningDetails',
				transferEmailLabel: 'giftCards.transfer.newOwnerEmail',
				transferEmailHint: 'giftCards.sharing.userMustBeRegistered',
				transferWhatHappens: 'giftCards.transfer.whatHappens',
				transferConfirmTitle: 'giftCards.transfer.confirmTitle',
				transferConfirmMessage: 'giftCards.transfer.confirmMessage',
				transferSuccess: 'giftCards.transfer.success',
				transferError: 'giftCards.transfer.error'
			}
		}
	};

	const cfg = $derived(CONFIG[kind]);
	const c = $derived(cfg.i18n);

	// Not-found branch keys. gift/voucher have a `.notFound` string; cards has
	// none, so fall back to the generic title. Back button reuses the shared
	// common.backToOverview (exists in all locales) — no new i18n keys needed.
	const notFoundKeys = $derived({
		title:
			kind === 'gift_card'
				? 'giftCards.notFound'
				: kind === 'voucher'
					? 'vouchers.notFound'
					: 'common.card',
		back: 'common.backToOverview'
	});

	// --- Typed accessors for the union DTO ---------------------------------
	const asCard = $derived(kind === 'card' ? (resource as CardDTO) : null);
	const asVoucher = $derived(
		kind === 'voucher' ? (resource as VoucherDTO) : null
	);
	const asGift = $derived(
		kind === 'gift_card' ? (resource as GiftCardDTO) : null
	);

	// Identifier value used for the barcode (card_number / code).
	const barcodeValue = $derived(
		asVoucher
			? asVoucher.code
			: ((resource as CardDTO | GiftCardDTO | null)?.card_number ?? '')
	);

	const status = $derived(resource?.status);
	const isDimmed = $derived(!!status && status !== cfg.activeSentinel);
	const merchantName = $derived(resource?.merchant?.name);
	const accentColor = $derived(resource?.merchant?.color || cfg.accentFallback);

	// Header eyebrow: card → program; voucher → single/multi-use label; gift → none.
	const eyebrow = $derived.by(() => {
		if (asCard) return asCard.program || undefined;
		if (asVoucher)
			return asVoucher.usage_limit_type === 'single_use'
				? tr('vouchers.singleUseOnly')
				: tr('vouchers.multipleUse');
		return undefined;
	});

	const pageTitle = $derived(merchantName || tr(c.titleFallback));
	const notes = $derived(
		cfg.notesPresent
			? (resource as CardDTO | GiftCardDTO | null)?.notes
			: undefined
	);

	function formatVoucherValue(
		value: number,
		type: string,
		currency?: string
	): string {
		if (type === 'percentage') return `${value}%`;
		if (type === 'fixed_amount')
			return formatCurrency(value, currency || 'CHF', $locale);
		if (type === 'points_multiplier')
			return `${value}${tr('vouchers.types.pointsMultiplierDisplay').trim()}`;
		if (type === 'bonus_points')
			return `+${value}${tr('vouchers.types.bonusPointsDisplay')}`;
		if (type === 'free') return tr('vouchers.types.freeDisplay');
		return `${value.toFixed(2)}`;
	}

	// Status badge sets differ per kind → per-kind map.
	function getStatusBadge(s: string): { class: string; text: string } {
		if (kind === 'card') {
			switch (s) {
				case 'inactive':
					return {
						class: 'bg-border text-text-ink2',
						text: tr('cards.status.inactive')
					};
				case 'expired':
					return {
						class: 'bg-danger-200 text-danger-700',
						text: tr('cards.status.expired')
					};
				case 'lost':
					return {
						class: 'bg-warning-200 text-warning-700',
						text: tr('cards.status.lost')
					};
				case 'blocked':
					return {
						class: 'bg-danger-200 text-danger-700',
						text: tr('cards.status.blocked')
					};
				default:
					return { class: '', text: '' };
			}
		}
		if (kind === 'voucher') {
			switch (s) {
				case 'inactive':
					return {
						class: 'bg-border text-text-ink2',
						text: tr('vouchers.status.inactive')
					};
				case 'expired':
					return {
						class: 'bg-danger-200 text-danger-700',
						text: tr('vouchers.status.expired')
					};
				case 'used':
					return {
						class: 'bg-success-200 text-success-700',
						text: tr('vouchers.status.used')
					};
				default:
					return { class: '', text: '' };
			}
		}
		switch (s) {
			case 'inactive':
				return {
					class: 'bg-border text-text-ink2',
					text: tr('giftCards.status.inactive')
				};
			case 'expired':
				return {
					class: 'bg-danger-200 text-danger-700',
					text: tr('giftCards.status.expired')
				};
			case 'depleted':
				return {
					class: 'bg-warning-200 text-warning-700',
					text: tr('giftCards.status.depleted')
				};
			default:
				return { class: '', text: '' };
		}
	}

	const statusBadge = $derived(
		isDimmed && status ? getStatusBadge(status) : undefined
	);

	// BarcodeDisplay prop adapter per kind (the component absorbs all fields).
	const barcodeProps = $derived.by(() => {
		if (asVoucher) {
			return {
				validFrom: asVoucher.valid_from,
				validUntil: asVoucher.valid_until,
				displayValue: formatVoucherValue(
					asVoucher.value,
					asVoucher.type,
					asVoucher.currency
				),
				minPurchaseInfo:
					asVoucher.min_purchase_amount && asVoucher.min_purchase_amount > 0
						? `${tr('vouchers.minPurchaseAmount')}: ${formatCurrency(asVoucher.min_purchase_amount, asVoucher.currency || 'CHF', $locale)}`
						: undefined,
				description: asVoucher.description
			};
		}
		if (asGift) {
			return {
				pin: asGift.pin,
				balance: asGift.current_balance.toFixed(2),
				currency: asGift.currency,
				expiresAt: asGift.expires_at
			};
		}
		return {};
	});

	// --- Share create form state ------------------------------------------
	let showShareForm = $state(false);
	let shareEmails = $state<string[]>([]);
	let canEdit = $state(false);
	let canDelete = $state(false);
	let canEditTransactions = $state(false);

	// Share editing state (editable kinds: cards, gift-cards).
	let editingShareId = $state<string | null>(null);
	let editShareCanEdit = $state(false);
	let editShareCanDelete = $state(false);
	let editShareCanEditTransactions = $state(false);
	// Voucher "manage" mode (delete-only, no permission editing).
	let managingShareId = $state<string | null>(null);

	let transferEmail = $state('');

	// Modal state
	let showDeleteModal = $state(false);
	let showDeleteShareModal = $state(false);
	let showRevokeAllModal = $state(false);
	let showTransferModal = $state(false);
	let shareToDelete: string | null = null;

	let isTogglingFavorite = $state(false);

	const shareApi = $derived(
		kind === 'card' ? cardsApi : kind === 'voucher' ? vouchersApi : giftCardsApi
	);

	// Back: return to where the user came from; fall back to wallet on deep link.
	function goBack() {
		if (history.length > 1) history.back();
		else goto(resolve('/wallet'));
	}

	// resolve() needs literal-prefixed routes for its typed-route check, so branch
	// on kind rather than interpolate cfg.listPath (which types as `string`).
	function listHref(): string {
		return kind === 'card'
			? resolve('/cards')
			: kind === 'voucher'
				? resolve('/vouchers')
				: resolve('/gift-cards');
	}

	function detailHref(id: string): string {
		return resourceDetailPath(kind, id);
	}

	// eslint can't see through the Href helpers that resolve() the target, so
	// wrap the navigations in handlers with a scoped disable.
	function goToDetail(id: string) {
		// eslint-disable-next-line svelte/no-navigation-without-resolve -- detailHref() already returns a resolve()'d path
		goto(detailHref(id));
	}

	function goToList() {
		// eslint-disable-next-line svelte/no-navigation-without-resolve -- listHref() already returns a resolve()'d path
		goto(listHref());
	}

	async function startEdit() {
		// Route populates its bound edit fields (merchants load + field copy),
		// then we flip into edit mode which renders the `edit` snippet.
		await onStartEdit();
		isEditing = true;
	}

	function cancelEdit() {
		isEditing = false;
	}

	async function toggleFavorite() {
		if (isTogglingFavorite || !resource) return;
		isTogglingFavorite = true;
		try {
			const response = await cfg.favoriteApi(resource.id);
			// Update favorite state directly from POST response (avoids stale SW GET).
			resource = { ...resource, is_favorite: response.is_favorite };
		} catch {
			toastStore.error(tr('common.favoriteError'));
		} finally {
			isTogglingFavorite = false;
		}
	}

	function promptDelete() {
		showDeleteModal = true;
	}

	async function confirmDelete() {
		if (!resource) return;
		try {
			await cfg.deleteApi(resource.id);
			toastStore.success(tr(c.deleteSuccess));
			// Force full page reload to refresh the list (SPA navigation caches data).
			window.location.href = cfg.listPath;
		} catch {
			toastStore.error(tr(c.deleteError));
		}
	}

	async function loadShares() {
		if (!resource?.permissions?.is_owner) return;
		try {
			const response = await shareApi.get(resource.id);
			shares = response.shares || [];
		} catch (err) {
			// Non-fatal: keep existing shares.
			console.error('Failed to load shares:', err);
		}
	}

	function buildSharePayload(): ShareCreateRequest {
		if (kind === 'voucher') return { emails: shareEmails };
		if (kind === 'gift_card')
			return {
				emails: shareEmails,
				can_edit: canEdit,
				can_delete: canDelete,
				can_edit_transactions: canEditTransactions
			};
		return { emails: shareEmails, can_edit: canEdit, can_delete: canDelete };
	}

	function resetShareForm() {
		showShareForm = false;
		shareEmails = [];
		canEdit = false;
		canDelete = false;
		canEditTransactions = false;
	}

	async function handleShare() {
		if (shareEmails.length === 0 || !resource) return;
		try {
			const response: ShareCreateResponse = await shareApi.createShare(
				resource.id,
				buildSharePayload()
			);
			shares = response.shares || [];
			const { message, isError } = formatShareResult(response, tr);
			if (isError) toastStore.error(message);
			else toastStore.success(message);
			resetShareForm();
		} catch (err: unknown) {
			const failed = shareResponseFromError(err);
			if (failed) {
				shares = failed.shares || shares;
				toastStore.error(formatShareResult(failed, tr).message);
			} else {
				toastStore.error(err instanceof Error ? err.message : tr(c.shareError));
			}
		}
	}

	function startEditShare(share: ShareDTO) {
		editingShareId = share.shared_with_user.id;
		editShareCanEdit = share.can_edit;
		editShareCanDelete = share.can_delete;
		editShareCanEditTransactions = share.can_edit_transactions || false;
	}

	function cancelEditShare() {
		editingShareId = null;
		editShareCanEdit = false;
		editShareCanDelete = false;
		editShareCanEditTransactions = false;
	}

	async function saveShareEdit(sharedWithID: string) {
		try {
			// updateShare only exists on editable kinds (cards, gift-cards).
			if (!resource || !('updateShare' in shareApi)) return;
			const response = await shareApi.updateShare(resource.id, sharedWithID, {
				can_edit: editShareCanEdit,
				can_delete: editShareCanDelete,
				...(cfg.showEditTransactions
					? { can_edit_transactions: editShareCanEditTransactions }
					: {})
			});
			shares = response.shares || [];
			editingShareId = null;
			toastStore.success(tr(c.updateSuccess));
		} catch (err: unknown) {
			toastStore.error(err instanceof Error ? err.message : tr(c.updateError));
		}
	}

	function startManageShare(shareId: string) {
		managingShareId = shareId;
	}

	function cancelManageShare() {
		managingShareId = null;
	}

	function promptDeleteShare(sharedWithID: string) {
		shareToDelete = sharedWithID;
		showDeleteShareModal = true;
	}

	async function confirmDeleteShare() {
		if (!shareToDelete || !resource) return;
		try {
			await shareApi.deleteShare(resource.id, shareToDelete);
			toastStore.success(tr(c.removeSuccess));
			managingShareId = null;
			showDeleteShareModal = false;
			await loadShares();
		} catch {
			toastStore.error(tr(c.removeError));
		} finally {
			shareToDelete = null;
			showDeleteShareModal = false;
		}
	}

	function promptRevokeAll() {
		showRevokeAllModal = true;
	}

	async function confirmRevokeAll() {
		if (!resource) return;
		try {
			await shareApi.deleteAllShares(resource.id);
			toastStore.success(tr(c.revokeAllSuccess));
			managingShareId = null;
			await loadShares();
		} catch {
			toastStore.error(tr(c.revokeAllError));
		} finally {
			showRevokeAllModal = false;
		}
	}

	function promptTransfer() {
		showTransferModal = true;
	}

	async function confirmTransfer() {
		if (!resource) return;
		try {
			await shareApi.transfer(resource.id, { new_owner_email: transferEmail });
			toastStore.success(tr(c.transferSuccess));
			// Remove from cache before redirect (prevents 403 on stale cache).
			await cfg.offlineDelete(resource.id);
			// Force full page reload (user lost access after transfer).
			window.location.href = cfg.listPath;
		} catch (err: unknown) {
			toastStore.error(
				err instanceof Error ? err.message : tr(c.transferError)
			);
		}
	}

	// Transfer details list per kind (cards/voucher share the same 4 keys under
	// their own namespace; gift uses detail1..4).
	const transferDetails = $derived.by(() => {
		if (kind === 'gift_card')
			return [
				tr('giftCards.transfer.detail1'),
				tr('giftCards.transfer.detail2'),
				tr('giftCards.transfer.detail3'),
				tr('giftCards.transfer.detail4')
			];
		const ns = kind === 'card' ? 'cards' : 'vouchers';
		return [
			tr(`${ns}.transfer.newOwnerGetsRights`),
			tr(`${ns}.transfer.allSharesDeleted`),
			tr(`${ns}.transfer.youLoseAccess`),
			tr(`${ns}.transfer.transferLogged`)
		];
	});

	const LOCK_PATH = ICON_LOCK;
</script>

{#if resource}
	<!-- Page header (view mode only; the edit form keeps its own title). -->
	{#if !isEditing}
		<PageHeader
			title={pageTitle}
			{eyebrow}
			mobileActions={false}
			showSearch
			onBack={goBack}
		>
			{#snippet actions()}
				<ResourceActions
					{isOffline}
					isFavorite={resource!.is_favorite}
					{isTogglingFavorite}
					canEdit={resource!.permissions?.can_edit}
					favoriteTitleAdd={tr(cfg.favoriteAdd)}
					favoriteTitleRemove={tr(cfg.favoriteRemove)}
					ontoggleFavorite={toggleFavorite}
					onstartEdit={startEdit}
				/>
			{/snippet}
		</PageHeader>
		{#if resource.owner && resource.owner.id !== $authStore.user?.id}
			<p class="-mt-6 mb-6 text-xs text-text-faint">
				{tr(c.sharedBy, {
					name: resource.owner.first_name || resource.owner.email
				})}
			</p>
		{/if}

		<!-- Android M3: edit is a bottom-right FAB. It opens the same edit mode by
	     toggling the form; the route renders the form via the edit snippet, but
	     the trigger lives here. Uses startEdit which the route overrides through
	     ResourceActions above — the FAB simply mirrors it. -->
		{#if resource.permissions?.can_edit && platform === 'android'}
			<button
				type="button"
				onclick={startEdit}
				disabled={isOffline}
				aria-label={tr('common.edit')}
				class="sm:hidden fixed right-4 z-50 h-14 w-14 flex items-center justify-center rounded-2xl bg-accent text-white shadow-lg mobile-nav-fab disabled:opacity-50 disabled:pointer-events-none"
			>
				<svg
					class="w-6 h-6"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
					viewBox="0 0 24 24"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
					/>
				</svg>
			</button>
		{/if}
	{/if}

	<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
		<!-- Left column: resource details / edit form -->
		<div class="lg:col-span-2">
			{#if !isEditing}
				<!-- View Mode -->
				<div
					class="overflow-hidden rounded-xl border border-border/80 bg-white"
					style="border-left: 3px solid color-mix(in srgb, {accentColor} 70%, transparent)"
				>
					<div class="p-6 {isDimmed ? 'opacity-50 grayscale' : ''}">
						{#if resource.duplicate_warning}
							<div class="mb-6">
								<DuplicateWarningBanner
									warning={resource.duplicate_warning}
									resourceType={kind}
									onNavigate={goToDetail}
								/>
							</div>
						{/if}

						<BarcodeDisplay
							value={barcodeValue}
							type={resource.barcode_type || 'CODE128'}
							{status}
							{statusBadge}
							{...barcodeProps}
						/>

						{#if notes}
							<div
								class="mt-4 bg-warning-50 border-l-4 border-warning-400 p-3 rounded"
							>
								<p class="text-sm text-text-ink2">{notes}</p>
							</div>
						{/if}
					</div>
				</div>
			{:else}
				<!-- Edit Mode: per-kind form slot + shared delete button. -->
				<div class="overflow-hidden rounded-xl border border-border bg-white">
					<div class="p-6">
						{@render edit({ cancel: cancelEdit, close: cancelEdit })}
						{#if resource.permissions?.can_delete}
							<div class="pt-4 mt-4 border-t border-border">
								<button
									type="button"
									onclick={promptDelete}
									disabled={isOffline}
									class="btn btn-text-danger w-full flex items-center justify-center gap-1.5 {isOffline
										? 'pointer-events-none blur-[0.5px]'
										: ''}"
								>
									{#if isOffline}
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
												d={LOCK_PATH}
											></path>
										</svg>
									{/if}
									{tr(c.deleteButton)}
								</button>
							</div>
						{/if}
					</div>
				</div>
			{/if}
		</div>

		<!-- Right column: ledger (gift), transfer & sharing (owners) -->
		<div class="lg:col-span-1 space-y-4">
			{#if ledger}
				{@render ledger()}
			{/if}

			{#if resource.permissions?.is_owner}
				<!-- Transfer Box -->
				<TransferBox
					{isOffline}
					title={kind === 'gift_card' ? tr(c.transferTitle) : undefined}
					openButtonLabel={kind === 'gift_card'
						? `→ ${tr(c.transferTransferButton)}`
						: tr(c.transferButton)}
					transferButtonLabel={tr(c.transferTransferButton)}
					warningTitle={tr(c.transferWarning)}
					warningDetails={tr(c.transferWarningDetails)}
					emailLabel={tr(c.transferEmailLabel)}
					emailHint={tr(c.transferEmailHint)}
					whatHappensLabel={tr(c.transferWhatHappens)}
					details={transferDetails}
					bind:email={transferEmail}
					ontransfer={promptTransfer}
				/>

				<!-- Sharing Box -->
				<div class="rounded-xl border border-border bg-white p-6">
					<div class="flex justify-between items-center mb-4">
						<h3 class="text-lg font-semibold text-text">
							{tr(c.shareTitle)}
						</h3>
						{#if !showShareForm}
							<button
								onclick={() => (showShareForm = true)}
								disabled={isOffline}
								class="btn btn-xs btn-primary whitespace-nowrap flex items-center gap-1.5 {isOffline
									? 'pointer-events-none blur-[0.5px]'
									: ''}"
							>
								{#if isOffline}
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
											d={LOCK_PATH}
										></path>
									</svg>
								{:else}
									<span>+</span>
								{/if}
								{tr(c.shareAddButton)}
							</button>
						{/if}
					</div>

					{#if showShareForm}
						<div
							class="border border-accent-200 bg-accent-50 rounded-lg p-4 space-y-4 mb-4"
						>
							<EmailAutocomplete
								multiple
								bind:values={shareEmails}
								label={tr(c.shareUserEmail)}
								hint={tr(c.shareHint)}
								inputId="share-email-input"
								disabled={isOffline}
							/>

							{#if cfg.sharePermissions}
								<SharePermissions
									bind:canEdit
									bind:canDelete
									bind:canEditTransactions
									showEditTransactions={cfg.showEditTransactions}
									labelEdit={tr(c.canEdit)}
									labelEditDesc={tr(c.canEditDesc)}
									labelDelete={tr(c.canDelete)}
									labelDeleteDesc={tr(c.canDeleteDesc)}
									labelEditTransactions={cfg.showEditTransactions
										? tr(c.canManageTransactions)
										: undefined}
									labelEditTransactionsDesc={cfg.showEditTransactions
										? tr(c.canManageTransactionsDesc)
										: undefined}
								/>
							{/if}

							<div class="bg-white border border-accent-200 rounded-lg p-3">
								<h4 class="font-medium text-accent-900 text-sm mb-2">
									{tr(c.whatIsShared)}
								</h4>
								{#if kind === 'voucher'}
									<ul class="text-xs text-accent-800 space-y-1">
										<li>{tr(c.sharedCode)}</li>
										<li>{tr(c.sharedDetails)}</li>
										<li>{tr(c.sharedDescription)}</li>
									</ul>
									<p class="text-xs text-accent-hover mt-2 italic">
										{tr(c.readOnlyNote)}
									</p>
								{:else if kind === 'gift_card'}
									<ul class="text-xs text-accent-800 space-y-1">
										<li>{tr('giftCards.sharing.sharedItemCardNumber')}</li>
										<li>{tr('giftCards.sharing.sharedItemBalance')}</li>
										<li>{tr('giftCards.sharing.sharedItemDetails')}</li>
										<li>{tr('giftCards.sharing.sharedItemTransactions')}</li>
										<li>{tr('giftCards.sharing.sharedItemNotes')}</li>
									</ul>
								{:else}
									<ul class="text-xs text-accent-800 space-y-1">
										<li>{tr('cards.sharing.sharedItemCardNumber')}</li>
										<li>{tr('cards.sharing.sharedItemDetails')}</li>
										<li>{tr('cards.sharing.sharedItemNotes')}</li>
									</ul>
								{/if}
							</div>

							<div class="flex gap-2">
								<button
									onclick={handleShare}
									disabled={isOffline}
									class="btn btn-primary flex-1 {isOffline
										? 'opacity-50 cursor-not-allowed'
										: ''}"
								>
									{tr(c.shareNow)}
								</button>
								<button onclick={resetShareForm} class="btn btn-ghost">
									{tr('common.cancel')}
								</button>
							</div>
						</div>
					{/if}

					{#if shares.length > 0}
						<div class="space-y-3">
							{#each shares as share (share.shared_with_user.id)}
								{#if shareMode === 'readonly'}
									<ShareListItem
										{share}
										isEditing={managingShareId === share.shared_with_user.id}
										{isOffline}
										editButtonLabel={tr(c.manage)}
										deleteButtonLabel={tr(c.removeShare)}
										alwaysViewOnly={true}
										onstartEdit={() =>
											startManageShare(share.shared_with_user.id)}
										oncancel={cancelManageShare}
										ondelete={() =>
											promptDeleteShare(share.shared_with_user.id)}
									>
										<div
											class="bg-warning-50 border border-warning-200 rounded-lg p-3"
										>
											<p class="text-xs font-medium text-warning-800 mb-1">
												{tr(c.alwaysReadOnly)}
											</p>
											<p class="text-xs text-warning-700">
												{tr(c.canOnlyRemove)}
											</p>
										</div>
									</ShareListItem>
								{:else}
									<ShareListItem
										{share}
										isEditing={editingShareId === share.shared_with_user.id}
										{isOffline}
										showTransactionsBadge={cfg.showEditTransactions}
										onstartEdit={() => startEditShare(share)}
										onsave={() => saveShareEdit(share.shared_with_user.id)}
										oncancel={cancelEditShare}
										ondelete={() =>
											promptDeleteShare(share.shared_with_user.id)}
									>
										<SharePermissions
											bind:canEdit={editShareCanEdit}
											bind:canDelete={editShareCanDelete}
											bind:canEditTransactions={editShareCanEditTransactions}
											showEditTransactions={cfg.showEditTransactions}
											labelEdit={tr(c.canEdit)}
											labelEditDesc={tr(c.canEditDesc)}
											labelDelete={tr(c.canDelete)}
											labelDeleteDesc={tr(c.canDeleteDesc)}
											labelEditTransactions={cfg.showEditTransactions
												? tr(c.canManageTransactions)
												: undefined}
											labelEditTransactionsDesc={cfg.showEditTransactions
												? tr(c.canManageTransactionsDesc)
												: undefined}
										/>
									</ShareListItem>
								{/if}
							{/each}
						</div>
						<button
							type="button"
							onclick={promptRevokeAll}
							disabled={isOffline}
							class="btn btn-ghost text-danger-600 mt-3 w-full disabled:opacity-50"
						>
							{tr(c.revokeAll)}
						</button>
					{:else}
						<p class="text-sm text-text-subtle text-center py-4">
							{tr(c.notSharedYet)}
						</p>
					{/if}
				</div>
			{/if}
		</div>
	</div>

	<!-- Confirmation Modals -->
	<ConfirmModal
		isOpen={showDeleteModal}
		title={tr(c.deleteConfirm)}
		message={tr(c.deleteConfirmMessage)}
		confirmText={tr('common.delete')}
		cancelText={tr('common.cancel')}
		variant="danger"
		onconfirm={confirmDelete}
		oncancel={() => (showDeleteModal = false)}
	/>

	<ConfirmModal
		isOpen={showDeleteShareModal}
		title={tr(c.removeConfirm)}
		message={tr(c.removeConfirmMessage)}
		confirmText={tr('common.remove')}
		cancelText={tr('common.cancel')}
		variant="danger"
		onconfirm={confirmDeleteShare}
		oncancel={() => (showDeleteShareModal = false)}
	/>

	<ConfirmModal
		isOpen={showRevokeAllModal}
		title={tr(c.revokeAllConfirm)}
		message={tr(c.revokeAllConfirmMessage)}
		confirmText={tr(c.revokeAll)}
		cancelText={tr('common.cancel')}
		variant="danger"
		onconfirm={confirmRevokeAll}
		oncancel={() => (showRevokeAllModal = false)}
	/>

	<ConfirmModal
		isOpen={showTransferModal}
		title={tr(c.transferConfirmTitle)}
		message={tr(c.transferConfirmMessage)}
		confirmText={tr(c.transferTransferButton)}
		cancelText={tr('common.cancel')}
		variant="transfer"
		onconfirm={confirmTransfer}
		oncancel={() => (showTransferModal = false)}
	/>
{:else}
	<!-- Not loading and no resource (e.g. after a transfer that removed access,
	     or a load that failed before redirect). Without this branch the page
	     rendered nothing → white screen (was gift-card-only, issue #121). -->
	<div class="flex flex-col items-center justify-center py-16 text-center">
		<p class="text-text-muted mb-4">{tr(notFoundKeys.title)}</p>
		<button type="button" onclick={goToList} class="btn btn-primary">
			{tr(notFoundKeys.back)}
		</button>
	</div>
{/if}
