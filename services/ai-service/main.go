package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"zodiac-ai-backend/pkg/config"
	"zodiac-ai-backend/pkg/middleware"
	"zodiac-ai-backend/pkg/queue"
	"zodiac-ai-backend/services/ai-service/client"
	"zodiac-ai-backend/services/ai-service/handlers"

	"github.com/gofiber/fiber/v2"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Validate Gemini API key
	if cfg.GeminiAPIKey == "" {
		log.Fatal("GEMINI_API_KEY environment variable is required")
	}

	// Initialize Gemini client
	geminiClient, err := client.NewGeminiClient(cfg.GeminiAPIKey)
	if err != nil {
		log.Fatalf("Failed to initialize Gemini client: %v", err)
	}
	defer geminiClient.Close()

	// Initialize request queue with processor
	requestQueue := queue.NewRequestQueue(queue.Config{
		QueueSize: 1000, // Buffer for 1000 requests
		Workers:   10,   // 10 concurrent workers
		Processor: func(ctx context.Context, data interface{}) (interface{}, error) {
			// Extract request data
			reqData := data.(map[string]interface{})
			zodiacSign := reqData["zodiac_sign"].(string)
			userMessage := reqData["user_message"].(string)

			// Generate AI response
			response, err := geminiClient.GenerateChatResponse(ctx, zodiacSign, userMessage)
			return response, err
		},
	})

	// Start queue workers
	requestQueue.Start()
	defer func() {
		log.Println("🛑 Stopping request queue...")
		if err := requestQueue.Stop(30 * time.Second); err != nil {
			log.Printf("⚠️ Error stopping queue: %v", err)
		}
	}()

	// Initialize handlers
	aiHandler := handlers.NewAIHandler(geminiClient, requestQueue)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "Zodiac AI - AI Service",
		ErrorHandler: customErrorHandler,
	})

	// Middleware
	app.Use(middleware.SetupRecover())
	app.Use(middleware.SetupLogger())
	app.Use(middleware.SetupCORS())
	app.Use(middleware.RequestID())

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		stats := requestQueue.Stats()
		return c.JSON(fiber.Map{
			"status":  "healthy",
			"service": "ai-service",
			"queue":   stats,
		})
	})

	// Routes
	api := app.Group("/api/v1")

	// AI routes (internal - called by other services)
	ai := api.Group("/ai")
	ai.Post("/chat", aiHandler.GenerateChatResponse)
	ai.Post("/insight", aiHandler.GenerateInsight)

	// Start server
	port := cfg.AIServicePort
	log.Printf("🚀 AI Service starting on port %s", port)
	log.Printf("📊 Queue: %d buffer, %d workers", 1000, 10)

	// Graceful shutdown
	go func() {
		if err := app.Listen(":" + port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down AI Service...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Fatalf("Failed to shutdown server: %v", err)
	}

	log.Println("✅ AI Service stopped gracefully")
}

// customErrorHandler handles errors globally
func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	return c.Status(code).JSON(fiber.Map{
		"success": false,
		"message": message,
		"error": fiber.Map{
			"code":    fmt.Sprintf("HTTP_%d", code),
			"message": message,
		},
	})
}
