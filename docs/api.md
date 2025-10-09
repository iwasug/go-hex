# API Documentation

## Base URL
```
http://localhost:8080/api/v1
```

## Authentication
Most endpoints require authentication using JWT tokens. Include the access token in the `Authorization` header:
```
Authorization: Bearer <access_token>
```

---

## Health & Monitoring

### Health Check
Get overall system health status.

**Endpoint:** `GET /health`

**Response:**
```json
{
  "status": "ok",
  "timestamp": "2024-01-01T00:00:00Z",
  "services": {
    "database": {
      "status": "up",
      "details": {
        "open_connections": "5",
        "in_use": "2",
        "idle": "3"
      }
    }
  }
}
```

### Readiness Check
Check if the service is ready to accept traffic.

**Endpoint:** `GET /health/ready`

**Response:**
```json
{
  "status": "ready"
}
```

### Liveness Check
Check if the service is alive.

**Endpoint:** `GET /health/live`

**Response:**
```json
{
  "status": "alive"
}
```

### Metrics
Prometheus metrics endpoint.

**Endpoint:** `GET /metrics`

---

## Authentication Endpoints

### Register
Create a new user account.

**Endpoint:** `POST /api/v1/auth/register`

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "name": "John Doe"
}
```

**Response:** `201 Created`
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "email": "user@example.com",
  "name": "John Doe",
  "is_deleted": false,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

**Errors:**
- `400 VALIDATION_ERROR` - Invalid input
- `409 CONFLICT` - Email already exists

---

### Login
Authenticate and receive tokens.

**Endpoint:** `POST /api/v1/auth/login`

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "SecurePass123!"
}
```

**Response:** `200 OK`
```json
{
  "access_token": "eyJhbGc...",
  "refresh_token": "eyJhbGc...",
  "user": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "email": "user@example.com",
    "name": "John Doe",
    "is_deleted": false,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

**Errors:**
- `400 VALIDATION_ERROR` - Invalid input
- `401 UNAUTHORIZED` - Invalid credentials

---

### Refresh Token
Get a new access token using refresh token.

**Endpoint:** `POST /api/v1/auth/refresh`

**Request Body:**
```json
{
  "refresh_token": "eyJhbGc..."
}
```

**Response:** `200 OK`
```json
{
  "access_token": "eyJhbGc..."
}
```

**Errors:**
- `401 UNAUTHORIZED` - Invalid or expired refresh token

---

### Logout
Invalidate the refresh token.

**Endpoint:** `POST /api/v1/auth/logout`

**Request Body:**
```json
{
  "refresh_token": "eyJhbGc..."
}
```

**Response:** `200 OK`
```json
{
  "message": "Logged out successfully"
}
```

---

### Forgot Password
Request a password reset token.

**Endpoint:** `POST /api/v1/auth/forgot-password`

**Request Body:**
```json
{
  "email": "user@example.com"
}
```

**Response:** `200 OK`
```json
{
  "message": "If the email exists, a password reset link has been sent"
}
```

---

### Reset Password
Reset password using reset token.

**Endpoint:** `POST /api/v1/auth/reset-password`

**Request Body:**
```json
{
  "token": "abc123...",
  "new_password": "NewSecurePass123!"
}
```

**Response:** `200 OK`
```json
{
  "message": "Password has been reset successfully"
}
```

**Errors:**
- `400 VALIDATION_ERROR` - Invalid or expired token

---

## User Endpoints

All user endpoints require authentication and appropriate permissions.

### List Users
Get paginated list of users.

**Endpoint:** `GET /api/v1/users`

**Required Permission:** `users:read`

**Query Parameters:**
- `search` (optional) - Search by name or email
- `order_by` (optional) - Sort field (name, email, created_at)
- `order` (optional) - Sort direction (asc, desc)
- `page` (optional) - Page number (default: 1)
- `limit` (optional) - Items per page (default: 10, max: 100)
- `include_deleted` (optional) - Include soft-deleted users (admin only)

**Response:** `200 OK`
```json
{
  "data": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "email": "user@example.com",
      "name": "John Doe",
      "is_deleted": false,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "metadata": {
    "current_page": 1,
    "page_size": 10,
    "total_records": 100,
    "total_pages": 10,
    "has_next": true,
    "has_previous": false
  }
}
```

---

### Get User by ID
Get a specific user by ID.

**Endpoint:** `GET /api/v1/users/:id`

**Required Permission:** `users:read`

**Response:** `200 OK`
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "email": "user@example.com",
  "name": "John Doe",
  "is_deleted": false,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

**Errors:**
- `404 NOT_FOUND` - User not found

---

### Update User
Update user information.

**Endpoint:** `PUT /api/v1/users/:id`

**Required Permission:** `users:write`

**Request Body:**
```json
{
  "name": "Jane Doe",
  "email": "jane@example.com"
}
```

**Response:** `200 OK`
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "email": "jane@example.com",
  "name": "Jane Doe",
  "is_deleted": false,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-02T00:00:00Z"
}
```

---

### Delete User
Soft delete a user.

**Endpoint:** `DELETE /api/v1/users/:id`

**Required Permission:** `users:delete`

**Response:** `204 No Content`

---

### Assign Role to User
Assign a role to a user.

**Endpoint:** `POST /api/v1/users/:id/roles`

**Required Permission:** `users:write`

**Request Body:**
```json
{
  "role_id": "123e4567-e89b-12d3-a456-426614174001"
}
```

**Response:** `200 OK`
```json
{
  "message": "Role assigned successfully"
}
```

---

### Remove Role from User
Remove a role from a user.

**Endpoint:** `DELETE /api/v1/users/:id/roles/:role_id`

**Required Permission:** `users:write`

**Response:** `204 No Content`

---

## Role Endpoints

All role endpoints require authentication and appropriate permissions.

### List Roles
Get list of all roles.

**Endpoint:** `GET /api/v1/roles`

**Required Permission:** `roles:read`

**Query Parameters:**
- `search` (optional) - Search by name or description
- `order_by` (optional) - Sort field (name, created_at)
- `order` (optional) - Sort direction (asc, desc)
- `include_deleted` (optional) - Include soft-deleted roles (admin only)

**Response:** `200 OK`
```json
[
  {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "admin",
    "description": "Administrator role",
    "is_deleted": false,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
]
```

---

### Create Role
Create a new role.

**Endpoint:** `POST /api/v1/roles`

**Required Permission:** `roles:write`

**Request Body:**
```json
{
  "name": "moderator",
  "description": "Moderator role with limited admin privileges"
}
```

**Response:** `201 Created`
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174002",
  "name": "moderator",
  "description": "Moderator role with limited admin privileges",
  "is_deleted": false,
  "created_at": "2024-01-02T00:00:00Z",
  "updated_at": "2024-01-02T00:00:00Z"
}
```

---

### Get Role by ID
Get a specific role by ID.

**Endpoint:** `GET /api/v1/roles/:id`

**Required Permission:** `roles:read`

**Response:** `200 OK`
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "name": "admin",
  "description": "Administrator role",
  "is_deleted": false,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

---

### Update Role
Update role information.

**Endpoint:** `PUT /api/v1/roles/:id`

**Required Permission:** `roles:write`

**Request Body:**
```json
{
  "name": "super_admin",
  "description": "Super administrator with full access"
}
```

**Response:** `200 OK`
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "name": "super_admin",
  "description": "Super administrator with full access",
  "is_deleted": false,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-02T00:00:00Z"
}
```

---

### Delete Role
Soft delete a role.

**Endpoint:** `DELETE /api/v1/roles/:id`

**Required Permission:** `roles:delete`

**Response:** `204 No Content`

---

## Error Responses

All error responses follow this format:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message",
    "details": {
      "field": "Additional error details"
    }
  }
}
```

### Common Error Codes

- `VALIDATION_ERROR` (400) - Request validation failed
- `UNAUTHORIZED` (401) - Authentication required or invalid token
- `FORBIDDEN` (403) - Insufficient permissions
- `NOT_FOUND` (404) - Resource not found
- `CONFLICT` (409) - Resource conflict (e.g., duplicate email)
- `RATE_LIMIT_EXCEEDED` (429) - Too many requests
- `INTERNAL_ERROR` (500) - Internal server error

---

## Rate Limiting

The API implements rate limiting to prevent abuse:
- **Default:** 100 requests per second per IP
- **Burst:** 200 requests

When rate limit is exceeded, the API returns:
```json
{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Too many requests. Please try again later."
  }
}
```

---

## Request Tracing

All requests include a `X-Request-ID` header for tracing:
- If provided in the request, it will be echoed back
- If not provided, a new UUID will be generated
- Use this ID for debugging and support inquiries
