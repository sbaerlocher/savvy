<script lang="ts">
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

	function handleBackdropClick() {
		handleCancel();
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape') {
			handleCancel();
		}
	}
</script>

{#if show}
	<!-- Backdrop -->
	<div
		class="fixed inset-0 z-40 {platform === 'ios'
			? 'bg-black/40 backdrop-blur-sm'
			: 'bg-black bg-opacity-50'}"
		onclick={handleBackdropClick}
		onkeydown={handleKeydown}
		role="presentation"
	></div>

	<!-- Modal -->
	<div
		class="fixed inset-0 z-50 flex items-center justify-center p-4 pb-52 sm:pb-4"
		role="dialog"
		aria-modal="true"
		aria-labelledby="modal-title"
	>
		<div
			class="max-w-md w-full p-6 shadow-xl {platform === 'ios'
				? 'bg-white/70 backdrop-blur-xl backdrop-saturate-150 rounded-2xl border border-white/30'
				: 'bg-white dark:bg-gray-800 rounded-lg'}"
		>
			<!-- Header -->
			<div class="flex items-start mb-4">
				{#if variant === 'danger'}
					<div class="flex-shrink-0 mr-3">
						<svg
							class="h-6 w-6 text-red-600"
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
							/>
						</svg>
					</div>
				{:else if variant === 'warning'}
					<div class="flex-shrink-0 mr-3">
						<svg
							class="h-6 w-6 text-yellow-600"
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
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
								d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4"
							/>
						</svg>
					</div>
				{:else}
					<div class="flex-shrink-0 mr-3">
						<svg
							class="h-6 w-6 text-cyan-600"
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
							/>
						</svg>
					</div>
				{/if}
				<div class="flex-1">
					<h3
						id="modal-title"
						class="text-lg font-semibold {variant === 'danger'
							? 'text-red-600 dark:text-red-400'
							: variant === 'warning'
								? 'text-yellow-600 dark:text-yellow-400'
								: variant === 'transfer'
									? 'text-purple-600 dark:text-purple-400'
									: 'text-cyan-600 dark:text-cyan-400'}"
					>
						{title}
					</h3>
				</div>
			</div>

			<!-- Message -->
			<p class="text-gray-600 dark:text-gray-300 mb-6 ml-9">
				{message}
			</p>

			<!-- Actions -->
			<div class="flex gap-3 justify-end">
				<button
					type="button"
					class="px-4 py-2 rounded-md border border-gray-300 dark:border-gray-600 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors text-gray-700 dark:text-gray-300"
					onclick={handleCancel}
					data-testid="modal-cancel"
				>
					{cancelText}
				</button>
				<button
					type="button"
					class="px-4 py-2 rounded-md text-white transition-colors
            {variant === 'danger'
						? 'bg-red-600 hover:bg-red-700'
						: variant === 'warning'
							? 'bg-yellow-600 hover:bg-yellow-700'
							: variant === 'transfer'
								? 'bg-purple-600 hover:bg-purple-700'
								: 'bg-cyan-600 hover:bg-cyan-700'}"
					onclick={handleConfirm}
					data-testid="modal-confirm"
				>
					{confirmText}
				</button>
			</div>
		</div>
	</div>
{/if}
