// Package migrations defines all database migrations using Gormigrate.
// This provides Laravel-like migration experience with up/down functions.
package migrations

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// validMigrationIdentifier matches safe PostgreSQL identifiers: lowercase letters, digits, and underscores.
// This mirrors the validIdentifier regex in internal/repository/base_repository.go.
var validMigrationIdentifier = regexp.MustCompile(`^[a-z0-9_]+$`)

// validateSQLIdentifiers checks that all provided identifier strings are safe for
// interpolation into DDL statements. Returns an error if any identifier is empty or
// contains characters outside [a-z0-9_].
func validateSQLIdentifiers(identifiers ...string) error {
	for _, id := range identifiers {
		if id == "" {
			return fmt.Errorf("SQL identifier must not be empty")
		}
		if !validMigrationIdentifier.MatchString(id) {
			return fmt.Errorf("SQL identifier %q contains invalid characters (only [a-z0-9_] allowed)", id)
		}
	}
	return nil
}

// Helper functions to reduce code duplication

var validTriggerTimings = map[string]bool{"BEFORE": true, "AFTER": true, "INSTEAD OF": true}
var validTriggerEvents = map[string]bool{"INSERT": true, "UPDATE": true, "DELETE": true, "TRUNCATE": true}

// validateTriggerEvent validates a trigger event string, which may be a compound
// expression like "INSERT OR UPDATE" or "INSERT OR UPDATE OR DELETE".
func validateTriggerEvent(event string) error {
	for _, part := range strings.Split(strings.ToUpper(event), " OR ") {
		part = strings.TrimSpace(part)
		if !validTriggerEvents[part] {
			return fmt.Errorf("createTrigger: invalid event %q", event)
		}
	}
	return nil
}

// createTrigger creates a database trigger
func createTrigger(tx *gorm.DB, triggerName, tableName, timing, event, functionName string) error {
	if err := validateSQLIdentifiers(triggerName, tableName, functionName); err != nil {
		return fmt.Errorf("createTrigger: %w", err)
	}
	if !validTriggerTimings[strings.ToUpper(timing)] {
		return fmt.Errorf("createTrigger: invalid timing %q", timing)
	}
	if err := validateTriggerEvent(event); err != nil {
		return err
	}
	// Drop existing trigger first (separate statement)
	if err := tx.Exec(fmt.Sprintf("DROP TRIGGER IF EXISTS %s ON %s", triggerName, tableName)).Error; err != nil {
		return err
	}

	// Create new trigger (separate statement)
	return tx.Exec(fmt.Sprintf(`
		CREATE TRIGGER %s
			%s %s ON %s
			FOR EACH ROW
			EXECUTE FUNCTION %s()
	`, triggerName, timing, event, tableName, functionName)).Error
}

// dropTrigger drops a database trigger
func dropTrigger(tx *gorm.DB, triggerName, tableName string) error {
	if err := validateSQLIdentifiers(triggerName, tableName); err != nil {
		return fmt.Errorf("dropTrigger: %w", err)
	}
	return tx.Exec(fmt.Sprintf("DROP TRIGGER IF EXISTS %s ON %s", triggerName, tableName)).Error
}

// createFunction creates a database function
func createFunction(tx *gorm.DB, functionSQL string) error {
	return tx.Exec(functionSQL).Error
}

// dropFunction drops a database function
func dropFunction(tx *gorm.DB, functionName string) error {
	if err := validateSQLIdentifiers(functionName); err != nil {
		return fmt.Errorf("dropFunction: %w", err)
	}
	return tx.Exec(fmt.Sprintf("DROP FUNCTION IF EXISTS %s()", functionName)).Error
}

// createIndex creates a database index
func createIndex(tx *gorm.DB, indexSQL string) error {
	return tx.Exec(indexSQL).Error
}

// dropIndex drops a database index
func dropIndex(tx *gorm.DB, indexName string) error {
	if err := validateSQLIdentifiers(indexName); err != nil {
		return fmt.Errorf("dropIndex: %w", err)
	}
	return tx.Exec(fmt.Sprintf("DROP INDEX IF EXISTS %s", indexName)).Error
}

// addComment adds a comment to a database object
func addComment(tx *gorm.DB, commentSQL string) error {
	return tx.Exec(commentSQL).Error
}

// GetMigrations returns all migrations in chronological order
func GetMigrations() []*gormigrate.Migration {
	return []*gormigrate.Migration{
		initSchema(),
		addGiftCardBalanceConstraint(),
		normalizeEmails(),
		addUserFavorites(),
		addCaseInsensitiveEmailIndex(),
		addGiftCardBalanceCache(),
		addUniqueConstraintsForRaceConditions(),
		replaceGlobalUniqueWithComposite(),
		addAuditLog(),
		addAuthProvider(),
		autoSetGiftCardCurrentBalance(),
		removeVoucherUsedCount(),
		removeUnusedColorColumns(),
		fixGiftCardBalanceExcludeSoftDeletes(),
		addNotifications(),
		addNotificationsSoftDelete(),
		fixShareUniqueConstraintsForSoftDelete(),
		addUserSecurityColumns(),
		addEmailVerificationAndTokens(),
		addUserLanguage(),
		addPushSubscriptions(),
		addExpiryReminderTracking(),
		addUserTOTP(),
		addVoucherCurrency(),
		addUserExpiryRemindersEnabled(),
		addUserShareEmailsEnabled(),
		addPerChannelNotificationPreferences(),
		addUserPasswordChangedAt(),
		addServerSideSessions(),
		fixNotificationPreferenceDefaults(),
		partialUniqueIndexesExcludeSoftDeleted(),
		addNotificationsArchivedAt(),
		addNotificationEmailDelivery(),
	}
}
