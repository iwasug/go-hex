# Golang Chi Hexagonal Base Project

A production-ready Go application using Chi router with Hexagonal Architecture (Ports and Adapters pattern). This project provides authentication, user management, role-based access control (RBAC), caching, observability, and comprehensive middleware stack with PostgreSQL and Redis.

## Table of Contents

- [Quick Start](#quick-start)
- [Architecture](#architecture)
- [Features](#features)
- [Project Structure](#project-structure)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Configuration](#configuration)
- [Database Setup](#database-setup)
- [Running the Application](#running-the-application)
- [API Endpoints](#api-endpoints)
- [Default Roles and Permissions](#default-roles-and-permissions)
- [Development](#development)
- [Testing](#testing)
- [Monitoring & Observability](#monitoring--observability)
- [API Documentation](#api-documentation)
- [Performance & Scalability](#performance--scalability)
- [Security Considerations](#security-considerations)
- [Troubleshooting](#troubleshooting)

## Quick Start

Get up and running in 5 minutes with Docker Compose:

```bash
# Clone the repository
git clone <repository-url>
cd go-hex

# Copy environment variables
cp .env.example .env

# Start all services (PostgreSQL + API)
docker-compose up -d

# View logs
docker-compose logs -f app

# Access the API
curl http://localhost:8080/health

# Access Swagger documentation
# Open in browser: http://localhost:8080/swagger/index.html

# Access Prometheus metrics
curl http://localhost:8080/metrics
```

The application will be available at `http://localhost:8080` with all migrations applied automatically.

**Test the API:**
```bash
# Register a new user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Test123!","name":"Test User"}'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Test123!"}'
```

## Architecture

This project follows **Hexagonal Architecture** (also known as Ports and Adapters pattern), which separates business logic from infrastructure concerns:

```
┌─────────────────────────────────────────────────────────────┐
│                     Infrastructure Layer                     │
│  (HTTP Handlers, Database, External Services)               │
└─────────────────────────────────────────────────────────────┘
                            │
┌─────────────────────────────────────────────────────────────┐
│                    Application/Ports Layer                   │
│              (Interfaces for Repositories, etc.)            │
└─────────────────────────────────────────────────────────────┘
                            │
┌─────────────────────────────────────────────────────────────┐
│                        Domain Layer                          │
│              (Business Logic, Entities, Services)           │
└─────────────────────────────────────────────────────────────┘
```

### Key Benefits

- **Testability**: Business logic can be tested independently
- **Flexibility**: Easy to swap infrastructure components
- **Maintainability**: Clear separation of concerns
- **Scalability**: Well-organized codebase that scales with complexity

## Features

- ✅ **Authentication**
  - User registration with email/password
  - Login with JWT tokens (access + refresh)
  - Password reset flow
  - Token refresh mechanism
  - Logout functionality

- ✅ **User Management**
  - List users with search, sorting, and pagination
  - Get user by ID
  - Update user profile
  - Soft delete users
  - Role assignment

- ✅ **Role Management**
  - Create, read, update, delete roles
  - Permission management
  - Role-permission mapping
  - Soft delete with restore capability

- ✅ **Security**
  - JWT-based authentication
  - Bcrypt password hashing
  - Role-Based Access Control (RBAC)
  - Security headers middleware
  - Rate limiting
  - Audit trail (created_by, updated_by, deleted_by)
  - Soft delete support
  - Input sanitization

- ✅ **Caching**
  - Redis integration
  - Cache repository pattern
  - Configurable TTL
  - Cache invalidation strategies

- ✅ **Observability**
  - Prometheus metrics
  - OpenTelemetry distributed tracing
  - Structured logging with multiple levels
  - Request ID tracking
  - Performance monitoring

- ✅ **Middleware Stack**
  - Authentication & Authorization (JWT + RBAC)
  - Request logging with timing
  - Panic recovery
  - CORS support
  - Rate limiting
  - Request ID generation
  - Response compression
  - Security headers (CSP, HSTS, etc.)
  - Distributed tracing
  - Metrics collection

- ✅ **Infrastructure**
  - PostgreSQL with raw SQL (no ORM)
  - Redis for caching
  - Connection pooling
  - Circuit breaker pattern
  - Graceful shutdown
  - Health check endpoint
  - Docker & Docker Compose support
  - Comprehensive error handling

- ✅ **Documentation**
  - Swagger/OpenAPI documentation
  - Auto-generated API specs
  - Interactive API explorer

## Project Structure

```
.
├── cmd/
│   └── api/
│       └── main.go                 # Application entry point
├── internal/
│   ├── domain/                     # Domain Layer (Business Logic)
│   │   ├── entity/                 # Domain entities
│   │   ├── service/                # Business services
│   │   └── repository/             # Repository interfaces (ports)
│   ├── adapter/                    # Adapters (Infrastructure)
│   │   ├── http/                   # HTTP layer
│   │   │   ├── handler/            # HTTP handlers
│   │   │   ├── middleware/         # HTTP middleware
│   │   │   ├── request/            # Request DTOs
│   │   │   └── response/           # Response DTOs
│   │   ├── postgres/               # PostgreSQL implementations
│   │   ├── cache/                  # Redis cache implementations
│   │   └── crypto/                 # JWT and bcrypt implementations
│   ├── config/                     # Configuration management
│   └── pkg/                        # Shared packages
│       ├── validator/              # Input validation
│       ├── logger/                 # Structured logging
│       ├── errors/                 # Custom error types
│       ├── pagination/             # Pagination utilities
│       ├── sanitizer/              # Input sanitization
│       ├── circuitbreaker/         # Circuit breaker pattern
│       ├── contextutil/            # Context utilities
│       └── constants/              # Application constants
├── migrations/                     # Database migrations
├── docs/                           # API documentation (Swagger)
├── Dockerfile                      # Docker image definition
├── docker-compose.yml              # Docker Compose configuration
├── .env.example                    # Example environment variables
├── .golangci.yml                   # Linter configuration
├── Makefile                        # Common commands
└── README.md                       # This file
```

## Prerequisites

- Go 1.25 or higher
- PostgreSQL 14 or higher
- Redis 7 or higher (optional, for caching)
- Docker & Docker Compose (optional, for containerized deployment)
- Make (optional, for using Makefile commands)

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd golang-chi-hexagonal-base
```

2. Install dependencies:
```bash
go mod download
```

3. Copy the example environment file:
```bash
cp .env.example .env
```

4. Update the `.env` file with your configuration (see [Configuration](#configuration))

## Configuration

The application uses environment variables for configuration. Create a `.env` file in the project root:

```env
# Server Configuration
PORT=8080
ENV=development

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=myapp
DB_SSLMODE=disable
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=5m

# Redis Configuration (optional)
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
REDIS_POOL_SIZE=10
CACHE_TTL=3600

# JWT Configuration
JWT_SECRET=your-secret-key-minimum-32-characters-long
JWT_ACCESS_TOKEN_EXPIRY=15m
JWT_REFRESH_TOKEN_EXPIRY=168h

# Bcrypt Configuration
BCRYPT_COST=10

# Rate Limiting
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=1m

# Logging Configuration
LOG_LEVEL=info
LOG_FORMAT=json

# Observability (optional)
ENABLE_METRICS=true
ENABLE_TRACING=false
TRACING_ENDPOINT=http://localhost:4318
```

### Configuration Details

- **PORT**: HTTP server port (default: 8080)
- **ENV**: Environment mode (development/production)
- **DB_***: PostgreSQL connection settings
- **REDIS_***: Redis connection settings (optional, required for caching)
- **CACHE_TTL**: Default cache TTL in seconds (default: 3600)
- **JWT_SECRET**: Secret key for JWT signing (minimum 32 characters)
- **JWT_ACCESS_TOKEN_EXPIRY**: Access token lifetime (default: 15m)
- **JWT_REFRESH_TOKEN_EXPIRY**: Refresh token lifetime (default: 168h = 7 days)
- **BCRYPT_COST**: Bcrypt hashing cost (4-31, default: 10)
- **RATE_LIMIT_REQUESTS**: Maximum requests per time window (default: 100)
- **RATE_LIMIT_WINDOW**: Rate limit time window (default: 1m)
- **LOG_LEVEL**: Logging level (debug/info/warn/error)
- **LOG_FORMAT**: Log format (json/text)
- **ENABLE_METRICS**: Enable Prometheus metrics endpoint (default: true)
- **ENABLE_TRACING**: Enable OpenTelemetry tracing (default: false)
- **TRACING_ENDPOINT**: OpenTelemetry collector endpoint

## Database Setup

### Create Database

```bash
# Connect to PostgreSQL
psql -U postgres

# Create database
CREATE DATABASE myapp;

# Exit psql
\q
```

### Run Migrations

The project includes database migration files in the `migrations/` directory.

#### Using Make (Linux/macOS):

```bash
# Run all migrations
make migrate-up

# Rollback migrations
make migrate-down
```

#### Using psql directly (Cross-platform):

**Linux/macOS:**
```bash
# Run migrations manually
for file in migrations/*_*.up.sql; do
    psql -h localhost -p 5432 -U postgres -d myapp -f "$file"
done
```

**Windows (PowerShell):**
```powershell
# Run migrations manually
Get-ChildItem migrations\*_*.up.sql | ForEach-Object {
    psql -h localhost -p 5432 -U postgres -d myapp -f $_.FullName
}
```

**Windows (CMD):**
```cmd
# Run migrations manually (one at a time)
psql -h localhost -p 5432 -U postgres -d myapp -f migrations\001_create_users_table.up.sql
psql -h localhost -p 5432 -U postgres -d myapp -f migrations\002_create_roles_table.up.sql
# ... continue for each migration file
```

### Migration Files

The following migrations are included:

1. `001_create_users_table` - Users table with audit fields
2. `002_create_roles_table` - Roles table with audit fields
3. `003_create_permissions_table` - Permissions table
4. `004_create_user_roles_table` - User-role junction table
5. `005_create_role_permissions_table` - Role-permission junction table
6. `006_create_refresh_tokens_table` - Refresh tokens storage
7. `007_create_password_reset_tokens_table` - Password reset tokens
8. `008_seed_default_roles_permissions` - Default roles and permissions

## Running the Application

### Using Docker Compose (Recommended)

The easiest way to run the application with all dependencies:

```bash
# Start all services (PostgreSQL + Application)
make docker-up

# Or using docker-compose directly
docker-compose up -d

# View logs
docker-compose logs -f app

# Stop all services
make docker-down
```

This will:
- Start PostgreSQL database
- Run database migrations automatically
- Start the application
- Expose the API on port 8080

### Using Make (Linux/macOS)

```bash
# Run the application
make run

# Run with auto-reload (requires air)
make dev

# Build the application
make build

# Run the built binary
./bin/api
```

### Using Go directly (Cross-platform)

```bash
# Run directly
go run cmd/api/main.go

# Build and run
go build -o bin/api cmd/api/main.go

# On Windows
bin\api.exe

# On Linux/macOS
./bin/api
```

### Verify the Application

Once running, you can verify the application is working:

```bash
# Health check
curl http://localhost:8080/health

# Expected response:
# {"status":"ok"}

# Metrics endpoint (Prometheus)
curl http://localhost:8080/metrics

# Swagger documentation (if running)
# Open in browser: http://localhost:8080/swagger/index.html
```

## API Endpoints

### Health & Monitoring (Public)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check endpoint |
| GET | `/metrics` | Prometheus metrics (if enabled) |
| GET | `/swagger/*` | Swagger UI documentation |

### Authentication Endpoints (Public)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/register` | Register new user |
| POST | `/api/v1/auth/login` | Login with email/password |
| POST | `/api/v1/auth/refresh` | Refresh access token |
| POST | `/api/v1/auth/forgot-password` | Request password reset |
| POST | `/api/v1/auth/reset-password` | Reset password with token |
| POST | `/api/v1/auth/logout` | Logout (invalidate refresh token) |

### User Endpoints (Protected)

| Method | Endpoint | Description | Permission Required |
|--------|----------|-------------|---------------------|
| GET | `/api/v1/users` | List users (with pagination) | `users:read` |
| GET | `/api/v1/users/:id` | Get user by ID | `users:read` |
| PUT | `/api/v1/users/:id` | Update user | `users:write` |
| DELETE | `/api/v1/users/:id` | Delete user (soft delete) | `users:delete` |
| POST | `/api/v1/users/:id/roles` | Assign role to user | `users:write` |
| DELETE | `/api/v1/users/:id/roles/:role_id` | Remove role from user | `users:write` |

### Role Endpoints (Protected)

| Method | Endpoint | Description | Permission Required |
|--------|----------|-------------|---------------------|
| GET | `/api/v1/roles` | List roles | `roles:read` |
| POST | `/api/v1/roles` | Create role | `roles:write` |
| GET | `/api/v1/roles/:id` | Get role by ID | `roles:read` |
| PUT | `/api/v1/roles/:id` | Update role | `roles:write` |
| DELETE | `/api/v1/roles/:id` | Delete role (soft delete) | `roles:delete` |

### Query Parameters

#### List Users (`GET /api/v1/users`)

- `search` - Search by name or email
- `order_by` - Sort field (name, email, created_at)
- `order` - Sort direction (asc, desc)
- `page` - Page number (default: 1)
- `limit` - Items per page (default: 10, max: 100)
- `include_deleted` - Include soft-deleted users (admin only)

#### List Roles (`GET /api/v1/roles`)

- `search` - Search by name or description
- `order_by` - Sort field (name, created_at)
- `order` - Sort direction (asc, desc)
- `include_deleted` - Include soft-deleted roles (admin only)

### Example API Calls

#### Register
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!",
    "name": "John Doe"
  }'
```

#### Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!"
  }'
```

#### List Users (with authentication)
```bash
curl -X GET "http://localhost:8080/api/v1/users?page=1&limit=10&search=john" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

## Default Roles and Permissions

After running migrations, the following default roles and permissions are created:

### Roles

1. **admin** - Full system access
2. **user** - Basic user access

### Permissions

- `users:read` - View users
- `users:write` - Create/update users
- `users:delete` - Delete users
- `roles:read` - View roles
- `roles:write` - Create/update roles
- `roles:delete` - Delete roles
- `profile:read` - View own profile
- `profile:write` - Update own profile

### Role-Permission Mapping

- **admin**: All permissions
- **user**: `profile:read`, `profile:write`

## Development

### Project Commands

**Using Make (Linux/macOS):**
```bash
# Run tests
make test

# Run tests with coverage
make test-coverage

# Run linter
make lint

# Generate Swagger documentation
make swagger

# Run application
make run

# Run with auto-reload
make dev

# Build application
make build

# Run migrations
make migrate-up

# Rollback migrations
make migrate-down

# Create new migration
make migrate-create NAME=add_new_table

# Docker commands
make docker-build
make docker-up
make docker-down

# Clean build artifacts
make clean
```

**Using Go directly (Cross-platform):**
```bash
# Run tests
go test -v -cover ./...

# Run tests with coverage report
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run application
go run cmd/api/main.go

# Build application
go build -o bin/api cmd/api/main.go

# Generate Swagger docs (requires swag)
swag init -g cmd/api/main.go -o docs

# Clean build artifacts
go clean
```

### Code Organization

- **Domain Layer** (`internal/domain/`): Contains business logic, entities, and repository interfaces
- **Adapter Layer** (`internal/adapter/`): Contains infrastructure implementations (HTTP, database, crypto)
- **Configuration** (`internal/config/`): Environment configuration management
- **Shared Packages** (`internal/pkg/`): Reusable utilities (logger, validator)

### Adding New Features

1. Define domain entities in `internal/domain/entity/`
2. Create repository interfaces in `internal/domain/repository/`
3. Implement business logic in `internal/domain/service/`
4. Create repository implementations in `internal/adapter/postgres/`
5. Add HTTP handlers in `internal/adapter/http/handler/`
6. Register routes in `internal/adapter/http/router.go`
7. Add Swagger annotations to document your API endpoints

### Middleware Stack

The application includes a comprehensive middleware stack that can be configured per route:

- **RequestID**: Generates unique request IDs for tracing
- **Logger**: Logs HTTP requests with timing information
- **Recovery**: Recovers from panics and returns error responses
- **Compress**: Compresses HTTP responses (gzip)
- **Security**: Adds security headers (CSP, HSTS, X-Frame-Options, etc.)
- **RateLimit**: Limits request rates per client (token bucket algorithm)
- **Auth**: JWT authentication middleware
- **RBAC**: Role-based access control middleware
- **Metrics**: Collects Prometheus metrics (requests, latency, status codes)
- **Tracing**: OpenTelemetry distributed tracing integration

### Available Utilities

**Error Handling** (`internal/pkg/errors/`):
- Structured error types with codes
- Error wrapping and unwrapping
- HTTP status code mapping

**Pagination** (`internal/pkg/pagination/`):
- Standardized pagination utilities
- Configurable page size limits
- Metadata calculation

**Sanitizer** (`internal/pkg/sanitizer/`):
- HTML/XSS sanitization
- SQL injection prevention helpers

**Circuit Breaker** (`internal/pkg/circuitbreaker/`):
- Circuit breaker pattern implementation
- Configurable failure thresholds
- Automatic recovery

**Context Utilities** (`internal/pkg/contextutil/`):
- Extract user information from context
- Request ID management
- Timeout and cancellation helpers

**Constants** (`internal/pkg/constants/`):
- Application-wide constants
- HTTP header names
- Error codes and messages

## Testing

### Run All Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Run specific package tests
go test ./internal/domain/service/...
```

### Test Structure

- Unit tests are located alongside the code they test
- Integration tests require a test database
- Use table-driven tests for comprehensive coverage

## Graceful Shutdown

The application handles graceful shutdown on receiving SIGTERM or SIGINT signals:

1. Stops accepting new requests
2. Waits for in-flight requests to complete (30-second timeout)
3. Closes database connections
4. Exits cleanly

## Security Considerations

- **Password Security**: Passwords are hashed using bcrypt with configurable cost (default: 10)
- **JWT Authentication**: Tokens are signed with a secret key (minimum 32 characters)
- **Token Management**: 
  - Access tokens are short-lived (default: 15 minutes)
  - Refresh tokens are stored in database for revocation
  - Token rotation on refresh
- **SQL Injection Protection**: All SQL queries use prepared statements
- **Input Validation & Sanitization**: 
  - Request validation using struct tags
  - XSS protection through input sanitization
  - SQL injection prevention helpers
- **RBAC**: Role-based access control ensures users can only access authorized resources
- **Security Headers**: 
  - Content Security Policy (CSP)
  - HTTP Strict Transport Security (HSTS)
  - X-Frame-Options (clickjacking protection)
  - X-Content-Type-Options (MIME sniffing protection)
  - X-XSS-Protection
- **Rate Limiting**: Token bucket algorithm to prevent abuse
- **CORS**: Configurable CORS policies
- **Audit Trail**: Tracks who created, updated, or deleted records
- **Error Handling**: Does not expose sensitive information in error messages
- **Graceful Degradation**: Circuit breaker pattern prevents cascade failures

## Monitoring & Observability

### Prometheus Metrics

The application exposes Prometheus metrics at `/metrics` endpoint when `ENABLE_METRICS=true`:

**Available Metrics:**
- `http_requests_total` - Total HTTP requests by method, path, and status
- `http_request_duration_seconds` - HTTP request latency histogram
- `http_requests_in_flight` - Current number of in-flight requests
- Standard Go runtime metrics (memory, goroutines, GC, etc.)

**Scrape Configuration** (`prometheus.yml`):
```yaml
scrape_configs:
  - job_name: 'go-hex-api'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
```

### Distributed Tracing

OpenTelemetry tracing can be enabled with `ENABLE_TRACING=true`:

- Automatic trace context propagation
- Spans for HTTP requests
- Custom span creation for business logic
- Export to Jaeger, Zipkin, or any OTLP-compatible backend

**Example Jaeger Setup:**
```bash
# Run Jaeger all-in-one
docker run -d --name jaeger \
  -p 16686:16686 \
  -p 4318:4318 \
  jaegertracing/all-in-one:latest

# Set tracing endpoint
export TRACING_ENDPOINT=http://localhost:4318
export ENABLE_TRACING=true
```

### Structured Logging

The application uses structured JSON logging with configurable levels:

```bash
# Set log level (debug, info, warn, error)
export LOG_LEVEL=info

# Set log format (json, text)
export LOG_FORMAT=json
```

**Log Fields:**
- `timestamp` - ISO 8601 timestamp
- `level` - Log level
- `message` - Log message
- `request_id` - Unique request identifier
- `method`, `path`, `status` - HTTP request details
- `duration` - Request duration in milliseconds
- Additional context fields

## API Documentation

### Swagger/OpenAPI

The API is documented using Swagger/OpenAPI 2.0 specifications.

**Accessing Swagger UI:**
```
http://localhost:8080/swagger/index.html
```

**Generating Documentation:**
```bash
# Install swag
go install github.com/swaggo/swag/cmd/swag@latest

# Generate docs
make swagger
# or
swag init -g cmd/api/main.go -o docs
```

**Adding Swagger Annotations:**
```go
// @Summary      Get user by ID
// @Description  Retrieve a user by their ID
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "User ID"
// @Success      200  {object}  response.UserResponse
// @Failure      404  {object}  response.ErrorResponse
// @Security     BearerAuth
// @Router       /api/v1/users/{id} [get]
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
    // implementation
}
```

### API Documentation Files

- `docs/swagger.json` - OpenAPI JSON specification
- `docs/swagger.yaml` - OpenAPI YAML specification
- `docs/docs.go` - Generated Go documentation
- `docs/api.md` - Additional API documentation

## Performance & Scalability

### Optimization Features

- **Connection Pooling**: PostgreSQL and Redis connection pools
- **Response Compression**: Gzip compression for HTTP responses
- **Rate Limiting**: Prevents resource exhaustion
- **Circuit Breaker**: Prevents cascade failures
- **Caching**: Redis caching reduces database load
- **Prepared Statements**: Reusable SQL query plans

### Horizontal Scaling

The application is stateless and can be horizontally scaled:

1. Run multiple instances behind a load balancer
2. Share PostgreSQL and Redis across instances
3. Use sticky sessions if needed (not required for JWT)
4. Metrics and logs are per-instance (aggregate with external tools)

### Resource Requirements

**Minimum (Development):**
- 256 MB RAM
- 1 CPU core
- 1 GB disk space

**Recommended (Production):**
- 512 MB - 1 GB RAM per instance
- 2 CPU cores
- 10 GB disk space (for logs and migrations)
- PostgreSQL with 2+ GB RAM
- Redis with 256+ MB RAM

## Troubleshooting

### Common Issues

**Database Connection Failed:**
```bash
# Check PostgreSQL is running
docker ps | grep postgres
# or
pg_isready -h localhost -p 5432

# Check credentials in .env file
# Verify DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME
```

**Redis Connection Failed:**
```bash
# Check Redis is running
redis-cli ping
# Expected: PONG

# Check credentials in .env file
# Verify REDIS_HOST, REDIS_PORT, REDIS_PASSWORD
```

**Port Already in Use:**
```bash
# Change PORT in .env file
# or kill process using the port
lsof -ti:8080 | xargs kill -9  # macOS/Linux
netstat -ano | findstr :8080   # Windows
```

**Migrations Not Running:**
```bash
# Manually run migrations
make migrate-up

# Check migration status in database
psql -U postgres -d myapp -c "SELECT * FROM schema_migrations;"
```

## License

[Your License Here]

## Contributing

[Your Contributing Guidelines Here]
