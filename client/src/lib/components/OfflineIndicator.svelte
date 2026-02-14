<script lang="ts">
	import { t } from '$lib/stores/i18n';
	import { isOnline as isOnlineStore } from '$lib/stores/offline';
	import { slide, fade } from 'svelte/transition';
	import { quintOut } from 'svelte/easing';

	let showBanner = $state(false);
	let bannerType = $state<'warning' | 'success'>('warning');
	let wasOffline = $state(false);

	const online = $derived($isOnlineStore);

	$effect(() => {
		if (!online) {
			showBanner = true;
			bannerType = 'warning';
			wasOffline = true;
		} else if (wasOffline) {
			bannerType = 'success';
			showBanner = true;
			const timeout = setTimeout(() => {
				showBanner = false;
				wasOffline = false;
			}, 3000);

			return () => clearTimeout(timeout);
		}
	});

	function dismiss() {
		showBanner = false;
	}
</script>

{#if showBanner}
	<div
		data-testid="offline-indicator"
		in:slide={{ duration: 300, easing: quintOut }}
		out:fade={{ duration: 200 }}
		class="fixed top-0 left-0 right-0 z-[100] px-4 py-3 border-b"
		class:bg-yellow-50={bannerType === 'warning'}
		class:bg-green-50={bannerType === 'success'}
		class:border-yellow-200={bannerType === 'warning'}
		class:border-green-200={bannerType === 'success'}
	>
		<div class="container mx-auto flex items-center justify-between">
			<div class="flex items-center gap-2">
				{#if bannerType === 'warning'}
					<svg
						class="h-4 w-4 text-yellow-600 shrink-0"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
						/>
					</svg>
				{:else}
					<svg
						class="h-4 w-4 text-green-600 shrink-0"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
						/>
					</svg>
				{/if}

				<span
					class="text-sm font-medium"
					class:text-yellow-800={bannerType === 'warning'}
					class:text-green-800={bannerType === 'success'}
				>
					{#if bannerType === 'warning'}
						{$t('common.offlineBannerMessage')}
					{:else}
						{$t('common.onlineAgainMessage')}
					{/if}
				</span>
			</div>

			{#if bannerType === 'warning'}
				<button
					onclick={dismiss}
					class="text-yellow-600 hover:text-yellow-800 p-1 rounded-md hover:bg-yellow-100 transition-colors"
					aria-label="Close"
				>
					<svg
						class="h-4 w-4"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M6 18L18 6M6 6l12 12"
						/>
					</svg>
				</button>
			{/if}
		</div>
	</div>
{/if}
