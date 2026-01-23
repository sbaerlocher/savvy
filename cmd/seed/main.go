// Package main seeds the database with sample data for testing.
package main

import (
	"log"
	"savvy/internal/config"
	"savvy/internal/database"
	"savvy/internal/models"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	log.Println("🌱 Seeding database with comprehensive test data...")

	// Load config
	cfg := config.Load()

	// Connect to database
	if err := database.Connect(cfg.DatabaseURL); err != nil {
		log.Fatal(err)
	}

	// Hash password "test123" for all test users
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("test123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}

	// Create test users
	users := []models.User{
		{
			Email:        "admin@example.com",
			PasswordHash: string(hashedPassword),
			FirstName:    "Admin",
			LastName:     "User",
			Role:         "admin",
		},
		{
			Email:        "anna.mueller@example.com",
			PasswordHash: string(hashedPassword),
			FirstName:    "Anna",
			LastName:     "Müller",
			Role:         "user",
		},
		{
			Email:        "thomas.schmidt@example.com",
			PasswordHash: string(hashedPassword),
			FirstName:    "Thomas",
			LastName:     "Schmidt",
			Role:         "user",
		},
		{
			Email:        "maria.garcia@example.com",
			PasswordHash: string(hashedPassword),
			FirstName:    "Maria",
			LastName:     "Garcia",
			Role:         "user",
		},
	}

	log.Println("Creating users...")
	for i := range users {
		var existing models.User
		if err := database.DB.Where("email = ?", users[i].Email).First(&existing).Error; err == nil {
			log.Printf("  • User already exists: %s", users[i].Email)
			users[i] = existing // Use existing user
		} else {
			if err := database.DB.Create(&users[i]).Error; err != nil {
				log.Fatal(err)
			}
			log.Printf("  ✓ Created user: %s (%s %s)", users[i].Email, users[i].FirstName, users[i].LastName)
		}
	}

	// Create test merchants
	log.Println("Creating merchants...")
	merchants := []models.Merchant{
		{Name: "Migros", LogoURL: "", Website: "https://www.migros.ch", Color: "#FF6B35"},
		{Name: "Coop", LogoURL: "", Website: "https://www.coop.ch", Color: "#F7931E"},
		{Name: "Manor", LogoURL: "", Website: "https://www.manor.ch", Color: "#C41E3A"},
		{Name: "Media Markt", LogoURL: "", Website: "https://www.mediamarkt.ch", Color: "#DC2626"},
		{Name: "Digitec", LogoURL: "", Website: "https://www.digitec.ch", Color: "#0066CC"},
		{Name: "Galaxus", LogoURL: "", Website: "https://www.galaxus.ch", Color: "#7C3AED"},
		{Name: "Interdiscount", LogoURL: "", Website: "https://www.interdiscount.ch", Color: "#059669"},
		{Name: "Denner", LogoURL: "", Website: "https://www.denner.ch", Color: "#DC2626"},
	}

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

	// Create comprehensive test cards covering all barcode types and statuses
	log.Println("Creating cards (all barcode types & statuses)...")
	cards := []models.Card{
		// Admin's cards - CODE128
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
		// Admin's cards - EAN13
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
		// Admin's cards - EAN8
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
		// Admin's cards - QR
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
		// Admin's card - inactive status
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
		// Anna's cards
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
		// Thomas's cards
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
		// Maria's card
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

	// Create comprehensive test vouchers covering all types and usage limits
	log.Println("Creating vouchers (all types & usage limits)...")
	vouchers := []models.Voucher{
		// Admin's vouchers - percentage type
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
			UsedCount:         2,
			BarcodeType:       "QR",
		},
		// Admin's vouchers - fixed_amount, single_use
		{
			UserID:            &users[0].ID,
			MerchantID:        &merchants[1].ID,
			MerchantName:      merchants[1].Name,
			Code:              "WELCOME50",
			Type:              "fixed_amount",
			Value:             50.0,
			Description:       "50 CHF Willkommensbonus",
			MinPurchaseAmount: 100.0,
			ValidFrom:         time.Now(),
			ValidUntil:        time.Now().AddDate(0, 1, 0),
			UsageLimitType:    "single_use",
			UsedCount:         0,
			BarcodeType:       "CODE128",
		},
		// Admin's vouchers - percentage, one_per_customer
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
			UsedCount:         0,
			BarcodeType:       "CODE128",
		},
		// Admin's vouchers - points_multiplier, multiple_use_without_card
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
			UsedCount:         5,
			BarcodeType:       "QR",
		},
		// Admin's vouchers - unlimited usage
		{
			UserID:            &users[0].ID,
			MerchantID:        &merchants[2].ID,
			MerchantName:      merchants[2].Name,
			Code:              "MANOR-FOREVER",
			Type:              "fixed_amount",
			Value:             10.0,
			Description:       "10 CHF Dauerrabatt",
			MinPurchaseAmount: 30.0,
			ValidFrom:         time.Now(),
			ValidUntil:        time.Now().AddDate(5, 0, 0),
			UsageLimitType:    "unlimited",
			UsedCount:         0,
			BarcodeType:       "CODE128",
		},
		// Admin's vouchers - expired
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
			UsedCount:         0,
			BarcodeType:       "CODE128",
		},
		// Anna's vouchers
		{
			UserID:            &users[1].ID,
			MerchantID:        &merchants[2].ID,
			MerchantName:      merchants[2].Name,
			Code:              "VIP2026",
			Type:              "fixed_amount",
			Value:             25.0,
			Description:       "VIP Bonus 25 CHF",
			MinPurchaseAmount: 75.0,
			ValidFrom:         time.Now(),
			ValidUntil:        time.Now().AddDate(0, 6, 0),
			UsageLimitType:    "one_per_customer",
			UsedCount:         0,
			BarcodeType:       "QR",
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
			UsedCount:         1,
			BarcodeType:       "QR",
		},
		// Thomas's vouchers
		{
			UserID:            &users[2].ID,
			MerchantID:        &merchants[6].ID,
			MerchantName:      merchants[6].Name,
			Code:              "ID-SAVE-100",
			Type:              "fixed_amount",
			Value:             100.0,
			Description:       "100 CHF Mega-Rabatt",
			MinPurchaseAmount: 500.0,
			ValidFrom:         time.Now(),
			ValidUntil:        time.Now().AddDate(0, 1, 0),
			UsageLimitType:    "single_use",
			UsedCount:         0,
			BarcodeType:       "CODE128",
		},
	}

	for _, voucher := range vouchers {
		var existing models.Voucher
		if err := database.DB.Where("code = ?", voucher.Code).First(&existing).Error; err == nil {
			log.Printf("  • Voucher already exists: %s", voucher.Code)
		} else {
			if err := database.DB.Create(&voucher).Error; err != nil {
				log.Fatal(err)
			}
			log.Printf("  ✓ Created voucher: %s (%s, %s, %s)", voucher.Code, voucher.Type, voucher.UsageLimitType, voucher.BarcodeType)
		}
	}

	// Create comprehensive test gift cards
	log.Println("Creating gift cards (with & without transactions, different statuses)...")
	giftCards := []models.GiftCard{
		// Admin's gift cards - with transactions
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
		// Admin's gift cards - with PIN, EAN13
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
		// Admin's gift cards - without PIN, QR
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
		// Admin's gift cards - expired
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
		// Admin's gift cards - fully used (balance 0)
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
		// Anna's gift cards
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
		// Thomas's gift cards
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
		// Maria's gift cards - no expiry
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

	for _, giftCard := range giftCards {
		var existing models.GiftCard
		if err := database.DB.Where("card_number = ?", giftCard.CardNumber).First(&existing).Error; err == nil {
			log.Printf("  • Gift card already exists: %s", giftCard.CardNumber)
		} else {
			if err := database.DB.Create(&giftCard).Error; err != nil {
				log.Fatal(err)
			}
			log.Printf("  ✓ Created gift card: %s - %s (%s)", giftCard.MerchantName, giftCard.CardNumber, giftCard.BarcodeType)
		}
	}

	// Add transactions to gift cards
	log.Println("Creating gift card transactions...")

	// Media Markt card - multiple transactions
	var mediaMarktGC models.GiftCard
	database.DB.Where("card_number = ?", "MM1234567890").First(&mediaMarktGC)
	if mediaMarktGC.ID.String() != "00000000-0000-0000-0000-000000000000" {
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
		for _, t := range transactions1 {
			var existing models.GiftCardTransaction
			if err := database.DB.Where("gift_card_id = ? AND description = ?", t.GiftCardID, t.Description).First(&existing).Error; err == nil {
				log.Printf("  • Transaction already exists: %s", t.Description)
			} else {
				database.DB.Create(&t)
				log.Printf("  ✓ Created transaction: %s (%.2f CHF)", t.Description, t.Amount)
			}
		}
	}

	// Coop fully used card - transactions that sum to initial balance
	var coopUsedGC models.GiftCard
	database.DB.Where("card_number = ?", "COOP-USED-777").First(&coopUsedGC)
	if coopUsedGC.ID.String() != "00000000-0000-0000-0000-000000000000" {
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
		for _, t := range transactions2 {
			var existing models.GiftCardTransaction
			if err := database.DB.Where("gift_card_id = ? AND description = ?", t.GiftCardID, t.Description).First(&existing).Error; err == nil {
				log.Printf("  • Transaction already exists: %s", t.Description)
			} else {
				database.DB.Create(&t)
				log.Printf("  ✓ Created transaction: %s (%.2f CHF)", t.Description, t.Amount)
			}
		}
	}

	// Create comprehensive shares covering all permission combinations
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

	log.Println()
	log.Println("✓ Comprehensive database seeding completed!")
	log.Println()
	log.Println("📧 Test credentials:")
	log.Println("  • admin@example.com / test123 (👑 Admin)")
	log.Println("  • anna.mueller@example.com / test123")
	log.Println("  • thomas.schmidt@example.com / test123")
	log.Println("  • maria.garcia@example.com / test123")
	log.Println()
	log.Println("📊 Data Summary:")
	log.Println("  📇 Cards: 9 total")
	log.Println("    • Barcode Types: CODE128 (5), EAN13 (2), EAN8 (1), QR (2)")
	log.Println("    • Statuses: active (8), inactive (1)")
	log.Println()
	log.Println("  🎟️  Vouchers: 9 total")
	log.Println("    • Types: percentage (4), fixed_amount (4), points_multiplier (2)")
	log.Println("    • Usage Limits: single_use (2), one_per_customer (2), multiple_use_with_card (2),")
	log.Println("                    multiple_use_without_card (2), unlimited (1)")
	log.Println("    • Includes: 1 expired voucher (EXPIRED2025)")
	log.Println()
	log.Println("  🎁 Gift Cards: 8 total")
	log.Println("    • With PIN: 6, Without PIN: 2")
	log.Println("    • Barcode Types: CODE128 (5), EAN13 (1), QR (2)")
	log.Println("    • Statuses: active (7), inactive/expired (1)")
	log.Println("    • With Transactions: 2 cards")
	log.Println()
	log.Println("  🤝 Shares:")
	log.Println("    • Card Shares: 5 (covering all permission combos)")
	log.Println("    • Voucher Shares: 5 (all read-only)")
	log.Println("    • Gift Card Shares: 4 (covering all permission combos)")
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
