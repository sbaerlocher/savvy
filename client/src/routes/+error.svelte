<script lang="ts">
	import { page } from '$app/stores';
	import { browser } from '$app/environment';
	import { resolve } from '$app/paths';
	import { t } from '$lib/stores/i18n';

	const isOffline = $derived(browser && !navigator.onLine);
	const status = $derived($page.status);
	const message = $derived($page.error?.message || '');

	function handleRetry() {
		if (browser) {
			window.location.reload();
		}
	}

	function handleGoBack() {
		if (browser) {
			if (window.history.length > 1) {
				window.history.back();
			} else {
				window.location.href = '/dashboard';
			}
		}
	}
</script>

<div
	class="min-h-screen bg-gradient-to-br from-gray-50 to-gray-100 flex items-center justify-center px-4"
>
	<div class="max-w-md w-full bg-white rounded-2xl shadow-xl p-8">
		<div class="text-center">
			{#if isOffline}
				<!-- Offline Error -->
				<div
					class="w-20 h-20 bg-orange-100 rounded-full flex items-center justify-center mx-auto mb-4"
				>
					<svg
						class="w-10 h-10 text-orange-600"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M18.364 5.636a9 9 0 010 12.728m0 0l-2.829-2.829m2.829 2.829L21 21M15.536 8.464a5 5 0 010 7.072m0 0l-2.829-2.829m-4.243 2.829a4.978 4.978 0 01-1.414-2.83m-1.414 5.658a9 9 0 01-2.167-9.238m7.824 2.167a1 1 0 111.414 1.414m-1.414-1.414L3 3m8.293 8.293l1.414 1.414"
						></path>
					</svg>
				</div>
				<h1 class="text-2xl font-bold text-gray-900 mb-2">
					{$t('errors.offline.title')}
				</h1>
				<p class="text-gray-600 mb-2">
					{$t('errors.offline.description')}
				</p>
				<p class="text-gray-500 text-sm mb-6">
					{$t('errors.offline.hint')}
				</p>
			{:else if status === 404}
				<!-- Not Found -->
				<div
					class="w-20 h-20 bg-cyan-100 rounded-full flex items-center justify-center mx-auto mb-4"
				>
					<svg
						class="w-10 h-10 text-cyan-600"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
						></path>
					</svg>
				</div>
				<h1 class="text-2xl font-bold text-gray-900 mb-2">
					{$t('errors.notFound.title')}
				</h1>
				<p class="text-gray-600 mb-6">
					{message || $t('errors.notFound.description')}
				</p>
			{:else}
				<!-- Generic Error (500, etc.) -->
				<div
					class="w-20 h-20 bg-red-100 rounded-full flex items-center justify-center mx-auto mb-4"
				>
					<svg
						class="w-10 h-10 text-red-600"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
						></path>
					</svg>
				</div>
				<h1 class="text-2xl font-bold text-gray-900 mb-2">
					{$t('errors.generic.title')}
				</h1>
				<p class="text-gray-600 mb-6">
					{message || $t('errors.generic.description')}
				</p>
			{/if}

			<!-- Actions -->
			<div class="space-y-3">
				{#if isOffline}
					<button
						type="button"
						onclick={handleGoBack}
						class="w-full inline-flex justify-center items-center px-6 py-3 border border-transparent text-base font-medium rounded-lg shadow-sm text-white bg-cyan-600 hover:bg-cyan-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-cyan-500 transition-colors"
					>
						{$t('errors.goBack')}
					</button>
				{:else}
					<button
						type="button"
						onclick={handleRetry}
						class="w-full inline-flex justify-center items-center px-6 py-3 border border-transparent text-base font-medium rounded-lg shadow-sm text-white bg-cyan-600 hover:bg-cyan-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-cyan-500 transition-colors"
					>
						{$t('errors.retry')}
					</button>
				{/if}
				<a
					href={resolve('/dashboard')}
					class="block w-full text-center px-6 py-3 border border-gray-300 text-base font-medium rounded-lg text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-cyan-500 transition-colors"
				>
					{$t('errors.toDashboard')}
				</a>
			</div>
		</div>
	</div>
</div>
