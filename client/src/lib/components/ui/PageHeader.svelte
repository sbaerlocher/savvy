<script lang="ts">
	import { ICON_CLOSE, ICON_SEARCH } from '$lib/icons';
	import type { Snippet } from 'svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/stores';
	import { t } from '$lib/stores/i18n';
	import { platform } from '$lib/utils/platform';
	import { get } from 'svelte/store';
	import MobileHeaderActions from './MobileHeaderActions.svelte';

	let {
		title,
		eyebrow,
		eyebrowAside,
		eyebrowVerbatim = false,
		actions,
		mobileActions = true,
		showSearch = false,
		onBack
	}: {
		/** Main heading, e.g. "Deine Favoriten". */
		title: string;
		/** Small line above the title, e.g. a greeting "Hallo Anna". */
		eyebrow?: string;
		/** Inline element next to the eyebrow (Android and iOS mockups put the
		 *  refresh indicator there). */
		eyebrowAside?: Snippet;
		/** Set when the eyebrow carries user-entered text (a card program name,
		 *  say). The native uppercase treatment is meant for fixed kickers and
		 *  would case-mangle whatever the user typed. */
		eyebrowVerbatim?: boolean;
		/** Optional trailing controls (buttons, links) rendered on the right. */
		actions?: Snippet;
		/** Render the mobile header actions (bell + New) on the title row. Top-level
		 *  screens keep this on; sub-screens without them can pass false. */
		mobileActions?: boolean;
		/** Show a standalone search icon (Android) next to custom actions — used by
		 *  detail pages that keep search but not the full mobile header actions. */
		showSearch?: boolean;
		/** When set, render a compact back chevron left of the title (detail pages). */
		onBack?: () => void;
	} = $props();

	const tr = (key: string) => get(t)(key);

	// Header chrome per platform (dashboard/wallet mockups). All three use the
	// uppercase 11px eyebrow; they differ in title step (Android ~26px
	// --text-title, iOS 28px --text-screen-title, desktop --text-title), bottom
	// gap (native 20px, desktop 32px) and action spacing (Android 4px between
	// round buttons, iOS 20px between bare glyphs). `platform` is a module
	// constant, so plain consts, not $derived.
	const NATIVE = platform === 'android' || platform === 'ios';
	const HEADER_MB = NATIVE ? 'mb-5' : 'mb-8';
	const EYEBROW_BASE = 'text-eyebrow text-text-subtle';
	// Uppercase is the kicker treatment; user-entered eyebrows opt out.
	const eyebrowClass = $derived(
		eyebrowVerbatim ? EYEBROW_BASE : `${EYEBROW_BASE} uppercase`
	);
	const TITLE_CLASS =
		platform === 'ios'
			? 'text-screen-title text-text'
			: 'mt-0.5 text-title text-text';
	const ACTIONS_GAP =
		platform === 'android' ? 'gap-1' : platform === 'ios' ? 'gap-0' : 'gap-2.5';

	// Android M3: the header search icon expands an inline docked search field
	// that replaces the title row. Typing drives /wallet?search=<query> so the
	// wallet list filters reactively (same flow as the iOS bottom-nav search).
	// Active-state is derived from the ?search param, not a local flag — that
	// way it survives the navigation/remount when the query changes.
	// Inline header search is the Android pattern; iOS uses the bottom-nav pill.
	const searchParam = $derived($page.url.searchParams.get('search'));
	const searchActive = $derived(platform === 'android' && searchParam !== null);
	let searchValue = $state('');
	let searchEl = $state<HTMLInputElement | null>(null);
	let searchDebounce: ReturnType<typeof setTimeout> | null = null;

	// Keep the input text in sync with the param (prefill on open / navigation).
	$effect(() => {
		if (searchParam && searchParam !== '1') searchValue = searchParam;
	});

	$effect(() => {
		if (searchActive && searchEl) searchEl.focus();
	});

	function openSearch() {
		// eslint-disable-next-line svelte/no-navigation-without-resolve -- base path is resolve()d; ?search is a query string
		goto(resolve('/wallet') + '?search=1', { replaceState: true });
	}

	function closeSearch() {
		searchValue = '';
		if (searchDebounce) clearTimeout(searchDebounce);
		if ($page.url.pathname.startsWith('/wallet')) goto(resolve('/wallet'));
	}

	function onSearchInput() {
		if (searchDebounce) clearTimeout(searchDebounce);
		const q = searchValue.trim();
		searchDebounce = setTimeout(() => {
			goto(
				// eslint-disable-next-line svelte/no-navigation-without-resolve -- base path is resolve()d; ?search is a query string
				resolve('/wallet') +
					(q ? `?search=${encodeURIComponent(q)}` : '?search=1'),
				{ keepFocus: true, replaceState: true }
			);
		}, 250);
	}
</script>

{#if searchActive}
	<!-- Inline M3 docked search field replacing the header row (Android). -->
	<div
		class="{HEADER_MB} flex h-12 items-center gap-3 rounded-full border border-border bg-white px-4"
	>
		<svg
			class="h-5 w-5 shrink-0 text-text-muted"
			fill="none"
			stroke="currentColor"
			stroke-width="2"
			viewBox="0 0 24 24"
		>
			<path stroke-linecap="round" stroke-linejoin="round" d={ICON_SEARCH} />
		</svg>
		<input
			bind:this={searchEl}
			bind:value={searchValue}
			oninput={onSearchInput}
			type="search"
			enterkeyhint="search"
			placeholder={tr('common.search')}
			aria-label={tr('common.search')}
			class="min-w-0 flex-1 bg-transparent text-text placeholder:text-text-subtle focus:outline-none"
		/>
		<button
			type="button"
			onclick={closeSearch}
			aria-label={tr('common.cancel')}
			class="shrink-0 text-text-muted hover:text-text-strong"
		>
			<svg
				class="h-5 w-5"
				fill="none"
				stroke="currentColor"
				stroke-width="2"
				viewBox="0 0 24 24"
			>
				<path stroke-linecap="round" stroke-linejoin="round" d={ICON_CLOSE} />
			</svg>
		</button>
	</div>
{:else}
	<!-- Page header: plain type hierarchy, no left accent bar (mockup). -->
	<div class="{HEADER_MB} flex items-start justify-between gap-4">
		<div class="flex min-w-0 items-center gap-2">
			{#if onBack}
				<button
					type="button"
					onclick={onBack}
					aria-label={tr('common.back')}
					class="-ml-1 inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-xl text-text-muted transition-colors hover:bg-surface-1"
				>
					<svg
						class="h-6 w-6"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							d="M15 19l-7-7 7-7"
						/>
					</svg>
				</button>
			{/if}
			<div class="min-w-0">
				{#if eyebrow}
					<!-- The native mockups put the greeting and the refresh hint on
					     one uppercase eyebrow row above the title. -->
					<div class="flex items-center gap-2">
						<p class={eyebrowClass}>{eyebrow}</p>
						{#if eyebrowAside}
							{@render eyebrowAside()}
						{/if}
					</div>
				{/if}
				<h1 class={TITLE_CLASS}>{title}</h1>
			</div>
		</div>
		{#if actions || mobileActions || (showSearch && platform === 'android')}
			<div class="flex shrink-0 items-center {ACTIONS_GAP}">
				{#if showSearch && platform === 'android'}
					<button
						type="button"
						onclick={openSearch}
						aria-label={tr('common.search')}
						class="inline-flex h-11 w-11 items-center justify-center rounded-full text-text-muted transition-colors hover:bg-surface-1"
					>
						<svg
							class="h-5.5 w-5.5"
							fill="none"
							stroke="currentColor"
							stroke-width="2"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								d={ICON_SEARCH}
							/>
						</svg>
					</button>
				{/if}
				{#if actions}
					{@render actions()}
				{/if}
				{#if mobileActions}
					<MobileHeaderActions onSearchOpen={openSearch} />
				{/if}
			</div>
		{/if}
	</div>
{/if}
