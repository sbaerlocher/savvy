import { describe, it, expect } from 'vitest';
import { render } from '@testing-library/svelte';
import Skeleton from './Skeleton.svelte';

describe('Skeleton', () => {
	it('carries the shimmer and neutral fill', () => {
		const { container } = render(Skeleton);

		expect(container.firstElementChild).toHaveClass(
			'animate-pulse',
			'rounded',
			'bg-border-soft'
		);
	});

	it('takes its shape from the call-site class', () => {
		const { container } = render(Skeleton, { props: { class: 'h-8 w-32' } });

		expect(container.firstElementChild).toHaveClass('h-8', 'w-32');
	});

	it('renders no content of its own', () => {
		const { container } = render(Skeleton);

		expect(container.firstElementChild?.textContent).toBe('');
		expect(container.firstElementChild?.children.length).toBe(0);
	});
});
