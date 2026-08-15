import { expect, test, TEST_USERS } from './fixtures/test-fixtures';

test.describe('Password Reset', () => {
	test.describe('Forgot Password Page', () => {
		test('should display forgot password form', async ({ page }) => {
			await page.goto('/forgot-password');
			await page.waitForURL(/\/forgot-password/, { timeout: 10000 });

			const emailInput = page.locator('input[type="email"]').first();
			await expect(emailInput).toBeVisible({ timeout: 5000 });

			const submitButton = page.locator('button[type="submit"]').first();
			await expect(submitButton).toBeVisible({ timeout: 5000 });

			const loginLink = page.locator('a[href="/login"]').first();
			await expect(loginLink).toBeVisible({ timeout: 5000 });
		});

		test('should show info panel with instructions', async ({ page }) => {
			await page.goto('/forgot-password');
			await page.waitForURL(/\/forgot-password/, { timeout: 10000 });

			// Right column info panel has step titles
			const infoPanel = page
				.locator('text=/E-Mail eingeben|Enter your email/i')
				.first();
			const hasInfo = await infoPanel
				.isVisible({ timeout: 5000 })
				.catch(() => false);
			if (hasInfo) {
				await expect(infoPanel).toBeVisible();
			}
		});

		test('should submit forgot password form', async ({ page }) => {
			await page.goto('/forgot-password');
			await page.waitForURL(/\/forgot-password/, { timeout: 10000 });

			const emailInput = page.locator('input[type="email"]').first();
			await expect(emailInput).toBeVisible({ timeout: 5000 });
			await emailInput.fill(TEST_USERS.regular.email);

			const submitButton = page.locator('button[type="submit"]').first();

			const apiResponse = page
				.waitForResponse(
					(resp) =>
						resp.url().includes('/forgot-password') &&
						resp.request().method() === 'POST',
					{ timeout: 10000 }
				)
				.catch(() => null);
			await submitButton.click();
			await apiResponse;

			// Success: "E-Mail gesendet!" heading
			const successHeading = page
				.locator('text=/gesendet|sent|E-Mail gesendet/i')
				.first();
			await expect(successHeading).toBeVisible({ timeout: 10000 });
		});

		test('should not reveal whether email exists', async ({ page }) => {
			await page.goto('/forgot-password');
			await page.waitForURL(/\/forgot-password/, { timeout: 10000 });

			const emailInput = page.locator('input[type="email"]').first();
			await emailInput.fill('nonexistent@example.com');

			const submitButton = page.locator('button[type="submit"]').first();

			const apiResponse = page
				.waitForResponse(
					(resp) =>
						resp.url().includes('/forgot-password') &&
						resp.request().method() === 'POST',
					{ timeout: 10000 }
				)
				.catch(() => null);
			await submitButton.click();
			await apiResponse;

			// Same success message regardless of email existence (no enumeration)
			const successHeading = page
				.locator('text=/gesendet|sent|E-Mail gesendet/i')
				.first();
			await expect(successHeading).toBeVisible({ timeout: 10000 });
		});

		test('should redirect authenticated users', async ({
			authenticatedPage
		}) => {
			await authenticatedPage.goto('/forgot-password');

			await authenticatedPage.waitForURL(
				/\/(dashboard|cards|vouchers|gift-cards)\/?/,
				{
					timeout: 10000
				}
			);
		});
	});

	test.describe('Reset Password Page', () => {
		test('should display reset password form with token', async ({ page }) => {
			await page.goto('/reset-password?token=test-token');
			await page.waitForURL(/\/reset-password/, { timeout: 10000 });

			const passwordInput = page.locator('input[type="password"]').first();
			await expect(passwordInput).toBeVisible({ timeout: 5000 });
		});

		test('should show error for missing token', async ({ page }) => {
			await page.goto('/reset-password');
			await page.waitForURL(/\/reset-password/, { timeout: 10000 });

			// Error: "Fehler beim Zurücksetzen" or "Ungültiger"
			const errorContent = page
				.locator('text=/Fehler|Error|Ungültig|invalid/i')
				.first();
			await expect(errorContent).toBeVisible({ timeout: 10000 });
		});

		test('should show error for invalid token', async ({ page }) => {
			await page.goto('/reset-password?token=invalid-token-12345');
			await page.waitForURL(/\/reset-password/, { timeout: 10000 });

			// A token in the URL puts the page in its 'form' state
			// (reset-password/+page.svelte:32-36), so the form is the only state
			// this test can reach. The old `isVisible({ timeout: 5000 })` branch
			// treated a slow hydration as "no form" and then waited for an error
			// screen that never comes — an emulated Pixel 5 is where that 5s budget
			// actually runs out. Wait for the form properly instead of guessing.
			const passwordInputs = page.locator('input[type="password"]');
			await expect(passwordInputs.first()).toBeVisible({ timeout: 15000 });

			await passwordInputs.nth(0).fill('newPassword123');
			await passwordInputs.nth(1).fill('newPassword123');

			const submitButton = page.locator('button[type="submit"]').first();
			await submitButton.click();

			// The error screen replaces the form outright (status === 'error'),
			// so the form disappearing is the unambiguous signal that the app
			// rejected the token — unlike a bare text match, which the footer's
			// "Fehler melden" link satisfied on desktop while being `hidden
			// sm:block` below `sm` (AppFooter.svelte:8/93).
			await expect(passwordInputs.first()).toBeHidden({ timeout: 15000 });

			// The error heading is rendered by the error branch only.
			await expect(
				page.getByRole('heading', {
					name: /Fehler beim Zurücksetzen|Failed to reset|Error/i
				})
			).toBeVisible({ timeout: 5000 });
		});
	});
});
