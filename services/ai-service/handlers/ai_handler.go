package handlers

import (
	"log"
	
	"zodiac-ai-backend/pkg/response"
	"zodiac-ai-backend/services/ai-service/client"

	"github.com/gofiber/fiber/v2"
)

// AIHandler handles AI-related HTTP requests
type AIHandler struct {
	geminiClient *client.GeminiClient
}

// NewAIHandler creates a new AI handler
func NewAIHandler(geminiClient *client.GeminiClient) *AIHandler {
	return &AIHandler{
		geminiClient: geminiClient,
	}
}

// GenerateChatResponse generates AI chat response
// POST /ai/chat
func (h *AIHandler) GenerateChatResponse(c *fiber.Ctx) error {
	var req struct {
		ZodiacSign  string `json:"zodiac_sign" validate:"required"`
		UserMessage string `json:"user_message" validate:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", nil)
	}

	log.Printf("🎯 AI Handler received request - Zodiac: %s, Message: %.50s...", req.ZodiacSign, req.UserMessage)

	aiResponse, err := h.geminiClient.GenerateChatResponse(
		c.Context(),
		req.ZodiacSign,
		req.UserMessage,
	)
	if err != nil {
		log.Printf("❌ AI Handler failed to generate response: %v", err)
		return response.InternalServerError(c, "Failed to generate AI response")
	}

	log.Printf("✅ AI Handler returning response: %.100s...", aiResponse)
	return response.Success(c, "AI response generated", fiber.Map{
		"response": aiResponse,
	})
}

// GenerateInsight generates insight from chat history
// POST /ai/insight
func (h *AIHandler) GenerateInsight(c *fiber.Ctx) error {
	var req struct {
		ChatHistory string `json:"chat_history" validate:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", nil)
	}

	insight, err := h.geminiClient.GenerateInsight(
		c.Context(),
		req.ChatHistory,
	)
	if err != nil {
		return response.InternalServerError(c, "Failed to generate insight")
	}

	return response.Success(c, "Insight generated", fiber.Map{
		"insight": insight,
	})
}
