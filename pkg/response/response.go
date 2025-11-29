package response

import (
	"github.com/gofiber/fiber/v2"
)

// APIResponse represents standard API response structure
// Reference: Pragmatic Programmer - DRY Principle
// Single response format ensures consistency across all services
type APIResponse struct {
	Success bool         `json:"success"`
	Message string       `json:"message"`
	Data    interface{}  `json:"data,omitempty"`
	Error   *ErrorDetail `json:"error,omitempty"`
	Meta    *MetaData    `json:"meta,omitempty"`
}

// ErrorDetail provides detailed error information
type ErrorDetail struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// MetaData contains pagination and additional metadata
type MetaData struct {
	// Cursor-based pagination
	NextCursor string `json:"next_cursor,omitempty"`
	PrevCursor string `json:"prev_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
	
	// Additional info
	Total int `json:"total,omitempty"`
	Limit int `json:"limit,omitempty"`
}

// Success sends a successful response
func Success(c *fiber.Ctx, message string, data interface{}) error {
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// SuccessWithMeta sends a successful response with metadata
func SuccessWithMeta(c *fiber.Ctx, message string, data interface{}, meta *MetaData) error {
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

// Created sends a 201 Created response
func Created(c *fiber.Ctx, message string, data interface{}) error {
	return c.Status(fiber.StatusCreated).JSON(APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// BadRequest sends a 400 Bad Request response
func BadRequest(c *fiber.Ctx, message string, details map[string]interface{}) error {
	return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
		Success: false,
		Message: message,
		Error: &ErrorDetail{
			Code:    "BAD_REQUEST",
			Message: message,
			Details: details,
		},
	})
}

// Unauthorized sends a 401 Unauthorized response
func Unauthorized(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
		Success: false,
		Message: message,
		Error: &ErrorDetail{
			Code:    "UNAUTHORIZED",
			Message: message,
		},
	})
}

// Forbidden sends a 403 Forbidden response
func Forbidden(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusForbidden).JSON(APIResponse{
		Success: false,
		Message: message,
		Error: &ErrorDetail{
			Code:    "FORBIDDEN",
			Message: message,
		},
	})
}

// NotFound sends a 404 Not Found response
func NotFound(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusNotFound).JSON(APIResponse{
		Success: false,
		Message: message,
		Error: &ErrorDetail{
			Code:    "NOT_FOUND",
			Message: message,
		},
	})
}

// Conflict sends a 409 Conflict response
func Conflict(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusConflict).JSON(APIResponse{
		Success: false,
		Message: message,
		Error: &ErrorDetail{
			Code:    "CONFLICT",
			Message: message,
		},
	})
}

// TooManyRequests sends a 429 Too Many Requests response
func TooManyRequests(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusTooManyRequests).JSON(APIResponse{
		Success: false,
		Message: message,
		Error: &ErrorDetail{
			Code:    "TOO_MANY_REQUESTS",
			Message: message,
		},
	})
}

// InternalServerError sends a 500 Internal Server Error response
func InternalServerError(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
		Success: false,
		Message: message,
		Error: &ErrorDetail{
			Code:    "INTERNAL_SERVER_ERROR",
			Message: message,
		},
	})
}

// ServiceUnavailable sends a 503 Service Unavailable response
func ServiceUnavailable(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusServiceUnavailable).JSON(APIResponse{
		Success: false,
		Message: message,
		Error: &ErrorDetail{
			Code:    "SERVICE_UNAVAILABLE",
			Message: message,
		},
	})
}
