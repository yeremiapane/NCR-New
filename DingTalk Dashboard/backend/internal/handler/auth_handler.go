package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

// AuthHandler handles auth proxy requests
type AuthHandler struct {
	authAPIBaseURL string
	httpClient     *http.Client
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authAPIBaseURL string) *AuthHandler {
	return &AuthHandler{
		authAPIBaseURL: authAPIBaseURL,
		httpClient:     &http.Client{},
	}
}

// Login proxies login request to external auth API
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	return h.proxyRequest(c, "/api/v1/auth/jwt/login")
}

// Register proxies register request to external auth API
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	return h.proxyRequest(c, "/api/v1/auth/register")
}

// ForgotPassword proxies forgot password request
func (h *AuthHandler) ForgotPassword(c *fiber.Ctx) error {
	return h.proxyRequest(c, "/api/v1/auth/forgot-password")
}

// ResetPassword proxies reset password request
func (h *AuthHandler) ResetPassword(c *fiber.Ctx) error {
	return h.proxyRequest(c, "/api/v1/auth/reset-password")
}

// RefreshToken proxies refresh token request
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	return h.proxyRequest(c, "/api/v1/auth/jwt/refresh")
}

// Logout proxies logout request
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	return h.proxyRequest(c, "/api/v1/auth/jwt/logout")
}

// proxyRequest forwards the request to the external auth API
func (h *AuthHandler) proxyRequest(c *fiber.Ctx, path string) error {
	url := h.authAPIBaseURL + path

	// Create new request
	req, err := http.NewRequest(c.Method(), url, bytes.NewReader(c.Body()))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to create request",
		})
	}

	// Copy headers
	req.Header.Set("Content-Type", "application/json")
	if auth := c.Get("Authorization"); auth != "" {
		req.Header.Set("Authorization", auth)
	}

	// Make request
	resp, err := h.httpClient.Do(req)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("Auth service unavailable: %v", err),
		})
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to read response",
		})
	}

	// Try to parse as JSON, otherwise return as-is
	var jsonResp interface{}
	if err := json.Unmarshal(body, &jsonResp); err != nil {
		return c.Status(resp.StatusCode).Send(body)
	}

	return c.Status(resp.StatusCode).JSON(jsonResp)
}
