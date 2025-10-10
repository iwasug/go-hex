# Task 20: Session Reconnection Logic - Implementation Summary

## Overview
Implemented automatic reconnection logic with exponential backoff for WhatsApp sessions that disconnect unexpectedly. The system now tracks reconnection attempts, applies exponential backoff delays, and marks sessions as failed after exceeding the maximum retry limit.

## Changes Made

### 1. Database Schema Updates

#### Migration Files Created
- **`migrations/018_add_reconnection_tracking.up.sql`**
  - Added `reconnect_attempts` column (INTEGER, default 0)
  - Added `last_reconnect_attempt` column (TIMESTAMP, nullable)
  - Created index on `(status, reconnect_attempts)` for efficient querying
  - Added column comments for documentation

- **`migrations/018_add_reconnection_tracking.down.sql`**
  - Rollback migration to remove reconnection tracking fields

### 2. Domain Entity Updates

#### `internal/domain/entity/whatsapp_session.go`
Added reconnection tracking fields to `WhatsAppSession`:
- `ReconnectAttempts int` - Counter for reconnection attempts
- `LastReconnectAttempt *time.Time` - Timestamp of last attempt

Added new methods:
- `IncrementReconnectAttempts()` - Increments counter and updates timestamp
- `ResetReconnectAttempts()` - Resets counter to 0 and clears timestamp
- `MarkAsFailed()` - Sets status to `SessionStatusFailed`
- `ShouldAttemptReconnect(maxAttempts int) bool` - Checks if reconnection should be attempted

### 3. Repository Interface Updates

#### `internal/domain/repository/whatsapp_session_repository.go`
Added new methods:
- `IncrementReconnectAttempts(ctx, id) error`
- `ResetReconnectAttempts(ctx, id) error`

### 4. Repository Implementation Updates

#### `internal/adapter/postgres/whatsapp_session_repository.go`
- Updated all SELECT queries to include `reconnect_attempts` and `last_reconnect_attempt`
- Updated all scan operations to populate the new fields
- Implemented `IncrementReconnectAttempts()` method
- Implemented `ResetReconnectAttempts()` method

### 5. Event Bus Reconnection Logic

#### `internal/adapter/whatsapp/event_bus.go`
Completely rewrote `attemptReconnection()` method with:

**Exponential Backoff Implementation:**
- Initial delay: 5 seconds
- Maximum delay: 5 minutes
- Backoff multiplier: 2x
- Delay calculation: `delay = initialDelay * (2 ^ attemptNumber)`

**Reconnection Flow:**
1. Check if session should attempt reconnection (based on max attempts)
2. Calculate delay using exponential backoff based on previous attempts
3. Wait for calculated delay
4. Increment reconnection attempts in database
5. Update session status to "connecting"
6. Get client and check if already connected
7. Attempt to connect
8. On success: Reset reconnection attempts and update status to "connected"
9. On failure: Update status to "disconnected" and schedule next attempt
10. After max attempts: Mark session as "failed"

**Updated Event Handlers:**
- `HandleConnectedEvent()` - Now resets reconnection attempts on successful connection
- `HandlePairSuccessEvent()` - Now resets reconnection attempts on successful pairing

## Reconnection Behavior

### Configuration
```go
const (
    maxAttempts       = 5              // Maximum reconnection attempts
    initialDelay      = 5 * time.Second  // Initial delay before first retry
    maxDelay          = 5 * time.Minute  // Maximum delay between retries
    backoffMultiplier = 2              // Exponential backoff multiplier
)
```

### Delay Schedule
| Attempt | Delay      |
|---------|------------|
| 1       | 5 seconds  |
| 2       | 10 seconds |
| 3       | 20 seconds |
| 4       | 40 seconds |
| 5       | 80 seconds |

### Session Status Transitions
```
Disconnected → Connecting → Connected (success)
                         → Disconnected → Connecting (retry)
                                       → Failed (max attempts reached)
```

## Database Schema

### New Columns in `whatsapp_sessions`
```sql
reconnect_attempts INTEGER NOT NULL DEFAULT 0
last_reconnect_attempt TIMESTAMP
```

### New Index
```sql
CREATE INDEX idx_whatsapp_sessions_reconnect 
ON whatsapp_sessions(status, reconnect_attempts)
WHERE status IN ('disconnected', 'connecting');
```

## Key Features

### 1. Automatic Reconnection
- Triggered automatically when a session disconnects
- Runs asynchronously in a goroutine
- Does not block other operations

### 2. Exponential Backoff
- Prevents overwhelming the WhatsApp servers
- Increases delay between attempts exponentially
- Caps at maximum delay to avoid excessive waiting

### 3. Attempt Tracking
- Persists reconnection attempts in database
- Survives application restarts
- Provides visibility into reconnection history

### 4. Status Management
- Updates session status during reconnection process
- Marks sessions as "failed" after max attempts
- Resets attempts counter on successful connection

### 5. Graceful Degradation
- Continues attempting reconnection up to max limit
- Provides clear logging at each step
- Handles edge cases (client not found, already connected)

## Error Handling

### Scenarios Handled
1. **Client Not Found**: Stops reconnection attempts
2. **Already Connected**: Resets attempts and updates status
3. **Connection Failure**: Schedules next attempt with backoff
4. **Max Attempts Reached**: Marks session as failed
5. **Database Errors**: Logs errors but continues operation

## Logging

All reconnection activities are logged with appropriate levels:
- **Info**: Successful connections, attempt starts
- **Warn**: Max attempts reached
- **Error**: Connection failures, database errors

Log fields include:
- `session_id`
- `attempt` number
- `delay` duration
- `reconnect_attempts` count
- Error details

## Requirements Satisfied

✅ **Requirement 5.2**: Automatic reconnection on unexpected disconnect
✅ **Requirement 5.5**: Exponential backoff for reconnection attempts
✅ **Requirement 9.5**: Track reconnection attempts and max limit
✅ **Task Detail**: Update session status during reconnection
✅ **Task Detail**: Mark session as failed after max attempts

## Testing Recommendations

### Manual Testing
1. Create a session and connect
2. Simulate disconnect (turn off network)
3. Observe reconnection attempts in logs
4. Verify exponential backoff delays
5. Verify session marked as failed after 5 attempts
6. Restore network and verify successful reconnection resets counter

### Integration Testing
```go
// Test reconnection with exponential backoff
// Test max attempts reached
// Test successful reconnection resets counter
// Test session status transitions
// Test database persistence of attempts
```

## Future Enhancements

### Potential Improvements
1. Make reconnection parameters configurable via environment variables
2. Add metrics/monitoring for reconnection success rates
3. Implement circuit breaker pattern for repeated failures
4. Add webhook notifications for failed sessions
5. Implement manual reconnection trigger via API

## Migration Instructions

### To Apply
```bash
# Run migration
make migrate-up

# Or manually
psql -d your_database -f migrations/018_add_reconnection_tracking.up.sql
```

### To Rollback
```bash
# Rollback migration
make migrate-down

# Or manually
psql -d your_database -f migrations/018_add_reconnection_tracking.down.sql
```

## Verification

### Check Implementation
```bash
# Build the application
go build -o bin/api.exe ./cmd/api

# Run tests (when available)
go test ./internal/adapter/whatsapp/...
go test ./internal/adapter/postgres/...
```

### Database Verification
```sql
-- Check new columns exist
SELECT column_name, data_type, column_default 
FROM information_schema.columns 
WHERE table_name = 'whatsapp_sessions' 
AND column_name IN ('reconnect_attempts', 'last_reconnect_attempt');

-- Check index exists
SELECT indexname, indexdef 
FROM pg_indexes 
WHERE tablename = 'whatsapp_sessions' 
AND indexname = 'idx_whatsapp_sessions_reconnect';
```

## Conclusion

The session reconnection logic has been successfully implemented with:
- ✅ Exponential backoff algorithm
- ✅ Attempt tracking in database
- ✅ Status management during reconnection
- ✅ Automatic failure marking after max attempts
- ✅ Proper error handling and logging
- ✅ Clean integration with existing event system

The implementation follows the hexagonal architecture pattern and integrates seamlessly with the existing WhatsApp multi-device integration.
