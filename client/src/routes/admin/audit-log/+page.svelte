<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { t } from '$lib/stores/i18n';
	import PageShell from '$lib/components/layout/PageShell.svelte';
	import { platform } from '$lib/utils/platform';
	import Section from './Section.svelte';
	import AdminTabs from '$lib/components/admin/AdminTabs.svelte';

	const IS_DESKTOP = platform === 'other';
	const IS_IOS = platform === 'ios';

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
	<title>{$t('admin.auditLog.title')} - {$t('common.appName')}</title>
</svelte:head>

{#snippet adminTabs()}
	<AdminTabs active="audit-log" />
{/snippet}

<PageShell
	title={$t('admin.auditLog.title')}
	eyebrow={$t('nav.admin')}
	actions={IS_DESKTOP ? adminTabs : undefined}
	mobileActions={!IS_IOS}
	onBack={IS_DESKTOP ? undefined : goBack}
>
	<Section />
</PageShell>
