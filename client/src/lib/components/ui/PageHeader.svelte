<script lang="ts">
	import type { Snippet } from 'svelte';
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
</script>

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
				<MobileHeaderActions />
			{/if}
		</div>
	{/if}
</div>
