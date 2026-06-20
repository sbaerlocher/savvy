<script lang="ts">
	import { onMount, tick } from 'svelte';
	import { page } from '$app/stores';
	import { goto, invalidateAll } from '$app/navigation';
	import { get } from 'svelte/store';
	import { authStore } from '$lib/stores/auth';
	import { isOnline } from '$lib/stores/offline';
	import { t, locale } from '$lib/stores/i18n';
	import { giftCardsApi, merchantsApi, ApiError } from '$lib/api';
	import { offlineDB } from '$lib/stores/offline-db';
	import { toastStore } from '$lib/stores/toast';
	import { formatCurrency } from '$lib/utils/currency';
	import BarcodeDisplay from '$lib/components/BarcodeDisplay.svelte';
	import DuplicateWarningBanner from '$lib/components/DuplicateWarningBanner.svelte';

	import ConfirmModal from '$lib/components/ConfirmModal.svelte';
	import GiftCardForm from '$lib/components/gift-cards/GiftCardForm.svelte';
	import type {
		GiftCardDTO,
		ShareDTO,
		MerchantDTO,
		TransactionDTO
	} from '$lib/types/api';
	import EmailAutocomplete from '$lib/components/EmailAutocomplete.svelte';
	import SharePermissions from '$lib/components/SharePermissions.svelte';
	import ShareListItem from '$lib/components/ShareListItem.svelte';
	import { logger } from '$lib/utils/logger';
	import { detectBarcodeType } from '$lib/utils/barcode';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import TransferBox from '$lib/components/TransferBox.svelte';
	import ResourceHeader from '$lib/components/ResourceHeader.svelte';

	// Svelte 5 compatible translation wrapper
	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);
	const currentLocale = $derived($locale || 'de-DE');
	const pageLogger = logger.child('GiftCardDetailsPage');

	const giftCardId = $derived($page.params.id);

	let giftCard = $state<GiftCardDTO | null>(null);
	let shares = $state<ShareDTO[]>([]);
	let transactions = $state<TransactionDTO[]>([]);
	let isLoading = $state(true);
	let isRefreshing = $state(false);
	let showShareForm = $state(false);
	let showTransactionForm = $state(false);
	let shareEmail = $state('');
	let canEdit = $state(false);
	let canDelete = $state(false);
	let canEditTransactions = $state(false);
	let transferEmail = $state('');
	let isEditing = $state(false);
	let merchants = $state<MerchantDTO[]>([]);

	// Transaction form
	let transactionAmount = $state(0);
	let transactionDescription = $state('');
	let transactionDate = $state(new Date().toISOString().split('T')[0]); // Today's date in YYYY-MM-DD format

	// Edit form fields
	let editMerchantId = $state('');
	let editCardNumber = $state('');
	let editInitialBalance = $state(0);
	let editCurrency = $state('CHF');
	let editPin = $state('');
	let editBarcodeType = $state('CODE128');
	let editExpiresAt = $state('');
	let editNotes = $state('');

	// Share editing state
	let editingShareId = $state<string | null>(null);
	let editShareCanEdit = $state(false);
	let editShareCanDelete = $state(false);
	let editShareCanEditTransactions = $state(false);

	// Scanner state
	let scannerOpen = $state(false);

	// Modal state
	let showDeleteModal = $state(false);
	let showDeleteShareModal = $state(false);
	let showTransferModal = $state(false);
	let showDeleteTransactionModal = $state(false);
	let shareToDelete: string | null = null;
	let transactionToDelete: string | null = null;

	const isOffline = $derived(!$isOnline);
	const percentageRemaining = $derived(
		giftCard
			? Math.round((giftCard.current_balance / giftCard.initial_balance) * 100)
			: 0
	);

	onMount(async () => {
		await Promise.all([loadGiftCard(), loadMerchants(), loadTransactions()]);
	});

	async function loadGiftCard() {
		isLoading = true;
		try {
			if (!giftCardId) {
				toastStore.error(tr('giftCards.loadError'));
				goto('/gift-cards');
				return;
			}

			// Phase 1: Show cached data immediately
			const cached = await giftCardsApi.getCached(giftCardId);
			if (cached) {
				giftCard = cached.gift_card;
				shares = cached.shares;
				transactions = cached.transactions || [];
				isLoading = false;
				isRefreshing = true;
			}

			// Phase 2: Fetch fresh data from network
			if (navigator.onLine) {
				try {
					const fresh = await giftCardsApi.get(giftCardId);
					giftCard = fresh.gift_card;
					shares = fresh.shares || [];
				} catch (err: unknown) {
					if (
						err instanceof ApiError &&
						(err.status === 403 || err.status === 404)
					) {
						await offlineDB.deleteGiftCard(giftCardId);
						toastStore.error(tr('giftCards.loadError'));
						goto('/gift-cards');
						return;
					}
					if (!cached) {
						toastStore.error(tr('giftCards.loadError'));
						goto('/gift-cards');
						return;
					}
					// Transient error with cached data available - show warning, don't redirect
					toastStore.warning(tr('common.offlineMode'));
				}
			} else if (!cached) {
				toastStore.error(tr('giftCards.loadError'));
				goto('/gift-cards');
			}
		} catch (err) {
			toastStore.error(tr('giftCards.loadError'));
			goto('/gift-cards');
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

	async function loadTransactions() {
		try {
			if (!giftCardId) return;
			const response = await giftCardsApi.listTransactions(giftCardId);
			transactions = response.transactions || [];
		} catch (err) {
			pageLogger.error('Failed to load transactions:', err);
		}
	}

	async function startEdit() {
		if (!giftCard) return;

		if (merchants.length === 0) {
			await loadMerchants();
		}

		isEditing = true;
		editMerchantId = giftCard.merchant?.id || '';
		editCardNumber = giftCard.card_number;
		editInitialBalance = giftCard.initial_balance;
		editCurrency = giftCard.currency;
		editPin = giftCard.pin || '';
		editBarcodeType = giftCard.barcode_type || 'CODE128';
		editExpiresAt = giftCard.expires_at
			? giftCard.expires_at.split('T')[0]
			: '';
		editNotes = giftCard.notes || '';
	}

	function cancelEdit() {
		isEditing = false;
	}

	async function saveEdit() {
		if (!giftCard || !giftCardId) return;
		try {
			const response = await giftCardsApi.update(giftCardId, {
				merchant_id: editMerchantId || undefined,
				card_number: editCardNumber,
				initial_balance: editInitialBalance,
				currency: editCurrency,
				pin: editPin || undefined,
				barcode_type: editBarcodeType,
				expires_at: editExpiresAt ? `${editExpiresAt}T00:00:00Z` : undefined,
				notes: editNotes || undefined
			});

			// Ensure permissions are set from the response
			if (response.permissions) {
				response.gift_card.permissions = response.permissions;
			}
			giftCard = response.gift_card;
			shares = response.shares || [];
			isEditing = false;
			toastStore.success(tr('giftCards.updateSuccess'));
		} catch (err: any) {
			toastStore.error(err.message || tr('giftCards.updateError'));
		}
	}

	function handleScan(event: { barcode: string; format?: string }) {
		editCardNumber = event.barcode;
		// Use detected format from scanner if available, otherwise fallback to detectBarcodeType
		editBarcodeType = event.format || detectBarcodeType(event.barcode);
		scannerOpen = false;
		toastStore.success(tr('common.scanSuccess'));
	}

	function handleCardNumberInput(event: Event) {
		const input = event.target as HTMLInputElement;
		editCardNumber = input.value;
		if (editCardNumber.trim()) {
			editBarcodeType = detectBarcodeType(editCardNumber);
		}
	}

	async function loadShares() {
		if (!giftCard?.permissions?.is_owner || !giftCardId) return;
		try {
			const response = await giftCardsApi.get(giftCardId);
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
			if (!giftCardId) {
				toastStore.error(tr('giftCards.deleteError'));
				return;
			}
			await giftCardsApi.delete(giftCardId);
			toastStore.success(tr('giftCards.deleteSuccess'));
			// Force full page reload to refresh the list (SPA navigation caches data)
			window.location.href = '/gift-cards';
		} catch (err) {
			toastStore.error(tr('giftCards.deleteError'));
		}
	}

	let isTogglingFavorite = $state(false);

	async function toggleFavorite() {
		if (isTogglingFavorite || !giftCard || !giftCardId) return;

		isTogglingFavorite = true;

		try {
			const response = await giftCardsApi.toggleFavorite(giftCardId);
			// Update favorite state directly from POST response
			// Avoids stale data from Service Worker cached GET responses
			giftCard = { ...giftCard, is_favorite: response.is_favorite };
		} catch (err) {
			toastStore.error(tr('common.favoriteError'));
		} finally {
			isTogglingFavorite = false;
		}
	}

	async function handleShare() {
		try {
			if (!giftCardId) {
				toastStore.error(tr('giftCards.sharing.shareError'));
				return;
			}
			const response = await giftCardsApi.createShare(giftCardId, {
				email: shareEmail,
				can_edit: canEdit,
				can_delete: canDelete,
				can_edit_transactions: canEditTransactions
			});
			shares = response.shares || [];
			toastStore.success(tr('giftCards.sharing.shareSuccess'));
			shareEmail = '';
			canEdit = false;
			canDelete = false;
			canEditTransactions = false;
			showShareForm = false;
		} catch (err: any) {
			toastStore.error(err.message || tr('giftCards.sharing.shareError'));
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
			if (!giftCardId) {
				toastStore.error(tr('giftCards.sharing.updateError'));
				return;
			}
			const response = await giftCardsApi.updateShare(
				giftCardId,
				sharedWithID,
				{
					can_edit: editShareCanEdit,
					can_delete: editShareCanDelete,
					can_edit_transactions: editShareCanEditTransactions
				}
			);
			shares = response.shares || [];
			editingShareId = null;
			toastStore.success(tr('giftCards.sharing.updateSuccess'));
		} catch (err: any) {
			toastStore.error(err.message || tr('giftCards.sharing.updateError'));
		}
	}

	function promptDeleteShare(sharedWithID: string) {
		shareToDelete = sharedWithID;
		showDeleteShareModal = true;
	}

	async function confirmDeleteShare() {
		if (!shareToDelete || !giftCardId) return;
		try {
			await giftCardsApi.deleteShare(giftCardId, shareToDelete);
			toastStore.success(tr('giftCards.sharing.removeSuccess'));
			showDeleteShareModal = false;
			await loadShares();
		} catch (err) {
			toastStore.error(tr('giftCards.sharing.removeError'));
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
			if (!giftCardId) {
				toastStore.error(tr('giftCards.transfer.error'));
				return;
			}
			await giftCardsApi.transfer(giftCardId, {
				new_owner_email: transferEmail
			});
			toastStore.success(tr('giftCards.transfer.success'));
			// Remove transferred gift card from cache before redirect (prevents 403 on stale cache)
			await offlineDB.deleteGiftCard(giftCardId);
			// Force full page reload (user lost access after transfer)
			window.location.href = '/gift-cards';
		} catch (err: any) {
			toastStore.error(err.message || tr('giftCards.transfer.error'));
		}
	}

	async function handleAddTransaction() {
		try {
			if (!giftCardId) {
				toastStore.error(tr('giftCards.transactions.error'));
				return;
			}
			// Convert YYYY-MM-DD to RFC3339 format (ISO 8601 with timezone)
			const transactionDateISO = `${transactionDate}T00:00:00Z`;

			// Automatisch debit (Ausgabe) verwenden
			await giftCardsApi.createTransaction(giftCardId, {
				type: 'debit',
				amount: transactionAmount,
				description: transactionDescription || undefined,
				transaction_date: transactionDateISO
			});
			toastStore.success(tr('giftCards.transactions.createSuccess'));
			showTransactionForm = false;
			transactionAmount = 0;
			transactionDescription = '';
			transactionDate = new Date().toISOString().split('T')[0];
			await Promise.all([loadGiftCard(), loadTransactions()]);
		} catch (err: any) {
			toastStore.error(err.message || tr('giftCards.transactions.createError'));
		}
	}

	function promptDeleteTransaction(transactionId: string) {
		transactionToDelete = transactionId;
		showDeleteTransactionModal = true;
	}

	async function confirmDeleteTransaction() {
		if (!transactionToDelete || !giftCardId) return;
		try {
			await giftCardsApi.deleteTransaction(giftCardId, transactionToDelete);
			toastStore.success(tr('giftCards.transactions.deleteSuccess'));
			showDeleteTransactionModal = false;
			await Promise.all([loadGiftCard(), loadTransactions()]);
		} catch (err: any) {
			toastStore.error(err.message || tr('giftCards.transactions.deleteError'));
		} finally {
			transactionToDelete = null;
			showDeleteTransactionModal = false;
		}
	}

	function getStatusBadge(status: string): { class: string; text: string } {
		switch (status) {
			case 'inactive':
				return {
					class: 'bg-gray-200 text-gray-700',
					text: tr('giftCards.status.inactive')
				};
			case 'expired':
				return {
					class: 'bg-red-200 text-red-700',
					text: tr('giftCards.status.expired')
				};
			case 'depleted':
				return {
					class: 'bg-orange-200 text-orange-700',
					text: tr('giftCards.status.depleted')
				};
			default:
				return { class: '', text: '' };
		}
	}
</script>

<svelte:head>
	<title
		>{giftCard
			? `${giftCard.merchant?.name || tr('giftCards.title')} - ${tr('common.appName')}`
			: `${tr('giftCards.title')} - ${tr('common.appName')}`}</title
	>
</svelte:head>

<div class="mb-6 flex items-center justify-between">
	<a href="/gift-cards" class="text-cyan-600 hover:text-cyan-700"
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
{:else if giftCard}
	<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
		<!-- Left column: Gift Card Details -->
		<div class="lg:col-span-2">
			{#if !isEditing}
				<!-- View Mode -->
				<div
					class="bg-white rounded-lg shadow-lg overflow-hidden"
					style="border-left: 6px solid {giftCard.merchant?.color || '#F59E0B'}"
				>
					<div
						class="p-6 {giftCard.status && giftCard.status !== 'active'
							? 'opacity-50 grayscale'
							: ''}"
					>
						<!-- Header -->
						<div class="mb-6">
							<ResourceHeader
								{isOffline}
								isFavorite={giftCard.is_favorite}
								{isTogglingFavorite}
								canEdit={giftCard.permissions?.can_edit}
								favoriteTitleAdd={tr('common.addFavorite')}
								favoriteTitleRemove={tr('common.removeFavorite')}
								ontoggleFavorite={toggleFavorite}
								onstartEdit={startEdit}
							>
								{#snippet children()}
									{@const gc = giftCard!}
									<div class="flex items-baseline gap-3 flex-wrap mb-2">
										<h1
											class="text-lg sm:text-xl md:text-2xl font-bold text-gray-900"
										>
											{#if gc.merchant}
												{gc.merchant.name}
											{:else}
												{tr('giftCards.title')}
											{/if}
										</h1>
										<span
											class="text-sm sm:text-base md:text-lg font-semibold"
											style="color: {gc.merchant?.color || '#F59E0B'}"
										>
											{formatCurrency(gc.current_balance, gc.currency, $locale)}
											<span class="text-sm text-gray-500 font-normal">
												({tr('giftCards.balance.remaining', {
													percent: Math.round(
														(gc.current_balance / gc.initial_balance) * 100
													)
												})})
											</span>
										</span>
										{#if gc.owner && gc.owner.id !== $authStore.user?.id}
											<span class="text-xs text-gray-400">
												{tr('giftCards.sharedBy', {
													name: gc.owner.first_name || gc.owner.email
												})}
											</span>
										{/if}
									</div>
								{/snippet}
							</ResourceHeader>

							<!-- Duplicate Warning -->
							{#if giftCard.duplicate_warning}
								<DuplicateWarningBanner
									warning={giftCard.duplicate_warning}
									resourceType="gift_card"
									onNavigate={(id) => goto(`/gift-cards/${id}`)}
								/>
							{/if}
						</div>

						<!-- Barcode Display -->
						<BarcodeDisplay
							value={giftCard.card_number}
							type={giftCard.barcode_type || 'CODE128'}
							status={giftCard.status}
							statusBadge={giftCard.status && giftCard.status !== 'active'
								? getStatusBadge(giftCard.status)
								: undefined}
							pin={giftCard.pin}
							balance={giftCard.current_balance.toFixed(2)}
							currency={giftCard.currency}
							expiresAt={giftCard.expires_at}
						/>

						<!-- Notes -->
						{#if giftCard.notes}
							<div
								class="mt-4 bg-yellow-50 border-l-4 border-yellow-400 p-3 rounded"
							>
								<p class="text-sm text-gray-700">{giftCard.notes}</p>
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
						<GiftCardForm
							bind:cardNumber={editCardNumber}
							bind:merchantId={editMerchantId}
							bind:initialBalance={editInitialBalance}
							bind:currency={editCurrency}
							bind:pin={editPin}
							bind:barcodeType={editBarcodeType}
							bind:expiresAt={editExpiresAt}
							bind:notes={editNotes}
							onSubmit={saveEdit}
							onCancel={cancelEdit}
							isLoading={false}
							submitLabel={tr('common.save')}
						/>
						{#if giftCard.permissions?.can_delete}
							<div class="pt-4 mt-4 border-t border-gray-200">
								<button
									type="button"
									onclick={promptDelete}
									disabled={isOffline}
									class="btn-text-danger w-full flex items-center justify-center gap-1.5 {isOffline
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
									{/if}
									{tr('giftCards.deleteButton')}
								</button>
							</div>
						{/if}
					</div>
				</div>
			{/if}
		</div>

		<!-- Right column: Balance, Transactions, Transfer & Sharing Info -->
		<div class="lg:col-span-1 space-y-4">
			<!-- Balance & Transactions Box -->
			<div class="bg-white rounded-lg shadow-lg p-6">
				<!-- Balance Display -->
				<div class="mb-6">
					<p class="text-sm text-gray-700 mb-1">
						{tr('giftCards.balance.current')}
					</p>
					<p
						class="text-3xl font-bold"
						style="color: {giftCard.merchant?.color || '#F59E0B'}"
					>
						{giftCard.current_balance.toFixed(2)}
						{giftCard.currency}
					</p>
					<!-- Progress Bar -->
					<div class="mt-3 bg-gray-200 rounded-full h-3">
						<div
							class="h-3 rounded-full transition-all {percentageRemaining > 50
								? 'bg-green-500'
								: percentageRemaining > 20
									? 'bg-orange-500'
									: 'bg-red-600'}"
							style="width: {percentageRemaining}%"
						></div>
					</div>
					<p class="text-xs text-gray-600 mt-2">
						{tr('giftCards.balance.initial')}: {giftCard.initial_balance.toFixed(
							2
						)}
						{giftCard.currency}
					</p>
				</div>

				<!-- Transactions Section -->
				<div class="border-t pt-6">
					<div class="flex justify-between items-center mb-4">
						<h3 class="text-sm font-medium text-gray-700">
							{tr('giftCards.transactions.title')}
						</h3>
						{#if giftCard.permissions?.can_edit_transactions && !showTransactionForm}
							<button
								onclick={() => (showTransactionForm = true)}
								disabled={isOffline}
								class="btn btn-xs btn-danger whitespace-nowrap flex items-center gap-1.5 {isOffline
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
								{tr('common.new')}
							</button>
						{/if}
					</div>

					{#if showTransactionForm}
						<div
							class="border border-red-200 bg-red-50 rounded-lg p-4 space-y-4 mb-4"
						>
							<div>
								<label
									for="transactionDate-input"
									class="block text-sm font-medium text-gray-700 mb-1"
								>
									{tr('giftCards.transactions.date')} *
								</label>
								<input
									id="transactionDate-input"
									type="date"
									required
									bind:value={transactionDate}
									max={new Date().toISOString().split('T')[0]}
									class="input w-full text-base bg-white"
								/>
							</div>

							<div>
								<label
									for="transactionAmount-input"
									class="block text-sm font-medium text-gray-700 mb-1"
								>
									{tr('giftCards.transactions.amount')} *
								</label>
								<input
									id="transactionAmount-input"
									type="number"
									step="0.01"
									min="0.01"
									required
									bind:value={transactionAmount}
									class="input bg-white"
									placeholder="10.00"
								/>
							</div>

							<div>
								<label
									for="transactionDescription-input"
									class="block text-sm font-medium text-gray-700 mb-1"
								>
									{tr('giftCards.transactions.description')}
								</label>
								<input
									id="transactionDescription-input"
									type="text"
									bind:value={transactionDescription}
									class="input bg-white"
									placeholder={tr(
										'giftCards.transactions.descriptionPlaceholder'
									)}
								/>
							</div>

							<div class="flex gap-2">
								<button
									onclick={handleAddTransaction}
									disabled={isOffline}
									class="btn btn-danger flex-1 {isOffline
										? 'opacity-50 cursor-not-allowed'
										: ''}"
								>
									{tr('common.save')}
								</button>
								<button
									onclick={() => (showTransactionForm = false)}
									class="btn btn-ghost"
								>
									{tr('common.cancel')}
								</button>
							</div>
						</div>
					{/if}

					{#if transactions.length > 0}
						<div class="space-y-2">
							{#each transactions as transaction}
								<div
									class="flex items-center justify-between p-3 bg-gray-50 rounded gap-3"
								>
									<div class="flex-1">
										<div class="font-medium text-red-600">
											-{transaction.amount.toFixed(2)}
											{giftCard.currency}
										</div>
										{#if transaction.description}
											<div class="text-sm text-gray-600">
												{transaction.description}
											</div>
										{/if}
									</div>
									<div class="flex items-center gap-3">
										<div class="text-xs text-gray-500">
											{new Date(
												transaction.transaction_date.split('T')[0]
											).toLocaleDateString(currentLocale)}
										</div>
										{#if giftCard.permissions?.can_edit_transactions}
											<button
												onclick={() => promptDeleteTransaction(transaction.id)}
												disabled={isOffline}
												class="btn-text-danger text-base flex items-center {isOffline
													? 'opacity-50 cursor-not-allowed'
													: ''}"
												title={tr('giftCards.transactions.deleteButton')}
											>
												{#if isOffline}
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
															d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"
														></path>
													</svg>
												{:else}
													×
												{/if}
											</button>
										{/if}
									</div>
								</div>
							{/each}
						</div>
					{:else}
						<div class="text-center py-8 bg-gray-50 rounded">
							{#if isOffline}
								<p class="text-amber-600 text-sm">
									{tr('giftCards.transactions.notCachedOffline')}
								</p>
							{:else}
								<p class="text-gray-500 text-sm">
									{tr('giftCards.transactions.noTransactions')}
								</p>
							{/if}
						</div>
					{/if}
				</div>
			</div>

			{#if giftCard.permissions?.is_owner}
				<!-- Transfer Box -->
				<TransferBox
					{isOffline}
					title={tr('giftCards.transfer.title')}
					openButtonLabel={`→ ${tr('giftCards.transfer.transferButton')}`}
					transferButtonLabel={tr('giftCards.transfer.transferButton')}
					warningTitle={tr('giftCards.transfer.warning')}
					warningDetails={tr('giftCards.transfer.warningDetails')}
					emailLabel={tr('giftCards.transfer.newOwnerEmail')}
					emailHint={tr('giftCards.sharing.userMustBeRegistered')}
					whatHappensLabel={tr('giftCards.transfer.whatHappens')}
					details={[
						tr('giftCards.transfer.detail1'),
						tr('giftCards.transfer.detail2'),
						tr('giftCards.transfer.detail3'),
						tr('giftCards.transfer.detail4')
					]}
					bind:email={transferEmail}
					ontransfer={promptTransfer}
				/>

				<!-- Sharing Box -->
				<div class="bg-white rounded-lg shadow-lg p-6">
					<div class="flex justify-between items-center mb-4">
						<h3 class="text-lg font-semibold text-gray-900">
							{tr('giftCards.sharing.title')}
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
								{tr('giftCards.sharing.addButton')}
							</button>
						{/if}
					</div>

					{#if showShareForm}
						<div
							class="border border-cyan-200 bg-cyan-50 rounded-lg p-4 space-y-4 mb-4"
						>
							<EmailAutocomplete
								bind:value={shareEmail}
								label={tr('giftCards.sharing.userEmail')}
								hint={tr('giftCards.sharing.userMustBeRegistered')}
								inputId="share-email-input"
								disabled={isOffline}
							/>

							<SharePermissions
								bind:canEdit
								bind:canDelete
								bind:canEditTransactions
								showEditTransactions={true}
								labelEdit={tr('giftCards.sharing.canEdit')}
								labelEditDesc={tr('giftCards.sharing.canEditDesc')}
								labelDelete={tr('giftCards.sharing.canDelete')}
								labelDeleteDesc={tr('giftCards.sharing.canDeleteDesc')}
								labelEditTransactions={tr(
									'giftCards.sharing.canManageTransactions'
								)}
								labelEditTransactionsDesc={tr(
									'giftCards.sharing.canManageTransactionsDesc'
								)}
							/>

							<div class="bg-white border border-cyan-200 rounded-lg p-3">
								<h4 class="font-medium text-cyan-900 text-sm mb-2">
									{tr('giftCards.sharing.whatIsShared')}
								</h4>
								<ul class="text-xs text-cyan-800 space-y-1">
									<li>{tr('giftCards.sharing.sharedItemCardNumber')}</li>
									<li>{tr('giftCards.sharing.sharedItemBalance')}</li>
									<li>{tr('giftCards.sharing.sharedItemDetails')}</li>
									<li>{tr('giftCards.sharing.sharedItemTransactions')}</li>
									<li>{tr('giftCards.sharing.sharedItemNotes')}</li>
								</ul>
							</div>

							<div class="flex gap-2">
								<button
									onclick={handleShare}
									disabled={isOffline}
									class="btn btn-primary flex-1 {isOffline
										? 'opacity-50 cursor-not-allowed'
										: ''}"
								>
									{tr('giftCards.sharing.shareNow')}
								</button>
								<button
									onclick={() => {
										showShareForm = false;
										shareEmail = '';
										canEdit = false;
										canDelete = false;
										canEditTransactions = false;
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
							{#each shares as share}
								<ShareListItem
									{share}
									isEditing={editingShareId === share.shared_with_user.id}
									{isOffline}
									showTransactionsBadge={true}
									onstartEdit={() => startEditShare(share)}
									onsave={() => saveShareEdit(share.shared_with_user.id)}
									oncancel={cancelEditShare}
									ondelete={() => promptDeleteShare(share.shared_with_user.id)}
								>
									{#snippet children()}
										<SharePermissions
											bind:canEdit={editShareCanEdit}
											bind:canDelete={editShareCanDelete}
											bind:canEditTransactions={editShareCanEditTransactions}
											showEditTransactions={true}
											labelEdit={tr('giftCards.sharing.canEdit')}
											labelEditDesc={tr('giftCards.sharing.canEditDesc')}
											labelDelete={tr('giftCards.sharing.canDelete')}
											labelDeleteDesc={tr('giftCards.sharing.canDeleteDesc')}
											labelEditTransactions={tr(
												'giftCards.sharing.canManageTransactions'
											)}
											labelEditTransactionsDesc={tr(
												'giftCards.sharing.canManageTransactionsDesc'
											)}
										/>
									{/snippet}
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
		title={tr('giftCards.deleteConfirm')}
		message={tr('giftCards.deleteConfirmMessage')}
		confirmText={tr('common.delete')}
		cancelText={tr('common.cancel')}
		variant="danger"
		onconfirm={confirmDelete}
		oncancel={() => (showDeleteModal = false)}
	/>

	<ConfirmModal
		isOpen={showDeleteShareModal}
		title={tr('giftCards.sharing.removeConfirm')}
		message={tr('giftCards.sharing.removeConfirmMessage')}
		confirmText={tr('common.remove')}
		cancelText={tr('common.cancel')}
		variant="danger"
		onconfirm={confirmDeleteShare}
		oncancel={() => (showDeleteShareModal = false)}
	/>

	<ConfirmModal
		isOpen={showTransferModal}
		title={tr('giftCards.transfer.confirmTitle')}
		message={tr('giftCards.transfer.confirmMessage')}
		confirmText={tr('giftCards.transfer.transferButton')}
		cancelText={tr('common.cancel')}
		variant="transfer"
		onconfirm={confirmTransfer}
		oncancel={() => (showTransferModal = false)}
	/>

	<ConfirmModal
		isOpen={showDeleteTransactionModal}
		title={tr('giftCards.transactions.deleteConfirm')}
		message={tr('giftCards.transactions.deleteConfirmMessage')}
		confirmText={tr('common.delete')}
		cancelText={tr('common.cancel')}
		variant="danger"
		onconfirm={confirmDeleteTransaction}
		oncancel={() => (showDeleteTransactionModal = false)}
	/>
{:else}
	<!-- Not loading and no gift card (e.g. after a transfer that removed
	     access, or a load that failed before redirect). Without this branch
	     the page rendered nothing → white screen (issue #121). -->
	<div class="flex flex-col items-center justify-center py-16 text-center">
		<p class="text-gray-600 mb-4">{tr('giftCards.notFound')}</p>
		<button type="button" onclick={() => goto('/gift-cards')} class="btn btn-primary">
			{tr('giftCards.backToList')}
		</button>
	</div>
{/if}
