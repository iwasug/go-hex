# Golang Chi Hexagonal Base Project

A production-ready Go application using Chi router with Hexagonal Architecture (Ports and Adapters pattern). This project provides authentication, user management, and role-based access control (RBAC) with PostgreSQL database.

## Table of Contents

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
  - Audit trail (created_by, updated_by, deleted_by)
  - Soft delete support

- ✅ **Infrastructure**
  - PostgreSQL with raw SQL (no ORM)
  - Connection pooling
  - Graceful shutdown
  - Health check endpoint
  - Structured logging
  - CORS support
  - Recovery middleware

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
│   │   ├── http/                   # HTTP handlers and middleware
│   │   ├── postgres/               # PostgreSQL implementations
│   │   └── crypto/                 # JWT and bcrypt implementations
│   ├── config/                     # Configuration management
│   └── pkg/                        # Shared packages
│       ├── validator/              # Input validation
│       └── logger/                 # Structured logging
├── migrations/                     # Database migrations
├── .env.example                    # Example environment variables
├── Makefile                        # Common commands
└── README.md                       # This file
```

## Prerequisites

- Go 1.21 or higher
- PostgreSQL 14 or higher
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

# JWT Configuration
JWT_SECRET=your-secret-key-minimum-32-characters-long
JWT_ACCESS_TOKEN_EXPIRY=15m
JWT_REFRESH_TOKEN_EXPIRY=168h

# Bcrypt Configuration
BCRYPT_COST=10

# Logging Configuration
LOG_LEVEL=info
LOG_FORMAT=json
```

### Configuration Details

- **PORT**: HTTP server port (default: 8080)
- **ENV**: Environment mode (development/production)
- **DB_***: PostgreSQL connection settings
- **JWT_SECRET**: Secret key for JWT signing (minimum 32 characters)
- **JWT_ACCESS_TOKEN_EXPIRY**: Access token lifetime (default: 15m)
- **JWT_REFRESH_TOKEN_EXPIRY**: Refresh token lifetime (default: 168h = 7 days)
- **BCRYPT_COST**: Bcrypt hashing cost (4-31, default: 10)
- **LOG_LEVEL**: Logging level (debug/info/warn/error)
- **LOG_FORMAT**: Log format (json/text)

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

### Using Make (Linux/macOS):

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

### Using Go directly (Cross-platform):

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
```

## API Endpoints

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

# Run application
make run

# Build application
make build

# Run migrations
make migrate-up

# Rollback migrations
make migrate-down

# Create new migration
make migrate-create NAME=add_new_table

# Clean build artifacts
make clean
```

**Using Go directly (Cross-platform):**
```bash
# Run tests
go test -v -cover ./...

# Run application
go run cmd/api/main.go

# Build application
go build -o bin/api cmd/api/main.go

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

- Passwords are hashed using bcrypt with configurable cost
- JWT tokens are signed with a secret key (minimum 32 characters)
- Access tokens are short-lived (15 minutes)
- Refresh tokens are stored in database for revocation
- All SQL queries use prepared statements to prevent SQL injection
- RBAC ensures users can only access authorized resources
- Audit trail tracks who created, updated, or deleted records

## License

[Your License Here]

## Contributing

[Your Contributing Guidelines Here]
