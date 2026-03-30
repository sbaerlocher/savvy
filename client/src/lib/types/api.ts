// Batch Operations
export interface BatchDeleteRequest {
	ids: string[];
}

export interface BatchShareRequest {
	ids: string[];
	email: string;
	can_edit?: boolean;
	can_delete?: boolean;
	can_edit_transactions?: boolean;
}

export interface BatchTransferRequest {
	ids: string[];
	new_owner_email: string;
}

export interface BatchResponse {
	success_count: number;
	failed: BatchFailedItem[];
}

export interface BatchFailedItem {
	id: string;
	error: string;
}

// Pagination
export interface PaginationMeta {
	total: number;
	page: number;
	per_page: number;
	total_pages: number;
}

// Duplicate Warning
export interface DuplicateWarning {
	has_duplicate: boolean;
	merchant_name?: string;
	resource_number?: string;
	existing_id?: string;
}

// API Response Types
export type ErrorCode =
	| 'unauthorized'
	| 'forbidden'
	| 'not_found'
	| 'validation_error'
	| 'server_error'
	| 'conflict'
	| 'bad_request';

export interface ErrorResponse {
	error: ErrorCode;
	message: string;
	details?: Record<string, string[]>; // Validation errors (field -> error messages)
}

export interface PermissionDTO {
	can_view: boolean;
	can_edit: boolean;
	can_delete: boolean;
	can_edit_transactions?: boolean;
	is_owner: boolean;
}

export interface UserDTO {
	id: string;
	email: string;
	first_name?: string;
	last_name?: string;
	is_admin: boolean;
	is_impersonating?: boolean;
	email_verified?: boolean;
	auth_provider?: string;
	language?: string;
}

export interface MerchantDTO {
	id: string;
	name: string;
	description?: string;
	logo_url?: string;
	website?: string;
	color?: string;
	created_at: string;
	updated_at: string;
}

export interface CardDTO {
	id: string;
	user_id?: string;
	merchant_id?: string;
	merchant?: MerchantDTO;
	owner?: UserDTO;
	card_number: string;
	program?: string;
	barcode_type?: string;
	status?: string;
	notes?: string;
	is_favorite: boolean;
	shared_with_count: number;
	permissions?: PermissionDTO;
	duplicate_warning?: DuplicateWarning;
	created_at: string;
	updated_at: string;
}

export interface VoucherDTO {
	id: string;
	user_id?: string;
	merchant_id?: string;
	merchant?: MerchantDTO;
	owner?: UserDTO;
	code: string;
	type: string;
	value: number;
	currency: string;
	description?: string;
	barcode_type?: string;
	valid_from?: string;
	valid_until?: string;
	min_purchase_amount?: number;
	usage_limit_type?: string;
	status?: string;
	is_favorite: boolean;
	shared_with_count: number;
	permissions?: PermissionDTO;
	duplicate_warning?: DuplicateWarning;
	created_at: string;
	updated_at: string;
}

export interface GiftCardDTO {
	id: string;
	user_id?: string;
	merchant_id?: string;
	merchant?: MerchantDTO;
	owner?: UserDTO;
	card_number: string;
	initial_balance: number;
	current_balance: number;
	currency: string;
	pin?: string;
	barcode_type?: string;
	expires_at?: string;
	status?: string;
	notes?: string;
	is_favorite: boolean;
	shared_with_count: number;
	permissions?: PermissionDTO;
	duplicate_warning?: DuplicateWarning;
	transactions?: TransactionDTO[]; // Cached for offline use (from detail endpoint)
	created_at: string;
	updated_at: string;
}

export interface TransactionDTO {
	id: string;
	gift_card_id: string;
	type: string;
	amount: number;
	description?: string;
	transaction_date: string;
	created_at: string;
}

export interface ShareDTO {
	shared_with_user: UserDTO;
	can_edit: boolean;
	can_delete: boolean;
	can_edit_transactions?: boolean;
	created_at: string;
}

export interface DashboardStats {
	cards_count: number;
	vouchers_count: number;
	gift_cards_count: number;
	shared_count: number;
	total_balance: number;
	favorite_counts: { [key: string]: number };
}

export interface DashboardResponse {
	stats: DashboardStats;
	recent_cards: CardDTO[];
	recent_vouchers: VoucherDTO[];
	recent_gift_cards: GiftCardDTO[];
	has_favorites: boolean;
	has_card_favorites: boolean;
	has_voucher_favorites: boolean;
	has_gift_card_favorites: boolean;
}

// Request Types
export interface LoginRequest {
	email: string;
	password: string;
}

export interface RegisterRequest {
	email: string;
	password: string;
	first_name?: string;
	last_name?: string;
}

export interface CardCreateRequest {
	merchant_id?: string;
	new_merchant_name?: string;
	program?: string;
	card_number: string;
	barcode_type?: string;
	notes?: string;
	status?: string;
	// Optional sharing on creation
	share_with_email?: string;
	share_can_edit?: boolean;
	share_can_delete?: boolean;
}

export interface CardUpdateRequest {
	merchant_id?: string;
	card_number?: string;
	program?: string;
	barcode_type?: string;
	status?: string;
	notes?: string;
}

export interface VoucherCreateRequest {
	merchant_id?: string;
	code: string;
	type: string;
	value: number;
	currency?: string;
	description?: string;
	barcode_type?: string;
	min_purchase_amount?: number;
	usage_limit_type?: string;
	valid_from?: string;
	valid_until?: string;
	status?: string;
	// Optional sharing on creation
	share_with_email?: string;
}

export interface VoucherUpdateRequest {
	merchant_id?: string;
	code?: string;
	type?: string;
	value?: number;
	currency?: string;
	description?: string;
	barcode_type?: string;
	min_purchase_amount?: number;
	usage_limit_type?: string;
	valid_from?: string;
	valid_until?: string;
	status?: string;
}

export interface GiftCardCreateRequest {
	merchant_id?: string;
	card_number: string;
	initial_balance: number;
	currency: string;
	pin?: string;
	barcode_type?: string;
	expires_at?: string;
	notes?: string;
	status?: string;
	// Optional sharing on creation
	share_with_email?: string;
	share_can_edit?: boolean;
	share_can_delete?: boolean;
	share_can_edit_transactions?: boolean;
}

export interface GiftCardUpdateRequest {
	merchant_id?: string;
	card_number?: string;
	initial_balance?: number;
	currency?: string;
	pin?: string;
	barcode_type?: string;
	expires_at?: string;
	status?: string;
	notes?: string;
}

export interface TransactionCreateRequest {
	type: 'debit' | 'credit';
	amount: number;
	description?: string;
	transaction_date?: string;
}

export interface ShareCreateRequest {
	email: string;
	can_edit?: boolean;
	can_delete?: boolean;
	can_edit_transactions?: boolean;
}

export interface TransferRequest {
	new_owner_email: string;
}

// Admin Types
export interface AdminUserDTO {
	id: string;
	email: string;
	first_name: string;
	last_name: string;
	role: 'user' | 'admin';
	auth_provider: 'local' | 'oauth';
	created_at: string;
	updated_at: string;
}

export interface AdminUserCreateRequest {
	email: string;
	password: string;
	first_name: string;
	last_name: string;
	role?: 'user' | 'admin';
}

export interface AdminUserUpdateRequest {
	email?: string;
	first_name?: string;
	last_name?: string;
	role?: 'user' | 'admin';
}

export interface AuditLogDTO {
	id: string;
	user_id?: string;
	user?: UserDTO;
	action: 'delete' | 'hard_delete' | 'restore';
	resource_type: string;
	resource_id: string;
	resource_data: string; // JSON string
	ip_address: string;
	user_agent: string;
	created_at: string;
}

export interface AuditLogFiltersRequest {
	user_id?: string;
	resource_type?: string;
	action?: string;
	date_from?: string;
	date_to?: string;
	search?: string;
	page?: number;
	per_page?: number;
}

// API Response Types
export interface CardResponse {
	card: CardDTO;
	shares?: ShareDTO[];
}

export interface VoucherResponse {
	voucher: VoucherDTO;
	shares?: ShareDTO[];
}

export interface GiftCardResponse {
	gift_card: GiftCardDTO;
	permissions: PermissionDTO;
	transactions?: TransactionDTO[];
	shares: ShareDTO[];
}

export interface TransactionListResponse {
	transactions: TransactionDTO[];
}

export interface ShareListResponse {
	shares: ShareDTO[];
}

export interface UserSearchResponse {
	users: UserDTO[];
}

// Import Types
export interface ImportResult {
	cards_imported: number;
	vouchers_imported: number;
	gift_cards_imported: number;
	skipped: number;
	errors?: ImportError[];
}

export interface ImportError {
	row?: number;
	field?: string;
	message: string;
}

export interface ImportPreview {
	cards: number;
	vouchers: number;
	gift_cards: number;
}

// Notification Types
export type NotificationType =
	| 'share_received'
	| 'transfer_received'
	| 'expiry_reminder';

export interface NotificationDTO {
	id: string;
	type: NotificationType;
	resource_type: 'card' | 'voucher' | 'gift_card';
	resource_id: string;
	metadata: {
		from_user_id?: string;
		from_user_name?: string;
		permissions?: {
			can_edit?: boolean;
			can_delete?: boolean;
			can_edit_transactions?: boolean;
		};
		[key: string]: any;
	};
	is_read: boolean;
	read_at?: string;
	created_at: string;
}

export interface NotificationListResponse {
	notifications: NotificationDTO[];
}

export interface NotificationUnreadCountResponse {
	count: number;
}

// Search
export interface SearchResultsResponse {
	cards: CardDTO[];
	vouchers: VoucherDTO[];
	gift_cards: GiftCardDTO[];
	total: number;
}
