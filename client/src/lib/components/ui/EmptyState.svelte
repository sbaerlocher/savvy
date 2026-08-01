<script lang="ts">
	import type { Snippet } from 'svelte';

	// Centered empty-state block for "no items" screens. Replaces the ad-hoc
	// `<div class="text-center py-… text-text-subtle">…</div>` repeated per route.
	// `icon` is a heroicons `d` path string (from $lib/icons); the <svg> wrapper
	// stays here so call sites pass only the path, matching repo icon idiom.
	let {
		icon = undefined as string | undefined,
		title,
		description = undefined as string | undefined,
		action = undefined as Snippet | undefined
	}: {
		icon?: string;
		title: string;
		description?: string;
		action?: Snippet;
	} = $props();
</script>

<div class="flex flex-col items-center py-12 text-center text-text-subtle">
	{#if icon}
		<svg
			class="mb-3 h-10 w-10 text-text-faint"
			viewBox="0 0 24 24"
			fill="none"
			stroke="currentColor"
			stroke-width="1.5"
			aria-hidden="true"
		>
			<path stroke-linecap="round" stroke-linejoin="round" d={icon} />
		</svg>
	{/if}
	<p class="text-sm font-medium text-text-ink2">{title}</p>
	{#if description}
		<p class="mt-1 max-w-sm text-sm">{description}</p>
	{/if}
	{#if action}
		<div class="mt-4">{@render action()}</div>
	{/if}
</div>
