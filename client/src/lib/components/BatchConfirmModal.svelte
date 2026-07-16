<script lang="ts">
	import { ICON_INFO_CIRCLE, ICON_SPINNER, ICON_WARNING } from '$lib/icons';
	import Modal from '$lib/components/ui/Modal.svelte';
	import { t } from '$lib/stores/i18n';
	import { get } from 'svelte/store';
	import { platform } from '$lib/utils/platform';
	import EmailAutocomplete from './EmailAutocomplete.svelte';
	import SharePermissions from './SharePermissions.svelte';

	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);

	type BatchAction = 'delete' | 'share' | 'transfer';

	interface Props {
		action: BatchAction;
		count: number;
		isOpen: boolean;
		isLoading: boolean;
		onConfirm: (
			email: string,
			permissions: {
				canEdit: boolean;
				canDelete: boolean;
				canEditTransactions: boolean;
			}
		) => void;
		onCancel: () => void;
		showTransactionPermission?: boolean;
		hidePermissions?: boolean;
	}

	let {
		action,
		count,
		isOpen,
		isLoading,
		onConfirm,
		onCancel,
		showTransactionPermission = false,
		hidePermissions = false
	}: Props = $props();

	let email = $state('');
	let canEdit = $state(false);
	let canDelete = $state(false);
	let canEditTransactions = $state(false);

	function handleConfirm() {
		onConfirm(email, { canEdit, canDelete, canEditTransactions });
	}

	function reset() {
		email = '';
		canEdit = false;
		canDelete = false;
		canEditTransactions = false;
	}

	// Reset state when modal opens
	$effect(() => {
		if (isOpen) {
			reset();
		}
	});
</script>

<Modal
	open={isOpen}
	onclose={onCancel}
	layer="elevated"
	mobileLayout="sheet"
	labelledby="batch-modal-title"
>
	<!-- ponytail: overflow-y-auto clips the email-autocomplete dropdown to the sheet on very short viewports (keyboard open). Acceptable — the email input sits near the top so it fits in practice; move the suggestion list to a portal if a real cutoff shows up. -->
	<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
	<div
		class="pointer-events-auto w-full sm:max-w-md max-h-[90vh] overflow-y-auto p-6 shadow-xl {platform ===
		'ios'
			? 'bg-white/70 backdrop-blur-xl backdrop-saturate-150 rounded-t-3xl sm:rounded-2xl border border-white/30'
			: 'bg-white rounded-t-3xl sm:rounded-lg'}"
		onclick={(e) => e.stopPropagation()}
		onkeydown={(e) => e.stopPropagation()}
		role="document"
	>
		{#if action === 'delete'}
			<!-- Delete Confirmation -->
			<div class="flex items-start mb-4">
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
				<div class="flex-1">
					<h3
						id="batch-modal-title"
						class="text-lg font-semibold text-danger-600"
					>
						{tr('batch.confirmDeleteTitle')}
					</h3>
				</div>
			</div>
			<p class="text-text-muted mb-6 ml-9">
				{tr('batch.confirmDeleteMessage', { count })}
			</p>
			<div class="flex gap-3 justify-end">
				<button
					type="button"
					onclick={onCancel}
					disabled={isLoading}
					class="px-4 py-2 rounded-md border border-border-field hover:bg-surface-1 transition-colors text-text-ink2"
				>
					{tr('common.cancel')}
				</button>
				<button
					type="button"
					onclick={handleConfirm}
					disabled={isLoading}
					class="px-4 py-2 rounded-md text-white bg-danger-600 hover:bg-danger-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
				>
					{#if isLoading}
						<span class="inline-flex items-center gap-2">
							<svg class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
								<circle
									class="opacity-25"
									cx="12"
									cy="12"
									r="10"
									stroke="currentColor"
									stroke-width="4"
								></circle>
								<path class="opacity-75" fill="currentColor" d={ICON_SPINNER}
								></path>
							</svg>
							{tr('common.loading')}
						</span>
					{:else}
						{tr('batch.deleteSelected')}
					{/if}
				</button>
			</div>
		{:else if action === 'share'}
			<!-- Share -->
			<h3 id="batch-modal-title" class="text-lg font-semibold text-text mb-2">
				{tr('batch.confirmShareTitle')}
			</h3>
			<p class="text-sm text-text-muted mb-4">
				{tr('batch.confirmShareMessage', { count })}
			</p>

			<div
				class="border border-accent-200 bg-accent-50 rounded-lg p-4 space-y-4"
			>
				<!-- Email with Autocomplete -->
				<EmailAutocomplete
					bind:value={email}
					label={tr('batch.email')}
					hint={tr('giftCards.sharing.userMustBeRegistered')}
					inputId="batch-email"
					disabled={isLoading}
				/>

				<!-- Permissions -->
				{#if hidePermissions}
					<div
						class="flex items-start gap-2 bg-surface-1 border border-border rounded-lg p-3"
					>
						<svg
							class="w-4 h-4 text-text-subtle mt-0.5 flex-shrink-0"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d={ICON_INFO_CIRCLE}
							></path>
						</svg>
						<p class="text-sm text-text-muted">{tr('batch.readOnlyShare')}</p>
					</div>
				{:else}
					<SharePermissions
						bind:canEdit
						bind:canDelete
						bind:canEditTransactions
						showEditTransactions={showTransactionPermission}
						labelEdit={tr('cards.sharing.canEdit')}
						labelEditDesc={tr('cards.sharing.canEditDesc')}
						labelDelete={tr('cards.sharing.canDelete')}
						labelDeleteDesc={tr('cards.sharing.canDeleteDesc')}
						labelEditTransactions={showTransactionPermission
							? tr('giftCards.sharing.canManageTransactions')
							: undefined}
						labelEditTransactionsDesc={showTransactionPermission
							? tr('giftCards.sharing.canManageTransactionsDesc')
							: undefined}
					/>
				{/if}

				<!-- Action Buttons -->
				<div class="flex gap-2">
					<button
						type="button"
						onclick={handleConfirm}
						disabled={isLoading || !email.trim()}
						class="btn btn-primary flex-1 disabled:opacity-50 disabled:cursor-not-allowed"
					>
						{#if isLoading}
							<span class="inline-flex items-center gap-2">
								<svg
									class="animate-spin w-4 h-4"
									fill="none"
									viewBox="0 0 24 24"
								>
									<circle
										class="opacity-25"
										cx="12"
										cy="12"
										r="10"
										stroke="currentColor"
										stroke-width="4"
									></circle>
									<path class="opacity-75" fill="currentColor" d={ICON_SPINNER}
									></path>
								</svg>
								{tr('common.loading')}
							</span>
						{:else}
							{tr('giftCards.sharing.shareNow')}
						{/if}
					</button>
					<button
						type="button"
						onclick={onCancel}
						disabled={isLoading}
						class="btn btn-ghost"
					>
						{tr('common.cancel')}
					</button>
				</div>
			</div>
		{:else}
			<!-- Transfer -->
			<h3
				id="batch-modal-title"
				class="text-lg font-semibold text-purple-900 mb-2"
			>
				{tr('batch.confirmTransferTitle')}
			</h3>
			<p class="text-sm text-text-muted mb-4">
				{tr('batch.confirmTransferMessage', { count })}
			</p>

			<div
				class="border border-purple-200 bg-purple-50 rounded-lg p-4 space-y-4"
			>
				<!-- Warning Banner -->
				<div class="bg-warning-50 border border-warning-200 rounded-lg p-3">
					<p class="text-sm font-medium text-warning-800">
						<strong>{tr('cards.transfer.warning')}</strong>
					</p>
					<p class="text-xs text-warning-700 mt-1">
						{tr('giftCards.transfer.warningDetails')}
					</p>
				</div>

				<!-- Email with Autocomplete -->
				<EmailAutocomplete
					bind:value={email}
					label={tr('cards.transfer.newOwnerEmail')}
					hint={tr('giftCards.sharing.userMustBeRegistered')}
					inputId="batch-transfer-email"
					disabled={isLoading}
				/>

				<!-- What Happens -->
				<div>
					<p class="text-sm font-medium text-text-ink2 mb-2">
						{tr('cards.transfer.whatHappens')}
					</p>
					<ul class="text-xs text-text-muted space-y-1">
						<li>{tr('cards.transfer.newOwnerGetsRights')}</li>
						<li>{tr('cards.transfer.allSharesDeleted')}</li>
						<li>{tr('cards.transfer.youLoseAccess')}</li>
						<li>{tr('cards.transfer.transferLogged')}</li>
					</ul>
				</div>

				<!-- Action Buttons -->
				<div class="flex gap-2">
					<button
						type="button"
						onclick={handleConfirm}
						disabled={isLoading || !email.trim()}
						class="btn btn-purple flex-1 disabled:opacity-50 disabled:cursor-not-allowed"
					>
						{#if isLoading}
							<span class="inline-flex items-center gap-2">
								<svg
									class="animate-spin w-4 h-4"
									fill="none"
									viewBox="0 0 24 24"
								>
									<circle
										class="opacity-25"
										cx="12"
										cy="12"
										r="10"
										stroke="currentColor"
										stroke-width="4"
									></circle>
									<path class="opacity-75" fill="currentColor" d={ICON_SPINNER}
									></path>
								</svg>
								{tr('common.loading')}
							</span>
						{:else}
							{tr('cards.transfer.transferButton')}
						{/if}
					</button>
					<button
						type="button"
						onclick={onCancel}
						disabled={isLoading}
						class="btn btn-ghost"
					>
						{tr('common.cancel')}
					</button>
				</div>
			</div>
		{/if}
	</div>
</Modal>
