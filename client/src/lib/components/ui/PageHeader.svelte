<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		title,
		eyebrow,
		actions
	}: {
		/** Main heading, e.g. "Deine Favoriten". */
		title: string;
		/** Small line above the title, e.g. a greeting "Hallo Anna". */
		eyebrow?: string;
		/** Optional trailing controls (buttons, links) rendered on the right. */
		actions?: Snippet;
	} = $props();
</script>

<!-- Page header: accent bar + type hierarchy, no heavy shadow. -->
<div class="mb-8 flex items-start justify-between gap-4">
	<div class="flex items-stretch gap-3">
		<div
			class="w-1.5 self-stretch rounded-full bg-cyan-600"
			aria-hidden="true"
		></div>
		<div>
			{#if eyebrow}
				<p class="text-sm text-gray-500">{eyebrow}</p>
			{/if}
			<h1 class="text-3xl font-bold text-gray-900">{title}</h1>
		</div>
	</div>
	{#if actions}
		<div class="flex shrink-0 items-center gap-2">
			{@render actions()}
		</div>
	{/if}
</div>
