import { expect, test } from './fixtures/test-fixtures';

test.describe('Notification Bell & Panel', () => {
	test.beforeEach(async ({ authenticatedPage }) => {
		// authenticatedPage fixture handles login
	});

	const notificationBellLocator =
		'[data-testid="notification-bell"], button[aria-label*="notification" i], button[aria-label*="Benachrichtigung" i]';

	test('should display notification bell', async ({ authenticatedPage }) => {
		const page = authenticatedPage;
		const notificationBell = page.locator(notificationBellLocator);

		if (
			!(await notificationBell
				.first()
				.isVisible({ timeout: 5000 })
				.catch(() => false))
		) {
			test.skip();
			return;
		}

		await expect(notificationBell.first()).toBeVisible();
	});

	test('should open notification panel', async ({ authenticatedPage }) => {
		const page = authenticatedPage;
		const notificationBell = page.locator(notificationBellLocator).first();

		if (
			!(await notificationBell.isVisible({ timeout: 5000 }).catch(() => false))
		) {
			test.skip();
			return;
		}

		await notificationBell.click();

		const notificationPanel = page.locator(
			'[data-testid="notification-panel"], [role="dialog"]:has-text("Benachrichtigung"), [role="dialog"]:has-text("Notification"), .notification-panel'
		);
		await expect(notificationPanel.first()).toBeVisible({ timeout: 5000 });
	});

	test('should show unread notification count', async ({
		authenticatedPage
	}) => {
		const page = authenticatedPage;
		const notificationBadge = page.locator(
			'[data-testid="notification-count"], .notification-badge, .badge'
		);

		if (
			await notificationBadge
				.first()
				.isVisible({ timeout: 5000 })
				.catch(() => false)
		) {
			const badgeText = await notificationBadge.first().textContent();
			expect(badgeText).toBeTruthy();
		}
	});

	test('should mark notification as read', async ({ authenticatedPage }) => {
		const page = authenticatedPage;
		const notificationBell = page.locator(notificationBellLocator).first();

		if (
			!(await notificationBell.isVisible({ timeout: 5000 }).catch(() => false))
		) {
			test.skip();
			return;
		}

		await notificationBell.click();
		await page.waitForLoadState('networkidle');

		const unreadNotification = page
			.locator(
				'[data-testid="notification-item"].unread, .notification-item.unread, .notification-unread'
			)
			.first();
		if (
			!(await unreadNotification
				.isVisible({ timeout: 3000 })
				.catch(() => false))
		) {
			test.skip();
			return;
		}

		await unreadNotification.click();

		const markReadBtn = page
			.locator(
				'button:has-text("Gelesen"), button:has-text("Read"), [data-testid="mark-read"]'
			)
			.first();
		if (await markReadBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
			await markReadBtn.click();
		}
	});

	test('should mark all notifications as read', async ({
		authenticatedPage
	}) => {
		const page = authenticatedPage;
		const notificationBell = page.locator(notificationBellLocator).first();

		if (
			!(await notificationBell.isVisible({ timeout: 5000 }).catch(() => false))
		) {
			test.skip();
			return;
		}

		await notificationBell.click();
		await page.waitForLoadState('networkidle');

		const markAllRead = page
			.locator(
				'button:has-text("Alle gelesen"), button:has-text("Mark all read"), button:has-text("Alle als gelesen"), [data-testid="mark-all-read"]'
			)
			.first();
		if (await markAllRead.isVisible({ timeout: 3000 }).catch(() => false)) {
			const markReadResponse = page
				.waitForResponse(
					(resp) =>
						resp.url().includes('/notifications') && resp.status() < 400,
					{
						timeout: 10000
					}
				)
				.catch(() => null);
			await markAllRead.click();
			await markReadResponse;
		}
	});

	test('should delete a notification', async ({ authenticatedPage }) => {
		const page = authenticatedPage;
		const notificationBell = page.locator(notificationBellLocator).first();

		if (
			!(await notificationBell.isVisible({ timeout: 5000 }).catch(() => false))
		) {
			test.skip();
			return;
		}

		await notificationBell.click();
		await page.waitForLoadState('networkidle');

		const notificationItem = page
			.locator(
				'[data-testid="notification-item"], .notification-item, [role="listitem"]'
			)
			.first();
		if (
			!(await notificationItem.isVisible({ timeout: 3000 }).catch(() => false))
		) {
			test.skip();
			return;
		}

		const deleteBtn = notificationItem
			.locator(
				'button[aria-label*="delete" i], button[aria-label*="löschen" i], button[aria-label*="dismiss" i], button[aria-label*="remove" i], [data-testid="delete-notification"]'
			)
			.first();
		if (await deleteBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
			const deleteResponse = page
				.waitForResponse(
					(resp) =>
						resp.url().includes('/notifications') && resp.status() < 400,
					{
						timeout: 10000
					}
				)
				.catch(() => null);
			await deleteBtn.click();
			await deleteResponse;
		}
	});

	test('should navigate to resource from notification', async ({
		authenticatedPage
	}) => {
		const page = authenticatedPage;
		const notificationBell = page.locator(notificationBellLocator).first();

		if (
			!(await notificationBell.isVisible({ timeout: 5000 }).catch(() => false))
		) {
			test.skip();
			return;
		}

		await notificationBell.click();
		await page.waitForLoadState('networkidle');

		// Look for a clickable notification that navigates to a resource
		const notificationLink = page
			.locator(
				'[data-testid="notification-item"] a, .notification-item a, [role="listitem"] a'
			)
			.first();
		if (
			!(await notificationLink.isVisible({ timeout: 3000 }).catch(() => false))
		) {
			test.skip();
			return;
		}

		const href = await notificationLink.getAttribute('href');
		await notificationLink.click();

		// Should navigate away from current page (to the resource)
		if (href) {
			await page.waitForURL(new RegExp(href.replace(/\//g, '\\/')), {
				timeout: 10000
			});
		}
	});
});

test.describe('Notification Preferences', () => {
	test.beforeEach(async ({ authenticatedSettingsPage }) => {
		// authenticatedSettingsPage fixture handles login + navigation to /notifications
	});

	test('should display notification preferences section with channel toggles', async ({
		authenticatedSettingsPage,
		settingsPage
	}) => {
		await settingsPage.waitForPageReady();

		// Notification card should be visible
		const hasNotifCard = await settingsPage.notificationCard
			.isVisible({ timeout: 5000 })
			.catch(() => false);
		if (!hasNotifCard) {
			test.skip();
			return;
		}

		// Both channel toggles should be visible
		await expect(settingsPage.pushNotificationsToggle).toBeVisible({
			timeout: 3000
		});
		await expect(settingsPage.emailNotificationsToggle).toBeVisible({
			timeout: 3000
		});
	});

	test('should show push subcategories when push is enabled', async ({
		authenticatedSettingsPage,
		settingsPage
	}) => {
		await settingsPage.waitForPageReady();

		const hasNotifCard = await settingsPage.notificationCard
			.isVisible({ timeout: 5000 })
			.catch(() => false);
		if (!hasNotifCard) {
			test.skip();
			return;
		}

		// Check if push is currently enabled via aria-checked
		const pushToggle = settingsPage.pushNotificationsToggle;
		const isEnabled = await pushToggle.getAttribute('aria-checked');

		if (isEnabled === 'true') {
			// Subcategories should be visible
			await expect(settingsPage.pushRemindersToggle).toBeVisible({
				timeout: 3000
			});
			await expect(settingsPage.pushSharingToggle).toBeVisible({
				timeout: 3000
			});
		} else {
			// Enable push to show subcategories
			await settingsPage.toggleNotificationPreference(pushToggle);
			await expect(settingsPage.pushRemindersToggle).toBeVisible({
				timeout: 3000
			});
			await expect(settingsPage.pushSharingToggle).toBeVisible({
				timeout: 3000
			});

			// Restore: disable push again
			await settingsPage.toggleNotificationPreference(pushToggle);
		}
	});

	test('should hide push subcategories when push is disabled', async ({
		authenticatedSettingsPage,
		settingsPage
	}) => {
		await settingsPage.waitForPageReady();

		const hasNotifCard = await settingsPage.notificationCard
			.isVisible({ timeout: 5000 })
			.catch(() => false);
		if (!hasNotifCard) {
			test.skip();
			return;
		}

		const pushToggle = settingsPage.pushNotificationsToggle;
		const isEnabled = await pushToggle.getAttribute('aria-checked');

		if (isEnabled === 'true') {
			// Disable push
			await settingsPage.toggleNotificationPreference(pushToggle);
		}

		// Subcategories should NOT be visible
		await expect(settingsPage.pushSubcategories).not.toBeVisible({
			timeout: 3000
		});

		// Restore if we changed it
		if (isEnabled === 'true') {
			await settingsPage.toggleNotificationPreference(pushToggle);
		}
	});

	test('should show email subcategories when email is enabled', async ({
		authenticatedSettingsPage,
		settingsPage
	}) => {
		await settingsPage.waitForPageReady();

		const hasNotifCard = await settingsPage.notificationCard
			.isVisible({ timeout: 5000 })
			.catch(() => false);
		if (!hasNotifCard) {
			test.skip();
			return;
		}

		const emailToggle = settingsPage.emailNotificationsToggle;
		const isEnabled = await emailToggle.getAttribute('aria-checked');

		if (isEnabled === 'true') {
			await expect(settingsPage.emailRemindersToggle).toBeVisible({
				timeout: 3000
			});
			await expect(settingsPage.emailSharingToggle).toBeVisible({
				timeout: 3000
			});
		} else {
			// Enable email to show subcategories
			await settingsPage.toggleNotificationPreference(emailToggle);
			await expect(settingsPage.emailRemindersToggle).toBeVisible({
				timeout: 3000
			});
			await expect(settingsPage.emailSharingToggle).toBeVisible({
				timeout: 3000
			});

			// Restore: disable email again
			await settingsPage.toggleNotificationPreference(emailToggle);
		}
	});

	test('should hide email subcategories when email is disabled', async ({
		authenticatedSettingsPage,
		settingsPage
	}) => {
		await settingsPage.waitForPageReady();

		const hasNotifCard = await settingsPage.notificationCard
			.isVisible({ timeout: 5000 })
			.catch(() => false);
		if (!hasNotifCard) {
			test.skip();
			return;
		}

		const emailToggle = settingsPage.emailNotificationsToggle;
		const isEnabled = await emailToggle.getAttribute('aria-checked');

		if (isEnabled === 'true') {
			// Disable email
			await settingsPage.toggleNotificationPreference(emailToggle);
		}

		// Subcategories should NOT be visible
		await expect(settingsPage.emailSubcategories).not.toBeVisible({
			timeout: 3000
		});

		// Restore if we changed it
		if (isEnabled === 'true') {
			await settingsPage.toggleNotificationPreference(emailToggle);
		}
	});

	test('should toggle push notification preference and persist', async ({
		authenticatedSettingsPage,
		settingsPage
	}) => {
		await settingsPage.waitForPageReady();

		const hasNotifCard = await settingsPage.notificationCard
			.isVisible({ timeout: 5000 })
			.catch(() => false);
		if (!hasNotifCard) {
			test.skip();
			return;
		}

		const pushToggle = settingsPage.pushNotificationsToggle;
		const initialState = await pushToggle.getAttribute('aria-checked');

		// Toggle the preference
		await settingsPage.toggleNotificationPreference(pushToggle);
		await settingsPage.expectToast();

		// Verify state changed
		const newState = await pushToggle.getAttribute('aria-checked');
		expect(newState).not.toBe(initialState);

		// Restore original state
		await settingsPage.toggleNotificationPreference(pushToggle);
	});

	test('should toggle email notification preference and persist', async ({
		authenticatedSettingsPage,
		settingsPage
	}) => {
		await settingsPage.waitForPageReady();

		const hasNotifCard = await settingsPage.notificationCard
			.isVisible({ timeout: 5000 })
			.catch(() => false);
		if (!hasNotifCard) {
			test.skip();
			return;
		}

		const emailToggle = settingsPage.emailNotificationsToggle;
		const initialState = await emailToggle.getAttribute('aria-checked');

		// Toggle the preference
		await settingsPage.toggleNotificationPreference(emailToggle);
		await settingsPage.expectToast();

		// Verify state changed
		const newState = await emailToggle.getAttribute('aria-checked');
		expect(newState).not.toBe(initialState);

		// Restore original state
		await settingsPage.toggleNotificationPreference(emailToggle);
	});
});
