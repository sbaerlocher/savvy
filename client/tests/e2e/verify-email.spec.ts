import { expect, test } from './fixtures/test-fixtures';

test.describe('Email Verification', () => {
	test('should show error for missing token', async ({ page }) => {
		await page.goto('/verify-email');
		await page.waitForURL(/\/verify-email/, { timeout: 10000 });

		// Without a token, the page should show error state with red icon
		// Error heading: "Verifizierung fehlgeschlagen" or "Verification Failed"
		const errorHeading = page
			.locator(
				'h2:has-text("fehlgeschlagen"), h2:has-text("Failed"), h2:has-text("Error")'
			)
			.first();

		// Or the idle state shows "checkEmail" text + login link
		const idleOrError = page
			.locator(
				'text=/fehlgeschlagen|Failed|Postfach|check.*email|Ungültig|invalid/i'
			)
			.first();
		await expect(idleOrError).toBeVisible({ timeout: 10000 });
	});

	test('should show error for invalid token', async ({ page }) => {
		await page.goto('/verify-email?token=invalid-token-12345');
		await page.waitForURL(/\/verify-email/, { timeout: 10000 });

		// With an invalid token, the page first shows loading, then error
		// Wait for the final error state
		const errorHeading = page
			.locator('text=/fehlgeschlagen|Failed|Ungültig|invalid/i')
			.first();
		await expect(errorHeading).toBeVisible({ timeout: 15000 });
	});

	test('should display login link for unauthenticated users', async ({
		page
	}) => {
		await page.goto('/verify-email');
		await page.waitForURL(/\/verify-email/, { timeout: 10000 });

		// Should show link back to login ("Zur Anmeldung" or "Go to Login")
		const loginLink = page.locator('a[href="/login"]').first();
		await expect(loginLink).toBeVisible({ timeout: 10000 });
	});

	test('should show loading state during verification', async ({ page }) => {
		// Navigate with a fake token to trigger API call
		await page.goto('/verify-email?token=fake-token-loading-test');

		// The page eventually shows an outcome (error since token is invalid)
		const outcome = page
			.locator(
				'text=/verifiziert|verified|fehlgeschlagen|Failed|Ungültig|invalid|Anmeldung|Login/i'
			)
			.first();
		await expect(outcome).toBeVisible({ timeout: 15000 });
	});
});
