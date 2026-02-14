import { api } from './client';
import type { DashboardResponse } from '$lib/types/api';
import { offlineDB } from '$lib/stores/offline-db';
import { browser } from '$app/environment';
import { logger } from '$lib/utils/logger';

const log = logger.child('DashboardAPI');

export const dashboardApi = {
	/**
	 * Get dashboard data from IndexedDB cache only (no network)
	 */
	async getCached(): Promise<DashboardResponse | null> {
		if (!browser) return null;
		return offlineDB.getDashboard();
	},

	/**
	 * Get dashboard data — Network-first with IndexedDB fallback
	 */
	async get(): Promise<DashboardResponse> {
		if (browser && !navigator.onLine) {
			const cached = await offlineDB.getDashboard();
			if (cached) return cached;
			throw new Error('Dashboard not available offline');
		}

		try {
			const result = await api.get<DashboardResponse>('/dashboard');

			if (browser) {
				await offlineDB.saveDashboard(result);
			}

			return result;
		} catch (error) {
			if (browser) {
				const cached = await offlineDB.getDashboard();
				if (cached) {
					log.warn(
						'[dashboard.get] Network failed, returning cached data:',
						error
					);
					return cached;
				}
			}
			throw error;
		}
	}
};
