<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { ICON_PLUS } from '$lib/icons';
	import { t } from '$lib/stores/i18n';
	import { isOnline } from '$lib/stores/offline';
	import PageShell from '$lib/components/layout/PageShell.svelte';
	import { platform } from '$lib/utils/platform';
	import Section from './Section.svelte';
	import AdminTabs from '$lib/components/admin/AdminTabs.svelte';

	const IS_DESKTOP = platform === 'other';
	const IS_IOS = platform === 'ios';

	// iOS puts create-merchant on the title row as a liquid-glass circle.
	const isOffline = $derived(!$isOnline);

	// Direct hits (bookmark, PWA start URL) have no history to go back to; the
	// chevron would be dead. Same guard as the other admin screens.
	function goBack() {
		if (history.length > 1) {
			history.back();
			return;
		}
		goto(resolve('/profile'));
	}
</script>

<svelte:head>
	<title>{$t('admin.merchants.title')} - {$t('common.appName')}</title>
</svelte:head>

{#snippet adminTabs()}
	<AdminTabs active="merchants" />
{/snippet}

{#snippet iosCreate()}
	<a
		href={resolve('/admin/merchants/new')}
		aria-label={$t('admin.merchants.createMerchant')}
		class="liquid-glass-surface inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-full text-purple-600 transition-colors active:text-purple-700 {isOffline
			? 'pointer-events-none opacity-50'
			: ''}"
	>
		<svg
			class="h-6.25 w-6.25"
			fill="none"
			stroke="currentColor"
			stroke-width="2.1"
			stroke-linecap="round"
			stroke-linejoin="round"
			viewBox="0 0 24 24"
			aria-hidden="true"
		>
			<path d={ICON_PLUS} />
		</svg>
	</a>
{/snippet}

<!-- The descriptive subtitle is gone: the small line above the title is the
     ADMIN kicker, same slot as KONTO on the settings pages. -->
<PageShell
	title={$t('admin.merchants.title')}
	eyebrow={$t('nav.admin')}
	actions={IS_DESKTOP ? adminTabs : IS_IOS ? iosCreate : undefined}
	mobileActions={!IS_IOS}
	onBack={IS_DESKTOP ? undefined : goBack}
>
	<Section />
</PageShell>
