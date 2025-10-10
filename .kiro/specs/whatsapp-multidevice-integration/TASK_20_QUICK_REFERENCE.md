# Task 20: Session Reconnection Logic - Quick Reference

## Overview
Automatic reconnection with exponential backoff for disconnected WhatsApp sessions.

## Key Files Modified

### Database
- `migrations/018_add_reconnection_tracking.up.sql` - Add reconnection fields
- `migrations/018_add_reconnection_tracking.down.sql` - Rollback migration

### Domain Layer
- `internal/domain/entity/whatsapp_session.go` - Added reconnection fields and methods
- `internal/domain/repository/whatsapp_session_repository.go` - Added reconnection methods

### Adapter Layer
- `internal/adapter/postgres/whatsapp_session_repository.go` - Implemented reconnection tracking
- `internal/adapter/whatsapp/event_bus.go` - Implemented reconnection logic

## New Database Fields

```sql
reconnect_attempts INTEGER NOT NULL DEFAULT 0
last_reconnect_attempt TIMESTAMP
```

## Reconnection Configuration

```go
maxAttempts       = 5              // Maximum retry attempts
initialDelay      = 5 * time.Second  // Initial delay
maxDelay          = 5 * time.Minute  // Maximum delay
backoffMultiplier = 2              // Exponential multiplier
```

## Delay Schedule

| Attempt | Delay      |
|---------|------------|
| 1       | 5 seconds  |
| 2       | 10 seconds |
| 3       | 20 seconds |
| 4       | 40 seconds |
| 5       | 80 seconds |

## Status Flow

```
Connected → Disconnected → Connecting → Connected (success)
                                     → Disconnected (retry)
                                     → Failed (max attempts)
```

## Key Methods

### Entity Methods
```go
session.IncrementReconnectAttempts()           // Increment counter
session.ResetReconnectAttempts()               // Reset to 0
session.MarkAsFailed()                         // Set status to failed
session.ShouldAttemptReconnect(maxAttempts)    // Check if should retry
```

### Repository Methods
```go
repo.IncrementReconnectAttempts(ctx, id)  // Increment in DB
repo.ResetReconnectAttempts(ctx, id)      // Reset in DB
```

### Event Bus Methods
```go
eb.HandleDisconnectedEvent(ctx, sessionID, userID, clientManager)  // Triggers reconnection
eb.attemptReconnection(ctx, sessionID, userID, clientManager)      // Performs reconnection
```

## Usage Example

### Automatic Reconnection
```go
// Triggered automatically on disconnect
// No manual intervention required
```

### Manual Check
```go
// Get session
session, err := sessionRepo.FindByID(ctx, sessionID)

// Check reconnection status
if session.ReconnectAttempts >= 5 {
    // Max attempts reached
    // Session is marked as failed
}

// Check if should attempt reconnection
if session.ShouldAttemptReconnect(5) {
    // Can still attempt reconnection
}
```

## Database Queries

### Check Reconnection Status
```sql
SELECT 
    id,
    session_name,
    status,
    reconnect_attempts,
    last_reconnect_attempt
FROM whatsapp_sessions
WHERE status IN ('disconnected', 'connecting', 'failed')
ORDER BY last_reconnect_attempt DESC;
```

### Find Failed Sessions
```sql
SELECT 
    id,
    session_name,
    phone_number,
    reconnect_attempts,
    last_reconnect_attempt
FROM whatsapp_sessions
WHERE status = 'failed'
ORDER BY last_reconnect_attempt DESC;
```

### Reset Reconnection Attempts
```sql
UPDATE whatsapp_sessions
SET reconnect_attempts = 0,
    last_reconnect_attempt = NULL,
    status = 'disconnected'
WHERE id = 'session-id-here';
```

## Monitoring

### Log Patterns

**Reconnection Started**:
```
"starting reconnection attempt" session_id=... current_attempts=0 max_attempts=5
```

**Reconnection Success**:
```
"reconnection successful" session_id=... attempt=2
```

**Max Attempts Reached**:
```
"max reconnection attempts reached, marking session as failed" session_id=... reconnect_attempts=5
```

### Metrics to Monitor
- Number of sessions in "failed" status
- Average reconnection attempts before success
- Reconnection success rate
- Time to successful reconnection

## Troubleshooting

### Session Stuck in "Connecting"
```sql
-- Check reconnection status
SELECT * FROM whatsapp_sessions WHERE status = 'connecting';

-- Reset if needed
UPDATE whatsapp_sessions 
SET status = 'disconnected', reconnect_attempts = 0 
WHERE id = 'session-id';
```

### Too Many Failed Sessions
```sql
-- Find sessions that failed recently
SELECT 
    id,
    session_name,
    reconnect_attempts,
    last_reconnect_attempt,
    updated_at
FROM whatsapp_sessions
WHERE status = 'failed'
AND updated_at > NOW() - INTERVAL '1 hour';

-- Reset for retry
UPDATE whatsapp_sessions
SET status = 'disconnected',
    reconnect_attempts = 0,
    last_reconnect_attempt = NULL
WHERE status = 'failed'
AND id IN ('session-id-1', 'session-id-2');
```

### Check Reconnection History
```sql
-- Sessions with multiple reconnection attempts
SELECT 
    id,
    session_name,
    status,
    reconnect_attempts,
    last_reconnect_attempt,
    EXTRACT(EPOCH FROM (NOW() - last_reconnect_attempt)) as seconds_since_last_attempt
FROM whatsapp_sessions
WHERE reconnect_attempts > 0
ORDER BY reconnect_attempts DESC;
```

## API Impact

### Session Status Response
```json
{
  "id": "session-id",
  "status": "connecting",
  "reconnect_attempts": 2,
  "last_reconnect_attempt": "2024-01-15T10:30:00Z"
}
```

### Failed Session Response
```json
{
  "id": "session-id",
  "status": "failed",
  "reconnect_attempts": 5,
  "last_reconnect_attempt": "2024-01-15T10:35:00Z",
  "error": "Max reconnection attempts reached"
}
```

## Configuration (Future Enhancement)

### Environment Variables (Proposed)
```env
WHATSAPP_MAX_RECONNECT_ATTEMPTS=5
WHATSAPP_RECONNECT_INITIAL_DELAY=5s
WHATSAPP_RECONNECT_MAX_DELAY=5m
WHATSAPP_RECONNECT_BACKOFF_MULTIPLIER=2
```

## Testing Commands

### Build
```bash
go build -o bin/api.exe ./cmd/api
```

### Run Migration
```bash
# Apply migration
psql -d your_database -f migrations/018_add_reconnection_tracking.up.sql

# Rollback
psql -d your_database -f migrations/018_add_reconnection_tracking.down.sql
```

### Test Reconnection
```bash
# 1. Start application
./bin/api.exe

# 2. Create and connect session via API
curl -X POST http://localhost:8080/api/v1/whatsapp/sessions \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"session_name": "test"}'

# 3. Simulate disconnect (turn off network or kill WhatsApp connection)

# 4. Monitor logs for reconnection attempts
tail -f logs/app.log | grep "reconnection"

# 5. Restore network and verify reconnection
```

## Performance Notes

- Reconnection runs asynchronously (non-blocking)
- Database index optimizes queries for disconnected sessions
- Exponential backoff prevents server overload
- Maximum 5 attempts prevents infinite loops
- Goroutines properly managed (no leaks)

## Security Notes

- No sensitive data in reconnection logs
- Session ownership validated before reconnection
- Proper error handling prevents information disclosure
- Database constraints prevent invalid states

## Related Tasks

- Task 8: Event handling system (foundation)
- Task 6: WhatsApp client manager (client lifecycle)
- Task 5: PostgreSQL repository implementations (data persistence)

## Documentation

- Full Summary: `TASK_20_SUMMARY.md`
- Verification Guide: `TASK_20_VERIFICATION.md`
- This Quick Reference: `TASK_20_QUICK_REFERENCE.md`
