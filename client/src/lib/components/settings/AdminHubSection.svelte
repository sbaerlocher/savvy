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
	import { platform } from '$lib/utils/platform';
	import { get } from 'svelte/store';

	const tr = (key: string) => get(t)(key);

	// Both native mockups (screen-AdminIOS and screen-AdminAndroid, frame
	// "Admin-Einstieg · Profil") put the admin entry points into the profile,
	// since neither three-tab bottom nav carries an admin tab. They differ in
	// chrome: iOS is a grouped-inset row group, Android a tonal M3 card led by
	// the identity row. Desktop reaches admin through DesktopNav's link, so
	// this section is native-only and gated on is_admin by the call site.
	// `platform` is a module constant, so a plain const.
	const IS_ANDROID = platform === 'android';

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

{#if IS_ANDROID}
	<!-- Android: the destinations as M3 list rows on the screen background —
	     same row anatomy as the M3SettingsRow sections below (40px tonal
	     circle, title over subtitle), so the admin hub reads as part of the
	     settings list, not a card. No identity card: the Profil section below
	     already carries name and email. -->
	<div class="flex items-center gap-2 px-6 pt-2.5 pb-2">
		<span class="text-label text-purple-600 font-semibold"
			>{tr('nav.admin')}</span
		>
		<span
			class="bg-purple-50 text-purple-600 rounded-m3-full text-tag inline-flex items-center px-2.25 py-0.5 font-semibold"
		>
			{tr('admin.hub.adminsOnly')}
		</span>
	</div>

	{#each items as item (item.href)}
		<!-- eslint-disable svelte/no-navigation-without-resolve -- item.href is produced by resolve() above -->
		<a
			href={item.href}
			class="active:bg-ground-active flex w-full items-center gap-4 px-6 py-3"
		>
			<span
				class="bg-purple-50 text-purple-600 rounded-m3-full flex h-10 w-10 shrink-0 items-center justify-center"
			>
				<svg
					class="h-4.75 w-4.75"
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
				<span class="text-subheading text-text block font-normal"
					>{tr(item.label)}</span
				>
				<span class="text-label text-text-muted mt-px block font-normal"
					>{tr(item.sub)}</span
				>
			</span>
			<svg
				class="text-text-subtle h-5 w-5 shrink-0"
				fill="none"
				stroke="currentColor"
				stroke-width="2"
				stroke-linecap="round"
				stroke-linejoin="round"
				viewBox="0 0 24 24"
				aria-hidden="true"
			>
				<path d="M9 5l7 7-7 7" />
			</svg>
		</a>
		<!-- eslint-enable svelte/no-navigation-without-resolve -->
	{/each}
{:else}
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
{/if}
