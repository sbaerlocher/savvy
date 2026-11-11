package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// partialUniqueIndexMerchantName allows a new active merchant to reuse a
// name held only by a soft-deleted merchant.
func partialUniqueIndexMerchantName() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202611110034_partial_unique_index_merchant_name",
		Migrate: func(tx *gorm.DB) error {
			if err := dropIndex(tx, "idx_merchants_name"); err != nil {
				return err
			}
			return createIndex(tx, "CREATE UNIQUE INDEX IF NOT EXISTS idx_merchants_name ON merchants (name) WHERE deleted_at IS NULL")
		},
		// Rollback restores global uniqueness. It intentionally fails if active
		// and soft-deleted rows now share a name, preserving existing data.
		Rollback: func(tx *gorm.DB) error {
			if err := dropIndex(tx, "idx_merchants_name"); err != nil {
				return err
			}
			return createIndex(tx, "CREATE UNIQUE INDEX IF NOT EXISTS idx_merchants_name ON merchants (name)")
		},
	}
}
