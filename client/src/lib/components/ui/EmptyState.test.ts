import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import { createRawSnippet } from 'svelte';
import EmptyState from './EmptyState.svelte';

describe('EmptyState', () => {
	it('renders the title', () => {
		render(EmptyState, { props: { title: 'No logs yet' } });

		expect(screen.getByText('No logs yet')).toBeInTheDocument();
	});

	it('renders nothing but the title when the optional props are omitted', () => {
		const { container } = render(EmptyState, {
			props: { title: 'Nothing here' }
		});

		expect(container.querySelector('svg')).toBeNull();
		expect(container.textContent?.trim()).toBe('Nothing here');
	});

	it('renders the description when given', () => {
		render(EmptyState, {
			props: {
				title: 'No cards',
				description: 'Add your first card to get started.'
			}
		});

		expect(
			screen.getByText('Add your first card to get started.')
		).toBeInTheDocument();
	});

	it('puts the icon path into the svg', () => {
		const { container } = render(EmptyState, {
			props: { title: 'No cards', icon: 'M4 6h16M4 12h16' }
		});

		const path = container.querySelector('svg path');
		expect(path).not.toBeNull();
		expect(path).toHaveAttribute('d', 'M4 6h16M4 12h16');
	});

	it('renders the action snippet', () => {
		render(EmptyState, {
			props: {
				title: 'No cards',
				action: createRawSnippet(() => ({
					render: () => '<button>Add card</button>'
				}))
			}
		});

		expect(
			screen.getByRole('button', { name: 'Add card' })
		).toBeInTheDocument();
	});
});
