<script lang="ts">
	import { t } from '$lib/stores/i18n';
	import { get } from 'svelte/store';
	import TypeFilterButtons from './TypeFilterButtons.svelte';

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
		showExpiringFilter = true
	}: Props = $props();
</script>

<!-- Type Filter -->
<div class="pb-4">
	<TypeFilterButtons
		bind:typeFilter
		{cardsCount}
		{vouchersCount}
		{giftCardsCount}
	/>
</div>

<!-- Favorites Toggle (detail page only) -->
{#if favoritesOnly !== undefined}
	<div class="border-t border-gray-100"></div>

	<div class="py-4">
		<button
			type="button"
			role="switch"
			aria-checked={favoritesOnly}
			onclick={() => (favoritesOnly = !favoritesOnly)}
			class="flex items-center justify-between w-full cursor-pointer group"
		>
			<div class="flex items-center gap-2">
				<svg
					class="w-4 h-4 transition-colors {favoritesOnly
						? 'text-amber-500'
						: 'text-gray-400 group-hover:text-amber-400'}"
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
				<span class="text-sm font-medium text-gray-700"
					>{tr('common.favoritesOnly')}</span
				>
			</div>
			<div
				class="relative w-9 h-5 rounded-full transition-colors {favoritesOnly
					? 'bg-cyan-600'
					: 'bg-gray-200'}"
			>
				<div
					class="absolute top-0.5 left-0.5 w-4 h-4 rounded-full bg-white shadow-sm transition-transform {favoritesOnly
						? 'translate-x-4'
						: 'translate-x-0'}"
				></div>
			</div>
		</button>
	</div>
{/if}
<div class="border-t border-gray-100"></div>

<!-- Sort -->
<div class="py-4">
	<label
		for="{idPrefix}-sort"
		class="block text-xs font-medium text-gray-500 uppercase tracking-wider mb-2"
	>
		{tr('merchantOverview.sortBy')}
	</label>
	<select
		id="{idPrefix}-sort"
		bind:value={sortBy}
		class="w-full bg-gray-50 border border-gray-200 rounded-lg px-3 py-2 text-sm text-gray-700 focus:ring-2 focus:ring-cyan-500 focus:border-cyan-500 transition-colors"
	>
		{#each sortOptions as opt}
			<option value={opt.value}>{opt.label}</option>
		{/each}
	</select>
</div>

<!-- Status Filter -->
{#if showStatusFilter}
	<div class="border-t border-gray-100"></div>

	<div class="py-4">
		<label
			for="{idPrefix}-status"
			class="block text-xs font-medium text-gray-500 uppercase tracking-wider mb-2"
		>
			{tr('merchantOverview.detail.statusFilter')}
		</label>
		<select
			id="{idPrefix}-status"
			bind:value={statusFilter}
			class="w-full bg-gray-50 border border-gray-200 rounded-lg px-3 py-2 text-sm text-gray-700 focus:ring-2 focus:ring-cyan-500 focus:border-cyan-500 transition-colors"
		>
			{#each statusOptions as opt}
				<option value={opt.value}>{opt.label}</option>
			{/each}
		</select>
	</div>
{/if}

<!-- Owner Filter (detail page only) -->
{#if ownerFilter !== undefined && ownerOptions}
	<div class="border-t border-gray-100"></div>

	<div class="py-4">
		<label
			for="{idPrefix}-owner"
			class="block text-xs font-medium text-gray-500 uppercase tracking-wider mb-2"
		>
			{tr('merchantOverview.detail.ownerFilter')}
		</label>
		<select
			id="{idPrefix}-owner"
			bind:value={ownerFilter}
			class="w-full bg-gray-50 border border-gray-200 rounded-lg px-3 py-2 text-sm text-gray-700 focus:ring-2 focus:ring-cyan-500 focus:border-cyan-500 transition-colors"
		>
			{#each ownerOptions as opt}
				<option value={opt.value}>{opt.label}</option>
			{/each}
		</select>
	</div>
{/if}

<!-- Expiring Filter (detail page only, hidden for cards) -->
{#if expiringFilter !== undefined && expiringOptions && showExpiringFilter}
	<div class="border-t border-gray-100"></div>

	<div class="py-4">
		<label
			for="{idPrefix}-expiring"
			class="block text-xs font-medium text-gray-500 uppercase tracking-wider mb-2"
		>
			{tr('merchantOverview.detail.expiringFilter')}
		</label>
		<select
			id="{idPrefix}-expiring"
			bind:value={expiringFilter}
			class="w-full bg-gray-50 border border-gray-200 rounded-lg px-3 py-2 text-sm text-gray-700 focus:ring-2 focus:ring-cyan-500 focus:border-cyan-500 transition-colors"
		>
			{#each expiringOptions as opt}
				<option value={opt.value}>{opt.label}</option>
			{/each}
		</select>
	</div>
{/if}

<!-- Reset Filters -->
{#if hasActiveFilters}
	<div class="border-t border-gray-100"></div>

	<div class="py-4">
		<button
			type="button"
			onclick={onReset}
			class="w-full btn btn-sm btn-ghost text-sm"
		>
			{tr('common.resetFilters')}
		</button>
	</div>
{/if}
