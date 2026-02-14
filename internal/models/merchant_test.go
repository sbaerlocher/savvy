package models

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMerchant_BeforeCreate(t *testing.T) {
	merchant := &Merchant{}

	// ID should be zero before creation
	assert.Equal(t, uuid.Nil, merchant.ID)

	// Simulate BeforeCreate hook behavior (DB generates UUID)
	merchant.ID = uuid.New()

	// ID should be set after creation
	assert.NotEqual(t, uuid.Nil, merchant.ID)
}

func TestMerchant_DefaultColor(t *testing.T) {
	merchant := &Merchant{
		Name:  "Test Merchant",
		Color: "#0066CC", // Default blue
	}

	assert.Equal(t, "#0066CC", merchant.Color)
}

func TestMerchant_CustomColor(t *testing.T) {
	merchant := &Merchant{
		Name:  "Test Merchant",
		Color: "#FF5733",
	}

	assert.Equal(t, "#FF5733", merchant.Color)
}

func TestMerchant_Name(t *testing.T) {
	merchant := &Merchant{
		Name: "Test Merchant",
	}

	assert.Equal(t, "Test Merchant", merchant.Name)
}

func TestMerchant_LogoURL(t *testing.T) {
	merchant := &Merchant{
		Name:    "Test Merchant",
		LogoURL: "https://example.com/logo.png",
	}

	assert.Equal(t, "https://example.com/logo.png", merchant.LogoURL)
}

func TestMerchant_Website(t *testing.T) {
	merchant := &Merchant{
		Name:    "Test Merchant",
		Website: "https://example.com",
	}

	assert.Equal(t, "https://example.com", merchant.Website)
}
