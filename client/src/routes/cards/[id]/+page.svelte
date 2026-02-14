<script lang="ts">
	import { onMount, onDestroy, tick } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { get } from 'svelte/store';
	import { authStore } from '$lib/stores/auth';
	import { isOnline } from '$lib/stores/offline';
	import { t } from '$lib/stores/i18n';
	import { cardsApi, sharedUsersApi, merchantsApi, ApiError } from '$lib/api';
	import { offlineDB } from '$lib/stores/offline-db';
	import { toastStore } from '$lib/stores/toast';
	import BarcodeDisplay from '$lib/components/BarcodeDisplay.svelte';
	import DuplicateWarningBanner from '$lib/components/DuplicateWarningBanner.svelte';
	import ConfirmModal from '$lib/components/ConfirmModal.svelte';

	import type { CardDTO, ShareDTO, MerchantDTO, UserDTO } from '$lib/types/api';
	import { detectBarcodeType } from '$lib/utils/barcode';
	import { logger } from '$lib/utils/logger';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

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
	let showTransferForm = $state(false);
	let shareEmail = $state('');
	let canEdit = $state(false);
	let canDelete = $state(false);
	let transferEmail = $state('');
	let isEditing = $state(false);
	let merchants = $state<MerchantDTO[]>([]);

	// Autocomplete state (for share form)
	let suggestedUsers = $state<UserDTO[]>([]);
	let showSuggestions = $state(false);
	let searchTimeout: ReturnType<typeof setTimeout> | null = null;

	// Autocomplete state (for transfer form)
	let transferSuggestedUsers = $state<UserDTO[]>([]);
	let showTransferSuggestions = $state(false);
	let transferSearchTimeout: ReturnType<typeof setTimeout> | null = null;

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

	// Scanner state
	let scannerOpen = $state(false);

	// Modal state
	let showDeleteModal = $state(false);
	let showDeleteShareModal = $state(false);
	let showTransferModal = $state(false);
	let shareToDelete: string | null = null;

	const isOffline = $derived(!$isOnline);

	onMount(async () => {
		await Promise.all([loadCard(), loadMerchants()]);
	});

	// Cleanup debounced search timeouts on unmount (SVL-PERF-003)
	onDestroy(() => {
		if (searchTimeout) clearTimeout(searchTimeout);
		if (transferSearchTimeout) clearTimeout(transferSearchTimeout);
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
						goto('/cards');
						return;
					}
					if (!cached) {
						toastStore.error(tr('cards.loadError'));
						goto('/cards');
						return;
					}
					// Transient error with cached data available - show warning, don't redirect
					toastStore.warning(tr('common.offlineMode'));
				}
			} else if (!cached) {
				toastStore.error(tr('cards.loadError'));
				goto('/cards');
			}
		} catch (err) {
			toastStore.error(tr('cards.loadError'));
			goto('/cards');
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
		} catch (err: any) {
			pageLogger.error('Save error:', err);
			toastStore.error(err.message || tr('cards.updateError'));
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

	function promptDelete() {
		showDeleteModal = true;
	}

	async function confirmDelete() {
		try {
			await cardsApi.delete(cardId);
			toastStore.success(tr('cards.deleteSuccess'));
			// Force full page reload to refresh the list (SPA navigation caches data)
			window.location.href = '/cards';
		} catch (err) {
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
		} catch (err) {
			toastStore.error(tr('common.favoriteError'));
		} finally {
			isTogglingFavorite = false;
		}
	}

	async function handleShare() {
		try {
			const response = await cardsApi.createShare(cardId, {
				email: shareEmail,
				can_edit: canEdit,
				can_delete: canDelete
			});
			shares = response.shares || [];
			toastStore.success(tr('cards.sharing.shareSuccess'));
			shareEmail = '';
			canEdit = false;
			canDelete = false;
			showShareForm = false;
		} catch (err: any) {
			toastStore.error(err.message || tr('cards.sharing.shareError'));
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
		} catch (err: any) {
			toastStore.error(err.message || tr('cards.sharing.updateError'));
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
		} catch (err) {
			toastStore.error(tr('cards.sharing.removeError'));
		} finally {
			shareToDelete = null;
			showDeleteShareModal = false;
		}
	}

	// Autocomplete functions
	async function searchSharedUsers(query: string) {
		if (searchTimeout) clearTimeout(searchTimeout);

		if (query.length < 2) {
			suggestedUsers = [];
			showSuggestions = false;
			return;
		}

		searchTimeout = setTimeout(async () => {
			try {
				const response = await sharedUsersApi.search(query);
				suggestedUsers = response.users;
				showSuggestions = true;
			} catch (err) {
				pageLogger.error('Failed to search users:', err);
				suggestedUsers = [];
			}
		}, 300); // 300ms debounce
	}

	function selectUser(user: UserDTO) {
		shareEmail = user.email;
		showSuggestions = false;
		suggestedUsers = [];
	}

	function onEmailInput(event: Event) {
		const input = event.target as HTMLInputElement;
		shareEmail = input.value;
		searchSharedUsers(input.value);
	}

	function onEmailFocus() {
		if (shareEmail.length >= 2) {
			searchSharedUsers(shareEmail);
		}
	}

	function onEmailBlur() {
		// Delay to allow click on suggestion
		setTimeout(() => {
			showSuggestions = false;
		}, 200);
	}

	// Transfer autocomplete functions
	async function searchTransferUsers(query: string) {
		if (transferSearchTimeout) clearTimeout(transferSearchTimeout);

		if (query.length < 2) {
			transferSuggestedUsers = [];
			showTransferSuggestions = false;
			return;
		}

		transferSearchTimeout = setTimeout(async () => {
			try {
				const response = await sharedUsersApi.search(query);
				transferSuggestedUsers = response.users;
				showTransferSuggestions = true;
			} catch (err) {
				pageLogger.error('Failed to search users:', err);
				transferSuggestedUsers = [];
			}
		}, 300);
	}

	function selectTransferUser(user: UserDTO) {
		transferEmail = user.email;
		showTransferSuggestions = false;
		transferSuggestedUsers = [];
	}

	function onTransferEmailInput(event: Event) {
		const input = event.target as HTMLInputElement;
		transferEmail = input.value;
		searchTransferUsers(input.value);
	}

	function onTransferEmailFocus() {
		if (transferEmail.length >= 2) {
			searchTransferUsers(transferEmail);
		}
	}

	function onTransferEmailBlur() {
		setTimeout(() => {
			showTransferSuggestions = false;
		}, 200);
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
		} catch (err: any) {
			toastStore.error(err.message || tr('cards.transfer.error'));
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
	<a href="/cards" class="text-cyan-600 hover:text-cyan-700"
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
							<div class="flex items-start justify-between gap-4 mb-3">
								<div class="flex-1 min-w-0">
									<div class="flex items-baseline gap-2 flex-wrap mb-1">
										<h1
											class="text-lg sm:text-xl md:text-2xl font-bold text-gray-900"
										>
											{#if card.merchant}
												{card.merchant.name}
											{:else}
												{tr('common.card')}
											{/if}
											{#if card.program}
												<span
													class="text-sm sm:text-base md:text-lg font-normal text-gray-500 ml-1"
													>{card.program}</span
												>
											{/if}
										</h1>
										{#if card.owner && card.owner.id !== $authStore.user?.id}
											<span class="text-xs text-gray-400">
												{tr('cards.sharedBy', {
													name: card.owner.first_name || card.owner.email
												})}
											</span>
										{/if}
									</div>
								</div>
								<div class="flex gap-2 flex-shrink-0">
									<!-- Favorite Button -->
									<button
										data-testid="favorite-button"
										onclick={() => toggleFavorite()}
										disabled={isOffline || isTogglingFavorite}
										class="btn btn-xs {card.is_favorite
											? 'btn-favorite'
											: 'bg-gray-200 hover:bg-gray-300 text-gray-700'} {isOffline ||
										isTogglingFavorite
											? 'opacity-50 cursor-not-allowed'
											: ''}"
										title={card.is_favorite
											? tr('common.removeFromFavorites')
											: tr('common.addToFavorites')}
									>
										<span class="inline-block w-4 text-center leading-none"
											>{card.is_favorite ? '★' : '☆'}</span
										>
									</button>
									{#if card.permissions?.can_edit}
										<button
											onclick={() => startEdit()}
											disabled={isOffline}
											class="btn btn-xs btn-gray whitespace-nowrap flex items-center gap-1.5 {isOffline
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
											{tr('common.edit')}
										</button>
									{/if}
								</div>
							</div>

							<!-- Duplicate Warning -->
							{#if card.duplicate_warning}
								<DuplicateWarningBanner
									warning={card.duplicate_warning}
									resourceType="card"
									onNavigate={(id) => goto(`/cards/${id}`)}
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
						<form
							class="space-y-6"
							onsubmit={(e) => {
								e.preventDefault();
								saveEdit();
							}}
						>
							<div>
								<label
									for="merchant-select"
									class="block text-sm font-medium text-gray-700 mb-1"
								>
									{tr('cards.merchant')} *
								</label>
								<select
									id="merchant-select"
									bind:value={editMerchantId}
									required
									class="input"
								>
									<option value="" disabled
										>{tr('merchants.selectMerchant')}</option
									>
									{#each merchants as merchant}
										<option value={merchant.id}>{merchant.name}</option>
									{/each}
								</select>
							</div>

							<div>
								<label
									for="program"
									class="block text-sm font-medium text-gray-700 mb-1"
								>
									{tr('cards.program')} *
								</label>
								<input
									type="text"
									id="program"
									bind:value={editProgram}
									required
									class="input"
								/>
							</div>

							<div>
								<label
									for="card_number"
									class="block text-sm font-medium text-gray-700 mb-1"
								>
									{tr('cards.cardNumber')} *
								</label>
								<div class="flex gap-2">
									<input
										type="text"
										id="card_number"
										bind:value={editCardNumber}
										oninput={handleCardNumberInput}
										required
										class="input flex-1 font-mono"
									/>
									<button
										type="button"
										onclick={() => (scannerOpen = true)}
										class="btn btn-primary flex-shrink-0"
										title="Barcode mit Kamera scannen"
									>
										<svg
											class="w-5 h-5"
											fill="none"
											stroke="currentColor"
											viewBox="0 0 24 24"
										>
											<path
												stroke-linecap="round"
												stroke-linejoin="round"
												stroke-width="2"
												d="M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z"
											></path>
											<path
												stroke-linecap="round"
												stroke-linejoin="round"
												stroke-width="2"
												d="M15 13a3 3 0 11-6 0 3 3 0 016 0z"
											></path>
										</svg>
										<span class="hidden sm:inline">{tr('common.scan')}</span>
									</button>
								</div>
							</div>

							<!-- Scanner Modal (lazy loaded) -->
							{#if scannerOpen}
								{#await import('$lib/components/BarcodeScanner.svelte') then module}
									{@const BarcodeScanner = module.default}
									<BarcodeScanner bind:open={scannerOpen} onscan={handleScan} />
								{/await}
							{/if}

							<div>
								<label
									for="barcode_type"
									class="block text-sm font-medium text-gray-700 mb-1"
								>
									{tr('cards.barcodeType')}
								</label>
								<select
									id="barcode_type"
									bind:value={editBarcodeType}
									class="input"
								>
									<option value="CODE128">CODE128</option>
									<option value="CODE39">CODE39</option>
									<option value="CODE93">CODE93</option>
									<option value="CODABAR">CODABAR</option>
									<option value="QR">QR Code</option>
									<option value="EAN13">EAN-13</option>
									<option value="EAN8">EAN-8</option>
									<option value="UPCA">UPC-A</option>
									<option value="UPCE">UPC-E</option>
									<option value="ITF">ITF</option>
									<option value="ITF14">ITF-14</option>
									<option value="ISBN13">ISBN-13</option>
									<option value="PDF417">PDF417</option>
									<option value="DATAMATRIX">Data Matrix</option>
									<option value="AZTEC">Aztec</option>
									<option value="MAXICODE">MaxiCode</option>
								</select>
							</div>

							<div>
								<label
									for="status"
									class="block text-sm font-medium text-gray-700 mb-1"
								>
									{tr('cards.statusLabel')}
								</label>
								<select id="status" bind:value={editStatus} class="input">
									<option value="active">{tr('giftCards.status.active')}</option
									>
									<option value="inactive">{tr('cards.status.inactive')}</option
									>
									<option value="expired">{tr('cards.status.expired')}</option>
									<option value="lost">{tr('cards.status.lost')}</option>
									<option value="blocked">{tr('cards.status.blocked')}</option>
								</select>
							</div>

							<div>
								<label
									for="notes"
									class="block text-sm font-medium text-gray-700 mb-1"
								>
									{tr('cards.notes')}
								</label>
								<textarea
									id="notes"
									bind:value={editNotes}
									rows="3"
									class="input"
								></textarea>
							</div>

							<div class="flex gap-2">
								<button
									type="submit"
									disabled={isOffline}
									class="btn btn-primary flex-1"
								>
									{tr('common.save')}
								</button>
								<button
									type="button"
									onclick={cancelEdit}
									class="btn btn-ghost"
								>
									{tr('common.cancel')}
								</button>
							</div>

							{#if card.permissions?.can_delete}
								<div class="pt-4 border-t border-gray-200">
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
						</form>
					</div>
				</div>
			{/if}
		</div>

		<!-- Right column: Transfer & Sharing Info (only for owners) -->
		<div class="lg:col-span-1 space-y-4">
			{#if card.permissions?.is_owner}
				<!-- Transfer Box -->
				<div
					class="bg-white rounded-lg shadow-lg p-6 border-2 border-purple-200"
				>
					<div class="flex justify-between items-center mb-4">
						<h3 class="text-lg font-semibold text-purple-900">
							{tr('common.transferOwnership')}
						</h3>
						{#if !showTransferForm}
							<button
								onclick={() => (showTransferForm = true)}
								disabled={isOffline}
								class="btn btn-xs btn-purple whitespace-nowrap flex items-center gap-1.5 {isOffline
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
									{tr('cards.transfer.transferButton')}
								{:else}
									{tr('cards.transfer.button')}
								{/if}
							</button>
						{/if}
					</div>

					{#if showTransferForm}
						<div
							class="border border-purple-200 bg-purple-50 rounded-lg p-4 space-y-4"
						>
							<div
								class="bg-yellow-50 border border-yellow-200 rounded-lg p-3 mb-4"
							>
								<p class="text-sm font-medium text-yellow-800">
									<strong>{tr('cards.transfer.warning')}</strong>
								</p>
								<p class="text-xs text-yellow-700 mt-1">
									{tr('giftCards.transfer.warningDetails')}
								</p>
							</div>
							<div class="relative">
								<label
									for="transfer-email-input"
									class="block text-sm font-medium text-gray-700 mb-1"
								>
									{tr('cards.transfer.newOwnerEmail')} *
								</label>
								<input
									id="transfer-email-input"
									type="email"
									value={transferEmail}
									oninput={onTransferEmailInput}
									onfocus={onTransferEmailFocus}
									onblur={onTransferEmailBlur}
									required
									placeholder="benutzer@example.com"
									autocomplete="off"
									class="input bg-white"
								/>
								<p class="text-xs text-gray-500 mt-1">
									{tr('giftCards.sharing.userMustBeRegistered')}
								</p>

								{#if showTransferSuggestions && transferSuggestedUsers.length > 0}
									<div
										class="absolute z-10 w-full mt-1 bg-white border border-gray-300 rounded-md shadow-lg max-h-48 overflow-y-auto"
									>
										{#each transferSuggestedUsers as user}
											<button
												type="button"
												onclick={() => selectTransferUser(user)}
												class="w-full text-left px-3 py-2 hover:bg-gray-100 focus:bg-gray-100 focus:outline-none"
											>
												<div class="font-medium text-sm text-gray-900">
													{#if user.first_name && user.last_name}
														{user.first_name} {user.last_name}
													{:else if user.first_name}
														{user.first_name}
													{:else}
														{user.email}
													{/if}
												</div>
												<div class="text-xs text-gray-500">{user.email}</div>
											</button>
										{/each}
									</div>
								{/if}
							</div>
							<div>
								<p class="text-sm font-medium text-gray-700 mb-2">
									{tr('cards.transfer.whatHappens')}
								</p>
								<ul class="text-xs text-gray-600 space-y-1">
									<li>{tr('cards.transfer.newOwnerGetsRights')}</li>
									<li>{tr('cards.transfer.allSharesDeleted')}</li>
									<li>{tr('cards.transfer.youLoseAccess')}</li>
									<li>{tr('cards.transfer.transferLogged')}</li>
								</ul>
							</div>
							<div class="flex gap-2">
								<button
									onclick={promptTransfer}
									disabled={isOffline}
									class="btn btn-purple flex-1"
								>
									{tr('cards.transfer.transferButton')}
								</button>
								<button
									onclick={() => (showTransferForm = false)}
									class="btn btn-ghost"
								>
									{tr('common.cancel')}
								</button>
							</div>
						</div>
					{/if}
				</div>

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
							<div class="relative">
								<label
									for="share-email-input"
									class="block text-sm font-medium text-gray-700 mb-1"
								>
									{tr('cards.sharing.userEmail')} *
								</label>
								<input
									id="share-email-input"
									type="email"
									value={shareEmail}
									oninput={onEmailInput}
									onfocus={onEmailFocus}
									onblur={onEmailBlur}
									required
									placeholder="benutzer@example.com"
									autocomplete="off"
									class="input bg-white"
								/>

								{#if showSuggestions && suggestedUsers.length > 0}
									<div
										class="absolute z-10 w-full mt-1 bg-white border border-gray-300 rounded-md shadow-lg max-h-48 overflow-y-auto"
									>
										{#each suggestedUsers as user}
											<button
												type="button"
												onclick={() => selectUser(user)}
												class="w-full text-left px-3 py-2 hover:bg-gray-100 focus:bg-gray-100 focus:outline-none"
											>
												<div class="font-medium text-sm text-gray-900">
													{#if user.first_name && user.last_name}
														{user.first_name} {user.last_name}
													{:else if user.first_name}
														{user.first_name}
													{:else}
														{user.email}
													{/if}
												</div>
												<div class="text-xs text-gray-500">{user.email}</div>
											</button>
										{/each}
									</div>
								{/if}

								<p class="text-xs text-gray-500 mt-1">
									{tr('giftCards.sharing.userMustBeRegistered')}
								</p>
							</div>

							<div class="space-y-2">
								<label class="flex items-start">
									<input
										type="checkbox"
										bind:checked={canEdit}
										class="mt-0.5 h-4 w-4 text-cyan-600 focus:ring-cyan-500 border-gray-300 rounded"
									/>
									<div class="ml-2">
										<span class="block text-sm font-medium text-gray-900"
											>{tr('cards.sharing.canEdit')}</span
										>
										<span class="text-xs text-gray-500">
											{tr('cards.sharing.canEditDesc')}
										</span>
									</div>
								</label>
								<label class="flex items-start">
									<input
										type="checkbox"
										bind:checked={canDelete}
										class="mt-0.5 h-4 w-4 text-cyan-600 focus:ring-cyan-500 border-gray-300 rounded"
									/>
									<div class="ml-2">
										<span class="block text-sm font-medium text-gray-900"
											>{tr('cards.sharing.canDelete')}</span
										>
										<span class="text-xs text-gray-500"
											>{tr('cards.sharing.canDeleteDesc')}</span
										>
									</div>
								</label>
							</div>

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
										shareEmail = '';
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
							{#each shares as share}
								{#if editingShareId === share.shared_with_user.id}
									<!-- Edit Mode -->
									<div
										class="border border-cyan-200 bg-cyan-50 rounded-lg p-4 space-y-4 mb-4"
									>
										<div>
											<p class="font-medium text-gray-900 text-sm">
												{#if share.shared_with_user?.first_name && share.shared_with_user?.last_name}
													{share.shared_with_user.first_name}
													{share.shared_with_user.last_name}
												{:else if share.shared_with_user?.first_name}
													{share.shared_with_user.first_name}
												{:else}
													{share.shared_with_user?.email || 'Unknown User'}
												{/if}
											</p>
											<p class="text-xs text-gray-500">
												{share.shared_with_user?.email || ''}
											</p>
										</div>

										<div class="space-y-2">
											<label class="flex items-start">
												<input
													type="checkbox"
													bind:checked={editShareCanEdit}
													class="mt-0.5 h-4 w-4 text-cyan-600 focus:ring-cyan-500 border-gray-300 rounded"
												/>
												<div class="ml-2">
													<span class="block text-sm font-medium text-gray-900"
														>{tr('cards.sharing.canEdit')}</span
													>
													<span class="text-xs text-gray-500"
														>{tr('cards.sharing.canEditDesc')}</span
													>
												</div>
											</label>
											<label class="flex items-start">
												<input
													type="checkbox"
													bind:checked={editShareCanDelete}
													class="mt-0.5 h-4 w-4 text-cyan-600 focus:ring-cyan-500 border-gray-300 rounded"
												/>
												<div class="ml-2">
													<span class="block text-sm font-medium text-gray-900"
														>{tr('cards.sharing.canDelete')}</span
													>
													<span class="text-xs text-gray-500"
														>{tr('cards.sharing.canDeleteDesc')}</span
													>
												</div>
											</label>
										</div>

										<div class="flex gap-2">
											<button
												onclick={() => saveShareEdit(share.shared_with_user.id)}
												disabled={isOffline}
												class="btn btn-primary flex-1"
											>
												{tr('common.save')}
											</button>
											<button onclick={cancelEditShare} class="btn btn-ghost">
												{tr('common.cancel')}
											</button>
										</div>

										<div class="pt-2 border-t border-cyan-200">
											<button
												type="button"
												onclick={() =>
													promptDeleteShare(share.shared_with_user.id)}
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
												{tr('giftCards.sharing.removeButton')}
											</button>
										</div>
									</div>
								{:else}
									<!-- View Mode -->
									<div class="border border-gray-200 rounded-lg p-3">
										<div class="flex justify-between items-start mb-2">
											<div class="flex-1">
												<p class="font-medium text-gray-900 text-sm">
													{#if share.shared_with_user?.first_name && share.shared_with_user?.last_name}
														{share.shared_with_user.first_name}
														{share.shared_with_user.last_name}
													{:else if share.shared_with_user?.first_name}
														{share.shared_with_user.first_name}
													{:else}
														{share.shared_with_user?.email || 'Unknown User'}
													{/if}
												</p>
												<p class="text-xs text-gray-500">
													{share.shared_with_user?.email || ''}
												</p>
											</div>
											<button
												onclick={() => startEditShare(share)}
												disabled={isOffline}
												class="btn-text text-xs flex items-center gap-1"
											>
												{#if isOffline}
													<svg
														class="w-3 h-3"
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
													{tr('common.edit')}
												{/if}
											</button>
										</div>
										<div class="flex flex-wrap gap-1">
											{#if share.can_edit}
												<span
													class="text-xs bg-green-100 text-green-800 px-2 py-0.5 rounded"
												>
													{tr('giftCards.sharing.permEdit')}
												</span>
											{/if}
											{#if share.can_delete}
												<span
													class="text-xs bg-red-100 text-red-800 px-2 py-0.5 rounded"
												>
													{tr('giftCards.sharing.permDelete')}
												</span>
											{/if}
											{#if !share.can_edit && !share.can_delete}
												<span
													class="text-xs bg-gray-100 text-gray-600 px-2 py-0.5 rounded"
												>
													{tr('common.viewOnly')}
												</span>
											{/if}
										</div>
									</div>
								{/if}
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
