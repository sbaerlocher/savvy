<script lang="ts">
	import { page } from '$app/stores';
	import { resolve } from '$app/paths';
	import { goto } from '$app/navigation';
	import NotificationPanel from '$lib/components/NotificationPanel.svelte';
	import { ICON_SEARCH } from '$lib/icons';
	import { authStore } from '$lib/stores/auth';
	import { t } from '$lib/stores/i18n';
	import { toastStore } from '$lib/stores/toast';
	import { showOfflineBanner } from '$lib/stores/offline';
	import { showNewDialog } from '$lib/stores/newDialog';
	import { logger } from '$lib/utils/logger';
	import type { UserDTO } from '$lib/types/api';

	const navLogger = logger.child('DesktopNav');

	// Server-provided user (data.user) — carries is_admin for the admin-link gate.
	let { user }: { user?: UserDTO } = $props();

	let desktopSearch = $state('');

	async function handleLogout() {
		await authStore.logout();
		window.location.href = '/login';
	}

	// Desktop search submits into Wallet (single global search entry).
	function submitDesktopSearch(e: SubmitEvent) {
		e.preventDefault();
		const q = desktopSearch.trim();
		const query = q ? `?search=${encodeURIComponent(q)}` : '?search=1';
		/* eslint-disable-next-line svelte/no-navigation-without-resolve -- base is resolve()d; the rest is a query string */
		goto(`${resolve('/wallet')}${query}`);
	}

	async function handleStopImpersonation() {
		navLogger.debug('Stop impersonation button clicked');
		try {
			await authStore.stopImpersonation();
			// Redirect happens in authStore.stopImpersonation
		} catch (err: unknown) {
			navLogger.error('Failed to stop impersonation', { error: err });
			const message =
				err instanceof Error
					? err.message
					: $t('admin.impersonate_info.stop_failed');
			toastStore.error(message || $t('admin.impersonate_info.stop_failed'));
		}
	}
</script>

<nav
	data-testid="desktop-nav"
	class="hidden transition-all duration-300 ease-out sm:block sm:border-b sm:border-border-soft sm:bg-surface"
	class:mt-16={$showOfflineBanner}
	class:sm:mt-12={$showOfflineBanner}
>
	<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
		<div class="flex justify-between h-16">
			<div class="flex items-center">
				<div class="flex-shrink-0">
					<a href={resolve('/dashboard')} class="flex items-center gap-2.5">
						<!-- The real logo next to the wordmark (the accent initial tile
						     was a placeholder). -->
						<img
							src="/logo.png"
							alt=""
							aria-hidden="true"
							class="hidden h-7.5 w-7.5 shrink-0 sm:block"
						/>
						<span
							class="hidden sm:inline text-heading font-bold tracking-tight text-text"
							>{$t('common.appName')}</span
						>
					</a>
				</div>

				<!-- App shell: three places — Start · Wallet · Profile. Desktop
				     mockup: 14px labels, active is weight-600 ink with a 2px accent
				     underline sitting on the bar's bottom edge. -->
				<div class="hidden sm:ml-6 sm:flex sm:gap-1 sm:self-stretch">
					<a
						href={resolve('/dashboard')}
						data-testid="nav-start-desktop"
						class="relative inline-flex items-center px-3.5 text-sm transition-colors hover:text-text-ink2"
						class:font-semibold={$page.url.pathname.startsWith('/dashboard')}
						class:font-medium={!$page.url.pathname.startsWith('/dashboard')}
						class:text-text={$page.url.pathname.startsWith('/dashboard')}
						class:text-text-subtle={!$page.url.pathname.startsWith(
							'/dashboard'
						)}
					>
						{$t('nav.start')}
						{#if $page.url.pathname.startsWith('/dashboard')}
							<span
								class="absolute inset-x-3.5 -bottom-px h-0.5 rounded-full bg-accent"
							></span>
						{/if}
					</a>
					<a
						href={resolve('/wallet')}
						data-testid="nav-wallet-desktop"
						class="relative inline-flex items-center px-3.5 text-sm transition-colors hover:text-text-ink2"
						class:font-semibold={$page.url.pathname.startsWith('/wallet')}
						class:font-medium={!$page.url.pathname.startsWith('/wallet')}
						class:text-text={$page.url.pathname.startsWith('/wallet')}
						class:text-text-subtle={!$page.url.pathname.startsWith('/wallet')}
					>
						{$t('nav.wallet')}
						{#if $page.url.pathname.startsWith('/wallet')}
							<span
								class="absolute inset-x-3.5 -bottom-px h-0.5 rounded-full bg-accent"
							></span>
						{/if}
					</a>
					<a
						href={resolve('/profile')}
						data-testid="nav-profile-desktop"
						class="relative inline-flex items-center px-3.5 text-sm transition-colors hover:text-text-ink2"
						class:font-semibold={$page.url.pathname.startsWith('/profile')}
						class:font-medium={!$page.url.pathname.startsWith('/profile')}
						class:text-text={$page.url.pathname.startsWith('/profile')}
						class:text-text-subtle={!$page.url.pathname.startsWith('/profile')}
					>
						{$t('nav.profile')}
						{#if $page.url.pathname.startsWith('/profile')}
							<span
								class="absolute inset-x-3.5 -bottom-px h-0.5 rounded-full bg-accent"
							></span>
						{/if}
					</a>
				</div>
			</div>

			<div class="flex items-center space-x-4">
				<!-- Desktop: inline search + New (single global entries). Mockup:
				     filled field on surface-2 with a leading magnifier, not a plain
				     bordered input. -->
				<form
					onsubmit={submitDesktopSearch}
					class="hidden sm:flex h-10 w-64 items-center gap-2 rounded-lg border border-border bg-surface-2 px-3 focus-within:border-accent focus-within:ring-2 focus-within:ring-accent/20"
					role="search"
				>
					<svg
						class="h-4 w-4 shrink-0 text-text-faint"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
						viewBox="0 0 24 24"
						aria-hidden="true"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							d={ICON_SEARCH}
						/>
					</svg>
					<input
						type="search"
						bind:value={desktopSearch}
						placeholder={$t('common.search')}
						aria-label={$t('common.search')}
						class="min-w-0 flex-1 bg-transparent text-sm text-text placeholder:text-text-placeholder focus:outline-none"
					/>
				</form>
				<button
					type="button"
					onclick={() => ($showNewDialog = true)}
					data-testid="nav-new-desktop"
					class="hidden sm:inline-flex h-10 items-center gap-1.5 rounded-lg bg-accent px-3.5 text-label text-white shadow-sm transition-colors hover:bg-accent-hover"
				>
					<svg
						class="w-4 h-4"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M12 4v16m8-8H4"
						/>
					</svg>
					{$t('common.new')}
				</button>

				<!-- Impersonation Indicator -->
				{#if $authStore.user?.is_impersonating}
					<div
						class="hidden sm:flex items-center bg-warning-50 border border-warning-300 rounded-lg px-3 py-2 gap-2"
					>
						<!-- Admin Icon -->
						<svg
							class="w-5 h-5 text-warning-700"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"
							></path>
						</svg>

						<!-- Arrow -->
						<svg
							class="w-4 h-4 text-warning-600"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M13 7l5 5m0 0l-5 5m5-5H6"
							></path>
						</svg>

						<!-- User Icon + Email -->
						<div class="flex items-center gap-1.5">
							<svg
								class="w-5 h-5 text-warning-700"
								fill="none"
								stroke="currentColor"
								viewBox="0 0 24 24"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"
								></path>
							</svg>
							<span class="text-sm text-warning-800 font-medium">
								{$authStore.user.email}
							</span>
						</div>

						<!-- Stop Button -->
						<button
							type="button"
							onclick={(e) => {
								navLogger.debug('Button onclick handler triggered');
								e.stopPropagation();
								navLogger.debug('About to call handleStopImpersonation');
								handleStopImpersonation();
							}}
							class="ml-2 text-xs bg-warning-600 hover:bg-warning-700 text-white px-3 py-1 rounded font-medium transition-colors"
						>
							{$t('admin.stop_impersonating')}
						</button>
					</div>
				{/if}

				{#if !$authStore.user?.is_impersonating}
					<!-- Admin entry (admins only, desktop only): a straight link into
					     the admin area — the pages themselves carry the AdminTabs
					     navigation, so the old dropdown is gone. -->
					{#if user?.is_admin}
						<a
							href={resolve('/admin/users')}
							class="hidden h-9.5 w-9.5 items-center justify-center rounded-sm bg-purple-50 text-purple-600 transition-colors hover:text-purple-700 sm:inline-flex"
							title={$t('nav.admin')}
							aria-label={$t('nav.admin')}
						>
							<svg
								class="h-5.5 w-5.5"
								fill="none"
								stroke="currentColor"
								viewBox="0 0 24 24"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"
								></path>
							</svg>
						</a>
					{/if}

					<!-- Notifications (hidden during impersonation) -->
					<!-- Mobile: bell sits before the "+" to match the mockup order (bell · +); desktop keeps its natural position. -->
					{#if !$authStore.user?.is_impersonating}
						<div class="relative order-first sm:order-none">
							<!-- Desktop mockup: the bell is a 40px boxed control on
							     surface-2, not a bare icon. -->
							<NotificationPanel
								triggerClass="notification-bell relative inline-flex h-10 w-10 items-center justify-center rounded-lg border border-border bg-surface-2 text-text-muted transition-colors hover:text-text-strong"
								iconClass="w-4.5 h-4.5"
							/>
						</div>
					{/if}
				{/if}

				<!-- Logout (hidden during impersonation; desktop only). Replaces the old
				     user dropdown, which only repeated links that already exist in the
				     nav and on /profile. -->
				{#if !$authStore.user?.is_impersonating}
					<button
						type="button"
						onclick={handleLogout}
						class="hidden items-center text-sm text-text-subtle transition-colors hover:text-danger-600 sm:flex"
						aria-label={$t('nav.logout')}
						title={$t('nav.logout')}
					>
						<svg
							class="w-6 h-6"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"
							></path>
						</svg>
					</button>
				{/if}
			</div>
		</div>
	</div>
</nav>
