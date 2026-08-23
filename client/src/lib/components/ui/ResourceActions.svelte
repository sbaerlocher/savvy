<script lang="ts">
	import { ICON_LOCK } from '$lib/icons';
	import { t } from '$lib/stores/i18n';
	import { platform } from '$lib/utils/platform';

	// Favorite + edit buttons for a resource detail page, rendered in the
	// PageHeader actions slot. The title/badges live in the PageHeader itself.
	//
	// iOS renders these as a flat accent-coloured glyph row (favorite · share ·
	// edit · more) per screen-ResourceDetailIOS; Android/desktop keep the boxed
	// 40px buttons. `platform` is a module constant, so plain consts.
	let {
		isOffline,
		isFavorite,
		isTogglingFavorite = false,
		canEdit = false,
		canShare = false,
		hasMore = false,
		favoriteTitleAdd,
		favoriteTitleRemove,
		ontoggleFavorite,
		onstartEdit,
		onshare,
		onmore
	}: {
		isOffline: boolean;
		isFavorite: boolean;
		isTogglingFavorite?: boolean;
		canEdit?: boolean;
		/** iOS: render the share glyph (owner only). */
		canShare?: boolean;
		/** iOS: render the ••• glyph opening the context menu (owner only). */
		hasMore?: boolean;
		favoriteTitleAdd?: string;
		favoriteTitleRemove?: string;
		ontoggleFavorite?: () => void;
		onstartEdit?: () => void;
		onshare?: () => void;
		onmore?: () => void;
	} = $props();

	const IS_IOS = platform === 'ios';

	const LOCK_ICON_PATH = ICON_LOCK;
	const EDIT_ICON_PATH =
		'M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z';
	// iOS glyph paths (mockup): thinner pencil, tray-with-arrow share, dotted
	// circle for "more". The star is shared with the boxed variant below.
	const IOS_EDIT_PATH = 'M12 20h9M16.5 3.5a2.1 2.1 0 013 3L7 19l-4 1 1-4z';
	const IOS_STAR_PATH =
		'M12 3l2.7 5.6 6.1.9-4.4 4.3 1 6.1-5.4-2.9-5.4 2.9 1-6.1-4.4-4.3 6.1-.9z';

	const favoriteTitle = $derived(
		isFavorite
			? (favoriteTitleRemove ?? $t('common.removeFromFavorites'))
			: (favoriteTitleAdd ?? $t('common.addToFavorites'))
	);
</script>

{#if IS_IOS}
	<!-- iOS: bare accent glyphs in the nav bar (mockup gap 17px, no chrome). -->
	<div class="flex flex-shrink-0 items-center gap-[17px]">
		<button
			type="button"
			data-testid="favorite-button"
			onclick={ontoggleFavorite}
			disabled={isOffline || isTogglingFavorite}
			aria-pressed={isFavorite}
			title={favoriteTitle}
			class="flex items-center justify-center text-accent disabled:opacity-40"
		>
			<svg
				class="h-[23px] w-[23px]"
				viewBox="0 0 24 24"
				fill={isFavorite ? 'currentColor' : 'none'}
				stroke="currentColor"
				stroke-width="1.4"
				stroke-linecap="round"
				stroke-linejoin="round"
			>
				<path d={IOS_STAR_PATH} />
			</svg>
		</button>

		{#if canShare}
			<button
				type="button"
				onclick={onshare}
				disabled={isOffline}
				aria-label={$t('common.share')}
				title={$t('common.share')}
				class="flex items-center justify-center text-accent disabled:opacity-40"
			>
				<svg
					class="h-[22px] w-[22px]"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="1.9"
					stroke-linecap="round"
					stroke-linejoin="round"
				>
					<path d="M12 15.5V3.4" />
					<path d="M8.4 7l3.6-3.6L15.6 7" />
					<path
						d="M7.6 10H6.2A2.2 2.2 0 004 12.2v6.6A2.2 2.2 0 006.2 21h11.6a2.2 2.2 0 002.2-2.2v-6.6A2.2 2.2 0 0017.8 10h-1.4"
					/>
				</svg>
			</button>
		{/if}

		{#if canEdit}
			<button
				type="button"
				onclick={onstartEdit}
				disabled={isOffline}
				aria-label={$t('common.edit')}
				title={$t('common.edit')}
				class="flex items-center justify-center text-accent disabled:opacity-40"
			>
				<svg
					class="h-[22px] w-[22px]"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
					stroke-linecap="round"
					stroke-linejoin="round"
				>
					<path d={isOffline ? LOCK_ICON_PATH : IOS_EDIT_PATH} />
				</svg>
			</button>
		{/if}

		{#if hasMore}
			<button
				type="button"
				onclick={onmore}
				disabled={isOffline}
				aria-label={$t('common.moreActions')}
				title={$t('common.moreActions')}
				class="flex items-center justify-center text-accent disabled:opacity-40"
			>
				<svg
					class="h-[23px] w-[23px]"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="1.8"
				>
					<circle cx="12" cy="12" r="9.2" />
					<circle cx="7.6" cy="12" r="1.15" fill="currentColor" stroke="none" />
					<circle cx="12" cy="12" r="1.15" fill="currentColor" stroke="none" />
					<circle
						cx="16.4"
						cy="12"
						r="1.15"
						fill="currentColor"
						stroke="none"
					/>
				</svg>
			</button>
		{/if}
	</div>
{:else}
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
			title={favoriteTitle}
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
		<!-- Android moves edit to a bottom-right FAB (M3); desktop keeps it here. -->
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
{/if}
