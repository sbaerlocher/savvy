// Package models defines the database models for the savvy system.
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserFavorite represents a user's favorite item (card, voucher, or gift card)
// This allows each user to have their own favorites, even for shared items
type UserFavorite struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_user_favorites_unique" json:"user_id"`
	User         *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	ResourceType string         `gorm:"not null;uniqueIndex:idx_user_favorites_unique" json:"resource_type"` // "card", "voucher", "gift_card"
	ResourceID   uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_user_favorites_unique" json:"resource_id"`
	CreatedAt    time.Time      `json:"created_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName specifies the table name for UserFavorite
func (UserFavorite) TableName() string {
	return "user_favorites"
}
