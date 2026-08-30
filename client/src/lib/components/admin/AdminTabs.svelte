<script lang="ts">
	import { resolve } from '$app/paths';
	import { configStore } from '$lib/stores/config';
	import { t } from '$lib/stores/i18n';

	export type AdminTab =
		'users' | 'merchants' | 'audit-log' | 'system-health' | 'email-templates';

	interface Props {
		active: AdminTab;
	}

	let { active }: Props = $props();

	// Literal path type keeps SvelteKit's typed `resolve()` satisfied.
	interface Tab {
		id: AdminTab;
		path:
			| '/admin/users'
			| '/admin/merchants'
			| '/admin/audit-log'
			| '/admin/system-health'
			| '/admin/email-templates';
		labelKey: string;
		icon: string;
	}

	// Icon paths match the admin dropdown in DesktopNav, so the two admin
	// navigations stay recognisably the same set.
	const BASE_TABS: Tab[] = [
		{
			id: 'users',
			path: '/admin/users',
			labelKey: 'nav.adminUsers',
			icon: 'M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z'
		},
		{
			id: 'merchants',
			path: '/admin/merchants',
			labelKey: 'nav.adminMerchants',
			icon: 'M3 3h2l.4 2M7 13h10l4-8H5.4M7 13L5.4 5M7 13l-2.293 2.293c-.63.63-.184 1.707.707 1.707H17m0 0a2 2 0 100 4 2 2 0 000-4zm-8 2a2 2 0 11-4 0 2 2 0 014 0z'
		},
		{
			id: 'audit-log',
			path: '/admin/audit-log',
			labelKey: 'nav.adminAuditLog',
			icon: 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01'
		},
		{
			id: 'system-health',
			path: '/admin/system-health',
			labelKey: 'nav.adminSystemHealth',
			icon: 'M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z'
		}
	];
	// Development-only route (the handler is gated the same way).
	const DEV_TAB: Tab = {
		id: 'email-templates',
		path: '/admin/email-templates',
		labelKey: 'nav.adminEmailTemplates',
		icon: 'M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z'
	};
	const tabs = $derived(
		$configStore.is_development ? [...BASE_TABS, DEV_TAB] : BASE_TABS
	);
</script>

<!-- Separate title-row actions, same chrome as the settings tabs and the
     wallet toolbar; the active tab carries the shared accent-ring state. -->
<div class="flex items-center gap-2.5">
	{#each tabs as tab (tab.id)}
		<a
			href={resolve(tab.path)}
			aria-current={active === tab.id ? 'page' : undefined}
			class="title-action whitespace-nowrap {active === tab.id
				? 'ring-2 ring-accent border-accent font-semibold text-accent-hover'
				: 'text-text-muted'}"
		>
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
				<path d={tab.icon} />
			</svg>
			{$t(tab.labelKey)}
		</a>
	{/each}
</div>
