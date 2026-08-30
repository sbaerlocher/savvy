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

		// Where the type is picked differs per chrome: desktop has no chip row at
		// all (WalletView.svelte:1204 gates it on `!DESKTOP_CHROME`) and offers the
		// type only inside the filter panel, while Android/iOS render the chip row
		// above the toolbar. Both render TypeFilterButtons, which stamps a stable
		// `data-testid` on every chip (TypeFilterButtons.svelte:130/151), so open
		// the filter panel only when the chip is not already on screen.
		const voucherChip = page
			.locator('[data-testid="type-chip-vouchers"]')
			.filter({ visible: true })
			.first();

		if (!(await voucherChip.isVisible().catch(() => false))) {
			// The filter control exists twice — `hidden sm:flex` desktop actions
			// come first in DOM order, so a plain .first() picks the display:none
			// copy on a 393px viewport. getByRole reads the accessibility tree,
			// which drops display:none subtrees, so it lands on the rendered one.
			await page
				.getByRole('button', { name: /^Filter/i })
				.filter({ visible: true })
				.first()
				.click();
		}

		await voucherChip.click();

		await expect(page).toHaveURL(/\/wallet/);
		// The chip must actually engage the filter, not just be clickable.
		await expect(voucherChip).toHaveAttribute('aria-pressed', 'true');
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

		// The tiles reveal the barcode inline (canvas) or, on the compact mobile
		// tiles, as a per-tile "show barcode" button — accept either signal.
		const inlineBarcode = page
			.locator('canvas')
			.filter({ visible: true })
			.first();
		const tileBarcodeButton = page
			.getByRole('button', { name: /Barcode anzeigen|Show barcode/i })
			.first();
		await expect(inlineBarcode.or(tileBarcodeButton).first()).toBeVisible({
			timeout: 10000
		});
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
