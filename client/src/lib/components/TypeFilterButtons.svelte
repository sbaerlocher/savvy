<script lang="ts">
	import { t } from '$lib/stores/i18n';
	import { get } from 'svelte/store';
	import { platform } from '$lib/utils/platform';

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
	// Android M3: inactive chips are outlined/transparent everywhere this row
	// renders (the tonal filter sheet and the wallet header), so the chip takes
	// the surface behind it instead of stamping a white block over it.
	const active = 'bg-accent text-white border border-accent';
	const inactive =
		platform === 'android'
			? 'bg-transparent text-text-muted border border-border-chip hover:bg-surface-1'
			: 'bg-white text-text-muted border border-border hover:bg-surface-1';
	// Android M3 small chip: 8px corners, 14px inset, semibold label with the
	// count trailing it (wallet mockup). Other platforms keep the pill row.
	const base = $derived(
		platform === 'android'
			? `inline-flex items-center gap-1.5 rounded-m3-sm px-3.5 py-2 text-label transition-colors whitespace-nowrap`
			: `inline-flex items-center py-1.5 text-sm font-medium transition-colors whitespace-nowrap ${
					variant === 'chip' ? 'rounded-lg px-3' : 'rounded-full px-4'
				}`
	);
	// Count sits inside the chip on Android only; it is what makes the row
	// readable without an "All" chip. On an active (filled) chip it inherits the
	// chip ink, on an outlined one it steps down to the faint tone.
	const SHOW_COUNT = platform === 'android';
	const countClass = (isActive: boolean) => (isActive ? '' : 'text-text-faint');
</script>

<div
	class="flex gap-2 {variant === 'chip'
		? 'flex-wrap'
		: platform === 'android'
			? 'scrollbar-none overflow-x-auto'
			: 'overflow-x-auto pb-1'}"
>
	{#if showAll}
		<button
			type="button"
			data-testid="type-chip-all"
			onclick={() => handleClick('all')}
			class="{base} {typeFilter === 'all' ? active : inactive}"
		>
			{tr('common.all')}
		</button>
	{/if}
	{#if cardsCount > 0}
		<button
			type="button"
			data-testid="type-chip-cards"
			onclick={() => handleClick('cards')}
			class="{base} {typeFilter === 'cards' ? active : inactive}"
		>
			{tr('merchantOverview.filterCards')}
			{#if SHOW_COUNT}
				<span class={countClass(typeFilter === 'cards')}>{cardsCount}</span>
			{/if}
		</button>
	{/if}
	{#if vouchersCount > 0}
		<button
			type="button"
			data-testid="type-chip-vouchers"
			onclick={() => handleClick('vouchers')}
			class="{base} {typeFilter === 'vouchers' ? active : inactive}"
		>
			{tr('merchantOverview.filterVouchers')}
			{#if SHOW_COUNT}
				<span class={countClass(typeFilter === 'vouchers')}
					>{vouchersCount}</span
				>
			{/if}
		</button>
	{/if}
	{#if giftCardsCount > 0}
		<button
			type="button"
			data-testid="type-chip-gift-cards"
			onclick={() => handleClick('gift-cards')}
			class="{base} {typeFilter === 'gift-cards' ? active : inactive}"
		>
			{tr('merchantOverview.filterGiftCards')}
			{#if SHOW_COUNT}
				<span class={countClass(typeFilter === 'gift-cards')}
					>{giftCardsCount}</span
				>
			{/if}
		</button>
	{/if}
</div>
