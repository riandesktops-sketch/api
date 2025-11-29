package middleware

import (
	"strings"
	"zodiac-ai-backend/pkg/jwt"
	"zodiac-ai-backend/pkg/response"

	"github.com/gofiber/fiber/v2"
)

// AuthMiddleware creates authentication middleware
// Verifies JWT token and injects user context
func AuthMiddleware(jwtManager *jwt.Manager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract token from Authorization header or query param (for WebSocket)
		authHeader := c.Get("Authorization")
		var tokenString string

		if authHeader != "" {
			// Check Bearer prefix
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return response.Unauthorized(c, "Invalid authorization header format")
			}
			tokenString = parts[1]
		} else {
			// Try to get token from query param (for WebSocket)
			tokenString = c.Query("token")
			if tokenString == "" {
				return response.Unauthorized(c, "Missing authorization header or token")
			}
		}

		// Verify token
		claims, err := jwtManager.VerifyToken(tokenString)
		if err != nil {
			if err == jwt.ErrExpiredToken {
				return response.Unauthorized(c, "Token has expired")
			}
			return response.Unauthorized(c, "Invalid token")
		}

		// Validate token type (must be access token)
		if err := jwtManager.ValidateTokenType(claims, jwt.AccessToken); err != nil {
			return response.Unauthorized(c, "Invalid token type")
		}

		// Inject user context into request
		c.Locals("user_id", claims.UserID)
		c.Locals("zodiac_sign", claims.ZodiacSign)

		return c.Next()
	}
}

// GetUserID extracts user ID from context
func GetUserID(c *fiber.Ctx) string {
	userID, ok := c.Locals("user_id").(string)
	if !ok {
		return ""
	}
	return userID
}

// GetZodiacSign extracts zodiac sign from context
func GetZodiacSign(c *fiber.Ctx) string {
	zodiacSign, ok := c.Locals("zodiac_sign").(string)
	if !ok {
		return ""
	}
	return zodiacSign
}
