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
	}

	let {
		typeFilter = $bindable(),
		cardsCount,
		vouchersCount,
		giftCardsCount,
		allowToggle = true
	}: Props = $props();

	function handleClick(type: 'all' | 'cards' | 'vouchers' | 'gift-cards') {
		if (type === 'all') {
			typeFilter = 'all';
		} else if (allowToggle && typeFilter === type) {
			typeFilter = 'all';
		} else {
			typeFilter = type;
		}
	}

	// Active pill = solid cyan; inactive = subtle warm chip. One consistent
	// accent (brand cyan) keeps the row calm — the type is read from the label,
	// not from a color per type.
	const active = 'bg-cyan-600 text-white';
	const inactive = 'bg-white text-gray-600 hover:bg-gray-50';
	const base =
		'inline-flex items-center gap-1.5 rounded-full px-4 py-1.5 text-sm font-medium transition-colors whitespace-nowrap';
</script>

<div class="flex gap-2 overflow-x-auto pb-1">
	<button
		type="button"
		onclick={() => handleClick('all')}
		class="{base} {typeFilter === 'all' ? active : inactive}"
	>
		{tr('common.all')}
	</button>
	{#if cardsCount > 0}
		<button
			type="button"
			onclick={() => handleClick('cards')}
			class="{base} {typeFilter === 'cards' ? active : inactive}"
		>
			{tr('merchantOverview.filterCards')}
			<span class="text-xs opacity-70">{cardsCount}</span>
		</button>
	{/if}
	{#if vouchersCount > 0}
		<button
			type="button"
			onclick={() => handleClick('vouchers')}
			class="{base} {typeFilter === 'vouchers' ? active : inactive}"
		>
			{tr('merchantOverview.filterVouchers')}
			<span class="text-xs opacity-70">{vouchersCount}</span>
		</button>
	{/if}
	{#if giftCardsCount > 0}
		<button
			type="button"
			onclick={() => handleClick('gift-cards')}
			class="{base} {typeFilter === 'gift-cards' ? active : inactive}"
		>
			{tr('merchantOverview.filterGiftCards')}
			<span class="text-xs opacity-70">{giftCardsCount}</span>
		</button>
	{/if}
</div>
