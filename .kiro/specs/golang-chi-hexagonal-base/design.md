# Design Document

## Overview

Proyek ini mengimplementasikan base project Golang dengan Go Chi router menggunakan Hexagonal Architecture (Ports and Adapters pattern). Sistem ini menyediakan authentication, user management, dan role-based access control (RBAC) dengan PostgreSQL sebagai database tanpa menggunakan ORM.

### Key Design Principles

1. **Hexagonal Architecture**: Memisahkan business logic (domain) dari infrastructure concerns
2. **Dependency Inversion**: Domain layer tidak bergantung pada infrastructure layer
3. **Raw SQL**: Menggunakan database/sql package tanpa ORM untuk kontrol penuh atas queries
4. **JWT-based Authentication**: Stateless authentication dengan access dan refresh tokens
5. **RBAC**: Role-Based Access Control dengan many-to-many relationship antara users, roles, dan permissions

## Architecture

### Hexagonal Architecture Layers

```
┌─────────────────────────────────────────────────────────────┐
│                     Infrastructure Layer                     │
│  (HTTP Handlers, Database, External Services)               │
│                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │   HTTP/Chi   │  │  PostgreSQL  │  │   JWT/Bcrypt │     │
│  │   Adapters   │  │   Adapters   │  │   Adapters   │     │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘     │
│         │                  │                  │              │
└─────────┼──────────────────┼──────────────────┼──────────────┘
          │                  │                  │
┌─────────┼──────────────────┼──────────────────┼──────────────┐
│         │    Application/Ports Layer          │              │
│         │                  │                  │              │
│  ┌──────▼───────┐  ┌──────▼───────┐  ┌──────▼───────┐     │
│  │   HTTP Port  │  │  Repository  │  │    Crypto    │     │
│  │  (Interface) │  │    Ports     │  │     Port     │     │
│  └──────────────┘  │ (Interfaces) │  │  (Interface) │     │
│                    └──────┬───────┘  └──────────────┘     │
│                           │                                 │
└───────────────────────────┼─────────────────────────────────┘
                            │
┌───────────────────────────┼─────────────────────────────────┐
│                    Domain Layer                              │
│                           │                                  │
│  ┌────────────┐  ┌────────▼────────┐  ┌────────────┐       │
│  │  Entities  │  │    Services      │  │   Value    │       │
│  │            │  │  (Business Logic)│  │  Objects   │       │
│  │ - User     │  │                  │  │            │       │
│  │ - Role     │  │ - AuthService    │  │ - Email    │       │
│  │ - Permission│ │ - UserService    │  │ - Password │       │
│  │            │  │ - RoleService    │  │            │       │
│  └────────────┘  └──────────────────┘  └────────────┘       │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

### Directory Structure

```
.
├── cmd/
│   └── api/
│       └── main.go                 # Application entry point
├── internal/
│   ├── domain/                     # Domain Layer (Business Logic)
│   │   ├── entity/
│   │   │   ├── user.go
│   │   │   ├── role.go
│   │   │   └── permission.go
│   │   ├── service/
│   │   │   ├── auth_service.go
│   │   │   ├── user_service.go
│   │   │   └── role_service.go
│   │   └── repository/             # Repository Interfaces (Ports)
│   │       ├── user_repository.go
│   │       ├── role_repository.go
│   │       ├── permission_repository.go
│   │       └── token_repository.go
│   ├── adapter/                    # Adapters (Infrastructure)
│   │   ├── http/
│   │   │   ├── handler/
│   │   │   │   ├── auth_handler.go
│   │   │   │   ├── user_handler.go
│   │   │   │   └── role_handler.go
│   │   │   ├── middleware/
│   │   │   │   ├── auth_middleware.go
│   │   │   │   ├── rbac_middleware.go
│   │   │   │   ├── logger_middleware.go
│   │   │   │   └── recovery_middleware.go
│   │   │   ├── request/
│   │   │   │   ├── auth_request.go
│   │   │   │   ├── user_request.go
│   │   │   │   └── role_request.go
│   │   │   ├── response/
│   │   │   │   ├── response.go
│   │   │   │   └── error_response.go
│   │   │   └── router.go
│   │   ├── postgres/
│   │   │   ├── user_repository.go
│   │   │   ├── role_repository.go
│   │   │   ├── permission_repository.go
│   │   │   ├── token_repository.go
│   │   │   └── db.go
│   │   └── crypto/
│   │       ├── jwt.go
│   ��       └── bcrypt.go
│   ├── config/
│   │   └── config.go               # Configuration management
│   └── pkg/
│       ├── validator/
│       │   └── validator.go
│       └── logger/
│           └── logger.go
├── migrations/
│   ├── 001_create_users_table.sql
│   ├── 002_create_roles_table.sql
│   ├── 003_create_permissions_table.sql
│   ├── 004_create_user_roles_table.sql
│   ├── 005_create_role_permissions_table.sql
│   ├── 006_create_refresh_tokens_table.sql
│   ├── 007_create_password_reset_tokens_table.sql
│   └── 008_seed_default_roles_permissions.sql
├── .env.example
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## Components and Interfaces

### Domain Layer

#### Entities

**User Entity**
```go
type User struct {
    ID        string
    Email     string
    Password  string // hashed
    Name      string
    IsDeleted bool
    CreatedBy *string
    UpdatedBy *string
    DeletedBy *string
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt *time.Time
}
```

**Role Entity**
```go
type Role struct {
    ID          string
    Name        string
    Description string
    IsDeleted   bool
    CreatedBy   *string
    UpdatedBy   *string
    DeletedBy   *string
    CreatedAt   time.Time
    UpdatedAt   time.Time
    DeletedAt   *time.Time
}
```

**Permission Entity**
```go
type Permission struct {
    ID          string
    Name        string
    Resource    string
    Action      string
    Description string
    CreatedAt   time.Time
}
```

#### Repository Interfaces (Ports)

**UserRepository**
```go
type ListUsersParams struct {
    Search        string // search by name or email
    OrderBy       string // field to order by (e.g., "name", "email", "created_at")
    Order         string // "asc" or "desc"
    Limit         int
    Offset        int
    IncludeDeleted bool  // include soft deleted records
}

type UserRepository interface {
    Create(ctx context.Context, user *User, createdBy string) error
    FindByID(ctx context.Context, id string) (*User, error)
    FindByEmail(ctx context.Context, email string) (*User, error)
    Update(ctx context.Context, user *User, updatedBy string) error
    Delete(ctx context.Context, id string, deletedBy string) error // soft delete
    HardDelete(ctx context.Context, id string) error // permanent delete
    Restore(ctx context.Context, id string) error // restore soft deleted user
    List(ctx context.Context, params ListUsersParams) ([]*User, int, error)
    GetUserRoles(ctx context.Context, userID string) ([]*Role, error)
    AssignRole(ctx context.Context, userID, roleID string) error
    RemoveRole(ctx context.Context, userID, roleID string) error
    GetUserPermissions(ctx context.Context, userID string) ([]*Permission, error)
}
```

**RoleRepository**
```go
type ListRolesParams struct {
    Search         string // search by name or description
    OrderBy        string // field to order by (e.g., "name", "created_at")
    Order          string // "asc" or "desc"
    IncludeDeleted bool   // include soft deleted records
}

type RoleRepository interface {
    Create(ctx context.Context, role *Role, createdBy string) error
    FindByID(ctx context.Context, id string) (*Role, error)
    FindByName(ctx context.Context, name string) (*Role, error)
    Update(ctx context.Context, role *Role, updatedBy string) error
    Delete(ctx context.Context, id string, deletedBy string) error // soft delete
    HardDelete(ctx context.Context, id string) error // permanent delete
    Restore(ctx context.Context, id string) error // restore soft deleted role
    List(ctx context.Context, params ListRolesParams) ([]*Role, error)
    GetRolePermissions(ctx context.Context, roleID string) ([]*Permission, error)
    AssignPermission(ctx context.Context, roleID, permissionID string) error
    RemovePermission(ctx context.Context, roleID, permissionID string) error
    IsRoleInUse(ctx context.Context, roleID string) (bool, error)
}
```

**PermissionRepository**
```go
type PermissionRepository interface {
    Create(ctx context.Context, permission *Permission) error
    FindByID(ctx context.Context, id string) (*Permission, error)
    FindByName(ctx context.Context, name string) (*Permission, error)
    List(ctx context.Context) ([]*Permission, error)
}
```

**TokenRepository**
```go
type TokenRepository interface {
    SaveRefreshToken(ctx context.Context, userID, token string, expiresAt time.Time) error
    FindRefreshToken(ctx context.Context, token string) (string, error) // returns userID
    DeleteRefreshToken(ctx context.Context, token string) error
    DeleteUserRefreshTokens(ctx context.Context, userID string) error
    SavePasswordResetToken(ctx context.Context, userID, token string, expiresAt time.Time) error
    FindPasswordResetToken(ctx context.Context, token string) (string, error) // returns userID
    DeletePasswordResetToken(ctx context.Context, token string) error
}
```

#### Services

**AuthService**
```go
type AuthService struct {
    userRepo  repository.UserRepository
    tokenRepo repository.TokenRepository
    jwtSecret string
    bcrypt    crypto.PasswordHasher
}

Methods:
- Register(ctx, email, password, name) (*User, error) // created_by will be nil for self-registration
- Login(ctx, email, password) (accessToken, refreshToken string, error)
- RefreshToken(ctx, refreshToken) (accessToken string, error)
- ForgotPassword(ctx, email) error
- ResetPassword(ctx, resetToken, newPassword) error
- Logout(ctx, refreshToken) error
```

**UserService**
```go
type UserService struct {
    userRepo repository.UserRepository
}

Methods:
- GetByID(ctx, id) (*User, error)
- List(ctx, search, orderBy, order string, page, limit int, includeDeleted bool) ([]*User, int, error)
- Update(ctx, id, updates, updatedBy string) (*User, error)
- Delete(ctx, id, deletedBy string) error // soft delete
- Restore(ctx, id) error // restore soft deleted user
- GetUserRoles(ctx, userID) ([]*Role, error)
- AssignRole(ctx, userID, roleID) error
- RemoveRole(ctx, userID, roleID) error
```

**RoleService**
```go
type RoleService struct {
    roleRepo repository.RoleRepository
}

Methods:
- Create(ctx, name, description, permissionIDs, createdBy string) (*Role, error)
- GetByID(ctx, id) (*Role, error)
- List(ctx, search, orderBy, order string, includeDeleted bool) ([]*Role, error)
- Update(ctx, id, updates, updatedBy string) (*Role, error)
- Delete(ctx, id, deletedBy string) error // soft delete
- Restore(ctx, id) error // restore soft deleted role
- GetRolePermissions(ctx, roleID) ([]*Permission, error)
- UpdatePermissions(ctx, roleID, permissionIDs) error
```

### Adapter Layer

#### HTTP Handlers

**AuthHandler**
```go
type AuthHandler struct {
    authService *service.AuthService
}

Endpoints:
- POST /api/v1/auth/register
- POST /api/v1/auth/login
- POST /api/v1/auth/refresh
- POST /api/v1/auth/forgot-password
- POST /api/v1/auth/reset-password
- POST /api/v1/auth/logout
```

**UserHandler**
```go
type UserHandler struct {
    userService *service.UserService
}

Endpoints:
- GET    /api/v1/users              # Query params: search, order_by, order, page, limit
- GET    /api/v1/users/:id
- PUT    /api/v1/users/:id
- DELETE /api/v1/users/:id
- POST   /api/v1/users/:id/roles
- DELETE /api/v1/users/:id/roles/:role_id

Query Parameters for GET /api/v1/users:
- search: string (optional) - search by name or email
- order_by: string (optional) - field to sort by (name, email, created_at), default: created_at
- order: string (optional) - sort direction (asc, desc), default: desc
- page: int (optional) - page number, default: 1
- limit: int (optional) - items per page, default: 10, max: 100
```

**RoleHandler**
```go
type RoleHandler struct {
    roleService *service.RoleService
}

Endpoints:
- GET    /api/v1/roles              # Query params: search, order_by, order
- POST   /api/v1/roles
- GET    /api/v1/roles/:id
- PUT    /api/v1/roles/:id
- DELETE /api/v1/roles/:id

Query Parameters for GET /api/v1/roles:
- search: string (optional) - search by name or description
- order_by: string (optional) - field to sort by (name, created_at), default: name
- order: string (optional) - sort direction (asc, desc), default: asc
```

#### Middleware

**AuthMiddleware**
- Validates JWT token dari Authorization header
- Extracts user information dan inject ke context
- Returns 401 jika token invalid/expired

**RBACMiddleware**
- Checks user permissions berdasarkan required permission
- Returns 403 jika user tidak memiliki permission
- Supports permission checking: `RequirePermission("users:read")`

**LoggerMiddleware**
- Logs setiap incoming request
- Logs response time dan status code

**RecoveryMiddleware**
- Recovers dari panic
- Returns 500 error response
- Logs panic stack trace

#### PostgreSQL Adapters

Implementasi repository interfaces menggunakan `database/sql` dengan raw SQL queries.

**Connection Management**
```go
type DB struct {
    *sql.DB
}

func NewDB(config *Config) (*DB, error)
func (db *DB) Ping() error
func (db *DB) Close() error
```

**Query Patterns**
- Prepared statements untuk security
- Context-aware queries untuk timeout/cancellation
- Transaction support untuk operations yang memerlukan atomicity
- Proper error handling dan connection pooling
- Dynamic query building untuk search dan sorting dengan whitelist validation
- SQL injection prevention dengan parameterized queries

#### Crypto Adapters

**JWT Implementation**
```go
type JWTManager struct {
    secret string
}

Methods:
- GenerateAccessToken(userID, email string, roles []string) (string, error)
- GenerateRefreshToken(userID string) (string, error)
- ValidateToken(token string) (*Claims, error)
- ParseToken(token string) (*Claims, error)
```

**Bcrypt Implementation**
```go
type BcryptHasher struct{}

Methods:
- HashPassword(password string) (string, error)
- ComparePassword(hashedPassword, password string) error
```

## Data Models

### Database Schema

**users table**
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    created_by UUID REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    deleted_by UUID REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_is_deleted ON users(is_deleted);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);
```

**roles table**
```sql
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    created_by UUID REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    deleted_by UUID REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_roles_name ON roles(name);
CREATE INDEX idx_roles_is_deleted ON roles(is_deleted);
```

**permissions table**
```sql
CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    resource VARCHAR(100) NOT NULL,
    action VARCHAR(50) NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_permissions_name ON permissions(name);
CREATE INDEX idx_permissions_resource_action ON permissions(resource, action);
```

**user_roles table**
```sql
CREATE TABLE user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id)
);

CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);
```

**role_permissions table**
```sql
CREATE TABLE role_permissions (
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (role_id, permission_id)
);

CREATE INDEX idx_role_permissions_role_id ON role_permissions(role_id);
```

**refresh_tokens table**
```sql
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(500) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_token ON refresh_tokens(token);
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
```

**password_reset_tokens table**
```sql
CREATE TABLE password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(500) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_password_reset_tokens_token ON password_reset_tokens(token);
CREATE INDEX idx_password_reset_tokens_expires_at ON password_reset_tokens(expires_at);
```

### Default Roles and Permissions

**Default Roles:**
1. `admin` - Full system access
2. `user` - Basic user access

**Default Permissions:**
- `users:read` - View users
- `users:write` - Create/update users
- `users:delete` - Delete users
- `roles:read` - View roles
- `roles:write` - Create/update roles
- `roles:delete` - Delete roles
- `profile:read` - View own profile
- `profile:write` - Update own profile

**Role-Permission Mapping:**
- `admin`: All permissions
- `user`: `profile:read`, `profile:write`

## Error Handling

### Error Response Format

```go
type ErrorResponse struct {
    Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
    Code    string                 `json:"code"`
    Message string                 `json:"message"`
    Details map[string]interface{} `json:"details,omitempty"`
}
```

### Error Codes

- `VALIDATION_ERROR` - Input validation failed
- `UNAUTHORIZED` - Authentication required or failed
- `FORBIDDEN` - Insufficient permissions
- `NOT_FOUND` - Resource not found
- `CONFLICT` - Resource conflict (e.g., duplicate email)
- `INTERNAL_ERROR` - Internal server error

### Error Handling Strategy

1. **Domain Layer**: Returns domain-specific errors
2. **Service Layer**: Wraps domain errors dengan context
3. **Handler Layer**: Converts errors ke HTTP responses dengan appropriate status codes
4. **Middleware**: Catches panics dan unhandled errors

### Logging

**Log Levels:**
- `ERROR`: Application errors, failed operations
- `WARN`: Recoverable issues, deprecated usage
- `INFO`: General information, request logs
- `DEBUG`: Detailed debugging information

**Log Format:**
```json
{
    "timestamp": "2025-10-08T10:30:00Z",
    "level": "INFO",
    "message": "Request completed",
    "method": "GET",
    "path": "/api/v1/users",
    "status": 200,
    "duration_ms": 45,
    "user_id": "uuid",
    "request_id": "uuid"
}
```

## Testing Strategy

### Unit Tests

**Domain Layer:**
- Test business logic dalam services
- Mock repository interfaces
- Test validation rules
- Test error conditions

**Adapter Layer:**
- Test HTTP handlers dengan mock services
- Test middleware behavior
- Test request/response serialization
- Test crypto implementations

### Integration Tests

**Database Layer:**
- Test repository implementations dengan test database
- Test transactions
- Test concurrent access
- Test query performance

**API Layer:**
- Test end-to-end API flows
- Test authentication/authorization
- Test error responses
- Test pagination

### Test Database

- Gunakan Docker container untuk PostgreSQL test database
- Run migrations sebelum tests
- Clean up data setelah setiap test
- Use transactions untuk test isolation

## Configuration

### Environment Variables

```
# Server
PORT=8080
ENV=development

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=myapp
DB_SSLMODE=disable
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=5m

# JWT
JWT_SECRET=your-secret-key
JWT_ACCESS_TOKEN_EXPIRY=15m
JWT_REFRESH_TOKEN_EXPIRY=168h

# Bcrypt
BCRYPT_COST=10

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
```

### Configuration Loading

- Load dari environment variables
- Support .env file untuk development
- Validate required configurations pada startup
- Provide sensible defaults

## Security Considerations

1. **Password Security**
   - Hash passwords dengan bcrypt (cost 10+)
   - Never return password hashes dalam API responses
   - Enforce password complexity rules

2. **JWT Security**
   - Use strong secret key (minimum 32 characters)
   - Short-lived access tokens (15 minutes)
   - Longer-lived refresh tokens (7 days)
   - Store refresh tokens di database untuk revocation

3. **SQL Injection Prevention**
   - Always use prepared statements
   - Never concatenate user input ke SQL queries
   - Validate dan sanitize all inputs

4. **CORS**
   - Configure allowed origins
   - Restrict methods dan headers
   - Handle preflight requests

5. **Rate Limiting**
   - Implement rate limiting untuk auth endpoints
   - Prevent brute force attacks
   - Use sliding window algorithm

6. **Input Validation**
   - Validate all user inputs
   - Use whitelist approach
   - Return clear validation errors

## Performance Considerations

1. **Database Connection Pooling**
   - Configure appropriate pool size
   - Monitor connection usage
   - Set connection lifetime limits

2. **Query Optimization**
   - Use indexes untuk frequently queried columns
   - Avoid N+1 queries
   - Use pagination untuk large result sets

3. **Caching Strategy**
   - Cache user permissions
   - Cache role-permission mappings
   - Use Redis untuk distributed caching (future enhancement)

4. **Middleware Ordering**
   - Recovery middleware first
   - Logger middleware early
   - Auth middleware before business logic
   - RBAC middleware after auth

## Deployment Considerations

1. **Database Migrations**
   - Run migrations sebelum deployment
   - Support rollback untuk failed migrations
   - Version control migration files

2. **Health Checks**
   - Implement `/health` endpoint
   - Check database connectivity
   - Check external dependencies

3. **Graceful Shutdown**
   - Handle SIGTERM/SIGINT signals
   - Close database connections
   - Complete in-flight requests
   - Set shutdown timeout

4. **Monitoring**
   - Log all errors dengan context
   - Track response times
   - Monitor database query performance
   - Alert on high error rates


## Pagination and Search Implementation

### Pagination Response Format

```go
type PaginatedResponse struct {
    Data       interface{} `json:"data"`
    Pagination Pagination  `json:"pagination"`
}

type Pagination struct {
    Total      int `json:"total"`
    Page       int `json:"page"`
    Limit      int `json:"limit"`
    TotalPages int `json:"total_pages"`
}
```

### Search Implementation

**Search Strategy:**
- Case-insensitive search menggunakan ILIKE (PostgreSQL)
- Search multiple fields dengan OR condition
- Trim whitespace dari search term
- Empty search term returns all results

**Example SQL Query for User List:**
```sql
SELECT 
    u.id, u.email, u.name, u.created_at, u.updated_at,
    COUNT(*) OVER() as total_count
FROM users u
WHERE 
    ($1 = '' OR u.name ILIKE '%' || $1 || '%' OR u.email ILIKE '%' || $1 || '%')
ORDER BY 
    CASE WHEN $2 = 'name' AND $3 = 'asc' THEN u.name END ASC,
    CASE WHEN $2 = 'name' AND $3 = 'desc' THEN u.name END DESC,
    CASE WHEN $2 = 'email' AND $3 = 'asc' THEN u.email END ASC,
    CASE WHEN $2 = 'email' AND $3 = 'desc' THEN u.email END DESC,
    CASE WHEN $2 = 'created_at' AND $3 = 'asc' THEN u.created_at END ASC,
    CASE WHEN $2 = 'created_at' AND $3 = 'desc' THEN u.created_at END DESC
LIMIT $4 OFFSET $5
```

### Sorting Implementation

**Allowed Sort Fields:**
- Users: `name`, `email`, `created_at`
- Roles: `name`, `created_at`

**Validation:**
- Whitelist validation untuk order_by field
- Only allow `asc` or `desc` for order direction
- Default values jika tidak provided
- Return validation error jika invalid values

**Security Considerations:**
- NEVER concatenate order_by field directly ke SQL query
- Use CASE statements atau map validation
- Prevent SQL injection melalui sort parameters


## Soft Delete and Audit Trail Implementation

### Soft Delete Strategy

**Soft Delete Behavior:**
- Set `is_deleted = true`
- Set `deleted_at = NOW()`
- Set `deleted_by = user_id` (user yang melakukan delete)
- Record tetap ada di database tapi tidak muncul di query default
- Dapat di-restore jika diperlukan

**Query Filtering:**
```sql
-- Default query (exclude soft deleted)
SELECT * FROM users WHERE is_deleted = false

-- Include soft deleted (untuk admin)
SELECT * FROM users WHERE ($1 = true OR is_deleted = false)
```

**Restore Functionality:**
- Set `is_deleted = false`
- Set `deleted_at = NULL`
- Set `deleted_by = NULL`
- Record kembali muncul di query default

**Hard Delete:**
- Permanent deletion dari database
- Hanya untuk admin dengan special permission
- Cascade delete ke related records

### Audit Trail Fields

**Audit Fields:**
- `created_by` - UUID user yang membuat record (NULL untuk self-registration)
- `updated_by` - UUID user yang terakhir update record
- `deleted_by` - UUID user yang soft delete record
- `created_at` - Timestamp pembuatan
- `updated_at` - Timestamp update terakhir
- `deleted_at` - Timestamp soft delete

**Audit Trail Usage:**
1. **Create Operations:**
   - Set `created_by` ke user ID yang membuat (NULL jika self-registration)
   - Set `created_at` ke NOW()
   - Set `updated_at` ke NOW()

2. **Update Operations:**
   - Set `updated_by` ke user ID yang update
   - Set `updated_at` ke NOW()

3. **Delete Operations:**
   - Set `deleted_by` ke user ID yang delete
   - Set `deleted_at` ke NOW()
   - Set `is_deleted` ke true

**Context Propagation:**
- Extract user ID dari JWT token di middleware
- Inject user ID ke request context
- Services extract user ID dari context untuk audit fields
- Repository methods accept audit parameters

**Example Context Usage:**
```go
// In middleware
ctx = context.WithValue(ctx, "user_id", userID)

// In service
userID := ctx.Value("user_id").(string)
err := s.userRepo.Update(ctx, user, userID)
```

### Soft Delete Considerations

**Login Prevention:**
- Check `is_deleted` flag saat login
- Return error jika user sudah di-soft delete
- Prevent token generation untuk deleted users

**Cascade Behavior:**
- Soft delete user → invalidate all refresh tokens
- Soft delete role → check if role in use first
- Maintain referential integrity

**Admin Features:**
- View soft deleted records dengan query parameter `include_deleted=true`
- Restore soft deleted records
- Hard delete untuk permanent removal (optional, dengan special permission)
