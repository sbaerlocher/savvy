import type { Locator, Page } from '@playwright/test';
import { expect } from '@playwright/test';
import { BasePage } from './base.page';
import type { ResourceType } from './resource-list.page';

const API_PATHS: Record<ResourceType, string> = {
	cards: '/api/v1/cards',
	vouchers: '/api/v1/vouchers',
	'gift-cards': '/api/v1/gift-cards'
};

const DELETE_LABELS: Record<ResourceType, RegExp> = {
	cards: /Karte löschen|Delete|Löschen/,
	vouchers: /Gutschein löschen|Delete|Löschen/,
	'gift-cards': /Geschenkkarte löschen|Delete|Löschen/
};

export class ResourceDetailPage extends BasePage {
	readonly resourceType: ResourceType;
	readonly apiPath: string;

	constructor(page: Page, resourceType: ResourceType) {
		super(page);
		this.resourceType = resourceType;
		this.apiPath = API_PATHS[resourceType];
	}

	get heading(): Locator {
		return this.page.locator('h1').first();
	}

	get editButton(): Locator {
		// The redesigned detail pages expose edit as an icon button (iOS/desktop
		// header action, Android FAB) labelled only via aria-label/title, plus a
		// legacy text button on some screens. Match by accessible name so all
		// variants resolve.
		return this.page
			.getByRole('button', { name: /Edit|Bearbeiten/ })
			.first();
	}

	get deleteButton(): Locator {
		return this.page
			.locator('button')
			.filter({ hasText: DELETE_LABELS[this.resourceType] })
			.first();
	}

	get submitButton(): Locator {
		return this.page.locator('button[type="submit"]');
	}

	get favoriteButton(): Locator {
		return this.page.getByTestId('favorite-button');
	}

	get barcodeCanvas(): Locator {
		return this.page.locator('canvas.barcode-canvas').first();
	}

	/**
	 * The share UI: the Android bottom sheet when one is open, otherwise the
	 * inline section the other layouts render.
	 */
	get shareSection(): Locator {
		const sheet = this.page.locator('[role="dialog"]');
		return sheet.or(this.page.locator('text=/Teilen|Share/i')).first();
	}

	get transferSection(): Locator {
		const sheet = this.page.locator('[role="dialog"]');
		return sheet
			.or(this.page.locator('text=/Besitzerwechsel|Transfer|Übertragen/i'))
			.first();
	}

	async enterEditMode() {
		await this.editButton.click();
		await this.waitForPageReady();
	}

	async save() {
		const updateResponse = this.waitForApi(this.apiPath, 'PATCH', 15000);
		await this.submitButton.click();
		await updateResponse;
		await this.waitForPageReady();
	}

	async deleteResource() {
		await expect(this.deleteButton).toBeVisible({ timeout: 3000 });
		await this.deleteButton.click();
		await this.confirmDeletion(this.apiPath);
		// After delete the detail page navigates to the legacy list route, which
		// redirects to /wallet?type=<resource>. Accept either as the end state.
		await this.page.waitForURL(
			new RegExp(
				`(\\/${this.resourceType}\\/?$|\\/wallet\\?type=${this.resourceType})`
			),
			{ timeout: 10000 }
		);
	}

	async toggleFavorite(expectedIcon: '★' | '☆') {
		await expect(this.favoriteButton).toBeEnabled({ timeout: 5000 });
		const favoriteResponse = this.page.waitForResponse(
			(resp) =>
				resp.url().includes('/favorite') &&
				resp.request().method() === 'POST' &&
				resp.status() < 500,
			{ timeout: 10000 }
		);
		await this.favoriteButton.click();
		await favoriteResponse;
		// The redesigned favorite control is an icon button that carries its
		// state on aria-pressed instead of a ★/☆ glyph in the text content.
		const expectPressed = expectedIcon === '★' ? 'true' : 'false';
		await expect(this.favoriteButton).toHaveAttribute(
			'aria-pressed',
			expectPressed,
			{ timeout: 15000 }
		);
	}

	async isFavorited(): Promise<boolean> {
		return (await this.favoriteButton.getAttribute('aria-pressed')) === 'true';
	}

	async ensureFavoriteState(shouldBeFavorited: boolean) {
		await expect(this.favoriteButton).toBeVisible({ timeout: 10000 });
		await expect(this.favoriteButton).toBeEnabled({ timeout: 5000 });

		const currentlyFavorited = await this.isFavorited();
		if (currentlyFavorited !== shouldBeFavorited) {
			const expectedIcon = shouldBeFavorited ? '★' : '☆';
			await this.toggleFavorite(expectedIcon);
		}
	}

	get moreActionsButton(): Locator {
		return this.page
			.getByRole('button', { name: /Weitere Aktionen|More actions/ })
			.first();
	}

	/**
	 * The native title rows tuck share and transfer behind a ••• more menu
	 * (Android: M3 bottom sheet, iOS: context menu). Opens it when present;
	 * returns false on layouts that keep the actions inline.
	 */
	private async openMoreMenu(): Promise<boolean> {
		if (!(await this.moreActionsButton.isVisible().catch(() => false))) {
			return false;
		}
		await this.moreActionsButton.click();
		return true;
	}

	/**
	 * Opens the share form. The native layouts reach it through the ••• menu;
	 * the inline layouts disclose it via an "Add" button.
	 */
	async openShareForm() {
		// Wait for whichever entry this layout renders — probing right after a
		// navigation would find neither and pick the wrong branch.
		const inlineOpener = this.page
			.locator(
				'button:has-text("Hinzufügen"), button:has-text("Add"), button:has-text("Teilen"), button:has-text("Share")'
			)
			.first();
		await expect(this.moreActionsButton.or(inlineOpener).first()).toBeVisible({
			timeout: 10000
		});
		if (await this.openMoreMenu()) {
			const entry = this.page
				.getByRole('menuitem', { name: /Teilen|Share/i })
				.or(this.page.getByRole('button', { name: /Teilen|Share/ }))
				.filter({ visible: true })
				.first();
			await entry.click();
			return;
		}
		const opener = this.page
			.locator(
				'button:has-text("Hinzufügen"), button:has-text("Add"), button:has-text("Teilen"), button:has-text("Share")'
			)
			.first();
		await opener.waitFor({ state: 'visible', timeout: 5000 });
		await opener.click();
	}

	/**
	 * Opens the transfer form (the email input is ready afterwards). Native
	 * goes through the ••• menu; the inline layouts keep a transfer button.
	 */
	async openTransferForm() {
		const inlineTransfer = this.page
			.locator(
				'button:has-text("Übergeben"), button:has-text("Übertragen"), button:has-text("Transfer")'
			)
			.first();
		await expect(
			this.moreActionsButton.or(inlineTransfer).first()
		).toBeVisible({ timeout: 10000 });
		if (await this.openMoreMenu()) {
			const entry = this.page
				.getByRole('menuitem', {
					name: /Besitzerwechsel|Übertragen|Transfer/i
				})
				.or(this.page.getByRole('button', { name: /Übertragen|Transfer/ }))
				.filter({ visible: true })
				.first();
			await entry.click();
			return;
		}
		const transferButton = this.page
			.locator(
				'button:has-text("Übergeben"), button:has-text("Übertragen"), button:has-text("Transfer")'
			)
			.first();
		await expect(transferButton).toBeVisible({ timeout: 5000 });
		await transferButton.click();
	}

	async addShare(
		email: string,
		options?: { canEdit?: boolean; canDelete?: boolean }
	) {
		await this.openShareForm();

		const emailInput = this.page.locator('input#share-email-input');
		await emailInput.waitFor({ state: 'visible', timeout: 5000 });
		await emailInput.fill(email);
		// Multi-recipient input: commit the typed email to a chip before sharing.
		await emailInput.press('Enter');

		const checkboxes = this.page.locator('input[type="checkbox"]');
		if (options?.canEdit) {
			await checkboxes.nth(0).check();
		}
		if (options?.canDelete) {
			await checkboxes.nth(1).check();
		}

		const shareResponse = this.page
			.waitForResponse(
				(resp) =>
					resp.url().includes(this.apiPath) &&
					resp.request().method() === 'POST',
				{ timeout: 5000 }
			)
			.catch(() => null);

		await this.page
			.locator('button:has-text("Jetzt teilen"), button:has-text("Share Now")')
			.first()
			.click();
		await shareResponse;
	}
}
