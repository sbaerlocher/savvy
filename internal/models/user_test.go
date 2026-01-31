package models

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUser_DisplayName(t *testing.T) {
	user := &User{
		FirstName: "John",
		LastName:  "Doe",
	}

	displayName := user.DisplayName()

	assert.Equal(t, "John Doe", displayName)
}

func TestUser_DisplayName_EmptyLastName(t *testing.T) {
	user := &User{
		FirstName: "John",
		LastName:  "",
	}

	displayName := user.DisplayName()

	assert.Equal(t, "John ", displayName)
}

func TestUser_IsAdmin_True(t *testing.T) {
	user := &User{
		Role: "admin",
	}

	assert.True(t, user.IsAdmin())
}

func TestUser_IsAdmin_False(t *testing.T) {
	user := &User{
		Role: "user",
	}

	assert.False(t, user.IsAdmin())
}

func TestUser_IsOAuthUser_True(t *testing.T) {
	user := &User{
		AuthProvider: "oauth",
	}

	assert.True(t, user.IsOAuthUser())
}

func TestUser_IsOAuthUser_False(t *testing.T) {
	user := &User{
		AuthProvider: "local",
	}

	assert.False(t, user.IsOAuthUser())
}

func TestUser_BeforeCreate(t *testing.T) {
	user := &User{
		Email: "test@example.com",
	}

	// ID should be Nil before BeforeCreate
	assert.Equal(t, uuid.Nil, user.ID)

	err := user.BeforeCreate(nil)

	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, user.ID)
}

func TestUser_BeforeCreate_ExistingID(t *testing.T) {
	existingID := uuid.New()
	user := &User{
		ID:    existingID,
		Email: "test@example.com",
	}

	err := user.BeforeCreate(nil)

	assert.NoError(t, err)
	assert.Equal(t, existingID, user.ID) // ID should not change
}
