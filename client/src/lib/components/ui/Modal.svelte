<script lang="ts">
	import { platform } from '$lib/utils/platform';
	import type { Snippet } from 'svelte';

	// Reusable modal shell: backdrop (iOS blur, Android opacity, desktop scrim), mobile
	// bottom-sheet / desktop-centered positioning, Escape + backdrop-click close.
	// Content (icon/title/buttons/etc.) is supplied via the children snippet.
	//
	// z-index is a named layer, not a raw number, so the full Tailwind class
	// strings are statically visible to the JIT scanner. 'default' matches the
	// ConfirmModal stack (backdrop z-55 / panel z-60); 'elevated' the batch/
	// import stack (z-70 / z-80). Add layers here when a later modal needs one.
	const LAYERS = {
		default: { backdrop: 'z-[55]', panel: 'z-[60]' },
		elevated: { backdrop: 'z-[70]', panel: 'z-[80]' }
	} as const;

	let {
		open = $bindable(false),
		onclose = () => {},
		layer = 'default' as keyof typeof LAYERS,
		mobileLayout = 'sheet' as 'sheet' | 'center',
		clearBottomNav = true,
		backdrop = 'platform' as 'platform' | 'glass' | 'ios-scrim',
		labelledby,
		label,
		children
	}: {
		open?: boolean;
		onclose?: () => void;
		layer?: keyof typeof LAYERS;
		mobileLayout?: 'sheet' | 'center';
		// mobileLayout 'center' only: keep the pb-40 mobile bottom-nav clearance
		// (default) or centre the panel exactly (M3 dialog).
		clearBottomNav?: boolean;
		// 'platform' = iOS blur, Android and desktop --scrim
		// (ConfirmModal/ImportDialog);
		// 'glass' = always iOS-style blur on both platforms (TypeChoiceDialog);
		// 'ios-scrim' = the iOS mockup's own dim (--color-glass-scrim + 3px blur)
		// on iOS, platform behaviour elsewhere (BatchConfirmModal).
		backdrop?: 'platform' | 'glass' | 'ios-scrim';
		// Pass whichever the caller uses: labelledby (id ref) or label (literal).
		labelledby?: string;
		label?: string;
		children: Snippet;
	} = $props();

	function handleClose() {
		onclose();
	}

	function handleKeydown(event: KeyboardEvent) {
		if (open && event.key === 'Escape') {
			handleClose();
		}
	}

	const z = $derived(LAYERS[layer]);
	// 'sheet' = bottom-sheet on mobile, centered on desktop (ConfirmModal).
	// 'center' = centered on all breakpoints (ImportDialog).
	const alignClass = $derived(
		mobileLayout === 'sheet' ? 'items-end sm:items-center' : 'items-center'
	);
	// 'sheet' hugs the safe area so the sheet meets the screen edge; 'center'
	// insets the panel (p-4) and clears the mobile bottom nav (pb-40),
	// collapsing to a plain p-4 on desktop — this reproduces ImportDialog's
	// original outer padding byte-for-byte. clearBottomNav={false} drops the
	// pb-40 lift for a caller that wants the panel truly centred (the M3 import
	// dialog): the panel already stacks above the bottom nav (z-80 vs z-50), so
	// the clearance only pushed it off centre.
	const padClass = $derived(
		mobileLayout === 'sheet'
			? 'pb-[env(safe-area-inset-bottom)] sm:p-4'
			: clearBottomNav
				? 'p-4 pb-40 sm:pb-4'
				: 'p-4'
	);
	// Full literal class strings so the Tailwind JIT scanner sees them.
	// Android and desktop share bg-scrim: --color-scrim is the M3 dialog/sheet
	// scrim the Android mockup asks for, and the desktop mockup lands on the
	// same value.
	const backdropClass = $derived(
		backdrop === 'ios-scrim' && platform === 'ios'
			? 'bg-[var(--color-glass-scrim)] backdrop-blur-[3px]'
			: backdrop === 'glass'
				? 'bg-black/40 backdrop-blur-sm'
				: platform === 'ios'
					? 'bg-black/40 backdrop-blur-sm'
					: 'bg-scrim'
	);
</script>

<!-- Escape closes from anywhere; the backdrop is not focusable so a window
     listener is the only way keydown reaches this shell (guarded by open). -->
<svelte:window onkeydown={handleKeydown} />

{#if open}
	<!-- Backdrop carries the outside-click close. The panel div below is stacked
	     above and covers the viewport, so the panel is pointer-events-none and the
	     content sets pointer-events-auto — empty-area clicks fall through here. -->
	<div
		class="fixed inset-0 {z.backdrop} {backdropClass}"
		onclick={handleClose}
		role="presentation"
	></div>

	<!-- Panel positioning: bottom sheet on mobile, centered dialog on desktop.
	     pointer-events-none lets outside clicks reach the backdrop; the content
	     re-enables pointer events (see caller wrappers). -->
	<div
		class="pointer-events-none fixed inset-0 {z.panel} flex {alignClass} justify-center {padClass}"
		role="dialog"
		aria-modal="true"
		aria-labelledby={labelledby}
		aria-label={label}
	>
		{@render children()}
	</div>
{/if}
