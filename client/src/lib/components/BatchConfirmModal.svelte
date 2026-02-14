<script lang="ts">
	import { t } from '$lib/stores/i18n';
	import { get } from 'svelte/store';
	import { sharedUsersApi } from '$lib/api';
	import type { UserDTO } from '$lib/types/api';
	import { logger } from '$lib/utils/logger';
	import { platform } from '$lib/utils/platform';

	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);
	const componentLogger = logger.child('BatchConfirmModal');

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

	// Autocomplete state
	let suggestedUsers = $state<UserDTO[]>([]);
	let showSuggestions = $state(false);
	let searchTimeout: ReturnType<typeof setTimeout> | null = null;

	const needsEmail = $derived(action === 'share' || action === 'transfer');

	function handleConfirm() {
		onConfirm(email, { canEdit, canDelete, canEditTransactions });
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') {
			onCancel();
		}
	}

	function reset() {
		email = '';
		canEdit = false;
		canDelete = false;
		canEditTransactions = false;
		suggestedUsers = [];
		showSuggestions = false;
		if (searchTimeout) clearTimeout(searchTimeout);
	}

	// Reset state when modal opens
	$effect(() => {
		if (isOpen) {
			reset();
		}
	});

	// Autocomplete functions
	async function searchSharedUsers(query: string) {
		if (searchTimeout) clearTimeout(searchTimeout);

		if (query.length < 2) {
			suggestedUsers = [];
			showSuggestions = false;
			return;
		}

		searchTimeout = setTimeout(async () => {
			try {
				const response = await sharedUsersApi.search(query);
				suggestedUsers = response.users;
				showSuggestions = true;
			} catch (err) {
				componentLogger.error('Failed to search users:', err);
				suggestedUsers = [];
			}
		}, 300); // 300ms debounce
	}

	function selectUser(user: UserDTO) {
		email = user.email;
		showSuggestions = false;
		suggestedUsers = [];
	}

	function onEmailInput(event: Event) {
		const input = event.target as HTMLInputElement;
		email = input.value;
		searchSharedUsers(input.value);
	}

	function onEmailFocus() {
		if (email.length >= 2) {
			searchSharedUsers(email);
		}
	}

	function onEmailBlur() {
		// Delay to allow click on suggestion
		setTimeout(() => {
			showSuggestions = false;
		}, 200);
	}
</script>

{#if isOpen}
	<!-- Backdrop -->
	<div
		class="fixed inset-0 z-[70] {platform === 'ios'
			? 'bg-black/40 backdrop-blur-sm'
			: 'bg-black bg-opacity-50'}"
		onclick={onCancel}
		role="presentation"
	></div>

	<!-- Modal -->
	<div
		class="fixed inset-0 z-[80] flex items-center justify-center p-4 pb-40 sm:pb-4"
		onkeydown={handleKeydown}
		role="dialog"
		aria-modal="true"
		aria-labelledby="batch-modal-title"
		tabindex="-1"
	>
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div
			class="max-w-md w-full p-6 shadow-xl {platform === 'ios'
				? 'bg-white/70 backdrop-blur-xl backdrop-saturate-150 rounded-2xl border border-white/30'
				: 'bg-white rounded-lg'}"
			onclick={(e) => e.stopPropagation()}
			onkeydown={(e) => e.stopPropagation()}
		>
			{#if action === 'delete'}
				<!-- Delete Confirmation -->
				<div class="flex items-start mb-4">
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
					<div class="flex-1">
						<h3
							id="batch-modal-title"
							class="text-lg font-semibold text-red-600"
						>
							{tr('batch.confirmDeleteTitle')}
						</h3>
					</div>
				</div>
				<p class="text-gray-600 mb-6 ml-9">
					{tr('batch.confirmDeleteMessage', { count })}
				</p>
				<div class="flex gap-3 justify-end">
					<button
						type="button"
						onclick={onCancel}
						disabled={isLoading}
						class="px-4 py-2 rounded-md border border-gray-300 hover:bg-gray-50 transition-colors text-gray-700"
					>
						{tr('common.cancel')}
					</button>
					<button
						type="button"
						onclick={handleConfirm}
						disabled={isLoading}
						class="px-4 py-2 rounded-md text-white bg-red-600 hover:bg-red-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
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
										d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
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
				<h3
					id="batch-modal-title"
					class="text-lg font-semibold text-gray-900 mb-2"
				>
					{tr('batch.confirmShareTitle')}
				</h3>
				<p class="text-sm text-gray-600 mb-4">
					{tr('batch.confirmShareMessage', { count })}
				</p>

				<div class="border border-cyan-200 bg-cyan-50 rounded-lg p-4 space-y-4">
					<!-- Email with Autocomplete -->
					<div class="relative">
						<label
							for="batch-email"
							class="block text-sm font-medium text-gray-700 mb-1"
						>
							{tr('batch.email')} *
						</label>
						<input
							id="batch-email"
							type="email"
							value={email}
							oninput={onEmailInput}
							onfocus={onEmailFocus}
							onblur={onEmailBlur}
							placeholder={tr('batch.emailPlaceholder')}
							autocomplete="off"
							class="input bg-white w-full"
							disabled={isLoading}
						/>

						{#if showSuggestions && suggestedUsers.length > 0}
							<div
								class="absolute z-10 w-full mt-1 bg-white border border-gray-300 rounded-md shadow-lg max-h-48 overflow-y-auto"
							>
								{#each suggestedUsers as user}
									<button
										type="button"
										onclick={() => selectUser(user)}
										class="w-full text-left px-3 py-2 hover:bg-gray-100 focus:bg-gray-100 focus:outline-none"
									>
										<div class="font-medium text-sm text-gray-900">
											{#if user.first_name && user.last_name}
												{user.first_name} {user.last_name}
											{:else if user.first_name}
												{user.first_name}
											{:else}
												{user.email}
											{/if}
										</div>
										<div class="text-xs text-gray-500">{user.email}</div>
									</button>
								{/each}
							</div>
						{/if}

						<p class="text-xs text-gray-500 mt-1">
							{tr('giftCards.sharing.userMustBeRegistered')}
						</p>
					</div>

					<!-- Permissions -->
					{#if hidePermissions}
						<div
							class="flex items-start gap-2 bg-gray-50 border border-gray-200 rounded-lg p-3"
						>
							<svg
								class="w-4 h-4 text-gray-500 mt-0.5 flex-shrink-0"
								fill="none"
								stroke="currentColor"
								viewBox="0 0 24 24"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
								></path>
							</svg>
							<p class="text-sm text-gray-600">{tr('batch.readOnlyShare')}</p>
						</div>
					{:else}
						<div class="space-y-2">
							<label class="flex items-start cursor-pointer">
								<input
									type="checkbox"
									bind:checked={canEdit}
									disabled={isLoading}
									class="mt-0.5 h-4 w-4 text-cyan-600 focus:ring-cyan-500 border-gray-300 rounded"
								/>
								<div class="ml-2">
									<span class="block text-sm font-medium text-gray-900"
										>{tr('cards.sharing.canEdit')}</span
									>
									<span class="text-xs text-gray-500"
										>{tr('cards.sharing.canEditDesc')}</span
									>
								</div>
							</label>
							<label class="flex items-start cursor-pointer">
								<input
									type="checkbox"
									bind:checked={canDelete}
									disabled={isLoading}
									class="mt-0.5 h-4 w-4 text-cyan-600 focus:ring-cyan-500 border-gray-300 rounded"
								/>
								<div class="ml-2">
									<span class="block text-sm font-medium text-gray-900"
										>{tr('cards.sharing.canDelete')}</span
									>
									<span class="text-xs text-gray-500"
										>{tr('cards.sharing.canDeleteDesc')}</span
									>
								</div>
							</label>
							{#if showTransactionPermission}
								<label class="flex items-start cursor-pointer">
									<input
										type="checkbox"
										bind:checked={canEditTransactions}
										disabled={isLoading}
										class="mt-0.5 h-4 w-4 text-cyan-600 focus:ring-cyan-500 border-gray-300 rounded"
									/>
									<div class="ml-2">
										<span class="block text-sm font-medium text-gray-900"
											>{tr('giftCards.sharing.canManageTransactions')}</span
										>
										<span class="text-xs text-gray-500"
											>{tr('giftCards.sharing.canManageTransactionsDesc')}</span
										>
									</div>
								</label>
							{/if}
						</div>
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
										<path
											class="opacity-75"
											fill="currentColor"
											d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
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
				<p class="text-sm text-gray-600 mb-4">
					{tr('batch.confirmTransferMessage', { count })}
				</p>

				<div
					class="border border-purple-200 bg-purple-50 rounded-lg p-4 space-y-4"
				>
					<!-- Warning Banner -->
					<div class="bg-yellow-50 border border-yellow-200 rounded-lg p-3">
						<p class="text-sm font-medium text-yellow-800">
							<strong>{tr('cards.transfer.warning')}</strong>
						</p>
						<p class="text-xs text-yellow-700 mt-1">
							{tr('giftCards.transfer.warningDetails')}
						</p>
					</div>

					<!-- Email with Autocomplete -->
					<div class="relative">
						<label
							for="batch-transfer-email"
							class="block text-sm font-medium text-gray-700 mb-1"
						>
							{tr('cards.transfer.newOwnerEmail')} *
						</label>
						<input
							id="batch-transfer-email"
							type="email"
							value={email}
							oninput={onEmailInput}
							onfocus={onEmailFocus}
							onblur={onEmailBlur}
							placeholder={tr('batch.emailPlaceholder')}
							autocomplete="off"
							class="input bg-white w-full"
							disabled={isLoading}
						/>

						{#if showSuggestions && suggestedUsers.length > 0}
							<div
								class="absolute z-10 w-full mt-1 bg-white border border-gray-300 rounded-md shadow-lg max-h-48 overflow-y-auto"
							>
								{#each suggestedUsers as user}
									<button
										type="button"
										onclick={() => selectUser(user)}
										class="w-full text-left px-3 py-2 hover:bg-gray-100 focus:bg-gray-100 focus:outline-none"
									>
										<div class="font-medium text-sm text-gray-900">
											{#if user.first_name && user.last_name}
												{user.first_name} {user.last_name}
											{:else if user.first_name}
												{user.first_name}
											{:else}
												{user.email}
											{/if}
										</div>
										<div class="text-xs text-gray-500">{user.email}</div>
									</button>
								{/each}
							</div>
						{/if}

						<p class="text-xs text-gray-500 mt-1">
							{tr('giftCards.sharing.userMustBeRegistered')}
						</p>
					</div>

					<!-- What Happens -->
					<div>
						<p class="text-sm font-medium text-gray-700 mb-2">
							{tr('cards.transfer.whatHappens')}
						</p>
						<ul class="text-xs text-gray-600 space-y-1">
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
										<path
											class="opacity-75"
											fill="currentColor"
											d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
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
	</div>
{/if}
