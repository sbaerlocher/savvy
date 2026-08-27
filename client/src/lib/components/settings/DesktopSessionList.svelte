<script lang="ts">
	import type { SessionDTO } from '$lib/api';
	import { t } from '$lib/stores/i18n';
	import { get } from 'svelte/store';

	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);

	interface Props {
		sessions: SessionDTO[];
		isLoading: boolean;
		revokingSessionId: string | null;
		isRevokingOthers: boolean;
		onRevoke: (sessionId: string) => void;
		onRevokeOthers: () => void;
	}

	let {
		sessions,
		isLoading,
		revokingSessionId,
		isRevokingOthers,
		onRevoke,
		onRevokeOthers
	}: Props = $props();

	// The mockup distinguishes a phone session from a desktop one by glyph.
	// device_info is server-parsed free text, so match on the mobile keywords
	// it produces and fall back to the desktop glyph.
	function isMobile(session: SessionDTO): boolean {
		return /mobile|phone|android|iphone|ipad|tablet/i.test(
			session.device_info || ''
		);
	}

	function formatRelativeTime(dateStr: string): string {
		try {
			const date = new Date(dateStr);
			const diffMs = Date.now() - date.getTime();
			const diffMinutes = Math.floor(diffMs / 60000);
			const diffHours = Math.floor(diffMs / 3600000);
			const diffDays = Math.floor(diffMs / 86400000);

			if (diffMinutes < 1) return tr('notifications.timeAgo.justNow');
			if (diffMinutes < 60)
				return tr('notifications.timeAgo.minutesAgo', { count: diffMinutes });
			if (diffHours < 24)
				return tr('notifications.timeAgo.hoursAgo', { count: diffHours });
			return tr('notifications.timeAgo.daysAgo', { count: diffDays });
		} catch {
			return dateStr;
		}
	}
</script>

<div class="rounded-xl border border-border bg-white p-6">
	<div class="mb-4 flex items-center justify-between gap-3">
		<h3 class="text-subheading font-semibold text-text">
			{$t('settings.sessions.title')}
		</h3>
		{#if sessions.length > 1}
			<button
				type="button"
				onclick={onRevokeOthers}
				disabled={isRevokingOthers}
				class="text-body-sm font-semibold text-accent transition-opacity hover:opacity-60 disabled:opacity-50"
			>
				{isRevokingOthers ? '…' : $t('settings.sessions.revokeOthers')}
			</button>
		{/if}
	</div>

	{#if isLoading}
		<div class="flex justify-center py-4">
			<span class="relative inline-flex h-4 w-4"
				><span
					class="absolute inline-flex h-full w-full animate-ping rounded-full bg-accent-400 opacity-75"
				></span><span
					class="relative inline-flex h-4 w-4 rounded-full bg-accent"
				></span></span
			>
		</div>
	{:else if sessions.length === 0}
		<p class="text-body text-text-subtle">
			{$t('settings.sessions.noOtherSessions')}
		</p>
	{:else}
		<div class="flex flex-col gap-2.5">
			{#each sessions as session (session.id)}
				<div
					class="flex items-start gap-3 rounded-xl p-3 {session.is_current
						? 'border border-accent-200 bg-accent-50'
						: 'bg-surface-1'}"
				>
					<span
						class="flex h-8.5 w-8.5 flex-none items-center justify-center rounded-md {session.is_current
							? 'bg-accent-100 text-accent-850'
							: 'bg-tile-tint text-text-subtle'}"
					>
						{#if isMobile(session)}
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
								<rect x="6" y="2" width="12" height="20" rx="2.5" />
								<path d="M11 18h2" />
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
								<rect x="3" y="4" width="18" height="12" rx="2" />
								<path d="M8 20h8M12 16v4" />
							</svg>
						{/if}
					</span>

					<div class="min-w-0 flex-1">
						<div class="flex flex-wrap items-center gap-1.5">
							<span class="text-body font-semibold text-text">
								{session.browser_info ||
									$t('settings.sessions.unknownBrowser')}{session.device_info
									? ` · ${session.device_info}`
									: ''}
							</span>
							{#if session.is_current}
								<span
									class="rounded-full bg-accent-100 px-2 py-px text-tag font-semibold text-accent-850"
								>
									{$t('settings.sessions.current')}
								</span>
							{/if}
						</div>
						<div class="mt-0.5 font-mono text-mono-sm text-text-muted">
							{#if session.ip_address}{session.ip_address} ·
							{/if}{formatRelativeTime(session.last_active_at)}
						</div>
					</div>

					{#if !session.is_current}
						<button
							type="button"
							onclick={() => onRevoke(session.id)}
							disabled={revokingSessionId === session.id}
							class="flex h-7 w-7 flex-none items-center justify-center rounded-sm border border-border-soft bg-white text-text-subtle transition-colors hover:bg-danger-50 hover:text-danger-600 disabled:opacity-50"
							title={$t('settings.sessions.revoke')}
							aria-label={$t('settings.sessions.revoke')}
						>
							{#if revokingSessionId === session.id}
								<span class="relative inline-flex h-3 w-3"
									><span
										class="absolute inline-flex h-full w-full animate-ping rounded-full bg-accent-400 opacity-75"
									></span><span
										class="relative inline-flex h-3 w-3 rounded-full bg-accent"
									></span></span
								>
							{:else}
								<svg
									class="h-3.5 w-3.5"
									viewBox="0 0 24 24"
									fill="none"
									stroke="currentColor"
									stroke-width="2.2"
									stroke-linecap="round"
									aria-hidden="true"
								>
									<path d="M6 6l12 12M18 6L6 18" />
								</svg>
							{/if}
						</button>
					{/if}
				</div>
			{/each}
		</div>
	{/if}
</div>
