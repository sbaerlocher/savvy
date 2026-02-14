import { writable } from 'svelte/store';

export type ToastType = 'success' | 'error' | 'info' | 'warning';

export interface Toast {
	id: string;
	type: ToastType;
	message: string;
	duration?: number;
}

const MAX_TOASTS = 5;

function createToastStore() {
	const { subscribe, update } = writable<Toast[]>([]);

	function remove(id: string): void {
		update((toasts) => toasts.filter((t) => t.id !== id));
	}

	function show(type: ToastType, message: string, duration = 5000): void {
		const id = crypto.randomUUID();
		const toast: Toast = { id, type, message, duration };

		update((toasts) => {
			const next = [...toasts, toast];
			// Drop oldest toasts if over limit
			return next.length > MAX_TOASTS ? next.slice(-MAX_TOASTS) : next;
		});

		if (duration > 0) {
			setTimeout(() => remove(id), duration);
		}
	}

	return {
		subscribe,
		show,
		remove,
		success: (message: string, duration?: number) =>
			show('success', message, duration),
		error: (message: string, duration?: number) =>
			show('error', message, duration),
		info: (message: string, duration?: number) =>
			show('info', message, duration),
		warning: (message: string, duration?: number) =>
			show('warning', message, duration)
	};
}

export const toastStore = createToastStore();
