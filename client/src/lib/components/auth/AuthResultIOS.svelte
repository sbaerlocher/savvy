<script lang="ts">
	import type { Snippet } from 'svelte';

	// Centred result block inside the iOS auth card (mockup frames 2, 4, 5):
	// 64px tinted circle, headline, muted copy, then the actions.
	let {
		tone,
		icon,
		heading,
		message,
		children
	}: {
		tone: 'success' | 'danger';
		// The mockup pairs a different glyph with each frame: envelope for
		// "mail sent" (frame 2), check for "password reset" (frame 4), cross
		// for the token error (frame 5).
		icon: 'envelope' | 'check' | 'cross';
		heading: string;
		message: string;
		children: Snippet;
	} = $props();
</script>

<div class="pt-card text-center">
	<span
		class="mb-4 inline-flex h-16 w-16 items-center justify-center rounded-full {tone ===
		'success'
			? 'bg-success-100'
			: 'bg-danger-100'}"
	>
		<svg
			width="32"
			height="32"
			viewBox="0 0 24 24"
			fill="none"
			stroke="currentColor"
			class={tone === 'success' ? 'text-success-600' : 'text-danger-600'}
			stroke-width="2"
			stroke-linecap="round"
			stroke-linejoin="round"
			aria-hidden="true"
		>
			{#if icon === 'envelope'}
				<path
					d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
				></path>
			{:else if icon === 'check'}
				<path d="M5 13l4 4L19 7"></path>
			{:else}
				<path d="M6 18L18 6M6 6l12 12"></path>
			{/if}
		</svg>
	</span>
	<h2 class="mb-2 text-heading font-semibold text-text">{heading}</h2>
	<p class="mb-6 text-body text-text-muted">{message}</p>
	{@render children()}
</div>
