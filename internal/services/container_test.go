package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestNewContainer(t *testing.T) {
	// Use nil DB since we're just testing that the container initializes
	// The actual services will be non-nil wrappers around repositories
	var db *gorm.DB = nil
	container := NewContainer(db)

	// Verify all services are initialized
	assert.NotNil(t, container)
	assert.NotNil(t, container.CardService)
	assert.NotNil(t, container.VoucherService)
	assert.NotNil(t, container.GiftCardService)
	assert.NotNil(t, container.MerchantService)
	assert.NotNil(t, container.ShareService)
	assert.NotNil(t, container.FavoriteService)
	assert.NotNil(t, container.AuthzService)
	assert.NotNil(t, container.DashboardService)

	// Verify services implement their interfaces
	var _ CardServiceInterface = container.CardService
	var _ VoucherServiceInterface = container.VoucherService
	var _ GiftCardServiceInterface = container.GiftCardService
	var _ MerchantServiceInterface = container.MerchantService
	var _ ShareServiceInterface = container.ShareService
	var _ FavoriteServiceInterface = container.FavoriteService
	var _ AuthzServiceInterface = container.AuthzService
	var _ DashboardServiceInterface = container.DashboardService
}
