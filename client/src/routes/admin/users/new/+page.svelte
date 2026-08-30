<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { get } from 'svelte/store';
	import { t } from '$lib/stores/i18n';
	import PageShell from '$lib/components/layout/PageShell.svelte';
	import { platform } from '$lib/utils/platform';
	import Section from './Section.svelte';

	// Android uses the M3 back chevron on the title row; the other platforms
	// keep the text link in the eyebrow slot. `platform` is a module constant.
	const IS_ANDROID = platform === 'android';

	// Svelte 5 compatible translation wrapper
	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);
</script>

<svelte:head>
	<title>{tr('admin.users.createUser')} - {tr('common.appName')}</title>
</svelte:head>

{#snippet backLine()}
	<a
		href={resolve('/admin/users')}
		class="text-text-subtle hover:text-text-ink2"
		>{tr('common.backToOverview')}</a
	>
{/snippet}

<PageShell
	title={tr('admin.users.createUser')}
	mobileActions={false}
	back={IS_ANDROID ? undefined : backLine}
	onBack={IS_ANDROID ? () => goto(resolve('/admin/users')) : undefined}
>
	<Section />
</PageShell>
