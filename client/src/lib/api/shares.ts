/**
 * Shared API utilities for resource sharing operations
 */
import { api } from './client';
import type {
	ShareDTO,
	ShareCreateRequest,
	ShareCreateResponse,
	TransferRequest
} from '$lib/types/api';

/**
 * Generic share API operations for any resource type
 */
export const createShareApi = (resourcePath: string) => ({
	/**
	 * List shares for resource
	 */
	async listShares(id: string): Promise<{ shares: ShareDTO[] }> {
		return api.get<{ shares: ShareDTO[] }>(`/${resourcePath}/${id}/shares`);
	},

	/**
	 * Create share
	 */
	async createShare(
		id: string,
		data: ShareCreateRequest
	): Promise<ShareCreateResponse> {
		return api.post<ShareCreateResponse>(`/${resourcePath}/${id}/share`, data);
	},

	/**
	 * Delete share
	 */
	async deleteShare(
		id: string,
		sharedWithID: string
	): Promise<{ message: string }> {
		return api.delete<{ message: string }>(
			`/${resourcePath}/${id}/share/${sharedWithID}`
		);
	},

	/**
	 * Transfer ownership
	 */
	async transfer(
		id: string,
		data: TransferRequest
	): Promise<{ message: string }> {
		return api.post<{ message: string }>(
			`/${resourcePath}/${id}/transfer`,
			data
		);
	}
});

/**
 * Share API with update permissions support (Cards, Gift Cards)
 */
export const createShareWithPermissionsApi = (resourcePath: string) => {
	const base = createShareApi(resourcePath);

	return {
		...base,

		/**
		 * Update share permissions
		 */
		async updateShare(
			id: string,
			sharedWithID: string,
			data: {
				can_edit?: boolean;
				can_delete?: boolean;
				can_edit_transactions?: boolean;
			}
		): Promise<{ message: string; shares: ShareDTO[] }> {
			return api.patch<{ message: string; shares: ShareDTO[] }>(
				`/${resourcePath}/${id}/share/${sharedWithID}`,
				data
			);
		}
	};
};
