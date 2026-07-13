<script lang="ts">
	import type { Snippet } from 'svelte';
	import { t } from '$lib/stores/i18n';
	import { isOnline } from '$lib/stores/offline';
	import { platform } from '$lib/utils/platform';
	import { get } from 'svelte/store';

	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);

	interface Props {
		selectedCount: number;
		totalCount: number;
		sharedSelectedCount?: number;
		hasNonDeletableShared?: boolean;
		onSelectAll: () => void;
		onDeselectAll: () => void;
		onDelete: () => void;
		onShare: () => void;
		onTransfer: () => void;
		onExport: () => void;
		onCancel: () => void;
		headerExtra?: Snippet;
	}

	let {
		selectedCount,
		totalCount,
		sharedSelectedCount = 0,
		hasNonDeletableShared = false,
		onSelectAll,
		onDeselectAll,
		onDelete,
		onShare,
		onTransfer,
		onExport,
		onCancel,
		headerExtra
	}: Props = $props();

	const isOffline = $derived(!$isOnline);
	const allSelected = $derived(selectedCount === totalCount && totalCount > 0);
	const hasSharedSelected = $derived(sharedSelectedCount > 0);
	const disableShareTransfer = $derived(
		isOffline || selectedCount === 0 || hasSharedSelected
	);
	const disableExport = $derived(isOffline || selectedCount === 0);
	const disableDelete = $derived(
		isOffline || selectedCount === 0 || hasNonDeletableShared
	);
</script>

<!-- Desktop Side-Panel -->
<div class="hidden lg:block lg:col-span-1">
	<div class="bg-white rounded-xl shadow-lg sticky top-4 overflow-hidden">
		<!-- Header -->
		<div class="px-5 py-4 bg-surface-1/80 border-b border-border-soft">
			<div class="flex items-center justify-between">
				<div class="flex items-center gap-2">
					<svg
						class="w-4 h-4 text-text-subtle"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"
						></path>
					</svg>
					<h3 class="text-sm font-semibold text-text">
						{tr('batch.selectMode')}
					</h3>
				</div>
				<div class="flex items-center gap-2.5">
					<span
						class="text-xs text-text-subtle bg-white px-2.5 py-1 rounded-full border border-border tabular-nums"
					>
						{selectedCount} / {totalCount}
					</span>
					<button
						type="button"
						onclick={onCancel}
						class="text-text-faint hover:text-text-muted transition-colors"
						aria-label={tr('batch.exitSelectMode')}
					>
						<svg
							class="w-4 h-4"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M6 18L18 6M6 6l12 12"
							></path>
						</svg>
					</button>
				</div>
			</div>
		</div>

		<div class="p-5">
			{#if headerExtra}
				<div class="pb-4">
					{@render headerExtra()}
				</div>
				<div class="border-t border-border-soft"></div>
			{/if}

			<!-- Select All / Deselect All -->
			<div class="{headerExtra ? 'pt-4 ' : ''}pb-4">
				<button
					type="button"
					onclick={allSelected ? onDeselectAll : onSelectAll}
					class="w-full flex items-center gap-2 cursor-pointer group"
				>
					<svg
						class="w-4 h-4 text-text-subtle"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M4 6h16M4 10h16M4 14h16M4 18h16"
						></path>
					</svg>
					<span class="text-sm font-medium text-text-ink2">
						{allSelected ? tr('batch.deselectAll') : tr('batch.selectAll')}
					</span>
				</button>
			</div>

			{#if sharedSelectedCount > 0}
				<div class="pb-4">
					<div
						class="flex items-start gap-2 bg-amber-50 border border-amber-200 rounded-lg p-3"
					>
						<svg
							class="w-4 h-4 text-amber-600 mt-0.5 flex-shrink-0"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
							></path>
						</svg>
						<p class="text-xs text-amber-800">
							{tr('batch.sharedItemsWarning', { count: sharedSelectedCount })}
						</p>
					</div>
				</div>
			{/if}

			<div class="border-t border-border-soft"></div>

			<!-- Actions -->
			<div class="pt-4 space-y-2">
				<p
					class="text-xs font-medium text-text-subtle uppercase tracking-wider mb-3"
				>
					{tr('batch.actions')}
				</p>

				<!-- Share -->
				<button
					type="button"
					onclick={onShare}
					disabled={disableShareTransfer}
					class="w-full flex items-center gap-3 px-3 py-2.5 text-sm text-text-ink2 bg-surface-1 border border-border rounded-lg hover:bg-accent-50 hover:border-accent-200 hover:text-accent-hover transition-colors disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:bg-surface-1 disabled:hover:border-border disabled:hover:text-text-ink2"
				>
					<svg
						class="w-4 h-4"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m0 2.684l6.632 3.316m-6.632-6l6.632-3.316m0 0a3 3 0 105.367-2.684 3 3 0 00-5.367 2.684zm0 9.316a3 3 0 105.368 2.684 3 3 0 00-5.368-2.684z"
						></path>
					</svg>
					{tr('batch.shareSelected')}
				</button>

				<!-- Transfer -->
				<button
					type="button"
					onclick={onTransfer}
					disabled={disableShareTransfer}
					class="w-full flex items-center gap-3 px-3 py-2.5 text-sm text-text-ink2 bg-surface-1 border border-border rounded-lg hover:bg-purple-50 hover:border-purple-200 hover:text-purple-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:bg-surface-1 disabled:hover:border-border disabled:hover:text-text-ink2"
				>
					<svg
						class="w-4 h-4"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4"
						></path>
					</svg>
					{tr('batch.transferSelected')}
				</button>

				<!-- Export -->
				<button
					type="button"
					onclick={onExport}
					disabled={disableExport}
					class="w-full flex items-center gap-3 px-3 py-2.5 text-sm text-text-ink2 bg-surface-1 border border-border rounded-lg hover:bg-emerald-50 hover:border-emerald-200 hover:text-emerald-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:bg-surface-1 disabled:hover:border-border disabled:hover:text-text-ink2"
				>
					<svg
						class="w-4 h-4"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"
						></path>
					</svg>
					{tr('batch.exportSelected')}
				</button>

				<!-- Delete -->
				<button
					type="button"
					onclick={onDelete}
					disabled={disableDelete}
					class="w-full flex items-center gap-3 px-3 py-2.5 text-sm text-text-ink2 bg-surface-1 border border-border rounded-lg hover:bg-red-50 hover:border-red-200 hover:text-red-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:bg-surface-1 disabled:hover:border-border disabled:hover:text-text-ink2"
				>
					<svg
						class="w-4 h-4"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
						></path>
					</svg>
					{tr('batch.deleteSelected')}
				</button>
			</div>

			<div class="border-t border-border-soft mt-4"></div>

			<!-- Exit Select Mode -->
			<div class="pt-4">
				<button
					type="button"
					onclick={onCancel}
					class="w-full text-sm text-text-subtle hover:text-text-ink2 py-2 rounded-lg hover:bg-surface-1 transition-colors flex items-center justify-center gap-1.5"
				>
					<svg
						class="w-3.5 h-3.5"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M6 18L18 6M6 6l12 12"
						></path>
					</svg>
					{tr('batch.exitSelectMode')}
				</button>
			</div>
		</div>
	</div>
</div>

<!-- Mobile Fixed Bottom Bar -->
<div
	class="lg:hidden fixed left-0 right-0 z-[55] {platform === 'ios'
		? 'mx-4 rounded-2xl bg-white/70 backdrop-blur-xl backdrop-saturate-150 border border-white/40 shadow-lg batch-panel-floating'
		: platform === 'android'
			? 'bottom-16 sm:bottom-0 bg-[#FFFBFE] border-t border-[#CAC4D0] shadow-[0_-2px_6px_rgba(0,0,0,0.08)]'
			: 'bottom-16 sm:bottom-0 bg-white border-t border-border shadow-[0_-4px_12px_rgba(0,0,0,0.1)]'}"
	style={platform === 'ios'
		? ''
		: 'padding-bottom: env(safe-area-inset-bottom);'}
	role="toolbar"
	aria-label={tr('batch.selectMode')}
>
	<!-- Top row: header matching filter style -->
	<div class="px-6 py-3 flex items-center justify-between">
		<div class="flex items-center gap-2">
			<svg
				class="w-4 h-4 text-text-subtle"
				fill="none"
				stroke="currentColor"
				viewBox="0 0 24 24"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"
				></path>
			</svg>
			<h3 class="text-sm font-semibold text-text">
				{tr('batch.selectMode')}
			</h3>
		</div>
		<div class="flex items-center gap-2.5">
			<span
				class="text-xs text-text-subtle bg-border-soft px-2.5 py-1 rounded-full border border-border tabular-nums"
			>
				{selectedCount} / {totalCount}
			</span>
			<button
				type="button"
				onclick={onCancel}
				class="text-text-faint hover:text-text-muted transition-colors"
				aria-label={tr('batch.exitSelectMode')}
			>
				<svg
					class="w-5 h-5"
					fill="none"
					stroke="currentColor"
					viewBox="0 0 24 24"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M6 18L18 6M6 6l12 12"
					></path>
				</svg>
			</button>
		</div>
	</div>

	{#if headerExtra}
		<div class="px-4 py-2.5 border-t border-border-soft">
			{@render headerExtra()}
		</div>
	{/if}

	{#if sharedSelectedCount > 0}
		<div class="px-4 py-2 bg-amber-50 border-t border-amber-200">
			<p class="text-xs text-amber-800 text-center">
				{tr('batch.sharedItemsWarning', { count: sharedSelectedCount })}
			</p>
		</div>
	{/if}

	<div class="border-t border-border-soft"></div>

	<!-- Bottom row: action buttons -->
	<div class="flex items-center justify-evenly px-2 py-2">
		<button
			type="button"
			onclick={allSelected ? onDeselectAll : onSelectAll}
			class="flex flex-col items-center gap-0.5 px-3 py-1.5 rounded-lg text-text-muted hover:bg-border-soft hover:text-text-strong transition-colors"
		>
			<svg
				class="w-5 h-5"
				fill="none"
				stroke="currentColor"
				viewBox="0 0 24 24"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d="M4 6h16M4 10h16M4 14h16M4 18h16"
				></path>
			</svg>
			<span class="text-[10px] font-medium"
				>{allSelected ? tr('batch.deselectAll') : tr('batch.selectAll')}</span
			>
		</button>

		<button
			type="button"
			onclick={onShare}
			disabled={disableShareTransfer}
			class="flex flex-col items-center gap-0.5 px-3 py-1.5 rounded-lg text-text-muted hover:bg-accent-50 hover:text-accent transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
		>
			<svg
				class="w-5 h-5"
				fill="none"
				stroke="currentColor"
				viewBox="0 0 24 24"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m0 2.684l6.632 3.316m-6.632-6l6.632-3.316m0 0a3 3 0 105.367-2.684 3 3 0 00-5.367 2.684zm0 9.316a3 3 0 105.368 2.684 3 3 0 00-5.368-2.684z"
				></path>
			</svg>
			<span class="text-[10px] font-medium">{tr('common.share')}</span>
		</button>

		<button
			type="button"
			onclick={onTransfer}
			disabled={disableShareTransfer}
			class="flex flex-col items-center gap-0.5 px-3 py-1.5 rounded-lg text-text-muted hover:bg-purple-50 hover:text-purple-600 transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
		>
			<svg
				class="w-5 h-5"
				fill="none"
				stroke="currentColor"
				viewBox="0 0 24 24"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4"
				></path>
			</svg>
			<span class="text-[10px] font-medium"
				>{tr('common.transferOwnership')}</span
			>
		</button>

		<button
			type="button"
			onclick={onDelete}
			disabled={disableDelete}
			class="flex flex-col items-center gap-0.5 px-3 py-1.5 rounded-lg text-text-muted hover:bg-red-50 hover:text-red-600 transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
		>
			<svg
				class="w-5 h-5"
				fill="none"
				stroke="currentColor"
				viewBox="0 0 24 24"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
				></path>
			</svg>
			<span class="text-[10px] font-medium">{tr('common.delete')}</span>
		</button>
	</div>
</div>
