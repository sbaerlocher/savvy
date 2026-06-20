package migrations

import (
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// addUserFavorites creates the user_favorites table
// Equivalent to: migrations/000005_add_user_favorites.up.sql
func addUserFavorites() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202601250005_add_user_favorites",
		Migrate: func(tx *gorm.DB) error {
			type UserFavorite struct {
				ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
				UserID       uuid.UUID  `gorm:"type:uuid;not null;index:idx_user_favorites"`
				ResourceType string     `gorm:"type:varchar(50);not null;index:idx_user_favorites"`
				ResourceID   uuid.UUID  `gorm:"type:uuid;not null;index:idx_user_favorites"`
				CreatedAt    time.Time  `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP"`
				DeletedAt    *time.Time `gorm:"type:timestamp with time zone;index:idx_user_favorites_deleted_at"`
			}

			// Create table
			if err := tx.AutoMigrate(&UserFavorite{}); err != nil {
				return err
			}

			// Add unique constraint
			if err := tx.Exec(`
				DO $$
				BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM pg_constraint WHERE conname = 'user_favorites_unique'
					) THEN
						ALTER TABLE user_favorites ADD CONSTRAINT user_favorites_unique UNIQUE (user_id, resource_type, resource_id);
					END IF;
				END $$;
			`).Error; err != nil {
				return err
			}

			// Add foreign key (idempotent)
			if err := tx.Exec(`
				DO $$
				BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM pg_constraint WHERE conname = 'fk_user_favorites_user'
					) THEN
						ALTER TABLE user_favorites
						ADD CONSTRAINT fk_user_favorites_user
						FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
					END IF;
				END $$;
			`).Error; err != nil {
				return err
			}

			// Add table comment
			if err := tx.Exec(`
				COMMENT ON TABLE user_favorites IS 'User-specific favorites for cards, vouchers, and gift cards'
			`).Error; err != nil {
				return err
			}

			// Add column comments
			if err := tx.Exec(`
				COMMENT ON COLUMN user_favorites.resource_type IS 'Type of resource: card, voucher, or gift_card'
			`).Error; err != nil {
				return err
			}

			if err := tx.Exec(`
				COMMENT ON COLUMN user_favorites.resource_id IS 'UUID of the favorited resource'
			`).Error; err != nil {
				return err
			}

			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Exec("DROP TABLE IF EXISTS user_favorites CASCADE").Error
		},
	}
}
