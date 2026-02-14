import { api } from './client';
import type { UserDTO } from '$lib/types/api';

export const sharedUsersApi = {
	/**
	 * Search for users that the current user has shared resources with
	 */
	async search(query: string = ''): Promise<{ users: UserDTO[] }> {
		const params = query ? `?q=${encodeURIComponent(query)}` : '';
		return api.get<{ users: UserDTO[] }>(`/shared-users${params}`);
	}
};
