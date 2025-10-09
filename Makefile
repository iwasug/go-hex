.PHONY: help run build test migrate-up migrate-down migrate-create clean dev

# Default target
help:
	@echo "Available commands:"
	@echo "  make run           - Run the application"
	@echo "  make dev           - Run the application with auto-reload (requires air)"
	@echo "  make build         - Build the application"
	@echo "  make test          - Run tests"
	@echo "  make migrate-up    - Run database migrations up"
	@echo "  make migrate-down  - Rollback database migrations"
	@echo "  make migrate-create NAME=<name> - Create a new migration file"
	@echo "  make clean         - Clean build artifacts"

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
