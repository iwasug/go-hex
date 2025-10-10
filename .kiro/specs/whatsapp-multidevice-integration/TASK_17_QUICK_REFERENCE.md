# Task 17: Router Configuration - Quick Reference

## Overview

Complete router configuration for WhatsApp multi-device integration with authentication, RBAC, and rate limiting.

## Route Structure

```
/api/v1/whatsapp/
├── sessions/
│   ├── POST   /                    - Create session (whatsapp:write)
│   ├── GET    /                    - List sessions (whatsapp:read)
│   └── {id}/
│       ├── GET    /                - Get session (whatsapp:read)
│       ├── DELETE /                - Delete session (whatsapp:delete)
│       ├── GET    /qr              - Get QR code (whatsapp:read)
│       ├── GET    /status          - Get status (whatsapp:read)
│       ├── POST   /reconnect       - Reconnect (whatsapp:write)
│       ├── POST   /messages        - Send text (whatsapp:write)
│       ├── POST   /media           - Send media (whatsapp:write)
│       └── GET    /messages        - Get messages (whatsapp:read)
```

## Middleware Stack

### Global Middleware (All Routes)
1. Recovery
2. RequestID
3. Logger
4. SecurityHeaders
5. Metrics
6. Tracing
7. **RateLimiter** (100 req/s, burst 200)
8. Compress
9. CORS
10. Timeout (60s)

### Protected Routes Middleware
11. **Auth** (JWT validation)
12. **InjectAdminStatus** (admin context)

### Per-Endpoint Middleware
13. **RequirePermission** (RBAC check)

### Session Routes Middleware
14. **ValidateSessionOwnership** (ownership check)

## Permission Matrix

| Action | Permission | Admin Bypass |
|--------|-----------|--------------|
| Create session | `whatsapp:write` | No |
| List sessions | `whatsapp:read` | Yes (sees all) |
| View session | `whatsapp:read` | Yes |
| Delete session | `whatsapp:delete` | Yes |
| Get QR code | `whatsapp:read` | Yes |
| Send message | `whatsapp:write` | Yes |
| View messages | `whatsapp:read` | Yes |
| Reconnect | `whatsapp:write` | Yes |

## Rate Limiting

**Configuration**:
- Rate: 100 requests/second per IP
- Burst: 200 requests
- Cleanup: 3 minutes

**Response on Limit Exceeded**:
```http
HTTP/1.1 429 Too Many Requests
Content-Type: application/json
Retry-After: 60

{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Too many requests. Please try again later."
  }
}
```

## Authentication

**Required Header**:
```
Authorization: Bearer <JWT_TOKEN>
```

**Error Responses**:
- `401 Unauthorized` - Missing or invalid token
- `403 Forbidden` - Missing permission or unauthorized access

## Session Ownership

**Rules**:
- Regular users: Can only access their own sessions
- Admin users: Can access all sessions
- Validation: Automatic via middleware

**Error Response**:
```json
{
  "error": {
    "code": "FORBIDDEN",
    "message": "You don't have permission to access this session"
  }
}
```

## Testing Commands

### Test Authentication
```bash
# Without token (should fail)
curl -X GET http://localhost:8080/api/v1/whatsapp/sessions

# With token (should succeed)
curl -X GET http://localhost:8080/api/v1/whatsapp/sessions \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Test RBAC
```bash
# Without permission (should fail)
curl -X GET http://localhost:8080/api/v1/whatsapp/sessions \
  -H "Authorization: Bearer TOKEN_WITHOUT_PERMISSION"

# With permission (should succeed)
curl -X GET http://localhost:8080/api/v1/whatsapp/sessions \
  -H "Authorization: Bearer TOKEN_WITH_WHATSAPP_READ"
```

### Test Rate Limiting
```bash
# Exceed rate limit
for i in {1..101}; do
  curl -X GET http://localhost:8080/api/v1/whatsapp/sessions \
    -H "Authorization: Bearer YOUR_TOKEN"
done
# 101st request should return 429
```

### Test Session Ownership
```bash
# Access own session (should succeed)
curl -X GET http://localhost:8080/api/v1/whatsapp/sessions/YOUR_SESSION_ID \
  -H "Authorization: Bearer YOUR_TOKEN"

# Access other's session (should fail)
curl -X GET http://localhost:8080/api/v1/whatsapp/sessions/OTHER_SESSION_ID \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## Configuration

### Adjust Rate Limiting

Edit `internal/adapter/http/router.go`:

```go
// Current: 100 req/s, burst 200
rateLimiter := custommiddleware.NewRateLimiter(100, 200)

// Stricter: 50 req/s, burst 100
rateLimiter := custommiddleware.NewRateLimiter(50, 100)

// More lenient: 200 req/s, burst 400
rateLimiter := custommiddleware.NewRateLimiter(200, 400)
```

### Adjust Timeout

Edit `internal/adapter/http/router.go`:

```go
// Current: 60 seconds
r.Use(middleware.Timeout(60 * time.Second))

// Longer: 120 seconds
r.Use(middleware.Timeout(120 * time.Second))
```

## Security Features

✅ **Defense in Depth**
- Multiple security layers
- Early rejection of invalid requests

✅ **Principle of Least Privilege**
- Granular permissions
- Session ownership validation

✅ **Rate Limiting**
- Prevents abuse and DoS
- IP-based tracking

✅ **Admin Controls**
- Admin bypass for support
- Audit trail via logging

## Files

```
internal/adapter/http/
├── router.go                    # Main router configuration
└── middleware/
    ├── auth.go                  # JWT authentication
    ├── rbac.go                  # RBAC permission checks
    ├── ratelimit.go             # Rate limiting
    └── whatsapp.go              # Session ownership validation
```

## Common Issues

### Issue: 401 Unauthorized
**Cause**: Missing or invalid JWT token
**Solution**: Include valid token in Authorization header

### Issue: 403 Forbidden
**Cause**: Missing permission or unauthorized session access
**Solution**: Ensure user has required permission and owns the session

### Issue: 429 Too Many Requests
**Cause**: Rate limit exceeded
**Solution**: Wait 60 seconds or adjust rate limit configuration

### Issue: 404 Not Found
**Cause**: Incorrect route path
**Solution**: Verify route matches `/api/v1/whatsapp/...` pattern

## Next Steps

- Task 18: Configuration management
- Task 19: Media file handling
- Task 20: Session reconnection logic
- Task 21: Application startup integration

## Summary

✅ All routes registered under `/api/v1/whatsapp`
✅ JWT authentication on all routes
✅ RBAC with granular permissions
✅ Session ownership validation
✅ Rate limiting with proper headers
✅ Production-ready and secure
