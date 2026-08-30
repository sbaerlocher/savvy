<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import PageShell from '$lib/components/layout/PageShell.svelte';
	import { t } from '$lib/stores/i18n';
	import { platform } from '$lib/utils/platform';
	import Section from './Section.svelte';
	import AdminTabs from '$lib/components/admin/AdminTabs.svelte';

	// `platform` is a module constant, so a plain const, not $derived.
	const IS_DESKTOP = platform === 'other';
	const IS_IOS = platform === 'ios';

	// Direct hits (bookmark, alerting mail) have no history to go back to; the
	// chevron would be dead. Same guard as the other iOS screens.
	function goBack() {
		if (history.length > 1) {
			history.back();
			return;
		}
		goto(resolve('/profile'));
	}
</script>

<svelte:head>
	<title>{$t('nav.adminSystemHealth')} - {$t('common.appName')}</title>
</svelte:head>

{#snippet adminTabs()}
	<AdminTabs active="system-health" />
{/snippet}

<PageShell
	title={$t('nav.adminSystemHealth')}
	eyebrow={$t('nav.admin')}
	actions={IS_DESKTOP ? adminTabs : undefined}
	mobileActions={!IS_IOS}
	onBack={IS_DESKTOP ? undefined : goBack}
>
	<Section />
</PageShell>
