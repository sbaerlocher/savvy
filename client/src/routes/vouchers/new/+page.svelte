<script lang="ts">
	import { goto } from '$app/navigation';
	import { get } from 'svelte/store';
	import { locale, t } from '$lib/stores/i18n';
	import { vouchersApi, merchantsApi, sharedUsersApi } from '$lib/api';
	import { toastStore } from '$lib/stores/toast';
	import { logger } from '$lib/utils/logger';
	import type { UserDTO, MerchantDTO } from '$lib/types/api';
	import { onMount } from 'svelte';

	import { detectBarcodeType } from '$lib/utils/barcode';

	// Svelte 5 compatible translation wrapper
	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);
	const pageLogger = logger.child('VoucherNewPage');

	let code = $state('');
	let merchantId = $state('');
	let type = $state('percentage');
	let value = $state(0);
	let currency = $state('CHF');
	let barcodeType = $state('CODE128');
	let usageLimitType = $state('single_use');
	let validFrom = $state('');
	let validUntil = $state('');
	let description = $state('');
	let isLoading = $state(false);

	// Validation errors
	let merchantError = $state('');
	let valueError = $state('');
	let validUntilError = $state('');

	// Merchants
	let merchants = $state<MerchantDTO[]>([]);

	// Sharing state
	let shareEmail = $state('');

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
			pageLogger.error('Merchants laden fehlgeschlagen', { error: err });
		}
	}

	function handleScan(event: { barcode: string; format?: string }) {
		code = event.barcode;
		// Use detected format from scanner if available, otherwise fallback to detectBarcodeType
		const detectedFormat = event.format || detectBarcodeType(event.barcode);
		pageLogger.info('Barcode scanned', {
			barcode: event.barcode,
			format: event.format,
			detectedFormat
		});
		barcodeType = detectedFormat;
		scannerOpen = false;
		toastStore.success(`${tr('common.scanSuccess')}: ${detectedFormat}`);
	}

	function handleCodeInput(event: Event) {
		const input = event.target as HTMLInputElement;
		code = input.value;
		if (code.trim()) {
			barcodeType = detectBarcodeType(code);
		}
	}

	function setExpiryOffset(days: number) {
		const date = new Date();
		date.setDate(date.getDate() + days);
		validUntil = date.toISOString().split('T')[0];
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
			pageLogger.error('Failed to fetch user suggestions', { error: err });
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

		// Reset errors
		merchantError = '';
		valueError = '';
		validUntilError = '';

		try {
			// Validate required fields
			let hasErrors = false;

			if (!validUntil) {
				validUntilError = tr('vouchers.validUntilRequired');
				hasErrors = true;
			}

			if (!merchantId) {
				merchantError = tr('vouchers.errors.merchantRequired');
				hasErrors = true;
			}

			if (value <= 0) {
				valueError = tr('vouchers.errors.valueRequired');
				hasErrors = true;
			}

			if (hasErrors) {
				isLoading = false;
				return;
			}

			// Use today for validFrom if not provided
			const today = new Date().toISOString().split('T')[0];

			const response = await vouchersApi.create({
				code,
				merchant_id: merchantId || undefined,
				type,
				value,
				currency: currency || undefined,
				barcode_type: barcodeType || undefined,
				usage_limit_type: usageLimitType || undefined,
				valid_from: validFrom ? `${validFrom}T00:00:00Z` : `${today}T00:00:00Z`,
				valid_until: `${validUntil}T23:59:59Z`,
				description: description || undefined,
				share_with_email: shareEmail || undefined
			});
			toastStore.success(tr('vouchers.createSuccess'));
			// Force full page reload to ensure fresh data in lists
			window.location.href = `/vouchers/${response.voucher.id}`;
		} catch (err: any) {
			toastStore.error(err.message || tr('vouchers.createError'));
		} finally {
			isLoading = false;
		}
	}
</script>

<svelte:head>
	<title>{tr('vouchers.newVoucher')} - {tr('common.appName')}</title>
</svelte:head>

<div class="mb-6">
	<a href="/vouchers" class="text-cyan-600 hover:text-cyan-700"
		>{tr('common.backToOverview')}</a
	>
</div>

<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
	<!-- Left column: Form (2/3 width) -->
	<div class="lg:col-span-2">
		<div class="bg-white rounded-lg shadow-lg p-6">
			<h1 class="text-3xl font-bold text-gray-900 mb-6">
				{tr('vouchers.newVoucher')}
			</h1>
			<form onsubmit={handleSubmit} class="space-y-6">
				<!-- Händler -->
				<div>
					<label
						for="merchant-select"
						class="block text-sm font-medium text-gray-700 mb-1"
					>
						{tr('vouchers.merchant')} *
					</label>
					<select
						id="merchant-select"
						bind:value={merchantId}
						oninput={() => (merchantError = '')}
						class="w-full px-4 py-2 bg-white border rounded-md {merchantError
							? 'border-red-500 focus:ring-red-500 focus:border-red-500'
							: 'border-gray-300 focus:ring-cyan-500 focus:border-cyan-500'}"
					>
						<option value="">{tr('vouchers.selectMerchant')}</option>
						{#each merchants as merchant}
							<option value={merchant.id}>{merchant.name}</option>
						{/each}
					</select>
					{#if merchantError}
						<p class="text-red-600 text-sm mt-1">{merchantError}</p>
					{:else}
						<p class="text-sm text-gray-500 mt-1">
							{tr('vouchers.merchantHint')}
						</p>
					{/if}
				</div>

				<!-- Code -->
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
							bind:value={code}
							oninput={handleCodeInput}
							required
							class="flex-1 px-4 py-2 bg-white border border-gray-300 rounded-md focus:ring-cyan-500 focus:border-cyan-500 font-mono"
							placeholder="SUMMER2024"
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

				<!-- Scanner Modal -->
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
					>
						{tr('vouchers.barcodeType')}
					</label>
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

				<!-- Beschreibung -->
				<div>
					<label
						for="description"
						class="block text-sm font-medium text-gray-700 mb-1"
					>
						{tr('vouchers.description')}
					</label>
					<textarea
						id="description"
						bind:value={description}
						rows="3"
						class="w-full px-4 py-2 bg-white border border-gray-300 rounded-md focus:ring-cyan-500 focus:border-cyan-500"
						placeholder={tr('vouchers.descriptionPlaceholder')}
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
							bind:value={type}
							required
							class="w-full px-4 py-2 bg-white border border-gray-300 rounded-md focus:ring-cyan-500 focus:border-cyan-500"
						>
							<option value="percentage">{tr('vouchers.typePercentage')}</option
							>
							<option value="fixed_amount"
								>{tr('vouchers.typeFixedAmount')}</option
							>
							<option value="points_multiplier"
								>{tr('vouchers.typePointsMultiplier')}</option
							>
						</select>
					</div>

					<div>
						<label
							for="value"
							class="block text-sm font-medium text-gray-700 mb-1"
						>
							{tr('vouchers.value')} *
						</label>
						<div
							class="flex gap-2 {$locale?.startsWith('en')
								? 'flex-row-reverse'
								: 'flex-row'}"
						>
							<input
								type="number"
								step="0.01"
								min="0"
								id="value"
								bind:value
								oninput={() => (valueError = '')}
								required
								class="flex-1 px-4 py-2 bg-white border rounded-md {valueError
									? 'border-red-500 focus:ring-red-500 focus:border-red-500'
									: 'border-gray-300 focus:ring-cyan-500 focus:border-cyan-500'}"
								placeholder="10.00"
							/>
							{#if type === 'fixed_amount'}
								<select
									id="currency"
									bind:value={currency}
									required
									class="w-28 px-2 py-2 bg-white border border-gray-300 rounded-md focus:ring-cyan-500 focus:border-cyan-500"
								>
									<option value="CHF">CHF</option>
									<option value="EUR">EUR</option>
									<option value="USD">USD</option>
									<option value="GBP">GBP</option>
								</select>
							{/if}
						</div>
						{#if valueError}
							<p class="text-red-600 text-sm mt-1">{valueError}</p>
						{:else}
							<p class="text-sm text-gray-500 mt-1">
								{type === 'percentage'
									? tr('vouchers.valueHintPercentage')
									: type === 'points_multiplier'
										? tr('vouchers.valueHintMultiplier')
										: tr('vouchers.valueHintAmount')}
							</p>
						{/if}
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
						bind:value={usageLimitType}
						class="w-full px-4 py-2 bg-white border border-gray-300 rounded-md focus:ring-cyan-500 focus:border-cyan-500"
					>
						<option value="single_use"
							>{tr('vouchers.usageLimitTypes.single_use')}</option
						>
						<option value="one_per_customer"
							>{tr('vouchers.usageLimitTypes.one_per_customer')}</option
						>
						<option value="multiple_use_with_card"
							>{tr('vouchers.usageLimitTypes.multiple_use_with_card')}</option
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

				<!-- Gültig von / Gültig bis -->
				<div class="space-y-4 md:grid md:grid-cols-2 md:gap-4 md:space-y-0">
					<div>
						<label
							for="valid_from"
							class="block text-sm font-medium text-gray-700 mb-1"
						>
							{tr('vouchers.validFrom')}
						</label>
						<input
							type="date"
							id="valid_from"
							bind:value={validFrom}
							class="w-full px-3 py-2 bg-white border border-gray-300 rounded-md focus:ring-cyan-500 focus:border-cyan-500 text-base"
						/>
						<p class="text-xs text-gray-500 mt-1 hidden sm:block">
							{tr('vouchers.validFromHint')}
						</p>
					</div>

					<div>
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
									onclick={() => setExpiryOffset(30)}
									class="px-2 py-0.5 text-xs bg-gray-100 hover:bg-gray-200 text-gray-600 hover:text-gray-800 rounded transition-colors"
								>
									{tr('vouchers.quickSelect.oneMonth')}
								</button>
								<button
									type="button"
									onclick={() => setExpiryOffset(90)}
									class="px-2 py-0.5 text-xs bg-gray-100 hover:bg-gray-200 text-gray-600 hover:text-gray-800 rounded transition-colors"
								>
									{tr('vouchers.quickSelect.threeMonths')}
								</button>
							</div>
						</div>

						<input
							type="date"
							id="valid_until"
							bind:value={validUntil}
							oninput={() => (validUntilError = '')}
							required
							class="w-full px-3 py-2 bg-white border rounded-md text-base {validUntilError
								? 'border-red-500 focus:ring-red-500 focus:border-red-500'
								: 'border-gray-300 focus:ring-cyan-500 focus:border-cyan-500'}"
						/>
						{#if validUntilError}
							<p class="text-red-600 text-sm mt-1">{validUntilError}</p>
						{:else}
							<p class="text-xs text-gray-500 mt-1 hidden sm:block">
								{tr('vouchers.validUntilHint')}
							</p>
						{/if}
					</div>
				</div>

				<!-- Buttons -->
				<div class="flex gap-3 pt-4">
					<button
						type="submit"
						disabled={isLoading}
						class="btn btn-sm btn-primary flex-1"
					>
						{isLoading ? tr('common.creating') : tr('vouchers.createButton')}
					</button>
					<a href="/vouchers" class="btn btn-sm btn-ghost">
						{tr('common.cancel')}
					</a>
				</div>
			</form>
		</div>
	</div>

	<!-- Right column: Sharing (1/3 width) -->
	<div class="lg:col-span-1 space-y-4">
		<!-- Sharing Box -->
		<div class="bg-white rounded-lg shadow-lg p-6">
			<h2 class="text-xl font-bold text-gray-900 mb-4">{tr('common.share')}</h2>
			<p class="text-sm text-gray-600 mb-4">
				{tr('vouchers.sharing.shareOnCreate')}
			</p>

			<div class="border border-cyan-200 bg-cyan-50 rounded-lg p-4 space-y-4">
				<!-- Email Input with Autocomplete -->
				<div class="relative">
					<label
						for="share_email"
						class="block text-sm font-medium text-gray-700 mb-1"
					>
						{tr('vouchers.sharing.userEmail')} *
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
						{tr('vouchers.sharing.hint')}
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

				<!-- Info Box -->
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
			</div>
		</div>
	</div>
</div>
