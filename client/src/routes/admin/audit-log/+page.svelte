<script lang="ts">
	import { onMount } from 'svelte';
	import { isOnline } from '$lib/stores/offline';
	import { t } from '$lib/stores/i18n';
	import { adminApi } from '$lib/api';
	import { toastStore } from '$lib/stores/toast';
	import { logger } from '$lib/utils/logger';
	import BottomSheet from '$lib/components/BottomSheet.svelte';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import type { AuditLogDTO, AdminUserDTO } from '$lib/types/api';

	const pageLogger = logger.child('AuditLogPage');

	// State
	let logs = $state<AuditLogDTO[]>([]);
	let users = $state<AdminUserDTO[]>([]);
	let isLoading = $state(true);
	let expandedLogId = $state<string | null>(null);
	let showFilterMenu = $state(false);
	let restoreConfirmId = $state<string | null>(null);

	// Filters
	let userFilter = $state('');
	let resourceTypeFilter = $state('');
	let actionFilter = $state('');
	let dateFrom = $state('');
	let dateTo = $state('');
	let search = $state('');

	// Pagination
	let currentPage = $state(1);
	let perPage = $state(20);
	let totalPages = $state(0);

	const isOffline = $derived(!$isOnline);
	const hasActiveFilters = $derived(
		userFilter !== '' ||
			resourceTypeFilter !== '' ||
			actionFilter !== '' ||
			dateFrom !== '' ||
			dateTo !== '' ||
			search !== ''
	);

	const resourceTypes = $derived([
		{ value: '', label: $t('admin.auditLog.allTypes') },
		{ value: 'cards', label: $t('admin.auditLog.typeCards') },
		{ value: 'vouchers', label: $t('admin.auditLog.typeVouchers') },
		{ value: 'gift_cards', label: $t('admin.auditLog.typeGiftCards') },
		{ value: 'card_shares', label: $t('admin.auditLog.typeCardShares') },
		{ value: 'voucher_shares', label: $t('admin.auditLog.typeVoucherShares') },
		{
			value: 'gift_card_shares',
			label: $t('admin.auditLog.typeGiftCardShares')
		},
		{
			value: 'gift_card_transactions',
			label: $t('admin.auditLog.typeTransactions')
		},
		{ value: 'merchants', label: $t('admin.auditLog.typeMerchants') }
	]);

	const actions = $derived([
		{ value: '', label: $t('admin.auditLog.allActions') },
		{ value: 'delete', label: $t('admin.auditLog.actionDelete') },
		{ value: 'hard_delete', label: $t('admin.auditLog.actionHardDelete') },
		{ value: 'restore', label: $t('admin.auditLog.actionRestore') }
	]);

	// ✅ Server-side admin check in +layout.server.ts (SVL-002 Fix)
	// No client-side check needed - user is already validated server-side
	onMount(async () => {
		pageLogger.debug(
			'Admin access granted (server-side check), loading audit logs'
		);
		loadFilters();
		await Promise.all([loadLogs(), loadUsers()]);
	});

	async function loadUsers() {
		try {
			const response = await adminApi.listUsers();
			users = response.users;
		} catch (err) {
			pageLogger.error('Failed to load users', { error: err });
		}
	}

	function loadFilters() {
		try {
			const saved = localStorage.getItem('savvy_audit_filters');
			if (saved) {
				const filters = JSON.parse(saved);
				search = filters.search || '';
				userFilter = filters.userFilter || '';
				resourceTypeFilter = filters.resourceTypeFilter || '';
				actionFilter = filters.actionFilter || '';
				dateFrom = filters.dateFrom || '';
				dateTo = filters.dateTo || '';
			}
		} catch (e) {
			pageLogger.error('Failed to load filters', { error: e });
		}
	}

	function saveFilters() {
		try {
			const filters = {
				search,
				userFilter,
				resourceTypeFilter,
				actionFilter,
				dateFrom,
				dateTo
			};
			localStorage.setItem('savvy_audit_filters', JSON.stringify(filters));
		} catch (e) {
			pageLogger.error('Failed to save filters', { error: e });
		}
	}

	async function loadLogs() {
		isLoading = true;
		saveFilters();
		try {
			const response = await adminApi.getAuditLogs({
				user_id: userFilter || undefined,
				resource_type: resourceTypeFilter || undefined,
				action: actionFilter || undefined,
				date_from: dateFrom || undefined,
				date_to: dateTo || undefined,
				search: search || undefined,
				page: currentPage,
				per_page: perPage
			});

			logs = response.logs;
			totalPages = response.total_pages;
		} catch {
			toastStore.error($t('admin.auditLog.loadError'));
		} finally {
			isLoading = false;
		}
	}

	function toggleExpandLog(logId: string) {
		expandedLogId = expandedLogId === logId ? null : logId;
	}

	function formatResourceData(data: string): string {
		try {
			const parsed = JSON.parse(data);
			return JSON.stringify(parsed, null, 2);
		} catch {
			return data;
		}
	}

	function resetFilters() {
		userFilter = '';
		resourceTypeFilter = '';
		actionFilter = '';
		dateFrom = '';
		dateTo = '';
		search = '';
		currentPage = 1;
	}

	function goToPage(page: number) {
		currentPage = page;
		loadLogs();
	}

	async function handleRestore(
		logId: string,
		resourceType: string,
		resourceId: string
	) {
		if (restoreConfirmId !== logId) {
			restoreConfirmId = logId;
			setTimeout(() => (restoreConfirmId = null), 3000);
			return;
		}

		try {
			await adminApi.restoreResource(resourceType, resourceId);
			toastStore.success($t('admin.auditLog.restoreSuccess'));
			await loadLogs();
			restoreConfirmId = null;
			expandedLogId = null;
		} catch (err) {
			toastStore.error($t('admin.auditLog.restoreError'));
			pageLogger.error('Failed to restore resource', { error: err });
		}
	}

	$effect(() => {
		// Reload when filters change
		currentPage = 1;
		loadLogs();
	});
</script>

<svelte:head>
	<title>{$t('admin.auditLog.title')} - {$t('common.appName')}</title>
</svelte:head>

<div class="px-4 pb-20 md:pb-4">
	<!-- Header -->
	<div class="mb-8">
		<h1 class="text-3xl font-bold text-text">
			{$t('admin.auditLog.title')}
		</h1>
	</div>

	<!-- Search bar and filter button -->
	<div class="flex flex-col sm:flex-row gap-3 mb-4">
		<!-- Search Bar -->
		<div class="flex-1">
			<input
				type="text"
				bind:value={search}
				placeholder={$t('admin.auditLog.searchInData')}
				class="input bg-white"
			/>
		</div>

		<!-- Filter Button -->
		<button
			type="button"
			onclick={(e: MouseEvent) => {
				e.stopPropagation();
				showFilterMenu = !showFilterMenu;
			}}
			class="flex items-center justify-center gap-2 h-[42px] px-4 bg-white border border-border-field rounded-md hover:bg-surface-1 transition-colors relative"
			title={$t('common.filter')}
			aria-label={$t('common.filter')}
			aria-expanded={showFilterMenu}
		>
			<svg
				class="w-5 h-5 text-text-muted"
				fill="none"
				stroke="currentColor"
				viewBox="0 0 24 24"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z"
				/>
			</svg>
			{#if hasActiveFilters}
				<span class="absolute -top-1 -right-1 w-3 h-3 bg-accent rounded-full"
				></span>
			{/if}
		</button>
	</div>

	<!-- Grid with optional Side-Panel -->
	<div class="grid grid-cols-1 {showFilterMenu ? 'lg:grid-cols-4' : ''} gap-6">
		<!-- Logs List (3/4 when filter is open on desktop) -->
		<div class={showFilterMenu ? 'lg:col-span-3' : ''}>
			<!-- Logs Table -->
			{#if isLoading}
				<LoadingSpinner />
			{:else if logs.length === 0}
				<div class="text-center py-12 text-text-subtle">
					{$t('admin.auditLog.noLogs')}
				</div>
			{:else}
				<div class="bg-white shadow rounded-lg overflow-hidden">
					<div class="overflow-x-auto">
						<table class="min-w-full divide-y divide-border">
							<thead class="bg-surface-1">
								<tr>
									<th
										class="px-6 py-3 text-left text-xs font-medium text-text-subtle uppercase tracking-wider"
									>
										{$t('admin.auditLog.action')}
									</th>
									<th
										class="hidden md:table-cell px-6 py-3 text-left text-xs font-medium text-text-subtle uppercase tracking-wider"
									>
										{$t('admin.auditLog.resourceType')}
									</th>
									<th
										class="hidden md:table-cell px-6 py-3 text-left text-xs font-medium text-text-subtle uppercase tracking-wider"
									>
										{$t('admin.auditLog.user')}
									</th>
									<th
										class="hidden md:table-cell px-6 py-3 text-left text-xs font-medium text-text-subtle uppercase tracking-wider"
									>
										{$t('admin.auditLog.timestamp')}
									</th>
									<th
										class="hidden md:table-cell px-6 py-3 text-right text-xs font-medium text-text-subtle uppercase tracking-wider"
									>
										{$t('admin.users.actions')}
									</th>
								</tr>
							</thead>
							<tbody class="bg-white divide-y divide-border">
								{#each logs as log (log.id)}
									<tr
										class="hover:bg-surface-1 transition-colors md:cursor-default cursor-pointer"
										onclick={() => {
											if (window.innerWidth < 768) toggleExpandLog(log.id);
										}}
									>
										<td class="px-6 py-4 whitespace-nowrap text-sm">
											<div class="flex items-center justify-between">
												<div class="flex items-center gap-2">
													<span
														class="px-2 py-1 text-xs font-medium rounded-full {log.action ===
														'delete'
															? 'bg-yellow-100 text-yellow-800'
															: log.action === 'hard_delete'
																? 'bg-red-100 text-red-800'
																: 'bg-green-100 text-green-800'}"
													>
														{log.action}
													</span>
													<span
														class="md:hidden text-sm font-medium text-text"
														>{log.resource_type}</span
													>
												</div>
												<svg
													class="w-4 h-4 text-text-faint ml-2 md:hidden flex-shrink-0 transition-transform {expandedLogId ===
													log.id
														? 'rotate-180'
														: ''}"
													fill="none"
													stroke="currentColor"
													viewBox="0 0 24 24"
												>
													<path
														stroke-linecap="round"
														stroke-linejoin="round"
														stroke-width="2"
														d="M19 9l-7 7-7-7"
													/>
												</svg>
											</div>
										</td>
										<td
											class="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm font-medium text-text"
										>
											{log.resource_type}
										</td>
										<td
											class="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-text-subtle"
										>
											{#if log.user}
												{log.user.email}
											{:else}
												<span class="text-text-faint italic"
													>{$t('admin.auditLog.systemAction')}</span
												>
											{/if}
										</td>
										<td
											class="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-text-subtle"
										>
											{new Date(log.created_at).toLocaleString()}
										</td>
										<td
											class="hidden md:table-cell px-6 py-4 whitespace-nowrap text-right text-sm"
										>
											<button
												onclick={() => toggleExpandLog(log.id)}
												class="text-accent hover:text-accent-900 font-medium transition-colors"
											>
												{expandedLogId === log.id
													? $t('common.close')
													: $t('admin.auditLog.viewData')}
											</button>
										</td>
									</tr>
									{#if expandedLogId === log.id}
										<tr class="bg-surface-1">
											<td colspan="5" class="px-6 py-4">
												<div class="space-y-3">
													<!-- Mobile-only: User & Timestamp -->
													<div class="grid grid-cols-1 gap-4 md:hidden">
														<div>
															<span class="text-xs font-medium text-text-subtle">
																{$t('admin.auditLog.user')}:
															</span>
															<p class="text-sm text-text">
																{#if log.user}
																	{log.user.email}
																{:else}
																	<span class="text-text-faint italic"
																		>{$t('admin.auditLog.systemAction')}</span
																	>
																{/if}
															</p>
														</div>
														<div>
															<span class="text-xs font-medium text-text-subtle">
																{$t('admin.auditLog.timestamp')}:
															</span>
															<p class="text-sm text-text">
																{new Date(log.created_at).toLocaleString()}
															</p>
														</div>
													</div>
													<!-- Log Details -->
													<div class="grid grid-cols-1 md:grid-cols-3 gap-4">
														<div>
															<span class="text-xs font-medium text-text-subtle">
																{$t('admin.auditLog.resourceId')}:
															</span>
															<p class="text-sm text-text font-mono">
																{log.resource_id}
															</p>
														</div>
														<div>
															<span class="text-xs font-medium text-text-subtle">
																{$t('admin.auditLog.ipAddress')}:
															</span>
															<p class="text-sm text-text font-mono">
																{log.ip_address}
															</p>
														</div>
														<div>
															<span class="text-xs font-medium text-text-subtle">
																{$t('admin.auditLog.userAgent')}:
															</span>
															<p
																class="text-sm text-text truncate"
																title={log.user_agent}
															>
																{log.user_agent}
															</p>
														</div>
													</div>

													<!-- Resource Data -->
													<div class="pt-3 border-t border-border">
														<span
															class="text-xs font-medium text-text-subtle block mb-2"
														>
															{$t('admin.auditLog.resourceData')}:
														</span>
														<div
															class="bg-white p-3 rounded border border-border text-xs overflow-x-auto"
														>
															<pre
																class="whitespace-pre-wrap text-text-ink2">{formatResourceData(
																	log.resource_data
																)}</pre>
														</div>
													</div>

													<!-- Actions -->
													{#if log.action === 'delete' || log.action === 'hard_delete'}
														<div class="pt-3 border-t border-border">
															<button
																onclick={() =>
																	handleRestore(
																		log.id,
																		log.resource_type,
																		log.resource_id
																	)}
																disabled={isOffline}
																class="btn btn-sm {restoreConfirmId === log.id
																	? 'btn-primary'
																	: 'btn-ghost'}"
															>
																{restoreConfirmId === log.id
																	? $t('admin.auditLog.confirmRestore')
																	: $t('admin.auditLog.restore')}
															</button>
														</div>
													{/if}
												</div>
											</td>
										</tr>
									{/if}
								{/each}
							</tbody>
						</table>
					</div>
				</div>
			{/if}

			<!-- Pagination -->
			{#if totalPages > 1}
				<div class="mt-8 flex justify-center items-center gap-3">
					<button
						onclick={() => goToPage(currentPage - 1)}
						disabled={currentPage === 1}
						class="btn btn-ghost disabled:opacity-50 disabled:cursor-not-allowed"
					>
						<svg
							class="w-5 h-5 mr-1"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M15 19l-7-7 7-7"
							></path>
						</svg>
						{$t('common.back')}
					</button>

					<span
						class="px-4 py-2 text-sm font-medium text-text-ink2 bg-surface-1 rounded-lg"
					>
						{$t('admin.auditLog.pageOf')
							.replace('{page}', currentPage.toString())
							.replace('{total}', totalPages.toString())}
					</span>

					<button
						onclick={() => goToPage(currentPage + 1)}
						disabled={currentPage === totalPages}
						class="btn btn-ghost disabled:opacity-50 disabled:cursor-not-allowed"
					>
						{$t('admin.auditLog.next')}
						<svg
							class="w-5 h-5 ml-1"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M9 5l7 7-7 7"
							></path>
						</svg>
					</button>
				</div>
			{/if}
		</div>

		<!-- Filter Side-Panel (Desktop only, 1/4 width, sticky) -->
		{#if showFilterMenu}
			<div class="hidden lg:block lg:col-span-1">
				<div class="bg-white rounded-lg shadow-lg p-6 sticky top-4">
					<div class="flex items-center justify-between mb-4">
						<h3 class="text-lg font-semibold text-text">
							{$t('common.filter')}
						</h3>
						<button
							type="button"
							onclick={() => (showFilterMenu = false)}
							class="text-text-faint hover:text-text-muted transition-colors"
							aria-label={$t('common.close')}
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

					<div class="space-y-4">
						<div>
							<label
								for="audit-log-user-filter-desktop"
								class="block text-sm font-medium text-text-ink2 mb-2"
							>
								{$t('admin.auditLog.user')}
							</label>
							<select
								id="audit-log-user-filter-desktop"
								bind:value={userFilter}
								class="input text-sm"
							>
								<option value="">{$t('admin.auditLog.allUsers')}</option>
								{#each users as user (user.id)}
									<option value={user.id}>{user.email}</option>
								{/each}
							</select>
						</div>

						<div>
							<label
								for="audit-log-resource-type-filter-desktop"
								class="block text-sm font-medium text-text-ink2 mb-2"
							>
								{$t('admin.auditLog.resourceType')}
							</label>
							<select
								id="audit-log-resource-type-filter-desktop"
								bind:value={resourceTypeFilter}
								class="input text-sm"
							>
								{#each resourceTypes as type (type.value)}
									<option value={type.value}>{type.label}</option>
								{/each}
							</select>
						</div>

						<div>
							<label
								for="audit-log-action-filter-desktop"
								class="block text-sm font-medium text-text-ink2 mb-2"
							>
								{$t('admin.auditLog.action')}
							</label>
							<select
								id="audit-log-action-filter-desktop"
								bind:value={actionFilter}
								class="input text-sm"
							>
								{#each actions as action (action.value)}
									<option value={action.value}>{action.label}</option>
								{/each}
							</select>
						</div>

						<div>
							<label
								for="audit-log-date-from-desktop"
								class="block text-sm font-medium text-text-ink2 mb-2"
							>
								{$t('admin.auditLog.dateFrom')}
							</label>
							<input
								id="audit-log-date-from-desktop"
								type="date"
								bind:value={dateFrom}
								class="input text-sm"
							/>
						</div>

						<div>
							<label
								for="audit-log-date-to-desktop"
								class="block text-sm font-medium text-text-ink2 mb-2"
							>
								{$t('admin.auditLog.dateTo')}
							</label>
							<input
								id="audit-log-date-to-desktop"
								type="date"
								bind:value={dateTo}
								class="input text-sm"
							/>
						</div>

						{#if hasActiveFilters}
							<button
								type="button"
								onclick={resetFilters}
								class="w-full btn btn-sm btn-ghost text-sm"
							>
								{$t('common.resetFilters')}
							</button>
						{/if}
					</div>
				</div>
			</div>
		{/if}
	</div>

	<!-- Mobile Filter Bottom Sheet -->
	<BottomSheet
		open={showFilterMenu}
		onClose={() => (showFilterMenu = false)}
		maxHeight="80vh"
		ariaLabel={$t('common.filter')}
	>
		<div class="p-6">
			<div class="flex items-center justify-between mb-4">
				<h3
					id="filter-dialog-title"
					class="text-lg font-semibold text-text"
				>
					{$t('common.filter')}
				</h3>
				<button
					type="button"
					onclick={() => (showFilterMenu = false)}
					class="text-text-faint hover:text-text-muted transition-colors"
					aria-label={$t('common.close')}
				>
					<svg
						class="w-6 h-6"
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

			<div class="space-y-4">
				<div>
					<label
						for="audit-log-user-filter-mobile"
						class="block text-sm font-medium text-text-ink2 mb-2"
					>
						{$t('admin.auditLog.user')}
					</label>
					<select
						id="audit-log-user-filter-mobile"
						bind:value={userFilter}
						class="input"
					>
						<option value="">{$t('admin.auditLog.allUsers')}</option>
						{#each users as user (user.id)}
							<option value={user.id}>{user.email}</option>
						{/each}
					</select>
				</div>

				<div>
					<label
						for="audit-log-resource-type-filter-mobile"
						class="block text-sm font-medium text-text-ink2 mb-2"
					>
						{$t('admin.auditLog.resourceType')}
					</label>
					<select
						id="audit-log-resource-type-filter-mobile"
						bind:value={resourceTypeFilter}
						class="input"
					>
						{#each resourceTypes as type (type.value)}
							<option value={type.value}>{type.label}</option>
						{/each}
					</select>
				</div>

				<div>
					<label
						for="audit-log-action-filter-mobile"
						class="block text-sm font-medium text-text-ink2 mb-2"
					>
						{$t('admin.auditLog.action')}
					</label>
					<select
						id="audit-log-action-filter-mobile"
						bind:value={actionFilter}
						class="input"
					>
						{#each actions as action (action.value)}
							<option value={action.value}>{action.label}</option>
						{/each}
					</select>
				</div>

				<div>
					<label
						for="audit-log-date-from-mobile"
						class="block text-sm font-medium text-text-ink2 mb-2"
					>
						{$t('admin.auditLog.dateFrom')}
					</label>
					<input
						id="audit-log-date-from-mobile"
						type="date"
						bind:value={dateFrom}
						class="input"
					/>
				</div>

				<div>
					<label
						for="audit-log-date-to-mobile"
						class="block text-sm font-medium text-text-ink2 mb-2"
					>
						{$t('admin.auditLog.dateTo')}
					</label>
					<input
						id="audit-log-date-to-mobile"
						type="date"
						bind:value={dateTo}
						class="input"
					/>
				</div>

				{#if hasActiveFilters}
					<button
						type="button"
						onclick={resetFilters}
						class="w-full btn btn-ghost"
					>
						{$t('common.resetFilters')}
					</button>
				{/if}

				<button
					type="button"
					onclick={() => (showFilterMenu = false)}
					class="w-full btn btn-primary"
				>
					{$t('common.done')}
				</button>
			</div>
		</div>
	</BottomSheet>
</div>
