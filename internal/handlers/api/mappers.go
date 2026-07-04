// Package api contains mappers to convert models to DTOs.
//
//nolint:revive // "api" is a meaningful package name for API handlers
package api

import (
	"time"

	"savvy/internal/models"
	"savvy/internal/services"
)

// appLocation is the configured timezone for date-based status calculations.
// Set via SetAppLocation during server startup. Defaults to UTC.
var appLocation = time.UTC

// SetAppLocation sets the timezone used for voucher status calculations.
func SetAppLocation(loc *time.Location) {
	if loc != nil {
		appLocation = loc
	}
}

// ==================== User Mappers ====================

// ToUserDTO converts a User model to UserDTO
func ToUserDTO(user *models.User) UserDTO {
	return UserDTO{
		ID:        user.ID.String(),
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		IsAdmin:   user.IsAdmin(), // Method call, not field
	}
}

// ==================== Merchant Mappers ====================

// ToMerchantDTO converts a Merchant model to MerchantDTO
func ToMerchantDTO(merchant *models.Merchant) MerchantDTO {
	var color, logoURL, website *string
	if merchant.Color != "" {
		color = &merchant.Color
	}
	if merchant.LogoURL != "" {
		logoURL = &merchant.LogoURL
	}
	if merchant.Website != "" {
		website = &merchant.Website
	}

	return MerchantDTO{
		ID:        merchant.ID.String(),
		Name:      merchant.Name,
		Color:     color,
		LogoURL:   logoURL,
		Website:   website,
		CreatedAt: FormatTime(merchant.CreatedAt),
		UpdatedAt: FormatTime(merchant.UpdatedAt),
	}
}

// ToMerchantDTOs converts a slice of Merchant models to DTOs
func ToMerchantDTOs(merchants []models.Merchant) []MerchantDTO {
	dtos := make([]MerchantDTO, len(merchants))
	for i, m := range merchants {
		dtos[i] = ToMerchantDTO(&m)
	}
	return dtos
}

// ==================== Permission Mappers ====================

// ToPermissionDTO converts ResourcePermissions to PermissionDTO
func ToPermissionDTO(perms *services.ResourcePermissions) PermissionDTO {
	return PermissionDTO{
		CanView:             perms.CanView,
		CanEdit:             perms.CanEdit,
		CanDelete:           perms.CanDelete,
		CanEditTransactions: perms.CanEditTransactions,
		IsOwner:             perms.IsOwner,
	}
}

// ==================== Card Mappers ====================

// ToCardDTO converts a Card model to CardDTO
func ToCardDTO(card *models.Card, isFavorite bool) CardDTO {
	var merchantID *string
	var merchant *MerchantDTO
	var owner *UserDTO
	var program, barcodeType, notes *string

	if card.MerchantID != nil {
		mid := card.MerchantID.String()
		merchantID = &mid
	}

	if card.Merchant != nil {
		merchantDTO := ToMerchantDTO(card.Merchant)
		merchant = &merchantDTO
	}

	if card.User != nil {
		userDTO := ToUserDTO(card.User)
		owner = &userDTO
	}

	if card.Program != "" {
		program = &card.Program
	}
	if card.BarcodeType != "" {
		barcodeType = &card.BarcodeType
	}
	if card.Notes != "" {
		notes = &card.Notes
	}

	return CardDTO{
		ID:          card.ID.String(),
		MerchantID:  merchantID,
		Merchant:    merchant,
		Owner:       owner,
		Program:     program,
		CardNumber:  card.CardNumber,
		BarcodeType: barcodeType,
		Notes:       notes,
		Status:      card.Status,
		IsFavorite:  isFavorite,
		IsShared:    card.UserID == nil, // Shared if no owner
		CreatedAt:   FormatTime(card.CreatedAt),
		UpdatedAt:   FormatTime(card.UpdatedAt),
	}
}

// ToCardDTOs converts a slice of Card models to DTOs
func ToCardDTOs(cards []models.Card, favoriteIDs map[string]bool) []CardDTO {
	dtos := make([]CardDTO, len(cards))
	for i, c := range cards {
		dtos[i] = ToCardDTO(&c, favoriteIDs[c.ID.String()])
	}
	return dtos
}

// ==================== Voucher Mappers ====================

// ToVoucherDTO converts a Voucher model to VoucherDTO
func ToVoucherDTO(voucher *models.Voucher, isFavorite bool) VoucherDTO {
	var merchantID *string
	var merchant *MerchantDTO
	var owner *UserDTO
	var description *string

	if voucher.MerchantID != nil {
		mid := voucher.MerchantID.String()
		merchantID = &mid
	}

	if voucher.Merchant != nil {
		merchantDTO := ToMerchantDTO(voucher.Merchant)
		merchant = &merchantDTO
	}

	if voucher.User != nil {
		userDTO := ToUserDTO(voucher.User)
		owner = &userDTO
	}

	if voucher.Description != "" {
		description = &voucher.Description
	}

	// Calculate status based on valid_from and valid_until using configured timezone.
	// Compare dates only (not timestamps) to avoid timezone boundary issues.
	now := time.Now().In(appLocation)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, appLocation)
	validFrom := time.Date(voucher.ValidFrom.Year(), voucher.ValidFrom.Month(), voucher.ValidFrom.Day(), 0, 0, 0, 0, appLocation)
	validUntil := time.Date(voucher.ValidUntil.Year(), voucher.ValidUntil.Month(), voucher.ValidUntil.Day(), 0, 0, 0, 0, appLocation)

	var status string
	if today.After(validUntil) {
		status = "expired"
	} else if today.Before(validFrom) {
		status = "inactive"
	} else {
		status = "valid"
	}

	return VoucherDTO{
		ID:                voucher.ID.String(),
		MerchantID:        merchantID,
		Merchant:          merchant,
		Owner:             owner,
		Code:              voucher.Code,
		Type:              voucher.Type,
		Value:             voucher.Value,
		Currency:          voucher.Currency,
		Description:       description,
		MinPurchaseAmount: voucher.MinPurchaseAmount,
		ValidFrom:         FormatTime(voucher.ValidFrom),
		ValidUntil:        FormatTime(voucher.ValidUntil),
		UsageLimitType:    voucher.UsageLimitType,
		BarcodeType:       voucher.BarcodeType,
		Status:            status,
		IsFavorite:        isFavorite,
		IsShared:          voucher.UserID == nil,
		CreatedAt:         FormatTime(voucher.CreatedAt),
		UpdatedAt:         FormatTime(voucher.UpdatedAt),
	}
}

// ToVoucherDTOs converts a slice of Voucher models to DTOs
func ToVoucherDTOs(vouchers []models.Voucher, favoriteIDs map[string]bool) []VoucherDTO {
	dtos := make([]VoucherDTO, len(vouchers))
	for i, v := range vouchers {
		dtos[i] = ToVoucherDTO(&v, favoriteIDs[v.ID.String()])
	}
	return dtos
}

// ==================== Gift Card Mappers ====================

// ToGiftCardDTO converts a GiftCard model to GiftCardDTO
func ToGiftCardDTO(giftCard *models.GiftCard, isFavorite bool) GiftCardDTO {
	var merchantID *string
	var merchant *MerchantDTO
	var owner *UserDTO
	var pin, expiresAt, notes, barcodeType *string

	if giftCard.MerchantID != nil {
		mid := giftCard.MerchantID.String()
		merchantID = &mid
	}

	if giftCard.Merchant != nil {
		merchantDTO := ToMerchantDTO(giftCard.Merchant)
		merchant = &merchantDTO
	}

	if giftCard.User != nil {
		userDTO := ToUserDTO(giftCard.User)
		owner = &userDTO
	}

	if giftCard.PIN != "" {
		pin = &giftCard.PIN
	}
	if giftCard.ExpiresAt != nil {
		expiresAt = FormatTimePtr(giftCard.ExpiresAt)
	}
	if giftCard.Notes != "" {
		notes = &giftCard.Notes
	}
	if giftCard.BarcodeType != "" {
		barcodeType = &giftCard.BarcodeType
	}

	return GiftCardDTO{
		ID:             giftCard.ID.String(),
		MerchantID:     merchantID,
		Merchant:       merchant,
		Owner:          owner,
		CardNumber:     giftCard.CardNumber,
		InitialBalance: giftCard.InitialBalance,
		CurrentBalance: giftCard.CurrentBalance,
		Currency:       giftCard.Currency,
		PIN:            pin,
		BarcodeType:    barcodeType,
		ExpiresAt:      expiresAt,
		Notes:          notes,
		IsFavorite:     isFavorite,
		IsShared:       giftCard.UserID == nil,
		CreatedAt:      FormatTime(giftCard.CreatedAt),
		UpdatedAt:      FormatTime(giftCard.UpdatedAt),
	}
}

// ToGiftCardDTOs converts a slice of GiftCard models to DTOs
func ToGiftCardDTOs(giftCards []models.GiftCard, favoriteIDs map[string]bool) []GiftCardDTO {
	dtos := make([]GiftCardDTO, len(giftCards))
	for i, g := range giftCards {
		dtos[i] = ToGiftCardDTO(&g, favoriteIDs[g.ID.String()])
	}
	return dtos
}

// ==================== Transaction Mappers ====================

// ToTransactionDTO converts a GiftCardTransaction model to DTO
func ToTransactionDTO(transaction *models.GiftCardTransaction) GiftCardTransactionDTO {
	var description *string
	if transaction.Description != "" {
		description = &transaction.Description
	}

	return GiftCardTransactionDTO{
		ID:              transaction.ID.String(),
		GiftCardID:      transaction.GiftCardID.String(),
		Amount:          transaction.Amount,
		Description:     description,
		TransactionDate: FormatTime(transaction.TransactionDate),
		CreatedAt:       FormatTime(transaction.CreatedAt),
	}
}

// ToTransactionDTOs converts a slice of Transaction models to DTOs
func ToTransactionDTOs(transactions []models.GiftCardTransaction) []GiftCardTransactionDTO {
	dtos := make([]GiftCardTransactionDTO, len(transactions))
	for i, t := range transactions {
		dtos[i] = ToTransactionDTO(&t)
	}
	return dtos
}

// ==================== Share Mappers ====================

// ToCardShareDTO converts a CardShare model to ShareDTO
func ToCardShareDTO(share *models.CardShare) ShareDTO {
	return ShareDTO{
		ID:             share.ID.String(),
		SharedWithUser: ToUserDTO(share.SharedWithUser),
		CanEdit:        share.CanEdit,
		CanDelete:      share.CanDelete,
		CreatedAt:      FormatTime(share.CreatedAt),
	}
}

// ToCardShareDTOs converts a slice of CardShare models to DTOs
func ToCardShareDTOs(shares []models.CardShare) []ShareDTO {
	dtos := make([]ShareDTO, len(shares))
	for i, s := range shares {
		dtos[i] = ToCardShareDTO(&s)
	}
	return dtos
}

// ToVoucherShareDTO converts a VoucherShare model to ShareDTO
func ToVoucherShareDTO(share *models.VoucherShare) ShareDTO {
	return ShareDTO{
		ID:             share.ID.String(),
		SharedWithUser: ToUserDTO(share.SharedWithUser),
		CanEdit:        false, // Vouchers are always read-only
		CanDelete:      false,
		CreatedAt:      FormatTime(share.CreatedAt),
	}
}

// ToVoucherShareDTOs converts a slice of VoucherShare models to DTOs
func ToVoucherShareDTOs(shares []models.VoucherShare) []ShareDTO {
	dtos := make([]ShareDTO, len(shares))
	for i, s := range shares {
		dtos[i] = ToVoucherShareDTO(&s)
	}
	return dtos
}

// ToGiftCardShareDTO converts a GiftCardShare model to ShareDTO
func ToGiftCardShareDTO(share *models.GiftCardShare) ShareDTO {
	return ShareDTO{
		ID:                  share.ID.String(),
		SharedWithUser:      ToUserDTO(share.SharedWithUser),
		CanEdit:             share.CanEdit,
		CanDelete:           share.CanDelete,
		CanEditTransactions: share.CanEditTransactions,
		CreatedAt:           FormatTime(share.CreatedAt),
	}
}

// ToGiftCardShareDTOs converts a slice of GiftCardShare models to DTOs
func ToGiftCardShareDTOs(shares []models.GiftCardShare) []ShareDTO {
	dtos := make([]ShareDTO, len(shares))
	for i, s := range shares {
		dtos[i] = ToGiftCardShareDTO(&s)
	}
	return dtos
}

// ==================== Admin User Mappers ====================

// ToAdminUserDTO converts User model to AdminUserDTO
func ToAdminUserDTO(user *models.User) AdminUserDTO {
	return AdminUserDTO{
		ID:           user.ID.String(),
		Email:        user.Email,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Role:         user.Role,
		AuthProvider: user.AuthProvider,
		CreatedAt:    FormatTime(user.CreatedAt),
		UpdatedAt:    FormatTime(user.UpdatedAt),
	}
}

// ToAdminUserDTOs converts slice of User models to DTOs
func ToAdminUserDTOs(users []models.User) []AdminUserDTO {
	dtos := make([]AdminUserDTO, len(users))
	for i, u := range users {
		dtos[i] = ToAdminUserDTO(&u)
	}
	return dtos
}

// ==================== Audit Log Mappers ====================

// ToAuditLogDTO converts AuditLog model to DTO
func ToAuditLogDTO(log *models.AuditLog) AuditLogDTO {
	var userID *string
	var user *UserDTO

	if log.UserID != nil {
		uid := log.UserID.String()
		userID = &uid
	}

	if log.User != nil {
		userDTO := ToUserDTO(log.User)
		user = &userDTO
	}

	return AuditLogDTO{
		ID:           log.ID.String(),
		UserID:       userID,
		User:         user,
		Action:       log.Action,
		ResourceType: log.ResourceType,
		ResourceID:   log.ResourceID.String(),
		ResourceData: log.ResourceData,
		IPAddress:    log.IPAddress,
		UserAgent:    log.UserAgent,
		CreatedAt:    FormatTime(log.CreatedAt),
	}
}

// ToAuditLogDTOs converts slice of AuditLog models to DTOs
func ToAuditLogDTOs(logs []models.AuditLog) []AuditLogDTO {
	dtos := make([]AuditLogDTO, len(logs))
	for i, l := range logs {
		dtos[i] = ToAuditLogDTO(&l)
	}
	return dtos
}
