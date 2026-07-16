<script lang="ts">
	import { ICON_CLOSE, ICON_SEARCH } from '$lib/icons';
	import { page } from '$app/stores';
	import { resolve } from '$app/paths';
	import { goto } from '$app/navigation';
	import { isOnline } from '$lib/stores/offline';
	import { t } from '$lib/stores/i18n';
	import { platform } from '$lib/utils/platform';

	interface Props {
		// Opens the global type-choice ("New") dialog.
		onNew: () => void;
	}

	let { onNew }: Props = $props();

	const isActive = (path: string) => $page.url.pathname.startsWith(path);
	const preloadStrategy = $derived($isOnline ? 'hover' : 'off');

	// Three places: Start (dashboard), Wallet, Profile.
	// href is resolved up-front (resolve() needs literal routes, not a variable).
	const places = [
		{
			path: '/dashboard',
			href: resolve('/dashboard'),
			label: 'nav.start',
			icon: 'M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6'
		},
		{
			path: '/wallet',
			href: resolve('/wallet'),
			label: 'nav.wallet',
			icon: 'M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z'
		},
		{
			path: '/profile',
			href: resolve('/profile'),
			label: 'nav.profile',
			icon: 'M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z'
		}
	];

	// iOS Liquid Glass: the search pill expands inline into a text field (no
	// navigation to a separate search bar). Typing drives /wallet?search=<query>
	// so the wallet list filters reactively; the field itself stays in the nav.
	let searchActive = $state(false);
	let searchValue = $state('');
	let searchEl = $state<HTMLInputElement | null>(null);
	let searchDebounce: ReturnType<typeof setTimeout> | null = null;

	// Prefill from the current ?search param so re-opening keeps the query.
	$effect(() => {
		const param = $page.url.searchParams.get('search');
		if (param && param !== '1') searchValue = param;
	});

	$effect(() => {
		if (searchActive && searchEl) searchEl.focus();
	});

	function openSearch() {
		searchActive = true;
	}

	function closeSearch() {
		searchActive = false;
		searchValue = '';
		if (searchDebounce) clearTimeout(searchDebounce);
		if ($page.url.pathname.startsWith('/wallet')) goto(resolve('/wallet'));
	}

	function onSearchInput() {
		if (searchDebounce) clearTimeout(searchDebounce);
		const q = searchValue.trim();
		searchDebounce = setTimeout(() => {
			const target =
				resolve('/wallet') +
				(q ? `?search=${encodeURIComponent(q)}` : '?search=1');
			// eslint-disable-next-line svelte/no-navigation-without-resolve -- base path is resolve()d; ?search is a query string
			goto(target, {
				keepFocus: true,
				replaceState: true
			});
		}, 250);
	}

	const linkClass = (path: string) => {
		const base =
			'flex flex-col items-center justify-center transition-colors relative';
		if (!isActive(path)) return `${base} text-text-muted`;
		if (platform === 'android')
			return `${base} text-accent-hover font-semibold`;
		return `${base} text-accent font-semibold`;
	};
</script>

{#if platform === 'ios'}
	<!-- iOS 26 tab-bar pattern: two separate floating glass pills (places + search) -->
	<div
		data-testid="mobile-nav"
		class="sm:hidden fixed bottom-0 left-0 right-0 mx-2 z-50 flex items-center gap-2 mobile-nav mobile-nav-floating"
		style="-webkit-tap-highlight-color: transparent;"
	>
		{#if !searchActive}
			<nav
				class="flex-1 grid grid-cols-3 h-16 rounded-full bg-white/70 backdrop-blur-xl backdrop-saturate-150 border border-white/40 shadow-lg"
			>
				{#each places as place (place.path)}
					<!-- eslint-disable svelte/no-navigation-without-resolve -- place.href is produced by resolve() above -->
					<a
						href={place.href}
						data-sveltekit-preload-data={preloadStrategy}
						class={linkClass(place.path)}
					>
						<!-- eslint-enable svelte/no-navigation-without-resolve -->
						{#if isActive(place.path)}
							<span class="liquid-glass-pill"></span>
						{/if}
						<svg
							class="w-6 h-6"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d={place.icon}
							/>
						</svg>
						<span class="text-[length:var(--text-tag)] leading-tight mt-1"
							>{$t(place.label)}</span
						>
					</a>
				{/each}
			</nav>
			<!-- Round search pill (collapsed) -->
			<button
				type="button"
				onclick={openSearch}
				aria-label={$t('common.search')}
				class="h-16 w-16 shrink-0 flex items-center justify-center rounded-full bg-white/70 backdrop-blur-xl backdrop-saturate-150 border border-white/40 shadow-lg text-text-muted"
			>
				<svg
					class="w-6 h-6"
					fill="none"
					stroke="currentColor"
					viewBox="0 0 24 24"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d={ICON_SEARCH}
					/>
				</svg>
			</button>
		{:else}
			<!-- Expanded inline search field (Liquid Glass) -->
			<div
				class="flex-1 flex items-center gap-2 h-16 px-5 rounded-full bg-white/70 backdrop-blur-xl backdrop-saturate-150 border border-white/40 shadow-lg"
			>
				<svg
					class="w-6 h-6 shrink-0 text-text-muted"
					fill="none"
					stroke="currentColor"
					viewBox="0 0 24 24"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d={ICON_SEARCH}
					/>
				</svg>
				<input
					bind:this={searchEl}
					bind:value={searchValue}
					oninput={onSearchInput}
					type="search"
					enterkeyhint="search"
					placeholder={$t('common.search')}
					aria-label={$t('common.search')}
					class="flex-1 min-w-0 bg-transparent text-text placeholder:text-text-subtle focus:outline-none"
				/>
				<button
					type="button"
					onclick={closeSearch}
					aria-label={$t('common.cancel')}
					class="shrink-0 text-text-muted hover:text-text-strong"
				>
					<svg
						class="w-6 h-6"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d={ICON_CLOSE}
						/>
					</svg>
				</button>
			</div>
		{/if}
	</div>
{:else}
	<!-- Android Material 3 edge-to-edge bar + FAB for New. Search lives in top bar. -->
	<button
		type="button"
		onclick={onNew}
		aria-label={$t('common.new')}
		class="sm:hidden fixed right-4 z-50 h-14 w-14 flex items-center justify-center rounded-2xl bg-accent text-white shadow-lg mobile-nav-fab"
	>
		<svg class="w-7 h-7" fill="none" stroke="currentColor" viewBox="0 0 24 24">
			<path
				stroke-linecap="round"
				stroke-linejoin="round"
				stroke-width="2"
				d="M12 4v16m8-8H4"
			/>
		</svg>
	</button>
	<nav
		data-testid="mobile-nav"
		class="sm:hidden fixed bottom-0 left-0 right-0 z-50 mobile-nav {platform ===
		'android'
			? 'bg-surface border-t border-border shadow-[var(--shadow-sheet)]'
			: 'bg-white/70 backdrop-blur-xl backdrop-saturate-150 border-t border-white/40 shadow-lg'}"
		style="-webkit-tap-highlight-color: transparent;"
	>
		<div class="grid grid-cols-3 h-16">
			{#each places as place (place.path)}
				<!-- eslint-disable svelte/no-navigation-without-resolve -- place.href is produced by resolve() above -->
				<a
					href={place.href}
					data-sveltekit-preload-data={preloadStrategy}
					class={linkClass(place.path)}
				>
					<!-- eslint-enable svelte/no-navigation-without-resolve -->
					{#if platform === 'android' && isActive(place.path)}
						<span
							class="absolute -inset-x-1.5 top-1 bottom-1 bg-accent-100 rounded-full -z-10"
						></span>
					{:else if isActive(place.path)}
						<span class="liquid-glass-pill"></span>
					{/if}
					<svg
						class="w-6 h-6"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d={place.icon}
						/>
					</svg>
					<span class="text-[length:var(--text-tag)] leading-tight mt-1"
						>{$t(place.label)}</span
					>
				</a>
			{/each}
		</div>
	</nav>
{/if}
