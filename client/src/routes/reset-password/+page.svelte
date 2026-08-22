<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/stores';
	import { ApiError, authApi } from '$lib/api';
	import { authStore } from '$lib/stores/auth';
	import { t } from '$lib/stores/i18n';
	import { toastStore } from '$lib/stores/toast';
	import { platform } from '$lib/utils/platform';
	import { onMount } from 'svelte';
	import { get } from 'svelte/store';

	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);

	// Desktop mockup (screen-PasswordResetDesktop, boards 3-5) wraps the narrow
	// card in a raised panel and swaps the logo for an accent card tile.
	const isDesktop = platform === 'other';

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
		<div
			class={isDesktop
				? 'bg-surface rounded-lg lg:rounded-2xl shadow-lg lg:shadow-card p-6 sm:p-8 lg:px-8 lg:py-8.5'
				: 'bg-white rounded-lg shadow-lg p-6 sm:p-8'}
		>
			<div
				class="flex items-center gap-4 lg:gap-3.5 {isDesktop &&
				status === 'form'
					? 'mb-8 lg:mb-4.5'
					: 'mb-8 lg:mb-2.5'}"
			>
				{#if isDesktop}
					<span
						class="flex h-11.5 w-11.5 flex-none items-center justify-center rounded-lg bg-accent shadow-accent"
					>
						<svg
							class="h-6 w-6 text-on-accent"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							stroke-width="2"
							stroke-linecap="round"
							stroke-linejoin="round"
						>
							<rect x="2" y="5" width="20" height="14" rx="3" />
							<path d="M2 10h20" />
							<path d="M6 15h4" />
						</svg>
					</span>
				{:else}
					<img src="/logo.png" alt="Savvy Logo" class="h-12 sm:h-16" />
				{/if}
				<h1
					class="font-bold text-text {isDesktop
						? 'text-2xl lg:text-stat lg:font-bold'
						: 'text-2xl'}"
				>
					{tr('auth.resetPassword.title')}
				</h1>
			</div>

			{#if status === 'success'}
				<div class="text-center py-8 {isDesktop ? 'lg:pt-7 lg:pb-1' : ''}">
					<div
						class="mx-auto flex items-center justify-center h-16 w-16 rounded-full bg-success-100 mb-4 lg:mb-4.5"
					>
						<svg
							class="h-8 w-8 text-success-600"
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
					<h2
						class="font-semibold text-text mb-2 {isDesktop
							? 'text-xl lg:text-heading lg:font-semibold'
							: 'text-xl'}"
					>
						{tr('auth.resetPassword.success')}
					</h2>
					<p
						class="text-text-muted mb-6 {isDesktop
							? 'lg:text-body lg:mb-6.5'
							: ''}"
					>
						{tr('auth.resetPassword.successMessage')}
					</p>
					<a
						href={resolve('/login')}
						class="btn btn-primary w-full {isDesktop
							? 'lg:h-12 lg:rounded-lg lg:text-subheading lg:shadow-accent'
							: ''}"
					>
						{tr('auth.resetPassword.goToLogin')}
					</a>
				</div>
			{:else if status === 'error'}
				<div class="text-center py-8 {isDesktop ? 'lg:pt-7 lg:pb-1' : ''}">
					<div
						class="mx-auto flex items-center justify-center h-16 w-16 rounded-full bg-danger-100 mb-4 lg:mb-4.5"
					>
						<svg
							class="h-8 w-8 text-danger-600"
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
					<h2
						class="font-semibold text-text mb-2 {isDesktop
							? 'text-xl lg:text-heading lg:font-semibold'
							: 'text-xl'}"
					>
						{tr('auth.resetPassword.error')}
					</h2>
					<p
						class="text-text-muted mb-6 {isDesktop
							? 'lg:text-body lg:mb-6.5'
							: ''}"
					>
						{errorMessage}
					</p>

					<div class="space-y-3">
						{#if errorCode === 'token_expired' || errorCode === 'invalid_token'}
							<a
								href={resolve('/forgot-password')}
								class="btn btn-primary w-full {isDesktop
									? 'lg:h-12 lg:rounded-lg lg:text-subheading lg:shadow-accent'
									: ''}"
							>
								{tr('auth.resetPassword.requestNew')}
							</a>
						{/if}
						<a
							href={resolve('/login')}
							class="btn btn-ghost w-full {isDesktop
								? 'lg:h-12 lg:rounded-lg lg:text-subheading lg:text-text'
								: ''}"
						>
							{tr('auth.resetPassword.goToLogin')}
						</a>
					</div>
				</div>
			{:else}
				<p class="text-text-muted mb-6 {isDesktop ? 'lg:text-body' : ''}">
					{tr('auth.resetPassword.description')}
				</p>

				<form
					class="space-y-6 {isDesktop ? 'lg:space-y-5' : ''}"
					onsubmit={handleSubmit}
				>
					<div>
						<label
							for="password"
							class="block font-medium text-text-ink2 mb-1 {isDesktop
								? 'text-sm lg:text-body lg:font-semibold lg:mb-1.5'
								: 'text-sm'}"
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
							class="input {isDesktop
								? 'lg:h-11.5 lg:rounded-lg lg:text-body'
								: ''}"
							placeholder={tr('auth.resetPassword.passwordPlaceholder')}
						/>
					</div>

					<div>
						<label
							for="confirmPassword"
							class="block font-medium text-text-ink2 mb-1 {isDesktop
								? 'text-sm lg:text-body lg:font-semibold lg:mb-1.5'
								: 'text-sm'}"
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
							class="input {isDesktop
								? 'lg:h-11.5 lg:rounded-lg lg:text-body'
								: ''}"
							placeholder={tr('auth.resetPassword.confirmPasswordPlaceholder')}
						/>
					</div>

					<div class={isDesktop ? 'pt-2 lg:pt-0' : 'pt-2'}>
						<button
							type="submit"
							disabled={isLoading}
							class="btn btn-primary w-full {isDesktop
								? 'lg:h-12 lg:rounded-lg lg:text-subheading lg:shadow-accent'
								: ''}"
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

					<div class="text-center pt-4 {isDesktop ? 'lg:pt-0' : ''}">
						<a
							href={resolve('/login')}
							class="font-medium text-accent hover:text-accent {isDesktop
								? 'lg:text-body lg:font-semibold'
								: ''}"
						>
							{tr('auth.resetPassword.goToLogin')}
						</a>
					</div>
				</form>
			{/if}
		</div>
	</div>
</div>
