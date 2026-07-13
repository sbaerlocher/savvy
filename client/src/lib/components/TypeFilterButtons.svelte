<script lang="ts">
	import { t } from '$lib/stores/i18n';
	import { get } from 'svelte/store';

	const tr = (key: string) => get(t)(key);

	interface Props {
		typeFilter: string;
		cardsCount: number;
		vouchersCount: number;
		giftCardsCount: number;
		allowToggle?: boolean;
		/** Render the explicit "All" chip. When off, tapping the active type
		 *  toggles back to all — keeps the row narrow on mobile. */
		showAll?: boolean;
		/** 'pill' (default, rounded-full) for the main list header; 'chip'
		 *  (compact rounded-lg) inside the filter menu to match FilterGroup. */
		variant?: 'pill' | 'chip';
	}

	let {
		typeFilter = $bindable(),
		cardsCount,
		vouchersCount,
		giftCardsCount,
		allowToggle = true,
		showAll = true,
		variant = 'pill'
	}: Props = $props();

	function handleClick(type: 'all' | 'cards' | 'vouchers' | 'gift-cards') {
		if (type === 'all') {
			typeFilter = 'all';
		} else {
			// Tapping the active type clears back to all; otherwise select it.
			typeFilter = allowToggle && typeFilter === type ? 'all' : type;
		}
	}

	// Active pill = solid cyan; inactive = subtle warm chip. One consistent
	// accent (brand cyan) keeps the row calm — the type is read from the label,
	// not from a color per type.
	// Mockup: solid teal active pill, white bordered inactive pill, no count.
	const active = 'bg-accent text-white border border-accent';
	const inactive =
		'bg-white text-text-muted border border-border hover:bg-surface-1';
	const base = $derived(
		`inline-flex items-center py-1.5 text-sm font-medium transition-colors whitespace-nowrap ${
			variant === 'chip' ? 'rounded-lg px-3' : 'rounded-full px-4'
		}`
	);
</script>

<div
	class="flex gap-2 {variant === 'chip' ? 'flex-wrap' : 'overflow-x-auto pb-1'}"
>
	{#if showAll}
		<button
			type="button"
			onclick={() => handleClick('all')}
			class="{base} {typeFilter === 'all' ? active : inactive}"
		>
			{tr('common.all')}
		</button>
	{/if}
	{#if cardsCount > 0}
		<button
			type="button"
			onclick={() => handleClick('cards')}
			class="{base} {typeFilter === 'cards' ? active : inactive}"
		>
			{tr('merchantOverview.filterCards')}
		</button>
	{/if}
	{#if vouchersCount > 0}
		<button
			type="button"
			onclick={() => handleClick('vouchers')}
			class="{base} {typeFilter === 'vouchers' ? active : inactive}"
		>
			{tr('merchantOverview.filterVouchers')}
		</button>
	{/if}
	{#if giftCardsCount > 0}
		<button
			type="button"
			onclick={() => handleClick('gift-cards')}
			class="{base} {typeFilter === 'gift-cards' ? active : inactive}"
		>
			{tr('merchantOverview.filterGiftCards')}
		</button>
	{/if}
</div>
