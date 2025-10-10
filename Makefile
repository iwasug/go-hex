.PHONY: help run build test test-coverage lint migrate-up migrate-down migrate-create clean dev docker-build docker-up docker-down swagger

# Load environment variables from .env file if it exists
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Default target
help:
	@echo "Available commands:"
	@echo "  make run                       - Run the application"
	@echo "  make dev                       - Run the application with auto-reload (requires air)"
	@echo "  make build                     - Build the application"
	@echo "  make test                      - Run tests"
	@echo "  make test-coverage             - Run tests with coverage report"
	@echo "  make lint                      - Run linter (requires golangci-lint)"
	@echo "  make swagger                   - Generate Swagger documentation"
	@echo "  make migrate-up                - Run database migrations up"
	@echo "  make migrate-down              - Rollback database migrations"
	@echo "  make migrate-create NAME=<name> - Create a new migration file"
	@echo "  make docker-build              - Build Docker image"
	@echo "  make docker-up                 - Start services with docker-compose"
	@echo "  make docker-down               - Stop services with docker-compose"
	@echo "  make clean                     - Clean build artifacts"

# Run the application
run:
	go run cmd/api/main.go

# Run the application in development mode with auto-reload
dev:
	@which air > /dev/null || (echo "air not found. Install with: go install github.com/air-verse/air@latest" && exit 1)
	air

# Build the application
build:
	go build -o bin/api cmd/api/main.go

# Run tests
test:
	go test -v -cover ./...

# Run tests with coverage report
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linter
lint:
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	golangci-lint run ./...

# Generate Swagger documentation
swagger:
	@which swag > /dev/null || (echo "swag not found. Install with: go install github.com/swaggo/swag/cmd/swag@latest" && exit 1)
	swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal
	@echo "Swagger documentation generated in docs/ directory"
	@echo "Access Swagger UI at: http://localhost:8080/swagger/index.html"

# Build Docker image
docker-build:
	docker build -t go-hex-api:latest .

# Start services with docker-compose
docker-up:
	docker-compose up -d

# Stop services with docker-compose
docker-down:
	docker-compose down

# Run database migrations up
migrate-up:
	@echo "Running migrations up..."
	@if [ -z "$(DB_HOST)" ]; then \
		echo "Error: DB_HOST not set. Load .env or set environment variables."; \
		exit 1; \
	fi
	@for file in migrations/*_*.up.sql; do \
		if [ -f "$$file" ]; then \
			echo "Executing $$file"; \
			PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME) -f $$file || exit 1; \
		fi \
	done
	@echo "Migrations completed successfully"

# Rollback database migrations
migrate-down:
	@echo "Rolling back migrations..."
	@if [ -z "$(DB_HOST)" ]; then \
		echo "Error: DB_HOST not set. Load .env or set environment variables."; \
		exit 1; \
	fi
	@for file in $$(ls -r migrations/*_*.down.sql); do \
		if [ -f "$$file" ]; then \
			echo "Executing $$file"; \
			PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME) -f $$file || exit 1; \
		fi \
	done
	@echo "Rollback completed successfully"

# Create a new migration file
migrate-create:
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required. Usage: make migrate-create NAME=create_users_table"; \
		exit 1; \
	fi
	@timestamp=$$(date +%Y%m%d%H%M%S); \
	up_file="migrations/$${timestamp}_$(NAME).up.sql"; \
	down_file="migrations/$${timestamp}_$(NAME).down.sql"; \
	touch $$up_file $$down_file; \
	echo "Created migration files:"; \
	echo "  $$up_file"; \
	echo "  $$down_file"

# Clean build artifacts
clean:
	rm -rf bin/
	go clean
