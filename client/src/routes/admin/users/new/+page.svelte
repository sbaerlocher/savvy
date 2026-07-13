<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { get } from 'svelte/store';
	import { t } from '$lib/stores/i18n';
	import { adminApi } from '$lib/api';
	import { toastStore } from '$lib/stores/toast';

	// Svelte 5 compatible translation wrapper
	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);

	let email = $state('');
	let password = $state('');
	let firstName = $state('');
	let lastName = $state('');
	let role = $state<'user' | 'admin'>('user');
	let isLoading = $state(false);

	async function handleSubmit(e: Event) {
		e.preventDefault();
		isLoading = true;
		try {
			await adminApi.createUser({
				email: email,
				password: password,
				first_name: firstName,
				last_name: lastName,
				role: role
			});
			toastStore.success(tr('admin.users.createSuccess'));
			goto(resolve('/admin/users'));
		} catch (err) {
			const message =
				err instanceof Error ? err.message : tr('admin.users.createError');
			toastStore.error(message);
		} finally {
			isLoading = false;
		}
	}
</script>

<svelte:head>
	<title>{tr('admin.users.createUser')} - {tr('common.appName')}</title>
</svelte:head>

<div class="mb-6">
	<a href={resolve('/admin/users')} class="text-accent hover:text-accent-hover"
		>{tr('common.backToOverview')}</a
	>
</div>

<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
	<!-- Left column: User Form (2/3 width) -->
	<div class="lg:col-span-2">
		<div class="bg-white rounded-lg shadow-lg p-6">
			<h1 class="text-3xl font-bold text-text mb-6">
				{tr('admin.users.createUser')}
			</h1>
			<form onsubmit={handleSubmit} class="space-y-6">
				<!-- Email -->
				<div>
					<label
						for="email"
						class="block text-sm font-medium text-text-ink2 mb-1"
						>{tr('admin.users.email')} *</label
					>
					<input
						type="email"
						id="email"
						bind:value={email}
						required
						class="w-full px-4 py-2 bg-white border border-border-field rounded-md focus:ring-accent focus:border-accent"
						placeholder={tr('admin.users.emailPlaceholder')}
					/>
				</div>

				<!-- Password -->
				<div>
					<label
						for="password"
						class="block text-sm font-medium text-text-ink2 mb-1"
						>{tr('admin.users.password')} *</label
					>
					<input
						type="password"
						id="password"
						bind:value={password}
						required
						class="w-full px-4 py-2 bg-white border border-border-field rounded-md focus:ring-accent focus:border-accent"
						placeholder={tr('admin.users.passwordPlaceholder')}
					/>
				</div>

				<!-- First Name -->
				<div>
					<label
						for="firstName"
						class="block text-sm font-medium text-text-ink2 mb-1"
						>{tr('admin.users.firstName')} *</label
					>
					<input
						type="text"
						id="firstName"
						bind:value={firstName}
						required
						class="w-full px-4 py-2 bg-white border border-border-field rounded-md focus:ring-accent focus:border-accent"
					/>
				</div>

				<!-- Last Name -->
				<div>
					<label
						for="lastName"
						class="block text-sm font-medium text-text-ink2 mb-1"
						>{tr('admin.users.lastName')} *</label
					>
					<input
						type="text"
						id="lastName"
						bind:value={lastName}
						required
						class="w-full px-4 py-2 bg-white border border-border-field rounded-md focus:ring-accent focus:border-accent"
					/>
				</div>

				<!-- Role -->
				<div>
					<label
						for="role"
						class="block text-sm font-medium text-text-ink2 mb-1"
						>{tr('admin.users.role')} *</label
					>
					<select
						id="role"
						bind:value={role}
						class="w-full px-4 py-2 bg-white border border-border-field rounded-md focus:ring-accent focus:border-accent"
					>
						<option value="user">{tr('admin.users.roleUser')}</option>
						<option value="admin">{tr('admin.users.roleAdmin')}</option>
					</select>
				</div>

				<!-- Buttons -->
				<div class="flex gap-3 pt-4">
					<button
						type="submit"
						disabled={isLoading}
						class="btn btn-sm btn-primary flex-1"
					>
						{isLoading ? tr('common.loading') : tr('common.save')}
					</button>
					<a href={resolve('/admin/users')} class="btn btn-sm btn-ghost">
						{tr('common.cancel')}
					</a>
				</div>
			</form>
		</div>
	</div>

	<!-- Right column: Role Information (1/3 width) -->
	<div class="lg:col-span-1">
		<div class="bg-white rounded-lg shadow-lg p-6">
			<h2 class="text-xl font-bold text-text mb-4">
				{tr('admin.users.roleInfo.title')}
			</h2>
			<p class="text-sm text-text-muted mb-4">
				{tr('admin.users.roleInfo.description')}
			</p>

			<div class="space-y-4">
				<!-- User Role -->
				<div class="border border-border rounded-lg p-4">
					<div class="flex items-center mb-2">
						<svg
							class="w-5 h-5 text-accent mr-2"
							fill="currentColor"
							viewBox="0 0 20 20"
						>
							<path
								fill-rule="evenodd"
								d="M10 9a3 3 0 100-6 3 3 0 000 6zm-7 9a7 7 0 1114 0H3z"
								clip-rule="evenodd"
							/>
						</svg>
						<h3 class="text-sm font-bold text-text">
							{tr('admin.users.roleUser')}
						</h3>
					</div>
					<p class="text-xs text-text-muted mb-2">
						{tr('admin.users.roleInfo.userDesc')}
					</p>
					<ul class="text-xs text-text-muted space-y-1 ml-4">
						<li class="flex items-start">
							<span class="mr-2">•</span>
							<span>{tr('admin.users.roleInfo.userPerm1')}</span>
						</li>
						<li class="flex items-start">
							<span class="mr-2">•</span>
							<span>{tr('admin.users.roleInfo.userPerm2')}</span>
						</li>
						<li class="flex items-start">
							<span class="mr-2">•</span>
							<span>{tr('admin.users.roleInfo.userPerm3')}</span>
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
							{tr('admin.users.roleAdmin')}
						</h3>
					</div>
					<p class="text-xs text-red-800 mb-2">
						{tr('admin.users.roleInfo.adminDesc')}
					</p>
					<ul class="text-xs text-red-800 space-y-1 ml-4">
						<li class="flex items-start">
							<span class="mr-2">•</span>
							<span>{tr('admin.users.roleInfo.adminPerm1')}</span>
						</li>
						<li class="flex items-start">
							<span class="mr-2">•</span>
							<span>{tr('admin.users.roleInfo.adminPerm2')}</span>
						</li>
						<li class="flex items-start">
							<span class="mr-2">•</span>
							<span>{tr('admin.users.roleInfo.adminPerm3')}</span>
						</li>
						<li class="flex items-start">
							<span class="mr-2">•</span>
							<span>{tr('admin.users.roleInfo.adminPerm4')}</span>
						</li>
					</ul>
				</div>
			</div>

			<div class="mt-6 pt-6 border-t border-border">
				<div class="flex items-start text-sm text-text-muted">
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
					<span class="text-xs">{tr('admin.users.roleInfo.warning')}</span>
				</div>
			</div>
		</div>
	</div>
</div>
