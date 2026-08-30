<script lang="ts">
	import { t } from '$lib/stores/i18n';
	import { toastStore } from '$lib/stores/toast';
	import { logger } from '$lib/utils/logger';
	import { platform } from '$lib/utils/platform';
	import { adminApi } from '$lib/api';

	const pageLogger = logger.child('EmailTemplatesPage');

	const DESKTOP = platform === 'other';

	// Android renders the M3 chrome the other admin screens use: a language
	// card plus one tonal card per template, expanding into description and
	// send action. `platform` is a module constant, so a plain const.
	const IS_ANDROID = platform === 'android';

	// iOS renders the liquid-glass chrome the admin-users screen uses: glass
	// cards with icon circle + chevron rows, expanding in place.
	const IS_IOS = platform === 'ios';

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

{#if DESKTOP}
	<!-- Desktop: language selector and table share one elevated panel
	     (same pattern as /admin/users). -->
	<div
		class="overflow-hidden rounded-4xl border border-border bg-surface shadow-panel"
	>
		<div class="px-7.5 pt-6 pb-4.5">
			<div class="flex items-center gap-3">
				<svg
					class="w-5 h-5 text-text-faint shrink-0"
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
				<label for="language" class="text-sm text-text-subtle shrink-0"
					>{$t('admin.emailTemplates.language')}:</label
				>
				<select
					id="language"
					bind:value={selectedLanguage}
					class="text-sm font-medium text-text bg-transparent border-none focus:ring-0 p-0 cursor-pointer"
				>
					<option value="de">Deutsch (DE)</option>
					<option value="en">English (EN)</option>
					<option value="fr">Français (FR)</option>
				</select>
			</div>
		</div>

		<div class="border-t border-border-soft">
			<table class="min-w-full">
				<thead>
					<tr class="border-b border-border-soft bg-surface-1">
						<th
							class="px-7.5 py-3 text-left text-section-eyebrow uppercase text-text-subtle"
						>
							{$t('admin.emailTemplates.template')}
						</th>
						<th
							class="px-7.5 py-3 text-left text-section-eyebrow uppercase text-text-subtle"
						>
							{$t('admin.emailTemplates.description')}
						</th>
						<th
							class="px-7.5 py-3 text-right text-section-eyebrow uppercase text-text-subtle"
						>
							{$t('admin.users.actions')}
						</th>
					</tr>
				</thead>
				<tbody>
					{#each templates as template (template.id)}
						<tr
							class="border-b border-border-soft transition-colors hover:bg-surface-1"
						>
							<td class="px-7.5 py-3.5 whitespace-nowrap">
								<div class="flex items-center gap-3">
									<div class="text-accent">
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
									<span class="text-body font-medium text-text"
										>{template.name}</span
									>
								</div>
							</td>
							<td class="px-7.5 py-3.5 text-body text-text-ink2">
								{template.description}
							</td>
							<td class="px-7.5 py-3.5 text-right whitespace-nowrap">
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
	</div>
{:else if IS_IOS}
	<!-- Language selector inside its own glass card; the .ios-glass-form
	     override restyles the shared input/label classes to glass fields. -->
	<div
		class="liquid-glass-card ios-glass-form mb-3.5 rounded-[var(--radius-inset)] p-3.75"
	>
		<label for="language" class="label"
			>{$t('admin.emailTemplates.language')}</label
		>
		<select id="language" bind:value={selectedLanguage} class="input">
			<option value="de">Deutsch (DE)</option>
			<option value="en">English (EN)</option>
			<option value="fr">Français (FR)</option>
		</select>
	</div>

	<!-- One glass card per template, expanding into description plus the
	     send-preview pill (same row anatomy as /admin/users on iOS). -->
	<div class="flex flex-col gap-2.5">
		{#each templates as template (template.id)}
			{@const expanded = expandedTemplate === template.id}
			<div
				class="liquid-glass-card overflow-hidden rounded-[var(--radius-inset)]"
			>
				<button
					type="button"
					onclick={() => toggleTemplate(template.id)}
					aria-expanded={expanded}
					class="flex w-full items-center gap-3 px-3.75 py-3.25 text-left transition-colors active:bg-surface-1"
				>
					<span
						class="flex h-9.5 w-9.5 shrink-0 items-center justify-center rounded-full bg-accent-50 text-accent"
					>
						<svg
							class="h-4.75 w-4.75"
							fill="none"
							stroke="currentColor"
							stroke-width="2"
							viewBox="0 0 24 24"
							aria-hidden="true"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
							/>
						</svg>
					</span>
					<span
						class="min-w-0 flex-1 truncate text-body font-semibold text-text"
						>{template.name}</span
					>
					<svg
						class="h-4 w-4 shrink-0 text-text-faint transition-transform {expanded
							? 'rotate-180'
							: ''}"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
						viewBox="0 0 24 24"
						aria-hidden="true"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							d="M19 9l-7 7-7-7"
						/>
					</svg>
				</button>

				{#if expanded}
					<div class="border-t border-border-soft bg-surface-2 px-3.75 py-3.5">
						<p class="text-body-sm text-text-muted">{template.description}</p>
						<button
							type="button"
							onclick={() => sendPreview(template.id)}
							disabled={sendingTemplate !== null}
							class="bg-accent text-on-accent text-body-sm mt-3 inline-flex h-9 items-center gap-1.75 rounded-full px-4 font-semibold disabled:opacity-50"
						>
							{#if sendingTemplate === template.id}
								{$t('admin.emailTemplates.sending')}
							{:else}
								<svg
									class="h-3.75 w-3.75"
									fill="none"
									stroke="currentColor"
									stroke-width="2"
									viewBox="0 0 24 24"
									aria-hidden="true"
								>
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8"
									/>
								</svg>
								{$t('admin.emailTemplates.sendPreview')}
							{/if}
						</button>
					</div>
				{/if}
			</div>
		{/each}
	</div>
{:else if IS_ANDROID}
	<!-- Language selector inside its own M3 card; the shared input/label
	     classes carry the field styling. -->
	<div class="rounded-m3-lg bg-m3-card border-border mb-3.5 border p-4">
		<label for="language" class="label"
			>{$t('admin.emailTemplates.language')}</label
		>
		<select id="language" bind:value={selectedLanguage} class="input">
			<option value="de">Deutsch (DE)</option>
			<option value="en">English (EN)</option>
			<option value="fr">Français (FR)</option>
		</select>
	</div>

	<!-- One tonal card per template, expanding into description plus the
	     send-preview pill. -->
	<div class="flex flex-col gap-2.5">
		{#each templates as template (template.id)}
			{@const expanded = expandedTemplate === template.id}
			<div
				class="rounded-m3-lg bg-m3-card border-border overflow-hidden border"
			>
				<button
					type="button"
					onclick={() => toggleTemplate(template.id)}
					aria-expanded={expanded}
					class="hover:bg-ground-active flex w-full items-center gap-3.5 px-4 py-3.25 text-left transition-colors"
				>
					<span
						class="bg-tile-tint text-accent rounded-m3-full flex h-10 w-10 shrink-0 items-center justify-center"
					>
						<svg
							class="h-4.75 w-4.75"
							fill="none"
							stroke="currentColor"
							stroke-width="2"
							viewBox="0 0 24 24"
							aria-hidden="true"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
							/>
						</svg>
					</span>
					<span
						class="text-body text-text min-w-0 flex-1 truncate font-semibold"
						>{template.name}</span
					>
					<svg
						class="text-text-subtle h-4 w-4 shrink-0 transition-transform {expanded
							? 'rotate-180'
							: ''}"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
						viewBox="0 0 24 24"
						aria-hidden="true"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							d="M19 9l-7 7-7-7"
						/>
					</svg>
				</button>

				{#if expanded}
					<div class="bg-surface-1 px-4 pt-0.5 pb-4">
						<div class="border-accent-100 border-l-2 pt-3 pl-3.5">
							<p class="text-body-sm text-text-muted">{template.description}</p>
							<button
								type="button"
								onclick={() => sendPreview(template.id)}
								disabled={sendingTemplate !== null}
								class="border-border-field bg-m3-card text-text rounded-m3-full text-body-sm mt-3 inline-flex h-9 items-center gap-1.75 border px-4 font-semibold disabled:opacity-50"
							>
								{#if sendingTemplate === template.id}
									{$t('admin.emailTemplates.sending')}
								{:else}
									<svg
										class="h-3.75 w-3.75"
										fill="none"
										stroke="currentColor"
										stroke-width="2"
										viewBox="0 0 24 24"
										aria-hidden="true"
									>
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8"
										/>
									</svg>
									{$t('admin.emailTemplates.sendPreview')}
								{/if}
							</button>
						</div>
					</div>
				{/if}
			</div>
		{/each}
	</div>
{:else}
	<!-- Controls Row: Language Selector -->
	<div class="flex flex-col sm:flex-row gap-3 mb-6">
		<div
			class="sm:flex-1 flex items-center gap-3 control px-4 bg-white border border-border-field rounded-md"
		>
			<svg
				class="w-5 h-5 text-text-faint shrink-0"
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
			<label for="language" class="text-sm text-text-subtle shrink-0"
				>{$t('admin.emailTemplates.language')}:</label
			>
			<select
				id="language"
				bind:value={selectedLanguage}
				class="text-sm font-medium text-text bg-transparent border-none focus:ring-0 p-0 cursor-pointer"
			>
				<option value="de">Deutsch (DE)</option>
				<option value="en">English (EN)</option>
				<option value="fr">Français (FR)</option>
			</select>
		</div>
	</div>

	<!-- Mobile: Card List -->
	<div class="md:hidden bg-white shadow rounded-lg divide-y divide-border">
		{#each templates as template (template.id)}
			<button
				class="w-full px-4 py-3 flex items-center gap-3 text-left"
				onclick={() => toggleTemplate(template.id)}
			>
				<div class="text-accent">
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
				<span class="text-sm font-medium text-text flex-1">{template.name}</span
				>
				<svg
					class="w-4 h-4 text-text-faint transition-transform {expandedTemplate ===
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
				<div class="px-4 py-3 bg-surface-1 space-y-3">
					<p class="text-sm text-text-muted">{template.description}</p>
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
		<table class="min-w-full divide-y divide-border">
			<thead class="bg-surface-1">
				<tr>
					<th
						class="px-6 py-3 text-left text-xs font-medium text-text-subtle uppercase tracking-wider"
					>
						{$t('admin.emailTemplates.template')}
					</th>
					<th
						class="px-6 py-3 text-left text-xs font-medium text-text-subtle uppercase tracking-wider"
					>
						{$t('admin.emailTemplates.description')}
					</th>
					<th
						class="px-6 py-3 text-right text-xs font-medium text-text-subtle uppercase tracking-wider"
					>
						{$t('admin.users.actions')}
					</th>
				</tr>
			</thead>
			<tbody class="bg-white divide-y divide-border">
				{#each templates as template (template.id)}
					<tr class="hover:bg-surface-1 transition-colors">
						<td class="px-6 py-4 whitespace-nowrap">
							<div class="flex items-center gap-3">
								<div class="text-accent">
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
								<span class="text-sm font-medium text-text"
									>{template.name}</span
								>
							</div>
						</td>
						<td class="px-6 py-4 text-sm text-text-subtle">
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
{/if}

<!-- Info Box -->
<div
	class="mt-6 bg-warning-50 border border-warning-200 p-4 {IS_ANDROID
		? 'rounded-m3-lg'
		: 'rounded-lg'}"
>
	<div class="flex items-start gap-3">
		<svg
			class="w-5 h-5 text-warning-600 mt-0.5"
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
			<h4 class="font-semibold text-warning-900 mb-1">
				{$t('admin.emailTemplates.devOnlyTitle')}
			</h4>
			<p class="text-sm text-warning-800">
				{$t('admin.emailTemplates.devOnlyDesc')}
			</p>
		</div>
	</div>
</div>
