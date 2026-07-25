import {
	BarcodeDetector as PolyfillBarcodeDetector,
	prepareZXingModule
} from 'barcode-detector/ponyfill';
import { logger } from '$lib/utils/logger';
import { SCANNER_FORMATS } from '$lib/utils/barcode';

// Override WASM location to serve from same origin instead of jsDelivr CDN.
// Without this, CSP `connect-src 'self'` blocks the CDN fetch, causing
// the polyfill to silently fail on iOS Safari and Firefox.
prepareZXingModule({
	overrides: {
		locateFile: (path: string, prefix: string) => {
			if (path.endsWith('.wasm')) {
				return `/${path}`;
			}
			return prefix + path;
		}
	}
});

const detectorLogger = logger.child('BarcodeDetector');

export interface DetectedBarcode {
	rawValue: string;
	format: string;
}

export interface BarcodeDetectorWrapper {
	readonly method: 'polyfill';
	detect(video: HTMLVideoElement): Promise<DetectedBarcode[]>;
	destroy(): void;
}

/**
 * Create a barcode detector backed by the WASM polyfill (ZXing-C++) with
 * canvas frame capture, on every platform.
 *
 * The native `BarcodeDetector` API is deliberately not used: on Chrome Android
 * it constructs successfully even when the Google Play Services barcode module
 * is missing/broken, in which case `detect()` returns `[]` forever (scanner
 * stuck in the orange `detecting` state, never turning green). The polyfill
 * path already runs reliably on iOS Safari and Firefox, so it is the single
 * path for all platforms.
 *
 * The polyfill path always uses canvas because `createImageBitmap(HTMLVideoElement)`
 * is unsupported on iOS Safari, causing `detect(video)` to silently return empty results.
 */
export function createBarcodeDetector(
	videoWidth: number,
	videoHeight: number
): BarcodeDetectorWrapper {
	const formats = [...SCANNER_FORMATS];

	const detector = new PolyfillBarcodeDetector({ formats });
	const canvas = document.createElement('canvas');
	canvas.width = videoWidth || 1280;
	canvas.height = videoHeight || 720;
	const ctx = canvas.getContext('2d', { willReadFrequently: true });
	if (!ctx) {
		throw Object.assign(new Error('Failed to get 2D canvas context'), {
			name: 'NotSupportedError'
		});
	}

	detectorLogger.info('Using polyfill BarcodeDetector with canvas capture', {
		canvas: `${canvas.width}x${canvas.height}`
	});

	return {
		method: 'polyfill',
		async detect(video: HTMLVideoElement): Promise<DetectedBarcode[]> {
			ctx.drawImage(video, 0, 0, canvas.width, canvas.height);
			const results = await detector.detect(canvas);
			return results.map((b) => ({ rawValue: b.rawValue, format: b.format }));
		},
		destroy() {
			canvas.width = 0;
			canvas.height = 0;
		}
	};
}
