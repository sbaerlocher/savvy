<script lang="ts">
	import PageShell from '$lib/components/layout/PageShell.svelte';
	import { ICON_FILTER_LINES } from '$lib/icons';
	import { t } from '$lib/stores/i18n';
	import { get } from 'svelte/store';
	import { platform } from '$lib/utils/platform';
	import Section from './Section.svelte';

	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);

	const IS_DESKTOP = platform === 'other';

	// Bound from the Section: the count eyebrow (Android: filtered count, iOS:
	// chrome eyebrow) derives from the list data living there.
	let eyebrow = $state<string | undefined>(undefined);

	// Desktop filter button on the title row. The panel it opens lives in the
	// section — this state gets bound to it when the section is re-attached.
	let filterOpen = $state(false);
	// Applied-filter indicator, bound up from the section (ring + dot).
	let filtersActive = $state(false);
</script>

<svelte:head>
	<title>{tr('merchantOverview.title')} - {tr('common.appName')}</title>
</svelte:head>

<!-- Desktop chrome row action: the lone filter button beside the title (mockup
     board 1A/1B). Search and the type/sort/status groups live in the panel it
     opens, which is part of the section. -->
{#snippet desktopFilterButton()}
	<button
		type="button"
		onclick={(e: MouseEvent) => {
			e.stopPropagation();
			filterOpen = !filterOpen;
		}}
		class="title-action relative {filtersActive
			? 'border-accent text-accent-hover ring-2 ring-accent'
			: 'text-text-muted'}"
		title={tr('common.filter')}
		aria-label={tr('common.filter')}
		aria-expanded={filterOpen}
	>
		<svg
			class="h-5 w-5"
			fill="none"
			stroke="currentColor"
			stroke-width="2"
			viewBox="0 0 24 24"
			aria-hidden="true"
		>
			<path
				stroke-linecap="round"
				stroke-linejoin="round"
				d={ICON_FILTER_LINES}
			/>
		</svg>
		{#if filtersActive}
			<span
				class="absolute -top-1 -right-1 h-3 w-3 rounded-full border-2 border-paper bg-accent"
			></span>
		{/if}
	</button>
{/snippet}

<!-- Header. Android carries the count as an eyebrow (mockup); the mobile
     header actions stay on their default for every platform — on iOS that row
     is the only create entry point on this screen. -->
<PageShell
	title={tr('merchantOverview.title')}
	{eyebrow}
	mobileActions={!IS_DESKTOP}
	actions={IS_DESKTOP ? desktopFilterButton : undefined}
>
	<Section bind:eyebrow bind:filterOpen bind:filtersActive />
</PageShell>
