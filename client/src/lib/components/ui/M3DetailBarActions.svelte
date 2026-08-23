<script lang="ts">
	import { t } from '$lib/stores/i18n';
	import { get } from 'svelte/store';

	// Trailing icon buttons of the Android detail top app bar: the favourite
	// star (filled accent once set, outlined otherwise) and the owner-only
	// overflow menu. A recipient sees only the star — the mockup's
	// "geteilt mit mir" frame drops the overflow entirely, since every entry
	// behind it is an owner action.
	let {
		isOffline,
		isFavorite,
		isTogglingFavorite = false,
		showOverflow = false,
		favoriteTitleAdd,
		favoriteTitleRemove,
		ontoggleFavorite,
		onoverflow
	}: {
		isOffline: boolean;
		isFavorite: boolean;
		isTogglingFavorite?: boolean;
		showOverflow?: boolean;
		favoriteTitleAdd?: string;
		favoriteTitleRemove?: string;
		ontoggleFavorite?: () => void;
		onoverflow?: () => void;
	} = $props();

	const tr = (key: string) => get(t)(key);

	const STAR_PATH =
		'M12 3l2.7 5.6 6.1.9-4.4 4.3 1 6.1-5.4-2.9-5.4 2.9 1-6.1-4.4-4.3 6.1-.9z';
</script>

<button
	type="button"
	data-testid="favorite-button"
	onclick={ontoggleFavorite}
	disabled={isOffline || isTogglingFavorite}
	aria-pressed={isFavorite}
	title={isFavorite
		? (favoriteTitleRemove ?? tr('common.removeFromFavorites'))
		: (favoriteTitleAdd ?? tr('common.addToFavorites'))}
	class="hover:bg-surface-1 inline-flex h-11 w-11 shrink-0 items-center justify-center rounded-m3-full transition-colors {isOffline ||
	isTogglingFavorite
		? 'cursor-not-allowed opacity-50'
		: ''} {isFavorite ? 'text-accent-600' : 'text-text-ink2'}"
>
	<svg
		class="h-5.5 w-5.5"
		viewBox="0 0 24 24"
		fill={isFavorite ? 'currentColor' : 'none'}
		stroke="currentColor"
		stroke-width="1.9"
		stroke-linecap="round"
		stroke-linejoin="round"
	>
		<path d={STAR_PATH} />
	</svg>
</button>

{#if showOverflow}
	<button
		type="button"
		onclick={onoverflow}
		aria-label={tr('common.more')}
		class="text-text-ink2 hover:bg-surface-1 inline-flex h-11 w-11 shrink-0 items-center justify-center rounded-m3-full transition-colors"
	>
		<svg class="h-5.5 w-5.5" viewBox="0 0 24 24" fill="currentColor">
			<circle cx="12" cy="5" r="2" />
			<circle cx="12" cy="12" r="2" />
			<circle cx="12" cy="19" r="2" />
		</svg>
	</button>
{/if}
