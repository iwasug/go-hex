# Task 21: Application Startup Integration - Quick Reference

## What Was Implemented

### 1. Application Startup Integration
- WhatsApp service initialization on app startup
- Automatic restoration of connected sessions
- Non-blocking initialization (app continues if WhatsApp fails)

### 2. Graceful Shutdown
- WhatsApp clients shut down before HTTP server
- All active connections closed gracefully
- 30-second timeout for shutdown operations

### 3. Health Monitoring
- WhatsApp service status in health endpoint
- Real-time client metrics (total, connected, disconnected)
- Service marked as "degraded" if WhatsApp down

## Key Files Modified

```
cmd/api/main.go                                    # Startup & shutdown
internal/adapter/http/router.go                    # Lifecycle methods
internal/adapter/http/handler/health_handler.go    # Health checks
```

## New Methods

### Router (internal/adapter/http/router.go)
```go
// Initialize WhatsApp and restore sessions
func (rt *Router) InitializeWhatsApp(ctx context.Context) error

// Gracefully shutdown all WhatsApp clients
func (rt *Router) ShutdownWhatsApp(ctx context.Context) error

// Get client manager for health checks
func (rt *Router) GetClientManager() *whatsapp.ClientManager
```

### Health Handler (internal/adapter/http/handler/health_handler.go)
```go
// Set client manager for health checks
func (h *HealthHandler) SetClientManager(cm *whatsapp.ClientManager)

// Check WhatsApp service health
func (h *HealthHandler) checkWhatsApp() Status
```

## Usage Examples

### Starting the Application
```bash
# Application automatically:
# 1. Initializes WhatsApp service
# 2. Restores connected sessions from database
# 3. Reconnects each session
# 4. Logs restoration results

./bin/api.exe

# Expected logs:
# INFO: WhatsApp service initialized successfully
# INFO: found sessions to restore count=2
# INFO: session restored successfully session_id=xxx
# INFO: client restoration completed restored=2 failed=0
```

### Checking Health
```bash
# Get service health including WhatsApp status
curl http://localhost:8080/health

# Response:
{
  "status": "ok",
  "timestamp": "2024-10-09T15:38:40Z",
  "services": {
    "database": {
      "status": "up",
      "details": {...}
    },
    "whatsapp": {
      "status": "up",
      "details": {
        "total_clients": "2",
        "connected_clients": "2",
        "disconnected_clients": "0"
      }
    }
  }
}
```

### Graceful Shutdown
```bash
# Send shutdown signal (SIGTERM or SIGINT)
kill -TERM <pid>

# Or use Ctrl+C

# Expected logs:
# INFO: Shutdown signal received signal=terminated
# INFO: Shutting down WhatsApp clients...
# INFO: shutting down all WhatsApp clients count=2
# INFO: all WhatsApp clients shut down successfully
# INFO: WhatsApp clients shut down successfully
# INFO: Server stopped gracefully
```

## Application Lifecycle Flow

### Startup Sequence
```
1. Load configuration
2. Initialize database
3. Setup router
4. Initialize WhatsApp service
   ├─ Restore connected sessions
   ├─ Reconnect each session
   └─ Log results
5. Start HTTP server
```

### Shutdown Sequence
```
1. Receive shutdown signal
2. Shutdown WhatsApp clients
   ├─ Disconnect all clients
   ├─ Clear clients map
   └─ Log results
3. Shutdown HTTP server
4. Close database connection
5. Exit
```

## Health Check Response Structure

```json
{
  "status": "ok|degraded",
  "timestamp": "ISO8601",
  "services": {
    "database": {
      "status": "up|down",
      "details": {
        "open_connections": "N",
        "in_use": "N",
        "idle": "N"
      }
    },
    "whatsapp": {
      "status": "up|down",
      "details": {
        "total_clients": "N",
        "connected_clients": "N",
        "disconnected_clients": "N"
      }
    }
  }
}
```

## Error Handling

### Initialization Errors
- **Scenario**: WhatsApp initialization fails
- **Behavior**: Application continues to run
- **Status**: Service marked as "degraded"
- **Logging**: Error logged with context

### Shutdown Errors
- **Scenario**: Client disconnect fails
- **Behavior**: Continue shutting down other clients
- **Status**: Error logged, shutdown completes
- **Logging**: All errors logged with context

### Health Check Errors
- **Scenario**: Client manager not initialized
- **Behavior**: WhatsApp status shows "down"
- **Status**: Service marked as "degraded"
- **Response**: Error details in health response

## Configuration

### Shutdown Timeout
```go
// In cmd/api/main.go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
```

**Default**: 30 seconds
**Configurable**: Yes, modify timeout value in main.go

### Health Check Interval
- Health endpoint: `/health`
- Readiness probe: `/health/ready`
- Liveness probe: `/health/live`

## Monitoring

### Metrics Available
- Total WhatsApp clients
- Connected clients count
- Disconnected clients count
- Database connection stats

### Log Messages to Monitor
```
# Successful startup
"WhatsApp service initialized successfully"
"client restoration completed"

# Successful shutdown
"all WhatsApp clients shut down successfully"
"Server stopped gracefully"

# Errors
"Failed to initialize WhatsApp service"
"failed to restore session"
"failed to disconnect client during shutdown"
```

## Kubernetes Integration

### Liveness Probe
```yaml
livenessProbe:
  httpGet:
    path: /health/live
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10
```

### Readiness Probe
```yaml
readinessProbe:
  httpGet:
    path: /health/ready
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
```

### Graceful Shutdown
```yaml
lifecycle:
  preStop:
    exec:
      command: ["/bin/sh", "-c", "sleep 5"]
terminationGracePeriodSeconds: 35
```

## Troubleshooting

### Issue: Sessions not restoring on startup
**Check**:
1. Database connectivity
2. Session status in database (should be "connected")
3. Application logs for restoration errors

**Solution**:
```sql
-- Check connected sessions
SELECT id, session_name, status, phone_number 
FROM whatsapp_sessions 
WHERE status = 'connected' AND is_active = true;
```

### Issue: Shutdown hangs
**Check**:
1. Shutdown timeout (default 30s)
2. Client disconnect errors in logs
3. Network connectivity

**Solution**:
- Increase timeout if needed
- Check for stuck connections
- Force kill if timeout exceeded

### Issue: Health check shows WhatsApp down
**Check**:
1. Client manager initialization
2. Database connectivity
3. Session restoration logs

**Solution**:
```bash
# Check health endpoint
curl http://localhost:8080/health | jq .

# Check application logs
grep "WhatsApp" application.log
```

## Best Practices

1. **Monitor health endpoint** - Set up alerts for degraded status
2. **Check logs on startup** - Verify session restoration
3. **Test graceful shutdown** - Ensure clean shutdowns in production
4. **Set appropriate timeouts** - Balance between graceful and forced shutdown
5. **Monitor client metrics** - Track connected/disconnected clients

## Related Tasks

- **Task 6**: WhatsApp client manager (provides RestoreClients)
- **Task 20**: Session reconnection logic (used during restoration)
- **Task 22**: Error handling and logging (enhances error handling)

## Requirements Satisfied

- ✅ **1.6**: Automatic session restoration on startup
- ✅ **5.6**: Session verification on startup
- ✅ **9.2**: Proper error handling and logging
