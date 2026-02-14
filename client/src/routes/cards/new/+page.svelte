<script lang="ts">
	import { goto } from '$app/navigation';
	import { get } from 'svelte/store';
	import { t } from '$lib/stores/i18n';
	import { cardsApi, merchantsApi, sharedUsersApi } from '$lib/api';
	import { toastStore } from '$lib/stores/toast';
	import type { MerchantDTO, UserDTO } from '$lib/types/api';
	import { onMount } from 'svelte';

	import { detectBarcodeType } from '$lib/utils/barcode';
	import { logger } from '$lib/utils/logger';

	// Svelte 5 compatible translation wrapper
	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);
	const pageLogger = logger.child('CardsNewPage');

	let cardNumber = $state('');
	let merchantId = $state('');
	let program = $state('');
	let barcodeType = $state('CODE128');
	let notes = $state('');
	let isLoading = $state(false);

	// Merchants
	let merchants = $state<MerchantDTO[]>([]);

	// Sharing state
	let shareEmail = $state('');
	let canEdit = $state(false);
	let canDelete = $state(false);

	// Email autocomplete state
	let suggestions = $state<UserDTO[]>([]);
	let showSuggestions = $state(false);
	let selectedIndex = $state(-1);

	// Scanner state
	let scannerOpen = $state(false);

	onMount(async () => {
		await loadMerchants();
	});

	async function loadMerchants() {
		try {
			const response = await merchantsApi.list();
			merchants = response.merchants;
		} catch (err) {
			pageLogger.error('Failed to load merchants:', err);
		}
	}

	function handleScan(event: { barcode: string; format?: string }) {
		cardNumber = event.barcode;
		// Use detected format from scanner if available, otherwise fallback to detectBarcodeType
		barcodeType = event.format || detectBarcodeType(event.barcode);
		scannerOpen = false;
		toastStore.success(tr('common.scanSuccess'));
	}

	function handleCardNumberInput(event: Event) {
		const input = event.target as HTMLInputElement;
		cardNumber = input.value;
		if (cardNumber.trim()) {
			barcodeType = detectBarcodeType(cardNumber);
		}
	}

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

	async function handleSubmit(e: Event) {
		e.preventDefault();
		isLoading = true;
		try {
			const response = await cardsApi.create({
				card_number: cardNumber,
				merchant_id: merchantId || undefined,
				program: program || undefined,
				barcode_type: barcodeType || undefined,
				notes: notes || undefined,
				share_with_email: shareEmail || undefined,
				share_can_edit: shareEmail ? canEdit : undefined,
				share_can_delete: shareEmail ? canDelete : undefined
			});
			toastStore.success(tr('cards.createSuccess'));
			// Force full page reload to ensure fresh data in lists
			window.location.href = `/cards/${response.card.id}`;
		} catch (err: any) {
			toastStore.error(err.message || tr('cards.createError'));
		} finally {
			isLoading = false;
		}
	}
</script>

<svelte:head>
	<title>{tr('cards.newCard')} - {tr('common.appName')}</title>
</svelte:head>

<div class="mb-6">
	<a href="/cards" class="text-cyan-600 hover:text-cyan-700"
		>{tr('common.backToOverview')}</a
	>
</div>

<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
	<!-- Left column: Form (2/3 width) -->
	<div class="lg:col-span-2">
		<div class="bg-white rounded-lg shadow-lg p-6">
			<h1 class="text-3xl font-bold text-gray-900 mb-6">
				{tr('cards.newCard')}
			</h1>
			<form onsubmit={handleSubmit} class="space-y-6">
				<!-- Händler -->
				<div>
					<label
						for="merchant-select"
						class="block text-sm font-medium text-gray-700 mb-1"
						>{tr('cards.merchant')} *</label
					>
					<select
						id="merchant-select"
						bind:value={merchantId}
						required
						class="w-full px-4 py-2 bg-white border border-gray-300 rounded-md focus:ring-cyan-500 focus:border-cyan-500"
					>
						<option value="">{tr('merchants.selectMerchant')}</option>
						{#each merchants as merchant}
							<option value={merchant.id}>{merchant.name}</option>
						{/each}
					</select>
					<p class="text-sm text-gray-500 mt-1">
						{tr('cards.merchantSelectHint')}
					</p>
				</div>

				<!-- Programm -->
				<div>
					<label
						for="program"
						class="block text-sm font-medium text-gray-700 mb-1"
						>{tr('cards.program')} *</label
					>
					<input
						type="text"
						id="program"
						bind:value={program}
						required
						class="w-full px-4 py-2 bg-white border border-gray-300 rounded-md focus:ring-cyan-500 focus:border-cyan-500"
						placeholder={tr('cards.programPlaceholder')}
					/>
				</div>

				<!-- Kartennummer -->
				<div>
					<label
						for="card_number"
						class="block text-sm font-medium text-gray-700 mb-1"
						>{tr('cards.cardNumber')} *</label
					>
					<div class="flex gap-2">
						<input
							type="text"
							id="card_number"
							bind:value={cardNumber}
							oninput={handleCardNumberInput}
							required
							class="flex-1 px-4 py-2 bg-white border border-gray-300 rounded-md focus:ring-cyan-500 focus:border-cyan-500 font-mono"
							placeholder="1234567890123"
						/>
						<button
							type="button"
							onclick={() => (scannerOpen = true)}
							class="btn btn-primary flex-shrink-0"
							title={tr('common.scanBarcode')}
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

				<!-- Barcode-Typ -->
				<div>
					<label
						for="barcode_type"
						class="block text-sm font-medium text-gray-700 mb-1"
						>{tr('cards.barcodeType')}</label
					>
					<select
						id="barcode_type"
						bind:value={barcodeType}
						class="w-full px-4 py-2 bg-white border border-gray-300 rounded-md focus:ring-cyan-500 focus:border-cyan-500"
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

				<!-- Notizen -->
				<div>
					<label
						for="notes"
						class="block text-sm font-medium text-gray-700 mb-1"
						>{tr('cards.notes')}</label
					>
					<textarea
						id="notes"
						bind:value={notes}
						rows="3"
						class="w-full px-4 py-2 bg-white border border-gray-300 rounded-md focus:ring-cyan-500 focus:border-cyan-500"
						placeholder={tr('cards.notesPlaceholder')}
					></textarea>
				</div>

				<!-- Buttons -->
				<div class="flex gap-3 pt-4">
					<button
						type="submit"
						disabled={isLoading}
						class="btn btn-sm btn-primary flex-1"
					>
						{isLoading ? tr('common.creating') : tr('cards.createButton')}
					</button>
					<a href="/cards" class="btn btn-sm btn-ghost">
						{tr('common.cancel')}
					</a>
				</div>
			</form>
		</div>
	</div>

	<!-- Right column: Sharing (1/3 width) -->
	<div class="lg:col-span-1">
		<div class="bg-white rounded-lg shadow-lg p-6">
			<h2 class="text-xl font-bold text-gray-900 mb-4">
				{tr('cards.sharing.title')}
			</h2>
			<p class="text-sm text-gray-600 mb-4">
				{tr('cards.sharing.shareOnCreate')}
			</p>

			<div class="border border-cyan-200 bg-cyan-50 rounded-lg p-4 space-y-4">
				<!-- Email Input with Autocomplete -->
				<div class="relative">
					<label
						for="share_email"
						class="block text-sm font-medium text-gray-700 mb-1"
					>
						{tr('cards.sharing.userEmail')} *
					</label>
					<input
						type="email"
						id="share_email"
						bind:value={shareEmail}
						oninput={() => fetchSuggestions()}
						onkeydown={handleKeydown}
						onblur={hideSuggestions}
						required
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

				<!-- Info Box -->
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
			</div>
		</div>
	</div>
</div>
