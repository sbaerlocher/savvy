<script lang="ts">
	import { platform } from '$lib/utils/platform';
	import type { Snippet } from 'svelte';

	// Reusable modal shell: backdrop (iOS blur vs Android opacity), mobile
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
		labelledby,
		children
	}: {
		open?: boolean;
		onclose?: () => void;
		layer?: keyof typeof LAYERS;
		mobileLayout?: 'sheet' | 'center';
		labelledby?: string;
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
	// original outer padding byte-for-byte.
	const padClass = $derived(
		mobileLayout === 'sheet'
			? 'pb-[env(safe-area-inset-bottom)] sm:p-4'
			: 'p-4 pb-40 sm:pb-4'
	);
</script>

<!-- Escape closes from anywhere; the backdrop is not focusable so a window
     listener is the only way keydown reaches this shell (guarded by open). -->
<svelte:window onkeydown={handleKeydown} />

{#if open}
	<!-- Backdrop -->
	<div
		class="fixed inset-0 {z.backdrop} {platform === 'ios'
			? 'bg-black/40 backdrop-blur-sm'
			: 'bg-black bg-opacity-50'}"
		onclick={handleClose}
		role="presentation"
	></div>

	<!-- Panel positioning: bottom sheet on mobile, centered dialog on desktop -->
	<div
		class="fixed inset-0 {z.panel} flex {alignClass} justify-center {padClass}"
		role="dialog"
		aria-modal="true"
		aria-labelledby={labelledby}
	>
		{@render children()}
	</div>
{/if}
