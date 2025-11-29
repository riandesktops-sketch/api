package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	// MongoDB
	MongoURI      string
	MongoDatabase string

	// JWT
	JWTSecret        string
	JWTAccessExpiry  time.Duration
	JWTRefreshExpiry time.Duration

	// Gemini AI
	GeminiAPIKey string

	// Service Ports
	APIGatewayPort  string
	AuthServicePort string
	ChatServicePort string
	SocialServicePort string
	AIServicePort   string

	// Service URLs (for gateway)
	AuthServiceURL   string
	ChatServiceURL   string
	SocialServiceURL string
	AIServiceURL     string

	// Rate Limiting
	RateLimitRequests int
	RateLimitWindow   time.Duration

	// Environment
	Environment string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	// Load .env file if exists (ignore error in production)
	_ = godotenv.Load()

	return &Config{
		// MongoDB
		MongoURI:      getEnv("MONGODB_URI", "mongodb://localhost:27017"),
		MongoDatabase: getEnv("MONGODB_DATABASE", "zodiac_ai"),

		// JWT
		JWTSecret:        getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-this-in-production"),
		JWTAccessExpiry:  parseDuration(getEnv("JWT_ACCESS_EXPIRY", "15m")),
		JWTRefreshExpiry: parseDuration(getEnv("JWT_REFRESH_EXPIRY", "720h")),

		// Gemini AI
		GeminiAPIKey: getEnv("GEMINI_API_KEY", ""),

		// Service Ports
		APIGatewayPort:    getEnv("API_GATEWAY_PORT", "8000"),
		AuthServicePort:   getEnv("AUTH_SERVICE_PORT", "8001"),
		ChatServicePort:   getEnv("CHAT_SERVICE_PORT", "8002"),
		SocialServicePort: getEnv("SOCIAL_SERVICE_PORT", "8003"),
		AIServicePort:     getEnv("AI_SERVICE_PORT", "8004"),

		// Service URLs
		AuthServiceURL:   getEnv("AUTH_SERVICE_URL", "http://localhost:8001"),
		ChatServiceURL:   getEnv("CHAT_SERVICE_URL", "http://localhost:8002"),
		SocialServiceURL: getEnv("SOCIAL_SERVICE_URL", "http://localhost:8003"),
		AIServiceURL:     getEnv("AI_SERVICE_URL", "http://localhost:8004"),

		// Rate Limiting
		RateLimitRequests: parseInt(getEnv("RATE_LIMIT_REQUESTS", "100")),
		RateLimitWindow:   parseDuration(getEnv("RATE_LIMIT_WINDOW", "60s")),

		// Environment
		Environment: getEnv("ENVIRONMENT", "development"),
	}
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// parseDuration parses duration string with error handling
func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		log.Printf("Warning: Invalid duration '%s', using default", s)
		return 15 * time.Minute
	}
	return d
}

// parseInt parses integer string with error handling
func parseInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		log.Printf("Warning: Invalid integer '%s', using default", s)
		return 100
	}
	return i
}
