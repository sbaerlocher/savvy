/**
 * Logs in once and stores the session, so the ~100 baseline captures do not
 * each pay for a login round-trip (and do not each depend on the login flow
 * still working).
 */
import { expect, test as setup } from '@playwright/test';

const ADMIN_STATE = 'tests/structure/.auth/admin.json';

setup('authenticate as admin', async ({ page }) => {
	await page.goto('/login');
	await page.locator('input[name="email"]').fill('admin@example.com');
	await page.locator('input[name="password"]').fill('test123');
	await page.locator('button[type="submit"]').click();
	await page.waitForURL((url) => !url.pathname.includes('/login'), {
		timeout: 15000
	});

	// Prove the session is real before freezing it — an unauthenticated state
	// file would silently baseline ~100 login redirects instead of the app.
	await page.goto('/dashboard');
	await expect(page.locator('main')).toBeVisible();
	expect(page.url()).not.toContain('/login');

	await page.context().storageState({ path: ADMIN_STATE });
});
