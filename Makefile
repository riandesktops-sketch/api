.PHONY: help setup migrate dev docker-up docker-down test clean deploy-koyeb build-all-in-one

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

setup: ## Install dependencies
	@echo "📦 Installing dependencies..."
	go mod download
	go mod tidy
	@echo "✅ Dependencies installed"

migrate: ## Run database migrations (create indexes)
	@echo "🔄 Running migrations..."
	go run scripts/migrate.go
	@echo "✅ Migrations completed"

# Development - Individual Services
dev-auth: ## Run Auth Service only
	@echo "🚀 Starting Auth Service..."
	cd services/auth-service && go run main.go

dev-ai: ## Run AI Service only
	@echo "🚀 Starting AI Service..."
	cd services/ai-service && go run main.go

dev-chat: ## Run Chat Service only
	@echo "🚀 Starting Chat Service..."
	cd services/chat-service && go run main.go

dev-social: ## Run Social Service only
	@echo "🚀 Starting Social Service..."
	cd services/social-service && go run main.go

dev-gateway: ## Run API Gateway only
	@echo "🚀 Starting API Gateway..."
	cd api-gateway && go run main.go

# Development - All-in-One
dev-all-in-one: ## Run all services in one app (Koyeb mode)
	@echo "🚀 Starting All-in-One server..."
	go run cmd/all-in-one/main.go

build-all-in-one: ## Build all-in-one binary
	@echo "🔨 Building all-in-one binary..."
	go build -o bin/zodiac-ai-all-in-one ./cmd/all-in-one
	@echo "✅ Binary created: bin/zodiac-ai-all-in-one"

# Docker
docker-build: ## Build Docker images
	@echo "🐳 Building Docker image..."
	docker build -t zodiac-ai-backend .
	@echo "✅ Docker image built"

docker-run: ## Run Docker container
	@echo "🐳 Running Docker container..."
	docker run -p 8080:8080 --env-file .env zodiac-ai-backend

docker-up: ## Start all services with Docker Compose
	@echo "🐳 Starting services with Docker Compose..."
	docker-compose up -d
	@echo "✅ Services started"

docker-down: ## Stop all services
	@echo "🛑 Stopping services..."
	docker-compose down
	@echo "✅ Services stopped"

docker-logs: ## View logs from all services
	docker-compose logs -f

docker-status: ## Check status of all services
	@echo "📊 Service Status:"
	@docker-compose ps

# Testing
test: ## Run tests
	@echo "🧪 Running tests..."
	go test -v ./...
	@echo "✅ Tests completed"

test-coverage: ## Run tests with coverage
	@echo "🧪 Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report: coverage.html"

# Deployment
deploy-koyeb: ## Deploy to Koyeb
	@echo "🚀 Deploying to Koyeb..."
	@echo "1. Push to GitHub: git push origin main"
	@echo "2. Go to https://koyeb.com"
	@echo "3. Create Web Service from GitHub repo"
	@echo "4. Build: go build -o app ./cmd/all-in-one"
	@echo "5. Run: ./app"
	@echo "📖 See KOYEB_DEPLOYMENT.md for detailed instructions"

deploy-check: ## Check if ready for deployment
	@echo "🔍 Checking deployment readiness..."
	@echo "✅ Checking Go version..."
	@go version
	@echo "✅ Checking dependencies..."
	@go mod verify
	@echo "✅ Checking all-in-one build..."
	@go build -o /dev/null ./cmd/all-in-one
	@echo "✅ All checks passed! Ready for deployment."

# Utilities
clean: ## Clean build artifacts
	@echo "🧹 Cleaning..."
	rm -rf bin/
	rm -rf dist/
	rm -f coverage.out coverage.html
	go clean
	@echo "✅ Cleaned"

lint: ## Run linter
	@echo "🔍 Running linter..."
	golangci-lint run ./...
	@echo "✅ Linting completed"

install-tools: ## Install development tools
	@echo "🔧 Installing tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "✅ Tools installed"

# Quick start
quick-start: setup migrate dev-all-in-one ## Setup and run all-in-one (quickest way to start)
