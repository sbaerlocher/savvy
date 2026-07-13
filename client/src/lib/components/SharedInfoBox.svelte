<script lang="ts">
	interface Props {
		title: string;
		items: string[];
		note?: string;
	}

	let { title, items, note }: Props = $props();
</script>

<div class="bg-white border border-accent-200 rounded-lg p-3">
	<h4 class="font-medium text-accent-900 text-sm mb-2">{title}</h4>
	<ul class="text-xs text-accent-800 space-y-1">
		{#each items as item (item)}
			<li>{item}</li>
		{/each}
	</ul>
	{#if note}
		<p class="text-xs text-accent-hover mt-2 italic">{note}</p>
	{/if}
</div>
