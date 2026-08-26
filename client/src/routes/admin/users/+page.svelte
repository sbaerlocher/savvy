<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import { authStore } from '$lib/stores/auth';
	import { isOnline } from '$lib/stores/offline';
	import { t } from '$lib/stores/i18n';
	import { adminApi } from '$lib/api';
	import { toastStore } from '$lib/stores/toast';
	import { logger } from '$lib/utils/logger';
	import { platform } from '$lib/utils/platform';
	import { ICON_FILTER_LINES, ICON_SEARCH } from '$lib/icons';
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

	// Desktop renders the mockup's single elevated panel with a six-column table;
	// the native platforms keep the compact list. `platform` is a module
	// constant, so a plain const, not $derived.
	const DESKTOP = platform === 'other';

	const isOffline = $derived(!$isOnline);
	const adminCount = $derived(users.filter((u) => u.role === 'admin').length);

	function initials(user: AdminUserDTO): string {
		const parts = [user.first_name, user.last_name].filter(Boolean);
		if (parts.length > 0) {
			return parts.map((p) => p!.charAt(0).toUpperCase()).join('');
		}
		return user.email.charAt(0).toUpperCase();
	}
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
	{#if DESKTOP}
		<!-- Desktop mockup (screen-AdminDesktop, "Benutzer-Verwaltung · Tabelle"):
		     title, toolbar and table live inside one elevated panel. -->
		<div
			class="grid grid-cols-1 {showFilterMenu ? 'lg:grid-cols-4' : ''} gap-6"
		>
			<div class={showFilterMenu ? 'lg:col-span-3' : ''}>
				<div
					class="overflow-hidden rounded-4xl border border-border bg-surface shadow-panel"
				>
					<div class="px-7.5 pt-6 pb-4.5">
						<div class="mb-4.5">
							<h1 class="text-heading text-text">{$t('admin.users.title')}</h1>
							<p class="mt-0.5 text-label font-normal text-text-subtle">
								{$t('admin.users.countSummary')
									.replace('{total}', users.length.toString())
									.replace('{admins}', adminCount.toString())}
							</p>
						</div>

						<div class="flex gap-2.5">
							<label
								class="flex h-10 flex-1 items-center gap-2.25 rounded-lg border border-border-field bg-surface-1 px-3.5 focus-within:border-accent focus-within:ring-2 focus-within:ring-accent/20"
							>
								<svg
									class="h-4.25 w-4.25 shrink-0 text-text-subtle"
									fill="none"
									stroke="currentColor"
									stroke-width="2"
									viewBox="0 0 24 24"
									aria-hidden="true"
								>
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										d={ICON_SEARCH}
									/>
								</svg>
								<input
									type="search"
									bind:value={search}
									placeholder={$t('admin.users.searchPlaceholder')}
									aria-label={$t('common.search')}
									class="min-w-0 flex-1 bg-transparent text-body text-text placeholder:text-text-placeholder focus:outline-none"
								/>
							</label>

							<button
								type="button"
								onclick={(e: MouseEvent) => {
									e.stopPropagation();
									showFilterMenu = !showFilterMenu;
								}}
								class="relative inline-flex h-10 items-center gap-2 rounded-lg border border-border-field bg-surface px-4 text-label text-text-ink2 transition-colors hover:bg-surface-1"
								aria-label={$t('common.filter')}
								aria-expanded={showFilterMenu}
							>
								<svg
									class="h-4.25 w-4.25"
									fill="none"
									stroke="currentColor"
									stroke-width="2"
									viewBox="0 0 24 24"
									aria-hidden="true"
								>
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										d={ICON_FILTER_LINES}
									/>
								</svg>
								{$t('common.filter')}
								{#if hasActiveFilters}
									<span
										class="absolute -top-0.75 -right-0.75 h-2.5 w-2.5 rounded-full bg-accent"
									></span>
								{/if}
							</button>

							{#if localLoginEnabled}
								<a
									href={resolve('/admin/users/new')}
									class="inline-flex h-10 items-center rounded-lg bg-accent px-4.5 text-label whitespace-nowrap text-white shadow-sm transition-colors hover:bg-accent-hover {isOffline
										? 'pointer-events-none cursor-not-allowed opacity-50 blur-[0.5px]'
										: ''}"
								>
									{$t('admin.users.createUser')}
								</a>
							{/if}
						</div>
					</div>

					{#if isLoading}
						<div class="border-t border-border-soft"><LoadingSpinner /></div>
					{:else if filteredUsers.length === 0}
						<div
							class="border-t border-border-soft py-12 text-center text-text-subtle"
						>
							{$t('admin.users.noUsers')}
						</div>
					{:else}
						<div class="border-t border-border-soft">
							<div
								class="grid grid-cols-[2.4fr_1.6fr_1fr_1fr_1.4fr_1.2fr] border-b border-border-soft bg-surface-1 px-7.5 py-3"
							>
								<span class="text-section-eyebrow uppercase text-text-subtle"
									>{$t('admin.users.email')}</span
								>
								<span class="text-section-eyebrow uppercase text-text-subtle"
									>{$t('admin.users.name')}</span
								>
								<span class="text-section-eyebrow uppercase text-text-subtle"
									>{$t('admin.users.role')}</span
								>
								<span class="text-section-eyebrow uppercase text-text-subtle"
									>{$t('admin.users.provider')}</span
								>
								<span class="text-section-eyebrow uppercase text-text-subtle"
									>{$t('admin.users.createdAt')}</span
								>
								<span
									class="text-section-eyebrow text-right uppercase text-text-subtle"
									>{$t('admin.users.actions')}</span
								>
							</div>

							{#each filteredUsers as user (user.id)}
								{@const isOAuth = user.auth_provider === 'oauth'}
								{@const expanded = expandedUserId === user.id}
								<div class="border-b border-border-soft">
									<button
										type="button"
										onclick={() => toggleExpandUser(user.id)}
										class="grid w-full grid-cols-[2.4fr_1.6fr_1fr_1fr_1.4fr_1.2fr] items-center px-7.5 py-3.5 text-left transition-colors hover:bg-surface-1"
										aria-expanded={expanded}
									>
										<span class="flex min-w-0 items-center gap-2.75">
											<span
												aria-hidden="true"
												class="flex h-8 w-8 flex-none items-center justify-center rounded-full bg-accent-50 text-body-sm font-semibold text-accent-800"
												>{initials(user)}</span
											>
											<span class="truncate text-body font-medium text-text"
												>{user.email}</span
											>
										</span>
										<span class="text-body text-text-ink2"
											>{user.first_name || ''}
											{user.last_name || ''}</span
										>
										<span>
											<span
												class="inline-flex flex-none items-center rounded-full px-2.5 py-0.5 text-eyebrow whitespace-nowrap {user.role ===
												'admin'
													? 'bg-danger-100 text-danger-800'
													: 'bg-border-soft text-text-strong'}"
											>
												{user.role === 'admin'
													? $t('admin.users.roleAdmin')
													: $t('admin.users.roleUser')}
											</span>
										</span>
										<span class="text-body text-text-ink2"
											>{isOAuth
												? $t('admin.users.providerOAuth')
												: $t('admin.users.providerLocal')}</span
										>
										<span class="text-body text-text-ink2"
											>{new Date(user.created_at).toLocaleDateString()}</span
										>
										<span class="flex items-center justify-end gap-2">
											<span class="text-label text-accent"
												>{expanded
													? $t('common.close')
													: $t('admin.users.showDetails')}</span
											>
											<svg
												class="h-3.75 w-3.75 flex-none text-text-faint transition-transform {expanded
													? 'rotate-180'
													: ''}"
												fill="none"
												stroke="currentColor"
												stroke-width="2"
												viewBox="0 0 24 24"
												aria-hidden="true"
											>
												<path
													stroke-linecap="round"
													stroke-linejoin="round"
													d="M19 9l-7 7-7-7"
												/>
											</svg>
										</span>
									</button>

									{#if expanded}
										<div
											class="flex items-start gap-10 bg-surface-1 px-7.5 pt-4.5 pb-5"
										>
											<div class="grid grid-cols-3 gap-x-10 gap-y-2">
												<div>
													<div
														class="mb-0.75 text-section-eyebrow uppercase text-text-faint"
													>
														{$t('admin.users.userId')}
													</div>
													<div class="font-mono text-body-sm text-text">
														{user.id}
													</div>
												</div>
												<div>
													<div
														class="mb-0.75 text-section-eyebrow uppercase text-text-faint"
													>
														{$t('admin.users.provider')}
													</div>
													<div class="text-label font-normal text-text">
														{isOAuth
															? $t('admin.users.providerOAuth')
															: $t('admin.users.providerLocal')}
													</div>
												</div>
												<div>
													<div
														class="mb-0.75 text-section-eyebrow uppercase text-text-faint"
													>
														{$t('admin.users.createdAt')}
													</div>
													<div class="text-label font-normal text-text">
														{new Date(user.created_at).toLocaleString()}
													</div>
												</div>
											</div>
											<div class="flex-1"></div>
											<div class="flex flex-none gap-2">
												{#if !isOAuth}
													<a
														href={resolve(`/admin/users/${user.id}/edit`)}
														class="inline-flex h-9 items-center rounded-full bg-accent px-4 text-body-sm font-semibold text-white transition-colors hover:bg-accent-hover"
													>
														{$t('common.edit')}
													</a>
												{:else}
													<button
														type="button"
														disabled
														class="inline-flex h-9 cursor-not-allowed items-center rounded-full bg-accent px-4 text-body-sm font-semibold text-white opacity-45"
														title={$t('admin.users.oauthUserCannotEdit')}
													>
														{$t('common.edit')}
													</button>
												{/if}

												<button
													type="button"
													onclick={() => toggleRole(user.id, user.role)}
													disabled={isOffline ||
														user.id === currentUser?.id ||
														isOAuth}
													title={isOAuth
														? $t('admin.users.oauthRoleManaged')
														: ''}
													class="inline-flex h-9 items-center rounded-full border border-border-field bg-surface px-4 text-body-sm font-semibold text-text transition-colors hover:bg-surface-1 disabled:cursor-not-allowed disabled:opacity-50"
												>
													{$t('admin.users.changeRole')}
												</button>

												<button
													type="button"
													onclick={() => handleImpersonate(user.id, user.email)}
													disabled={isOffline || user.id === currentUser?.id}
													class="inline-flex h-9 items-center rounded-full border border-border-field bg-surface px-4 text-body-sm font-semibold text-text transition-colors hover:bg-surface-1 disabled:cursor-not-allowed disabled:opacity-50"
												>
													{$t('admin.impersonate_user')}
												</button>
											</div>
										</div>
									{/if}
								</div>
							{/each}
						</div>
					{/if}
				</div>
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
									<option value="local"
										>{$t('admin.users.providerLocal')}</option
									>
									<option value="oauth"
										>{$t('admin.users.providerOAuth')}</option
									>
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
	{:else}
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
					class="flex items-center justify-center gap-2 control px-4 bg-white border border-border-field rounded-md hover:bg-surface-1 transition-colors relative"
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
						<span
							class="absolute -top-1 -right-1 w-3 h-3 bg-accent rounded-full"
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
					class="flex-1 flex items-center justify-center control px-3 bg-white border border-border-field rounded-md hover:bg-surface-1 transition-colors relative"
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
						<span
							class="absolute -top-1 -right-1 w-3 h-3 bg-accent rounded-full"
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
		<div
			class="grid grid-cols-1 {showFilterMenu ? 'lg:grid-cols-4' : ''} gap-6"
		>
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
														? 'bg-danger-100 text-danger-800'
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
																<span
																	class="text-xs font-medium text-text-subtle"
																	>{$t('admin.users.name')}:</span
																>
																<p class="text-sm text-text">
																	{user.first_name || ''}
																	{user.last_name || ''}
																</p>
															</div>
															<div class="md:hidden">
																<span
																	class="text-xs font-medium text-text-subtle"
																	>{$t('admin.users.role')}:</span
																>
																<p class="text-sm">
																	<span
																		class="px-2 py-1 text-xs font-medium rounded-full {user.role ===
																		'admin'
																			? 'bg-danger-100 text-danger-800'
																			: 'bg-border-soft text-text-strong'}"
																	>
																		{user.role === 'admin'
																			? $t('admin.users.roleAdmin')
																			: $t('admin.users.roleUser')}
																	</span>
																</p>
															</div>
															<div>
																<span
																	class="text-xs font-medium text-text-subtle"
																	>{$t('admin.users.userId')}:</span
																>
																<p class="text-sm text-text font-mono">
																	{user.id}
																</p>
															</div>
															<div>
																<span
																	class="text-xs font-medium text-text-subtle"
																	>{$t('admin.users.provider')}:</span
																>
																<p class="text-sm text-text">
																	{user.auth_provider}
																</p>
															</div>
															<div>
																<span
																	class="text-xs font-medium text-text-subtle"
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
		</div>
	{/if}

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
