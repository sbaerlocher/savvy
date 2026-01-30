// Package vouchers contains HTTP request handlers for voucher management.
package vouchers

import (
	"savvy/internal/services"

	"gorm.io/gorm"
)

const (
	newMerchantValue = "new"
)

// Handler handles HTTP requests for voucher operations.
type Handler struct {
	voucherService services.VoucherServiceInterface
	authzService   services.AuthzServiceInterface
	db             *gorm.DB
}

// NewHandler creates a new voucher handler with the provided services.
func NewHandler(voucherService services.VoucherServiceInterface, authzService services.AuthzServiceInterface, db *gorm.DB) *Handler {
	return &Handler{
		voucherService: voucherService,
		authzService:   authzService,
		db:             db,
	}
}
