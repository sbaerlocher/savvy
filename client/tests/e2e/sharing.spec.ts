import {
	testCards,
	testVouchers,
	testGiftCards,
	createCardAndNavigate
} from './fixtures/test-data';
import { expect, test, TEST_USERS } from './fixtures/test-fixtures';
import { LoginPage } from './pages/login.page';
import { ResourceDetailPage } from './pages/resource-detail.page';
import { ResourceFormPage } from './pages/resource-form.page';
import { ResourceListPage } from './pages/resource-list.page';

test.describe('Sharing', () => {
	test.beforeEach(async ({ authenticatedPage }) => {
		// authenticatedPage fixture handles login
	});

	test('should share a card with another user', async ({
		authenticatedPage,
		cardsListPage,
		cardFormPage,
		cardDetailPage
	}) => {
		await createCardAndNavigate(authenticatedPage, cardsListPage, cardFormPage);
		await cardDetailPage.waitForPageReady();

		await cardDetailPage.addShare(TEST_USERS.shared.email);
		await cardDetailPage.expectToast('geteilt');
	});

	test('should update share permissions', async ({
		authenticatedPage,
		cardsListPage,
		cardFormPage,
		cardDetailPage
	}) => {
		const page = authenticatedPage;
		await createCardAndNavigate(page, cardsListPage, cardFormPage);
		await cardDetailPage.waitForPageReady();

		await cardDetailPage.addShare(TEST_USERS.shared.email, {});

		const shareSection = cardDetailPage.shareSection;
		await expect(shareSection).toBeVisible({ timeout: 5000 });

		const editShareBtn = shareSection
			.locator(
				'button[aria-label*="edit" i], button[aria-label*="bearbeiten" i], [data-testid="edit-share"]'
			)
			.first();
		if (await editShareBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
			await editShareBtn.click();

			const permissionSelect = page
				.locator('select[name="permission"], select#permission')
				.first();
			if (await permissionSelect.isVisible({ timeout: 3000 })) {
				await permissionSelect.selectOption('write');
				const saveBtn = page
					.locator(
						'button[type="submit"]:has-text("Speichern"), button[type="submit"]:has-text("Save")'
					)
					.first();
				const updateResponse = page.waitForResponse(
					(resp) =>
						resp.url().includes('/shares') &&
						resp.request().method() === 'PATCH' &&
						resp.status() < 400,
					{ timeout: 10000 }
				);
				await saveBtn.click();
				await updateResponse;
			}
		}
	});

	test('should remove a share', async ({
		authenticatedPage,
		cardsListPage,
		cardFormPage,
		cardDetailPage
	}) => {
		const page = authenticatedPage;
		await createCardAndNavigate(page, cardsListPage, cardFormPage);
		await cardDetailPage.waitForPageReady();

		await cardDetailPage.addShare(TEST_USERS.shared.email, {});

		const shareSection = cardDetailPage.shareSection;
		await expect(shareSection).toBeVisible({ timeout: 5000 });

		const removeShareBtn = shareSection
			.locator(
				'button[aria-label*="remove" i], button[aria-label*="entfernen" i], button[aria-label*="delete" i], button[aria-label*="löschen" i], [data-testid="remove-share"]'
			)
			.first();
		if (await removeShareBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
			const deleteResponse = page.waitForResponse(
				(resp) =>
					resp.url().includes('/shares') &&
					resp.request().method() === 'DELETE' &&
					resp.status() < 400,
				{ timeout: 10000 }
			);
			await removeShareBtn.click();

			const confirmButton = page.locator('[data-testid="modal-confirm"]');
			if (await confirmButton.isVisible({ timeout: 3000 }).catch(() => false)) {
				await confirmButton.click();
			}
			await deleteResponse;
		}
	});

	test('should transfer ownership of a card', async ({
		authenticatedPage,
		cardsListPage,
		cardFormPage,
		cardDetailPage
	}) => {
		const page = authenticatedPage;
		await createCardAndNavigate(page, cardsListPage, cardFormPage);
		await cardDetailPage.waitForPageReady();

		const transferSection = cardDetailPage.transferSection;
		if (
			!(await transferSection.isVisible({ timeout: 3000 }).catch(() => false))
		) {
			test.skip();
			return;
		}

		// Button is a sibling of the heading, find it on the page near the transfer section
		const transferButton = page
			.locator(
				'button:has-text("Übergeben"), button:has-text("Übertragen"), button:has-text("Transfer")'
			)
			.first();
		await expect(transferButton).toBeVisible({ timeout: 3000 });
		await transferButton.click();

		const emailInput = page
			.locator('input[type="email"], input#transfer_email, input[name="email"]')
			.first();
		await expect(emailInput).toBeVisible({ timeout: 5000 });
		await emailInput.fill(TEST_USERS.shared.email);

		// Click autocomplete suggestion if it appears
		const suggestion = page
			.locator(`button:has-text("${TEST_USERS.shared.email}")`)
			.first();
		if (await suggestion.isVisible({ timeout: 2000 }).catch(() => false)) {
			await suggestion.click();
		}

		// Confirm inside the form. Android renders it in a bottom sheet, so the
		// search is scoped to the open dialog when there is one — otherwise the
		// row button behind the scrim matches first and swallows the click.
		const transferForm = page.locator('[role="dialog"]');
		const transferFormScope = (await transferForm
			.first()
			.isVisible({ timeout: 1000 })
			.catch(() => false))
			? transferForm.first()
			: page.locator('body');
		const confirmTransfer = transferFormScope
			.locator(
				'button:has-text("übertragen"), button:has-text("Übergeben"), button:has-text("Transfer"), button:has-text("Transférer")'
			)
			.first();
		await expect(confirmTransfer).toBeVisible({ timeout: 5000 });
		await confirmTransfer.click();

		// Handle confirmation dialog
		const dialog = page.locator('dialog, [role="dialog"]').last();
		const dialogConfirmButton = dialog
			.locator(
				'button:has-text("übertragen"), button:has-text("Transfer"), button:has-text("Confirm")'
			)
			.first();
		if (
			await dialogConfirmButton.isVisible({ timeout: 3000 }).catch(() => false)
		) {
			const transferResponse = page.waitForResponse(
				(resp) => resp.url().includes('/transfer') && resp.status() < 400,
				{ timeout: 10000 }
			);
			await dialogConfirmButton.click();
			await transferResponse;
		}
	});

	test('should share a voucher with another user (read-only)', async ({
		authenticatedPage,
		vouchersListPage,
		voucherFormPage,
		voucherDetailPage
	}) => {
		const page = authenticatedPage;

		// Create a voucher first
		await vouchersListPage.goto();
		await vouchersListPage.clickNewButton();

		const voucherCode = `SHARE-${Date.now()}`;
		await voucherFormPage.fillVoucherForm({
			merchantName: testVouchers.amazon.merchant_name,
			code: voucherCode,
			type: testVouchers.amazon.type,
			value: testVouchers.amazon.value,
			validFrom: testVouchers.amazon.valid_from,
			validUntil: testVouchers.amazon.valid_until,
			notes: 'Share test voucher'
		});
		await voucherFormPage.submit();
		await page.waitForURL(/\/vouchers\/[a-f0-9-]+$/, { timeout: 10000 });

		await voucherDetailPage.waitForPageReady();
		await voucherDetailPage.addShare(TEST_USERS.shared.email);
		await voucherDetailPage.expectToast();
	});

	test('should share a gift card with another user', async ({
		authenticatedPage,
		giftCardsListPage,
		giftCardFormPage,
		giftCardDetailPage
	}) => {
		const page = authenticatedPage;

		// Create a gift card first
		await giftCardsListPage.goto();
		await giftCardsListPage.clickNewButton();

		const cardNumber = `SHARE-GC-${Date.now()}`;
		await giftCardFormPage.fillGiftCardForm({
			merchantName: testGiftCards.appleStore.merchant_name,
			cardNumber,
			initialBalance: testGiftCards.appleStore.initial_balance,
			notes: 'Share test gift card'
		});
		await giftCardFormPage.submit();
		await page.waitForURL(/\/gift-cards\/[a-f0-9-]+$/, { timeout: 10000 });

		await giftCardDetailPage.waitForPageReady();
		await giftCardDetailPage.addShare(TEST_USERS.shared.email, {
			canEdit: true
		});
		await giftCardDetailPage.expectToast();
	});

	test('should transfer ownership of a voucher', async ({
		authenticatedPage,
		vouchersListPage,
		voucherFormPage,
		voucherDetailPage
	}) => {
		const page = authenticatedPage;

		await vouchersListPage.goto();
		await vouchersListPage.clickNewButton();

		const voucherCode = `TRANSFER-${Date.now()}`;
		await voucherFormPage.fillVoucherForm({
			merchantName: testVouchers.amazon.merchant_name,
			code: voucherCode,
			type: testVouchers.amazon.type,
			value: testVouchers.amazon.value,
			validFrom: testVouchers.amazon.valid_from,
			validUntil: testVouchers.amazon.valid_until,
			notes: 'Transfer test voucher'
		});
		await voucherFormPage.submit();
		await page.waitForURL(/\/vouchers\/[a-f0-9-]+$/, { timeout: 10000 });

		await voucherDetailPage.waitForPageReady();

		const transferSection = voucherDetailPage.transferSection;
		if (
			!(await transferSection.isVisible({ timeout: 3000 }).catch(() => false))
		) {
			test.skip();
			return;
		}

		const transferButton = page
			.locator(
				'button:has-text("Übergeben"), button:has-text("Übertragen"), button:has-text("Transfer")'
			)
			.first();
		await expect(transferButton).toBeVisible({ timeout: 3000 });
		await transferButton.click();

		const emailInput = page
			.locator('input[type="email"], input#transfer_email, input[name="email"]')
			.first();
		await expect(emailInput).toBeVisible({ timeout: 5000 });
		await emailInput.fill(TEST_USERS.shared.email);

		const suggestion = page
			.locator(`button:has-text("${TEST_USERS.shared.email}")`)
			.first();
		if (await suggestion.isVisible({ timeout: 2000 }).catch(() => false)) {
			await suggestion.click();
		}

		// Confirm inside the form. Android renders it in a bottom sheet, so the
		// search is scoped to the open dialog when there is one — otherwise the
		// row button behind the scrim matches first and swallows the click.
		const transferForm = page.locator('[role="dialog"]');
		const transferFormScope = (await transferForm
			.first()
			.isVisible({ timeout: 1000 })
			.catch(() => false))
			? transferForm.first()
			: page.locator('body');
		const confirmTransfer = transferFormScope
			.locator(
				'button:has-text("übertragen"), button:has-text("Übergeben"), button:has-text("Transfer"), button:has-text("Transférer")'
			)
			.first();
		await expect(confirmTransfer).toBeVisible({ timeout: 5000 });
		await confirmTransfer.click();

		const dialog = page.locator('dialog, [role="dialog"]').last();
		const dialogConfirmButton = dialog
			.locator(
				'button:has-text("übertragen"), button:has-text("Transfer"), button:has-text("Confirm")'
			)
			.first();
		if (
			await dialogConfirmButton.isVisible({ timeout: 3000 }).catch(() => false)
		) {
			const transferResponse = page.waitForResponse(
				(resp) => resp.url().includes('/transfer') && resp.status() < 400,
				{ timeout: 10000 }
			);
			await dialogConfirmButton.click();
			await transferResponse;
		}
	});

	test('should transfer ownership of a gift card', async ({
		authenticatedPage,
		giftCardsListPage,
		giftCardFormPage,
		giftCardDetailPage
	}) => {
		const page = authenticatedPage;

		await giftCardsListPage.goto();
		await giftCardsListPage.clickNewButton();

		const cardNumber = `TRANSFER-GC-${Date.now()}`;
		await giftCardFormPage.fillGiftCardForm({
			merchantName: testGiftCards.appleStore.merchant_name,
			cardNumber,
			initialBalance: testGiftCards.appleStore.initial_balance,
			notes: 'Transfer test gift card'
		});
		await giftCardFormPage.submit();
		await page.waitForURL(/\/gift-cards\/[a-f0-9-]+$/, { timeout: 10000 });

		await giftCardDetailPage.waitForPageReady();

		const transferSection = giftCardDetailPage.transferSection;
		if (
			!(await transferSection.isVisible({ timeout: 3000 }).catch(() => false))
		) {
			test.skip();
			return;
		}

		const transferButton = page
			.locator(
				'button:has-text("Übergeben"), button:has-text("Übertragen"), button:has-text("Transfer")'
			)
			.first();
		await expect(transferButton).toBeVisible({ timeout: 3000 });
		await transferButton.click();

		const emailInput = page
			.locator('input[type="email"], input#transfer_email, input[name="email"]')
			.first();
		await expect(emailInput).toBeVisible({ timeout: 5000 });
		await emailInput.fill(TEST_USERS.shared.email);

		const suggestion = page
			.locator(`button:has-text("${TEST_USERS.shared.email}")`)
			.first();
		if (await suggestion.isVisible({ timeout: 2000 }).catch(() => false)) {
			await suggestion.click();
		}

		// Confirm inside the form. Android renders it in a bottom sheet, so the
		// search is scoped to the open dialog when there is one — otherwise the
		// row button behind the scrim matches first and swallows the click.
		const transferForm = page.locator('[role="dialog"]');
		const transferFormScope = (await transferForm
			.first()
			.isVisible({ timeout: 1000 })
			.catch(() => false))
			? transferForm.first()
			: page.locator('body');
		const confirmTransfer = transferFormScope
			.locator(
				'button:has-text("übertragen"), button:has-text("Übergeben"), button:has-text("Transfer"), button:has-text("Transférer")'
			)
			.first();
		await expect(confirmTransfer).toBeVisible({ timeout: 5000 });
		await confirmTransfer.click();

		const dialog = page.locator('dialog, [role="dialog"]').last();
		const dialogConfirmButton = dialog
			.locator(
				'button:has-text("übertragen"), button:has-text("Transfer"), button:has-text("Confirm")'
			)
			.first();
		if (
			await dialogConfirmButton.isVisible({ timeout: 3000 }).catch(() => false)
		) {
			const transferResponse = page.waitForResponse(
				(resp) => resp.url().includes('/transfer') && resp.status() < 400,
				{ timeout: 10000 }
			);
			await dialogConfirmButton.click();
			await transferResponse;

			// Regression guard for #121: after a successful transfer the user
			// must land back on the gift-cards list, not a white screen. The
			// detail page reloads to /gift-cards because the owner just lost
			// access — assert the redirect actually happens, the list renders,
			// and the transferred card is gone from it.
			await page.waitForURL(/\/gift-cards\/?$/, { timeout: 10000 });
			// Assert the list's own <h1> rendered — not just the persistent nav
			// shell, which carries the same "Gift Cards" label on every page and
			// would stay visible even on the blank content area #121 is about.
			await expect(
				page.locator('h1', { hasText: /Geschenkkarten|Gift Cards|Cartes/i })
			).toBeVisible({ timeout: 10000 });
			await expect(page.locator(`text=${cardNumber}`)).toHaveCount(0);
		}
	});

	test('should show shared items as read-only for recipient', async ({
		browser
	}) => {
		const ownerContext = await browser.newContext();
		const ownerPage = await ownerContext.newPage();
		const ownerLogin = new LoginPage(ownerPage);
		await ownerLogin.goto();
		await ownerLogin.login(
			TEST_USERS.regular.email,
			TEST_USERS.regular.password
		);
		await ownerPage.waitForURL(/\/(dashboard|cards|vouchers|gift-cards)\/?/, {
			timeout: 15000
		});

		const ownerCardsList = new ResourceListPage(ownerPage, 'cards');
		const ownerCardForm = new ResourceFormPage(ownerPage, 'cards');
		const ownerCardDetail = new ResourceDetailPage(ownerPage, 'cards');

		await createCardAndNavigate(ownerPage, ownerCardsList, ownerCardForm);
		await ownerCardDetail.waitForPageReady();
		await ownerCardDetail.addShare(TEST_USERS.shared.email, {});

		// Login as shared user
		const sharedContext = await browser.newContext();
		const sharedPage = await sharedContext.newPage();
		const sharedLogin = new LoginPage(sharedPage);
		await sharedLogin.goto();
		await sharedLogin.login(
			TEST_USERS.shared.email,
			TEST_USERS.shared.password
		);
		await sharedPage.waitForURL(/\/(dashboard|cards|vouchers|gift-cards)\/?/, {
			timeout: 15000
		});

		const sharedCardsList = new ResourceListPage(sharedPage, 'cards');
		await sharedCardsList.goto();

		const sharedCard = sharedPage
			.locator(`[data-owner]:has-text("${testCards.ikea.merchant_name}")`)
			.first();
		if (await sharedCard.isVisible({ timeout: 5000 }).catch(() => false)) {
			await sharedCard.click();
			await sharedPage.waitForURL(/\/cards\/[a-f0-9-]+$/, { timeout: 10000 });

			const editButton = sharedPage
				.locator('button:has-text("Bearbeiten"), button:has-text("Edit")')
				.first();
			const isEditable = await editButton
				.isVisible({ timeout: 3000 })
				.catch(() => false);
			expect(isEditable).toBeFalsy();
		}

		await ownerContext.close();
		await sharedContext.close();
	});
});
