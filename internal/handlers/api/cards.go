// Package api contains JSON API handlers for cards.
package api //nolint:revive // "api" is a meaningful package name for API handlers

//

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"savvy/internal/audit"
	"savvy/internal/logsafe"
	"savvy/internal/models"
	"savvy/internal/services"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"gorm.io/gorm"
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
//
//	@Summary		List cards
//	@Description	Returns all cards the user owns or has been given access to. Omitting both pagination parameters returns the full list without a pagination block.
//	@Tags			cards
//	@Produce		json
//	@Param			page		query		int	false	"Page number, 1-based. Enables pagination."
//	@Param			per_page	query		int	false	"Items per page, capped at 100."	default(25)
//	@Success		200			{object}	CardListResponse
//	@Failure		401			{object}	ErrorResponse	"Not authenticated"
//	@Failure		500			{object}	ErrorResponse	"Failed to load cards"
//	@Security		cookieAuth
//	@Router			/cards [get]
func (h *CardsHandler) List(c *echo.Context) error {
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
//
//	@Summary		Get a card
//	@Description	Returns a single card including the caller's permissions. Shares are only included when the caller owns the card.
//	@Tags			cards
//	@Produce		json
//	@Param			id	path		string	true	"Card ID"	format(uuid)
//	@Success		200	{object}	CardDetailResponse
//	@Failure		400	{object}	ErrorResponse	"Invalid card ID"
//	@Failure		401	{object}	ErrorResponse	"Not authenticated"
//	@Failure		403	{object}	ErrorResponse	"No access to this card"
//	@Failure		404	{object}	ErrorResponse	"Card not found"
//	@Security		cookieAuth
//	@Router			/cards/{id} [get]
func (h *CardsHandler) Show(c *echo.Context) error {
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
		slog.WarnContext(c.Request().Context(), "failed to check favorite status", "resource_type", "card", "resource_id", cardID, "error", err)
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
			slog.WarnContext(c.Request().Context(), "failed to load shares", "resource_type", "card", "resource_id", cardID, "error", err)
		}
		shares = ToCardShareDTOs(cardShares)
	}

	return c.JSON(http.StatusOK, CardDetailResponse{
		Card:        cardDTO,
		Permissions: permDTO,
		Shares:      shares,
	})
}

// Create creates a new card
// POST /api/v1/cards
//
//	@Summary		Create a card
//	@Description	Creates a card for the current user. Either merchant_id or new_merchant_name must be set. A card number already owned by the user is rejected with 409; a card number only shared with the user is allowed and reported back as an advisory duplicate warning.
//	@Tags			cards
//	@Accept			json
//	@Produce		json
//	@Param			X-CSRF-Token	header		string					true	"CSRF token"
//	@Param			card			body		CardCreateRequest		true	"Card to create"
//	@Success		201				{object}	CardDetailResponse
//	@Failure		400				{object}	ErrorResponse			"Invalid or missing fields"
//	@Failure		401				{object}	ErrorResponse			"Not authenticated"
//	@Failure		403				{object}	ErrorResponse			"Invalid CSRF token"
//	@Failure		409				{object}	DuplicateErrorResponse	"Card number already exists, possibly soft-deleted and restorable"
//	@Failure		500				{object}	ErrorResponse			"Failed to create card"
//	@Security		cookieAuth
//	@Router			/cards [post]
func (h *CardsHandler) Create(c *echo.Context) error {
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
			slog.ErrorContext(c.Request().Context(), "failed to parse merchant ID", "error", err)
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_merchant_id",
				Message: "Invalid merchant ID",
			})
		}
		merchantID = &mid

		// Get merchant name
		merchant, err := h.merchantService.GetMerchantByID(c.Request().Context(), mid)
		if err != nil {
			slog.ErrorContext(c.Request().Context(), "failed to get merchant", "merchant_id", mid, "error", err)
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
			slog.ErrorContext(c.Request().Context(), "failed to create merchant", "name", logsafe.String(*req.NewMerchantName), "error", logsafe.Error(err))
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

	// Check for duplicates — blocks creation to prevent DB unique constraint violation
	duplicate, err := h.cardService.CheckDuplicate(c.Request().Context(), req.CardNumber, user.ID, nil)
	if err != nil {
		slog.ErrorContext(c.Request().Context(), "failed to check duplicate", "error", err)
		// Don't fail the request, just log — worst case the DB constraint catches it
	}
	if duplicate != nil {
		slog.InfoContext(c.Request().Context(), "duplicate card blocked", "existing_id", duplicate.ID)
		return c.JSON(http.StatusConflict, DuplicateErrorResponse{
			Error:   "duplicate_barcode",
			Message: "A card with this number already exists",
			Duplicate: &DuplicateWarning{
				HasDuplicate:   true,
				MerchantName:   duplicate.MerchantName,
				ResourceNumber: duplicate.CardNumber,
				ExistingID:     duplicate.ID.String(),
			},
		})
	}

	// Soft-deleted twin owned by this user → offer restore instead of a hard failure.
	// Checked before the shared-duplicate advisory: restoring one's own card is the
	// more actionable path and must not be shadowed by a same-numbered shared card.
	deletedDup, err := h.cardService.FindDeletedDuplicate(c.Request().Context(), req.CardNumber, user.ID)
	if err != nil {
		slog.ErrorContext(c.Request().Context(), "failed to check deleted duplicate", "error", err)
	}
	if deletedDup != nil {
		return c.JSON(http.StatusConflict, DuplicateErrorResponse{
			Error:   "duplicate_barcode",
			Message: "A soft-deleted card with this number exists and can be restored",
			Duplicate: &DuplicateWarning{
				HasDuplicate:   true,
				MerchantName:   deletedDup.MerchantName,
				ResourceNumber: deletedDup.CardNumber,
				ExistingID:     deletedDup.ID.String(),
				Deleted:        true,
			},
		})
	}

	// Shared duplicate: a card with this number was already shared with the user by
	// another owner. This is advisory only — a shared card belongs to a different
	// user_id and violates no unique constraint, so creation proceeds (family cards
	// are intentionally allowed). The warning is attached to the created response.
	var sharedWarning *DuplicateWarning
	sharedDup, err := h.cardService.CheckSharedDuplicate(c.Request().Context(), req.CardNumber, merchantID, user.ID)
	if err != nil {
		slog.ErrorContext(c.Request().Context(), "failed to check shared duplicate", "error", err)
	}
	if sharedDup != nil {
		sharedWarning = &DuplicateWarning{
			HasDuplicate:   true,
			MerchantName:   sharedDup.MerchantName,
			ResourceNumber: sharedDup.CardNumber,
			ExistingID:     sharedDup.ID.String(),
			IsShared:       true,
		}
		if sharedDup.User != nil {
			owner := ToUserDTO(sharedDup.User)
			sharedWarning.SharedBy = &owner
		}
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
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			// TOCTOU race: two concurrent requests with the same barcode — re-fetch to return details
			if existing, _ := h.cardService.CheckDuplicate(c.Request().Context(), req.CardNumber, user.ID, nil); existing != nil {
				return c.JSON(http.StatusConflict, DuplicateErrorResponse{
					Error:   "duplicate_barcode",
					Message: "A card with this number already exists",
					Duplicate: &DuplicateWarning{
						HasDuplicate:   true,
						MerchantName:   existing.MerchantName,
						ResourceNumber: existing.CardNumber,
						ExistingID:     existing.ID.String(),
					},
				})
			}
		}
		slog.ErrorContext(c.Request().Context(), "failed to create card", "error", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to create card",
		})
	}

	slog.InfoContext(c.Request().Context(), "card created", "card_id", card.ID)

	// Handle optional sharing on creation
	if req.ShareWithEmail != nil && *req.ShareWithEmail != "" {
		email := strings.ToLower(strings.TrimSpace(*req.ShareWithEmail))
		sharedUser, err := h.userService.GetUserByEmail(c.Request().Context(), email)
		if err == nil && sharedUser.ID != user.ID {
			// Create share with specified permissions
			canEdit := req.ShareCanEdit != nil && *req.ShareCanEdit
			canDelete := req.ShareCanDelete != nil && *req.ShareCanDelete
			if err := h.shareService.CreateCardShare(c.Request().Context(), user.ID, card.ID, sharedUser.ID, canEdit, canDelete); err != nil {
				slog.WarnContext(c.Request().Context(), "failed to create share on card creation", "card_id", card.ID, "error", err)
			} else {
				slog.InfoContext(c.Request().Context(), "share created on card creation", "card_id", card.ID, "shared_with_id", sharedUser.ID)
			}
		}
		// Silently ignore errors during share creation - card is still created
	}

	// Reload with merchant relation
	card, err = h.cardService.GetCard(c.Request().Context(), card.ID)
	if err != nil {
		slog.ErrorContext(c.Request().Context(), "failed to reload card after creation", "card_id", card.ID, "error", err)
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
		DuplicateWarning: sharedWarning, // advisory: same number already shared with the user
	})
}

// Update updates a card
// PATCH /api/v1/cards/:id
//
//	@Summary		Update a card
//	@Description	Applies a partial update; omitted fields keep their current value. A card number colliding with another card owned by the caller is reported as an advisory warning, not an error.
//	@Tags			cards
//	@Accept			json
//	@Produce		json
//	@Param			id				path		string				true	"Card ID"	format(uuid)
//	@Param			X-CSRF-Token	header		string				true	"CSRF token"
//	@Param			card			body		CardUpdateRequest	true	"Fields to update"
//	@Success		200				{object}	CardDetailResponse
//	@Failure		400				{object}	ErrorResponse	"Invalid request body or card ID"
//	@Failure		401				{object}	ErrorResponse	"Not authenticated"
//	@Failure		403				{object}	ErrorResponse	"No edit permission or invalid CSRF token"
//	@Failure		404				{object}	ErrorResponse	"Card not found"
//	@Failure		500				{object}	ErrorResponse	"Failed to update card"
//	@Security		cookieAuth
//	@Router			/cards/{id} [patch]
func (h *CardsHandler) Update(c *echo.Context) error {
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
			slog.WarnContext(c.Request().Context(), "failed to check duplicate", "error", err)
		}
		if duplicate != nil {
			duplicateWarning = &DuplicateWarning{
				HasDuplicate:   true,
				MerchantName:   duplicate.MerchantName,
				ResourceNumber: duplicate.CardNumber,
				ExistingID:     duplicate.ID.String(),
			}
			slog.InfoContext(c.Request().Context(), "duplicate card detected during update", "existing_id", duplicate.ID)
		}
	}

	slog.DebugContext(c.Request().Context(), "updating card", "card_id", cardID, "merchant_id", card.MerchantID, "merchant_name", card.MerchantName)

	if err := h.cardService.UpdateCard(c.Request().Context(), card); err != nil {
		slog.ErrorContext(c.Request().Context(), "failed to update card", "card_id", cardID, "error", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to update card",
		})
	}

	// Reload with relations
	card, err = h.cardService.GetCard(c.Request().Context(), cardID)
	if err != nil {
		slog.ErrorContext(c.Request().Context(), "failed to reload card after update", "card_id", cardID, "error", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Card updated but failed to reload",
		})
	}
	isFavorite, err := h.favoriteService.IsFavorite(c.Request().Context(), user.ID, "card", cardID)
	if err != nil {
		slog.WarnContext(c.Request().Context(), "failed to check favorite status", "resource_type", "card", "resource_id", cardID, "error", err)
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
			slog.WarnContext(c.Request().Context(), "failed to load shares", "resource_type", "card", "resource_id", cardID, "error", err)
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
//
//	@Summary		Delete a card
//	@Description	Soft-deletes the card. A deleted card owned by the caller can be brought back via the restore endpoint.
//	@Tags			cards
//	@Produce		json
//	@Param			id				path		string	true	"Card ID"	format(uuid)
//	@Param			X-CSRF-Token	header		string	true	"CSRF token"
//	@Success		200				{object}	MessageResponse
//	@Failure		400				{object}	ErrorResponse	"Invalid card ID"
//	@Failure		401				{object}	ErrorResponse	"Not authenticated"
//	@Failure		403				{object}	ErrorResponse	"No delete permission or invalid CSRF token"
//	@Failure		500				{object}	ErrorResponse	"Failed to delete card"
//	@Security		cookieAuth
//	@Router			/cards/{id} [delete]
func (h *CardsHandler) Delete(c *echo.Context) error {
	return handleResourceDelete(c, "card", h.authzService.CheckCardAccess, h.cardService.DeleteCard)
}

// ToggleFavorite toggles favorite status
// POST /api/v1/cards/:id/favorite
//
//	@Summary		Toggle favorite
//	@Description	Flips the caller's favorite flag for this card and returns the resulting state.
//	@Tags			cards
//	@Produce		json
//	@Param			id				path		string	true	"Card ID"	format(uuid)
//	@Param			X-CSRF-Token	header		string	true	"CSRF token"
//	@Success		200				{object}	FavoriteResponse
//	@Failure		400				{object}	ErrorResponse	"Invalid card ID"
//	@Failure		401				{object}	ErrorResponse	"Not authenticated"
//	@Failure		403				{object}	ErrorResponse	"No access to this card or invalid CSRF token"
//	@Failure		500				{object}	ErrorResponse	"Failed to toggle favorite"
//	@Security		cookieAuth
//	@Router			/cards/{id}/favorite [post]
func (h *CardsHandler) ToggleFavorite(c *echo.Context) error {
	return handleResourceToggleFavorite(c, "card", h.authzService.CheckCardAccess, h.favoriteService)
}

// CreateShare creates a new share
// POST /api/v1/cards/:id/share
//
//	@Summary		Share a card
//	@Description	Shares the card with one or more recipients by email, all with the same permissions. Partial success returns 201 with the failures listed; if every recipient fails the status is 422.
//	@Tags			cards
//	@Accept			json
//	@Produce		json
//	@Param			id				path		string				true	"Card ID"	format(uuid)
//	@Param			X-CSRF-Token	header		string				true	"CSRF token"
//	@Param			share			body		ShareCreateRequest	true	"Recipients and permissions"
//	@Success		201				{object}	ShareCreateResponse
//	@Failure		400				{object}	ErrorResponse		"Invalid request body or card ID"
//	@Failure		401				{object}	ErrorResponse		"Not authenticated"
//	@Failure		403				{object}	ErrorResponse		"Only the owner can share, or invalid CSRF token"
//	@Failure		422				{object}	ShareCreateResponse	"Every recipient failed"
//	@Security		cookieAuth
//	@Router			/cards/{id}/share [post]
func (h *CardsHandler) CreateShare(c *echo.Context) error {
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

	canEdit := req.CanEdit != nil && *req.CanEdit
	canDelete := req.CanDelete != nil && *req.CanDelete

	return handleResourceMultiShare(c, "card", cardID, req, h.userService,
		func(ctx context.Context, sharedWithID uuid.UUID) error {
			return h.shareService.CreateCardShare(ctx, user.ID, cardID, sharedWithID, canEdit, canDelete)
		},
		func(ctx context.Context) ([]ShareDTO, error) {
			shares, err := h.shareService.GetCardShares(ctx, cardID)
			return ToCardShareDTOs(shares), err
		},
	)
}

// UpdateShare updates share permissions
// PATCH /api/v1/cards/:id/share/:sharedWithID
//
//	@Summary		Update share permissions
//	@Description	Replaces the permissions of an existing share. Omitted permission flags are treated as false, and the change is written to the audit log.
//	@Tags			cards
//	@Accept			json
//	@Produce		json
//	@Param			id				path		string				true	"Card ID"			format(uuid)
//	@Param			sharedWithID	path		string				true	"Recipient user ID"	format(uuid)
//	@Param			X-CSRF-Token	header		string				true	"CSRF token"
//	@Param			permissions		body		ShareUpdateRequest	true	"New permissions"
//	@Success		200				{object}	SharesResponse
//	@Failure		400				{object}	ErrorResponse	"Invalid request body, card ID or user ID"
//	@Failure		401				{object}	ErrorResponse	"Not authenticated"
//	@Failure		403				{object}	ErrorResponse	"Only the owner can update shares, or invalid CSRF token"
//	@Failure		500				{object}	ErrorResponse	"Failed to update share permissions"
//	@Security		cookieAuth
//	@Router			/cards/{id}/share/{sharedWithID} [patch]
func (h *CardsHandler) UpdateShare(c *echo.Context) error {
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
		slog.ErrorContext(c.Request().Context(), "failed to update card share", "card_id", cardID, "shared_with_id", sharedWithID, "error", err)
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
		slog.WarnContext(c.Request().Context(), "failed to load shares", "resource_type", "card", "resource_id", cardID, "error", err)
	}

	return c.JSON(http.StatusOK, SharesResponse{
		Message: "Share permissions updated successfully",
		Shares:  ToCardShareDTOs(shares),
	})
}

// DeleteShare removes a share
// DELETE /api/v1/cards/:id/share/:sharedWithID
//
//	@Summary		Remove a share
//	@Description	Revokes one recipient's access to the card.
//	@Tags			cards
//	@Produce		json
//	@Param			id				path		string	true	"Card ID"			format(uuid)
//	@Param			sharedWithID	path		string	true	"Recipient user ID"	format(uuid)
//	@Param			X-CSRF-Token	header		string	true	"CSRF token"
//	@Success		200				{object}	MessageResponse
//	@Failure		400				{object}	ErrorResponse	"Invalid card ID or user ID"
//	@Failure		401				{object}	ErrorResponse	"Not authenticated"
//	@Failure		403				{object}	ErrorResponse	"Only the owner can remove shares, or invalid CSRF token"
//	@Failure		500				{object}	ErrorResponse	"Failed to remove share"
//	@Security		cookieAuth
//	@Router			/cards/{id}/share/{sharedWithID} [delete]
func (h *CardsHandler) DeleteShare(c *echo.Context) error {
	return handleResourceDeleteShare(c, "card", h.authzService.CheckCardAccess, h.shareService.DeleteCardShare)
}

// DeleteAllShares removes all shares for a card
// DELETE /api/v1/cards/:id/shares
//
//	@Summary		Remove all shares
//	@Description	Revokes access for every recipient of the card at once.
//	@Tags			cards
//	@Produce		json
//	@Param			id				path		string	true	"Card ID"	format(uuid)
//	@Param			X-CSRF-Token	header		string	true	"CSRF token"
//	@Success		200				{object}	MessageResponse
//	@Failure		400				{object}	ErrorResponse	"Invalid card ID"
//	@Failure		401				{object}	ErrorResponse	"Not authenticated"
//	@Failure		403				{object}	ErrorResponse	"Only the owner can remove shares, or invalid CSRF token"
//	@Failure		500				{object}	ErrorResponse	"Failed to remove shares"
//	@Security		cookieAuth
//	@Router			/cards/{id}/shares [delete]
func (h *CardsHandler) DeleteAllShares(c *echo.Context) error {
	return handleResourceDeleteAllShares(c, "card", h.authzService.CheckCardAccess, h.shareService.DeleteAllCardShares)
}

// Transfer transfers ownership
// POST /api/v1/cards/:id/transfer
//
//	@Summary		Transfer ownership
//	@Description	Hands the card over to another user by email. Only the current owner can transfer, and the new owner must already have an account.
//	@Tags			cards
//	@Accept			json
//	@Produce		json
//	@Param			id				path		string			true	"Card ID"	format(uuid)
//	@Param			X-CSRF-Token	header		string			true	"CSRF token"
//	@Param			transfer		body		TransferRequest	true	"New owner"
//	@Success		200				{object}	MessageResponse
//	@Failure		400				{object}	ErrorResponse	"Invalid request body, card ID or unknown recipient"
//	@Failure		401				{object}	ErrorResponse	"Not authenticated"
//	@Failure		403				{object}	ErrorResponse	"Only the owner can transfer, or invalid CSRF token"
//	@Failure		500				{object}	ErrorResponse	"Failed to transfer card"
//	@Security		cookieAuth
//	@Router			/cards/{id}/transfer [post]
func (h *CardsHandler) Transfer(c *echo.Context) error {
	return handleResourceTransfer(c, "card", h.authzService.CheckCardAccess, h.transferService.TransferCardOwnership, h.userService)
}

// Restore restores a soft-deleted card owned by the current user
// POST /api/v1/cards/:id/restore
//
//	@Summary		Restore a deleted card
//	@Description	Brings a soft-deleted card back. Only the original owner can restore, and the restored card always comes back with owner permissions.
//	@Tags			cards
//	@Produce		json
//	@Param			id				path		string	true	"Card ID"	format(uuid)
//	@Param			X-CSRF-Token	header		string	true	"CSRF token"
//	@Success		200				{object}	CardDetailResponse
//	@Failure		400				{object}	ErrorResponse	"Invalid card ID"
//	@Failure		401				{object}	ErrorResponse	"Not authenticated"
//	@Failure		403				{object}	ErrorResponse	"Invalid CSRF token"
//	@Failure		404				{object}	ErrorResponse	"No restorable card found"
//	@Failure		500				{object}	ErrorResponse	"Failed to restore card"
//	@Security		cookieAuth
//	@Router			/cards/{id}/restore [post]
func (h *CardsHandler) Restore(c *echo.Context) error {
	user := c.Get("current_user").(*models.User)

	cardID, err := parseResourceID(c, "card")
	if err != nil {
		return err
	}

	restored, err := h.cardService.RestoreCard(c.Request().Context(), cardID, user.ID)
	if err != nil {
		slog.ErrorContext(c.Request().Context(), "failed to restore card", "card_id", cardID, "error", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to restore card",
		})
	}
	if restored == nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "No restorable card found",
		})
	}

	isFavorite, _ := h.favoriteService.IsFavorite(c.Request().Context(), user.ID, "card", cardID)
	dto := ToCardDTO(restored, isFavorite)
	// A restored card is always owned by the caller — return owner permissions.
	perms := PermissionDTO{CanView: true, CanEdit: true, CanDelete: true, IsOwner: true}
	dto.Permissions = &perms
	return c.JSON(http.StatusOK, CardDetailResponse{Card: dto, Permissions: perms})
}

// stringOrDefault returns the dereferenced string or default if nil
func stringOrDefault(s *string, defaultValue string) string {
	if s == nil {
		return defaultValue
	}
	return *s
}
