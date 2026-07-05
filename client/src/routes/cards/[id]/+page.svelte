<script lang="ts">
	import { onMount, tick } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { get } from 'svelte/store';
	import { authStore } from '$lib/stores/auth';
	import { isOnline } from '$lib/stores/offline';
	import { t } from '$lib/stores/i18n';
	import { cardsApi, merchantsApi, ApiError } from '$lib/api';
	import { offlineDB } from '$lib/stores/offline-db';
	import { toastStore } from '$lib/stores/toast';
	import BarcodeDisplay from '$lib/components/BarcodeDisplay.svelte';
	import DuplicateWarningBanner from '$lib/components/DuplicateWarningBanner.svelte';
	import ConfirmModal from '$lib/components/ConfirmModal.svelte';
	import CardForm from '$lib/components/cards/CardForm.svelte';
	import TransferBox from '$lib/components/TransferBox.svelte';
	import ResourceHeader from '$lib/components/ResourceHeader.svelte';

	import type { CardDTO, ShareDTO, MerchantDTO } from '$lib/types/api';
	import { logger } from '$lib/utils/logger';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import EmailAutocomplete from '$lib/components/EmailAutocomplete.svelte';
	import SharePermissions from '$lib/components/SharePermissions.svelte';
	import ShareListItem from '$lib/components/ShareListItem.svelte';
	import {
		formatShareResult,
		shareResponseFromError
	} from '$lib/utils/share-result';

	// Svelte 5 compatible translation wrapper
	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);
	const pageLogger = logger.child('CardDetailsPage');

	const cardId = $derived($page.params.id!);

	let card = $state<CardDTO | null>(null);
	let shares = $state<ShareDTO[]>([]);
	let isLoading = $state(true);
	let isRefreshing = $state(false);
	let showShareForm = $state(false);
	let shareEmails = $state<string[]>([]);
	let canEdit = $state(false);
	let canDelete = $state(false);
	let transferEmail = $state('');
	let isEditing = $state(false);
	let merchants = $state<MerchantDTO[]>([]);

	// Share editing state
	let editingShareId = $state<string | null>(null);
	let editShareCanEdit = $state(false);
	let editShareCanDelete = $state(false);

	// Edit form fields
	let editMerchantId = $state('');
	let editProgram = $state('');
	let editCardNumber = $state('');
	let editBarcodeType = $state('CODE128');
	let editStatus = $state('active');
	let editNotes = $state('');

	// Modal state
	let showDeleteModal = $state(false);
	let showDeleteShareModal = $state(false);
	let showTransferModal = $state(false);
	let shareToDelete: string | null = null;

	const isOffline = $derived(!$isOnline);

	onMount(async () => {
		await Promise.all([loadCard(), loadMerchants()]);
	});

	async function loadCard() {
		isLoading = true;
		try {
			// Phase 1: Show cached data immediately
			const cached = await cardsApi.getCached(cardId);
			if (cached) {
				card = cached.card;
				shares = cached.shares;
				isLoading = false;
				isRefreshing = true;
			}

			// Phase 2: Fetch fresh data from network
			if (navigator.onLine) {
				try {
					const fresh = await cardsApi.get(cardId);
					card = fresh.card;
					shares = fresh.shares || [];
				} catch (err: unknown) {
					if (
						err instanceof ApiError &&
						(err.status === 403 || err.status === 404)
					) {
						await offlineDB.deleteCard(cardId);
						toastStore.error(tr('cards.loadError'));
						goto(resolve('/cards'));
						return;
					}
					if (!cached) {
						toastStore.error(tr('cards.loadError'));
						goto(resolve('/cards'));
						return;
					}
					// Transient error with cached data available - show warning, don't redirect
					toastStore.warning(tr('common.offlineMode'));
				}
			} else if (!cached) {
				toastStore.error(tr('cards.loadError'));
				goto(resolve('/cards'));
			}
		} catch {
			toastStore.error(tr('cards.loadError'));
			goto(resolve('/cards'));
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
			pageLogger.error('Failed to load merchants:', err);
		}
	}

	async function startEdit() {
		if (!card) return;

		// Ensure merchants are loaded before entering edit mode
		if (merchants.length === 0) {
			await loadMerchants();
		}

		isEditing = true;
		editMerchantId = card.merchant?.id || '';
		editProgram = card.program || '';
		editCardNumber = card.card_number;
		editBarcodeType = card.barcode_type || 'CODE128';
		editStatus = card.status || 'active';
		editNotes = card.notes || '';

		pageLogger.debug('Edit initialized:', {
			cardMerchantId: card.merchant?.id,
			editMerchantId: editMerchantId,
			merchantsLoaded: merchants.length,
			cardMerchant: card.merchant
		});
	}

	function cancelEdit() {
		isEditing = false;
	}

	async function saveEdit() {
		if (!card) return;
		try {
			pageLogger.debug('Saving edit with data:', {
				merchant_id: editMerchantId || undefined,
				program: editProgram || undefined,
				card_number: editCardNumber,
				barcode_type: editBarcodeType,
				status: editStatus,
				notes: editNotes || undefined
			});

			const response = await cardsApi.update(cardId, {
				merchant_id: editMerchantId || undefined,
				program: editProgram || undefined,
				card_number: editCardNumber,
				barcode_type: editBarcodeType,
				status: editStatus,
				notes: editNotes || undefined
			});

			pageLogger.debug('Update response:', response);

			// Ensure permissions are set from the response
			if (response.permissions) {
				response.card.permissions = response.permissions;
			}
			card = response.card;
			shares = response.shares || [];
			isEditing = false;
			toastStore.success(tr('cards.updateSuccess'));

			// Wait for DOM to update before completing
			await tick();
		} catch (err: unknown) {
			pageLogger.error('Save error:', err);
			toastStore.error(
				err instanceof Error ? err.message : tr('cards.updateError')
			);
		}
	}

	async function loadShares() {
		if (!card?.permissions?.is_owner) return;
		try {
			const response = await cardsApi.get(cardId);
			shares = response.shares || [];
		} catch (err) {
			pageLogger.error('Failed to load shares:', err);
		}
	}

	function promptDelete() {
		showDeleteModal = true;
	}

	async function confirmDelete() {
		try {
			await cardsApi.delete(cardId);
			toastStore.success(tr('cards.deleteSuccess'));
			// Force full page reload to refresh the list (SPA navigation caches data)
			window.location.href = '/cards';
		} catch {
			toastStore.error(tr('cards.deleteError'));
		}
	}

	let isTogglingFavorite = $state(false);

	async function toggleFavorite() {
		if (isTogglingFavorite || !card || !cardId) return;

		isTogglingFavorite = true;

		try {
			const response = await cardsApi.toggleFavorite(cardId);
			// Update favorite state directly from POST response
			// Avoids stale data from Service Worker cached GET responses
			card = { ...card, is_favorite: response.is_favorite };
		} catch {
			toastStore.error(tr('common.favoriteError'));
		} finally {
			isTogglingFavorite = false;
		}
	}

	async function handleShare() {
		if (shareEmails.length === 0) return;
		try {
			const response = await cardsApi.createShare(cardId, {
				emails: shareEmails,
				can_edit: canEdit,
				can_delete: canDelete
			});
			shares = response.shares || [];
			const { message, isError } = formatShareResult(response, tr);
			if (isError) toastStore.error(message);
			else toastStore.success(message);
			shareEmails = [];
			canEdit = false;
			canDelete = false;
			showShareForm = false;
		} catch (err: unknown) {
			const failed = shareResponseFromError(err);
			if (failed) {
				shares = failed.shares || shares;
				toastStore.error(formatShareResult(failed, tr).message);
			} else {
				toastStore.error(
					err instanceof Error ? err.message : tr('cards.sharing.shareError')
				);
			}
		}
	}

	function startEditShare(share: ShareDTO) {
		editingShareId = share.shared_with_user.id;
		editShareCanEdit = share.can_edit;
		editShareCanDelete = share.can_delete;
	}

	function cancelEditShare() {
		editingShareId = null;
		editShareCanEdit = false;
		editShareCanDelete = false;
	}

	async function saveShareEdit(sharedWithID: string) {
		try {
			const response = await cardsApi.updateShare(cardId, sharedWithID, {
				can_edit: editShareCanEdit,
				can_delete: editShareCanDelete
			});
			shares = response.shares || [];
			editingShareId = null;
			toastStore.success(tr('cards.sharing.updateSuccess'));
		} catch (err: unknown) {
			toastStore.error(
				err instanceof Error ? err.message : tr('cards.sharing.updateError')
			);
		}
	}

	function promptDeleteShare(sharedWithID: string) {
		shareToDelete = sharedWithID;
		showDeleteShareModal = true;
	}

	async function confirmDeleteShare() {
		if (!shareToDelete) return;
		try {
			await cardsApi.deleteShare(cardId, shareToDelete);
			toastStore.success(tr('cards.sharing.removeSuccess'));
			showDeleteShareModal = false;
			await loadShares();
		} catch {
			toastStore.error(tr('cards.sharing.removeError'));
		} finally {
			shareToDelete = null;
			showDeleteShareModal = false;
		}
	}

	function promptTransfer() {
		showTransferModal = true;
	}

	async function confirmTransfer() {
		try {
			await cardsApi.transfer(cardId, { new_owner_email: transferEmail });
			toastStore.success(tr('cards.transfer.success'));
			// Remove transferred card from cache before redirect (prevents 403 on stale cache)
			await offlineDB.deleteCard(cardId);
			// Force full page reload (user lost access after transfer)
			window.location.href = '/cards';
		} catch (err: unknown) {
			toastStore.error(
				err instanceof Error ? err.message : tr('cards.transfer.error')
			);
		}
	}

	function getStatusBadge(status: string): { class: string; text: string } {
		switch (status) {
			case 'inactive':
				return {
					class: 'bg-gray-200 text-gray-700',
					text: tr('cards.status.inactive')
				};
			case 'expired':
				return {
					class: 'bg-red-200 text-red-700',
					text: tr('cards.status.expired')
				};
			case 'lost':
				return {
					class: 'bg-orange-200 text-orange-700',
					text: tr('cards.status.lost')
				};
			case 'blocked':
				return {
					class: 'bg-red-200 text-red-700',
					text: tr('cards.status.blocked')
				};
			default:
				return { class: '', text: '' };
		}
	}
</script>

<svelte:head>
	<title
		>{card
			? `${card.merchant?.name || tr('cards.title')} - ${tr('common.appName')}`
			: `${tr('cards.title')} - ${tr('common.appName')}`}</title
	>
</svelte:head>

<div class="mb-6 flex items-center justify-between">
	<a href={resolve('/cards')} class="text-cyan-600 hover:text-cyan-700"
		>{tr('common.backToOverview')}</a
	>
	{#if isRefreshing}
		<span class="text-xs text-gray-400 animate-pulse"
			>{tr('common.refreshing')}</span
		>
	{/if}
</div>

{#if isLoading}
	<LoadingSpinner />
{:else if card}
	<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
		<!-- Left column: Card Details -->
		<div class="lg:col-span-2">
			{#if !isEditing}
				<!-- View Mode -->
				<div
					class="bg-white rounded-lg shadow-lg overflow-hidden"
					style="border-left: 6px solid {card.merchant?.color || '#3B82F6'}"
				>
					<div
						class="p-6 {card.status && card.status !== 'active'
							? 'opacity-50 grayscale'
							: ''}"
					>
						<!-- Header -->
						<div class="mb-6">
							<ResourceHeader
								{isOffline}
								isFavorite={card.is_favorite}
								{isTogglingFavorite}
								canEdit={card.permissions?.can_edit}
								favoriteTitleAdd={tr('common.addToFavorites')}
								favoriteTitleRemove={tr('common.removeFromFavorites')}
								ontoggleFavorite={toggleFavorite}
								onstartEdit={startEdit}
							>
								{@const c = card!}
								<div class="flex items-baseline gap-2 flex-wrap mb-1">
									<h1
										class="text-lg sm:text-xl md:text-2xl font-bold text-gray-900"
									>
										{#if c.merchant}
											{c.merchant.name}
										{:else}
											{tr('common.card')}
										{/if}
										{#if c.program}
											<span
												class="text-sm sm:text-base md:text-lg font-normal text-gray-500 ml-1"
												>{c.program}</span
											>
										{/if}
									</h1>
									{#if c.owner && c.owner.id !== $authStore.user?.id}
										<span class="text-xs text-gray-400">
											{tr('cards.sharedBy', {
												name: c.owner.first_name || c.owner.email
											})}
										</span>
									{/if}
								</div>
							</ResourceHeader>

							<!-- Duplicate Warning -->
							{#if card.duplicate_warning}
								<DuplicateWarningBanner
									warning={card.duplicate_warning}
									resourceType="card"
									onNavigate={(id) => goto(resolve(`/cards/${id}`))}
								/>
							{/if}
						</div>

						<!-- Barcode Display -->
						<BarcodeDisplay
							value={card.card_number}
							type={card.barcode_type || 'CODE128'}
							status={card.status}
							statusBadge={card.status !== 'active'
								? getStatusBadge(card.status ?? 'active')
								: undefined}
						/>

						<!-- Notes -->
						{#if card.notes}
							<div
								class="mt-4 bg-yellow-50 border-l-4 border-yellow-400 p-3 rounded"
							>
								<p class="text-sm text-gray-700">{card.notes}</p>
							</div>
						{/if}
					</div>
				</div>
			{:else}
				<!-- Edit Mode -->
				<div
					class="bg-white rounded-lg shadow-lg overflow-hidden"
					style="border-top: 4px solid #3B82F6"
				>
					<div class="p-6">
						<CardForm
							bind:cardNumber={editCardNumber}
							bind:merchantId={editMerchantId}
							bind:program={editProgram}
							bind:barcodeType={editBarcodeType}
							bind:status={editStatus}
							bind:notes={editNotes}
							onSubmit={saveEdit}
							onCancel={cancelEdit}
							isLoading={false}
							submitLabel={tr('common.save')}
						/>
						{#if card.permissions?.can_delete}
							<div class="pt-4 mt-4 border-t border-gray-200">
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
									{tr('cards.deleteButton')}
								</button>
							</div>
						{/if}
					</div>
				</div>
			{/if}
		</div>

		<!-- Right column: Transfer & Sharing Info (only for owners) -->
		<div class="lg:col-span-1 space-y-4">
			{#if card.permissions?.is_owner}
				<!-- Transfer Box -->
				<TransferBox
					{isOffline}
					openButtonLabel={tr('cards.transfer.button')}
					transferButtonLabel={tr('cards.transfer.transferButton')}
					warningTitle={tr('cards.transfer.warning')}
					warningDetails={tr('giftCards.transfer.warningDetails')}
					emailLabel={tr('cards.transfer.newOwnerEmail')}
					emailHint={tr('giftCards.sharing.userMustBeRegistered')}
					whatHappensLabel={tr('cards.transfer.whatHappens')}
					details={[
						tr('cards.transfer.newOwnerGetsRights'),
						tr('cards.transfer.allSharesDeleted'),
						tr('cards.transfer.youLoseAccess'),
						tr('cards.transfer.transferLogged')
					]}
					bind:email={transferEmail}
					ontransfer={promptTransfer}
				/>

				<!-- Sharing Box -->
				<div class="bg-white rounded-lg shadow-lg p-6">
					<div class="flex justify-between items-center mb-4">
						<h3 class="text-lg font-semibold text-gray-900">
							{tr('common.share')}
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
							class="border border-cyan-200 bg-cyan-50 rounded-lg p-4 space-y-4 mb-4"
						>
							<EmailAutocomplete
								multiple
								bind:values={shareEmails}
								label={tr('cards.sharing.userEmail')}
								hint={tr('giftCards.sharing.userMustBeRegistered')}
								inputId="share-email-input"
								disabled={isOffline}
							/>

							<SharePermissions
								bind:canEdit
								bind:canDelete
								labelEdit={tr('cards.sharing.canEdit')}
								labelEditDesc={tr('cards.sharing.canEditDesc')}
								labelDelete={tr('cards.sharing.canDelete')}
								labelDeleteDesc={tr('cards.sharing.canDeleteDesc')}
							/>

							<div class="bg-white border border-cyan-200 rounded-lg p-3">
								<h4 class="font-medium text-cyan-900 text-sm mb-2">
									{tr('cards.sharing.whatIsShared')}
								</h4>
								<ul class="text-xs text-cyan-800 space-y-1">
									<li>{tr('cards.sharing.sharedItemCardNumber')}</li>
									<li>{tr('cards.sharing.sharedItemDetails')}</li>
									<li>{tr('cards.sharing.sharedItemNotes')}</li>
								</ul>
							</div>

							<div class="flex gap-2">
								<button
									onclick={handleShare}
									disabled={isOffline}
									class="btn btn-primary flex-1"
								>
									{tr('giftCards.sharing.shareNow')}
								</button>
								<button
									onclick={() => {
										showShareForm = false;
										shareEmails = [];
										canEdit = false;
										canDelete = false;
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
									isEditing={editingShareId === share.shared_with_user.id}
									{isOffline}
									onstartEdit={() => startEditShare(share)}
									onsave={() => saveShareEdit(share.shared_with_user.id)}
									oncancel={cancelEditShare}
									ondelete={() => promptDeleteShare(share.shared_with_user.id)}
								>
									<SharePermissions
										bind:canEdit={editShareCanEdit}
										bind:canDelete={editShareCanDelete}
										labelEdit={tr('cards.sharing.canEdit')}
										labelEditDesc={tr('cards.sharing.canEditDesc')}
										labelDelete={tr('cards.sharing.canDelete')}
										labelDeleteDesc={tr('cards.sharing.canDeleteDesc')}
									/>
								</ShareListItem>
							{/each}
						</div>
					{:else}
						<p class="text-sm text-gray-500 text-center py-4">
							{tr('giftCards.sharing.notSharedYet')}
						</p>
					{/if}
				</div>
			{/if}
		</div>
	</div>

	<!-- Confirmation Modals -->
	<ConfirmModal
		isOpen={showDeleteModal}
		title={tr('cards.deleteConfirm')}
		message={tr('cards.deleteConfirmMessage')}
		confirmText={tr('common.delete')}
		cancelText={tr('common.cancel')}
		variant="danger"
		onconfirm={confirmDelete}
		oncancel={() => (showDeleteModal = false)}
	/>

	<ConfirmModal
		isOpen={showDeleteShareModal}
		title={tr('cards.sharing.removeConfirm')}
		message={tr('cards.sharing.removeConfirmMessage')}
		confirmText={tr('common.remove')}
		cancelText={tr('common.cancel')}
		variant="danger"
		onconfirm={confirmDeleteShare}
		oncancel={() => (showDeleteShareModal = false)}
	/>

	<ConfirmModal
		isOpen={showTransferModal}
		title={tr('cards.transfer.confirmTitle')}
		message={tr('cards.transfer.confirmMessage')}
		confirmText={tr('cards.transfer.transferButton')}
		cancelText={tr('common.cancel')}
		variant="transfer"
		onconfirm={confirmTransfer}
		oncancel={() => (showTransferModal = false)}
	/>
{/if}
