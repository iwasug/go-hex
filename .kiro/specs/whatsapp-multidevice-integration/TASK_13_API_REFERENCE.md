# Task 13: WhatsApp Session Management API Reference

## Base URL
```
http://localhost:8080/api/v1/whatsapp
```

## Authentication
All endpoints require a valid JWT token in the Authorization header:
```
Authorization: Bearer <your_jwt_token>
```

## Permissions Required
- `whatsapp:read` - View sessions and QR codes
- `whatsapp:write` - Create sessions
- `whatsapp:delete` - Delete sessions

---

## Endpoints

### 1. Create Session

**POST** `/sessions`

Creates a new WhatsApp session for the authenticated user.

**Request Body:**
```json
{
  "session_name": "My WhatsApp Session",
  "webhook_url": "https://example.com/webhook",
  "webhook_secret": "my_secret_key"
}
```

**Response (201 Created):**
```json
{
  "success": true,
  "message": "Session created successfully",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "session_name": "My WhatsApp Session",
    "phone_number": "",
    "status": "disconnected",
    "qr_code": "",
    "qr_code_expiry": null,
    "device_info": "",
    "last_activity": "2024-10-09T14:30:00Z",
    "webhook_url": "https://example.com/webhook",
    "is_active": true,
    "created_at": "2024-10-09T14:30:00Z",
    "updated_at": "2024-10-09T14:30:00Z"
  }
}
```

**Errors:**
- `400` - Validation error (missing/invalid fields)
- `401` - Unauthorized (missing/invalid token)
- `403` - Forbidden (missing whatsapp:write permission)
- `409` - Conflict (session name already exists)

---

### 2. List Sessions

**GET** `/sessions`

Retrieves a paginated list of WhatsApp sessions.

**Query Parameters:**
- `status` (optional) - Filter by status: `connected`, `disconnected`, `connecting`, `authenticating`, `failed`
- `is_active` (optional) - Filter by active status: `true`, `false`
- `search` (optional) - Search by session name or phone number
- `order_by` (optional) - Sort field: `session_name`, `status`, `created_at`, `last_activity` (default: `created_at`)
- `order` (optional) - Sort direction: `asc`, `desc` (default: `desc`)
- `page` (optional) - Page number (default: `1`)
- `limit` (optional) - Items per page (default: `10`, max: `100`)

**Example Request:**
```bash
GET /sessions?status=connected&page=1&limit=10&order_by=created_at&order=desc
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Sessions retrieved successfully",
  "data": {
    "data": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "user_id": "123e4567-e89b-12d3-a456-426614174000",
        "session_name": "My WhatsApp Session",
        "phone_number": "+1234567890",
        "status": "connected",
        "last_activity": "2024-10-09T14:30:00Z",
        "webhook_url": "https://example.com/webhook",
        "is_active": true,
        "created_at": "2024-10-09T14:30:00Z",
        "updated_at": "2024-10-09T14:30:00Z"
      }
    ],
    "pagination": {
      "total": 1,
      "page": 1,
      "limit": 10,
      "total_pages": 1
    }
  }
}
```

**Errors:**
- `400` - Invalid query parameters
- `401` - Unauthorized
- `403` - Forbidden (missing whatsapp:read permission)

---

### 3. Get Session

**GET** `/sessions/{id}`

Retrieves detailed information about a specific WhatsApp session.

**Path Parameters:**
- `id` (required) - Session ID (UUID)

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Session retrieved successfully",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "session_name": "My WhatsApp Session",
    "phone_number": "+1234567890",
    "status": "connected",
    "qr_code": "",
    "qr_code_expiry": null,
    "device_info": "{\"platform\":\"go\",\"push_name\":\"My Device\"}",
    "last_activity": "2024-10-09T14:30:00Z",
    "webhook_url": "https://example.com/webhook",
    "is_active": true,
    "created_at": "2024-10-09T14:30:00Z",
    "updated_at": "2024-10-09T14:30:00Z"
  }
}
```

**Errors:**
- `400` - Invalid session ID
- `401` - Unauthorized
- `403` - Forbidden (user doesn't own session or missing permission)
- `404` - Session not found

---

### 4. Get QR Code

**GET** `/sessions/{id}/qr`

Generates a QR code for linking a WhatsApp account to the session.

**Path Parameters:**
- `id` (required) - Session ID (UUID)

**Query Parameters:**
- `timeout` (optional) - QR code timeout in seconds (default: `60`, max: `300`)

**Example Request:**
```bash
GET /sessions/550e8400-e29b-41d4-a716-446655440000/qr?timeout=60
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "QR code generated successfully",
  "data": {
    "session_id": "550e8400-e29b-41d4-a716-446655440000",
    "qr_code": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAA...",
    "expires_at": "2024-10-09T14:31:00Z",
    "status": "authenticating"
  }
}
```

**Errors:**
- `400` - Invalid parameters (timeout exceeds 5 minutes)
- `401` - Unauthorized
- `403` - Forbidden (user doesn't own session or missing permission)
- `404` - Session not found
- `409` - Conflict (session is already connected)

**Notes:**
- The QR code is returned as a base64-encoded PNG image
- The QR code expires after the specified timeout
- Scan the QR code with WhatsApp mobile app to link the account
- Once scanned, the session status will change to "connected"

---

### 5. Delete Session

**DELETE** `/sessions/{id}`

Deletes a WhatsApp session and disconnects the client.

**Path Parameters:**
- `id` (required) - Session ID (UUID)

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Session deleted successfully",
  "data": null
}
```

**Errors:**
- `400` - Invalid session ID
- `401` - Unauthorized
- `403` - Forbidden (user doesn't own session or missing whatsapp:delete permission)
- `404` - Session not found

**Notes:**
- Deleting a session will:
  - Disconnect the WhatsApp client
  - Remove all session data from the database
  - Clean up associated resources
  - Delete all messages associated with the session (cascade delete)

---

## Session Status Values

| Status | Description |
|--------|-------------|
| `disconnected` | Session is not connected to WhatsApp |
| `connecting` | Session is attempting to connect |
| `authenticating` | Session is waiting for QR code scan |
| `connected` | Session is connected and authenticated |
| `failed` | Session connection failed |

---

## Error Response Format

All error responses follow this structure:

```json
{
  "success": false,
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

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `VALIDATION_ERROR` | 400 | Request validation failed |
| `UNAUTHORIZED` | 401 | Missing or invalid authentication |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `NOT_FOUND` | 404 | Resource not found |
| `CONFLICT` | 409 | Resource conflict (duplicate, already exists) |
| `INTERNAL_ERROR` | 500 | Internal server error |

---

## cURL Examples

### Create Session
```bash
curl -X POST http://localhost:8080/api/v1/whatsapp/sessions \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "session_name": "My WhatsApp",
    "webhook_url": "https://example.com/webhook",
    "webhook_secret": "secret123"
  }'
```

### List Sessions
```bash
curl -X GET "http://localhost:8080/api/v1/whatsapp/sessions?page=1&limit=10" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Get Session
```bash
curl -X GET http://localhost:8080/api/v1/whatsapp/sessions/SESSION_ID \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Get QR Code
```bash
curl -X GET "http://localhost:8080/api/v1/whatsapp/sessions/SESSION_ID/qr?timeout=60" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Delete Session
```bash
curl -X DELETE http://localhost:8080/api/v1/whatsapp/sessions/SESSION_ID \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

---

## Admin Access

Users with admin privileges can:
- View all sessions across all users (not just their own)
- Access any session regardless of ownership
- Delete any session

Admin status is determined by the RBAC system and set in the request context by the authentication middleware.

---

## Rate Limiting

All endpoints are subject to the global rate limiting configured in the application:
- Default: 100 requests per minute per IP
- Burst: 200 requests

Rate limit headers are included in responses:
- `X-RateLimit-Limit` - Maximum requests allowed
- `X-RateLimit-Remaining` - Remaining requests in current window
- `X-RateLimit-Reset` - Time when the rate limit resets

---

## Webhook Configuration

When creating a session, you can optionally configure a webhook URL to receive incoming messages:

- `webhook_url` - HTTPS endpoint to receive webhook events
- `webhook_secret` - Secret key for HMAC signature verification

Webhook events will be sent with:
- `X-Webhook-Signature` header containing HMAC-SHA256 signature
- JSON payload with event data

See Task 12 documentation for webhook payload format and verification.
