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
	"zodiac-ai-backend/services/social-service/handlers"
	"zodiac-ai-backend/services/social-service/repositories"
	"zodiac-ai-backend/services/social-service/services"

	"github.com/gofiber/fiber/v2"
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
	postRepo := repositories.NewPostRepository(db)
	commentRepo := repositories.NewCommentRepository(db)

	// Initialize services
	socialService := services.NewSocialService(postRepo, commentRepo)

	// Initialize handlers
	socialHandler := handlers.NewSocialHandler(socialService)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "Zodiac AI - Social Service",
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
			"service": "social-service",
		})
	})

	// Routes
	api := app.Group("/api/v1")

	// Post routes
	posts := api.Group("/posts")
	
	// Public routes
	posts.Get("", socialHandler.GetFeed)
	posts.Get("/:id", socialHandler.GetPost)
	posts.Get("/:id/comments", socialHandler.GetComments)

	// Protected routes
	posts.Use(middleware.AuthMiddleware(jwtManager))
	posts.Post("", socialHandler.PublishPost)
	posts.Post("/:id/like", socialHandler.LikePost)
	posts.Delete("/:id/like", socialHandler.UnlikePost)
	posts.Post("/:id/comments", socialHandler.AddComment)

	// Start server
	port := cfg.SocialServicePort
	log.Printf("🚀 Social Service starting on port %s", port)

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

	log.Println("🛑 Shutting down Social Service...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Fatalf("Failed to shutdown server: %v", err)
	}

	log.Println("✅ Social Service stopped gracefully")
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
