// Package api contains JSON API handlers for gift cards.
package api

//
//nolint:revive // "api" is a meaningful package name for API handlers

import (
	"net/http"
	"savvy/internal/audit"
	"savvy/internal/models"
	"savvy/internal/services"
	"savvy/internal/validation"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// GiftCardsHandler handles gift card API endpoints.
type GiftCardsHandler struct {
	giftCardService services.GiftCardServiceInterface
	authzService    services.AuthzServiceInterface
	merchantService services.MerchantServiceInterface
	userService     services.UserServiceInterface
	favoriteService services.FavoriteServiceInterface
	shareService    services.ShareServiceInterface
	transferService services.TransferServiceInterface
}

// NewGiftCardsHandler creates a new gift cards API handler.
func NewGiftCardsHandler(
	giftCardService services.GiftCardServiceInterface,
	authzService services.AuthzServiceInterface,
	merchantService services.MerchantServiceInterface,
	userService services.UserServiceInterface,
	favoriteService services.FavoriteServiceInterface,
	shareService services.ShareServiceInterface,
	transferService services.TransferServiceInterface,
) *GiftCardsHandler {
	return &GiftCardsHandler{
		giftCardService: giftCardService,
		authzService:    authzService,
		merchantService: merchantService,
		userService:     userService,
		favoriteService: favoriteService,
		shareService:    shareService,
		transferService: transferService,
	}
}

// List returns all gift cards (owned + shared)
// GET /api/v1/gift-cards
// Optional query params: ?page=1&per_page=25 for pagination
func (h *GiftCardsHandler) List(c echo.Context) error {
	user := c.Get("current_user").(*models.User)

	page, perPage, isPaginated := parsePaginationParams(c)

	if isPaginated {
		result, err := h.giftCardService.GetUserGiftCardsPaginated(c.Request().Context(), user.ID, page, perPage)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "server_error",
				Message: "Failed to load gift cards",
			})
		}

		giftCardIDs := make([]uuid.UUID, len(result.Items))
		for i, gc := range result.Items {
			giftCardIDs[i] = gc.ID
		}
		favoriteIDs := getFavoriteIDsForResources(c.Request().Context(), h.favoriteService, user.ID, "gift_card", giftCardIDs)
		shareCounts, _ := h.shareService.GetGiftCardShareCounts(c.Request().Context(), giftCardIDs)

		dtos := ToGiftCardDTOs(result.Items, favoriteIDs)
		for i := range dtos {
			dtos[i].SharedWithCount = int(shareCounts[result.Items[i].ID])
			perms, err := h.authzService.CheckGiftCardAccess(c.Request().Context(), user.ID, result.Items[i].ID)
			if err == nil {
				dtos[i].Permissions = &PermissionDTO{
					CanView:             perms.CanView,
					CanEdit:             perms.CanEdit,
					CanDelete:           perms.CanDelete,
					CanEditTransactions: perms.CanEditTransactions,
					IsOwner:             perms.IsOwner,
				}
			}
		}

		return c.JSON(http.StatusOK, GiftCardListResponse{
			GiftCards: dtos,
			Pagination: &PaginationMeta{
				Total:      result.Total,
				Page:       result.Page,
				PerPage:    result.PerPage,
				TotalPages: result.TotalPages,
			},
		})
	}

	// Non-paginated: return all (backward-compatible)
	giftCards, err := h.giftCardService.GetUserGiftCards(c.Request().Context(), user.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to load gift cards",
		})
	}

	giftCardIDs := make([]uuid.UUID, len(giftCards))
	for i, giftCard := range giftCards {
		giftCardIDs[i] = giftCard.ID
	}
	favoriteIDs := getFavoriteIDsForResources(c.Request().Context(), h.favoriteService, user.ID, "gift_card", giftCardIDs)
	shareCounts, _ := h.shareService.GetGiftCardShareCounts(c.Request().Context(), giftCardIDs)

	dtos := ToGiftCardDTOs(giftCards, favoriteIDs)
	for i := range dtos {
		dtos[i].SharedWithCount = int(shareCounts[giftCards[i].ID])
		perms, err := h.authzService.CheckGiftCardAccess(c.Request().Context(), user.ID, giftCards[i].ID)
		if err == nil {
			dtos[i].Permissions = &PermissionDTO{
				CanView:             perms.CanView,
				CanEdit:             perms.CanEdit,
				CanDelete:           perms.CanDelete,
				CanEditTransactions: perms.CanEditTransactions,
				IsOwner:             perms.IsOwner,
			}
		}
	}

	return c.JSON(http.StatusOK, GiftCardListResponse{
		GiftCards: dtos,
	})
}

// Show returns a single gift card with permissions and transactions
// GET /api/v1/gift-cards/:id
func (h *GiftCardsHandler) Show(c echo.Context) error {
	user := c.Get("current_user").(*models.User)

	giftCardID, err := parseResourceID(c, "gift card")
	if err != nil {
		return err
	}

	// Check authorization
	perms, err := h.authzService.CheckGiftCardAccess(c.Request().Context(), user.ID, giftCardID)
	if err != nil {
		return c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "You don't have access to this gift card",
		})
	}

	giftCard, err := h.giftCardService.GetGiftCard(c.Request().Context(), giftCardID)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "Gift card not found",
		})
	}

	// Check if favorited
	isFavorite, err := h.favoriteService.IsFavorite(c.Request().Context(), user.ID, "gift_card", giftCardID)
	if err != nil {
		c.Logger().Warnf("Failed to check favorite status for gift card %s: %v", giftCardID, err)
	}

	giftCardDTO := ToGiftCardDTO(giftCard, isFavorite)
	giftCardDTO.Permissions = &PermissionDTO{
		CanView:             perms.CanView,
		CanEdit:             perms.CanEdit,
		CanDelete:           perms.CanDelete,
		CanEditTransactions: perms.CanEditTransactions,
		IsOwner:             perms.IsOwner,
	}

	// Get transactions (if available - transactions are loaded separately via Preload)
	var transactions []GiftCardTransactionDTO
	if giftCard.Transactions != nil {
		transactions = ToTransactionDTOs(giftCard.Transactions)
	}

	// Get shares if owner
	var shares []ShareDTO
	if perms.IsOwner {
		giftCardShares, err := h.shareService.GetGiftCardShares(c.Request().Context(), giftCardID)
		if err != nil {
			c.Logger().Warnf("Failed to load shares for gift card %s: %v", giftCardID, err)
		}
		shares = ToGiftCardShareDTOs(giftCardShares)
	}

	// No duplicate check for Show
	var duplicateWarning *DuplicateWarning

	return c.JSON(http.StatusOK, GiftCardDetailResponse{
		GiftCard:         giftCardDTO,
		Permissions:      *giftCardDTO.Permissions,
		Transactions:     transactions,
		Shares:           shares,
		DuplicateWarning: duplicateWarning,
	})
}

// Create creates a new gift card
// POST /api/v1/gift-cards
func (h *GiftCardsHandler) Create(c echo.Context) error {
	user := c.Get("current_user").(*models.User)

	var req GiftCardCreateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Validate required fields
	if req.CardNumber == "" || req.Currency == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_fields",
			Message: "Card number and currency are required",
		})
	}

	// Validate string lengths
	if err := validateStringLength(req.CardNumber, "card_number"); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "field_too_long", Message: err.Error()})
	}
	if err := validateStringLengthPtr(req.PIN, "pin"); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "field_too_long", Message: err.Error()})
	}
	if err := validateStringLengthPtr(req.Notes, "notes"); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "field_too_long", Message: err.Error()})
	}
	if err := validateStringLengthPtr(req.NewMerchantName, "name"); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "field_too_long", Message: err.Error()})
	}

	// Validate enum fields
	if err := validateEnum(c, req.Currency, validCurrencies, "currency"); err != nil {
		return err
	}

	// Validate initial balance bounds
	if err := validation.ValidateMonetaryAmount(req.InitialBalance, "initial_balance"); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_amount", Message: err.Error()})
	}

	// Handle merchant
	merchantID, merchantName, err := resolveMerchant(c, h.merchantService, req.MerchantID, req.NewMerchantName)
	if err != nil {
		return err
	}

	// Parse expires_at if provided
	var expiresAt *time.Time
	if req.ExpiresAt != nil {
		expTime, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_date",
				Message: "Invalid expires_at date format (use ISO 8601/RFC3339)",
			})
		}
		expiresAt = &expTime
	}

	// Check for duplicates (warning only)
	var duplicateWarning *DuplicateWarning
	duplicate, err := h.giftCardService.CheckDuplicate(c.Request().Context(), req.CardNumber, user.ID, nil)
	if err != nil {
		c.Logger().Warnf("Failed to check duplicate: %v", err)
	}
	if duplicate != nil {
		duplicateWarning = &DuplicateWarning{
			HasDuplicate:   true,
			MerchantName:   duplicate.MerchantName,
			ResourceNumber: duplicate.CardNumber,
			ExistingID:     duplicate.ID.String(),
		}
		c.Logger().Infof("Duplicate gift card detected: existing gift card %s", duplicate.ID)
	}

	// Create gift card
	giftCard := &models.GiftCard{
		UserID:         &user.ID,
		MerchantID:     merchantID,
		MerchantName:   merchantName,
		CardNumber:     req.CardNumber,
		InitialBalance: req.InitialBalance,
		CurrentBalance: req.InitialBalance, // Starts at initial balance
		Currency:       req.Currency,
		PIN:            stringOrDefault(req.PIN, ""),
		ExpiresAt:      expiresAt,
		Notes:          stringOrDefault(req.Notes, ""),
	}

	if err := h.giftCardService.CreateGiftCard(c.Request().Context(), giftCard); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to create gift card",
		})
	}

	// Reload with merchant relation
	giftCard, err = h.giftCardService.GetGiftCard(c.Request().Context(), giftCard.ID)
	if err != nil {
		c.Logger().Errorf("Failed to reload gift card %s after creation: %v", giftCard.ID, err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Gift card created but failed to reload",
		})
	}

	return c.JSON(http.StatusCreated, GiftCardDetailResponse{
		GiftCard: ToGiftCardDTO(giftCard, false),
		Permissions: PermissionDTO{
			CanView:             true,
			CanEdit:             true,
			CanDelete:           true,
			CanEditTransactions: true,
			IsOwner:             true,
		},
		Transactions:     []GiftCardTransactionDTO{}, // Empty initially
		DuplicateWarning: duplicateWarning,
	})
}

// Update updates a gift card
// PUT /api/v1/gift-cards/:id
func (h *GiftCardsHandler) Update(c echo.Context) error {
	user := c.Get("current_user").(*models.User)

	giftCardID, err := parseResourceID(c, "gift card")
	if err != nil {
		return err
	}

	// Check authorization
	perms, err := h.authzService.CheckGiftCardAccess(c.Request().Context(), user.ID, giftCardID)
	if err != nil || !perms.CanEdit {
		return c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "You don't have permission to edit this gift card",
		})
	}

	var req GiftCardUpdateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Validate string lengths
	if err := validateStringLengthPtr(req.CardNumber, "card_number"); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "field_too_long", Message: err.Error()})
	}
	if err := validateStringLengthPtr(req.PIN, "pin"); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "field_too_long", Message: err.Error()})
	}
	if err := validateStringLengthPtr(req.Notes, "notes"); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "field_too_long", Message: err.Error()})
	}

	// Get existing gift card
	giftCard, err := h.giftCardService.GetGiftCard(c.Request().Context(), giftCardID)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "Gift card not found",
		})
	}

	// Apply updates
	if err := h.applyGiftCardUpdates(c, giftCard, &req); err != nil {
		return err
	}

	// Check for duplicates if card number changed (warning only)
	duplicateWarning := checkGiftCardDuplicate(c, h.giftCardService, req.CardNumber, user.ID, &giftCardID)

	if err := h.giftCardService.UpdateGiftCard(c.Request().Context(), giftCard); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to update gift card",
		})
	}

	// Reload
	giftCard, err = h.giftCardService.GetGiftCard(c.Request().Context(), giftCardID)
	if err != nil {
		c.Logger().Errorf("Failed to reload gift card %s after update: %v", giftCardID, err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Gift card updated but failed to reload",
		})
	}
	isFavorite, err := h.favoriteService.IsFavorite(c.Request().Context(), user.ID, "gift_card", giftCardID)
	if err != nil {
		c.Logger().Warnf("Failed to check favorite status for gift card %s: %v", giftCardID, err)
	}

	// Map transactions
	var transactions []GiftCardTransactionDTO
	if giftCard.Transactions != nil {
		transactions = ToTransactionDTOs(giftCard.Transactions)
	}

	// Get shares if owner
	var shares []ShareDTO
	if perms.IsOwner {
		giftCardShares, err := h.shareService.GetGiftCardShares(c.Request().Context(), giftCardID)
		if err != nil {
			c.Logger().Warnf("Failed to load shares for gift card %s: %v", giftCardID, err)
		}
		shares = ToGiftCardShareDTOs(giftCardShares)
	}

	return c.JSON(http.StatusOK, GiftCardDetailResponse{
		GiftCard: ToGiftCardDTO(giftCard, isFavorite),
		Permissions: PermissionDTO{
			CanView:             perms.CanView,
			CanEdit:             perms.CanEdit,
			CanDelete:           perms.CanDelete,
			CanEditTransactions: perms.CanEditTransactions,
			IsOwner:             perms.IsOwner,
		},
		Transactions:     transactions,
		Shares:           shares,
		DuplicateWarning: duplicateWarning,
	})
}

// Delete deletes a gift card
// DELETE /api/v1/gift-cards/:id
func (h *GiftCardsHandler) Delete(c echo.Context) error {
	return handleResourceDelete(c, "gift card", h.authzService.CheckGiftCardAccess, h.giftCardService.DeleteGiftCard)
}

// ToggleFavorite toggles favorite status
// POST /api/v1/gift-cards/:id/favorite
func (h *GiftCardsHandler) ToggleFavorite(c echo.Context) error {
	return handleResourceToggleFavorite(c, "gift_card", h.authzService.CheckGiftCardAccess, h.favoriteService)
}

// CreateShare creates a new share
// POST /api/v1/gift-cards/:id/share
func (h *GiftCardsHandler) CreateShare(c echo.Context) error {
	user := c.Get("current_user").(*models.User)

	giftCardID, err := parseResourceID(c, "gift card")
	if err != nil {
		return err
	}

	// Check authorization - only owner can share
	perms, err := h.authzService.CheckGiftCardAccess(c.Request().Context(), user.ID, giftCardID)
	if err != nil || !perms.IsOwner {
		return c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "Only the owner can share this gift card",
		})
	}

	var req ShareCreateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Find user by email
	email := strings.ToLower(strings.TrimSpace(req.Email))
	sharedUser, err := h.userService.GetUserByEmail(c.Request().Context(), email)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "share_failed",
			Message: "Could not share with this email address",
		})
	}

	// Prevent self-sharing
	if sharedUser.ID == user.ID {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "self_share",
			Message: "Cannot share with yourself",
		})
	}

	// Create share with granular permissions
	canEdit := req.CanEdit != nil && *req.CanEdit
	canDelete := req.CanDelete != nil && *req.CanDelete
	canEditTransactions := req.CanEditTransactions != nil && *req.CanEditTransactions

	if err := h.shareService.CreateGiftCardShare(c.Request().Context(), user.ID, giftCardID, sharedUser.ID, canEdit, canDelete, canEditTransactions); err != nil {
		c.Logger().Errorf("Failed to create gift card share for gift card %s: %v", giftCardID, err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to share gift card",
		})
	}

	// Return updated shares list
	shares, err := h.shareService.GetGiftCardShares(c.Request().Context(), giftCardID)
	if err != nil {
		c.Logger().Warnf("Failed to load shares for gift card %s: %v", giftCardID, err)
	}

	return c.JSON(http.StatusCreated, map[string]any{
		"message": "Gift card shared successfully",
		"shares":  ToGiftCardShareDTOs(shares),
	})
}

// UpdateShare updates share permissions
// PATCH /api/v1/gift-cards/:id/share/:sharedWithID
func (h *GiftCardsHandler) UpdateShare(c echo.Context) error {
	user := c.Get("current_user").(*models.User)

	giftCardID, err := parseResourceID(c, "gift card")
	if err != nil {
		return err
	}

	sharedWithID, err := parseUUIDParam(c, "sharedWithID", "invalid_user_id", "Invalid user ID")
	if err != nil {
		return err
	}

	// Check authorization - only owner can update shares
	perms, err := h.authzService.CheckGiftCardAccess(c.Request().Context(), user.ID, giftCardID)
	if err != nil || !perms.IsOwner {
		return c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "Only the owner can update share permissions",
		})
	}

	var req ShareUpdateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Update share
	canEdit := req.CanEdit != nil && *req.CanEdit
	canDelete := req.CanDelete != nil && *req.CanDelete
	canEditTransactions := req.CanEditTransactions != nil && *req.CanEditTransactions

	// Add audit context (user ID, IP address, user agent) for audit logging
	ctx := audit.AddAuditContextToContext(c.Request().Context(), user.ID, c.RealIP(), c.Request().UserAgent())

	if err := h.shareService.UpdateGiftCardShare(ctx, user.ID, giftCardID, sharedWithID, canEdit, canDelete, canEditTransactions); err != nil {
		c.Logger().Errorf("Failed to update gift card share for gift card %s, user %s: %v", giftCardID, sharedWithID, err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to update share permissions",
		})
	}

	// Return updated shares list
	shares, err := h.shareService.GetGiftCardShares(c.Request().Context(), giftCardID)
	if err != nil {
		c.Logger().Warnf("Failed to load shares for gift card %s: %v", giftCardID, err)
	}

	return c.JSON(http.StatusOK, map[string]any{
		"message": "Share permissions updated successfully",
		"shares":  ToGiftCardShareDTOs(shares),
	})
}

// DeleteShare removes a share
// DELETE /api/v1/gift-cards/:id/share/:sharedWithID
func (h *GiftCardsHandler) DeleteShare(c echo.Context) error {
	return handleResourceDeleteShare(c, "gift card", h.authzService.CheckGiftCardAccess, h.shareService.DeleteGiftCardShare)
}

// Transfer transfers ownership
// POST /api/v1/gift-cards/:id/transfer
func (h *GiftCardsHandler) Transfer(c echo.Context) error {
	return handleResourceTransfer(c, "gift card", h.authzService.CheckGiftCardAccess, h.transferService.TransferGiftCardOwnership, h.userService)
}

// applyGiftCardUpdates applies partial update fields from the request to the gift card model.
func (h *GiftCardsHandler) applyGiftCardUpdates(c echo.Context, giftCard *models.GiftCard, req *GiftCardUpdateRequest) error {
	if req.InitialBalance != nil {
		if err := validation.ValidateMonetaryAmount(*req.InitialBalance, "initial_balance"); err != nil {
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_amount", Message: err.Error()})
		}
		if *req.InitialBalance != giftCard.InitialBalance {
			// Adjust current_balance by the same delta (current = initial - SUM(transactions))
			delta := *req.InitialBalance - giftCard.InitialBalance
			giftCard.InitialBalance = *req.InitialBalance
			giftCard.CurrentBalance += delta
		}
	}
	if req.MerchantID != nil {
		mid, name, err := resolveMerchantUpdate(c, h.merchantService, *req.MerchantID)
		if err != nil {
			return err
		}
		giftCard.MerchantID = mid
		giftCard.MerchantName = name
		if mid != nil {
			// CRITICAL: Clear the Merchant relation to prevent GORM from overwriting MerchantID
			giftCard.Merchant = nil
		}
	}
	if req.CardNumber != nil {
		giftCard.CardNumber = *req.CardNumber
	}
	if req.Currency != nil {
		if err := validateEnum(c, *req.Currency, validCurrencies, "currency"); err != nil {
			return err
		}
		giftCard.Currency = *req.Currency
	}
	if req.PIN != nil {
		giftCard.PIN = *req.PIN
	}
	if req.ExpiresAt != nil {
		expTime, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_date",
				Message: "Invalid expires_at date format (use ISO 8601/RFC3339)",
			})
		}
		giftCard.ExpiresAt = &expTime
	}
	if req.BarcodeType != nil {
		if err := validateEnum(c, *req.BarcodeType, validBarcodeTypes, "barcode_type"); err != nil {
			return err
		}
		giftCard.BarcodeType = *req.BarcodeType
	}
	if req.Notes != nil {
		giftCard.Notes = *req.Notes
	}
	return nil
}

// checkGiftCardDuplicate checks for duplicate gift cards and returns a warning if found.
func checkGiftCardDuplicate(c echo.Context, svc services.GiftCardServiceInterface, cardNumber *string, userID uuid.UUID, excludeID *uuid.UUID) *DuplicateWarning {
	if cardNumber == nil || *cardNumber == "" {
		return nil
	}
	duplicate, err := svc.CheckDuplicate(c.Request().Context(), *cardNumber, userID, excludeID)
	if err != nil {
		c.Logger().Warnf("Failed to check duplicate: %v", err)
		return nil
	}
	if duplicate == nil {
		return nil
	}
	return &DuplicateWarning{
		HasDuplicate:   true,
		MerchantName:   duplicate.MerchantName,
		ResourceNumber: duplicate.CardNumber,
		ExistingID:     duplicate.ID.String(),
	}
}

// ==================== Transaction Endpoints ====================

// ListTransactions returns all transactions for a gift card
// GET /api/v1/gift-cards/:id/transactions
func (h *GiftCardsHandler) ListTransactions(c echo.Context) error {
	user := c.Get("current_user").(*models.User)

	giftCardID, err := parseResourceID(c, "gift card")
	if err != nil {
		return err
	}

	// Check authorization
	_, err = h.authzService.CheckGiftCardAccess(c.Request().Context(), user.ID, giftCardID)
	if err != nil {
		return c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "You don't have access to this gift card",
		})
	}

	// Get gift card with transactions
	giftCard, err := h.giftCardService.GetGiftCard(c.Request().Context(), giftCardID)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "Gift card not found",
		})
	}

	// Convert transactions to DTOs
	var transactions []GiftCardTransactionDTO
	if giftCard.Transactions != nil {
		transactions = ToTransactionDTOs(giftCard.Transactions)
	} else {
		transactions = []GiftCardTransactionDTO{}
	}

	return c.JSON(http.StatusOK, map[string]any{
		"transactions": transactions,
	})
}

// CreateTransaction creates a new transaction (updates balance via DB trigger)
// POST /api/v1/gift-cards/:id/transactions
func (h *GiftCardsHandler) CreateTransaction(c echo.Context) error {
	user := c.Get("current_user").(*models.User)

	giftCardID, err := parseResourceID(c, "gift card")
	if err != nil {
		return err
	}

	// Check authorization - needs can_edit_transactions permission
	perms, err := h.authzService.CheckGiftCardAccess(c.Request().Context(), user.ID, giftCardID)
	if err != nil || !perms.CanEditTransactions {
		return c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "You don't have permission to edit transactions for this gift card",
		})
	}

	var req TransactionCreateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Validate amount bounds (negative for spending, positive for refunds)
	if err := validation.ValidateTransactionAmount(req.Amount); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_amount",
			Message: err.Error(),
		})
	}

	// Parse transaction date (defaults to now)
	transactionDate := time.Now()
	if req.TransactionDate != nil {
		txDate, err := time.Parse(time.RFC3339, *req.TransactionDate)
		if err != nil {
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_date",
				Message: "Invalid transaction_date format (use ISO 8601/RFC3339)",
			})
		}
		transactionDate = txDate
	}

	// Create transaction
	transaction := &models.GiftCardTransaction{
		GiftCardID:      giftCardID,
		Amount:          req.Amount,
		Description:     stringOrDefault(req.Description, ""),
		TransactionDate: transactionDate,
	}

	if err := h.giftCardService.CreateTransaction(c.Request().Context(), transaction); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to create transaction",
		})
	}

	// Return created transaction + updated balance
	giftCard, err := h.giftCardService.GetGiftCard(c.Request().Context(), giftCardID)
	if err != nil {
		c.Logger().Errorf("Failed to reload gift card %s after transaction: %v", giftCardID, err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Transaction created but failed to reload gift card",
		})
	}

	return c.JSON(http.StatusCreated, map[string]any{
		"message":         "Transaction created successfully",
		"transaction":     ToTransactionDTO(transaction),
		"current_balance": giftCard.CurrentBalance,
	})
}

// DeleteTransaction deletes a transaction (balance recalculated via DB trigger)
// DELETE /api/v1/gift-cards/:id/transactions/:transactionID
func (h *GiftCardsHandler) DeleteTransaction(c echo.Context) error {
	user := c.Get("current_user").(*models.User)

	giftCardID, err := parseResourceID(c, "gift card")
	if err != nil {
		return err
	}

	transactionID, err := parseUUIDParam(c, "transactionID", "invalid_transaction_id", "Invalid transaction ID")
	if err != nil {
		return err
	}

	// Check authorization - needs can_edit_transactions permission
	perms, err := h.authzService.CheckGiftCardAccess(c.Request().Context(), user.ID, giftCardID)
	if err != nil || !perms.CanEditTransactions {
		return c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "You don't have permission to edit transactions for this gift card",
		})
	}

	// Verify transaction belongs to this gift card
	transaction, err := h.giftCardService.GetTransaction(c.Request().Context(), transactionID, giftCardID)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "Transaction not found",
		})
	}

	if transaction.GiftCardID != giftCardID {
		return c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "Transaction does not belong to this gift card",
		})
	}

	if err := h.giftCardService.DeleteTransaction(c.Request().Context(), transactionID); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to delete transaction",
		})
	}

	// Return updated balance
	giftCard, err := h.giftCardService.GetGiftCard(c.Request().Context(), giftCardID)
	if err != nil {
		c.Logger().Errorf("Failed to reload gift card %s after transaction deletion: %v", giftCardID, err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Transaction deleted but failed to reload gift card",
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"message":         "Transaction deleted successfully",
		"current_balance": giftCard.CurrentBalance,
	})
}
