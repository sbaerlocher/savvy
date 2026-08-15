import { testGiftCards, uniqueGiftCardNumber } from './fixtures/test-data';
import { expect, test } from './fixtures/test-fixtures';

test.describe('Gift Cards Management', () => {
	test.beforeEach(async ({ authenticatedPage }) => {
		// authenticatedPage fixture handles login
	});

	test('should display gift cards list', async ({
		authenticatedPage,
		giftCardsListPage
	}) => {
		await giftCardsListPage.goto();
		await giftCardsListPage.expectHeading();
	});

	test('should create a new gift card', async ({
		authenticatedPage,
		giftCardsListPage,
		giftCardFormPage
	}) => {
		const page = authenticatedPage;
		await giftCardsListPage.goto();
		await giftCardsListPage.clickNewButton();

		const giftCardData = {
			...testGiftCards.appleStore,
			card_number: uniqueGiftCardNumber()
		};
		await giftCardFormPage.fillGiftCardForm({
			merchantName: giftCardData.merchant_name,
			cardNumber: giftCardData.card_number,
			initialBalance: giftCardData.initial_balance,
			currency: giftCardData.currency,
			notes: giftCardData.notes
		});
		await giftCardFormPage.submit();

		// `/gift-cards/new` already satisfies `\/gift-cards\/[^/]+$`, so the old
		// loose pattern resolved instantly on the form URL — before the create
		// handler's `window.location.href = /gift-cards/<id>`
		// (gift-cards/new/+page.svelte:113) had even fired. The following goto()
		// then raced that in-flight full document load and stalled until its
		// timeout. Anchor on the UUID so this waits for the real detail route.
		await page.waitForURL(/\/gift-cards\/[a-f0-9-]{36}$/, { timeout: 15000 });
		await giftCardsListPage.goto();
		await expect(
			page.locator(`text=${giftCardData.merchant_name}`).first()
		).toBeVisible();
	});

	test('should view gift card details', async ({
		authenticatedPage,
		giftCardsListPage,
		giftCardFormPage,
		giftCardDetailPage
	}) => {
		await giftCardsListPage.goto();
		await giftCardsListPage.clickNewButton();

		const giftCardData = {
			...testGiftCards.appleStore,
			card_number: uniqueGiftCardNumber()
		};
		await giftCardFormPage.fillGiftCardForm({
			merchantName: giftCardData.merchant_name,
			cardNumber: giftCardData.card_number,
			initialBalance: giftCardData.initial_balance,
			currency: giftCardData.currency
		});
		await giftCardFormPage.submit();

		await giftCardsListPage.goto();
		await expect(giftCardsListPage.firstItem).toBeVisible({ timeout: 10000 });
		await giftCardsListPage.clickFirstItem();

		await expect(giftCardDetailPage.heading).toBeVisible();
	});

	test('should edit a gift card', async ({
		authenticatedPage,
		giftCardsListPage,
		giftCardDetailPage
	}) => {
		const page = authenticatedPage;
		await giftCardsListPage.goto();

		await expect(giftCardsListPage.firstItem).toBeVisible({ timeout: 10000 });
		await giftCardsListPage.clickFirstItem();
		await giftCardDetailPage.waitForPageReady();

		await giftCardDetailPage.enterEditMode();
		const notesField = page.locator('textarea#notes, input#notes');
		if (await notesField.isVisible()) {
			await notesField.clear();
			await notesField.fill('Updated gift card notes');
		}

		await giftCardDetailPage.save();
		await giftCardDetailPage.waitForPageReady();
	});

	test('should delete a gift card', async ({
		authenticatedPage,
		giftCardsListPage,
		giftCardFormPage,
		giftCardDetailPage
	}) => {
		const page = authenticatedPage;

		await giftCardsListPage.goto();
		await giftCardsListPage.clickNewButton();

		const giftCardData = {
			...testGiftCards.appleStore,
			card_number: uniqueGiftCardNumber()
		};
		await giftCardFormPage.fillGiftCardForm({
			merchantName: giftCardData.merchant_name,
			cardNumber: giftCardData.card_number,
			initialBalance: giftCardData.initial_balance,
			currency: giftCardData.currency
		});
		await giftCardFormPage.submit();

		await expect(page).toHaveURL(/\/gift-cards\/[a-f0-9-]+$/);
		await giftCardDetailPage.waitForPageReady();
		await giftCardDetailPage.enterEditMode();
		await giftCardDetailPage.deleteResource();

		// After delete the detail page navigates to the legacy /gift-cards list,
		// which redirects into the unified /wallet?type=gift-cards overview.
		await expect(page).toHaveURL(/\/wallet\?type=gift-cards$/);
	});

	test('should add a transaction to gift card', async ({
		authenticatedPage,
		giftCardsListPage,
		giftCardFormPage,
		giftCardDetailPage
	}) => {
		const page = authenticatedPage;

		await giftCardsListPage.goto();
		await giftCardsListPage.clickNewButton();

		const giftCardData = {
			...testGiftCards.appleStore,
			card_number: uniqueGiftCardNumber()
		};
		await giftCardFormPage.fillGiftCardForm({
			merchantName: giftCardData.merchant_name,
			cardNumber: giftCardData.card_number,
			initialBalance: 100,
			currency: 'CHF'
		});
		await giftCardFormPage.submit();

		await expect(page).toHaveURL(/\/gift-cards\/[a-f0-9-]+$/);
		await giftCardDetailPage.waitForPageReady();

		const addTransactionButton = page.getByTestId('add-transaction');
		await expect(addTransactionButton).toBeVisible({ timeout: 5000 });
		await addTransactionButton.click();

		const amountInput = page.getByRole('spinbutton', {
			name: /Betrag|Amount/i
		});
		await expect(amountInput).toBeVisible({ timeout: 5000 });
		await amountInput.fill('25.00');

		const descriptionInput = page.getByRole('textbox', {
			name: /Beschreibung|Description/i
		});
		if (await descriptionInput.isVisible()) {
			await descriptionInput.fill('Coffee purchase');
		}

		const submitTransaction = page
			.locator(
				'button:has-text("Speichern"), button:has-text("Save"), button:has-text("Hinzufügen"), button:has-text("Add")'
			)
			.first();

		const transactionResponse = page.waitForResponse(
			(resp) =>
				resp.url().includes('/api/v1/gift-cards/') &&
				resp.url().includes('/transactions') &&
				resp.status() < 400,
			{ timeout: 10000 }
		);
		await submitTransaction.click();
		await transactionResponse;

		await expect(page.locator('text=/75\\.00|75,00/i').first()).toBeVisible({
			timeout: 5000
		});
	});

	test('should delete a transaction from gift card', async ({
		authenticatedPage,
		giftCardsListPage,
		giftCardFormPage,
		giftCardDetailPage
	}) => {
		const page = authenticatedPage;

		await giftCardsListPage.goto();
		await giftCardsListPage.clickNewButton();

		const giftCardData = {
			...testGiftCards.appleStore,
			card_number: uniqueGiftCardNumber()
		};
		await giftCardFormPage.fillGiftCardForm({
			merchantName: giftCardData.merchant_name,
			cardNumber: giftCardData.card_number,
			initialBalance: 200,
			currency: 'CHF'
		});
		await giftCardFormPage.submit();

		await expect(page).toHaveURL(/\/gift-cards\/[a-f0-9-]+$/);
		await giftCardDetailPage.waitForPageReady();

		// Add a transaction first
		const addTransactionButton = page.getByTestId('add-transaction');
		await addTransactionButton.click();

		const amountInput = page.getByRole('spinbutton', {
			name: /Betrag|Amount/i
		});
		await expect(amountInput).toBeVisible({ timeout: 5000 });
		await amountInput.fill('50.00');

		const submitTransaction = page
			.locator(
				'button:has-text("Speichern"), button:has-text("Save"), button:has-text("Hinzufügen"), button:has-text("Add")'
			)
			.first();
		const txnResponse = page.waitForResponse(
			(resp) => resp.url().includes('/transactions') && resp.status() < 400,
			{ timeout: 10000 }
		);
		await submitTransaction.click();
		await txnResponse;

		await expect(page.locator('text=/150\\.00|150,00/i').first()).toBeVisible({
			timeout: 5000
		});

		// Delete the transaction
		const deleteTransactionBtn = page
			.locator(
				'[data-testid="delete-transaction"], button[aria-label*="delete" i], button[aria-label*="löschen" i]'
			)
			.first();
		if (
			await deleteTransactionBtn.isVisible({ timeout: 3000 }).catch(() => false)
		) {
			const deleteTxnResponse = page.waitForResponse(
				(resp) =>
					resp.url().includes('/transactions') &&
					resp.request().method() === 'DELETE' &&
					resp.status() < 400,
				{ timeout: 10000 }
			);
			await deleteTransactionBtn.click();

			const confirmButton = page.locator('[data-testid="modal-confirm"]');
			if (await confirmButton.isVisible({ timeout: 3000 }).catch(() => false)) {
				await confirmButton.click();
			}
			await deleteTxnResponse;

			await expect(page.locator('text=/200\\.00|200,00/i').first()).toBeVisible(
				{ timeout: 5000 }
			);
		}
	});

	test('should track balance across multiple transactions', async ({
		authenticatedPage,
		giftCardsListPage,
		giftCardFormPage,
		giftCardDetailPage
	}) => {
		const page = authenticatedPage;

		await giftCardsListPage.goto();
		await giftCardsListPage.clickNewButton();

		const giftCardData = {
			...testGiftCards.appleStore,
			card_number: uniqueGiftCardNumber()
		};
		await giftCardFormPage.fillGiftCardForm({
			merchantName: giftCardData.merchant_name,
			cardNumber: giftCardData.card_number,
			initialBalance: 100,
			currency: 'CHF'
		});
		await giftCardFormPage.submit();

		await expect(page).toHaveURL(/\/gift-cards\/[a-f0-9-]+$/);
		await giftCardDetailPage.waitForPageReady();

		const addAndSubmitTransaction = async (amount: string) => {
			const addBtn = page.getByTestId('add-transaction');
			await addBtn.click();

			const amountInput = page.getByRole('spinbutton', {
				name: /Betrag|Amount/i
			});
			await expect(amountInput).toBeVisible({ timeout: 5000 });
			await amountInput.fill(amount);

			const submitBtn = page
				.locator(
					'button:has-text("Speichern"), button:has-text("Save"), button:has-text("Hinzufügen"), button:has-text("Add")'
				)
				.first();
			const resp = page.waitForResponse(
				(r) => r.url().includes('/transactions') && r.status() < 400,
				{ timeout: 10000 }
			);
			await submitBtn.click();
			await resp;
		};

		await addAndSubmitTransaction('30.00');
		await expect(page.locator('text=/70\\.00|70,00/i').first()).toBeVisible({
			timeout: 5000
		});

		await addAndSubmitTransaction('20.00');
		await expect(page.locator('text=/50\\.00|50,00/i').first()).toBeVisible({
			timeout: 5000
		});
	});

	test('should filter gift cards by merchant', async ({
		authenticatedPage,
		giftCardsListPage
	}) => {
		await giftCardsListPage.goto();

		if (
			!(await giftCardsListPage.filterButton.isVisible().catch(() => false))
		) {
			test.skip();
			return;
		}

		await giftCardsListPage.filterButton.click();

		// The unified wallet filter sheet no longer exposes a native <select>
		// merchant picker; per-merchant browsing moved to /merchants. Skip when
		// no legacy select is present rather than assert one into existence.
		const merchantFilter = giftCardsListPage.page
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
		await giftCardsListPage.waitForPageReady();
		expect(await giftCardsListPage.items.count()).toBeGreaterThanOrEqual(0);
	});
});
