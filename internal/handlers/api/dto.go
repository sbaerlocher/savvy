// Package api contains JSON API handlers and DTOs for the SvelteKit frontend.
package api

import "time"

// ==================== Pagination ====================

// PaginationMeta represents pagination metadata in list responses
type PaginationMeta struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	TotalPages int   `json:"total_pages"`
}

// ==================== Error Responses ====================

// ErrorResponse represents a generic API error
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// DuplicateErrorResponse represents an error response when a duplicate resource is detected
type DuplicateErrorResponse struct {
	Error     string            `json:"error"`
	Message   string            `json:"message,omitempty"`
	Duplicate *DuplicateWarning `json:"duplicate"`
}

// ==================== Permission DTOs ====================

// PermissionDTO represents resource access permissions
type PermissionDTO struct {
	CanView             bool `json:"can_view"`
	CanEdit             bool `json:"can_edit"`
	CanDelete           bool `json:"can_delete"`
	CanEditTransactions bool `json:"can_edit_transactions,omitempty"` // Gift Cards only
	IsOwner             bool `json:"is_owner"`
}

// DuplicateWarning indicates a potential duplicate resource
type DuplicateWarning struct {
	HasDuplicate   bool   `json:"has_duplicate"`
	MerchantName   string `json:"merchant_name,omitempty"`
	ResourceNumber string `json:"resource_number,omitempty"` // Card number, voucher code, etc.
	ExistingID     string `json:"existing_id,omitempty"`
	Deleted        bool   `json:"deleted"` // true = existing_id refers to a soft-deleted resource that can be restored
}

// ==================== User DTOs ====================

// UserDTO represents a user in API responses
type UserDTO struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	IsAdmin   bool   `json:"is_admin"`
}

// LoginRequest represents login credentials
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"` // #nosec G117 -- struct field name, not a hardcoded secret
}

// RegisterRequest represents registration data
type RegisterRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"` // #nosec G117 -- struct field name, not a hardcoded secret
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// ==================== Merchant DTOs ====================

// MerchantDTO represents a merchant/brand
type MerchantDTO struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Color     *string `json:"color,omitempty"`
	LogoURL   *string `json:"logo_url,omitempty"`
	Website   *string `json:"website,omitempty"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

// MerchantCreateRequest represents merchant creation data (Admin only)
type MerchantCreateRequest struct {
	Name    string  `json:"name"`
	Color   *string `json:"color,omitempty"`
	LogoURL *string `json:"logo_url,omitempty"`
	Website *string `json:"website,omitempty"`
}

// MerchantUpdateRequest represents merchant update data (Admin only)
type MerchantUpdateRequest struct {
	Name    *string `json:"name,omitempty"`
	Color   *string `json:"color,omitempty"`
	LogoURL *string `json:"logo_url,omitempty"`
	Website *string `json:"website,omitempty"`
}

// ==================== Card DTOs ====================

// CardDTO represents a customer card
type CardDTO struct {
	ID              string         `json:"id"`
	MerchantID      *string        `json:"merchant_id,omitempty"`
	Merchant        *MerchantDTO   `json:"merchant,omitempty"`
	Owner           *UserDTO       `json:"owner,omitempty"`
	Program         *string        `json:"program,omitempty"`
	CardNumber      string         `json:"card_number"`
	BarcodeType     *string        `json:"barcode_type,omitempty"`
	Notes           *string        `json:"notes,omitempty"`
	Status          string         `json:"status"`
	IsFavorite      bool           `json:"is_favorite"`
	IsShared        bool           `json:"is_shared"`
	SharedWithCount int            `json:"shared_with_count"`
	CreatedAt       string         `json:"created_at"`
	UpdatedAt       string         `json:"updated_at"`
	Permissions     *PermissionDTO `json:"permissions,omitempty"` // Included in detail views
}

// CardCreateRequest represents card creation data
type CardCreateRequest struct {
	MerchantID      *string `json:"merchant_id,omitempty"`
	NewMerchantName *string `json:"new_merchant_name,omitempty"`
	Program         *string `json:"program,omitempty"`
	CardNumber      string  `json:"card_number"`
	BarcodeType     *string `json:"barcode_type,omitempty"`
	Notes           *string `json:"notes,omitempty"`
	Status          *string `json:"status,omitempty"`
	// Optional sharing on creation
	ShareWithEmail *string `json:"share_with_email,omitempty"`
	ShareCanEdit   *bool   `json:"share_can_edit,omitempty"`
	ShareCanDelete *bool   `json:"share_can_delete,omitempty"`
}

// CardUpdateRequest represents card update data
type CardUpdateRequest struct {
	MerchantID  *string `json:"merchant_id,omitempty"`
	Program     *string `json:"program,omitempty"`
	CardNumber  *string `json:"card_number,omitempty"`
	BarcodeType *string `json:"barcode_type,omitempty"`
	Notes       *string `json:"notes,omitempty"`
	Status      *string `json:"status,omitempty"`
}

// CardListResponse represents a list of cards
type CardListResponse struct {
	Cards      []CardDTO       `json:"cards"`
	Pagination *PaginationMeta `json:"pagination,omitempty"`
}

// CardDetailResponse represents a single card with permissions
type CardDetailResponse struct {
	Card             CardDTO           `json:"card"`
	Permissions      PermissionDTO     `json:"permissions"`
	Shares           []ShareDTO        `json:"shares,omitempty"`            // If owner
	DuplicateWarning *DuplicateWarning `json:"duplicate_warning,omitempty"` // Duplicate detection
}

// ==================== Voucher DTOs ====================

// VoucherDTO represents a voucher
type VoucherDTO struct {
	ID                string         `json:"id"`
	MerchantID        *string        `json:"merchant_id,omitempty"`
	Merchant          *MerchantDTO   `json:"merchant,omitempty"`
	Owner             *UserDTO       `json:"owner,omitempty"`
	Code              string         `json:"code"`
	Type              string         `json:"type"` // percentage, fixed_amount, points_multiplier
	Value             float64        `json:"value"`
	Currency          string         `json:"currency"` // Currency for fixed_amount vouchers (CHF, EUR, USD, GBP)
	Description       *string        `json:"description,omitempty"`
	MinPurchaseAmount float64        `json:"min_purchase_amount"`
	ValidFrom         string         `json:"valid_from"`
	ValidUntil        string         `json:"valid_until"`
	UsageLimitType    string         `json:"usage_limit_type"` // single_use, one_per_customer, etc.
	BarcodeType       string         `json:"barcode_type"`
	Status            string         `json:"status"` // valid, expired (computed from valid_from/valid_until)
	IsFavorite        bool           `json:"is_favorite"`
	IsShared          bool           `json:"is_shared"`
	SharedWithCount   int            `json:"shared_with_count"`
	CreatedAt         string         `json:"created_at"`
	UpdatedAt         string         `json:"updated_at"`
	Permissions       *PermissionDTO `json:"permissions,omitempty"`
}

// VoucherCreateRequest represents voucher creation data
type VoucherCreateRequest struct {
	MerchantID        *string  `json:"merchant_id,omitempty"`
	NewMerchantName   *string  `json:"new_merchant_name,omitempty"`
	Code              string   `json:"code"`
	Type              string   `json:"type"` // percentage, fixed_amount, points_multiplier
	Value             float64  `json:"value"`
	Currency          *string  `json:"currency,omitempty"` // Currency for fixed_amount vouchers (CHF, EUR, USD, GBP)
	Description       *string  `json:"description,omitempty"`
	MinPurchaseAmount *float64 `json:"min_purchase_amount,omitempty"`
	ValidFrom         string   `json:"valid_from"`
	ValidUntil        string   `json:"valid_until"`
	UsageLimitType    *string  `json:"usage_limit_type,omitempty"`
	BarcodeType       *string  `json:"barcode_type,omitempty"`
	// Optional sharing on creation
	ShareWithEmail *string `json:"share_with_email,omitempty"`
}

// VoucherUpdateRequest represents voucher update data
type VoucherUpdateRequest struct {
	MerchantID        *string  `json:"merchant_id,omitempty"`
	Code              *string  `json:"code,omitempty"`
	Type              *string  `json:"type,omitempty"`
	Value             *float64 `json:"value,omitempty"`
	Currency          *string  `json:"currency,omitempty"` // Currency for fixed_amount vouchers (CHF, EUR, USD, GBP)
	Description       *string  `json:"description,omitempty"`
	MinPurchaseAmount *float64 `json:"min_purchase_amount,omitempty"`
	ValidFrom         *string  `json:"valid_from,omitempty"`
	ValidUntil        *string  `json:"valid_until,omitempty"`
	UsageLimitType    *string  `json:"usage_limit_type,omitempty"`
	BarcodeType       *string  `json:"barcode_type,omitempty"`
}

// VoucherListResponse represents a list of vouchers
type VoucherListResponse struct {
	Vouchers   []VoucherDTO    `json:"vouchers"`
	Pagination *PaginationMeta `json:"pagination,omitempty"`
}

// VoucherDetailResponse represents a single voucher with permissions
type VoucherDetailResponse struct {
	Voucher          VoucherDTO        `json:"voucher"`
	Permissions      PermissionDTO     `json:"permissions"`
	Shares           []ShareDTO        `json:"shares,omitempty"`
	DuplicateWarning *DuplicateWarning `json:"duplicate_warning,omitempty"` // Duplicate detection
}

// ==================== Gift Card DTOs ====================

// GiftCardDTO represents a gift card
type GiftCardDTO struct {
	ID              string         `json:"id"`
	MerchantID      *string        `json:"merchant_id,omitempty"`
	Merchant        *MerchantDTO   `json:"merchant,omitempty"`
	Owner           *UserDTO       `json:"owner,omitempty"`
	CardNumber      string         `json:"card_number"`
	InitialBalance  float64        `json:"initial_balance"`
	CurrentBalance  float64        `json:"current_balance"`
	Currency        string         `json:"currency"`
	PIN             *string        `json:"pin,omitempty"`
	BarcodeType     *string        `json:"barcode_type,omitempty"`
	ExpiresAt       *string        `json:"expires_at,omitempty"`
	Notes           *string        `json:"notes,omitempty"`
	IsFavorite      bool           `json:"is_favorite"`
	IsShared        bool           `json:"is_shared"`
	SharedWithCount int            `json:"shared_with_count"`
	CreatedAt       string         `json:"created_at"`
	UpdatedAt       string         `json:"updated_at"`
	Permissions     *PermissionDTO `json:"permissions,omitempty"`
}

// GiftCardCreateRequest represents gift card creation data
type GiftCardCreateRequest struct {
	MerchantID      *string `json:"merchant_id,omitempty"`
	NewMerchantName *string `json:"new_merchant_name,omitempty"`
	CardNumber      string  `json:"card_number"`
	InitialBalance  float64 `json:"initial_balance"`
	Currency        string  `json:"currency"`
	PIN             *string `json:"pin,omitempty"`
	BarcodeType     *string `json:"barcode_type,omitempty"`
	ExpiresAt       *string `json:"expires_at,omitempty"`
	Notes           *string `json:"notes,omitempty"`
}

// GiftCardUpdateRequest represents gift card update data
type GiftCardUpdateRequest struct {
	MerchantID     *string  `json:"merchant_id,omitempty"`
	CardNumber     *string  `json:"card_number,omitempty"`
	InitialBalance *float64 `json:"initial_balance,omitempty"`
	Currency       *string  `json:"currency,omitempty"`
	PIN            *string  `json:"pin,omitempty"`
	BarcodeType    *string  `json:"barcode_type,omitempty"`
	ExpiresAt      *string  `json:"expires_at,omitempty"`
	Notes          *string  `json:"notes,omitempty"`
}

// GiftCardListResponse represents a list of gift cards
type GiftCardListResponse struct {
	GiftCards  []GiftCardDTO   `json:"gift_cards"`
	Pagination *PaginationMeta `json:"pagination,omitempty"`
}

// GiftCardDetailResponse represents a single gift card with permissions and transactions
type GiftCardDetailResponse struct {
	GiftCard         GiftCardDTO              `json:"gift_card"`
	Permissions      PermissionDTO            `json:"permissions"`
	Transactions     []GiftCardTransactionDTO `json:"transactions"`
	Shares           []ShareDTO               `json:"shares,omitempty"`
	DuplicateWarning *DuplicateWarning        `json:"duplicate_warning,omitempty"` // Duplicate detection
}

// ==================== Transaction DTOs ====================

// GiftCardTransactionDTO represents a gift card transaction
type GiftCardTransactionDTO struct {
	ID              string  `json:"id"`
	GiftCardID      string  `json:"gift_card_id"`
	Amount          float64 `json:"amount"`
	Description     *string `json:"description,omitempty"`
	TransactionDate string  `json:"transaction_date"`
	CreatedAt       string  `json:"created_at"`
}

// TransactionCreateRequest represents transaction creation data
type TransactionCreateRequest struct {
	Amount          float64 `json:"amount"`
	Description     *string `json:"description,omitempty"`
	TransactionDate *string `json:"transaction_date,omitempty"` // ISO 8601, defaults to now
}

// TransactionUpdateRequest represents transaction update data
type TransactionUpdateRequest struct {
	Amount          *float64 `json:"amount,omitempty"`
	Description     *string  `json:"description,omitempty"`
	TransactionDate *string  `json:"transaction_date,omitempty"`
}

// ==================== Share DTOs ====================

// ShareDTO represents a share record
type ShareDTO struct {
	ID                  string  `json:"id"`
	SharedWithUser      UserDTO `json:"shared_with_user"`
	CanEdit             bool    `json:"can_edit"`
	CanDelete           bool    `json:"can_delete"`
	CanEditTransactions bool    `json:"can_edit_transactions,omitempty"` // Gift Cards only
	CreatedAt           string  `json:"created_at"`
}

// ShareCreateRequest represents share creation data.
// One or more recipients are given in Emails; shares are created for each with
// the same permissions and a partial-success response is returned.
type ShareCreateRequest struct {
	Emails              []string `json:"emails"`
	CanEdit             *bool    `json:"can_edit,omitempty"`
	CanDelete           *bool    `json:"can_delete,omitempty"`
	CanEditTransactions *bool    `json:"can_edit_transactions,omitempty"`
}

// ShareCreateResponse is the partial-success result of a multi-recipient share.
type ShareCreateResponse struct {
	SuccessCount int               `json:"success_count"`
	Failed       []BatchFailedItem `json:"failed"`
	Shares       []ShareDTO        `json:"shares"`
}

// ShareUpdateRequest represents share permission updates
type ShareUpdateRequest struct {
	CanEdit             *bool `json:"can_edit,omitempty"`
	CanDelete           *bool `json:"can_delete,omitempty"`
	CanEditTransactions *bool `json:"can_edit_transactions,omitempty"`
}

// ==================== Transfer DTOs ====================

// TransferRequest represents ownership transfer data
type TransferRequest struct {
	NewOwnerEmail string `json:"new_owner_email"`
}

// ==================== Dashboard DTOs ====================

// DashboardStats represents aggregated stats
type DashboardStats struct {
	CardsCount     int            `json:"cards_count"`
	VouchersCount  int            `json:"vouchers_count"`
	GiftCardsCount int            `json:"gift_cards_count"`
	SharedCount    int            `json:"shared_count"`
	TotalBalance   float64        `json:"total_balance"`
	FavoriteCounts map[string]int `json:"favorite_counts"` // {"card": 3, "voucher": 2, "gift_card": 1}
}

// DashboardResponse represents the dashboard data
type DashboardResponse struct {
	Stats                DashboardStats `json:"stats"`
	RecentCards          []CardDTO      `json:"recent_cards"`
	RecentVouchers       []VoucherDTO   `json:"recent_vouchers"`
	RecentGiftCards      []GiftCardDTO  `json:"recent_gift_cards"`
	HasFavorites         bool           `json:"has_favorites"`
	HasCardFavorites     bool           `json:"has_card_favorites"`
	HasVoucherFavorites  bool           `json:"has_voucher_favorites"`
	HasGiftCardFavorites bool           `json:"has_gift_card_favorites"`
}

// ==================== Admin DTOs ====================

// AdminUserDTO represents a user in admin panel (includes role and auth provider)
type AdminUserDTO struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Role         string `json:"role"`          // "user" or "admin"
	AuthProvider string `json:"auth_provider"` // "local" or "oauth"
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// AdminUserListResponse represents list of users
type AdminUserListResponse struct {
	Users []AdminUserDTO `json:"users"`
}

// AdminUserCreateRequest represents user creation request
type AdminUserCreateRequest struct {
	Email     string  `json:"email"`
	Password  string  `json:"password"` // #nosec G117 -- password input field, not a secret
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	Role      *string `json:"role,omitempty"` // Optional, defaults to "user"
}

// AdminUserUpdateRequest represents user update request
type AdminUserUpdateRequest struct {
	Email     *string `json:"email,omitempty"`
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	Role      *string `json:"role,omitempty"`
}

// AuditLogDTO represents an audit log entry
type AuditLogDTO struct {
	ID           string   `json:"id"`
	UserID       *string  `json:"user_id,omitempty"`
	User         *UserDTO `json:"user,omitempty"`
	Action       string   `json:"action"` // "delete", "hard_delete", "restore"
	ResourceType string   `json:"resource_type"`
	ResourceID   string   `json:"resource_id"`
	ResourceData string   `json:"resource_data"` // JSONB snapshot
	IPAddress    string   `json:"ip_address"`
	UserAgent    string   `json:"user_agent"`
	CreatedAt    string   `json:"created_at"`
}

// AuditLogListResponse represents paginated audit log
type AuditLogListResponse struct {
	Logs       []AuditLogDTO `json:"logs"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	PerPage    int           `json:"per_page"`
	TotalPages int           `json:"total_pages"`
}

// AuditLogFiltersRequest represents audit log filter parameters
type AuditLogFiltersRequest struct {
	UserID       *string `query:"user_id"`
	ResourceType string  `query:"resource_type"`
	Action       string  `query:"action"`
	DateFrom     string  `query:"date_from"` // ISO 8601
	DateTo       string  `query:"date_to"`   // ISO 8601
	SearchQuery  string  `query:"search"`
	Page         int     `query:"page"`
	PerPage      int     `query:"per_page"`
}

// ==================== Batch DTOs ====================

// BatchDeleteRequest represents a batch delete request
type BatchDeleteRequest struct {
	IDs []string `json:"ids"`
}

// BatchShareRequest represents a batch share request
type BatchShareRequest struct {
	IDs                 []string `json:"ids"`
	Email               string   `json:"email"`
	CanEdit             *bool    `json:"can_edit,omitempty"`
	CanDelete           *bool    `json:"can_delete,omitempty"`
	CanEditTransactions *bool    `json:"can_edit_transactions,omitempty"`
}

// BatchTransferRequest represents a batch transfer request
type BatchTransferRequest struct {
	IDs           []string `json:"ids"`
	NewOwnerEmail string   `json:"new_owner_email"`
}

// BatchResponse represents the result of a batch operation
type BatchResponse struct {
	SuccessCount int               `json:"success_count"`
	Failed       []BatchFailedItem `json:"failed"`
}

// BatchFailedItem represents a single failed item in a batch operation
type BatchFailedItem struct {
	ID    string `json:"id"`
	Error string `json:"error"`
}

// ==================== Utility Functions ====================

// FormatTime converts time.Time to ISO 8601 string
func FormatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

// FormatTimePtr converts *time.Time to *string (ISO 8601)
func FormatTimePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format(time.RFC3339)
	return &s
}
