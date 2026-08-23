<script lang="ts">
	import type { Snippet } from 'svelte';
	import { platform } from '$lib/utils/platform';
	import { merchantsApi } from '$lib/api';
	import { onMount } from 'svelte';
	import { locale, t } from '$lib/stores/i18n';

	import type { MerchantDTO } from '$lib/types/api';
	import { logger } from '$lib/utils/logger';
	import MerchantSelect from '$lib/components/MerchantSelect.svelte';
	import BarcodeFields from '$lib/components/forms/BarcodeFields.svelte';

	const componentLogger = logger.child('GiftCardForm');

	interface Props {
		cardNumber: string;
		merchantId?: string;
		initialBalance?: number;
		currency?: string;
		pin?: string;
		barcodeType?: string;
		expiresAt?: string;
		notes?: string;
		onSubmit: () => void;
		onCancel: () => void;
		isLoading: boolean;
		submitLabel?: string;
		/** Trailing action rendered right-aligned in the desktop action row. */
		trailingActions?: Snippet;
		/** Desktop detail edit board: pair fields into rows and move the action
		 *  row onto a divider (mockup screen-ResourceDetailDesktop). The create
		 *  screens are not part of that mockup and keep the stacked layout. */
		pairedLayout?: boolean;
	}

	let {
		cardNumber = $bindable(''),
		merchantId = $bindable(''),
		initialBalance = $bindable(0),
		currency = $bindable('EUR'),
		pin = $bindable(''),
		barcodeType = $bindable('CODE128'),
		expiresAt = $bindable(''),
		notes = $bindable(''),
		onSubmit,
		onCancel,
		isLoading,
		submitLabel,
		trailingActions,
		pairedLayout = false
	}: Props = $props();

	// Paired rows and the divided action row are the desktop detail-edit board;
	// they need both the desktop platform and the caller's opt-in.
	const PAIRED = $derived(platform === 'other' && pairedLayout);

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
</script>

<form
	onsubmit={(e) => {
		e.preventDefault();
		onSubmit();
	}}
	class="space-y-4"
>
	<div>
		<label for="merchant" class="label">{$t('giftCards.merchant')}</label>
		<MerchantSelect {merchants} bind:value={merchantId} id="merchant" />
	</div>

	<BarcodeFields
		bind:value={cardNumber}
		bind:barcodeType
		label={$t('giftCards.cardNumber')}
		typeLabel={$t('giftCards.barcodeType')}
		inputId="cardNumber"
		placeholder="1234567890123"
		part={PAIRED ? 'value' : 'both'}
	/>

	{#snippet pinField()}
		<div>
			<label for="pin" class="label">{$t('giftCards.pin')}</label>
			<input
				id="pin"
				type="text"
				bind:value={pin}
				class="input"
				placeholder="1234"
			/>
			<p class="text-sm text-text-subtle mt-1">{$t('giftCards.pinDesc')}</p>
		</div>
	{/snippet}

	<!-- Desktop pairs barcode type with the PIN on one row (mockup). -->
	{#if PAIRED}
		<div class="grid grid-cols-2 gap-4">
			<BarcodeFields
				bind:value={cardNumber}
				bind:barcodeType
				label={$t('giftCards.cardNumber')}
				typeLabel={$t('giftCards.barcodeType')}
				inputId="cardNumber"
				part="type"
			/>
			{@render pinField()}
		</div>
	{/if}

	<div>
		<label for="initialBalance" class="label"
			>{$t('giftCards.initialBalance')} *</label
		>
		<div
			class="flex gap-2 {$locale?.startsWith('en')
				? 'flex-row-reverse'
				: 'flex-row'}"
		>
			<input
				id="initialBalance"
				type="number"
				step="0.01"
				min="0"
				required
				bind:value={initialBalance}
				placeholder="50.00"
				class="input flex-1"
			/>
			<select id="currency" bind:value={currency} required class="w-28 input">
				<option value="CHF">CHF</option>
				<option value="EUR">EUR</option>
				<option value="USD">USD</option>
				<option value="GBP">GBP</option>
			</select>
		</div>
		<p class="text-sm text-text-subtle mt-1">
			{$t('giftCards.initialBalanceDesc')}
		</p>
	</div>

	{#if !PAIRED}
		{@render pinField()}
	{/if}

	<div>
		<label for="expiresAt" class="label">{$t('giftCards.expiresAt')}</label>
		<input
			id="expiresAt"
			type="date"
			bind:value={expiresAt}
			class="input w-full text-base"
		/>
		<p class="text-xs text-text-subtle mt-1 hidden sm:block">
			{$t('giftCards.expiresAtDesc')}
		</p>
	</div>

	<div>
		<label for="notes" class="label">{$t('giftCards.notes')}</label>
		<textarea
			id="notes"
			bind:value={notes}
			rows="3"
			class="input"
			placeholder={$t('giftCards.notesPlaceholder')}></textarea>
	</div>

	<!-- Desktop (mockup) puts save + cancel left and the trailing action (delete)
	     hard right on a divided row; the native layouts keep the full-width save. -->
	<div
		class={PAIRED
			? 'mt-6 flex items-center gap-2.5 border-t border-border-soft pt-5'
			: 'flex gap-2'}
	>
		<button
			type="submit"
			class="btn btn-primary {PAIRED ? 'px-6' : 'flex-1'}"
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
