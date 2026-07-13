<script lang="ts">
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
		actions,
		mobileActions = true
	}: {
		/** Main heading, e.g. "Deine Favoriten". */
		title: string;
		/** Small line above the title, e.g. a greeting "Hallo Anna". */
		eyebrow?: string;
		/** Optional trailing controls (buttons, links) rendered on the right. */
		actions?: Snippet;
		/** Render the mobile header actions (bell + New) on the title row. Top-level
		 *  screens keep this on; sub-screens without them can pass false. */
		mobileActions?: boolean;
	} = $props();

	const tr = (key: string) => get(t)(key);

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
		// eslint-disable-next-line svelte/no-navigation-without-resolve -- base path is resolve()d
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
		class="mb-8 flex h-12 items-center gap-3 rounded-full border border-border bg-white px-4"
	>
		<svg
			class="h-5 w-5 shrink-0 text-text-muted"
			fill="none"
			stroke="currentColor"
			stroke-width="2"
			viewBox="0 0 24 24"
		>
			<path
				stroke-linecap="round"
				stroke-linejoin="round"
				d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
			/>
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
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					d="M6 18L18 6M6 6l12 12"
				/>
			</svg>
		</button>
	</div>
{:else}
	<!-- Page header: plain type hierarchy, no left accent bar (mockup). -->
	<div class="mb-8 flex items-start justify-between gap-4">
		<div>
			{#if eyebrow}
				<p class="text-sm text-text-subtle">{eyebrow}</p>
			{/if}
			<h1 class="text-3xl font-bold tracking-tight text-text">{title}</h1>
		</div>
		{#if actions || mobileActions}
			<div class="flex shrink-0 items-center gap-2">
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
