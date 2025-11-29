package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"zodiac-ai-backend/api-gateway/proxy"
	"zodiac-ai-backend/pkg/config"
	"zodiac-ai-backend/pkg/jwt"
	"zodiac-ai-backend/pkg/middleware"

	"github.com/gofiber/fiber/v2"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize JWT manager for auth middleware
	jwtManager := jwt.NewManager(
		cfg.JWTSecret,
		cfg.JWTAccessExpiry,
		cfg.JWTRefreshExpiry,
	)

	// Initialize service proxy
	serviceProxy := proxy.NewServiceProxy(
		cfg.AuthServiceURL,
		cfg.ChatServiceURL,
		cfg.SocialServiceURL,
		cfg.AIServiceURL,
	)

	// Initialize rate limiter
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimitRequests, cfg.RateLimitWindow)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "Zodiac AI - API Gateway",
		ErrorHandler: customErrorHandler,
	})

	// Global middleware
	app.Use(middleware.SetupRecover())
	app.Use(middleware.SetupLogger())
	app.Use(middleware.SetupCORS())
	app.Use(middleware.RequestID())

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "healthy",
			"service": "api-gateway",
		})
	})

	// API routes
	api := app.Group("/api/v1")

	// Auth routes (public)
	auth := api.Group("/auth")
	auth.All("/*", serviceProxy.ProxyToAuth)

	// User routes (protected)
	users := api.Group("/users")
	users.Use(middleware.AuthMiddleware(jwtManager))
	users.Use(rateLimiter.RateLimitMiddleware())
	users.All("/*", serviceProxy.ProxyToAuth)

	// Friend routes (protected)
	friends := api.Group("/friends")
	friends.Use(middleware.AuthMiddleware(jwtManager))
	friends.Use(rateLimiter.RateLimitMiddleware())
	friends.All("/*", serviceProxy.ProxyToAuth)

	// Chat routes (protected)
	chat := api.Group("/chat")
	chat.Use(middleware.AuthMiddleware(jwtManager))
	chat.Use(rateLimiter.RateLimitMiddleware())
	chat.All("/*", serviceProxy.ProxyToChat)

	// Room routes (protected)
	rooms := api.Group("/rooms")
	rooms.Use(middleware.AuthMiddleware(jwtManager))
	rooms.Use(rateLimiter.RateLimitMiddleware())
	rooms.All("/*", serviceProxy.ProxyToChat)

	// Post routes (mixed: public read, protected write)
	posts := api.Group("/posts")
	
	// Public routes (no auth)
	posts.Get("", serviceProxy.ProxyToSocial)
	posts.Get("/:id", serviceProxy.ProxyToSocial)
	posts.Get("/:id/comments", serviceProxy.ProxyToSocial)

	// Protected routes
	postsProtected := posts.Group("")
	postsProtected.Use(middleware.AuthMiddleware(jwtManager))
	postsProtected.Use(rateLimiter.RateLimitMiddleware())
	postsProtected.Post("", serviceProxy.ProxyToSocial)
	postsProtected.Post("/:id/like", serviceProxy.ProxyToSocial)
	postsProtected.Delete("/:id/like", serviceProxy.ProxyToSocial)
	postsProtected.Post("/:id/comments", serviceProxy.ProxyToSocial)

	// AI routes (internal - no rate limit for service-to-service)
	ai := api.Group("/ai")
	ai.All("/*", serviceProxy.ProxyToAI)

	// Start server
	port := cfg.APIGatewayPort
	log.Printf("🚀 API Gateway starting on port %s", port)
	log.Printf("📡 Routing to services:")
	log.Printf("   - Auth Service: %s", cfg.AuthServiceURL)
	log.Printf("   - Chat Service: %s", cfg.ChatServiceURL)
	log.Printf("   - Social Service: %s", cfg.SocialServiceURL)
	log.Printf("   - AI Service: %s", cfg.AIServiceURL)

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

	log.Println("🛑 Shutting down API Gateway...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Fatalf("Failed to shutdown server: %v", err)
	}

	log.Println("✅ API Gateway stopped gracefully")
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
