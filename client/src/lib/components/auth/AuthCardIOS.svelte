<script lang="ts">
	import type { Snippet } from 'svelte';

	// iOS grouped-inset auth card (mockup screen-PasswordResetIOS, frames 1-5):
	// solid card on the system background, accent icon tile next to the title.
	// Frames differ only in the card body, so the chrome lives here once.
	let {
		title,
		compact = false,
		children
	}: {
		title: string;
		// Success/error frames pair a 14px title gap with 28px card bottom
		// padding; the form frames use 18px and 24px.
		compact?: boolean;
		children: Snippet;
	} = $props();
</script>

<!-- The app shell pads `main` by 16px; the mockup insets this screen by 22px
     from the device edge, so cancel the shell padding before applying ours. -->
<div
	class="-mx-4 flex min-h-[calc(100dvh-var(--spacing-page-y))] items-center justify-center px-[var(--spacing-card)] sm:-mx-6 lg:-mx-8"
>
	<div
		class="w-full max-w-[344px] rounded-inset bg-surface px-6 pt-[26px] shadow-card {compact
			? 'pb-7'
			: 'pb-6'}"
	>
		<div
			class="flex items-center gap-[11px] {compact ? 'mb-3.5' : 'mb-[18px]'}"
		>
			<span
				class="flex h-[38px] w-[38px] shrink-0 items-center justify-center rounded-lg bg-accent-600 shadow-accent"
			>
				<svg
					width="21"
					height="21"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					class="text-on-accent"
					stroke-width="2"
					stroke-linecap="round"
					stroke-linejoin="round"
					aria-hidden="true"
				>
					<rect x="2" y="5" width="20" height="14" rx="3"></rect>
					<path d="M2 10h20"></path>
					<path d="M6 15h4"></path>
				</svg>
			</span>
			<h1 class="text-heading text-text">{title}</h1>
		</div>

		{@render children()}
	</div>
</div>
