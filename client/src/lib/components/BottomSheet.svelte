<script lang="ts">
	import { platform } from '$lib/utils/platform';
	import type { Snippet } from 'svelte';

	interface Props {
		open: boolean;
		onClose: () => void;
		maxHeight?: string;
		ariaLabel?: string;
		children: Snippet;
		dialogRef?: HTMLDivElement;
	}

	let {
		open,
		onClose,
		maxHeight = '85vh',
		ariaLabel,
		children,
		dialogRef = $bindable()
	}: Props = $props();

	const backdropClass = $derived.by(() => {
		switch (platform) {
			case 'ios':
				return 'lg:hidden fixed inset-0 bg-black/30 backdrop-blur-sm z-[60]';
			default:
				return 'lg:hidden fixed inset-0 bg-black/50 z-[60]';
		}
	});

	const sheetClass = $derived.by(() => {
		const base = 'absolute bottom-0 left-0 right-0 overflow-y-auto';

		switch (platform) {
			case 'ios':
				return `${base} liquid-glass-surface rounded-t-3xl`;
			case 'android':
				return `${base} bg-surface rounded-t-2xl shadow-[var(--shadow-sheet)]`;
			default:
				return `${base} bg-white rounded-t-2xl shadow-2xl`;
		}
	});

	const handleClass = $derived.by(() => {
		switch (platform) {
			case 'ios':
				return 'w-10 h-1 bg-text-faint/50 rounded-full';
			case 'android':
				return 'w-8 h-1 bg-border rounded-full';
			default:
				return 'w-12 h-1.5 bg-border-field rounded-full';
		}
	});
</script>

{#if open}
	<div
		class={backdropClass}
		onclick={onClose}
		onkeydown={(e) => {
			if (e.key === 'Escape' || e.key === 'Enter' || e.key === ' ') {
				e.preventDefault();
				onClose();
			}
		}}
		role="button"
		tabindex="0"
		aria-hidden="true"
	>
		<div
			bind:this={dialogRef}
			class={sheetClass}
			style="max-height: {maxHeight}; padding-bottom: max(1.5rem, env(safe-area-inset-bottom));"
			onclick={(e) => e.stopPropagation()}
			onkeydown={(e) => {
				e.stopPropagation();
				if (e.key === 'Escape') {
					e.preventDefault();
					onClose();
				}
			}}
			role="dialog"
			aria-modal="true"
			aria-label={ariaLabel}
			tabindex="-1"
		>
			<!-- Drag Handle -->
			<div class="flex justify-center pt-3 pb-2">
				<div class={handleClass}></div>
			</div>

			{@render children()}
		</div>
	</div>
{/if}
