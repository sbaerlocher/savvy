<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { ICON_PLUS } from '$lib/icons';
	import PageShell from '$lib/components/layout/PageShell.svelte';
	import { t } from '$lib/stores/i18n';
	import { isOnline } from '$lib/stores/offline';
	import { platform } from '$lib/utils/platform';
	import Section from './Section.svelte';
	import AdminTabs from '$lib/components/admin/AdminTabs.svelte';

	// `platform` is a module constant, so a plain const, not $derived.
	const IS_DESKTOP = platform === 'other';
	const IS_IOS = platform === 'ios';

	// iOS puts create-user on the title row as a liquid-glass circle; the
	// config gate comes up from the section, which already loads it.
	let localLoginEnabled = $state(true);
	const isOffline = $derived(!$isOnline);

	// Direct hits (bookmark, PWA start URL) have no history to go back to; the
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
	<title>{$t('admin.users.title')} - {$t('common.appName')}</title>
</svelte:head>

{#snippet adminTabs()}
	<AdminTabs active="users" />
{/snippet}

{#snippet iosCreate()}
	{#if localLoginEnabled}
		<a
			href={resolve('/admin/users/new')}
			aria-label={$t('admin.users.createUser')}
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
	{/if}
{/snippet}

<PageShell
	title={$t('admin.users.title')}
	eyebrow={$t('nav.admin')}
	actions={IS_DESKTOP ? adminTabs : IS_IOS ? iosCreate : undefined}
	mobileActions={!IS_IOS}
	onBack={IS_DESKTOP ? undefined : goBack}
>
	<Section bind:localLoginEnabled />
</PageShell>
