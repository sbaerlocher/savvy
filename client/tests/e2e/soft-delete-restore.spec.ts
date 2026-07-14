import { uniqueCardNumber, testCards } from './fixtures/test-data';
import { expect, test } from './fixtures/test-fixtures';

/**
 * E2E test: soft-delete duplicate restore flow
 *
 * Flow:
 * 1. Create a card with a unique card number → navigate to detail
 * 2. Delete the card (soft-delete via detail page)
 * 3. Navigate to /cards/new and submit a new card with the SAME card number
 * 4. Assert the 409 duplicate-warning banner appears with the restore button
 * 5. Click "Restore deleted entry"
 * 6. Assert navigation to /cards/<id> (card detail) and card is active/visible
 */
test.describe('Soft-delete duplicate restore flow', () => {
	test('should show restore banner and restore a soft-deleted card', async ({
		authenticatedPage,
		cardsListPage,
		cardFormPage,
		cardDetailPage
	}) => {
		const page = authenticatedPage;

		// ── Step 1: Create a card with a unique number ──────────────────────────
		const cardNumber = uniqueCardNumber();

		await cardsListPage.goto();
		await cardsListPage.clickNewButton();

		await cardFormPage.fillCardForm({
			merchantName: testCards.ikea.merchant_name,
			program: testCards.ikea.merchant_name + ' Card',
			cardNumber
		});
		await cardFormPage.submit();

		// Wait for redirect to the new card's detail page
		await expect(page).toHaveURL(/\/cards\/[a-f0-9-]+$/, { timeout: 10000 });
		await cardDetailPage.waitForPageReady();

		// ── Step 2: Delete the card (soft-delete) ────────────────────────────────
		await cardDetailPage.enterEditMode();
		await cardDetailPage.deleteResource();

		// After deletion the app redirects to the wallet list filtered to cards
		await expect(page).toHaveURL(/\/wallet\?type=cards$/, { timeout: 10000 });

		// ── Step 3: Attempt to create a new card with the SAME card number ───────
		await cardsListPage.clickNewButton();

		// Fill the form with the identical card number — will trigger a 409
		await cardFormPage.fillCardForm({
			merchantName: testCards.ikea.merchant_name,
			program: testCards.ikea.merchant_name + ' Card',
			cardNumber
		});

		// Submit and wait for the 409 response (waitForApi accepts any status)
		const createResponse = page.waitForResponse(
			(resp) =>
				resp.url().includes('/api/v1/cards') &&
				resp.request().method() === 'POST',
			{ timeout: 15000 }
		);
		await page.locator('button[type="submit"]').click();
		const resp = await createResponse;
		expect(resp.status()).toBe(409);

		// ── Step 4: Assert the duplicate-warning banner with restore button ──────
		// Banner renders via DuplicateWarningBanner.svelte when warning.deleted === true
		const banner = page.locator('[role="alert"].duplicate-warning');
		await expect(banner).toBeVisible({ timeout: 5000 });

		// Locate the restore button by data-testid (locale-independent — the label is i18n)
		const restoreButton = page.getByTestId('restore-duplicate');
		await expect(restoreButton).toBeVisible({ timeout: 5000 });

		// ── Step 5: Click restore ────────────────────────────────────────────────
		// The handler calls cardsApi.restore(id) then goto(`/cards/${id}`)
		await restoreButton.click();

		// ── Step 6: Assert navigation to the restored card detail page ───────────
		await expect(page).toHaveURL(/\/cards\/[a-f0-9-]+$/, { timeout: 10000 });
		await cardDetailPage.waitForPageReady();

		// The card detail heading should be visible (card is active/restored)
		await expect(cardDetailPage.heading).toBeVisible({ timeout: 5000 });
	});
});
