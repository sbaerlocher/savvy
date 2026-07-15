<script lang="ts">
	import { get } from 'svelte/store';
	import { authStore } from '$lib/stores/auth';
	import { t } from '$lib/stores/i18n';
	import { toastStore } from '$lib/stores/toast';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { onMount } from 'svelte';
	import { logger } from '$lib/utils/logger';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	const pageLogger = logger.child('RegisterPage');

	// Svelte 5 compatible translation wrapper
	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);

	let email = $state('');
	let password = $state('');
	let firstName = $state('');
	let lastName = $state('');
	let isLoading = $state(false);
	let configLoaded = $state(false);

	// Redirect if already logged in or registration disabled
	onMount(async () => {
		if ($authStore.isAuthenticated) {
			goto(resolve('/dashboard'));
			return;
		}

		// Load app config to check if registration is enabled
		try {
			const response = await fetch('/api/v1/config');
			if (response.ok) {
				const config = await response.json();
				const registrationEnabled = config.registration_enabled ?? false;

				if (!registrationEnabled) {
					goto(resolve('/login'));
					return;
				}
			} else {
				// Config endpoint failed - redirect to login for safety
				goto(resolve('/login'));
				return;
			}
		} catch (error) {
			pageLogger.error('Failed to load config', { error });
			goto(resolve('/login'));
			return;
		} finally {
			configLoaded = true;
		}
	});

	async function handleRegister(e: Event) {
		e.preventDefault();
		isLoading = true;

		try {
			await authStore.register({
				email,
				password,
				first_name: firstName || undefined,
				last_name: lastName || undefined
			});
			toastStore.success(tr('auth.register.success'));
			goto(resolve('/dashboard'));
		} catch {
			toastStore.error($authStore.error || tr('auth.register.error'));
		} finally {
			isLoading = false;
		}
	}
</script>

<svelte:head>
	<title>{tr('auth.register.title')} - {tr('common.appName')}</title>
</svelte:head>

<div class="min-h-screen py-12 px-4 sm:px-6 lg:px-8">
	<div class="max-w-5xl mx-auto">
		{#if !configLoaded}
			<LoadingSpinner fullPage />
		{:else}
			<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
				<!-- Left column: Registration Form (2/3 width) -->
				<div class="lg:col-span-2">
					<div class="bg-white rounded-lg shadow-lg p-6 sm:p-8">
						<div class="mb-8 flex items-center gap-4">
							<img src="/logo.png" alt="Savvy Logo" class="h-12 sm:h-16" />
							<h1 class="text-3xl font-bold text-text">
								{tr('auth.register.title')}
							</h1>
						</div>

						<form class="space-y-6" onsubmit={handleRegister}>
							<div>
								<label
									for="email"
									class="block text-sm font-medium text-text-ink2 mb-1"
								>
									{tr('auth.register.email')} *
								</label>
								<input
									id="email"
									name="email"
									type="email"
									autocomplete="email"
									required
									bind:value={email}
									disabled={isLoading}
									class="w-full px-4 py-2 bg-white border border-border-field rounded-md focus:ring-accent focus:border-accent"
									placeholder={tr('auth.register.email')}
								/>
							</div>

							<div>
								<label
									for="password"
									class="block text-sm font-medium text-text-ink2 mb-1"
								>
									{tr('auth.register.password')} *
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
									placeholder={tr('auth.register.password')}
								/>
							</div>

							<div>
								<label
									for="first_name"
									class="block text-sm font-medium text-text-ink2 mb-1"
								>
									{tr('auth.register.firstName')} *
								</label>
								<input
									id="first_name"
									name="first_name"
									type="text"
									autocomplete="given-name"
									required
									bind:value={firstName}
									disabled={isLoading}
									class="w-full px-4 py-2 bg-white border border-border-field rounded-md focus:ring-accent focus:border-accent"
									placeholder={tr('auth.register.firstName')}
								/>
							</div>

							<div>
								<label
									for="last_name"
									class="block text-sm font-medium text-text-ink2 mb-1"
								>
									{tr('auth.register.lastName')} *
								</label>
								<input
									id="last_name"
									name="last_name"
									type="text"
									autocomplete="family-name"
									required
									bind:value={lastName}
									disabled={isLoading}
									class="w-full px-4 py-2 bg-white border border-border-field rounded-md focus:ring-accent focus:border-accent"
									placeholder={tr('auth.register.lastName')}
								/>
							</div>

							{#if $authStore.error}
								<div class="rounded-md bg-danger-50 p-4">
									<div class="flex">
										<div class="flex-shrink-0">
											<svg
												class="h-5 w-5 text-danger-400"
												viewBox="0 0 20 20"
												fill="currentColor"
											>
												<path
													fill-rule="evenodd"
													d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
													clip-rule="evenodd"
												/>
											</svg>
										</div>
										<div class="ml-3">
											<p class="text-sm font-medium text-danger-800">
												{$authStore.error}
											</p>
										</div>
									</div>
								</div>
							{/if}

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
										{tr('auth.register.registering')}
									{:else}
										{tr('auth.register.registerButton')}
									{/if}
								</button>
							</div>

							<div class="text-center pt-4">
								<a
									href={resolve('/login')}
									class="font-medium text-accent hover:text-accent"
								>
									{tr('auth.register.hasAccount')}
								</a>
							</div>
						</form>
					</div>
				</div>

				<!-- Right column: Information (1/3 width) -->
				<div class="lg:col-span-1">
					<div class="bg-white rounded-lg shadow-lg p-6">
						<h2 class="text-xl font-bold text-text mb-4">
							{tr('auth.register.benefitsTitle')}
						</h2>
						<p class="text-sm text-text-muted mb-4">
							{tr('auth.register.benefitsDescription')}
						</p>

						<div class="space-y-4">
							<div class="flex items-start">
								<svg
									class="w-5 h-5 text-success-500 mt-0.5 mr-3 flex-shrink-0"
									fill="currentColor"
									viewBox="0 0 20 20"
								>
									<path
										fill-rule="evenodd"
										d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
										clip-rule="evenodd"
									/>
								</svg>
								<div>
									<h3 class="text-sm font-medium text-text">
										{tr('auth.register.benefit1Title')}
									</h3>
									<p class="text-xs text-text-muted mt-1">
										{tr('auth.register.benefit1Desc')}
									</p>
								</div>
							</div>

							<div class="flex items-start">
								<svg
									class="w-5 h-5 text-success-500 mt-0.5 mr-3 flex-shrink-0"
									fill="currentColor"
									viewBox="0 0 20 20"
								>
									<path
										fill-rule="evenodd"
										d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
										clip-rule="evenodd"
									/>
								</svg>
								<div>
									<h3 class="text-sm font-medium text-text">
										{tr('auth.register.benefit2Title')}
									</h3>
									<p class="text-xs text-text-muted mt-1">
										{tr('auth.register.benefit2Desc')}
									</p>
								</div>
							</div>

							<div class="flex items-start">
								<svg
									class="w-5 h-5 text-success-500 mt-0.5 mr-3 flex-shrink-0"
									fill="currentColor"
									viewBox="0 0 20 20"
								>
									<path
										fill-rule="evenodd"
										d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
										clip-rule="evenodd"
									/>
								</svg>
								<div>
									<h3 class="text-sm font-medium text-text">
										{tr('auth.register.benefit3Title')}
									</h3>
									<p class="text-xs text-text-muted mt-1">
										{tr('auth.register.benefit3Desc')}
									</p>
								</div>
							</div>

							<div class="flex items-start">
								<svg
									class="w-5 h-5 text-success-500 mt-0.5 mr-3 flex-shrink-0"
									fill="currentColor"
									viewBox="0 0 20 20"
								>
									<path
										fill-rule="evenodd"
										d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
										clip-rule="evenodd"
									/>
								</svg>
								<div>
									<h3 class="text-sm font-medium text-text">
										{tr('auth.register.benefit4Title')}
									</h3>
									<p class="text-xs text-text-muted mt-1">
										{tr('auth.register.benefit4Desc')}
									</p>
								</div>
							</div>
						</div>

						<div class="mt-6 pt-6 border-t border-border">
							<div class="flex items-center text-sm text-text-muted">
								<svg
									class="w-5 h-5 text-accent mr-2 flex-shrink-0"
									fill="none"
									stroke="currentColor"
									viewBox="0 0 24 24"
								>
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										stroke-width="2"
										d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"
									/>
								</svg>
								<span class="text-xs">{tr('auth.register.securityNote')}</span>
							</div>
						</div>
					</div>
				</div>
			</div>
		{/if}
	</div>
</div>
