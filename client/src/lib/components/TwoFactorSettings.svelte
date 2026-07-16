<script lang="ts">
	import { get } from 'svelte/store';
	import { authApi } from '$lib/api';
	import { t } from '$lib/stores/i18n';
	import { toastStore } from '$lib/stores/toast';
	import { logger } from '$lib/utils/logger';
	import Modal from '$lib/components/ui/Modal.svelte';

	const componentLogger = logger.child('TwoFactorSettings');
	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);

	interface Props {
		authProvider: string;
		// Whether *new* enrollment is allowed (server ENABLE_2FA). When false,
		// users with existing 2FA still see status + disable/rotate, but no new
		// setup is offered. Defaults to true.
		enrollmentEnabled?: boolean;
	}

	let { authProvider, enrollmentEnabled = true }: Props = $props();

	// State
	let isEnabled = $state(false);
	let isLocalAuth = $state(false);
	let isLoading = $state(true);
	let showSetup = $state(false);
	let showDisable = $state(false);
	let showRegenerate = $state(false);
	let showBackupCodes = $state(false);

	// Setup state
	let qrCodeUrl = $state('');
	let secret = $state('');
	let backupCodes = $state<string[]>([]);
	let setupCode = $state('');
	let isSettingUp = $state(false);
	let showManualEntry = $state(false);

	// Disable state
	let disableCode = $state('');
	let isDisabling = $state(false);

	// Regenerate state
	let regenerateCode = $state('');
	let isRegenerating = $state(false);

	// Backup codes confirmed
	let backupCodesConfirmed = $state(false);

	$effect(() => {
		loadStatus();
	});

	async function loadStatus() {
		try {
			const status = await authApi.get2FAStatus();
			isEnabled = status.enabled;
			isLocalAuth = status.is_local_auth;
		} catch (error) {
			componentLogger.error('Failed to load 2FA status', { error });
		} finally {
			isLoading = false;
		}
	}

	async function startSetup() {
		isSettingUp = true;
		try {
			const result = await authApi.setup2FA();
			qrCodeUrl = result.qr_code_url;
			secret = result.secret;
			backupCodes = result.backup_codes;
			showSetup = true;
		} catch (error) {
			componentLogger.error('Failed to start 2FA setup', { error });
			toastStore.error(tr('settings.twoFactor.enableError'));
		} finally {
			isSettingUp = false;
		}
	}

	async function verifyAndEnable(e: Event) {
		e.preventDefault();
		isSettingUp = true;
		try {
			await authApi.verify2FASetup(setupCode.trim());
			isEnabled = true;
			showSetup = false;
			showBackupCodes = true;
			toastStore.success(tr('settings.twoFactor.enableSuccess'));
			setupCode = '';
		} catch (error) {
			componentLogger.error('Failed to verify 2FA', { error });
			toastStore.error(tr('settings.twoFactor.enableError'));
		} finally {
			isSettingUp = false;
		}
	}

	async function handleDisable(e: Event) {
		e.preventDefault();
		isDisabling = true;
		try {
			await authApi.disable2FA(disableCode.trim());
			isEnabled = false;
			showDisable = false;
			disableCode = '';
			toastStore.success(tr('settings.twoFactor.disableSuccess'));
		} catch (error) {
			componentLogger.error('Failed to disable 2FA', { error });
			toastStore.error(tr('settings.twoFactor.disableError'));
		} finally {
			isDisabling = false;
		}
	}

	async function handleRegenerate(e: Event) {
		e.preventDefault();
		isRegenerating = true;
		try {
			const result = await authApi.regenerateBackupCodes(regenerateCode.trim());
			backupCodes = result.backup_codes;
			showRegenerate = false;
			showBackupCodes = true;
			backupCodesConfirmed = false;
			regenerateCode = '';
			toastStore.success(tr('settings.twoFactor.regenerateSuccess'));
		} catch (error) {
			componentLogger.error('Failed to regenerate backup codes', { error });
			toastStore.error(tr('settings.twoFactor.regenerateError'));
		} finally {
			isRegenerating = false;
		}
	}

	function closeBackupCodes() {
		showBackupCodes = false;
		backupCodes = [];
		backupCodesConfirmed = false;
	}

	function closeSetup() {
		showSetup = false;
		setupCode = '';
	}

	function closeDisable() {
		showDisable = false;
		disableCode = '';
	}

	function closeRegenerate() {
		showRegenerate = false;
		regenerateCode = '';
	}
</script>

{#snippet pingDot()}
	<span class="relative inline-flex h-3 w-3 mr-2"
		><span
			class="animate-ping absolute inline-flex h-full w-full rounded-full bg-accent-400 opacity-75"
		></span><span class="relative inline-flex rounded-full h-3 w-3 bg-accent"
		></span></span
	>
{/snippet}

<div class="overflow-hidden rounded-xl border border-border bg-white p-6">
	<h2 class="text-lg font-semibold text-text mb-2">
		{tr('settings.twoFactor.title')}
	</h2>
	<p class="text-sm text-text-muted mb-4">
		{tr('settings.twoFactor.description')}
	</p>

	{#if isLoading}
		<div class="flex justify-center py-4">
			<span class="relative inline-flex h-5 w-5"
				><span
					class="animate-ping absolute inline-flex h-full w-full rounded-full bg-accent-400 opacity-75"
				></span><span
					class="relative inline-flex rounded-full h-5 w-5 bg-accent"
				></span></span
			>
		</div>
	{:else if authProvider !== 'local'}
		<p class="text-sm text-text-subtle">
			{tr('settings.twoFactor.oauthNote')}
		</p>
	{:else if !isLocalAuth}
		<p class="text-sm text-text-subtle">
			{tr('settings.twoFactor.oauthNote')}
		</p>
	{:else}
		<!-- Status -->
		<div class="flex items-center justify-between mb-4">
			<div class="flex items-center gap-2">
				{#if isEnabled}
					<span
						class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-success-100 text-success-800"
					>
						<svg class="w-3 h-3 mr-1" fill="currentColor" viewBox="0 0 20 20">
							<path
								fill-rule="evenodd"
								d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
								clip-rule="evenodd"
							/>
						</svg>
						{tr('settings.twoFactor.enabled')}
					</span>
				{:else}
					<span
						class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-border-soft text-text-strong"
					>
						{tr('settings.twoFactor.disabled')}
					</span>
				{/if}
			</div>
		</div>

		<!-- Actions -->
		{#if isEnabled}
			<div class="space-y-2">
				<button
					type="button"
					onclick={() => {
						showRegenerate = true;
					}}
					class="btn btn-ghost text-sm w-full"
				>
					{tr('settings.twoFactor.regenerateBackupCodes')}
				</button>
				<button
					type="button"
					onclick={() => {
						showDisable = true;
					}}
					class="btn btn-ghost text-sm w-full text-danger-600 border-danger-300 hover:bg-danger-50"
				>
					{tr('settings.twoFactor.disableButton')}
				</button>
			</div>
		{:else if enrollmentEnabled}
			<button
				type="button"
				onclick={startSetup}
				disabled={isSettingUp}
				class="btn btn-primary text-sm w-full"
			>
				{#if isSettingUp}
					{@render pingDot()}
				{/if}
				{tr('settings.twoFactor.enableButton')}
			</button>
		{:else}
			<p class="text-sm text-text-subtle">
				{tr('settings.twoFactor.enrollmentDisabled')}
			</p>
		{/if}
	{/if}
</div>

<!-- Setup Modal -->
<Modal
	open={showSetup}
	onclose={closeSetup}
	layer="default"
	mobileLayout="center"
	label={tr('settings.twoFactor.setupTitle')}
>
	<div
		class="pointer-events-auto bg-white rounded-lg shadow-xl max-w-md w-full p-6 max-h-[90vh] overflow-y-auto"
	>
		<h3 class="text-lg font-semibold text-text mb-4">
			{tr('settings.twoFactor.setupTitle')}
		</h3>

		<div class="space-y-4 mb-6">
			<p class="text-sm text-text-muted">
				{tr('settings.twoFactor.setupStep1')}
			</p>
			<p class="text-sm text-text-muted">
				{tr('settings.twoFactor.setupStep2')}
			</p>
			<p class="text-sm text-text-muted">
				{tr('settings.twoFactor.setupStep3')}
			</p>
		</div>

		<!-- QR Code -->
		<div class="flex flex-col items-center mb-6">
			{#if !showManualEntry}
				<p class="text-sm font-medium text-text-ink2 mb-2">
					{tr('settings.twoFactor.scanQrCode')}
				</p>
				<img
					src={qrCodeUrl}
					alt="QR Code"
					class="w-48 h-48 border rounded-lg"
				/>
				<button
					type="button"
					onclick={() => {
						showManualEntry = true;
					}}
					class="text-sm text-accent hover:text-accent mt-2"
				>
					{tr('settings.twoFactor.manualEntry')}
				</button>
			{:else}
				<p class="text-sm font-medium text-text-ink2 mb-2">
					{tr('settings.twoFactor.secretKey')}
				</p>
				<code
					class="bg-border-soft px-4 py-2 rounded text-sm font-mono break-all select-all"
				>
					{secret}
				</code>
				<button
					type="button"
					onclick={() => {
						showManualEntry = false;
					}}
					class="text-sm text-accent hover:text-accent mt-2"
				>
					{tr('settings.twoFactor.scanQrCode')}
				</button>
			{/if}
		</div>

		<!-- Verify Code -->
		<form onsubmit={verifyAndEnable} class="space-y-4">
			<div>
				<label
					for="setupCode"
					class="block text-sm font-medium text-text-ink2 mb-1"
				>
					{tr('settings.twoFactor.enterCode')}
				</label>
				<input
					id="setupCode"
					type="text"
					inputmode="numeric"
					pattern="[0-9]*"
					maxlength="6"
					autocomplete="one-time-code"
					required
					bind:value={setupCode}
					disabled={isSettingUp}
					class="w-full px-4 py-3 bg-white border border-border-field rounded-md focus:ring-accent focus:border-accent text-center text-2xl font-mono tracking-widest"
					placeholder={tr('settings.twoFactor.codePlaceholder')}
				/>
			</div>

			<div class="flex gap-3">
				<button
					type="button"
					onclick={closeSetup}
					disabled={isSettingUp}
					class="btn btn-ghost flex-1"
				>
					{tr('common.cancel')}
				</button>
				<button
					type="submit"
					disabled={isSettingUp || setupCode.length !== 6}
					class="btn btn-primary flex-1"
				>
					{#if isSettingUp}
						{@render pingDot()}
						{tr('settings.twoFactor.verifying')}
					{:else}
						{tr('settings.twoFactor.verifyAndEnable')}
					{/if}
				</button>
			</div>
		</form>
	</div>
</Modal>

<!-- Backup Codes Modal — deliberately NOT dismissable via Escape/backdrop:
     the user must tick "I saved my codes" and click Done (onclose is a no-op
     so the confirmation gate can't be bypassed). -->
<Modal
	open={showBackupCodes && backupCodes.length > 0}
	onclose={() => {}}
	layer="default"
	mobileLayout="center"
	label={tr('settings.twoFactor.backupCodesTitle')}
>
	<div
		class="pointer-events-auto bg-white rounded-lg shadow-xl max-w-md w-full p-6"
	>
		<h3 class="text-lg font-semibold text-text mb-2">
			{tr('settings.twoFactor.backupCodesTitle')}
		</h3>
		<p class="text-sm text-text-muted mb-4">
			{tr('settings.twoFactor.backupCodesDescription')}
		</p>

		<div class="bg-surface-1 rounded-lg p-4 mb-4">
			<div class="grid grid-cols-2 gap-2">
				{#each backupCodes as code (code)}
					<code
						class="text-sm font-mono text-center bg-white px-3 py-1.5 rounded border select-all"
					>
						{code}
					</code>
				{/each}
			</div>
		</div>

		<label class="flex items-center gap-2 mb-4 cursor-pointer">
			<input
				type="checkbox"
				bind:checked={backupCodesConfirmed}
				class="rounded border-border-field text-accent focus:ring-accent"
			/>
			<span class="text-sm text-text-ink2"
				>{tr('settings.twoFactor.backupCodesSaved')}</span
			>
		</label>

		<button
			type="button"
			onclick={closeBackupCodes}
			disabled={!backupCodesConfirmed}
			class="btn btn-primary w-full"
		>
			{tr('common.done')}
		</button>
	</div>
</Modal>

<!-- Disable Modal -->
<Modal
	open={showDisable}
	onclose={closeDisable}
	layer="default"
	mobileLayout="center"
	label={tr('settings.twoFactor.disableButton')}
>
	<div
		class="pointer-events-auto bg-white rounded-lg shadow-xl max-w-md w-full p-6"
	>
		<h3 class="text-lg font-semibold text-danger-600 mb-2">
			{tr('settings.twoFactor.disableButton')}
		</h3>
		<p class="text-sm text-text-muted mb-4">
			{tr('settings.twoFactor.disableConfirm')}
		</p>

		<form onsubmit={handleDisable} class="space-y-4">
			<div>
				<input
					type="text"
					inputmode="numeric"
					pattern="[0-9]*"
					maxlength="6"
					autocomplete="one-time-code"
					required
					bind:value={disableCode}
					disabled={isDisabling}
					class="w-full px-4 py-3 bg-white border border-border-field rounded-md focus:ring-accent focus:border-accent text-center text-2xl font-mono tracking-widest"
					placeholder={tr('settings.twoFactor.codePlaceholder')}
				/>
			</div>

			<div class="flex gap-3">
				<button
					type="button"
					onclick={closeDisable}
					disabled={isDisabling}
					class="btn btn-ghost flex-1"
				>
					{tr('common.cancel')}
				</button>
				<button
					type="submit"
					disabled={isDisabling || disableCode.length !== 6}
					class="flex-1 px-4 py-2 bg-danger-600 text-white rounded-md hover:bg-danger-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors text-sm font-medium"
				>
					{#if isDisabling}
						{@render pingDot()}
					{/if}
					{tr('settings.twoFactor.disableButton')}
				</button>
			</div>
		</form>
	</div>
</Modal>

<!-- Regenerate Backup Codes Modal -->
<Modal
	open={showRegenerate}
	onclose={closeRegenerate}
	layer="default"
	mobileLayout="center"
	label={tr('settings.twoFactor.regenerateBackupCodes')}
>
	<div
		class="pointer-events-auto bg-white rounded-lg shadow-xl max-w-md w-full p-6"
	>
		<h3 class="text-lg font-semibold text-text mb-2">
			{tr('settings.twoFactor.regenerateBackupCodes')}
		</h3>
		<p class="text-sm text-text-muted mb-4">
			{tr('settings.twoFactor.regenerateConfirm')}
		</p>

		<form onsubmit={handleRegenerate} class="space-y-4">
			<div>
				<input
					type="text"
					inputmode="numeric"
					pattern="[0-9]*"
					maxlength="6"
					autocomplete="one-time-code"
					required
					bind:value={regenerateCode}
					disabled={isRegenerating}
					class="w-full px-4 py-3 bg-white border border-border-field rounded-md focus:ring-accent focus:border-accent text-center text-2xl font-mono tracking-widest"
					placeholder={tr('settings.twoFactor.codePlaceholder')}
				/>
			</div>

			<div class="flex gap-3">
				<button
					type="button"
					onclick={closeRegenerate}
					disabled={isRegenerating}
					class="btn btn-ghost flex-1"
				>
					{tr('common.cancel')}
				</button>
				<button
					type="submit"
					disabled={isRegenerating || regenerateCode.length !== 6}
					class="btn btn-primary flex-1"
				>
					{#if isRegenerating}
						{@render pingDot()}
					{/if}
					{tr('settings.twoFactor.regenerateBackupCodes')}
				</button>
			</div>
		</form>
	</div>
</Modal>
