<script lang="ts">
	import type { ProfileDTO } from '$lib/api';
	import { profileApi } from '$lib/api';
	import ToggleSwitch from './ToggleSwitch.svelte';
	import M3SettingsRow from './M3SettingsRow.svelte';
	import SectionLabel from '$lib/components/ui/SectionLabel.svelte';
	import { configStore } from '$lib/stores/config';
	import { t } from '$lib/stores/i18n';
	import { pushStore } from '$lib/stores/push';
	import { toastStore } from '$lib/stores/toast';
	import { logger } from '$lib/utils/logger';
	import { platform } from '$lib/utils/platform';
	import { onMount, onDestroy } from 'svelte';
	import { SvelteMap } from 'svelte/reactivity';
	import { get } from 'svelte/store';

	const pageLogger = logger.child('NotificationsSection');
	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);

	interface Props {
		profile: ProfileDTO;
		onProfileUpdated: (profile: ProfileDTO) => void;
		/** Render the Android section header. The combined settings screen sets
		 *  it; the standalone /notifications/settings route already has a page
		 *  header with the same title. */
		sectionHeader?: boolean;
		/** The desktop settings page already names the section through its tab,
		 *  so it renders the card without the redundant heading. */
		showTitle?: boolean;
	}

	let {
		profile,
		onProfileUpdated,
		sectionHeader = false,
		showTitle = true
	}: Props = $props();

	// iOS renders the grouped-inset channel groups from screen-SettingsIOS:
	// one group per channel, subcategories on a recessed sub-surface that is
	// only present while the channel is on. `platform` is a module constant,
	// so this is a plain const, not $derived.
	const IS_IOS = platform === 'ios';
	const IS_ANDROID = platform === 'android';

	// Mockup glyph paths (screen-SettingsAndroid).
	const ICON_BELL_BODY = 'M18 8a6 6 0 10-12 0c0 7-3 9-3 9h18s-3-2-3-9';
	const ICON_BELL_CLAPPER = 'M13.5 21a1.7 1.7 0 01-3 0';
	const ICON_ENVELOPE_BOX = 'M3 5h18v14H3z';
	const ICON_ENVELOPE_FLAP = 'M3 7l9 6 9-6';

	type PreferenceKey =
		| 'push_notifications_enabled'
		| 'email_notifications_enabled'
		| 'push_reminders_enabled'
		| 'push_sharing_enabled'
		| 'email_reminders_enabled'
		| 'email_sharing_enabled';

	// Derived from profile prop, but locally writable for optimistic toggles
	let preferences: Record<PreferenceKey, boolean> = $derived({
		push_notifications_enabled: profile.push_notifications_enabled ?? true,
		email_notifications_enabled: profile.email_notifications_enabled ?? true,
		push_reminders_enabled: profile.push_reminders_enabled ?? true,
		push_sharing_enabled: profile.push_sharing_enabled ?? true,
		email_reminders_enabled: profile.email_reminders_enabled ?? true,
		email_sharing_enabled: profile.email_sharing_enabled ?? true
	});

	let savingKeys = $state<Set<PreferenceKey>>(new Set());
	let debounceTimers = new SvelteMap<
		PreferenceKey,
		ReturnType<typeof setTimeout>
	>();

	onMount(async () => {
		if ($configStore.push_enabled) {
			await pushStore.init();

			if (preferences.push_notifications_enabled && !$pushStore.isSubscribed) {
				try {
					const resp = await profileApi.update({
						push_notifications_enabled: false
					});
					onProfileUpdated(resp.profile);
					preferences.push_notifications_enabled = false;
					pageLogger.info(
						'Auto-disabled push preference (browser not subscribed)'
					);
				} catch {
					pageLogger.warn('Failed to sync push preference');
				}
			}
		}

		if (typeof document !== 'undefined') {
			document.addEventListener('visibilitychange', handleVisibilityChange);
		}
	});

	onDestroy(() => {
		if (typeof document !== 'undefined') {
			document.removeEventListener('visibilitychange', handleVisibilityChange);
		}
		for (const timer of debounceTimers.values()) {
			clearTimeout(timer);
		}
	});

	async function handleVisibilityChange() {
		if (document.visibilityState !== 'visible') return;
		if (!$configStore.push_enabled) return;

		await pushStore.init();

		if (preferences.push_notifications_enabled && !$pushStore.isSubscribed) {
			try {
				const resp = await profileApi.update({
					push_notifications_enabled: false
				});
				onProfileUpdated(resp.profile);
				preferences.push_notifications_enabled = false;
			} catch {
				pageLogger.warn('Failed to sync push preference on visibility change');
			}
		}
	}

	async function handleToggle(key: PreferenceKey) {
		// Special handling for push_notifications_enabled (browser permission)
		if (key === 'push_notifications_enabled') {
			await handleTogglePush();
			return;
		}

		const newValue = !preferences[key];
		preferences[key] = newValue;

		// Debounce: cancel pending save for this key
		const existing = debounceTimers.get(key);
		if (existing) clearTimeout(existing);

		const timer = setTimeout(() => {
			debounceTimers.delete(key);
			savePreference(key, newValue);
		}, 400);
		debounceTimers.set(key, timer);
	}

	async function savePreference(key: PreferenceKey, value: boolean) {
		savingKeys = new Set([...savingKeys, key]);

		try {
			const response = await profileApi.update({ [key]: value });
			onProfileUpdated(response.profile);
			preferences[key] = response.profile[key] as boolean;
			toastStore.success(tr('settings.notifications.saved'));
		} catch {
			// Revert on error
			preferences[key] = !value;
			toastStore.error(tr('settings.notifications.error'));
		} finally {
			savingKeys = new Set([...savingKeys].filter((k) => k !== key));
		}
	}

	async function handleTogglePush() {
		const key: PreferenceKey = 'push_notifications_enabled';
		savingKeys = new Set([...savingKeys, key]);
		const newValue = !preferences[key];

		try {
			if (
				newValue &&
				$configStore.push_enabled &&
				$pushStore.isSupported &&
				!$pushStore.isSubscribed
			) {
				const subscribed = await pushStore.enablePush();
				if (!subscribed && $pushStore.permission === 'denied') {
					toastStore.error(tr('settings.pushNotifications.permissionDenied'));
					savingKeys = new Set([...savingKeys].filter((k) => k !== key));
					return;
				}
			}

			const response = await profileApi.update({
				push_notifications_enabled: newValue
			});
			onProfileUpdated(response.profile);
			preferences[key] = response.profile.push_notifications_enabled;

			if (!newValue && $pushStore.isSubscribed) {
				await pushStore.disablePush();
			}

			toastStore.success(tr('settings.notifications.saved'));
		} catch {
			toastStore.error(tr('settings.notifications.error'));
		} finally {
			savingKeys = new Set([...savingKeys].filter((k) => k !== key));
		}
	}
</script>

{#if IS_ANDROID}
	{#if sectionHeader}
		<h2 class="text-label px-6 pt-1.5 pb-2 text-accent">
			{tr('settings.sections.notifications')}
		</h2>
	{/if}

	<M3SettingsRow
		icon="{ICON_BELL_BODY} {ICON_BELL_CLAPPER}"
		title={tr('settings.notifications.androidPush')}
		subtitle={tr('settings.notifications.androidPushDesc')}
	>
		{#snippet trailing()}
			<ToggleSwitch
				bare
				checked={preferences.push_notifications_enabled}
				label={tr('settings.notifications.pushNotifications')}
				isSaving={savingKeys.has('push_notifications_enabled')}
				onToggle={() => handleToggle('push_notifications_enabled')}
			/>
		{/snippet}
	</M3SettingsRow>

	{#if $pushStore.permission === 'denied' && preferences.push_notifications_enabled}
		<p class="px-6 pb-1 text-body-sm text-danger-500">
			{tr('settings.pushNotifications.permissionDenied')}
		</p>
	{/if}

	{#if preferences.push_notifications_enabled}
		<div
			data-testid="push-subcategories"
			class="mx-6 mb-1 border-l-2 border-accent-100 pl-4"
		>
			<div class="flex items-center gap-3 px-2 py-2.5">
				<span class="flex-1 text-body text-text-ink2">
					{tr('settings.notifications.androidReminders')}
				</span>
				<ToggleSwitch
					bare
					checked={preferences.push_reminders_enabled}
					label={tr('settings.notifications.pushReminders')}
					isSaving={savingKeys.has('push_reminders_enabled')}
					onToggle={() => handleToggle('push_reminders_enabled')}
				/>
			</div>
			<div class="flex items-center gap-3 px-2 py-2.5">
				<span class="flex-1 text-body text-text-ink2">
					{tr('settings.notifications.androidSharing')}
				</span>
				<ToggleSwitch
					bare
					checked={preferences.push_sharing_enabled}
					label={tr('settings.notifications.pushSharing')}
					isSaving={savingKeys.has('push_sharing_enabled')}
					onToggle={() => handleToggle('push_sharing_enabled')}
				/>
			</div>
		</div>
	{/if}

	<M3SettingsRow
		icon="{ICON_ENVELOPE_BOX} {ICON_ENVELOPE_FLAP}"
		title={tr('settings.notifications.androidEmail')}
		subtitle={tr('settings.notifications.androidEmailDesc', {
			email: profile.email
		})}
	>
		{#snippet trailing()}
			<ToggleSwitch
				bare
				checked={preferences.email_notifications_enabled}
				label={tr('settings.notifications.emailNotifications')}
				isSaving={savingKeys.has('email_notifications_enabled')}
				onToggle={() => handleToggle('email_notifications_enabled')}
			/>
		{/snippet}
	</M3SettingsRow>

	{#if preferences.email_notifications_enabled}
		<div
			data-testid="email-subcategories"
			class="mx-6 mb-1 border-l-2 border-accent-100 pl-4"
		>
			<div class="flex items-center gap-3 px-2 py-2.5">
				<span class="flex-1 text-body text-text-ink2">
					{tr('settings.notifications.androidReminders')}
				</span>
				<ToggleSwitch
					bare
					checked={preferences.email_reminders_enabled}
					label={tr('settings.notifications.emailReminders')}
					isSaving={savingKeys.has('email_reminders_enabled')}
					onToggle={() => handleToggle('email_reminders_enabled')}
				/>
			</div>
			<div class="flex items-center gap-3 px-2 py-2.5">
				<span class="flex-1 text-body text-text-ink2">
					{tr('settings.notifications.androidSharing')}
				</span>
				<ToggleSwitch
					bare
					checked={preferences.email_sharing_enabled}
					label={tr('settings.notifications.emailSharing')}
					isSaving={savingKeys.has('email_sharing_enabled')}
					onToggle={() => handleToggle('email_sharing_enabled')}
				/>
			</div>
		</div>
	{/if}
{:else if IS_IOS}
	<!-- Grouped-inset channel groups (screen-SettingsIOS). The subcategory
	     block sits on the recessed sub-surface and is only rendered while its
	     channel is on, exactly as the mockup's nested state shows. -->
	<SectionLabel inset>{tr('settings.notifications.title')}</SectionLabel>

	<div class="mb-3.5 overflow-hidden rounded-inset bg-surface">
		<div class="flex items-center gap-3 px-4 py-3.5">
			<span class="flex-1">
				<span
					class="block text-[length:var(--text-code)] font-normal text-text"
				>
					{tr('settings.notifications.pushNotifications')}
				</span>
				<span class="mt-0.25 block text-body-sm text-text-subtle">
					{tr('settings.notifications.pushNotificationsDesc')}
				</span>
			</span>
			<ToggleSwitch
				bare
				checked={preferences.push_notifications_enabled}
				label={tr('settings.notifications.pushNotifications')}
				isSaving={savingKeys.has('push_notifications_enabled')}
				onToggle={() => handleToggle('push_notifications_enabled')}
			/>
		</div>

		{#if $pushStore.permission === 'denied' && preferences.push_notifications_enabled}
			<p class="px-4 pb-3 text-body-sm text-danger-500">
				{tr('settings.pushNotifications.permissionDenied')}
			</p>
		{/if}

		{#if preferences.push_notifications_enabled}
			<div
				data-testid="push-subcategories"
				class="border-t border-border-soft bg-surface-2"
			>
				<div
					class="flex items-center gap-3 border-b border-border-soft py-3 pl-7.5 pr-4"
				>
					<span class="flex-1 text-body text-text-ink2">
						{tr('settings.notifications.pushReminders')}
					</span>
					<ToggleSwitch
						bare
						checked={preferences.push_reminders_enabled}
						label={tr('settings.notifications.pushReminders')}
						isSaving={savingKeys.has('push_reminders_enabled')}
						onToggle={() => handleToggle('push_reminders_enabled')}
					/>
				</div>
				<div class="flex items-center gap-3 py-3 pl-7.5 pr-4">
					<span class="flex-1 text-body text-text-ink2">
						{tr('settings.notifications.pushSharing')}
					</span>
					<ToggleSwitch
						bare
						checked={preferences.push_sharing_enabled}
						label={tr('settings.notifications.pushSharing')}
						isSaving={savingKeys.has('push_sharing_enabled')}
						onToggle={() => handleToggle('push_sharing_enabled')}
					/>
				</div>
			</div>
		{/if}
	</div>

	<div class="mb-2 overflow-hidden rounded-inset bg-surface">
		<div class="flex items-center gap-3 px-4 py-3.5">
			<span class="flex-1">
				<span
					class="block text-[length:var(--text-code)] font-normal text-text"
				>
					{tr('settings.notifications.emailNotifications')}
				</span>
				<!-- Mockup shows the real target address on this subtitle. -->
				<span class="mt-0.25 block text-body-sm text-text-subtle">
					{profile.email}
				</span>
			</span>
			<ToggleSwitch
				bare
				checked={preferences.email_notifications_enabled}
				label={tr('settings.notifications.emailNotifications')}
				isSaving={savingKeys.has('email_notifications_enabled')}
				onToggle={() => handleToggle('email_notifications_enabled')}
			/>
		</div>

		{#if preferences.email_notifications_enabled}
			<div
				data-testid="email-subcategories"
				class="border-t border-border-soft bg-surface-2"
			>
				<div
					class="flex items-center gap-3 border-b border-border-soft py-3 pl-7.5 pr-4"
				>
					<span class="flex-1 text-body text-text-ink2">
						{tr('settings.notifications.emailReminders')}
					</span>
					<ToggleSwitch
						bare
						checked={preferences.email_reminders_enabled}
						label={tr('settings.notifications.emailReminders')}
						isSaving={savingKeys.has('email_reminders_enabled')}
						onToggle={() => handleToggle('email_reminders_enabled')}
					/>
				</div>
				<div class="flex items-center gap-3 py-3 pl-7.5 pr-4">
					<span class="flex-1 text-body text-text-ink2">
						{tr('settings.notifications.emailSharing')}
					</span>
					<ToggleSwitch
						bare
						checked={preferences.email_sharing_enabled}
						label={tr('settings.notifications.emailSharing')}
						isSaving={savingKeys.has('email_sharing_enabled')}
						onToggle={() => handleToggle('email_sharing_enabled')}
					/>
				</div>
			</div>
		{/if}
	</div>

	<p class="px-1.5 text-body-sm text-text-faint">
		{tr('settings.notifications.ios.subcategoryHint')}
	</p>
{:else}
	<div>
		<div class="overflow-hidden rounded-xl border border-border bg-white">
			<div class="p-6">
				{#if showTitle}
					<h3 class="text-lg font-semibold text-text mb-4">
						{tr('settings.notifications.title')}
					</h3>
				{/if}

				<!-- Push Notifications Channel -->
				<ToggleSwitch
					checked={preferences.push_notifications_enabled}
					label={tr('settings.notifications.pushNotifications')}
					description={tr('settings.notifications.pushNotificationsDesc')}
					isSaving={savingKeys.has('push_notifications_enabled')}
					onToggle={() => handleToggle('push_notifications_enabled')}
				/>

				{#if $pushStore.permission === 'denied' && preferences.push_notifications_enabled}
					<p class="text-xs text-danger-500 mt-2">
						{tr('settings.pushNotifications.permissionDenied')}
					</p>
				{/if}

				<!-- Push subcategories -->
				{#if preferences.push_notifications_enabled}
					<div
						data-testid="push-subcategories"
						class="ml-4 mt-3 pl-4 border-l-2 border-border-soft space-y-3"
					>
						<ToggleSwitch
							checked={preferences.push_reminders_enabled}
							label={tr('settings.notifications.pushReminders')}
							description={tr('settings.notifications.pushRemindersDesc')}
							isSaving={savingKeys.has('push_reminders_enabled')}
							onToggle={() => handleToggle('push_reminders_enabled')}
						/>
						<ToggleSwitch
							checked={preferences.push_sharing_enabled}
							label={tr('settings.notifications.pushSharing')}
							description={tr('settings.notifications.pushSharingDesc')}
							isSaving={savingKeys.has('push_sharing_enabled')}
							onToggle={() => handleToggle('push_sharing_enabled')}
						/>
					</div>
				{/if}

				<!-- Email Notifications Channel -->
				<div class="mt-4 pt-4 border-t border-border-soft">
					<ToggleSwitch
						checked={preferences.email_notifications_enabled}
						label={tr('settings.notifications.emailNotifications')}
						description={tr('settings.notifications.emailNotificationsDesc')}
						isSaving={savingKeys.has('email_notifications_enabled')}
						onToggle={() => handleToggle('email_notifications_enabled')}
					/>
				</div>

				<!-- Email subcategories -->
				{#if preferences.email_notifications_enabled}
					<div
						data-testid="email-subcategories"
						class="ml-4 mt-3 pl-4 border-l-2 border-border-soft space-y-3"
					>
						<ToggleSwitch
							checked={preferences.email_reminders_enabled}
							label={tr('settings.notifications.emailReminders')}
							description={tr('settings.notifications.emailRemindersDesc')}
							isSaving={savingKeys.has('email_reminders_enabled')}
							onToggle={() => handleToggle('email_reminders_enabled')}
						/>
						<ToggleSwitch
							checked={preferences.email_sharing_enabled}
							label={tr('settings.notifications.emailSharing')}
							description={tr('settings.notifications.emailSharingDesc')}
							isSaving={savingKeys.has('email_sharing_enabled')}
							onToggle={() => handleToggle('email_sharing_enabled')}
						/>
					</div>
				{/if}
			</div>
		</div>
	</div>
{/if}
