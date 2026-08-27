<script lang="ts">
	import {
		ICON_INFO_CIRCLE,
		ICON_LOCK,
		ICON_SPINNER,
		ICON_TRANSFER,
		ICON_WARNING
	} from '$lib/icons';
	import Modal from '$lib/components/ui/Modal.svelte';
	import { t } from '$lib/stores/i18n';
	import { get } from 'svelte/store';
	import { platform } from '$lib/utils/platform';
	import EmailAutocomplete from './EmailAutocomplete.svelte';
	import SharePermissions from './SharePermissions.svelte';

	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);

	// Module-level constant, so this is a plain boolean, not reactive state.
	const isIOS = platform === 'ios';
	// Android renders the BatchShareAndroid mockup: centred M3 extra-large
	// dialogs on the tonal container, pill actions with a text button for the
	// dismissive one, and a danger box on the transfer arm. iOS and Android
	// each have their own arm below, so the final branch is desktop by
	// construction and needs no flag of its own — it renders the
	// BatchShareDesktop mockup: a --surface panel on --radius-3xl with
	// --shadow-panel and a --border hairline.
	const isAndroid = platform === 'android';

	// M3 dialog actions: a text button for the dismissive action, a filled pill
	// for the confirming one — same convention as the Android import dialog.
	const androidTextButtonClass =
		'text-label text-text-muted rounded-m3-full inline-flex h-10 items-center justify-center px-4 disabled:opacity-50';
	const androidDangerTextClass =
		'text-label text-danger-700 rounded-m3-full inline-flex h-10 items-center justify-center px-4 disabled:opacity-50';
	const androidAccentFillClass =
		'bg-accent-600 text-on-accent text-label rounded-m3-full inline-flex h-10 items-center justify-center px-5.5 disabled:opacity-50 disabled:cursor-not-allowed';
	const androidTransferFillClass =
		'bg-transfer-600 text-on-accent text-label rounded-m3-full inline-flex h-10 items-center justify-center px-5.5 disabled:opacity-50 disabled:cursor-not-allowed';

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
		/**
		 * Selection holds vouchers alongside cards or gift cards. Permissions
		 * still apply (to the non-voucher items), so the voucher read-only note
		 * runs next to them rather than replacing them — that is the
		 * `hidePermissions` case, a selection of vouchers only.
		 */
		mixedWithVouchers?: boolean;
		/**
		 * Size of the largest resource-type group in the selection. The batch
		 * endpoints cap a request at 50 items and each type group is dispatched
		 * as its own request, so this — not `count` — is the number that can
		 * exceed the cap. Defaults to `count` for a caller that does not split
		 * by type.
		 */
		largestGroupCount?: number;
	}

	let {
		action,
		count,
		isOpen,
		isLoading,
		onConfirm,
		onCancel,
		showTransactionPermission = false,
		hidePermissions = false,
		mixedWithVouchers = false,
		largestGroupCount
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

	// Mirrors maxBatchSize in internal/handlers/api/batch.go.
	const BATCH_MAX_ITEMS = 50;

	const groupCount = $derived(largestGroupCount ?? count);
	const overBatchLimit = $derived(groupCount > BATCH_MAX_ITEMS);

	// Over the cap the request comes back as an opaque 400, so block confirm
	// here instead of letting the user run into it.
	const confirmDisabled = $derived(
		isLoading || !email.trim() || overBatchLimit
	);

	// iOS draws delete as a centred alert and the two email flows as bottom
	// sheets; every other platform keeps the single sheet/dialog shell.
	const iosAlert = $derived(isIOS && action === 'delete');

	const desktopChrome = $derived(
		action === 'transfer'
			? 'bg-surface rounded-t-3xl sm:rounded-3xl shadow-panel border-2 border-transfer-200'
			: 'bg-surface rounded-t-3xl sm:rounded-3xl shadow-panel border border-border'
	);
</script>

<!-- Shown on every platform when the largest type group is over the batch cap:
     confirm is blocked there, so say why. -->
{#snippet limitNotice(extraClass: string)}
	<div
		class="flex items-start gap-2 rounded-lg border border-danger-200 bg-danger-50 p-3 {extraClass}"
	>
		<svg
			class="mt-0.5 h-4 w-4 flex-none text-danger-700"
			fill="none"
			stroke="currentColor"
			stroke-width="2"
			viewBox="0 0 24 24"
			aria-hidden="true"
		>
			<path stroke-linecap="round" stroke-linejoin="round" d={ICON_WARNING} />
		</svg>
		<p class="text-body-sm text-danger-800">
			{tr('batch.tooManyItems', { max: BATCH_MAX_ITEMS })}
		</p>
	</div>
{/snippet}

<!-- Android read-only voucher note: same M3 warning box whether it replaces the
     permission rows (vouchers only) or runs beside them (mixed selection). -->
{#snippet androidVoucherNote(extraClass: string)}
	<div
		class="flex items-start gap-2 rounded-m3-sm border border-warning-200 bg-warning-50 px-3 py-2.75 {extraClass}"
	>
		<svg
			class="mt-0.25 h-3.75 w-3.75 flex-none text-warning-700"
			fill="none"
			stroke="currentColor"
			stroke-width="2"
			viewBox="0 0 24 24"
			aria-hidden="true"
		>
			<path stroke-linecap="round" stroke-linejoin="round" d={ICON_LOCK} />
		</svg>
		<p class="text-chip font-normal text-warning-800">
			{tr('batch.readOnlyShare')}
		</p>
	</div>
{/snippet}

<!-- Sheet chrome shared by the iOS share and transfer flows: grabber plus a
     nav bar carrying Cancel on the left and the confirm action on the right
     (mockup screen-BatchShareIOS). -->
{#snippet iosSheetBar(
	title: string,
	titleClass: string,
	confirmLabel: string,
	confirmClass: string
)}
	<div class="flex justify-center pt-2 pb-0.5">
		<span
			class="h-1 w-10 rounded-full bg-[var(--color-glass-grabber)]"
			aria-hidden="true"
		></span>
	</div>
	<!-- Three items on one 390pt line: every cell keeps its own line so a longer
	     translation shortens the title instead of wrapping the whole bar. -->
	<div
		class="flex items-center justify-between gap-2 border-b border-border-soft px-4.5 pt-1.5 pb-3"
	>
		<button
			type="button"
			onclick={onCancel}
			disabled={isLoading}
			class="shrink-0 whitespace-nowrap text-subheading font-normal tracking-normal text-accent disabled:opacity-50"
		>
			{tr('common.cancel')}
		</button>
		<h3
			id="batch-modal-title"
			class="truncate text-subheading {titleClass} whitespace-nowrap"
		>
			{title}
		</h3>
		<button
			type="button"
			onclick={handleConfirm}
			disabled={confirmDisabled}
			class="shrink-0 whitespace-nowrap text-subheading font-bold tracking-normal {confirmClass} disabled:opacity-50"
		>
			{#if isLoading}
				<span class="inline-flex items-center gap-2">
					<svg class="h-4 w-4 animate-spin" fill="none" viewBox="0 0 24 24">
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
				{confirmLabel}
			{/if}
		</button>
	</div>
{/snippet}

<Modal
	open={isOpen}
	onclose={onCancel}
	layer="elevated"
	mobileLayout={iosAlert || isAndroid ? 'center' : 'sheet'}
	clearBottomNav={!(iosAlert || isAndroid)}
	backdrop={isIOS ? 'ios-scrim' : 'platform'}
	labelledby="batch-modal-title"
>
	{#if iosAlert}
		<!-- iOS alert: a 270px centred card whose two actions sit side by side
		     under a hairline, not a sheet (mockup screen-BatchShareIOS, frame 1). -->
		<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
		<div
			class="liquid-glass-surface pointer-events-auto w-[270px] overflow-hidden rounded-xl shadow-modal"
			onclick={(e) => e.stopPropagation()}
			onkeydown={(e) => e.stopPropagation()}
			role="document"
		>
			<div class="px-4 pt-4.75 pb-3.75 text-center">
				<h3
					id="batch-modal-title"
					class="mb-1 text-subheading font-semibold text-text"
				>
					{tr('batch.confirmDeleteTitle')}
				</h3>
				<!-- The alert is 270pt wide, so over the cap the limit message replaces
				     the confirmation copy instead of stacking a banner under it. -->
				<p
					class="text-label font-normal {overBatchLimit
						? 'text-danger-800'
						: 'text-text-muted'}"
				>
					{overBatchLimit
						? tr('batch.tooManyItems', { max: BATCH_MAX_ITEMS })
						: tr('batch.confirmDeleteMessage', { count })}
				</p>
			</div>
			<div class="flex border-t border-[var(--color-glass-edge)]">
				<button
					type="button"
					onclick={onCancel}
					disabled={isLoading}
					class="flex-1 border-r border-[var(--color-glass-edge)] py-3 text-subheading font-normal tracking-normal text-accent disabled:opacity-50"
				>
					{tr('common.cancel')}
				</button>
				<button
					type="button"
					onclick={handleConfirm}
					disabled={isLoading || overBatchLimit}
					class="flex-1 py-3 text-subheading font-semibold text-danger-600 disabled:opacity-50"
				>
					{#if isLoading}
						<span class="inline-flex items-center gap-2">
							<svg class="h-4 w-4 animate-spin" fill="none" viewBox="0 0 24 24">
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
						</span>
					{:else}
						{tr('common.delete')}
					{/if}
				</button>
			</div>
		</div>
	{:else if isIOS}
		<!-- iOS share / transfer sheet: nav-bar actions, content flat on the
		     glass (mockup screen-BatchShareIOS, frames 2 and 3). -->
		<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
		<div
			class="liquid-glass-surface pointer-events-auto flex max-h-[92vh] w-full flex-col overflow-hidden rounded-t-[var(--radius-sheet)] shadow-sheet sm:max-w-md sm:rounded-[var(--radius-sheet)]"
			onclick={(e) => e.stopPropagation()}
			onkeydown={(e) => e.stopPropagation()}
			role="document"
		>
			{#if action === 'share'}
				{@render iosSheetBar(
					tr('batch.confirmShareTitle'),
					'font-semibold text-text',
					tr('common.share'),
					'text-accent'
				)}
				<div class="overflow-y-auto px-5 pt-4.5 pb-6.5">
					<div class="mb-4 flex items-center justify-between gap-2">
						<p class="text-label font-normal text-text-muted">
							{tr('batch.confirmShareMessage', { count })}
						</p>
						<!-- The largest type group, not the total: each group goes out as
						     its own request and the cap applies per request. -->
						<span
							class="flex-none rounded-full border px-2.25 py-0.75 text-eyebrow font-normal tracking-normal normal-case tabular-nums {overBatchLimit
								? 'border-danger-200 bg-danger-50 text-danger-800'
								: 'border-border bg-surface text-text-subtle'}"
						>
							{groupCount} / {BATCH_MAX_ITEMS}
						</span>
					</div>

					{#if overBatchLimit}
						{@render limitNotice('mb-4')}
					{/if}

					<EmailAutocomplete
						bind:value={email}
						label={tr('batch.email')}
						hint={tr('giftCards.sharing.userMustBeRegistered')}
						inputId="batch-email"
						disabled={isLoading}
						iosField
					/>

					{#if hidePermissions}
						<div
							class="mt-4 flex items-start gap-2 rounded-lg border border-border bg-surface-1 p-3"
						>
							<svg
								class="mt-0.5 h-4 w-4 flex-shrink-0 text-text-subtle"
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
							<p class="text-body-sm text-text-muted">
								{tr('batch.readOnlyShare')}
							</p>
						</div>
					{:else}
						<p
							class="mt-4 mb-2.75 text-eyebrow font-bold uppercase tracking-[0.09em] text-text-subtle"
						>
							{tr('batch.permissions')}
						</p>
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
							iosBoxes
						/>
						{#if mixedWithVouchers}
							<div
								class="mt-4 flex items-start gap-2 rounded-md border border-warning-200 bg-warning-50 px-3 py-2.75"
							>
								<svg
									class="mt-0.25 h-3.75 w-3.75 flex-none text-warning-700"
									fill="none"
									stroke="currentColor"
									stroke-width="2"
									viewBox="0 0 24 24"
									aria-hidden="true"
								>
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										d={ICON_LOCK}
									/>
								</svg>
								<p class="text-body-sm text-warning-800">
									{tr('batch.readOnlyShare')}
								</p>
							</div>
						{/if}
					{/if}
				</div>
			{:else}
				<!-- The nav bar is one line wide, so it takes the short wordings the
				     mockup uses: `common.transferOwnership` over the longer
				     `batch.confirmTransferTitle`, and the gift-card namespace's
				     "Übergeben" over the card namespace's "Jetzt übertragen". -->
				{@render iosSheetBar(
					tr('common.transferOwnership'),
					'font-semibold text-transfer-900',
					tr('giftCards.transfer.transferButton'),
					'text-transfer-700'
				)}
				<div class="overflow-y-auto px-5 pt-4.5 pb-6.5">
					<p class="mb-3.5 text-label font-normal text-text-muted">
						{tr('batch.confirmTransferMessage', { count })}
					</p>
					{#if overBatchLimit}
						{@render limitNotice('mb-4')}
					{/if}
					<div
						class="mb-4 flex items-start gap-2.5 rounded-lg border border-danger-200 bg-danger-50 px-3.5 py-3.25"
					>
						<svg
							class="mt-0.25 h-4.5 w-4.5 flex-none text-danger-700"
							fill="none"
							stroke="currentColor"
							stroke-width="2"
							viewBox="0 0 24 24"
							aria-hidden="true"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								d={ICON_WARNING}
							/>
						</svg>
						<p class="text-body-sm text-danger-800">
							<!-- The gift-card wording, not `cards.transfer.warning`: that one
							     carries a ⚠️ emoji and would double the icon next to it. -->
							<strong>{tr('giftCards.transfer.warning')}</strong>
							{tr('giftCards.transfer.warningDetails')}
						</p>
					</div>

					<!-- Last element of the scrolling sheet body, so the suggestion list
					     opens upwards — downwards it would render past the sheet's
					     bottom edge and be clipped away. -->
					<EmailAutocomplete
						bind:value={email}
						label={tr('cards.transfer.newOwnerEmail')}
						hint={tr('giftCards.sharing.userMustBeRegistered')}
						inputId="batch-transfer-email"
						disabled={isLoading}
						iosField
						dropUp
					/>
				</div>
			{/if}
		</div>
	{:else if isAndroid}
		<!-- Android M3 dialog: extra-large tonal container, centred rather than a
		     sheet, with a text button for the dismissive action and a filled pill
		     for the confirming one (mockup screen-BatchShareAndroid). -->
		<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
		<div
			class="m3-outlined-form bg-m3-surface-container pointer-events-auto max-h-[90vh] w-full overflow-y-auto rounded-m3-xl p-5.5 shadow-m3-dialog sm:max-w-md"
			onclick={(e) => e.stopPropagation()}
			onkeydown={(e) => e.stopPropagation()}
			role="document"
		>
			{#if action === 'delete'}
				<h3
					id="batch-modal-title"
					class="mb-2 text-heading font-semibold text-text"
				>
					{tr('batch.confirmDeleteTitle')}
				</h3>
				<p class="mb-6 text-body-sm text-text-muted">
					{tr('batch.confirmDeleteMessage', { count })}
				</p>
				{#if overBatchLimit}
					{@render limitNotice('mb-6')}
				{/if}
				<div class="flex justify-end gap-1.5">
					<button
						type="button"
						onclick={onCancel}
						disabled={isLoading}
						class={androidTextButtonClass}
					>
						{tr('common.cancel')}
					</button>
					<button
						type="button"
						onclick={handleConfirm}
						disabled={isLoading || overBatchLimit}
						class={androidDangerTextClass}
					>
						{#if isLoading}
							<span class="inline-flex items-center gap-2">
								<svg
									class="h-4 w-4 animate-spin"
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
							{tr('common.delete')}
						{/if}
					</button>
				</div>
			{:else if action === 'share'}
				<div class="mb-1 flex items-center justify-between gap-2">
					<h3
						id="batch-modal-title"
						class="text-heading font-semibold text-text"
					>
						{tr('batch.confirmShareTitle')}
					</h3>
					<!-- The largest type group, not the total: each group goes out as
					     its own request and the cap applies per request. -->
					<span
						class="flex-none rounded-m3-full border px-2.25 py-0.75 text-eyebrow font-normal tracking-normal normal-case tabular-nums {overBatchLimit
							? 'border-danger-200 bg-danger-50 text-danger-800'
							: 'border-border bg-surface-2 text-text-subtle'}"
					>
						{groupCount} / {BATCH_MAX_ITEMS}
					</span>
				</div>
				<p class="mb-4 text-label font-normal text-text-muted">
					{tr('batch.confirmShareMessage', { count })}
				</p>

				{#if overBatchLimit}
					{@render limitNotice('mb-4')}
				{/if}

				<EmailAutocomplete
					bind:value={email}
					label={tr('batch.email')}
					hint={tr('giftCards.sharing.userMustBeRegistered')}
					inputId="batch-email"
					disabled={isLoading}
				/>

				{#if hidePermissions}
					{@render androidVoucherNote('mt-4 mb-5')}
				{:else}
					<p
						class="mt-4 mb-2.75 text-tag font-bold uppercase tracking-[0.09em] text-text-subtle"
					>
						{tr('batch.permissions')}
					</p>
					<SharePermissions
						class="m3-permissions"
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
					{#if mixedWithVouchers}
						{@render androidVoucherNote('mt-4 mb-5')}
					{/if}
				{/if}

				<div
					class="flex justify-end gap-1.5 {hidePermissions || mixedWithVouchers
						? ''
						: 'mt-5'}"
				>
					<button
						type="button"
						onclick={onCancel}
						disabled={isLoading}
						class={androidTextButtonClass}
					>
						{tr('common.cancel')}
					</button>
					<button
						type="button"
						onclick={handleConfirm}
						disabled={confirmDisabled}
						class={androidAccentFillClass}
					>
						{#if isLoading}
							<span class="inline-flex items-center gap-2">
								<svg
									class="h-4 w-4 animate-spin"
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
				</div>
			{:else}
				<div class="mb-3 flex items-center gap-2.25">
					<svg
						class="h-5.5 w-5.5 flex-none text-transfer-600"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
						viewBox="0 0 24 24"
						aria-hidden="true"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							d={ICON_TRANSFER}
						/>
					</svg>
					<h3
						id="batch-modal-title"
						class="text-heading font-semibold text-transfer-900"
					>
						{tr('giftCards.transfer.title')}
					</h3>
				</div>
				<p class="mb-3.5 text-label font-normal text-text-muted">
					{tr('batch.confirmTransferMessage', { count })}
				</p>

				{#if overBatchLimit}
					{@render limitNotice('mb-4')}
				{/if}

				<!-- The gift-card wording, not `cards.transfer.warning`: that one
				     carries a ⚠️ emoji the M3 box has no icon slot for. -->
				<div
					class="mb-4 rounded-m3-md border border-danger-200 bg-danger-50 px-3.25 py-3"
				>
					<p class="mb-1 text-chip text-danger-800">
						{tr('giftCards.transfer.warning')}
					</p>
					<p class="text-chip font-normal text-danger-800">
						{tr('giftCards.transfer.warningDetails')}
					</p>
				</div>

				<EmailAutocomplete
					bind:value={email}
					label={tr('giftCards.transfer.newOwnerEmail')}
					hint={tr('giftCards.sharing.userMustBeRegistered')}
					inputId="batch-transfer-email"
					disabled={isLoading}
				/>

				<div class="mt-4.5 flex justify-end gap-1.5">
					<button
						type="button"
						onclick={onCancel}
						disabled={isLoading}
						class={androidTextButtonClass}
					>
						{tr('common.cancel')}
					</button>
					<button
						type="button"
						data-testid="batch-transfer-confirm"
						onclick={handleConfirm}
						disabled={confirmDisabled}
						class={androidTransferFillClass}
					>
						{#if isLoading}
							<span class="inline-flex items-center gap-2">
								<svg
									class="h-4 w-4 animate-spin"
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
							{tr('giftCards.transfer.transferButton')}
						{/if}
					</button>
				</div>
			{/if}
		</div>
	{:else}
		<!-- ponytail: overflow-y-auto clips the email-autocomplete dropdown to the sheet on very short viewports (keyboard open). Acceptable — the email input sits near the top so it fits in practice; move the suggestion list to a portal if a real cutoff shows up. -->
		<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
		<div
			class="pointer-events-auto w-full sm:max-w-md max-h-[90vh] overflow-y-auto p-6 {desktopChrome}"
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
							class="text-lg font-semibold text-danger-800"
						>
							{tr('batch.confirmDeleteTitle')}
						</h3>
					</div>
				</div>
				<p
					class="text-text-muted {overBatchLimit
						? 'mb-4'
						: 'mb-6'} ml-9 text-label font-normal"
				>
					{tr('batch.confirmDeleteMessage', { count })}
				</p>
				{#if overBatchLimit}
					{@render limitNotice('mb-6 ml-9')}
				{/if}
				<div class="flex gap-3 justify-end">
					<button
						type="button"
						onclick={onCancel}
						disabled={isLoading}
						class="text-label h-11 rounded-lg border border-border-field px-4.5 text-text-ink2 transition-colors hover:bg-surface-1"
					>
						{tr('common.cancel')}
					</button>
					<button
						type="button"
						onclick={handleConfirm}
						disabled={isLoading || overBatchLimit}
						class="text-label text-on-accent h-11 rounded-lg bg-danger-600 px-4.5 transition-colors hover:bg-danger-700 disabled:cursor-not-allowed disabled:opacity-50"
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
							{tr('batch.deleteSelected')}
						{/if}
					</button>
				</div>
			{:else if action === 'share'}
				<!-- Share -->
				<div class="flex items-center justify-between gap-2 mb-2">
					<h3 id="batch-modal-title" class="text-lg font-semibold text-text">
						{tr('batch.confirmShareTitle')}
					</h3>
					<!-- The largest type group, not the total: each group goes out as
					     its own request and the cap applies per request. -->
					<span
						class="flex-none rounded-full border px-2.5 py-0.75 text-eyebrow font-normal tracking-normal normal-case tabular-nums {overBatchLimit
							? 'border-danger-200 bg-danger-50 text-danger-800'
							: 'border-border bg-surface-1 text-text-subtle'}"
					>
						{groupCount} / {BATCH_MAX_ITEMS}
					</span>
				</div>
				<p class="text-label mb-4 font-normal text-text-muted">
					{tr('batch.confirmShareMessage', { count })}
				</p>

				{#if overBatchLimit}
					{@render limitNotice('mb-4')}
				{/if}

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

					<!-- Permissions. A vouchers-only selection has none — its read-only
					     note renders outside the accent inset below. -->
					{#if !hidePermissions}
						<p
							class="text-eyebrow font-bold uppercase tracking-[0.09em] text-text-subtle"
						>
							{tr('batch.permissions')}
						</p>
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
							disabled={confirmDisabled}
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
										<path
											class="opacity-75"
											fill="currentColor"
											d={ICON_SPINNER}
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

				<!-- Vouchers in the selection are always shared read-only. The desktop
				     mockup puts this outside the accent inset, under the actions. -->
				{#if mixedWithVouchers || hidePermissions}
					<div
						class="mt-3.5 flex items-start gap-2.25 rounded-md border border-warning-200 bg-warning-50 px-3.25 py-2.75"
					>
						<svg
							class="mt-0.25 h-4 w-4 flex-none text-warning-700"
							fill="none"
							stroke="currentColor"
							stroke-width="2"
							viewBox="0 0 24 24"
							aria-hidden="true"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								d={ICON_LOCK}
							/>
						</svg>
						<p class="text-body-sm text-warning-800">
							{tr('batch.readOnlyShare')}
						</p>
					</div>
				{/if}
			{:else}
				<!-- Transfer -->
				<h3
					id="batch-modal-title"
					class="text-lg font-semibold text-transfer-900 mb-2"
				>
					{tr('batch.confirmTransferTitle')}
				</h3>
				<p class="text-label mb-4 font-normal text-text-muted">
					{tr('batch.confirmTransferMessage', { count })}
				</p>

				{#if overBatchLimit}
					{@render limitNotice('mb-4')}
				{/if}

				<div
					class="border border-transfer-200 bg-transfer-50 rounded-lg p-4 space-y-4"
				>
					<!-- Warning Banner -->
					<div
						class="bg-danger-50 border border-danger-200 rounded-md px-3.25 py-3"
					>
						<!-- The gift-card wording, not `cards.transfer.warning`: that
						     one carries a warning emoji the mockup does not show. -->
						<p class="text-body-sm font-bold text-danger-800 mb-1">
							{tr('giftCards.transfer.warning')}
						</p>
						<p class="text-body-sm text-danger-800">
							{tr('giftCards.transfer.warningDetails')}
						</p>
					</div>

					<!-- Email with Autocomplete -->
					<EmailAutocomplete
						bind:value={email}
						label={tr('giftCards.transfer.newOwnerEmail')}
						hint={tr('giftCards.sharing.userMustBeRegistered')}
						inputId="batch-transfer-email"
						disabled={isLoading}
					/>

					<!-- Action Buttons -->
					<div class="flex gap-2">
						<button
							type="button"
							data-testid="batch-transfer-confirm"
							onclick={handleConfirm}
							disabled={confirmDisabled}
							class="btn btn-transfer flex-1 disabled:opacity-50 disabled:cursor-not-allowed"
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
										<path
											class="opacity-75"
											fill="currentColor"
											d={ICON_SPINNER}
										></path>
									</svg>
									{tr('common.loading')}
								</span>
							{:else}
								{tr('giftCards.transfer.transferButton')}
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
	{/if}
</Modal>
