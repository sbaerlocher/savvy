<script lang="ts">
	import { merchantsApi } from '$lib/api';
	import { onMount } from 'svelte';
	import { toastStore } from '$lib/stores/toast';

	import type { MerchantDTO } from '$lib/types/api';
	import { logger } from '$lib/utils/logger';
	import { locale, t } from '$lib/stores/i18n';
	import MerchantSelect from '$lib/components/MerchantSelect.svelte';

	const componentLogger = logger.child('VoucherForm');

	interface Props {
		code: string;
		merchantId?: string;
		type?: string;
		value?: number;
		currency?: string;
		minPurchaseAmount?: number;
		barcodeType?: string;
		validFrom?: string;
		validUntil?: string;
		usageLimitType?: string;
		description?: string;
		errors?: { merchant?: string; value?: string; validUntil?: string };
		onSubmit: () => void;
		onCancel: () => void;
		isLoading: boolean;
		submitLabel?: string;
	}

	let {
		code = $bindable(''),
		merchantId = $bindable(''),
		type = $bindable('percentage'),
		value = $bindable(0),
		currency = $bindable('CHF'),
		minPurchaseAmount = $bindable(0),
		barcodeType = $bindable('CODE128'),
		validFrom = $bindable(''),
		validUntil = $bindable(''),
		usageLimitType = $bindable('single_use'),
		description = $bindable(''),
		errors = $bindable({}),
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
			componentLogger.error('Merchants laden fehlgeschlagen:', err);
		}
	}

	function handleScan(event: { barcode: string; format?: string }) {
		code = event.barcode;
		// Automatically set barcode type from scan result
		if (event.format) {
			barcodeType = event.format;
		}
		toastStore.success($t('common.scanSuccess') + ': ' + (event.format || ''));
	}

	function handleScanError(event: { message: string }) {
		toastStore.error(event.message);
	}

	function setExpiryOffset(days: number) {
		const date = new Date();
		date.setDate(date.getDate() + days);
		validUntil = date.toISOString().split('T')[0];
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
		<label for="merchant" class="label">{$t('vouchers.merchant')} *</label>
		<MerchantSelect
			{merchants}
			bind:value={merchantId}
			id="merchant"
			onchange={() => {
				if (errors) errors = { ...errors, merchant: undefined };
			}}
		/>
		{#if errors?.merchant}
			<p class="text-red-600 text-sm mt-1">{errors.merchant}</p>
		{/if}
	</div>

	<div>
		<label for="code" class="label">{$t('vouchers.code')} *</label>
		<div class="flex gap-2">
			<input
				id="code"
				type="text"
				required
				bind:value={code}
				placeholder={$t('vouchers.codePlaceholder')}
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
		<label for="barcodeType" class="label">{$t('vouchers.barcodeType')}</label>
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

	<div>
		<label for="description" class="label">{$t('vouchers.description')}</label>
		<textarea
			id="description"
			bind:value={description}
			rows="3"
			class="input"
			placeholder={$t('vouchers.descriptionPlaceholder')}
		></textarea>
	</div>

	<!-- Typ / Wert -->
	<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
		<div>
			<label for="type" class="label">{$t('vouchers.type')} *</label>
			<select
				id="type"
				bind:value={type}
				required
				class="input"
				style="font-size: 16px;"
			>
				<option value="percentage">{$t('vouchers.typePercentage')}</option>
				<option value="fixed_amount">{$t('vouchers.typeFixedAmount')}</option>
				<option value="points_multiplier"
					>{$t('vouchers.typePointsMultiplier')}</option
				>
				<option value="bonus_points">{$t('vouchers.typeBonusPoints')}</option>
			</select>
		</div>

		<div>
			<label for="value" class="label">{$t('vouchers.value')} *</label>
			<div
				class="flex gap-2 {$locale?.startsWith('en')
					? 'flex-row-reverse'
					: 'flex-row'}"
			>
				<input
					id="value"
					type="number"
					step="0.01"
					min="0"
					required
					bind:value
					oninput={() => {
						if (errors) errors = { ...errors, value: undefined };
					}}
					class="flex-1 input {errors?.value
						? 'border-red-500 focus:ring-red-500 focus:border-red-500'
						: ''}"
					placeholder="10.00"
				/>
				{#if type === 'fixed_amount'}
					<select
						id="currency"
						bind:value={currency}
						required
						class="w-28 input"
						style="font-size: 16px;"
					>
						<option value="CHF">CHF</option>
						<option value="EUR">EUR</option>
						<option value="USD">USD</option>
						<option value="GBP">GBP</option>
					</select>
				{/if}
			</div>
			{#if errors?.value}
				<p class="text-red-600 text-sm mt-1">{errors.value}</p>
			{:else}
				<p class="text-sm text-gray-500 mt-1">
					{type === 'percentage'
						? $t('vouchers.valueHintPercentage')
						: type === 'points_multiplier'
							? $t('vouchers.valueHintMultiplier')
							: type === 'bonus_points'
								? $t('vouchers.valueHintBonusPoints')
								: $t('vouchers.valueHintAmount')}
				</p>
			{/if}
		</div>
	</div>

	<!-- Mindesteinkauf / Verwendungsart -->
	<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
		<div>
			<label for="minPurchaseAmount" class="label"
				>{$t('vouchers.minPurchaseAmount')}</label
			>
			<div class="flex gap-2">
				<input
					id="minPurchaseAmount"
					type="number"
					step="0.01"
					min="0"
					bind:value={minPurchaseAmount}
					class="flex-1 input"
					placeholder="0.00"
				/>
				<select
					id="minPurchaseCurrency"
					bind:value={currency}
					class="w-28 input"
					style="font-size: 16px;"
				>
					<option value="CHF">CHF</option>
					<option value="EUR">EUR</option>
					<option value="USD">USD</option>
					<option value="GBP">GBP</option>
				</select>
			</div>
			<p class="text-sm text-gray-500 mt-1">
				{$t('vouchers.minPurchaseAmountHint')}
			</p>
		</div>

		<div>
			<label for="usageLimitType" class="label"
				>{$t('vouchers.usageLimitType')}</label
			>
			<select
				id="usageLimitType"
				bind:value={usageLimitType}
				class="input"
				style="font-size: 16px;"
			>
				<option value="single_use"
					>{$t('vouchers.usageLimitTypes.single_use')}</option
				>
				<option value="one_per_customer"
					>{$t('vouchers.usageLimitTypes.one_per_customer')}</option
				>
				<option value="multiple_use_with_card"
					>{$t('vouchers.usageLimitTypes.multiple_use_with_card')}</option
				>
				<option value="multiple_use_without_card"
					>{$t('vouchers.usageLimitTypes.multiple_use_without_card')}</option
				>
			</select>
			<p class="text-sm text-gray-500 mt-1">
				{$t('vouchers.usageLimitTypeHint')}
			</p>
		</div>
	</div>

	<!-- Gültig von / Gültig bis -->
	<div class="space-y-4 md:grid md:grid-cols-2 md:gap-4 md:space-y-0">
		<div>
			<label for="validFrom" class="label">{$t('vouchers.validFrom')}</label>
			<input
				type="date"
				id="validFrom"
				bind:value={validFrom}
				class="input w-full text-base"
				style="min-width: 0;"
			/>
			<p class="text-xs text-gray-500 mt-1 hidden sm:block">
				{$t('vouchers.validFromHint')}
			</p>
		</div>

		<div>
			<div class="flex items-center justify-between mb-1">
				<label for="validUntil" class="text-sm font-medium text-gray-700">
					{$t('vouchers.validUntil')} *
				</label>
				<!-- Quick-Select Buttons (inline with label) -->
				<div class="flex gap-1.5">
					<button
						type="button"
						onclick={() => setExpiryOffset(30)}
						class="px-2 py-0.5 text-xs bg-gray-100 hover:bg-gray-200 text-gray-600 hover:text-gray-800 rounded transition-colors"
					>
						{$t('vouchers.quickSelect.oneMonth')}
					</button>
					<button
						type="button"
						onclick={() => setExpiryOffset(90)}
						class="px-2 py-0.5 text-xs bg-gray-100 hover:bg-gray-200 text-gray-600 hover:text-gray-800 rounded transition-colors"
					>
						{$t('vouchers.quickSelect.threeMonths')}
					</button>
				</div>
			</div>
			<input
				type="date"
				id="validUntil"
				bind:value={validUntil}
				oninput={() => {
					if (errors) errors = { ...errors, validUntil: undefined };
				}}
				required
				class="input w-full text-base {errors?.validUntil
					? 'border-red-500 focus:ring-red-500 focus:border-red-500'
					: ''}"
				style="min-width: 0;"
			/>
			{#if errors?.validUntil}
				<p class="text-red-600 text-sm mt-1">{errors.validUntil}</p>
			{:else}
				<p class="text-xs text-gray-500 mt-1 hidden sm:block">
					{$t('vouchers.validUntilHint')}
				</p>
			{/if}
		</div>
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
