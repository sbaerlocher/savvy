// Package api contains JSON API handlers for cards.
package api //nolint:revive // "api" is a meaningful package name for API handlers

//

import (
	"encoding/json"
	"net/http"
	"savvy/internal/audit"
	"savvy/internal/models"
	"savvy/internal/services"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// CardsHandler handles card API endpoints.
type CardsHandler struct {
	cardService     services.CardServiceInterface
	authzService    services.AuthzServiceInterface
	merchantService services.MerchantServiceInterface
	userService     services.UserServiceInterface
	favoriteService services.FavoriteServiceInterface
	shareService    services.ShareServiceInterface
	transferService services.TransferServiceInterface
	adminService    services.AdminServiceInterface
}

// NewCardsHandler creates a new cards API handler.
func NewCardsHandler(
	cardService services.CardServiceInterface,
	authzService services.AuthzServiceInterface,
	merchantService services.MerchantServiceInterface,
	userService services.UserServiceInterface,
	favoriteService services.FavoriteServiceInterface,
	shareService services.ShareServiceInterface,
	transferService services.TransferServiceInterface,
	adminService services.AdminServiceInterface,
) *CardsHandler {
	return &CardsHandler{
		cardService:     cardService,
		authzService:    authzService,
		merchantService: merchantService,
		userService:     userService,
		favoriteService: favoriteService,
		shareService:    shareService,
		transferService: transferService,
		adminService:    adminService,
	}
}

// List returns all cards (owned + shared)
// GET /api/v1/cards
// Optional query params: ?page=1&per_page=25 for pagination
func (h *CardsHandler) List(c echo.Context) error {
	user := c.Get("current_user").(*models.User)

	page, perPage, isPaginated := parsePaginationParams(c)

	if isPaginated {
		result, err := h.cardService.GetUserCardsPaginated(c.Request().Context(), user.ID, page, perPage)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "server_error",
				Message: "Failed to load cards",
			})
		}

		cardIDs := make([]uuid.UUID, len(result.Items))
		for i, card := range result.Items {
			cardIDs[i] = card.ID
		}
		favoriteIDs := getFavoriteIDsForResources(c.Request().Context(), h.favoriteService, user.ID, "card", cardIDs)
		shareCounts, _ := h.shareService.GetCardShareCounts(c.Request().Context(), cardIDs)

		dtos := ToCardDTOs(result.Items, favoriteIDs)
		for i, card := range result.Items {
			dtos[i].SharedWithCount = int(shareCounts[card.ID])
		}

		return c.JSON(http.StatusOK, CardListResponse{
			Cards: dtos,
			Pagination: &PaginationMeta{
				Total:      result.Total,
				Page:       result.Page,
				PerPage:    result.PerPage,
				TotalPages: result.TotalPages,
			},
		})
	}

	// Non-paginated: return all (backward-compatible)
	cards, err := h.cardService.GetUserCards(c.Request().Context(), user.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to load cards",
		})
	}

	cardIDs := make([]uuid.UUID, len(cards))
	for i, card := range cards {
		cardIDs[i] = card.ID
	}
	favoriteIDs := getFavoriteIDsForResources(c.Request().Context(), h.favoriteService, user.ID, "card", cardIDs)
	shareCounts, _ := h.shareService.GetCardShareCounts(c.Request().Context(), cardIDs)

	dtos := ToCardDTOs(cards, favoriteIDs)
	for i, card := range cards {
		dtos[i].SharedWithCount = int(shareCounts[card.ID])
	}

	return c.JSON(http.StatusOK, CardListResponse{
		Cards: dtos,
	})
}

// Show returns a single card with permissions
// GET /api/v1/cards/:id
func (h *CardsHandler) Show(c echo.Context) error {
	user := c.Get("current_user").(*models.User)

	cardID, err := parseResourceID(c, "card")
	if err != nil {
		return err
	}

	// Check authorization
	perms, err := h.authzService.CheckCardAccess(c.Request().Context(), user.ID, cardID)
	if err != nil {
		return c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "You don't have access to this card",
		})
	}

	card, err := h.cardService.GetCard(c.Request().Context(), cardID)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "Card not found",
		})
	}

	// Check if favorited
	isFavorite, err := h.favoriteService.IsFavorite(c.Request().Context(), user.ID, "card", cardID)
	if err != nil {
		c.Logger().Warnf("Failed to check favorite status for card %s: %v", cardID, err)
	}

	cardDTO := ToCardDTO(card, isFavorite)
	permDTO := PermissionDTO{
		CanView:   perms.CanView,
		CanEdit:   perms.CanEdit,
		CanDelete: perms.CanDelete,
		IsOwner:   perms.IsOwner,
	}
	cardDTO.Permissions = &permDTO

	// Get shares if owner
	var shares []ShareDTO
	if perms.IsOwner {
		cardShares, err := h.shareService.GetCardShares(c.Request().Context(), cardID)
		if err != nil {
			c.Logger().Warnf("Failed to load shares for card %s: %v", cardID, err)
		}
		shares = ToCardShareDTOs(cardShares)
	}

	// No duplicate check for Show (only for Create/Update)
	var duplicateWarning *DuplicateWarning

	return c.JSON(http.StatusOK, CardDetailResponse{
		Card:             cardDTO,
		Permissions:      permDTO,
		Shares:           shares,
		DuplicateWarning: duplicateWarning,
	})
}

// Create creates a new card
// POST /api/v1/cards
func (h *CardsHandler) Create(c echo.Context) error {
	user := c.Get("current_user").(*models.User)

	var req CardCreateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Validate required fields
	if req.CardNumber == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_fields",
			Message: "Card number is required",
		})
	}

	// Validate string lengths
	if err := validateStringLength(req.CardNumber, "card_number"); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "field_too_long", Message: err.Error()})
	}
	if err := validateStringLengthPtr(req.Notes, "notes"); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "field_too_long", Message: err.Error()})
	}
	if err := validateStringLengthPtr(req.Program, "program"); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "field_too_long", Message: err.Error()})
	}
	if err := validateStringLengthPtr(req.NewMerchantName, "name"); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "field_too_long", Message: err.Error()})
	}

	// Handle merchant (existing or new)
	var merchantID *uuid.UUID
	var merchantName string

	if req.MerchantID != nil && *req.MerchantID != "" {
		// Use existing merchant
		mid, err := uuid.Parse(*req.MerchantID)
		if err != nil {
			c.Logger().Errorf("Failed to parse merchant ID: %v", err)
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_merchant_id",
				Message: "Invalid merchant ID",
			})
		}
		merchantID = &mid

		// Get merchant name
		merchant, err := h.merchantService.GetMerchantByID(c.Request().Context(), mid)
		if err != nil {
			c.Logger().Errorf("Failed to get merchant: %v", err)
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_merchant",
				Message: "Merchant not found",
			})
		}
		merchantName = merchant.Name
	} else if req.NewMerchantName != nil && *req.NewMerchantName != "" {
		// Create new merchant
		merchant := &models.Merchant{Name: *req.NewMerchantName}
		if err := h.merchantService.CreateMerchant(c.Request().Context(), merchant); err != nil {
			c.Logger().Errorf("Failed to create merchant: %v", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "server_error",
				Message: "Failed to create merchant",
			})
		}
		merchantID = &merchant.ID
		merchantName = merchant.Name
	} else {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_merchant",
			Message: "Either merchant_id or new_merchant_name is required",
		})
	}

	// Validate enum fields
	barcodeType := stringOrDefault(req.BarcodeType, "CODE128")
	if err := validateEnum(c, barcodeType, validBarcodeTypes, "barcode_type"); err != nil {
		return err
	}
	status := stringOrDefault(req.Status, "active")
	if err := validateEnum(c, status, validStatuses, "status"); err != nil {
		return err
	}

	// Check for duplicates (warning only, doesn't block creation)
	var duplicateWarning *DuplicateWarning
	duplicate, err := h.cardService.CheckDuplicate(c.Request().Context(), req.CardNumber, user.ID, nil)
	if err != nil {
		c.Logger().Warnf("Failed to check duplicate: %v", err)
		// Don't fail the request, just log
	}
	if duplicate != nil {
		duplicateWarning = &DuplicateWarning{
			HasDuplicate:   true,
			MerchantName:   duplicate.MerchantName,
			ResourceNumber: duplicate.CardNumber,
			ExistingID:     duplicate.ID.String(),
		}
		c.Logger().Infof("Duplicate card detected: existing card %s", duplicate.ID)
	}

	// Create card
	card := &models.Card{
		UserID:       &user.ID,
		MerchantID:   merchantID,
		MerchantName: merchantName,
		Program:      stringOrDefault(req.Program, ""),
		CardNumber:   req.CardNumber,
		BarcodeType:  barcodeType,
		Status:       status,
		Notes:        stringOrDefault(req.Notes, ""),
	}

	if err := h.cardService.CreateCard(c.Request().Context(), card); err != nil {
		c.Logger().Errorf("Failed to create card: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to create card",
		})
	}

	c.Logger().Infof("Card created successfully: %s", card.ID)

	// Handle optional sharing on creation
	if req.ShareWithEmail != nil && *req.ShareWithEmail != "" {
		email := strings.ToLower(strings.TrimSpace(*req.ShareWithEmail))
		sharedUser, err := h.userService.GetUserByEmail(c.Request().Context(), email)
		if err == nil && sharedUser.ID != user.ID {
			// Create share with specified permissions
			canEdit := req.ShareCanEdit != nil && *req.ShareCanEdit
			canDelete := req.ShareCanDelete != nil && *req.ShareCanDelete
			if err := h.shareService.CreateCardShare(c.Request().Context(), user.ID, card.ID, sharedUser.ID, canEdit, canDelete); err != nil {
				c.Logger().Warnf("Failed to create share for card %s: %v", card.ID, err)
			} else {
				c.Logger().Infof("Share created for card %s with user %s", card.ID, sharedUser.ID)
			}
		}
		// Silently ignore errors during share creation - card is still created
	}

	// Reload with merchant relation
	card, err = h.cardService.GetCard(c.Request().Context(), card.ID)
	if err != nil {
		c.Logger().Errorf("Failed to reload card %s after creation: %v", card.ID, err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Card created but failed to reload",
		})
	}

	cardDTO := ToCardDTO(card, false)
	permDTO := PermissionDTO{
		CanView:   true,
		CanEdit:   true,
		CanDelete: true,
		IsOwner:   true,
	}
	cardDTO.Permissions = &permDTO

	return c.JSON(http.StatusCreated, CardDetailResponse{
		Card:             cardDTO,
		Permissions:      permDTO,
		DuplicateWarning: duplicateWarning,
	})
}

// Update updates a card
// PUT /api/v1/cards/:id
func (h *CardsHandler) Update(c echo.Context) error {
	user := c.Get("current_user").(*models.User)

	cardID, err := parseResourceID(c, "card")
	if err != nil {
		return err
	}

	// Check authorization
	perms, err := h.authzService.CheckCardAccess(c.Request().Context(), user.ID, cardID)
	if err != nil || !perms.CanEdit {
		return c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "You don't have permission to edit this card",
		})
	}

	var req CardUpdateRequest
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
	if err := validateStringLengthPtr(req.Notes, "notes"); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "field_too_long", Message: err.Error()})
	}
	if err := validateStringLengthPtr(req.Program, "program"); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "field_too_long", Message: err.Error()})
	}

	// Get existing card
	card, err := h.cardService.GetCard(c.Request().Context(), cardID)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "Card not found",
		})
	}

	// Apply updates
	if req.MerchantID != nil {
		mid, name, err := resolveMerchantUpdate(c, h.merchantService, *req.MerchantID)
		if err != nil {
			return err
		}
		card.MerchantID = mid
		card.MerchantName = name
		if mid != nil {
			// CRITICAL: Clear the Merchant relation to prevent GORM from overwriting MerchantID
			// from the old preloaded relation during Save()
			card.Merchant = nil
		}
	}
	if req.Program != nil {
		card.Program = *req.Program
	}
	if req.CardNumber != nil {
		card.CardNumber = *req.CardNumber
	}
	if req.BarcodeType != nil {
		if err := validateEnum(c, *req.BarcodeType, validBarcodeTypes, "barcode_type"); err != nil {
			return err
		}
		card.BarcodeType = *req.BarcodeType
	}
	if req.Notes != nil {
		card.Notes = *req.Notes
	}
	if req.Status != nil {
		if err := validateEnum(c, *req.Status, validStatuses, "status"); err != nil {
			return err
		}
		card.Status = *req.Status
	}

	// Check for duplicates if card number changed (warning only)
	var duplicateWarning *DuplicateWarning
	if req.CardNumber != nil && *req.CardNumber != "" {
		duplicate, err := h.cardService.CheckDuplicate(c.Request().Context(), card.CardNumber, user.ID, &cardID)
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
			c.Logger().Infof("Duplicate card detected during update: existing card %s", duplicate.ID)
		}
	}

	// Debug: Log before update
	c.Logger().Infof("UPDATING Card %s with MerchantID: %v, MerchantName: %s", cardID, card.MerchantID, card.MerchantName)

	if err := h.cardService.UpdateCard(c.Request().Context(), card); err != nil {
		c.Logger().Errorf("Update failed: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to update card",
		})
	}

	c.Logger().Infof("Update successful, reloading card %s", cardID)

	// Reload with relations
	card, err = h.cardService.GetCard(c.Request().Context(), cardID)
	if err != nil {
		c.Logger().Errorf("Failed to reload card %s after update: %v", cardID, err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Card updated but failed to reload",
		})
	}
	isFavorite, err := h.favoriteService.IsFavorite(c.Request().Context(), user.ID, "card", cardID)
	if err != nil {
		c.Logger().Warnf("Failed to check favorite status for card %s: %v", cardID, err)
	}

	// Debug: Log merchant info
	if card.Merchant != nil {
		c.Logger().Infof("Card %s has merchant %s (%s)", card.ID, card.Merchant.Name, card.Merchant.ID)
	} else if card.MerchantID != nil {
		c.Logger().Warnf("Card %s has MerchantID %s but Merchant is nil", card.ID, *card.MerchantID)
	} else {
		c.Logger().Infof("Card %s has no merchant", card.ID)
	}

	cardDTO := ToCardDTO(card, isFavorite)
	permDTO := PermissionDTO{
		CanView:   perms.CanView,
		CanEdit:   perms.CanEdit,
		CanDelete: perms.CanDelete,
		IsOwner:   perms.IsOwner,
	}
	cardDTO.Permissions = &permDTO

	// Get shares if owner
	var shares []ShareDTO
	if perms.IsOwner {
		cardShares, err := h.shareService.GetCardShares(c.Request().Context(), cardID)
		if err != nil {
			c.Logger().Warnf("Failed to load shares for card %s: %v", cardID, err)
		}
		shares = ToCardShareDTOs(cardShares)
	}

	return c.JSON(http.StatusOK, CardDetailResponse{
		Card:             cardDTO,
		Permissions:      permDTO,
		Shares:           shares,
		DuplicateWarning: duplicateWarning,
	})
}

// Delete deletes a card
// DELETE /api/v1/cards/:id
func (h *CardsHandler) Delete(c echo.Context) error {
	return handleResourceDelete(c, "card", h.authzService.CheckCardAccess, h.cardService.DeleteCard)
}

// ToggleFavorite toggles favorite status
// POST /api/v1/cards/:id/favorite
func (h *CardsHandler) ToggleFavorite(c echo.Context) error {
	return handleResourceToggleFavorite(c, "card", h.authzService.CheckCardAccess, h.favoriteService)
}

// CreateShare creates a new share
// POST /api/v1/cards/:id/share
func (h *CardsHandler) CreateShare(c echo.Context) error {
	user := c.Get("current_user").(*models.User)

	cardID, err := parseResourceID(c, "card")
	if err != nil {
		return err
	}

	// Check authorization - only owner can share
	perms, err := h.authzService.CheckCardAccess(c.Request().Context(), user.ID, cardID)
	if err != nil || !perms.IsOwner {
		return c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "Only the owner can share this card",
		})
	}

	var req ShareCreateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Validate and find user by email
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

	// Create share
	canEdit := req.CanEdit != nil && *req.CanEdit
	canDelete := req.CanDelete != nil && *req.CanDelete

	if err := h.shareService.CreateCardShare(c.Request().Context(), user.ID, cardID, sharedUser.ID, canEdit, canDelete); err != nil {
		c.Logger().Errorf("Failed to create card share for card %s: %v", cardID, err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to share card",
		})
	}

	// Return updated shares list
	shares, err := h.shareService.GetCardShares(c.Request().Context(), cardID)
	if err != nil {
		c.Logger().Warnf("Failed to load shares for card %s: %v", cardID, err)
	}

	return c.JSON(http.StatusCreated, map[string]any{
		"message": "Card shared successfully",
		"shares":  ToCardShareDTOs(shares),
	})
}

// UpdateShare updates share permissions
// PATCH /api/v1/cards/:id/share/:sharedWithID
func (h *CardsHandler) UpdateShare(c echo.Context) error {
	user := c.Get("current_user").(*models.User)

	cardID, err := parseResourceID(c, "card")
	if err != nil {
		return err
	}

	sharedWithID, err := parseUUIDParam(c, "sharedWithID", "invalid_user_id", "Invalid user ID")
	if err != nil {
		return err
	}

	// Check authorization - only owner can update shares
	perms, err := h.authzService.CheckCardAccess(c.Request().Context(), user.ID, cardID)
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

	// Add audit context (user ID, IP address, user agent) for audit logging
	ctx := audit.AddAuditContextToContext(c.Request().Context(), user.ID, c.RealIP(), c.Request().UserAgent())

	if err := h.shareService.UpdateCardShare(ctx, user.ID, cardID, sharedWithID, canEdit, canDelete); err != nil {
		c.Logger().Errorf("Failed to update card share for card %s, user %s: %v", cardID, sharedWithID, err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to update share permissions",
		})
	}

	// Audit log for permission change
	resourceData := map[string]any{
		"shared_with_id": sharedWithID.String(),
		"can_edit":       canEdit,
		"can_delete":     canDelete,
	}
	resourceDataJSON, _ := json.Marshal(resourceData)
	auditLog := &models.AuditLog{
		UserID:       &user.ID,
		Action:       "share_permission_update",
		ResourceType: "card_shares",
		ResourceID:   cardID,
		ResourceData: string(resourceDataJSON),
		IPAddress:    c.RealIP(),
		UserAgent:    c.Request().UserAgent(),
	}
	_ = h.adminService.CreateAuditLog(c.Request().Context(), auditLog)

	// Return updated shares list
	shares, err := h.shareService.GetCardShares(c.Request().Context(), cardID)
	if err != nil {
		c.Logger().Warnf("Failed to load shares for card %s: %v", cardID, err)
	}

	return c.JSON(http.StatusOK, map[string]any{
		"message": "Share permissions updated successfully",
		"shares":  ToCardShareDTOs(shares),
	})
}

// DeleteShare removes a share
// DELETE /api/v1/cards/:id/share/:sharedWithID
func (h *CardsHandler) DeleteShare(c echo.Context) error {
	return handleResourceDeleteShare(c, "card", h.authzService.CheckCardAccess, h.shareService.DeleteCardShare)
}

// Transfer transfers ownership
// POST /api/v1/cards/:id/transfer
func (h *CardsHandler) Transfer(c echo.Context) error {
	return handleResourceTransfer(c, "card", h.authzService.CheckCardAccess, h.transferService.TransferCardOwnership, h.userService)
}

// stringOrDefault returns the dereferenced string or default if nil
func stringOrDefault(s *string, defaultValue string) string {
	if s == nil {
		return defaultValue
	}
	return *s
}
