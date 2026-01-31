package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCard_GetColor_WithMerchant(t *testing.T) {
	card := &Card{
		Merchant: &Merchant{
			Color: "#FF5733",
		},
	}

	color := card.GetColor()

	assert.Equal(t, "#FF5733", color)
}

func TestCard_GetColor_WithoutMerchant(t *testing.T) {
	card := &Card{}

	color := card.GetColor()

	assert.Equal(t, "#0066CC", color) // Default color
}

func TestCard_GetColor_MerchantWithoutColor(t *testing.T) {
	card := &Card{
		Merchant: &Merchant{
			Color: "",
		},
	}

	color := card.GetColor()

	assert.Equal(t, "#0066CC", color) // Default color
}
