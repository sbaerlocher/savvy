<script lang="ts">
	import { t } from '$lib/stores/i18n';

	// Favorite + edit buttons for a resource detail page, rendered in the
	// PageHeader actions slot. The title/badges live in the PageHeader itself.
	let {
		isOffline,
		isFavorite,
		isTogglingFavorite = false,
		canEdit = false,
		favoriteTitleAdd,
		favoriteTitleRemove,
		ontoggleFavorite,
		onstartEdit
	}: {
		isOffline: boolean;
		isFavorite: boolean;
		isTogglingFavorite?: boolean;
		canEdit?: boolean;
		favoriteTitleAdd?: string;
		favoriteTitleRemove?: string;
		ontoggleFavorite?: () => void;
		onstartEdit?: () => void;
	} = $props();

	const LOCK_ICON_PATH =
		'M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z';
</script>

<div class="flex gap-2 flex-shrink-0">
	<button
		data-testid="favorite-button"
		onclick={ontoggleFavorite}
		disabled={isOffline || isTogglingFavorite}
		class="btn btn-xs {isFavorite
			? 'btn-favorite'
			: 'bg-border hover:bg-border-field text-text-ink2'} {isOffline ||
		isTogglingFavorite
			? 'opacity-50 cursor-not-allowed'
			: ''}"
		title={isFavorite
			? (favoriteTitleRemove ?? $t('common.removeFromFavorites'))
			: (favoriteTitleAdd ?? $t('common.addToFavorites'))}
	>
		<span class="inline-block w-4 text-center leading-none"
			>{isFavorite ? '★' : '☆'}</span
		>
	</button>
	{#if canEdit}
		<button
			onclick={onstartEdit}
			disabled={isOffline}
			class="btn btn-xs btn-gray whitespace-nowrap flex items-center gap-1.5 {isOffline
				? 'opacity-50 cursor-not-allowed pointer-events-none blur-[0.5px]'
				: ''}"
		>
			{#if isOffline}
				<svg
					class="w-3.5 h-3.5"
					fill="none"
					stroke="currentColor"
					viewBox="0 0 24 24"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d={LOCK_ICON_PATH}
					></path>
				</svg>
			{/if}
			{$t('common.edit')}
		</button>
	{/if}
</div>
