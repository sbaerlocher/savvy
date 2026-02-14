import { api, getCSRFToken } from './client';
import type { ImportResult, ImportPreview } from '$lib/types/api';

export const importApi = {
	previewJSON: async (data: unknown): Promise<ImportPreview> => {
		return api.post<ImportPreview>('/import/json/preview', data);
	},

	importJSON: async (data: unknown): Promise<ImportResult> => {
		return api.post<ImportResult>('/import/json', data);
	},

	importCardsCSV: async (file: File): Promise<ImportResult> => {
		return uploadCSV('/import/csv/cards', file);
	},

	importVouchersCSV: async (file: File): Promise<ImportResult> => {
		return uploadCSV('/import/csv/vouchers', file);
	},

	importGiftCardsCSV: async (file: File): Promise<ImportResult> => {
		return uploadCSV('/import/csv/gift-cards', file);
	}
};

async function uploadCSV(path: string, file: File): Promise<ImportResult> {
	const formData = new FormData();
	formData.append('file', file);

	const headers: Record<string, string> = {};
	const csrfToken = getCSRFToken();
	if (csrfToken) {
		headers['X-CSRF-Token'] = csrfToken;
	}

	const response = await fetch(`/api/v1${path}`, {
		method: 'POST',
		credentials: 'include',
		headers,
		body: formData
	});

	if (!response.ok) {
		const error = await response
			.json()
			.catch(() => ({ message: 'Import failed' }));
		throw new Error(error.message || 'Import failed');
	}

	return response.json();
}
