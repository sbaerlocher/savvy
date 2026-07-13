<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import { authStore } from '$lib/stores/auth';
	import { isOnline } from '$lib/stores/offline';
	import { t } from '$lib/stores/i18n';
	import { adminApi } from '$lib/api';
	import { toastStore } from '$lib/stores/toast';
	import { logger } from '$lib/utils/logger';
	import BottomSheet from '$lib/components/BottomSheet.svelte';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import type { AdminUserDTO } from '$lib/types/api';

	const pageLogger = logger.child('AdminPage');

	// State
	let users = $state<AdminUserDTO[]>([]);
	let filteredUsers = $state<AdminUserDTO[]>([]);
	let isLoading = $state(true);
	let search = $state('');
	let sortBy = $state('email-asc');
	let roleFilter = $state('all');
	let providerFilter = $state('all');
	let showFilterMenu = $state(false);
	let expandedUserId = $state<string | null>(null);
	let localLoginEnabled = $state(true);

	const isOffline = $derived(!$isOnline);
	const currentUser = $derived($authStore.user);
	const hasActiveFilters = $derived(
		search !== '' ||
			roleFilter !== 'all' ||
			providerFilter !== 'all' ||
			sortBy !== 'email-asc'
	);

	// ✅ Server-side admin check in +layout.server.ts (SVL-002 Fix)
	// No client-side check needed - user is already validated server-side
	onMount(async () => {
		pageLogger.debug('Admin access granted (server-side check), loading users');

		// Load app config
		try {
			const response = await fetch('/api/v1/config');
			if (response.ok) {
				const config = await response.json();
				localLoginEnabled = config.local_login_enabled ?? true;
			}
		} catch (error) {
			pageLogger.error('Failed to load config', { error });
		}

		loadFilters();
		await loadUsers();
	});

	async function loadUsers() {
		isLoading = true;
		try {
			const response = await adminApi.listUsers();
			users = response.users;
			applyFilters();
		} catch {
			toastStore.error($t('admin.users.loadError'));
		} finally {
			isLoading = false;
		}
	}

	function loadFilters() {
		try {
			const saved = localStorage.getItem('savvy_admin_filters');
			if (saved) {
				const filters = JSON.parse(saved);
				search = filters.search || '';
				sortBy = filters.sortBy || 'email-asc';
				roleFilter = filters.roleFilter || 'all';
				providerFilter = filters.providerFilter || 'all';
			}
		} catch (e) {
			pageLogger.error('Failed to load filters', { error: e });
		}
	}

	function saveFilters() {
		try {
			const filters = { search, sortBy, roleFilter, providerFilter };
			localStorage.setItem('savvy_admin_filters', JSON.stringify(filters));
		} catch (e) {
			pageLogger.error('Failed to save filters', { error: e });
		}
	}

	function applyFilters() {
		saveFilters();
		let result = users;

		// Search filter
		if (search) {
			const query = search.toLowerCase();
			result = result.filter(
				(u) =>
					u.email.toLowerCase().includes(query) ||
					u.first_name?.toLowerCase().includes(query) ||
					u.last_name?.toLowerCase().includes(query)
			);
		}

		// Role filter
		if (roleFilter !== 'all') {
			result = result.filter((u) => u.role === roleFilter);
		}

		// Provider filter
		if (providerFilter !== 'all') {
			result = result.filter((u) => u.auth_provider === providerFilter);
		}

		// Sort
		result = [...result].sort((a, b) => {
			switch (sortBy) {
				case 'email-asc':
					return a.email.localeCompare(b.email);
				case 'email-desc':
					return b.email.localeCompare(a.email);
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

		filteredUsers = result;
	}

	async function toggleRole(userId: string, currentRole: 'user' | 'admin') {
		const newRole = currentRole === 'admin' ? 'user' : 'admin';

		try {
			await adminApi.updateUser(userId, { role: newRole });
			toastStore.success($t('admin.users.roleUpdateSuccess'));
			await loadUsers();
		} catch {
			toastStore.error($t('admin.users.roleUpdateError'));
		}
	}

	async function handleImpersonate(userId: string, _userEmail: string) {
		try {
			await authStore.startImpersonation(userId);
			// Redirect happens in authStore
		} catch (err) {
			const message = err instanceof Error ? err.message : '';
			toastStore.error(message || $t('admin.impersonate_info.failed'));
		}
	}

	function toggleExpandUser(userId: string) {
		expandedUserId = expandedUserId === userId ? null : userId;
	}

	$effect(() => {
		applyFilters();
	});
</script>

<svelte:head>
	<title>{$t('admin.users.title')} - {$t('common.appName')}</title>
</svelte:head>

<div class="px-4 pb-20 md:pb-4">
	<!-- Header -->
	<div class="mb-8">
		<h1 class="text-3xl font-bold text-text">{$t('admin.users.title')}</h1>
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

			<!-- Create User Button -->
			{#if localLoginEnabled}
				<a
					href={resolve('/admin/users/new')}
					class="btn btn-primary whitespace-nowrap {isOffline
						? 'opacity-50 cursor-not-allowed pointer-events-none blur-[0.5px]'
						: ''}"
				>
					{$t('admin.users.createUser')}
				</a>
			{/if}
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

			<!-- Create User Button (Mobile) -->
			{#if localLoginEnabled}
				<a
					href={resolve('/admin/users/new')}
					class="btn btn-sm btn-primary flex-1 text-center {isOffline
						? 'opacity-50 cursor-not-allowed pointer-events-none blur-[0.5px]'
						: ''}"
				>
					{$t('admin.users.createUser')}
				</a>
			{/if}
		</div>
	</div>

	<!-- Grid with optional Side-Panel -->
	<div class="grid grid-cols-1 {showFilterMenu ? 'lg:grid-cols-4' : ''} gap-6">
		<!-- Users Table (3/4 when filter is open on desktop) -->
		<div class={showFilterMenu ? 'lg:col-span-3' : ''}>
			{#if isLoading}
				<LoadingSpinner />
			{:else if filteredUsers.length === 0}
				<div class="text-center py-12 text-text-subtle">
					{$t('admin.users.noUsers')}
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
										{$t('admin.users.email')}
									</th>
									<th
										class="hidden md:table-cell px-6 py-3 text-left text-xs font-medium text-text-subtle uppercase tracking-wider"
									>
										{$t('admin.users.name')}
									</th>
									<th
										class="hidden md:table-cell px-6 py-3 text-left text-xs font-medium text-text-subtle uppercase tracking-wider"
									>
										{$t('admin.users.role')}
									</th>
									<th
										class="hidden md:table-cell px-6 py-3 text-right text-xs font-medium text-text-subtle uppercase tracking-wider"
									>
										{$t('admin.users.actions')}
									</th>
								</tr>
							</thead>
							<tbody class="bg-white divide-y divide-border">
								{#each filteredUsers as user (user.id)}
									{@const isOAuth = user.auth_provider === 'oauth'}
									<tr
										class="hover:bg-surface-1 transition-colors md:cursor-default cursor-pointer"
										onclick={() => {
											if (window.innerWidth < 768) toggleExpandUser(user.id);
										}}
									>
										<td
											class="px-6 py-4 whitespace-nowrap text-sm font-medium text-text"
										>
											<div class="flex items-center justify-between">
												<span class="truncate">{user.email}</span>
												<svg
													class="w-4 h-4 text-text-faint ml-2 md:hidden flex-shrink-0 transition-transform {expandedUserId ===
													user.id
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
											class="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-text-subtle"
										>
											{user.first_name || ''}
											{user.last_name || ''}
										</td>
										<td
											class="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm"
										>
											<span
												class="px-2 py-1 text-xs font-medium rounded-full {user.role ===
												'admin'
													? 'bg-red-100 text-red-800'
													: 'bg-border-soft text-text-strong'}"
											>
												{user.role === 'admin'
													? $t('admin.users.roleAdmin')
													: $t('admin.users.roleUser')}
											</span>
										</td>
										<td
											class="hidden md:table-cell px-6 py-4 whitespace-nowrap text-right text-sm"
										>
											<button
												onclick={() => toggleExpandUser(user.id)}
												class="text-accent hover:text-accent-900 font-medium transition-colors"
											>
												{expandedUserId === user.id
													? $t('common.close')
													: $t('admin.users.showDetails')}
											</button>
										</td>
									</tr>
									{#if expandedUserId === user.id}
										<tr class="bg-surface-1">
											<td colspan="4" class="px-6 py-4">
												<div class="space-y-3">
													<!-- User Details -->
													<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
														<div class="md:hidden">
															<span class="text-xs font-medium text-text-subtle"
																>{$t('admin.users.name')}:</span
															>
															<p class="text-sm text-text">
																{user.first_name || ''}
																{user.last_name || ''}
															</p>
														</div>
														<div class="md:hidden">
															<span class="text-xs font-medium text-text-subtle"
																>{$t('admin.users.role')}:</span
															>
															<p class="text-sm">
																<span
																	class="px-2 py-1 text-xs font-medium rounded-full {user.role ===
																	'admin'
																		? 'bg-red-100 text-red-800'
																		: 'bg-border-soft text-text-strong'}"
																>
																	{user.role === 'admin'
																		? $t('admin.users.roleAdmin')
																		: $t('admin.users.roleUser')}
																</span>
															</p>
														</div>
														<div>
															<span class="text-xs font-medium text-text-subtle"
																>{$t('admin.users.userId')}:</span
															>
															<p class="text-sm text-text font-mono">
																{user.id}
															</p>
														</div>
														<div>
															<span class="text-xs font-medium text-text-subtle"
																>{$t('admin.users.provider')}:</span
															>
															<p class="text-sm text-text">
																{user.auth_provider}
															</p>
														</div>
														<div>
															<span class="text-xs font-medium text-text-subtle"
																>{$t('admin.users.createdAt')}:</span
															>
															<p class="text-sm text-text">
																{new Date(user.created_at).toLocaleString()}
															</p>
														</div>
													</div>

													<!-- Actions -->
													<div
														class="flex flex-wrap gap-2 pt-2 border-t border-border"
													>
														{#if !isOAuth}
															<a
																href={resolve(`/admin/users/${user.id}/edit`)}
																class="btn btn-sm btn-primary"
															>
																{$t('common.edit')}
															</a>
														{:else}
															<button
																disabled
																class="btn btn-sm btn-ghost opacity-50 cursor-not-allowed"
																title={$t('admin.users.oauthUserCannotEdit')}
															>
																{$t('common.edit')}
															</button>
														{/if}

														<button
															onclick={() => toggleRole(user.id, user.role)}
															disabled={isOffline ||
																user.id === currentUser?.id ||
																isOAuth}
															title={isOAuth
																? $t('admin.users.oauthRoleManaged')
																: ''}
															class="btn btn-sm btn-ghost"
														>
															{$t('admin.users.changeRole')}
														</button>

														<button
															onclick={() =>
																handleImpersonate(user.id, user.email)}
															disabled={isOffline ||
																user.id === currentUser?.id}
															class="btn btn-sm btn-ghost"
														>
															{$t('admin.impersonate_user')}
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
								for="roleFilter"
								class="block text-sm font-medium text-text-ink2 mb-2"
							>
								{$t('admin.users.role')}
							</label>
							<select
								id="roleFilter"
								bind:value={roleFilter}
								class="input text-sm"
							>
								<option value="all">{$t('admin.users.allRoles')}</option>
								<option value="user">{$t('admin.users.roleUser')}</option>
								<option value="admin">{$t('admin.users.roleAdmin')}</option>
							</select>
						</div>

						<div>
							<label
								for="providerFilter"
								class="block text-sm font-medium text-text-ink2 mb-2"
							>
								{$t('admin.users.provider')}
							</label>
							<select
								id="providerFilter"
								bind:value={providerFilter}
								class="input text-sm"
							>
								<option value="all">{$t('admin.users.allProviders')}</option>
								<option value="local">{$t('admin.users.providerLocal')}</option>
								<option value="oauth">{$t('admin.users.providerOAuth')}</option>
							</select>
						</div>

						<div>
							<label
								for="sortBy"
								class="block text-sm font-medium text-text-ink2 mb-2"
							>
								{$t('common.sort')}
							</label>
							<select id="sortBy" bind:value={sortBy} class="input text-sm">
								<option value="email-asc"
									>{$t('admin.users.sortEmailAsc')}</option
								>
								<option value="email-desc"
									>{$t('admin.users.sortEmailDesc')}</option
								>
								<option value="newest">{$t('admin.users.sortNewest')}</option>
								<option value="oldest">{$t('admin.users.sortOldest')}</option>
							</select>
						</div>

						{#if hasActiveFilters}
							<button
								type="button"
								onclick={() => {
									search = '';
									sortBy = 'email-asc';
									roleFilter = 'all';
									providerFilter = 'all';
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
						for="roleFilterMobile"
						class="block text-sm font-medium text-text-ink2 mb-2"
					>
						{$t('admin.users.role')}
					</label>
					<select
						id="roleFilterMobile"
						bind:value={roleFilter}
						class="input bg-white"
					>
						<option value="all">{$t('admin.users.allRoles')}</option>
						<option value="user">{$t('admin.users.roleUser')}</option>
						<option value="admin">{$t('admin.users.roleAdmin')}</option>
					</select>
				</div>

				<div>
					<label
						for="providerFilterMobile"
						class="block text-sm font-medium text-text-ink2 mb-2"
					>
						{$t('admin.users.provider')}
					</label>
					<select
						id="providerFilterMobile"
						bind:value={providerFilter}
						class="input bg-white"
					>
						<option value="all">{$t('admin.users.allProviders')}</option>
						<option value="local">{$t('admin.users.providerLocal')}</option>
						<option value="oauth">{$t('admin.users.providerOAuth')}</option>
					</select>
				</div>

				<div>
					<label
						for="sortByMobile"
						class="block text-sm font-medium text-text-ink2 mb-2"
					>
						{$t('common.sort')}
					</label>
					<select id="sortByMobile" bind:value={sortBy} class="input bg-white">
						<option value="email-asc">{$t('admin.users.sortEmailAsc')}</option>
						<option value="email-desc">{$t('admin.users.sortEmailDesc')}</option
						>
						<option value="newest">{$t('admin.users.sortNewest')}</option>
						<option value="oldest">{$t('admin.users.sortOldest')}</option>
					</select>
				</div>

				{#if hasActiveFilters}
					<button
						type="button"
						onclick={() => {
							search = '';
							sortBy = 'email-asc';
							roleFilter = 'all';
							providerFilter = 'all';
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
</div>
