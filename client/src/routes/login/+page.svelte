<script lang="ts">
	import { get } from 'svelte/store';
	import { authStore } from '$lib/stores/auth';
	import { t } from '$lib/stores/i18n';
	import { toastStore } from '$lib/stores/toast';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { logger } from '$lib/utils/logger';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	const pageLogger = logger.child('LoginPage');

	// Svelte 5 compatible translation wrapper
	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);

	let email = $state('');
	let password = $state('');
	let isLoading = $state(false);
	let configLoaded = $state(false);
	let oauthEnabled = $state(false);
	let oauthLoginURL = $state('');
	let localLoginEnabled = $state(false);
	let registrationEnabled = $state(false);
	let oauthError = $state('');

	// Redirect if already logged in
	onMount(async () => {
		if ($authStore.isAuthenticated) {
			goto('/dashboard');
			return;
		}

		// Check for OAuth error from callback redirect
		oauthError = $page.url.searchParams.get('error') || '';

		// Clean error from URL to prevent showing stale errors on page reload
		if (oauthError) {
			const cleanUrl = new URL(window.location.href);
			cleanUrl.searchParams.delete('error');
			window.history.replaceState({}, '', cleanUrl.pathname);
		}

		// Load app config before rendering login options
		try {
			const response = await fetch('/api/v1/config');
			if (response.ok) {
				const config = await response.json();
				oauthEnabled = config.oauth?.enabled || false;
				oauthLoginURL = config.oauth?.login_url || '/auth/oauth/login';
				localLoginEnabled = config.local_login_enabled ?? true;
				registrationEnabled = config.registration_enabled ?? true;

				// Auto-redirect to OAuth if local login is disabled
				// but NOT if we just came back from a failed OAuth attempt
				if (!localLoginEnabled && oauthEnabled && !oauthError) {
					window.location.href = oauthLoginURL;
					return;
				}
			} else {
				// Fallback: show local login if config endpoint fails
				localLoginEnabled = true;
			}
		} catch (error) {
			pageLogger.error('Failed to load config', { error });
			// Fallback: show local login if config endpoint unreachable
			localLoginEnabled = true;
		} finally {
			configLoaded = true;
		}
	});

	async function handleLogin(e: Event) {
		e.preventDefault();
		isLoading = true;

		try {
			await authStore.login({ email, password });

			// Check if 2FA is required
			if ($authStore.requires2FA) {
				goto('/login/2fa');
				return;
			}

			toastStore.success(tr('auth.login.success'));
			goto('/dashboard');
		} catch (error) {
			toastStore.error($authStore.error || tr('auth.login.error'));
		} finally {
			isLoading = false;
		}
	}
</script>

<svelte:head>
	<title>{tr('auth.login.title')} - {tr('common.appName')}</title>
</svelte:head>

<div class="min-h-screen py-12 px-4 sm:px-6 lg:px-8">
	<div class="max-w-5xl mx-auto">
		{#if !configLoaded}
			<LoadingSpinner fullPage />
		{:else}
			<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
				<!-- Left column: Login Form (2/3 width) -->
				<div class="lg:col-span-2">
					<div class="bg-white rounded-lg shadow-lg p-6 sm:p-8">
						<div class="mb-8 flex items-center gap-4">
							<img src="/logo.png" alt="Savvy Logo" class="h-12 sm:h-16" />
							<h1 class="text-3xl font-bold text-gray-900">
								{tr('auth.login.title')}
							</h1>
						</div>

						<!-- OAuth Error Message -->
						{#if oauthError}
							<div class="rounded-md bg-red-50 p-4 mb-6">
								<div class="flex">
									<div class="flex-shrink-0">
										<svg
											class="h-5 w-5 text-red-400"
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
										<p class="text-sm font-medium text-red-800">
											{tr('auth.login.oauthError')}
										</p>
									</div>
								</div>
							</div>
						{/if}

						<!-- OAuth Button (if enabled) -->
						{#if oauthEnabled}
							<div class="mb-6">
								<a href={oauthLoginURL} class="btn btn-ghost w-full">
									<svg class="h-5 w-5" viewBox="0 0 24 24" fill="currentColor">
										<path
											d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 18c-4.41 0-8-3.59-8-8s3.59-8 8-8 8 3.59 8 8-3.59 8-8 8zm-1-13h2v6h-2zm0 8h2v2h-2z"
										></path>
									</svg>
									{tr('auth.login.oauthButton')}
								</a>
							</div>

							<!-- Divider only if both OAuth and local login are enabled -->
							{#if localLoginEnabled}
								<div class="relative my-6">
									<div class="absolute inset-0 flex items-center">
										<div class="w-full border-t border-gray-300"></div>
									</div>
									<div class="relative flex justify-center text-sm">
										<span class="px-2 bg-white text-gray-500"
											>{tr('auth.login.orDivider')}</span
										>
									</div>
								</div>
							{/if}
						{/if}

						<!-- Local Login Form (only if enabled) -->
						{#if localLoginEnabled}
							<form class="space-y-6" onsubmit={handleLogin}>
								<div>
									<label
										for="email"
										class="block text-sm font-medium text-gray-700 mb-1"
									>
										{tr('auth.login.email')}
									</label>
									<input
										id="email"
										name="email"
										type="email"
										autocomplete="email"
										required
										bind:value={email}
										disabled={isLoading}
										class="w-full px-4 py-2 bg-white border border-gray-300 rounded-md focus:ring-cyan-500 focus:border-cyan-500"
										placeholder={tr('auth.login.email')}
									/>
								</div>

								<div>
									<label
										for="password"
										class="block text-sm font-medium text-gray-700 mb-1"
									>
										{tr('auth.login.password')}
									</label>
									<input
										id="password"
										name="password"
										type="password"
										autocomplete="current-password"
										required
										bind:value={password}
										disabled={isLoading}
										class="w-full px-4 py-2 bg-white border border-gray-300 rounded-md focus:ring-cyan-500 focus:border-cyan-500"
										placeholder={tr('auth.login.password')}
									/>
								</div>

								{#if $authStore.error}
									<div class="rounded-md bg-red-50 p-4">
										<div class="flex">
											<div class="flex-shrink-0">
												<svg
													class="h-5 w-5 text-red-400"
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
												<p class="text-sm font-medium text-red-800">
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
													class="animate-ping absolute inline-flex h-full w-full rounded-full bg-cyan-400 opacity-75"
												></span><span
													class="relative inline-flex rounded-full h-3 w-3 bg-cyan-500"
												></span></span
											>
											{tr('auth.login.loggingIn')}
										{:else}
											{tr('auth.login.loginButton')}
										{/if}
									</button>
								</div>

								<div class="flex items-center justify-between pt-4">
									{#if registrationEnabled}
										<a
											href="/register"
											class="font-medium text-cyan-600 hover:text-cyan-500 text-sm"
										>
											{tr('auth.login.noAccountYet')}
											{tr('auth.login.registerNow')}
										</a>
									{:else}
										<span></span>
									{/if}
									<a
										href="/forgot-password"
										class="font-medium text-cyan-600 hover:text-cyan-500 text-sm"
									>
										{tr('auth.login.forgotPassword')}
									</a>
								</div>
							</form>
						{/if}
					</div>
				</div>

				<!-- Right column: Information (1/3 width) -->
				<div class="lg:col-span-1">
					<div class="bg-white rounded-lg shadow-lg p-6">
						<h2 class="text-xl font-bold text-gray-900 mb-4">
							{tr('auth.login.infoTitle')}
						</h2>
						<p class="text-sm text-gray-600 mb-4">
							{tr('auth.login.infoDescription')}
						</p>

						<div class="space-y-4">
							<div class="flex items-start">
								<svg
									class="w-5 h-5 text-green-500 mt-0.5 mr-3 flex-shrink-0"
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
									<h3 class="text-sm font-medium text-gray-900">
										{tr('auth.login.info1Title')}
									</h3>
									<p class="text-xs text-gray-600 mt-1">
										{tr('auth.login.info1Desc')}
									</p>
								</div>
							</div>

							<div class="flex items-start">
								<svg
									class="w-5 h-5 text-green-500 mt-0.5 mr-3 flex-shrink-0"
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
									<h3 class="text-sm font-medium text-gray-900">
										{tr('auth.login.info2Title')}
									</h3>
									<p class="text-xs text-gray-600 mt-1">
										{tr('auth.login.info2Desc')}
									</p>
								</div>
							</div>

							<div class="flex items-start">
								<svg
									class="w-5 h-5 text-green-500 mt-0.5 mr-3 flex-shrink-0"
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
									<h3 class="text-sm font-medium text-gray-900">
										{tr('auth.login.info3Title')}
									</h3>
									<p class="text-xs text-gray-600 mt-1">
										{tr('auth.login.info3Desc')}
									</p>
								</div>
							</div>
						</div>

						<div class="mt-6 pt-6 border-t border-gray-200">
							<div class="flex items-center text-sm text-gray-600">
								<svg
									class="w-5 h-5 text-cyan-500 mr-2 flex-shrink-0"
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
								<span class="text-xs">{tr('auth.login.securityNote')}</span>
							</div>
						</div>
					</div>
				</div>
			</div>
		{/if}
	</div>
</div>
