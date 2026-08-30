<script lang="ts">
	import { resolve } from '$app/paths';
	import { get } from 'svelte/store';
	import { t } from '$lib/stores/i18n';

	import PageShell from '$lib/components/layout/PageShell.svelte';
	import Section from './Section.svelte';

	// Svelte 5 compatible translation wrapper
	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);
</script>

<svelte:head>
	<title>{tr('cards.newCard')} - {tr('common.appName')}</title>
</svelte:head>

<PageShell title={tr('cards.newCard')} mobileActions={false}>
	<Section />
	{#snippet back()}
		<a href={resolve('/cards')} class="text-text-subtle hover:text-text-ink2"
			>{tr('common.backToOverview')}</a
		>
	{/snippet}
</PageShell>
