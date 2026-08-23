import { describe, it, expect } from 'vitest';
import { FullscreenState } from './fullscreen-state.svelte';

// Guards the rule the enlarged barcode overlay runs on: dismissing must not
// contradict the real orientation, or a device that is still sideways (an iPad
// held in its usual landscape) is left with the overlay stuck open or stuck
// hidden until it is turned twice.
describe('FullscreenState', () => {
	it('opens when the device turns landscape', () => {
		const s = new FullscreenState();
		s.setOrientation(true);

		expect(s.visible).toBe(true);
	});

	it('reports whether the orientation actually changed', () => {
		const s = new FullscreenState();

		expect(s.setOrientation(true)).toBe(true);
		// An incidental resize re-reports the same orientation.
		expect(s.setOrientation(true)).toBe(false);
	});

	it('closes on dismissal and stays closed while still landscape', () => {
		const s = new FullscreenState();
		s.setOrientation(true);
		s.close();

		expect(s.visible).toBe(false);
		s.setOrientation(true);
		expect(s.visible).toBe(false);
	});

	it('reopens after turning back and out again', () => {
		const s = new FullscreenState();
		s.setOrientation(true);
		s.close();
		s.setOrientation(false);
		s.setOrientation(true);

		expect(s.visible).toBe(true);
	});

	it('is dismissable when the device is already landscape at load', () => {
		const s = new FullscreenState();
		// checkOrientation() on mount, e.g. an iPad in its usual orientation.
		s.setOrientation(true);
		expect(s.visible).toBe(true);

		s.close();
		expect(s.visible).toBe(false);
	});

	it('closes a manually opened overlay in portrait', () => {
		const s = new FullscreenState();
		s.openManually();
		expect(s.visible).toBe(true);

		s.close();
		expect(s.visible).toBe(false);
	});
});
