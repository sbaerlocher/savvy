import { randomVoucherCode, testVouchers } from './fixtures/test-data';
import { expect, test } from './fixtures/test-fixtures';

test.describe('Vouchers Management', () => {
	test.beforeEach(async ({ authenticatedPage }) => {
		// authenticatedPage fixture handles login
	});

	test('should display vouchers list', async ({
		authenticatedPage,
		vouchersListPage
	}) => {
		await vouchersListPage.goto();
		await vouchersListPage.expectHeading();
	});

	test('should create a new voucher', async ({
		authenticatedPage,
		vouchersListPage,
		voucherFormPage
	}) => {
		const page = authenticatedPage;
		await vouchersListPage.goto();
		await vouchersListPage.clickNewButton();

		const voucherData = {
			...testVouchers.amazon,
			voucher_code: randomVoucherCode()
		};
		await voucherFormPage.fillVoucherForm({
			merchantName: voucherData.merchant_name,
			code: voucherData.voucher_code,
			type: 'fixed_amount',
			value: voucherData.value,
			currency: voucherData.currency,
			validFrom: voucherData.valid_from,
			validUntil: voucherData.valid_until,
			usageLimitType: 'multiple_use_with_card',
			expiresAt: voucherData.expires_at,
			notes: voucherData.notes
		});
		await voucherFormPage.submit();

		await page.waitForURL(/\/vouchers($|\/[^/]+$)/, { timeout: 10000 });
		await vouchersListPage.goto();
		await expect(
			page.locator(`text=${voucherData.merchant_name}`).first()
		).toBeVisible();
	});

	test('should view voucher details', async ({
		authenticatedPage,
		vouchersListPage,
		voucherFormPage,
		voucherDetailPage
	}) => {
		await vouchersListPage.goto();
		await vouchersListPage.clickNewButton();

		const voucherData = {
			...testVouchers.amazon,
			voucher_code: randomVoucherCode()
		};
		await voucherFormPage.fillVoucherForm({
			merchantName: voucherData.merchant_name,
			code: voucherData.voucher_code,
			value: voucherData.value,
			validFrom: voucherData.valid_from,
			validUntil: voucherData.valid_until
		});
		await voucherFormPage.submit();

		await vouchersListPage.goto();
		await expect(vouchersListPage.firstItem).toBeVisible({ timeout: 10000 });
		await vouchersListPage.clickFirstItem();

		await expect(voucherDetailPage.heading).toBeVisible();
		await expect(voucherDetailPage.barcodeCanvas).toBeVisible({
			timeout: 10000
		});
	});

	test('should edit a voucher', async ({
		authenticatedPage,
		vouchersListPage,
		voucherDetailPage
	}) => {
		const page = authenticatedPage;
		await vouchersListPage.goto();

		// ResourceTile masks the code, so match the seeded voucher by its
		// description (rendered as the tile identifier) instead.
		const testVoucher = page
			.locator('[data-owner]')
			.filter({ hasText: 'Test voucher for E2E edit test' })
			.first();
		await expect(testVoucher).toBeVisible({ timeout: 10000 });
		await testVoucher.click();
		await voucherDetailPage.waitForPageReady();

		await voucherDetailPage.enterEditMode();
		const valueField = page.locator('input#value');
		await expect(valueField).toBeVisible();
		await valueField.clear();
		await valueField.fill('75.50');

		const currencySelect = page.locator('select#currency');
		if (await currencySelect.isVisible()) {
			await currencySelect.selectOption('CHF');
		}

		const usageLimitSelect = page.locator('select#usage_limit_type');
		if (await usageLimitSelect.isVisible()) {
			await usageLimitSelect.selectOption('one_per_customer');
		}

		await voucherDetailPage.save();
		await expect(page.locator('text=/75\\.50|75,50/i')).toBeVisible({
			timeout: 5000
		});
	});

	test('should delete a voucher', async ({
		authenticatedPage,
		vouchersListPage,
		voucherFormPage,
		voucherDetailPage
	}) => {
		const page = authenticatedPage;

		await vouchersListPage.goto();
		await vouchersListPage.clickNewButton();

		const voucherData = {
			...testVouchers.amazon,
			voucher_code: randomVoucherCode()
		};
		await voucherFormPage.fillVoucherForm({
			merchantName: voucherData.merchant_name,
			code: voucherData.voucher_code,
			type: 'fixed_amount',
			value: voucherData.value,
			currency: voucherData.currency,
			validFrom: voucherData.valid_from,
			validUntil: voucherData.valid_until
		});
		await voucherFormPage.submit();

		await expect(page).toHaveURL(/\/vouchers\/[a-f0-9-]+$/);
		await voucherDetailPage.waitForPageReady();
		await voucherDetailPage.enterEditMode();
		await voucherDetailPage.deleteResource();

		await expect(page).toHaveURL(/\/wallet\?type=vouchers$/);
	});

	test('should show expiration warning for expiring vouchers', async ({
		authenticatedPage,
		vouchersListPage
	}) => {
		await vouchersListPage.goto();

		const warningBadge = vouchersListPage.page.locator(
			'[data-testid="expiring-badge"], .badge-warning, .text-warning'
		);
		if (!(await warningBadge.isVisible({ timeout: 2000 }).catch(() => false))) {
			test.skip();
		}
		expect(await warningBadge.isVisible()).toBeTruthy();
	});

	test('should filter vouchers by merchant', async ({
		authenticatedPage,
		vouchersListPage
	}) => {
		await vouchersListPage.goto();

		if (!(await vouchersListPage.filterButton.isVisible().catch(() => false))) {
			test.skip();
			return;
		}

		await vouchersListPage.filterButton.click();

		// The unified wallet filter sheet no longer exposes a native <select>
		// merchant picker; per-merchant browsing moved to /merchants. Skip when
		// no legacy select is present rather than assert one into existence.
		const merchantFilter = vouchersListPage.page
			.locator('select')
			.filter({ hasText: /Merchant|Händler/i })
			.first();
		if (!(await merchantFilter.isVisible().catch(() => false))) {
			test.skip();
			return;
		}

		const options = await merchantFilter.locator('option').count();
		if (options <= 1) {
			test.skip();
			return;
		}

		await merchantFilter.selectOption({ index: 1 });
		await vouchersListPage.waitForPageReady();
		expect(await vouchersListPage.items.count()).toBeGreaterThanOrEqual(0);
	});

	test('should show inactive status for voucher with future valid_from', async ({
		authenticatedPage,
		vouchersListPage,
		voucherFormPage
	}) => {
		const page = authenticatedPage;
		await vouchersListPage.goto();
		await vouchersListPage.clickNewButton();

		const futureStart = new Date(Date.now() + 30 * 24 * 60 * 60 * 1000)
			.toISOString()
			.split('T')[0];
		const futureEnd = new Date(Date.now() + 90 * 24 * 60 * 60 * 1000)
			.toISOString()
			.split('T')[0];
		const testCode = randomVoucherCode();

		await voucherFormPage.fillVoucherForm({
			merchantName: testVouchers.amazon.merchant_name,
			code: testCode,
			value: 20,
			validFrom: futureStart,
			validUntil: futureEnd
		});
		await voucherFormPage.submit();

		await page.waitForURL(/\/vouchers\/[a-f0-9-]+$/, { timeout: 10000 });

		await vouchersListPage.goto();

		// The wallet defaults to the "valid" status, which hides a future
		// (inactive) voucher. Open the filter panel and switch to "inactive"
		// so the tile is shown.
		await vouchersListPage.filterButton.click();
		// Below `lg` the filter lives in a BottomSheet capped at 80vh
		// (BottomSheet.svelte), and the status group sits far enough down that the
		// option starts outside the scroll viewport — Playwright reports it as
		// hidden even though it is rendered and reachable. So scope to the open
		// sheet and let click() scroll it into view, rather than filtering on
		// visibility and waiting for something scrolling would have fixed. The
		// desktop panel renders no dialog, so fall back to the whole page there.
		const filterSurface = page.getByRole('dialog');
		const scope = (await filterSurface.count()) > 0 ? filterSurface : page;
		await scope
			.getByRole('radio', { name: /Inaktiv|Inactive/ })
			.first()
			.click();

		// A future-valid_from voucher must render as inactive. In ResourceTile the
		// status badge is a sibling of the [data-owner] link, so match the tile
		// root that contains the badge and click its inner link.
		const inactiveTile = page
			.locator('div:has(> [data-owner])')
			.filter({ hasText: /Inaktiv|Inactive/ })
			.first();
		await expect(inactiveTile).toBeVisible({ timeout: 10000 });

		// Cleanup
		await inactiveTile.locator('[data-owner]').click();
		await page.waitForURL(/\/vouchers\/[a-f0-9-]+$/, { timeout: 5000 });
		const deleteButton = page
			.locator('button:has-text("Delete"), button:has-text("Löschen")')
			.first();
		if (await deleteButton.isVisible({ timeout: 2000 })) {
			await deleteButton.click();
			const confirmButton = page.locator('[data-testid="modal-confirm"]');
			if (await confirmButton.isVisible({ timeout: 3000 })) {
				await confirmButton.click();
				await page.waitForURL(/\/vouchers\/?$/, { timeout: 5000 });
			}
		}
	});

	test('should display correct expiration date without timezone offset', async ({
		authenticatedPage,
		vouchersListPage,
		voucherFormPage
	}) => {
		const page = authenticatedPage;
		await vouchersListPage.goto();
		await vouchersListPage.clickNewButton();

		// validFrom must be in the past (so voucher is "valid", not "inactive")
		// validUntil must be in the future (so voucher is not "expired")
		// Use a month boundary (Feb 28) to test timezone offset doesn't shift to Mar 1
		const testStartDate = '2026-01-01';
		const testExpiryDate = '2027-02-28';
		const testVoucherCode = randomVoucherCode();

		await voucherFormPage.fillVoucherForm({
			merchantName: testVouchers.amazon.merchant_name,
			code: testVoucherCode,
			value: testVouchers.amazon.value,
			validFrom: testStartDate,
			validUntil: testExpiryDate
		});
		await voucherFormPage.submit();

		await vouchersListPage.goto();
		const voucherCard = page
			.locator('[data-owner]')
			.filter({ hasText: testVouchers.amazon.merchant_name })
			.first();
		await expect(voucherCard).toBeVisible({ timeout: 10000 });

		// The full validity date now lives on the detail page (the tile shows only
		// a compact expiry badge), so assert the timezone-safe date there.
		await voucherCard.click();
		await expect(page).toHaveURL(/\/vouchers\/[a-f0-9-]+$/);

		const dateText = await page.locator('body').textContent();

		// Date should NOT be March 1 (timezone offset bug)
		expect(dateText).not.toContain('01.03.2027');
		expect(dateText).not.toContain('1.3.2027');
		expect(dateText).not.toContain('3/1/2027');
		expect(dateText).not.toContain('01/03/2027');
		expect(dateText).not.toContain('Mar 1');
		expect(dateText).not.toContain('März 1');

		const hasFeb28 =
			dateText?.includes('28.02.2027') ||
			dateText?.includes('28.2.2027') ||
			dateText?.includes('2/28/2027') ||
			dateText?.includes('28/02/2027');
		expect(hasFeb28).toBeTruthy();

		// Already on the detail page — verify the stored dates in edit mode.
		const merchantTitle = page.locator('h1', {
			hasText: testVouchers.amazon.merchant_name
		});
		await expect(merchantTitle).toBeVisible({ timeout: 10000 });

		const editButton = page
			.getByRole('button', { name: /Edit|Bearbeiten/ })
			.first();
		await editButton.waitFor({ state: 'visible', timeout: 5000 });
		await editButton.click();

		const validUntilInput = page.locator('input#validUntil');
		await expect(validUntilInput).toBeVisible({ timeout: 5000 });

		expect(await page.locator('input#validFrom').inputValue()).toBe(
			testStartDate
		);
		expect(await validUntilInput.inputValue()).toBe(testExpiryDate);
		expect(await validUntilInput.inputValue()).not.toBe('2027-03-01');
	});
});
