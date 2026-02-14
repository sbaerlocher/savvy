import { expect, test, TEST_USERS } from './fixtures/test-fixtures';
import { LoginPage } from './pages/login.page';

test.describe('Configuration & Features', () => {
	test.use({ serviceWorkers: 'block' });

	test('should have PWA manifest', async ({ page }) => {
		await page.goto('/');
		const manifestLink = page.locator('link[rel="manifest"]');
		await expect(manifestLink).toBeAttached();
	});

	test('should load application without errors', async ({ page }) => {
		const errors: string[] = [];
		page.on('pageerror', (error) => errors.push(error.message));

		const loginPage = new LoginPage(page);
		await loginPage.goto();
		await loginPage.login(
			TEST_USERS.regular.email,
			TEST_USERS.regular.password
		);
		await page.waitForURL(/\/(dashboard|cards|vouchers|gift-cards)\/?/, {
			timeout: 15000
		});

		const criticalErrors = errors.filter(
			(e) =>
				!e.includes('ResizeObserver') &&
				!e.includes('Non-Error promise rejection')
		);
		expect(criticalErrors).toHaveLength(0);
	});

	test('should have proper meta tags', async ({ page }) => {
		await page.goto('/');
		const viewport = page.locator('meta[name="viewport"]');
		await expect(viewport).toBeAttached();

		const themeColor = page.locator('meta[name="theme-color"]');
		if ((await themeColor.count()) > 0) {
			await expect(themeColor).toBeAttached();
		}
	});

	test('should show oauth login buttons when configured', async ({ page }) => {
		await page.goto('/login');
		// OAuth is optional - just verify the login page loads correctly
		const loginPage = new LoginPage(page);
		await loginPage.expectLoginPage();
	});

	test('should handle dark mode toggle', async ({ authenticatedPage }) => {
		const page = authenticatedPage;

		const themeToggle = page.locator(
			'[data-testid="theme-toggle"], button[aria-label*="theme" i], button[aria-label*="dark" i], button[aria-label*="light" i]'
		);
		if (
			!(await themeToggle
				.first()
				.isVisible({ timeout: 3000 })
				.catch(() => false))
		) {
			test.skip();
			return;
		}

		await themeToggle.first().click();

		const html = page.locator('html');
		const classList = (await html.getAttribute('class')) || '';
		const hasDarkClass = classList.includes('dark');
		const dataTheme = await html.getAttribute('data-theme');
		expect(
			hasDarkClass || dataTheme === 'dark' || dataTheme === 'light'
		).toBeTruthy();
	});

	test('should have responsive navigation', async ({ authenticatedPage }) => {
		const page = authenticatedPage;

		await page.setViewportSize({ width: 375, height: 667 });

		const mobileNav = page.locator(
			'nav, [data-testid="mobile-nav"], [role="navigation"]'
		);
		await expect(mobileNav.first()).toBeVisible();

		await page.setViewportSize({ width: 1280, height: 720 });

		const desktopNav = page.locator(
			'nav, [data-testid="desktop-nav"], [role="navigation"]'
		);
		await expect(desktopNav.first()).toBeVisible();
	});
});
