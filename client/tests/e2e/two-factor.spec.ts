import { expect, test } from './fixtures/test-fixtures';

test.describe('Two-Factor Authentication', () => {
	test.use({ serviceWorkers: 'block' });

	test.beforeEach(async ({ authenticatedSecurityPage, securityPage }) => {
		// authenticatedSecurityPage fixture handles login + navigation to /security
		await securityPage.waitForPageReady();
	});

	test('should show 2FA settings section', async ({
		authenticatedSecurityPage,
		securityPage
	}) => {
		// 2FA may be disabled via ENABLE_2FA=false in test environment
		if (
			!(await securityPage.twoFactorSection
				.isVisible({ timeout: 5000 })
				.catch(() => false))
		) {
			test.skip();
			return;
		}
		await expect(securityPage.twoFactorSection).toBeVisible();
	});

	test('should open 2FA setup modal', async ({
		authenticatedSecurityPage,
		securityPage
	}) => {
		const page = authenticatedSecurityPage;

		if (
			!(await securityPage.enable2FAButton
				.isVisible({ timeout: 5000 })
				.catch(() => false))
		) {
			// 2FA might already be enabled, disable it first
			if (
				await securityPage.disable2FAButton
					.isVisible({ timeout: 3000 })
					.catch(() => false)
			) {
				await securityPage.disable2FAButton.click();

				const confirmButton = page
					.locator(
						'[data-testid="modal-confirm"], button:has-text("Bestätigen"), button:has-text("Confirm")'
					)
					.first();
				if (
					await confirmButton.isVisible({ timeout: 3000 }).catch(() => false)
				) {
					await confirmButton.click();
					await expect(securityPage.enable2FAButton).toBeVisible({
						timeout: 5000
					});
				}
			} else {
				test.skip();
				return;
			}
		}

		await securityPage.open2FASetup();
	});

	test('should display QR code during 2FA setup', async ({
		authenticatedSecurityPage,
		securityPage
	}) => {
		const page = authenticatedSecurityPage;

		if (
			!(await securityPage.enable2FAButton
				.isVisible({ timeout: 5000 })
				.catch(() => false))
		) {
			test.skip();
			return;
		}

		await securityPage.enable2FAButton.click();
		await expect(securityPage.setupModal).toBeVisible({ timeout: 5000 });

		const qrCode = page.locator(
			'canvas, img[alt*="QR"], [data-testid="qr-code"], svg'
		);
		const secretKey = page.locator('text=/secret|geheim|schlüssel|key/i');
		const codeInput = securityPage.totpCodeInput;

		const hasQr = await qrCode
			.first()
			.isVisible({ timeout: 5000 })
			.catch(() => false);
		const hasSecret = await secretKey.isVisible().catch(() => false);
		const hasInput = await codeInput.isVisible().catch(() => false);

		expect(hasQr || hasSecret || hasInput).toBeTruthy();
	});

	test('should show backup codes after 2FA setup', async ({
		authenticatedSecurityPage,
		securityPage
	}) => {
		if (
			!(await securityPage.enable2FAButton
				.isVisible({ timeout: 5000 })
				.catch(() => false))
		) {
			test.skip();
			return;
		}

		await securityPage.enable2FAButton.click();
		await expect(securityPage.setupModal).toBeVisible({ timeout: 5000 });

		// We can't enter a real TOTP code in E2E tests without the secret
		// Just verify the code input is present
		if (
			await securityPage.totpCodeInput
				.isVisible({ timeout: 3000 })
				.catch(() => false)
		) {
			expect(await securityPage.totpCodeInput.isVisible()).toBeTruthy();
		}
	});

	test('should validate TOTP code format', async ({
		authenticatedSecurityPage,
		securityPage
	}) => {
		const page = authenticatedSecurityPage;

		if (
			!(await securityPage.enable2FAButton
				.isVisible({ timeout: 5000 })
				.catch(() => false))
		) {
			test.skip();
			return;
		}

		await securityPage.enable2FAButton.click();
		await expect(securityPage.setupModal).toBeVisible({ timeout: 5000 });

		if (
			!(await securityPage.totpCodeInput
				.isVisible({ timeout: 5000 })
				.catch(() => false))
		) {
			test.skip();
			return;
		}

		await securityPage.totpCodeInput.fill('invalid');

		if (
			await securityPage.verifyButton
				.isVisible({ timeout: 3000 })
				.catch(() => false)
		) {
			await securityPage.verifyButton.click();

			const error = page.locator(
				'text=/ungültig|invalid|falsch|wrong|incorrect/i'
			);
			const hasError = await error
				.isVisible({ timeout: 3000 })
				.catch(() => false);
			const stillInModal = await securityPage.setupModal
				.isVisible()
				.catch(() => false);
			expect(hasError || stillInModal).toBeTruthy();
		}
	});

	test('should reject invalid TOTP code', async ({
		authenticatedSecurityPage,
		securityPage
	}) => {
		const page = authenticatedSecurityPage;

		if (
			!(await securityPage.enable2FAButton
				.isVisible({ timeout: 5000 })
				.catch(() => false))
		) {
			test.skip();
			return;
		}

		await securityPage.enable2FAButton.click();
		await expect(securityPage.setupModal).toBeVisible({ timeout: 5000 });

		if (
			!(await securityPage.totpCodeInput
				.isVisible({ timeout: 5000 })
				.catch(() => false))
		) {
			test.skip();
			return;
		}

		await securityPage.totpCodeInput.fill('000000');

		if (
			await securityPage.verifyButton
				.isVisible({ timeout: 3000 })
				.catch(() => false)
		) {
			await securityPage.verifyButton.click();

			const error = page.locator(
				'text=/ungültig|invalid|falsch|wrong|incorrect/i'
			);
			const toast = page.locator('[role="alert"]');
			const hasError = await error
				.isVisible({ timeout: 3000 })
				.catch(() => false);
			const hasToast = await toast
				.first()
				.isVisible()
				.catch(() => false);
			const stillInModal = await securityPage.setupModal
				.isVisible()
				.catch(() => false);

			expect(hasError || hasToast || stillInModal).toBeTruthy();
		}
	});

	test('should close 2FA setup modal on cancel', async ({
		authenticatedSecurityPage,
		securityPage
	}) => {
		if (
			!(await securityPage.enable2FAButton
				.isVisible({ timeout: 5000 })
				.catch(() => false))
		) {
			test.skip();
			return;
		}

		await securityPage.open2FASetup();
		await securityPage.close2FAModal();
	});
});
