<script lang="ts">
	import { ICON_INFO_CIRCLE, ICON_TRANSFER, ICON_WARNING } from '$lib/icons';
	import Modal from '$lib/components/ui/Modal.svelte';
	import { platform } from '$lib/utils/platform';

	let {
		isOpen = false,
		title = '',
		message = '',
		confirmText = 'Bestätigen',
		cancelText = 'Abbrechen',
		variant = 'warning' as 'danger' | 'warning' | 'info' | 'transfer',
		onconfirm = () => {},
		oncancel = () => {}
	}: {
		isOpen?: boolean;
		title?: string;
		message?: string;
		confirmText?: string;
		cancelText?: string;
		variant?: 'danger' | 'warning' | 'info' | 'transfer';
		onconfirm?: () => void;
		oncancel?: () => void;
	} = $props();

	// Make isOpen explicitly reactive
	let show = $derived(isOpen);

	function handleConfirm() {
		onconfirm();
	}

	function handleCancel() {
		oncancel();
	}
</script>

<Modal open={show} onclose={handleCancel} labelledby="modal-title">
	<div
		class="pointer-events-auto w-full sm:max-w-md p-6 {platform === 'ios'
			? 'liquid-glass-surface rounded-t-3xl sm:rounded-2xl'
			: 'bg-white dark:bg-text-strong rounded-t-3xl sm:rounded-lg shadow-xl'}"
	>
		<!-- Header -->
		<div class="flex items-start mb-4">
			{#if variant === 'danger'}
				<div class="flex-shrink-0 mr-3">
					<svg
						class="h-6 w-6 text-danger-600"
						fill="none"
						viewBox="0 0 24 24"
						stroke="currentColor"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d={ICON_WARNING}
						/>
					</svg>
				</div>
			{:else if variant === 'warning'}
				<div class="flex-shrink-0 mr-3">
					<svg
						class="h-6 w-6 text-warning-600"
						fill="none"
						viewBox="0 0 24 24"
						stroke="currentColor"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d={ICON_WARNING}
						/>
					</svg>
				</div>
			{:else if variant === 'transfer'}
				<div class="flex-shrink-0 mr-3">
					<svg
						class="h-6 w-6 text-purple-600"
						fill="none"
						viewBox="0 0 24 24"
						stroke="currentColor"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d={ICON_TRANSFER}
						/>
					</svg>
				</div>
			{:else}
				<div class="flex-shrink-0 mr-3">
					<svg
						class="h-6 w-6 text-accent"
						fill="none"
						viewBox="0 0 24 24"
						stroke="currentColor"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d={ICON_INFO_CIRCLE}
						/>
					</svg>
				</div>
			{/if}
			<div class="flex-1">
				<h3
					id="modal-title"
					class="text-lg font-semibold {variant === 'danger'
						? 'text-danger-600 dark:text-danger-400'
						: variant === 'warning'
							? 'text-warning-600 dark:text-warning-400'
							: variant === 'transfer'
								? 'text-purple-600 dark:text-purple-400'
								: 'text-accent dark:text-accent-400'}"
				>
					{title}
				</h3>
			</div>
		</div>

		<!-- Message -->
		<p class="text-text-muted dark:text-text-placeholder mb-6 ml-9">
			{message}
		</p>

		<!-- Actions -->
		<div class="flex gap-3 justify-end">
			<button
				type="button"
				class="px-4 py-2 rounded-md border border-border-field dark:border-text-muted hover:bg-surface-1 dark:hover:bg-text-ink2 transition-colors text-text-ink2 dark:text-text-placeholder"
				onclick={handleCancel}
				data-testid="modal-cancel"
			>
				{cancelText}
			</button>
			<button
				type="button"
				class="px-4 py-2 rounded-md text-white transition-colors
            {variant === 'danger'
					? 'bg-danger-600 hover:bg-danger-700'
					: variant === 'warning'
						? 'bg-warning-600 hover:bg-warning-700'
						: variant === 'transfer'
							? 'bg-purple-600 hover:bg-purple-700'
							: 'bg-accent hover:bg-accent-hover'}"
				onclick={handleConfirm}
				data-testid="modal-confirm"
			>
				{confirmText}
			</button>
		</div>
	</div>
</Modal>
