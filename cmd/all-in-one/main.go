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

	// Auth
	authHandlers "zodiac-ai-backend/services/auth-service/handlers"
	authRepos "zodiac-ai-backend/services/auth-service/repositories"
	authServices "zodiac-ai-backend/services/auth-service/services"

	// Chat
	chatHandlers "zodiac-ai-backend/services/chat-service/handlers"
	chatRepos "zodiac-ai-backend/services/chat-service/repositories"
	chatServices "zodiac-ai-backend/services/chat-service/services"
	"zodiac-ai-backend/services/chat-service/websocket"

	// Social
	socialHandlers "zodiac-ai-backend/services/social-service/handlers"
	socialRepos "zodiac-ai-backend/services/social-service/repositories"
	socialServices "zodiac-ai-backend/services/social-service/services"

	// AI
	aiHandlers "zodiac-ai-backend/services/ai-service/handlers"
	"zodiac-ai-backend/services/ai-service/client"

	"github.com/gofiber/fiber/v2"
	ws "github.com/gofiber/websocket/v2"
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

	// Initialize rate limiter
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimitRequests, cfg.RateLimitWindow)

	// ========== AUTH SERVICE ==========
	userRepo := authRepos.NewUserRepository(db)
	refreshTokenRepo := authRepos.NewRefreshTokenRepository(db)
	friendshipRepo := authRepos.NewFriendshipRepository(db)

	authService := authServices.NewAuthService(userRepo, refreshTokenRepo, jwtManager)
	friendshipService := authServices.NewFriendshipService(friendshipRepo, userRepo)

	authHandler := authHandlers.NewAuthHandler(authService)
	friendHandler := authHandlers.NewFriendHandler(friendshipService)

	// ========== AI SERVICE ==========
	geminiClient, err := client.NewGeminiClient(cfg.GeminiAPIKey)
	if err != nil {
		log.Printf("Warning: Failed to initialize Gemini client: %v", err)
	}
	defer geminiClient.Close()

	aiHandler := aiHandlers.NewAIHandler(geminiClient)

	// ========== CHAT SERVICE ==========
	sessionRepo := chatRepos.NewChatSessionRepository(db)
	messageRepo := chatRepos.NewMessageRepository(db)
	roomRepo := chatRepos.NewRoomRepository(db)

	// AI service URL is internal (same app) - use direct handler call instead of HTTP
	// For simplicity, we'll keep HTTP but use localhost
	aiServiceURL := os.Getenv("AI_SERVICE_URL")
	if aiServiceURL == "" {
		aiServiceURL = "http://localhost:8080"
	}
	chatService := chatServices.NewChatService(sessionRepo, messageRepo, aiServiceURL)

	chatHandler := chatHandlers.NewChatHandler(chatService)

	// WebSocket Hub
	hub := websocket.NewHub()
	go hub.Run()

	roomHandler := chatHandlers.NewRoomHandler(roomRepo, hub)

	// ========== SOCIAL SERVICE ==========
	postRepo := socialRepos.NewPostRepository(db)
	commentRepo := socialRepos.NewCommentRepository(db)

	socialService := socialServices.NewSocialService(postRepo, commentRepo)

	socialHandler := socialHandlers.NewSocialHandler(socialService)

	// ========== FIBER APP ==========
	app := fiber.New(fiber.Config{
		AppName:      "Zodiac AI - All-in-One",
		ErrorHandler: customErrorHandler,
	})

	// Global middleware
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
			"service": "zodiac-ai-all-in-one",
		})
	})

	// API routes
	api := app.Group("/api/v1")

	// ========== AUTH ROUTES ==========
	auth := api.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh", authHandler.RefreshToken)

	// User routes (protected)
	users := api.Group("/users")
	users.Use(middleware.AuthMiddleware(jwtManager))
	users.Use(rateLimiter.RateLimitMiddleware())
	users.Get("/me", authHandler.GetProfile)
	users.Put("/me", authHandler.UpdateProfile)

	// Friend routes (protected)
	friends := api.Group("/friends")
	friends.Use(middleware.AuthMiddleware(jwtManager))
	friends.Use(rateLimiter.RateLimitMiddleware())
	friends.Post("/requests", friendHandler.SendFriendRequest)
	friends.Put("/requests/:id", friendHandler.AcceptRejectRequest)
	friends.Get("", friendHandler.GetFriends)
	friends.Get("/status/:user_id", friendHandler.CheckFriendshipStatus)

	// ========== CHAT ROUTES ==========
	chat := api.Group("/chat")
	chat.Use(middleware.AuthMiddleware(jwtManager))
	chat.Use(rateLimiter.RateLimitMiddleware())
	chat.Post("/sessions", chatHandler.CreateSession)
	chat.Get("/sessions", chatHandler.GetSessions)
	chat.Post("/sessions/:id/messages", chatHandler.SendMessage)
	chat.Get("/sessions/:id/messages", chatHandler.GetMessages)
	chat.Post("/sessions/:id/generate-insight", chatHandler.GenerateInsight)

	// ========== ROOM ROUTES ==========
	rooms := api.Group("/rooms")
	rooms.Use(middleware.AuthMiddleware(jwtManager))
	rooms.Use(rateLimiter.RateLimitMiddleware())
	rooms.Post("", roomHandler.CreateRoom)
	rooms.Get("", roomHandler.GetRooms)

	// WebSocket route
	app.Get("/api/v1/rooms/:id/ws", ws.New(func(c *ws.Conn) {
		roomHandler.JoinRoom(c)
	}))

	// ========== SOCIAL ROUTES ==========
	posts := api.Group("/posts")

	// Public routes
	posts.Get("", socialHandler.GetFeed)
	posts.Get("/:id", socialHandler.GetPost)
	posts.Get("/:id/comments", socialHandler.GetComments)

	// Protected routes
	postsProtected := posts.Group("")
	postsProtected.Use(middleware.AuthMiddleware(jwtManager))
	postsProtected.Use(rateLimiter.RateLimitMiddleware())
	postsProtected.Post("", socialHandler.PublishPost)
	postsProtected.Post("/:id/like", socialHandler.LikePost)
	postsProtected.Delete("/:id/like", socialHandler.UnlikePost)
	postsProtected.Post("/:id/comments", socialHandler.AddComment)

	// ========== AI ROUTES (Internal) ==========
	ai := api.Group("/ai")
	ai.Post("/chat", aiHandler.GenerateChatResponse)
	ai.Post("/insight", aiHandler.GenerateInsight)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Zodiac AI All-in-One starting on port %s", port)
	log.Printf("📡 All services running in single application")

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

	log.Println("🛑 Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Fatalf("Failed to shutdown server: %v", err)
	}

	log.Println("✅ Server stopped gracefully")
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
