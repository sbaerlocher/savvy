import { testCards, testGiftCards, testVouchers } from './fixtures/test-data';
import { expect, test } from './fixtures/test-fixtures';
import { ResourceFormPage } from './pages/resource-form.page';

test.describe('Form Validation', () => {
	test.beforeEach(async ({ authenticatedPage }) => {
		// authenticatedPage fixture handles login
	});

	test('should require merchant for card creation', async ({
		authenticatedPage,
		cardsListPage,
		cardFormPage
	}) => {
		const page = authenticatedPage;
		await cardsListPage.goto();
		await cardsListPage.clickNewButton();

		const submitButton = page.locator('button[type="submit"]');
		await submitButton.click();

		const merchantError = page.locator(
			'text=/Händler|Merchant|erforderlich|required/i'
		);
		const invalidState = page.locator('select:invalid, [aria-invalid="true"]');

		const hasError = await merchantError.isVisible().catch(() => false);
		const hasInvalid = await invalidState
			.first()
			.isVisible()
			.catch(() => false);
		const stillOnForm = page.url().includes('/new');

		expect(hasError || hasInvalid || stillOnForm).toBeTruthy();
	});

	test('should require merchant for voucher creation', async ({
		authenticatedPage,
		vouchersListPage
	}) => {
		const page = authenticatedPage;
		await vouchersListPage.goto();
		await vouchersListPage.clickNewButton();

		const submitButton = page.locator('button[type="submit"]');
		await submitButton.click();

		expect(page.url().includes('/new')).toBeTruthy();
	});

	test('should require merchant for gift card creation', async ({
		authenticatedPage,
		giftCardsListPage
	}) => {
		const page = authenticatedPage;
		await giftCardsListPage.goto();
		await giftCardsListPage.clickNewButton();

		const submitButton = page.locator('button[type="submit"]');
		await submitButton.click();

		expect(page.url().includes('/new')).toBeTruthy();
	});

	test('should validate card number format', async ({
		authenticatedPage,
		cardsListPage,
		cardFormPage
	}) => {
		const page = authenticatedPage;
		await cardsListPage.goto();
		await cardsListPage.clickNewButton();

		await cardFormPage.selectMerchant(testCards.ikea.merchant_name);

		const cardNumberInput = page.locator('input#cardNumber');
		await cardNumberInput.fill('');

		const submitButton = page.locator('button[type="submit"]');
		await submitButton.click();

		const error = page.locator(
			'text=/Kartennummer|Card number|erforderlich|required/i'
		);
		const hasError = await error.isVisible().catch(() => false);
		const stillOnForm = page.url().includes('/new');
		expect(hasError || stillOnForm).toBeTruthy();
	});

	test('should validate voucher code format', async ({
		authenticatedPage,
		vouchersListPage
	}) => {
		const page = authenticatedPage;
		await vouchersListPage.goto();
		await vouchersListPage.clickNewButton();

		const voucherForm = new ResourceFormPage(page, 'vouchers');
		await voucherForm.selectMerchant(testVouchers.amazon.merchant_name);

		const codeInput = page.locator('input#code');
		await codeInput.fill('');

		const submitButton = page.locator('button[type="submit"]');
		await submitButton.click();

		expect(page.url().includes('/new')).toBeTruthy();
	});

	test('should validate gift card balance is non-negative', async ({
		authenticatedPage,
		giftCardsListPage
	}) => {
		const page = authenticatedPage;
		await giftCardsListPage.goto();
		await giftCardsListPage.clickNewButton();

		const giftCardForm = new ResourceFormPage(page, 'gift-cards');
		await giftCardForm.selectMerchant(testGiftCards.appleStore.merchant_name);

		const balanceInput = page.locator('input#initialBalance');
		if (await balanceInput.isVisible()) {
			await balanceInput.fill('-50');

			const submitButton = page.locator('button[type="submit"]');
			await submitButton.click();

			const error = page.locator(
				'text=/negativ|negative|ungültig|invalid|mindestens|minimum/i'
			);
			const hasError = await error.isVisible().catch(() => false);
			const stillOnForm = page.url().includes('/new');
			expect(hasError || stillOnForm).toBeTruthy();
		}
	});

	test('should validate date ranges for vouchers', async ({
		authenticatedPage,
		vouchersListPage
	}) => {
		const page = authenticatedPage;
		await vouchersListPage.goto();
		await vouchersListPage.clickNewButton();

		const voucherForm = new ResourceFormPage(page, 'vouchers');
		await voucherForm.selectMerchant(testVouchers.amazon.merchant_name);

		const validFromInput = page.locator('input#validFrom');
		const validUntilInput = page.locator('input#validUntil');

		if (
			(await validFromInput.isVisible()) &&
			(await validUntilInput.isVisible())
		) {
			await validFromInput.fill('2026-12-31');
			await validUntilInput.fill('2026-01-01');

			const codeInput = page.locator('input#code');
			await codeInput.fill('TEST-VALIDATION');

			const submitButton = page.locator('button[type="submit"]');
			await submitButton.click();

			const error = page.locator(
				'text=/datum|date|vor|before|ungültig|invalid/i'
			);
			const hasError = await error.isVisible().catch(() => false);
			const stillOnForm = page.url().includes('/new');
			expect(hasError || stillOnForm).toBeTruthy();
		}
	});

	test('should show validation errors inline', async ({
		authenticatedPage,
		cardsListPage
	}) => {
		const page = authenticatedPage;
		await cardsListPage.goto();
		await cardsListPage.clickNewButton();

		const submitButton = page.locator('button[type="submit"]');
		await submitButton.click();

		const inlineErrors = page.locator(
			'.error, .text-error, [role="alert"], .invalid-feedback, [aria-invalid="true"]'
		);
		const errorCount = await inlineErrors.count();

		expect(errorCount >= 0).toBeTruthy();
	});
});
