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
		return this.page
			.locator('button:has-text("Edit"), button:has-text("Bearbeiten")')
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

	get shareSection(): Locator {
		return this.page.locator('text=/Teilen|Share/i').first();
	}

	get transferSection(): Locator {
		return this.page.locator('text=/Besitzerwechsel|Transfer/i').first();
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
		await this.page.waitForURL(new RegExp(`\\/${this.resourceType}\\/?$`), {
			timeout: 10000
		});
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
		await expect(this.favoriteButton).toContainText(expectedIcon, {
			timeout: 15000
		});
	}

	async isFavorited(): Promise<boolean> {
		const text = await this.favoriteButton.textContent();
		return text?.includes('★') || false;
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

	async addShare(
		email: string,
		options?: { canEdit?: boolean; canDelete?: boolean }
	) {
		const addShareButton = this.page
			.locator('button:has-text("Hinzufügen"), button:has-text("Add")')
			.first();
		await addShareButton.waitFor({ state: 'visible', timeout: 5000 });
		await addShareButton.click();

		const emailInput = this.page.locator('input#share-email-input');
		await emailInput.waitFor({ state: 'visible', timeout: 5000 });
		await emailInput.fill(email);

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
