<script lang="ts">
	import type { Snippet } from 'svelte';

	/**
	 * Full-screen state page: offline, 404, server error, back-online.
	 *
	 * These screens render outside the app shell — no nav, no container, no
	 * page title — so they are not PageShell's job. What they do share is the
	 * frame: a centred card with a coloured icon disc, a heading, a line or
	 * two of explanation and a pair of actions. That frame was copied
	 * character-for-character between +error.svelte and offline/+page.svelte,
	 * including the icon markup and both buttons.
	 *
	 * Tone picks the icon disc's colour. The icon path itself is passed in,
	 * because each state has its own glyph.
	 */
	let {
		tone = 'accent',
		icon,
		title,
		description,
		hint,
		children,
		actions
	}: {
		/** Colour of the icon disc: what kind of state this is. */
		tone?: 'accent' | 'warning' | 'danger' | 'success';
		/** SVG path data for the icon (24×24 viewBox, stroked). */
		icon: string;
		/** The screen's single `<h1>`. */
		title: string;
		/** Main explanatory line. */
		description?: string;
		/** Secondary line in smaller, quieter type. */
		hint?: string;
		/** Extra content between the text and the actions (offline: the
		 *  what-still-works box). */
		children?: Snippet;
		/** The buttons. Stacked with consistent spacing. */
		actions?: Snippet;
	} = $props();

	const DISC: Record<typeof tone, string> = {
		accent: 'bg-accent-100 text-accent',
		warning: 'bg-warning-100 text-warning-600',
		danger: 'bg-danger-100 text-danger-600',
		success: 'bg-success-100 text-success-600'
	};
</script>

<div
	class="min-h-screen bg-gradient-to-br from-surface-1 to-border-soft flex items-center justify-center px-4"
>
	<div class="max-w-md w-full bg-white rounded-2xl shadow-xl p-8">
		<div class="text-center">
			<div
				class="w-20 h-20 {DISC[
					tone
				]} rounded-full flex items-center justify-center mx-auto mb-4"
			>
				<svg
					class="w-10 h-10"
					fill="none"
					stroke="currentColor"
					viewBox="0 0 24 24"
					aria-hidden="true"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d={icon}
					></path>
				</svg>
			</div>
			<h1 class="text-2xl font-bold text-text mb-2">{title}</h1>
			{#if description}
				<p class="text-text-muted {hint ? 'mb-2' : 'mb-6'}">{description}</p>
			{/if}
			{#if hint}
				<p class="text-text-subtle text-sm mb-6">{hint}</p>
			{/if}
		</div>

		{#if children}
			{@render children()}
		{/if}

		{#if actions}
			<div class="space-y-3">
				{@render actions()}
			</div>
		{/if}
	</div>
</div>
