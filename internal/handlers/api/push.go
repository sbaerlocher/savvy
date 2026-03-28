// Package api contains JSON API handlers for the SvelteKit frontend.
package api

//
//nolint:revive // "api" is a meaningful package name for API handlers

import (
	"net/http"
	"savvy/internal/models"
	"savvy/internal/services"

	"github.com/labstack/echo/v5"
)

// PushHandler handles Web Push API endpoints.
type PushHandler struct {
	pushService services.PushServiceInterface
}

// NewPushHandler creates a new push API handler.
func NewPushHandler(pushService services.PushServiceInterface) *PushHandler {
	return &PushHandler{pushService: pushService}
}

// SubscribeRequest is the request body for subscribing to push notifications.
type SubscribeRequest struct {
	Endpoint string `json:"endpoint"`
	Keys     struct {
		P256dh string `json:"p256dh"`
		Auth   string `json:"auth"`
	} `json:"keys"`
}

// Subscribe registers a push subscription.
// POST /api/v1/push/subscribe
func (h *PushHandler) Subscribe(c echo.Context) error {
	user, ok := c.Get("current_user").(*models.User)
	if !ok || user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "Not authenticated",
		})
	}

	var req SubscribeRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request body",
		})
	}

	if req.Endpoint == "" || req.Keys.P256dh == "" || req.Keys.Auth == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Missing required fields: endpoint, keys.p256dh, keys.auth",
		})
	}

	if err := h.pushService.Subscribe(
		c.Request().Context(),
		user.ID,
		req.Endpoint,
		req.Keys.P256dh,
		req.Keys.Auth,
		c.Request().UserAgent(),
	); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to subscribe",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "subscribed"})
}

// Unsubscribe removes a push subscription.
// POST /api/v1/push/unsubscribe
func (h *PushHandler) Unsubscribe(c echo.Context) error {
	user, ok := c.Get("current_user").(*models.User)
	if !ok || user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "Not authenticated",
		})
	}

	var req struct {
		Endpoint string `json:"endpoint"`
	}
	if err := c.Bind(&req); err != nil || req.Endpoint == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Missing required field: endpoint",
		})
	}

	if err := h.pushService.Unsubscribe(c.Request().Context(), req.Endpoint); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to unsubscribe",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "unsubscribed"})
}

// GetVAPIDKey returns the VAPID public key for client subscription.
// GET /api/v1/push/vapid-key
func (h *PushHandler) GetVAPIDKey(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"public_key": h.pushService.GetVAPIDPublicKey(),
	})
}
