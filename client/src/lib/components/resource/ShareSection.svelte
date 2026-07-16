<script lang="ts">
	import { ICON_LOCK } from '$lib/icons';
	import { get } from 'svelte/store';
	import { t } from '$lib/stores/i18n';
	import { toastStore } from '$lib/stores/toast';
	import { cardsApi, vouchersApi, giftCardsApi } from '$lib/api';
	import { CONFIG, type Kind } from '$lib/resource/config';
	import ConfirmModal from '$lib/components/ConfirmModal.svelte';
	import EmailAutocomplete from '$lib/components/EmailAutocomplete.svelte';
	import SharePermissions from '$lib/components/SharePermissions.svelte';
	import ShareListItem from '$lib/components/ShareListItem.svelte';
	import {
		formatShareResult,
		shareResponseFromError
	} from '$lib/utils/share-result';
	import type {
		CardDTO,
		VoucherDTO,
		GiftCardDTO,
		ShareDTO,
		ShareCreateRequest,
		ShareCreateResponse
	} from '$lib/types/api';

	type ResourceDTO = CardDTO | VoucherDTO | GiftCardDTO;

	interface Props {
		kind: Kind;
		resource: ResourceDTO;
		/** Owned by the route; ShareSection reads and writes it back via bind. */
		shares: ShareDTO[];
		isOffline: boolean;
		/** 'editable' shows edit/delete permission checkboxes; 'readonly' (voucher). */
		shareMode?: 'editable' | 'readonly';
	}

	let {
		kind,
		resource,
		shares = $bindable(),
		isOffline,
		shareMode = 'editable'
	}: Props = $props();

	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);

	const cfg = $derived(CONFIG[kind]);
	const c = $derived(cfg.i18n);

	const shareApi = $derived(
		kind === 'card' ? cardsApi : kind === 'voucher' ? vouchersApi : giftCardsApi
	);

	// --- Share create form state ------------------------------------------
	let showShareForm = $state(false);
	let shareEmails = $state<string[]>([]);
	let canEdit = $state(false);
	let canDelete = $state(false);
	let canEditTransactions = $state(false);

	// Share editing state (editable kinds: cards, gift-cards).
	let editingShareId = $state<string | null>(null);
	let editShareCanEdit = $state(false);
	let editShareCanDelete = $state(false);
	let editShareCanEditTransactions = $state(false);
	// Voucher "manage" mode (delete-only, no permission editing).
	let managingShareId = $state<string | null>(null);

	// Modal state
	let showDeleteShareModal = $state(false);
	let showRevokeAllModal = $state(false);
	let shareToDelete: string | null = null;

	async function loadShares() {
		if (!resource?.permissions?.is_owner) return;
		try {
			const response = await shareApi.get(resource.id);
			shares = response.shares || [];
		} catch (err) {
			// Non-fatal: keep existing shares.
			console.error('Failed to load shares:', err);
		}
	}

	function buildSharePayload(): ShareCreateRequest {
		if (kind === 'voucher') return { emails: shareEmails };
		if (kind === 'gift_card')
			return {
				emails: shareEmails,
				can_edit: canEdit,
				can_delete: canDelete,
				can_edit_transactions: canEditTransactions
			};
		return { emails: shareEmails, can_edit: canEdit, can_delete: canDelete };
	}

	function resetShareForm() {
		showShareForm = false;
		shareEmails = [];
		canEdit = false;
		canDelete = false;
		canEditTransactions = false;
	}

	async function handleShare() {
		if (shareEmails.length === 0 || !resource) return;
		try {
			const response: ShareCreateResponse = await shareApi.createShare(
				resource.id,
				buildSharePayload()
			);
			shares = response.shares || [];
			const { message, isError } = formatShareResult(response, tr);
			if (isError) toastStore.error(message);
			else toastStore.success(message);
			resetShareForm();
		} catch (err: unknown) {
			const failed = shareResponseFromError(err);
			if (failed) {
				shares = failed.shares || shares;
				toastStore.error(formatShareResult(failed, tr).message);
			} else {
				toastStore.error(err instanceof Error ? err.message : tr(c.shareError));
			}
		}
	}

	function startEditShare(share: ShareDTO) {
		editingShareId = share.shared_with_user.id;
		editShareCanEdit = share.can_edit;
		editShareCanDelete = share.can_delete;
		editShareCanEditTransactions = share.can_edit_transactions || false;
	}

	function cancelEditShare() {
		editingShareId = null;
		editShareCanEdit = false;
		editShareCanDelete = false;
		editShareCanEditTransactions = false;
	}

	async function saveShareEdit(sharedWithID: string) {
		try {
			// updateShare only exists on editable kinds (cards, gift-cards).
			if (!resource || !('updateShare' in shareApi)) return;
			const response = await shareApi.updateShare(resource.id, sharedWithID, {
				can_edit: editShareCanEdit,
				can_delete: editShareCanDelete,
				...(cfg.showEditTransactions
					? { can_edit_transactions: editShareCanEditTransactions }
					: {})
			});
			shares = response.shares || [];
			editingShareId = null;
			toastStore.success(tr(c.updateSuccess));
		} catch (err: unknown) {
			toastStore.error(err instanceof Error ? err.message : tr(c.updateError));
		}
	}

	function startManageShare(shareId: string) {
		managingShareId = shareId;
	}

	function cancelManageShare() {
		managingShareId = null;
	}

	function promptDeleteShare(sharedWithID: string) {
		shareToDelete = sharedWithID;
		showDeleteShareModal = true;
	}

	async function confirmDeleteShare() {
		if (!shareToDelete || !resource) return;
		try {
			await shareApi.deleteShare(resource.id, shareToDelete);
			toastStore.success(tr(c.removeSuccess));
			managingShareId = null;
			showDeleteShareModal = false;
			await loadShares();
		} catch {
			toastStore.error(tr(c.removeError));
		} finally {
			shareToDelete = null;
			showDeleteShareModal = false;
		}
	}

	function promptRevokeAll() {
		showRevokeAllModal = true;
	}

	async function confirmRevokeAll() {
		if (!resource) return;
		try {
			await shareApi.deleteAllShares(resource.id);
			toastStore.success(tr(c.revokeAllSuccess));
			managingShareId = null;
			await loadShares();
		} catch {
			toastStore.error(tr(c.revokeAllError));
		} finally {
			showRevokeAllModal = false;
		}
	}

	const LOCK_PATH = ICON_LOCK;
</script>

<!-- Sharing Box -->
<div class="rounded-xl border border-border bg-white p-6">
	<div class="flex justify-between items-center mb-4">
		<h3 class="text-lg font-semibold text-text">
			{tr(c.shareTitle)}
		</h3>
		{#if !showShareForm}
			<button
				onclick={() => (showShareForm = true)}
				disabled={isOffline}
				class="btn btn-xs btn-primary whitespace-nowrap flex items-center gap-1.5 {isOffline
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
							d={LOCK_PATH}
						></path>
					</svg>
				{:else}
					<span>+</span>
				{/if}
				{tr(c.shareAddButton)}
			</button>
		{/if}
	</div>

	{#if showShareForm}
		<div
			class="border border-accent-200 bg-accent-50 rounded-lg p-4 space-y-4 mb-4"
		>
			<EmailAutocomplete
				multiple
				bind:values={shareEmails}
				label={tr(c.shareUserEmail)}
				hint={tr(c.shareHint)}
				inputId="share-email-input"
				disabled={isOffline}
			/>

			{#if cfg.sharePermissions}
				<SharePermissions
					bind:canEdit
					bind:canDelete
					bind:canEditTransactions
					showEditTransactions={cfg.showEditTransactions}
					labelEdit={tr(c.canEdit)}
					labelEditDesc={tr(c.canEditDesc)}
					labelDelete={tr(c.canDelete)}
					labelDeleteDesc={tr(c.canDeleteDesc)}
					labelEditTransactions={cfg.showEditTransactions
						? tr(c.canManageTransactions)
						: undefined}
					labelEditTransactionsDesc={cfg.showEditTransactions
						? tr(c.canManageTransactionsDesc)
						: undefined}
				/>
			{/if}

			<div class="bg-white border border-accent-200 rounded-lg p-3">
				<h4 class="font-medium text-accent-900 text-sm mb-2">
					{tr(c.whatIsShared)}
				</h4>
				{#if kind === 'voucher'}
					<ul class="text-xs text-accent-800 space-y-1">
						<li>{tr(c.sharedCode)}</li>
						<li>{tr(c.sharedDetails)}</li>
						<li>{tr(c.sharedDescription)}</li>
					</ul>
					<p class="text-xs text-accent-hover mt-2 italic">
						{tr(c.readOnlyNote)}
					</p>
				{:else if kind === 'gift_card'}
					<ul class="text-xs text-accent-800 space-y-1">
						<li>{tr('giftCards.sharing.sharedItemCardNumber')}</li>
						<li>{tr('giftCards.sharing.sharedItemBalance')}</li>
						<li>{tr('giftCards.sharing.sharedItemDetails')}</li>
						<li>{tr('giftCards.sharing.sharedItemTransactions')}</li>
						<li>{tr('giftCards.sharing.sharedItemNotes')}</li>
					</ul>
				{:else}
					<ul class="text-xs text-accent-800 space-y-1">
						<li>{tr('cards.sharing.sharedItemCardNumber')}</li>
						<li>{tr('cards.sharing.sharedItemDetails')}</li>
						<li>{tr('cards.sharing.sharedItemNotes')}</li>
					</ul>
				{/if}
			</div>

			<div class="flex gap-2">
				<button
					onclick={handleShare}
					disabled={isOffline}
					class="btn btn-primary flex-1 {isOffline
						? 'opacity-50 cursor-not-allowed'
						: ''}"
				>
					{tr(c.shareNow)}
				</button>
				<button onclick={resetShareForm} class="btn btn-ghost">
					{tr('common.cancel')}
				</button>
			</div>
		</div>
	{/if}

	{#if shares.length > 0}
		<div class="space-y-3">
			{#each shares as share (share.shared_with_user.id)}
				{#if shareMode === 'readonly'}
					<ShareListItem
						{share}
						isEditing={managingShareId === share.shared_with_user.id}
						{isOffline}
						editButtonLabel={tr(c.manage)}
						deleteButtonLabel={tr(c.removeShare)}
						alwaysViewOnly={true}
						onstartEdit={() => startManageShare(share.shared_with_user.id)}
						oncancel={cancelManageShare}
						ondelete={() => promptDeleteShare(share.shared_with_user.id)}
					>
						<div class="bg-warning-50 border border-warning-200 rounded-lg p-3">
							<p class="text-xs font-medium text-warning-800 mb-1">
								{tr(c.alwaysReadOnly)}
							</p>
							<p class="text-xs text-warning-700">
								{tr(c.canOnlyRemove)}
							</p>
						</div>
					</ShareListItem>
				{:else}
					<ShareListItem
						{share}
						isEditing={editingShareId === share.shared_with_user.id}
						{isOffline}
						showTransactionsBadge={cfg.showEditTransactions}
						onstartEdit={() => startEditShare(share)}
						onsave={() => saveShareEdit(share.shared_with_user.id)}
						oncancel={cancelEditShare}
						ondelete={() => promptDeleteShare(share.shared_with_user.id)}
					>
						<SharePermissions
							bind:canEdit={editShareCanEdit}
							bind:canDelete={editShareCanDelete}
							bind:canEditTransactions={editShareCanEditTransactions}
							showEditTransactions={cfg.showEditTransactions}
							labelEdit={tr(c.canEdit)}
							labelEditDesc={tr(c.canEditDesc)}
							labelDelete={tr(c.canDelete)}
							labelDeleteDesc={tr(c.canDeleteDesc)}
							labelEditTransactions={cfg.showEditTransactions
								? tr(c.canManageTransactions)
								: undefined}
							labelEditTransactionsDesc={cfg.showEditTransactions
								? tr(c.canManageTransactionsDesc)
								: undefined}
						/>
					</ShareListItem>
				{/if}
			{/each}
		</div>
		<button
			type="button"
			onclick={promptRevokeAll}
			disabled={isOffline}
			class="btn btn-ghost text-danger-600 mt-3 w-full disabled:opacity-50"
		>
			{tr(c.revokeAll)}
		</button>
	{:else}
		<p class="text-sm text-text-subtle text-center py-4">
			{tr(c.notSharedYet)}
		</p>
	{/if}
</div>

<ConfirmModal
	isOpen={showDeleteShareModal}
	title={tr(c.removeConfirm)}
	message={tr(c.removeConfirmMessage)}
	confirmText={tr('common.remove')}
	cancelText={tr('common.cancel')}
	variant="danger"
	onconfirm={confirmDeleteShare}
	oncancel={() => (showDeleteShareModal = false)}
/>

<ConfirmModal
	isOpen={showRevokeAllModal}
	title={tr(c.revokeAllConfirm)}
	message={tr(c.revokeAllConfirmMessage)}
	confirmText={tr(c.revokeAll)}
	cancelText={tr('common.cancel')}
	variant="danger"
	onconfirm={confirmRevokeAll}
	oncancel={() => (showRevokeAllModal = false)}
/>
