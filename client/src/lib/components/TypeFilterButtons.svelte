<script lang="ts">
	import { categoryColors } from '$lib/utils/category-colors';
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

	function handleClick(type: 'cards' | 'vouchers' | 'gift-cards') {
		if (allowToggle && typeFilter === type) {
			typeFilter = 'all';
		} else {
			typeFilter = type;
		}
	}
</script>

<div class="grid grid-cols-3 gap-2">
	{#if cardsCount > 0}
		<button
			type="button"
			onclick={() => handleClick('cards')}
			class="flex flex-col items-center justify-center py-2 rounded-xl text-xs font-semibold transition-colors {typeFilter ===
			'cards'
				? categoryColors.cards.filter
				: 'bg-gray-50 text-gray-500'}"
		>
			<svg
				class="w-5 h-5 mb-0.5"
				fill="none"
				stroke="currentColor"
				viewBox="0 0 24 24"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z"
				></path>
			</svg>
			{tr('merchantOverview.filterCards')}
			<span class="text-[10px] font-normal opacity-70">{cardsCount}</span>
		</button>
	{/if}
	{#if vouchersCount > 0}
		<button
			type="button"
			onclick={() => handleClick('vouchers')}
			class="flex flex-col items-center justify-center py-2 rounded-xl text-xs font-semibold transition-colors {typeFilter ===
			'vouchers'
				? categoryColors.vouchers.filter
				: 'bg-gray-50 text-gray-500'}"
		>
			<svg
				class="w-5 h-5 mb-0.5"
				fill="none"
				stroke="currentColor"
				viewBox="0 0 24 24"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d="M15 5v2m0 4v2m0 4v2M5 5a2 2 0 00-2 2v3a2 2 0 110 4v3a2 2 0 002 2h14a2 2 0 002-2v-3a2 2 0 110-4V7a2 2 0 00-2-2H5z"
				></path>
			</svg>
			{tr('merchantOverview.filterVouchers')}
			<span class="text-[10px] font-normal opacity-70">{vouchersCount}</span>
		</button>
	{/if}
	{#if giftCardsCount > 0}
		<button
			type="button"
			onclick={() => handleClick('gift-cards')}
			class="flex flex-col items-center justify-center py-2 rounded-xl text-xs font-semibold transition-colors {typeFilter ===
			'gift-cards'
				? categoryColors.giftCards.filter
				: 'bg-gray-50 text-gray-500'}"
		>
			<svg
				class="w-5 h-5 mb-0.5"
				fill="none"
				stroke="currentColor"
				viewBox="0 0 24 24"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d="M12 8v13m0-13V6a2 2 0 112 2h-2zm0 0V5.5A2.5 2.5 0 109.5 8H12zm-7 4h14M5 12a2 2 0 110-4h14a2 2 0 110 4M5 12v7a2 2 0 002 2h10a2 2 0 002-2v-7"
				></path>
			</svg>
			{tr('merchantOverview.filterGiftCards')}
			<span class="text-[10px] font-normal opacity-70">{giftCardsCount}</span>
		</button>
	{/if}
</div>
