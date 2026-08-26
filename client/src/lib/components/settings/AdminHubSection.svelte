<script lang="ts">
	import { resolve } from '$app/paths';
	import {
		ICON_CART,
		ICON_CHART,
		ICON_CHEVRON_RIGHT,
		ICON_DOCUMENT,
		ICON_USERS
	} from '$lib/icons';
	import { t } from '$lib/stores/i18n';
	import { get } from 'svelte/store';

	const tr = (key: string) => get(t)(key);

	// The iOS mockup (screen-AdminIOS, frame "Admin-Einstieg · Profil") puts the
	// admin entry points into the profile as a grouped-inset row group, since the
	// three-tab bottom nav carries no admin tab. Every other platform reaches
	// admin through DesktopNav's dropdown, so this section is iOS-only and gated
	// on is_admin by the call site.
	const items = [
		{
			href: resolve('/admin/users'),
			icon: ICON_USERS,
			label: 'nav.adminUsers',
			sub: 'admin.hub.usersSub'
		},
		{
			href: resolve('/admin/merchants'),
			icon: ICON_CART,
			label: 'nav.adminMerchants',
			sub: 'admin.hub.merchantsSub'
		},
		{
			href: resolve('/admin/audit-log'),
			icon: ICON_DOCUMENT,
			label: 'nav.adminAuditLog',
			sub: 'admin.hub.auditLogSub'
		},
		{
			href: resolve('/admin/system-health'),
			icon: ICON_CHART,
			label: 'nav.adminSystemHealth',
			sub: 'admin.hub.systemHealthSub'
		}
	];
</script>

<section>
	<div class="flex items-center gap-1.75 px-1.5 pb-2">
		<span class="text-body-sm font-semibold text-purple-600 uppercase">
			{tr('nav.admin')}
		</span>
		<span
			class="inline-flex items-center rounded-full bg-purple-50 px-2 py-0.5 text-tag font-semibold tracking-normal text-purple-600"
		>
			{tr('admin.hub.adminsOnly')}
		</span>
	</div>

	<div class="overflow-hidden rounded-inset bg-surface">
		{#each items as item, i (item.href)}
			<!-- eslint-disable svelte/no-navigation-without-resolve -- item.href is produced by resolve() above -->
			<a
				href={item.href}
				class="flex items-center gap-3 px-3.75 py-3.25 transition-colors active:bg-surface-1 {i <
				items.length - 1
					? 'border-b border-border-soft'
					: ''}"
			>
				<span
					class="flex h-7.5 w-7.5 shrink-0 items-center justify-center rounded-lg bg-purple-50 text-purple-600"
				>
					<svg
						class="h-4 w-4"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
						stroke-linecap="round"
						stroke-linejoin="round"
						viewBox="0 0 24 24"
						aria-hidden="true"
					>
						<path d={item.icon} />
					</svg>
				</span>
				<span class="min-w-0 flex-1">
					<span class="block text-[length:var(--text-code)] text-text"
						>{tr(item.label)}</span
					>
					<span class="mt-px block text-chip font-normal text-text-subtle"
						>{tr(item.sub)}</span
					>
				</span>
				<svg
					class="h-3.5 w-2 shrink-0 text-text-mono-faint"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
					stroke-linecap="round"
					stroke-linejoin="round"
					viewBox="0 0 8 14"
					aria-hidden="true"
				>
					<path d={ICON_CHEVRON_RIGHT} />
				</svg>
			</a>
			<!-- eslint-enable svelte/no-navigation-without-resolve -->
		{/each}
	</div>
</section>
