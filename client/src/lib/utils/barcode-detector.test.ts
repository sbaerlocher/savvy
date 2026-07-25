import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';

// Mock the WASM ponyfill so importing the module under test never touches
// ZXing/WASM. `detect` is a spy we drive per test; `prepareZXingModule` is a
// no-op here.
const detectSpy = vi.fn();
vi.mock('barcode-detector/ponyfill', () => ({
	BarcodeDetector: vi.fn().mockImplementation(() => ({ detect: detectSpy })),
	prepareZXingModule: vi.fn()
}));

import { createBarcodeDetector } from './barcode-detector';

describe('createBarcodeDetector', () => {
	beforeEach(() => {
		detectSpy.mockReset();
	});

	afterEach(() => {
		// eslint-disable-next-line @typescript-eslint/no-explicit-any
		delete (globalThis as any).BarcodeDetector;
	});

	it('always reports method "polyfill", even when a native BarcodeDetector exists', () => {
		// Simulate Chrome Android exposing the (potentially broken) native API.
		// eslint-disable-next-line @typescript-eslint/no-explicit-any
		(globalThis as any).BarcodeDetector = vi.fn();

		const detector = createBarcodeDetector(640, 480);
		expect(detector.method).toBe('polyfill');
	});

	it('delegates detect() to the ZXing polyfill and maps to DetectedBarcode[]', async () => {
		detectSpy.mockResolvedValue([
			{ rawValue: '4006381333931', format: 'ean_13', extra: 'ignored' }
		]);

		const detector = createBarcodeDetector(640, 480);
		const video = document.createElement('video');
		const result = await detector.detect(video);

		expect(detectSpy).toHaveBeenCalledOnce();
		expect(result).toEqual([{ rawValue: '4006381333931', format: 'ean_13' }]);
	});
});
