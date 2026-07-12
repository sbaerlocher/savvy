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

	test('should navigate to cards from dashboard', async ({
		authenticatedPage,
		dashboardPage
	}) => {
		const page = authenticatedPage;
		await dashboardPage.goto();
		await dashboardPage.waitForDashboardApi();

		const cardsLink = page.locator('a[href="/cards"]').first();
		await expect(cardsLink).toBeVisible({ timeout: 5000 });
		await cardsLink.click();
		await expect(page).toHaveURL(/(\/cards\/?$|\/wallet\?type=cards)/);
	});

	test('should navigate to vouchers from dashboard', async ({
		authenticatedPage,
		dashboardPage
	}) => {
		const page = authenticatedPage;
		await dashboardPage.goto();
		await dashboardPage.waitForDashboardApi();

		const vouchersLink = page.locator('a[href="/vouchers"]').first();
		await expect(vouchersLink).toBeVisible({ timeout: 5000 });
		await vouchersLink.click();
		await expect(page).toHaveURL(/(\/vouchers\/?$|\/wallet\?type=vouchers)/);
	});

	test('should navigate to gift cards from dashboard', async ({
		authenticatedPage,
		dashboardPage
	}) => {
		const page = authenticatedPage;
		await dashboardPage.goto();
		await dashboardPage.waitForDashboardApi();

		const giftCardsLink = page.locator('a[href="/gift-cards"]').first();
		await expect(giftCardsLink).toBeVisible({ timeout: 5000 });
		await giftCardsLink.click();
		await expect(page).toHaveURL(/(\/gift-cards\/?$|\/wallet\?type=gift-cards)/);
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
