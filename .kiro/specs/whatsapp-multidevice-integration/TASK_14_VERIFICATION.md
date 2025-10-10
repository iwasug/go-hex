# Task 14: HTTP Handlers - Messaging - Verification Guide

## Overview
This document provides verification steps and test examples for the messaging HTTP handlers implementation.

## Prerequisites

1. Application running on `http://localhost:8080`
2. Valid JWT token with `whatsapp:read` and `whatsapp:write` permissions
3. Active WhatsApp session (connected and authenticated)
4. Session ID from a created session

## Environment Setup

```bash
# Set your JWT token
export TOKEN="your_jwt_token_here"

# Set your session ID
export SESSION_ID="your_session_id_here"
```

## Test Cases

### 1. Send Text Message

#### Valid Text Message
```bash
curl -X POST http://localhost:8080/api/v1/whatsapp/sessions/${SESSION_ID}/messages \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "to": "+1234567890",
    "content": "Hello from WhatsApp API!"
  }'
```

**Expected Response (200 OK)**:
```json
{
  "success": true,
  "message": "Message sent successfully",
  "data": {
    "message_id": "3EB0XXXXXXXXXXXXX",
    "status": "sent",
    "timestamp": "2024-10-09T14:35:00Z",
    "message": {
      "id": "uuid-here",
      "session_id": "session-uuid",
      "message_id": "3EB0XXXXXXXXXXXXX",
      "direction": "outbound",
      "status": "sent",
      "from": "+1234567890",
      "to": "+0987654321",
      "content": "Hello from WhatsApp API!",
      "message_type": "text",
      "timestamp": "2024-10-09T14:35:00Z",
      "webhook_delivered": false,
      "webhook_attempts": 0,
      "created_at": "2024-10-09T14:35:00Z",
      "updated_at": "2024-10-09T14:35:00Z"
    }
  }
}
```

#### Missing Required Fields
```bash
curl -X POST http://localhost:8080/api/v1/whatsapp/sessions/${SESSION_ID}/messages \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "to": "+1234567890"
  }'
```

**Expected Response (400 Bad Request)**:
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": {
      "fields": {
        "content": "content is required"
      }
    }
  }
}
```

#### Invalid Phone Number Format
```bash
curl -X POST http://localhost:8080/api/v1/whatsapp/sessions/${SESSION_ID}/messages \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "to": "1234567890",
    "content": "Test message"
  }'
```

**Expected Response (400 Bad Request)**:
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "phone number must start with + (E.164 format)",
    "details": null
  }
}
```

#### Session Not Connected
```bash
# Using a disconnected session
curl -X POST http://localhost:8080/api/v1/whatsapp/sessions/${DISCONNECTED_SESSION_ID}/messages \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "to": "+1234567890",
    "content": "Test message"
  }'
```

**Expected Response (400 Bad Request)**:
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Session is not connected",
    "details": null
  }
}
```

### 2. Send Media Message

#### Send Image with Caption
```bash
curl -X POST http://localhost:8080/api/v1/whatsapp/sessions/${SESSION_ID}/media \
  -H "Authorization: Bearer ${TOKEN}" \
  -F "to=+1234567890" \
  -F "caption=Check out this image!" \
  -F "file=@/path/to/image.jpg"
```

**Expected Response (200 OK)**:
```json
{
  "success": true,
  "message": "Media message sent successfully",
  "data": {
    "message_id": "3EB0XXXXXXXXXXXXX",
    "status": "sent",
    "timestamp": "2024-10-09T14:35:00Z",
    "message": {
      "id": "uuid-here",
      "session_id": "session-uuid",
      "message_id": "3EB0XXXXXXXXXXXXX",
      "direction": "outbound",
      "status": "sent",
      "from": "+1234567890",
      "to": "+0987654321",
      "content": "Check out this image!",
      "message_type": "image",
      "timestamp": "2024-10-09T14:35:00Z",
      "webhook_delivered": false,
      "webhook_attempts": 0,
      "created_at": "2024-10-09T14:35:00Z",
      "updated_at": "2024-10-09T14:35:00Z"
    }
  }
}
```

#### Send Video
```bash
curl -X POST http://localhost:8080/api/v1/whatsapp/sessions/${SESSION_ID}/media \
  -H "Authorization: Bearer ${TOKEN}" \
  -F "to=+1234567890" \
  -F "caption=Watch this video" \
  -F "file=@/path/to/video.mp4"
```

#### Send Audio
```bash
curl -X POST http://localhost:8080/api/v1/whatsapp/sessions/${SESSION_ID}/media \
  -H "Authorization: Bearer ${TOKEN}" \
  -F "to=+1234567890" \
  -F "file=@/path/to/audio.mp3"
```

#### Send Document with Filename
```bash
curl -X POST http://localhost:8080/api/v1/whatsapp/sessions/${SESSION_ID}/media \
  -H "Authorization: Bearer ${TOKEN}" \
  -F "to=+1234567890" \
  -F "filename=report.pdf" \
  -F "file=@/path/to/document.pdf"
```

#### File Too Large
```bash
# Upload a file larger than 16MB
curl -X POST http://localhost:8080/api/v1/whatsapp/sessions/${SESSION_ID}/media \
  -H "Authorization: Bearer ${TOKEN}" \
  -F "to=+1234567890" \
  -F "file=@/path/to/large_file.mp4"
```

**Expected Response (413 Request Entity Too Large)**:
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "File size exceeds maximum allowed size of 16MB",
    "details": null
  }
}
```

#### Missing File
```bash
curl -X POST http://localhost:8080/api/v1/whatsapp/sessions/${SESSION_ID}/media \
  -H "Authorization: Bearer ${TOKEN}" \
  -F "to=+1234567890" \
  -F "caption=Test"
```

**Expected Response (400 Bad Request)**:
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "File is required",
    "details": null
  }
}
```

### 3. Get Message History

#### Get All Messages (Default Pagination)
```bash
curl -X GET "http://localhost:8080/api/v1/whatsapp/sessions/${SESSION_ID}/messages" \
  -H "Authorization: Bearer ${TOKEN}"
```

**Expected Response (200 OK)**:
```json
{
  "success": true,
  "message": "Messages retrieved successfully",
  "data": {
    "data": [
      {
        "id": "uuid-1",
        "session_id": "session-uuid",
        "message_id": "3EB0XXXXXXXXXXXXX",
        "direction": "outbound",
        "status": "sent",
        "from": "+1234567890",
        "to": "+0987654321",
        "content": "Hello World",
        "message_type": "text",
        "timestamp": "2024-10-09T14:35:00Z",
        "webhook_delivered": false,
        "webhook_attempts": 0,
        "created_at": "2024-10-09T14:35:00Z",
        "updated_at": "2024-10-09T14:35:00Z"
      },
      {
        "id": "uuid-2",
        "session_id": "session-uuid",
        "message_id": "3EB0YYYYYYYYYYYYY",
        "direction": "inbound",
        "status": "delivered",
        "from": "+0987654321",
        "to": "+1234567890",
        "content": "Hi there!",
        "message_type": "text",
        "timestamp": "2024-10-09T14:30:00Z",
        "webhook_delivered": true,
        "webhook_attempts": 1,
        "created_at": "2024-10-09T14:30:00Z",
        "updated_at": "2024-10-09T14:30:00Z"
      }
    ],
    "pagination": {
      "total": 25,
      "page": 1,
      "limit": 50,
      "total_pages": 1
    }
  }
}
```

#### Filter by Direction (Inbound Only)
```bash
curl -X GET "http://localhost:8080/api/v1/whatsapp/sessions/${SESSION_ID}/messages?direction=inbound" \
  -H "Authorization: Bearer ${TOKEN}"
```

#### Filter by Direction (Outbound Only)
```bash
curl -X GET "http://localhost:8080/api/v1/whatsapp/sessions/${SESSION_ID}/messages?direction=outbound" \
  -H "Authorization: Bearer ${TOKEN}"
```

#### Filter by Date Range
```bash
curl -X GET "http://localhost:8080/api/v1/whatsapp/sessions/${SESSION_ID}/messages?start_date=2024-10-01T00:00:00Z&end_date=2024-10-09T23:59:59Z" \
  -H "Authorization: Bearer ${TOKEN}"
```

#### Custom Pagination
```bash
# Get page 2 with 20 items per page
curl -X GET "http://localhost:8080/api/v1/whatsapp/sessions/${SESSION_ID}/messages?page=2&limit=20" \
  -H "Authorization: Bearer ${TOKEN}"
```

#### Combined Filters
```bash
curl -X GET "http://localhost:8080/api/v1/whatsapp/sessions/${SESSION_ID}/messages?direction=inbound&start_date=2024-10-01T00:00:00Z&page=1&limit=10" \
  -H "Authorization: Bearer ${TOKEN}"
```

#### Invalid Direction
```bash
curl -X GET "http://localhost:8080/api/v1/whatsapp/sessions/${SESSION_ID}/messages?direction=invalid" \
  -H "Authorization: Bearer ${TOKEN}"
```

**Expected Response (400 Bad Request)**:
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid direction. Must be 'inbound' or 'outbound'",
    "details": null
  }
}
```

#### Invalid Date Format
```bash
curl -X GET "http://localhost:8080/api/v1/whatsapp/sessions/${SESSION_ID}/messages?start_date=2024-10-01" \
  -H "Authorization: Bearer ${TOKEN}"
```

**Expected Response (400 Bad Request)**:
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid start_date format. Use ISO 8601 format",
    "details": null
  }
}
```

### 4. Authorization Tests

#### Missing JWT Token
```bash
curl -X POST http://localhost:8080/api/v1/whatsapp/sessions/${SESSION_ID}/messages \
  -H "Content-Type: application/json" \
  -d '{
    "to": "+1234567890",
    "content": "Test"
  }'
```

**Expected Response (401 Unauthorized)**:
```json
{
  "success": false,
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Missing or invalid token",
    "details": null
  }
}
```

#### Insufficient Permissions
```bash
# Using a token without whatsapp:write permission
curl -X POST http://localhost:8080/api/v1/whatsapp/sessions/${SESSION_ID}/messages \
  -H "Authorization: Bearer ${TOKEN_WITHOUT_PERMISSION}" \
  -H "Content-Type: application/json" \
  -d '{
    "to": "+1234567890",
    "content": "Test"
  }'
```

**Expected Response (403 Forbidden)**:
```json
{
  "success": false,
  "error": {
    "code": "FORBIDDEN",
    "message": "Insufficient permissions",
    "details": null
  }
}
```

#### Access Another User's Session
```bash
# Using a valid token but trying to access another user's session
curl -X POST http://localhost:8080/api/v1/whatsapp/sessions/${OTHER_USER_SESSION_ID}/messages \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "to": "+1234567890",
    "content": "Test"
  }'
```

**Expected Response (403 Forbidden)**:
```json
{
  "success": false,
  "error": {
    "code": "FORBIDDEN",
    "message": "You don't have access to this session",
    "details": null
  }
}
```

### 5. Error Handling Tests

#### Session Not Found
```bash
curl -X POST http://localhost:8080/api/v1/whatsapp/sessions/non-existent-id/messages \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "to": "+1234567890",
    "content": "Test"
  }'
```

**Expected Response (404 Not Found)**:
```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "Session not found",
    "details": null
  }
}
```

#### Invalid JSON Body
```bash
curl -X POST http://localhost:8080/api/v1/whatsapp/sessions/${SESSION_ID}/messages \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d 'invalid json'
```

**Expected Response (400 Bad Request)**:
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request body",
    "details": null
  }
}
```

## Verification Checklist

### Text Message Sending
- [ ] Successfully send text message with valid input
- [ ] Validate phone number format (E.164)
- [ ] Validate content is required
- [ ] Validate content max length (65536 characters)
- [ ] Check session ownership
- [ ] Verify session is connected
- [ ] Return message ID and status
- [ ] Handle session not found error
- [ ] Handle unauthorized access error
- [ ] Handle message send failure

### Media Message Sending
- [ ] Successfully send image with caption
- [ ] Successfully send video
- [ ] Successfully send audio
- [ ] Successfully send document with filename
- [ ] Validate file is required
- [ ] Validate file size (max 16MB)
- [ ] Validate MIME type
- [ ] Determine media type from MIME type
- [ ] Handle file too large error
- [ ] Handle unsupported media type error
- [ ] Check session ownership
- [ ] Verify session is connected

### Message History Retrieval
- [ ] Get all messages with default pagination
- [ ] Filter by direction (inbound)
- [ ] Filter by direction (outbound)
- [ ] Filter by date range
- [ ] Custom pagination (page and limit)
- [ ] Combined filters
- [ ] Validate direction values
- [ ] Validate date format (ISO 8601)
- [ ] Check session ownership
- [ ] Return paginated response with metadata

### Authorization & Security
- [ ] Require JWT authentication
- [ ] Check whatsapp:write permission for sending
- [ ] Check whatsapp:read permission for retrieving
- [ ] Validate session ownership
- [ ] Prevent access to other users' sessions
- [ ] Admin can access all sessions

### Error Handling
- [ ] Handle missing required fields
- [ ] Handle invalid input formats
- [ ] Handle session not found
- [ ] Handle session not connected
- [ ] Handle unauthorized access
- [ ] Handle insufficient permissions
- [ ] Handle file upload errors
- [ ] Return structured error responses

## Performance Considerations

1. **File Upload**: Large files (up to 16MB) may take time to upload
2. **Message History**: Pagination helps with large message volumes
3. **Date Range Filtering**: Improves query performance for large datasets
4. **Connection Check**: Validates session status before sending

## Next Steps

After verification:
1. Test with real WhatsApp sessions
2. Monitor message delivery status
3. Test webhook delivery for incoming messages
4. Verify message persistence in database
5. Test concurrent message sending
6. Load test with multiple sessions

## Notes

- All timestamps are in ISO 8601 format (UTC)
- Phone numbers must be in E.164 format (+[country][number])
- Message IDs are generated by WhatsApp
- Media files are validated for type and size
- Pagination defaults: page=1, limit=50, max=100
- Date filtering uses inclusive ranges
