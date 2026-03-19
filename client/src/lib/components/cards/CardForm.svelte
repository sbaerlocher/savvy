<script lang="ts">
	import { merchantsApi } from '$lib/api';
	import { onMount } from 'svelte';
	import { toastStore } from '$lib/stores/toast';
	import { t } from '$lib/stores/i18n';

	import type { MerchantDTO } from '$lib/types/api';
	import { logger } from '$lib/utils/logger';
	import MerchantSelect from '$lib/components/MerchantSelect.svelte';

	const componentLogger = logger.child('CardForm');

	interface Props {
		cardNumber: string;
		merchantId?: string;
		program?: string;
		barcodeType?: string;
		notes?: string;
		status?: string;
		onSubmit: () => void;
		onCancel: () => void;
		isLoading: boolean;
		submitLabel?: string;
	}

	let {
		cardNumber = $bindable(''),
		merchantId = $bindable(''),
		program = $bindable(''),
		barcodeType = $bindable('CODE128'),
		notes = $bindable(''),
		status = $bindable(undefined),
		onSubmit,
		onCancel,
		isLoading,
		submitLabel
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
			componentLogger.error('Failed to load merchants:', err);
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
		<label for="merchant" class="label">{$t('cards.merchant')} *</label>
		<MerchantSelect
			{merchants}
			bind:value={merchantId}
			required
			id="merchant"
		/>
	</div>

	<div>
		<label for="program" class="label">{$t('cards.program')}</label>
		<input
			id="program"
			type="text"
			bind:value={program}
			placeholder={$t('cards.programPlaceholder')}
			class="input"
		/>
	</div>

	<div>
		<label for="cardNumber" class="label">{$t('cards.cardNumber')} *</label>
		<div class="flex gap-2">
			<input
				id="cardNumber"
				type="text"
				required
				bind:value={cardNumber}
				placeholder="1234567890"
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
						d="M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z"
					></path>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M15 13a3 3 0 11-6 0 3 3 0 016 0z"
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
		<label for="barcodeType" class="label">{$t('cards.barcodeType')}</label>
		<select
			id="barcodeType"
			bind:value={barcodeType}
			class="input"
			style="font-size: 16px;"
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

	{#if status !== undefined}
		<div>
			<label for="status" class="label">{$t('cards.statusLabel')}</label>
			<select
				id="status"
				bind:value={status}
				class="input"
				style="font-size: 16px;"
			>
				<option value="active">{$t('giftCards.status.active')}</option>
				<option value="inactive">{$t('cards.status.inactive')}</option>
				<option value="expired">{$t('cards.status.expired')}</option>
				<option value="lost">{$t('cards.status.lost')}</option>
				<option value="blocked">{$t('cards.status.blocked')}</option>
			</select>
		</div>
	{/if}

	<div>
		<label for="notes" class="label">{$t('cards.notes')}</label>
		<textarea id="notes" bind:value={notes} rows="3" class="input"></textarea>
	</div>

	<div class="flex gap-2">
		<button type="submit" class="flex-1 btn btn-primary" disabled={isLoading}>
			{isLoading ? $t('common.saving') : submitLabel || $t('common.save')}
		</button>
		<button type="button" onclick={onCancel} class="btn btn-ghost"
			>{$t('common.cancel')}</button
		>
	</div>
</form>
