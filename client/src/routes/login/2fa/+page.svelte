<script lang="ts">
	import { get } from 'svelte/store';
	import { authStore } from '$lib/stores/auth';
	import { t } from '$lib/stores/i18n';
	import { toastStore } from '$lib/stores/toast';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { onMount } from 'svelte';

	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);

	let code = $state('');
	let backupCode = $state('');
	let isLoading = $state(false);
	let useBackup = $state(false);
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

<div class="min-h-screen py-12 px-4 sm:px-6 lg:px-8">
	<div class="max-w-md mx-auto">
		<div class="bg-white rounded-lg shadow-lg p-6 sm:p-8">
			<div class="mb-6 text-center">
				<div
					class="mx-auto w-12 h-12 bg-cyan-100 rounded-full flex items-center justify-center mb-4"
				>
					<svg
						class="w-6 h-6 text-cyan-600"
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
				<h1 class="text-2xl font-bold text-gray-900">
					{tr('auth.twoFactor.title')}
				</h1>
				<p class="text-sm text-gray-600 mt-2">
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
							class="block text-sm font-medium text-gray-700 mb-1"
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
							class="w-full px-4 py-2 bg-white border border-gray-300 rounded-md focus:ring-cyan-500 focus:border-cyan-500 font-mono"
							placeholder={tr('auth.twoFactor.backupCodePlaceholder')}
						/>
					</div>
				{:else}
					<div>
						<label
							for="totpCode"
							class="block text-sm font-medium text-gray-700 mb-1"
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
							class="w-full px-4 py-3 bg-white border border-gray-300 rounded-md focus:ring-cyan-500 focus:border-cyan-500 text-center text-2xl font-mono tracking-widest"
							placeholder={tr('auth.twoFactor.codePlaceholder')}
						/>
					</div>
				{/if}

				{#if error}
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
								<p class="text-sm font-medium text-red-800">{error}</p>
							</div>
						</div>
					</div>
				{/if}

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
						{tr('auth.twoFactor.verifying')}
					{:else}
						{tr('auth.twoFactor.verifyButton')}
					{/if}
				</button>
			</form>

			<div class="mt-4 space-y-2">
				<button
					type="button"
					onclick={toggleBackupMode}
					class="w-full text-sm text-cyan-600 hover:text-cyan-500 text-center"
				>
					{#if useBackup}
						{tr('auth.twoFactor.useAuthenticator')}
					{:else}
						{tr('auth.twoFactor.useBackupCode')}
					{/if}
				</button>

				<a
					href={resolve('/login')}
					class="block w-full text-sm text-gray-500 hover:text-gray-700 text-center"
				>
					{tr('auth.twoFactor.backToLogin')}
				</a>
			</div>
		</div>
	</div>
</div>
