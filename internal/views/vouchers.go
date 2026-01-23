// Package views contains view models for templates.
package views

import (
	"savvy/internal/models"
)

// VoucherPermissions represents user permissions for a voucher
type VoucherPermissions struct {
	CanEdit    bool
	CanDelete  bool
	IsFavorite bool
}

// VoucherShowView contains all data needed for vouchers/show template
type VoucherShowView struct {
	Voucher         models.Voucher
	Merchants       []models.Merchant
	Shares          []models.VoucherShare
	User            *models.User
	Permissions     VoucherPermissions
	IsImpersonating bool
}

// VoucherEditView contains all data needed for vouchers/edit template
type VoucherEditView struct {
	Voucher         models.Voucher
	Merchants       []models.Merchant
	User            *models.User
	IsImpersonating bool
}

// VoucherIndexView contains all data needed for vouchers/index template
type VoucherIndexView struct {
	Vouchers        []models.Voucher
	User            *models.User
	IsImpersonating bool
}
