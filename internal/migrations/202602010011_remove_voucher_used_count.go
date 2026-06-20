package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// removeVoucherUsedCount drops the unused used_count column from vouchers table
// Migration 000011 - 2026-02-01
func removeVoucherUsedCount() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202602010011_remove_voucher_used_count",
		Migrate: func(tx *gorm.DB) error {
			// Drop used_count column (no longer used after removing redeem functionality)
			return tx.Exec("ALTER TABLE vouchers DROP COLUMN IF EXISTS used_count").Error
		},
		Rollback: func(tx *gorm.DB) error {
			// Re-add used_count column with default value
			return tx.Exec("ALTER TABLE vouchers ADD COLUMN used_count BIGINT DEFAULT 0").Error
		},
	}
}
