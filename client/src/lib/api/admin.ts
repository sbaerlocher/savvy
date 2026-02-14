import { api } from './client';
import type {
	AdminUserDTO,
	AdminUserCreateRequest,
	AdminUserUpdateRequest,
	AuditLogDTO,
	AuditLogFiltersRequest
} from '$lib/types/api';

export const adminApi = {
	// User Management

	/**
	 * List all users
	 */
	async listUsers(): Promise<{ users: AdminUserDTO[] }> {
		return api.get<{ users: AdminUserDTO[] }>('/admin/users');
	},

	/**
	 * Get single user by ID
	 */
	async getUser(id: string): Promise<{ user: AdminUserDTO }> {
		return api.get<{ user: AdminUserDTO }>(`/admin/users/${id}`);
	},

	/**
	 * Create new user
	 */
	async createUser(
		data: AdminUserCreateRequest
	): Promise<{ message: string; user: AdminUserDTO }> {
		return api.post<{ message: string; user: AdminUserDTO }>(
			'/admin/users',
			data
		);
	},

	/**
	 * Update user (email, name, role)
	 */
	async updateUser(
		id: string,
		data: AdminUserUpdateRequest
	): Promise<{ message: string }> {
		return api.patch<{ message: string }>(`/admin/users/${id}`, data);
	},

	// Impersonation

	/**
	 * Start impersonating a user
	 */
	async startImpersonation(
		userId: string
	): Promise<{ message: string; user: AdminUserDTO }> {
		return api.post<{ message: string; user: AdminUserDTO }>(
			`/admin/users/${userId}/impersonate`,
			{}
		);
	},

	/**
	 * Stop impersonating and return to admin account
	 */
	async stopImpersonation(): Promise<{ message: string }> {
		return api.post<{ message: string }>('/admin/impersonate/stop', {});
	},

	// Audit Log

	/**
	 * Get audit logs with filters and pagination
	 */
	async getAuditLogs(filters: AuditLogFiltersRequest): Promise<{
		logs: AuditLogDTO[];
		total: number;
		page: number;
		per_page: number;
		total_pages: number;
	}> {
		const queryParams = new URLSearchParams();

		if (filters.user_id) queryParams.set('user_id', filters.user_id);
		if (filters.resource_type)
			queryParams.set('resource_type', filters.resource_type);
		if (filters.action) queryParams.set('action', filters.action);
		if (filters.date_from) queryParams.set('date_from', filters.date_from);
		if (filters.date_to) queryParams.set('date_to', filters.date_to);
		if (filters.search) queryParams.set('search', filters.search);
		if (filters.page) queryParams.set('page', filters.page.toString());
		if (filters.per_page)
			queryParams.set('per_page', filters.per_page.toString());

		const query = queryParams.toString() ? `?${queryParams.toString()}` : '';

		return api.get(`/admin/audit-log${query}`);
	},

	/**
	 * Restore a soft-deleted resource
	 */
	async restoreResource(
		resourceType: string,
		resourceId: string
	): Promise<{ message: string }> {
		return api.post<{ message: string }>(
			`/admin/restore/${resourceType}/${resourceId}`,
			{}
		);
	},

	// System Health

	/**
	 * Send test email to verify SMTP configuration
	 */
	async sendTestEmail(): Promise<{ message: string }> {
		return api.post<{ message: string }>('/admin/test-email', {});
	},

	/**
	 * Send test push notification to verify VAPID configuration
	 */
	async sendTestPush(): Promise<{ message: string }> {
		return api.post<{ message: string }>('/admin/test-push', {});
	},

	/**
	 * Send a preview email of a specific template type (dev mode only)
	 */
	async sendPreviewEmail(
		template: string,
		language?: string
	): Promise<{ message: string }> {
		return api.post<{ message: string }>('/admin/preview-email', {
			template,
			language
		});
	}
};
