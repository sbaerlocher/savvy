<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { t } from '$lib/stores/i18n';
	import PageShell from '$lib/components/layout/PageShell.svelte';
	import { platform } from '$lib/utils/platform';
	import Section from './Section.svelte';

	// Android uses the M3 back chevron on the title row; the other platforms
	// keep the text link in the eyebrow slot. `platform` is a module constant.
	const IS_ANDROID = platform === 'android';
</script>

<svelte:head>
	<title>{$t('admin.merchants.createMerchant')} - {$t('common.appName')}</title>
</svelte:head>

{#snippet backLine()}
	<a
		href={resolve('/admin/merchants')}
		class="text-text-subtle hover:text-text-ink2 transition-colors"
	>
		{$t('common.backToOverview')}
	</a>
{/snippet}

<PageShell
	title={$t('admin.merchants.createMerchant')}
	mobileActions={false}
	back={IS_ANDROID ? undefined : backLine}
	onBack={IS_ANDROID ? () => goto(resolve('/admin/merchants')) : undefined}
>
	<Section />
</PageShell>
