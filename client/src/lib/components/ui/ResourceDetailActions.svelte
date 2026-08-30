<script lang="ts">
	import { t } from '$lib/stores/i18n';
	import { ICON_PENCIL, ICON_TRASH } from '$lib/icons';
	import { platform } from '$lib/utils/platform';
	import { get } from 'svelte/store';

	const tr = (key: string) => get(t)(key);

	// iOS renders the title-row actions as floating liquid-glass circles —
	// the same 36px chrome as PageHeader's iOS back button — icon-only, no
	// labels. The other platforms keep the shared title-action chrome.
	// `platform` is a module constant, so a plain const.
	const IS_IOS = platform === 'ios';
	const BTN = IS_IOS
		? 'liquid-glass-surface inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-full'
		: 'title-action max-sm:px-3';

	// Offline swaps the action glyphs for a lock (matches the former in-detail
	// header buttons).
	const LOCK_PATH =
		'M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z';

	interface Props {
		isFavorite: boolean;
		isTogglingFavorite: boolean;
		canEdit: boolean;
		canDelete: boolean;
		isOffline: boolean;
		favoriteAddLabel: string;
		favoriteRemoveLabel: string;
		onFavorite: () => void;
		onEdit: () => void;
		onDelete: () => void;
		/** The more menu holds the secondary actions on both native platforms
		 *  (share + transfer + delete), so the delete button yields to it.
		 *  Omitted on desktop. */
		onMore?: () => void;
	}

	let {
		isFavorite,
		isTogglingFavorite,
		canEdit,
		canDelete,
		isOffline,
		favoriteAddLabel,
		favoriteRemoveLabel,
		onFavorite,
		onEdit,
		onDelete,
		onMore
	}: Props = $props();
</script>

<!-- Detail title-row actions: favourite · edit · (more) · delete. Labels only
     from `sm` up — on a phone the buttons go icon-only so the row fits beside
     the title. -->
<div class="flex items-center gap-2">
	<button
		type="button"
		data-testid="favorite-button"
		onclick={onFavorite}
		disabled={isOffline || isTogglingFavorite}
		aria-pressed={isFavorite}
		title={isFavorite ? favoriteRemoveLabel : favoriteAddLabel}
		class="{BTN} text-accent disabled:cursor-not-allowed disabled:opacity-50"
	>
		<svg
			class="h-5 w-5"
			viewBox="0 0 24 24"
			fill={isFavorite ? 'currentColor' : 'none'}
			stroke="currentColor"
			stroke-width="1.9"
		>
			<path
				stroke-linecap="round"
				stroke-linejoin="round"
				d="M12 3l2.7 5.6 6.1.9-4.4 4.3 1 6.1-5.4-2.9-5.4 2.9 1-6.1-4.4-4.3 6.1-.9z"
			/>
		</svg>
	</button>
	{#if canEdit}
		<button
			type="button"
			onclick={onEdit}
			disabled={isOffline}
			aria-label={tr('common.edit')}
			class="{BTN} whitespace-nowrap text-text-ink2 disabled:cursor-not-allowed disabled:opacity-50"
		>
			<svg
				class="h-4 w-4"
				fill="none"
				stroke="currentColor"
				stroke-width="2"
				viewBox="0 0 24 24"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					d={isOffline ? LOCK_PATH : ICON_PENCIL}
				/>
			</svg>
			{#if !IS_IOS}
				<span class="hidden sm:inline">{tr('common.edit')}</span>
			{/if}
		</button>
	{/if}
	{#if onMore}
		<button
			type="button"
			onclick={onMore}
			aria-label={tr('common.moreActions')}
			title={tr('common.moreActions')}
			class="{BTN} text-text-ink2"
		>
			<svg class="h-5 w-5" fill="currentColor" viewBox="0 0 24 24">
				<circle cx="5" cy="12" r="1.7" />
				<circle cx="12" cy="12" r="1.7" />
				<circle cx="19" cy="12" r="1.7" />
			</svg>
		</button>
	{/if}
	{#if canDelete && !onMore}
		<button
			type="button"
			onclick={onDelete}
			disabled={isOffline}
			aria-label={tr('common.delete')}
			class="{BTN} whitespace-nowrap {IS_IOS
				? 'text-danger-600'
				: 'border-danger-200 bg-danger-50 text-danger-700 hover:bg-danger-100'} disabled:cursor-not-allowed disabled:opacity-50"
		>
			<svg
				class="h-4 w-4"
				fill="none"
				stroke="currentColor"
				stroke-width="2"
				viewBox="0 0 24 24"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					d={isOffline ? LOCK_PATH : ICON_TRASH}
				/>
			</svg>
			{#if !IS_IOS}
				<span class="hidden sm:inline">{tr('common.delete')}</span>
			{/if}
		</button>
	{/if}
</div>
