<script lang="ts">
	import { ICON_LOCK, ICON_TRASH } from '$lib/icons';
	import type { Snippet } from 'svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { get } from 'svelte/store';
	import { authStore } from '$lib/stores/auth';
	import { t, locale } from '$lib/stores/i18n';
	import { toastStore } from '$lib/stores/toast';
	import { formatCurrency } from '$lib/utils/currency';
	import { platform } from '$lib/utils/platform';
	import { cardsApi, vouchersApi, giftCardsApi } from '$lib/api';
	import { resourceDetailPath } from '$lib/resource/routes';
	import { CONFIG, type Kind } from '$lib/resource/config';
	import DuplicateWarningBanner from '$lib/components/DuplicateWarningBanner.svelte';
	import BarcodeDisplay from '$lib/components/BarcodeDisplay.svelte';
	import ConfirmModal from '$lib/components/ConfirmModal.svelte';
	import TransferBox from '$lib/components/TransferBox.svelte';
	import ShareSection from '$lib/components/resource/ShareSection.svelte';
	import ResourceActions from '$lib/components/ui/ResourceActions.svelte';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import BottomSheet from '$lib/components/BottomSheet.svelte';
	import M3DetailAppBar from '$lib/components/ui/M3DetailAppBar.svelte';
	import M3DetailBarActions from '$lib/components/ui/M3DetailBarActions.svelte';
	import M3DetailOverflowSheet from '$lib/components/ui/M3DetailOverflowSheet.svelte';
	import type {
		CardDTO,
		VoucherDTO,
		GiftCardDTO,
		ShareDTO
	} from '$lib/types/api';

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
		edit: Snippet<
			[
				{
					cancel: () => void;
					close: () => void;
					/** Desktop: pass to the form's `trailingActions` so delete joins its
					 *  action row (mockup). Undefined when the user cannot delete or on
					 *  the native layouts, where delete stays below the form. */
					deleteAction?: Snippet;
				}
			]
		>;
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

	// Android detail chrome (screen-ResourceDetailAndroid): the M3 top app bar
	// replaces the PageHeader, sharing and transfer move from permanent cards
	// into bottom sheets opened from a button row under the resource card, and
	// the overflow menu holds delete.
	const isAndroid = platform === 'android';
	// Desktop renders its own detail chrome (screen-ResourceDetailDesktop): a
	// header row carrying the type icon, eyebrow, title and the edit/delete text
	// buttons, and a two-column body.
	const IS_DESKTOP = platform === 'other';
	let showOverflowSheet = $state(false);
	let showShareSheet = $state(false);
	let showTransferSheet = $state(false);

	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);

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
	// The card program is whatever the user typed, so it must not get the
	// native uppercase kicker treatment; the voucher label is translated copy.
	const eyebrowVerbatim = $derived(!!asCard);

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

	let transferEmail = $state('');

	// Modal state
	let showDeleteModal = $state(false);
	let showTransferModal = $state(false);

	let isTogglingFavorite = $state(false);

	// Transfer targets the same per-kind resource API as sharing.
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

	// Desktop header (mockup): the type icon sits in a 44px tinted square left of
	// the eyebrow + title pair. Same neutral line icons the wallet tiles use.
	const TYPE_ICON: Record<Kind, string> = {
		card: 'M3 10h18M7 15h1m4 0h1m-7 4h12a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z',
		voucher:
			'M15 5v2m0 4v2m0 4v2M5 5a2 2 0 00-2 2v3a2 2 0 010 4v3a2 2 0 002 2h14a2 2 0 002-2v-3a2 2 0 010-4V7a2 2 0 00-2-2H5z',
		gift_card:
			'M12 8v13m0-13V6a2 2 0 112-2 2 2 0 01-2 2zm0 0V5.5A2.5 2.5 0 109.5 8H12zm-7 4h14M5 12a2 2 0 110-4h14a2 2 0 110 4M5 12v7a2 2 0 002 2h10a2 2 0 002-2v-7'
	};

	// Back link copy: view mode returns to the list, edit mode to the resource.
	const backToResourceKey = $derived(
		kind === 'card'
			? 'common.backToCard'
			: kind === 'voucher'
				? 'common.backToVoucher'
				: 'common.backToGiftCard'
	);

	// Section title over the desktop edit form ("Kartendaten" / …). Lives in the
	// per-kind i18n namespace, not in CONFIG's key map.
	const dataSectionKey = $derived(
		kind === 'card'
			? 'cards.dataSection'
			: kind === 'voucher'
				? 'vouchers.dataSection'
				: 'giftCards.dataSection'
	);

	const canDelete = $derived(!!resource?.permissions?.can_delete);
	const canEditResource = $derived(!!resource?.permissions?.can_edit);
	const isOwner = $derived(!!resource?.permissions?.is_owner);
	// Desktop only: card and voucher lose the right column entirely when viewed as
	// a recipient, and the gift card keeps just its read-only ledger there (mockup
	// boards C and D). The native layouts keep the column whenever there is a
	// ledger, edit mode included.
	const hasRightColumn = $derived(
		IS_DESKTOP ? isOwner || (!!ledger && !isEditing) : isOwner || !!ledger
	);
	// The desktop edit board drops the ledger; panel order keys on this, not on
	// the snippet merely being passed in.
	const showLedger = $derived(!!ledger && !(IS_DESKTOP && isEditing));
</script>

{#snippet deleteAction()}
	<button
		type="button"
		onclick={promptDelete}
		disabled={isOffline}
		class="inline-flex h-11 items-center gap-2 rounded-lg border border-danger-200 bg-danger-50 px-4 text-label text-danger-700 transition-colors hover:bg-danger-100 disabled:cursor-not-allowed disabled:opacity-50"
	>
		<svg
			class="h-4 w-4"
			fill="none"
			stroke="currentColor"
			stroke-width="2"
			viewBox="0 0 24 24"
		>
			<path
				stroke-linecap="round"
				stroke-linejoin="round"
				d={isOffline ? LOCK_PATH : ICON_TRASH}
			/>
		</svg>
		{tr('common.delete')}
	</button>
{/snippet}

{#if resource}
	<!-- Android: M3 small top app bar in both modes — view mode carries the
	     resource title with star + overflow, edit mode swaps to a close cross
	     and the "<kind> bearbeiten" title (mockup frames 4-6). -->
	{#if isAndroid}
		<M3DetailAppBar
			title={isEditing ? tr(c.editTitle) : pageTitle}
			nav={isEditing ? 'close' : 'back'}
			onNav={isEditing ? cancelEdit : goBack}
		>
			{#snippet actions()}
				{#if !isEditing}
					<M3DetailBarActions
						{isOffline}
						isFavorite={resource!.is_favorite}
						{isTogglingFavorite}
						showOverflow={!!resource!.permissions?.can_delete}
						favoriteTitleAdd={tr(cfg.favoriteAdd)}
						favoriteTitleRemove={tr(cfg.favoriteRemove)}
						ontoggleFavorite={toggleFavorite}
						onoverflow={() => (showOverflowSheet = true)}
					/>
				{/if}
			{/snippet}
		</M3DetailAppBar>
	{/if}

	{#if IS_DESKTOP}
		<!-- Desktop header (mockup): back link, type icon + eyebrow/title on the
		     left, favourite/edit/delete as text buttons on the right. Edit mode
		     swaps in its own title and drops the action buttons. -->
		<div class="mb-6 flex items-start justify-between gap-5">
			<div class="min-w-0">
				<button
					type="button"
					onclick={isEditing ? cancelEdit : goBack}
					class="mb-3 inline-flex items-center gap-1.5 text-label text-accent hover:text-accent-hover"
				>
					<svg
						class="h-4 w-4"
						fill="none"
						stroke="currentColor"
						stroke-width="2.2"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							d="M15 18l-6-6 6-6"
						/>
					</svg>
					{isEditing ? tr(backToResourceKey) : tr('common.backToOverviewPlain')}
				</button>
				{#if isEditing}
					<h1 class="text-title text-text">{tr(c.editTitle)}</h1>
				{:else}
					<div class="flex items-center gap-3">
						<span
							class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl"
							style="background: color-mix(in srgb, {accentColor} 16%, transparent); color: {accentColor}"
						>
							<svg
								class="h-5.5 w-5.5"
								fill="none"
								stroke="currentColor"
								stroke-width="1.9"
								viewBox="0 0 24 24"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									d={TYPE_ICON[kind]}
								/>
							</svg>
						</span>
						<div class="min-w-0">
							<p
								class="text-section-eyebrow text-text-faint {eyebrowVerbatim
									? ''
									: 'uppercase'}"
							>
								{eyebrow
									? `${tr(`common.${kind}`)} · ${eyebrow}`
									: tr(`common.${kind}`)}
							</p>
							<h1 class="text-title text-text">{pageTitle}</h1>
						</div>
					</div>
				{/if}
			</div>
			{#if !isEditing}
				<div class="flex shrink-0 items-center gap-2.5">
					<button
						type="button"
						data-testid="favorite-button"
						onclick={toggleFavorite}
						disabled={isOffline || isTogglingFavorite}
						aria-pressed={resource.is_favorite}
						title={resource.is_favorite
							? tr(cfg.favoriteRemove)
							: tr(cfg.favoriteAdd)}
						class="inline-flex h-10 w-10 items-center justify-center rounded-lg border border-border-field bg-white text-accent transition-colors hover:bg-surface-1 disabled:cursor-not-allowed disabled:opacity-50"
					>
						<svg
							class="h-5 w-5"
							viewBox="0 0 24 24"
							fill={resource.is_favorite ? 'currentColor' : 'none'}
							stroke="currentColor"
							stroke-width="1.9"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								d="M12 3l2.7 5.6 6.1.9-4.4 4.3 1 6.1-5.4-2.9-5.4 2.9 1-6.1-4.4-4.3 6.1-.9z"
							/>
						</svg>
					</button>
					{#if canEditResource}
						<button
							type="button"
							onclick={startEdit}
							disabled={isOffline}
							class="inline-flex h-10 items-center gap-2 rounded-lg border border-border-field bg-white px-4 text-label text-text-ink2 transition-colors hover:bg-surface-1 disabled:cursor-not-allowed disabled:opacity-50"
						>
							<svg
								class="h-4 w-4"
								fill="none"
								stroke="currentColor"
								stroke-width="2"
								viewBox="0 0 24 24"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									d={isOffline
										? LOCK_PATH
										: 'M12 20h9M16.5 3.5a2.1 2.1 0 013 3L7 19l-4 1 1-4z'}
								/>
							</svg>
							{tr('common.edit')}
						</button>
					{/if}
					{#if canDelete}
						<button
							type="button"
							onclick={promptDelete}
							disabled={isOffline}
							class="inline-flex h-10 items-center gap-2 rounded-lg border border-danger-200 bg-danger-50 px-4 text-label text-danger-700 transition-colors hover:bg-danger-100 disabled:cursor-not-allowed disabled:opacity-50"
						>
							<svg
								class="h-4 w-4"
								fill="none"
								stroke="currentColor"
								stroke-width="2"
								viewBox="0 0 24 24"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									d={isOffline ? LOCK_PATH : ICON_TRASH}
								/>
							</svg>
							{tr('common.delete')}
						</button>
					{/if}
				</div>
			{/if}
		</div>
		{#if !isEditing && resource.owner && resource.owner.id !== $authStore.user?.id}
			<p class="-mt-4 mb-6 text-body-sm text-text-faint">
				{tr(c.sharedBy, {
					name: resource.owner.first_name || resource.owner.email
				})}
			</p>
		{/if}
	{/if}

	<!-- Page header (view mode only; the edit form keeps its own title). -->
	{#if !isEditing && !isAndroid && !IS_DESKTOP}
		<PageHeader
			title={pageTitle}
			{eyebrow}
			{eyebrowVerbatim}
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
	{/if}

	{#if !isEditing}
		<!-- Android M3: edit is a bottom-right FAB. It opens the same edit mode by
	     toggling the form; the route renders the form via the edit snippet, but
	     the trigger lives here. Uses startEdit which the route overrides through
	     ResourceActions above — the FAB simply mirrors it. The detail route has
	     no bottom nav (mockup), so the FAB sits on the screen edge. -->
		{#if resource.permissions?.can_edit && isAndroid}
			<button
				type="button"
				onclick={startEdit}
				disabled={isOffline}
				aria-label={tr('common.edit')}
				style="bottom: calc(1.375rem + env(safe-area-inset-bottom))"
				class="bg-accent text-on-accent fixed right-4.5 z-50 flex h-14 w-14 items-center justify-center rounded-m3-lg shadow-[var(--shadow-fab)] disabled:pointer-events-none disabled:opacity-50"
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

	{#if isAndroid && !isEditing}
		<!-- Body eyebrow: the M3 top app bar carries only the title, so the kind
		     kicker and the "geteilt von" line sit above the card (mockup). -->
		{#if eyebrow}
			<p
				class="text-eyebrow text-text-faint mx-0.5 mb-3.5 {eyebrowVerbatim
					? ''
					: 'uppercase'}"
			>
				{eyebrow}
			</p>
		{/if}
		{#if resource.owner && resource.owner.id !== $authStore.user?.id}
			<p class="text-text-faint mx-0.5 -mt-2 mb-3.5 text-xs">
				{tr(c.sharedBy, {
					name: resource.owner.first_name || resource.owner.email
				})}
			</p>
		{/if}
	{/if}

	<!-- Body: two columns on desktop (2fr / 1fr per mockup). A recipient without
	     a right column (card / voucher shared with me) gets the single narrower
	     column instead. -->
	<div
		class={IS_DESKTOP
			? hasRightColumn
				? 'grid grid-cols-1 items-start gap-6 lg:grid-cols-[2fr_1fr]'
				: 'max-w-3xl'
			: 'grid grid-cols-1 lg:grid-cols-3 gap-6'}
	>
		<!-- Left column: resource details / edit form -->
		<div class={IS_DESKTOP ? 'min-w-0' : 'lg:col-span-2'}>
			{#if !isEditing}
				<!-- View Mode -->
				<div
					class={isAndroid
						? 'rounded-m3-lg bg-m3-card overflow-hidden'
						: 'overflow-hidden rounded-xl border border-border/80 bg-white'}
					style="border-left: 3px solid color-mix(in srgb, {accentColor} 70%, transparent)"
				>
					<div
						class="{isAndroid ? 'p-5' : 'p-6'} {isDimmed
							? 'opacity-50 grayscale'
							: ''}"
					>
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
								class="mt-3.5 bg-warning-50 border-l-4 border-warning-400 px-3.5 py-2.5 {isAndroid
									? 'rounded-m3-xs'
									: 'rounded'}"
							>
								<p class="text-body-sm text-text-ink2">{notes}</p>
							</div>
						{/if}
					</div>
				</div>

				<!-- Android stacks the gift-card ledger directly under the barcode
				     card, above the action row (mockup frame 3) — the grid's right
				     column would otherwise push it below the buttons. -->
				{#if isAndroid && ledger}
					<div class="mt-3">
						{@render ledger()}
					</div>
				{/if}

				<!-- Android: share / transfer sit as a button row under the card and
				     open M3 bottom sheets (mockup frames 1-3, 7, 8). A recipient
				     (is_owner false) gets no row at all — frame 9. -->
				{#if isAndroid && resource.permissions?.is_owner}
					<div class="mt-3 flex gap-2.5">
						<button
							type="button"
							onclick={() => (showShareSheet = true)}
							disabled={isOffline}
							class="bg-accent-600 text-on-accent text-label flex h-12 flex-1 items-center justify-center gap-2 rounded-m3-full disabled:opacity-50"
						>
							<svg
								class="h-4.5 w-4.5"
								fill="none"
								stroke="currentColor"
								stroke-width="2"
								stroke-linecap="round"
								stroke-linejoin="round"
								viewBox="0 0 24 24"
							>
								<circle cx="18" cy="5" r="3" />
								<circle cx="6" cy="12" r="3" />
								<circle cx="18" cy="19" r="3" />
								<path d="M8.6 13.5l6.8 4M15.4 6.5l-6.8 4" />
							</svg>
							{tr(c.shareTitle)}
						</button>
						<button
							type="button"
							onclick={() => (showTransferSheet = true)}
							disabled={isOffline}
							class="bg-m3-card border-transfer-200 text-transfer-700 text-label flex h-12 flex-1 items-center justify-center gap-2 rounded-m3-full border disabled:opacity-50"
						>
							<svg
								class="h-4.5 w-4.5"
								fill="none"
								stroke="currentColor"
								stroke-width="2"
								stroke-linecap="round"
								stroke-linejoin="round"
								viewBox="0 0 24 24"
							>
								<path d="M7 4L3 8l4 4" />
								<path d="M3 8h13" />
								<path d="M17 20l4-4-4-4" />
								<path d="M21 16H8" />
							</svg>
							{tr('common.transfer')}
						</button>
					</div>
				{/if}
			{:else}
				<!-- Edit Mode: per-kind form slot + shared delete button. On desktop the
				     form card carries a section title and the delete button joins the
				     form's action row (mockup), so it is rendered by the form there. -->
				<div
					class={isAndroid
						? 'rounded-m3-lg bg-m3-card overflow-hidden'
						: IS_DESKTOP
							? 'overflow-hidden rounded-2xl border border-border bg-white p-7'
							: 'overflow-hidden rounded-xl border border-border bg-white'}
				>
					<div
						class={isAndroid ? 'm3-filled-form p-4' : IS_DESKTOP ? '' : 'p-6'}
					>
						{#if IS_DESKTOP}
							<h2 class="mb-5 text-subheading font-bold text-text">
								{tr(dataSectionKey)}
							</h2>
						{/if}
						{@render edit({
							cancel: cancelEdit,
							close: cancelEdit,
							deleteAction:
								IS_DESKTOP && resource.permissions?.can_delete
									? deleteAction
									: undefined
						})}
						{#if resource.permissions?.can_delete && !isAndroid && !IS_DESKTOP}
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

				<!-- Android: delete leaves the form card — the mockup puts it below a
				     divider as a plain danger text button (frames 4-6). -->
				{#if isAndroid && resource.permissions?.can_delete}
					<div class="border-border-soft mt-5 border-t pt-3">
						<button
							type="button"
							onclick={promptDelete}
							disabled={isOffline}
							class="text-label text-danger-600 inline-flex h-10 items-center gap-2 rounded-m3-full px-3 disabled:opacity-50"
						>
							<svg
								class="h-4.25 w-4.25"
								fill="none"
								stroke="currentColor"
								stroke-width="2"
								stroke-linecap="round"
								stroke-linejoin="round"
								viewBox="0 0 24 24"
							>
								{#if isOffline}
									<path d={LOCK_PATH} />
								{:else}
									<path d="M4 7h16" />
									<path d="M9 7V5a1 1 0 011-1h4a1 1 0 011 1v2" />
									<path d="M6 7l1 13a1 1 0 001 1h8a1 1 0 001-1l1-13" />
								{/if}
							</svg>
							{tr(c.deleteButton)}
						</button>
					</div>
				{/if}
			{/if}
		</div>

		<!-- Right column: ledger (gift), transfer & sharing (owners) -->
		{#if hasRightColumn}
			<div class={IS_DESKTOP ? 'min-w-0 space-y-4' : 'lg:col-span-1 space-y-4'}>
				<!-- The desktop edit board drops the ledger; only sharing and transfer
				     stay beside the form (mockup). -->
				{#if showLedger && !isAndroid}
					{@render ledger!()}
				{/if}

				{#if resource.permissions?.is_owner && !isAndroid}
					<!-- Panel order follows the mockup: without a ledger sharing leads the
					     column, with one (gift card) transfer sits between ledger and
					     sharing. Snippets keep both boxes defined once. -->
					{#snippet transferBox()}
						<TransferBox
							{isOffline}
							title={kind === 'gift_card' ? tr(c.transferTitle) : undefined}
							subtitle={IS_DESKTOP ? tr('common.transferSubtitle') : undefined}
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
					{/snippet}
					{#snippet shareBox()}
						<ShareSection
							{kind}
							resource={resource!}
							bind:shares
							{isOffline}
							{shareMode}
						/>
					{/snippet}
					{#if showLedger}
						{@render transferBox()}
						{@render shareBox()}
					{:else}
						{@render shareBox()}
						{@render transferBox()}
					{/if}
				{/if}
			</div>
		{/if}
	</div>

	<!-- Android sheets: overflow (delete), share and transfer. The share and
	     transfer bodies are the same components the other platforms show inline,
	     rendered sheet-flavoured so the data flow stays identical. -->
	{#if isAndroid}
		<M3DetailOverflowSheet
			open={showOverflowSheet}
			{isOffline}
			deleteLabel={tr(c.deleteButton)}
			onClose={() => (showOverflowSheet = false)}
			ondelete={promptDelete}
		/>

		<BottomSheet
			open={showShareSheet}
			onClose={() => (showShareSheet = false)}
			tonalAndroid
			allowWide
			maxHeight="90%"
			ariaLabel={tr(c.shareTitle)}
		>
			<div class="px-4.5 pb-2">
				<ShareSection
					{kind}
					resource={resource!}
					bind:shares
					{isOffline}
					{shareMode}
					variant="sheet"
				/>
			</div>
		</BottomSheet>

		<BottomSheet
			open={showTransferSheet}
			onClose={() => (showTransferSheet = false)}
			tonalAndroid
			allowWide
			maxHeight="90%"
			ariaLabel={tr(c.transferTransferButton)}
		>
			<div class="px-4.5 pb-2">
				<TransferBox
					{isOffline}
					variant="sheet"
					title={tr('common.transferOwnershipTitle')}
					openButtonLabel={tr(c.transferTransferButton)}
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
			</div>
		</BottomSheet>
	{/if}

	<!-- Confirmation Modals -->
	<!-- Elevated while a sheet is open: the sheet's own z-60 backdrop would
	     otherwise sit above the modal and swallow its outside-click. Closing the
	     sheet instead would drop the state the user is about to confirm. -->
	<ConfirmModal
		isOpen={showDeleteModal}
		layer={showShareSheet || showTransferSheet ? 'elevated' : 'default'}
		title={tr(c.deleteConfirm)}
		message={tr(c.deleteConfirmMessage)}
		confirmText={tr('common.delete')}
		cancelText={tr('common.cancel')}
		variant="danger"
		onconfirm={confirmDelete}
		oncancel={() => (showDeleteModal = false)}
	/>

	<ConfirmModal
		isOpen={showTransferModal}
		layer={showTransferSheet ? 'elevated' : 'default'}
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
