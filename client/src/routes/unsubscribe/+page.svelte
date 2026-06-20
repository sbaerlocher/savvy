<script lang="ts">
	import { get } from 'svelte/store';
	import { authStore } from '$lib/stores/auth';
	import { authApi } from '$lib/api';
	import { t } from '$lib/stores/i18n';
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { resolve } from '$app/paths';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	const tr = (key: string) => get(t)(key);

	let status = $state<'loading' | 'success' | 'error' | 'idle'>('idle');
	let errorMessage = $state('');
	let unsubType = $state<'notifications' | 'reminders'>('notifications');

	onMount(async () => {
		const token = $page.url.searchParams.get('token');
		const type = $page.url.searchParams.get('type');
		if (type === 'reminders') {
			unsubType = 'reminders';
		}

		if (!token) {
			status = 'error';
			errorMessage = tr('unsubscribe.invalidToken');
			return;
		}

		status = 'loading';

		try {
			if (unsubType === 'reminders') {
				await authApi.unsubscribeReminders(token);
			} else {
				await authApi.unsubscribeNotifications(token);
			}
			status = 'success';
		} catch (error: unknown) {
			status = 'error';
			const code =
				typeof error === 'object' && error !== null && 'code' in error
					? String((error as { code?: unknown }).code ?? '')
					: '';

			if (code === 'token_expired') {
				errorMessage = tr('unsubscribe.tokenExpired');
			} else if (code === 'token_used') {
				errorMessage = tr('unsubscribe.tokenUsed');
			} else {
				errorMessage = tr('unsubscribe.invalidToken');
			}
		}
	});
</script>

<svelte:head>
	<title>{tr('unsubscribe.title')} - {tr('common.appName')}</title>
</svelte:head>

<div class="min-h-screen py-12 px-4 sm:px-6 lg:px-8">
	<div class="max-w-5xl mx-auto">
		<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
			<!-- Left column: Status Content (2/3 width) -->
			<div class="lg:col-span-2">
				<div class="bg-white rounded-lg shadow-lg p-6 sm:p-8">
					<div class="mb-8 flex items-center gap-4">
						<img src="/logo.png" alt="Savvy Logo" class="h-12 sm:h-16" />
						<h1 class="text-2xl font-bold text-gray-900">
							{tr('unsubscribe.title')}
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
							<h2 class="text-xl font-semibold text-gray-900 mb-2">
								{tr('unsubscribe.success')}
							</h2>
							<p class="text-gray-600 mb-6">
								{unsubType === 'reminders'
									? tr('unsubscribe.successMessageReminders')
									: tr('unsubscribe.successMessage')}
							</p>

							{#if $authStore.isAuthenticated}
								<a
									href={resolve('/notifications')}
									class="btn btn-primary w-full"
								>
									{tr('unsubscribe.goToSettings')}
								</a>
							{:else}
								<a href={resolve('/login')} class="btn btn-primary w-full">
									{tr('unsubscribe.goToLogin')}
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
							<h2 class="text-xl font-semibold text-gray-900 mb-2">
								{tr('unsubscribe.error')}
							</h2>
							<p class="text-gray-600 mb-6">{errorMessage}</p>

							{#if $authStore.isAuthenticated}
								<a
									href={resolve('/notifications')}
									class="btn btn-ghost w-full"
								>
									{tr('unsubscribe.goToSettings')}
								</a>
							{:else}
								<a href={resolve('/login')} class="btn btn-ghost w-full">
									{tr('unsubscribe.goToLogin')}
								</a>
							{/if}
						</div>
					{:else}
						<!-- idle state - no token provided -->
						<div class="text-center py-8">
							<p class="text-gray-600 mb-6">{tr('unsubscribe.invalidToken')}</p>
							<a href={resolve('/login')} class="btn btn-primary w-full">
								{tr('unsubscribe.goToLogin')}
							</a>
						</div>
					{/if}
				</div>
			</div>

			<!-- Right column: Information (1/3 width) -->
			<div class="lg:col-span-1">
				<div class="bg-white rounded-lg shadow-lg p-6">
					<h2 class="text-xl font-bold text-gray-900 mb-4">
						{unsubType === 'reminders'
							? tr('unsubscribe.infoTitleReminders')
							: tr('unsubscribe.infoTitleNotifications')}
					</h2>
					<p class="text-sm text-gray-600 mb-4">
						{unsubType === 'reminders'
							? tr('unsubscribe.infoDescReminders')
							: tr('unsubscribe.infoDescNotifications')}
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
									{tr('unsubscribe.info1Title')}
								</h3>
								<p class="text-xs text-gray-600 mt-1">
									{tr('unsubscribe.info1Desc')}
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
									{tr('unsubscribe.info2Title')}
								</h3>
								<p class="text-xs text-gray-600 mt-1">
									{tr('unsubscribe.info2Desc')}
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
									{tr('unsubscribe.info3Title')}
								</h3>
								<p class="text-xs text-gray-600 mt-1">
									{tr('unsubscribe.info3Desc')}
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
							<span class="text-xs">{tr('unsubscribe.securityNote')}</span>
						</div>
					</div>
				</div>
			</div>
		</div>
	</div>
</div>
