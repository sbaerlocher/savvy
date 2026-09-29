import '@testing-library/jest-dom';
import { vi, beforeEach } from 'vitest';

// Mock SvelteKit modules
vi.mock('$app/environment', () => ({
	browser: true,
	building: false,
	dev: true,
	version: '1.0.0'
}));

vi.mock('$app/stores', () => ({
	page: {
		subscribe: vi.fn()
	},
	navigating: {
		subscribe: vi.fn()
	},
	updated: {
		subscribe: vi.fn()
	}
}));

// Mock localStorage with actual storage implementation
const storage: Record<string, string> = {};

// Vitest 5 forwards globalThis assignments to happy-dom, whose window exposes
// localStorage/sessionStorage as getter-only — a plain assignment throws, so
// the mocks are installed via defineProperty instead.
const localStorageMock = {
	getItem: (key: string) => storage[key] || null,
	setItem: (key: string, value: string) => {
		storage[key] = value;
	},
	removeItem: (key: string) => {
		delete storage[key];
	},
	clear: () => {
		for (const key in storage) {
			delete storage[key];
		}
	},
	get length() {
		return Object.keys(storage).length;
	},
	key: (index: number) => {
		const keys = Object.keys(storage);
		return keys[index] || null;
	}
} as Storage;

Object.defineProperty(globalThis, 'localStorage', {
	configurable: true,
	writable: true,
	value: localStorageMock
});

// Mock sessionStorage
const sessionStorageData: Record<string, string> = {};

const sessionStorageMock = {
	getItem: (key: string) => sessionStorageData[key] || null,
	setItem: (key: string, value: string) => {
		sessionStorageData[key] = value;
	},
	removeItem: (key: string) => {
		delete sessionStorageData[key];
	},
	clear: () => {
		for (const key in sessionStorageData) {
			delete sessionStorageData[key];
		}
	},
	get length() {
		return Object.keys(sessionStorageData).length;
	},
	key: (index: number) => {
		const keys = Object.keys(sessionStorageData);
		return keys[index] || null;
	}
} as Storage;

Object.defineProperty(globalThis, 'sessionStorage', {
	configurable: true,
	writable: true,
	value: sessionStorageMock
});

// Mock fetch
global.fetch = vi.fn();

// Mock navigator.onLine
Object.defineProperty(global.navigator, 'onLine', {
	writable: true,
	value: true
});

// Reset storage and mocks before each test
beforeEach(() => {
	// Clear storage
	for (const key in storage) {
		delete storage[key];
	}
	for (const key in sessionStorageData) {
		delete sessionStorageData[key];
	}

	// Reset navigator.onLine to true
	Object.defineProperty(global.navigator, 'onLine', {
		writable: true,
		value: true
	});
});
