<script lang="ts">
	import { merchantsApi } from '$lib/api';
	import { onMount } from 'svelte';
	import { toastStore } from '$lib/stores/toast';
	import { t } from '$lib/stores/i18n';

	import type { MerchantDTO } from '$lib/types/api';
	import { logger } from '$lib/utils/logger';

	const componentLogger = logger.child('GiftCardForm');

	interface Props {
		cardNumber: string;
		merchantId?: string;
		initialBalance?: number;
		currency?: string;
		barcodeType?: string;
		notes?: string;
		onSubmit: () => void;
		onCancel: () => void;
		isLoading: boolean;
		submitLabel?: string;
	}

	let {
		cardNumber = $bindable(''),
		merchantId = $bindable(''),
		initialBalance = $bindable(0),
		currency = $bindable('EUR'),
		barcodeType = $bindable('CODE128'),
		notes = $bindable(''),
		onSubmit,
		onCancel,
		isLoading,
		submitLabel = 'Speichern'
	}: Props = $props();

	let merchants = $state<MerchantDTO[]>([]);
	let scanning = $state(false);

	onMount(async () => {
		await loadMerchants();
	});

	async function loadMerchants() {
		try {
			const response = await merchantsApi.list();
			merchants = response.merchants;
		} catch (err) {
			componentLogger.error('Merchants laden fehlgeschlagen:', err);
		}
	}

	function handleScan(event: { barcode: string; format?: string }) {
		cardNumber = event.barcode;
		// Automatically set barcode type from scan result
		if (event.format) {
			barcodeType = event.format;
		}
		toastStore.success($t('common.scanSuccess') + ': ' + (event.format || ''));
	}

	function handleScanError(event: { message: string }) {
		toastStore.error(event.message);
	}
</script>

<form
	onsubmit={(e) => {
		e.preventDefault();
		onSubmit();
	}}
	class="space-y-4"
>
	<div>
		<label for="merchant" class="label">Händler</label>
		<select
			id="merchant"
			bind:value={merchantId}
			class="input"
			style="font-size: 16px; line-height: 2.5;"
		>
			<option value="">-- Händler auswählen --</option>
			{#each merchants as merchant}
				<option value={merchant.id}>{merchant.name}</option>
			{/each}
		</select>
		<p class="text-sm text-gray-500 mt-1">
			Optional: Wählen Sie einen Händler aus
		</p>
	</div>

	<div>
		<label for="cardNumber" class="label">Kartennummer *</label>
		<div class="flex gap-2">
			<input
				id="cardNumber"
				type="text"
				required
				bind:value={cardNumber}
				placeholder="1234567890123"
				class="input flex-1"
			/>
			<button
				type="button"
				onclick={() => (scanning = true)}
				class="btn btn-primary"
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
				<span class="hidden sm:inline">Scannen</span>
			</button>
		</div>
	</div>

	<!-- Barcode Scanner (lazy loaded) -->
	{#if scanning}
		{#await import('$lib/components/BarcodeScanner.svelte') then module}
			{@const BarcodeScanner = module.default}
			<BarcodeScanner
				bind:open={scanning}
				onscan={handleScan}
				onerror={handleScanError}
			/>
		{/await}
	{/if}

	<div>
		<label for="initialBalance" class="label">Anfangsguthaben *</label>
		<input
			id="initialBalance"
			type="number"
			step="0.01"
			required
			bind:value={initialBalance}
			placeholder="50.00"
			class="input"
		/>
		<p class="text-sm text-gray-500 mt-1">
			Der ursprüngliche Wert der Geschenkkarte
		</p>
	</div>

	<div>
		<label for="currency" class="label">Währung *</label>
		<select
			id="currency"
			bind:value={currency}
			required
			class="input"
			style="font-size: 16px; line-height: 2.5;"
		>
			<option value="EUR">EUR (€)</option>
			<option value="USD">USD ($)</option>
			<option value="CHF">CHF (Fr.)</option>
			<option value="GBP">GBP (£)</option>
		</select>
	</div>

	<div>
		<label for="barcodeType" class="label">Barcode-Typ</label>
		<select
			id="barcodeType"
			bind:value={barcodeType}
			class="input"
			style="font-size: 16px; line-height: 2.5;"
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
			<option value="ISBN10">ISBN-10</option>
			<option value="ISSN">ISSN</option>
			<option value="PDF417">PDF417</option>
			<option value="DATAMATRIX">Data Matrix</option>
			<option value="AZTEC">Aztec</option>
			<option value="MAXICODE">MaxiCode</option>
		</select>
	</div>

	<div>
		<label for="notes" class="label">Notizen</label>
		<textarea
			id="notes"
			bind:value={notes}
			rows="3"
			class="input"
			placeholder="Weitere Details zur Geschenkkarte..."
		></textarea>
	</div>

	<div class="flex gap-2">
		<button type="submit" class="btn btn-primary" disabled={isLoading}>
			{isLoading ? 'Speichere...' : submitLabel}
		</button>
		<button type="button" onclick={onCancel} class="btn btn-secondary"
			>Abbrechen</button
		>
	</div>
</form>
