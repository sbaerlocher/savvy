<script lang="ts">
	import { ICON_LOCK, ICON_SHARE, ICON_TRANSFER, ICON_TRASH } from '$lib/icons';
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
	import { portal } from '$lib/actions/portal';
	import BottomSheet from '$lib/components/BottomSheet.svelte';
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
		/** Bindable, so the page can render the desktop title row (edit-aware
		 *  title, back link, action buttons) on the shell while this component
		 *  keeps owning the behaviour. */
		isEditing?: boolean;
		isTogglingFavorite?: boolean;
	}

	let {
		kind,
		resource = $bindable(),
		shares = $bindable(),
		isOffline,
		shareMode = 'editable',
		onStartEdit,
		edit,
		ledger,
		isEditing = $bindable(false),
		isTogglingFavorite = $bindable(false)
	}: Props = $props();

	// Android detail chrome (screen-ResourceDetailAndroid): the M3 top app bar
	// replaces the PageHeader, sharing and transfer move from permanent cards
	// into bottom sheets opened from a button row under the resource card, and
	// the overflow menu holds delete.
	const isAndroid = platform === 'android';
	// Desktop renders its own detail chrome (screen-ResourceDetailDesktop): a
	// header row carrying the type icon, eyebrow, title and the edit/delete text
	// buttons, and a two-column body.
	const IS_DESKTOP = platform === 'other';
	// iOS puts the same secondary actions behind its own chrome (mockup
	// screen-ResourceDetailIOS): share and transfer are sheets too, the •••
	// glyph opens a context menu, delete confirms through an action sheet.
	const IS_IOS = platform === 'ios';
	let showOverflowSheet = $state(false);
	let showShareSheet = $state(false);
	let showTransferSheet = $state(false);
	let showMoreMenu = $state(false);
	let moreMenuEl = $state<HTMLDivElement>();

	function onMoreMenuKeydown(event: KeyboardEvent) {
		if (!showMoreMenu || event.key !== 'Escape') return;
		event.preventDefault();
		showMoreMenu = false;
	}

	// Move focus into the menu when it opens, so Escape reaches the handler and
	// the tab order does not walk the page behind the scrim.
	$effect(() => {
		if (showMoreMenu) moreMenuEl?.focus();
	});

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

	// Body eyebrow (Android view mode): card → program; voucher →
	// single/multi-use label; gift cards get none here.
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
	const accentColor = $derived(resource?.merchant?.color || cfg.accentFallback);

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

	// iOS detail cards are solid grouped-inset cards (mockup: bg-surface,
	// --radius-inset, 3px accent edge, no surrounding border and no glass).
	// Gift cards carry the gold giftcard edge instead of the merchant accent.
	const viewCardClass = $derived(
		isAndroid
			? 'rounded-m3-lg bg-m3-card overflow-hidden'
			: IS_IOS
				? 'relative overflow-hidden rounded-[var(--radius-inset)] bg-surface'
				: 'overflow-hidden rounded-xl border border-border/80 bg-white'
	);
	const viewCardStyle = $derived(
		IS_IOS && kind === 'gift_card'
			? 'border-left: 3px solid var(--color-giftcard-edge)'
			: `border-left: 3px solid color-mix(in srgb, ${accentColor} 70%, transparent)`
	);
	const editCardClass = $derived(
		isAndroid
			? 'rounded-m3-lg bg-m3-card overflow-hidden'
			: IS_DESKTOP
				? 'overflow-hidden rounded-2xl border border-border bg-white p-7'
				: IS_IOS
					? 'overflow-hidden rounded-[var(--radius-inset)] bg-surface'
					: 'overflow-hidden rounded-xl border border-border bg-white'
	);

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

	// Transfer targets the same per-kind resource API as sharing.
	const shareApi = $derived(
		kind === 'card' ? cardsApi : kind === 'voucher' ? vouchersApi : giftCardsApi
	);

	// Back: return to where the user came from; fall back to wallet on deep link.
	export function goBack() {
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

	export async function startEdit() {
		// Route populates its bound edit fields (merchants load + field copy),
		// then we flip into edit mode which renders the `edit` snippet.
		await onStartEdit();
		isEditing = true;
	}

	export function cancelEdit() {
		isEditing = false;
	}

	export async function toggleFavorite() {
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

	// Opened from the Android overflow sheet and the iOS ••• menu.
	export function openShareSheet() {
		showMoreMenu = false;
		showShareSheet = true;
	}

	// ••• on the title row: Android opens the M3 overflow bottom sheet (share /
	// transfer / delete), iOS its glass context menu (transfer / delete).
	export function openMoreMenu() {
		if (isAndroid) showOverflowSheet = true;
		else showMoreMenu = true;
	}

	export function promptDelete() {
		showMoreMenu = false;
		showDeleteModal = true;
	}

	// iOS routes transfer through its own sheet; the other platforms keep the
	// inline TransferBox in the right-hand column.
	export function openTransferSheet() {
		showMoreMenu = false;
		showTransferSheet = true;
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
		// The sheet stays open: the confirm modal takes the elevated layer above
		// it, so cancelling returns the user to the filled form instead of a bare
		// detail page.
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

	// Section title over the desktop edit form ("Kartendaten" / …). Lives in the
	// per-kind i18n namespace, not in CONFIG's key map.
	const dataSectionKey = $derived(
		kind === 'card'
			? 'cards.dataSection'
			: kind === 'voucher'
				? 'vouchers.dataSection'
				: 'giftCards.dataSection'
	);

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

<svelte:window onkeydown={onMoreMenuKeydown} />

{#if resource}
	<!-- The title row (back link, title, actions) lives on the page's
	     PageShell for every platform; this block starts with the shared-by
	     line. -->
	{#if !isEditing && resource.owner && resource.owner.id !== $authStore.user?.id}
		<p class="mb-6 text-body-sm text-text-faint">
			{tr(c.sharedBy, {
				name: resource.owner.first_name || resource.owner.email
			})}
		</p>
	{/if}

	{#if !isEditing}
		<!-- Android M3: edit is a bottom-right FAB. It opens the same edit mode by
	     toggling the form; the route renders the form via the edit snippet, but
	     the trigger lives here. Uses startEdit which the route overrides through
	     the title-row actions — the FAB simply mirrors it. The bottom nav stays
	     visible on detail routes and drops its own New FAB there (MobileNav),
	     so this one takes the nav-FAB slot above the bar. -->
		{#if resource.permissions?.can_edit && isAndroid}
			<button
				type="button"
				onclick={startEdit}
				disabled={isOffline}
				aria-label={tr('common.edit')}
				class="mobile-nav-fab-android bg-accent text-on-accent fixed right-4.5 z-50 flex h-14 w-14 items-center justify-center rounded-m3-lg shadow-[var(--shadow-fab)] disabled:pointer-events-none disabled:opacity-50"
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
		     kicker sits above the card (mockup). The "shared by" line renders in
		     the shared block above, for every platform. -->
		{#if eyebrow}
			<p
				class="text-eyebrow text-text-faint mx-0.5 mb-3.5 {eyebrowVerbatim
					? ''
					: 'uppercase'}"
			>
				{eyebrow}
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
				<div class={viewCardClass} style={viewCardStyle}>
					<div
						class="{isAndroid || IS_IOS ? 'p-5' : 'p-6'} {isDimmed
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
			{:else}
				<!-- Edit Mode: per-kind form slot + shared delete button. On desktop the
				     form card carries a section title and the delete button joins the
				     form's action row (mockup), so it is rendered by the form there;
				     iOS puts it in its own grouped-inset card below. -->
				<div class={editCardClass}>
					<div
						class={isAndroid
							? 'm3-filled-form p-4'
							: IS_DESKTOP
								? ''
								: IS_IOS
									? 'p-[18px]'
									: 'p-6'}
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
						{#if resource.permissions?.can_delete && !isAndroid && !IS_DESKTOP && !IS_IOS}
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

				<!-- iOS: delete becomes its own grouped-inset card (mockup). -->
				{#if IS_IOS && resource.permissions?.can_delete}
					<div
						class="mt-3 overflow-hidden rounded-[var(--radius-inset)] bg-surface"
					>
						<button
							type="button"
							onclick={promptDelete}
							disabled={isOffline}
							class="flex min-h-11 w-full items-center justify-center px-4 text-[17px] text-danger-600 disabled:opacity-40"
						>
							{tr(c.deleteButton)}
						</button>
					</div>
				{/if}

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

				{#if resource.permissions?.is_owner && !isAndroid && !IS_IOS}
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
			shareLabel={tr(c.shareTitle)}
			deleteLabel={tr(c.deleteButton)}
			onClose={() => (showOverflowSheet = false)}
			onshare={resource.permissions?.is_owner ? openShareSheet : undefined}
			ontransfer={resource.permissions?.is_owner
				? openTransferSheet
				: undefined}
			ondelete={resource.permissions?.can_delete ? promptDelete : undefined}
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

	{#if IS_IOS && resource.permissions?.is_owner}
		<!-- ••• context menu: transfer + delete behind a glass popover. The
		     overlay is portalled to <body> so an ancestor backdrop-filter (the
		     glass chrome) cannot trap the position:fixed layer. -->
		{#if showMoreMenu}
			<!-- Escape closes and focus moves into the menu, matching Modal and
			     BottomSheet: without it a keyboard or VoiceOver user is left
			     tabbing the page behind the scrim. -->
			<div use:portal class="fixed inset-0 z-[70]">
				<div
					class="absolute inset-0 bg-[var(--color-glass-scrim)] backdrop-blur-[3px]"
					onclick={() => (showMoreMenu = false)}
					role="presentation"
				></div>
				<div
					bind:this={moreMenuEl}
					class="liquid-glass-menu absolute right-3.5 top-[72px] w-[252px] overflow-hidden rounded-[var(--radius-xl)] focus:outline-none"
					role="menu"
					tabindex="-1"
					aria-label={tr('common.moreActions')}
				>
					<button
						type="button"
						role="menuitem"
						onclick={openShareSheet}
						disabled={isOffline}
						class="flex min-h-[46px] w-full items-center justify-between gap-3 px-4 text-[17px] text-text disabled:opacity-40"
					>
						<span>{tr('common.share')}</span>
						<svg
							class="h-[19px] w-[19px] flex-none"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							stroke-width="1.9"
							stroke-linecap="round"
							stroke-linejoin="round"
						>
							<path d={ICON_SHARE} />
						</svg>
					</button>
					<div class="h-px bg-[var(--color-glass-edge)]"></div>
					<button
						type="button"
						role="menuitem"
						onclick={openTransferSheet}
						disabled={isOffline}
						class="flex min-h-[46px] w-full items-center justify-between gap-3 px-4 text-[17px] text-text disabled:opacity-40"
					>
						<span>{tr('common.transferOwnership')}</span>
						<svg
							class="h-[19px] w-[19px] flex-none"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							stroke-width="1.9"
							stroke-linecap="round"
							stroke-linejoin="round"
						>
							<path d={ICON_TRANSFER} />
						</svg>
					</button>
					<div class="h-px bg-[var(--color-glass-edge)]"></div>
					<button
						type="button"
						role="menuitem"
						onclick={promptDelete}
						disabled={isOffline}
						class="flex min-h-[46px] w-full items-center justify-between gap-3 px-4 text-[17px] text-danger-600 disabled:opacity-40"
					>
						<span>{tr('common.delete')}</span>
						<svg
							class="h-[18px] w-[18px] flex-none"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							stroke-width="1.9"
							stroke-linecap="round"
							stroke-linejoin="round"
						>
							<path d={ICON_TRASH} />
						</svg>
					</button>
				</div>
			</div>
		{/if}

		<!-- Share sheet (iOS): the same ShareSection, presented as a sheet. -->
		<BottomSheet
			open={showShareSheet}
			onClose={() => (showShareSheet = false)}
			maxHeight="88%"
			ariaLabel={tr(c.shareTitle)}
		>
			<!-- Sheet header: cancel · title · confirm (mockup). The confirm side
			     is handled by ShareSection's own add flow, so the right slot
			     stays empty rather than duplicating a second submit. -->
			<div
				class="flex items-center justify-between border-b border-border-soft px-[18px] pb-3 pt-1.5"
			>
				<button
					type="button"
					onclick={() => (showShareSheet = false)}
					class="flex-1 text-left text-[17px] text-accent"
				>
					{tr('common.cancel')}
				</button>
				<span class="text-[17px] font-semibold text-text"
					>{tr(c.shareTitle)}</span
				>
				<span class="flex-1"></span>
			</div>
			<div class="px-[18px] pb-6 pt-4">
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

		<!-- Transfer sheet (iOS). -->
		<BottomSheet
			open={showTransferSheet}
			onClose={() => (showTransferSheet = false)}
			maxHeight="88%"
			ariaLabel={tr(c.transferTitle)}
		>
			<!-- Transfer sheet header in the transfer palette (mockup). -->
			<div
				class="flex items-center justify-between border-b border-border-soft px-[18px] pb-3 pt-1.5"
			>
				<button
					type="button"
					onclick={() => (showTransferSheet = false)}
					class="flex-1 text-left text-[17px] text-accent"
				>
					{tr('common.cancel')}
				</button>
				<span class="text-[17px] font-semibold text-transfer-900"
					>{tr(c.transferTitle)}</span
				>
				<span class="flex-1"></span>
			</div>
			<div class="px-[18px] pb-6 pt-4">
				<TransferBox
					variant="sheet"
					{isOffline}
					title={tr(c.transferTitle)}
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
