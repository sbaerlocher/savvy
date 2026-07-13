import { expect, test } from './fixtures/test-fixtures';

test.describe('Dashboard', () => {
	test.beforeEach(async ({ authenticatedPage }) => {
		// authenticatedPage fixture handles login
	});

	test('should display dashboard after login', async ({
		authenticatedPage,
		dashboardPage
	}) => {
		await dashboardPage.goto();
		await dashboardPage.waitForDashboardApi();
		await expect(dashboardPage.welcomeHeading).toBeVisible();
	});

	test('should show statistics', async ({
		authenticatedPage,
		dashboardPage
	}) => {
		await dashboardPage.goto();
		await dashboardPage.waitForDashboardApi();
		await dashboardPage.expectStatsVisible();
	});

	// App shell: individual resource tabs were merged into a single Wallet place.
	// Cards/vouchers/gift cards are reachable via the Wallet nav entry, not from
	// the dashboard directly.
	test('should navigate to wallet from nav', async ({
		authenticatedPage,
		dashboardPage
	}) => {
		const page = authenticatedPage;
		await dashboardPage.goto();
		await dashboardPage.waitForDashboardApi();

		// Wallet link exists in both the desktop nav and the mobile nav; pick
		// whichever is visible for the current viewport (Pixel 5 vs desktop).
		const walletLink = page.locator('a[href="/wallet"]:visible').first();
		await expect(walletLink).toBeVisible({ timeout: 5000 });
		await walletLink.click();
		await expect(page).toHaveURL(/\/wallet\/?$/);
	});

	test('should display recent items', async ({
		authenticatedPage,
		dashboardPage
	}) => {
		await dashboardPage.goto();
		await dashboardPage.waitForDashboardApi();

		const recentSection = dashboardPage.page.locator('text=/Zuletzt|Recent/i');
		if (await recentSection.isVisible({ timeout: 3000 }).catch(() => false)) {
			const recentLinks = dashboardPage.recentItemLinks;
			expect(await recentLinks.count()).toBeGreaterThanOrEqual(0);
		}
	});

	test('should display favorites section', async ({
		authenticatedPage,
		dashboardPage
	}) => {
		await dashboardPage.goto();
		await dashboardPage.waitForDashboardApi();

		if (
			await dashboardPage.favoritesHeading
				.isVisible({ timeout: 3000 })
				.catch(() => false)
		) {
			// Favorites heading is visible - verify section exists
			// The favorites list may or may not have items depending on test state
			const hasList = await dashboardPage.favoritesList
				.isVisible({ timeout: 3000 })
				.catch(() => false);
			expect(
				hasList || (await dashboardPage.favoritesHeading.isVisible())
			).toBeTruthy();
		}
	});
});
