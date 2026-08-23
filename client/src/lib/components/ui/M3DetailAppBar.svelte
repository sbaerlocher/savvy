<script lang="ts">
	import type { Snippet } from 'svelte';
	import { t } from '$lib/stores/i18n';
	import { get } from 'svelte/store';

	// Material 3 small top app bar for the Android resource-detail screens
	// (screen-ResourceDetailAndroid). One navigation glyph, a single-line title
	// that ellipsises, and up to two trailing 44px icon buttons. The eyebrow
	// that iOS/desktop put above the title lives in the page body here, so this
	// bar never carries one.
	let {
		title,
		/** Back arrow in view mode, close cross while editing. */
		nav = 'back',
		onNav,
		actions
	}: {
		title: string;
		nav?: 'back' | 'close';
		onNav: () => void;
		actions?: Snippet;
	} = $props();

	const tr = (key: string) => get(t)(key);

	const NAV_PATHS = {
		back: 'M19 12H5M12 19l-7-7 7-7',
		close: 'M18 6L6 18M6 6l12 12'
	};
</script>

<div class="-mx-4 mb-1.5 flex items-center gap-1.5 py-2 pr-2 pl-3">
	<button
		type="button"
		onclick={onNav}
		aria-label={nav === 'close' ? tr('common.cancel') : tr('common.back')}
		class="text-text-ink2 hover:bg-surface-1 inline-flex h-11 w-11 shrink-0 items-center justify-center rounded-m3-full transition-colors"
	>
		<svg
			class="h-5.5 w-5.5"
			fill="none"
			stroke="currentColor"
			stroke-width="2"
			stroke-linecap="round"
			stroke-linejoin="round"
			viewBox="0 0 24 24"
		>
			<path d={NAV_PATHS[nav]} />
		</svg>
	</button>
	<h1
		class="text-heading text-text min-w-0 flex-1 overflow-hidden font-semibold tracking-tight text-ellipsis whitespace-nowrap"
	>
		{title}
	</h1>
	{#if actions}
		{@render actions()}
	{/if}
</div>
