package handlers

import (
	"zodiac-ai-backend/pkg/middleware"
	"zodiac-ai-backend/pkg/response"
	"zodiac-ai-backend/services/chat-service/models"
	"zodiac-ai-backend/services/chat-service/services"

	"github.com/gofiber/fiber/v2"
)

// ChatHandler handles chat HTTP requests
type ChatHandler struct {
	chatService *services.ChatService
}

// NewChatHandler creates a new chat handler
func NewChatHandler(chatService *services.ChatService) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
	}
}

// CreateSession creates a new chat session
// POST /chat/sessions
func (h *ChatHandler) CreateSession(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return response.Unauthorized(c, "User not authenticated")
	}

	var req struct {
		Title string `json:"title"`
	}
	c.BodyParser(&req) // Optional title

	session, err := h.chatService.CreateSession(c.Context(), userID, req.Title)
	if err != nil {
		return response.InternalServerError(c, "Failed to create session")
	}

	return response.Created(c, "Chat session created", session)
}

// SendMessage sends a message to AI
// POST /chat/sessions/:id/messages
func (h *ChatHandler) SendMessage(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	zodiacSign := middleware.GetZodiacSign(c)
	if userID == "" {
		return response.Unauthorized(c, "User not authenticated")
	}

	sessionID := c.Params("id")
	if sessionID == "" {
		return response.BadRequest(c, "Session ID required", nil)
	}

	var req models.SendMessageRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", nil)
	}

	messageResp, err := h.chatService.SendMessage(c.Context(), sessionID, userID, zodiacSign, req.Message)
	if err != nil {
		if err == services.ErrSessionNotFound {
			return response.NotFound(c, "Chat session not found")
		}
		if err == services.ErrAIServiceDown {
			return response.ServiceUnavailable(c, "AI service temporarily unavailable")
		}
		return response.InternalServerError(c, "Failed to send message")
	}

	return response.Success(c, "Message sent successfully", messageResp)
}

// GetMessages gets chat history with pagination
// GET /chat/sessions/:id/messages
func (h *ChatHandler) GetMessages(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return response.Unauthorized(c, "User not authenticated")
	}

	sessionID := c.Params("id")
	if sessionID == "" {
		return response.BadRequest(c, "Session ID required", nil)
	}

	cursor := c.Query("cursor", "")
	limit := c.QueryInt("limit", 20)

	messages, nextCursor, err := h.chatService.GetMessages(c.Context(), sessionID, userID, cursor, limit)
	if err != nil {
		if err == services.ErrSessionNotFound {
			return response.NotFound(c, "Chat session not found")
		}
		return response.InternalServerError(c, "Failed to get messages")
	}

	meta := &response.MetaData{
		NextCursor: nextCursor,
		HasMore:    nextCursor != "",
		Limit:      limit,
	}

	return response.SuccessWithMeta(c, "Messages retrieved successfully", messages, meta)
}

// GenerateInsight generates insight from chat
// POST /chat/sessions/:id/generate-insight
func (h *ChatHandler) GenerateInsight(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return response.Unauthorized(c, "User not authenticated")
	}

	sessionID := c.Params("id")
	if sessionID == "" {
		return response.BadRequest(c, "Session ID required", nil)
	}

	insight, err := h.chatService.GenerateInsight(c.Context(), sessionID, userID)
	if err != nil {
		if err == services.ErrSessionNotFound {
			return response.NotFound(c, "Chat session not found")
		}
		if err == services.ErrAIServiceDown {
			return response.ServiceUnavailable(c, "AI service temporarily unavailable")
		}
		return response.InternalServerError(c, "Failed to generate insight")
	}

	return response.Success(c, "Insight generated successfully", fiber.Map{
		"insight": insight,
	})
}

// GetSessions gets all user's chat sessions
// GET /chat/sessions
func (h *ChatHandler) GetSessions(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return response.Unauthorized(c, "User not authenticated")
	}

	sessions, err := h.chatService.GetUserSessions(c.Context(), userID)
	if err != nil {
		return response.InternalServerError(c, "Failed to get sessions")
	}

	return response.Success(c, "Sessions retrieved successfully", sessions)
}
