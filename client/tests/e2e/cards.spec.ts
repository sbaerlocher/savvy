import { randomCardNumber, testCards } from './fixtures/test-data';
import { expect, test } from './fixtures/test-fixtures';

test.describe('Cards Management', () => {
	test.beforeEach(async ({ authenticatedPage }) => {
		// authenticatedPage fixture handles login
	});

	test('should display cards list', async ({
		authenticatedPage,
		cardsListPage
	}) => {
		await cardsListPage.goto();
		await cardsListPage.expectHeading();
	});

	test('should create a new card', async ({
		authenticatedPage,
		cardsListPage,
		cardFormPage
	}) => {
		const page = authenticatedPage;
		await cardsListPage.goto();
		await cardsListPage.clickNewButton();

		const cardData = { ...testCards.ikea, card_number: randomCardNumber() };
		await cardFormPage.fillCardForm({
			merchantName: cardData.merchant_name,
			program: cardData.merchant_name + ' Card',
			cardNumber: cardData.card_number,
			barcodeType: cardData.barcode_type,
			notes: cardData.notes
		});
		await cardFormPage.submit();

		await page.waitForURL(/\/cards($|\/[^/]+$)/, { timeout: 10000 });
		await cardsListPage.goto();
		await expect(
			page.locator(`text=${cardData.merchant_name}`).first()
		).toBeVisible();
	});

	test('should view card details', async ({
		authenticatedPage,
		cardsListPage,
		cardDetailPage
	}) => {
		await cardsListPage.goto();
		await cardsListPage.clickFirstItem();
		await expect(cardDetailPage.heading).toBeVisible();
		await expect(cardDetailPage.barcodeCanvas).toBeVisible({ timeout: 10000 });
	});

	test('should edit a card', async ({
		authenticatedPage,
		cardsListPage,
		cardFormPage,
		cardDetailPage
	}) => {
		const page = authenticatedPage;

		await cardsListPage.goto();
		await cardsListPage.clickNewButton();

		const cardData = { ...testCards.ikea, card_number: randomCardNumber() };
		await cardFormPage.fillCardForm({
			merchantName: cardData.merchant_name,
			program: cardData.merchant_name + ' Card',
			cardNumber: cardData.card_number
		});
		await cardFormPage.submit();

		await expect(page).toHaveURL(/\/cards\/[a-f0-9-]+$/);
		await cardDetailPage.waitForPageReady();
		await cardDetailPage.enterEditMode();

		const programField = page.locator('input#program');
		await expect(programField).toBeVisible();
		await programField.clear();
		await programField.fill('Updated Program');

		await cardDetailPage.save();
		await expect(page.locator('text=Updated Program')).toBeVisible({
			timeout: 5000
		});
	});

	test('should delete a card', async ({
		authenticatedPage,
		cardsListPage,
		cardDetailPage
	}) => {
		const page = authenticatedPage;
		await cardsListPage.goto();
		await cardsListPage.clickFirstItem();
		await cardDetailPage.waitForPageReady();

		await cardDetailPage.enterEditMode();
		await cardDetailPage.deleteResource();

		await expect(page).toHaveURL(/(\/cards\/?$|\/wallet\?type=cards)/);
	});

	test('should search/filter cards', async ({
		authenticatedPage,
		cardsListPage
	}) => {
		const page = authenticatedPage;
		await cardsListPage.goto();
		await cardsListPage.search('Migros');

		const count = await cardsListPage.items.count();

		if (count > 0) {
			await expect(page.locator('text=Migros').first()).toBeVisible();
		} else {
			const noResults = page.locator(
				'text=/no.*found|keine.*gefunden|keine.*vorhanden/i'
			);
			await expect(noResults).toBeVisible();
		}
	});

	test('should validate required fields', async ({
		authenticatedPage,
		cardsListPage
	}) => {
		const page = authenticatedPage;
		await cardsListPage.goto();
		await cardsListPage.clickNewButton();

		await page.click('button[type="submit"]');
		await expect(page).toHaveURL(/\/cards\/new/);

		const cardNumberField = page.locator('input#cardNumber');
		const isInvalid = await cardNumberField.evaluate(
			(el: HTMLInputElement) => !el.validity.valid
		);
		expect(isInvalid).toBe(true);
	});

	test('should display barcode correctly', async ({
		authenticatedPage,
		cardsListPage,
		cardDetailPage
	}) => {
		await cardsListPage.goto();
		await cardsListPage.clickFirstItem();

		await expect(cardDetailPage.barcodeCanvas).toBeVisible({ timeout: 10000 });

		const hasContent = await cardDetailPage.barcodeCanvas.evaluate(
			(canvas: HTMLCanvasElement) => {
				return canvas.width > 0 && canvas.height > 0;
			}
		);
		expect(hasContent).toBe(true);
	});
});
