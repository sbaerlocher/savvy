<script lang="ts">
	import type { ProfileDTO } from '$lib/api';
	import { profileApi } from '$lib/api';
	import ToggleSwitch from './ToggleSwitch.svelte';
	import { configStore } from '$lib/stores/config';
	import { t } from '$lib/stores/i18n';
	import { pushStore } from '$lib/stores/push';
	import { toastStore } from '$lib/stores/toast';
	import { logger } from '$lib/utils/logger';
	import { onMount, onDestroy } from 'svelte';
	import { get } from 'svelte/store';

	const pageLogger = logger.child('NotificationsSection');
	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);

	interface Props {
		profile: ProfileDTO;
		onProfileUpdated: (profile: ProfileDTO) => void;
	}

	let { profile, onProfileUpdated }: Props = $props();

	type PreferenceKey =
		| 'push_notifications_enabled'
		| 'email_notifications_enabled'
		| 'push_reminders_enabled'
		| 'push_sharing_enabled'
		| 'email_reminders_enabled'
		| 'email_sharing_enabled';

	let preferences = $state<Record<PreferenceKey, boolean>>({
		push_notifications_enabled: profile.push_notifications_enabled ?? true,
		email_notifications_enabled: profile.email_notifications_enabled ?? true,
		push_reminders_enabled: profile.push_reminders_enabled ?? true,
		push_sharing_enabled: profile.push_sharing_enabled ?? true,
		email_reminders_enabled: profile.email_reminders_enabled ?? true,
		email_sharing_enabled: profile.email_sharing_enabled ?? true
	});

	let savingKeys = $state<Set<PreferenceKey>>(new Set());
	let debounceTimers = new Map<PreferenceKey, ReturnType<typeof setTimeout>>();

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

<div>
	<div
		class="bg-white rounded-lg shadow-lg overflow-hidden"
		style="border-left: 6px solid #06b6d4"
	>
		<div class="p-6">
			<h3 class="text-lg font-semibold text-gray-900 mb-4">
				{tr('settings.notifications.title')}
			</h3>

			<!-- Push Notifications Channel -->
			<ToggleSwitch
				checked={preferences.push_notifications_enabled}
				label={tr('settings.notifications.pushNotifications')}
				description={tr('settings.notifications.pushNotificationsDesc')}
				isSaving={savingKeys.has('push_notifications_enabled')}
				onToggle={() => handleToggle('push_notifications_enabled')}
			/>

			{#if $pushStore.permission === 'denied' && preferences.push_notifications_enabled}
				<p class="text-xs text-red-500 mt-2">
					{tr('settings.pushNotifications.permissionDenied')}
				</p>
			{/if}

			<!-- Push subcategories -->
			{#if preferences.push_notifications_enabled}
				<div class="ml-4 mt-3 pl-4 border-l-2 border-gray-100 space-y-3">
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
			<div class="mt-4 pt-4 border-t border-gray-100">
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
				<div class="ml-4 mt-3 pl-4 border-l-2 border-gray-100 space-y-3">
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
