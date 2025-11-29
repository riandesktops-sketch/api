package handlers

import (
	"zodiac-ai-backend/pkg/middleware"
	"zodiac-ai-backend/pkg/response"
	"zodiac-ai-backend/services/auth-service/models"
	"zodiac-ai-backend/services/auth-service/services"

	"github.com/gofiber/fiber/v2"
)

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
	authService *services.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register handles user registration
// POST /auth/register
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req models.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", nil)
	}

	authResp, err := h.authService.Register(c.Context(), &req)
	if err != nil {
		if err == services.ErrEmailAlreadyExists {
			return response.Conflict(c, "Email already exists")
		}
		// Log the actual error for debugging
		// In a real app, use a proper logger
		println("Registration error:", err.Error())
		return response.InternalServerError(c, "Failed to register user: "+err.Error())
	}

	return response.Created(c, "User registered successfully", authResp)
}

// Login handles user login
// POST /auth/login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req models.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", nil)
	}

	authResp, err := h.authService.Login(c.Context(), &req)
	if err != nil {
		if err == services.ErrInvalidCredentials {
			return response.Unauthorized(c, "Invalid email or password")
		}
		return response.InternalServerError(c, "Failed to login")
	}

	return response.Success(c, "Login successful", authResp)
}

// RefreshToken handles access token refresh
// POST /auth/refresh
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", nil)
	}

	accessToken, err := h.authService.RefreshAccessToken(c.Context(), req.RefreshToken)
	if err != nil {
		return response.Unauthorized(c, "Invalid or expired refresh token")
	}

	return response.Success(c, "Token refreshed successfully", fiber.Map{
		"access_token": accessToken,
	})
}

// GetProfile handles get user profile
// GET /users/me
func (h *AuthHandler) GetProfile(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return response.Unauthorized(c, "User not authenticated")
	}

	user, err := h.authService.GetProfile(c.Context(), userID)
	if err != nil {
		return response.InternalServerError(c, "Failed to get profile")
	}

	return response.Success(c, "Profile retrieved successfully", user)
}

// UpdateProfile handles update user profile
// PUT /users/me
func (h *AuthHandler) UpdateProfile(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return response.Unauthorized(c, "User not authenticated")
	}

	var req models.UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", nil)
	}

	user, err := h.authService.UpdateProfile(c.Context(), userID, &req)
	if err != nil {
		return response.InternalServerError(c, "Failed to update profile")
	}

	return response.Success(c, "Profile updated successfully", user)
}
