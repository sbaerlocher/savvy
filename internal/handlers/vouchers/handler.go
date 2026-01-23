// Package vouchers contains HTTP request handlers for voucher management.
package vouchers

import (
	"savvy/internal/services"

	"gorm.io/gorm"
)

type Handler struct {
	voucherService services.VoucherServiceInterface
	authzService   services.AuthzServiceInterface
	db             *gorm.DB
}

func NewHandler(voucherService services.VoucherServiceInterface, authzService services.AuthzServiceInterface, db *gorm.DB) *Handler {
	return &Handler{
		voucherService: voucherService,
		authzService:   authzService,
		db:             db,
	}
}
