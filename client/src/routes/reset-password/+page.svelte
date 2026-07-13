<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/stores';
	import { ApiError, authApi } from '$lib/api';
	import { authStore } from '$lib/stores/auth';
	import { t } from '$lib/stores/i18n';
	import { toastStore } from '$lib/stores/toast';
	import { onMount } from 'svelte';
	import { get } from 'svelte/store';

	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);

	let password = $state('');
	let confirmPassword = $state('');
	let isLoading = $state(false);
	let status = $state<'form' | 'success' | 'error'>('form');
	let errorMessage = $state('');
	let errorCode = $state('');
	let token = $state('');

	onMount(async () => {
		if ($authStore.isAuthenticated) {
			goto(resolve('/dashboard'));
			return;
		}

		// Initialize CSRF cookie by hitting a public GET endpoint
		await fetch('/api/v1/config').catch(() => {});

		token = $page.url.searchParams.get('token') || '';
		if (!token) {
			status = 'error';
			errorMessage = tr('auth.resetPassword.invalidToken');
		}
	});

	async function handleSubmit(e: Event) {
		e.preventDefault();

		if (password !== confirmPassword) {
			toastStore.error(tr('auth.resetPassword.passwordMismatch'));
			return;
		}

		isLoading = true;

		try {
			await authApi.resetPassword(token, password);
			status = 'success';
		} catch (error) {
			if (error instanceof ApiError) {
				errorCode = error.error || '';
				if (errorCode === 'token_expired') {
					errorMessage = tr('auth.resetPassword.tokenExpired');
				} else if (errorCode === 'token_used') {
					errorMessage = tr('auth.resetPassword.tokenUsed');
				} else if (errorCode === 'invalid_token') {
					errorMessage = tr('auth.resetPassword.invalidToken');
				} else {
					errorMessage = error.message || tr('auth.resetPassword.error');
				}
			} else {
				errorMessage = tr('auth.resetPassword.error');
			}
			status = 'error';
		} finally {
			isLoading = false;
		}
	}
</script>

<svelte:head>
	<title>{tr('auth.resetPassword.title')} - {tr('common.appName')}</title>
</svelte:head>

<div class="min-h-screen py-12 px-4 sm:px-6 lg:px-8">
	<div class="max-w-md mx-auto">
		<div class="bg-white rounded-lg shadow-lg p-6 sm:p-8">
			<div class="mb-8 flex items-center gap-4">
				<img src="/logo.png" alt="Savvy Logo" class="h-12 sm:h-16" />
				<h1 class="text-2xl font-bold text-text">
					{tr('auth.resetPassword.title')}
				</h1>
			</div>

			{#if status === 'success'}
				<div class="text-center py-8">
					<div
						class="mx-auto flex items-center justify-center h-16 w-16 rounded-full bg-green-100 mb-4"
					>
						<svg
							class="h-8 w-8 text-green-600"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M5 13l4 4L19 7"
							/>
						</svg>
					</div>
					<h2 class="text-xl font-semibold text-text mb-2">
						{tr('auth.resetPassword.success')}
					</h2>
					<p class="text-text-muted mb-6">
						{tr('auth.resetPassword.successMessage')}
					</p>
					<a href={resolve('/login')} class="btn btn-primary w-full">
						{tr('auth.resetPassword.goToLogin')}
					</a>
				</div>
			{:else if status === 'error'}
				<div class="text-center py-8">
					<div
						class="mx-auto flex items-center justify-center h-16 w-16 rounded-full bg-red-100 mb-4"
					>
						<svg
							class="h-8 w-8 text-red-600"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M6 18L18 6M6 6l12 12"
							/>
						</svg>
					</div>
					<h2 class="text-xl font-semibold text-text mb-2">
						{tr('auth.resetPassword.error')}
					</h2>
					<p class="text-text-muted mb-6">{errorMessage}</p>

					<div class="space-y-3">
						{#if errorCode === 'token_expired' || errorCode === 'invalid_token'}
							<a
								href={resolve('/forgot-password')}
								class="btn btn-primary w-full"
							>
								{tr('auth.resetPassword.requestNew')}
							</a>
						{/if}
						<a href={resolve('/login')} class="btn btn-ghost w-full">
							{tr('auth.resetPassword.goToLogin')}
						</a>
					</div>
				</div>
			{:else}
				<p class="text-text-muted mb-6">{tr('auth.resetPassword.description')}</p>

				<form class="space-y-6" onsubmit={handleSubmit}>
					<div>
						<label
							for="password"
							class="block text-sm font-medium text-text-ink2 mb-1"
						>
							{tr('auth.resetPassword.password')}
						</label>
						<input
							id="password"
							name="password"
							type="password"
							autocomplete="new-password"
							required
							bind:value={password}
							disabled={isLoading}
							class="w-full px-4 py-2 bg-white border border-border-field rounded-md focus:ring-accent focus:border-accent"
							placeholder={tr('auth.resetPassword.passwordPlaceholder')}
						/>
					</div>

					<div>
						<label
							for="confirmPassword"
							class="block text-sm font-medium text-text-ink2 mb-1"
						>
							{tr('auth.resetPassword.confirmPassword')}
						</label>
						<input
							id="confirmPassword"
							name="confirmPassword"
							type="password"
							autocomplete="new-password"
							required
							bind:value={confirmPassword}
							disabled={isLoading}
							class="w-full px-4 py-2 bg-white border border-border-field rounded-md focus:ring-accent focus:border-accent"
							placeholder={tr('auth.resetPassword.confirmPasswordPlaceholder')}
						/>
					</div>

					<div class="pt-2">
						<button
							type="submit"
							disabled={isLoading}
							class="btn btn-primary w-full"
						>
							{#if isLoading}
								<span class="relative inline-flex h-3 w-3 mr-2"
									><span
										class="animate-ping absolute inline-flex h-full w-full rounded-full bg-accent-400 opacity-75"
									></span><span
										class="relative inline-flex rounded-full h-3 w-3 bg-accent"
									></span></span
								>
								{tr('auth.resetPassword.submitting')}
							{:else}
								{tr('auth.resetPassword.submitButton')}
							{/if}
						</button>
					</div>

					<div class="text-center pt-4">
						<a
							href={resolve('/login')}
							class="font-medium text-accent hover:text-accent"
						>
							{tr('auth.resetPassword.goToLogin')}
						</a>
					</div>
				</form>
			{/if}
		</div>
	</div>
</div>
