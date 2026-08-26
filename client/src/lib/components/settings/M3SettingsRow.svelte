<script lang="ts">
	import type { Snippet } from 'svelte';

	interface Props {
		/** Stroke path(s) for the 24x24 leading glyph. Omitted when `avatar` is set. */
		icon?: string;
		/** Initials rendered instead of a glyph (the mockup's name row). */
		avatar?: string;
		title: string;
		subtitle?: string;
		/** Danger treatment: tinted circle, danger ink on both text lines. */
		danger?: boolean;
		/** Turns the row into a button. Non-interactive rows stay a plain div. */
		onclick?: () => void;
		disabled?: boolean;
		/** Trailing control (chevron, toggle, badge). */
		trailing?: Snippet;
	}

	let {
		icon,
		avatar,
		title,
		subtitle,
		danger = false,
		onclick,
		disabled = false,
		trailing
	}: Props = $props();

	// M3 list item: 40px leading circle, 15px title over 13px subtitle, the
	// trailing control flush right (mockup screen-SettingsAndroid).
	const ROW = 'flex w-full items-center gap-4 px-6 py-3 text-left';
	const CIRCLE =
		'flex h-10 w-10 shrink-0 items-center justify-center rounded-m3-full';
	const press = $derived(
		danger ? 'active:bg-danger-50' : 'active:bg-ground-active'
	);
	const circleTone = $derived(
		danger
			? 'bg-danger-50 text-danger-600'
			: avatar
				? 'bg-accent-100 text-accent-850 text-subheading'
				: 'bg-tile-tint text-text-muted'
	);
</script>

{#snippet body()}
	<span class="{CIRCLE} {circleTone}">
		{#if avatar}
			{avatar}
		{:else}
			<svg
				class="h-4.75 w-4.75"
				viewBox="0 0 24 24"
				fill="none"
				stroke="currentColor"
				stroke-width="2"
				stroke-linecap="round"
				stroke-linejoin="round"
				aria-hidden="true"
			>
				<path d={icon} />
			</svg>
		{/if}
	</span>
	<span class="min-w-0 flex-1">
		<span
			class="text-subheading block font-normal {danger
				? 'text-danger-600'
				: 'text-text'}">{title}</span
		>
		{#if subtitle}
			<span
				class="mt-px block text-label font-normal {danger
					? 'text-danger-400'
					: 'text-text-muted'}">{subtitle}</span
			>
		{/if}
	</span>
	{#if trailing}
		{@render trailing()}
	{/if}
{/snippet}

{#if onclick}
	<button type="button" {onclick} {disabled} class="{ROW} {press}">
		{@render body()}
	</button>
{:else}
	<div class="{ROW} {press}">
		{@render body()}
	</div>
{/if}
