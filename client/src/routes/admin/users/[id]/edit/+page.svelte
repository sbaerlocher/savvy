<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { adminApi } from '$lib/api';
	import { authStore } from '$lib/stores/auth';
	import { toastStore } from '$lib/stores/toast';
	import { isOnline } from '$lib/stores/offline';
	import { t } from '$lib/stores/i18n';
	import ConfirmModal from '$lib/components/ConfirmModal.svelte';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import type { AdminUserDTO } from '$lib/types/api';

	const userId = $derived($page.params.id);

	let user = $state<AdminUserDTO | null>(null);
	let email = $state('');
	let firstName = $state('');
	let lastName = $state('');
	let role = $state<'user' | 'admin'>('user');
	let isLoading = $state(false);
	let isLoadingUser = $state(true);

	// Modal state
	let showImpersonateModal = $state(false);

	const isOffline = $derived(!$isOnline);
	const currentUser = $derived($authStore.user);
	const isAdmin = $derived(currentUser?.is_admin || false);
	const isEditingSelf = $derived(user?.id === currentUser?.id);
	const isOAuthUser = $derived(user?.auth_provider === 'oauth');

	// ✅ Server-side admin check in +layout.server.ts (SVL-002 Fix)
	// No client-side check needed - user is already validated server-side
	onMount(async () => {
		await loadUser();
	});

	async function loadUser() {
		isLoadingUser = true;
		try {
			if (!userId) {
				toastStore.error($t('admin.users.loadError'));
				goto(resolve('/admin/users'));
				return;
			}
			const response = await adminApi.getUser(userId);
			user = response.user;
			email = user.email;
			firstName = user.first_name;
			lastName = user.last_name;
			role = user.role;
		} catch {
			toastStore.error($t('admin.users.loadError'));
			goto(resolve('/admin/users'));
		} finally {
			isLoadingUser = false;
		}
	}

	async function handleSubmit() {
		if (!email || !firstName || !lastName) {
			toastStore.error($t('admin.users.missingFields'));
			return;
		}

		if (!userId) {
			toastStore.error($t('admin.users.updateError'));
			return;
		}

		isLoading = true;
		try {
			await adminApi.updateUser(userId, {
				email,
				first_name: firstName,
				last_name: lastName,
				...(isOAuthUser ? {} : { role })
			});
			toastStore.success($t('admin.users.updateSuccess'));
			goto(resolve('/admin/users'));
		} catch (err) {
			const message =
				err instanceof Error ? err.message : $t('admin.users.updateError');
			toastStore.error(message);
		} finally {
			isLoading = false;
		}
	}

	function handleCancel() {
		goto(resolve('/admin/users'));
	}

	function promptImpersonate() {
		showImpersonateModal = true;
	}

	async function confirmImpersonate() {
		if (!user) return;

		try {
			await authStore.startImpersonation(user.id);
			// Redirect happens in authStore
			showImpersonateModal = false;
		} catch (err) {
			const message =
				err instanceof Error
					? err.message
					: $t('admin.impersonate_info.failed');
			toastStore.error(message);
			showImpersonateModal = false;
		}
	}
</script>

<svelte:head>
	<title>{$t('admin.users.editUser')} - {$t('common.appName')}</title>
</svelte:head>

<div class="px-4 pb-20 md:pb-4">
	<!-- Header -->
	<div class="mb-6">
		<button onclick={handleCancel} class="text-cyan-600 hover:text-cyan-700">
			{$t('common.backToOverview')}
		</button>
	</div>

	{#if isLoadingUser}
		<LoadingSpinner />
	{:else if user}
		<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
			<!-- Left column: Edit Form (2/3 width) -->
			<div class="lg:col-span-2">
				<div class="bg-white shadow-lg rounded-lg p-6">
					<h1 class="text-3xl font-bold text-gray-900 mb-6">
						{$t('admin.users.editUser')}
					</h1>

					<form
						onsubmit={(e) => {
							e.preventDefault();
							handleSubmit();
						}}
						class="space-y-6"
					>
						<!-- Email -->
						<div>
							<label
								for="email"
								class="block text-sm font-medium text-gray-700 mb-1"
							>
								{$t('admin.users.email')} *
							</label>
							<input
								id="email"
								type="email"
								bind:value={email}
								required
								disabled={!isAdmin}
								class="w-full px-4 py-2 bg-white border border-gray-300 rounded-md focus:ring-cyan-500 focus:border-cyan-500 disabled:bg-gray-100 disabled:cursor-not-allowed"
							/>
						</div>

						<!-- First Name -->
						<div>
							<label
								for="firstName"
								class="block text-sm font-medium text-gray-700 mb-1"
							>
								{$t('admin.users.firstName')} *
							</label>
							<input
								id="firstName"
								type="text"
								bind:value={firstName}
								required
								disabled={!isAdmin}
								class="w-full px-4 py-2 bg-white border border-gray-300 rounded-md focus:ring-cyan-500 focus:border-cyan-500 disabled:bg-gray-100 disabled:cursor-not-allowed"
							/>
						</div>

						<!-- Last Name -->
						<div>
							<label
								for="lastName"
								class="block text-sm font-medium text-gray-700 mb-1"
							>
								{$t('admin.users.lastName')} *
							</label>
							<input
								id="lastName"
								type="text"
								bind:value={lastName}
								required
								disabled={!isAdmin}
								class="w-full px-4 py-2 bg-white border border-gray-300 rounded-md focus:ring-cyan-500 focus:border-cyan-500 disabled:bg-gray-100 disabled:cursor-not-allowed"
							/>
						</div>

						<!-- Role -->
						<div>
							<label
								for="role"
								class="block text-sm font-medium text-gray-700 mb-1"
							>
								{$t('admin.users.role')} *
							</label>
							<select
								id="role"
								bind:value={role}
								disabled={!isAdmin || isEditingSelf || isOAuthUser}
								class="w-full px-4 py-2 bg-white border border-gray-300 rounded-md focus:ring-cyan-500 focus:border-cyan-500 disabled:bg-gray-100 disabled:cursor-not-allowed"
							>
								<option value="user">{$t('admin.users.roleUser')}</option>
								<option value="admin">{$t('admin.users.roleAdmin')}</option>
							</select>
							{#if isOAuthUser}
								<p class="text-sm text-gray-500 mt-1">
									{$t('admin.users.oauthRoleManaged')}
								</p>
							{:else if isEditingSelf}
								<p class="text-sm text-gray-500 mt-1">
									{$t('admin.users.cannotChangeOwnRole')}
								</p>
							{/if}
						</div>

						<!-- Provider Info (Read-Only) -->
						<div>
							<label
								for="user-provider"
								class="block text-sm font-medium text-gray-700 mb-1"
							>
								{$t('admin.users.provider')}
							</label>
							<input
								id="user-provider"
								type="text"
								value={user.auth_provider}
								disabled
								class="w-full px-4 py-2 bg-gray-100 border border-gray-300 rounded-md cursor-not-allowed"
							/>
						</div>

						<!-- Action Buttons -->
						<div class="flex gap-3 pt-4">
							<button
								type="submit"
								disabled={isLoading || isOffline || !isAdmin}
								class="btn btn-sm btn-primary flex-1"
							>
								{isLoading ? $t('common.loading') : $t('common.save')}
							</button>

							<button
								type="button"
								onclick={handleCancel}
								disabled={isLoading}
								class="btn btn-sm btn-ghost flex-1"
							>
								{$t('common.cancel')}
							</button>
						</div>

						{#if isOffline}
							<p class="text-sm text-amber-600 text-center">
								{$t('common.offlineEditDisabled')}
							</p>
						{/if}
					</form>

					<!-- Impersonation Section -->
					{#if !isEditingSelf}
						<div class="pt-6 mt-6 border-t border-gray-200">
							<h3 class="text-lg font-bold text-gray-900 mb-2">
								{$t('admin.impersonate_info.title')}
							</h3>
							<p class="text-sm text-gray-600 mb-4">
								{$t('admin.impersonate_info.description')}
							</p>
							<button
								type="button"
								onclick={promptImpersonate}
								disabled={isOffline || !isAdmin}
								class="btn btn-sm btn-primary w-full"
							>
								{$t('admin.impersonate_user')}
							</button>
							{#if isOffline}
								<p class="text-sm text-amber-600 text-center mt-2">
									{$t('common.offlineEditDisabled')}
								</p>
							{/if}
						</div>
					{/if}
				</div>
			</div>

			<!-- Right column: Role Information (1/3 width) -->
			<div class="lg:col-span-1">
				<div class="bg-white rounded-lg shadow-lg p-6">
					<h2 class="text-xl font-bold text-gray-900 mb-4">
						{$t('admin.users.roleInfo.title')}
					</h2>
					<p class="text-sm text-gray-600 mb-4">
						{$t('admin.users.roleInfo.description')}
					</p>

					<div class="space-y-4">
						<!-- User Role -->
						<div class="border border-gray-200 rounded-lg p-4">
							<div class="flex items-center mb-2">
								<svg
									class="w-5 h-5 text-cyan-500 mr-2"
									fill="currentColor"
									viewBox="0 0 20 20"
								>
									<path
										fill-rule="evenodd"
										d="M10 9a3 3 0 100-6 3 3 0 000 6zm-7 9a7 7 0 1114 0H3z"
										clip-rule="evenodd"
									/>
								</svg>
								<h3 class="text-sm font-bold text-gray-900">
									{$t('admin.users.roleUser')}
								</h3>
							</div>
							<p class="text-xs text-gray-600 mb-2">
								{$t('admin.users.roleInfo.userDesc')}
							</p>
							<ul class="text-xs text-gray-600 space-y-1 ml-4">
								<li class="flex items-start">
									<span class="mr-2">•</span>
									<span>{$t('admin.users.roleInfo.userPerm1')}</span>
								</li>
								<li class="flex items-start">
									<span class="mr-2">•</span>
									<span>{$t('admin.users.roleInfo.userPerm2')}</span>
								</li>
								<li class="flex items-start">
									<span class="mr-2">•</span>
									<span>{$t('admin.users.roleInfo.userPerm3')}</span>
								</li>
							</ul>
						</div>

						<!-- Admin Role -->
						<div class="border border-red-200 bg-red-50 rounded-lg p-4">
							<div class="flex items-center mb-2">
								<svg
									class="w-5 h-5 text-red-500 mr-2"
									fill="currentColor"
									viewBox="0 0 20 20"
								>
									<path
										fill-rule="evenodd"
										d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-6-3a2 2 0 11-4 0 2 2 0 014 0zm-2 4a5 5 0 00-4.546 2.916A5.986 5.986 0 0010 16a5.986 5.986 0 004.546-2.084A5 5 0 0010 11z"
										clip-rule="evenodd"
									/>
								</svg>
								<h3 class="text-sm font-bold text-red-900">
									{$t('admin.users.roleAdmin')}
								</h3>
							</div>
							<p class="text-xs text-red-800 mb-2">
								{$t('admin.users.roleInfo.adminDesc')}
							</p>
							<ul class="text-xs text-red-800 space-y-1 ml-4">
								<li class="flex items-start">
									<span class="mr-2">•</span>
									<span>{$t('admin.users.roleInfo.adminPerm1')}</span>
								</li>
								<li class="flex items-start">
									<span class="mr-2">•</span>
									<span>{$t('admin.users.roleInfo.adminPerm2')}</span>
								</li>
								<li class="flex items-start">
									<span class="mr-2">•</span>
									<span>{$t('admin.users.roleInfo.adminPerm3')}</span>
								</li>
								<li class="flex items-start">
									<span class="mr-2">•</span>
									<span>{$t('admin.users.roleInfo.adminPerm4')}</span>
								</li>
							</ul>
						</div>
					</div>

					<div class="mt-6 pt-6 border-t border-gray-200">
						<div class="flex items-start text-sm text-gray-600">
							<svg
								class="w-5 h-5 text-yellow-500 mr-2 flex-shrink-0 mt-0.5"
								fill="currentColor"
								viewBox="0 0 20 20"
							>
								<path
									fill-rule="evenodd"
									d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z"
									clip-rule="evenodd"
								/>
							</svg>
							<span class="text-xs">{$t('admin.users.roleInfo.warning')}</span>
						</div>
					</div>
				</div>
			</div>
		</div>
	{/if}

	<!-- Confirmation Modal -->
	<ConfirmModal
		isOpen={showImpersonateModal}
		title={$t('admin.impersonate_info.confirm_title')}
		message={user
			? $t('admin.impersonate_info.actions_as_user').replace(
					'{email}',
					user.email
				)
			: ''}
		confirmText={$t('admin.impersonate_info.confirm_button')}
		cancelText={$t('common.cancel')}
		variant="warning"
		onconfirm={confirmImpersonate}
		oncancel={() => (showImpersonateModal = false)}
	/>
</div>
