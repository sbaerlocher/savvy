<script lang="ts">
	import {
		ICON_CHECK_CIRCLE,
		ICON_CLOSE,
		ICON_INFO_CIRCLE,
		ICON_WARNING
	} from '$lib/icons';
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
				: 'bg-success-50',
			border: isIOS ? 'border-white/30' : 'border-success-200',
			text: 'text-success-800',
			icon: 'text-success-500'
		},
		error: {
			bg: isIOS
				? 'bg-white/60 backdrop-blur-xl backdrop-saturate-150'
				: 'bg-danger-50',
			border: isIOS ? 'border-white/30' : 'border-danger-200',
			text: 'text-danger-800',
			icon: 'text-danger-500'
		},
		warning: {
			bg: isIOS
				? 'bg-white/60 backdrop-blur-xl backdrop-saturate-150'
				: 'bg-warning-50',
			border: isIOS ? 'border-white/30' : 'border-warning-200',
			text: 'text-warning-800',
			icon: 'text-warning-500'
		},
		info: {
			bg: isIOS
				? 'bg-white/60 backdrop-blur-xl backdrop-saturate-150'
				: 'bg-accent-50',
			border: isIOS ? 'border-white/30' : 'border-accent-200',
			text: 'text-accent-800',
			icon: 'text-accent'
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
						d={ICON_CHECK_CIRCLE}
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
						d={ICON_WARNING}
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
						d={ICON_INFO_CIRCLE}
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
						d={ICON_CLOSE}
					/>
				</svg>
			</button>
		</div>
	{/each}
</div>
