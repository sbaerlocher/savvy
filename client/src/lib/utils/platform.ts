import { browser } from '$app/environment';

export type Platform = 'ios' | 'android' | 'other';

function detectPlatform(): Platform {
	if (!browser) return 'other';

	const ua = navigator.userAgent;

	// Android is checked first: the MacIntel fallback below identifies iPadOS,
	// which reports a desktop Safari UA, but it also matches a device-emulating
	// desktop browser sending an Android UA — and iPadOS never sends one. With
	// the old order such a client was classified as iOS.
	if (/Android/.test(ua)) {
		return 'android';
	}

	if (
		/iPad|iPhone|iPod/.test(ua) ||
		(navigator.platform === 'MacIntel' && navigator.maxTouchPoints > 1)
	) {
		return 'ios';
	}

	return 'other';
}

export const platform: Platform = detectPlatform();
