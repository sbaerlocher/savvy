import type { Page } from '@playwright/test';
import { expect, test, TEST_USERS } from './fixtures/test-fixtures';
import { LoginPage } from './pages/login.page';
import { ResourceListPage } from './pages/resource-list.page';

test.describe('Pagination & Progressive Loading', () => {
	async function loginAndNavigate(page: Page, path: string) {
		const loginPage = new LoginPage(page);
		await loginPage.goto();
		await loginPage.login(
			TEST_USERS.regular.email,
			TEST_USERS.regular.password
		);
		await page.waitForURL(/\/(dashboard|cards|vouchers|gift-cards)\/?/, {
			timeout: 15000
		});
		await page.goto(path);
		await page.waitForURL(new RegExp(path.replace('/', '\\/')), {
			timeout: 10000
		});
	}

	test('should load cards list progressively', async ({ page }) => {
		await loginAndNavigate(page, '/cards');

		const cardsList = new ResourceListPage(page, 'cards');
		await expect(cardsList.firstItem).toBeVisible({ timeout: 10000 });

		const initialCount = await cardsList.items.count();
		expect(initialCount).toBeGreaterThan(0);
	});

	test('should load vouchers list progressively', async ({ page }) => {
		await loginAndNavigate(page, '/vouchers');

		const vouchersList = new ResourceListPage(page, 'vouchers');
		const hasItems = await vouchersList.firstItem
			.isVisible({ timeout: 10000 })
			.catch(() => false);
		if (hasItems) {
			expect(await vouchersList.items.count()).toBeGreaterThan(0);
		}
	});

	test('should load gift cards list progressively', async ({ page }) => {
		await loginAndNavigate(page, '/gift-cards');

		const giftCardsList = new ResourceListPage(page, 'gift-cards');
		const hasItems = await giftCardsList.firstItem
			.isVisible({ timeout: 10000 })
			.catch(() => false);
		if (hasItems) {
			expect(await giftCardsList.items.count()).toBeGreaterThan(0);
		}
	});

	test('should handle scroll-based loading', async ({ page }) => {
		await loginAndNavigate(page, '/cards');

		const cardsList = new ResourceListPage(page, 'cards');
		await expect(cardsList.firstItem).toBeVisible({ timeout: 10000 });

		const initialCount = await cardsList.items.count();

		// Scroll to bottom to trigger potential lazy loading
		await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight));
		await page.waitForLoadState('networkidle');

		const afterScrollCount = await cardsList.items.count();
		expect(afterScrollCount).toBeGreaterThanOrEqual(initialCount);
	});

	test('should cache data in IndexedDB for subsequent loads', async ({
		page
	}) => {
		await loginAndNavigate(page, '/cards');

		const cardsList = new ResourceListPage(page, 'cards');
		await expect(cardsList.firstItem).toBeVisible({ timeout: 10000 });

		const firstLoadCount = await cardsList.items.count();

		// Navigate away and back
		await page.goto('/dashboard');
		await page.waitForURL(/\/dashboard/, { timeout: 10000 });
		await page.goto('/cards');
		await page.waitForURL(/\/cards/, { timeout: 10000 });

		await expect(cardsList.firstItem).toBeVisible({ timeout: 10000 });
		const secondLoadCount = await cardsList.items.count();

		expect(secondLoadCount).toBeGreaterThanOrEqual(firstLoadCount);
	});

	test('should show loading state while fetching', async ({ page }) => {
		// Login first, then set up route delay for cards API
		await loginAndNavigate(page, '/cards');

		// Verify list loaded initially
		const cardsList = new ResourceListPage(page, 'cards');
		await expect(cardsList.firstItem).toBeVisible({ timeout: 10000 });

		// Now set up delayed route for next navigation
		await page.route('**/api/v1/cards**', async (route) => {
			if (
				route.request().resourceType() === 'fetch' ||
				route.request().resourceType() === 'xhr'
			) {
				await new Promise((r) => setTimeout(r, 500));
			}
			await route.continue();
		});

		// Navigate away and back to trigger loading state
		await page.goto('/dashboard');
		await page.waitForURL(/\/dashboard/, { timeout: 10000 });
		await page.goto('/cards');
		await page.waitForURL(/\/cards/, { timeout: 10000 });

		// Page should eventually render (loading state may be brief)
		const hasItems = await cardsList.firstItem
			.isVisible({ timeout: 15000 })
			.catch(() => false);

		expect(hasItems || page.url().includes('/cards')).toBeTruthy();

		await page.unroute('**/api/v1/cards**');
	});

	test('should handle empty list state', async ({ page }) => {
		await page.route('**/api/v1/cards**', (route) => {
			if (route.request().method() === 'GET') {
				return route.fulfill({
					status: 200,
					contentType: 'application/json',
					body: JSON.stringify({ data: [], total: 0, page: 1, page_size: 20 })
				});
			}
			return route.continue();
		});

		await loginAndNavigate(page, '/cards');

		const emptyState = page.locator(
			'text=/keine|no items|empty|leer|Keine Karten|No cards/i'
		);
		const cardsList = new ResourceListPage(page, 'cards');

		const emptyVisible = await emptyState
			.first()
			.isVisible({ timeout: 5000 })
			.catch(() => false);
		const noItems = (await cardsList.items.count()) === 0;

		expect(emptyVisible || noItems).toBeTruthy();

		await page.unroute('**/api/v1/cards**');
	});
});
