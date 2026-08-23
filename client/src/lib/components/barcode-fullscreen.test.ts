import { describe, it, expect } from 'vitest';

// Mirrors the fullscreen state rule in BarcodeDisplay.svelte. Dismissing sets
// `dismissed` instead of clearing `isLandscape`, so the overlay never
// contradicts the real orientation — the bug this guards against left a device
// that was already landscape with no way to close the overlay.
function createFullscreenState() {
	let isLandscape = false;
	let manualFullscreen = false;
	let dismissed = false;

	return {
		get visible() {
			return (isLandscape && !dismissed) || manualFullscreen;
		},
		/** checkOrientation(): only a genuine flip resets the dismissal. */
		setOrientation(landscape: boolean) {
			if (landscape === isLandscape) return;
			isLandscape = landscape;
			dismissed = false;
		},
		openManually() {
			manualFullscreen = true;
		},
		close() {
			dismissed = true;
			manualFullscreen = false;
		}
	};
}

describe('barcode fullscreen visibility', () => {
	it('opens when the device turns landscape', () => {
		const s = createFullscreenState();
		s.setOrientation(true);

		expect(s.visible).toBe(true);
	});

	it('closes on dismissal and stays closed while still landscape', () => {
		const s = createFullscreenState();
		s.setOrientation(true);
		s.close();

		expect(s.visible).toBe(false);
		// An incidental resize re-reports the same orientation; no re-open.
		s.setOrientation(true);
		expect(s.visible).toBe(false);
	});

	it('reopens after turning back and out again', () => {
		const s = createFullscreenState();
		s.setOrientation(true);
		s.close();
		s.setOrientation(false);
		s.setOrientation(true);

		expect(s.visible).toBe(true);
	});

	it('is dismissable when the device is already landscape at load', () => {
		const s = createFullscreenState();
		// checkOrientation() on mount, e.g. an iPad in its usual orientation.
		s.setOrientation(true);
		expect(s.visible).toBe(true);

		s.close();
		expect(s.visible).toBe(false);
	});

	it('closes a manually opened overlay in portrait', () => {
		const s = createFullscreenState();
		s.openManually();
		expect(s.visible).toBe(true);

		s.close();
		expect(s.visible).toBe(false);
	});
});
