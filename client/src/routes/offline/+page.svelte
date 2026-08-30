<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import { languageStore, t } from '$lib/stores/i18n';
	import { translations } from '$lib/i18n';
	import Button from '$lib/components/ui/Button.svelte';
	import StateScreen from '$lib/components/layout/StateScreen.svelte';
	import { ICON_CHECK, ICON_INFO_CIRCLE, ICON_WIFI_OFF } from '$lib/icons';

	let retrying = $state(false);
	let isOnline = $state(false);

	// This page used to carry a hand-written German string table to avoid
	// depending on $t while offline. That dependency is safe: src/lib/i18n
	// imports all three locales statically, so they ship inside the bundle and
	// need no network. The table meant the one page a French or English user
	// sees when their connection drops was the only German one in the app.

	// `$t` resolves string keys only — it warns and returns the key for
	// anything else — so the bullet list is read off the translation table
	// directly rather than widening the store for one call site.
	const availableItems = $derived(
		translations[$languageStore].errors.offlinePage.availableItems
	);

	function handleRetry() {
		if (navigator.onLine) {
			window.location.href = '/';
		} else {
			retrying = true;
			setTimeout(() => {
				retrying = false;
			}, 2000);
		}
	}

	onMount(() => {
		// Check initial online status
		isOnline = navigator.onLine;

		// Auto-redirect when back online
		const handleOnline = () => {
			isOnline = true;
			setTimeout(() => {
				window.location.href = '/';
			}, 1000);
		};

		const handleOffline = () => {
			isOnline = false;
		};

		window.addEventListener('online', handleOnline);
		window.addEventListener('offline', handleOffline);

		return () => {
			window.removeEventListener('online', handleOnline);
			window.removeEventListener('offline', handleOffline);
		};
	});
</script>

<svelte:head>
	<title
		>{isOnline
			? $t('errors.offlinePage.backOnlineTitle')
			: $t('errors.offline.title')} - {$t('common.appName')}</title
	>
</svelte:head>

{#if isOnline}
	<StateScreen
		tone="success"
		icon={ICON_CHECK}
		title={$t('errors.offlinePage.backOnlineTitle')}
		description={$t('errors.offlinePage.backOnlineMessage')}
	/>
{:else}
	<StateScreen
		tone="warning"
		icon={ICON_WIFI_OFF}
		title={$t('errors.offline.title')}
		description={$t('errors.offline.description')}
		hint={$t('errors.offline.hint')}
	>
		<!-- What still works without a connection, so the page is more than an
		     apology. -->
		<div
			class="bg-accent-50 border border-accent-200 rounded-lg p-4 mb-6 text-left"
		>
			<div class="flex items-start gap-3">
				<svg
					class="w-5 h-5 text-accent flex-shrink-0 mt-0.5"
					fill="none"
					stroke="currentColor"
					viewBox="0 0 24 24"
					aria-hidden="true"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d={ICON_INFO_CIRCLE}
					></path>
				</svg>
				<div class="text-sm text-accent-800">
					<p class="font-semibold mb-1">
						{$t('errors.offlinePage.availableTitle')}
					</p>
					<ul class="list-disc list-inside space-y-1">
						{#each availableItems as feature (feature)}
							<li>{feature}</li>
						{/each}
					</ul>
				</div>
			</div>
		</div>

		{#snippet actions()}
			<Button onclick={handleRetry} loading={retrying} class="w-full">
				{retrying ? $t('errors.offlinePage.retrying') : $t('errors.retry')}
			</Button>
			<Button variant="secondary" href={resolve('/')} class="w-full">
				{$t('errors.offlinePage.goHome')}
			</Button>
		{/snippet}
	</StateScreen>
{/if}
