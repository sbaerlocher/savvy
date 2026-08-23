import { describe, it, expect, vi } from 'vitest';
import { render } from '@testing-library/svelte';

// `platform` is a module constant evaluated on first import, so the user agent
// is set before the component (and the platform module) load. Resetting the
// whole module graph per case would reload Svelte itself and split its runtime
// state, so each platform gets its own isolated test file section instead.
const IOS_UA =
	'Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1';

Object.defineProperty(navigator, 'userAgent', {
	value: IOS_UA,
	configurable: true
});

const { default: ResourceActions } = await import('./ResourceActions.svelte');

const base = { isOffline: false, isFavorite: false, canEdit: true };

describe('ResourceActions on iOS', () => {
	it('renders the bare glyph row for an owner', () => {
		const { container } = render(ResourceActions, {
			props: { ...base, canShare: true, hasMore: true }
		});

		// Four bare accent glyphs: favorite, share, edit, more — no boxed chrome.
		expect(container.querySelectorAll('button')).toHaveLength(4);
		expect(container.querySelector('button')?.className).not.toContain(
			'border'
		);
	});

	it('omits share and more for a recipient', () => {
		const { container } = render(ResourceActions, {
			props: { ...base, canShare: false, hasMore: false }
		});

		expect(container.querySelectorAll('button')).toHaveLength(2);
	});

	it('hides edit when the recipient may not edit', () => {
		const { container } = render(ResourceActions, {
			props: { ...base, canEdit: false, canShare: false, hasMore: false }
		});

		expect(container.querySelectorAll('button')).toHaveLength(1);
	});

	it('fires the share and more handlers', () => {
		const onshare = vi.fn();
		const onmore = vi.fn();
		const { container } = render(ResourceActions, {
			props: { ...base, canShare: true, hasMore: true, onshare, onmore }
		});

		const buttons = container.querySelectorAll('button');
		(buttons[1] as HTMLButtonElement).click();
		(buttons[3] as HTMLButtonElement).click();

		expect(onshare).toHaveBeenCalledOnce();
		expect(onmore).toHaveBeenCalledOnce();
	});

	it('disables the actions while offline', () => {
		const { container } = render(ResourceActions, {
			props: { ...base, isOffline: true, canShare: true, hasMore: true }
		});

		for (const button of container.querySelectorAll('button')) {
			expect(button).toBeDisabled();
		}
	});
});
