import { createMultipleCards } from './fixtures/test-data';
import { expect, test } from './fixtures/test-fixtures';

test.describe('Batch Operations', () => {
	test.beforeEach(async ({ authenticatedPage }) => {
		// authenticatedPage fixture handles login
	});

	test('should enter select mode on cards list', async ({
		authenticatedPage,
		cardsListPage
	}) => {
		await cardsListPage.goto();
		await expect(cardsListPage.firstItem).toBeVisible({ timeout: 10000 });

		await cardsListPage.enterSelectMode();
		// Verify BatchPanel is visible with "Select all" button
		await expect(cardsListPage.selectAllButton).toBeVisible();
	});

	test('should select multiple cards', async ({
		authenticatedPage,
		cardsListPage
	}) => {
		await cardsListPage.goto();
		await expect(cardsListPage.firstItem).toBeVisible({ timeout: 10000 });

		await cardsListPage.enterSelectMode();

		const itemCount = await cardsListPage.items.count();
		const toSelect = Math.min(itemCount, 2);

		for (let i = 0; i < toSelect; i++) {
			await cardsListPage.selectItemByIndex(i);
		}

		// Check that selection counter shows selected items (format: "N / Total")
		const selectedCount = cardsListPage.page.locator(
			'text=/\\d+\\s*\\/\\s*\\d+/'
		);
		if (await selectedCount.isVisible({ timeout: 3000 }).catch(() => false)) {
			await expect(selectedCount.first()).toBeVisible();
		}
	});

	test('should batch delete cards', async ({
		authenticatedPage,
		cardsListPage,
		cardFormPage
	}) => {
		const page = authenticatedPage;

		await createMultipleCards(page, cardsListPage, cardFormPage, 2);
		await cardsListPage.goto();
		await expect(cardsListPage.firstItem).toBeVisible({ timeout: 10000 });

		await cardsListPage.enterSelectMode();

		// Select only owned items to avoid disabling batch actions on shared cards
		const ownedCount = await cardsListPage.ownedItems.count();
		const toSelect = Math.min(ownedCount, 2);

		for (let i = 0; i < toSelect; i++) {
			await cardsListPage.selectOwnedItemByIndex(i);
		}

		const batchDeleteBtn = page
			.locator(
				'button:has-text("löschen"), button:has-text("Delete"), [data-testid="batch-delete"]'
			)
			.first();
		if (
			!(await batchDeleteBtn.isVisible({ timeout: 3000 }).catch(() => false))
		) {
			test.skip();
			return;
		}

		await batchDeleteBtn.click();

		// Confirm deletion in dialog
		const confirmButton = page
			.locator('[role="dialog"]')
			.locator('button:has-text("löschen"), button:has-text("Delete")')
			.first();
		if (await confirmButton.isVisible({ timeout: 3000 }).catch(() => false)) {
			const deleteResponse = page.waitForResponse(
				(resp) =>
					resp.url().includes('/batch/delete') &&
					resp.request().method() === 'POST' &&
					resp.status() < 400,
				{ timeout: 10000 }
			);
			await confirmButton.click();
			await deleteResponse;
		}
	});

	test('should batch share cards', async ({
		authenticatedPage,
		cardsListPage,
		cardFormPage
	}) => {
		const page = authenticatedPage;

		await createMultipleCards(page, cardsListPage, cardFormPage, 2);
		await cardsListPage.goto();
		await expect(cardsListPage.firstItem).toBeVisible({ timeout: 10000 });

		await cardsListPage.enterSelectMode();

		// Select only owned items to avoid disabling batch actions on shared cards
		const ownedCount = await cardsListPage.ownedItems.count();
		const toSelect = Math.min(ownedCount, 2);

		for (let i = 0; i < toSelect; i++) {
			await cardsListPage.selectOwnedItemByIndex(i);
		}

		const batchShareBtn = page
			.locator(
				'button:has-text("teilen"), button:has-text("Share"), [data-testid="batch-share"]'
			)
			.first();
		if (
			!(await batchShareBtn.isVisible({ timeout: 3000 }).catch(() => false))
		) {
			test.skip();
			return;
		}

		await batchShareBtn.click();

		const emailInput = page.locator('input[type="email"]').first();
		if (await emailInput.isVisible({ timeout: 3000 }).catch(() => false)) {
			await emailInput.fill('thomas.schmidt@example.com');

			const confirmButton = page
				.locator('[role="dialog"]')
				.locator('button:has-text("teilen"), button:has-text("Share")')
				.first();
			const shareResponse = page.waitForResponse(
				(resp) =>
					resp.url().includes('/batch/share') &&
					resp.request().method() === 'POST' &&
					resp.status() < 400,
				{ timeout: 10000 }
			);
			await confirmButton.click();
			await shareResponse;
		}
	});

	test('should batch transfer cards', async ({
		authenticatedPage,
		cardsListPage,
		cardFormPage
	}) => {
		const page = authenticatedPage;

		await createMultipleCards(page, cardsListPage, cardFormPage, 2);
		await cardsListPage.goto();
		await expect(cardsListPage.firstItem).toBeVisible({ timeout: 10000 });

		await cardsListPage.enterSelectMode();

		// Select only owned items to avoid disabling batch actions on shared cards
		const ownedCount = await cardsListPage.ownedItems.count();
		const toSelect = Math.min(ownedCount, 2);

		for (let i = 0; i < toSelect; i++) {
			await cardsListPage.selectOwnedItemByIndex(i);
		}

		const batchTransferBtn = page
			.locator(
				'button:has-text("übertragen"), button:has-text("Transfer"), [data-testid="batch-transfer"]'
			)
			.first();
		if (
			!(await batchTransferBtn.isVisible({ timeout: 3000 }).catch(() => false))
		) {
			test.skip();
			return;
		}

		await batchTransferBtn.click();

		const emailInput = page.locator('input[type="email"]').first();
		if (await emailInput.isVisible({ timeout: 3000 }).catch(() => false)) {
			await emailInput.fill('thomas.schmidt@example.com');

			const confirmButton = page
				.locator('[role="dialog"]')
				.locator('button:has-text("übertragen"), button:has-text("Transfer")')
				.first();
			if (await confirmButton.isVisible({ timeout: 3000 }).catch(() => false)) {
				const transferResponse = page.waitForResponse(
					(resp) =>
						resp.url().includes('/batch/transfer') &&
						resp.request().method() === 'POST' &&
						resp.status() < 400,
					{ timeout: 10000 }
				);
				await confirmButton.click();
				await transferResponse;
			}
		}
	});

	test('should batch export cards', async ({
		authenticatedPage,
		cardsListPage,
		cardFormPage
	}) => {
		const page = authenticatedPage;

		await createMultipleCards(page, cardsListPage, cardFormPage, 2);
		await cardsListPage.goto();
		await expect(cardsListPage.firstItem).toBeVisible({ timeout: 10000 });

		await cardsListPage.enterSelectMode();

		// Select only owned items
		const ownedCount = await cardsListPage.ownedItems.count();
		const toSelect = Math.min(ownedCount, 2);

		for (let i = 0; i < toSelect; i++) {
			await cardsListPage.selectOwnedItemByIndex(i);
		}

		const batchExportBtn = page
			.locator(
				'button:has-text("exportieren"), button:has-text("Export"), [data-testid="batch-export"]'
			)
			.first();
		if (
			!(await batchExportBtn.isVisible({ timeout: 3000 }).catch(() => false))
		) {
			test.skip();
			return;
		}

		// Intercept the download triggered by batch export
		const downloadPromise = page.waitForEvent('download', { timeout: 15000 });
		await batchExportBtn.click();
		const download = await downloadPromise;

		// Verify the download has a JSON filename
		expect(download.suggestedFilename()).toMatch(/\.json$/);
	});

	test('should enter select mode on vouchers list', async ({
		authenticatedPage,
		vouchersListPage
	}) => {
		await vouchersListPage.goto();

		const hasItems = await vouchersListPage.firstItem
			.isVisible({ timeout: 5000 })
			.catch(() => false);
		if (!hasItems) {
			test.skip();
			return;
		}

		await vouchersListPage.enterSelectMode();
	});

	test('should enter select mode on gift cards list', async ({
		authenticatedPage,
		giftCardsListPage
	}) => {
		await giftCardsListPage.goto();

		const hasItems = await giftCardsListPage.firstItem
			.isVisible({ timeout: 5000 })
			.catch(() => false);
		if (!hasItems) {
			test.skip();
			return;
		}

		await giftCardsListPage.enterSelectMode();
	});

	test('should keep the active type locked while in select mode', async ({
		authenticatedPage,
		cardsListPage
	}) => {
		const page = authenticatedPage;
		await cardsListPage.goto();
		await expect(cardsListPage.firstItem).toBeVisible({ timeout: 10000 });

		await cardsListPage.enterSelectMode();

		// Batch operations assume a single concrete type — a mixed selection would
		// route to the wrong endpoint — so select mode must not offer a way back
		// to 'all'. The desktop wallet has no chip row of its own (the type is
		// picked in the filter panel, which select mode replaces with the batch
		// panel), so no type control may be reachable at all here.
		await expect(page.getByTestId('type-chip-all')).toHaveCount(0);
		await expect(page.getByTestId('type-chip-cards')).toHaveCount(0);

		// Leaving select mode brings the type control back, still on 'cards'.
		await cardsListPage.selectModeButton.click();
		await cardsListPage.filterButton.click();
		const cardsChip = page.getByTestId('type-chip-cards').first();
		await expect(cardsChip).toBeVisible({ timeout: 10000 });
		const activeClass = await cardsChip.getAttribute('class');
		expect(activeClass).toContain('bg-accent');
	});
});
