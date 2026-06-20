package migrations

import (
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// addAuditLog creates the audit_logs table for tracking all deletion operations
// Migration 000010 - 2026-01-26
func addAuditLog() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202601260010_add_audit_log",
		Migrate: func(tx *gorm.DB) error {
			type AuditLog struct {
				ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
				UserID       *uuid.UUID `gorm:"type:uuid;index"`
				Action       string     `gorm:"type:varchar(50);not null;index"`
				ResourceType string     `gorm:"type:varchar(50);not null;index"`
				ResourceID   uuid.UUID  `gorm:"type:uuid;not null;index"`
				ResourceData string     `gorm:"type:jsonb"`
				IPAddress    string     `gorm:"type:varchar(45)"`
				UserAgent    string     `gorm:"type:text"`
				CreatedAt    time.Time  `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP;index"`
			}

			// Create table
			if err := tx.AutoMigrate(&AuditLog{}); err != nil {
				return err
			}

			// Add foreign key constraint for user_id (idempotent)
			if err := tx.Exec(`
				DO $$
				BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM pg_constraint WHERE conname = 'fk_audit_logs_user'
					) THEN
						ALTER TABLE audit_logs
						ADD CONSTRAINT fk_audit_logs_user
						FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;
					END IF;
				END $$;
			`).Error; err != nil {
				return err
			}

			// Add composite index for common queries
			if err := tx.Exec(`
				CREATE INDEX IF NOT EXISTS idx_audit_logs_resource
				ON audit_logs (resource_type, resource_id, created_at DESC);
			`).Error; err != nil {
				return err
			}

			// Add index for user queries
			if err := tx.Exec(`
				CREATE INDEX IF NOT EXISTS idx_audit_logs_user_created
				ON audit_logs (user_id, created_at DESC)
				WHERE user_id IS NOT NULL;
			`).Error; err != nil {
				return err
			}

			// Add table comment
			if err := tx.Exec(`
				COMMENT ON TABLE audit_logs IS 'Audit trail for all deletion operations in the system for compliance and traceability';
			`).Error; err != nil {
				return err
			}

			// Add column comments (each must be a separate statement)
			if err := tx.Exec("COMMENT ON COLUMN audit_logs.action IS 'Type of action: delete, hard_delete, restore'").Error; err != nil {
				return err
			}
			if err := tx.Exec("COMMENT ON COLUMN audit_logs.resource_type IS 'Type of resource: cards, vouchers, gift_cards, etc.'").Error; err != nil {
				return err
			}
			if err := tx.Exec("COMMENT ON COLUMN audit_logs.resource_data IS 'JSON snapshot of the deleted resource'").Error; err != nil {
				return err
			}
			if err := tx.Exec("COMMENT ON COLUMN audit_logs.ip_address IS 'IP address of the user who performed the action'").Error; err != nil {
				return err
			}
			if err := tx.Exec("COMMENT ON COLUMN audit_logs.user_agent IS 'Browser user agent string'").Error; err != nil {
				return err
			}

			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Exec("DROP TABLE IF EXISTS audit_logs CASCADE").Error
		},
	}
}
