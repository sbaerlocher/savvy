<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { get } from 'svelte/store';
	import { authStore } from '$lib/stores/auth';
	import { isOnline } from '$lib/stores/offline';
	import { locale, t } from '$lib/stores/i18n';
	import { vouchersApi, merchantsApi, ApiError } from '$lib/api';
	import { offlineDB } from '$lib/stores/offline-db';
	import { toastStore } from '$lib/stores/toast';
	import { formatCurrency } from '$lib/utils/currency';
	import { logger } from '$lib/utils/logger';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import BarcodeDisplay from '$lib/components/BarcodeDisplay.svelte';
	import DuplicateWarningBanner from '$lib/components/DuplicateWarningBanner.svelte';

	import ConfirmModal from '$lib/components/ConfirmModal.svelte';
	import VoucherForm from '$lib/components/vouchers/VoucherForm.svelte';
	import type { VoucherDTO, ShareDTO, MerchantDTO } from '$lib/types/api';
	import EmailAutocomplete from '$lib/components/EmailAutocomplete.svelte';
	import ShareListItem from '$lib/components/ShareListItem.svelte';
	import TransferBox from '$lib/components/TransferBox.svelte';
	import {
		formatShareResult,
		shareResponseFromError
	} from '$lib/utils/share-result';
	import ResourceActions from '$lib/components/ui/ResourceActions.svelte';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';

	// Svelte 5 compatible translation wrapper
	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);
	const pageLogger = logger.child('VoucherDetailPage');

	const voucherId = $derived($page.params.id);

	let voucher = $state<VoucherDTO | null>(null);
	let shares = $state<ShareDTO[]>([]);
	let isLoading = $state(true);
	let isRefreshing = $state(false);
	let showShareForm = $state(false);
	let shareEmails = $state<string[]>([]);
	let transferEmail = $state('');
	let isEditing = $state(false);
	let merchants = $state<MerchantDTO[]>([]);

	// Edit form fields
	let editMerchantId = $state('');
	let editCode = $state('');
	let editType = $state('percentage');
	let editValue = $state(0);
	let editCurrency = $state('CHF');
	let editMinPurchaseAmount = $state(0);
	let editBarcodeType = $state('CODE128');
	let editUsageLimitType = $state('single_use');
	let editValidFrom = $state('');
	let editValidUntil = $state('');

	// Share managing state (for vouchers, only delete is possible)
	let managingShareId = $state<string | null>(null);
	let editStatus = $state('active');
	let editDescription = $state('');

	// Modal state
	let showDeleteModal = $state(false);
	let showDeleteShareModal = $state(false);
	let showRevokeAllModal = $state(false);
	let showTransferModal = $state(false);
	let shareToDelete: string | null = null;

	const isOffline = $derived(!$isOnline);

	onMount(async () => {
		await Promise.all([loadVoucher(), loadMerchants()]);
	});

	async function loadVoucher() {
		isLoading = true;
		try {
			if (!voucherId) {
				toastStore.error(tr('vouchers.loadError'));
				goto(resolve('/vouchers'));
				return;
			}

			// Phase 1: Show cached data immediately
			const cached = await vouchersApi.getCached(voucherId);
			if (cached) {
				voucher = cached.voucher;
				shares = cached.shares;
				isLoading = false;
				isRefreshing = true;
			}

			// Phase 2: Fetch fresh data from network
			if (navigator.onLine) {
				try {
					const fresh = await vouchersApi.get(voucherId);
					voucher = fresh.voucher;
					shares = fresh.shares || [];
				} catch (err: unknown) {
					if (
						err instanceof ApiError &&
						(err.status === 403 || err.status === 404)
					) {
						await offlineDB.deleteVoucher(voucherId);
						toastStore.error(tr('vouchers.loadError'));
						goto(resolve('/vouchers'));
						return;
					}
					if (!cached) {
						toastStore.error(tr('vouchers.loadError'));
						goto(resolve('/vouchers'));
						return;
					}
					// Transient error with cached data available - show warning, don't redirect
					toastStore.warning(tr('common.offlineMode'));
				}
			} else if (!cached) {
				toastStore.error(tr('vouchers.loadError'));
				goto(resolve('/vouchers'));
			}
		} catch {
			toastStore.error(tr('vouchers.loadError'));
			goto(resolve('/vouchers'));
		} finally {
			isLoading = false;
			isRefreshing = false;
		}
	}

	async function loadMerchants() {
		// Merchants are only needed for editing, skip when offline
		if (!navigator.onLine) return;
		try {
			const response = await merchantsApi.list();
			merchants = response.merchants || [];
		} catch (err) {
			pageLogger.error('Failed to load merchants', { error: err });
		}
	}

	async function startEdit() {
		if (!voucher) return;

		if (merchants.length === 0) {
			await loadMerchants();
		}

		isEditing = true;
		editMerchantId = voucher.merchant?.id || '';
		editCode = voucher.code;
		editType = voucher.type || 'percentage';
		editValue = voucher.value;
		editCurrency = voucher.currency || 'CHF';
		editMinPurchaseAmount = voucher.min_purchase_amount || 0;
		// Ensure barcode_type always has a valid value (handle null, undefined, empty string)
		editBarcodeType =
			(voucher.barcode_type && voucher.barcode_type.trim()) || 'CODE128';
		editUsageLimitType = voucher.usage_limit_type || 'single_use';
		editValidFrom = voucher.valid_from ? voucher.valid_from.split('T')[0] : '';
		editValidUntil = voucher.valid_until
			? voucher.valid_until.split('T')[0]
			: '';
		editStatus = voucher.status || 'active';
		editDescription = voucher.description || '';
	}

	function cancelEdit() {
		isEditing = false;
	}

	async function saveEdit() {
		if (!voucher || !voucherId) return;

		// Validate required fields
		if (!editValidUntil) {
			toastStore.error(tr('vouchers.validUntilRequired'));
			return;
		}

		try {
			const response = await vouchersApi.update(voucherId, {
				merchant_id: editMerchantId || undefined,
				code: editCode,
				type: editType,
				value: editValue,
				currency: editCurrency || undefined,
				min_purchase_amount: editMinPurchaseAmount || undefined,
				barcode_type: editBarcodeType,
				usage_limit_type: editUsageLimitType,
				valid_from: editValidFrom ? `${editValidFrom}T00:00:00Z` : undefined,
				valid_until: `${editValidUntil}T23:59:59Z`,
				status: editStatus,
				description: editDescription || undefined
			});

			// Ensure permissions are set from the response
			if (response.permissions) {
				response.voucher.permissions = response.permissions;
			}
			voucher = response.voucher;
			shares = response.shares || [];
			isEditing = false;
			toastStore.success(tr('vouchers.updateSuccess'));
		} catch (err) {
			const message = err instanceof Error ? err.message : '';
			toastStore.error(message || tr('vouchers.updateError'));
		}
	}

	async function loadShares() {
		if (!voucher?.permissions?.is_owner || !voucherId) return;
		try {
			const response = await vouchersApi.get(voucherId);
			shares = response.shares || [];
		} catch (err) {
			pageLogger.error('Failed to load shares', { error: err });
		}
	}

	function promptDelete() {
		showDeleteModal = true;
	}

	async function confirmDelete() {
		try {
			if (!voucherId) {
				toastStore.error(tr('vouchers.deleteError'));
				return;
			}
			await vouchersApi.delete(voucherId);
			toastStore.success(tr('vouchers.deleteSuccess'));
			// Force full page reload to refresh the list (SPA navigation caches data)
			window.location.href = '/vouchers';
		} catch {
			toastStore.error(tr('vouchers.deleteError'));
		}
	}

	let isTogglingFavorite = $state(false);

	async function toggleFavorite() {
		if (isTogglingFavorite || !voucher || !voucherId) return;

		isTogglingFavorite = true;

		try {
			const response = await vouchersApi.toggleFavorite(voucherId);
			// Update favorite state directly from POST response
			// Avoids stale data from Service Worker cached GET responses
			voucher = { ...voucher, is_favorite: response.is_favorite };
		} catch {
			toastStore.error(tr('common.favoriteError'));
		} finally {
			isTogglingFavorite = false;
		}
	}

	async function handleShare() {
		try {
			if (!voucherId) {
				toastStore.error(tr('vouchers.sharing.shareError'));
				return;
			}
			if (shareEmails.length === 0) return;
			const response = await vouchersApi.createShare(voucherId, {
				emails: shareEmails
			});
			shares = response.shares || [];
			const { message, isError } = formatShareResult(response, tr);
			if (isError) toastStore.error(message);
			else toastStore.success(message);
			shareEmails = [];
			showShareForm = false;
		} catch (err) {
			const failed = shareResponseFromError(err);
			if (failed) {
				shares = failed.shares || shares;
				toastStore.error(formatShareResult(failed, tr).message);
				return;
			}
			const message = err instanceof Error ? err.message : '';
			toastStore.error(message || tr('vouchers.sharing.shareError'));
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
		if (!shareToDelete || !voucherId) return;
		try {
			await vouchersApi.deleteShare(voucherId, shareToDelete);
			toastStore.success(tr('vouchers.sharing.removeSuccess'));
			managingShareId = null;
			showDeleteShareModal = false;
			await loadShares();
		} catch {
			toastStore.error(tr('vouchers.sharing.removeError'));
		} finally {
			shareToDelete = null;
			showDeleteShareModal = false;
		}
	}

	function promptRevokeAll() {
		showRevokeAllModal = true;
	}

	async function confirmRevokeAll() {
		if (!voucherId) return;
		try {
			await vouchersApi.deleteAllShares(voucherId);
			toastStore.success(tr('vouchers.sharing.revokeAllSuccess'));
			managingShareId = null;
			await loadShares();
		} catch {
			toastStore.error(tr('vouchers.sharing.revokeAllError'));
		} finally {
			showRevokeAllModal = false;
		}
	}

	function promptTransfer() {
		showTransferModal = true;
	}

	async function confirmTransfer() {
		try {
			if (!voucherId) {
				toastStore.error(tr('vouchers.transfer.error'));
				return;
			}
			await vouchersApi.transfer(voucherId, { new_owner_email: transferEmail });
			toastStore.success(tr('vouchers.transfer.success'));
			// Remove transferred voucher from cache before redirect (prevents 403 on stale cache)
			await offlineDB.deleteVoucher(voucherId);
			// Force full page reload (user lost access after transfer)
			window.location.href = '/vouchers';
		} catch (err) {
			const message = err instanceof Error ? err.message : '';
			toastStore.error(message || tr('vouchers.transfer.error'));
		}
	}

	function getStatusBadge(status: string): { class: string; text: string } {
		switch (status) {
			case 'inactive':
				return {
					class: 'bg-border text-text-ink2',
					text: tr('vouchers.status.inactive')
				};
			case 'expired':
				return {
					class: 'bg-red-200 text-red-700',
					text: tr('vouchers.status.expired')
				};
			case 'used':
				return {
					class: 'bg-green-200 text-green-700',
					text: tr('vouchers.status.used')
				};
			default:
				return { class: '', text: '' };
		}
	}

	function formatValue(value: number, type: string, currency?: string): string {
		if (type === 'percentage') {
			return `${value}%`;
		} else if (type === 'fixed_amount') {
			return formatCurrency(value, currency || 'CHF', $locale);
		} else if (type === 'points_multiplier') {
			return `${value}${tr('vouchers.types.pointsMultiplierDisplay').trim()}`;
		} else if (type === 'bonus_points') {
			return `+${value}${tr('vouchers.types.bonusPointsDisplay')}`;
		} else if (type === 'free') {
			return tr('vouchers.types.freeDisplay');
		}
		return `${value.toFixed(2)}`;
	}
</script>

<svelte:head>
	<title
		>{voucher
			? `${voucher.merchant?.name || tr('common.voucher')} - ${tr('common.appName')}`
			: `${tr('common.voucher')} - ${tr('common.appName')}`}</title
	>
</svelte:head>

{#if isRefreshing}
	<div class="mb-6 flex justify-end">
		<span class="text-xs text-text-faint animate-pulse"
			>{tr('common.refreshing')}</span
		>
	</div>
{/if}

{#if isLoading}
	<LoadingSpinner />
{:else if voucher}
	<!-- Page header (view mode only; edit mode keeps its own form title). -->
	{#if !isEditing}
		<PageHeader
			title={voucher.merchant?.name || tr('vouchers.title')}
			eyebrow={voucher.usage_limit_type === 'single_use'
				? tr('vouchers.singleUseOnly')
				: tr('vouchers.multipleUse')}
			mobileActions={false}
		>
			{#snippet actions()}
				<ResourceActions
					{isOffline}
					isFavorite={voucher!.is_favorite}
					{isTogglingFavorite}
					canEdit={voucher!.permissions?.can_edit}
					favoriteTitleAdd={tr('common.addToFavorites')}
					favoriteTitleRemove={tr('common.removeFromFavorites')}
					ontoggleFavorite={toggleFavorite}
					onstartEdit={startEdit}
				/>
			{/snippet}
		</PageHeader>
		{#if voucher.owner && voucher.owner.id !== $authStore.user?.id}
			<p class="-mt-6 mb-6 text-xs text-text-faint">
				{tr('vouchers.sharedBy', {
					name: voucher.owner.first_name || voucher.owner.email
				})}
			</p>
		{/if}
	{/if}

	<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
		<!-- Left column: Voucher Details -->
		<div class="lg:col-span-2">
			{#if !isEditing}
				<!-- View Mode -->
				<div
					class="overflow-hidden rounded-xl border border-border/80 bg-white"
					style="border-left: 3px solid color-mix(in srgb, {voucher.merchant
						?.color || '#10B981'} 70%, transparent)"
				>
					<div
						class="p-6 {voucher.status && voucher.status !== 'valid'
							? 'opacity-50 grayscale'
							: ''}"
					>
						<!-- Duplicate Warning -->
						{#if voucher.duplicate_warning}
							<div class="mb-6">
								<DuplicateWarningBanner
									warning={voucher.duplicate_warning}
									resourceType="voucher"
									onNavigate={(id) => goto(resolve(`/vouchers/${id}`))}
								/>
							</div>
						{/if}

						<!-- Barcode Display -->
						<BarcodeDisplay
							value={voucher.code}
							type={voucher.barcode_type || 'CODE128'}
							status={voucher.status}
							statusBadge={voucher.status && voucher.status !== 'valid'
								? getStatusBadge(voucher.status)
								: undefined}
							validFrom={voucher.valid_from}
							validUntil={voucher.valid_until}
							displayValue={formatValue(
								voucher.value,
								voucher.type,
								voucher.currency
							)}
							minPurchaseInfo={voucher.min_purchase_amount &&
							voucher.min_purchase_amount > 0
								? `${tr('vouchers.minPurchaseAmount')}: ${formatCurrency(voucher.min_purchase_amount, voucher.currency || 'CHF', $locale)}`
								: undefined}
							description={voucher.description}
						/>
					</div>
				</div>
			{:else}
				<!-- Edit Mode -->
				<div class="overflow-hidden rounded-xl border border-border bg-white">
					<div class="p-6">
						<VoucherForm
							bind:code={editCode}
							bind:merchantId={editMerchantId}
							bind:type={editType}
							bind:value={editValue}
							bind:currency={editCurrency}
							bind:minPurchaseAmount={editMinPurchaseAmount}
							bind:barcodeType={editBarcodeType}
							bind:validFrom={editValidFrom}
							bind:validUntil={editValidUntil}
							bind:usageLimitType={editUsageLimitType}
							bind:description={editDescription}
							onSubmit={saveEdit}
							onCancel={cancelEdit}
							isLoading={false}
							submitLabel={tr('common.save')}
						/>
						{#if voucher.permissions?.can_delete}
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
												d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"
											></path>
										</svg>
									{/if}
									{tr('vouchers.deleteButton')}
								</button>
							</div>
						{/if}
					</div>
				</div>
			{/if}
		</div>

		<!-- Right column: Transfer & Sharing Info (only for owners) -->
		<div class="lg:col-span-1 space-y-4">
			{#if voucher.permissions?.is_owner}
				<!-- Transfer Box -->
				<TransferBox
					{isOffline}
					openButtonLabel={tr('vouchers.transfer.button')}
					transferButtonLabel={tr('vouchers.transfer.transferButton')}
					warningTitle={tr('vouchers.transfer.warning')}
					warningDetails={tr('vouchers.transfer.warningDetail')}
					emailLabel={tr('vouchers.transfer.newOwner')}
					emailHint={tr('giftCards.sharing.userMustBeRegistered')}
					whatHappensLabel={tr('vouchers.transfer.whatHappens')}
					details={[
						tr('vouchers.transfer.newOwnerGetsRights'),
						tr('vouchers.transfer.allSharesDeleted'),
						tr('vouchers.transfer.youLoseAccess'),
						tr('vouchers.transfer.transferLogged')
					]}
					bind:email={transferEmail}
					ontransfer={promptTransfer}
				/>

				<!-- Sharing Box -->
				<div class="rounded-xl border border-border bg-white p-6">
					<div class="flex justify-between items-center mb-4">
						<h3 class="text-lg font-semibold text-text">
							{tr('common.share')}
						</h3>
						{#if !showShareForm}
							<button
								onclick={() => (showShareForm = true)}
								disabled={isOffline}
								class="btn btn-xs btn-primary whitespace-nowrap flex items-center gap-1.5 {isOffline
									? 'opacity-50 cursor-not-allowed pointer-events-none blur-[0.5px]'
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
											d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"
										></path>
									</svg>
								{:else}
									<span>+</span>
								{/if}
								{tr('common.add')}
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
								label={tr('vouchers.sharing.userEmail')}
								hint={tr('vouchers.sharing.hint')}
								inputId="share-email-input"
								disabled={isOffline}
							/>

							<div class="bg-white border border-accent-200 rounded-lg p-3">
								<h4 class="font-medium text-accent-900 text-sm mb-2">
									{tr('vouchers.sharing.whatIsShared')}
								</h4>
								<ul class="text-xs text-accent-800 space-y-1">
									<li>{tr('vouchers.sharing.sharedCode')}</li>
									<li>{tr('vouchers.sharing.sharedDetails')}</li>
									<li>{tr('vouchers.sharing.sharedDescription')}</li>
								</ul>
								<p class="text-xs text-accent-hover mt-2 italic">
									{tr('vouchers.sharing.readOnlyNote')}
								</p>
							</div>

							<div class="flex gap-2">
								<button
									onclick={handleShare}
									disabled={isOffline}
									class="btn btn-primary flex-1 {isOffline
										? 'opacity-50 cursor-not-allowed'
										: ''}"
								>
									{tr('vouchers.sharing.shareNow')}
								</button>
								<button
									onclick={() => {
										showShareForm = false;
										shareEmails = [];
									}}
									class="btn btn-ghost"
								>
									{tr('common.cancel')}
								</button>
							</div>
						</div>
					{/if}

					{#if shares.length > 0}
						<div class="space-y-3">
							{#each shares as share (share.shared_with_user.id)}
								<ShareListItem
									{share}
									isEditing={managingShareId === share.shared_with_user.id}
									{isOffline}
									editButtonLabel={tr('common.manage')}
									deleteButtonLabel={tr('vouchers.sharing.removeShare')}
									alwaysViewOnly={true}
									onstartEdit={() =>
										startManageShare(share.shared_with_user.id)}
									oncancel={cancelManageShare}
									ondelete={() => promptDeleteShare(share.shared_with_user.id)}
								>
									<div
										class="bg-yellow-50 border border-yellow-200 rounded-lg p-3"
									>
										<p class="text-xs font-medium text-yellow-800 mb-1">
											{tr('vouchers.sharing.alwaysReadOnly')}
										</p>
										<p class="text-xs text-yellow-700">
											{tr('vouchers.sharing.canOnlyRemove')}
										</p>
									</div>
								</ShareListItem>
							{/each}
						</div>
						<button
							type="button"
							onclick={promptRevokeAll}
							disabled={isOffline}
							class="btn btn-ghost text-red-600 mt-3 w-full disabled:opacity-50"
						>
							{tr('vouchers.sharing.revokeAll')}
						</button>
					{:else}
						<p class="text-sm text-text-subtle text-center py-4">
							{tr('vouchers.sharing.notSharedYet')}
						</p>
					{/if}
				</div>
			{/if}
		</div>
	</div>

	<!-- Confirmation Modals -->
	<ConfirmModal
		isOpen={showDeleteModal}
		title={tr('vouchers.deleteConfirm')}
		message={tr('vouchers.deleteConfirmMessage')}
		confirmText={tr('common.delete')}
		cancelText={tr('common.cancel')}
		variant="danger"
		onconfirm={confirmDelete}
		oncancel={() => (showDeleteModal = false)}
	/>

	<ConfirmModal
		isOpen={showDeleteShareModal}
		title={tr('vouchers.sharing.removeConfirm')}
		message={tr('vouchers.sharing.removeConfirmMessage')}
		confirmText={tr('common.remove')}
		cancelText={tr('common.cancel')}
		variant="danger"
		onconfirm={confirmDeleteShare}
		oncancel={() => (showDeleteShareModal = false)}
	/>

	<ConfirmModal
		isOpen={showRevokeAllModal}
		title={tr('vouchers.sharing.revokeAllConfirm')}
		message={tr('vouchers.sharing.revokeAllConfirmMessage')}
		confirmText={tr('vouchers.sharing.revokeAll')}
		cancelText={tr('common.cancel')}
		variant="danger"
		onconfirm={confirmRevokeAll}
		oncancel={() => (showRevokeAllModal = false)}
	/>

	<ConfirmModal
		isOpen={showTransferModal}
		title={tr('vouchers.transfer.confirmTitle')}
		message={tr('vouchers.transfer.confirmMessage')}
		confirmText={tr('vouchers.transfer.transferButton')}
		cancelText={tr('common.cancel')}
		variant="transfer"
		onconfirm={confirmTransfer}
		oncancel={() => (showTransferModal = false)}
	/>
{/if}
