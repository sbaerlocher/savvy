<script lang="ts">
	import { toastStore } from '$lib/stores/toast';
	import type { Toast } from '$lib/stores/toast';
	import { fly, fade } from 'svelte/transition';
	import { quintOut } from 'svelte/easing';
	import { platform } from '$lib/utils/platform';

	const isIOS = platform === 'ios';

	const styles: Record<
		Toast['type'],
		{ bg: string; border: string; text: string; icon: string }
	> = {
		success: {
			bg: isIOS
				? 'bg-white/60 backdrop-blur-xl backdrop-saturate-150'
				: 'bg-green-50',
			border: isIOS ? 'border-white/30' : 'border-green-200',
			text: 'text-green-800',
			icon: 'text-green-500'
		},
		error: {
			bg: isIOS
				? 'bg-white/60 backdrop-blur-xl backdrop-saturate-150'
				: 'bg-red-50',
			border: isIOS ? 'border-white/30' : 'border-red-200',
			text: 'text-red-800',
			icon: 'text-red-500'
		},
		warning: {
			bg: isIOS
				? 'bg-white/60 backdrop-blur-xl backdrop-saturate-150'
				: 'bg-yellow-50',
			border: isIOS ? 'border-white/30' : 'border-yellow-200',
			text: 'text-yellow-800',
			icon: 'text-yellow-500'
		},
		info: {
			bg: isIOS
				? 'bg-white/60 backdrop-blur-xl backdrop-saturate-150'
				: 'bg-cyan-50',
			border: isIOS ? 'border-white/30' : 'border-cyan-200',
			text: 'text-cyan-800',
			icon: 'text-cyan-500'
		}
	};
</script>

<div
	class="toast-container fixed sm:bottom-6 left-1/2 -translate-x-1/2 z-50 flex flex-col gap-2 w-[calc(100%-2rem)] max-w-md"
	aria-live="polite"
>
	{#each $toastStore as toast (toast.id)}
		<div
			in:fly={{ y: 20, duration: 250, easing: quintOut }}
			out:fade={{ duration: 150 }}
			class="{isIOS
				? 'rounded-2xl'
				: 'rounded-lg'} border shadow-xl px-4 py-3 flex items-start gap-3 {styles[
				toast.type
			].bg} {styles[toast.type].border}"
			role="status"
		>
			<!-- Icon -->
			{#if toast.type === 'success'}
				<svg
					class="h-5 w-5 shrink-0 mt-0.5 {styles[toast.type].icon}"
					fill="none"
					stroke="currentColor"
					viewBox="0 0 24 24"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
					/>
				</svg>
			{:else if toast.type === 'error'}
				<svg
					class="h-5 w-5 shrink-0 mt-0.5 {styles[toast.type].icon}"
					fill="none"
					stroke="currentColor"
					viewBox="0 0 24 24"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z"
					/>
				</svg>
			{:else if toast.type === 'warning'}
				<svg
					class="h-5 w-5 shrink-0 mt-0.5 {styles[toast.type].icon}"
					fill="none"
					stroke="currentColor"
					viewBox="0 0 24 24"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
					/>
				</svg>
			{:else}
				<svg
					class="h-5 w-5 shrink-0 mt-0.5 {styles[toast.type].icon}"
					fill="none"
					stroke="currentColor"
					viewBox="0 0 24 24"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
					/>
				</svg>
			{/if}

			<!-- Message -->
			<p class="flex-1 text-sm font-medium {styles[toast.type].text}">
				{toast.message}
			</p>

			<!-- Dismiss -->
			<button
				onclick={() => toastStore.remove(toast.id)}
				class="shrink-0 p-0.5 rounded-md transition-colors {styles[toast.type]
					.text} hover:opacity-70"
				aria-label="Close"
			>
				<svg
					class="h-4 w-4"
					fill="none"
					stroke="currentColor"
					viewBox="0 0 24 24"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M6 18L18 6M6 6l12 12"
					/>
				</svg>
			</button>
		</div>
	{/each}
</div>
