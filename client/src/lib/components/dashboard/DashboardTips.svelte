<script lang="ts">
	import { t } from '$lib/stores/i18n';
</script>

<div class="bg-white rounded-lg shadow-md p-6 hidden lg:block">
	<h2 class="text-lg font-semibold text-gray-900 mb-3">
		{$t('dashboard.tipsTitle')}
	</h2>
	<div class="space-y-3">
		<div class="flex items-start gap-2">
			<svg
				class="w-5 h-5 text-cyan-500 mt-0.5 flex-shrink-0"
				fill="none"
				stroke="currentColor"
				viewBox="0 0 24 24"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
				/>
			</svg>
			<p class="text-xs text-gray-600">{$t('dashboard.tip1')}</p>
		</div>
		<div class="flex items-start gap-2">
			<svg
				class="w-5 h-5 text-green-500 mt-0.5 flex-shrink-0"
				fill="none"
				stroke="currentColor"
				viewBox="0 0 24 24"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m0 2.684l6.632 3.316m-6.632-6l6.632-3.316m0 0a3 3 0 105.367-2.684 3 3 0 00-5.367 2.684zm0 9.316a3 3 0 105.368 2.684 3 3 0 00-5.368-2.684z"
				/>
			</svg>
			<p class="text-xs text-gray-600">{$t('dashboard.tip2')}</p>
		</div>
		<div class="flex items-start gap-2">
			<svg
				class="w-5 h-5 text-yellow-500 mt-0.5 flex-shrink-0"
				fill="none"
				stroke="currentColor"
				viewBox="0 0 24 24"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d="M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z"
				/>
			</svg>
			<p class="text-xs text-gray-600">{$t('dashboard.tip3')}</p>
		</div>
	</div>
</div>
