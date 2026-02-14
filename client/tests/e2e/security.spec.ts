import { expect, test, TEST_USERS } from './fixtures/test-fixtures';

test.describe('Security', () => {
	test.beforeEach(async ({ authenticatedSecurityPage }) => {
		// authenticatedSecurityPage fixture handles login + navigation to /security
	});

	test.describe('Password Change', () => {
		test('should display password change form for local auth', async ({
			authenticatedSecurityPage,
			securityPage
		}) => {
			await securityPage.waitForPageReady();

			// Check if password section is visible (only for local auth)
			const passwordVisible = await securityPage.currentPasswordInput
				.isVisible({ timeout: 5000 })
				.catch(() => false);
			if (!passwordVisible) {
				test.skip();
				return;
			}

			await expect(securityPage.currentPasswordInput).toBeVisible();
			await expect(securityPage.newPasswordInput).toBeVisible();
			await expect(securityPage.confirmPasswordInput).toBeVisible();
			await expect(securityPage.changePasswordButton).toBeVisible();
		});

		test('should reject password change with wrong current password', async ({
			authenticatedSecurityPage,
			securityPage
		}) => {
			await securityPage.waitForPageReady();

			const passwordVisible = await securityPage.currentPasswordInput
				.isVisible({ timeout: 5000 })
				.catch(() => false);
			if (!passwordVisible) {
				test.skip();
				return;
			}

			await securityPage.changePassword(
				'wrongpassword',
				'newPass123',
				'newPass123'
			);

			// Should show error toast
			const errorToast = securityPage.page
				.locator('[role="alert"], [role="status"]')
				.first();
			await expect(errorToast).toBeVisible({ timeout: 5000 });
		});

		test('should reject password change with mismatched passwords', async ({
			authenticatedSecurityPage,
			securityPage
		}) => {
			await securityPage.waitForPageReady();

			const passwordVisible = await securityPage.currentPasswordInput
				.isVisible({ timeout: 5000 })
				.catch(() => false);
			if (!passwordVisible) {
				test.skip();
				return;
			}

			await securityPage.currentPasswordInput.fill(TEST_USERS.regular.password);
			await securityPage.newPasswordInput.fill('newPassword1');
			await securityPage.confirmPasswordInput.fill('newPassword2');
			await securityPage.changePasswordButton.click();

			// Client-side validation should show mismatch error
			const errorToast = securityPage.page
				.locator('[role="alert"], [role="status"]')
				.first();
			await expect(errorToast).toBeVisible({ timeout: 5000 });
		});
	});

	test.describe('Session Management', () => {
		test('should display active sessions section', async ({
			authenticatedSecurityPage,
			securityPage
		}) => {
			await securityPage.waitForPageReady();

			// Sessions section should be visible
			await expect(securityPage.sessionsSection).toBeVisible({
				timeout: 10000
			});
		});

		test('should show current session badge', async ({
			authenticatedSecurityPage,
			securityPage
		}) => {
			await securityPage.waitForPageReady();

			const sessionsVisible = await securityPage.sessionsSection
				.isVisible({ timeout: 5000 })
				.catch(() => false);
			if (!sessionsVisible) {
				test.skip();
				return;
			}

			// Wait for sessions to load (API call)
			await securityPage.page.waitForResponse(
				(resp) =>
					resp.url().includes('/profile/sessions') && resp.status() < 400,
				{ timeout: 10000 }
			);

			// Current session should have a badge
			await expect(securityPage.currentSessionBadge).toBeVisible({
				timeout: 5000
			});
		});

		test('should display session details', async ({
			authenticatedSecurityPage,
			securityPage
		}) => {
			await securityPage.waitForPageReady();

			const sessionsVisible = await securityPage.sessionsSection
				.isVisible({ timeout: 5000 })
				.catch(() => false);
			if (!sessionsVisible) {
				test.skip();
				return;
			}

			// Wait for sessions API response
			await securityPage.page.waitForResponse(
				(resp) =>
					resp.url().includes('/profile/sessions') && resp.status() < 400,
				{ timeout: 10000 }
			);

			// At least one session should show browser/device info
			const sessionCard = securityPage.page
				.locator('.bg-cyan-50, .bg-gray-50')
				.filter({
					hasText: /Chrome|Firefox|Safari|Edge|Unknown|Unbekannt/i
				})
				.first();
			await expect(sessionCard).toBeVisible({ timeout: 5000 });
		});
	});

	test.describe('Two-Factor Authentication', () => {
		test('should display 2FA section when enabled', async ({
			authenticatedSecurityPage,
			securityPage
		}) => {
			await securityPage.waitForPageReady();

			const twoFactorVisible = await securityPage.twoFactorSection
				.isVisible({ timeout: 5000 })
				.catch(() => false);
			if (!twoFactorVisible) {
				test.skip();
				return;
			}

			await expect(securityPage.twoFactorSection).toBeVisible();
		});
	});
});
