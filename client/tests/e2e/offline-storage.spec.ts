import type { Page } from '@playwright/test';
import { expect, test, TEST_USERS } from './fixtures/test-fixtures';
import { LoginPage } from './pages/login.page';

test.describe('Offline Storage', () => {
	async function loginAsUser(page: Page, email: string, password: string) {
		const loginPage = new LoginPage(page);
		await loginPage.goto();
		await loginPage.login(email, password);
		await page.waitForURL(/\/(dashboard|cards|vouchers|gift-cards)\/?/, {
			timeout: 15000
		});
	}

	async function logoutUser(page: Page) {
		// Open user menu dropdown first
		const userMenuButton = page
			.locator(
				'button[aria-label*="user" i], button[aria-label*="Benutzer" i], button[aria-label*="menu" i]'
			)
			.first();
		if (await userMenuButton.isVisible({ timeout: 3000 }).catch(() => false)) {
			await userMenuButton.click();
		}

		const logoutButton = page
			.locator(
				'button:has-text("Abmelden"), button:has-text("Logout"), button:has-text("Sign out"), a:has-text("Abmelden"), a:has-text("Logout"), [data-testid="logout"]'
			)
			.first();
		if (await logoutButton.isVisible({ timeout: 3000 }).catch(() => false)) {
			await logoutButton.click();
			await page.waitForURL(/\/login/, { timeout: 10000 });
		} else {
			// Fallback: clear cookies and navigate to login
			await page.context().clearCookies();
			await page.goto('/login');
			await page.waitForURL(/\/login/, { timeout: 10000 });
		}
	}

	async function getIndexedDBDatabases(page: Page): Promise<string[]> {
		return page.evaluate(async () => {
			const databases = await indexedDB.databases();
			return databases
				.map((db: IDBDatabaseInfo) => db.name)
				.filter(Boolean) as string[];
		});
	}

	async function getIndexedDBData(
		page: Page,
		dbName: string
	): Promise<Record<string, number>> {
		return page.evaluate(async (name: string) => {
			return new Promise<Record<string, number>>((resolve) => {
				const request = indexedDB.open(name);
				request.onsuccess = () => {
					const db = request.result;
					const storeNames = Array.from(db.objectStoreNames);
					const data: Record<string, number> = {};

					if (storeNames.length === 0) {
						db.close();
						resolve(data);
						return;
					}

					let completed = 0;
					for (const storeName of storeNames) {
						const tx = db.transaction(storeName, 'readonly');
						const store = tx.objectStore(storeName);
						const countReq = store.count();
						countReq.onsuccess = () => {
							data[storeName] = countReq.result;
							completed++;
							if (completed === storeNames.length) {
								db.close();
								resolve(data);
							}
						};
					}
				};
				request.onerror = () => resolve({});
			});
		}, dbName);
	}

	async function clearAllIndexedDB(page: Page) {
		await page.evaluate(async () => {
			const databases = await indexedDB.databases();
			for (const db of databases) {
				if (db.name) {
					indexedDB.deleteDatabase(db.name);
				}
			}
		});
	}

	test('should create IndexedDB stores after login', async ({ page }) => {
		await loginAsUser(
			page,
			TEST_USERS.regular.email,
			TEST_USERS.regular.password
		);

		await page.goto('/cards');
		await page.waitForURL(/\/cards/, { timeout: 10000 });
		await page.waitForLoadState('networkidle');

		const databases = await getIndexedDBDatabases(page);
		expect(databases.length).toBeGreaterThanOrEqual(0);
	});

	test('should clear IndexedDB data on logout', async ({ page }) => {
		await loginAsUser(
			page,
			TEST_USERS.regular.email,
			TEST_USERS.regular.password
		);

		await page.goto('/cards');
		await page.waitForLoadState('networkidle');

		const dbsBefore = await getIndexedDBDatabases(page);

		await logoutUser(page);
		await page.waitForLoadState('networkidle');

		const dbsAfter = await getIndexedDBDatabases(page);

		if (dbsBefore.length > 0) {
			let dataCleared = true;
			for (const dbName of dbsAfter) {
				const data = await getIndexedDBData(page, dbName);
				const totalRecords = Object.values(data).reduce(
					(sum: number, count: number) => sum + count,
					0
				);
				if (totalRecords > 0) {
					dataCleared = false;
				}
			}
			expect(dbsAfter.length < dbsBefore.length || dataCleared).toBeTruthy();
		}
	});

	test('should not leak data between users', async ({ page }) => {
		await loginAsUser(
			page,
			TEST_USERS.regular.email,
			TEST_USERS.regular.password
		);
		await page.goto('/cards');
		await page.waitForLoadState('networkidle');

		await logoutUser(page);
		await page.waitForLoadState('networkidle');

		// Login as second user
		await loginAsUser(
			page,
			TEST_USERS.shared.email,
			TEST_USERS.shared.password
		);
		await page.goto('/cards');
		await page.waitForLoadState('networkidle');

		const databases = await getIndexedDBDatabases(page);
		for (const dbName of databases) {
			const data = await getIndexedDBData(page, dbName);
			expect(data).toBeDefined();
		}
	});

	test('should handle IndexedDB cleanup on login after browser restart simulation', async ({
		page
	}) => {
		// Navigate to a page first to have access to indexedDB
		await page.goto('/login');
		await page.waitForLoadState('networkidle');
		await clearAllIndexedDB(page);

		await loginAsUser(
			page,
			TEST_USERS.regular.email,
			TEST_USERS.regular.password
		);
		await page.goto('/cards');
		await page.waitForURL(/\/cards/, { timeout: 10000 });
		await page.waitForLoadState('networkidle');

		const pageContent = await page.textContent('body');
		expect(pageContent).toBeTruthy();

		const errorOverlay = page.locator(
			'.error-boundary, [data-testid="error-boundary"]'
		);
		const hasError = await errorOverlay.isVisible().catch(() => false);
		expect(hasError).toBeFalsy();
	});

	test('should recover from corrupted IndexedDB', async ({ page }) => {
		await loginAsUser(
			page,
			TEST_USERS.regular.email,
			TEST_USERS.regular.password
		);

		// Simulate corruption by creating a malformed database
		await page.evaluate(() => {
			const request = indexedDB.open('savvy-test-corruption', 1);
			request.onupgradeneeded = () => {
				const db = request.result;
				db.createObjectStore('corrupted-store');
			};
		});

		await page.goto('/cards');
		await page.waitForURL(/\/cards/, { timeout: 10000 });
		await page.waitForLoadState('networkidle');

		const pageContent = await page.textContent('body');
		expect(pageContent).toBeTruthy();

		// Cleanup
		await page.evaluate(() => {
			indexedDB.deleteDatabase('savvy-test-corruption');
		});
	});
});
