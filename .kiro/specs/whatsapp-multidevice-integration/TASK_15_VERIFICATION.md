# Task 15 Verification: HTTP Handlers - Session Operations

## Overview
This document provides verification steps for Task 15: HTTP handlers for session operations (reconnect and status endpoints).

## Implementation Summary

### 1. Service Layer
- ✅ Added `ReconnectSession` method to `WhatsAppService`
  - Validates session ownership
  - Checks if session is already connected
  - Creates client if needed
  - Attempts connection
  - Updates session status appropriately

### 2. Response Types
- ✅ Added `SessionStatusResponse` struct in `internal/adapter/http/response/whatsapp_response.go`
  - Contains session ID, status, phone numbers, activity, and connection flags

### 3. HTTP Handlers
- ✅ Added `ReconnectSession` handler
  - POST `/api/v1/whatsapp/sessions/:id/reconnect`
  - Validates session ID and user authentication
  - Calls service to reconnect
  - Returns appropriate HTTP status codes
  - Handles all error cases

- ✅ Added `GetSessionStatus` handler
  - GET `/api/v1/whatsapp/sessions/:id/status`
  - Validates session ID and user authentication
  - Retrieves real-time session status
  - Returns structured status information

### 4. Router Configuration
- ✅ Registered reconnect endpoint with `whatsapp:write` permission
- ✅ Registered status endpoint with `whatsapp:read` permission

## API Endpoints

### 1. Reconnect Session
```http
POST /api/v1/whatsapp/sessions/{id}/reconnect
Authorization: Bearer <token>
```

**Success Response (200 OK):**
```json
{
  "success": true,
  "message": "Session reconnection initiated successfully"
}
```

**Error Responses:**
- `400 Bad Request` - Session already connected
- `401 Unauthorized` - Missing or invalid token
- `403 Forbidden` - Insufficient permissions or not session owner
- `404 Not Found` - Session not found
- `500 Internal Server Error` - Connection failed

### 2. Get Session Status
```http
GET /api/v1/whatsapp/sessions/{id}/status
Authorization: Bearer <token>
```

**Success Response (200 OK):**
```json
{
  "success": true,
  "message": "Session status retrieved successfully",
  "data": {
    "session_id": "uuid",
    "status": "connected",
    "phone_number": "+1234567890",
    "connected_phone": "+1234567890",
    "last_activity": "2024-01-15T10:30:00Z",
    "is_active": true,
    "is_connected": true,
    "is_logged_in": true
  }
}
```

**Error Responses:**
- `400 Bad Request` - Invalid session ID
- `401 Unauthorized` - Missing or invalid token
- `403 Forbidden` - Insufficient permissions or not session owner
- `404 Not Found` - Session not found
- `500 Internal Server Error` - Failed to retrieve status

## Testing Steps

### Prerequisites
1. Start the application
2. Create a user account and obtain JWT token
3. Ensure user has `whatsapp:read` and `whatsapp:write` permissions
4. Create a WhatsApp session

### Test 1: Get Session Status
```bash
# Get status of a session
curl -X GET "http://localhost:8080/api/v1/whatsapp/sessions/{session_id}/status" \
  -H "Authorization: Bearer {token}"

# Expected: 200 OK with session status details
```

### Test 2: Reconnect Disconnected Session
```bash
# First, ensure session is disconnected (or create a new one)
# Then attempt reconnection
curl -X POST "http://localhost:8080/api/v1/whatsapp/sessions/{session_id}/reconnect" \
  -H "Authorization: Bearer {token}"

# Expected: 200 OK with success message
```

### Test 3: Reconnect Already Connected Session
```bash
# Try to reconnect a session that's already connected
curl -X POST "http://localhost:8080/api/v1/whatsapp/sessions/{session_id}/reconnect" \
  -H "Authorization: Bearer {token}"

# Expected: 400 Bad Request - "Session is already connected"
```

### Test 4: Unauthorized Access
```bash
# Try to access another user's session
curl -X GET "http://localhost:8080/api/v1/whatsapp/sessions/{other_user_session_id}/status" \
  -H "Authorization: Bearer {token}"

# Expected: 403 Forbidden
```

### Test 5: Missing Permissions
```bash
# Use a token without whatsapp:read permission
curl -X GET "http://localhost:8080/api/v1/whatsapp/sessions/{session_id}/status" \
  -H "Authorization: Bearer {token_without_permission}"

# Expected: 403 Forbidden
```

### Test 6: Invalid Session ID
```bash
# Try with non-existent session ID
curl -X GET "http://localhost:8080/api/v1/whatsapp/sessions/invalid-uuid/status" \
  -H "Authorization: Bearer {token}"

# Expected: 404 Not Found
```

### Test 7: Admin Access
```bash
# Admin should be able to access any session
curl -X GET "http://localhost:8080/api/v1/whatsapp/sessions/{any_session_id}/status" \
  -H "Authorization: Bearer {admin_token}"

# Expected: 200 OK with session status
```

## HTTP Status Codes

### Reconnect Session Endpoint
- `200 OK` - Reconnection initiated successfully
- `400 Bad Request` - Invalid input or session already connected
- `401 Unauthorized` - Authentication required
- `403 Forbidden` - Insufficient permissions
- `404 Not Found` - Session not found
- `500 Internal Server Error` - Server error during reconnection

### Get Session Status Endpoint
- `200 OK` - Status retrieved successfully
- `400 Bad Request` - Invalid session ID
- `401 Unauthorized` - Authentication required
- `403 Forbidden` - Insufficient permissions
- `404 Not Found` - Session not found
- `500 Internal Server Error` - Server error retrieving status

## Error Response Structure

All error responses follow this structure:
```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message",
    "details": null
  }
}
```

## Requirements Coverage

### Requirement 5.2: Session Status and Health Monitoring
✅ **WHEN checking session status THEN the system SHALL return connection state**
- Implemented via `GetSessionStatus` endpoint
- Returns detailed status including connection state, logged-in state, and activity

### Requirement 5.4: Session Status Monitoring
✅ **WHEN monitoring sessions THEN the system SHALL track last activity timestamp**
- Status endpoint returns `last_activity` timestamp
- Real-time connection status from client

### Requirement 9.1: Error Handling
✅ **WHEN any WhatsApp operation fails THEN the system SHALL return structured error responses**
- All endpoints return structured error responses
- Proper error codes and messages

### Requirement 9.2: Service Availability
✅ **WHEN the WhatsApp service is unavailable THEN the system SHALL return appropriate HTTP status codes**
- Returns 500 Internal Server Error for service failures
- Returns 503 Service Unavailable when appropriate

### Requirement 9.3: Rate Limiting
✅ **WHEN rate limits are exceeded THEN the system SHALL return 429 Too Many Requests**
- Rate limiting middleware already applied to all routes
- Proper HTTP status codes implemented

## Integration Points

### Service Layer
- `WhatsAppService.ReconnectSession()` - Handles reconnection logic
- `WhatsAppService.GetSessionStatus()` - Retrieves session status

### Client Manager
- `ClientManager.CreateClient()` - Creates new client if needed
- `ClientManager.GetClient()` - Retrieves existing client
- `Client.Connect()` - Initiates connection
- `Client.IsConnected()` - Checks connection status
- `Client.IsLoggedIn()` - Checks authentication status

### Repository Layer
- `WhatsAppSessionRepository.FindByID()` - Retrieves session
- `WhatsAppSessionRepository.Update()` - Updates session status

## Security Considerations

1. **Authentication**: Both endpoints require valid JWT token
2. **Authorization**: 
   - Reconnect requires `whatsapp:write` permission
   - Status requires `whatsapp:read` permission
3. **Ownership**: Users can only access their own sessions (unless admin)
4. **Input Validation**: Session ID validated before processing

## Performance Considerations

1. **Status Endpoint**: Lightweight operation, queries database and checks client state
2. **Reconnect Endpoint**: May take time to establish connection, but returns immediately after initiating
3. **Concurrent Access**: Thread-safe client manager handles concurrent requests

## Monitoring and Logging

The implementation logs:
- Session reconnection attempts
- Connection failures
- Status check requests
- Authorization failures

## Next Steps

After verification:
1. Test with real WhatsApp sessions
2. Monitor reconnection success rates
3. Verify status accuracy during various connection states
4. Test under load with multiple concurrent requests
5. Validate error handling in edge cases

## Completion Checklist

- [x] Service method `ReconnectSession` implemented
- [x] Service method `GetSessionStatus` already exists
- [x] Response type `SessionStatusResponse` added
- [x] Handler `ReconnectSession` implemented
- [x] Handler `GetSessionStatus` implemented
- [x] Routes registered in router
- [x] Proper HTTP status codes implemented
- [x] Structured error responses implemented
- [x] Authentication and authorization applied
- [x] Swagger documentation added
- [x] Requirements coverage verified
