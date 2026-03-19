<script lang="ts">
	import type { Snippet } from 'svelte';
	import { t } from '$lib/stores/i18n';
	import type { ShareDTO } from '$lib/types/api';

	interface Props {
		share: ShareDTO;
		isEditing: boolean;
		isOffline: boolean;
		/** Button label in view mode. Defaults to t('common.edit'). */
		editButtonLabel?: string;
		/** Delete button label in edit mode. Defaults to t('giftCards.sharing.removeButton'). */
		deleteButtonLabel?: string;
		/** viewOnly badge label. Defaults to t('common.viewOnly'). */
		viewOnlyLabel?: string;
		/** Always show viewOnly badge regardless of permissions (e.g. vouchers). */
		alwaysViewOnly?: boolean;
		/** Show can_edit_transactions badge in view mode. */
		showTransactionsBadge?: boolean;
		/** Label for transactions permission badge. */
		labelPermTransactions?: string;
		/** Edit-mode body content (permission checkboxes or info box). */
		children?: Snippet;
		onstartEdit?: () => void;
		/** If not provided, the save button is hidden (e.g. vouchers manage-only mode). */
		onsave?: () => void;
		oncancel?: () => void;
		ondelete?: () => void;
	}

	const LOCK_ICON_PATH =
		'M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z';

	let {
		share,
		isEditing,
		isOffline,
		editButtonLabel,
		deleteButtonLabel,
		viewOnlyLabel,
		alwaysViewOnly = false,
		showTransactionsBadge = false,
		labelPermTransactions,
		children,
		onstartEdit,
		onsave,
		oncancel,
		ondelete
	}: Props = $props();

	function userName(s: ShareDTO) {
		const u = s.shared_with_user;
		if (u?.first_name && u?.last_name) return `${u.first_name} ${u.last_name}`;
		if (u?.first_name) return u.first_name;
		return u?.email || 'Unknown User';
	}
</script>

{#if isEditing}
	<div class="border border-cyan-200 bg-cyan-50 rounded-lg p-4 space-y-4 mb-4">
		<div>
			<p class="font-medium text-gray-900 text-sm">{userName(share)}</p>
			<p class="text-xs text-gray-500">{share.shared_with_user?.email || ''}</p>
		</div>

		{@render children?.()}

		<div class="flex gap-2">
			{#if onsave}
				<button
					onclick={onsave}
					disabled={isOffline}
					class="btn btn-primary flex-1"
				>
					{$t('common.save')}
				</button>
			{/if}
			<button onclick={oncancel} class="btn btn-ghost {onsave ? '' : 'flex-1'}">
				{$t('common.cancel')}
			</button>
		</div>

		<div class="pt-2 border-t border-cyan-200">
			<button
				type="button"
				onclick={ondelete}
				disabled={isOffline}
				class="btn btn-text-danger w-full flex items-center justify-center gap-1.5 {isOffline
					? 'pointer-events-none blur-[0.5px]'
					: ''}"
			>
				{#if isOffline}
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
							d={LOCK_ICON_PATH}
						></path>
					</svg>
				{/if}
				{deleteButtonLabel ?? $t('giftCards.sharing.removeButton')}
			</button>
		</div>
	</div>
{:else}
	<div class="border border-gray-200 rounded-lg p-3">
		<div class="flex justify-between items-start mb-2">
			<div class="flex-1">
				<p class="font-medium text-gray-900 text-sm">{userName(share)}</p>
				<p class="text-xs text-gray-500">
					{share.shared_with_user?.email || ''}
				</p>
			</div>
			<button
				onclick={onstartEdit}
				disabled={isOffline}
				class="btn-text text-xs flex items-center gap-1"
			>
				{#if isOffline}
					<svg
						class="w-3 h-3"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d={LOCK_ICON_PATH}
						></path>
					</svg>
				{:else}
					{editButtonLabel ?? $t('common.edit')}
				{/if}
			</button>
		</div>
		<div class="flex flex-wrap gap-1">
			{#if alwaysViewOnly}
				<span class="text-xs bg-gray-100 text-gray-600 px-2 py-0.5 rounded">
					{viewOnlyLabel ?? $t('common.viewOnly')}
				</span>
			{:else}
				{#if share.can_edit}
					<span class="text-xs bg-green-100 text-green-800 px-2 py-0.5 rounded">
						{$t('giftCards.sharing.permEdit')}
					</span>
				{/if}
				{#if share.can_delete}
					<span class="text-xs bg-red-100 text-red-800 px-2 py-0.5 rounded">
						{$t('giftCards.sharing.permDelete')}
					</span>
				{/if}
				{#if showTransactionsBadge && share.can_edit_transactions}
					<span class="text-xs bg-cyan-100 text-cyan-800 px-2 py-0.5 rounded">
						{labelPermTransactions ?? $t('giftCards.sharing.permTransactions')}
					</span>
				{/if}
				{#if !share.can_edit && !share.can_delete && !(showTransactionsBadge && share.can_edit_transactions)}
					<span class="text-xs bg-gray-100 text-gray-600 px-2 py-0.5 rounded">
						{viewOnlyLabel ?? $t('common.viewOnly')}
					</span>
				{/if}
			{/if}
		</div>
	</div>
{/if}
