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
		// Migrated lists render the ResourceTile clickable as <a>/<button>;
		// not-yet-migrated lists still use div[role="button"]. Both carry
		// data-owner, so match on that alone.
		return this.page.locator('[data-owner]');
	}

	get ownedItems(): Locator {
		return this.page.locator('[data-owner="owned"]');
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
			// The legacy list route redirects to /wallet?type=<resource>.
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
		await this.page.waitForURL(
			new RegExp(`\\/wallet\\?type=${this.resourceType}`),
			{ timeout: 5000 }
		);
		await apiResponsePromise;
		await this.waitForPageReady();
	}

	async expectHeading() {
		// Wallet renders a single "Wallet" heading; the active resource is
		// reflected by the ?type= query. Assert we landed on the filtered view.
		await expect(this.page).toHaveURL(
			new RegExp(`\\/wallet\\?type=${this.resourceType}`)
		);
		await expect(this.heading).toContainText(/Wallet/i);
	}

	async clickFirstItem() {
		await this.firstItem.click();
		await expect(this.page).toHaveURL(
			new RegExp(`\\/${this.resourceType}\\/[a-f0-9-]+$`)
		);
		await this.waitForPageReady();
	}

	async clickNewButton() {
		// The unified wallet toolbar no longer carries a per-type "New" button
		// (creation is reached via the nav). Navigate to the create route directly.
		await this.page.goto(`/${this.resourceType}/new`, {
			waitUntil: 'domcontentloaded',
			timeout: 10000
		});
		await expect(this.page).toHaveURL(
			new RegExp(`\\/${this.resourceType}\\/new`)
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
