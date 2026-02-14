package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"savvy/internal/mocks"
	"savvy/internal/models"
	"savvy/internal/services"
)

// ==================== Helper Functions ====================

func setupBatchTest() (
	*BatchHandler,
	*mocks.MockCardServiceInterface,
	*mocks.MockVoucherServiceInterface,
	*mocks.MockGiftCardServiceInterface,
	*mocks.MockAuthzServiceInterface,
	*mocks.MockShareServiceInterface,
	*mocks.MockTransferServiceInterface,
	*mocks.MockUserServiceInterface,
	*mocks.MockExportServiceInterface,
) {
	mockCard := new(mocks.MockCardServiceInterface)
	mockVoucher := new(mocks.MockVoucherServiceInterface)
	mockGiftCard := new(mocks.MockGiftCardServiceInterface)
	mockAuthz := new(mocks.MockAuthzServiceInterface)
	mockShare := new(mocks.MockShareServiceInterface)
	mockTransfer := new(mocks.MockTransferServiceInterface)
	mockUser := new(mocks.MockUserServiceInterface)
	mockExport := new(mocks.MockExportServiceInterface)

	handler := NewBatchHandler(mockCard, mockVoucher, mockGiftCard, mockAuthz, mockShare, mockTransfer, mockUser, mockExport)
	return handler, mockCard, mockVoucher, mockGiftCard, mockAuthz, mockShare, mockTransfer, mockUser, mockExport
}

// ==================== Batch Delete Tests ====================

func TestBatchHandler_DeleteCards_Success(t *testing.T) {
	handler, mockCard, _, _, mockAuthz, _, _, _, _ := setupBatchTest()

	cardID1 := uuid.New()
	cardID2 := uuid.New()
	body := `{"ids":["` + cardID1.String() + `","` + cardID2.String() + `"]}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/cards/batch/delete", body)
	user := createTestUser()
	c.Set("current_user", user)

	perms := &services.ResourcePermissions{
		CanView: true, CanEdit: true, CanDelete: true, IsOwner: true,
	}

	mockAuthz.On("CheckCardAccess", mock.Anything, user.ID, cardID1).Return(perms, nil)
	mockAuthz.On("CheckCardAccess", mock.Anything, user.ID, cardID2).Return(perms, nil)
	mockCard.On("DeleteCard", mock.Anything, cardID1).Return(nil)
	mockCard.On("DeleteCard", mock.Anything, cardID2).Return(nil)

	err := handler.DeleteCards(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response BatchResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, 2, response.SuccessCount)
	assert.Empty(t, response.Failed)
	mockAuthz.AssertExpectations(t)
	mockCard.AssertExpectations(t)
}

func TestBatchHandler_DeleteCards_Forbidden(t *testing.T) {
	handler, _, _, _, mockAuthz, _, _, _, _ := setupBatchTest()

	cardID1 := uuid.New()
	cardID2 := uuid.New()
	body := `{"ids":["` + cardID1.String() + `","` + cardID2.String() + `"]}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/cards/batch/delete", body)
	user := createTestUser()
	c.Set("current_user", user)

	permsOK := &services.ResourcePermissions{
		CanView: true, CanEdit: true, CanDelete: true, IsOwner: true,
	}
	permsDenied := &services.ResourcePermissions{
		CanView: true, CanEdit: false, CanDelete: false, IsOwner: false,
	}

	// First card passes, second card fails permission check
	mockAuthz.On("CheckCardAccess", mock.Anything, user.ID, cardID1).Return(permsOK, nil)
	mockAuthz.On("CheckCardAccess", mock.Anything, user.ID, cardID2).Return(permsDenied, nil)

	err := handler.DeleteCards(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "forbidden", response.Error)
	assert.Contains(t, response.Message, "permission to delete")
	mockAuthz.AssertExpectations(t)
}

func TestBatchHandler_DeleteCards_EmptyIDs(t *testing.T) {
	handler, _, _, _, _, _, _, _, _ := setupBatchTest()

	body := `{"ids":[]}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/cards/batch/delete", body)
	user := createTestUser()
	c.Set("current_user", user)

	err := handler.DeleteCards(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "invalid_request", response.Error)
}

func TestBatchHandler_DeleteCards_TooManyIDs(t *testing.T) {
	handler, _, _, _, _, _, _, _, _ := setupBatchTest()

	// Generate 51 UUIDs (exceeds maxBatchSize of 50)
	ids := make([]string, 51)
	for i := range ids {
		ids[i] = `"` + uuid.New().String() + `"`
	}
	body := `{"ids":[` + strings.Join(ids, ",") + `]}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/cards/batch/delete", body)
	user := createTestUser()
	c.Set("current_user", user)

	err := handler.DeleteCards(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "invalid_request", response.Error)
}

func TestBatchHandler_DeleteCards_InvalidID(t *testing.T) {
	handler, _, _, _, _, _, _, _, _ := setupBatchTest()

	body := `{"ids":["not-a-valid-uuid"]}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/cards/batch/delete", body)
	user := createTestUser()
	c.Set("current_user", user)

	err := handler.DeleteCards(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "invalid_request", response.Error)
}

func TestBatchHandler_DeleteCards_PartialDeleteFailure(t *testing.T) {
	handler, mockCard, _, _, mockAuthz, _, _, _, _ := setupBatchTest()

	cardID1 := uuid.New()
	cardID2 := uuid.New()
	body := `{"ids":["` + cardID1.String() + `","` + cardID2.String() + `"]}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/cards/batch/delete", body)
	user := createTestUser()
	c.Set("current_user", user)

	perms := &services.ResourcePermissions{
		CanView: true, CanEdit: true, CanDelete: true, IsOwner: true,
	}

	// Both cards pass permission check
	mockAuthz.On("CheckCardAccess", mock.Anything, user.ID, cardID1).Return(perms, nil)
	mockAuthz.On("CheckCardAccess", mock.Anything, user.ID, cardID2).Return(perms, nil)

	// First delete succeeds, second fails
	mockCard.On("DeleteCard", mock.Anything, cardID1).Return(nil)
	mockCard.On("DeleteCard", mock.Anything, cardID2).Return(errors.New("database error"))

	err := handler.DeleteCards(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response BatchResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, 1, response.SuccessCount)
	assert.Len(t, response.Failed, 1)
	assert.Equal(t, cardID2.String(), response.Failed[0].ID)
	assert.Contains(t, response.Failed[0].Error, "database error")
	mockAuthz.AssertExpectations(t)
	mockCard.AssertExpectations(t)
}

func TestBatchHandler_DeleteCards_AuthzError(t *testing.T) {
	handler, _, _, _, mockAuthz, _, _, _, _ := setupBatchTest()

	cardID1 := uuid.New()
	body := `{"ids":["` + cardID1.String() + `"]}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/cards/batch/delete", body)
	user := createTestUser()
	c.Set("current_user", user)

	// CheckCardAccess returns error (resource not found, etc.)
	mockAuthz.On("CheckCardAccess", mock.Anything, user.ID, cardID1).
		Return((*services.ResourcePermissions)(nil), errors.New("not found"))

	err := handler.DeleteCards(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockAuthz.AssertExpectations(t)
}

func TestBatchHandler_DeleteCards_InvalidJSON(t *testing.T) {
	handler, _, _, _, _, _, _, _, _ := setupBatchTest()

	c, rec := createTestContext(http.MethodPost, "/api/v1/cards/batch/delete", "invalid-json")
	user := createTestUser()
	c.Set("current_user", user)

	err := handler.DeleteCards(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "invalid_request", response.Error)
}

func TestBatchHandler_DeleteVouchers_Success(t *testing.T) {
	handler, _, mockVoucher, _, mockAuthz, _, _, _, _ := setupBatchTest()

	voucherID1 := uuid.New()
	voucherID2 := uuid.New()
	body := `{"ids":["` + voucherID1.String() + `","` + voucherID2.String() + `"]}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/vouchers/batch/delete", body)
	user := createTestUser()
	c.Set("current_user", user)

	perms := &services.ResourcePermissions{
		CanView: true, CanEdit: true, CanDelete: true, IsOwner: true,
	}

	mockAuthz.On("CheckVoucherAccess", mock.Anything, user.ID, voucherID1).Return(perms, nil)
	mockAuthz.On("CheckVoucherAccess", mock.Anything, user.ID, voucherID2).Return(perms, nil)
	mockVoucher.On("DeleteVoucher", mock.Anything, voucherID1).Return(nil)
	mockVoucher.On("DeleteVoucher", mock.Anything, voucherID2).Return(nil)

	err := handler.DeleteVouchers(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response BatchResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, 2, response.SuccessCount)
	assert.Empty(t, response.Failed)
	mockAuthz.AssertExpectations(t)
	mockVoucher.AssertExpectations(t)
}

func TestBatchHandler_DeleteGiftCards_Success(t *testing.T) {
	handler, _, _, mockGiftCard, mockAuthz, _, _, _, _ := setupBatchTest()

	gcID1 := uuid.New()
	gcID2 := uuid.New()
	body := `{"ids":["` + gcID1.String() + `","` + gcID2.String() + `"]}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/gift-cards/batch/delete", body)
	user := createTestUser()
	c.Set("current_user", user)

	perms := &services.ResourcePermissions{
		CanView: true, CanEdit: true, CanDelete: true, IsOwner: true,
	}

	mockAuthz.On("CheckGiftCardAccess", mock.Anything, user.ID, gcID1).Return(perms, nil)
	mockAuthz.On("CheckGiftCardAccess", mock.Anything, user.ID, gcID2).Return(perms, nil)
	mockGiftCard.On("DeleteGiftCard", mock.Anything, gcID1).Return(nil)
	mockGiftCard.On("DeleteGiftCard", mock.Anything, gcID2).Return(nil)

	err := handler.DeleteGiftCards(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response BatchResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, 2, response.SuccessCount)
	assert.Empty(t, response.Failed)
	mockAuthz.AssertExpectations(t)
	mockGiftCard.AssertExpectations(t)
}

// ==================== Batch Share Tests ====================

func TestBatchHandler_ShareCards_Success(t *testing.T) {
	handler, _, _, _, mockAuthz, mockShare, _, mockUser, _ := setupBatchTest()

	cardID1 := uuid.New()
	cardID2 := uuid.New()
	targetUserID := uuid.New()
	body := `{"ids":["` + cardID1.String() + `","` + cardID2.String() + `"],"email":"target@example.com","can_edit":true,"can_delete":false}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/cards/batch/share", body)
	user := createTestUser()
	c.Set("current_user", user)

	targetUser := &models.User{ID: targetUserID, Email: "target@example.com"}
	perms := &services.ResourcePermissions{
		CanView: true, CanEdit: true, CanDelete: true, IsOwner: true,
	}

	mockUser.On("GetUserByEmail", mock.Anything, "target@example.com").Return(targetUser, nil)
	mockAuthz.On("CheckCardAccess", mock.Anything, user.ID, cardID1).Return(perms, nil)
	mockAuthz.On("CheckCardAccess", mock.Anything, user.ID, cardID2).Return(perms, nil)
	mockShare.On("CreateCardShare", mock.Anything, user.ID, cardID1, targetUserID, true, false).Return(nil)
	mockShare.On("CreateCardShare", mock.Anything, user.ID, cardID2, targetUserID, true, false).Return(nil)

	err := handler.ShareCards(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response BatchResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, 2, response.SuccessCount)
	assert.Empty(t, response.Failed)
	mockUser.AssertExpectations(t)
	mockAuthz.AssertExpectations(t)
	mockShare.AssertExpectations(t)
}

func TestBatchHandler_ShareCards_MissingEmail(t *testing.T) {
	handler, _, _, _, _, _, _, _, _ := setupBatchTest()

	cardID := uuid.New()
	body := `{"ids":["` + cardID.String() + `"]}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/cards/batch/share", body)
	user := createTestUser()
	c.Set("current_user", user)

	err := handler.ShareCards(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "missing_email", response.Error)
}

func TestBatchHandler_ShareCards_SelfShare(t *testing.T) {
	handler, _, _, _, _, _, _, mockUser, _ := setupBatchTest()

	cardID := uuid.New()
	user := createTestUser()
	body := `{"ids":["` + cardID.String() + `"],"email":"` + user.Email + `"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/cards/batch/share", body)
	c.Set("current_user", user)

	// GetUserByEmail returns the same user
	mockUser.On("GetUserByEmail", mock.Anything, user.Email).Return(user, nil)

	err := handler.ShareCards(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "self_share", response.Error)
	assert.Contains(t, response.Message, "yourself")
	mockUser.AssertExpectations(t)
}

func TestBatchHandler_ShareCards_UserNotFound(t *testing.T) {
	handler, _, _, _, _, _, _, mockUser, _ := setupBatchTest()

	cardID := uuid.New()
	body := `{"ids":["` + cardID.String() + `"],"email":"unknown@example.com"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/cards/batch/share", body)
	user := createTestUser()
	c.Set("current_user", user)

	mockUser.On("GetUserByEmail", mock.Anything, "unknown@example.com").
		Return((*models.User)(nil), errors.New("not found"))

	err := handler.ShareCards(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "share_failed", response.Error)
	mockUser.AssertExpectations(t)
}

func TestBatchHandler_ShareCards_NotOwner(t *testing.T) {
	handler, _, _, _, mockAuthz, mockShare, _, mockUser, _ := setupBatchTest()

	cardID1 := uuid.New()
	cardID2 := uuid.New()
	targetUserID := uuid.New()
	body := `{"ids":["` + cardID1.String() + `","` + cardID2.String() + `"],"email":"target@example.com"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/cards/batch/share", body)
	user := createTestUser()
	c.Set("current_user", user)

	targetUser := &models.User{ID: targetUserID, Email: "target@example.com"}
	permsOwner := &services.ResourcePermissions{
		CanView: true, CanEdit: true, CanDelete: true, IsOwner: true,
	}
	permsNotOwner := &services.ResourcePermissions{
		CanView: true, CanEdit: false, CanDelete: false, IsOwner: false,
	}

	mockUser.On("GetUserByEmail", mock.Anything, "target@example.com").Return(targetUser, nil)
	mockAuthz.On("CheckCardAccess", mock.Anything, user.ID, cardID1).Return(permsOwner, nil)
	mockAuthz.On("CheckCardAccess", mock.Anything, user.ID, cardID2).Return(permsNotOwner, nil)
	mockShare.On("CreateCardShare", mock.Anything, user.ID, cardID1, targetUserID, false, false).Return(nil)

	err := handler.ShareCards(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response BatchResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, 1, response.SuccessCount)
	assert.Len(t, response.Failed, 1)
	assert.Equal(t, cardID2.String(), response.Failed[0].ID)
	assert.Contains(t, response.Failed[0].Error, "Not the owner")
	mockUser.AssertExpectations(t)
	mockAuthz.AssertExpectations(t)
	mockShare.AssertExpectations(t)
}

func TestBatchHandler_ShareCards_DuplicateIDs(t *testing.T) {
	handler, _, _, _, mockAuthz, mockShare, _, mockUser, _ := setupBatchTest()

	cardID := uuid.New()
	targetUserID := uuid.New()
	// Same ID twice in request - should be deduplicated
	body := `{"ids":["` + cardID.String() + `","` + cardID.String() + `"],"email":"target@example.com"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/cards/batch/share", body)
	user := createTestUser()
	c.Set("current_user", user)

	targetUser := &models.User{ID: targetUserID, Email: "target@example.com"}
	perms := &services.ResourcePermissions{
		CanView: true, CanEdit: true, CanDelete: true, IsOwner: true,
	}

	mockUser.On("GetUserByEmail", mock.Anything, "target@example.com").Return(targetUser, nil)
	mockAuthz.On("CheckCardAccess", mock.Anything, user.ID, cardID).Return(perms, nil)
	mockShare.On("CreateCardShare", mock.Anything, user.ID, cardID, targetUserID, false, false).Return(nil)

	err := handler.ShareCards(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response BatchResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	// After deduplication, only 1 unique ID should be processed
	assert.Equal(t, 1, response.SuccessCount)
	assert.Empty(t, response.Failed)
	mockUser.AssertExpectations(t)
	mockAuthz.AssertExpectations(t)
	mockShare.AssertExpectations(t)
}

func TestBatchHandler_ShareCards_EmptyIDs(t *testing.T) {
	handler, _, _, _, _, _, _, _, _ := setupBatchTest()

	body := `{"ids":[],"email":"target@example.com"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/cards/batch/share", body)
	user := createTestUser()
	c.Set("current_user", user)

	err := handler.ShareCards(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "invalid_request", response.Error)
}

func TestBatchHandler_ShareCards_ShareServiceError(t *testing.T) {
	handler, _, _, _, mockAuthz, mockShare, _, mockUser, _ := setupBatchTest()

	cardID := uuid.New()
	targetUserID := uuid.New()
	body := `{"ids":["` + cardID.String() + `"],"email":"target@example.com"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/cards/batch/share", body)
	user := createTestUser()
	c.Set("current_user", user)

	targetUser := &models.User{ID: targetUserID, Email: "target@example.com"}
	perms := &services.ResourcePermissions{
		CanView: true, CanEdit: true, CanDelete: true, IsOwner: true,
	}

	mockUser.On("GetUserByEmail", mock.Anything, "target@example.com").Return(targetUser, nil)
	mockAuthz.On("CheckCardAccess", mock.Anything, user.ID, cardID).Return(perms, nil)
	mockShare.On("CreateCardShare", mock.Anything, user.ID, cardID, targetUserID, false, false).
		Return(errors.New("duplicate share"))

	err := handler.ShareCards(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response BatchResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, 0, response.SuccessCount)
	assert.Len(t, response.Failed, 1)
	assert.Contains(t, response.Failed[0].Error, "duplicate share")
	mockUser.AssertExpectations(t)
	mockAuthz.AssertExpectations(t)
	mockShare.AssertExpectations(t)
}

func TestBatchHandler_ShareCards_EmailNormalization(t *testing.T) {
	handler, _, _, _, mockAuthz, mockShare, _, mockUser, _ := setupBatchTest()

	cardID := uuid.New()
	targetUserID := uuid.New()
	// Email with spaces and uppercase
	body := `{"ids":["` + cardID.String() + `"],"email":"  Target@Example.COM  "}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/cards/batch/share", body)
	user := createTestUser()
	c.Set("current_user", user)

	targetUser := &models.User{ID: targetUserID, Email: "target@example.com"}
	perms := &services.ResourcePermissions{
		CanView: true, CanEdit: true, CanDelete: true, IsOwner: true,
	}

	// The handler normalizes the email to lowercase and trimmed
	mockUser.On("GetUserByEmail", mock.Anything, "target@example.com").Return(targetUser, nil)
	mockAuthz.On("CheckCardAccess", mock.Anything, user.ID, cardID).Return(perms, nil)
	mockShare.On("CreateCardShare", mock.Anything, user.ID, cardID, targetUserID, false, false).Return(nil)

	err := handler.ShareCards(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockUser.AssertExpectations(t)
}

func TestBatchHandler_ShareVouchers_Success(t *testing.T) {
	handler, _, _, _, mockAuthz, mockShare, _, mockUser, _ := setupBatchTest()

	voucherID1 := uuid.New()
	voucherID2 := uuid.New()
	targetUserID := uuid.New()
	// Voucher shares are always read-only, can_edit/can_delete in request are ignored
	body := `{"ids":["` + voucherID1.String() + `","` + voucherID2.String() + `"],"email":"target@example.com","can_edit":true,"can_delete":true}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/vouchers/batch/share", body)
	user := createTestUser()
	c.Set("current_user", user)

	targetUser := &models.User{ID: targetUserID, Email: "target@example.com"}
	perms := &services.ResourcePermissions{
		CanView: true, CanEdit: true, CanDelete: true, IsOwner: true,
	}

	mockUser.On("GetUserByEmail", mock.Anything, "target@example.com").Return(targetUser, nil)
	mockAuthz.On("CheckVoucherAccess", mock.Anything, user.ID, voucherID1).Return(perms, nil)
	mockAuthz.On("CheckVoucherAccess", mock.Anything, user.ID, voucherID2).Return(perms, nil)
	// CreateVoucherShare only takes 4 args (always read-only, no permission flags)
	mockShare.On("CreateVoucherShare", mock.Anything, user.ID, voucherID1, targetUserID).Return(nil)
	mockShare.On("CreateVoucherShare", mock.Anything, user.ID, voucherID2, targetUserID).Return(nil)

	err := handler.ShareVouchers(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response BatchResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, 2, response.SuccessCount)
	assert.Empty(t, response.Failed)
	mockUser.AssertExpectations(t)
	mockAuthz.AssertExpectations(t)
	mockShare.AssertExpectations(t)
}

func TestBatchHandler_ShareGiftCards_Success(t *testing.T) {
	handler, _, _, _, mockAuthz, mockShare, _, mockUser, _ := setupBatchTest()

	gcID1 := uuid.New()
	gcID2 := uuid.New()
	targetUserID := uuid.New()
	body := `{"ids":["` + gcID1.String() + `","` + gcID2.String() + `"],"email":"target@example.com","can_edit":true,"can_delete":false,"can_edit_transactions":true}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/gift-cards/batch/share", body)
	user := createTestUser()
	c.Set("current_user", user)

	targetUser := &models.User{ID: targetUserID, Email: "target@example.com"}
	perms := &services.ResourcePermissions{
		CanView: true, CanEdit: true, CanDelete: true, IsOwner: true,
	}

	mockUser.On("GetUserByEmail", mock.Anything, "target@example.com").Return(targetUser, nil)
	mockAuthz.On("CheckGiftCardAccess", mock.Anything, user.ID, gcID1).Return(perms, nil)
	mockAuthz.On("CheckGiftCardAccess", mock.Anything, user.ID, gcID2).Return(perms, nil)
	// CreateGiftCardShare takes 7 args: ctx, callerID, resourceID, sharedWithID, canEdit, canDelete, canEditTx
	mockShare.On("CreateGiftCardShare", mock.Anything, user.ID, gcID1, targetUserID, true, false, true).Return(nil)
	mockShare.On("CreateGiftCardShare", mock.Anything, user.ID, gcID2, targetUserID, true, false, true).Return(nil)

	err := handler.ShareGiftCards(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response BatchResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, 2, response.SuccessCount)
	assert.Empty(t, response.Failed)
	mockUser.AssertExpectations(t)
	mockAuthz.AssertExpectations(t)
	mockShare.AssertExpectations(t)
}

// ==================== Batch Transfer Tests ====================

func TestBatchHandler_TransferCards_Success(t *testing.T) {
	handler, _, _, _, mockAuthz, _, mockTransfer, mockUser, _ := setupBatchTest()

	cardID1 := uuid.New()
	cardID2 := uuid.New()
	newOwnerID := uuid.New()
	body := `{"ids":["` + cardID1.String() + `","` + cardID2.String() + `"],"new_owner_email":"newowner@example.com"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/cards/batch/transfer", body)
	user := createTestUser()
	c.Set("current_user", user)

	newOwner := &models.User{ID: newOwnerID, Email: "newowner@example.com"}
	perms := &services.ResourcePermissions{
		CanView: true, CanEdit: true, CanDelete: true, IsOwner: true,
	}

	mockUser.On("GetUserByEmail", mock.Anything, "newowner@example.com").Return(newOwner, nil)
	mockAuthz.On("CheckCardAccess", mock.Anything, user.ID, cardID1).Return(perms, nil)
	mockAuthz.On("CheckCardAccess", mock.Anything, user.ID, cardID2).Return(perms, nil)
	mockTransfer.On("TransferCardOwnership", mock.Anything, cardID1, newOwnerID, user.ID).Return(nil)
	mockTransfer.On("TransferCardOwnership", mock.Anything, cardID2, newOwnerID, user.ID).Return(nil)

	err := handler.TransferCards(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response BatchResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, 2, response.SuccessCount)
	assert.Empty(t, response.Failed)
	mockUser.AssertExpectations(t)
	mockAuthz.AssertExpectations(t)
	mockTransfer.AssertExpectations(t)
}

func TestBatchHandler_TransferCards_MissingEmail(t *testing.T) {
	handler, _, _, _, _, _, _, _, _ := setupBatchTest()

	cardID := uuid.New()
	body := `{"ids":["` + cardID.String() + `"]}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/cards/batch/transfer", body)
	user := createTestUser()
	c.Set("current_user", user)

	err := handler.TransferCards(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "missing_email", response.Error)
	assert.Contains(t, response.Message, "New owner email is required")
}

func TestBatchHandler_TransferCards_SelfTransfer(t *testing.T) {
	handler, _, _, _, _, _, _, mockUser, _ := setupBatchTest()

	cardID := uuid.New()
	user := createTestUser()
	body := `{"ids":["` + cardID.String() + `"],"new_owner_email":"` + user.Email + `"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/cards/batch/transfer", body)
	c.Set("current_user", user)

	mockUser.On("GetUserByEmail", mock.Anything, user.Email).Return(user, nil)

	err := handler.TransferCards(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "self_transfer", response.Error)
	assert.Contains(t, response.Message, "yourself")
	mockUser.AssertExpectations(t)
}

func TestBatchHandler_TransferCards_NotOwner(t *testing.T) {
	handler, _, _, _, mockAuthz, _, mockTransfer, mockUser, _ := setupBatchTest()

	cardID1 := uuid.New()
	cardID2 := uuid.New()
	newOwnerID := uuid.New()
	body := `{"ids":["` + cardID1.String() + `","` + cardID2.String() + `"],"new_owner_email":"newowner@example.com"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/cards/batch/transfer", body)
	user := createTestUser()
	c.Set("current_user", user)

	newOwner := &models.User{ID: newOwnerID, Email: "newowner@example.com"}
	permsOwner := &services.ResourcePermissions{
		CanView: true, CanEdit: true, CanDelete: true, IsOwner: true,
	}
	permsNotOwner := &services.ResourcePermissions{
		CanView: true, CanEdit: false, CanDelete: false, IsOwner: false,
	}

	mockUser.On("GetUserByEmail", mock.Anything, "newowner@example.com").Return(newOwner, nil)
	// First card: owner, second card: not owner
	mockAuthz.On("CheckCardAccess", mock.Anything, user.ID, cardID1).Return(permsNotOwner, nil)
	mockAuthz.On("CheckCardAccess", mock.Anything, user.ID, cardID2).Return(permsOwner, nil)
	mockTransfer.On("TransferCardOwnership", mock.Anything, cardID2, newOwnerID, user.ID).Return(nil)

	err := handler.TransferCards(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response BatchResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, 1, response.SuccessCount)
	assert.Len(t, response.Failed, 1)
	assert.Equal(t, cardID1.String(), response.Failed[0].ID)
	assert.Contains(t, response.Failed[0].Error, "Not the owner")
	mockUser.AssertExpectations(t)
	mockAuthz.AssertExpectations(t)
}

func TestBatchHandler_TransferCards_TransferFailure(t *testing.T) {
	handler, _, _, _, mockAuthz, _, mockTransfer, mockUser, _ := setupBatchTest()

	cardID1 := uuid.New()
	cardID2 := uuid.New()
	newOwnerID := uuid.New()
	body := `{"ids":["` + cardID1.String() + `","` + cardID2.String() + `"],"new_owner_email":"newowner@example.com"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/cards/batch/transfer", body)
	user := createTestUser()
	c.Set("current_user", user)

	newOwner := &models.User{ID: newOwnerID, Email: "newowner@example.com"}
	perms := &services.ResourcePermissions{
		CanView: true, CanEdit: true, CanDelete: true, IsOwner: true,
	}

	mockUser.On("GetUserByEmail", mock.Anything, "newowner@example.com").Return(newOwner, nil)
	mockAuthz.On("CheckCardAccess", mock.Anything, user.ID, cardID1).Return(perms, nil)
	mockAuthz.On("CheckCardAccess", mock.Anything, user.ID, cardID2).Return(perms, nil)
	// First transfer succeeds, second fails
	mockTransfer.On("TransferCardOwnership", mock.Anything, cardID1, newOwnerID, user.ID).Return(nil)
	mockTransfer.On("TransferCardOwnership", mock.Anything, cardID2, newOwnerID, user.ID).
		Return(errors.New("database error"))

	err := handler.TransferCards(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response BatchResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, 1, response.SuccessCount)
	assert.Len(t, response.Failed, 1)
	assert.Equal(t, cardID2.String(), response.Failed[0].ID)
	assert.Contains(t, response.Failed[0].Error, "database error")
	mockUser.AssertExpectations(t)
	mockAuthz.AssertExpectations(t)
	mockTransfer.AssertExpectations(t)
}

func TestBatchHandler_TransferCards_UserNotFound(t *testing.T) {
	handler, _, _, _, _, _, _, mockUser, _ := setupBatchTest()

	cardID := uuid.New()
	body := `{"ids":["` + cardID.String() + `"],"new_owner_email":"unknown@example.com"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/cards/batch/transfer", body)
	user := createTestUser()
	c.Set("current_user", user)

	mockUser.On("GetUserByEmail", mock.Anything, "unknown@example.com").
		Return((*models.User)(nil), errors.New("not found"))

	err := handler.TransferCards(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "transfer_failed", response.Error)
	mockUser.AssertExpectations(t)
}

func TestBatchHandler_TransferCards_EmptyIDs(t *testing.T) {
	handler, _, _, _, _, _, _, _, _ := setupBatchTest()

	body := `{"ids":[],"new_owner_email":"newowner@example.com"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/cards/batch/transfer", body)
	user := createTestUser()
	c.Set("current_user", user)

	err := handler.TransferCards(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "invalid_request", response.Error)
}

func TestBatchHandler_TransferCards_InvalidJSON(t *testing.T) {
	handler, _, _, _, _, _, _, _, _ := setupBatchTest()

	c, rec := createTestContext(http.MethodPost, "/api/v1/cards/batch/transfer", "not-json")
	user := createTestUser()
	c.Set("current_user", user)

	err := handler.TransferCards(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "invalid_request", response.Error)
}

func TestBatchHandler_TransferCards_EmailNormalization(t *testing.T) {
	handler, _, _, _, mockAuthz, _, mockTransfer, mockUser, _ := setupBatchTest()

	cardID := uuid.New()
	newOwnerID := uuid.New()
	// Email with leading/trailing spaces and mixed case
	body := `{"ids":["` + cardID.String() + `"],"new_owner_email":"  NewOwner@Example.COM  "}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/cards/batch/transfer", body)
	user := createTestUser()
	c.Set("current_user", user)

	newOwner := &models.User{ID: newOwnerID, Email: "newowner@example.com"}
	perms := &services.ResourcePermissions{
		CanView: true, CanEdit: true, CanDelete: true, IsOwner: true,
	}

	// Handler normalizes email to lowercase and trimmed
	mockUser.On("GetUserByEmail", mock.Anything, "newowner@example.com").Return(newOwner, nil)
	mockAuthz.On("CheckCardAccess", mock.Anything, user.ID, cardID).Return(perms, nil)
	mockTransfer.On("TransferCardOwnership", mock.Anything, cardID, newOwnerID, user.ID).Return(nil)

	err := handler.TransferCards(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockUser.AssertExpectations(t)
}

func TestBatchHandler_TransferVouchers_Success(t *testing.T) {
	handler, _, _, _, mockAuthz, _, mockTransfer, mockUser, _ := setupBatchTest()

	voucherID1 := uuid.New()
	voucherID2 := uuid.New()
	newOwnerID := uuid.New()
	body := `{"ids":["` + voucherID1.String() + `","` + voucherID2.String() + `"],"new_owner_email":"newowner@example.com"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/vouchers/batch/transfer", body)
	user := createTestUser()
	c.Set("current_user", user)

	newOwner := &models.User{ID: newOwnerID, Email: "newowner@example.com"}
	perms := &services.ResourcePermissions{
		CanView: true, CanEdit: true, CanDelete: true, IsOwner: true,
	}

	mockUser.On("GetUserByEmail", mock.Anything, "newowner@example.com").Return(newOwner, nil)
	mockAuthz.On("CheckVoucherAccess", mock.Anything, user.ID, voucherID1).Return(perms, nil)
	mockAuthz.On("CheckVoucherAccess", mock.Anything, user.ID, voucherID2).Return(perms, nil)
	mockTransfer.On("TransferVoucherOwnership", mock.Anything, voucherID1, newOwnerID, user.ID).Return(nil)
	mockTransfer.On("TransferVoucherOwnership", mock.Anything, voucherID2, newOwnerID, user.ID).Return(nil)

	err := handler.TransferVouchers(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response BatchResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, 2, response.SuccessCount)
	assert.Empty(t, response.Failed)
	mockUser.AssertExpectations(t)
	mockAuthz.AssertExpectations(t)
	mockTransfer.AssertExpectations(t)
}

func TestBatchHandler_TransferGiftCards_Success(t *testing.T) {
	handler, _, _, _, mockAuthz, _, mockTransfer, mockUser, _ := setupBatchTest()

	gcID1 := uuid.New()
	gcID2 := uuid.New()
	newOwnerID := uuid.New()
	body := `{"ids":["` + gcID1.String() + `","` + gcID2.String() + `"],"new_owner_email":"newowner@example.com"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/gift-cards/batch/transfer", body)
	user := createTestUser()
	c.Set("current_user", user)

	newOwner := &models.User{ID: newOwnerID, Email: "newowner@example.com"}
	perms := &services.ResourcePermissions{
		CanView: true, CanEdit: true, CanDelete: true, IsOwner: true,
	}

	mockUser.On("GetUserByEmail", mock.Anything, "newowner@example.com").Return(newOwner, nil)
	mockAuthz.On("CheckGiftCardAccess", mock.Anything, user.ID, gcID1).Return(perms, nil)
	mockAuthz.On("CheckGiftCardAccess", mock.Anything, user.ID, gcID2).Return(perms, nil)
	mockTransfer.On("TransferGiftCardOwnership", mock.Anything, gcID1, newOwnerID, user.ID).Return(nil)
	mockTransfer.On("TransferGiftCardOwnership", mock.Anything, gcID2, newOwnerID, user.ID).Return(nil)

	err := handler.TransferGiftCards(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response BatchResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, 2, response.SuccessCount)
	assert.Empty(t, response.Failed)
	mockUser.AssertExpectations(t)
	mockAuthz.AssertExpectations(t)
	mockTransfer.AssertExpectations(t)
}

// ==================== Batch Export Tests ====================

func TestBatchHandler_ExportCards_Success(t *testing.T) {
	handler, _, _, _, mockAuthz, _, _, _, mockExport := setupBatchTest()

	cardID1 := uuid.New()
	cardID2 := uuid.New()
	body := `{"ids":["` + cardID1.String() + `","` + cardID2.String() + `"]}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/cards/batch/export", body)
	user := createTestUser()
	c.Set("current_user", user)

	perms := &services.ResourcePermissions{
		CanView: true, CanEdit: true, CanDelete: true, IsOwner: true,
	}

	mockAuthz.On("CheckCardAccess", mock.Anything, user.ID, cardID1).Return(perms, nil)
	mockAuthz.On("CheckCardAccess", mock.Anything, user.ID, cardID2).Return(perms, nil)

	exportData := &services.BatchExportData{
		ExportedAt: "2026-03-04T00:00:00Z",
		Cards: []services.ExportCard{
			{MerchantName: "IKEA", CardNumber: "123"},
			{MerchantName: "Migros", CardNumber: "456"},
		},
	}
	mockExport.On("ExportCardsByIDs", mock.Anything, user.ID, []uuid.UUID{cardID1, cardID2}).Return(exportData, nil)

	err := handler.ExportCards(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify Content-Disposition header for file download
	contentDisposition := rec.Header().Get("Content-Disposition")
	assert.Contains(t, contentDisposition, "attachment")
	assert.Contains(t, contentDisposition, "savvy-export-cards-")
	assert.Contains(t, contentDisposition, ".json")

	// Verify response body contains export data
	var responseData services.BatchExportData
	_ = json.Unmarshal(rec.Body.Bytes(), &responseData)
	assert.Equal(t, "2026-03-04T00:00:00Z", responseData.ExportedAt)
	assert.Len(t, responseData.Cards, 2)
	mockAuthz.AssertExpectations(t)
	mockExport.AssertExpectations(t)
}

func TestBatchHandler_ExportCards_Forbidden(t *testing.T) {
	handler, _, _, _, mockAuthz, _, _, _, _ := setupBatchTest()

	cardID1 := uuid.New()
	cardID2 := uuid.New()
	body := `{"ids":["` + cardID1.String() + `","` + cardID2.String() + `"]}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/cards/batch/export", body)
	user := createTestUser()
	c.Set("current_user", user)

	permsOK := &services.ResourcePermissions{
		CanView: true, CanEdit: true, CanDelete: true, IsOwner: true,
	}
	permsDenied := &services.ResourcePermissions{
		CanView: false, CanEdit: false, CanDelete: false, IsOwner: false,
	}

	// First card passes, second card fails view permission check
	mockAuthz.On("CheckCardAccess", mock.Anything, user.ID, cardID1).Return(permsOK, nil)
	mockAuthz.On("CheckCardAccess", mock.Anything, user.ID, cardID2).Return(permsDenied, nil)

	err := handler.ExportCards(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "forbidden", response.Error)
	assert.Contains(t, response.Message, "permission to export")
	mockAuthz.AssertExpectations(t)
}

func TestBatchHandler_ExportCards_ServiceError(t *testing.T) {
	handler, _, _, _, mockAuthz, _, _, _, mockExport := setupBatchTest()

	cardID := uuid.New()
	body := `{"ids":["` + cardID.String() + `"]}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/cards/batch/export", body)
	user := createTestUser()
	c.Set("current_user", user)

	perms := &services.ResourcePermissions{
		CanView: true, CanEdit: true, CanDelete: true, IsOwner: true,
	}

	mockAuthz.On("CheckCardAccess", mock.Anything, user.ID, cardID).Return(perms, nil)
	mockExport.On("ExportCardsByIDs", mock.Anything, user.ID, []uuid.UUID{cardID}).
		Return((*services.BatchExportData)(nil), errors.New("export failed"))

	err := handler.ExportCards(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "export_failed", response.Error)
	assert.Contains(t, response.Message, "Failed to export data")
	mockAuthz.AssertExpectations(t)
	mockExport.AssertExpectations(t)
}

func TestBatchHandler_ExportVouchers_Success(t *testing.T) {
	handler, _, _, _, mockAuthz, _, _, _, mockExport := setupBatchTest()

	voucherID1 := uuid.New()
	voucherID2 := uuid.New()
	body := `{"ids":["` + voucherID1.String() + `","` + voucherID2.String() + `"]}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/vouchers/batch/export", body)
	user := createTestUser()
	c.Set("current_user", user)

	perms := &services.ResourcePermissions{
		CanView: true, CanEdit: true, CanDelete: true, IsOwner: true,
	}

	mockAuthz.On("CheckVoucherAccess", mock.Anything, user.ID, voucherID1).Return(perms, nil)
	mockAuthz.On("CheckVoucherAccess", mock.Anything, user.ID, voucherID2).Return(perms, nil)

	exportData := &services.BatchExportData{
		ExportedAt: "2026-03-04T00:00:00Z",
		Vouchers: []services.ExportVoucher{
			{MerchantName: "Amazon", Code: "ABC123"},
			{MerchantName: "Steam", Code: "DEF456"},
		},
	}
	mockExport.On("ExportVouchersByIDs", mock.Anything, user.ID, []uuid.UUID{voucherID1, voucherID2}).Return(exportData, nil)

	err := handler.ExportVouchers(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify Content-Disposition header for file download
	contentDisposition := rec.Header().Get("Content-Disposition")
	assert.Contains(t, contentDisposition, "attachment")
	assert.Contains(t, contentDisposition, "savvy-export-vouchers-")
	assert.Contains(t, contentDisposition, ".json")

	var responseData services.BatchExportData
	_ = json.Unmarshal(rec.Body.Bytes(), &responseData)
	assert.Equal(t, "2026-03-04T00:00:00Z", responseData.ExportedAt)
	assert.Len(t, responseData.Vouchers, 2)
	mockAuthz.AssertExpectations(t)
	mockExport.AssertExpectations(t)
}

func TestBatchHandler_ExportGiftCards_Success(t *testing.T) {
	handler, _, _, _, mockAuthz, _, _, _, mockExport := setupBatchTest()

	gcID1 := uuid.New()
	gcID2 := uuid.New()
	body := `{"ids":["` + gcID1.String() + `","` + gcID2.String() + `"]}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/gift-cards/batch/export", body)
	user := createTestUser()
	c.Set("current_user", user)

	perms := &services.ResourcePermissions{
		CanView: true, CanEdit: true, CanDelete: true, IsOwner: true,
	}

	mockAuthz.On("CheckGiftCardAccess", mock.Anything, user.ID, gcID1).Return(perms, nil)
	mockAuthz.On("CheckGiftCardAccess", mock.Anything, user.ID, gcID2).Return(perms, nil)

	exportData := &services.BatchExportData{
		ExportedAt: "2026-03-04T00:00:00Z",
		GiftCards: []services.ExportGiftCard{
			{MerchantName: "Apple", CardNumber: "GC001"},
			{MerchantName: "Google", CardNumber: "GC002"},
		},
	}
	mockExport.On("ExportGiftCardsByIDs", mock.Anything, user.ID, []uuid.UUID{gcID1, gcID2}).Return(exportData, nil)

	err := handler.ExportGiftCards(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify Content-Disposition header for file download
	contentDisposition := rec.Header().Get("Content-Disposition")
	assert.Contains(t, contentDisposition, "attachment")
	assert.Contains(t, contentDisposition, "savvy-export-gift-cards-")
	assert.Contains(t, contentDisposition, ".json")

	var responseData services.BatchExportData
	_ = json.Unmarshal(rec.Body.Bytes(), &responseData)
	assert.Equal(t, "2026-03-04T00:00:00Z", responseData.ExportedAt)
	assert.Len(t, responseData.GiftCards, 2)
	mockAuthz.AssertExpectations(t)
	mockExport.AssertExpectations(t)
}

// ==================== sanitizeBatchError Tests ====================

func TestSanitizeBatchError_HidesSQLErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		fallback string
	}{
		{
			name:     "SQLSTATE error",
			err:      errors.New("ERROR: duplicate key value violates unique constraint (SQLSTATE 23505)"),
			fallback: "Failed to process item",
		},
		{
			name:     "pq: error",
			err:      errors.New("pq: connection refused"),
			fallback: "Database unavailable",
		},
		{
			name:     "sql: error",
			err:      errors.New("sql: no rows in result set"),
			fallback: "Item not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeBatchError(tt.err, tt.fallback)
			assert.Equal(t, tt.fallback, result)
			// Ensure the raw SQL error is NOT exposed
			assert.NotContains(t, result, "SQLSTATE")
			assert.NotContains(t, result, "pq:")
			assert.NotContains(t, result, "sql:")
		})
	}
}

func TestSanitizeBatchError_PassesBusinessErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "not found",
			err:      errors.New("resource not found"),
			expected: "resource not found",
		},
		{
			name:     "permission denied",
			err:      errors.New("permission denied"),
			expected: "permission denied",
		},
		{
			name:     "duplicate share",
			err:      errors.New("already shared with this user"),
			expected: "already shared with this user",
		},
		{
			name:     "validation error",
			err:      errors.New("invalid resource ID"),
			expected: "invalid resource ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeBatchError(tt.err, "fallback message")
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ==================== parseBatchIDs Tests ====================

func TestParseBatchIDs_Success(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()
	ids, err := parseBatchIDs([]string{id1.String(), id2.String()})

	assert.NoError(t, err)
	assert.Len(t, ids, 2)
	assert.Equal(t, id1, ids[0])
	assert.Equal(t, id2, ids[1])
}

func TestParseBatchIDs_Empty(t *testing.T) {
	ids, err := parseBatchIDs([]string{})

	assert.Error(t, err)
	assert.Nil(t, ids)
}

func TestParseBatchIDs_Nil(t *testing.T) {
	ids, err := parseBatchIDs(nil)

	assert.Error(t, err)
	assert.Nil(t, ids)
}

func TestParseBatchIDs_TooMany(t *testing.T) {
	rawIDs := make([]string, 51)
	for i := range rawIDs {
		rawIDs[i] = uuid.New().String()
	}
	ids, err := parseBatchIDs(rawIDs)

	assert.Error(t, err)
	assert.Nil(t, ids)
}

func TestParseBatchIDs_InvalidUUID(t *testing.T) {
	ids, err := parseBatchIDs([]string{"not-a-uuid"})

	assert.Error(t, err)
	assert.Nil(t, ids)
}

func TestParseBatchIDs_Deduplication(t *testing.T) {
	id := uuid.New()
	ids, err := parseBatchIDs([]string{id.String(), id.String(), id.String()})

	assert.NoError(t, err)
	assert.Len(t, ids, 1)
	assert.Equal(t, id, ids[0])
}

func TestParseBatchIDs_MaxBoundary(t *testing.T) {
	// Exactly 50 IDs should be fine
	rawIDs := make([]string, 50)
	for i := range rawIDs {
		rawIDs[i] = uuid.New().String()
	}
	ids, err := parseBatchIDs(rawIDs)

	assert.NoError(t, err)
	assert.Len(t, ids, 50)
}

func TestParseBatchIDs_MixedValidInvalid(t *testing.T) {
	validID := uuid.New()
	ids, err := parseBatchIDs([]string{validID.String(), "invalid"})

	assert.Error(t, err)
	assert.Nil(t, ids)
}
