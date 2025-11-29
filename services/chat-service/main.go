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
	"zodiac-ai-backend/pkg/database"
	"zodiac-ai-backend/pkg/jwt"
	"zodiac-ai-backend/pkg/middleware"
	"zodiac-ai-backend/services/chat-service/handlers"
	"zodiac-ai-backend/services/chat-service/repositories"
	"zodiac-ai-backend/services/chat-service/services"
	"zodiac-ai-backend/services/chat-service/websocket"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Connect to MongoDB
	_, err := database.Connect(database.MongoConfig{
		URI:      cfg.MongoURI,
		Database: cfg.MongoDatabase,
	})
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer database.Disconnect()

	db := database.GetDatabase(cfg.MongoDatabase)

	// Initialize JWT manager
	jwtManager := jwt.NewManager(
		cfg.JWTSecret,
		cfg.JWTAccessExpiry,
		cfg.JWTRefreshExpiry,
	)

	// Initialize repositories
	sessionRepo := repositories.NewChatSessionRepository(db)
	messageRepo := repositories.NewMessageRepository(db)
	roomRepo := repositories.NewRoomRepository(db)

	// Initialize services
	chatService := services.NewChatService(sessionRepo, messageRepo, cfg.AIServiceURL)

	// Initialize WebSocket Hub
	hub := websocket.NewHub()
	go hub.Run()

	// Initialize handlers
	chatHandler := handlers.NewChatHandler(chatService)
	roomHandler := handlers.NewRoomHandler(roomRepo, hub)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "Zodiac AI - Chat Service",
		ErrorHandler: customErrorHandler,
	})

	// Middleware
	app.Use(middleware.SetupRecover())
	app.Use(middleware.SetupLogger())
	app.Use(middleware.SetupCORS())
	app.Use(middleware.RequestID())

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := database.HealthCheck(ctx); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status": "unhealthy",
				"error":  err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"status":  "healthy",
			"service": "chat-service",
		})
	})

	// Routes
	api := app.Group("/api/v1")

	// Chat routes (protected)
	chat := api.Group("/chat")
	chat.Use(middleware.AuthMiddleware(jwtManager))

	chat.Post("/sessions", chatHandler.CreateSession)
	chat.Get("/sessions", chatHandler.GetSessions)
	chat.Post("/sessions/:id/messages", chatHandler.SendMessage)
	chat.Get("/sessions/:id/messages", chatHandler.GetMessages)
	chat.Post("/sessions/:id/generate-insight", chatHandler.GenerateInsight)

	// Room routes
	rooms := api.Group("/rooms")
	rooms.Use(middleware.AuthMiddleware(jwtManager))
	rooms.Post("", roomHandler.CreateRoom)
	rooms.Get("", roomHandler.GetRooms)

	// WebSocket route
	// Note: Middleware is applied inside the handler for WebSocket upgrade
	app.Get("/rooms/:id/ws", websocket.New(roomHandler.JoinRoom))

	// Start server
	port := cfg.ChatServicePort
	log.Printf("🚀 Chat Service starting on port %s", port)

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

	log.Println("🛑 Shutting down Chat Service...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Fatalf("Failed to shutdown server: %v", err)
	}

	log.Println("✅ Chat Service stopped gracefully")
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
