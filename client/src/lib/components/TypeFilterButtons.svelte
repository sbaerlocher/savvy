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

	// iOS renders the wallet's type filter as a UIKit segmented control: equal
	// segments inside one tinted track, the active segment raised in white
	// (mockup screen-WalletIOS). Only the 'pill' variant switches — the 'chip'
	// variant inside the filter sheet stays a chip row on every platform.
	const segmented = $derived(platform === 'ios' && variant === 'pill');

	function handleClick(type: 'all' | 'cards' | 'vouchers' | 'gift-cards') {
		if (type === 'all') {
			typeFilter = 'all';
			return;
		}
		// A segmented control always keeps exactly one selection, so tapping the
		// active segment is a no-op there. The chip row instead toggles back to
		// 'all', where "nothing highlighted" legibly means "every type".
		if (segmented) {
			typeFilter = type;
			return;
		}
		typeFilter = allowToggle && typeFilter === type ? 'all' : type;
	}

	// The three resource types plus the optional explicit "All" entry. One list
	// drives both the chip row and the iOS segmented control, so the two layouts
	// can never drift in which types they offer.
	type TypeKey = 'all' | 'cards' | 'vouchers' | 'gift-cards';
	const entries = $derived(
		(
			[
				// The segmented control has no "nothing selected" reading, so it
				// always carries the 'all' segment — that is the wallet's default
				// filter and it needs somewhere to show.
				{ key: 'all', label: 'common.all', show: showAll || segmented },
				{
					key: 'cards',
					label: 'merchantOverview.filterCards',
					show: cardsCount > 0,
					count: cardsCount
				},
				{
					key: 'vouchers',
					label: 'merchantOverview.filterVouchers',
					show: vouchersCount > 0,
					count: vouchersCount
				},
				{
					key: 'gift-cards',
					label: 'merchantOverview.filterGiftCards',
					show: giftCardsCount > 0,
					count: giftCardsCount
				}
			] satisfies {
				key: TypeKey;
				label: string;
				show: boolean;
				count?: number;
			}[]
		).filter((e) => e.show)
	);

	const segmentBase =
		'flex-1 min-w-0 truncate rounded-sm py-1.75 text-center text-label font-medium transition-colors';

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
	// readable without an "All" chip. With the "All" chip present the row would
	// read "All | Cards 4 | …" — the one chip that has no count of its own — so
	// the counts drop there. On an active (filled) chip the count inherits the
	// chip ink, on an outlined one it steps down to the faint tone.
	const SHOW_COUNT = $derived(platform === 'android' && !showAll);
	const countClass = (isActive: boolean) => (isActive ? '' : 'text-text-faint');
</script>

{#if segmented}
	<!-- iOS segmented control: one tinted track, equal-width segments. -->
	<div class="flex gap-0.5 rounded-md bg-tile-tint p-0.5">
		{#each entries as entry (entry.key)}
			<button
				type="button"
				data-testid="type-chip-{entry.key}"
				onclick={() => handleClick(entry.key)}
				aria-pressed={typeFilter === entry.key}
				class="{segmentBase} {typeFilter === entry.key
					? 'bg-surface text-text shadow-sm'
					: 'text-text-muted'}"
			>
				{tr(entry.label)}
			</button>
		{/each}
	</div>
{:else}
	<div
		class="flex gap-2 {variant === 'chip'
			? 'flex-wrap'
			: platform === 'android'
				? 'scrollbar-none overflow-x-auto'
				: 'overflow-x-auto pb-1'}"
	>
		{#each entries as entry (entry.key)}
			<button
				type="button"
				data-testid="type-chip-{entry.key}"
				onclick={() => handleClick(entry.key)}
				class="{base} {typeFilter === entry.key ? active : inactive}"
			>
				{tr(entry.label)}
				{#if SHOW_COUNT && entry.count !== undefined}
					<span class={countClass(typeFilter === entry.key)}>{entry.count}</span
					>
				{/if}
			</button>
		{/each}
	</div>
{/if}
