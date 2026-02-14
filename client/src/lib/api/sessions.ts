import { api } from './client';

export interface SessionDTO {
	id: string;
	ip_address: string;
	user_agent: string;
	device_info: string;
	browser_info: string;
	is_current: boolean;
	created_at: string;
	last_active_at: string;
}

export const sessionsApi = {
	async list(): Promise<{ sessions: SessionDTO[] }> {
		return api.get('/profile/sessions');
	},

	async revoke(id: string): Promise<{ message: string }> {
		return api.delete(`/profile/sessions/${id}`);
	},

	async revokeOthers(): Promise<{ message: string; revoked_count: number }> {
		return api.post('/profile/sessions/revoke-others');
	}
};
