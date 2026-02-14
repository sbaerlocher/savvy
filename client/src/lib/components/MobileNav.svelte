<script lang="ts">
	import { page } from '$app/stores';
	import { isOnline } from '$lib/stores/offline';
	import { t } from '$lib/stores/i18n';
	import { platform } from '$lib/utils/platform';
	import { get } from 'svelte/store';

	const tr = (key: string) => get(t)(key);

	const isActive = (path: string) => $page.url.pathname.startsWith(path);
	const preloadStrategy = $derived($isOnline ? 'hover' : 'off');

	const navClass = $derived.by(() => {
		switch (platform) {
			case 'ios':
				// Apple Glass: floating frosted glass pill
				return 'sm:hidden fixed bottom-0 left-0 right-0 mx-2 rounded-full bg-white/70 backdrop-blur-xl backdrop-saturate-150 border border-white/40 shadow-lg z-50 mobile-nav mobile-nav-floating';
			case 'android':
				// Material Design 3: elevated surface, edge-to-edge
				return 'sm:hidden fixed bottom-0 left-0 right-0 bg-[#FFFBFE] border-t border-[#CAC4D0] shadow-[0_-2px_6px_rgba(0,0,0,0.08)] z-50 mobile-nav';
			default:
				// Desktop/other: clean neutral style
				return 'sm:hidden fixed bottom-0 left-0 right-0 bg-white/70 backdrop-blur-xl backdrop-saturate-150 border-t border-white/40 shadow-lg z-50 mobile-nav';
		}
	});

	const linkClass = (path: string) => {
		const base =
			'flex flex-col items-center justify-center transition-colors relative';

		if (!isActive(path)) {
			return `${base} text-gray-600`;
		}

		if (platform === 'android') {
			// Material Design 3: active item with pill indicator
			return `${base} text-cyan-700 font-semibold`;
		}

		// iOS / Default: Liquid Glass active state
		return `${base} text-cyan-600 font-semibold`;
	};
</script>

<nav class={navClass} style="-webkit-tap-highlight-color: transparent;">
	<div class="grid grid-cols-4 h-16">
		<a
			href="/cards"
			data-sveltekit-preload-data={preloadStrategy}
			class={linkClass('/cards')}
		>
			{#if platform === 'android' && isActive('/cards')}
				<span
					class="absolute -inset-x-1.5 top-1 bottom-1 bg-cyan-100 rounded-full -z-10"
				></span>
			{:else if platform !== 'android' && isActive('/cards')}
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
					d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z"
				></path>
			</svg>
			<span class="text-[10px] leading-tight mt-1">{tr('nav.cards')}</span>
		</a>

		<a
			href="/vouchers"
			data-sveltekit-preload-data={preloadStrategy}
			class={linkClass('/vouchers')}
		>
			{#if platform === 'android' && isActive('/vouchers')}
				<span
					class="absolute -inset-x-1.5 top-1 bottom-1 bg-cyan-100 rounded-full -z-10"
				></span>
			{:else if platform !== 'android' && isActive('/vouchers')}
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
					d="M15 5v2m0 4v2m0 4v2M5 5a2 2 0 00-2 2v3a2 2 0 110 4v3a2 2 0 002 2h14a2 2 0 002-2v-3a2 2 0 110-4V7a2 2 0 00-2-2H5z"
				></path>
			</svg>
			<span class="text-[10px] leading-tight mt-1">{tr('nav.vouchers')}</span>
		</a>

		<a
			href="/gift-cards"
			data-sveltekit-preload-data={preloadStrategy}
			class={linkClass('/gift-cards')}
		>
			{#if platform === 'android' && isActive('/gift-cards')}
				<span
					class="absolute -inset-x-1.5 top-1 bottom-1 bg-cyan-100 rounded-full -z-10"
				></span>
			{:else if platform !== 'android' && isActive('/gift-cards')}
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
					d="M12 8v13m0-13V6a2 2 0 112 2h-2zm0 0V5.5A2.5 2.5 0 109.5 8H12zm-7 4h14M5 12a2 2 0 110-4h14a2 2 0 110 4M5 12v7a2 2 0 002 2h10a2 2 0 002-2v-7"
				></path>
			</svg>
			<span class="text-[10px] leading-tight mt-1">{tr('nav.giftCards')}</span>
		</a>

		<a
			href="/merchants"
			data-sveltekit-preload-data={preloadStrategy}
			class={linkClass('/merchants')}
		>
			{#if platform === 'android' && isActive('/merchants')}
				<span
					class="absolute -inset-x-1.5 top-1 bottom-1 bg-cyan-100 rounded-full -z-10"
				></span>
			{:else if platform !== 'android' && isActive('/merchants')}
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
					d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4"
				></path>
			</svg>
			<span class="text-[10px] leading-tight mt-1">{tr('nav.merchants')}</span>
		</a>
	</div>
</nav>
