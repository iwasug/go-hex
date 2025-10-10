# Task 14: Messaging API Reference

## Quick Reference for WhatsApp Messaging Endpoints

### Base URL
```
http://localhost:8080/api/v1/whatsapp/sessions
```

### Authentication
All endpoints require JWT Bearer token:
```
Authorization: Bearer {your_jwt_token}
```

---

## 1. Send Text Message

**Endpoint**: `POST /{session_id}/messages`

**Permission**: `whatsapp:write`

**Request Body**:
```json
{
  "to": "+1234567890",
  "content": "Your message here"
}
```

**Success Response (200)**:
```json
{
  "success": true,
  "message": "Message sent successfully",
  "data": {
    "message_id": "3EB0XXXXXXXXXXXXX",
    "status": "sent",
    "timestamp": "2024-10-09T14:35:00Z",
    "message": { /* full message object */ }
  }
}
```

**cURL Example**:
```bash
curl -X POST http://localhost:8080/api/v1/whatsapp/sessions/{session_id}/messages \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{"to": "+1234567890", "content": "Hello!"}'
```

---

## 2. Send Media Message

**Endpoint**: `POST /{session_id}/media`

**Permission**: `whatsapp:write`

**Content-Type**: `multipart/form-data`

**Form Fields**:
- `to` (required): Recipient phone number (E.164 format)
- `file` (required): Media file
- `caption` (optional): Caption for media
- `filename` (optional): Filename (required for documents)

**Success Response (200)**:
```json
{
  "success": true,
  "message": "Media message sent successfully",
  "data": {
    "message_id": "3EB0XXXXXXXXXXXXX",
    "status": "sent",
    "timestamp": "2024-10-09T14:35:00Z",
    "message": { /* full message object */ }
  }
}
```

**cURL Examples**:

**Image**:
```bash
curl -X POST http://localhost:8080/api/v1/whatsapp/sessions/{session_id}/media \
  -H "Authorization: Bearer {token}" \
  -F "to=+1234567890" \
  -F "caption=Check this out!" \
  -F "file=@image.jpg"
```

**Video**:
```bash
curl -X POST http://localhost:8080/api/v1/whatsapp/sessions/{session_id}/media \
  -H "Authorization: Bearer {token}" \
  -F "to=+1234567890" \
  -F "caption=Watch this" \
  -F "file=@video.mp4"
```

**Document**:
```bash
curl -X POST http://localhost:8080/api/v1/whatsapp/sessions/{session_id}/media \
  -H "Authorization: Bearer {token}" \
  -F "to=+1234567890" \
  -F "filename=report.pdf" \
  -F "file=@document.pdf"
```

---

## 3. Get Message History

**Endpoint**: `GET /{session_id}/messages`

**Permission**: `whatsapp:read`

**Query Parameters**:
- `direction` (optional): Filter by `inbound` or `outbound`
- `start_date` (optional): Start date (ISO 8601 format)
- `end_date` (optional): End date (ISO 8601 format)
- `page` (optional): Page number (default: 1)
- `limit` (optional): Items per page (default: 50, max: 100)

**Success Response (200)**:
```json
{
  "success": true,
  "message": "Messages retrieved successfully",
  "data": {
    "data": [
      {
        "id": "uuid",
        "session_id": "session-uuid",
        "message_id": "3EB0XXXXXXXXXXXXX",
        "direction": "outbound",
        "status": "sent",
        "from": "+1234567890",
        "to": "+0987654321",
        "content": "Message content",
        "message_type": "text",
        "timestamp": "2024-10-09T14:35:00Z",
        "webhook_delivered": false,
        "webhook_attempts": 0,
        "created_at": "2024-10-09T14:35:00Z",
        "updated_at": "2024-10-09T14:35:00Z"
      }
    ],
    "pagination": {
      "total": 100,
      "page": 1,
      "limit": 50,
      "total_pages": 2
    }
  }
}
```

**cURL Examples**:

**All Messages**:
```bash
curl -X GET "http://localhost:8080/api/v1/whatsapp/sessions/{session_id}/messages" \
  -H "Authorization: Bearer {token}"
```

**Inbound Only**:
```bash
curl -X GET "http://localhost:8080/api/v1/whatsapp/sessions/{session_id}/messages?direction=inbound" \
  -H "Authorization: Bearer {token}"
```

**Date Range**:
```bash
curl -X GET "http://localhost:8080/api/v1/whatsapp/sessions/{session_id}/messages?start_date=2024-10-01T00:00:00Z&end_date=2024-10-09T23:59:59Z" \
  -H "Authorization: Bearer {token}"
```

**Pagination**:
```bash
curl -X GET "http://localhost:8080/api/v1/whatsapp/sessions/{session_id}/messages?page=2&limit=20" \
  -H "Authorization: Bearer {token}"
```

**Combined Filters**:
```bash
curl -X GET "http://localhost:8080/api/v1/whatsapp/sessions/{session_id}/messages?direction=inbound&start_date=2024-10-01T00:00:00Z&page=1&limit=10" \
  -H "Authorization: Bearer {token}"
```

---

## Common Error Responses

### 400 Bad Request
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": {
      "fields": {
        "to": "to is required"
      }
    }
  }
}
```

### 401 Unauthorized
```json
{
  "success": false,
  "error": {
    "code": "UNAUTHORIZED",
    "message": "User ID not found in context",
    "details": null
  }
}
```

### 403 Forbidden
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

### 404 Not Found
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

### 413 Request Entity Too Large
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

### 500 Internal Server Error
```json
{
  "success": false,
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "Failed to send message",
    "details": null
  }
}
```

---

## Validation Rules

### Phone Numbers
- **Format**: E.164 (+[country][number])
- **Example**: `+1234567890`
- **Length**: 8-16 characters (including +)

### Text Messages
- **Content**: Required, max 65,536 characters
- **To**: Required, E.164 format

### Media Messages
- **File**: Required
- **Max Size**: 16MB (configurable)
- **Supported Types**:
  - **Images**: JPEG, PNG, GIF, WebP
  - **Videos**: MP4, 3GPP, QuickTime, Matroska
  - **Audio**: AAC, MP4, MPEG, AMR, OGG
  - **Documents**: PDF, Word, Excel, PowerPoint, Text, ZIP
- **Caption**: Optional, for all media types
- **Filename**: Required for documents

### Message History
- **Direction**: `inbound` or `outbound`
- **Dates**: ISO 8601 format (e.g., `2024-10-09T14:35:00Z`)
- **Page**: Positive integer
- **Limit**: 1-100 items per page

---

## Message Object Structure

```json
{
  "id": "uuid",
  "session_id": "session-uuid",
  "message_id": "3EB0XXXXXXXXXXXXX",
  "direction": "outbound",
  "status": "sent",
  "from": "+1234567890",
  "to": "+0987654321",
  "content": "Message content",
  "message_type": "text",
  "timestamp": "2024-10-09T14:35:00Z",
  "webhook_delivered": false,
  "webhook_attempts": 0,
  "error_message": "",
  "created_at": "2024-10-09T14:35:00Z",
  "updated_at": "2024-10-09T14:35:00Z"
}
```

### Field Descriptions

- **id**: Internal database ID (UUID)
- **session_id**: WhatsApp session ID
- **message_id**: WhatsApp message ID (from WhatsApp)
- **direction**: `inbound` or `outbound`
- **status**: `pending`, `sent`, `delivered`, `read`, `failed`
- **from**: Sender phone number (E.164)
- **to**: Recipient phone number (E.164)
- **content**: Message text content
- **message_type**: `text`, `image`, `video`, `audio`, `document`
- **timestamp**: Message timestamp from WhatsApp
- **webhook_delivered**: Whether webhook was delivered successfully
- **webhook_attempts**: Number of webhook delivery attempts
- **error_message**: Error message if failed
- **created_at**: Database creation timestamp
- **updated_at**: Database update timestamp

---

## Media Types

### Supported MIME Types

**Images**:
- `image/jpeg`, `image/jpg`
- `image/png`
- `image/gif`
- `image/webp`

**Videos**:
- `video/mp4`
- `video/3gpp`
- `video/quicktime`
- `video/x-matroska`

**Audio**:
- `audio/aac`
- `audio/mp4`
- `audio/mpeg`
- `audio/amr`
- `audio/ogg`

**Documents**:
- `application/pdf`
- `application/msword`
- `application/vnd.ms-excel`
- `application/vnd.openxmlformats-officedocument.wordprocessingml.document`
- `application/vnd.openxmlformats-officedocument.spreadsheetml.sheet`
- `application/vnd.openxmlformats-officedocument.presentationml.presentation`
- `text/plain`
- `application/zip`
- `application/x-zip-compressed`

---

## Rate Limiting

- Default: 100 requests per minute per IP
- Burst: 200 requests
- Response Header: `X-RateLimit-Remaining`

---

## Best Practices

1. **Phone Numbers**: Always use E.164 format with country code
2. **File Size**: Keep media files under 16MB for best performance
3. **Pagination**: Use reasonable page sizes (10-50 items)
4. **Date Filtering**: Use date ranges to improve query performance
5. **Error Handling**: Always check response status and handle errors
6. **Session Status**: Verify session is connected before sending messages
7. **Webhooks**: Configure webhooks to receive incoming messages
8. **Message IDs**: Store WhatsApp message IDs for tracking

---

## Testing Tips

1. **Use Postman**: Import cURL commands into Postman for easier testing
2. **Test Files**: Prepare test files of different types and sizes
3. **Phone Numbers**: Use test phone numbers or your own WhatsApp number
4. **Pagination**: Test with different page sizes and page numbers
5. **Date Ranges**: Test with various date ranges including edge cases
6. **Error Cases**: Test invalid inputs to verify error handling
7. **Concurrent Requests**: Test multiple simultaneous message sends

---

## Troubleshooting

### Message Not Sending
- Check session is connected
- Verify phone number format (E.164)
- Check user has `whatsapp:write` permission
- Verify session ownership

### File Upload Fails
- Check file size (max 16MB)
- Verify MIME type is supported
- Ensure multipart form is properly formatted

### Messages Not Appearing
- Check pagination parameters
- Verify date range filters
- Confirm session ownership
- Check direction filter

### Unauthorized Errors
- Verify JWT token is valid
- Check token has required permissions
- Ensure session belongs to user

---

## Related Documentation

- [Task 14 Summary](./TASK_14_SUMMARY.md)
- [Task 14 Verification Guide](./TASK_14_VERIFICATION.md)
- [WhatsApp Integration Architecture](./ARCHITECTURE_DECISION.md)
- [Webhook Integration Guide](./WEBHOOK_INTEGRATION.md)
