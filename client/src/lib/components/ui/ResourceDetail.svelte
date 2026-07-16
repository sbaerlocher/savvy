<script lang="ts">
	import { ICON_LOCK } from '$lib/icons';
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

				<!-- Sharing (own state + API calls → dedicated component). -->
				<ShareSection
					{kind}
					resource={resource!}
					bind:shares
					{isOffline}
					{shareMode}
				/>
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
