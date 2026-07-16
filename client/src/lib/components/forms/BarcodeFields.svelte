<script lang="ts">
	import { ICON_CAMERA, ICON_CAMERA_LENS } from '$lib/icons';
	import { toastStore } from '$lib/stores/toast';
	import { t } from '$lib/stores/i18n';
	import { checkSymbologySuitability } from '$lib/utils/barcode';

	interface Props {
		value: string;
		barcodeType?: string;
		label: string;
		typeLabel: string;
		inputId?: string;
		placeholder?: string;
	}

	let {
		value = $bindable(''),
		barcodeType = $bindable('CODE128'),
		label,
		typeLabel,
		inputId = 'cardNumber',
		placeholder = '1234567890'
	}: Props = $props();

	let scanning = $state(false);

	// Warn when the entered content cannot be encoded by the chosen symbology.
	const symbologyWarning = $derived(
		checkSymbologySuitability(value, barcodeType)
	);

	function handleScan(event: { barcode: string; format?: string }) {
		value = event.barcode;
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

<div>
	<label for={inputId} class="label">{label} *</label>
	<div class="flex gap-2">
		<input
			id={inputId}
			type="text"
			required
			bind:value
			{placeholder}
			class="input flex-1"
		/>
		<button
			type="button"
			onclick={() => (scanning = true)}
			class="btn btn-primary"
			title={$t('common.scanBarcode')}
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
					d={ICON_CAMERA}
				></path>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d={ICON_CAMERA_LENS}
				></path>
			</svg>
			<span class="hidden sm:inline">{$t('common.scan')}</span>
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
	<label for="barcodeType" class="label">{typeLabel}</label>
	<select id="barcodeType" bind:value={barcodeType} class="input">
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
	{#if symbologyWarning}
		<p class="mt-1 text-sm text-warning-600 dark:text-warning-400" role="alert">
			{$t(symbologyWarning)}
		</p>
	{/if}
</div>
