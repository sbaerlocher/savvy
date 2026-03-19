<script lang="ts">
	interface Props {
		title: string;
		items: string[];
		note?: string;
	}

	let { title, items, note }: Props = $props();
</script>

<div class="bg-white border border-cyan-200 rounded-lg p-3">
	<h4 class="font-medium text-cyan-900 text-sm mb-2">{title}</h4>
	<ul class="text-xs text-cyan-800 space-y-1">
		{#each items as item}
			<li>{item}</li>
		{/each}
	</ul>
	{#if note}
		<p class="text-xs text-cyan-700 mt-2 italic">{note}</p>
	{/if}
</div>
