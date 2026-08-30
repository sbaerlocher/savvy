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

	function handleCancel() {
		goto(resolve('/admin/users'));
	}
</script>

<svelte:head>
	<title>{$t('admin.users.editUser')} - {$t('common.appName')}</title>
</svelte:head>

{#snippet backLine()}
	<button onclick={handleCancel} class="text-text-subtle hover:text-text-ink2">
		{$t('common.backToOverview')}
	</button>
{/snippet}

<PageShell
	title={$t('admin.users.editUser')}
	mobileActions={false}
	back={IS_ANDROID ? undefined : backLine}
	onBack={IS_ANDROID ? handleCancel : undefined}
>
	<Section />
</PageShell>
