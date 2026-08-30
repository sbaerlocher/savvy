<script lang="ts">
	import { platform } from '$lib/utils/platform';
	import type { Snippet } from 'svelte';

	/**
	 * The uppercase kicker above a group of rows — "An der Kasse", "NEU",
	 * "FRÜHER", a settings group title, a filter group caption.
	 *
	 * Pages used to carry their own copies: four typography spellings for the
	 * same 11px/600 uppercase step (`text-eyebrow`, `text-section-eyebrow`,
	 * `text-body-sm font-semibold`, a raw `text-[length:var(--text-eyebrow)]
	 * font-semibold tracking-wider`) and eight different gaps below it
	 * (mb-1.5 / mb-2 / mb-2.5 / mb-3.5 / pb-1.5 / pb-2 …). The label is the
	 * same thing on every screen, so it gets one definition here — the same
	 * move PageHeader makes for the title row.
	 *
	 * `--text-section-eyebrow` is the token for exactly this kicker
	 * (tokens.css), so it is the one spelling used, on every platform.
	 */
	let {
		inset = false,
		spaced = false,
		id,
		size = 'eyebrow',
		children
	}: {
		/**
		 * Sits above a list that carries its own horizontal gutter rather than
		 * above full-width content, so the label lines up with the rows it
		 * captions. The gutter is the list's, not the label's: iOS grouped-inset
		 * cards indent 6px, the Android full-bleed rows 20px (`pl-5` on the row
		 * itself) — hence a platform value here, not one fixed number.
		 */
		inset?: boolean;
		/** Separates this group from one above it, rather than opening the page. */
		spaced?: boolean;
		/** Set when a group's controls reference the label via aria-labelledby. */
		id?: string;
		/** Type step. `label` matches the wallet's M3 filter chips (13px) — the
		 *  Android dashboard kicker uses it so the two screens line up. */
		size?: 'eyebrow' | 'label';
		children: Snippet;
	} = $props();

	// One gap, on every screen. The pages used to carry eight values for the
	// same kicker (mb-1.5 / mb-2 / mb-2.5 / mb-3.5 / pb-1.5 / pb-2 …), each
	// traced to a different mockup board. A label above a list is the same
	// element everywhere, so it keeps the same gap everywhere.
	const GAP = 'mb-2.5';
	// The gutter belongs to the list, not to the label: the kicker lines up
	// with the rows it captions. The iOS grouped-inset cards indent 6px, the
	// Android full-bleed rows 20px (`pl-5` on the row itself).
	const INSET = $derived(
		inset ? (platform === 'android' ? 'px-5' : 'px-1.5') : ''
	);
	const SPACED = $derived(spaced ? 'mt-6' : '');
</script>

<p
	{id}
	class="{GAP} {INSET} {SPACED} {size === 'label'
		? 'text-label'
		: 'text-section-eyebrow'} uppercase text-text-subtle"
>
	{@render children()}
</p>
