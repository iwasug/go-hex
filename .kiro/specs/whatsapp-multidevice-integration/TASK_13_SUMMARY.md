# Task 13: HTTP Handlers - Session Management

## Summary

Successfully implemented HTTP handlers for WhatsApp session management with complete CRUD operations, QR code generation, and proper integration with the existing authentication and RBAC system.

## Implementation Details

### 1. Request/Response Structures

**Created `internal/adapter/http/request/whatsapp_request.go`:**
- `CreateSessionRequest`: Request body for creating sessions
  - `session_name`: Required session name
  - `webhook_url`: Optional webhook URL
  - `webhook_secret`: Optional webhook secret
- `ListSessionsQueryParams`: Query parameters for listing sessions
- `GetQRCodeQueryParams`: Query parameters for QR code generation

**Created `internal/adapter/http/response/whatsapp_response.go`:**
- `WhatsAppSessionResponse`: Session details in API responses
  - Includes all session fields (ID, user_id, session_name, phone_number, status, etc.)
  - Excludes sensitive webhook_secret from response
- `QRCodeResponse`: QR code generation response
  - Includes session_id, qr_code (base64), expires_at, status

### 2. WhatsApp Handler

**Created `internal/adapter/http/handler/whatsapp_handler.go`:**

#### Handler Structure
```go
type WhatsAppHandler struct {
    whatsappService *service.WhatsAppService
}
```

#### Implemented Endpoints

1. **CreateSession** - `POST /api/v1/whatsapp/sessions`
   - Validates session_name (required, max 255 chars)
   - Validates webhook_url (optional, valid URL format)
   - Validates webhook_secret (optional, max 255 chars)
   - Calls `whatsappService.CreateSession()`
   - Returns 201 Created with session details
   - Error handling:
     - 409 Conflict: Session name already exists
     - 404 Not Found: User not found
     - 400 Bad Request: Validation errors

2. **ListSessions** - `GET /api/v1/whatsapp/sessions`
   - Query parameters:
     - `status`: Filter by status (connected, disconnected, etc.)
     - `is_active`: Filter by active status (true/false)
     - `search`: Search by session name or phone number
     - `order_by`: Sort field (session_name, status, created_at, last_activity)
     - `order`: Sort direction (asc, desc)
     - `page`: Page number (default: 1)
     - `limit`: Items per page (default: 10)
   - Validates order_by against whitelist
   - Validates order direction (asc/desc)
   - Supports admin view (all sessions) vs user view (own sessions only)
   - Returns paginated response with metadata

3. **GetSession** - `GET /api/v1/whatsapp/sessions/{id}`
   - Validates session ID from URL parameter
   - Checks session ownership (or admin access)
   - Returns session details
   - Error handling:
     - 404 Not Found: Session doesn't exist
     - 403 Forbidden: User doesn't own session

4. **DeleteSession** - `DELETE /api/v1/whatsapp/sessions/{id}`
   - Validates session ID from URL parameter
   - Checks session ownership (or admin access)
   - Calls `whatsappService.DeleteSession()` (includes client cleanup)
   - Returns 200 OK with success message
   - Error handling:
     - 404 Not Found: Session doesn't exist
     - 403 Forbidden: User doesn't own session

5. **GetQRCode** - `GET /api/v1/whatsapp/sessions/{id}/qr`
   - Validates session ID from URL parameter
   - Query parameter: `timeout` (seconds, default: 60, max: 300)
   - Checks session ownership (or admin access)
   - Validates timeout doesn't exceed 5 minutes
   - Calls `whatsappService.GenerateQRCode()`
   - Returns QR code (base64 PNG), expiry time, and session status
   - Error handling:
     - 404 Not Found: Session doesn't exist
     - 403 Forbidden: User doesn't own session
     - 409 Conflict: Session already connected

#### Helper Functions
- `toWhatsAppSessionResponse()`: Converts entity to response format

### 3. Validator Enhancement

**Updated `internal/pkg/validator/validator.go`:**
- Added `ValidateURL()` method for webhook URL validation
- Validates URL format (must start with http:// or https://)
- Uses regex pattern for URL validation

### 4. Router Integration

**Updated `internal/adapter/http/router.go`:**

#### Added Dependencies
- Imported `whatsapp` adapter package
- Imported `webhook` adapter package
- Imported `logger` package
- Imported `webhookpkg` for webhook client

#### Initialized WhatsApp Components
```go
// Repositories
whatsappSessionRepo := postgres.NewWhatsAppSessionRepository(rt.db)
whatsappMessageRepo := postgres.NewWhatsAppMessageRepository(rt.db)

// WhatsApp adapters
log := logger.NewFromConfig(rt.config.Logging.Level, rt.config.Logging.Format)
clientManager := whatsapp.NewClientManager(rt.db.DB, whatsappSessionRepo, whatsappMessageRepo, log)
webhookConfig := webhookpkg.DefaultConfig()
webhookClient := webhook.NewClientAdapter(webhookConfig, log)

// WhatsApp service
whatsappService := service.NewWhatsAppService(
    whatsappSessionRepo,
    whatsappMessageRepo,
    userRepo,
    clientManager,
    webhookClient,
)

// Handler
whatsappHandler := handler.NewWhatsAppHandler(whatsappService)
```

#### Registered Routes
```go
r.Route("/whatsapp", func(r chi.Router) {
    r.Route("/sessions", func(r chi.Router) {
        // Create session - requires whatsapp:write permission
        r.With(custommiddleware.RequirePermission(userRepo, "whatsapp:write")).Post("/", whatsappHandler.CreateSession)

        // List sessions - requires whatsapp:read permission
        r.With(custommiddleware.RequirePermission(userRepo, "whatsapp:read")).Get("/", whatsappHandler.ListSessions)

        // Get session by ID - requires whatsapp:read permission
        r.With(custommiddleware.RequirePermission(userRepo, "whatsapp:read")).Get("/{id}", whatsappHandler.GetSession)

        // Delete session - requires whatsapp:delete permission
        r.With(custommiddleware.RequirePermission(userRepo, "whatsapp:delete")).Delete("/{id}", whatsappHandler.DeleteSession)

        // Get QR code - requires whatsapp:read permission
        r.With(custommiddleware.RequirePermission(userRepo, "whatsapp:read")).Get("/{id}/qr", whatsappHandler.GetQRCode)
    })
})
```

## Security Features

### Authentication & Authorization
- All endpoints require JWT authentication (via `Auth` middleware)
- Permission-based access control:
  - `whatsapp:read`: View sessions and QR codes
  - `whatsapp:write`: Create sessions
  - `whatsapp:delete`: Delete sessions
- Session ownership validation (users can only access their own sessions)
- Admin bypass (admins can access all sessions)

### Input Validation
- Required field validation (session_name)
- Length validation (session_name, webhook_url, webhook_secret)
- URL format validation (webhook_url)
- Enum validation (order_by, order direction)
- Timeout range validation (QR code timeout)

### Error Handling
- Structured error responses with error codes
- Appropriate HTTP status codes:
  - 200 OK: Success
  - 201 Created: Session created
  - 400 Bad Request: Validation errors
  - 401 Unauthorized: Missing/invalid authentication
  - 403 Forbidden: Insufficient permissions or unauthorized access
  - 404 Not Found: Resource not found
  - 409 Conflict: Duplicate session name or already connected
  - 500 Internal Server Error: Server errors
- Detailed error messages for debugging

## API Documentation

All endpoints include Swagger annotations with:
- Summary and description
- Request/response schemas
- Parameter descriptions
- Success and error responses
- Security requirements (BearerAuth)
- Tags for organization

## Requirements Coverage

✅ **Requirement 1.1**: Session creation with unique identifier and metadata storage
✅ **Requirement 1.3**: List all active sessions with connection status and user info
✅ **Requirement 1.4**: Delete session with proper cleanup
✅ **Requirement 2.1**: Generate QR code for device pairing
✅ **Requirement 2.2**: Return QR code as base64-encoded PNG
✅ **Requirement 6.3**: Session ownership validation
✅ **Requirement 7.1**: JWT authentication required
✅ **Requirement 7.4**: Permission-based access control
✅ **Requirement 9.1**: Structured error responses with error codes

## Testing Recommendations

### Manual Testing
1. **Create Session**:
   ```bash
   curl -X POST http://localhost:8080/api/v1/whatsapp/sessions \
     -H "Authorization: Bearer <token>" \
     -H "Content-Type: application/json" \
     -d '{"session_name": "My WhatsApp"}'
   ```

2. **List Sessions**:
   ```bash
   curl -X GET "http://localhost:8080/api/v1/whatsapp/sessions?page=1&limit=10" \
     -H "Authorization: Bearer <token>"
   ```

3. **Get Session**:
   ```bash
   curl -X GET http://localhost:8080/api/v1/whatsapp/sessions/<session_id> \
     -H "Authorization: Bearer <token>"
   ```

4. **Get QR Code**:
   ```bash
   curl -X GET "http://localhost:8080/api/v1/whatsapp/sessions/<session_id>/qr?timeout=60" \
     -H "Authorization: Bearer <token>"
   ```

5. **Delete Session**:
   ```bash
   curl -X DELETE http://localhost:8080/api/v1/whatsapp/sessions/<session_id> \
     -H "Authorization: Bearer <token>"
   ```

### Integration Testing
- Test with valid JWT tokens
- Test with missing/invalid tokens (401)
- Test with insufficient permissions (403)
- Test session ownership validation
- Test admin access to all sessions
- Test pagination and filtering
- Test validation errors
- Test duplicate session name (409)
- Test QR code generation for connected session (409)

## Files Created/Modified

### Created
1. `internal/adapter/http/request/whatsapp_request.go` - Request structures
2. `internal/adapter/http/response/whatsapp_response.go` - Response structures
3. `internal/adapter/http/handler/whatsapp_handler.go` - HTTP handlers
4. `internal/adapter/whatsapp/client_manager_adapter.go` - Adapter for ClientManager to service interface
5. `internal/adapter/whatsapp/client_adapter.go` - Adapter for Client to service interface

### Modified
1. `internal/adapter/http/router.go` - Added WhatsApp routes and initialization
2. `internal/pkg/validator/validator.go` - Added ValidateURL method

## Adapter Pattern Implementation

To bridge the gap between the adapter layer and service layer interfaces, we implemented two adapter classes:

### ClientManagerAdapter
Wraps `whatsapp.ClientManager` to implement `service.WhatsAppClientManager` interface:
- Delegates all method calls to the underlying ClientManager
- Wraps returned Client instances with ClientAdapter

### ClientAdapter
Wraps `whatsapp.Client` to implement `service.WhatsAppClient` interface:
- Handles type conversion for SendMediaMessage (interface{} to MediaMessage)
- Delegates all other method calls to the underlying Client
- Ensures type safety between layers

## Next Steps

The next task (Task 14) will implement:
- Message sending handlers (text and media)
- Message history retrieval
- Multipart form handling for media uploads
- File size and MIME type validation
