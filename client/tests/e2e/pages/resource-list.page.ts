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

	get loadingSpinner(): Locator {
		// LoadingSpinner.svelte — the wallet route's only pulsing logo.
		return this.page.locator('img[alt="Savvy"].animate-pulse');
	}

	get newButton(): Locator {
		return this.page.locator(`a[href="/${this.resourceType}/new"]`).first();
	}

	get searchInput(): Locator {
		// The wallet has no search field of its own — the input lives in the
		// platform chrome (DesktopNav bar, Android header, iOS bottom-nav pill)
		// and only the one for the current platform is rendered/visible. Filter
		// on visibility so we never land on a `hidden sm:flex` desktop field
		// while on mobile.
		return this.page
			.locator('input[placeholder*="Search" i], input[placeholder*="Suchen" i]')
			.filter({ visible: true })
			.first();
	}

	get selectModeButton(): Locator {
		// The entry into select mode is rendered per layout (desktop chrome row,
		// Android M3 chip, plain mobile toolbar button), all labelled
		// `batch.selectMode`. getByRole drops `display:none` subtrees, but
		// Tailwind's `invisible`/opacity variants stay in the accessibility tree,
		// so filter on visibility too rather than trusting DOM order.
		return this.page
			.getByRole('button', { name: /Select|Auswählen/i })
			.filter({ visible: true })
			.first();
	}

	get selectAllButton(): Locator {
		// Order matters: filter to visible buttons BEFORE .first(). BatchPanel
		// renders its desktop side panel (hidden below lg) ahead of the mobile
		// bar in the DOM, so a leading .first() pins the hidden element and any
		// later filter operates on the wrong match. getByRole (not getByText)
		// because WalletView's mobile top bar is an icon-only button whose name
		// comes from aria-label. .first() stays: on a small viewport both that
		// top bar and BatchPanel's bottom bar can be visible at once, which
		// would trip strict mode.
		return this.page
			.getByRole('button', { name: /Select All|Alle auswählen/i })
			.filter({ visible: true })
			.first();
	}

	get filterButton(): Locator {
		// Same layout duplication as selectModeButton: the filter control is
		// rendered per platform (desktop chrome row `hidden sm:flex`, Android M3
		// chip, plain mobile toolbar button), desktop first in DOM order. The old
		// CSS-selector `.first()` is not accessibility-filtered, so below `sm` it
		// resolved to the `display:none` desktop copy and its clicks timed out.
		return this.page
			.getByRole('button', { name: /Filter|filtern/i })
			.filter({ visible: true })
			.first();
	}

	async goto() {
		// Navigate to the wallet directly. Going through the legacy /<resource>
		// route made this flaky: its redirect aborts the initial navigation, so
		// waiting for the wallet URL afterwards could check the URL we came from.
		// The legacy redirect itself is covered by wallet.spec.ts.
		const url = `/wallet?type=${this.resourceType}`;
		// Callers reach goto() right after a submit, while the app is still
		// navigating to the new resource's detail route. Racing that with our own
		// goto() is what makes this flaky, and the observed CI failure is a goto
		// left pending until its timeout rather than one rejecting with an abort —
		// so retry on all three shapes.
		//
		// Budget: these timeouts plus the assertions below must stay under the 60s
		// per-test timeout from playwright.config.ts, otherwise the final attempt
		// is unreachable and a real failure surfaces as the far less diagnostic
		// "Test timeout exceeded". 2 × 10s here + 3 × 10s below = 50s.
		const MAX_GOTO_ATTEMPTS = 2;
		for (let attempt = 0; attempt < MAX_GOTO_ATTEMPTS; attempt++) {
			try {
				await this.page.goto(url, {
					waitUntil: 'domcontentloaded',
					timeout: 10000
				});
				break;
			} catch (error) {
				const message = error instanceof Error ? error.message : String(error);
				const racy =
					message.includes('interrupted by another navigation') ||
					message.includes('net::ERR_ABORTED') ||
					message.includes('Timeout');
				if (!racy || attempt === MAX_GOTO_ATTEMPTS - 1) {
					throw error;
				}
			}
		}
		await expect(this.page).toHaveURL(
			new RegExp(`\\/wallet\\?type=${this.resourceType}`),
			{ timeout: 10000 }
		);
		await this.waitForPageReady();
		// Wait for the list to settle rather than for an API response: the
		// offline-first loader may serve cached data without issuing a fresh
		// request, in which case waiting on the API call hangs until its timeout.
		// The wallet route renders its PageHeader in both the loading and the
		// loaded branch, so the heading proves the component mounted. The spinner
		// clears via revealFirstPage() as soon as the first enabled type has its
		// first page — later pages keep streaming in — so this asserts the list is
		// rendered, not that every page arrived. That covers every valid outcome
		// (tiles, the unfiltered empty state, the filtered "no results" state)
		// without enumerating locale-specific copy.
		await expect(this.heading).toBeVisible({ timeout: 10000 });
		await expect(this.loadingSpinner).toBeHidden({ timeout: 10000 });
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
		// Desktop keeps a permanently visible field in the nav bar; the native
		// layouts collapse it behind a search icon (Android expands an inline
		// header field via the mobile header actions, iOS expands the bottom-nav
		// pill). Open it through the chrome's own control when it isn't showing,
		// so the app's real flow runs. Only the button the current layout renders
		// is in the accessibility tree, so getByRole picks the right one.
		if (!(await this.searchInput.isVisible())) {
			await this.page
				.getByRole('button', { name: /Search|Suchen/i })
				.first()
				.click();
		}
		await expect(this.searchInput).toBeVisible({ timeout: 5000 });
		await this.searchInput.fill(term);
		// How the query reaches the wallet differs per chrome: the native fields
		// debounce their input (250ms) and navigate on their own, while the
		// desktop nav field is a <form> that only navigates on submit. Enter
		// covers both — it submits the desktop form and is a no-op for the others.
		await this.searchInput.press('Enter');
		// The wallet mirrors ?search= into its filter state, so the URL carrying
		// the term is the proof that the filter was actually applied. Waiting for
		// it also absorbs the debounce before the caller counts items.
		await expect(this.page).toHaveURL(
			new RegExp(`search=${encodeURIComponent(term)}`),
			{ timeout: 5000 }
		);
		await this.waitForPageReady();
	}

	async enterSelectMode() {
		await this.selectModeButton.click();
		// App uses ring borders for selection (no checkboxes) - wait for BatchPanel with "Select all" button
		await expect(this.selectAllButton).toBeVisible({ timeout: 3000 });
		// Second, independent signal, and the one the callers actually rely on:
		// ResourceTile swaps its clickable element from the navigating `<a href>`
		// to a `<button>` that toggles the selection. Same `data-owner`, so
		// `items` covers both — assert the tag, which only select mode produces.
		// Not aria-pressed on the toggle: Android hides the whole toolbar row below
		// `sm` once selecting (the contextual top app bar replaces it), so no
		// select-mode *button* stays visible there to carry the attribute.
		await expect(this.firstItem).toHaveJSProperty('tagName', 'BUTTON', {
			timeout: 3000
		});
	}

	async exitSelectMode() {
		// The way out of select mode differs per layout. Desktop keeps the
		// toolbar's toggle (`Select`/`Auswählen`) on screen, but Android below
		// `sm` hides that whole row (ANDROID_SELECT_HIDDEN, WalletView.svelte)
		// and replaces it with a contextual top app bar whose close button is
		// labelled `batch.exitSelectMode` (`Cancel`/`Abbrechen`,
		// WalletView.svelte:1106). Both call the same toggleSelectMode.
		// getByRole is accessibility-filtered, so only the control the current
		// layout actually renders is matched.
		await this.page
			.getByRole('button', {
				name: /Select|Auswählen|Cancel|Abbrechen/i
			})
			.filter({ visible: true })
			.first()
			.click();
		// Leaving select mode restores the navigating <a href> tiles.
		await expect(this.firstItem).toHaveJSProperty('tagName', 'A', {
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
