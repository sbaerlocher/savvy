<script lang="ts">
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import type { Snippet } from 'svelte';

	/**
	 * The page-level container: one definition of content width, horizontal
	 * padding and page title, for every screen that renders inside the app
	 * shell.
	 *
	 * The root layout's `<main>` already sets `max-w-7xl mx-auto px-4 sm:px-6
	 * lg:px-8`. Pages used to re-state parts of that themselves — five
	 * competing dialects, several stacked *inside* the shell's own container,
	 * which doubled the padding (iOS /dashboard rendered 311px of content
	 * instead of 343px). This component exists so a page never has to say
	 * anything about its own width again.
	 *
	 * A page passes a title and content. Everything else has a default.
	 */
	let {
		title,
		subtitle,
		eyebrow,
		eyebrowAside,
		eyebrowVerbatim = false,
		aside,
		width = 'default',
		header = true,
		headerPlaceholder,
		mobileActions = true,
		headerClass = '',
		actions,
		onBack,
		back,
		children
	}: {
		/** Page title. Renders the page's single `<h1>` via PageHeader. */
		title?: string;
		/** Supporting line under the title (admin list pages use this). */
		subtitle?: string;
		/** Small line above the title — a greeting, a section kicker. */
		eyebrow?: string;
		/** Inline element next to the eyebrow (the dashboard's refresh hint on
		 *  native). Passed straight through to PageHeader. */
		eyebrowAside?: Snippet;
		/** Set when the eyebrow carries user- or data-driven text rather than a
		 *  fixed kicker, so the uppercase treatment is skipped. Passed through. */
		eyebrowVerbatim?: boolean;
		/**
		 * Content that sits *beside* the header rather than under it — the
		 * dashboard's stat tiles. The page used to build this itself with a
		 * two-column grid and place PageHeader into one of the cells, which is
		 * why it was the last screen rendering its own header. "Something to the
		 * right of my title" is a property of the title row, so the shell owns
		 * it.
		 *
		 * From `lg` up it shares the header's row. On a phone there is no second
		 * column, and it goes *after* the content rather than under the title —
		 * the dashboard mockups put the favourites before the statistics
		 * (AGENTS.md, "Mobile: Favorites appear BEFORE statistics").
		 */
		aside?: Snippet;
		/**
		 * `default` — the shared container: shell width, shell padding.
		 * `narrow`  — the reading column: forms, settings, the notification list.
		 *             680px is the desktop mockups' own measure; the screens used
		 *             to carry three values for it (640 / 672 / 680).
		 * `full`    — no width cap; the page manages its own horizontal space
		 *             (Android list screens whose rows bleed to the edge).
		 * `bleed`   — no width cap, content centred in the viewport. For the
		 *             Android auth screens: the root layout already drops the
		 *             shell's padding and footer for them (ANDROID_FULL_BLEED
		 *             in +layout.svelte), and their cards are max-w-88, so they
		 *             need centring rather than a container.
		 */
		width?: 'default' | 'narrow' | 'full' | 'bleed';
		/**
		 * Render the title row. `false` for pages that supply their own header
		 * through a section component (WalletView, ResourceDetail) — they still
		 * want the container, just not a second heading.
		 */
		header?: boolean;
		/**
		 * Stands in for the title row while the page is loading: keeps the header
		 * cell, draws no `<h1>`. A loading screen has no title yet but has the
		 * same shape, and without the cell its skeleton lines drop into the
		 * content slot — below the `aside` instead of beside it.
		 */
		headerPlaceholder?: Snippet;
		/** Passed through to PageHeader: bell + New on the title row. */
		mobileActions?: boolean;
		/** Extra classes on the header cell — e.g. Android hides the title row
		 *  below `sm` while its fixed select-mode app bar replaces it. */
		headerClass?: string;
		/** Trailing controls on the title row. */
		actions?: Snippet;
		/** Back chevron left of the title. */
		onBack?: () => void;
		/**
		 * Back-to-overview link in the eyebrow slot, above the title (the form
		 * pages' text link). A snippet rather than an href, because one caller
		 * navigates through a cancel handler instead of a plain link.
		 */
		back?: Snippet;
		/** Optional while the desktop sections are unmounted (skeleton phase):
		 *  a page may render just the shell and the title row. */
		children?: Snippet;
	} = $props();

	/**
	 * Layout mode. A grid only when there is an `aside` to place beside the
	 * header (the dashboard's stat tiles); every other page is a plain block.
	 *
	 * Deliberately NOT a column gap: the vertical rhythm comes from the
	 * children's own margins (PageHeader's platform-specific mb, the back
	 * line's mb-8). A shell-level gap would force one distance between EVERY
	 * pair of top-level children — the Android notification rows, a section
	 * label and its list — and cannot express the native 20px vs desktop 32px
	 * header gap the mockups specify.
	 */
	const RHYTHM = $derived(
		aside ? 'grid grid-cols-1 gap-x-4 lg:grid-cols-[1fr_auto]' : ''
	);

	// `default` adds nothing: the shell's <main> is already the shared
	// container, and re-stating it here would recreate the very duplication
	// this component removes. `full` lives on the CONTENT wrapper below, not
	// here — the title row always keeps the shell gutter, only the page's
	// area bleeds to the edge.
	const WIDTH_CLASS: Record<typeof width, string> = {
		default: '',
		narrow: 'mx-auto w-full max-w-[680px]',
		full: '',
		// px-5 is the auth cards' own gutter, identical across all five screens.
		bleed: 'flex min-h-dvh items-center justify-center px-5'
	};
	const FULL_BLEED = '-mx-4 sm:-mx-6 lg:-mx-8';
</script>

<!-- `data-page-shell` marks this wrapper as the shared shell container, so the
     structure suite can tell it apart from a page re-stating its own. -->
<div class="{RHYTHM} {WIDTH_CLASS[width]}" data-page-shell={width}>
	{#if headerPlaceholder}
		<div class="min-w-0 lg:col-start-1 lg:row-start-1 {headerClass}">
			{@render headerPlaceholder()}
		</div>
	{:else if header && title}
		<div class="min-w-0 lg:col-start-1 lg:row-start-1 {headerClass}">
			<PageHeader
				{title}
				{eyebrow}
				{eyebrowAside}
				eyebrowSnippet={back}
				{eyebrowVerbatim}
				{actions}
				{onBack}
				{mobileActions}
			/>
			{#if subtitle}
				<!-- Sits under the title inside PageHeader's own bottom gap, so the
				     spacing stays a property of the header, not of each page. -->
				<p class="-mt-4 mb-8 text-text-subtle">{subtitle}</p>
			{/if}
		</div>
	{/if}
	{#if aside}
		<!-- Wrapped only when there is an `aside` to place: without one the shell
		     is a plain block and the content stays a direct child, so pages that
		     rely on that keep working unchanged. `order`: content before the
		     aside on a phone, aside beside the header from `lg` up. -->
		<div class="order-2 lg:order-none lg:col-span-2 lg:row-start-2">
			{@render children?.()}
		</div>
		<div
			class="order-3 mt-6 lg:order-none lg:col-start-2 lg:row-start-1 lg:mt-0"
		>
			{@render aside()}
		</div>
	{:else if width === 'full'}
		<div class={FULL_BLEED}>
			{@render children?.()}
		</div>
	{:else}
		{@render children?.()}
	{/if}
</div>
