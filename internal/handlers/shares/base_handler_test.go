package shares

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	goI18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"
	"savvy/internal/i18n"
	"savvy/internal/models"
	"savvy/internal/services"
)

func TestMain(m *testing.M) {
	// Initialize i18n bundle with minimal messages for tests.
	i18n.Bundle = goI18n.NewBundle(language.German)
	i18n.Bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	msgs := map[string]string{
		"error.invalid_id":            "Invalid ID",
		"error.unauthorized":          "Unauthorized",
		"error.user_not_found":        "User not found",
		"error.already_shared":        "Already shared",
		"error.server_error":          "Server error",
		"error.invalid_resource_id":   "Invalid resource ID",
		"error.invalid_share_id":      "Invalid share ID",
		"error.updates_not_supported": "Updates not supported",
		"error.update_share_failed":   "Update share failed",
		"error.delete_share_failed":   "Delete share failed",
		"error.editing_not_supported": "Editing not supported",
		"success.created":             "Created",
		"success.updated":             "Updated",
		"success.deleted":             "Deleted",
	}

	for id, msg := range msgs {
		i18n.Bundle.MustAddMessages(language.German, &goI18n.Message{ID: id, Other: msg})
	}

	os.Exit(m.Run())
}

// ==================== Mock Adapter ====================

type mockAdapter struct {
	resourceType           string
	resourceName           string
	supportsEdit           bool
	hasTransactionPerm     bool
	checkOwnershipResult   bool
	checkOwnershipErr      error
	createShareErr         error
	updateShareErr         error
	deleteShareErr         error
	listSharesResult       []ShareView
	listSharesErr          error
	lastCreateReq          *CreateShareRequest
	lastUpdateReq          *UpdateShareRequest
	lastDeleteCallerUserID uuid.UUID
	lastDeleteSharedWithID uuid.UUID
	lastDeleteResourceID   uuid.UUID
}

var _ ShareAdapter = (*mockAdapter)(nil)

func (m *mockAdapter) ResourceType() string           { return m.resourceType }
func (m *mockAdapter) ResourceName() string           { return m.resourceName }
func (m *mockAdapter) SupportsEdit() bool             { return m.supportsEdit }
func (m *mockAdapter) HasTransactionPermission() bool { return m.hasTransactionPerm }

func (m *mockAdapter) CheckOwnership(_ context.Context, _, _ uuid.UUID) (bool, error) {
	return m.checkOwnershipResult, m.checkOwnershipErr
}

func (m *mockAdapter) ListShares(_ context.Context, _ uuid.UUID) ([]ShareView, error) {
	return m.listSharesResult, m.listSharesErr
}

func (m *mockAdapter) CreateShare(_ context.Context, req CreateShareRequest) error {
	m.lastCreateReq = &req
	return m.createShareErr
}

func (m *mockAdapter) UpdateShare(_ context.Context, req UpdateShareRequest) error {
	m.lastUpdateReq = &req
	return m.updateShareErr
}

func (m *mockAdapter) DeleteShare(_ context.Context, callerUserID, sharedWithID, resourceID uuid.UUID) error {
	m.lastDeleteCallerUserID = callerUserID
	m.lastDeleteSharedWithID = sharedWithID
	m.lastDeleteResourceID = resourceID
	return m.deleteShareErr
}

// ==================== Mock UserService ====================

type mockUserService struct {
	services.UserServiceInterface
	user *models.User
	err  error
}

func (m *mockUserService) GetUserByEmail(_ context.Context, _ string) (*models.User, error) {
	return m.user, m.err
}

// ==================== Test Helpers ====================

func newTestUser() *models.User {
	return &models.User{
		ID:        uuid.New(),
		Email:     "owner@example.com",
		FirstName: "Test",
		LastName:  "Owner",
		Role:      "user",
	}
}

func newEchoContext(method, path string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, strings.NewReader(""))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

func newFormContext(method string, formData string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, "/shares", strings.NewReader(formData))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

func newHTMXContext(method, path string, formData string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, strings.NewReader(formData))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

// ==================== NewBaseShareHandler ====================

func TestNewBaseShareHandler(t *testing.T) {
	adapter := &mockAdapter{}
	userSvc := &mockUserService{}
	h := NewBaseShareHandler(adapter, userSvc)

	assert.NotNil(t, h)
	assert.Equal(t, adapter, h.adapter)
	assert.Equal(t, userSvc, h.userService)
}

// ==================== Create Tests ====================

func TestCreate_Success(t *testing.T) {
	sharedUser := &models.User{ID: uuid.New(), Email: "shared@example.com"}
	adapter := &mockAdapter{checkOwnershipResult: true}
	userSvc := &mockUserService{user: sharedUser}
	h := NewBaseShareHandler(adapter, userSvc)

	resourceID := uuid.New()
	formData := "shared_with_email=shared@example.com&can_edit=on&can_delete=on"
	c, rec := newFormContext(http.MethodPost, formData)
	user := newTestUser()
	c.Set("current_user", user)
	c.SetParamNames("id")
	c.SetParamValues(resourceID.String())

	err := h.Create(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotNil(t, adapter.lastCreateReq)
	assert.Equal(t, user.ID, adapter.lastCreateReq.UserID)
	assert.Equal(t, resourceID, adapter.lastCreateReq.ResourceID)
	assert.Equal(t, "shared@example.com", adapter.lastCreateReq.SharedWithEmail)
	assert.True(t, adapter.lastCreateReq.CanEdit)
	assert.True(t, adapter.lastCreateReq.CanDelete)
	assert.False(t, adapter.lastCreateReq.CanEditTransactions)
}

func TestCreate_Success_HTMX(t *testing.T) {
	sharedUser := &models.User{ID: uuid.New(), Email: "shared@example.com"}
	adapter := &mockAdapter{checkOwnershipResult: true}
	userSvc := &mockUserService{user: sharedUser}
	h := NewBaseShareHandler(adapter, userSvc)

	resourceID := uuid.New()
	c, rec := newHTMXContext(http.MethodPost, "/shares", "shared_with_email=shared@example.com")
	user := newTestUser()
	c.Set("current_user", user)
	c.SetParamNames("id")
	c.SetParamValues(resourceID.String())

	err := h.Create(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "true", rec.Header().Get("HX-Refresh"))
	assert.Empty(t, rec.Body.String())
}

func TestCreate_WithTransactionPermission(t *testing.T) {
	sharedUser := &models.User{ID: uuid.New(), Email: "shared@example.com"}
	adapter := &mockAdapter{checkOwnershipResult: true, hasTransactionPerm: true}
	userSvc := &mockUserService{user: sharedUser}
	h := NewBaseShareHandler(adapter, userSvc)

	resourceID := uuid.New()
	formData := "shared_with_email=shared@example.com&can_edit_transactions=on"
	c, rec := newFormContext(http.MethodPost, formData)
	user := newTestUser()
	c.Set("current_user", user)
	c.SetParamNames("id")
	c.SetParamValues(resourceID.String())

	err := h.Create(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, adapter.lastCreateReq.CanEditTransactions)
}

func TestCreate_InvalidResourceID(t *testing.T) {
	adapter := &mockAdapter{}
	userSvc := &mockUserService{}
	h := NewBaseShareHandler(adapter, userSvc)

	c, rec := newFormContext(http.MethodPost, "")
	c.Set("current_user", newTestUser())
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")

	err := h.Create(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreate_NotOwner(t *testing.T) {
	adapter := &mockAdapter{checkOwnershipResult: false}
	userSvc := &mockUserService{}
	h := NewBaseShareHandler(adapter, userSvc)

	resourceID := uuid.New()
	c, rec := newFormContext(http.MethodPost, "")
	c.Set("current_user", newTestUser())
	c.SetParamNames("id")
	c.SetParamValues(resourceID.String())

	err := h.Create(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestCreate_OwnershipCheckError(t *testing.T) {
	adapter := &mockAdapter{checkOwnershipResult: false, checkOwnershipErr: errors.New("db error")}
	userSvc := &mockUserService{}
	h := NewBaseShareHandler(adapter, userSvc)

	resourceID := uuid.New()
	c, rec := newFormContext(http.MethodPost, "")
	c.Set("current_user", newTestUser())
	c.SetParamNames("id")
	c.SetParamValues(resourceID.String())

	err := h.Create(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestCreate_UserNotFound(t *testing.T) {
	adapter := &mockAdapter{checkOwnershipResult: true}
	userSvc := &mockUserService{err: errors.New("user not found")}
	h := NewBaseShareHandler(adapter, userSvc)

	resourceID := uuid.New()
	formData := "shared_with_email=nonexistent@example.com"
	c, rec := newFormContext(http.MethodPost, formData)
	c.Set("current_user", newTestUser())
	c.SetParamNames("id")
	c.SetParamValues(resourceID.String())

	err := h.Create(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreate_CreateShareError_UserNotFound(t *testing.T) {
	sharedUser := &models.User{ID: uuid.New(), Email: "shared@example.com"}
	adapter := &mockAdapter{checkOwnershipResult: true, createShareErr: errors.New("user not found")}
	userSvc := &mockUserService{user: sharedUser}
	h := NewBaseShareHandler(adapter, userSvc)

	resourceID := uuid.New()
	formData := "shared_with_email=shared@example.com"
	c, rec := newFormContext(http.MethodPost, formData)
	c.Set("current_user", newTestUser())
	c.SetParamNames("id")
	c.SetParamValues(resourceID.String())

	err := h.Create(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreate_CreateShareError_AlreadyShared(t *testing.T) {
	sharedUser := &models.User{ID: uuid.New(), Email: "shared@example.com"}
	adapter := &mockAdapter{checkOwnershipResult: true, createShareErr: errors.New("already shared with this user")}
	userSvc := &mockUserService{user: sharedUser}
	h := NewBaseShareHandler(adapter, userSvc)

	resourceID := uuid.New()
	formData := "shared_with_email=shared@example.com"
	c, rec := newFormContext(http.MethodPost, formData)
	c.Set("current_user", newTestUser())
	c.SetParamNames("id")
	c.SetParamValues(resourceID.String())

	err := h.Create(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreate_CreateShareError_ServerError(t *testing.T) {
	sharedUser := &models.User{ID: uuid.New(), Email: "shared@example.com"}
	adapter := &mockAdapter{checkOwnershipResult: true, createShareErr: errors.New("database connection lost")}
	userSvc := &mockUserService{user: sharedUser}
	h := NewBaseShareHandler(adapter, userSvc)

	resourceID := uuid.New()
	formData := "shared_with_email=shared@example.com"
	c, rec := newFormContext(http.MethodPost, formData)
	c.Set("current_user", newTestUser())
	c.SetParamNames("id")
	c.SetParamValues(resourceID.String())

	err := h.Create(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreate_EmailTrimmedAndLowercased(t *testing.T) {
	sharedUser := &models.User{ID: uuid.New(), Email: "shared@example.com"}
	adapter := &mockAdapter{checkOwnershipResult: true}
	userSvc := &mockUserService{user: sharedUser}
	h := NewBaseShareHandler(adapter, userSvc)

	resourceID := uuid.New()
	formData := "shared_with_email=  Shared@Example.COM  "
	c, rec := newFormContext(http.MethodPost, formData)
	c.Set("current_user", newTestUser())
	c.SetParamNames("id")
	c.SetParamValues(resourceID.String())

	err := h.Create(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "shared@example.com", adapter.lastCreateReq.SharedWithEmail)
}

func TestCreate_NoTransactionPermForCards(t *testing.T) {
	sharedUser := &models.User{ID: uuid.New(), Email: "shared@example.com"}
	adapter := &mockAdapter{checkOwnershipResult: true, hasTransactionPerm: false}
	userSvc := &mockUserService{user: sharedUser}
	h := NewBaseShareHandler(adapter, userSvc)

	resourceID := uuid.New()
	formData := "shared_with_email=shared@example.com&can_edit_transactions=on"
	c, rec := newFormContext(http.MethodPost, formData)
	c.Set("current_user", newTestUser())
	c.SetParamNames("id")
	c.SetParamValues(resourceID.String())

	err := h.Create(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.False(t, adapter.lastCreateReq.CanEditTransactions)
}

// ==================== Update Tests ====================

func TestUpdate_Success(t *testing.T) {
	adapter := &mockAdapter{supportsEdit: true, checkOwnershipResult: true}
	h := NewBaseShareHandler(adapter, nil)

	resourceID := uuid.New()
	sharedWithID := uuid.New()
	formData := "can_edit=on&can_delete=on"
	c, rec := newFormContext(http.MethodPut, formData)
	user := newTestUser()
	c.Set("current_user", user)
	c.SetParamNames("id", "shared_with_id")
	c.SetParamValues(resourceID.String(), sharedWithID.String())

	err := h.Update(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotNil(t, adapter.lastUpdateReq)
	assert.Equal(t, user.ID, adapter.lastUpdateReq.CallerUserID)
	assert.Equal(t, sharedWithID, adapter.lastUpdateReq.SharedWithID)
	assert.Equal(t, resourceID, adapter.lastUpdateReq.ResourceID)
	assert.True(t, adapter.lastUpdateReq.CanEdit)
	assert.True(t, adapter.lastUpdateReq.CanDelete)
}

func TestUpdate_Success_HTMX(t *testing.T) {
	adapter := &mockAdapter{supportsEdit: true, checkOwnershipResult: true}
	h := NewBaseShareHandler(adapter, nil)

	resourceID := uuid.New()
	sharedWithID := uuid.New()
	c, rec := newHTMXContext(http.MethodPut, "/shares", "can_edit=on")
	c.Set("current_user", newTestUser())
	c.SetParamNames("id", "shared_with_id")
	c.SetParamValues(resourceID.String(), sharedWithID.String())

	err := h.Update(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "true", rec.Header().Get("HX-Refresh"))
	assert.Empty(t, rec.Body.String())
}

func TestUpdate_WithTransactionPermission(t *testing.T) {
	adapter := &mockAdapter{supportsEdit: true, checkOwnershipResult: true, hasTransactionPerm: true}
	h := NewBaseShareHandler(adapter, nil)

	resourceID := uuid.New()
	sharedWithID := uuid.New()
	formData := "can_edit_transactions=on"
	c, rec := newFormContext(http.MethodPut, formData)
	c.Set("current_user", newTestUser())
	c.SetParamNames("id", "shared_with_id")
	c.SetParamValues(resourceID.String(), sharedWithID.String())

	err := h.Update(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, adapter.lastUpdateReq.CanEditTransactions)
}

func TestUpdate_NotSupported(t *testing.T) {
	adapter := &mockAdapter{supportsEdit: false}
	h := NewBaseShareHandler(adapter, nil)

	c, rec := newFormContext(http.MethodPut, "")
	c.Set("current_user", newTestUser())

	err := h.Update(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestUpdate_InvalidResourceID(t *testing.T) {
	adapter := &mockAdapter{supportsEdit: true}
	h := NewBaseShareHandler(adapter, nil)

	c, rec := newFormContext(http.MethodPut, "")
	c.Set("current_user", newTestUser())
	c.SetParamNames("id", "shared_with_id")
	c.SetParamValues("bad-uuid", uuid.New().String())

	err := h.Update(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUpdate_InvalidShareID(t *testing.T) {
	adapter := &mockAdapter{supportsEdit: true}
	h := NewBaseShareHandler(adapter, nil)

	c, rec := newFormContext(http.MethodPut, "")
	c.Set("current_user", newTestUser())
	c.SetParamNames("id", "shared_with_id")
	c.SetParamValues(uuid.New().String(), "bad-uuid")

	err := h.Update(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUpdate_NotOwner(t *testing.T) {
	adapter := &mockAdapter{supportsEdit: true, checkOwnershipResult: false}
	h := NewBaseShareHandler(adapter, nil)

	c, rec := newFormContext(http.MethodPut, "")
	c.Set("current_user", newTestUser())
	c.SetParamNames("id", "shared_with_id")
	c.SetParamValues(uuid.New().String(), uuid.New().String())

	err := h.Update(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestUpdate_OwnershipCheckError(t *testing.T) {
	adapter := &mockAdapter{supportsEdit: true, checkOwnershipErr: errors.New("db error")}
	h := NewBaseShareHandler(adapter, nil)

	c, rec := newFormContext(http.MethodPut, "")
	c.Set("current_user", newTestUser())
	c.SetParamNames("id", "shared_with_id")
	c.SetParamValues(uuid.New().String(), uuid.New().String())

	err := h.Update(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestUpdate_AdapterError(t *testing.T) {
	adapter := &mockAdapter{supportsEdit: true, checkOwnershipResult: true, updateShareErr: errors.New("update failed")}
	h := NewBaseShareHandler(adapter, nil)

	c, rec := newFormContext(http.MethodPut, "")
	c.Set("current_user", newTestUser())
	c.SetParamNames("id", "shared_with_id")
	c.SetParamValues(uuid.New().String(), uuid.New().String())

	err := h.Update(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// ==================== Delete Tests ====================

func TestDelete_Success(t *testing.T) {
	adapter := &mockAdapter{checkOwnershipResult: true}
	h := NewBaseShareHandler(adapter, nil)

	resourceID := uuid.New()
	sharedWithID := uuid.New()
	c, rec := newEchoContext(http.MethodDelete, "/shares")
	user := newTestUser()
	c.Set("current_user", user)
	c.SetParamNames("id", "shared_with_id")
	c.SetParamValues(resourceID.String(), sharedWithID.String())

	err := h.Delete(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, user.ID, adapter.lastDeleteCallerUserID)
	assert.Equal(t, sharedWithID, adapter.lastDeleteSharedWithID)
	assert.Equal(t, resourceID, adapter.lastDeleteResourceID)
}

func TestDelete_Success_HTMX(t *testing.T) {
	adapter := &mockAdapter{checkOwnershipResult: true}
	h := NewBaseShareHandler(adapter, nil)

	resourceID := uuid.New()
	sharedWithID := uuid.New()
	c, rec := newHTMXContext(http.MethodDelete, "/shares", "")
	c.Set("current_user", newTestUser())
	c.SetParamNames("id", "shared_with_id")
	c.SetParamValues(resourceID.String(), sharedWithID.String())

	err := h.Delete(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "true", rec.Header().Get("HX-Refresh"))
	assert.Empty(t, rec.Body.String())
}

func TestDelete_InvalidResourceID(t *testing.T) {
	adapter := &mockAdapter{}
	h := NewBaseShareHandler(adapter, nil)

	c, rec := newEchoContext(http.MethodDelete, "/shares")
	c.Set("current_user", newTestUser())
	c.SetParamNames("id", "shared_with_id")
	c.SetParamValues("bad-uuid", uuid.New().String())

	err := h.Delete(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestDelete_InvalidShareID(t *testing.T) {
	adapter := &mockAdapter{}
	h := NewBaseShareHandler(adapter, nil)

	c, rec := newEchoContext(http.MethodDelete, "/shares")
	c.Set("current_user", newTestUser())
	c.SetParamNames("id", "shared_with_id")
	c.SetParamValues(uuid.New().String(), "bad-uuid")

	err := h.Delete(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestDelete_NotOwner(t *testing.T) {
	adapter := &mockAdapter{checkOwnershipResult: false}
	h := NewBaseShareHandler(adapter, nil)

	c, rec := newEchoContext(http.MethodDelete, "/shares")
	c.Set("current_user", newTestUser())
	c.SetParamNames("id", "shared_with_id")
	c.SetParamValues(uuid.New().String(), uuid.New().String())

	err := h.Delete(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestDelete_OwnershipCheckError(t *testing.T) {
	adapter := &mockAdapter{checkOwnershipErr: errors.New("db error")}
	h := NewBaseShareHandler(adapter, nil)

	c, rec := newEchoContext(http.MethodDelete, "/shares")
	c.Set("current_user", newTestUser())
	c.SetParamNames("id", "shared_with_id")
	c.SetParamValues(uuid.New().String(), uuid.New().String())

	err := h.Delete(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestDelete_AdapterError(t *testing.T) {
	adapter := &mockAdapter{checkOwnershipResult: true, deleteShareErr: errors.New("delete failed")}
	h := NewBaseShareHandler(adapter, nil)

	c, rec := newEchoContext(http.MethodDelete, "/shares")
	c.Set("current_user", newTestUser())
	c.SetParamNames("id", "shared_with_id")
	c.SetParamValues(uuid.New().String(), uuid.New().String())

	err := h.Delete(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// ==================== NewInline / Cancel Tests ====================

func TestNewInline(t *testing.T) {
	h := NewBaseShareHandler(&mockAdapter{}, nil)

	c, rec := newEchoContext(http.MethodGet, "/shares/new")

	err := h.NewInline(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestCancel(t *testing.T) {
	h := NewBaseShareHandler(&mockAdapter{}, nil)

	c, rec := newEchoContext(http.MethodGet, "/shares/cancel")

	err := h.Cancel(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// ==================== EditInline / CancelEdit Tests ====================

func TestEditInline_Supported(t *testing.T) {
	adapter := &mockAdapter{supportsEdit: true}
	h := NewBaseShareHandler(adapter, nil)

	c, rec := newEchoContext(http.MethodGet, "/shares/edit")

	err := h.EditInline(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestEditInline_NotSupported(t *testing.T) {
	adapter := &mockAdapter{supportsEdit: false}
	h := NewBaseShareHandler(adapter, nil)

	c, rec := newEchoContext(http.MethodGet, "/shares/edit")

	err := h.EditInline(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestCancelEdit_Supported(t *testing.T) {
	adapter := &mockAdapter{supportsEdit: true}
	h := NewBaseShareHandler(adapter, nil)

	c, rec := newEchoContext(http.MethodGet, "/shares/cancel-edit")

	err := h.CancelEdit(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestCancelEdit_NotSupported(t *testing.T) {
	adapter := &mockAdapter{supportsEdit: false}
	h := NewBaseShareHandler(adapter, nil)

	c, rec := newEchoContext(http.MethodGet, "/shares/cancel-edit")

	err := h.CancelEdit(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}
