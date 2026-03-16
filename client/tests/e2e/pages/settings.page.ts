import type { Locator, Page } from '@playwright/test';
import { expect } from '@playwright/test';
import { BasePage } from './base.page';

export class SettingsPage extends BasePage {
	constructor(page: Page) {
		super(page);
	}

	// --- Page heading ---
	get heading(): Locator {
		return this.page
			.locator('h1, h2')
			.filter({ hasText: /Benachrichtigungen|Notifications/i })
			.first();
	}

	// --- Notification preference toggles ---

	/** The notification preferences card (scoped by cyan border) */
	get notificationCard(): Locator {
		return this.page.locator('div[style*="border-left: 6px solid #06b6d4"]');
	}

	/** Helper: finds a toggle row by matching text within the notification card */
	private notifToggleRow(text: RegExp): Locator {
		return this.notificationCard
			.locator('.flex.items-center.justify-between')
			.filter({ hasText: text });
	}

	/** Push Notifications channel toggle */
	get pushNotificationsToggle(): Locator {
		return this.notifToggleRow(
			/Push.?Benachrichtigungen|Push notifications|Notifications push/i
		).locator('button[role="switch"]');
	}

	/** Email Notifications channel toggle */
	get emailNotificationsToggle(): Locator {
		return this.notifToggleRow(
			/E-Mail.?Benachrichtigungen|Email notifications|Notifications par e-mail/i
		).locator('button[role="switch"]');
	}

	/** Push subcategories container (only visible when push enabled) */
	get pushSubcategories(): Locator {
		return this.notificationCard.locator('[data-testid="push-subcategories"]');
	}

	/** Email subcategories container (only visible when email enabled) */
	get emailSubcategories(): Locator {
		return this.notificationCard.locator('[data-testid="email-subcategories"]');
	}

	/** Push Reminders subcategory toggle */
	get pushRemindersToggle(): Locator {
		return this.pushSubcategories.locator('button[role="switch"]').first();
	}

	/** Push Sharing subcategory toggle */
	get pushSharingToggle(): Locator {
		return this.pushSubcategories.locator('button[role="switch"]').nth(1);
	}

	/** Email Reminders subcategory toggle */
	get emailRemindersToggle(): Locator {
		return this.emailSubcategories.locator('button[role="switch"]').first();
	}

	/** Email Sharing subcategory toggle */
	get emailSharingToggle(): Locator {
		return this.emailSubcategories.locator('button[role="switch"]').nth(1);
	}

	/** Notification preferences section heading */
	get notificationSection(): Locator {
		return this.notificationCard.locator('h2').first();
	}

	/** Toggle a notification switch and wait for the API response */
	async toggleNotificationPreference(toggle: Locator): Promise<void> {
		const response = this.page
			.waitForResponse(
				(resp) =>
					resp.url().includes('/api/v1/profile') &&
					resp.request().method() === 'PATCH',
				{ timeout: 10000 }
			)
			.catch(() => null);
		await toggle.click();
		await response;
	}

	// Legacy alias
	get pushToggle(): Locator {
		return this.pushNotificationsToggle;
	}

	// --- Navigation ---
	async goto() {
		const profileResponse = this.page
			.waitForResponse((resp) => resp.url().includes('/api/v1/profile'), {
				timeout: 15000
			})
			.catch(() => null);
		await this.page.goto('/notifications');
		await this.page.waitForURL(/\/notifications/, { timeout: 10000 });
		await profileResponse;
		await this.waitForPageReady();
	}

	async expectSettingsPage() {
		await expect(this.heading).toBeVisible({ timeout: 5000 });
	}
}
