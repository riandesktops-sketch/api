package handlers

import (
	"context"
	"log"
	"time"
	
	"zodiac-ai-backend/pkg/queue"
	"zodiac-ai-backend/pkg/response"
	"zodiac-ai-backend/services/ai-service/client"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// AIHandler handles AI-related HTTP requests
type AIHandler struct {
	geminiClient *client.GeminiClient
	requestQueue *queue.RequestQueue
}

// NewAIHandler creates a new AI handler
func NewAIHandler(geminiClient *client.GeminiClient, requestQueue *queue.RequestQueue) *AIHandler {
	return &AIHandler{
		geminiClient: geminiClient,
		requestQueue: requestQueue,
	}
}

// GenerateChatResponse generates AI chat response using request queue
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

	// Create request ID
	requestID := uuid.New().String()

	// Create request data
	requestData := map[string]interface{}{
		"zodiac_sign":  req.ZodiacSign,
		"user_message": req.UserMessage,
	}

	// Create result channel
	resultChan := make(chan queue.Result, 1)

	// Create queue request
	queueReq := &queue.Request{
		ID:        requestID,
		Data:      requestData,
		Context:   c.Context(),
		Result:    resultChan,
		EnqueueAt: time.Now(),
	}

	// Enqueue request
	if err := h.requestQueue.Enqueue(queueReq); err != nil {
		if err == queue.ErrQueueFull {
			log.Printf("❌ Queue full - rejecting request %s", requestID)
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"message": "Server is busy, please try again later",
			})
		}
		log.Printf("❌ Failed to enqueue request %s: %v", requestID, err)
		return response.InternalServerError(c, "Failed to process request")
	}

	log.Printf("📥 Request %s enqueued, waiting for result...", requestID)

	// Wait for result with timeout (60 seconds)
	ctx, cancel := context.WithTimeout(c.Context(), 60*time.Second)
	defer cancel()

	select {
	case result := <-resultChan:
		if result.Error != nil {
			log.Printf("❌ Request %s failed: %v", requestID, result.Error)
			
			// Check if it's a fallback response (still return success)
			if aiResponse, ok := result.Data.(string); ok && aiResponse != "" {
				log.Printf("✅ Request %s returning fallback response", requestID)
				return response.Success(c, "AI response generated (fallback)", fiber.Map{
					"response": aiResponse,
				})
			}
			
			return response.InternalServerError(c, "Failed to generate AI response")
		}

		aiResponse := result.Data.(string)
		log.Printf("✅ Request %s completed: %.100s...", requestID, aiResponse)
		return response.Success(c, "AI response generated", fiber.Map{
			"response": aiResponse,
		})

	case <-ctx.Done():
		log.Printf("⏱️ Request %s timeout", requestID)
		return c.Status(fiber.StatusRequestTimeout).JSON(fiber.Map{
			"success": false,
			"message": "Request timeout - please try again",
		})
	}
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
