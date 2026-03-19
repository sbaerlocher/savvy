import type { Page } from '@playwright/test';
import { BasePage } from './base.page';
import type { ResourceType } from './resource-list.page';

const API_PATHS: Record<ResourceType, string> = {
	cards: '/api/v1/cards',
	vouchers: '/api/v1/vouchers',
	'gift-cards': '/api/v1/gift-cards'
};

export class ResourceFormPage extends BasePage {
	readonly resourceType: ResourceType;
	readonly apiPath: string;

	constructor(page: Page, resourceType: ResourceType) {
		super(page);
		this.resourceType = resourceType;
		this.apiPath = API_PATHS[resourceType];
	}

	get merchantSelect() {
		return this.page.locator('input#merchant[role="combobox"]');
	}
	get submitButton() {
		return this.page.locator('button[type="submit"]');
	}
	get programInput() {
		return this.page.locator('input#program');
	}
	get cardNumberInput() {
		return this.page.locator('input#cardNumber');
	}
	get barcodeTypeSelect() {
		return this.page.locator('select#barcodeType');
	}
	get notesField() {
		return this.page.locator('textarea#notes');
	}
	get codeInput() {
		return this.page.locator('input#code');
	}
	get typeSelect() {
		return this.page.locator('select#type');
	}
	get valueInput() {
		return this.page.locator('input#value');
	}
	get currencySelect() {
		return this.page.locator('select#currency');
	}
	get validFromInput() {
		return this.page.locator('input#validFrom');
	}
	get validUntilInput() {
		return this.page.locator('input#validUntil');
	}
	get usageLimitSelect() {
		return this.page.locator('select#usageLimitType');
	}
	get expiresAtInput() {
		return this.page.locator('input#expiresAt');
	}
	get initialBalanceInput() {
		return this.page.locator('input#initialBalance');
	}

	async selectMerchant(merchantName: string) {
		await this.merchantSelect.waitFor({ state: 'visible' });
		await this.merchantSelect.click();
		await this.merchantSelect.fill(merchantName);
		const option = this.page.locator('[role="listbox"] [role="option"]', {
			hasText: merchantName
		});
		await option.waitFor({ state: 'visible', timeout: 5000 });
		await option.click();
		await this.page.locator('[role="listbox"]').waitFor({ state: 'hidden', timeout: 3000 });
	}

	async fillNotes(text: string) {
		if (await this.notesField.isVisible()) {
			await this.notesField.fill(text);
		}
	}

	async submit(): Promise<void> {
		const createResponse = this.waitForApi(this.apiPath, 'POST', 15000);
		await this.submitButton.click();
		await createResponse;
	}

	async fillCardForm(data: {
		merchantName: string;
		program?: string;
		cardNumber: string;
		barcodeType?: string;
		notes?: string;
	}) {
		await this.selectMerchant(data.merchantName);
		if (data.program) {
			await this.programInput.fill(data.program);
		}
		await this.cardNumberInput.fill(data.cardNumber);
		if (data.barcodeType && (await this.barcodeTypeSelect.isVisible())) {
			await this.barcodeTypeSelect.selectOption(data.barcodeType);
		}
		if (data.notes) {
			await this.fillNotes(data.notes);
		}
	}

	async fillVoucherForm(data: {
		merchantName: string;
		code: string;
		type?: string;
		value?: number;
		currency?: string;
		validFrom: string;
		validUntil: string;
		usageLimitType?: string;
		expiresAt?: string;
		notes?: string;
	}) {
		await this.selectMerchant(data.merchantName);
		await this.codeInput.fill(data.code);

		if (data.type) {
			await this.typeSelect.selectOption(data.type);
		}
		if (data.value !== undefined) {
			await this.valueInput.fill(data.value.toString());
		}
		if (data.currency && (await this.currencySelect.isVisible())) {
			await this.currencySelect.waitFor({ state: 'visible', timeout: 2000 });
			await this.currencySelect.selectOption(data.currency);
		}
		await this.validFromInput.fill(data.validFrom);
		await this.validUntilInput.fill(data.validUntil);

		if (data.usageLimitType && (await this.usageLimitSelect.isVisible())) {
			await this.usageLimitSelect.selectOption(data.usageLimitType);
		}
		if (data.expiresAt && (await this.expiresAtInput.isVisible())) {
			await this.expiresAtInput.fill(data.expiresAt);
		}
		if (data.notes) {
			await this.fillNotes(data.notes);
		}
	}

	async fillGiftCardForm(data: {
		merchantName: string;
		cardNumber: string;
		initialBalance: number;
		currency?: string;
		notes?: string;
	}) {
		await this.selectMerchant(data.merchantName);
		await this.cardNumberInput.fill(data.cardNumber);
		await this.initialBalanceInput.fill(data.initialBalance.toString());

		if (data.currency && (await this.currencySelect.isVisible())) {
			await this.currencySelect.selectOption(data.currency);
		}
		if (data.notes) {
			await this.fillNotes(data.notes);
		}
	}
}
