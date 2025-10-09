# Implementation Plan

- [x] 1. Setup project structure dan dependencies





  - Initialize Go module dengan `go mod init`
  - Install dependencies: go-chi, pq (PostgreSQL driver), golang-jwt, bcrypt, godotenv
  - Create directory structure sesuai hexagonal architecture (internal/domain, internal/adapter, cmd/api)
  - Create .env.example file dengan semua environment variables yang diperlukan
  - Create Makefile dengan common commands (run, migrate, test)
  - _Requirements: 1.1, 1.2, 1.3, 1.4_

- [x] 2. Implement configuration management





  - Create config package untuk load environment variables
  - Implement config struct dengan validation
  - Support .env file loading untuk development
  - Provide default values untuk optional configs
  - _Requirements: 1.3_

- [x] 3. Setup database connection dan migrations




  - [x] 3.1 Implement database connection dengan connection pooling


    - Create DB struct wrapper untuk sql.DB
    - Implement connection dengan proper timeout dan retry logic
    - Configure connection pool settings (max open, max idle, conn lifetime)
    - Implement Ping() untuk health check
    - _Requirements: 1.2_
  

  - [x] 3.2 Create migration files untuk database schema

    - Create migration 001: users table dengan audit fields dan soft delete
    - Create migration 002: roles table dengan audit fields dan soft delete
    - Create migration 003: permissions table
    - Create migration 004: user_roles junction table
    - Create migration 005: role_permissions junction table
    - Create migration 006: refresh_tokens table
    - Create migration 007: password_reset_tokens table
    - Create migration 008: seed default roles dan permissions
    - Add proper indexes untuk performa
    - _Requirements: 17.1, 17.2, 17.3_

- [x] 4. Implement domain entities





  - Create User entity dengan semua fields termasuk audit dan soft delete
  - Create Role entity dengan semua fields termasuk audit dan soft delete
  - Create Permission entity
  - Create value objects jika diperlukan (Email, Password)
  - _Requirements: 1.1_

- [x] 5. Define repository interfaces (ports)





  - Create UserRepository interface dengan semua methods termasuk soft delete dan restore
  - Create RoleRepository interface dengan semua methods termasuk soft delete dan restore
  - Create PermissionRepository interface
  - Create TokenRepository interface untuk refresh dan reset tokens
  - Define ListUsersParams dan ListRolesParams structs untuk query parameters
  - _Requirements: 1.1_

- [x] 6. Implement crypto adapters






  - [x] 6.1 Implement bcrypt password hasher

    - Create BcryptHasher struct
    - Implement HashPassword method dengan configurable cost
    - Implement ComparePassword method
    - _Requirements: 2.3_
  


  - [x] 6.2 Implement JWT token manager





    - Create JWTManager struct dengan secret key
    - Implement GenerateAccessToken dengan 15 menit expiry
    - Implement GenerateRefreshToken dengan 7 hari expiry
    - Implement ValidateToken untuk verify token signature dan expiry
    - Implement ParseToken untuk extract claims
    - Include user_id, email, roles dalam token payload
    - _Requirements: 3.3, 3.4, 3.6_

- [x] 7. Implement PostgreSQL repository adapters








  - [x] 7.1 Implement UserRepository dengan raw SQL


    - Implement Create dengan audit fields (created_by bisa NULL)
    - Implement FindByID dengan filter is_deleted
    - Implement FindByEmail dengan filter is_deleted
    - Implement Update dengan updated_by dan updated_at
    - Implement Delete (soft delete) dengan deleted_by dan deleted_at
    - Implement Restore untuk set is_deleted=false
    - Implement List dengan search, sorting, pagination, dan include_deleted option
    - Implement GetUserRoles, AssignRole, RemoveRole
    - Implement GetUserPermissions dengan join ke roles dan permissions
    - Use prepared statements untuk semua queries
    - _Requirements: 2.1, 6.1, 6.2, 6.3, 6.4, 6.5, 6.10, 7.1, 8.1, 8.6, 9.1, 9.5, 9.6_
  


  - [x] 7.2 Implement RoleRepository dengan raw SQL

    - Implement Create dengan audit fields
    - Implement FindByID dengan filter is_deleted
    - Implement FindByName dengan filter is_deleted
    - Implement Update dengan updated_by dan updated_at
    - Implement Delete (soft delete) dengan deleted_by dan deleted_at
    - Implement Restore untuk set is_deleted=false
    - Implement List dengan search, sorting, dan include_deleted option
    - Implement GetRolePermissions, AssignPermission, RemovePermission
    - Implement IsRoleInUse untuk check sebelum delete
    - Use prepared statements untuk semua queries

    - _Requirements: 10.1, 10.2, 10.3, 10.4, 10.5, 11.1, 11.3, 12.1, 12.3, 13.1, 13.2, 13.5_
  
  - [x] 7.3 Implement PermissionRepository dengan raw SQL

    - Implement Create
    - Implement FindByID
    - Implement FindByName
    - Implement List
    - Use prepared statements untuk semua queries
    - _Requirements: 17.2_
  


  - [x] 7.4 Implement TokenRepository dengan raw SQL





    - Implement SaveRefreshToken dengan expiry
    - Implement FindRefreshToken untuk return user_id
    - Implement DeleteRefreshToken
    - Implement DeleteUserRefreshTokens untuk cascade delete
    - Implement SavePasswordResetToken dengan expiry
    - Implement FindPasswordResetToken untuk return user_id
    - Implement DeletePasswordResetToken
    - Use prepared statements untuk semua queries
    - _Requirements: 3.5, 4.3, 5.1, 5.2, 5.5_

- [x] 8. Implement domain services






  - [x] 8.1 Implement AuthService

    - Implement Register: validate input, hash password, create user dengan default role, created_by=NULL
    - Implement Login: validate credentials, check is_deleted, generate tokens, save refresh token
    - Implement RefreshToken: validate refresh token, check expiry, generate new access token
    - Implement ForgotPassword: generate reset token, save dengan expiry 1 jam
    - Implement ResetPassword: validate reset token, update password, delete reset token
    - Implement Logout: delete refresh token
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 3.1, 3.2, 3.3, 3.4, 3.5, 4.1, 4.2, 4.4, 5.1, 5.3, 5.4, 5.5_
  

  - [x] 8.2 Implement UserService

    - Implement GetByID
    - Implement List dengan search, sorting, pagination, include_deleted
    - Implement Update dengan updated_by dari context
    - Implement Delete (soft delete) dengan deleted_by dari context
    - Implement Restore
    - Implement GetUserRoles, AssignRole, RemoveRole
    - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5, 6.6, 6.7, 6.8, 6.9, 6.10, 7.1, 8.1, 8.2, 8.3, 8.4, 8.5, 8.6, 9.1, 9.2, 9.3, 9.4, 9.5, 9.6, 14.1, 14.2, 14.3, 14.4, 15.1, 15.2, 15.3_
  

  - [x] 8.3 Implement RoleService





    - Implement Create dengan created_by dari context
    - Implement GetByID
    - Implement List dengan search, sorting, include_deleted
    - Implement Update dengan updated_by dari context
    - Implement Delete (soft delete) dengan deleted_by dari context, check IsRoleInUse
    - Implement Restore
    - Implement GetRolePermissions, UpdatePermissions
    - _Requirements: 10.1, 10.2, 10.3, 10.4, 10.5, 10.6, 10.7, 11.1, 11.2, 11.3, 11.4, 12.1, 12.2, 12.3, 12.4, 13.1, 13.2, 13.3, 13.4, 13.5_

- [x] 9. Implement HTTP request/response models





  - Create request structs untuk auth (RegisterRequest, LoginRequest, dll)
  - Create request structs untuk user (UpdateUserRequest, dll)
  - Create request structs untuk role (CreateRoleRequest, UpdateRoleRequest, dll)
  - Create response structs (UserResponse, RoleResponse, dll) tanpa sensitive data
  - Create PaginatedResponse struct dengan pagination metadata
  - Create ErrorResponse struct dengan consistent format
  - _Requirements: 18.1_

- [x] 10. Implement validation helper





  - Create validator package untuk input validation
  - Implement email validation
  - Implement password strength validation
  - Implement required field validation
  - Implement enum validation (order_by, order direction)
  - Return structured validation errors
  - _Requirements: 2.5, 18.1_

- [x] 11. Implement middleware






  - [x] 11.1 Implement RecoveryMiddleware

    - Catch panics dan recover
    - Log panic dengan stack trace
    - Return 500 error response
    - _Requirements: 18.4_
  

  - [x] 11.2 Implement LoggerMiddleware

    - Log setiap incoming request (method, path, user_id)
    - Log response status dan duration
    - Generate request_id untuk tracing
    - _Requirements: 18.3_
  

  - [x] 11.3 Implement AuthMiddleware

    - Extract JWT token dari Authorization header
    - Validate token signature dan expiry
    - Extract user information dari token
    - Inject user_id ke request context untuk audit trail
    - Return 401 jika token invalid/missing
    - _Requirements: 16.1, 16.2, 16.3_

  
  - [x] 11.4 Implement RBACMiddleware

    - Extract user permissions dari context atau database
    - Check required permission untuk endpoint
    - Return 403 jika user tidak memiliki permission
    - Support permission format: "resource:action"
    - _Requirements: 16.4, 16.5_

- [x] 12. Implement HTTP handlers






  - [x] 12.1 Implement AuthHandler

    - POST /api/v1/auth/register: validate input, call AuthService.Register
    - POST /api/v1/auth/login: validate input, call AuthService.Login
    - POST /api/v1/auth/refresh: validate input, call AuthService.RefreshToken
    - POST /api/v1/auth/forgot-password: validate input, call AuthService.ForgotPassword
    - POST /api/v1/auth/reset-password: validate input, call AuthService.ResetPassword
    - POST /api/v1/auth/logout: validate input, call AuthService.Logout
    - Return appropriate HTTP status codes dan error responses
    - _Requirements: 2.1, 2.2, 2.4, 2.5, 3.1, 3.2, 4.1, 4.2, 5.1, 5.3, 5.4_
  

  - [x] 12.2 Implement UserHandler

    - GET /api/v1/users: parse query params (search, order_by, order, page, limit, include_deleted), call UserService.List, return paginated response
    - GET /api/v1/users/:id: validate id, call UserService.GetByID
    - PUT /api/v1/users/:id: validate input dan permissions, call UserService.Update dengan updated_by dari context
    - DELETE /api/v1/users/:id: validate permissions, call UserService.Delete dengan deleted_by dari context
    - POST /api/v1/users/:id/roles: validate input, call UserService.AssignRole
    - DELETE /api/v1/users/:id/roles/:role_id: validate input, call UserService.RemoveRole
    - Validate order_by field dengan whitelist
    - Return appropriate HTTP status codes dan error responses
    - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5, 6.6, 6.7, 6.8, 6.9, 6.10, 7.1, 7.2, 7.3, 8.1, 8.2, 8.3, 8.4, 8.5, 9.1, 9.2, 9.3, 14.1, 14.2, 14.3, 14.4, 15.1, 15.2, 15.3_
  

  - [x] 12.3 Implement RoleHandler

    - GET /api/v1/roles: parse query params (search, order_by, order, include_deleted), call RoleService.List
    - POST /api/v1/roles: validate input, call RoleService.Create dengan created_by dari context
    - GET /api/v1/roles/:id: validate id, call RoleService.GetByID
    - PUT /api/v1/roles/:id: validate input, call RoleService.Update dengan updated_by dari context
    - DELETE /api/v1/roles/:id: validate permissions, call RoleService.Delete dengan deleted_by dari context
    - Validate order_by field dengan whitelist
    - Return appropriate HTTP status codes dan error responses
    - _Requirements: 10.1, 10.2, 10.3, 10.4, 10.5, 10.6, 10.7, 11.1, 11.2, 11.3, 11.4, 12.1, 12.2, 12.3, 12.4, 13.1, 13.2, 13.3, 13.4_

- [x] 13. Setup router dan wire dependencies





  - Create router dengan go-chi
  - Setup middleware chain (Recovery → Logger → CORS)
  - Define public routes (auth endpoints)
  - Define protected routes dengan AuthMiddleware
  - Define admin routes dengan AuthMiddleware + RBACMiddleware
  - Wire all dependencies (repositories → services → handlers)
  - Setup CORS configuration
  - _Requirements: 1.4, 16.1, 16.2, 16.3, 16.4, 16.5_

- [x] 14. Implement logger package





  - Create structured logger dengan levels (ERROR, WARN, INFO, DEBUG)
  - Support JSON format untuk production
  - Include timestamp, level, message, dan context fields
  - _Requirements: 18.2, 18.3_

- [x] 15. Implement main application entry point









  - Load configuration dari environment
  - Initialize database connection
  - Run migrations (atau provide migration command)
  - Initialize all dependencies
  - Setup HTTP server dengan graceful shutdown
  - Handle SIGTERM/SIGINT untuk graceful shutdown
  - Implement health check endpoint
  - _Requirements: 1.1, 1.2, 1.3, 1.4_


- [x] 16. Create README documentation




  - Document project structure dan architecture
  - Document API endpoints dengan examples
  - Document environment variables
  - Document how to run migrations
  - Document how to run the application
  - Document default roles dan permissions
  - _Requirements: 1.1_
