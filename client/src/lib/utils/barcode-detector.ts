import { BarcodeDetector as PolyfillBarcodeDetector } from 'barcode-detector/ponyfill';
import { logger } from '$lib/utils/logger';
import { SCANNER_FORMATS } from '$lib/utils/barcode';

const detectorLogger = logger.child('BarcodeDetector');

export interface DetectedBarcode {
	rawValue: string;
	format: string;
}

export interface BarcodeDetectorWrapper {
	readonly method: 'native' | 'polyfill';
	detect(video: HTMLVideoElement): Promise<DetectedBarcode[]>;
	destroy(): void;
}

/**
 * Create a barcode detector that prefers the native BarcodeDetector API
 * (Chrome 83+, Chrome Android) and falls back to the WASM polyfill
 * with canvas frame capture for iOS Safari and Firefox compatibility.
 *
 * The polyfill path always uses canvas because `createImageBitmap(HTMLVideoElement)`
 * is unsupported on iOS Safari, causing `detect(video)` to silently return empty results.
 */
export function createBarcodeDetector(
	videoWidth: number,
	videoHeight: number
): BarcodeDetectorWrapper {
	// eslint-disable-next-line @typescript-eslint/no-explicit-any
	const NativeDetector = (globalThis as any).BarcodeDetector as
		| typeof PolyfillBarcodeDetector
		| undefined;

	const formats = [...SCANNER_FORMATS];

	if (NativeDetector) {
		const detector = new NativeDetector({ formats });
		detectorLogger.info('Using native BarcodeDetector');

		return {
			method: 'native',
			async detect(video: HTMLVideoElement): Promise<DetectedBarcode[]> {
				const results = await detector.detect(video);
				return results.map((b) => ({ rawValue: b.rawValue, format: b.format }));
			},
			destroy() {
				// Native detector has no resources to clean up
			}
		};
	}

	// Polyfill path: Firefox, all iOS browsers (WebKit limitation)
	// Always use canvas frame capture — detect(video) fails on iOS Safari
	// because createImageBitmap(HTMLVideoElement) is not supported in WebKit.
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
