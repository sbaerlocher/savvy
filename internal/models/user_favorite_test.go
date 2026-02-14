package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserFavorite_TableName(t *testing.T) {
	favorite := UserFavorite{}

	tableName := favorite.TableName()

	assert.Equal(t, "user_favorites", tableName)
}
