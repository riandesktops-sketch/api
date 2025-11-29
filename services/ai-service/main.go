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

	// Initialize handlers
	aiHandler := handlers.NewAIHandler(geminiClient)

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
		return c.JSON(fiber.Map{
			"status":  "healthy",
			"service": "ai-service",
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
