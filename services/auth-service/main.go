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
	"zodiac-ai-backend/services/auth-service/handlers"
	"zodiac-ai-backend/services/auth-service/repositories"
	"zodiac-ai-backend/services/auth-service/services"

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
	userRepo := repositories.NewUserRepository(db)
	refreshTokenRepo := repositories.NewRefreshTokenRepository(db)

	// Initialize services
	authService := services.NewAuthService(userRepo, refreshTokenRepo, jwtManager)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "Zodiac AI - Auth Service",
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
			"service": "auth-service",
		})
	})

	// Routes
	api := app.Group("/api/v1")

	// Auth routes (public)
	auth := api.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh", authHandler.RefreshToken)

	// User routes (protected)
	users := api.Group("/users")
	users.Use(middleware.AuthMiddleware(jwtManager))
	users.Get("/me", authHandler.GetProfile)
	users.Put("/me", authHandler.UpdateProfile)

	// Start server
	port := cfg.AuthServicePort
	log.Printf("🚀 Auth Service starting on port %s", port)

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

	log.Println("🛑 Shutting down Auth Service...")

	if err := app.Shutdown(); err != nil {
		log.Fatalf("Failed to shutdown server: %v", err)
	}

	log.Println("✅ Auth Service stopped gracefully")
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
