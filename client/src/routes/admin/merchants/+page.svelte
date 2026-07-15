<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import { isOnline } from '$lib/stores/offline';
	import { t } from '$lib/stores/i18n';
	import { merchantsApi } from '$lib/api';
	import { toastStore } from '$lib/stores/toast';
	import { logger } from '$lib/utils/logger';
	import BottomSheet from '$lib/components/BottomSheet.svelte';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import ConfirmModal from '$lib/components/ConfirmModal.svelte';
	import { MERCHANT_DEFAULT_COLOR } from '$lib/utils/merchant-color';
	import type { MerchantDTO } from '$lib/types/api';

	const pageLogger = logger.child('AdminMerchantsPage');

	// State
	let merchants = $state<MerchantDTO[]>([]);
	let filteredMerchants = $state<MerchantDTO[]>([]);
	let isLoading = $state(true);
	let search = $state('');
	let sortBy = $state('name-asc');
	let showFilterMenu = $state(false);
	let expandedMerchantId = $state<string | null>(null);
	let deleteTarget = $state<MerchantDTO | null>(null);

	const isOffline = $derived(!$isOnline);
	const hasActiveFilters = $derived(search !== '' || sortBy !== 'name-asc');

	// ✅ Server-side admin check in +layout.server.ts (same as /admin/users)
	onMount(async () => {
		pageLogger.debug(
			'Admin access granted (server-side check), loading merchants'
		);
		await loadMerchants();
	});

	async function loadMerchants() {
		isLoading = true;
		try {
			const response = await merchantsApi.list();
			// $effect (below) re-runs applyFilters() reactively on merchants change.
			merchants = response.merchants;
		} catch {
			toastStore.error($t('admin.merchants.loadError'));
		} finally {
			isLoading = false;
		}
	}

	function applyFilters() {
		let result = merchants;

		// Search filter (by name, client-side like users page)
		if (search) {
			const query = search.toLowerCase();
			result = result.filter((m) => m.name.toLowerCase().includes(query));
		}

		// Sort
		result = [...result].sort((a, b) => {
			switch (sortBy) {
				case 'name-asc':
					return a.name.localeCompare(b.name);
				case 'name-desc':
					return b.name.localeCompare(a.name);
				case 'newest':
					return (
						new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
					);
				case 'oldest':
					return (
						new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
					);
				default:
					return 0;
			}
		});

		filteredMerchants = result;
	}

	function toggleExpandMerchant(id: string) {
		expandedMerchantId = expandedMerchantId === id ? null : id;
	}

	async function handleDelete() {
		if (!deleteTarget) return;
		const merchant = deleteTarget;
		deleteTarget = null;
		try {
			await merchantsApi.delete(merchant.id);
			toastStore.success(
				$t('admin.merchants.deleteSuccess', { name: merchant.name })
			);
			await loadMerchants();
		} catch {
			toastStore.error($t('admin.merchants.deleteError'));
		}
	}

	$effect(() => {
		applyFilters();
	});
</script>

<svelte:head>
	<title>{$t('admin.merchants.title')} - {$t('common.appName')}</title>
</svelte:head>

<div class="px-4 pb-20 md:pb-4">
	<!-- Header -->
	<div class="mb-8">
		<h1 class="text-3xl font-bold text-text">{$t('admin.merchants.title')}</h1>
		<p class="text-text-subtle mt-1">{$t('admin.merchants.subtitle')}</p>
	</div>

	<!-- Search bar and action buttons -->
	<div class="mb-6 flex flex-col sm:flex-row gap-3">
		<!-- Search Bar -->
		<div class="flex-1">
			<input
				type="text"
				bind:value={search}
				placeholder={$t('common.search')}
				class="input bg-white"
			/>
		</div>

		<!-- Action Buttons (Desktop) -->
		<div class="hidden sm:flex gap-3">
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

			<!-- Create Merchant Button -->
			<a
				href={resolve('/admin/merchants/new')}
				class="btn btn-primary whitespace-nowrap {isOffline
					? 'opacity-50 cursor-not-allowed pointer-events-none blur-[0.5px]'
					: ''}"
			>
				{$t('admin.merchants.createMerchant')}
			</a>
		</div>

		<!-- Action Buttons (Mobile) -->
		<div class="flex sm:hidden gap-3">
			<!-- Filter Button (Mobile) -->
			<button
				type="button"
				onclick={(e: MouseEvent) => {
					e.stopPropagation();
					showFilterMenu = !showFilterMenu;
				}}
				class="flex-1 flex items-center justify-center h-[42px] px-3 bg-white border border-border-field rounded-md hover:bg-surface-1 transition-colors relative"
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

			<!-- Create Merchant Button (Mobile) -->
			<a
				href={resolve('/admin/merchants/new')}
				class="btn btn-sm btn-primary flex-1 text-center {isOffline
					? 'opacity-50 cursor-not-allowed pointer-events-none blur-[0.5px]'
					: ''}"
			>
				{$t('admin.merchants.createMerchant')}
			</a>
		</div>
	</div>

	<!-- Grid with optional Side-Panel -->
	<div class="grid grid-cols-1 {showFilterMenu ? 'lg:grid-cols-4' : ''} gap-6">
		<!-- Merchants Table (3/4 when filter is open on desktop) -->
		<div class={showFilterMenu ? 'lg:col-span-3' : ''}>
			{#if isLoading}
				<LoadingSpinner />
			{:else if filteredMerchants.length === 0}
				<div class="text-center py-12 text-text-subtle">
					{$t('admin.merchants.noMerchants')}
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
										{$t('admin.merchants.name')}
									</th>
									<th
										class="hidden md:table-cell px-6 py-3 text-left text-xs font-medium text-text-subtle uppercase tracking-wider"
									>
										{$t('admin.merchants.website')}
									</th>
									<th
										class="hidden md:table-cell px-6 py-3 text-left text-xs font-medium text-text-subtle uppercase tracking-wider"
									>
										{$t('admin.merchants.createdAt')}
									</th>
									<th
										class="hidden md:table-cell px-6 py-3 text-right text-xs font-medium text-text-subtle uppercase tracking-wider"
									>
										{$t('admin.merchants.actions')}
									</th>
								</tr>
							</thead>
							<tbody class="bg-white divide-y divide-border">
								{#each filteredMerchants as merchant (merchant.id)}
									<tr
										class="hover:bg-surface-1 transition-colors md:cursor-default cursor-pointer"
										onclick={() => {
											if (window.innerWidth < 768)
												toggleExpandMerchant(merchant.id);
										}}
									>
										<td
											class="px-6 py-4 whitespace-nowrap text-sm font-medium text-text"
										>
											<div class="flex items-center justify-between">
												<span class="flex items-center gap-3 min-w-0">
													<span
														class="w-4 h-4 rounded-full flex-shrink-0 border border-border"
														style="background-color: {merchant.color ||
															MERCHANT_DEFAULT_COLOR}"
													></span>
													<span class="truncate">{merchant.name}</span>
												</span>
												<svg
													class="w-4 h-4 text-text-faint ml-2 md:hidden flex-shrink-0 transition-transform {expandedMerchantId ===
													merchant.id
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
											class="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-text-subtle max-w-xs truncate"
										>
											{#if merchant.website}
												<!-- eslint-disable svelte/no-navigation-without-resolve -- external merchant website URL, not an app route -->
												<a
													href={merchant.website}
													target="_blank"
													rel="noopener noreferrer"
													class="text-accent hover:text-accent-900 transition-colors"
													onclick={(e) => e.stopPropagation()}
												>
													{merchant.website}
												</a>
												<!-- eslint-enable svelte/no-navigation-without-resolve -->
											{:else}
												—
											{/if}
										</td>
										<td
											class="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-text-subtle"
										>
											{new Date(merchant.created_at).toLocaleDateString()}
										</td>
										<td
											class="hidden md:table-cell px-6 py-4 whitespace-nowrap text-right text-sm"
										>
											<button
												onclick={() => toggleExpandMerchant(merchant.id)}
												class="text-accent hover:text-accent-900 font-medium transition-colors"
											>
												{expandedMerchantId === merchant.id
													? $t('common.close')
													: $t('admin.merchants.showDetails')}
											</button>
										</td>
									</tr>
									{#if expandedMerchantId === merchant.id}
										<tr class="bg-surface-1">
											<td colspan="4" class="px-6 py-4">
												<div class="space-y-3">
													<!-- Merchant Details -->
													<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
														{#if merchant.description}
															<div class="md:col-span-2">
																<span
																	class="text-xs font-medium text-text-subtle"
																	>{$t('common.description')}:</span
																>
																<p class="text-sm text-text">
																	{merchant.description}
																</p>
															</div>
														{/if}
														<div class="md:hidden">
															<span class="text-xs font-medium text-text-subtle"
																>{$t('admin.merchants.website')}:</span
															>
															<p class="text-sm text-text">
																{#if merchant.website}
																	<!-- eslint-disable svelte/no-navigation-without-resolve -- external merchant website URL, not an app route -->
																	<a
																		href={merchant.website}
																		target="_blank"
																		rel="noopener noreferrer"
																		class="text-accent hover:text-accent-900 transition-colors break-all"
																	>
																		{merchant.website}
																	</a>
																	<!-- eslint-enable svelte/no-navigation-without-resolve -->
																{:else}
																	—
																{/if}
															</p>
														</div>
														<div class="md:hidden">
															<span class="text-xs font-medium text-text-subtle"
																>{$t('admin.merchants.createdAt')}:</span
															>
															<p class="text-sm text-text">
																{new Date(merchant.created_at).toLocaleString()}
															</p>
														</div>
														<div>
															<span class="text-xs font-medium text-text-subtle"
																>{$t('admin.merchants.id')}:</span
															>
															<p class="text-sm text-text font-mono break-all">
																{merchant.id}
															</p>
														</div>
														<div>
															<span class="text-xs font-medium text-text-subtle"
																>{$t('admin.merchants.color')}:</span
															>
															<p
																class="text-sm text-text font-mono flex items-center gap-2"
															>
																<span
																	class="w-4 h-4 rounded-full inline-block border border-border"
																	style="background-color: {merchant.color ||
																		MERCHANT_DEFAULT_COLOR}"
																></span>
																{merchant.color || '—'}
															</p>
														</div>
													</div>

													<!-- Actions -->
													<div
														class="flex flex-wrap gap-2 pt-2 border-t border-border"
													>
														<a
															href={resolve(
																`/admin/merchants/${merchant.id}/edit`
															)}
															class="btn btn-sm btn-primary {isOffline
																? 'opacity-50 cursor-not-allowed pointer-events-none'
																: ''}"
														>
															{$t('common.edit')}
														</a>
														<button
															onclick={() => (deleteTarget = merchant)}
															disabled={isOffline}
															class="btn btn-sm btn-ghost text-danger-600 hover:text-danger-700 {isOffline
																? 'opacity-50 cursor-not-allowed'
																: ''}"
														>
															{$t('common.delete')}
														</button>
													</div>
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
								for="sortBy"
								class="block text-sm font-medium text-text-ink2 mb-2"
							>
								{$t('common.sort')}
							</label>
							<select id="sortBy" bind:value={sortBy} class="input text-sm">
								<option value="name-asc"
									>{$t('admin.merchants.sortNameAsc')}</option
								>
								<option value="name-desc"
									>{$t('admin.merchants.sortNameDesc')}</option
								>
								<option value="newest"
									>{$t('admin.merchants.sortNewest')}</option
								>
								<option value="oldest"
									>{$t('admin.merchants.sortOldest')}</option
								>
							</select>
						</div>

						{#if hasActiveFilters}
							<button
								type="button"
								onclick={() => {
									search = '';
									sortBy = 'name-asc';
								}}
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
						for="sortByMobile"
						class="block text-sm font-medium text-text-ink2 mb-2"
					>
						{$t('common.sort')}
					</label>
					<select id="sortByMobile" bind:value={sortBy} class="input bg-white">
						<option value="name-asc">{$t('admin.merchants.sortNameAsc')}</option
						>
						<option value="name-desc"
							>{$t('admin.merchants.sortNameDesc')}</option
						>
						<option value="newest">{$t('admin.merchants.sortNewest')}</option>
						<option value="oldest">{$t('admin.merchants.sortOldest')}</option>
					</select>
				</div>

				{#if hasActiveFilters}
					<button
						type="button"
						onclick={() => {
							search = '';
							sortBy = 'name-asc';
						}}
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

	<!-- Delete confirmation -->
	<ConfirmModal
		isOpen={deleteTarget !== null}
		title={$t('admin.merchants.confirmDeleteTitle')}
		message={$t('admin.merchants.confirmDeleteMessage', {
			name: deleteTarget?.name ?? ''
		})}
		confirmText={$t('common.delete')}
		cancelText={$t('common.cancel')}
		variant="danger"
		onconfirm={handleDelete}
		oncancel={() => (deleteTarget = null)}
	/>
</div>
