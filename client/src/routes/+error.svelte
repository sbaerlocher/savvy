<script lang="ts">
	import { page } from '$app/stores';
	import { browser } from '$app/environment';
	import { resolve } from '$app/paths';
	import { t } from '$lib/stores/i18n';
	import Button from '$lib/components/ui/Button.svelte';
	import StateScreen from '$lib/components/layout/StateScreen.svelte';
	import {
		ICON_ALERT_TRIANGLE,
		ICON_FACE_CONFUSED,
		ICON_WIFI_OFF
	} from '$lib/icons';

	const isOffline = $derived(browser && !navigator.onLine);
	const status = $derived($page.status);
	const message = $derived($page.error?.message || '');

	// The three cases differ only in icon, tone and copy — so they pick a
	// variant rather than each rebuilding the screen. Previously each branch
	// carried its own <h1>, which is why this page rendered three of them.
	const variant = $derived(
		isOffline
			? {
					tone: 'warning' as const,
					icon: ICON_WIFI_OFF,
					title: $t('errors.offline.title'),
					description: $t('errors.offline.description'),
					hint: $t('errors.offline.hint')
				}
			: status === 404
				? {
						tone: 'accent' as const,
						icon: ICON_FACE_CONFUSED,
						title: $t('errors.notFound.title'),
						description: message || $t('errors.notFound.description'),
						hint: undefined
					}
				: {
						tone: 'danger' as const,
						icon: ICON_ALERT_TRIANGLE,
						title: $t('errors.generic.title'),
						description: message || $t('errors.generic.description'),
						hint: undefined
					}
	);

	function handleRetry() {
		if (browser) {
			window.location.reload();
		}
	}

	function handleGoBack() {
		if (browser) {
			if (window.history.length > 1) {
				window.history.back();
			} else {
				window.location.href = '/dashboard';
			}
		}
	}
</script>

<StateScreen
	tone={variant.tone}
	icon={variant.icon}
	title={variant.title}
	description={variant.description}
	hint={variant.hint}
>
	{#snippet actions()}
		{#if isOffline}
			<Button onclick={handleGoBack} class="w-full">
				{$t('errors.goBack')}
			</Button>
		{:else}
			<Button onclick={handleRetry} class="w-full">
				{$t('errors.retry')}
			</Button>
		{/if}
		<Button variant="secondary" href={resolve('/dashboard')} class="w-full">
			{$t('errors.toDashboard')}
		</Button>
	{/snippet}
</StateScreen>
