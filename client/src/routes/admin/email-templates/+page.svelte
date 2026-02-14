<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { t } from '$lib/stores/i18n';
	import { configStore } from '$lib/stores/config';
	import { toastStore } from '$lib/stores/toast';
	import { logger } from '$lib/utils/logger';
	import { adminApi } from '$lib/api';

	const pageLogger = logger.child('EmailTemplatesPage');

	let selectedLanguage = $state('de');
	let sendingTemplate = $state<string | null>(null);
	let expandedTemplate = $state<string | null>(null);

	function toggleTemplate(id: string) {
		expandedTemplate = expandedTemplate === id ? null : id;
	}

	const templates = $derived([
		{
			id: 'test_email',
			name: $t('admin.emailTemplates.templates.testEmail'),
			description: $t('admin.emailTemplates.templates.testEmailDesc')
		},
		{
			id: 'password_reset',
			name: $t('admin.emailTemplates.templates.passwordReset'),
			description: $t('admin.emailTemplates.templates.passwordResetDesc')
		},
		{
			id: 'email_verification',
			name: $t('admin.emailTemplates.templates.emailVerification'),
			description: $t('admin.emailTemplates.templates.emailVerificationDesc')
		},
		{
			id: 'account_deleted',
			name: $t('admin.emailTemplates.templates.accountDeleted'),
			description: $t('admin.emailTemplates.templates.accountDeletedDesc')
		},
		{
			id: 'expiry_reminder',
			name: $t('admin.emailTemplates.templates.expiryReminder'),
			description: $t('admin.emailTemplates.templates.expiryReminderDesc')
		},
		{
			id: 'share_notification',
			name: $t('admin.emailTemplates.templates.shareNotification'),
			description: $t('admin.emailTemplates.templates.shareNotificationDesc')
		},
		{
			id: 'transfer_notification',
			name: $t('admin.emailTemplates.templates.transferNotification'),
			description: $t('admin.emailTemplates.templates.transferNotificationDesc')
		},
		{
			id: 'validity_start',
			name: $t('admin.emailTemplates.templates.validityStart'),
			description: $t('admin.emailTemplates.templates.validityStartDesc')
		}
	]);

	onMount(async () => {
		await configStore.loaded;
		if (!$configStore.is_development) {
			goto('/admin/users');
		}
	});

	async function sendPreview(templateId: string) {
		if (sendingTemplate) return;

		sendingTemplate = templateId;
		try {
			const data = await adminApi.sendPreviewEmail(
				templateId,
				selectedLanguage
			);
			toastStore.success(
				data.message || $t('admin.emailTemplates.sendSuccess')
			);
		} catch (error) {
			pageLogger.error('Failed to send preview email', { error, templateId });
			toastStore.error($t('admin.emailTemplates.sendError'));
		} finally {
			sendingTemplate = null;
		}
	}
</script>

<svelte:head>
	<title>{$t('nav.adminEmailTemplates')} - {$t('common.appName')}</title>
</svelte:head>

<div class="px-4 pb-20 md:pb-4">
	<!-- Header -->
	<div class="mb-8">
		<h1 class="text-3xl font-bold text-gray-900">
			{$t('nav.adminEmailTemplates')}
		</h1>
	</div>

	<!-- Controls Row: Language Selector -->
	<div class="flex flex-col sm:flex-row gap-3 mb-6">
		<div
			class="sm:flex-1 flex items-center gap-3 h-[42px] px-4 bg-white border border-gray-300 rounded-md"
		>
			<svg
				class="w-5 h-5 text-gray-400 shrink-0"
				fill="none"
				stroke="currentColor"
				viewBox="0 0 24 24"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d="M3 5h12M9 3v2m1.048 9.5A18.022 18.022 0 016.412 9m6.088 9h7M11 21l5-10 5 10M12.751 5C11.783 10.77 8.07 15.61 3 18.129"
				/>
			</svg>
			<label for="language" class="text-sm text-gray-500 shrink-0"
				>{$t('admin.emailTemplates.language')}:</label
			>
			<select
				id="language"
				bind:value={selectedLanguage}
				class="text-sm font-medium text-gray-900 bg-transparent border-none focus:ring-0 p-0 cursor-pointer"
			>
				<option value="de">Deutsch (DE)</option>
				<option value="en">English (EN)</option>
				<option value="fr">Français (FR)</option>
			</select>
		</div>
	</div>

	<!-- Mobile: Card List -->
	<div class="md:hidden bg-white shadow rounded-lg divide-y divide-gray-200">
		{#each templates as template (template.id)}
			<button
				class="w-full px-4 py-3 flex items-center gap-3 text-left"
				onclick={() => toggleTemplate(template.id)}
			>
				<div class="text-cyan-600">
					<svg
						class="w-5 h-5"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
						/>
					</svg>
				</div>
				<span class="text-sm font-medium text-gray-900 flex-1"
					>{template.name}</span
				>
				<svg
					class="w-4 h-4 text-gray-400 transition-transform {expandedTemplate ===
					template.id
						? 'rotate-180'
						: ''}"
					fill="none"
					stroke="currentColor"
					viewBox="0 0 24 24"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M19 9l-7 7-7-7"
					/>
				</svg>
			</button>
			{#if expandedTemplate === template.id}
				<div class="px-4 py-3 bg-gray-50 space-y-3">
					<p class="text-sm text-gray-600">{template.description}</p>
					<button
						class="btn btn-sm btn-ghost w-full"
						onclick={() => sendPreview(template.id)}
						disabled={sendingTemplate !== null}
					>
						{#if sendingTemplate === template.id}
							{$t('admin.emailTemplates.sending')}
						{:else}
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
									d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8"
								/>
							</svg>
							{$t('admin.emailTemplates.sendPreview')}
						{/if}
					</button>
				</div>
			{/if}
		{/each}
	</div>

	<!-- Desktop: Table -->
	<div class="hidden md:block bg-white shadow rounded-lg overflow-hidden">
		<table class="min-w-full divide-y divide-gray-200">
			<thead class="bg-gray-50">
				<tr>
					<th
						class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
					>
						{$t('admin.emailTemplates.template')}
					</th>
					<th
						class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
					>
						{$t('admin.emailTemplates.description')}
					</th>
					<th
						class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider"
					>
						{$t('admin.users.actions')}
					</th>
				</tr>
			</thead>
			<tbody class="bg-white divide-y divide-gray-200">
				{#each templates as template (template.id)}
					<tr class="hover:bg-gray-50 transition-colors">
						<td class="px-6 py-4 whitespace-nowrap">
							<div class="flex items-center gap-3">
								<div class="text-cyan-600">
									<svg
										class="w-5 h-5"
										fill="none"
										stroke="currentColor"
										viewBox="0 0 24 24"
									>
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											stroke-width="2"
											d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
										/>
									</svg>
								</div>
								<span class="text-sm font-medium text-gray-900"
									>{template.name}</span
								>
							</div>
						</td>
						<td class="px-6 py-4 text-sm text-gray-500">
							{template.description}
						</td>
						<td class="px-6 py-4 whitespace-nowrap text-right">
							<button
								class="btn btn-sm btn-ghost"
								onclick={() => sendPreview(template.id)}
								disabled={sendingTemplate !== null}
							>
								{#if sendingTemplate === template.id}
									{$t('admin.emailTemplates.sending')}
								{:else}
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
											d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8"
										/>
									</svg>
									{$t('admin.emailTemplates.sendPreview')}
								{/if}
							</button>
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>

	<!-- Info Box -->
	<div class="mt-6 bg-yellow-50 border border-yellow-200 rounded-lg p-4">
		<div class="flex items-start gap-3">
			<svg
				class="w-5 h-5 text-yellow-600 mt-0.5"
				fill="none"
				stroke="currentColor"
				viewBox="0 0 24 24"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
				/>
			</svg>
			<div>
				<h4 class="font-semibold text-yellow-900 mb-1">
					{$t('admin.emailTemplates.devOnlyTitle')}
				</h4>
				<p class="text-sm text-yellow-800">
					{$t('admin.emailTemplates.devOnlyDesc')}
				</p>
			</div>
		</div>
	</div>
</div>
