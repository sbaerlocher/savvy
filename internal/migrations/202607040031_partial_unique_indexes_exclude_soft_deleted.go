package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// partialUniqueIndexesExcludeSoftDeleted rewrites the composite unique indexes
// on (user_id, card_number/code) to also exclude soft-deleted rows, so a user
// can re-create an entry whose number matches one they previously soft-deleted.
// Migration 000031 - 2026-07-04
func partialUniqueIndexesExcludeSoftDeleted() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202607040031_partial_unique_indexes_exclude_soft_deleted",
		Migrate: func(tx *gorm.DB) error {
			steps := []struct{ drop, create string }{
				{"idx_cards_user_card_number",
					"CREATE UNIQUE INDEX IF NOT EXISTS idx_cards_user_card_number ON cards (user_id, card_number) WHERE user_id IS NOT NULL AND deleted_at IS NULL"},
				{"idx_vouchers_user_code",
					"CREATE UNIQUE INDEX IF NOT EXISTS idx_vouchers_user_code ON vouchers (user_id, code) WHERE user_id IS NOT NULL AND deleted_at IS NULL"},
				{"idx_gift_cards_user_card_number",
					"CREATE UNIQUE INDEX IF NOT EXISTS idx_gift_cards_user_card_number ON gift_cards (user_id, card_number) WHERE user_id IS NOT NULL AND deleted_at IS NULL"},
			}
			for _, s := range steps {
				if err := dropIndex(tx, s.drop); err != nil {
					return err
				}
				if err := createIndex(tx, s.create); err != nil {
					return err
				}
			}
			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			steps := []struct{ drop, create string }{
				{"idx_cards_user_card_number",
					"CREATE UNIQUE INDEX IF NOT EXISTS idx_cards_user_card_number ON cards (user_id, card_number) WHERE user_id IS NOT NULL"},
				{"idx_vouchers_user_code",
					"CREATE UNIQUE INDEX IF NOT EXISTS idx_vouchers_user_code ON vouchers (user_id, code) WHERE user_id IS NOT NULL"},
				{"idx_gift_cards_user_card_number",
					"CREATE UNIQUE INDEX IF NOT EXISTS idx_gift_cards_user_card_number ON gift_cards (user_id, card_number) WHERE user_id IS NOT NULL"},
			}
			for _, s := range steps {
				if err := dropIndex(tx, s.drop); err != nil {
					return err
				}
				if err := createIndex(tx, s.create); err != nil {
					return err
				}
			}
			return nil
		},
	}
}
