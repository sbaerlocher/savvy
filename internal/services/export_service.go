// Package services contains business logic.
package services

import (
	"context"
	"time"

	"savvy/internal/repository"

	"github.com/google/uuid"
)

// ExportData contains all user data for export.
type ExportData struct {
	ExportedAt string           `json:"exported_at"`
	User       ExportUser       `json:"user"`
	Cards      []ExportCard     `json:"cards"`
	Vouchers   []ExportVoucher  `json:"vouchers"`
	GiftCards  []ExportGiftCard `json:"gift_cards"`
	Favorites  []ExportFavorite `json:"favorites"`
}

// ExportUser contains user profile data.
type ExportUser struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	AuthProvider  string `json:"auth_provider"`
	EmailVerified bool   `json:"email_verified"`
	CreatedAt     string `json:"created_at"`
}

// ExportCard contains card data.
type ExportCard struct {
	ID           string `json:"id"`
	MerchantName string `json:"merchant_name"`
	Program      string `json:"program"`
	CardNumber   string `json:"card_number"`
	BarcodeType  string `json:"barcode_type"`
	Status       string `json:"status"`
	Notes        string `json:"notes"`
	CreatedAt    string `json:"created_at"`
}

// ExportVoucher contains voucher data.
type ExportVoucher struct {
	ID                string  `json:"id"`
	MerchantName      string  `json:"merchant_name"`
	Code              string  `json:"code"`
	Type              string  `json:"type"`
	Value             float64 `json:"value"`
	Description       string  `json:"description"`
	MinPurchaseAmount float64 `json:"min_purchase_amount"`
	ValidFrom         string  `json:"valid_from"`
	ValidUntil        string  `json:"valid_until"`
	UsageLimitType    string  `json:"usage_limit_type"`
	BarcodeType       string  `json:"barcode_type"`
	CreatedAt         string  `json:"created_at"`
}

// ExportGiftCard contains gift card data with transactions.
type ExportGiftCard struct {
	ID             string                      `json:"id"`
	MerchantName   string                      `json:"merchant_name"`
	CardNumber     string                      `json:"card_number"`
	InitialBalance float64                     `json:"initial_balance"`
	CurrentBalance float64                     `json:"current_balance"`
	Currency       string                      `json:"currency"`
	PIN            string                      `json:"pin"`
	ExpiresAt      string                      `json:"expires_at,omitempty"`
	Status         string                      `json:"status"`
	BarcodeType    string                      `json:"barcode_type"`
	Notes          string                      `json:"notes"`
	CreatedAt      string                      `json:"created_at"`
	Transactions   []ExportGiftCardTransaction `json:"transactions"`
}

// ExportGiftCardTransaction contains transaction data.
type ExportGiftCardTransaction struct {
	ID              string  `json:"id"`
	Amount          float64 `json:"amount"`
	Description     string  `json:"description"`
	TransactionDate string  `json:"transaction_date"`
	CreatedAt       string  `json:"created_at"`
}

// ExportFavorite contains favorite data.
type ExportFavorite struct {
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
}

// BatchExportData contains only the exported resources (no user data, no empty arrays).
type BatchExportData struct {
	ExportedAt string           `json:"exported_at"`
	Cards      []ExportCard     `json:"cards,omitempty"`
	Vouchers   []ExportVoucher  `json:"vouchers,omitempty"`
	GiftCards  []ExportGiftCard `json:"gift_cards,omitempty"`
}

// ExportServiceInterface defines export operations.
type ExportServiceInterface interface {
	ExportUserData(ctx context.Context, userID uuid.UUID) (*ExportData, error)
	ExportCardsByIDs(ctx context.Context, userID uuid.UUID, ids []uuid.UUID) (*BatchExportData, error)
	ExportVouchersByIDs(ctx context.Context, userID uuid.UUID, ids []uuid.UUID) (*BatchExportData, error)
	ExportGiftCardsByIDs(ctx context.Context, userID uuid.UUID, ids []uuid.UUID) (*BatchExportData, error)
}

// ExportService implements ExportServiceInterface.
type ExportService struct {
	userService  UserServiceInterface
	cardRepo     repository.CardRepository
	voucherRepo  repository.VoucherRepository
	giftCardRepo repository.GiftCardRepository
	favoriteRepo repository.FavoriteRepository
}

// NewExportService creates a new export service.
func NewExportService(
	userService UserServiceInterface,
	cardRepo repository.CardRepository,
	voucherRepo repository.VoucherRepository,
	giftCardRepo repository.GiftCardRepository,
	favoriteRepo repository.FavoriteRepository,
) ExportServiceInterface {
	return &ExportService{
		userService:  userService,
		cardRepo:     cardRepo,
		voucherRepo:  voucherRepo,
		giftCardRepo: giftCardRepo,
		favoriteRepo: favoriteRepo,
	}
}

// ExportUserData collects all user-owned data for export.
func (s *ExportService) ExportUserData(ctx context.Context, userID uuid.UUID) (*ExportData, error) {
	// Get user profile
	user, err := s.userService.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get own cards (not shared)
	cards, err := s.cardRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get own vouchers (not shared)
	vouchers, err := s.voucherRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get own gift cards with transactions (not shared)
	giftCards, err := s.giftCardRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get favorites
	favorites, err := s.favoriteRepo.GetByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Build export data
	export := &ExportData{
		ExportedAt: time.Now().UTC().Format(time.RFC3339),
		User: ExportUser{
			ID:            user.ID.String(),
			Email:         user.Email,
			FirstName:     user.FirstName,
			LastName:      user.LastName,
			AuthProvider:  user.AuthProvider,
			EmailVerified: user.EmailVerified,
			CreatedAt:     user.CreatedAt.Format(time.RFC3339),
		},
	}

	// Map cards
	export.Cards = make([]ExportCard, len(cards))
	for i, c := range cards {
		merchantName := c.MerchantName
		if c.Merchant != nil && c.Merchant.Name != "" {
			merchantName = c.Merchant.Name
		}
		export.Cards[i] = ExportCard{
			ID:           c.ID.String(),
			MerchantName: merchantName,
			Program:      c.Program,
			CardNumber:   c.CardNumber,
			BarcodeType:  c.BarcodeType,
			Status:       c.Status,
			Notes:        c.Notes,
			CreatedAt:    c.CreatedAt.Format(time.RFC3339),
		}
	}

	// Map vouchers
	export.Vouchers = make([]ExportVoucher, len(vouchers))
	for i, v := range vouchers {
		merchantName := v.MerchantName
		if v.Merchant != nil && v.Merchant.Name != "" {
			merchantName = v.Merchant.Name
		}
		export.Vouchers[i] = ExportVoucher{
			ID:                v.ID.String(),
			MerchantName:      merchantName,
			Code:              v.Code,
			Type:              v.Type,
			Value:             v.Value,
			Description:       v.Description,
			MinPurchaseAmount: v.MinPurchaseAmount,
			ValidFrom:         v.ValidFrom.Format(time.RFC3339),
			ValidUntil:        v.ValidUntil.Format(time.RFC3339),
			UsageLimitType:    v.UsageLimitType,
			BarcodeType:       v.BarcodeType,
			CreatedAt:         v.CreatedAt.Format(time.RFC3339),
		}
	}

	// Map gift cards with transactions
	export.GiftCards = make([]ExportGiftCard, len(giftCards))
	for i, gc := range giftCards {
		merchantName := gc.MerchantName
		if gc.Merchant != nil && gc.Merchant.Name != "" {
			merchantName = gc.Merchant.Name
		}

		exportGC := ExportGiftCard{
			ID:             gc.ID.String(),
			MerchantName:   merchantName,
			CardNumber:     gc.CardNumber,
			InitialBalance: gc.InitialBalance,
			CurrentBalance: gc.GetCurrentBalance(),
			Currency:       gc.Currency,
			PIN:            gc.PIN,
			Status:         gc.GetComputedStatus(),
			BarcodeType:    gc.BarcodeType,
			Notes:          gc.Notes,
			CreatedAt:      gc.CreatedAt.Format(time.RFC3339),
		}

		if gc.ExpiresAt != nil {
			exportGC.ExpiresAt = gc.ExpiresAt.Format(time.RFC3339)
		}

		// Map transactions
		exportGC.Transactions = make([]ExportGiftCardTransaction, len(gc.Transactions))
		for j, tx := range gc.Transactions {
			exportGC.Transactions[j] = ExportGiftCardTransaction{
				ID:              tx.ID.String(),
				Amount:          tx.Amount,
				Description:     tx.Description,
				TransactionDate: tx.TransactionDate.Format(time.RFC3339),
				CreatedAt:       tx.CreatedAt.Format(time.RFC3339),
			}
		}

		export.GiftCards[i] = exportGC
	}

	// Map favorites
	export.Favorites = make([]ExportFavorite, len(favorites))
	for i, f := range favorites {
		export.Favorites[i] = ExportFavorite{
			ResourceType: f.ResourceType,
			ResourceID:   f.ResourceID.String(),
		}
	}

	return export, nil
}

// ExportCardsByIDs exports selected cards owned by or shared with the user.
func (s *ExportService) ExportCardsByIDs(ctx context.Context, userID uuid.UUID, ids []uuid.UUID) (*BatchExportData, error) {
	export := &BatchExportData{
		ExportedAt: time.Now().UTC().Format(time.RFC3339),
		Cards:      make([]ExportCard, 0, len(ids)),
	}

	// Get user's own and shared card IDs for ownership verification
	ownCards, _ := s.cardRepo.GetByUserID(ctx, userID)
	ownCardIDs := make(map[uuid.UUID]bool, len(ownCards))
	for _, c := range ownCards {
		ownCardIDs[c.ID] = true
	}

	for _, id := range ids {
		if !ownCardIDs[id] {
			continue // skip cards not owned by user
		}
		c, err := s.cardRepo.GetByID(ctx, id, "Merchant")
		if err != nil {
			continue
		}
		merchantName := c.MerchantName
		if c.Merchant != nil && c.Merchant.Name != "" {
			merchantName = c.Merchant.Name
		}
		export.Cards = append(export.Cards, ExportCard{
			ID:           c.ID.String(),
			MerchantName: merchantName,
			Program:      c.Program,
			CardNumber:   c.CardNumber,
			BarcodeType:  c.BarcodeType,
			Status:       c.Status,
			Notes:        c.Notes,
			CreatedAt:    c.CreatedAt.Format(time.RFC3339),
		})
	}

	return export, nil
}

// ExportVouchersByIDs exports selected vouchers owned by the user.
func (s *ExportService) ExportVouchersByIDs(ctx context.Context, userID uuid.UUID, ids []uuid.UUID) (*BatchExportData, error) {
	export := &BatchExportData{
		ExportedAt: time.Now().UTC().Format(time.RFC3339),
		Vouchers:   make([]ExportVoucher, 0, len(ids)),
	}

	// Get user's own voucher IDs for ownership verification
	ownVouchers, _ := s.voucherRepo.GetByUserID(ctx, userID)
	ownVoucherIDs := make(map[uuid.UUID]bool, len(ownVouchers))
	for _, v := range ownVouchers {
		ownVoucherIDs[v.ID] = true
	}

	for _, id := range ids {
		if !ownVoucherIDs[id] {
			continue // skip vouchers not owned by user
		}
		v, err := s.voucherRepo.GetByID(ctx, id, "Merchant")
		if err != nil {
			continue
		}
		merchantName := v.MerchantName
		if v.Merchant != nil && v.Merchant.Name != "" {
			merchantName = v.Merchant.Name
		}
		export.Vouchers = append(export.Vouchers, ExportVoucher{
			ID:                v.ID.String(),
			MerchantName:      merchantName,
			Code:              v.Code,
			Type:              v.Type,
			Value:             v.Value,
			Description:       v.Description,
			MinPurchaseAmount: v.MinPurchaseAmount,
			ValidFrom:         v.ValidFrom.Format(time.RFC3339),
			ValidUntil:        v.ValidUntil.Format(time.RFC3339),
			UsageLimitType:    v.UsageLimitType,
			BarcodeType:       v.BarcodeType,
			CreatedAt:         v.CreatedAt.Format(time.RFC3339),
		})
	}

	return export, nil
}

// ExportGiftCardsByIDs exports selected gift cards owned by the user.
func (s *ExportService) ExportGiftCardsByIDs(ctx context.Context, userID uuid.UUID, ids []uuid.UUID) (*BatchExportData, error) {
	export := &BatchExportData{
		ExportedAt: time.Now().UTC().Format(time.RFC3339),
		GiftCards:  make([]ExportGiftCard, 0, len(ids)),
	}

	// Get user's own gift card IDs for ownership verification
	ownGiftCards, _ := s.giftCardRepo.GetByUserID(ctx, userID)
	ownGiftCardIDs := make(map[uuid.UUID]bool, len(ownGiftCards))
	for _, gc := range ownGiftCards {
		ownGiftCardIDs[gc.ID] = true
	}

	for _, id := range ids {
		if !ownGiftCardIDs[id] {
			continue // skip gift cards not owned by user
		}
		gc, err := s.giftCardRepo.GetByID(ctx, id, "Merchant", "Transactions")
		if err != nil {
			continue
		}
		merchantName := gc.MerchantName
		if gc.Merchant != nil && gc.Merchant.Name != "" {
			merchantName = gc.Merchant.Name
		}

		exportGC := ExportGiftCard{
			ID:             gc.ID.String(),
			MerchantName:   merchantName,
			CardNumber:     gc.CardNumber,
			InitialBalance: gc.InitialBalance,
			CurrentBalance: gc.GetCurrentBalance(),
			Currency:       gc.Currency,
			PIN:            gc.PIN,
			Status:         gc.GetComputedStatus(),
			BarcodeType:    gc.BarcodeType,
			Notes:          gc.Notes,
			CreatedAt:      gc.CreatedAt.Format(time.RFC3339),
		}

		if gc.ExpiresAt != nil {
			exportGC.ExpiresAt = gc.ExpiresAt.Format(time.RFC3339)
		}

		exportGC.Transactions = make([]ExportGiftCardTransaction, len(gc.Transactions))
		for j, tx := range gc.Transactions {
			exportGC.Transactions[j] = ExportGiftCardTransaction{
				ID:              tx.ID.String(),
				Amount:          tx.Amount,
				Description:     tx.Description,
				TransactionDate: tx.TransactionDate.Format(time.RFC3339),
				CreatedAt:       tx.CreatedAt.Format(time.RFC3339),
			}
		}

		export.GiftCards = append(export.GiftCards, exportGC)
	}

	return export, nil
}
