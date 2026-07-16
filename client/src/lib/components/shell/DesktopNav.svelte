<script lang="ts">
	import { page } from '$app/stores';
	import { resolve } from '$app/paths';
	import { goto } from '$app/navigation';
	import NotificationPanel from '$lib/components/NotificationPanel.svelte';
	import { authStore } from '$lib/stores/auth';
	import { configStore } from '$lib/stores/config';
	import { t } from '$lib/stores/i18n';
	import { toastStore } from '$lib/stores/toast';
	import { showOfflineBanner } from '$lib/stores/offline';
	import { showNewDialog } from '$lib/stores/newDialog';
	import { logger } from '$lib/utils/logger';
	import { platform } from '$lib/utils/platform';
	import type { UserDTO } from '$lib/types/api';

	const navLogger = logger.child('DesktopNav');

	// Server-provided user (data.user) — carries is_admin for the admin dropdown gate.
	let { user }: { user?: UserDTO } = $props();

	let showUserMenu = $state(false);
	let showAdminMenu = $state(false);
	let desktopSearch = $state('');

	async function handleLogout() {
		showUserMenu = false;
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
	class="hidden transition-all duration-300 ease-out sm:block"
	class:mt-16={$showOfflineBanner}
	class:sm:mt-12={$showOfflineBanner}
>
	<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
		<div class="flex justify-between h-16">
			<div class="flex items-center">
				<div class="flex-shrink-0">
					<a href={resolve('/dashboard')} class="flex items-center space-x-3">
						<img src="/logo.png" alt="Savvy Logo" class="h-8 w-auto" />
						<span class="hidden sm:inline text-2xl font-bold text-accent"
							>{$t('common.appName')}</span
						>
					</a>
				</div>

				<!-- App shell: three places — Start · Wallet · Profile -->
				<div class="hidden sm:ml-6 sm:flex sm:space-x-8">
					<a
						href={resolve('/dashboard')}
						data-testid="nav-start-desktop"
						class="border-transparent text-text-subtle hover:border-border-field hover:text-text-ink2 inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium transition-colors"
						class:border-accent={$page.url.pathname.startsWith('/dashboard')}
						class:text-text={$page.url.pathname.startsWith('/dashboard')}
					>
						{$t('nav.start')}
					</a>
					<a
						href={resolve('/wallet')}
						data-testid="nav-wallet-desktop"
						class="border-transparent text-text-subtle hover:border-border-field hover:text-text-ink2 inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium transition-colors"
						class:border-accent={$page.url.pathname.startsWith('/wallet')}
						class:text-text={$page.url.pathname.startsWith('/wallet')}
					>
						{$t('nav.wallet')}
					</a>
					<a
						href={resolve('/profile')}
						data-testid="nav-profile-desktop"
						class="border-transparent text-text-subtle hover:border-border-field hover:text-text-ink2 inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium transition-colors"
						class:border-accent={$page.url.pathname.startsWith('/profile')}
						class:text-text={$page.url.pathname.startsWith('/profile')}
					>
						{$t('nav.profile')}
					</a>
				</div>
			</div>

			<div class="flex items-center space-x-4">
				<!-- Desktop: inline search + New (single global entries) -->
				<form
					onsubmit={submitDesktopSearch}
					class="hidden sm:flex items-center"
					role="search"
				>
					<input
						type="search"
						bind:value={desktopSearch}
						placeholder={$t('common.search')}
						aria-label={$t('common.search')}
						class="w-48 px-3 py-1.5 text-sm bg-white border border-border-field rounded-md focus:ring-accent focus:border-accent"
					/>
				</form>
				<button
					type="button"
					onclick={() => ($showNewDialog = true)}
					data-testid="nav-new-desktop"
					class="hidden sm:inline-flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium bg-accent text-white rounded-md hover:bg-accent-hover transition-colors"
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

				<!-- Admin Dropdown (only for admins, hidden during impersonation; desktop only — mockup mobile header has no admin controls) -->
				{#if !$authStore.user?.is_impersonating}
					{#if user?.is_admin && !$authStore.user?.is_impersonating}
						<div class="relative hidden sm:block">
							<button
								type="button"
								onclick={(e) => {
									e.stopPropagation();
									showAdminMenu = !showAdminMenu;
								}}
								class="inline-flex text-purple-600 hover:text-purple-700 p-2 rounded-md hover:bg-purple-50 transition-colors"
								title={$t('nav.admin')}
								aria-expanded={showAdminMenu}
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
										d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"
									></path>
								</svg>
							</button>

							{#if showAdminMenu}
								<div
									role="menu"
									tabindex="-1"
									class="absolute right-0 mt-2 w-56 rounded-lg shadow-xl py-2 z-50 {platform ===
									'ios'
										? 'bg-white/70 backdrop-blur-xl backdrop-saturate-150 border border-white/30'
										: 'bg-white border border-border'}"
									onclick={(e) => e.stopPropagation()}
									onkeydown={(e) => {
										if (e.key === 'Escape') showAdminMenu = false;
									}}
								>
									<div
										class="px-4 py-3 text-sm border-b {platform === 'ios'
											? 'border-white/30'
											: 'border-border'}"
									>
										<div class="font-semibold text-purple-700">
											{$t('nav.admin')}
										</div>
									</div>

									<a
										href={resolve('/admin/users')}
										onclick={() => (showAdminMenu = false)}
										class="flex items-center w-full px-4 py-3 text-sm text-text-ink2 hover:bg-surface-1 transition-colors"
									>
										<svg
											class="w-5 h-5 mr-3"
											fill="none"
											stroke="currentColor"
											viewBox="0 0 24 24"
										>
											<path
												stroke-linecap="round"
												stroke-linejoin="round"
												stroke-width="2"
												d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z"
											></path>
										</svg>
										{$t('nav.adminUsers')}
									</a>

									<a
										href={resolve('/admin/merchants')}
										onclick={() => (showAdminMenu = false)}
										class="flex items-center w-full px-4 py-3 text-sm text-text-ink2 hover:bg-surface-1 transition-colors"
									>
										<svg
											class="w-5 h-5 mr-3"
											fill="none"
											stroke="currentColor"
											viewBox="0 0 24 24"
										>
											<path
												stroke-linecap="round"
												stroke-linejoin="round"
												stroke-width="2"
												d="M3 3h2l.4 2M7 13h10l4-8H5.4M7 13L5.4 5M7 13l-2.293 2.293c-.63.63-.184 1.707.707 1.707H17m0 0a2 2 0 100 4 2 2 0 000-4zm-8 2a2 2 0 11-4 0 2 2 0 014 0z"
											></path>
										</svg>
										{$t('nav.adminMerchants')}
									</a>

									<a
										href={resolve('/admin/audit-log')}
										onclick={() => (showAdminMenu = false)}
										class="flex items-center w-full px-4 py-3 text-sm text-text-ink2 hover:bg-surface-1 transition-colors"
									>
										<svg
											class="w-5 h-5 mr-3"
											fill="none"
											stroke="currentColor"
											viewBox="0 0 24 24"
										>
											<path
												stroke-linecap="round"
												stroke-linejoin="round"
												stroke-width="2"
												d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01"
											></path>
										</svg>
										{$t('nav.adminAuditLog')}
									</a>

									<a
										href={resolve('/admin/system-health')}
										onclick={() => (showAdminMenu = false)}
										class="flex items-center w-full px-4 py-3 text-sm text-text-ink2 hover:bg-surface-1 transition-colors"
									>
										<svg
											class="w-5 h-5 mr-3"
											fill="none"
											stroke="currentColor"
											viewBox="0 0 24 24"
										>
											<path
												stroke-linecap="round"
												stroke-linejoin="round"
												stroke-width="2"
												d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
											></path>
										</svg>
										{$t('nav.adminSystemHealth')}
									</a>

									{#if $configStore.is_development}
										<a
											href={resolve('/admin/email-templates')}
											onclick={() => (showAdminMenu = false)}
											class="flex items-center w-full px-4 py-3 text-sm text-text-ink2 hover:bg-surface-1 transition-colors"
										>
											<svg
												class="w-5 h-5 mr-3"
												fill="none"
												stroke="currentColor"
												viewBox="0 0 24 24"
											>
												<path
													stroke-linecap="round"
													stroke-linejoin="round"
													stroke-width="2"
													d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
												></path>
											</svg>
											{$t('nav.adminEmailTemplates')}
										</a>
									{/if}
								</div>
							{/if}
						</div>
					{/if}

					<!-- Notifications (hidden during impersonation) -->
					<!-- Mobile: bell sits before the "+" to match the mockup order (bell · +); desktop keeps its natural position. -->
					{#if !$authStore.user?.is_impersonating}
						<div class="relative order-first sm:order-none">
							<NotificationPanel />
						</div>
					{/if}
				{/if}

				<!-- User Menu (hidden during impersonation; desktop only — Profile is a bottom-nav place on mobile, so no redundant user icon in the mobile header) -->
				{#if !$authStore.user?.is_impersonating}
					<div class="relative hidden sm:block">
						<button
							type="button"
							onclick={(e) => {
								e.stopPropagation();
								showUserMenu = !showUserMenu;
							}}
							class="flex items-center text-sm text-text-subtle hover:text-text-ink2 transition-colors"
							aria-label={$t('aria.openUserMenu')}
							aria-expanded={showUserMenu}
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
									d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"
								></path>
							</svg>
						</button>

						{#if showUserMenu}
							<div
								role="menu"
								tabindex="-1"
								class="absolute right-0 mt-2 w-56 rounded-lg shadow-xl py-2 z-50 {platform ===
								'ios'
									? 'bg-white/70 backdrop-blur-xl backdrop-saturate-150 border border-white/30'
									: 'bg-white border border-border'}"
								onclick={(e) => e.stopPropagation()}
								onkeydown={(e) => {
									if (e.key === 'Escape') showUserMenu = false;
								}}
							>
								{#if $authStore.user}
									<div
										class="px-4 py-3 text-sm border-b {platform === 'ios'
											? 'border-white/30'
											: 'border-border'}"
									>
										{#if $authStore.user.first_name || $authStore.user.last_name}
											<div class="font-semibold text-text mb-1">
												{$authStore.user.first_name || ''}
												{$authStore.user.last_name || ''}
											</div>
										{/if}
										<div class="text-xs text-text-muted break-all">
											{$authStore.user.email}
										</div>
									</div>
								{/if}

								<a
									href={resolve('/profile')}
									onclick={() => (showUserMenu = false)}
									class="flex items-center w-full px-4 py-3 text-sm text-text-ink2 hover:bg-surface-1 transition-colors"
								>
									<svg
										class="w-5 h-5 mr-3"
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
									{$t('nav.profile')}
								</a>

								<a
									href={resolve('/security')}
									onclick={() => (showUserMenu = false)}
									class="flex items-center w-full px-4 py-3 text-sm text-text-ink2 hover:bg-surface-1 transition-colors"
								>
									<svg
										class="w-5 h-5 mr-3"
										fill="none"
										stroke="currentColor"
										viewBox="0 0 24 24"
									>
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											stroke-width="2"
											d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"
										></path>
									</svg>
									{$t('nav.security')}
								</a>

								<a
									href={resolve('/notifications')}
									onclick={() => (showUserMenu = false)}
									class="flex items-center w-full px-4 py-3 text-sm text-text-ink2 hover:bg-surface-1 transition-colors"
								>
									<svg
										class="w-5 h-5 mr-3"
										fill="none"
										stroke="currentColor"
										viewBox="0 0 24 24"
									>
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											stroke-width="2"
											d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"
										></path>
									</svg>
									{$t('nav.notifications')}
								</a>

								<button
									type="button"
									onclick={handleLogout}
									class="flex items-center w-full px-4 py-3 text-sm text-danger-600 hover:bg-surface-1 transition-colors"
								>
									<svg
										class="w-5 h-5 mr-3"
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
									{$t('nav.logout')}
								</button>
							</div>
						{/if}
					</div>
				{/if}
			</div>
		</div>
	</div>
</nav>

<svelte:window
	onclick={() => {
		if (showUserMenu) showUserMenu = false;
		if (showAdminMenu) showAdminMenu = false;
	}}
/>
