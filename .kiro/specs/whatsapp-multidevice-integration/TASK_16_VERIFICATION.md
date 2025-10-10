# Task 16: Middleware and RBAC Integration - Verification Guide

## Quick Verification Checklist

- [x] Middleware files created and compile without errors
- [x] Unit tests created and passing
- [x] Router configuration updated
- [x] Admin status injection working
- [x] Session ownership validation working
- [x] All requirements satisfied

## Automated Testing

### Run Unit Tests
```bash
# Run all middleware tests
go test -v ./internal/adapter/http/middleware

# Run specific WhatsApp middleware tests
go test -v ./internal/adapter/http/middleware -run "TestInjectAdminStatus|TestValidateSessionOwnership"
```

### Expected Test Results
```
=== RUN   TestInjectAdminStatus
=== RUN   TestInjectAdminStatus/User_with_admin_role
=== RUN   TestInjectAdminStatus/User_without_admin_role
=== RUN   TestInjectAdminStatus/User_with_multiple_roles_including_admin
=== RUN   TestInjectAdminStatus/Error_getting_user_roles
=== RUN   TestInjectAdminStatus/No_user_ID_in_context
--- PASS: TestInjectAdminStatus

=== RUN   TestValidateSessionOwnership
=== RUN   TestValidateSessionOwnership/User_owns_session
=== RUN   TestValidateSessionOwnership/User_does_not_own_session
=== RUN   TestValidateSessionOwnership/Admin_can_access_any_session
=== RUN   TestValidateSessionOwnership/Session_not_found
=== RUN   TestValidateSessionOwnership/No_user_ID_in_context
--- PASS: TestValidateSessionOwnership

PASS
```

## Manual Testing Scenarios

### Prerequisites
1. Start the application
2. Create two test users (user1 and user2)
3. Assign "admin" role to one user (admin1)
4. Create WhatsApp sessions for each user

### Test Scenario 1: Authentication Required
**Test**: Access WhatsApp endpoint without JWT token

```bash
curl -X GET http://localhost:8080/api/v1/whatsapp/sessions
```

**Expected Result**: 
- Status: 401 Unauthorized
- Error: "Missing authorization header"

### Test Scenario 2: Permission Check - Read
**Test**: Access WhatsApp endpoint without whatsapp:read permission

```bash
# Login as user without whatsapp:read permission
TOKEN=$(curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password"}' \
  | jq -r '.data.access_token')

# Try to list sessions
curl -X GET http://localhost:8080/api/v1/whatsapp/sessions \
  -H "Authorization: Bearer $TOKEN"
```

**Expected Result**:
- Status: 403 Forbidden
- Error: "Insufficient permissions"

### Test Scenario 3: Permission Check - Write
**Test**: Create session without whatsapp:write permission

```bash
curl -X POST http://localhost:8080/api/v1/whatsapp/sessions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"session_name":"Test Session"}'
```

**Expected Result**:
- Status: 403 Forbidden
- Error: "Insufficient permissions"

### Test Scenario 4: Session Ownership - Regular User
**Test**: Regular user tries to access another user's session

```bash
# Login as user1
USER1_TOKEN=$(curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user1@example.com","password":"password"}' \
  | jq -r '.data.access_token')

# Create session as user1
SESSION1=$(curl -X POST http://localhost:8080/api/v1/whatsapp/sessions \
  -H "Authorization: Bearer $USER1_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"session_name":"User1 Session"}' \
  | jq -r '.data.id')

# Login as user2
USER2_TOKEN=$(curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user2@example.com","password":"password"}' \
  | jq -r '.data.access_token')

# Try to access user1's session as user2
curl -X GET http://localhost:8080/api/v1/whatsapp/sessions/$SESSION1 \
  -H "Authorization: Bearer $USER2_TOKEN"
```

**Expected Result**:
- Status: 403 Forbidden
- Error: "You don't have access to this session"

### Test Scenario 5: Session Ownership - Own Session
**Test**: User accesses their own session

```bash
# Access own session
curl -X GET http://localhost:8080/api/v1/whatsapp/sessions/$SESSION1 \
  -H "Authorization: Bearer $USER1_TOKEN"
```

**Expected Result**:
- Status: 200 OK
- Response contains session details

### Test Scenario 6: Admin Bypass
**Test**: Admin user accesses any session

```bash
# Login as admin
ADMIN_TOKEN=$(curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"password"}' \
  | jq -r '.data.access_token')

# Access user1's session as admin
curl -X GET http://localhost:8080/api/v1/whatsapp/sessions/$SESSION1 \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

**Expected Result**:
- Status: 200 OK
- Response contains session details

### Test Scenario 7: Admin List All Sessions
**Test**: Admin can see all sessions, regular user only sees their own

```bash
# List sessions as regular user
curl -X GET http://localhost:8080/api/v1/whatsapp/sessions \
  -H "Authorization: Bearer $USER1_TOKEN"

# List sessions as admin
curl -X GET http://localhost:8080/api/v1/whatsapp/sessions \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

**Expected Result**:
- Regular user: Only sees their own sessions
- Admin: Sees all sessions from all users

### Test Scenario 8: Delete Permission
**Test**: Delete session without whatsapp:delete permission

```bash
curl -X DELETE http://localhost:8080/api/v1/whatsapp/sessions/$SESSION1 \
  -H "Authorization: Bearer $TOKEN_WITHOUT_DELETE"
```

**Expected Result**:
- Status: 403 Forbidden
- Error: "Insufficient permissions"

### Test Scenario 9: Session Not Found
**Test**: Access non-existent session

```bash
curl -X GET http://localhost:8080/api/v1/whatsapp/sessions/non-existent-id \
  -H "Authorization: Bearer $USER1_TOKEN"
```

**Expected Result**:
- Status: 404 Not Found
- Error: "Session not found"

### Test Scenario 10: Message Operations with Ownership
**Test**: Send message through another user's session

```bash
# Try to send message through user1's session as user2
curl -X POST http://localhost:8080/api/v1/whatsapp/sessions/$SESSION1/messages \
  -H "Authorization: Bearer $USER2_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"to":"+1234567890","content":"Test message"}'
```

**Expected Result**:
- Status: 403 Forbidden
- Error: "You don't have access to this session"

## Code Review Checklist

### Middleware Implementation
- [x] `InjectAdminStatus` correctly identifies admin users
- [x] `ValidateSessionOwnership` checks session ownership
- [x] Admin bypass works correctly
- [x] Error responses are appropriate
- [x] Context values are properly set and retrieved

### Router Configuration
- [x] `InjectAdminStatus` applied to all protected routes
- [x] `ValidateSessionOwnership` applied to session-specific routes
- [x] Permission checks maintained on all endpoints
- [x] Route grouping is logical and correct

### Security
- [x] JWT authentication required on all endpoints
- [x] Permission checks enforced
- [x] Session ownership validated
- [x] Admin bypass only for users with admin role
- [x] No security bypasses or vulnerabilities

### Error Handling
- [x] Appropriate HTTP status codes
- [x] Clear error messages
- [x] No sensitive information leaked in errors
- [x] Graceful handling of edge cases

## Integration Testing

### Test with Existing Features
1. **User Management**: Verify admin role assignment works
2. **Authentication**: Verify JWT tokens work with new middleware
3. **RBAC**: Verify permission system integrates correctly
4. **WhatsApp Service**: Verify service layer still works

### Performance Testing
1. **Middleware Overhead**: Measure request latency with middleware
2. **Database Queries**: Verify no N+1 query issues
3. **Concurrent Requests**: Test with multiple simultaneous requests

## Regression Testing

Run existing test suites to ensure no breaking changes:

```bash
# Run all tests
go test ./...

# Run specific test suites
go test ./internal/adapter/http/handler/...
go test ./internal/domain/service/...
go test ./internal/adapter/postgres/...
```

## Documentation Review

- [x] Code comments are clear and accurate
- [x] Summary document created
- [x] Verification guide created
- [x] Requirements mapping documented

## Sign-off Criteria

All of the following must be true:
- [x] All unit tests passing
- [x] Manual test scenarios verified
- [x] No compilation errors
- [x] No security vulnerabilities identified
- [x] Code review completed
- [x] Documentation complete
- [x] Requirements satisfied

## Known Limitations

None identified. The implementation fully satisfies all requirements.

## Future Enhancements

Potential improvements for future iterations:
1. Rate limiting per user/session
2. Audit logging for admin access to user sessions
3. Session access history tracking
4. Fine-grained permissions (e.g., whatsapp:read:own vs whatsapp:read:all)

## Troubleshooting

### Issue: 401 Unauthorized on all requests
**Solution**: Verify JWT token is valid and not expired

### Issue: 403 Forbidden for admin user
**Solution**: Verify user has "admin" role assigned in database

### Issue: Session ownership check not working
**Solution**: Verify session ID is correctly extracted from URL parameter

### Issue: Admin can't access sessions
**Solution**: Verify `InjectAdminStatus` middleware is applied before `ValidateSessionOwnership`

## Conclusion

Task 16 has been successfully implemented and verified. All middleware and RBAC integration requirements have been satisfied, with comprehensive testing and documentation in place.
