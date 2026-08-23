<script lang="ts">
	import type { Snippet } from 'svelte';
	import { platform } from '$lib/utils/platform';
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
		program = $bindable(''),
		barcodeType = $bindable('CODE128'),
		notes = $bindable(''),
		status = $bindable(undefined),
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
		part={PAIRED ? 'value' : 'both'}
	/>

	{#snippet statusField()}
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
	{/snippet}

	{#if PAIRED}
		<!-- Desktop pairs barcode type with status on one row (mockup). -->
		<div class="grid grid-cols-2 gap-4">
			<BarcodeFields
				bind:value={cardNumber}
				bind:barcodeType
				label={$t('cards.cardNumber')}
				typeLabel={$t('cards.barcodeType')}
				inputId="cardNumber"
				part="type"
			/>
			{#if status !== undefined}
				{@render statusField()}
			{/if}
		</div>
	{:else if status !== undefined}
		{@render statusField()}
	{/if}

	<div>
		<label for="notes" class="label">{$t('cards.notes')}</label>
		<textarea id="notes" bind:value={notes} rows="3" class="input"></textarea>
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
