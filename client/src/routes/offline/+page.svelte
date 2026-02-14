<script lang="ts">
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';

	let retrying = $state(false);
	let isOnline = $state(false);

	// Simple translations (avoiding dependency on $t for offline reliability)
	const text = {
		title: 'Du bist offline',
		titleOnline: 'Du bist wieder online!',
		message:
			'Diese Seite ist offline nicht verfügbar. Bitte überprüfe deine Internetverbindung.',
		messageOnline: "Gleich geht's weiter...",
		retry: 'Erneut versuchen',
		retrying: 'Verbinde...',
		goHome: 'Zur Startseite',
		offlineFeaturesTitle: 'Offline verfügbare Funktionen:',
		offlineFeatures: [
			'Bereits besuchte Seiten ansehen',
			'Gespeicherte Karten durchsuchen',
			'Gutscheine im Cache anzeigen',
			'Geschenkkarten abrufen'
		]
	};

	function handleRetry() {
		if (navigator.onLine) {
			window.location.href = '/';
		} else {
			retrying = true;
			setTimeout(() => {
				retrying = false;
			}, 2000);
		}
	}

	onMount(() => {
		// Check initial online status
		isOnline = navigator.onLine;

		// Auto-redirect when back online
		const handleOnline = () => {
			isOnline = true;
			setTimeout(() => {
				window.location.href = '/';
			}, 1000);
		};

		const handleOffline = () => {
			isOnline = false;
		};

		window.addEventListener('online', handleOnline);
		window.addEventListener('offline', handleOffline);

		return () => {
			window.removeEventListener('online', handleOnline);
			window.removeEventListener('offline', handleOffline);
		};
	});
</script>

<svelte:head>
	<title>{text.title} - Savvy</title>
</svelte:head>

<div
	class="min-h-screen bg-gradient-to-br from-gray-50 to-gray-100 flex items-center justify-center px-4"
>
	<div class="max-w-md w-full bg-white rounded-2xl shadow-xl p-8">
		{#if isOnline}
			<!-- Back Online -->
			<div class="text-center">
				<div
					class="w-20 h-20 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-4"
				>
					<svg
						class="w-10 h-10 text-green-600"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M5 13l4 4L19 7"
						></path>
					</svg>
				</div>
				<h1 class="text-2xl font-bold text-gray-900 mb-2">
					{text.titleOnline}
				</h1>
				<p class="text-gray-600">
					{text.messageOnline}
				</p>
			</div>
		{:else}
			<!-- Offline -->
			<div class="text-center mb-6">
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
					{text.title}
				</h1>
				<p class="text-gray-600 mb-6">
					{text.message}
				</p>
			</div>

			<!-- Info Box -->
			<div
				class="bg-cyan-50 border border-cyan-200 rounded-lg p-4 mb-6 text-left"
			>
				<div class="flex items-start gap-3">
					<svg
						class="w-5 h-5 text-cyan-600 flex-shrink-0 mt-0.5"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
						></path>
					</svg>
					<div class="text-sm text-cyan-800">
						<p class="font-semibold mb-1">{text.offlineFeaturesTitle}</p>
						<ul class="list-disc list-inside space-y-1">
							{#each text.offlineFeatures as feature}
								<li>{feature}</li>
							{/each}
						</ul>
					</div>
				</div>
			</div>

			<!-- Actions -->
			<div class="space-y-3">
				<button
					type="button"
					onclick={handleRetry}
					disabled={retrying}
					class="w-full inline-flex justify-center items-center px-6 py-3 border border-transparent text-base font-medium rounded-lg shadow-sm text-white bg-cyan-600 hover:bg-cyan-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-cyan-500 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
				>
					{#if retrying}
						<svg
							class="animate-spin -ml-1 mr-3 h-5 w-5 text-white"
							xmlns="http://www.w3.org/2000/svg"
							fill="none"
							viewBox="0 0 24 24"
						>
							<circle
								class="opacity-25"
								cx="12"
								cy="12"
								r="10"
								stroke="currentColor"
								stroke-width="4"
							/>
							<path
								class="opacity-75"
								fill="currentColor"
								d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
							/>
						</svg>
						{text.retrying}
					{:else}
						{text.retry}
					{/if}
				</button>
				<a
					href="/"
					class="block w-full text-center px-6 py-3 border border-gray-300 text-base font-medium rounded-lg text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-cyan-500 transition-colors"
				>
					{text.goHome}
				</a>
			</div>
		{/if}
	</div>
</div>
