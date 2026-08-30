<script lang="ts">
	import PageShell from '$lib/components/layout/PageShell.svelte';
	import { t } from '$lib/stores/i18n';
	import { platform } from '$lib/utils/platform';
	import { get } from 'svelte/store';
	import Section from './Section.svelte';
	import SettingsTabs from '$lib/components/settings/SettingsTabs.svelte';

	// `platform` is a module constant, so a plain const, not $derived. Desktop
	// renders this route as the security tab of the merged settings page.
	const IS_DESKTOP = platform === 'other';

	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);
</script>

<svelte:head>
	<title>{tr('nav.security')} - {tr('common.appName')}</title>
</svelte:head>

{#snippet settingsTabs()}
	<SettingsTabs active="security" />
{/snippet}

<PageShell
	title={IS_DESKTOP ? tr('settings.title') : tr('nav.security')}
	eyebrow={IS_DESKTOP ? tr('settings.sections.account') : undefined}
	actions={IS_DESKTOP ? settingsTabs : undefined}
>
	<Section />
</PageShell>
