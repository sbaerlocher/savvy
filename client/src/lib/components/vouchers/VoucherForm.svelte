<script lang="ts">
	import type { Snippet } from 'svelte';
	import { platform } from '$lib/utils/platform';
	import { merchantsApi } from '$lib/api';
	import { onMount } from 'svelte';

	import type { MerchantDTO } from '$lib/types/api';
	import { logger } from '$lib/utils/logger';
	import { locale, t } from '$lib/stores/i18n';
	import MerchantSelect from '$lib/components/MerchantSelect.svelte';
	import BarcodeFields from '$lib/components/forms/BarcodeFields.svelte';

	const componentLogger = logger.child('VoucherForm');
	const SUPPORTED_CURRENCIES = ['CHF', 'EUR', 'USD', 'GBP'] as const;

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
		/** Trailing action rendered right-aligned in the desktop action row. */
		trailingActions?: Snippet;
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
		submitLabel,
		trailingActions
	}: Props = $props();

	// `platform` is a module constant, so a plain const, not $derived.
	const IS_DESKTOP = platform === 'other';

	let merchants = $state<MerchantDTO[]>([]);

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

	function setExpiryOffset(days: number) {
		// Transient local, not reactive state — svelte/prefer-svelte-reactivity
		// is a false positive here (per PR review).
		// eslint-disable-next-line svelte/prefer-svelte-reactivity
		const date = new Date();
		date.setDate(date.getDate() + days);
		validUntil = date.toISOString().split('T')[0];
	}
</script>

<form
	onsubmit={(e) => {
		e.preventDefault();
		// Free vouchers are gratis: coerce the value to 0 at submit time
		// (the field is hidden for free, the server enforces this too).
		if (type === 'free') value = 0;
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
			<p class="text-danger-600 text-sm mt-1">{errors.merchant}</p>
		{/if}
	</div>

	<BarcodeFields
		bind:value={code}
		bind:barcodeType
		label={$t('vouchers.code')}
		typeLabel={$t('vouchers.barcodeType')}
		inputId="code"
		placeholder={$t('vouchers.codePlaceholder')}
		part={IS_DESKTOP ? 'value' : 'both'}
	/>

	{#snippet descriptionField()}
		<div>
			<label for="description" class="label">{$t('vouchers.description')}</label
			>
			<textarea
				id="description"
				bind:value={description}
				rows="3"
				class="input"
				placeholder={$t('vouchers.descriptionPlaceholder')}></textarea>
		</div>
	{/snippet}

	<!-- Desktop moves the description to the end of the form (mockup); the native
	     layouts keep it right after the code. -->
	{#if !IS_DESKTOP}
		{@render descriptionField()}
	{/if}

	<!-- Typ / Wert — desktop pairs the barcode type with the voucher type first. -->
	{#if IS_DESKTOP}
		<div class="grid grid-cols-2 gap-4">
			<BarcodeFields
				bind:value={code}
				bind:barcodeType
				label={$t('vouchers.code')}
				typeLabel={$t('vouchers.barcodeType')}
				inputId="code"
				part="type"
			/>
			<div>
				<label for="type" class="label">{$t('vouchers.type')} *</label>
				<select id="type" bind:value={type} required class="input">
					<option value="percentage">{$t('vouchers.typePercentage')}</option>
					<option value="fixed_amount">{$t('vouchers.typeFixedAmount')}</option>
					<option value="points_multiplier"
						>{$t('vouchers.typePointsMultiplier')}</option
					>
					<option value="bonus_points">{$t('vouchers.typeBonusPoints')}</option>
					<option value="free">{$t('vouchers.typeFree')}</option>
				</select>
			</div>
		</div>
	{/if}

	<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
		{#if !IS_DESKTOP}
			<div>
				<label for="type" class="label">{$t('vouchers.type')} *</label>
				<select id="type" bind:value={type} required class="input">
					<option value="percentage">{$t('vouchers.typePercentage')}</option>
					<option value="fixed_amount">{$t('vouchers.typeFixedAmount')}</option>
					<option value="points_multiplier"
						>{$t('vouchers.typePointsMultiplier')}</option
					>
					<option value="bonus_points">{$t('vouchers.typeBonusPoints')}</option>
					<option value="free">{$t('vouchers.typeFree')}</option>
				</select>
			</div>
		{/if}

		{#if type !== 'free'}
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
							? 'border-danger-500 focus:ring-danger-500 focus:border-danger-500'
							: ''}"
						placeholder="10.00"
					/>
					{#if type === 'fixed_amount'}
						<select
							id="currency"
							bind:value={currency}
							required
							class="w-28 input"
						>
							{#each SUPPORTED_CURRENCIES as c (c)}
								<option value={c}>{c}</option>
							{/each}
						</select>
					{/if}
				</div>
				{#if errors?.value}
					<p class="text-danger-600 text-sm mt-1">{errors.value}</p>
				{:else}
					<p class="text-sm text-text-subtle mt-1">
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
		{/if}
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
				>
					{#each SUPPORTED_CURRENCIES as c (c)}
						<option value={c}>{c}</option>
					{/each}
				</select>
			</div>
			<p class="text-sm text-text-subtle mt-1">
				{$t('vouchers.minPurchaseAmountHint')}
			</p>
		</div>

		<div>
			<label for="usageLimitType" class="label"
				>{$t('vouchers.usageLimitType')}</label
			>
			<select id="usageLimitType" bind:value={usageLimitType} class="input">
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
			<p class="text-sm text-text-subtle mt-1">
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
			<p class="text-xs text-text-subtle mt-1 hidden sm:block">
				{$t('vouchers.validFromHint')}
			</p>
		</div>

		<div>
			<div class="flex items-center justify-between mb-1">
				<label for="validUntil" class="text-sm font-medium text-text-ink2">
					{$t('vouchers.validUntil')} *
				</label>
				<!-- Quick-Select Buttons (inline with label) -->
				<div class="flex gap-1.5">
					<button
						type="button"
						onclick={() => setExpiryOffset(30)}
						class="px-2 py-0.5 text-xs bg-border-soft hover:bg-border text-text-muted hover:text-text-strong rounded transition-colors"
					>
						{$t('vouchers.quickSelect.oneMonth')}
					</button>
					<button
						type="button"
						onclick={() => setExpiryOffset(90)}
						class="px-2 py-0.5 text-xs bg-border-soft hover:bg-border text-text-muted hover:text-text-strong rounded transition-colors"
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
					? 'border-danger-500 focus:ring-danger-500 focus:border-danger-500'
					: ''}"
				style="min-width: 0;"
			/>
			{#if errors?.validUntil}
				<p class="text-danger-600 text-sm mt-1">{errors.validUntil}</p>
			{:else}
				<p class="text-xs text-text-subtle mt-1 hidden sm:block">
					{$t('vouchers.validUntilHint')}
				</p>
			{/if}
		</div>
	</div>

	{#if IS_DESKTOP}
		{@render descriptionField()}
	{/if}

	<!-- Desktop (mockup) puts save + cancel left and the trailing action (delete)
	     hard right on a divided row; the native layouts keep the full-width save. -->
	<div
		class={IS_DESKTOP
			? 'mt-6 flex items-center gap-2.5 border-t border-border-soft pt-5'
			: 'flex gap-2'}
	>
		<button
			type="submit"
			class="btn btn-primary {IS_DESKTOP ? 'px-6' : 'flex-1'}"
			disabled={isLoading}
		>
			{isLoading ? $t('common.saving') : submitLabel || $t('common.save')}
		</button>
		<button type="button" onclick={onCancel} class="btn btn-ghost"
			>{$t('common.cancel')}</button
		>
		{#if trailingActions}
			<div class="ms-auto">{@render trailingActions()}</div>
		{/if}
	</div>
</form>
