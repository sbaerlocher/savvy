<script lang="ts">
	import { goto } from '$app/navigation';
	import { get } from 'svelte/store';
	import { giftCardsApi, sharedUsersApi } from '$lib/api';
	import { toastStore } from '$lib/stores/toast';
	import { t } from '$lib/stores/i18n';
	import type { UserDTO, DuplicateWarning } from '$lib/types/api';
	import { extractDuplicate } from '$lib/utils/api-errors';

	import { logger } from '$lib/utils/logger';
	import GiftCardForm from '$lib/components/gift-cards/GiftCardForm.svelte';
	import SharedInfoBox from '$lib/components/SharedInfoBox.svelte';
	import DuplicateWarningBanner from '$lib/components/DuplicateWarningBanner.svelte';

	// Svelte 5 compatible translation wrapper
	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);
	const pageLogger = logger.child('GiftCardsNewPage');

	let cardNumber = $state('');
	let merchantId = $state('');
	let initialBalance = $state(0);
	let currency = $state('EUR');
	let pin = $state('');
	let barcodeType = $state('CODE128');
	let expiresAt = $state('');
	let notes = $state('');
	let isLoading = $state(false);
	let duplicateWarning = $state<DuplicateWarning | null>(null);

	// Sharing state
	let shareEmail = $state('');
	let canEdit = $state(false);
	let canDelete = $state(false);
	let canEditTransactions = $state(false);

	// Email autocomplete state
	let suggestions = $state<UserDTO[]>([]);
	let showSuggestions = $state(false);
	let selectedIndex = $state(-1);

	async function fetchSuggestions() {
		if (shareEmail.length < 2) {
			suggestions = [];
			showSuggestions = false;
			return;
		}

		try {
			const response = await sharedUsersApi.search(shareEmail);
			suggestions = response.users || [];
			showSuggestions = suggestions.length > 0;
			selectedIndex = -1;
		} catch (err) {
			pageLogger.error('Failed to fetch user suggestions:', err);
			suggestions = [];
			showSuggestions = false;
		}
	}

	function selectSuggestion(user: UserDTO) {
		shareEmail = user.email;
		showSuggestions = false;
		selectedIndex = -1;
	}

	function handleKeydown(event: KeyboardEvent) {
		if (!showSuggestions) return;

		if (event.key === 'ArrowDown') {
			event.preventDefault();
			selectedIndex = Math.min(selectedIndex + 1, suggestions.length - 1);
		} else if (event.key === 'ArrowUp') {
			event.preventDefault();
			selectedIndex = Math.max(selectedIndex - 1, -1);
		} else if (event.key === 'Enter' && selectedIndex >= 0) {
			event.preventDefault();
			selectSuggestion(suggestions[selectedIndex]);
		} else if (event.key === 'Escape') {
			showSuggestions = false;
			selectedIndex = -1;
		}
	}

	function hideSuggestions() {
		setTimeout(() => {
			showSuggestions = false;
		}, 200);
	}

	async function handleSubmit() {
		isLoading = true;
		duplicateWarning = null;
		try {
			const response = await giftCardsApi.create({
				card_number: cardNumber,
				merchant_id: merchantId || undefined,
				initial_balance: initialBalance,
				currency,
				pin: pin || undefined,
				barcode_type: barcodeType || undefined,
				expires_at: expiresAt ? `${expiresAt}T00:00:00Z` : undefined,
				notes: notes || undefined,
				share_with_email: shareEmail || undefined,
				share_can_edit: shareEmail ? canEdit : undefined,
				share_can_delete: shareEmail ? canDelete : undefined,
				share_can_edit_transactions: shareEmail
					? canEditTransactions
					: undefined
			});
			toastStore.success(tr('giftCards.createSuccess'));
			// Force full page reload to ensure fresh data in lists
			window.location.href = `/gift-cards/${response.gift_card.id}`;
		} catch (err: any) {
			const duplicate = extractDuplicate(err);
			if (duplicate) {
				duplicateWarning = duplicate;
			} else {
				toastStore.error(err.message || tr('giftCards.createError'));
			}
		} finally {
			isLoading = false;
		}
	}

	function handleCancel() {
		goto('/gift-cards');
	}
</script>

<svelte:head>
	<title>{tr('giftCards.newGiftCard')} - {tr('common.appName')}</title>
</svelte:head>

<div class="mb-6">
	<a href="/gift-cards" class="text-cyan-600 hover:text-cyan-700"
		>{tr('common.backToOverview')}</a
	>
</div>

<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
	<!-- Left column: Form (2/3 width) -->
	<div class="lg:col-span-2">
		<div class="bg-white rounded-lg shadow-lg p-6">
			<h1 class="text-3xl font-bold text-gray-900 mb-6">
				{tr('giftCards.newGiftCard')}
			</h1>
			<DuplicateWarningBanner
				warning={duplicateWarning}
				resourceType="gift_card"
				onNavigate={(id) => goto(`/gift-cards/${id}`)}
			/>
			<GiftCardForm
				bind:cardNumber
				bind:merchantId
				bind:initialBalance
				bind:currency
				bind:pin
				bind:barcodeType
				bind:expiresAt
				bind:notes
				onSubmit={handleSubmit}
				onCancel={handleCancel}
				{isLoading}
				submitLabel={tr('giftCards.createButton')}
			/>
		</div>
	</div>

	<!-- Right column: Sharing (1/3 width) -->
	<div class="lg:col-span-1 space-y-4">
		<!-- Sharing Box -->
		<div class="bg-white rounded-lg shadow-lg p-6">
			<h2 class="text-xl font-bold text-gray-900 mb-4">
				{tr('giftCards.sharing.title')}
			</h2>
			<p class="text-sm text-gray-600 mb-4">
				{tr('giftCards.sharing.shareOnCreate')}
			</p>

			<div class="border border-cyan-200 bg-cyan-50 rounded-lg p-4 space-y-4">
				<!-- Email Input with Autocomplete -->
				<div class="relative">
					<label
						for="share_email"
						class="block text-sm font-medium text-gray-700 mb-1"
					>
						{tr('giftCards.sharing.userEmail')} *
					</label>
					<input
						type="email"
						id="share_email"
						bind:value={shareEmail}
						oninput={() => fetchSuggestions()}
						onkeydown={handleKeydown}
						onblur={hideSuggestions}
						class="w-full px-3 py-2 bg-white border border-gray-300 rounded-md focus:ring-cyan-500 focus:border-cyan-500"
						placeholder="benutzer@example.com"
						autocomplete="off"
					/>
					<p class="text-xs text-gray-500 mt-1">
						{tr('forms.userMustBeRegistered')}
					</p>

					<!-- Autocomplete Dropdown -->
					{#if showSuggestions}
						<div
							class="absolute z-10 w-full mt-1 bg-white border border-gray-300 rounded-lg shadow-lg max-h-60 overflow-y-auto"
						>
							{#each suggestions as suggestion, index}
								<button
									type="button"
									onclick={() => selectSuggestion(suggestion)}
									class="w-full text-left px-4 py-2 hover:bg-cyan-50 border-b border-gray-100 last:border-b-0 {index ===
									selectedIndex
										? 'bg-cyan-50'
										: ''}"
								>
									<div class="flex flex-col">
										<span class="text-sm font-medium text-gray-900"
											>{suggestion.first_name}
											{suggestion.last_name}</span
										>
										<span class="text-xs text-gray-500">{suggestion.email}</span
										>
									</div>
								</button>
							{/each}
						</div>
					{/if}
				</div>

				<!-- Permissions -->
				<div class="space-y-2">
					<label class="flex items-start">
						<input
							type="checkbox"
							bind:checked={canEdit}
							class="mt-0.5 h-4 w-4 text-cyan-600 focus:ring-cyan-500 border-gray-300 rounded"
						/>
						<div class="ml-2">
							<span class="block text-sm font-medium text-gray-900"
								>{tr('giftCards.sharing.canEdit')}</span
							>
							<span class="text-xs text-gray-500">
								{tr('giftCards.sharing.canEditDesc')}
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
								>{tr('giftCards.sharing.canDelete')}</span
							>
							<span class="text-xs text-gray-500"
								>{tr('giftCards.sharing.canDeleteDesc')}</span
							>
						</div>
					</label>
					<label class="flex items-start">
						<input
							type="checkbox"
							bind:checked={canEditTransactions}
							class="mt-0.5 h-4 w-4 text-cyan-600 focus:ring-cyan-500 border-gray-300 rounded"
						/>
						<div class="ml-2">
							<span class="block text-sm font-medium text-gray-900"
								>{tr('giftCards.sharing.canManageTransactions')}</span
							>
							<span class="text-xs text-gray-500">
								{tr('giftCards.sharing.canManageTransactionsDesc')}
							</span>
						</div>
					</label>
				</div>

				<!-- Info Box -->
				<SharedInfoBox
					title={tr('giftCards.sharing.whatIsShared')}
					items={[
						tr('giftCards.sharing.sharedItemCardNumber'),
						tr('giftCards.sharing.sharedItemBalance'),
						tr('giftCards.sharing.sharedItemDetails'),
						tr('giftCards.sharing.sharedItemTransactions'),
						tr('giftCards.sharing.sharedItemNotes')
					]}
				/>
			</div>
		</div>
	</div>
</div>
