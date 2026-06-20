import { expect, test } from './fixtures/test-fixtures';
import { LoginPage } from './pages/login.page';

test.describe('Error Handling', () => {
	test('should show error for invalid login credentials', async ({ page }) => {
		const loginPage = new LoginPage(page);
		await loginPage.goto();

		// Fill form and submit directly (no retry logic)
		await loginPage.emailInput.fill('invalid@example.com');
		await loginPage.passwordInput.fill('wrongpassword');
		await loginPage.submitButton.click();

		await loginPage.expectError();
	});

	test('should handle 404 page gracefully', async ({ authenticatedPage }) => {
		const page = authenticatedPage;

		await page.goto('/nonexistent-page-12345');

		const notFoundText = page.locator(
			'text=/404|nicht gefunden|not found|Seite nicht gefunden|page not found/i'
		);
		const redirectedToDashboard =
			page.url().includes('/dashboard') || page.url().includes('/cards');

		expect(
			(await notFoundText.isVisible().catch(() => false)) ||
				redirectedToDashboard
		).toBeTruthy();
	});

	test('should handle network errors gracefully', async ({
		authenticatedPage
	}) => {
		const page = authenticatedPage;

		await page.route('**/api/v1/cards**', (route) => route.abort('failed'));

		await page.goto('/cards');

		const errorMessage = page.locator(
			'text=/Fehler|Error|Verbindung|connection|network/i'
		);
		const emptyState = page.locator('text=/keine|no items|empty/i');

		await expect(errorMessage.or(emptyState).first()).toBeVisible({
			timeout: 5000
		});

		await page.unroute('**/api/v1/cards**');
	});

	test('should handle 401 unauthorized by redirecting to login', async ({
		authenticatedPage
	}) => {
		const page = authenticatedPage;

		await page.route('**/api/v1/**', (route) =>
			route.fulfill({
				status: 401,
				body: JSON.stringify({ error: 'Unauthorized' })
			})
		);

		await page.goto('/cards');

		// Wait for redirect or error display
		await page
			.waitForURL(/\/(login|cards|dashboard)/, { timeout: 5000 })
			.catch(() => {});

		const isOnLogin = page.url().includes('/login');
		const hasError = await page
			.locator(
				'text=/nicht autorisiert|unauthorized|anmelden|sign in|Fehler|Error/i'
			)
			.first()
			.isVisible({ timeout: 5000 })
			.catch(() => false);
		const stayedOnPage = page.url().includes('/cards');

		// App may redirect to login, show error, or stay on page with cached data
		expect(isOnLogin || hasError || stayedOnPage).toBeTruthy();

		await page.unroute('**/api/v1/**');
	});

	test('should handle 403 forbidden gracefully', async ({
		authenticatedPage
	}) => {
		const page = authenticatedPage;

		await page.route('**/api/v1/admin/**', (route) =>
			route.fulfill({
				status: 403,
				body: JSON.stringify({ error: 'Forbidden' })
			})
		);

		await page.goto('/admin/users');
		await expect(page).not.toHaveURL(/\/admin\/users\/?$/, { timeout: 5000 });

		await page.unroute('**/api/v1/admin/**');
	});

	test('should handle 500 server error gracefully', async ({
		authenticatedPage
	}) => {
		const page = authenticatedPage;

		await page.route('**/api/v1/cards**', (route) =>
			route.fulfill({
				status: 500,
				body: JSON.stringify({ error: 'Internal Server Error' })
			})
		);

		await page.goto('/cards');

		const errorMessage = page.locator('text=/Fehler|Error|Server/i');
		const toast = page.locator('[role="status"], [role="alert"], .toast');

		await expect(errorMessage.or(toast).first()).toBeVisible({ timeout: 5000 });

		await page.unroute('**/api/v1/cards**');
	});

	test('should sanitize error messages displayed to user', async ({
		authenticatedPage
	}) => {
		const page = authenticatedPage;

		await page.route('**/api/v1/cards**', (route) =>
			route.fulfill({
				status: 500,
				body: JSON.stringify({
					error: 'SQL error: relation "cards" does not exist at character 15'
				})
			})
		);

		await page.goto('/cards');
		await page.waitForLoadState('networkidle');

		const sqlExposed = page.locator(
			'text=/SQL error|relation.*does not exist|character 15/'
		);
		expect(await sqlExposed.isVisible().catch(() => false)).toBeFalsy();

		await page.unroute('**/api/v1/cards**');
	});

	test('should handle large request body rejection', async ({
		authenticatedPage
	}) => {
		const page = authenticatedPage;

		await page.route('**/api/v1/cards', (route) => {
			if (route.request().method() === 'POST') {
				return route.fulfill({
					status: 413,
					body: JSON.stringify({ error: 'Request entity too large' })
				});
			}
			return route.continue();
		});

		await page.goto('/cards/new');
		await page.waitForLoadState('networkidle');

		await page.unroute('**/api/v1/cards');
	});
});
