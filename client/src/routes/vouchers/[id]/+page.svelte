<script lang="ts">
	import { onMount, onDestroy, tick } from 'svelte';
	import { page } from '$app/stores';
	import { goto, invalidateAll } from '$app/navigation';
	import { get } from 'svelte/store';
	import { authStore } from '$lib/stores/auth';
	import { isOnline } from '$lib/stores/offline';
	import { locale, t } from '$lib/stores/i18n';
	import {
		vouchersApi,
		sharedUsersApi,
		merchantsApi,
		ApiError
	} from '$lib/api';
	import { offlineDB } from '$lib/stores/offline-db';
	import { toastStore } from '$lib/stores/toast';
	import { formatCurrency } from '$lib/utils/currency';
	import { logger } from '$lib/utils/logger';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import BarcodeDisplay from '$lib/components/BarcodeDisplay.svelte';
	import DuplicateWarningBanner from '$lib/components/DuplicateWarningBanner.svelte';

	import ConfirmModal from '$lib/components/ConfirmModal.svelte';
	import type {
		VoucherDTO,
		ShareDTO,
		MerchantDTO,
		UserDTO
	} from '$lib/types/api';
	import { detectBarcodeType } from '$lib/utils/barcode';

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
	let showTransferForm = $state(false);
	let shareEmail = $state('');
	let transferEmail = $state('');
	let isEditing = $state(false);
	let merchants = $state<MerchantDTO[]>([]);

	// Edit form fields
	let editMerchantId = $state('');
	let editCode = $state('');
	let editType = $state('percentage');
	let editValue = $state(0);
	let editCurrency = $state('CHF');
	let editBarcodeType = $state('CODE128');
	let editUsageLimitType = $state('single_use');
	let editValidFrom = $state('');
	let editValidUntil = $state('');

	// Share managing state (for vouchers, only delete is possible)
	let managingShareId = $state<string | null>(null);
	let editStatus = $state('active');
	let editDescription = $state('');

	// Scanner state
	let scannerOpen = $state(false);

	// Modal state
	let showDeleteModal = $state(false);
	let showDeleteShareModal = $state(false);
	let showTransferModal = $state(false);
	let shareToDelete: string | null = null;

	// Autocomplete state (for share form)
	let suggestedUsers = $state<UserDTO[]>([]);
	let showSuggestions = $state(false);
	let searchTimeout: ReturnType<typeof setTimeout> | null = null;

	// Autocomplete state (for transfer form)
	let transferSuggestedUsers = $state<UserDTO[]>([]);
	let showTransferSuggestions = $state(false);
	let transferSearchTimeout: ReturnType<typeof setTimeout> | null = null;

	const isOffline = $derived(!$isOnline);

	onMount(async () => {
		await Promise.all([loadVoucher(), loadMerchants()]);
	});

	// Cleanup debounced search timeouts on unmount (SVL-PERF-003)
	onDestroy(() => {
		if (searchTimeout) clearTimeout(searchTimeout);
		if (transferSearchTimeout) clearTimeout(transferSearchTimeout);
	});

	async function loadVoucher() {
		isLoading = true;
		try {
			if (!voucherId) {
				toastStore.error(tr('vouchers.loadError'));
				goto('/vouchers');
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
						goto('/vouchers');
						return;
					}
					if (!cached) {
						toastStore.error(tr('vouchers.loadError'));
						goto('/vouchers');
						return;
					}
					// Transient error with cached data available - show warning, don't redirect
					toastStore.warning(tr('common.offlineMode'));
				}
			} else if (!cached) {
				toastStore.error(tr('vouchers.loadError'));
				goto('/vouchers');
			}
		} catch (err) {
			toastStore.error(tr('vouchers.loadError'));
			goto('/vouchers');
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

	function setEditExpiryOffset(days: number) {
		const date = new Date();
		date.setDate(date.getDate() + days);
		editValidUntil = date.toISOString().split('T')[0];
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
		} catch (err: any) {
			toastStore.error(err.message || tr('vouchers.updateError'));
		}
	}

	function handleScan(event: { barcode: string; format?: string }) {
		editCode = event.barcode;
		// Use detected format from scanner if available, otherwise fallback to detectBarcodeType
		editBarcodeType = event.format || detectBarcodeType(event.barcode);
		scannerOpen = false;
		toastStore.success(tr('common.scanSuccess'));
	}

	function handleCodeInput(event: Event) {
		const input = event.target as HTMLInputElement;
		editCode = input.value;
		if (editCode.trim()) {
			editBarcodeType = detectBarcodeType(editCode);
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
		} catch (err) {
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
		} catch (err) {
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
			const response = await vouchersApi.createShare(voucherId, {
				email: shareEmail
			});
			shares = response.shares || [];
			toastStore.success(tr('vouchers.sharing.shareSuccess'));
			shareEmail = '';
			showShareForm = false;
		} catch (err: any) {
			toastStore.error(err.message || tr('vouchers.sharing.shareError'));
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
		} catch (err) {
			toastStore.error(tr('vouchers.sharing.removeError'));
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
		} catch (err: any) {
			toastStore.error(err.message || tr('vouchers.transfer.error'));
		}
	}

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
				pageLogger.error('Failed to search users', { error: err });
				suggestedUsers = [];
			}
		}, 300);
	}

	function selectUser(user: UserDTO) {
		shareEmail = user.email;
		showSuggestions = false;
		suggestedUsers = [];
	}

	function onEmailInput(e: Event) {
		const input = e.target as HTMLInputElement;
		shareEmail = input.value;
		searchSharedUsers(input.value);
	}

	function onEmailFocus(e: Event) {
		const input = e.target as HTMLInputElement;
		if (input.value.length >= 2) {
			searchSharedUsers(input.value);
		}
	}

	function onEmailBlur() {
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
				pageLogger.error('Failed to search users', { error: err });
				transferSuggestedUsers = [];
			}
		}, 300);
	}

	function selectTransferUser(user: UserDTO) {
		transferEmail = user.email;
		showTransferSuggestions = false;
		transferSuggestedUsers = [];
	}

	function onTransferEmailInput(e: Event) {
		const input = e.target as HTMLInputElement;
		transferEmail = input.value;
		searchTransferUsers(input.value);
	}

	function onTransferEmailFocus(e: Event) {
		const input = e.target as HTMLInputElement;
		if (input.value.length >= 2) {
			searchTransferUsers(input.value);
		}
	}

	function onTransferEmailBlur() {
		setTimeout(() => {
			showTransferSuggestions = false;
		}, 200);
	}

	function getStatusBadge(status: string): { class: string; text: string } {
		switch (status) {
			case 'inactive':
				return {
					class: 'bg-gray-200 text-gray-700',
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
			return `${value}x Punkte`;
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

<div class="mb-6 flex items-center justify-between">
	<a href="/vouchers" class="text-cyan-600 hover:text-cyan-700"
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
{:else if voucher}
	<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
		<!-- Left column: Voucher Details -->
		<div class="lg:col-span-2">
			{#if !isEditing}
				<!-- View Mode -->
				<div
					class="bg-white rounded-lg shadow-lg overflow-hidden"
					style="border-left: 6px solid {voucher.merchant?.color || '#10B981'}"
				>
					<div
						class="p-6 {voucher.status && voucher.status !== 'valid'
							? 'opacity-50 grayscale'
							: ''}"
					>
						<!-- Header -->
						<div class="mb-6">
							<div class="flex items-start justify-between gap-4 mb-3">
								<div class="flex-1 min-w-0">
									<div class="flex items-baseline gap-3 flex-wrap mb-2">
										<h1
											class="text-lg sm:text-xl md:text-2xl font-bold text-gray-900"
										>
											{#if voucher.merchant}
												{voucher.merchant.name}
											{:else}
												{tr('vouchers.title')}
											{/if}
										</h1>
										<span class="text-sm text-gray-600">
											{voucher.usage_limit_type === 'single_use'
												? tr('vouchers.singleUseOnly')
												: tr('vouchers.multipleUse')}
										</span>
										{#if voucher.owner && voucher.owner.id !== $authStore.user?.id}
											<span class="text-xs text-gray-400">
												{tr('vouchers.sharedBy', {
													name: voucher.owner.first_name || voucher.owner.email
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
										class="btn btn-xs {voucher.is_favorite
											? 'btn-favorite'
											: 'bg-gray-200 hover:bg-gray-300 text-gray-700'} {isOffline ||
										isTogglingFavorite
											? 'opacity-50 cursor-not-allowed'
											: ''}"
										title={voucher.is_favorite
											? tr('common.removeFromFavorites')
											: tr('common.addToFavorites')}
									>
										<span class="inline-block w-4 text-center leading-none"
											>{voucher.is_favorite ? '★' : '☆'}</span
										>
									</button>
									{#if voucher.permissions?.can_edit}
										<button
											onclick={() => startEdit()}
											disabled={isOffline}
											class="btn btn-xs btn-gray whitespace-nowrap flex items-center gap-1.5 {isOffline
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
											{tr('common.edit')}
										</button>
									{/if}
								</div>
							</div>

							<!-- Duplicate Warning -->
							{#if voucher.duplicate_warning}
								<DuplicateWarningBanner
									warning={voucher.duplicate_warning}
									resourceType="voucher"
									onNavigate={(id) => goto(`/vouchers/${id}`)}
								/>
							{/if}
						</div>

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
							description={voucher.description}
						/>
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
									{tr('vouchers.merchant')}
								</label>
								<select
									id="merchant-select"
									bind:value={editMerchantId}
									class="input"
								>
									<option value="">{tr('vouchers.merchantSelect')}</option>
									{#each merchants as merchant}
										<option value={merchant.id}>{merchant.name}</option>
									{/each}
								</select>
							</div>

							<div>
								<label
									for="code"
									class="block text-sm font-medium text-gray-700 mb-1"
								>
									{tr('vouchers.code')} *
								</label>
								<div class="flex gap-2">
									<input
										type="text"
										id="code"
										bind:value={editCode}
										oninput={handleCodeInput}
										required
										class="input font-mono flex-1"
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

							<!-- Barcode Type -->
							<div>
								<label
									for="barcode_type_edit"
									class="block text-sm font-medium text-gray-700 mb-1"
								>
									{tr('vouchers.barcodeType')} *
								</label>
								<select
									id="barcode_type_edit"
									bind:value={editBarcodeType}
									class="input"
									required
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
									for="description"
									class="block text-sm font-medium text-gray-700 mb-1"
								>
									{tr('vouchers.description')}
								</label>
								<textarea
									id="description"
									bind:value={editDescription}
									rows="3"
									class="input"
								></textarea>
							</div>

							<!-- Typ / Wert -->
							<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
								<div>
									<label
										for="type"
										class="block text-sm font-medium text-gray-700 mb-1"
									>
										{tr('vouchers.type')} *
									</label>
									<select
										id="type"
										bind:value={editType}
										required
										class="input"
									>
										<option value="percentage"
											>{tr('vouchers.types.percentage')}</option
										>
										<option value="fixed_amount"
											>{tr('vouchers.types.fixedAmount')}</option
										>
										<option value="points_multiplier"
											>{tr('vouchers.types.pointsMultiplier')}</option
										>
									</select>
								</div>

								<div>
									<label
										for="value"
										class="block text-sm font-medium text-gray-700 mb-1"
									>
										Wert *
									</label>
									<div
										class="flex gap-2 {$locale?.startsWith('en')
											? 'flex-row-reverse'
											: 'flex-row'}"
									>
										<input
											type="number"
											step="0.01"
											id="value"
											bind:value={editValue}
											required
											class="input flex-1"
										/>
										{#if editType === 'fixed_amount'}
											<select
												id="currency"
												bind:value={editCurrency}
												required
												class="input w-28"
											>
												<option value="CHF">CHF</option>
												<option value="EUR">EUR</option>
												<option value="USD">USD</option>
												<option value="GBP">GBP</option>
											</select>
										{/if}
									</div>
									<p class="text-sm text-gray-500 mt-1">
										{editType === 'percentage'
											? tr('vouchers.types.percentageHint')
											: editType === 'points_multiplier'
												? tr('vouchers.types.pointsMultiplierHint')
												: tr('vouchers.types.fixedAmountHint')}
									</p>
								</div>
							</div>

							<!-- Gültig von / Gültig bis -->
							<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
								<div class="min-w-0">
									<label
										for="valid_from"
										class="block text-sm font-medium text-gray-700 mb-1"
									>
										{tr('vouchers.validFrom')}
									</label>
									<input
										type="date"
										id="valid_from"
										bind:value={editValidFrom}
										class="input w-full max-w-full"
										style="min-width: 0;"
									/>
									<p class="text-xs text-gray-500 mt-1 hidden sm:block">
										{tr('vouchers.validFromHint')}
									</p>
								</div>

								<div class="min-w-0">
									<div class="flex items-center justify-between mb-1">
										<label
											for="valid_until"
											class="text-sm font-medium text-gray-700"
										>
											{tr('vouchers.validUntil')} *
										</label>
										<!-- Quick-Select Buttons (inline with label) -->
										<div class="flex gap-1.5">
											<button
												type="button"
												onclick={() => setEditExpiryOffset(30)}
												class="px-2 py-0.5 text-xs bg-gray-100 hover:bg-gray-200 text-gray-600 hover:text-gray-800 rounded transition-colors"
											>
												{tr('vouchers.quickSelect.oneMonth')}
											</button>
											<button
												type="button"
												onclick={() => setEditExpiryOffset(90)}
												class="px-2 py-0.5 text-xs bg-gray-100 hover:bg-gray-200 text-gray-600 hover:text-gray-800 rounded transition-colors"
											>
												{tr('vouchers.quickSelect.threeMonths')}
											</button>
										</div>
									</div>

									<input
										type="date"
										id="valid_until"
										bind:value={editValidUntil}
										required
										class="input w-full max-w-full"
										style="min-width: 0;"
									/>
									<p class="text-xs text-gray-500 mt-1 hidden sm:block">
										{tr('vouchers.validUntilHint')}
									</p>
								</div>
							</div>

							<!-- Verwendungsart -->
							<div>
								<label
									for="usage_limit_type"
									class="block text-sm font-medium text-gray-700 mb-1"
								>
									{tr('vouchers.usageLimitType')}
								</label>
								<select
									id="usage_limit_type"
									bind:value={editUsageLimitType}
									class="input"
								>
									<option value="single_use"
										>{tr('vouchers.usageLimitTypes.single_use')}</option
									>
									<option value="one_per_customer"
										>{tr('vouchers.usageLimitTypes.one_per_customer')}</option
									>
									<option value="multiple_use_with_card"
										>{tr(
											'vouchers.usageLimitTypes.multiple_use_with_card'
										)}</option
									>
									<option value="multiple_use_without_card"
										>{tr(
											'vouchers.usageLimitTypes.multiple_use_without_card'
										)}</option
									>
								</select>
								<p class="text-sm text-gray-500 mt-1">
									{tr('vouchers.usageLimitTypeHint')}
								</p>
							</div>

							<div class="flex gap-2">
								<button
									type="submit"
									disabled={isOffline}
									class="btn btn-primary flex-1 {isOffline
										? 'opacity-50 cursor-not-allowed'
										: ''}"
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

							{#if voucher.permissions?.can_delete}
								<div class="pt-4 border-t border-gray-200">
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
										{tr('vouchers.deleteButton')}
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
			{#if voucher.permissions?.is_owner}
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
									{tr('vouchers.transfer.transferButton')}
								{:else}
									{tr('vouchers.transfer.button')}
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
									<strong>{tr('vouchers.transfer.warning')}</strong>
								</p>
								<p class="text-xs text-yellow-700 mt-1">
									{tr('vouchers.transfer.warningDetail')}
								</p>
							</div>
							<div class="relative">
								<label
									for="transfer-email-input"
									class="block text-sm font-medium text-gray-700 mb-1"
								>
									{tr('vouchers.transfer.newOwner')} *
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
									{tr('vouchers.transfer.whatHappens')}
								</p>
								<ul class="text-xs text-gray-600 space-y-1">
									<li>{tr('vouchers.transfer.newOwnerGetsRights')}</li>
									<li>{tr('vouchers.transfer.allSharesDeleted')}</li>
									<li>{tr('vouchers.transfer.youLoseAccess')}</li>
									<li>{tr('vouchers.transfer.transferLogged')}</li>
								</ul>
							</div>
							<div class="flex gap-2">
								<button
									onclick={promptTransfer}
									disabled={isOffline}
									class="btn btn-purple flex-1 {isOffline
										? 'opacity-50 cursor-not-allowed'
										: ''}"
								>
									{tr('vouchers.transfer.transferNow')}
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
							class="border border-cyan-200 bg-cyan-50 rounded-lg p-4 space-y-4 mb-4"
						>
							<div class="relative">
								<label
									for="share-email-input"
									class="block text-sm font-medium text-gray-700 mb-1"
								>
									{tr('vouchers.sharing.userEmail')} *
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
								<p class="text-xs text-gray-500 mt-1">
									{tr('vouchers.sharing.hint')}
								</p>

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
							</div>

							<div class="bg-white border border-cyan-200 rounded-lg p-3">
								<h4 class="font-medium text-cyan-900 text-sm mb-2">
									{tr('vouchers.sharing.whatIsShared')}
								</h4>
								<ul class="text-xs text-cyan-800 space-y-1">
									<li>{tr('vouchers.sharing.sharedCode')}</li>
									<li>{tr('vouchers.sharing.sharedDetails')}</li>
									<li>{tr('vouchers.sharing.sharedDescription')}</li>
								</ul>
								<p class="text-xs text-cyan-700 mt-2 italic">
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
										shareEmail = '';
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
								{#if managingShareId === share.shared_with_user.id}
									<!-- Manage Mode (Delete only for vouchers) -->
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

										<div class="flex gap-2">
											<button
												onclick={cancelManageShare}
												class="btn btn-ghost flex-1"
											>
												{tr('common.cancel')}
											</button>
										</div>

										<div class="pt-2 border-t border-cyan-200">
											<button
												type="button"
												onclick={() =>
													promptDeleteShare(share.shared_with_user.id)}
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
												{tr('vouchers.sharing.removeShare')}
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
												onclick={() =>
													startManageShare(share.shared_with_user.id)}
												disabled={isOffline}
												class="btn-text text-xs flex items-center gap-1 {isOffline
													? 'opacity-50 cursor-not-allowed'
													: ''}"
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
													{tr('common.manage')}
												{/if}
											</button>
										</div>
										<div class="flex flex-wrap gap-1">
											<span
												class="text-xs bg-gray-100 text-gray-600 px-2 py-0.5 rounded"
											>
												{tr('common.viewOnly')}
											</span>
										</div>
									</div>
								{/if}
							{/each}
						</div>
					{:else}
						<p class="text-sm text-gray-500 text-center py-4">
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
		isOpen={showTransferModal}
		title={tr('vouchers.transfer.confirmTitle')}
		message={tr('vouchers.transfer.confirmMessage')}
		confirmText={tr('vouchers.transfer.transferButton')}
		cancelText={tr('common.cancel')}
		variant="transfer"
		onconfirm={confirmTransfer}
	/>
{/if}
