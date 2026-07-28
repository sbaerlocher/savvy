import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/svelte';
import Button from './Button.svelte';
import { createRawSnippet } from 'svelte';

// Children snippet helper: the component takes a required `children` Snippet,
// so every render needs one.
const label = (text: string) =>
	createRawSnippet(() => ({ render: () => `<span>${text}</span>` }));

describe('Button', () => {
	it('renders the literal recipe classes, not interpolated ones', () => {
		render(Button, { props: { children: label('Save') } });

		const button = screen.getByRole('button', { name: 'Save' });
		expect(button).toHaveClass('btn');
		expect(button).toHaveClass('btn-primary');
	});

	it('maps variant and size to literal classes without an undefined slot', () => {
		render(Button, {
			props: { variant: 'text-danger', size: 'xs', children: label('Delete') }
		});

		const button = screen.getByRole('button', { name: 'Delete' });
		expect(button).toHaveClass('btn', 'btn-text-danger', 'btn-xs');
		expect(button.className).not.toContain('undefined');
	});

	it('passes the call-site class through alongside the recipe classes', () => {
		render(Button, {
			props: { class: 'w-full', children: label('Login') }
		});

		expect(screen.getByRole('button', { name: 'Login' })).toHaveClass(
			'btn',
			'btn-primary',
			'w-full'
		);
	});

	it('loading disables the button and shows a spinner', () => {
		render(Button, { props: { loading: true, children: label('Saving') } });

		const button = screen.getByRole('button', { name: /Saving/ });
		expect(button).toBeDisabled();
		expect(button).toHaveAttribute('aria-busy', 'true');
		expect(button.querySelector('svg.animate-spin')).not.toBeNull();
	});

	it('renders an anchor when href is given', () => {
		render(Button, {
			props: { href: '/cards', children: label('All cards') }
		});

		const link = screen.getByRole('link', { name: 'All cards' });
		expect(link).toHaveAttribute('href', '/cards');
		expect(link).toHaveClass('btn', 'btn-primary');
	});

	it('blocks navigation and onclick on a disabled anchor', () => {
		const onclick = vi.fn();
		render(Button, {
			props: {
				href: '/cards',
				disabled: true,
				onclick,
				children: label('All cards')
			}
		});

		const link = screen.getByRole('link', { name: 'All cards' });
		expect(link).toHaveAttribute('aria-disabled', 'true');

		// An <a> has no native `disabled`, so the click must be prevented in code.
		const event = new MouseEvent('click', { bubbles: true, cancelable: true });
		link.dispatchEvent(event);

		expect(event.defaultPrevented).toBe(true);
		expect(onclick).not.toHaveBeenCalled();
	});

	it('fires onclick when enabled', async () => {
		const onclick = vi.fn();
		render(Button, { props: { onclick, children: label('Save') } });

		await fireEvent.click(screen.getByRole('button', { name: 'Save' }));

		expect(onclick).toHaveBeenCalledTimes(1);
	});
});
