/**
 * Visibility rule for the enlarged barcode overlay.
 *
 * Dismissing records a separate flag rather than clearing the orientation:
 * setting `isLandscape = false` on a device that is still sideways contradicts
 * reality, so the overlay would stay hidden until the device was turned twice
 * and could reappear on an incidental resize. The flag suppresses only the
 * current landscape session and resets on a genuine orientation change, which
 * also keeps a device that is already landscape at load dismissable.
 */
export class FullscreenState {
	#isLandscape = $state(false);
	#manual = $state(false);
	#dismissed = $state(false);

	/** True while the overlay should be mounted. */
	get visible(): boolean {
		return (this.#isLandscape && !this.#dismissed) || this.#manual;
	}

	get isLandscape(): boolean {
		return this.#isLandscape;
	}

	/** Report the measured orientation; only a real flip clears a dismissal. */
	setOrientation(landscape: boolean): boolean {
		if (landscape === this.#isLandscape) return false;
		this.#isLandscape = landscape;
		this.#dismissed = false;
		return true;
	}

	/** Tap-to-enlarge, used where the platform offers that affordance. */
	openManually(): void {
		this.#manual = true;
	}

	close(): void {
		this.#dismissed = true;
		this.#manual = false;
	}
}
