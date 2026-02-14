import type { Page } from '@playwright/test';

/**
 * Test data for Cards
 * Merchant names must match seeded data in cmd/seed/main.go
 */
export const testCards = {
	ikea: {
		merchant_name: 'Migros',
		card_number: '1234567890123',
		barcode_type: 'CODE128',
		notes: 'Test Migros loyalty card'
	},
	mediamarkt: {
		merchant_name: 'Media Markt',
		card_number: '9876543210982',
		barcode_type: 'EAN13',
		notes: 'Test Media Markt card'
	},
	starbucks: {
		merchant_name: 'Coop',
		card_number: '5555555555555',
		barcode_type: 'CODE128',
		notes: 'Test Coop card'
	}
};

/**
 * Test data for Vouchers
 * Merchant names must match seeded data in cmd/seed/main.go
 */
export const testVouchers = {
	amazon: {
		merchant_name: 'Digitec',
		voucher_code: 'DGTC-TEST-1234',
		barcode_type: 'CODE128',
		type: 'fixed_amount',
		value: 50.0,
		currency: 'EUR',
		valid_from: new Date().toISOString().split('T')[0],
		valid_until: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000)
			.toISOString()
			.split('T')[0],
		expires_at: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000)
			.toISOString()
			.split('T')[0],
		notes: 'Test Digitec voucher'
	},
	zalando: {
		merchant_name: 'Manor',
		voucher_code: 'MNR-TEST-5678',
		barcode_type: 'CODE128',
		type: 'fixed_amount',
		value: 25.0,
		currency: 'EUR',
		valid_from: new Date().toISOString().split('T')[0],
		valid_until: new Date(Date.now() + 60 * 24 * 60 * 60 * 1000)
			.toISOString()
			.split('T')[0],
		expires_at: new Date(Date.now() + 60 * 24 * 60 * 60 * 1000)
			.toISOString()
			.split('T')[0],
		notes: 'Test Manor voucher'
	}
};

/**
 * Test data for Gift Cards
 * Merchant names must match seeded data in cmd/seed/main.go
 */
export const testGiftCards = {
	appleStore: {
		merchant_name: 'Galaxus',
		card_number: 'GLXS-GIFT-1234',
		barcode_type: 'CODE128',
		initial_balance: 100.0,
		currency: 'EUR',
		notes: 'Test Galaxus gift card'
	},
	steam: {
		merchant_name: 'Interdiscount',
		card_number: 'INTD-GIFT-5678',
		barcode_type: 'CODE128',
		initial_balance: 50.0,
		currency: 'EUR',
		notes: 'Test Interdiscount gift card'
	}
};

/**
 * Test data for Gift Card Transactions
 */
export const testTransactions = {
	purchase1: {
		transaction_type: 'purchase',
		amount: 15.5,
		description: 'Test purchase - Coffee'
	},
	purchase2: {
		transaction_type: 'purchase',
		amount: 25.0,
		description: 'Test purchase - Book'
	},
	topup: {
		transaction_type: 'topup',
		amount: 50.0,
		description: 'Test top-up'
	}
};

export function randomEmail(): string {
	const timestamp = Date.now();
	return `test-${timestamp}@example.com`;
}

export function uniqueEmail(browserPrefix = ''): string {
	const timestamp = Date.now();
	const random = Math.random().toString(36).substring(2, 7);
	const prefix = browserPrefix ? `${browserPrefix}-` : '';
	return `test-${prefix}${timestamp}-${random}@example.com`;
}

export function randomCardNumber(): string {
	return Math.random().toString().slice(2, 15);
}

export function uniqueCardNumber(browserPrefix = ''): string {
	const timestamp = Date.now();
	const random = Math.random().toString().slice(2, 8);
	const prefix = browserPrefix ? `${browserPrefix.toUpperCase()}-` : '';
	return `${prefix}${timestamp}${random}`;
}

export function randomVoucherCode(): string {
	const prefix = 'TEST';
	const random = Math.random().toString(36).substring(2, 10).toUpperCase();
	return `${prefix}-${random}`;
}

export function uniqueVoucherCode(browserPrefix = ''): string {
	const timestamp = Date.now();
	const random = Math.random().toString(36).substring(2, 8).toUpperCase();
	const prefix = browserPrefix ? `${browserPrefix.toUpperCase()}-` : 'TEST-';
	return `${prefix}${timestamp}-${random}`;
}

export function uniqueMerchantName(
	baseName = 'TestMerchant',
	browserPrefix = ''
): string {
	const timestamp = Date.now();
	const random = Math.random().toString(36).substring(2, 5).toUpperCase();
	const prefix = browserPrefix ? `[${browserPrefix}] ` : '';
	return `${prefix}${baseName} ${timestamp}-${random}`;
}

export function uniqueGiftCardNumber(browserPrefix = ''): string {
	const timestamp = Date.now();
	const random = Math.random().toString(36).substring(2, 8).toUpperCase();
	const prefix = browserPrefix ? `${browserPrefix.toUpperCase()}-` : '';
	return `${prefix}GIFT-${timestamp}-${random}`;
}

/**
 * Helper to create a card and navigate to its detail page.
 * Returns the card number used for creation.
 */
export async function createCardAndNavigate(
	page: Page,
	cardsListPage: {
		goto: () => Promise<void>;
		clickNewButton: () => Promise<void>;
	},
	cardFormPage: {
		fillCardForm: (data: {
			merchantName: string;
			program?: string;
			cardNumber: string;
			notes?: string;
		}) => Promise<void>;
		submit: () => Promise<void>;
	}
): Promise<string> {
	await cardsListPage.goto();
	await cardsListPage.clickNewButton();

	const cardNumber = uniqueCardNumber();
	await cardFormPage.fillCardForm({
		merchantName: testCards.ikea.merchant_name,
		program: testCards.ikea.merchant_name + ' Card',
		cardNumber,
		notes: 'Test card'
	});
	await cardFormPage.submit();
	await page.waitForURL(/\/cards\/[a-f0-9-]+$/, { timeout: 10000 });
	return cardNumber;
}

/**
 * Helper to create multiple cards for batch operation tests.
 */
export async function createMultipleCards(
	page: Page,
	cardsListPage: {
		goto: () => Promise<void>;
		clickNewButton: () => Promise<void>;
	},
	cardFormPage: {
		fillCardForm: (data: {
			merchantName: string;
			program?: string;
			cardNumber: string;
			notes?: string;
		}) => Promise<void>;
		submit: () => Promise<void>;
	},
	count: number
): Promise<void> {
	for (let i = 0; i < count; i++) {
		await cardsListPage.goto();
		await cardsListPage.clickNewButton();
		await cardFormPage.fillCardForm({
			merchantName: testCards.ikea.merchant_name,
			program: testCards.ikea.merchant_name + ' Card',
			cardNumber: uniqueCardNumber(),
			notes: `Batch test card ${i + 1}`
		});
		await cardFormPage.submit();
		await page.waitForURL(/\/cards\/[a-f0-9-]+$/, { timeout: 10000 });
	}
}
