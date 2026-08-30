<script lang="ts">
	import { resolve } from '$app/paths';
	import { t } from '$lib/stores/i18n';

	export type SettingsTab = 'profile' | 'security' | 'notifications';

	interface Props {
		active: SettingsTab;
	}

	let { active }: Props = $props();

	// Each tab keeps its own route so mobile deep links and the existing
	// /security and /notifications/settings pages stay reachable; on desktop
	// they render as this same shell with a different tab preselected.
	const tabs = [
		{ id: 'profile', path: '/profile', labelKey: 'nav.profile' },
		{ id: 'security', path: '/security', labelKey: 'nav.security' },
		{
			id: 'notifications',
			path: '/notifications/settings',
			labelKey: 'settings.sections.notifications'
		}
	] as const;
</script>

<!-- Three separate title-row actions, same chrome as the wallet toolbar and
     the dashboard tiles; the active tab carries the shared accent-ring state. -->
<div class="flex items-center gap-2.5">
	{#each tabs as tab (tab.id)}
		<a
			href={resolve(tab.path)}
			aria-current={active === tab.id ? 'page' : undefined}
			class="title-action whitespace-nowrap {active === tab.id
				? 'ring-2 ring-accent border-accent font-semibold text-accent-hover'
				: 'text-text-muted'}"
		>
			{#if tab.id === 'profile'}
				<svg
					class="h-4 w-4"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
					stroke-linecap="round"
					stroke-linejoin="round"
					aria-hidden="true"
				>
					<circle cx="12" cy="8" r="4" />
					<path d="M4 21a8 8 0 0116 0" />
				</svg>
			{:else if tab.id === 'security'}
				<svg
					class="h-4 w-4"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
					stroke-linecap="round"
					stroke-linejoin="round"
					aria-hidden="true"
				>
					<rect x="4" y="10" width="16" height="11" rx="2.5" />
					<path d="M8 10V7a4 4 0 018 0v3" />
				</svg>
			{:else}
				<svg
					class="h-4 w-4"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
					stroke-linecap="round"
					stroke-linejoin="round"
					aria-hidden="true"
				>
					<path d="M18 8a6 6 0 10-12 0c0 7-3 9-3 9h18s-3-2-3-9" />
					<path d="M13.5 21a1.7 1.7 0 01-3 0" />
				</svg>
			{/if}
			{$t(tab.labelKey)}
		</a>
	{/each}
</div>
