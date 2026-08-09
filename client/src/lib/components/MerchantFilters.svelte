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
		showAll = true
	}: Props = $props();

	const isIos = platform === 'ios';
	const isAndroid = platform === 'android';
	// iOS grouped-inset: each filter group sits in its own translucent card on
	// the glass sheet. Android: groups sit flat on the tonal sheet itself, only
	// separated by their uppercase label (wallet mockup). Desktop keeps the
	// hairline-divided flat layout.
	const groupClass = isIos
		? 'liquid-glass-card rounded-[var(--radius-inset)] px-4'
		: '';
	const dividerClass =
		isIos || isAndroid ? 'hidden' : 'border-t border-border-soft';
</script>

<div class={isAndroid ? 'space-y-4.5' : isIos ? 'space-y-3' : ''}>
	<!-- Type Filter -->
	<div class="{isAndroid ? '' : 'pb-4'} {groupClass} {isIos ? 'pt-4' : ''}">
		<TypeFilterButtons
			bind:typeFilter
			{cardsCount}
			{vouchersCount}
			{giftCardsCount}
			{showAll}
			variant="chip"
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
						class="absolute rounded-full bg-white transition-transform {isAndroid
							? 'top-0.5 left-0.5 h-5.5 w-5.5 shadow-[var(--shadow-toggle)]'
							: 'top-0.5 left-0.5 h-4 w-4 shadow-sm'} {favoritesOnly
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

	<!-- Sort -->
	<div class={groupClass}>
		<FilterGroup
			label={tr('merchantOverview.sortBy')}
			bind:value={sortBy}
			options={sortOptions}
			idPrefix="{idPrefix}-sort"
		/>
	</div>

	<!-- Status Filter -->
	{#if showStatusFilter}
		<div class={dividerClass}></div>

		<div class={groupClass}>
			<FilterGroup
				label={tr('merchantOverview.detail.statusFilter')}
				bind:value={statusFilter}
				options={statusOptions}
				idPrefix="{idPrefix}-status"
			/>
		</div>
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
			/>
		</div>
	{/if}

	<!-- Reset Filters. Android carries the reset action in the sheet header
	     instead (mockup), so the in-panel button would be a duplicate. -->
	{#if hasActiveFilters && !isAndroid}
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
