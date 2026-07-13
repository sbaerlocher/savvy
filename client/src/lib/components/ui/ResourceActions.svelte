<script lang="ts">
	import { t } from '$lib/stores/i18n';
	import { platform } from '$lib/utils/platform';

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
	const EDIT_ICON_PATH =
		'M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z';
</script>

<!-- 40px boxed icon buttons matching the top-screen header actions
     (bell / plus): favorite = star, edit = pencil. -->
<div class="flex items-center gap-2.5 flex-shrink-0">
	<button
		type="button"
		data-testid="favorite-button"
		onclick={ontoggleFavorite}
		disabled={isOffline || isTogglingFavorite}
		aria-pressed={isFavorite}
		class="inline-flex h-10 w-10 items-center justify-center rounded-xl border transition-colors {isFavorite
			? 'border-accent bg-accent text-white'
			: 'border-border bg-white text-text-muted hover:bg-surface-1'} {isOffline ||
		isTogglingFavorite
			? 'opacity-50 cursor-not-allowed'
			: ''}"
		title={isFavorite
			? (favoriteTitleRemove ?? $t('common.removeFromFavorites'))
			: (favoriteTitleAdd ?? $t('common.addToFavorites'))}
	>
		<svg
			class="h-5 w-5"
			viewBox="0 0 24 24"
			fill={isFavorite ? 'currentColor' : 'none'}
			stroke="currentColor"
			stroke-width="2"
		>
			<path
				stroke-linecap="round"
				stroke-linejoin="round"
				d="M11.48 3.5l2.3 4.66 5.14.75-3.72 3.62.88 5.12-4.6-2.42-4.6 2.42.88-5.12L4.04 8.9l5.14-.75 2.3-4.66z"
			/>
		</svg>
	</button>
	<!-- Android moves edit to a bottom-right FAB (M3); iOS/desktop keep it here. -->
	{#if canEdit && platform !== 'android'}
		<button
			type="button"
			onclick={onstartEdit}
			disabled={isOffline}
			aria-label={$t('common.edit')}
			title={$t('common.edit')}
			class="inline-flex h-10 w-10 items-center justify-center rounded-xl border border-border bg-white text-text-muted transition-colors hover:bg-surface-1 {isOffline
				? 'opacity-50 cursor-not-allowed pointer-events-none blur-[0.5px]'
				: ''}"
		>
			<svg
				class="h-5 w-5"
				fill="none"
				stroke="currentColor"
				stroke-width="2"
				viewBox="0 0 24 24"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					d={isOffline ? LOCK_ICON_PATH : EDIT_ICON_PATH}
				/>
			</svg>
		</button>
	{/if}
</div>
