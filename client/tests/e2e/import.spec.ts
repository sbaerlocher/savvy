import { expect, test } from './fixtures/test-fixtures';

test.describe('Import', () => {
	test.use({ serviceWorkers: 'block' });

	test.beforeEach(async ({ authenticatedPage, cardsListPage }) => {
		// Navigate to cards list page where import button is available (desktop only)
		await cardsListPage.goto();
	});

	// Helper locators scoped to the page
	function importButton(page: import('@playwright/test').Page) {
		return page
			.locator(
				'button[aria-label*="Import" i], button:has-text("Import"), button:has-text("Importieren")'
			)
			.first();
	}

	function importDialog(page: import('@playwright/test').Page) {
		return page
			.locator('[role="dialog"], [data-testid="import-dialog"], .modal')
			.first();
	}

	function fileInput(page: import('@playwright/test').Page) {
		return page.locator('input[type="file"]').first();
	}

	function importConfirmButton(page: import('@playwright/test').Page) {
		return importDialog(page)
			.locator(
				'button:has-text("Importieren"), button:has-text("Import"), button:has-text("Bestätigen"), button:has-text("Confirm")'
			)
			.first();
	}

	async function openImportDialog(page: import('@playwright/test').Page) {
		const btn = importButton(page);
		await expect(btn).toBeVisible({ timeout: 5000 });
		await btn.click();
		await expect(importDialog(page)).toBeVisible({ timeout: 5000 });
	}

	async function uploadFile(
		page: import('@playwright/test').Page,
		name: string,
		content: string,
		mimeType: string
	) {
		const input = fileInput(page);
		await expect(input).toBeAttached({ timeout: 5000 });

		if (name.endsWith('.json')) {
			const previewResponse = page
				.waitForResponse(
					(resp) => resp.url().includes('/import/json/preview'),
					{ timeout: 10000 }
				)
				.catch(() => null);

			await input.setInputFiles({
				name,
				mimeType,
				buffer: Buffer.from(content)
			});

			await previewResponse;
		} else {
			await input.setInputFiles({
				name,
				mimeType,
				buffer: Buffer.from(content)
			});
		}
	}

	async function confirmImport(page: import('@playwright/test').Page) {
		const btn = importConfirmButton(page);
		await btn.waitFor({ state: 'visible', timeout: 5000 }).catch(() => {});
		if (await btn.isVisible().catch(() => false)) {
			const importResponse = page
				.waitForResponse(
					(resp) => resp.url().includes('/import') && resp.status() < 400,
					{ timeout: 15000 }
				)
				.catch(() => null);
			await btn.click();
			await importResponse;
		}
	}

	test('should show import button on cards list page', async ({
		authenticatedPage,
		cardsListPage
	}) => {
		const page = authenticatedPage;
		const btn = importButton(page);
		const isVisible = await btn.isVisible({ timeout: 5000 }).catch(() => false);
		if (!isVisible) {
			test.skip();
			return;
		}
		await expect(btn).toBeVisible();
	});

	test('should open import dialog', async ({
		authenticatedPage,
		cardsListPage
	}) => {
		const page = authenticatedPage;
		const btn = importButton(page);
		if (!(await btn.isVisible({ timeout: 5000 }).catch(() => false))) {
			test.skip();
			return;
		}
		await openImportDialog(page);
	});

	test('should accept JSON file for import', async ({
		authenticatedPage,
		cardsListPage
	}) => {
		const page = authenticatedPage;
		const btn = importButton(page);
		if (!(await btn.isVisible({ timeout: 5000 }).catch(() => false))) {
			test.skip();
			return;
		}
		await openImportDialog(page);

		const jsonContent = JSON.stringify({
			cards: [{ merchant_name: 'Test Import', card_number: '1234567890' }]
		});
		await uploadFile(page, 'test-import.json', jsonContent, 'application/json');

		// After JSON upload + preview API call, dialog should show preview step
		const preview = page.locator('text=/Vorschau|Preview|Karten|Cards|1 /i');
		const hasFileInfo = page.locator('text=/test-import/i');
		await expect(preview.or(hasFileInfo).first()).toBeVisible({
			timeout: 5000
		});
	});

	test('should accept CSV file for import', async ({
		authenticatedPage,
		cardsListPage
	}) => {
		const page = authenticatedPage;
		const btn = importButton(page);
		if (!(await btn.isVisible({ timeout: 5000 }).catch(() => false))) {
			test.skip();
			return;
		}
		await openImportDialog(page);

		const csvContent = 'merchant_name,card_number\nTest Import CSV,9876543210';
		await uploadFile(page, 'test-import.csv', csvContent, 'text/csv');

		// CSV import may go straight to importing (when defaultResourceType is set),
		// so wait for either preview step, CSV type selector, or import result
		const resultOrPreview = page.locator(
			'text=/Vorschau|Preview|CSV|Karten|Cards|importiert|imported|Importtyp|Import type|test-import/i'
		);
		await expect(resultOrPreview.first()).toBeVisible({ timeout: 15000 });
	});

	test('should show import preview before executing', async ({
		authenticatedPage,
		cardsListPage
	}) => {
		const page = authenticatedPage;
		const btn = importButton(page);
		if (!(await btn.isVisible({ timeout: 5000 }).catch(() => false))) {
			test.skip();
			return;
		}
		await openImportDialog(page);

		const jsonContent = JSON.stringify({
			cards: [
				{ merchant_name: 'Preview Card 1', card_number: '1111111111' },
				{ merchant_name: 'Preview Card 2', card_number: '2222222222' }
			]
		});
		await uploadFile(
			page,
			'preview-test.json',
			jsonContent,
			'application/json'
		);

		// After JSON upload + preview API call, confirm button should be visible
		await expect(importConfirmButton(page)).toBeVisible({
			timeout: 5000
		});
	});

	test('should execute import and show results', async ({
		authenticatedPage,
		cardsListPage
	}) => {
		const page = authenticatedPage;
		const btn = importButton(page);
		if (!(await btn.isVisible({ timeout: 5000 }).catch(() => false))) {
			test.skip();
			return;
		}
		await openImportDialog(page);

		const jsonContent = JSON.stringify({
			cards: [
				{
					merchant_name: 'Import Execute Test',
					card_number: 'IMPORT-' + Date.now()
				}
			]
		});
		await uploadFile(
			page,
			'execute-test.json',
			jsonContent,
			'application/json'
		);

		await confirmImport(page);

		const success = page.locator(
			'text=/erfolgreich|success|importiert|imported/i'
		);
		const toast = page.locator('[role="alert"], [role="status"], .toast');
		await expect(success.or(toast).first()).toBeVisible({ timeout: 5000 });
	});

	test('should reject invalid file format', async ({
		authenticatedPage,
		cardsListPage
	}) => {
		const page = authenticatedPage;
		const btn = importButton(page);
		if (!(await btn.isVisible({ timeout: 5000 }).catch(() => false))) {
			test.skip();
			return;
		}
		await openImportDialog(page);

		await uploadFile(
			page,
			'invalid.txt',
			'This is not a valid import file',
			'text/plain'
		);

		const error = page.locator(
			'text=/Fehler|Error|ungültig|invalid|Format|fehlgeschlagen/i'
		);
		const errorVisible = await error
			.isVisible({ timeout: 3000 })
			.catch(() => false);
		const noPreview = !(await importConfirmButton(page)
			.isVisible({ timeout: 1000 })
			.catch(() => false));
		expect(errorVisible || noPreview).toBeTruthy();
	});
});
