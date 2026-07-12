import { expect, test } from './fixtures/test-fixtures';

// The unified /wallet overview lists cards, vouchers and gift cards together
// with a type filter. The legacy /cards /vouchers /gift-cards routes redirect
// here with a matching ?type= query.
test.describe('Wallet Overview', () => {
	test.beforeEach(async ({ authenticatedPage }) => {
		// authenticatedPage fixture handles login
	});

	const items = (page: import('@playwright/test').Page) =>
		page.locator('[data-owner]');

	async function gotoWallet(page: import('@playwright/test').Page) {
		await page.goto('/wallet', { waitUntil: 'domcontentloaded' });
		await expect(page.locator('h1').first()).toContainText(/Wallet/i);
		// Wait for tiles to render rather than a specific API call: the
		// offline-first loader may serve cached data without a fresh request.
		await expect(items(page).first()).toBeVisible({ timeout: 15000 });
	}

	test('should display a mixed list of cards, vouchers and gift cards', async ({
		authenticatedPage
	}) => {
		const page = authenticatedPage;
		await gotoWallet(page);

		// The seeded account owns items of all three types, so the unfiltered
		// wallet shows more entries than any single type would.
		await expect(items(page).first()).toBeVisible({ timeout: 10000 });
		const total = await items(page).count();
		expect(total).toBeGreaterThan(0);
	});

	test('should filter by type and update the count', async ({
		authenticatedPage
	}) => {
		const page = authenticatedPage;
		await gotoWallet(page);
		const totalCount = await items(page).count();

		// Open the filter panel and pick the vouchers type.
		await page
			.locator('button[aria-label*="Filter" i], button:has-text("Filter")')
			.first()
			.click();
		await page
			.getByRole('button', { name: /Vouchers|Gutscheine/i })
			.first()
			.click();

		await expect(page).toHaveURL(/\/wallet/);
		const vouchersCount = await items(page).count();
		expect(vouchersCount).toBeGreaterThan(0);
		expect(vouchersCount).toBeLessThanOrEqual(totalCount);
	});

	test('should toggle barcodes on the tiles', async ({ authenticatedPage }) => {
		const page = authenticatedPage;
		await gotoWallet(page);

		const toggle = page
			.getByRole('button', { name: /Barcodes? (ein|aus)blenden|barcodes?/i })
			.first();
		await expect(toggle).toBeVisible();
		await toggle.click();

		// Barcode canvases render inside the tiles once the toggle is on.
		await expect(page.locator('canvas, svg[role="img"], img').first()).toBeVisible(
			{ timeout: 10000 }
		);
	});

	test('should redirect legacy list routes to the wallet with a type filter', async ({
		authenticatedPage
	}) => {
		const page = authenticatedPage;

		await page.goto('/cards', { waitUntil: 'domcontentloaded' });
		await expect(page).toHaveURL(/\/wallet\?type=cards/);

		await page.goto('/vouchers', { waitUntil: 'domcontentloaded' });
		await expect(page).toHaveURL(/\/wallet\?type=vouchers/);

		await page.goto('/gift-cards', { waitUntil: 'domcontentloaded' });
		await expect(page).toHaveURL(/\/wallet\?type=gift-cards/);
	});

	test('should keep detail routes reachable', async ({ authenticatedPage }) => {
		const page = authenticatedPage;
		await gotoWallet(page);

		await items(page).first().click();
		// Tiles link to the type-specific detail route, which stays unchanged.
		await expect(page).toHaveURL(
			/\/(cards|vouchers|gift-cards)\/[a-f0-9-]+$/
		);
	});
});
