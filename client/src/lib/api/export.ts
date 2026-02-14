import { getCSRFToken, ApiError } from './client';

export const exportApi = {
	async download(): Promise<{ blob: Blob; filename: string }> {
		const response = await fetch('/api/v1/export', {
			credentials: 'include'
		});

		if (!response.ok) {
			let error = 'export_failed';
			let message = 'Export failed';
			try {
				const data = await response.json();
				error = data.error || error;
				message = data.message || message;
			} catch {
				// non-JSON error response
			}
			throw new ApiError(response.status, error, message);
		}

		const blob = await response.blob();
		const filename =
			response.headers
				.get('Content-Disposition')
				?.match(/filename="?([^"]+)"?/)?.[1] || 'savvy-export.json';

		return { blob, filename };
	}
};
