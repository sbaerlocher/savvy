<script lang="ts">
	import { ICON_SPINNER } from '$lib/icons';
	import type { Snippet } from 'svelte';

	// Thin wrapper over the app.css .btn* recipe classes: the CSS stays the
	// single source of truth (no visual diff), the component carries behaviour
	// that today is duplicated inline at every call site — loading→disabled+
	// spinner, and <a>/<button> polymorphism.
	//
	// VARIANT/SIZE map to full literal class strings, never `btn-${variant}`, so
	// the Tailwind JIT scanner sees them (same technique as Modal.svelte).
	const VARIANT = {
		primary: 'btn-primary',
		secondary: 'btn-secondary',
		danger: 'btn-danger',
		success: 'btn-success',
		warning: 'btn-warning',
		purple: 'btn-purple',
		gray: 'btn-gray',
		ghost: 'btn-ghost',
		text: 'btn-text',
		'text-danger': 'btn-text-danger'
	} as const;
	const SIZE = {
		default: '',
		sm: 'btn-sm',
		xs: 'btn-xs'
	} as const;

	let {
		variant = 'primary' as keyof typeof VARIANT,
		size = 'default' as keyof typeof SIZE,
		type = 'button' as 'button' | 'submit' | 'reset',
		href = undefined as string | undefined,
		loading = false,
		disabled = false,
		class: klass = '',
		onclick = undefined as ((e: MouseEvent) => void) | undefined,
		children,
		...rest
	}: {
		variant?: keyof typeof VARIANT;
		size?: keyof typeof SIZE;
		type?: 'button' | 'submit' | 'reset';
		href?: string;
		loading?: boolean;
		disabled?: boolean;
		class?: string;
		onclick?: (e: MouseEvent) => void;
		children: Snippet;
		[key: string]: unknown;
	} = $props();

	const classes = $derived(
		['btn', VARIANT[variant], SIZE[size], klass].filter(Boolean).join(' ')
	);
	// loading implies disabled — a click while a request is in flight is the bug
	// the manual `disabled={isLoading}` at 15 call sites guards against.
	const isDisabled = $derived(disabled || loading);

	// An <a> has no native `disabled`, so aria-disabled alone would leave it
	// clickable/navigable — block the click ourselves to match the <button>
	// semantics the loading guard promises.
	function handleAnchorClick(e: MouseEvent) {
		if (isDisabled) {
			e.preventDefault();
			return;
		}
		onclick?.(e);
	}
</script>

{#if href}
	<!-- eslint-disable svelte/no-navigation-without-resolve -- href is the caller's responsibility to resolve() (generic wrapper) -->
	<a
		{href}
		class={classes}
		aria-disabled={isDisabled}
		aria-busy={loading}
		onclick={handleAnchorClick}
		{...rest}
	>
		{#if loading}
			<svg
				class="h-4 w-4 animate-spin"
				viewBox="0 0 24 24"
				fill="none"
				aria-hidden="true"
			>
				<circle
					class="opacity-25"
					cx="12"
					cy="12"
					r="10"
					stroke="currentColor"
					stroke-width="4"
				/>
				<path class="opacity-75" fill="currentColor" d={ICON_SPINNER} />
			</svg>
		{/if}
		{@render children()}
	</a>
	<!-- eslint-enable svelte/no-navigation-without-resolve -->
{:else}
	<button
		{type}
		class={classes}
		disabled={isDisabled}
		aria-busy={loading}
		{onclick}
		{...rest}
	>
		{#if loading}
			<svg
				class="h-4 w-4 animate-spin"
				viewBox="0 0 24 24"
				fill="none"
				aria-hidden="true"
			>
				<circle
					class="opacity-25"
					cx="12"
					cy="12"
					r="10"
					stroke="currentColor"
					stroke-width="4"
				/>
				<path class="opacity-75" fill="currentColor" d={ICON_SPINNER} />
			</svg>
		{/if}
		{@render children()}
	</button>
{/if}
