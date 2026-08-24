<script lang="ts">
	import { t } from '$lib/stores/i18n';
	import { get } from 'svelte/store';
	import { platform } from '$lib/utils/platform';
	import TypeFilterButtons from './TypeFilterButtons.svelte';
	import FilterGroup from './ui/FilterGroup.svelte';

	const tr = (key: string) => get(t)(key);

	interface SelectOption {
		value: string;
		label: string;
	}

	interface Props {
		typeFilter: string;
		statusFilter: string;
		sortBy: string;
		sortOptions: SelectOption[];
		statusOptions: SelectOption[];
		cardsCount: number;
		vouchersCount: number;
		giftCardsCount: number;
		showStatusFilter?: boolean;
		hasActiveFilters: boolean;
		onReset: () => void;
		idPrefix?: string;
		ownerFilter?: string;
		ownerOptions?: SelectOption[];
		favoritesOnly?: boolean;
		expiringFilter?: string;
		expiringOptions?: SelectOption[];
		showExpiringFilter?: boolean;
		showAll?: boolean;
		/** Let tapping the active type clear back to "all". Off while a batch
		 *  selection is active — batch endpoints are per type, so a mixed
		 *  selection would route to the wrong one. */
		allowTypeToggle?: boolean;
		/** Drop the in-panel reset button. Set by call sites whose own chrome
		 *  already carries a reset action (the Android wallet filter sheet). */
		hideReset?: boolean;
		/** Label the type chip row like the other groups. The Android merchants
		 *  sheet does (mockup); the wallet sheet leaves the row unlabelled. The
		 *  iOS flat sheet turns it on implicitly, see iosFlatGroups. */
		showTypeLabel?: boolean;
		/** iOS only: render each group with its own uppercase caption and lay the
		 *  options out flat — status as a chip row, sort as a checkmark list, all
		 *  expanded at once (mockup screen-MerchantsIOS). The default is the
		 *  wallet's accordion, where one row expands at a time. */
		iosFlatGroups?: boolean;
	}

	let {
		typeFilter = $bindable(),
		statusFilter = $bindable(),
		sortBy = $bindable(),
		sortOptions,
		statusOptions,
		cardsCount,
		vouchersCount,
		giftCardsCount,
		showStatusFilter = true,
		hasActiveFilters,
		onReset,
		idPrefix = 'filter',
		ownerFilter = $bindable(undefined),
		ownerOptions,
		favoritesOnly = $bindable(undefined),
		expiringFilter = $bindable(undefined),
		expiringOptions,
		showExpiringFilter = true,
		showAll = true,
		allowTypeToggle = true,
		hideReset = false,
		showTypeLabel = false,
		iosFlatGroups = false
	}: Props = $props();

	const isIos = platform === 'ios';
	const isAndroid = platform === 'android';
	// iOS grouped-inset: each filter group sits in its own translucent card on
	// the glass sheet. Android: groups sit flat on the tonal sheet itself, only
	// separated by their uppercase label (wallet mockup). Desktop keeps the
	// hairline-divided flat layout.
	// The flat variant only exists on iOS; elsewhere the prop is inert. A prop,
	// so $derived — unlike `platform`, which is a module constant.
	const iosFlat = $derived(isIos && iosFlatGroups);
	const groupClass = $derived(
		isIos
			? `liquid-glass-card rounded-[var(--radius-inset)] px-4${iosFlat ? ' py-3.5' : ''}`
			: ''
	);
	const dividerClass =
		isIos || isAndroid ? 'hidden' : 'border-t border-border-soft';

	// iOS accordion: which filter row is expanded. Held here so opening one
	// collapses the others. Sort starts open (mockup screen-WalletIOS).
	let openGroup = $state('sort');
</script>

<div class={isAndroid ? 'space-y-4.5' : isIos ? 'space-y-3' : ''}>
	<!-- Type Filter. The iOS flat sheet captions it like every other group. -->
	<div
		class="{isAndroid ? '' : 'pb-4'} {groupClass} {isIos && !iosFlat
			? 'pt-4'
			: ''}"
	>
		{#if showTypeLabel || iosFlat}
			<!-- The other groups get their label from FilterGroup; the chip row has
			     none of its own. The merchants sheets label it (mockups
			     screen-MerchantsAndroid, screen-MerchantsIOS), the wallet sheet
			     does not. -->
			<div
				class="{isIos
					? 'text-section-eyebrow'
					: 'text-eyebrow'} text-text-subtle mb-2.5 uppercase"
			>
				{tr('merchantOverview.detail.typeFilter')}
			</div>
		{/if}
		<TypeFilterButtons
			bind:typeFilter
			{cardsCount}
			{vouchersCount}
			{giftCardsCount}
			{showAll}
			allowToggle={allowTypeToggle}
			variant="chip"
			longAllLabel={iosFlat}
		/>
	</div>

	<!-- Favorites Toggle (detail page only) -->
	{#if favoritesOnly !== undefined}
		<div class={dividerClass}></div>

		<div class="{isAndroid ? '' : 'py-4'} {groupClass}">
			<button
				type="button"
				role="switch"
				aria-checked={favoritesOnly}
				onclick={() => (favoritesOnly = !favoritesOnly)}
				class="flex items-center justify-between w-full cursor-pointer group"
			>
				<div class="flex items-center {isAndroid ? 'gap-2.25' : 'gap-2'}">
					<svg
						class="{isAndroid
							? 'h-4.25 w-4.25'
							: 'h-4 w-4'} transition-colors {favoritesOnly
							? 'text-accent-600'
							: 'text-text-faint group-hover:text-accent-400'}"
						fill={favoritesOnly ? 'currentColor' : 'none'}
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z"
						></path>
					</svg>
					<span
						class="{isAndroid
							? 'text-body font-medium'
							: 'text-sm font-medium'} text-text-ink2"
						>{tr('common.favoritesOnly')}</span
					>
				</div>
				<!-- Android M3 switch: 44x26 track with a 22px thumb (mockup). -->
				<div
					class="relative rounded-full transition-colors {isAndroid
						? 'h-6.5 w-11'
						: 'h-5 w-9'} {favoritesOnly ? 'bg-accent' : 'bg-border'}"
				>
					<div
						class="absolute top-0.5 left-0.5 rounded-full bg-white shadow-sm transition-transform {isAndroid
							? 'h-5.5 w-5.5'
							: 'h-4 w-4'} {favoritesOnly
							? isAndroid
								? 'translate-x-4.5'
								: 'translate-x-4'
							: 'translate-x-0'}"
					></div>
				</div>
			</button>
		</div>
	{/if}
	<div class={dividerClass}></div>

	<!-- Status and Sort. The iOS flat sheet puts Status first (mockup
	     screen-MerchantsIOS); the accordion call sites — the desktop panel and the
	     Android tonal sheet — keep their established Sort-then-Status order, which
	     this iOS-scoped change has no mandate to reshuffle. -->
	{#if iosFlat}
		{@render statusGroup()}
		<div class={dividerClass}></div>
		{@render sortGroup()}
	{:else}
		{@render sortGroup()}
		{#if showStatusFilter}
			<div class={dividerClass}></div>
			{@render statusGroup()}
		{/if}
	{/if}

	<!-- Owner Filter (detail page only) -->
	{#if ownerFilter !== undefined && ownerOptions}
		<div class={dividerClass}></div>

		<div class={groupClass}>
			<FilterGroup
				label={tr('merchantOverview.detail.ownerFilter')}
				bind:value={ownerFilter}
				options={ownerOptions}
				idPrefix="{idPrefix}-owner"
				groupKey="owner"
				bind:openGroup
			/>
		</div>
	{/if}

	<!-- Expiring Filter (detail page only, hidden for cards) -->
	{#if expiringFilter !== undefined && expiringOptions && showExpiringFilter}
		<div class={dividerClass}></div>

		<div class={groupClass}>
			<FilterGroup
				label={tr('merchantOverview.detail.expiringFilter')}
				bind:value={expiringFilter}
				options={expiringOptions}
				idPrefix="{idPrefix}-expiring"
				groupKey="expiring"
				bind:openGroup
			/>
		</div>
	{/if}

	<!-- Reset Filters. Hidden where the call site's own header already carries a
	     reset action (Android wallet sheet), otherwise it would be a duplicate. -->
	{#if hasActiveFilters && !hideReset}
		<div class={dividerClass}></div>

		<div class="py-4 {groupClass}">
			<button
				type="button"
				onclick={onReset}
				class="w-full btn btn-sm btn-ghost text-sm"
			>
				{tr('common.resetFilters')}
			</button>
		</div>
	{/if}
</div>

{#snippet sortGroup()}
	<!-- Sort — iOS shows the options expanded with checkmarks (mockup). -->
	<div class={groupClass}>
		<FilterGroup
			label={tr('merchantOverview.sortBy')}
			bind:value={sortBy}
			options={sortOptions}
			idPrefix="{idPrefix}-sort"
			groupKey="sort"
			flat={iosFlat}
			bind:openGroup
		/>
	</div>
{/snippet}

{#snippet statusGroup()}
	<!-- Status Filter. The flat iOS groups lay it out as a chip row (mockup
	     screen-MerchantsIOS); the accordion keeps the checkmark list. -->
	{#if showStatusFilter}
		<div class={groupClass}>
			<FilterGroup
				label={tr('merchantOverview.detail.statusFilter')}
				bind:value={statusFilter}
				options={statusOptions}
				idPrefix="{idPrefix}-status"
				groupKey="status"
				flat={iosFlat}
				variant={iosFlat ? 'chips' : 'list'}
				bind:openGroup
			/>
		</div>
	{/if}
{/snippet}
