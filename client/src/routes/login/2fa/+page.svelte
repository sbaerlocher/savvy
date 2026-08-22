<script lang="ts">
	import { get } from 'svelte/store';
	import { authStore } from '$lib/stores/auth';
	import { t } from '$lib/stores/i18n';
	import { toastStore } from '$lib/stores/toast';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { onMount } from 'svelte';
	import { platform } from '$lib/utils/platform';
	import M3TextField from '$lib/components/ui/M3TextField.svelte';

	// Android mockup (screen-AuthAndroid, 2FA frame) replaces the plain card
	// with a centered M3 card and a six-box code display.
	const ANDROID = platform === 'android';

	// iOS mockup (screen-AuthIOS) replaces the desktop two-column
	// layout with a grouped-inset card, so it branches at the top level.
	const IOS = platform === 'ios';

	// Desktop mockup (screen-AuthDesktop, board C) enlarges the centered card's
	// chrome, icon circle and code field. Mobile keeps its layout.
	const isDesktop = platform === 'other';

	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);

	let code = $state('');
	let backupCode = $state('');
	let isLoading = $state(false);
	let useBackup = $state(false);
	// The real input is transparent, so the boxes render the focus ring.
	let focused = $state(false);
	let error = $state('');

	onMount(() => {
		// Redirect if not in 2FA flow
		if (!$authStore.requires2FA) {
			goto(resolve('/login'));
		}
	});

	async function handleVerify(e: Event) {
		e.preventDefault();
		isLoading = true;
		error = '';

		try {
			const data = useBackup
				? { backup_code: backupCode.trim() }
				: { code: code.trim() };

			await authStore.verify2FA(data);
			toastStore.success(tr('auth.login.success'));
			goto(resolve('/dashboard'));
		} catch {
			error = $authStore.error || tr('auth.twoFactor.invalidCode');
		} finally {
			isLoading = false;
		}
	}

	// The mockup shows six separate boxes; the flow still submits one code
	// string, so a single hidden input drives six rendered cells.
	const codeCells = $derived(
		Array.from({ length: 6 }, (_, i) => code[i] ?? '')
	);

	function toggleBackupMode() {
		useBackup = !useBackup;
		error = '';
		code = '';
		backupCode = '';
	}
</script>

<svelte:head>
	<title>{tr('auth.twoFactor.title')} - {tr('common.appName')}</title>
</svelte:head>

{#if IOS}
	<!-- The app shell pads `main` by 16px; the mockup insets this screen by
	     22px from the device edge, so cancel the shell padding first. -->
	<div
		class="-mx-4 flex min-h-[calc(100dvh-var(--spacing-page-y))] items-center justify-center px-card sm:-mx-6 lg:-mx-8"
	>
		<div
			class="w-full max-w-86 rounded-inset bg-surface px-6 pt-7 pb-6 shadow-card"
		>
			<div class="mb-card text-center">
				<span
					class="mb-3.5 inline-flex h-14 w-14 items-center justify-center rounded-full bg-accent-100"
				>
					<svg
						width="26"
						height="26"
						viewBox="0 0 24 24"
						fill="none"
						stroke="currentColor"
						class="text-accent-700"
						stroke-width="2"
						stroke-linecap="round"
						stroke-linejoin="round"
						aria-hidden="true"
					>
						<rect x="4" y="11" width="16" height="10" rx="2"></rect>
						<path d="M8 11V7a4 4 0 018 0v4"></path>
					</svg>
				</span>
				<h1 class="text-heading text-text">{tr('auth.twoFactor.heading')}</h1>
				<p class="mt-1.25 text-label font-normal text-text-muted">
					{#if useBackup}
						{tr('auth.twoFactor.backupCodeLabel')}
					{:else}
						{tr('auth.twoFactor.subtitle')}
					{/if}
				</p>
			</div>

			<form onsubmit={handleVerify}>
				{#if useBackup}
					<div class="mb-5 flex flex-col gap-1.75">
						<label
							for="backup-code-ios"
							class="pl-0.5 text-label text-text-ink2"
						>
							{tr('auth.twoFactor.backupCodeLabel')}
						</label>
						<input
							id="backup-code-ios"
							type="text"
							autocomplete="off"
							required
							bind:value={backupCode}
							disabled={isLoading}
							class="h-12.5 rounded-lg border border-border-field bg-surface-2 px-3.75 font-mono text-amount text-text placeholder:text-text-placeholder focus:border-accent focus:outline-none"
							placeholder={tr('auth.twoFactor.backupCodePlaceholder')}
						/>
					</div>
				{:else}
					<!-- One real input drives the six boxes: the boxes only mirror
					     `code`, so autofill and the numeric keyboard keep working. -->
					<div class="relative mb-5">
						<label for="totp-code-ios" class="sr-only">
							{tr('auth.twoFactor.codeLabel')}
						</label>
						<!-- The mockup's own measures do not fit: six 44px boxes plus five
						     9px gaps need 309px, while its 344px card leaves 296px between
						     the 24px paddings. The gap gives way first (6x44 + 5x6 = 294px),
						     and the columns cap at the mockup's 44px so narrower screens
						     shrink the boxes evenly instead of overhanging the card. -->
						<div class="grid grid-cols-6 justify-items-center gap-1.5">
							{#each Array(6) as _, i (i)}
								<span
									class="flex h-14 w-full max-w-11 items-center justify-center rounded-lg font-mono text-title font-semibold {focused &&
									code.length === i
										? 'border-2 border-accent-600 bg-surface text-accent-700'
										: 'border border-border-field bg-surface-2 text-text'}"
								>
									{#if focused && code.length === i}
										<!-- The real input is transparent, so the mockup's caret
										     is drawn on the active box instead. -->
										<span class="h-6.5 w-0.5 rounded-full bg-accent-600"></span>
									{:else}
										{code[i] ?? ''}
									{/if}
								</span>
							{/each}
						</div>
						<input
							id="totp-code-ios"
							type="text"
							inputmode="numeric"
							pattern="[0-9]*"
							maxlength="6"
							autocomplete="one-time-code"
							required
							bind:value={code}
							disabled={isLoading}
							onfocus={() => (focused = true)}
							onblur={() => (focused = false)}
							class="absolute inset-0 h-full w-full cursor-default text-transparent caret-transparent opacity-0"
						/>
					</div>
				{/if}

				{#if error}
					<p
						class="mb-4 rounded-lg bg-danger-50 p-3 text-body-sm text-danger-800"
					>
						{error}
					</p>
				{/if}

				<button
					type="submit"
					disabled={isLoading}
					class="flex h-13 w-full items-center justify-center rounded-xl bg-accent-600 text-amount font-semibold text-on-accent shadow-accent disabled:opacity-50"
				>
					{#if isLoading}
						{tr('auth.twoFactor.verifying')}
					{:else}
						{tr('auth.twoFactor.verifyButton')}
					{/if}
				</button>
			</form>

			<div class="mt-4.5 flex flex-col items-center gap-3">
				<button
					type="button"
					onclick={toggleBackupMode}
					class="inline-flex items-center gap-1.5 text-label text-accent-600"
				>
					<svg
						width="15"
						height="15"
						viewBox="0 0 24 24"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
						stroke-linecap="round"
						stroke-linejoin="round"
						aria-hidden="true"
					>
						<path d="M4 7h16M4 12h10M4 17h7"></path>
					</svg>
					{#if useBackup}
						{tr('auth.twoFactor.useAuthenticator')}
					{:else}
						{tr('auth.twoFactor.useBackupCode')}
					{/if}
				</button>
				<a
					href={resolve('/login')}
					class="text-label font-normal text-text-subtle"
				>
					{tr('auth.twoFactor.backToLogin')}
				</a>
			</div>
		</div>
	</div>
{:else if ANDROID}
	<!-- Android M3: centered tonal card with a six-cell code display. -->
	<div class="flex min-h-dvh items-center justify-center px-5">
		<div
			class="bg-m3-card rounded-m3-lg w-full max-w-88 px-[var(--spacing-card)] pt-7 pb-6"
		>
			<div class="mb-6 text-center">
				<span
					class="bg-accent-100 rounded-m3-full mb-3.5 inline-flex h-14 w-14 items-center justify-center"
				>
					<svg
						class="text-accent-700 h-6.5 w-6.5"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
						stroke-linecap="round"
						stroke-linejoin="round"
						viewBox="0 0 24 24"
					>
						<rect x="4" y="11" width="16" height="10" rx="2" />
						<path d="M8 11V7a4 4 0 018 0v4" />
					</svg>
				</span>
				<h1 class="text-heading text-text">{tr('auth.twoFactor.title')}</h1>
				<p class="text-label text-text-muted mt-1.5 font-normal">
					{useBackup
						? tr('auth.twoFactor.backupCodeLabel')
						: tr('auth.twoFactor.subtitle')}
				</p>
			</div>

			<form class="flex flex-col gap-6" onsubmit={handleVerify}>
				{#if useBackup}
					<M3TextField
						id="backupCode"
						name="backupCode"
						autocomplete="off"
						required
						disabled={isLoading}
						bind:value={backupCode}
						label={tr('auth.twoFactor.backupCodeLabel')}
					/>
				{:else}
					<div class="relative">
						<input
							id="totpCode"
							type="text"
							inputmode="numeric"
							pattern="[0-9]*"
							maxlength="6"
							autocomplete="one-time-code"
							required
							bind:value={code}
							disabled={isLoading}
							aria-label={tr('auth.twoFactor.codeLabel')}
							class="absolute inset-0 z-10 h-full w-full opacity-0"
						/>
						<div class="flex justify-center gap-2.5" aria-hidden="true">
							{#each codeCells as cell, i (i)}
								<span
									class="text-stat rounded-m3-xs flex h-14 w-11 items-center justify-center font-mono {i ===
									code.length
										? 'border-accent-600 text-accent-700 border-2'
										: 'border-border-field text-text border'}"
								>
									{cell}
								</span>
							{/each}
						</div>
					</div>
				{/if}

				{#if error}
					<div class="bg-danger-50 rounded-m3-sm -mt-2 px-4 py-3">
						<p class="text-body-sm text-danger-800 font-medium">{error}</p>
					</div>
				{/if}

				<button
					type="submit"
					disabled={isLoading}
					class="text-subheading bg-accent-600 text-on-accent rounded-m3-full shadow-accent h-12.5 w-full disabled:opacity-50"
				>
					{isLoading
						? tr('auth.twoFactor.verifying')
						: tr('auth.twoFactor.verifyButton')}
				</button>
			</form>

			<div class="mt-5 flex flex-col items-center gap-3.5">
				<button
					type="button"
					onclick={toggleBackupMode}
					class="text-label text-accent-600 inline-flex items-center gap-1.5 font-semibold"
				>
					<svg
						class="h-3.75 w-3.75"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
						stroke-linecap="round"
						stroke-linejoin="round"
						viewBox="0 0 24 24"
					>
						<path d="M4 7h16M4 12h10M4 17h7" />
					</svg>
					{useBackup
						? tr('auth.twoFactor.useAuthenticator')
						: tr('auth.twoFactor.useBackupCode')}
				</button>
				<a
					href={resolve('/login')}
					class="text-label text-text-subtle font-normal"
				>
					{tr('auth.twoFactor.backToLogin')}
				</a>
			</div>
		</div>
	</div>
{:else}
	<div class="min-h-screen py-12 px-4 sm:px-6 lg:px-8">
		<div class="max-w-md mx-auto">
			<div
				class={isDesktop
					? 'bg-surface rounded-lg lg:rounded-2xl shadow-lg lg:shadow-card p-6 sm:p-8 lg:p-10'
					: 'bg-white rounded-lg shadow-lg p-6 sm:p-8'}
			>
				<div class="mb-6 text-center {isDesktop ? 'lg:mb-7' : ''}">
					<div
						class="mx-auto bg-accent-100 rounded-full flex items-center justify-center mb-4 {isDesktop
							? 'w-12 h-12 lg:h-14 lg:w-14'
							: 'w-12 h-12'}"
					>
						<svg
							class="text-accent {isDesktop
								? 'w-6 h-6 lg:h-6.5 lg:w-6.5 lg:text-accent-700'
								: 'w-6 h-6'}"
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
					</div>
					<h1
						class="font-bold text-text {isDesktop
							? 'text-2xl lg:text-title'
							: 'text-2xl'}"
					>
						{tr('auth.twoFactor.title')}
					</h1>
					<p
						class="text-text-muted mt-2 {isDesktop
							? 'text-sm lg:text-label lg:font-normal'
							: 'text-sm'}"
					>
						{#if useBackup}
							{tr('auth.twoFactor.backupCodeLabel')}
						{:else}
							{tr('auth.twoFactor.subtitle')}
						{/if}
					</p>
				</div>

				<form onsubmit={handleVerify} class="space-y-6">
					{#if useBackup}
						<div>
							<label
								for="backupCode"
								class="block font-medium text-text-ink2 mb-1 {isDesktop
									? 'text-sm lg:text-body lg:font-semibold lg:mb-2'
									: 'text-sm'}"
							>
								{tr('auth.twoFactor.backupCodeLabel')}
							</label>
							<input
								id="backupCode"
								type="text"
								autocomplete="off"
								required
								bind:value={backupCode}
								disabled={isLoading}
								class="w-full px-4 py-2 bg-white border border-border-field rounded-md focus:ring-accent focus:border-accent font-mono {isDesktop
									? 'lg:h-15 lg:rounded-lg lg:py-0 lg:text-center lg:text-body lg:font-semibold'
									: ''}"
								placeholder={tr('auth.twoFactor.backupCodePlaceholder')}
							/>
						</div>
					{:else}
						<div>
							<label
								for="totpCode"
								class="block font-medium text-text-ink2 mb-1 {isDesktop
									? 'text-sm lg:text-body lg:font-semibold lg:mb-2'
									: 'text-sm'}"
							>
								{tr('auth.twoFactor.codeLabel')}
							</label>
							<input
								id="totpCode"
								type="text"
								inputmode="numeric"
								pattern="[0-9]*"
								maxlength="6"
								autocomplete="one-time-code"
								required
								bind:value={code}
								disabled={isLoading}
								class="w-full px-4 py-3 bg-white border border-border-field rounded-md focus:ring-accent focus:border-accent text-center font-mono tracking-widest {isDesktop
									? 'text-2xl lg:h-15 lg:rounded-lg lg:py-0 lg:text-stat lg:font-semibold lg:tracking-[0.32em]'
									: 'text-2xl'}"
								placeholder={tr('auth.twoFactor.codePlaceholder')}
							/>
						</div>
					{/if}

					{#if error}
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
									<p class="text-sm font-medium text-danger-800">{error}</p>
								</div>
							</div>
						</div>
					{/if}

					<button
						type="submit"
						disabled={isLoading}
						class="btn btn-primary w-full {isDesktop
							? 'lg:h-12 lg:rounded-lg lg:text-subheading lg:font-semibold lg:shadow-accent'
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
							{tr('auth.twoFactor.verifying')}
						{:else}
							{tr('auth.twoFactor.verifyButton')}
						{/if}
					</button>
				</form>

				<div
					class="mt-4 space-y-2 {isDesktop ? 'lg:mt-5.5 lg:space-y-3.5' : ''}"
				>
					<button
						type="button"
						onclick={toggleBackupMode}
						class="w-full text-sm text-accent hover:text-accent text-center {isDesktop
							? 'lg:inline-flex lg:items-center lg:justify-center lg:gap-1.5 lg:text-body lg:font-semibold'
							: ''}"
					>
						{#if isDesktop}
							<svg
								class="hidden h-4 w-4 lg:block"
								viewBox="0 0 24 24"
								fill="none"
								stroke="currentColor"
								stroke-width="2"
								stroke-linecap="round"
								stroke-linejoin="round"
							>
								<path d="M4 7h16M4 12h10M4 17h7" />
							</svg>
						{/if}
						{#if useBackup}
							{tr('auth.twoFactor.useAuthenticator')}
						{:else}
							{tr('auth.twoFactor.useBackupCode')}
						{/if}
					</button>

					<a
						href={resolve('/login')}
						class="block w-full text-sm text-text-subtle hover:text-text-ink2 text-center {isDesktop
							? 'lg:text-label lg:font-normal'
							: ''}"
					>
						{tr('auth.twoFactor.backToLogin')}
					</a>
				</div>
			</div>
		</div>
	</div>
{/if}
