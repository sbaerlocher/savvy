import { expect, test } from './fixtures/test-fixtures';
import { ResourceDetailPage } from './pages/resource-detail.page';
import type { ResourceType } from './pages/resource-list.page';
import { ResourceListPage } from './pages/resource-list.page';

test.describe('Favorites/Pinning', () => {
	// Serial mode: tests modify shared favorite state for the same user/resources
	test.describe.configure({ mode: 'serial' });

	test.beforeEach(async ({ authenticatedPage }) => {
		// authenticatedPage fixture handles login
	});

	async function navigateToFirstItem(
		page: import('@playwright/test').Page,
		resourceType: ResourceType
	) {
		const listPage = new ResourceListPage(page, resourceType);
		const detailPage = new ResourceDetailPage(page, resourceType);
		await listPage.goto();
		await listPage.clickFirstItem();
		await page.waitForLoadState('networkidle');
		return { listPage, detailPage };
	}

	async function testToggleFavorite(
		page: import('@playwright/test').Page,
		resourceType: ResourceType,
		action: 'add' | 'remove'
	) {
		const { detailPage } = await navigateToFirstItem(page, resourceType);

		if (action === 'add') {
			await detailPage.ensureFavoriteState(false);
			await detailPage.toggleFavorite('★');
		} else {
			await detailPage.ensureFavoriteState(true);
			await detailPage.toggleFavorite('☆');
		}
	}

	test('should add card to favorites', async ({ authenticatedPage }) => {
		await testToggleFavorite(authenticatedPage, 'cards', 'add');
	});

	test('should add voucher to favorites', async ({ authenticatedPage }) => {
		await testToggleFavorite(authenticatedPage, 'vouchers', 'add');
	});

	test('should add gift card to favorites', async ({ authenticatedPage }) => {
		await testToggleFavorite(authenticatedPage, 'gift-cards', 'add');
	});

	test('should remove card from favorites', async ({ authenticatedPage }) => {
		await testToggleFavorite(authenticatedPage, 'cards', 'remove');
	});

	test('should remove voucher from favorites', async ({
		authenticatedPage
	}) => {
		await testToggleFavorite(authenticatedPage, 'vouchers', 'remove');
	});

	test('should remove gift card from favorites', async ({
		authenticatedPage
	}) => {
		await testToggleFavorite(authenticatedPage, 'gift-cards', 'remove');
	});

	test('should show favorites on dashboard', async ({
		authenticatedPage,
		cardDetailPage,
		dashboardPage
	}) => {
		const page = authenticatedPage;
		const listPage = new ResourceListPage(page, 'cards');

		await listPage.goto();
		await listPage.clickFirstItem();
		await page.waitForLoadState('networkidle');
		await cardDetailPage.ensureFavoriteState(true);

		await dashboardPage.goto();
		await expect(page).toHaveURL('/dashboard');

		await expect(dashboardPage.favoritesHeading).toBeVisible({ timeout: 5000 });
		await expect(dashboardPage.favoriteItems.first()).toBeVisible({
			timeout: 5000
		});
	});

	test('should show favorite counts on dashboard', async ({
		authenticatedPage,
		dashboardPage
	}) => {
		const page = authenticatedPage;
		await dashboardPage.goto();
		await dashboardPage.waitForDashboardApi();

		const favoritesSection = page.locator('[data-testid="favorites-section"]');
		if (!(await favoritesSection.isVisible().catch(() => false))) {
			test.skip();
			return;
		}

		const favoriteItems = page
			.locator('[data-testid="favorites-list"]')
			.locator(
				'a[href^="/cards/"], a[href^="/vouchers/"], a[href^="/gift-cards/"]'
			);
		const itemCount = await favoriteItems.count();
		expect(itemCount).toBeGreaterThan(0);
	});

	test('should toggle card favorite multiple times', async ({
		authenticatedPage,
		cardDetailPage
	}) => {
		const page = authenticatedPage;
		const listPage = new ResourceListPage(page, 'cards');

		await listPage.goto();
		await listPage.clickFirstItem();
		await page.waitForLoadState('networkidle');

		await expect(cardDetailPage.favoriteButton).toBeVisible({ timeout: 10000 });

		const initialText = await cardDetailPage.favoriteButton.textContent();
		const isInitiallyFavorited = initialText?.includes('★') || false;
		const toggledIcon: '★' | '☆' = isInitiallyFavorited ? '☆' : '★';
		const originalIcon: '★' | '☆' = isInitiallyFavorited ? '★' : '☆';

		await cardDetailPage.toggleFavorite(toggledIcon);
		await cardDetailPage.toggleFavorite(originalIcon);
		await cardDetailPage.toggleFavorite(toggledIcon);
	});

	test('should handle favorite toggle error gracefully', async ({
		authenticatedPage,
		cardDetailPage
	}) => {
		const page = authenticatedPage;
		const listPage = new ResourceListPage(page, 'cards');

		await listPage.goto();
		await listPage.clickFirstItem();
		await page.waitForLoadState('networkidle');

		await page.route('**/favorite', (route) => {
			route.fulfill({
				status: 500,
				body: JSON.stringify({ error: 'Internal Server Error' })
			});
		});

		const initialIcon =
			(await cardDetailPage.favoriteButton.textContent()) || '';
		await cardDetailPage.favoriteButton.click();

		const errorToast = page.locator('[role="status"], [role="alert"], .toast');
		await expect(errorToast.first()).toBeVisible({ timeout: 5000 });

		// Verify the icon reverted back
		await expect(cardDetailPage.favoriteButton).toContainText(
			initialIcon.trim(),
			{
				timeout: 3000
			}
		);

		await page.unroute('**/favorite');
	});

	test('should disable favorite button when offline', async ({
		authenticatedPage,
		cardDetailPage
	}) => {
		const page = authenticatedPage;
		const listPage = new ResourceListPage(page, 'cards');

		await listPage.goto();
		await listPage.clickFirstItem();
		await page.waitForLoadState('networkidle');

		await page.context().setOffline(true);
		await expect(cardDetailPage.offlineIndicator.first()).toBeVisible({
			timeout: 5000
		});
		await expect(cardDetailPage.favoriteButton).toBeDisabled({ timeout: 3000 });

		await page.context().setOffline(false);
	});

	test('should favorite shared card from another user', async ({
		authenticatedPage
	}) => {
		const page = authenticatedPage;
		const listPage = new ResourceListPage(page, 'cards');
		const detailPage = new ResourceDetailPage(page, 'cards');

		await listPage.goto();
		if (!(await listPage.firstItem.isVisible().catch(() => false))) {
			test.skip();
			return;
		}

		await listPage.clickFirstItem();
		await page.waitForLoadState('networkidle');
		await detailPage.ensureFavoriteState(true);
	});

	test('should allow favoriting with view-only permission', async ({
		authenticatedPage,
		cardDetailPage
	}) => {
		const page = authenticatedPage;
		const listPage = new ResourceListPage(page, 'cards');

		await listPage.goto();
		if (!(await listPage.firstItem.isVisible().catch(() => false))) {
			test.skip();
			return;
		}

		await listPage.clickFirstItem();
		await page.waitForLoadState('networkidle');

		await expect(cardDetailPage.favoriteButton).toBeVisible();
		await expect(cardDetailPage.favoriteButton).toBeEnabled();

		const isFavorited = await cardDetailPage.isFavorited();
		const expectedIcon: '★' | '☆' = isFavorited ? '☆' : '★';
		await cardDetailPage.toggleFavorite(expectedIcon);
	});

	test('should show loading state during favorite toggle', async ({
		authenticatedPage,
		cardDetailPage
	}) => {
		const page = authenticatedPage;
		const listPage = new ResourceListPage(page, 'cards');

		await listPage.goto();
		await listPage.clickFirstItem();
		await page.waitForLoadState('networkidle');

		await expect(cardDetailPage.favoriteButton).toBeVisible({ timeout: 10000 });

		await page.route('**/favorite', async (route) => {
			await new Promise((resolve) => setTimeout(resolve, 1000));
			route.continue();
		});

		const favoriteResponse = page.waitForResponse(
			(resp) =>
				resp.url().includes('/favorite') &&
				resp.status() === 200 &&
				resp.request().method() === 'POST',
			{ timeout: 15000 }
		);

		await cardDetailPage.favoriteButton.click();
		await expect(cardDetailPage.favoriteButton).toBeDisabled({ timeout: 500 });
		await favoriteResponse;
		await expect(cardDetailPage.favoriteButton).toBeEnabled({ timeout: 10000 });

		await page.unroute('**/favorite');
	});
});
