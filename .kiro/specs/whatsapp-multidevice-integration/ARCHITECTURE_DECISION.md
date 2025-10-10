# Architecture Decision: Full PostgreSQL Implementation

## Decision

We will implement a **custom PostgreSQL store** for whatsmeow instead of using the default SQLite storage.

## Context

The whatsmeow library by default uses SQLite for storing cryptographic keys and device data. While this works for simple deployments, it creates significant challenges for production Kubernetes deployments:

1. **Scaling Limitations**: SQLite file locks prevent horizontal scaling
2. **High Availability**: Single file = single point of failure
3. **Kubernetes Challenges**: Requires ReadWriteOnce volumes, limiting to single pod
4. **Operational Complexity**: File-based backups and disaster recovery

## Solution

Implement a custom `PostgresStore` that satisfies the `store.Store` interface from whatsmeow, storing all data in PostgreSQL.

### What Gets Stored in PostgreSQL

1. **Device Keys** (whatsmeow requirement):
   - Identity keys (public/private key pairs)
   - Pre-keys and signed pre-keys
   - Session keys (Signal protocol)
   - Sender keys (group encryption)
   - App state sync keys

2. **Application Data**:
   - Session metadata
   - Message history
   - User associations
   - Webhook configuration

### Implementation Approach

```go
type PostgresStore struct {
    db        *sql.DB
    sessionID string
    log       logger.Logger
}

// Implements store.Store interface
// - GetDevice() / PutDevice()
// - GetIdentity() / PutIdentity()
// - GetPreKeys() / PutPreKey()
// - GetSession() / PutSession()
// - GetSenderKey() / PutSenderKey()
// etc.
```

## Benefits

### For Kubernetes Deployment

1. **True Horizontal Scaling**
   - Scale from 1 to 100+ pods seamlessly
   - No session affinity required
   - Any pod can handle any request

2. **High Availability**
   - PostgreSQL clustering for redundancy
   - Automatic failover
   - No single point of failure

3. **Stateless Pods**
   - No local file storage
   - Easy rolling updates
   - Fast pod startup/shutdown

4. **Cloud Native**
   - Works with managed PostgreSQL (RDS, Cloud SQL, Azure Database)
   - Standard backup/restore procedures
   - Geographic replication support

### For Operations

1. **Simplified Management**
   - One database system (PostgreSQL)
   - Standard monitoring tools
   - Familiar backup procedures

2. **Better Observability**
   - Query performance metrics
   - Connection pool monitoring
   - Standard PostgreSQL dashboards

3. **Disaster Recovery**
   - Point-in-time recovery
   - WAL archiving
   - Cross-region replication

### For Development

1. **Consistent Stack**
   - Same database for all data
   - No SQLite dependencies
   - Easier local development

2. **Better Testing**
   - Use test PostgreSQL instances
   - Easy data seeding
   - Parallel test execution

## Trade-offs

### Advantages ✅

- Unlimited horizontal scaling
- High availability out of the box
- Cloud-native architecture
- Standard operational procedures
- Better performance at scale
- No file storage dependencies

### Disadvantages ⚠️

- Custom implementation required (~500-1000 LOC)
- Need to maintain compatibility with whatsmeow updates
- Slightly more complex initial setup
- Must thoroughly test cryptographic operations

## Implementation Complexity

**Estimated Effort**: 3-5 days

1. **Day 1-2**: Implement core store interface methods
2. **Day 3**: Testing with whatsmeow client
3. **Day 4**: Integration testing with real WhatsApp
4. **Day 5**: Performance optimization and edge cases

**Maintenance**: Low - whatsmeow store interface is stable

## Alternatives Considered

### Alternative 1: Default SQLite (Per-Session)
- ❌ Cannot scale horizontally
- ❌ File management complexity
- ❌ Not Kubernetes-friendly
- ✅ Zero implementation effort

### Alternative 2: Default SQLite (Centralized)
- ⚠️ Better than per-session but still limited
- ❌ Still requires ReadWriteOnce volume
- ❌ Single pod limitation
- ✅ Zero implementation effort

### Alternative 3: Fork whatsmeow
- ❌ High maintenance burden
- ❌ Must track upstream changes
- ❌ Complex merge conflicts
- ⚠️ Full control over implementation

### Alternative 4: PostgreSQL Store (Chosen)
- ✅ Best for production Kubernetes
- ✅ True horizontal scaling
- ✅ Standard operations
- ⚠️ Custom implementation needed

## Decision Rationale

Given the requirements for:
- High performance
- Production Kubernetes deployment
- Horizontal scaling capability
- High availability

The **custom PostgreSQL store** is the clear winner despite the initial implementation effort. The long-term operational benefits far outweigh the upfront development cost.

## Success Criteria

1. ✅ All whatsmeow store interface methods implemented
2. ✅ Successful QR code pairing
3. ✅ Send/receive messages working
4. ✅ Multiple concurrent sessions supported
5. ✅ Horizontal scaling verified (3+ pods)
6. ✅ Session persistence across pod restarts
7. ✅ Performance: <10ms for key lookups
8. ✅ Zero data corruption under load

## References

- [whatsmeow store interface](https://github.com/tulir/whatsmeow/blob/main/store/store.go)
- [Signal Protocol](https://signal.org/docs/)
- [PostgreSQL BYTEA type](https://www.postgresql.org/docs/current/datatype-binary.html)
- [Kubernetes StatefulSet vs Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/)
