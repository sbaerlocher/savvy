<script lang="ts">
	import { t } from '$lib/stores/i18n';
	import { resolve } from '$app/paths';
	import { categoryColors } from '$lib/utils/category-colors';

	let { isOffline }: { isOffline: boolean } = $props();
</script>

<div class="bg-white rounded-lg shadow-md p-6">
	<h2 class="text-xl font-semibold text-gray-900 mb-3">
		{$t('dashboard.quickActions')}
	</h2>
	<div class="flex flex-col gap-2">
		<a
			href={resolve('/cards/new')}
			onclick={(e) => {
				if (isOffline) e.preventDefault();
			}}
			class="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-medium transition {isOffline
				? 'opacity-50 cursor-not-allowed bg-gray-50 text-gray-400'
				: categoryColors.cards.action}"
		>
			<span>+</span>
			{$t('dashboard.addCard')}
		</a>
		<a
			href={resolve('/vouchers/new')}
			onclick={(e) => {
				if (isOffline) e.preventDefault();
			}}
			class="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-medium transition {isOffline
				? 'opacity-50 cursor-not-allowed bg-gray-50 text-gray-400'
				: categoryColors.vouchers.action}"
		>
			<span>+</span>
			{$t('dashboard.addVoucher')}
		</a>
		<a
			href={resolve('/gift-cards/new')}
			onclick={(e) => {
				if (isOffline) e.preventDefault();
			}}
			class="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-medium transition {isOffline
				? 'opacity-50 cursor-not-allowed bg-gray-50 text-gray-400'
				: categoryColors.giftCards.action}"
		>
			<span>+</span>
			{$t('dashboard.addGiftCard')}
		</a>
	</div>
</div>
