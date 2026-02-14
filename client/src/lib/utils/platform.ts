import { browser } from '$app/environment';

export type Platform = 'ios' | 'android' | 'other';

function detectPlatform(): Platform {
	if (!browser) return 'other';

	const ua = navigator.userAgent;

	if (
		/iPad|iPhone|iPod/.test(ua) ||
		(navigator.platform === 'MacIntel' && navigator.maxTouchPoints > 1)
	) {
		return 'ios';
	}

	if (/Android/.test(ua)) {
		return 'android';
	}

	return 'other';
}

export const platform: Platform = detectPlatform();
