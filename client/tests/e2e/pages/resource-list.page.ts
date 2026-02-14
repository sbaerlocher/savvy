import type { Locator, Page } from '@playwright/test';
import { expect } from '@playwright/test';
import { BasePage } from './base.page';

export type ResourceType = 'cards' | 'vouchers' | 'gift-cards';

const RESOURCE_CONFIG = {
	cards: { apiPath: '/api/v1/cards', headingPattern: /Cards|Karten/i },
	vouchers: {
		apiPath: '/api/v1/vouchers',
		headingPattern: /Vouchers|Gutscheine/i
	},
	'gift-cards': {
		apiPath: '/api/v1/gift-cards',
		headingPattern: /Gift Cards|Geschenkkarten/i
	}
} as const;

export class ResourceListPage extends BasePage {
	readonly resourceType: ResourceType;
	readonly config: (typeof RESOURCE_CONFIG)[ResourceType];

	constructor(page: Page, resourceType: ResourceType) {
		super(page);
		this.resourceType = resourceType;
		this.config = RESOURCE_CONFIG[resourceType];
	}

	get heading(): Locator {
		return this.page.locator('h1').first();
	}

	get items(): Locator {
		return this.page.locator('div[role="button"][data-owner]');
	}

	get ownedItems(): Locator {
		return this.page.locator('div[role="button"][data-owner="owned"]');
	}

	get firstItem(): Locator {
		return this.items.first();
	}

	get newButton(): Locator {
		return this.page.locator(`a[href="/${this.resourceType}/new"]`).first();
	}

	get searchInput(): Locator {
		return this.page
			.locator('input[placeholder*="Search" i], input[placeholder*="Suchen" i]')
			.first();
	}

	get selectModeButton(): Locator {
		return this.page.getByRole('button', { name: /Select|Auswählen/i }).first();
	}

	get selectAllButton(): Locator {
		return this.page.getByText(/Select All|Alle auswählen/i).first();
	}

	get filterButton(): Locator {
		return this.page
			.locator(
				'button:has-text("filtern"), button:has-text("Filter"), button[aria-label*="Filter" i]'
			)
			.first();
	}

	async goto() {
		// Start listening for API response BEFORE navigating to avoid race condition
		// The "New" button only renders after cards data is loaded (cards.length > 0)
		const apiResponsePromise = this.page.waitForResponse(
			(resp) => resp.url().includes(this.config.apiPath) && resp.status() < 400,
			{ timeout: 15000 }
		);

		try {
			await this.page.goto(`/${this.resourceType}`, {
				waitUntil: 'domcontentloaded',
				timeout: 10000
			});
		} catch (error) {
			const errorMessage =
				error instanceof Error ? error.message : String(error);
			if (!errorMessage.includes('interrupted by another navigation')) {
				throw error;
			}
		}
		await this.page.waitForURL(new RegExp(`\\/${this.resourceType}`), {
			timeout: 5000
		});
		await apiResponsePromise;
		await this.waitForPageReady();
	}

	async expectHeading() {
		await expect(this.heading).toContainText(this.config.headingPattern);
	}

	async clickFirstItem() {
		await this.firstItem.click();
		await expect(this.page).toHaveURL(
			new RegExp(`\\/${this.resourceType}\\/[a-f0-9-]+$`)
		);
		await this.waitForPageReady();
	}

	async clickNewButton() {
		await this.newButton.waitFor({ state: 'visible', timeout: 5000 });
		await this.newButton.click();
		const resourceSingular = this.resourceType.replace(
			'gift-cards',
			'gift-cards'
		);
		await expect(this.page).toHaveURL(
			new RegExp(`\\/${resourceSingular}\\/new`)
		);
		await this.waitForPageReady();
	}

	async search(term: string) {
		await expect(this.searchInput).toBeVisible({ timeout: 5000 });
		await this.searchInput.fill(term);
		// Wait for debounce to settle - search filters client-side
		await this.page.waitForFunction(() => true, null, { timeout: 600 });
	}

	async enterSelectMode() {
		await this.selectModeButton.click();
		// App uses ring borders for selection (no checkboxes) - wait for BatchPanel with "Select all" button
		await expect(
			this.page.locator('text=/Alle auswählen|Select all/i').first()
		).toBeVisible({
			timeout: 3000
		});
	}

	async selectItemByIndex(index: number) {
		// In select mode, clicking a card item toggles its selection (ring-2 ring-cyan-500)
		await this.items.nth(index).click();
	}

	async selectOwnedItemByIndex(index: number) {
		// Select only owned items (not shared from other users)
		await this.ownedItems.nth(index).click();
	}
}
