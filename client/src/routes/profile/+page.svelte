<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import PageShell from '$lib/components/layout/PageShell.svelte';
	import SettingsTabs from '$lib/components/settings/SettingsTabs.svelte';
	import AndroidSettingsScreen from '$lib/components/settings/AndroidSettingsScreen.svelte';
	import { t } from '$lib/stores/i18n';
	import { platform } from '$lib/utils/platform';
	import { get } from 'svelte/store';
	import Section from './Section.svelte';

	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);

	// iOS renders the whole settings screen from screen-SettingsIOS (profile,
	// security, sessions and notifications as grouped-inset sections) instead of
	// the two-column card layout. `platform` is a module constant, so this is a
	// plain const, not $derived.
	const IS_IOS = platform === 'ios';
	// Desktop merges profile, security and notification preferences into one
	// tabbed settings page (screen-SettingsDesktop); the routes stay in place
	// and each renders the shell with its own tab preselected.
	const IS_DESKTOP = platform === 'other';

	// Android back chevron on the title row; guarded like the other screens so
	// a deep link (PWA start URL) does not leave the chevron dead.
	function goBack() {
		if (history.length > 1) history.back();
		else goto(resolve('/dashboard'));
	}
</script>

<svelte:head>
	<title>{tr('profile.title')} - {tr('common.appName')}</title>
</svelte:head>

{#if IS_IOS}
	<!-- No title: IOSSettingsScreen renders the screen's own <h1>. -->
	<PageShell>
		<Section />
	</PageShell>
{:else if IS_DESKTOP}
	<!-- The settings tabs sit on the title row like every other page's
	     right-hand actions. -->
	<PageShell
		title={tr('settings.title')}
		eyebrow={tr('settings.sections.account')}
	>
		{#snippet actions()}
			<SettingsTabs active="profile" />
		{/snippet}
		<Section />
	</PageShell>
{:else}
	<!-- Android: the M3 settings screen (mockup screen-SettingsAndroid) is the
	     account destination; `width="full"` because its list rows carry their
	     own inset and the hand-built top app bar replaces the shell title. -->
	<PageShell
		width="full"
		title={tr('settings.title')}
		mobileActions={false}
		onBack={goBack}
	>
		<AndroidSettingsScreen />
	</PageShell>
{/if}
