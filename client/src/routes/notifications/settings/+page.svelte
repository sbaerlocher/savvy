<script lang="ts">
	import PageShell from '$lib/components/layout/PageShell.svelte';
	import { t } from '$lib/stores/i18n';
	import { platform } from '$lib/utils/platform';
	import { get } from 'svelte/store';
	import Section from './Section.svelte';
	import SettingsTabs from '$lib/components/settings/SettingsTabs.svelte';

	// `platform` is a module constant, so a plain const, not $derived. Desktop
	// renders this route as the notifications tab of the merged settings page.
	const IS_DESKTOP = platform === 'other';

	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);
</script>

<svelte:head>
	<title>{tr('settings.notifications.title')} - {tr('common.appName')}</title>
</svelte:head>

<!-- Full shell width like every other title row; the settings column width is
     the section's own wrapper, not the shell's. -->
{#snippet settingsTabs()}
	<SettingsTabs active="notifications" />
{/snippet}

<PageShell
	title={IS_DESKTOP ? tr('settings.title') : tr('settings.notifications.title')}
	eyebrow={IS_DESKTOP ? tr('settings.sections.account') : undefined}
	actions={IS_DESKTOP ? settingsTabs : undefined}
>
	<Section />
</PageShell>
