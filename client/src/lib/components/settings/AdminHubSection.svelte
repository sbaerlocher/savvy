<script lang="ts">
	import { resolve } from '$app/paths';
	import {
		ICON_CART,
		ICON_CHART,
		ICON_CHEVRON_RIGHT,
		ICON_DOCUMENT,
		ICON_USERS
	} from '$lib/icons';
	import type { ProfileDTO } from '$lib/api';
	import { t } from '$lib/stores/i18n';
	import { platform } from '$lib/utils/platform';
	import { get } from 'svelte/store';

	// The Android caller passes the signed-in profile so the section can lead
	// with the identity card from its mockup; iOS renders the row group alone.
	let { profile }: { profile?: ProfileDTO } = $props();

	const tr = (key: string) => get(t)(key);

	// Both native mockups (screen-AdminIOS and screen-AdminAndroid, frame
	// "Admin-Einstieg · Profil") put the admin entry points into the profile,
	// since neither three-tab bottom nav carries an admin tab. They differ in
	// chrome: iOS is a grouped-inset row group, Android a tonal M3 card led by
	// the identity row. Desktop reaches admin through DesktopNav's dropdown, so
	// this section is native-only and gated on is_admin by the call site.
	// `platform` is a module constant, so a plain const.
	const IS_ANDROID = platform === 'android';

	const fullName = $derived(
		[profile?.first_name, profile?.last_name].filter(Boolean).join(' ')
	);
	const initials = $derived(
		(fullName || profile?.email || '')
			.split(/[\s@.]+/)
			.filter(Boolean)
			.slice(0, 2)
			.map((part) => part[0]?.toUpperCase() ?? '')
			.join('')
	);
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
	<!-- Android (mockup screen-AdminAndroid, frame "Admin-Einstieg · Profil"):
	     the identity card with its role badge, then the destinations as one
	     tonal M3 card. -->
	{#if profile}
		<div
			class="rounded-m3-lg bg-m3-card border-border flex items-center gap-3.5 border px-4 py-3.25"
		>
			<span
				class="bg-accent-100 text-accent-850 rounded-m3-full text-amount flex h-11 w-11 shrink-0 items-center justify-center font-semibold"
			>
				{initials}
			</span>
			<div class="min-w-0 flex-1">
				<div class="text-amount text-text truncate">
					{fullName || profile.email}
				</div>
				<div class="text-label text-text-muted mt-px truncate">
					{profile.email}
				</div>
			</div>
			<span
				class="bg-danger-100 text-danger-800 rounded-m3-full text-eyebrow inline-flex shrink-0 items-center px-2.25 py-0.75 font-semibold"
			>
				{tr('admin.users.roleAdmin')}
			</span>
		</div>
	{/if}

	<div class="flex items-center gap-2 px-2 pt-5 pb-2">
		<span class="text-label text-purple-600 font-semibold"
			>{tr('nav.admin')}</span
		>
		<span
			class="bg-purple-50 text-purple-600 rounded-m3-full text-tag inline-flex items-center px-2.25 py-0.5 font-semibold"
		>
			{tr('admin.hub.adminsOnly')}
		</span>
	</div>

	<div class="rounded-m3-lg bg-m3-card border-border overflow-hidden border">
		{#each items as item, i (item.href)}
			<!-- eslint-disable svelte/no-navigation-without-resolve -- item.href is produced by resolve() above -->
			<a
				href={item.href}
				class="hover:bg-ground-active flex items-center gap-3.5 px-4 py-3 transition-colors {i >
				0
					? 'border-border-soft border-t'
					: ''}"
			>
				<span
					class="bg-purple-50 text-purple-600 rounded-m3-full flex h-11 w-11 shrink-0 items-center justify-center"
				>
					<svg
						class="h-5 w-5"
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
				<div class="min-w-0 flex-1">
					<div class="text-subheading text-text">{tr(item.label)}</div>
					<div class="text-label text-text-muted mt-px">{tr(item.sub)}</div>
				</div>
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
	</div>
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
