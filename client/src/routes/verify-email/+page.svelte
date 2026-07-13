<script lang="ts">
	import { get } from 'svelte/store';
	import { authStore } from '$lib/stores/auth';
	import { authApi } from '$lib/api';
	import { ApiError } from '$lib/api/client';
	import { t } from '$lib/stores/i18n';
	import { toastStore } from '$lib/stores/toast';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);

	let status = $state<'loading' | 'success' | 'error' | 'idle'>('idle');
	let errorMessage = $state('');
	let errorCode = $state('');
	let isResending = $state(false);

	onMount(async () => {
		const token = $page.url.searchParams.get('token');
		if (!token) {
			status = 'error';
			errorMessage = tr('auth.verification.invalidToken');
			return;
		}

		status = 'loading';

		try {
			await authApi.verifyEmail(token);
			status = 'success';

			// Refresh user data if authenticated
			if ($authStore.isAuthenticated) {
				await authStore.checkAuth();
			}
		} catch (error) {
			status = 'error';
			errorCode = error instanceof ApiError ? error.error || '' : '';

			if (errorCode === 'token_expired') {
				errorMessage = tr('auth.verification.tokenExpired');
			} else if (errorCode === 'token_used') {
				errorMessage = tr('auth.verification.tokenUsed');
			} else {
				errorMessage = tr('auth.verification.invalidToken');
			}
		}
	});

	async function handleResend() {
		if (!$authStore.isAuthenticated) {
			goto(resolve('/login'));
			return;
		}

		isResending = true;
		try {
			await authApi.requestVerification();
			toastStore.success(tr('auth.verification.resendSuccess'));
		} catch {
			toastStore.error(tr('auth.verification.resendError'));
		} finally {
			isResending = false;
		}
	}
</script>

<svelte:head>
	<title>{tr('auth.verification.title')} - {tr('common.appName')}</title>
</svelte:head>

<div class="min-h-screen py-12 px-4 sm:px-6 lg:px-8">
	<div class="max-w-md mx-auto">
		<div class="bg-white rounded-lg shadow-lg p-6 sm:p-8">
			<div class="mb-8 flex items-center gap-4">
				<img src="/logo.png" alt="Savvy Logo" class="h-12 sm:h-16" />
				<h1 class="text-2xl font-bold text-text">
					{tr('auth.verification.title')}
				</h1>
			</div>

			{#if status === 'loading'}
				<LoadingSpinner />
			{:else if status === 'success'}
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
						{tr('auth.verification.success')}
					</h2>
					<p class="text-text-muted mb-6">
						{tr('auth.verification.successMessage')}
					</p>

					{#if $authStore.isAuthenticated}
						<a href={resolve('/dashboard')} class="btn btn-primary w-full">
							{tr('auth.verification.goToDashboard')}
						</a>
					{:else}
						<a href={resolve('/login')} class="btn btn-primary w-full">
							{tr('auth.verification.goToLogin')}
						</a>
					{/if}
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
						{tr('auth.verification.error')}
					</h2>
					<p class="text-text-muted mb-6">{errorMessage}</p>

					<div class="space-y-3">
						{#if $authStore.isAuthenticated && (errorCode === 'token_expired' || errorCode === 'invalid_token')}
							<button
								onclick={handleResend}
								disabled={isResending}
								class="btn btn-primary w-full"
							>
								{#if isResending}
									<span class="relative inline-flex h-3 w-3 mr-2"
										><span
											class="animate-ping absolute inline-flex h-full w-full rounded-full bg-accent-400 opacity-75"
										></span><span
											class="relative inline-flex rounded-full h-3 w-3 bg-accent"
										></span></span
									>
								{/if}
								{tr('auth.verification.requestNew')}
							</button>
						{/if}

						{#if $authStore.isAuthenticated}
							<a href={resolve('/dashboard')} class="btn btn-ghost w-full">
								{tr('auth.verification.goToDashboard')}
							</a>
						{:else}
							<a href={resolve('/login')} class="btn btn-ghost w-full">
								{tr('auth.verification.goToLogin')}
							</a>
						{/if}
					</div>
				</div>
			{:else}
				<!-- idle state - no token provided -->
				<div class="text-center py-8">
					<p class="text-text-muted mb-6">{tr('auth.verification.checkEmail')}</p>
					<a href={resolve('/login')} class="btn btn-primary w-full">
						{tr('auth.verification.goToLogin')}
					</a>
				</div>
			{/if}
		</div>
	</div>
</div>
