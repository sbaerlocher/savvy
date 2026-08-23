// Package main seeds the database with sample data for testing.
package main

import (
	"encoding/json"
	"log"
	"savvy/internal/config"
	"savvy/internal/database"
	"savvy/internal/models"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// createGiftCardTransactions creates transactions for a gift card and handles duplicates
func createGiftCardTransactions(transactions []models.GiftCardTransaction) {
	for _, t := range transactions {
		var existing models.GiftCardTransaction
		if err := database.DB.Where("gift_card_id = ? AND description = ?", t.GiftCardID, t.Description).First(&existing).Error; err == nil {
			log.Printf("  • Transaction already exists: %s", t.Description)
		} else {
			database.DB.Create(&t)
			log.Printf("  ✓ Created transaction: %s (%.2f)", t.Description, t.Amount)
		}
	}
}

// createUsers creates test users and returns them
func createUsers(hashedPassword string) []models.User {
	now := time.Now()

	users := []models.User{
		{
			Email:        "admin@example.com",
			PasswordHash: hashedPassword,
			FirstName:    "Admin",
			LastName:     "User",
			Role:         "admin",
			AuthProvider: "local",
			Language:     "de",
			// Admin: all notifications enabled (defaults), email verified
			EmailVerified:             true,
			EmailVerifiedAt:           &now,
			PushNotificationsEnabled:  true,
			EmailNotificationsEnabled: true,
			PushRemindersEnabled:      true,
			PushSharingEnabled:        true,
			EmailRemindersEnabled:     true,
			EmailSharingEnabled:       true,
		},
		{
			Email:        "anna.mueller@example.com",
			PasswordHash: hashedPassword,
			FirstName:    "Anna",
			LastName:     "Müller",
			Role:         "user",
			AuthProvider: "local",
			Language:     "de",
			// Anna: email verified, push only (email notifications off)
			EmailVerified:             true,
			EmailVerifiedAt:           &now,
			PushNotificationsEnabled:  true,
			EmailNotificationsEnabled: false,
			PushRemindersEnabled:      true,
			PushSharingEnabled:        true,
			EmailRemindersEnabled:     false,
			EmailSharingEnabled:       false,
		},
		{
			Email:        "thomas.schmidt@example.com",
			PasswordHash: hashedPassword,
			FirstName:    "Thomas",
			LastName:     "Schmidt",
			Role:         "user",
			AuthProvider: "local",
			Language:     "en",
			// Thomas: email only (push off), reminders on, sharing off
			EmailVerified:             true,
			EmailVerifiedAt:           &now,
			PushNotificationsEnabled:  false,
			EmailNotificationsEnabled: true,
			PushRemindersEnabled:      false,
			PushSharingEnabled:        false,
			EmailRemindersEnabled:     true,
			EmailSharingEnabled:       false,
		},
		{
			Email:        "maria.garcia@example.com",
			PasswordHash: hashedPassword,
			FirstName:    "Maria",
			LastName:     "Garcia",
			Role:         "user",
			AuthProvider: "local",
			Language:     "fr",
			// Maria: all notifications off, email not verified
			EmailVerified:             false,
			PushNotificationsEnabled:  false,
			EmailNotificationsEnabled: false,
			PushRemindersEnabled:      false,
			PushSharingEnabled:        false,
			EmailRemindersEnabled:     false,
			EmailSharingEnabled:       false,
		},
	}

	log.Println("Creating users...")
	for i := range users {
		var existing models.User
		if err := database.DB.Where("email = ?", users[i].Email).First(&existing).Error; err == nil {
			log.Printf("  • User already exists: %s (updating settings)", users[i].Email)
			// Update user settings to match seed
			database.DB.Model(&existing).Updates(map[string]interface{}{
				"language":                    users[i].Language,
				"email_verified":              users[i].EmailVerified,
				"email_verified_at":           users[i].EmailVerifiedAt,
				"push_notifications_enabled":  users[i].PushNotificationsEnabled,
				"email_notifications_enabled": users[i].EmailNotificationsEnabled,
				"push_reminders_enabled":      users[i].PushRemindersEnabled,
				"push_sharing_enabled":        users[i].PushSharingEnabled,
				"email_reminders_enabled":     users[i].EmailRemindersEnabled,
				"email_sharing_enabled":       users[i].EmailSharingEnabled,
			})
			users[i] = existing
		} else {
			if err := database.DB.Create(&users[i]).Error; err != nil {
				log.Fatal(err)
			}
			log.Printf("  ✓ Created user: %s (%s %s, lang=%s)", users[i].Email, users[i].FirstName, users[i].LastName, users[i].Language)
		}
	}
	return users
}

// createMerchants creates test merchants and returns them
func createMerchants() []models.Merchant {
	merchants := []models.Merchant{
		{Name: "Migros", LogoURL: "", Website: "https://www.migros.ch", Color: "#FF6B35"},
		{Name: "Coop", LogoURL: "", Website: "https://www.coop.ch", Color: "#F7931E"},
		{Name: "Manor", LogoURL: "", Website: "https://www.manor.ch", Color: "#C41E3A"},
		{Name: "Media Markt", LogoURL: "", Website: "https://www.mediamarkt.ch", Color: "#DC2626"},
		{Name: "Digitec", LogoURL: "", Website: "https://www.digitec.ch", Color: "#0066CC"},
		{Name: "Galaxus", LogoURL: "", Website: "https://www.galaxus.ch", Color: "#7C3AED"},
		{Name: "Interdiscount", LogoURL: "", Website: "https://www.interdiscount.ch", Color: "#059669"},
		{Name: "Denner", LogoURL: "", Website: "https://www.denner.ch", Color: "#DC2626"},
		// New merchants for additional use cases
		{Name: "IKEA", LogoURL: "", Website: "https://www.ikea.ch", Color: "#0058A3"},
		{Name: "H&M", LogoURL: "", Website: "", Color: "#E50010"},       // No website
		{Name: "Starbucks", LogoURL: "", Website: "", Color: "#00704A"}, // No website
		{Name: "Apple Store", LogoURL: "", Website: "https://www.apple.com", Color: "#555555"},
	}

	log.Println("Creating merchants...")
	for i := range merchants {
		var existing models.Merchant
		if err := database.DB.Where("name = ?", merchants[i].Name).First(&existing).Error; err == nil {
			log.Printf("  • Merchant already exists: %s", merchants[i].Name)
			merchants[i] = existing // Use existing merchant
		} else {
			if err := database.DB.Create(&merchants[i]).Error; err != nil {
				log.Fatal(err)
			}
			log.Printf("  ✓ Created merchant: %s", merchants[i].Name)
		}
	}
	return merchants
}

// merchantByName returns the merchant with the given name from the slice
func merchantByName(merchants []models.Merchant, name string) *models.Merchant {
	for i := range merchants {
		if merchants[i].Name == name {
			return &merchants[i]
		}
	}
	return nil
}

// createCards creates test cards for all users
func createCards(users []models.User, merchants []models.Merchant) {
	ikea := merchantByName(merchants, "IKEA")
	hm := merchantByName(merchants, "H&M")
	starbucks := merchantByName(merchants, "Starbucks")
	apple := merchantByName(merchants, "Apple Store")

	cards := []models.Card{
		// ===== Admin's cards - all barcode types & statuses =====
		// CODE128
		{
			UserID:       &users[0].ID,
			MerchantID:   &merchants[0].ID,
			MerchantName: merchants[0].Name,
			Program:      "Cumulus",
			CardNumber:   "7610000000001",
			BarcodeType:  "CODE128",
			Status:       "active",
			Notes:        "Hauptkarte Migros (CODE128)",
		},
		// EAN13
		{
			UserID:       &users[0].ID,
			MerchantID:   &merchants[1].ID,
			MerchantName: merchants[1].Name,
			Program:      "Supercard",
			CardNumber:   "7612345678900",
			BarcodeType:  "EAN13",
			Status:       "active",
			Notes:        "Coop Supercard (EAN13)",
		},
		// EAN8
		{
			UserID:       &users[0].ID,
			MerchantID:   &merchants[7].ID,
			MerchantName: merchants[7].Name,
			Program:      "Denner Card",
			CardNumber:   "12345670",
			BarcodeType:  "EAN8",
			Status:       "active",
			Notes:        "Denner Kundenkarte (EAN8)",
		},
		// QR
		{
			UserID:       &users[0].ID,
			MerchantID:   &merchants[2].ID,
			MerchantName: merchants[2].Name,
			Program:      "Manor Card",
			CardNumber:   "MANOR-QR-123456",
			BarcodeType:  "QR",
			Status:       "active",
			Notes:        "Manor Kundenkarte (QR Code)",
		},
		// Inactive status
		{
			UserID:       &users[0].ID,
			MerchantID:   &merchants[3].ID,
			MerchantName: merchants[3].Name,
			Program:      "Media Markt Club",
			CardNumber:   "MM-OLD-CARD-999",
			BarcodeType:  "CODE128",
			Status:       "inactive",
			Notes:        "Alte Karte - nicht mehr aktiv",
		},
		// PDF417
		{
			UserID:       &users[0].ID,
			MerchantID:   &ikea.ID,
			MerchantName: ikea.Name,
			Program:      "IKEA Family",
			CardNumber:   "IKEA-PDF417-001",
			BarcodeType:  "PDF417",
			Status:       "active",
			Notes:        "IKEA Family Karte (PDF417 Barcode)",
		},
		// DATAMATRIX
		{
			UserID:       &users[0].ID,
			MerchantID:   &hm.ID,
			MerchantName: hm.Name,
			Program:      "H&M Member",
			CardNumber:   "HM-DATAMATRIX-002",
			BarcodeType:  "DATAMATRIX",
			Status:       "active",
			Notes:        "H&M Mitgliedskarte (DataMatrix)",
		},
		// CODE39
		{
			UserID:       &users[0].ID,
			MerchantID:   &starbucks.ID,
			MerchantName: starbucks.Name,
			Program:      "Starbucks Rewards",
			CardNumber:   "SBUX-CODE39-003",
			BarcodeType:  "CODE39",
			Status:       "active",
			Notes:        "Starbucks Rewards (CODE39)",
		},
		// AZTEC
		{
			UserID:       &users[0].ID,
			MerchantID:   &apple.ID,
			MerchantName: apple.Name,
			Program:      "Apple Wallet Card",
			CardNumber:   "APPLE-AZTEC-004",
			BarcodeType:  "AZTEC",
			Status:       "active",
			Notes:        "Apple Store Card (Aztec Code)",
		},
		// UPCA
		{
			UserID:       &users[0].ID,
			MerchantID:   &merchants[4].ID,
			MerchantName: merchants[4].Name,
			Program:      "Digitec Bonus",
			CardNumber:   "012345678905",
			BarcodeType:  "UPCA",
			Status:       "active",
			Notes:        "Digitec Bonus Karte (UPC-A)",
		},
		// ITF
		{
			UserID:       &users[0].ID,
			MerchantID:   &merchants[5].ID,
			MerchantName: merchants[5].Name,
			Program:      "Galaxus Rewards",
			CardNumber:   "1234567890123456",
			BarcodeType:  "ITF",
			Status:       "active",
			Notes:        "Galaxus Rewards (ITF Barcode)",
		},
		// Card without notes
		{
			UserID:       &users[0].ID,
			MerchantID:   &merchants[6].ID,
			MerchantName: merchants[6].Name,
			Program:      "Interdiscount Card",
			CardNumber:   "ID-NO-NOTES-001",
			BarcodeType:  "CODE128",
			Status:       "active",
			Notes:        "",
		},
		// Card without merchant reference (free text only)
		{
			UserID:       &users[0].ID,
			MerchantID:   nil,
			MerchantName: "Lokaler Bäcker",
			Program:      "Treuekarte",
			CardNumber:   "BAECKER-STAMP-001",
			BarcodeType:  "QR",
			Status:       "active",
			Notes:        "Stempelkarte vom lokalen Bäcker (kein Merchant)",
		},
		// Expired card
		{
			UserID:       &users[0].ID,
			MerchantID:   &merchants[3].ID,
			MerchantName: merchants[3].Name,
			Program:      "Media Markt Aktion 2024",
			CardNumber:   "MM-EXPIRED-2024",
			BarcodeType:  "CODE128",
			Status:       "expired",
			Notes:        "Abgelaufene Aktionskarte",
		},

		// ===== Anna's cards =====
		{
			UserID:       &users[1].ID,
			MerchantID:   &merchants[4].ID,
			MerchantName: merchants[4].Name,
			Program:      "Digitec Club",
			CardNumber:   "DT-2024-ANNA-001",
			BarcodeType:  "CODE128",
			Status:       "active",
			Notes:        "Anna's Digitec Karte",
		},
		{
			UserID:       &users[1].ID,
			MerchantID:   &merchants[5].ID,
			MerchantName: merchants[5].Name,
			Program:      "Galaxus Club",
			CardNumber:   "GX-QR-ANNA-456",
			BarcodeType:  "QR",
			Status:       "active",
			Notes:        "Anna's Galaxus Karte (QR)",
		},
		// Anna's E2E test card - specifically for sharing tests
		{
			UserID:       &users[1].ID,
			MerchantID:   &merchants[0].ID,
			MerchantName: merchants[0].Name,
			Program:      "Cumulus",
			CardNumber:   "E2E-TEST-SHARING-001",
			BarcodeType:  "CODE128",
			Status:       "active",
			Notes:        "E2E Test Card for Sharing (owned by Anna)",
		},

		// ===== Thomas's cards =====
		{
			UserID:       &users[2].ID,
			MerchantID:   &merchants[6].ID,
			MerchantName: merchants[6].Name,
			Program:      "Interdiscount Club",
			CardNumber:   "ID-789-THOMAS",
			BarcodeType:  "CODE128",
			Status:       "active",
			Notes:        "Thomas's Interdiscount Karte",
		},

		// ===== Maria's card =====
		{
			UserID:       &users[3].ID,
			MerchantID:   &merchants[0].ID,
			MerchantName: merchants[0].Name,
			Program:      "Cumulus",
			CardNumber:   "7610000999888",
			BarcodeType:  "CODE128",
			Status:       "active",
			Notes:        "Maria's Migros Karte",
		},
	}

	log.Println("Creating cards (all barcode types & statuses)...")
	for _, card := range cards {
		var existing models.Card
		if err := database.DB.Where("card_number = ?", card.CardNumber).First(&existing).Error; err == nil {
			log.Printf("  • Card already exists: %s", card.CardNumber)
		} else {
			if err := database.DB.Create(&card).Error; err != nil {
				log.Fatal(err)
			}
			log.Printf("  ✓ Created card: %s - %s (%s, %s)", card.MerchantName, card.Program, card.BarcodeType, card.Status)
		}
	}
}

// createVouchers creates test vouchers for all users
func createVouchers(users []models.User, merchants []models.Merchant) {
	ikea := merchantByName(merchants, "IKEA")
	hm := merchantByName(merchants, "H&M")
	starbucks := merchantByName(merchants, "Starbucks")
	apple := merchantByName(merchants, "Apple Store")

	vouchers := []models.Voucher{
		// ===== Admin's vouchers =====

		// percentage, multiple_use_with_card, QR, CHF
		{
			UserID:            &users[0].ID,
			MerchantID:        &merchants[0].ID,
			MerchantName:      merchants[0].Name,
			Code:              "SUMMER2026",
			Type:              "percentage",
			Value:             20.0,
			Description:       "20% Sommerrabatt auf alle Artikel",
			MinPurchaseAmount: 50.0,
			ValidFrom:         time.Now().AddDate(0, 0, -7),
			ValidUntil:        time.Now().AddDate(0, 3, 0),
			UsageLimitType:    "multiple_use_with_card",
			BarcodeType:       "QR",
		},
		// fixed_amount CHF, single_use, CODE128
		{
			UserID:            &users[0].ID,
			MerchantID:        &merchants[1].ID,
			MerchantName:      merchants[1].Name,
			Code:              "WELCOME50",
			Type:              "fixed_amount",
			Value:             50.0,
			Currency:          "CHF",
			Description:       "50 CHF Willkommensbonus",
			MinPurchaseAmount: 100.0,
			ValidFrom:         time.Now(),
			ValidUntil:        time.Now().AddDate(0, 1, 0),
			UsageLimitType:    "single_use",
			BarcodeType:       "CODE128",
		},
		// percentage, one_per_customer, CODE128
		{
			UserID:            &users[0].ID,
			MerchantID:        &merchants[3].ID,
			MerchantName:      merchants[3].Name,
			Code:              "TECH15",
			Type:              "percentage",
			Value:             15.0,
			Description:       "15% Rabatt auf Elektronik",
			MinPurchaseAmount: 0.0,
			ValidFrom:         time.Now(),
			ValidUntil:        time.Now().AddDate(0, 2, 0),
			UsageLimitType:    "one_per_customer",
			BarcodeType:       "CODE128",
		},
		// points_multiplier, multiple_use_without_card, QR
		{
			UserID:            &users[0].ID,
			MerchantID:        &merchants[0].ID,
			MerchantName:      merchants[0].Name,
			Code:              "DOUBLE-POINTS",
			Type:              "points_multiplier",
			Value:             2.0,
			Description:       "Doppelte Cumulus-Punkte sammeln",
			MinPurchaseAmount: 0.0,
			ValidFrom:         time.Now(),
			ValidUntil:        time.Now().AddDate(0, 1, 0),
			UsageLimitType:    "multiple_use_without_card",
			BarcodeType:       "QR",
		},
		// fixed_amount CHF, multiple_use_without_card, long validity
		{
			UserID:            &users[0].ID,
			MerchantID:        &merchants[2].ID,
			MerchantName:      merchants[2].Name,
			Code:              "MANOR-FOREVER",
			Type:              "fixed_amount",
			Value:             10.0,
			Currency:          "CHF",
			Description:       "10 CHF Dauerrabatt",
			MinPurchaseAmount: 30.0,
			ValidFrom:         time.Now(),
			ValidUntil:        time.Now().AddDate(5, 0, 0),
			UsageLimitType:    "multiple_use_without_card",
			BarcodeType:       "CODE128",
		},
		// EXPIRED voucher
		{
			UserID:            &users[0].ID,
			MerchantID:        &merchants[1].ID,
			MerchantName:      merchants[1].Name,
			Code:              "EXPIRED2025",
			Type:              "percentage",
			Value:             30.0,
			Description:       "Abgelaufener Gutschein",
			MinPurchaseAmount: 0.0,
			ValidFrom:         time.Now().AddDate(0, -2, 0),
			ValidUntil:        time.Now().AddDate(0, -1, 0),
			UsageLimitType:    "single_use",
			BarcodeType:       "CODE128",
		},
		// FUTURE-VALID voucher (not yet valid)
		{
			UserID:            &users[0].ID,
			MerchantID:        &ikea.ID,
			MerchantName:      ikea.Name,
			Code:              "IKEA-FUTURE-2026",
			Type:              "percentage",
			Value:             25.0,
			Description:       "25% auf alles - gültig ab nächsten Monat",
			MinPurchaseAmount: 0.0,
			ValidFrom:         time.Now().AddDate(0, 1, 0),
			ValidUntil:        time.Now().AddDate(0, 3, 0),
			UsageLimitType:    "single_use",
			BarcodeType:       "PDF417",
		},
		// BECOMES VALID TOMORROW (for validity_start notification)
		{
			UserID:            &users[0].ID,
			MerchantID:        &hm.ID,
			MerchantName:      hm.Name,
			Code:              "HM-TOMORROW",
			Type:              "percentage",
			Value:             30.0,
			Description:       "30% Rabatt - wird morgen gültig",
			MinPurchaseAmount: 0.0,
			ValidFrom:         time.Now().AddDate(0, 0, 1),
			ValidUntil:        time.Now().AddDate(0, 1, 1),
			UsageLimitType:    "single_use",
			BarcodeType:       "DATAMATRIX",
		},
		// fixed_amount EUR
		{
			UserID:            &users[0].ID,
			MerchantID:        &apple.ID,
			MerchantName:      apple.Name,
			Code:              "APPLE-EUR-20",
			Type:              "fixed_amount",
			Value:             20.0,
			Currency:          "EUR",
			Description:       "20 EUR Apple Store Gutschein",
			MinPurchaseAmount: 50.0,
			ValidFrom:         time.Now(),
			ValidUntil:        time.Now().AddDate(0, 6, 0),
			UsageLimitType:    "single_use",
			BarcodeType:       "AZTEC",
		},
		// fixed_amount USD
		{
			UserID:            &users[0].ID,
			MerchantID:        &merchants[4].ID,
			MerchantName:      merchants[4].Name,
			Code:              "DIGITEC-USD-15",
			Type:              "fixed_amount",
			Value:             15.0,
			Currency:          "USD",
			Description:       "15 USD Digitec International",
			MinPurchaseAmount: 0.0,
			ValidFrom:         time.Now(),
			ValidUntil:        time.Now().AddDate(0, 3, 0),
			UsageLimitType:    "one_per_customer",
			BarcodeType:       "CODE128",
		},
		// fixed_amount GBP
		{
			UserID:            &users[0].ID,
			MerchantID:        &hm.ID,
			MerchantName:      hm.Name,
			Code:              "HM-GBP-10",
			Type:              "fixed_amount",
			Value:             10.0,
			Currency:          "GBP",
			Description:       "", // no description
			MinPurchaseAmount: 0.0,
			ValidFrom:         time.Now(),
			ValidUntil:        time.Now().AddDate(0, 2, 0),
			UsageLimitType:    "multiple_use_without_card",
			BarcodeType:       "CODE39",
		},
		// Voucher without description, no min purchase
		{
			UserID:            &users[0].ID,
			MerchantID:        &starbucks.ID,
			MerchantName:      starbucks.Name,
			Code:              "SBUX-5X-POINTS",
			Type:              "points_multiplier",
			Value:             5.0,
			Description:       "", // no description
			MinPurchaseAmount: 0.0,
			ValidFrom:         time.Now(),
			ValidUntil:        time.Now().AddDate(0, 1, 0),
			UsageLimitType:    "multiple_use_with_card",
			BarcodeType:       "QR",
		},
		// Voucher without merchant reference (free text)
		{
			UserID:         &users[0].ID,
			MerchantID:     nil,
			MerchantName:   "Restaurant Sternen",
			Code:           "STERNEN-LUNCH-50",
			Type:           "fixed_amount",
			Value:          50.0,
			Currency:       "CHF",
			Description:    "Mittagessen-Gutschein vom Restaurant",
			ValidFrom:      time.Now(),
			ValidUntil:     time.Now().AddDate(0, 6, 0),
			UsageLimitType: "single_use",
			BarcodeType:    "QR",
		},
		// Voucher expiring soon (3 days) - for expiry reminder testing
		{
			UserID:            &users[0].ID,
			MerchantID:        &merchants[5].ID,
			MerchantName:      merchants[5].Name,
			Code:              "GALAXUS-EXPIRING",
			Type:              "percentage",
			Value:             10.0,
			Description:       "Läuft bald ab!",
			MinPurchaseAmount: 0.0,
			ValidFrom:         time.Now().AddDate(0, -1, 0),
			ValidUntil:        time.Now().AddDate(0, 0, 3),
			UsageLimitType:    "single_use",
			BarcodeType:       "CODE128",
		},

		// ===== Anna's vouchers =====
		{
			UserID:            &users[1].ID,
			MerchantID:        &merchants[2].ID,
			MerchantName:      merchants[2].Name,
			Code:              "VIP2026",
			Type:              "fixed_amount",
			Value:             25.0,
			Currency:          "CHF",
			Description:       "VIP Bonus 25 CHF",
			MinPurchaseAmount: 75.0,
			ValidFrom:         time.Now(),
			ValidUntil:        time.Now().AddDate(0, 6, 0),
			UsageLimitType:    "one_per_customer",
			BarcodeType:       "QR",
		},
		{
			UserID:            &users[1].ID,
			MerchantID:        &merchants[1].ID,
			MerchantName:      merchants[1].Name,
			Code:              "TEST-EDIT-50",
			Type:              "fixed_amount",
			Value:             50.0,
			Currency:          "CHF",
			Description:       "Test voucher for E2E edit test",
			MinPurchaseAmount: 0.0,
			ValidFrom:         time.Now(),
			ValidUntil:        time.Now().AddDate(0, 6, 0),
			UsageLimitType:    "single_use",
			BarcodeType:       "CODE128",
		},
		{
			UserID:            &users[1].ID,
			MerchantID:        &merchants[4].ID,
			MerchantName:      merchants[4].Name,
			Code:              "DIGITEC-3X",
			Type:              "points_multiplier",
			Value:             3.0,
			Description:       "3x Punkte auf alle Einkäufe",
			MinPurchaseAmount: 0.0,
			ValidFrom:         time.Now(),
			ValidUntil:        time.Now().AddDate(0, 2, 0),
			UsageLimitType:    "multiple_use_with_card",
			BarcodeType:       "QR",
		},

		// ===== Thomas's vouchers =====
		{
			UserID:            &users[2].ID,
			MerchantID:        &merchants[6].ID,
			MerchantName:      merchants[6].Name,
			Code:              "ID-SAVE-100",
			Type:              "fixed_amount",
			Value:             100.0,
			Currency:          "CHF",
			Description:       "100 CHF Mega-Rabatt",
			MinPurchaseAmount: 500.0,
			ValidFrom:         time.Now(),
			ValidUntil:        time.Now().AddDate(0, 1, 0),
			UsageLimitType:    "single_use",
			BarcodeType:       "CODE128",
		},
	}

	log.Println("Creating vouchers (all types, currencies & usage limits)...")
	for _, voucher := range vouchers {
		var existing models.Voucher
		if err := database.DB.Where("code = ?", voucher.Code).First(&existing).Error; err == nil {
			log.Printf("  • Voucher already exists: %s", voucher.Code)
		} else {
			if err := database.DB.Create(&voucher).Error; err != nil {
				log.Fatal(err)
			}
			log.Printf("  ✓ Created voucher: %s (%s, %s, %s, %s)", voucher.Code, voucher.Type, voucher.Currency, voucher.UsageLimitType, voucher.BarcodeType)
		}
	}
}

// createGiftCards creates test gift cards and adds transactions
func createGiftCards(users []models.User, merchants []models.Merchant) {
	ikea := merchantByName(merchants, "IKEA")
	hm := merchantByName(merchants, "H&M")
	apple := merchantByName(merchants, "Apple Store")

	giftCards := []models.GiftCard{
		// ===== Admin's gift cards =====
		// CHF, active, with transactions, PIN, CODE128
		{
			UserID:         &users[0].ID,
			MerchantID:     &merchants[3].ID,
			MerchantName:   merchants[3].Name,
			CardNumber:     "MM1234567890",
			InitialBalance: 100.0,
			Currency:       "CHF",
			PIN:            "1234",
			ExpiresAt:      ptrTime(time.Now().AddDate(1, 0, 0)),
			Status:         "active",
			BarcodeType:    "CODE128",
			Notes:          "Geschenk zum Geburtstag - mit Transaktionen",
		},
		// CHF, active, with PIN, EAN13
		{
			UserID:         &users[0].ID,
			MerchantID:     &merchants[4].ID,
			MerchantName:   merchants[4].Name,
			CardNumber:     "7610200000002",
			InitialBalance: 200.0,
			Currency:       "CHF",
			PIN:            "5678",
			ExpiresAt:      ptrTime(time.Now().AddDate(2, 0, 0)),
			Status:         "active",
			BarcodeType:    "EAN13",
			Notes:          "Tech Shopping Karte (EAN13)",
		},
		// CHF, active, without PIN, QR
		{
			UserID:         &users[0].ID,
			MerchantID:     &merchants[5].ID,
			MerchantName:   merchants[5].Name,
			CardNumber:     "GX-CARD-QR-001",
			InitialBalance: 150.0,
			Currency:       "CHF",
			PIN:            "",
			ExpiresAt:      ptrTime(time.Now().AddDate(1, 6, 0)),
			Status:         "active",
			BarcodeType:    "QR",
			Notes:          "Galaxus Karte ohne PIN (QR)",
		},
		// CHF, expired/inactive
		{
			UserID:         &users[0].ID,
			MerchantID:     &merchants[2].ID,
			MerchantName:   merchants[2].Name,
			CardNumber:     "MANOR-EXPIRED-99",
			InitialBalance: 50.0,
			Currency:       "CHF",
			PIN:            "9999",
			ExpiresAt:      ptrTime(time.Now().AddDate(0, -1, 0)),
			Status:         "inactive",
			BarcodeType:    "CODE128",
			Notes:          "Abgelaufene Geschenkkarte",
		},
		// CHF, fully used (balance 0)
		{
			UserID:         &users[0].ID,
			MerchantID:     &merchants[1].ID,
			MerchantName:   merchants[1].Name,
			CardNumber:     "COOP-USED-777",
			InitialBalance: 75.0,
			Currency:       "CHF",
			PIN:            "0000",
			ExpiresAt:      ptrTime(time.Now().AddDate(0, 6, 0)),
			Status:         "active",
			BarcodeType:    "CODE128",
			Notes:          "Komplett aufgebrauchte Karte",
		},
		// EUR gift card
		{
			UserID:         &users[0].ID,
			MerchantID:     &apple.ID,
			MerchantName:   apple.Name,
			CardNumber:     "APPLE-EUR-500",
			InitialBalance: 500.0,
			Currency:       "EUR",
			PIN:            "4321",
			ExpiresAt:      ptrTime(time.Now().AddDate(2, 0, 0)),
			Status:         "active",
			BarcodeType:    "AZTEC",
			Notes:          "Apple Store EUR Geschenkkarte",
		},
		// USD gift card
		{
			UserID:         &users[0].ID,
			MerchantID:     &merchants[4].ID,
			MerchantName:   merchants[4].Name,
			CardNumber:     "DIGITEC-USD-100",
			InitialBalance: 100.0,
			Currency:       "USD",
			PIN:            "",
			ExpiresAt:      ptrTime(time.Now().AddDate(1, 0, 0)),
			Status:         "active",
			BarcodeType:    "DATAMATRIX",
			Notes:          "Digitec USD Gift Card",
		},
		// GBP gift card
		{
			UserID:         &users[0].ID,
			MerchantID:     &hm.ID,
			MerchantName:   hm.Name,
			CardNumber:     "HM-GBP-75",
			InitialBalance: 75.0,
			Currency:       "GBP",
			PIN:            "7777",
			ExpiresAt:      nil, // no expiry
			Status:         "active",
			BarcodeType:    "PDF417",
			Notes:          "H&M GBP Karte ohne Ablaufdatum",
		},
		// Gift card without notes, without PIN, no merchant ref
		{
			UserID:         &users[0].ID,
			MerchantID:     nil,
			MerchantName:   "Lokales Café",
			CardNumber:     "CAFE-GIFT-001",
			InitialBalance: 30.0,
			Currency:       "CHF",
			PIN:            "",
			ExpiresAt:      nil,
			Status:         "active",
			BarcodeType:    "QR",
			Notes:          "",
		},
		// Gift card with reload (negative transaction) - IKEA
		{
			UserID:         &users[0].ID,
			MerchantID:     &ikea.ID,
			MerchantName:   ikea.Name,
			CardNumber:     "IKEA-RELOAD-001",
			InitialBalance: 200.0,
			Currency:       "CHF",
			PIN:            "5555",
			ExpiresAt:      ptrTime(time.Now().AddDate(3, 0, 0)),
			Status:         "active",
			BarcodeType:    "CODE128",
			Notes:          "IKEA Karte mit Aufladungen und Einkäufen",
		},
		// Gift card expiring soon (for reminder testing)
		{
			UserID:         &users[0].ID,
			MerchantID:     &merchants[6].ID,
			MerchantName:   merchants[6].Name,
			CardNumber:     "ID-EXPIRING-SOON",
			InitialBalance: 40.0,
			Currency:       "CHF",
			PIN:            "",
			ExpiresAt:      ptrTime(time.Now().AddDate(0, 0, 5)),
			Status:         "active",
			BarcodeType:    "CODE128",
			Notes:          "Interdiscount - läuft in 5 Tagen ab",
		},

		// ===== Anna's gift cards =====
		{
			UserID:         &users[1].ID,
			MerchantID:     &merchants[2].ID,
			MerchantName:   merchants[2].Name,
			CardNumber:     "MANOR-ANNA-555",
			InitialBalance: 120.0,
			Currency:       "CHF",
			PIN:            "1111",
			ExpiresAt:      ptrTime(time.Now().AddDate(1, 0, 0)),
			Status:         "active",
			BarcodeType:    "QR",
			Notes:          "Anna's Manor Karte",
		},

		// ===== Thomas's gift cards =====
		{
			UserID:         &users[2].ID,
			MerchantID:     &merchants[6].ID,
			MerchantName:   merchants[6].Name,
			CardNumber:     "ID-THOMAS-888",
			InitialBalance: 80.0,
			Currency:       "CHF",
			PIN:            "2222",
			ExpiresAt:      ptrTime(time.Now().AddDate(0, 9, 0)),
			Status:         "active",
			BarcodeType:    "CODE128",
			Notes:          "Thomas's Interdiscount Karte",
		},

		// ===== Maria's gift cards =====
		// no expiry
		{
			UserID:         &users[3].ID,
			MerchantID:     &merchants[0].ID,
			MerchantName:   merchants[0].Name,
			CardNumber:     "MIGROS-MARIA-333",
			InitialBalance: 60.0,
			Currency:       "CHF",
			PIN:            "",
			ExpiresAt:      nil,
			Status:         "active",
			BarcodeType:    "CODE128",
			Notes:          "Maria's Migros Karte ohne Ablaufdatum",
		},
	}

	log.Println("Creating gift cards (all currencies, statuses & barcode types)...")
	for _, giftCard := range giftCards {
		var existing models.GiftCard
		if err := database.DB.Where("card_number = ?", giftCard.CardNumber).First(&existing).Error; err == nil {
			log.Printf("  • Gift card already exists: %s", giftCard.CardNumber)
		} else {
			if err := database.DB.Create(&giftCard).Error; err != nil {
				log.Fatal(err)
			}
			log.Printf("  ✓ Created gift card: %s - %s (%s, %s)", giftCard.MerchantName, giftCard.CardNumber, giftCard.Currency, giftCard.BarcodeType)
		}
	}

	// Add transactions to gift cards
	log.Println("Creating gift card transactions...")

	// Media Markt card - multiple spending transactions
	var mediaMarktGC models.GiftCard
	database.DB.Where("card_number = ?", "MM1234567890").First(&mediaMarktGC)
	if mediaMarktGC.ID.String() != "00000000-0000-0000-0000-000000000000" { //nolint:goconst
		transactions1 := []models.GiftCardTransaction{
			{
				GiftCardID:      mediaMarktGC.ID,
				Amount:          25.50,
				Description:     "Kopfhörer gekauft",
				TransactionDate: time.Now().AddDate(0, 0, -5),
			},
			{
				GiftCardID:      mediaMarktGC.ID,
				Amount:          10.00,
				Description:     "USB Kabel",
				TransactionDate: time.Now().AddDate(0, 0, -2),
			},
			{
				GiftCardID:      mediaMarktGC.ID,
				Amount:          50.00,
				Description:     "Aufladung",
				TransactionDate: time.Now().AddDate(0, 0, -1),
			},
		}
		createGiftCardTransactions(transactions1)
	}

	// Digitec card - add test transaction for E2E delete test
	var digitecGC models.GiftCard
	database.DB.Where("card_number = ?", "7610200000002").First(&digitecGC)
	if digitecGC.ID.String() != "00000000-0000-0000-0000-000000000000" { //nolint:goconst
		transactionsDigitec := []models.GiftCardTransaction{
			{
				GiftCardID:      digitecGC.ID,
				Amount:          5.00,
				Description:     "Test transaction for deletion",
				TransactionDate: time.Now().AddDate(0, 0, -1),
			},
		}
		createGiftCardTransactions(transactionsDigitec)
	}

	// Coop fully used card - transactions that sum to initial balance (75 CHF)
	var coopUsedGC models.GiftCard
	database.DB.Where("card_number = ?", "COOP-USED-777").First(&coopUsedGC)
	if coopUsedGC.ID.String() != "00000000-0000-0000-0000-000000000000" { //nolint:goconst
		transactions2 := []models.GiftCardTransaction{
			{
				GiftCardID:      coopUsedGC.ID,
				Amount:          30.00,
				Description:     "Einkauf 1",
				TransactionDate: time.Now().AddDate(0, 0, -10),
			},
			{
				GiftCardID:      coopUsedGC.ID,
				Amount:          25.00,
				Description:     "Einkauf 2",
				TransactionDate: time.Now().AddDate(0, 0, -7),
			},
			{
				GiftCardID:      coopUsedGC.ID,
				Amount:          20.00,
				Description:     "Einkauf 3 (letzter Rest)",
				TransactionDate: time.Now().AddDate(0, 0, -3),
			},
		}
		createGiftCardTransactions(transactions2)
	}

	// Apple EUR card - partial spend
	var appleGC models.GiftCard
	database.DB.Where("card_number = ?", "APPLE-EUR-500").First(&appleGC)
	if appleGC.ID.String() != "00000000-0000-0000-0000-000000000000" {
		transactionsApple := []models.GiftCardTransaction{
			{
				GiftCardID:      appleGC.ID,
				Amount:          149.00,
				Description:     "AirPods Pro",
				TransactionDate: time.Now().AddDate(0, 0, -14),
			},
			{
				GiftCardID:      appleGC.ID,
				Amount:          29.99,
				Description:     "Lightning Kabel",
				TransactionDate: time.Now().AddDate(0, 0, -7),
			},
		}
		createGiftCardTransactions(transactionsApple)
	}

	// IKEA card - spend + reload (negative = credit)
	var ikeaGC models.GiftCard
	database.DB.Where("card_number = ?", "IKEA-RELOAD-001").First(&ikeaGC)
	if ikeaGC.ID.String() != "00000000-0000-0000-0000-000000000000" {
		transactionsIKEA := []models.GiftCardTransaction{
			{
				GiftCardID:      ikeaGC.ID,
				Amount:          85.50,
				Description:     "Billy Regal",
				TransactionDate: time.Now().AddDate(0, -1, 0),
			},
			{
				GiftCardID:      ikeaGC.ID,
				Amount:          45.00,
				Description:     "KALLAX Regal",
				TransactionDate: time.Now().AddDate(0, 0, -20),
			},
			{
				GiftCardID:      ikeaGC.ID,
				Amount:          -100.00, // Reload (negative = credit)
				Description:     "Aufladung Geburtstag",
				TransactionDate: time.Now().AddDate(0, 0, -10),
			},
			{
				GiftCardID:      ikeaGC.ID,
				Amount:          22.90,
				Description:     "LACK Tisch",
				TransactionDate: time.Now().AddDate(0, 0, -3),
			},
		}
		createGiftCardTransactions(transactionsIKEA)
	}

	// Anna's Manor card - transaction so the regular user owns a gift card
	// with nested transactions (needed by the export-completeness E2E test)
	var manorAnnaGC models.GiftCard
	database.DB.Where("card_number = ?", "MANOR-ANNA-555").First(&manorAnnaGC)
	if manorAnnaGC.ID.String() != "00000000-0000-0000-0000-000000000000" {
		transactionsManorAnna := []models.GiftCardTransaction{
			{
				GiftCardID:      manorAnnaGC.ID,
				Amount:          20.00,
				Description:     "Einkauf Manor",
				TransactionDate: time.Now().AddDate(0, 0, -3),
			},
		}
		createGiftCardTransactions(transactionsManorAnna)
	}
}

// createShares creates test shares for all resource types
func createShares(users []models.User) {
	log.Println("Creating shares (all permission combinations)...")

	// Get created items for sharing
	var migrosCard, coopCard, dennerCard, manorCard models.Card
	database.DB.Where("merchant_name = ? AND user_id = ?", "Migros", users[0].ID).First(&migrosCard)
	database.DB.Where("merchant_name = ? AND user_id = ?", "Coop", users[0].ID).First(&coopCard)
	database.DB.Where("merchant_name = ? AND user_id = ?", "Denner", users[0].ID).First(&dennerCard)
	database.DB.Where("merchant_name = ? AND user_id = ?", "Manor", users[0].ID).First(&manorCard)

	var summerVoucher, welcomeVoucher, techVoucher, doublePointsVoucher models.Voucher
	database.DB.Where("code = ?", "SUMMER2026").First(&summerVoucher)
	database.DB.Where("code = ?", "WELCOME50").First(&welcomeVoucher)
	database.DB.Where("code = ?", "TECH15").First(&techVoucher)
	database.DB.Where("code = ?", "DOUBLE-POINTS").First(&doublePointsVoucher)

	var digitecGC, galaxusGC models.GiftCard
	database.DB.Where("card_number = ?", "7610200000002").First(&digitecGC)
	database.DB.Where("card_number = ?", "GX-CARD-QR-001").First(&galaxusGC)

	// Card Shares - all permission combinations
	cardShares := []models.CardShare{
		// Full permissions (edit + delete)
		{
			CardID:       migrosCard.ID,
			SharedWithID: users[1].ID, // Anna
			CanEdit:      true,
			CanDelete:    true,
		},
		// View only (no permissions)
		{
			CardID:       migrosCard.ID,
			SharedWithID: users[2].ID, // Thomas
			CanEdit:      false,
			CanDelete:    false,
		},
		// Edit only (no delete)
		{
			CardID:       coopCard.ID,
			SharedWithID: users[2].ID, // Thomas
			CanEdit:      true,
			CanDelete:    false,
		},
		// Delete only (no edit) - edge case
		{
			CardID:       dennerCard.ID,
			SharedWithID: users[3].ID, // Maria
			CanEdit:      false,
			CanDelete:    true,
		},
		// Share QR card with full permissions
		{
			CardID:       manorCard.ID,
			SharedWithID: users[1].ID, // Anna
			CanEdit:      true,
			CanDelete:    true,
		},
	}

	for _, share := range cardShares {
		var existing models.CardShare
		if err := database.DB.Where("card_id = ? AND shared_with_id = ?", share.CardID, share.SharedWithID).First(&existing).Error; err == nil {
			log.Printf("  • Card share already exists")
		} else {
			if err := database.DB.Create(&share).Error; err != nil {
				log.Printf("  ⚠ Failed to create card share: %v", err)
			} else {
				log.Printf("  ✓ Created card share (edit: %v, delete: %v)", share.CanEdit, share.CanDelete)
			}
		}
	}

	// Voucher Shares - always read-only but different voucher types
	voucherShares := []models.VoucherShare{
		// Percentage voucher
		{
			VoucherID:    summerVoucher.ID,
			SharedWithID: users[1].ID, // Anna
		},
		{
			VoucherID:    summerVoucher.ID,
			SharedWithID: users[2].ID, // Thomas
		},
		// Fixed amount voucher
		{
			VoucherID:    welcomeVoucher.ID,
			SharedWithID: users[1].ID, // Anna
		},
		// Percentage voucher different usage type
		{
			VoucherID:    techVoucher.ID,
			SharedWithID: users[2].ID, // Thomas
		},
		// Points multiplier voucher
		{
			VoucherID:    doublePointsVoucher.ID,
			SharedWithID: users[3].ID, // Maria
		},
	}

	for _, share := range voucherShares {
		var existing models.VoucherShare
		if err := database.DB.Where("voucher_id = ? AND shared_with_id = ?", share.VoucherID, share.SharedWithID).First(&existing).Error; err == nil {
			log.Printf("  • Voucher share already exists")
		} else {
			if err := database.DB.Create(&share).Error; err != nil {
				log.Printf("  ⚠ Failed to create voucher share: %v", err)
			} else {
				log.Printf("  ✓ Created voucher share (read-only)")
			}
		}
	}

	// Gift Card Shares - all permission combinations
	giftCardShares := []models.GiftCardShare{
		// Full permissions (edit + delete + transactions)
		{
			GiftCardID:          digitecGC.ID,
			SharedWithID:        users[1].ID, // Anna
			CanEdit:             true,
			CanDelete:           true,
			CanEditTransactions: true,
		},
		// View only
		{
			GiftCardID:          digitecGC.ID,
			SharedWithID:        users[2].ID, // Thomas
			CanEdit:             false,
			CanDelete:           false,
			CanEditTransactions: false,
		},
		// Edit + transactions (no delete)
		{
			GiftCardID:          galaxusGC.ID,
			SharedWithID:        users[2].ID, // Thomas
			CanEdit:             true,
			CanDelete:           false,
			CanEditTransactions: true,
		},
		// Transactions only
		{
			GiftCardID:          galaxusGC.ID,
			SharedWithID:        users[3].ID, // Maria
			CanEdit:             false,
			CanDelete:           false,
			CanEditTransactions: true,
		},
	}

	for _, share := range giftCardShares {
		var existing models.GiftCardShare
		if err := database.DB.Where("gift_card_id = ? AND shared_with_id = ?", share.GiftCardID, share.SharedWithID).First(&existing).Error; err == nil {
			log.Printf("  • Gift card share already exists")
		} else {
			if err := database.DB.Create(&share).Error; err != nil {
				log.Printf("  ⚠ Failed to create gift card share: %v", err)
			} else {
				log.Printf("  ✓ Created gift card share (edit: %v, delete: %v, transactions: %v)",
					share.CanEdit, share.CanDelete, share.CanEditTransactions)
			}
		}
	}
}

// createFavorites creates test favorites for all resource types
func createFavorites(users []models.User) {
	log.Println("Creating favorites (all resource types)...")

	// Admin's favorites
	var migrosCard models.Card
	database.DB.Where("card_number = ?", "7610000000001").First(&migrosCard)
	var ikeaCard models.Card
	database.DB.Where("card_number = ?", "IKEA-PDF417-001").First(&ikeaCard)

	var summerVoucher models.Voucher
	database.DB.Where("code = ?", "SUMMER2026").First(&summerVoucher)
	var appleVoucher models.Voucher
	database.DB.Where("code = ?", "APPLE-EUR-20").First(&appleVoucher)

	var galaxusGC models.GiftCard
	database.DB.Where("card_number = ?", "GX-CARD-QR-001").First(&galaxusGC)
	var appleGC models.GiftCard
	database.DB.Where("card_number = ?", "APPLE-EUR-500").First(&appleGC)

	// Anna's resources for her favorites
	var annaCard models.Card
	database.DB.Where("card_number = ?", "DT-2024-ANNA-001").First(&annaCard)
	var annaVoucher models.Voucher
	database.DB.Where("code = ?", "VIP2026").First(&annaVoucher)

	favorites := []models.UserFavorite{
		// Admin favorites: 2 cards, 2 vouchers, 2 gift cards
		{UserID: users[0].ID, ResourceType: "card", ResourceID: migrosCard.ID},
		{UserID: users[0].ID, ResourceType: "card", ResourceID: ikeaCard.ID},
		{UserID: users[0].ID, ResourceType: "voucher", ResourceID: summerVoucher.ID},
		{UserID: users[0].ID, ResourceType: "voucher", ResourceID: appleVoucher.ID},
		{UserID: users[0].ID, ResourceType: "gift_card", ResourceID: galaxusGC.ID},
		{UserID: users[0].ID, ResourceType: "gift_card", ResourceID: appleGC.ID},

		// Anna favorites: 1 card, 1 voucher (+ shared card from admin)
		{UserID: users[1].ID, ResourceType: "card", ResourceID: annaCard.ID},
		{UserID: users[1].ID, ResourceType: "voucher", ResourceID: annaVoucher.ID},
		{UserID: users[1].ID, ResourceType: "card", ResourceID: migrosCard.ID}, // Shared card favorited

		// Thomas favorites: 1 shared voucher
		{UserID: users[2].ID, ResourceType: "voucher", ResourceID: summerVoucher.ID}, // Shared voucher favorited
	}

	for _, fav := range favorites {
		if fav.ResourceID.String() == "00000000-0000-0000-0000-000000000000" {
			continue
		}
		var existing models.UserFavorite
		if err := database.DB.Where("user_id = ? AND resource_type = ? AND resource_id = ?",
			fav.UserID, fav.ResourceType, fav.ResourceID).First(&existing).Error; err == nil {
			log.Printf("  • Favorite already exists (%s)", fav.ResourceType)
		} else {
			if err := database.DB.Create(&fav).Error; err != nil {
				log.Printf("  ⚠ Failed to create favorite: %v", err)
			} else {
				log.Printf("  ✓ Created favorite: %s for user %s", fav.ResourceType, users[findUserIdx(users, fav.UserID)].Email)
			}
		}
	}
}

// createNotifications creates test notifications for all types
func createNotifications(users []models.User) {
	log.Println("Creating notifications (all types, read/unread)...")

	// Get some resource IDs for notifications
	var migrosCard models.Card
	database.DB.Where("card_number = ?", "7610000000001").First(&migrosCard)
	var summerVoucher models.Voucher
	database.DB.Where("code = ?", "SUMMER2026").First(&summerVoucher)
	var galaxusGC models.GiftCard
	database.DB.Where("card_number = ?", "GX-CARD-QR-001").First(&galaxusGC)
	var digitecGC models.GiftCard
	database.DB.Where("card_number = ?", "7610200000002").First(&digitecGC)
	var expiringVoucher models.Voucher
	database.DB.Where("code = ?", "GALAXUS-EXPIRING").First(&expiringVoucher)
	var tomorrowVoucher models.Voucher
	database.DB.Where("code = ?", "HM-TOMORROW").First(&tomorrowVoucher)
	var expiringGC models.GiftCard
	database.DB.Where("card_number = ?", "ID-EXPIRING-SOON").First(&expiringGC)

	now := time.Now()
	readAt := time.Now().Add(-2 * time.Hour)

	notifications := []models.Notification{
		// ===== Admin's notifications (received from others) =====

		// share_received - card (unread)
		{
			UserID:       users[0].ID,
			Type:         models.NotificationTypeShareReceived,
			ResourceType: "card",
			ResourceID:   migrosCard.ID,
			Metadata: models.NotificationMetadata{
				"from_user_id":   users[1].ID.String(),
				"from_user_name": "Anna Müller",
				"merchant_name":  migrosCard.MerchantName,
				"permissions":    map[string]interface{}{"can_edit": true, "can_delete": false},
			},
			IsRead:    false,
			CreatedAt: now.Add(-24 * time.Hour),
		},
		// share_received - voucher (read)
		{
			UserID:       users[0].ID,
			Type:         models.NotificationTypeShareReceived,
			ResourceType: "voucher",
			ResourceID:   summerVoucher.ID,
			Metadata: models.NotificationMetadata{
				"from_user_id":   users[2].ID.String(),
				"from_user_name": "Thomas Schmidt",
				"merchant_name":  summerVoucher.MerchantName,
				"permissions":    map[string]interface{}{"can_edit": false, "can_delete": false},
			},
			IsRead:    true,
			ReadAt:    &readAt,
			CreatedAt: now.Add(-48 * time.Hour),
		},
		// share_received - gift card (unread)
		{
			UserID:       users[0].ID,
			Type:         models.NotificationTypeShareReceived,
			ResourceType: "gift_card",
			ResourceID:   galaxusGC.ID,
			Metadata: models.NotificationMetadata{
				"from_user_id":   users[3].ID.String(),
				"from_user_name": "Maria Garcia",
				"merchant_name":  galaxusGC.MerchantName,
				"permissions":    map[string]interface{}{"can_edit": true, "can_delete": true, "can_edit_transactions": true},
			},
			IsRead:    false,
			CreatedAt: now.Add(-6 * time.Hour),
		},
		// transfer_received - card (unread)
		{
			UserID:       users[0].ID,
			Type:         models.NotificationTypeTransferReceived,
			ResourceType: "card",
			ResourceID:   migrosCard.ID,
			Metadata: models.NotificationMetadata{
				"from_user_id":   users[1].ID.String(),
				"from_user_name": "Anna Müller",
				"merchant_name":  migrosCard.MerchantName,
			},
			IsRead:    false,
			CreatedAt: now.Add(-12 * time.Hour),
		},
		// transfer_received - gift card (read)
		{
			UserID:       users[0].ID,
			Type:         models.NotificationTypeTransferReceived,
			ResourceType: "gift_card",
			ResourceID:   digitecGC.ID,
			Metadata: models.NotificationMetadata{
				"from_user_id":   users[2].ID.String(),
				"from_user_name": "Thomas Schmidt",
				"merchant_name":  digitecGC.MerchantName,
			},
			IsRead:    true,
			ReadAt:    &readAt,
			CreatedAt: now.Add(-72 * time.Hour),
		},
		// expiry_reminder - voucher 7 days (unread)
		{
			UserID:       users[0].ID,
			Type:         models.NotificationTypeExpiryReminder,
			ResourceType: "voucher",
			ResourceID:   expiringVoucher.ID,
			Metadata: models.NotificationMetadata{
				"days_before":   float64(7),
				"days_left":     float64(7),
				"merchant_name": expiringVoucher.MerchantName,
				"resource_name": "GALAXUS-EXPIRING",
				"expires_at":    time.Now().AddDate(0, 0, 3).Format(time.RFC3339),
			},
			IsRead:    false,
			CreatedAt: now.Add(-4 * 24 * time.Hour),
		},
		// expiry_reminder - voucher 3 days (unread)
		{
			UserID:       users[0].ID,
			Type:         models.NotificationTypeExpiryReminder,
			ResourceType: "voucher",
			ResourceID:   expiringVoucher.ID,
			Metadata: models.NotificationMetadata{
				"days_before":   float64(3),
				"days_left":     float64(3),
				"merchant_name": expiringVoucher.MerchantName,
				"resource_name": "GALAXUS-EXPIRING",
				"expires_at":    time.Now().AddDate(0, 0, 3).Format(time.RFC3339),
			},
			IsRead:    false,
			CreatedAt: now.Add(-1 * time.Hour),
		},
		// expiry_reminder - gift card 7 days (read)
		{
			UserID:       users[0].ID,
			Type:         models.NotificationTypeExpiryReminder,
			ResourceType: "gift_card",
			ResourceID:   expiringGC.ID,
			Metadata: models.NotificationMetadata{
				"days_before":   float64(7),
				"days_left":     float64(7),
				"merchant_name": expiringGC.MerchantName,
				"resource_name": "ID-EXPIRING-SOON",
				"expires_at":    time.Now().AddDate(0, 0, 5).Format(time.RFC3339),
			},
			IsRead:    true,
			ReadAt:    &readAt,
			CreatedAt: now.Add(-2 * 24 * time.Hour),
		},
		// validity_start - voucher becomes valid tomorrow (unread)
		{
			UserID:       users[0].ID,
			Type:         models.NotificationTypeValidityStart,
			ResourceType: "voucher",
			ResourceID:   tomorrowVoucher.ID,
			Metadata: models.NotificationMetadata{
				"merchant_name": tomorrowVoucher.MerchantName,
				"valid_from":    time.Now().AddDate(0, 0, 1).Format(time.RFC3339),
				"code":          "HM-TOMORROW",
			},
			IsRead:    false,
			CreatedAt: now.Add(-30 * time.Minute),
		},

		// ===== Anna's notifications =====
		// share_received - card from admin (unread)
		{
			UserID:       users[1].ID,
			Type:         models.NotificationTypeShareReceived,
			ResourceType: "card",
			ResourceID:   migrosCard.ID,
			Metadata: models.NotificationMetadata{
				"from_user_id":   users[0].ID.String(),
				"from_user_name": "Admin User",
				"merchant_name":  migrosCard.MerchantName,
				"permissions":    map[string]interface{}{"can_edit": true, "can_delete": true},
			},
			IsRead:    false,
			CreatedAt: now.Add(-3 * time.Hour),
		},

		// ===== Thomas's notifications =====
		// transfer_received - voucher (read)
		{
			UserID:       users[2].ID,
			Type:         models.NotificationTypeTransferReceived,
			ResourceType: "voucher",
			ResourceID:   summerVoucher.ID,
			Metadata: models.NotificationMetadata{
				"from_user_id":   users[0].ID.String(),
				"from_user_name": "Admin User",
				"merchant_name":  summerVoucher.MerchantName,
			},
			IsRead:    true,
			ReadAt:    &readAt,
			CreatedAt: now.Add(-5 * 24 * time.Hour),
		},
	}

	for _, notif := range notifications {
		if notif.ResourceID.String() == "00000000-0000-0000-0000-000000000000" {
			continue
		}
		var existing models.Notification
		if err := database.DB.Where("user_id = ? AND type = ? AND resource_type = ? AND resource_id = ?",
			notif.UserID, notif.Type, notif.ResourceType, notif.ResourceID).First(&existing).Error; err == nil {
			log.Printf("  • Notification already exists (%s %s)", notif.Type, notif.ResourceType)
		} else {
			if err := database.DB.Create(&notif).Error; err != nil {
				log.Printf("  ⚠ Failed to create notification: %v", err)
			} else {
				readStatus := "unread"
				if notif.IsRead {
					readStatus = "read"
				}
				log.Printf("  ✓ Created notification: %s %s (%s)", notif.Type, notif.ResourceType, readStatus)
			}
		}
	}
}

// createAuditLogs creates sample audit log entries
func createAuditLogs(users []models.User) {
	log.Println("Creating audit logs...")

	// Simulate deletion audit entries
	cardData, _ := json.Marshal(map[string]interface{}{
		"merchant_name": "Alte Firma AG",
		"program":       "Kundenkarte",
		"card_number":   "DELETED-CARD-001",
		"barcode_type":  "CODE128",
		"status":        "inactive",
	})
	voucherData, _ := json.Marshal(map[string]interface{}{
		"merchant_name": "Shop XYZ",
		"code":          "DELETED-VOUCHER-001",
		"type":          "percentage",
		"value":         15.0,
	})
	giftCardData, _ := json.Marshal(map[string]interface{}{
		"merchant_name":   "Geschenkhaus",
		"card_number":     "DELETED-GC-001",
		"initial_balance": 50.0,
		"currency":        "CHF",
	})

	fakeResourceID1 := uuid.New()
	fakeResourceID2 := uuid.New()
	fakeResourceID3 := uuid.New()

	auditLogs := []models.AuditLog{
		// Card deletion by admin
		{
			UserID:       &users[0].ID,
			Action:       "delete",
			ResourceType: "card",
			ResourceID:   fakeResourceID1,
			ResourceData: string(cardData),
			IPAddress:    "192.168.1.100",
			UserAgent:    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) Chrome/120.0.0.0",
			CreatedAt:    time.Now().AddDate(0, 0, -14),
		},
		// Voucher deletion by Anna
		{
			UserID:       &users[1].ID,
			Action:       "delete",
			ResourceType: "voucher",
			ResourceID:   fakeResourceID2,
			ResourceData: string(voucherData),
			IPAddress:    "10.0.0.50",
			UserAgent:    "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0) Safari/604.1",
			CreatedAt:    time.Now().AddDate(0, 0, -7),
		},
		// Gift card deletion by Thomas
		{
			UserID:       &users[2].ID,
			Action:       "delete",
			ResourceType: "gift_card",
			ResourceID:   fakeResourceID3,
			ResourceData: string(giftCardData),
			IPAddress:    "172.16.0.25",
			UserAgent:    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Firefox/121.0",
			CreatedAt:    time.Now().AddDate(0, 0, -3),
		},
	}

	for _, al := range auditLogs {
		var count int64
		database.DB.Model(&models.AuditLog{}).Where("resource_id = ? AND action = ?", al.ResourceID, al.Action).Count(&count)
		if count > 0 {
			log.Printf("  • Audit log already exists (%s %s)", al.Action, al.ResourceType)
		} else {
			if err := database.DB.Create(&al).Error; err != nil {
				log.Printf("  ⚠ Failed to create audit log: %v", err)
			} else {
				log.Printf("  ✓ Created audit log: %s %s", al.Action, al.ResourceType)
			}
		}
	}
}

// createExpiryReminders creates sample expiry reminder tracking records
func createExpiryReminders(users []models.User) {
	log.Println("Creating expiry reminder tracking records...")

	var expiringVoucher models.Voucher
	database.DB.Where("code = ?", "GALAXUS-EXPIRING").First(&expiringVoucher)
	var expiringGC models.GiftCard
	database.DB.Where("card_number = ?", "ID-EXPIRING-SOON").First(&expiringGC)

	if expiringVoucher.ID.String() == "00000000-0000-0000-0000-000000000000" {
		log.Println("  ⚠ Skipping expiry reminders - resources not found")
		return
	}

	reminders := []models.ExpiryReminderSent{
		// Voucher: 7-day reminder already sent
		{
			UserID:       users[0].ID,
			ResourceType: "voucher",
			ResourceID:   expiringVoucher.ID,
			DaysBefore:   7,
			SentAt:       time.Now().AddDate(0, 0, -4),
		},
		// Voucher: 3-day reminder already sent
		{
			UserID:       users[0].ID,
			ResourceType: "voucher",
			ResourceID:   expiringVoucher.ID,
			DaysBefore:   3,
			SentAt:       time.Now().Add(-1 * time.Hour),
		},
		// Gift card: 7-day reminder already sent
		{
			UserID:       users[0].ID,
			ResourceType: "gift_card",
			ResourceID:   expiringGC.ID,
			DaysBefore:   7,
			SentAt:       time.Now().AddDate(0, 0, -2),
		},
	}

	for _, r := range reminders {
		var count int64
		database.DB.Model(&models.ExpiryReminderSent{}).Where(
			"user_id = ? AND resource_type = ? AND resource_id = ? AND days_before = ?",
			r.UserID, r.ResourceType, r.ResourceID, r.DaysBefore,
		).Count(&count)
		if count > 0 {
			log.Printf("  • Expiry reminder already exists (%s, %d days)", r.ResourceType, r.DaysBefore)
		} else {
			if err := database.DB.Create(&r).Error; err != nil {
				log.Printf("  ⚠ Failed to create expiry reminder: %v", err)
			} else {
				log.Printf("  ✓ Created expiry reminder: %s %d-day", r.ResourceType, r.DaysBefore)
			}
		}
	}
}

// findUserIdx returns the index of a user in the slice by ID
func findUserIdx(users []models.User, userID uuid.UUID) int {
	for i, u := range users {
		if u.ID == userID {
			return i
		}
	}
	return 0
}

// printSummary prints the final summary of created data
func printSummary() {
	log.Println()
	log.Println("✓ Comprehensive database seeding completed!")
	log.Println()
	log.Println("📧 Test credentials:")
	log.Println("  • admin@example.com / test123 (👑 Admin, DE, email verified, all notifications)")
	log.Println("  • anna.mueller@example.com / test123 (DE, email verified, push only)")
	log.Println("  • thomas.schmidt@example.com / test123 (EN, email verified, email reminders only)")
	log.Println("  • maria.garcia@example.com / test123 (FR, email NOT verified, all notifications OFF)")
	log.Println()
	log.Println("📊 Data Summary:")
	log.Println()
	log.Println("  🏪 Merchants: 12 total")
	log.Println("    • With website: 9 (Migros, Coop, Manor, Media Markt, Digitec, Galaxus, Interdiscount, Denner, Apple Store, IKEA)")
	log.Println("    • Without website: 2 (H&M, Starbucks)")
	log.Println()
	log.Println("  📇 Cards: 19 total")
	log.Println("    • Barcode Types: CODE128 (7), QR (4), EAN13 (1), EAN8 (1), PDF417 (1),")
	log.Println("                     DATAMATRIX (1), CODE39 (1), AZTEC (1), UPCA (1), ITF (1)")
	log.Println("    • Statuses: active (16), inactive (1), expired (1)")
	log.Println("    • Special: 1 without notes, 1 without merchant reference")
	log.Println()
	log.Println("  🎟️  Vouchers: 19 total")
	log.Println("    • Types: percentage (6), fixed_amount (10), points_multiplier (3)")
	log.Println("    • Currencies: CHF (6), EUR (2), USD (1), GBP (1)")
	log.Println("    • Statuses: active/valid (13), expired (1), future-valid (2), expiring soon (1)")
	log.Println("    • Usage Limits: single_use (7), one_per_customer (3), multiple_use_with_card (3),")
	log.Println("                    multiple_use_without_card (4)")
	log.Println("    • Special: 2 without description, 1 without merchant ref, 1 becomes valid tomorrow")
	log.Println()
	log.Println("  🎁 Gift Cards: 14 total")
	log.Println("    • Currencies: CHF (9), EUR (1), USD (1), GBP (1)")
	log.Println("    • With PIN: 8, Without PIN: 6")
	log.Println("    • Barcode Types: CODE128 (7), QR (3), EAN13 (1), AZTEC (1), DATAMATRIX (1), PDF417 (1)")
	log.Println("    • Statuses: active (12), inactive/expired (1), redeemed/empty (1)")
	log.Println("    • Special: 1 without notes, 1 without merchant ref, 3 no expiry, 1 expiring soon")
	log.Println("    • Transactions: 5 cards with transactions (incl. reload/negative amount)")
	log.Println()
	log.Println("  🤝 Shares:")
	log.Println("    • Card Shares: 5 (all permission combos: view-only, edit, delete, full)")
	log.Println("    • Voucher Shares: 5 (all read-only)")
	log.Println("    • Gift Card Shares: 4 (all permission combos incl. transactions)")
	log.Println()
	log.Println("  ⭐ Favorites: 10 total")
	log.Println("    • Admin: 6 (2 cards, 2 vouchers, 2 gift cards)")
	log.Println("    • Anna: 3 (2 cards incl. shared, 1 voucher)")
	log.Println("    • Thomas: 1 (shared voucher)")
	log.Println("    • Maria: 0 (no favorites)")
	log.Println()
	log.Println("  🔔 Notifications: 12 total")
	log.Println("    • share_received: 4 (card, voucher, gift card - read/unread)")
	log.Println("    • transfer_received: 3 (card, gift card, voucher - read/unread)")
	log.Println("    • expiry_reminder: 3 (7-day, 3-day for voucher, 7-day for gift card)")
	log.Println("    • validity_start: 1 (voucher becomes valid tomorrow)")
	log.Println("    • Admin: 9 notifications (5 unread, 4 read)")
	log.Println()
	log.Println("  📋 Audit Logs: 3 total")
	log.Println("    • Card deletion, Voucher deletion, Gift card deletion")
	log.Println()
	log.Println("  ⏰ Expiry Reminders: 3 tracking records")
	log.Println("    • Voucher: 7-day + 3-day sent")
	log.Println("    • Gift Card: 7-day sent")
	log.Println()
	log.Println("  👤 User Settings:")
	log.Println("    • Languages: DE (2), EN (1), FR (1)")
	log.Println("    • Email Verified: 3 of 4")
	log.Println("    • Notification Preferences: all on, push only, email reminders only, all off")
}

func main() {
	log.Println("🌱 Seeding database with comprehensive test data...")

	// Load config
	cfg := config.Load()

	// Connect to database
	if err := database.Connect(cfg.DatabaseURL, cfg.LogLevel); err != nil {
		log.Fatal(err)
	}

	// Hash password "test123" for all test users with cost 12 for enhanced security
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("test123"), 12)
	if err != nil {
		log.Fatal(err)
	}

	// Create all test data using helper functions
	users := createUsers(string(hashedPassword))
	merchants := createMerchants()
	createCards(users, merchants)
	createVouchers(users, merchants)
	createGiftCards(users, merchants)
	createShares(users)
	createFavorites(users)
	createNotifications(users)
	createAuditLogs(users)
	createExpiryReminders(users)
	printSummary()
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
