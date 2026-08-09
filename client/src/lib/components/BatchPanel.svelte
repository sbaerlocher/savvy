<script lang="ts">
	import {
		ICON_CLIPBOARD_CHECK,
		ICON_CLOSE,
		ICON_EXPORT,
		ICON_LINES,
		ICON_SHARE,
		ICON_TRANSFER,
		ICON_TRASH,
		ICON_WARNING
	} from '$lib/icons';
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

<!-- One M3 batch-bar action: stacked icon over an 11px label (wallet mockup). -->
{#snippet androidAction(
	label: string,
	d: string,
	onclick: () => void,
	disabled: boolean,
	danger: boolean
)}
	<button
		type="button"
		{onclick}
		{disabled}
		class="flex flex-col items-center gap-1 rounded-m3-sm px-2 py-1 transition-colors disabled:cursor-not-allowed disabled:opacity-40 {danger
			? 'text-danger-600'
			: 'text-text-muted'}"
	>
		<svg
			class="h-5.5 w-5.5"
			fill="none"
			stroke="currentColor"
			stroke-width="2"
			viewBox="0 0 24 24"
			aria-hidden="true"
		>
			<path stroke-linecap="round" stroke-linejoin="round" {d} />
		</svg>
		<span class="text-eyebrow font-medium normal-case tracking-normal"
			>{label}</span
		>
	</button>
{/snippet}

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
							d={ICON_CLIPBOARD_CHECK}
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
								d={ICON_CLOSE}
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
						class="flex items-start gap-2 bg-warning-50 border border-warning-200 rounded-lg p-3"
					>
						<svg
							class="w-4 h-4 text-warning-600 mt-0.5 flex-shrink-0"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d={ICON_WARNING}
							></path>
						</svg>
						<p class="text-xs text-warning-800">
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
							d={ICON_SHARE}
						></path>
					</svg>
					{tr('batch.shareSelected')}
				</button>

				<!-- Transfer -->
				<button
					type="button"
					onclick={onTransfer}
					disabled={disableShareTransfer}
					class="w-full flex items-center gap-3 px-3 py-2.5 text-sm text-text-ink2 bg-surface-1 border border-border rounded-lg hover:bg-transfer-50 hover:border-transfer-200 hover:text-transfer-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:bg-surface-1 disabled:hover:border-border disabled:hover:text-text-ink2"
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
							d={ICON_TRANSFER}
						></path>
					</svg>
					{tr('batch.transferSelected')}
				</button>

				<!-- Export -->
				<button
					type="button"
					onclick={onExport}
					disabled={disableExport}
					class="w-full flex items-center gap-3 px-3 py-2.5 text-sm text-text-ink2 bg-surface-1 border border-border rounded-lg hover:bg-success-50 hover:border-success-200 hover:text-success-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:bg-surface-1 disabled:hover:border-border disabled:hover:text-text-ink2"
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
					class="w-full flex items-center gap-3 px-3 py-2.5 text-sm text-text-ink2 bg-surface-1 border border-border rounded-lg hover:bg-danger-50 hover:border-danger-200 hover:text-danger-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:bg-surface-1 disabled:hover:border-border disabled:hover:text-text-ink2"
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
							d={ICON_TRASH}
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
							d={ICON_CLOSE}
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
		? 'liquid-glass-surface mx-4 rounded-full batch-panel-floating'
		: platform === 'android'
			? 'batch-panel-android bg-m3-surface-container border-t border-border'
			: 'bottom-16 sm:bottom-0 bg-white border-t border-border shadow-[var(--shadow-batch)]'}"
	style={platform === 'ios'
		? ''
		: 'padding-bottom: env(safe-area-inset-bottom);'}
	role="toolbar"
	aria-label={tr('batch.selectMode')}
>
	{#if platform === 'ios'}
		<!-- iOS multi-select bar (mockup screen-WalletIOS, Phone 3): a single row
		     of evenly spaced icon actions in a floating glass pill. Select-all,
		     the count and Done sit at the top of the list, not in this bar. -->
		{#if sharedSelectedCount > 0}
			<p class="px-4 pt-2 text-center text-xs text-warning-700">
				{tr('batch.sharedItemsWarning', { count: sharedSelectedCount })}
			</p>
		{/if}

		<!-- Matches the 64px bottom-nav pill this bar replaces. The glass surface
		     adds a 1px border top and bottom, so the row itself is 62px. -->
		<div class="flex h-15.5 items-center justify-around px-2.5">
			<button
				type="button"
				onclick={onShare}
				disabled={disableShareTransfer}
				aria-label={tr('common.share')}
				class="flex min-w-16 flex-col items-center justify-center gap-1 rounded-lg text-accent transition-colors active:opacity-60 disabled:opacity-40"
			>
				<svg
					class="h-5.5 w-5.5"
					fill="none"
					stroke="currentColor"
					stroke-width="1.8"
					viewBox="0 0 24 24"
				>
					<path stroke-linecap="round" stroke-linejoin="round" d={ICON_SHARE} />
				</svg>
				<span class="text-[length:var(--text-tag)]">{tr('common.share')}</span>
			</button>
			<button
				type="button"
				onclick={onTransfer}
				disabled={disableShareTransfer}
				aria-label={tr('common.transferOwnership')}
				class="flex min-w-16 flex-col items-center justify-center gap-1 rounded-lg text-accent transition-colors active:opacity-60 disabled:opacity-40"
			>
				<svg
					class="h-5.5 w-5.5"
					fill="none"
					stroke="currentColor"
					stroke-width="1.8"
					viewBox="0 0 24 24"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						d={ICON_TRANSFER}
					/>
				</svg>
				<span class="text-[length:var(--text-tag)]"
					>{tr('common.transferOwnership')}</span
				>
			</button>
			<button
				type="button"
				onclick={onExport}
				disabled={disableExport}
				aria-label={tr('batch.exportSelected')}
				class="flex min-w-16 flex-col items-center justify-center gap-1 rounded-lg text-accent transition-colors active:opacity-60 disabled:opacity-40"
			>
				<svg
					class="h-5.5 w-5.5"
					fill="none"
					stroke="currentColor"
					stroke-width="1.8"
					viewBox="0 0 24 24"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"
					/>
				</svg>
				<span class="text-[length:var(--text-tag)]">{tr('common.export')}</span>
			</button>
			<button
				type="button"
				onclick={onDelete}
				disabled={disableDelete}
				aria-label={tr('common.delete')}
				class="flex min-w-16 flex-col items-center justify-center gap-1 rounded-lg text-danger-600 transition-colors active:opacity-60 disabled:opacity-40"
			>
				<svg
					class="h-5.5 w-5.5"
					fill="none"
					stroke="currentColor"
					stroke-width="1.8"
					viewBox="0 0 24 24"
				>
					<path stroke-linecap="round" stroke-linejoin="round" d={ICON_TRASH} />
				</svg>
				<span class="text-[length:var(--text-tag)]">{tr('common.delete')}</span>
			</button>
		</div>
	{:else if platform === 'android'}
		<!-- M3 batch bar: five evenly spaced icon actions, edge-to-edge. The
		     selection count and the exit action live in the contextual top app
		     bar instead (wallet mockup), so no header row here. -->
		{#if headerExtra}
			<!-- Carries the type filter on the merchant detail screen, where it has
			     no other home while a selection is active. -->
			<div class="border-b border-border-soft px-4 py-2.5">
				{@render headerExtra()}
			</div>
		{/if}
		{#if sharedSelectedCount > 0}
			<p class="px-4 pt-2 text-center text-body-sm text-warning-700">
				{tr('batch.sharedItemsWarning', { count: sharedSelectedCount })}
			</p>
		{/if}
		<div class="flex items-center justify-around px-1.5 pt-3 pb-1.5">
			{@render androidAction(
				allSelected ? tr('batch.deselectAll') : tr('batch.selectAll'),
				ICON_LINES,
				allSelected ? onDeselectAll : onSelectAll,
				false,
				false
			)}
			{@render androidAction(
				tr('common.share'),
				ICON_SHARE,
				onShare,
				disableShareTransfer,
				false
			)}
			{@render androidAction(
				tr('common.transferOwnership'),
				ICON_TRANSFER,
				onTransfer,
				disableShareTransfer,
				false
			)}
			{@render androidAction(
				tr('common.export'),
				ICON_EXPORT,
				onExport,
				disableExport,
				false
			)}
			{@render androidAction(
				tr('common.delete'),
				ICON_TRASH,
				onDelete,
				disableDelete,
				true
			)}
		</div>
	{:else}
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
						d={ICON_CLIPBOARD_CHECK}
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
							d={ICON_CLOSE}
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
			<div class="px-4 py-2 bg-warning-50 border-t border-warning-200">
				<p class="text-xs text-warning-800 text-center">
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
				<span class="text-[length:var(--text-tag)] font-medium"
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
						d={ICON_SHARE}
					></path>
				</svg>
				<span class="text-[length:var(--text-tag)] font-medium"
					>{tr('common.share')}</span
				>
			</button>

			<button
				type="button"
				onclick={onTransfer}
				disabled={disableShareTransfer}
				class="flex flex-col items-center gap-0.5 px-3 py-1.5 rounded-lg text-text-muted hover:bg-transfer-50 hover:text-transfer-600 transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
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
						d={ICON_TRANSFER}
					></path>
				</svg>
				<span class="text-[length:var(--text-tag)] font-medium"
					>{tr('common.transferOwnership')}</span
				>
			</button>

			<button
				type="button"
				onclick={onDelete}
				disabled={disableDelete}
				class="flex flex-col items-center gap-0.5 px-3 py-1.5 rounded-lg text-text-muted hover:bg-danger-50 hover:text-danger-600 transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
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
						d={ICON_TRASH}
					></path>
				</svg>
				<span class="text-[length:var(--text-tag)] font-medium"
					>{tr('common.delete')}</span
				>
			</button>
		</div>
	{/if}
</div>
