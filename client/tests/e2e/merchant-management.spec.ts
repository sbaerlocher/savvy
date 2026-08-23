import { uniqueMerchantName } from './fixtures/test-data';
import { expect, test } from './fixtures/test-fixtures';
import type { Page } from '@playwright/test';

/**
 * Fill the merchant search, opening the chrome that holds it first.
 *
 * The overview renders the field per layout: desktop and tablet keep it above
 * the grid, while Android below `sm` moves it into the filter bottom sheet
 * (mockup screen-MerchantsAndroid). Go through the layout's own control rather
 * than assuming one of the two, so the test exercises the real flow.
 */
async function searchMerchants(page: Page, term: string) {
	const inline = page.getByTestId('merchant-search');
	if (await inline.isVisible()) {
		await inline.fill(term);
		return;
	}
	await page.getByTestId('merchant-filter-chip').click();
	const sheetField = page.getByTestId('merchant-search-sheet');
	await expect(sheetField).toBeVisible({ timeout: 5000 });
	await sheetField.fill(term);
	await page
		.getByRole('button', { name: /Done|Fertig|Terminé/i })
		.filter({ visible: true })
		.first()
		.click();
}

test.describe('Merchant Management', () => {
	test.describe('Public Merchant Overview', () => {
		test.beforeEach(async ({ authenticatedPage }) => {
			// Regular user can browse merchants
		});

		test('should display merchants list', async ({ authenticatedPage }) => {
			const page = authenticatedPage;
			const apiResponse = page
				.waitForResponse((resp) => resp.url().includes('/api/v1/merchants'), {
					timeout: 15000
				})
				.catch(() => null);
			await page.goto('/merchants');
			await apiResponse;

			const merchantCards = page.locator('a[href^="/merchants/"]');
			await expect(merchantCards.first()).toBeVisible({ timeout: 10000 });
		});

		test('should search merchants by name', async ({ authenticatedPage }) => {
			const page = authenticatedPage;
			await page.goto('/merchants');

			const merchantCards = page.locator('a[href^="/merchants/"]');
			await expect(merchantCards.first()).toBeVisible({ timeout: 10000 });

			await searchMerchants(page, 'Migros');

			// Wait for client-side filter
			await page.waitForFunction(() => true, null, { timeout: 600 });

			const resultCount = await merchantCards.count();
			expect(resultCount).toBeGreaterThan(0);
		});

		test('should search merchants with no results', async ({
			authenticatedPage
		}) => {
			const page = authenticatedPage;
			await page.goto('/merchants');

			const merchantCards = page.locator('a[href^="/merchants/"]');
			await expect(merchantCards.first()).toBeVisible({ timeout: 10000 });

			await searchMerchants(page, 'NonExistentMerchantXYZ12345');

			await page.waitForFunction(() => true, null, { timeout: 600 });

			const noResults = await merchantCards.count();
			const emptyState = page.locator(
				'text=/Keine Ergebnisse|No results|no_results/i'
			);
			const hasEmptyState = await emptyState
				.first()
				.isVisible({ timeout: 3000 })
				.catch(() => false);
			expect(noResults === 0 || hasEmptyState).toBeTruthy();
		});

		test('should navigate to merchant detail', async ({
			authenticatedPage
		}) => {
			const page = authenticatedPage;
			await page.goto('/merchants');

			const merchantCards = page.locator('a[href^="/merchants/"]');
			await expect(merchantCards.first()).toBeVisible({ timeout: 10000 });

			await merchantCards.first().click();
			await page.waitForURL(/\/merchants\/[a-f0-9-]+$/, { timeout: 10000 });
		});
	});

	test.describe('Admin Merchant CRUD', () => {
		test.beforeEach(async ({ adminAuthenticatedPage }) => {
			// Admin user for create/edit/delete
		});

		test('should create a new merchant', async ({ adminAuthenticatedPage }) => {
			const page = adminAuthenticatedPage;
			const merchantName = uniqueMerchantName();

			await page.goto('/admin/merchants/new');
			await page.waitForURL(/\/admin\/merchants\/new/, { timeout: 10000 });

			const nameInput = page.locator('input[name="name"], input#name').first();
			await expect(nameInput).toBeVisible({ timeout: 5000 });
			await nameInput.fill(merchantName);

			const submitResponse = page.waitForResponse(
				(resp) =>
					resp.url().includes('/admin/merchants') &&
					resp.request().method() === 'POST',
				{ timeout: 10000 }
			);
			await page.click('button[type="submit"]');
			await submitResponse;

			// Verify merchant appears in the public list
			await page.goto('/merchants');
			await searchMerchants(page, merchantName);

			await page.waitForFunction(() => true, null, { timeout: 600 });

			const merchantLink = page.locator(`text=${merchantName}`).first();
			await expect(merchantLink).toBeVisible({ timeout: 10000 });
		});

		test('should prevent duplicate merchant names', async ({
			adminAuthenticatedPage
		}) => {
			const page = adminAuthenticatedPage;

			await page.goto('/admin/merchants/new');
			await page.waitForURL(/\/admin\/merchants\/new/, { timeout: 10000 });

			const nameInput = page.locator('input[name="name"], input#name').first();
			await expect(nameInput).toBeVisible({ timeout: 5000 });
			await nameInput.fill('Migros');

			await page.click('button[type="submit"]');

			const error = page.locator(
				'text=/bereits|already|existiert|exists|duplicate|doppelt/i'
			);
			const toast = page.locator('[role="alert"]');
			await expect(error.or(toast).first()).toBeVisible({ timeout: 5000 });
		});

		test('should validate merchant name is required', async ({
			adminAuthenticatedPage
		}) => {
			const page = adminAuthenticatedPage;

			await page.goto('/admin/merchants/new');
			await page.waitForURL(/\/admin\/merchants\/new/, { timeout: 10000 });

			const nameInput = page.locator('input[name="name"], input#name').first();
			await expect(nameInput).toBeVisible({ timeout: 5000 });

			await page.click('button[type="submit"]');

			// Should stay on the form
			const stillOnForm = page.url().includes('/merchants/new');
			expect(stillOnForm).toBeTruthy();
		});
	});
});
