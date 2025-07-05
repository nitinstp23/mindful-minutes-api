# Makefile for mindful-minutes-api

# Variables
BINARY_NAME=mindful-minutes-api
GO_VERSION=1.23
MIGRATE_VERSION=v4.18.1

# Database configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=mindful_user
DB_PASSWORD=mindful_pass
DB_NAME=mindful_minutes
DB_SSL_MODE=disable
DATABASE_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)

# Test database configuration
TEST_DB_NAME=mindful_minutes_test
TEST_DATABASE_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(TEST_DB_NAME)?sslmode=$(DB_SSL_MODE)

# Coverage configuration
COVERAGE_DIR=coverage
COVERAGE_FILE=$(COVERAGE_DIR)/coverage.out
COVERAGE_HTML=$(COVERAGE_DIR)/coverage.html

.PHONY: help build test test-coverage test-coverage-html clean run dev migrate-up migrate-down migrate-create migrate-force migrate-version install-migrate-tool docker-build docker-run docker-compose-up docker-compose-down

# Default target
help: ## Show this help message
	@echo "Available commands:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build commands
build: ## Build the application
	@echo "Building $(BINARY_NAME)..."
	go build -o bin/$(BINARY_NAME) cmd/server/main.go

clean: ## Clean build artifacts and coverage files
	@echo "Cleaning up..."
	rm -rf bin/
	rm -rf $(COVERAGE_DIR)/
	rm -f coverage.out coverage.html

# Run commands
run: build ## Build and run the application
	@echo "Running $(BINARY_NAME)..."
	./bin/$(BINARY_NAME)

dev: ## Run the application in development mode
	@echo "Running in development mode..."
	go run cmd/server/main.go

# Test commands
test: ## Run all tests
	@echo "Running tests..."
	go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@mkdir -p $(COVERAGE_DIR)
	go test -v -coverprofile=$(COVERAGE_FILE) ./...
	go tool cover -func=$(COVERAGE_FILE)

test-coverage-html: test-coverage ## Run tests with coverage and generate HTML report
	@echo "Generating HTML coverage report..."
	go tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Coverage report generated: $(COVERAGE_HTML)"
	@echo "Open file://$(shell pwd)/$(COVERAGE_HTML) in your browser"

# Database migration commands
install-migrate-tool: ## Install golang-migrate tool if not present
	@echo "Checking for golang-migrate..."
	@which migrate > /dev/null || (echo "Installing golang-migrate..." && \
		curl -L https://github.com/golang-migrate/migrate/releases/download/$(MIGRATE_VERSION)/migrate.linux-amd64.tar.gz | tar xvz && \
		sudo mv migrate /usr/local/bin/migrate || \
		echo "Note: If installation failed, please install golang-migrate manually from https://github.com/golang-migrate/migrate")

migrate-up: ## Run database migrations up
	@echo "Running database migrations up..."
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down: ## Run database migrations down (rollback one migration)
	@echo "Rolling back one migration..."
	migrate -path migrations -database "$(DATABASE_URL)" down 1

migrate-create: ## Create a new migration file (usage: make migrate-create NAME=migration_name)
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required. Usage: make migrate-create NAME=migration_name"; \
		exit 1; \
	fi
	@echo "Creating new migration: $(NAME)"
	migrate create -ext sql -dir migrations -seq $(NAME)

migrate-force: ## Force migration version (usage: make migrate-force VERSION=1)
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION is required. Usage: make migrate-force VERSION=1"; \
		exit 1; \
	fi
	@echo "Forcing migration version to $(VERSION)..."
	migrate -path migrations -database "$(DATABASE_URL)" force $(VERSION)

migrate-version: ## Show current migration version
	@echo "Current migration version:"
	migrate -path migrations -database "$(DATABASE_URL)" version

# Test database migration commands
migrate-test-up: ## Run database migrations up for test database
	@echo "Running test database migrations up..."
	migrate -path migrations -database "$(TEST_DATABASE_URL)" up

migrate-test-down: ## Run database migrations down for test database
	@echo "Rolling back test database migration..."
	migrate -path migrations -database "$(TEST_DATABASE_URL)" down 1

migrate-test-reset: ## Reset test database (down all, then up)
	@echo "Resetting test database..."
	-migrate -path migrations -database "$(TEST_DATABASE_URL)" down -all
	migrate -path migrations -database "$(TEST_DATABASE_URL)" up

# Docker commands
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME) .

docker-run: ## Run application in Docker container
	@echo "Running Docker container..."
	docker run -p 8080:8080 --env-file .env $(BINARY_NAME)

docker-compose-up: ## Start services with docker-compose
	@echo "Starting services with docker-compose..."
	docker-compose up -d

docker-compose-down: ## Stop services with docker-compose
	@echo "Stopping services with docker-compose..."
	docker-compose down

# Linting and formatting
fmt: ## Format Go code
	@echo "Formatting Go code..."
	go fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

# Dependencies
mod-tidy: ## Run go mod tidy
	@echo "Running go mod tidy..."
	go mod tidy

mod-download: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download

# Quality checks
quality: fmt vet test-coverage ## Run all quality checks (format, vet, test with coverage)

# Development workflow
setup: mod-download migrate-up ## Setup development environment
	@echo "Development environment setup complete!"

ci: quality ## Run CI pipeline (quality checks)
	@echo "CI pipeline completed successfully!"

# Environment specific commands
env-example: ## Create .env.example file
	@echo "Creating .env.example file..."
	@echo "# Server Configuration" > .env.example
	@echo "PORT=8080" >> .env.example
	@echo "GIN_MODE=release" >> .env.example
	@echo "" >> .env.example
	@echo "# Database Configuration" >> .env.example
	@echo "DATABASE_URL=postgres://mindful_user:mindful_pass@localhost:5432/mindful_minutes?sslmode=disable" >> .env.example
	@echo "" >> .env.example
	@echo "# Authentication Configuration" >> .env.example
	@echo "CLERK_SECRET_KEY=your_clerk_secret_key_here" >> .env.example
	@echo "CLERK_VERIFY_URL=https://api.clerk.com/v1/verify_token" >> .env.example
	@echo "" >> .env.example
	@echo "# Environment" >> .env.example
	@echo "ENVIRONMENT=development" >> .env.example
	@echo ".env.example created!"

# Show current configuration
show-config: ## Show current configuration
	@echo "Current configuration:"
	@echo "  Binary name: $(BINARY_NAME)"
	@echo "  Go version: $(GO_VERSION)"
	@echo "  Database URL: $(DATABASE_URL)"
	@echo "  Test Database URL: $(TEST_DATABASE_URL)"
	@echo "  Coverage file: $(COVERAGE_FILE)"
	@echo "  Coverage HTML: $(COVERAGE_HTML)"