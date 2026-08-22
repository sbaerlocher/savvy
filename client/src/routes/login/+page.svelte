<script lang="ts">
	import { get } from 'svelte/store';
	import { authStore } from '$lib/stores/auth';
	import { t } from '$lib/stores/i18n';
	import { toastStore } from '$lib/stores/toast';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { logger } from '$lib/utils/logger';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import M3TextField from '$lib/components/ui/M3TextField.svelte';
	import AuthCardIOS from '$lib/components/auth/AuthCardIOS.svelte';
	import { platform } from '$lib/utils/platform';

	// Android mockup (screen-AuthAndroid) replaces the split layout with a
	// centered M3 card, so it branches at the top level instead of overriding.
	const ANDROID = platform === 'android';

	// iOS mockup (screen-AuthIOS) replaces the desktop two-column
	// layout with a grouped-inset card, so it branches at the top level.
	const IOS = platform === 'ios';

	// Desktop mockup (screen-AuthDesktop, board A) swaps the logo for an accent
	// card tile and raises the type steps. Mobile keeps its layout.
	const isDesktop = platform === 'other';

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
	let showPassword = $state(false);
	let passwordVisible = $state(false);

	// Redirect if already logged in
	onMount(async () => {
		if ($authStore.isAuthenticated) {
			goto(resolve('/dashboard'));
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
				goto(resolve('/login/2fa'));
				return;
			}

			toastStore.success(tr('auth.login.success'));
			goto(resolve('/dashboard'));
		} catch {
			toastStore.error($authStore.error || tr('auth.login.error'));
		} finally {
			isLoading = false;
		}
	}
</script>

<svelte:head>
	<title>{tr('auth.login.title')} - {tr('common.appName')}</title>
</svelte:head>

{#if IOS}
	{#if !configLoaded}
		<LoadingSpinner fullPage />
	{:else}
		<AuthCardIOS
			title={tr('auth.login.title')}
			subtitle={tr('auth.login.subtitle')}
		>
			{#if oauthError}
				<p
					class="mb-4 rounded-lg bg-danger-50 p-3 text-body-sm text-danger-800"
				>
					{tr('auth.login.oauthError')}
				</p>
			{/if}

			{#if oauthEnabled}
				<a
					href={oauthLoginURL}
					rel="external"
					class="mb-4 flex h-13 w-full items-center justify-center gap-2.5 rounded-lg border border-border-field bg-surface-2 text-subheading font-semibold text-text"
				>
					<svg
						width="18"
						height="18"
						viewBox="0 0 24 24"
						fill="currentColor"
						aria-hidden="true"
					>
						<path
							d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 18c-4.41 0-8-3.59-8-8s3.59-8 8-8 8 3.59 8 8-3.59 8-8 8zm-1-13h2v6h-2zm0 8h2v2h-2z"
						></path>
					</svg>
					{tr('auth.login.oauthButton')}
				</a>

				{#if localLoginEnabled}
					<div
						class="mb-4.5 flex items-center gap-3 text-body-sm font-medium text-text-subtle"
					>
						<span class="h-px flex-1 bg-border"></span>
						{tr('auth.login.orDivider')}
						<span class="h-px flex-1 bg-border"></span>
					</div>
				{/if}
			{/if}

			{#if localLoginEnabled}
				<form class="flex flex-col gap-3.5" onsubmit={handleLogin}>
					<div class="flex flex-col gap-1.75">
						<label for="email-ios" class="pl-0.5 text-label text-text-ink2">
							{tr('auth.login.email')}
						</label>
						<input
							id="email-ios"
							name="email"
							type="email"
							autocomplete="email"
							required
							bind:value={email}
							disabled={isLoading}
							class="h-12.5 rounded-lg border border-border-field bg-surface-2 px-3.75 text-amount font-normal text-text placeholder:text-text-placeholder focus:border-accent focus:outline-none"
							placeholder={tr('auth.login.emailPlaceholder')}
						/>
					</div>

					<div class="flex flex-col gap-1.75">
						<label for="password-ios" class="pl-0.5 text-label text-text-ink2">
							{tr('auth.login.password')}
						</label>
						<div class="relative">
							<input
								id="password-ios"
								name="password"
								type={showPassword ? 'text' : 'password'}
								autocomplete="current-password"
								required
								bind:value={password}
								disabled={isLoading}
								class="h-12.5 w-full rounded-lg border border-border-field bg-surface-2 pl-3.75 pr-11.5 text-amount font-normal text-text placeholder:text-text-placeholder focus:border-accent focus:outline-none"
								placeholder={tr('auth.login.passwordPlaceholder')}
							/>
							<!-- Centre a 24x24 hit area (WCAG 2.2 AA) on the mockup's
							     19px glyph, so the padding grows outward only. -->
							<button
								type="button"
								disabled={isLoading}
								onclick={() => (showPassword = !showPassword)}
								aria-label={showPassword
									? tr('common.hidePassword')
									: tr('common.showPassword')}
								aria-pressed={showPassword}
								class="absolute right-3.125 top-1/2 flex size-6 -translate-y-1/2 items-center justify-center rounded-xs text-text-faint focus-visible:outline focus-visible:outline-2 focus-visible:outline-accent-600 disabled:opacity-50"
							>
								<svg
									width="19"
									height="19"
									viewBox="0 0 24 24"
									fill="none"
									stroke="currentColor"
									stroke-width="2"
									stroke-linecap="round"
									stroke-linejoin="round"
									aria-hidden="true"
								>
									<path d="M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7-10-7-10-7z"
									></path>
									<circle cx="12" cy="12" r="3"></circle>
									{#if showPassword}
										<path d="M4 20L20 4"></path>
									{/if}
								</svg>
							</button>
						</div>
					</div>

					{#if $authStore.error}
						<p class="rounded-lg bg-danger-50 p-3 text-body-sm text-danger-800">
							{$authStore.error}
						</p>
					{/if}

					<button
						type="submit"
						disabled={isLoading}
						class="mt-0.5 flex h-13 w-full items-center justify-center rounded-xl bg-accent-600 text-amount font-semibold text-on-accent shadow-accent disabled:opacity-50"
					>
						{#if isLoading}
							{tr('auth.login.loggingIn')}
						{:else}
							{tr('auth.login.loginButton')}
						{/if}
					</button>

					<div class="flex items-center justify-between pt-0.5">
						{#if registrationEnabled}
							<a href={resolve('/register')} class="text-label text-accent-600"
								>{tr('auth.register.title')}</a
							>
						{:else}
							<span></span>
						{/if}
						<a
							href={resolve('/forgot-password')}
							class="text-label text-accent-600"
						>
							{tr('auth.login.forgotPassword')}
						</a>
					</div>
				</form>
			{:else}
				<p class="text-center text-body-sm text-text-subtle">
					{tr('auth.login.oauthOnlyNote')}
				</p>
			{/if}
		</AuthCardIOS>
	{/if}
{:else if ANDROID}
	<!-- Android M3: centered tonal card, no info column, no logo header. -->
	<div class="flex min-h-dvh items-center justify-center px-5">
		{#if !configLoaded}
			<LoadingSpinner fullPage />
		{:else}
			<div
				class="bg-m3-card rounded-m3-lg w-full max-w-88 px-[var(--spacing-card)] pt-6.5 pb-6"
			>
				<div class="mb-6 flex flex-col items-center text-center">
					<span
						class="bg-accent-600 rounded-m3-md shadow-accent mb-3.5 flex h-13 w-13 items-center justify-center"
					>
						<svg
							class="text-on-accent h-7 w-7"
							fill="none"
							stroke="currentColor"
							stroke-width="2"
							stroke-linecap="round"
							stroke-linejoin="round"
							viewBox="0 0 24 24"
						>
							<rect x="2" y="5" width="20" height="14" rx="3" />
							<path d="M2 10h20" />
							<path d="M6 15h4" />
						</svg>
					</span>
					<h1 class="text-heading text-text">{tr('auth.login.title')}</h1>
					<p class="text-label text-text-muted mt-1 font-normal">
						{tr('auth.login.subtitle')}
					</p>
				</div>

				{#if oauthError}
					<div class="bg-danger-50 rounded-m3-sm mb-4.5 px-4 py-3">
						<p class="text-body-sm text-danger-800 font-medium">
							{tr('auth.login.oauthError')}
						</p>
					</div>
				{/if}

				{#if oauthEnabled}
					<a
						href={oauthLoginURL}
						rel="external"
						class="text-body bg-accent-100 text-accent-850 rounded-m3-full mb-4.5 flex h-12 w-full items-center justify-center gap-2.5 font-semibold"
					>
						<svg class="h-4.5 w-4.5" viewBox="0 0 24 24" fill="currentColor">
							<path
								d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 18c-4.41 0-8-3.59-8-8s3.59-8 8-8 8 3.59 8 8-3.59 8-8 8zm-1-13h2v6h-2zm0 8h2v2h-2z"
							/>
						</svg>
						{tr('auth.login.oauthButton')}
					</a>

					{#if localLoginEnabled}
						<div class="text-text-subtle mb-5 flex items-center gap-3">
							<span class="bg-border h-px flex-1"></span>
							<span class="text-chip font-medium">
								{tr('auth.login.orDivider')}
							</span>
							<span class="bg-border h-px flex-1"></span>
						</div>
					{/if}
				{/if}

				{#if localLoginEnabled}
					<form class="flex flex-col gap-5" onsubmit={handleLogin}>
						<M3TextField
							id="email"
							name="email"
							type="email"
							autocomplete="email"
							required
							disabled={isLoading}
							bind:value={email}
							label={tr('auth.login.email')}
						/>
						<M3TextField
							id="password"
							name="password"
							type="password"
							autocomplete="current-password"
							required
							trailingIcon
							disabled={isLoading}
							bind:value={password}
							label={tr('auth.login.password')}
						/>

						{#if $authStore.error}
							<div class="bg-danger-50 rounded-m3-sm px-4 py-3">
								<p class="text-body-sm text-danger-800 font-medium">
									{$authStore.error}
								</p>
							</div>
						{/if}

						<button
							type="submit"
							disabled={isLoading}
							class="text-subheading bg-accent-600 text-on-accent rounded-m3-full shadow-accent h-12.5 w-full disabled:opacity-50"
						>
							{isLoading
								? tr('auth.login.loggingIn')
								: tr('auth.login.loginButton')}
						</button>

						<div class="-mt-1.5 flex items-center justify-between">
							{#if registrationEnabled}
								<a
									href={resolve('/register')}
									class="text-label text-accent-600 font-semibold"
								>
									{tr('auth.login.registerNow')}
								</a>
							{:else}
								<span></span>
							{/if}
							<a
								href={resolve('/forgot-password')}
								class="text-label text-accent-600 font-semibold"
							>
								{tr('auth.login.forgotPassword')}
							</a>
						</div>
					</form>
				{:else}
					<p class="text-body-sm text-text-subtle text-center">
						{tr('auth.login.oauthOnlyNote')}
					</p>
				{/if}
			</div>
		{/if}
	</div>
{:else}
	<div class="min-h-screen py-12 px-4 sm:px-6 lg:px-8">
		<div class="max-w-5xl mx-auto">
			{#if !configLoaded}
				<LoadingSpinner fullPage />
			{:else}
				<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
					<!-- Left column: Login Form (2/3 width) -->
					<div class="lg:col-span-2">
						<div
							class={isDesktop
								? 'bg-surface rounded-lg lg:rounded-2xl shadow-lg lg:shadow-card p-6 sm:p-8 lg:p-10'
								: 'bg-white rounded-lg shadow-lg p-6 sm:p-8'}
						>
							<div
								class="mb-8 flex items-center gap-4 {isDesktop
									? 'lg:mb-7.5 lg:gap-3.5'
									: ''}"
							>
								{#if isDesktop}
									<span
										class="flex h-13 w-13 flex-none items-center justify-center rounded-lg bg-accent shadow-accent"
									>
										<svg
											class="h-7 w-7 text-on-accent"
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
										? 'text-3xl lg:text-screen-title'
										: 'text-3xl'}"
								>
									{tr('auth.login.title')}
								</h1>
							</div>

							<!-- OAuth Error Message -->
							{#if oauthError}
								<div class="rounded-md bg-danger-50 p-4 mb-6">
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
												{tr('auth.login.oauthError')}
											</p>
										</div>
									</div>
								</div>
							{/if}

							<!-- OAuth Button (if enabled) -->
							{#if oauthEnabled}
								<div class="mb-6 {isDesktop ? 'lg:mb-5' : ''}">
									<a
										href={oauthLoginURL}
										rel="external"
										class="btn btn-ghost w-full {isDesktop
											? 'lg:h-12 lg:gap-2.5 lg:rounded-lg lg:border-border-field lg:text-body lg:font-semibold'
											: ''}"
									>
										<svg
											class="h-5 w-5"
											viewBox="0 0 24 24"
											fill="currentColor"
										>
											<path
												d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 18c-4.41 0-8-3.59-8-8s3.59-8 8-8 8 3.59 8 8-3.59 8-8 8zm-1-13h2v6h-2zm0 8h2v2h-2z"
											></path>
										</svg>
										{tr('auth.login.oauthButton')}
									</a>
								</div>

								<!-- Divider only if both OAuth and local login are enabled -->
								{#if localLoginEnabled}
									<div
										class="relative my-6 {isDesktop
											? 'lg:my-0 lg:mb-5.5 lg:flex lg:items-center lg:gap-3.5'
											: ''}"
									>
										<div
											class="absolute inset-0 flex items-center {isDesktop
												? 'lg:static lg:flex-1'
												: ''}"
										>
											<div class="w-full border-t border-border-field"></div>
										</div>
										<div
											class="relative flex justify-center text-sm {isDesktop
												? 'lg:static lg:text-label lg:font-normal'
												: ''}"
										>
											<span
												class="px-2 text-text-subtle {isDesktop
													? 'bg-surface lg:px-0'
													: 'bg-white'}">{tr('auth.login.orDivider')}</span
											>
										</div>
										{#if isDesktop}
											<div
												class="hidden lg:block lg:h-px lg:flex-1 lg:bg-border-field"
											></div>
										{/if}
									</div>
								{/if}
							{/if}

							<!-- Local Login Form (only if enabled) -->
							{#if localLoginEnabled}
								<form
									class="space-y-6 {isDesktop ? 'lg:space-y-5' : ''}"
									onsubmit={handleLogin}
								>
									<div>
										<label
											for="email"
											class="block font-medium text-text-ink2 mb-1 {isDesktop
												? 'text-sm lg:text-body lg:font-semibold lg:mb-1.5'
												: 'text-sm'}"
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
											class="w-full px-4 py-2 bg-white border border-border-field rounded-md focus:ring-accent focus:border-accent {isDesktop
												? 'lg:h-11.5 lg:rounded-lg lg:py-0 lg:text-body'
												: ''}"
											placeholder={tr('auth.login.email')}
										/>
									</div>

									<div>
										<label
											for="password"
											class="block font-medium text-text-ink2 mb-1 {isDesktop
												? 'text-sm lg:text-body lg:font-semibold lg:mb-1.5'
												: 'text-sm'}"
										>
											{tr('auth.login.password')}
										</label>
										<div class="relative">
											<input
												id="password"
												name="password"
												type={passwordVisible ? 'text' : 'password'}
												autocomplete="current-password"
												required
												bind:value={password}
												disabled={isLoading}
												class="w-full px-4 py-2 pr-11.5 bg-white border border-border-field rounded-md focus:ring-accent focus:border-accent {isDesktop
													? 'lg:h-11.5 lg:rounded-lg lg:py-0 lg:text-body'
													: ''}"
												placeholder={tr('auth.login.password')}
											/>
											<button
												type="button"
												onclick={() => (passwordVisible = !passwordVisible)}
												aria-label={passwordVisible
													? tr('common.hidePassword')
													: tr('common.showPassword')}
												class="absolute right-3.5 top-1/2 -mr-1.5 -translate-y-1/2 p-1.5 text-text-faint hover:text-text-ink2"
											>
												{#if passwordVisible}
													<svg
														class="h-4.5 w-4.5"
														viewBox="0 0 24 24"
														fill="none"
														stroke="currentColor"
														stroke-width="2"
														stroke-linecap="round"
														stroke-linejoin="round"
													>
														<path
															d="M2 12s3.5-7 10-7c1.9 0 3.5.6 4.9 1.4M22 12s-3.5 7-10 7c-1.9 0-3.5-.6-4.9-1.4"
														/>
														<path d="M3 3l18 18" />
													</svg>
												{:else}
													<svg
														class="h-4.5 w-4.5"
														viewBox="0 0 24 24"
														fill="none"
														stroke="currentColor"
														stroke-width="2"
														stroke-linecap="round"
														stroke-linejoin="round"
													>
														<path
															d="M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7-10-7-10-7z"
														/>
														<circle cx="12" cy="12" r="3" />
													</svg>
												{/if}
											</button>
										</div>
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

									<div class={isDesktop ? 'pt-2 lg:pt-0' : 'pt-2'}>
										<Button
											type="submit"
											class="w-full {isDesktop
												? 'lg:h-12 lg:rounded-lg lg:text-subheading lg:font-semibold lg:shadow-accent'
												: ''}"
											loading={isLoading}
										>
											{#if isLoading}
												{tr('auth.login.loggingIn')}
											{:else}
												{tr('auth.login.loginButton')}
											{/if}
										</Button>
									</div>

									<div
										class="flex items-center justify-between pt-4 {isDesktop
											? 'lg:pt-0'
											: ''}"
									>
										{#if registrationEnabled}
											<a
												href={resolve('/register')}
												class="font-medium text-accent hover:text-accent text-sm {isDesktop
													? 'lg:text-body lg:font-semibold'
													: ''}"
											>
												{tr('auth.login.noAccountYet')}
												{tr('auth.login.registerNow')}
											</a>
										{:else}
											<span></span>
										{/if}
										<a
											href={resolve('/forgot-password')}
											class="font-medium text-accent hover:text-accent text-sm {isDesktop
												? 'lg:text-body lg:font-semibold'
												: ''}"
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
						<div
							class={isDesktop
								? 'bg-surface rounded-lg lg:rounded-2xl shadow-lg lg:shadow-card p-6 lg:p-7'
								: 'bg-white rounded-lg shadow-lg p-6'}
						>
							<h2
								class="font-bold text-text {isDesktop
									? 'text-xl lg:text-heading mb-4 lg:mb-1.5'
									: 'text-xl mb-4'}"
							>
								{tr('auth.login.infoTitle')}
							</h2>
							<p
								class="text-text-muted mb-4 {isDesktop
									? 'text-sm lg:text-label lg:font-normal lg:mb-5'
									: 'text-sm'}"
							>
								{tr('auth.login.infoDescription')}
							</p>

							<div class="space-y-4">
								<div class="flex items-start">
									<svg
										class="w-5 h-5 text-success-600 mt-0.5 mr-3 flex-shrink-0"
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
										<h3
											class="font-medium text-text {isDesktop
												? 'text-sm lg:text-body lg:font-semibold'
												: 'text-sm'}"
										>
											{tr('auth.login.info1Title')}
										</h3>
										<p
											class="text-text-muted mt-1 {isDesktop
												? 'text-xs lg:text-body-sm'
												: 'text-xs'}"
										>
											{tr('auth.login.info1Desc')}
										</p>
									</div>
								</div>

								<div class="flex items-start">
									<svg
										class="w-5 h-5 text-success-600 mt-0.5 mr-3 flex-shrink-0"
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
										<h3
											class="font-medium text-text {isDesktop
												? 'text-sm lg:text-body lg:font-semibold'
												: 'text-sm'}"
										>
											{tr('auth.login.info2Title')}
										</h3>
										<p
											class="text-text-muted mt-1 {isDesktop
												? 'text-xs lg:text-body-sm'
												: 'text-xs'}"
										>
											{tr('auth.login.info2Desc')}
										</p>
									</div>
								</div>

								<div class="flex items-start">
									<svg
										class="w-5 h-5 text-success-600 mt-0.5 mr-3 flex-shrink-0"
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
										<h3
											class="font-medium text-text {isDesktop
												? 'text-sm lg:text-body lg:font-semibold'
												: 'text-sm'}"
										>
											{tr('auth.login.info3Title')}
										</h3>
										<p
											class="text-text-muted mt-1 {isDesktop
												? 'text-xs lg:text-body-sm'
												: 'text-xs'}"
										>
											{tr('auth.login.info3Desc')}
										</p>
									</div>
								</div>
							</div>

							<div
								class="border-t border-border {isDesktop
									? 'mt-6 pt-6 lg:mt-5.5 lg:pt-5'
									: 'mt-6 pt-6'}"
							>
								<div class="flex items-center text-sm text-text-muted">
									<svg
										class="text-accent mr-2 flex-shrink-0 {isDesktop
											? 'w-5 h-5 lg:w-4.5 lg:h-4.5'
											: 'w-5 h-5'}"
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
{/if}
