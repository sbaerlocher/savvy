<script lang="ts">
	import { merchantsApi } from '$lib/api';
	import { onMount } from 'svelte';
	import { t } from '$lib/stores/i18n';

	import type { MerchantDTO } from '$lib/types/api';
	import { logger } from '$lib/utils/logger';
	import MerchantSelect from '$lib/components/MerchantSelect.svelte';
	import BarcodeFields from '$lib/components/forms/BarcodeFields.svelte';

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

	<BarcodeFields
		bind:value={cardNumber}
		bind:barcodeType
		label={$t('cards.cardNumber')}
		typeLabel={$t('cards.barcodeType')}
		inputId="cardNumber"
		placeholder="1234567890"
	/>

	{#if status !== undefined}
		<div>
			<label for="status" class="label">{$t('cards.statusLabel')}</label>
			<select id="status" bind:value={status} class="input">
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
