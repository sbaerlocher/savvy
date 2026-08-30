<script lang="ts">
	import PageShell from '$lib/components/layout/PageShell.svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { authApi } from '$lib/api';
	import AuthCardIOS from '$lib/components/auth/AuthCardIOS.svelte';
	import AuthResultIOS from '$lib/components/auth/AuthResultIOS.svelte';
	import M3TextField from '$lib/components/ui/M3TextField.svelte';
	import { authStore } from '$lib/stores/auth';
	import { t } from '$lib/stores/i18n';
	import { toastStore } from '$lib/stores/toast';
	import { platform } from '$lib/utils/platform';
	import { onMount } from 'svelte';
	import { get } from 'svelte/store';

	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);

	// Desktop mockup (screen-PasswordResetDesktop) wraps the split in a raised
	// panel and swaps the logo for an accent card tile. Mobile keeps its layout.
	const isDesktop = platform === 'other';

	// Android mockup (screen-PasswordResetAndroid) replaces the split with a
	// centered M3 card, so it branches at the top level instead of overriding.
	const ANDROID = platform === 'android';

	// iOS mockup (screen-PasswordResetIOS) uses its own grouped-inset card,
	// so it branches at the top level like Android does.
	const IOS = platform === 'ios';

	let email = $state('');
	let isLoading = $state(false);
	let submitted = $state(false);
	let configLoaded = $state(false);

	onMount(async () => {
		if ($authStore.isAuthenticated) {
			goto(resolve('/dashboard'));
			return;
		}
		// Initialize CSRF cookie by hitting a public GET endpoint. Gate the
		// form on this so the submit button can't fire a mutation before the
		// csrf_token cookie exists (the API client throws without it).
		try {
			await fetch('/api/v1/config');
		} catch {
			// Ignore: a failed config fetch still lets the user try; the
			// mutation surfaces its own error if CSRF is genuinely missing.
		} finally {
			configLoaded = true;
		}
	});

	async function handleSubmit(e: Event) {
		e.preventDefault();
		isLoading = true;

		try {
			await authApi.forgotPassword(email);
			submitted = true;
		} catch {
			toastStore.error(tr('auth.forgotPassword.error'));
		} finally {
			isLoading = false;
		}
	}
</script>

<svelte:head>
	<title>{tr('auth.forgotPassword.title')} - {tr('common.appName')}</title>
</svelte:head>

{#if IOS}
	<AuthCardIOS title={tr('auth.forgotPassword.title')} compact={submitted}>
		{#if submitted}
			<AuthResultIOS
				tone="success"
				icon="envelope"
				heading={tr('auth.forgotPassword.success')}
				message={tr('auth.forgotPassword.successMessage')}
			>
				<a
					href={resolve('/login')}
					class="flex h-13 w-full items-center justify-center rounded-xl bg-accent-600 text-amount font-semibold text-on-accent shadow-accent"
				>
					{tr('auth.forgotPassword.backToLogin')}
				</a>
			</AuthResultIOS>
		{:else}
			<p class="mb-5 text-body text-text-muted">
				{tr('auth.forgotPassword.description')}
			</p>

			<form class="flex flex-col gap-4" onsubmit={handleSubmit}>
				<div class="flex flex-col gap-1.75">
					<label for="email-ios" class="pl-0.5 text-label text-text-ink2">
						{tr('auth.forgotPassword.email')}
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
						placeholder={tr('auth.forgotPassword.emailPlaceholder')}
					/>
				</div>

				<button
					type="submit"
					disabled={isLoading || !configLoaded}
					class="flex h-13 w-full items-center justify-center rounded-xl bg-accent-600 text-amount font-semibold text-on-accent shadow-accent disabled:opacity-50"
				>
					{#if isLoading}
						{tr('auth.forgotPassword.submitting')}
					{:else}
						{tr('auth.forgotPassword.submitButton')}
					{/if}
				</button>

				<div class="pt-1 text-center">
					<a href={resolve('/login')} class="text-label text-accent-600">
						{tr('auth.forgotPassword.backToLogin')}
					</a>
				</div>
			</form>
		{/if}
	</AuthCardIOS>
{:else if ANDROID}
	<!-- Android M3: centered tonal card, no info column, no logo header. -->
	<PageShell width="bleed">
		{#if submitted}
			<div
				class="bg-m3-card rounded-m3-lg w-full max-w-88 px-card pt-8 pb-7 text-center"
			>
				<span
					class="bg-success-100 rounded-m3-full mb-4 inline-flex h-16 w-16 items-center justify-center"
				>
					<svg
						class="text-success-600 h-8 w-8"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
						stroke-linecap="round"
						stroke-linejoin="round"
						viewBox="0 0 24 24"
					>
						<path
							d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
						/>
					</svg>
				</span>
				<h1 class="text-heading text-text mb-2">
					{tr('auth.forgotPassword.success')}
				</h1>
				<p class="text-body text-text-muted mb-6">
					{tr('auth.forgotPassword.successMessage')}
				</p>
				<a
					href={resolve('/login')}
					class="text-subheading bg-accent-600 text-on-accent rounded-m3-full shadow-accent flex h-12 w-full items-center justify-center"
				>
					{tr('auth.forgotPassword.backToLogin')}
				</a>
			</div>
		{:else}
			<div class="bg-m3-card rounded-m3-lg w-full max-w-88 px-card pt-6 pb-6">
				<div class="mb-4.5 flex flex-col items-center text-center">
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
					<h1 class="text-heading text-text">
						{tr('auth.forgotPassword.title')}
					</h1>
				</div>

				<p class="text-body text-text-muted mb-6 text-center">
					{tr('auth.forgotPassword.description')}
				</p>

				<form class="flex flex-col gap-5" onsubmit={handleSubmit}>
					<M3TextField
						id="email"
						name="email"
						type="email"
						autocomplete="email"
						required
						disabled={isLoading}
						bind:value={email}
						label={tr('auth.forgotPassword.email')}
					/>

					<button
						type="submit"
						disabled={isLoading || !configLoaded}
						class="text-subheading bg-accent-600 text-on-accent rounded-m3-full shadow-accent h-12 w-full disabled:opacity-50"
					>
						{isLoading
							? tr('auth.forgotPassword.submitting')
							: tr('auth.forgotPassword.submitButton')}
					</button>

					<div class="-mt-1.5 text-center">
						<a
							href={resolve('/login')}
							class="text-label text-accent-600 font-semibold"
						>
							{tr('auth.forgotPassword.backToLogin')}
						</a>
					</div>
				</form>
			</div>
		{/if}
	</PageShell>
{:else}
	<div class="min-h-screen py-12 px-4 sm:px-6 lg:px-8">
		<div class="max-w-5xl mx-auto">
			<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
				<!-- Left column: Form (2/3 width) -->
				<div class="lg:col-span-2">
					<div
						class={isDesktop
							? 'bg-surface rounded-lg lg:rounded-2xl shadow-lg lg:shadow-card p-6 sm:p-8 lg:p-10'
							: 'bg-white rounded-lg shadow-lg p-6 sm:p-8'}
					>
						<div
							class="mb-8 flex items-center gap-4 {isDesktop
								? submitted
									? 'lg:mb-3 lg:gap-3.5'
									: 'lg:mb-6 lg:gap-3.5'
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
									? 'text-2xl lg:text-screen-title'
									: 'text-2xl'}"
							>
								{tr('auth.forgotPassword.title')}
							</h1>
						</div>

						{#if submitted}
							<div
								class="text-center py-8 {isDesktop
									? 'lg:mx-auto lg:max-w-md lg:pt-8 lg:pb-2'
									: ''}"
							>
								<div
									class="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-success-100 {isDesktop
										? 'lg:mb-4.5'
										: ''}"
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
											d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
										/>
									</svg>
								</div>
								<h2
									class="font-semibold text-text mb-2 {isDesktop
										? 'text-xl lg:text-heading lg:font-semibold'
										: 'text-xl'}"
								>
									{tr('auth.forgotPassword.success')}
								</h2>
								<p
									class="text-text-muted mb-6 {isDesktop
										? 'lg:text-body lg:mb-6.5'
										: ''}"
								>
									{tr('auth.forgotPassword.successMessage')}
								</p>
								<a
									href={resolve('/login')}
									class="btn btn-primary w-full {isDesktop
										? 'lg:h-12 lg:rounded-lg lg:text-subheading lg:shadow-accent'
										: ''}"
								>
									{tr('auth.forgotPassword.backToLogin')}
								</a>
							</div>
						{:else}
							<p
								class="text-text-muted mb-6 {isDesktop
									? 'lg:text-body lg:max-w-lg lg:mb-6.5'
									: ''}"
							>
								{tr('auth.forgotPassword.description')}
							</p>

							<form
								class="space-y-6 {isDesktop ? 'lg:space-y-5.5' : ''}"
								onsubmit={handleSubmit}
							>
								<div>
									<label
										for="email"
										class="block font-medium text-text-ink2 mb-1 {isDesktop
											? 'text-sm lg:text-body lg:font-semibold lg:mb-1.5'
											: 'text-sm'}"
									>
										{tr('auth.forgotPassword.email')}
									</label>
									<input
										id="email"
										name="email"
										type="email"
										autocomplete="email"
										required
										bind:value={email}
										disabled={isLoading}
										class="input {isDesktop
											? 'lg:h-11.5 lg:rounded-lg lg:text-body'
											: ''}"
										placeholder={tr('auth.forgotPassword.emailPlaceholder')}
									/>
								</div>

								<div class={isDesktop ? 'pt-2 lg:pt-0' : 'pt-2'}>
									<button
										type="submit"
										disabled={isLoading || !configLoaded}
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
											{tr('auth.forgotPassword.submitting')}
										{:else}
											{tr('auth.forgotPassword.submitButton')}
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
										{tr('auth.forgotPassword.backToLogin')}
									</a>
								</div>
							</form>
						{/if}
					</div>
				</div>

				<!-- Right column: Info (1/3 width) -->
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
							{tr('auth.forgotPassword.infoTitle')}
						</h2>
						<p
							class="text-text-muted mb-4 {isDesktop
								? 'text-sm lg:text-label lg:mb-5'
								: 'text-sm'}"
						>
							{tr('auth.forgotPassword.infoDescription')}
						</p>

						<div class="space-y-4">
							<div class="flex items-start">
								<svg
									class="w-5 h-5 text-accent mt-0.5 mr-3 flex-shrink-0"
									fill="none"
									stroke="currentColor"
									viewBox="0 0 24 24"
								>
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										stroke-width="2"
										d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
									/>
								</svg>
								<div>
									<h3
										class="font-medium text-text {isDesktop
											? 'text-sm lg:text-body lg:font-semibold'
											: 'text-sm'}"
									>
										{tr('auth.forgotPassword.step1Title')}
									</h3>
									<p
										class="text-text-muted mt-1 {isDesktop
											? 'text-xs lg:text-body-sm'
											: 'text-xs'}"
									>
										{tr('auth.forgotPassword.step1Desc')}
									</p>
								</div>
							</div>

							<div class="flex items-start">
								<svg
									class="w-5 h-5 text-accent mt-0.5 mr-3 flex-shrink-0"
									fill="none"
									stroke="currentColor"
									viewBox="0 0 24 24"
								>
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										stroke-width="2"
										d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z"
									/>
								</svg>
								<div>
									<h3
										class="font-medium text-text {isDesktop
											? 'text-sm lg:text-body lg:font-semibold'
											: 'text-sm'}"
									>
										{tr('auth.forgotPassword.step2Title')}
									</h3>
									<p
										class="text-text-muted mt-1 {isDesktop
											? 'text-xs lg:text-body-sm'
											: 'text-xs'}"
									>
										{tr('auth.forgotPassword.step2Desc')}
									</p>
								</div>
							</div>

							<div class="flex items-start">
								<svg
									class="w-5 h-5 text-accent mt-0.5 mr-3 flex-shrink-0"
									fill="none"
									stroke="currentColor"
									viewBox="0 0 24 24"
								>
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										stroke-width="2"
										d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
									/>
								</svg>
								<div>
									<h3
										class="font-medium text-text {isDesktop
											? 'text-sm lg:text-body lg:font-semibold'
											: 'text-sm'}"
									>
										{tr('auth.forgotPassword.step3Title')}
									</h3>
									<p
										class="text-text-muted mt-1 {isDesktop
											? 'text-xs lg:text-body-sm'
											: 'text-xs'}"
									>
										{tr('auth.forgotPassword.step3Desc')}
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
								<span class="text-xs"
									>{tr('auth.forgotPassword.securityNote')}</span
								>
							</div>
						</div>
					</div>
				</div>
			</div>
		</div>
	</div>
{/if}
