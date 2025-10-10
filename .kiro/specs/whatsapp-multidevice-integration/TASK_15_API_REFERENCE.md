# Task 15 API Reference: Session Operations

## Overview
This document provides detailed API reference for the session operations endpoints (reconnect and status).

---

## Reconnect Session

Attempts to reconnect a disconnected WhatsApp session.

### Endpoint
```
POST /api/v1/whatsapp/sessions/{id}/reconnect
```

### Authentication
Requires valid JWT token in Authorization header.

### Authorization
Requires `whatsapp:write` permission.

### Path Parameters
| Parameter | Type   | Required | Description |
|-----------|--------|----------|-------------|
| id        | string | Yes      | Session UUID |

### Request Headers
```http
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

### Request Body
None

### Success Response

**Code**: `200 OK`

**Content**:
```json
{
  "success": true,
  "message": "Session reconnection initiated successfully"
}
```

### Error Responses

#### Session Not Found
**Code**: `404 Not Found`

**Content**:
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

#### Already Connected
**Code**: `400 Bad Request`

**Content**:
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Session is already connected",
    "details": null
  }
}
```

#### Unauthorized Access
**Code**: `403 Forbidden`

**Content**:
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

#### Connection Failed
**Code**: `500 Internal Server Error`

**Content**:
```json
{
  "success": false,
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "Failed to reconnect session: connection error",
    "details": null
  }
}
```

#### Missing Authentication
**Code**: `401 Unauthorized`

**Content**:
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

#### Invalid Session ID
**Code**: `400 Bad Request`

**Content**:
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Session ID is required",
    "details": null
  }
}
```

### Example Requests

#### cURL
```bash
curl -X POST "http://localhost:8080/api/v1/whatsapp/sessions/123e4567-e89b-12d3-a456-426614174000/reconnect" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

#### JavaScript (Fetch)
```javascript
fetch('http://localhost:8080/api/v1/whatsapp/sessions/123e4567-e89b-12d3-a456-426614174000/reconnect', {
  method: 'POST',
  headers: {
    'Authorization': 'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...',
    'Content-Type': 'application/json'
  }
})
.then(response => response.json())
.then(data => console.log(data));
```

#### Python (Requests)
```python
import requests

url = "http://localhost:8080/api/v1/whatsapp/sessions/123e4567-e89b-12d3-a456-426614174000/reconnect"
headers = {
    "Authorization": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}

response = requests.post(url, headers=headers)
print(response.json())
```

### Notes
- Reconnection is asynchronous; the endpoint returns immediately after initiating the connection
- Session status will be updated to "connecting" during the reconnection process
- If reconnection succeeds, status will be updated to "connected"
- If reconnection fails, status will be updated to "failed"
- Monitor the session status endpoint to track reconnection progress

---

## Get Session Status

Retrieves the current status and connection information for a WhatsApp session.

### Endpoint
```
GET /api/v1/whatsapp/sessions/{id}/status
```

### Authentication
Requires valid JWT token in Authorization header.

### Authorization
Requires `whatsapp:read` permission.

### Path Parameters
| Parameter | Type   | Required | Description |
|-----------|--------|----------|-------------|
| id        | string | Yes      | Session UUID |

### Request Headers
```http
Authorization: Bearer <jwt_token>
```

### Request Body
None

### Success Response

**Code**: `200 OK`

**Content**:
```json
{
  "success": true,
  "message": "Session status retrieved successfully",
  "data": {
    "session_id": "123e4567-e89b-12d3-a456-426614174000",
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

### Response Fields

| Field            | Type    | Description |
|------------------|---------|-------------|
| session_id       | string  | Session UUID |
| status           | string  | Session status (disconnected, connecting, connected, authenticating, failed) |
| phone_number     | string  | Phone number stored in database (may be empty if not yet connected) |
| connected_phone  | string  | Currently connected phone number from WhatsApp client (real-time) |
| last_activity    | string  | ISO 8601 timestamp of last activity |
| is_active        | boolean | Whether session is marked as active in database |
| is_connected     | boolean | Real-time connection status from WhatsApp client |
| is_logged_in     | boolean | Real-time authentication status from WhatsApp client |

### Status Values

| Status         | Description |
|----------------|-------------|
| disconnected   | Session is not connected to WhatsApp |
| connecting     | Session is attempting to connect |
| connected      | Session is connected and authenticated |
| authenticating | Session is in the process of authenticating |
| failed         | Session connection or authentication failed |

### Error Responses

#### Session Not Found
**Code**: `404 Not Found`

**Content**:
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

#### Unauthorized Access
**Code**: `403 Forbidden`

**Content**:
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

#### Missing Authentication
**Code**: `401 Unauthorized`

**Content**:
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

#### Invalid Session ID
**Code**: `400 Bad Request`

**Content**:
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Session ID is required",
    "details": null
  }
}
```

#### Server Error
**Code**: `500 Internal Server Error`

**Content**:
```json
{
  "success": false,
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "Failed to get session status",
    "details": null
  }
}
```

### Example Requests

#### cURL
```bash
curl -X GET "http://localhost:8080/api/v1/whatsapp/sessions/123e4567-e89b-12d3-a456-426614174000/status" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

#### JavaScript (Fetch)
```javascript
fetch('http://localhost:8080/api/v1/whatsapp/sessions/123e4567-e89b-12d3-a456-426614174000/status', {
  method: 'GET',
  headers: {
    'Authorization': 'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...'
  }
})
.then(response => response.json())
.then(data => console.log(data));
```

#### Python (Requests)
```python
import requests

url = "http://localhost:8080/api/v1/whatsapp/sessions/123e4567-e89b-12d3-a456-426614174000/status"
headers = {
    "Authorization": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}

response = requests.get(url, headers=headers)
print(response.json())
```

### Notes
- Status information combines database state with real-time client state
- `is_connected` and `is_logged_in` reflect real-time status from the WhatsApp client
- If the client doesn't exist, these fields will be `false`
- `phone_number` is the stored value, while `connected_phone` is the real-time value
- Use this endpoint to monitor session health and connection status

---

## Common Use Cases

### 1. Monitor Session Health
Poll the status endpoint periodically to monitor session health:

```javascript
async function monitorSession(sessionId, token) {
  const response = await fetch(
    `http://localhost:8080/api/v1/whatsapp/sessions/${sessionId}/status`,
    {
      headers: { 'Authorization': `Bearer ${token}` }
    }
  );
  const data = await response.json();
  
  if (data.success) {
    console.log('Session status:', data.data.status);
    console.log('Connected:', data.data.is_connected);
    console.log('Logged in:', data.data.is_logged_in);
  }
}

// Poll every 30 seconds
setInterval(() => monitorSession(sessionId, token), 30000);
```

### 2. Auto-Reconnect on Disconnect
Automatically reconnect when a session becomes disconnected:

```javascript
async function checkAndReconnect(sessionId, token) {
  // Check status
  const statusResponse = await fetch(
    `http://localhost:8080/api/v1/whatsapp/sessions/${sessionId}/status`,
    {
      headers: { 'Authorization': `Bearer ${token}` }
    }
  );
  const statusData = await statusResponse.json();
  
  // If disconnected, attempt reconnection
  if (statusData.success && !statusData.data.is_connected) {
    console.log('Session disconnected, attempting reconnection...');
    
    const reconnectResponse = await fetch(
      `http://localhost:8080/api/v1/whatsapp/sessions/${sessionId}/reconnect`,
      {
        method: 'POST',
        headers: { 'Authorization': `Bearer ${token}` }
      }
    );
    const reconnectData = await reconnectResponse.json();
    
    if (reconnectData.success) {
      console.log('Reconnection initiated');
    } else {
      console.error('Reconnection failed:', reconnectData.error.message);
    }
  }
}
```

### 3. Dashboard Status Display
Display session status in a dashboard:

```javascript
async function displaySessionStatus(sessionId, token) {
  const response = await fetch(
    `http://localhost:8080/api/v1/whatsapp/sessions/${sessionId}/status`,
    {
      headers: { 'Authorization': `Bearer ${token}` }
    }
  );
  const data = await response.json();
  
  if (data.success) {
    const status = data.data;
    return {
      statusBadge: getStatusBadge(status.status),
      connectionIndicator: status.is_connected ? '🟢 Online' : '🔴 Offline',
      phoneNumber: status.connected_phone || status.phone_number || 'Not connected',
      lastSeen: new Date(status.last_activity).toLocaleString()
    };
  }
}

function getStatusBadge(status) {
  const badges = {
    'connected': '✅ Connected',
    'connecting': '🔄 Connecting',
    'disconnected': '❌ Disconnected',
    'authenticating': '🔐 Authenticating',
    'failed': '⚠️ Failed'
  };
  return badges[status] || status;
}
```

---

## Rate Limiting

Both endpoints are subject to the application's rate limiting:
- Default: 100 requests per minute per IP
- Burst: 200 requests

If rate limit is exceeded:
```json
{
  "success": false,
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Too many requests",
    "details": null
  }
}
```

**Response Code**: `429 Too Many Requests`

**Headers**:
```
Retry-After: 60
```

---

## Security Considerations

1. **Authentication**: Always include valid JWT token
2. **Authorization**: Ensure user has required permissions
3. **Ownership**: Users can only access their own sessions (unless admin)
4. **Token Expiry**: Handle token expiration and refresh
5. **HTTPS**: Use HTTPS in production to protect tokens

---

## Best Practices

1. **Polling**: Don't poll status too frequently (recommended: 30-60 seconds)
2. **Error Handling**: Always handle error responses gracefully
3. **Reconnection**: Implement exponential backoff for reconnection attempts
4. **Monitoring**: Log status changes for debugging
5. **User Feedback**: Provide clear feedback to users about session status

---

## Troubleshooting

### Reconnection Fails Repeatedly
- Check if session credentials are still valid
- Verify WhatsApp account is not banned
- Check network connectivity
- Review server logs for detailed error messages

### Status Shows Connected but Messages Fail
- Check if phone is still paired with WhatsApp
- Verify session hasn't been logged out from phone
- Try reconnecting the session

### Unauthorized Errors
- Verify JWT token is valid and not expired
- Check user has required permissions
- Ensure user owns the session (or is admin)

---

## Related Endpoints

- [Create Session](./TASK_13_API_REFERENCE.md#create-session)
- [Get Session](./TASK_13_API_REFERENCE.md#get-session)
- [Delete Session](./TASK_13_API_REFERENCE.md#delete-session)
- [Get QR Code](./TASK_13_API_REFERENCE.md#get-qr-code)
- [Send Message](./TASK_14_API_REFERENCE.md#send-text-message)
