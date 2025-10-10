# Design Document: WhatsApp Multi-Device Integration

## Overview

This design document outlines the integration of WhatsApp multi-device functionality into the existing Go Chi Hexagonal Base project using the whatsmeow library. The integration maintains the hexagonal architecture pattern, ensuring clean separation between business logic, infrastructure, and external dependencies.

The whatsmeow library provides a Go implementation of the WhatsApp Web Multi-Device protocol, enabling the application to act as a WhatsApp client. This integration will support multiple concurrent sessions, allowing different users to connect their WhatsApp accounts independently.

### Key Design Principles

1. **Hexagonal Architecture Compliance**: All WhatsApp functionality follows the ports and adapters pattern
2. **Session Isolation**: Each user's WhatsApp session is completely isolated from others
3. **Persistence**: Session data persists across application restarts using SQLite (whatsmeow requirement)
4. **Scalability**: Design supports multiple concurrent sessions with efficient resource management
5. **Security**: Integration with existing JWT authentication and RBAC system
6. **Resilience**: Automatic reconnection, error handling, and graceful degradation

## Architecture

### High-Level Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                     HTTP Layer (Chi Router)                      │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  WhatsApp Handlers (QR, Send, Receive, Session Mgmt)    │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                            │
┌─────────────────────────────────────────────────────────────────┐
│                    Application/Domain Layer                      │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  WhatsApp Service (Business Logic)                       │  │
│  │  - Session Management                                     │  │
│  │  - Message Handling                                       │  │
│  │  - QR Code Generation                                     │  │
│  │  - Webhook Processing                                     │  │
│  └──────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Repository Interfaces (Ports)                           │  │
│  │  - WhatsAppSessionRepository                             │  │
│  │  - WhatsAppMessageRepository                             │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                            │
┌─────────────────────────────────────────────────────────────────┐
│                    Infrastructure/Adapter Layer                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  WhatsApp Client Manager (whatsmeow wrapper)             │  │
│  │  - Client lifecycle management                           │  │
│  │  - Event handling                                        │  │
│  │  - Connection management                                 │  │
│  │  - Uses custom PostgresStore                             │  │
│  └──────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  PostgreSQL Repositories & Custom Store                  │  │
│  │  - Session metadata storage                              │  │
│  │  - Message history storage                               │  │
│  │  - Device keys and authentication data (PostgresStore)   │  │
│  │  - Cryptographic keys (identity, pre-keys, sessions)     │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

### Layer Responsibilities

#### Domain Layer (`internal/domain/`)
- **Entities**: WhatsAppSession, WhatsAppMessage
- **Services**: WhatsAppService (business logic)
- **Repository Interfaces**: Define contracts for data access

#### Adapter Layer (`internal/adapter/`)
- **HTTP Handlers**: REST API endpoints for WhatsApp operations
- **WhatsApp Client**: Wrapper around whatsmeow library
- **PostgreSQL Repositories**: Implementation of repository interfaces
- **SQLite Store**: whatsmeow-specific storage adapter

## Components and Interfaces

### 1. Domain Entities

#### WhatsAppSession Entity

```go
package entity

import "time"

type SessionStatus string

const (
    SessionStatusDisconnected   SessionStatus = "disconnected"
    SessionStatusConnecting     SessionStatus = "connecting"
    SessionStatusConnected      SessionStatus = "connected"
    SessionStatusAuthenticating SessionStatus = "authenticating"
    SessionStatusFailed         SessionStatus = "failed"
)

type WhatsAppSession struct {
    ID              string        // UUID
    UserID          string        // Foreign key to users table
    SessionName     string        // User-friendly name
    PhoneNumber     string        // Connected WhatsApp number (E.164 format)
    Status          SessionStatus
    QRCode          string        // Base64 encoded PNG (temporary, for pairing)
    QRCodeExpiry    *time.Time
    DeviceInfo      string        // JSON: device details from WhatsApp
    LastActivity    time.Time
    WebhookURL      string        // Optional webhook for incoming messages
    WebhookSecret   string        // Optional secret for webhook authentication
    IsActive        bool
    CreatedBy       *string
    UpdatedBy       *string
    CreatedAt       time.Time
    UpdatedAt       time.Time
}
```

#### WhatsAppMessage Entity

```go
package entity

import "time"

type MessageDirection string
type MessageStatus string

const (
    MessageDirectionInbound  MessageDirection = "inbound"
    MessageDirectionOutbound MessageDirection = "outbound"
    
    MessageStatusPending   MessageStatus = "pending"
    MessageStatusSent      MessageStatus = "sent"
    MessageStatusDelivered MessageStatus = "delivered"
    MessageStatusRead      MessageStatus = "read"
    MessageStatusFailed    MessageStatus = "failed"
)

type WhatsAppMessage struct {
    ID              string
    SessionID       string           // Foreign key to whatsapp_sessions
    MessageID       string           // WhatsApp message ID
    Direction       MessageDirection
    Status          MessageStatus
    From            string           // Phone number (E.164)
    To              string           // Phone number (E.164)
    Content         string           // Message text content
    MessageType     string           // text, image, video, etc.
    Timestamp       time.Time        // Message timestamp from WhatsApp
    WebhookDelivered bool
    WebhookAttempts int
    ErrorMessage    string
    CreatedAt       time.Time
    UpdatedAt       time.Time
}
```

### 2. Repository Interfaces (Ports)

#### WhatsAppSessionRepository

```go
package repository

import (
    "context"
    "golang-chi-hexagonal-base/internal/domain/entity"
)

type ListSessionsParams struct {
    UserID         string
    Status         entity.SessionStatus
    IncludeAll     bool // Admin: view all sessions
    Limit          int
    Offset         int
}

type WhatsAppSessionRepository interface {
    Create(ctx context.Context, session *entity.WhatsAppSession) error
    FindByID(ctx context.Context, id string) (*entity.WhatsAppSession, error)
    FindByUserID(ctx context.Context, userID string) ([]*entity.WhatsAppSession, error)
    Update(ctx context.Context, session *entity.WhatsAppSession) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, params ListSessionsParams) ([]*entity.WhatsAppSession, int, error)
    UpdateStatus(ctx context.Context, id string, status entity.SessionStatus) error
    UpdateQRCode(ctx context.Context, id string, qrCode string, expiry time.Time) error
}
```

#### WhatsAppMessageRepository

```go
package repository

import (
    "context"
    "golang-chi-hexagonal-base/internal/domain/entity"
    "time"
)

type ListMessagesParams struct {
    SessionID  string
    Direction  entity.MessageDirection
    StartDate  *time.Time
    EndDate    *time.Time
    Limit      int
    Offset     int
}

type WhatsAppMessageRepository interface {
    Create(ctx context.Context, message *entity.WhatsAppMessage) error
    FindByID(ctx context.Context, id string) (*entity.WhatsAppMessage, error)
    FindByMessageID(ctx context.Context, messageID string) (*entity.WhatsAppMessage, error)
    Update(ctx context.Context, message *entity.WhatsAppMessage) error
    List(ctx context.Context, params ListMessagesParams) ([]*entity.WhatsAppMessage, int, error)
    UpdateStatus(ctx context.Context, id string, status entity.MessageStatus) error
    IncrementWebhookAttempts(ctx context.Context, id string) error
}
```

### 3. Domain Service

#### WhatsAppService

```go
package service

type WhatsAppService struct {
    sessionRepo    repository.WhatsAppSessionRepository
    messageRepo    repository.WhatsAppMessageRepository
    clientManager  WhatsAppClientManager // Interface to adapter layer
    userRepo       repository.UserRepository
    webhookClient  WebhookClient
}

// Core methods:
// - CreateSession(ctx, userID, sessionName) (*entity.WhatsAppSession, error)
// - GetSession(ctx, sessionID, userID) (*entity.WhatsAppSession, error)
// - ListSessions(ctx, userID, isAdmin) ([]*entity.WhatsAppSession, error)
// - DeleteSession(ctx, sessionID, userID) error
// - GenerateQRCode(ctx, sessionID, userID) (string, error)
// - SendTextMessage(ctx, sessionID, userID, to, content) (*entity.WhatsAppMessage, error)
// - SendMediaMessage(ctx, sessionID, userID, to string, media MediaMessage) (*entity.WhatsAppMessage, error)
// - GetMessages(ctx, sessionID, userID, params) ([]*entity.WhatsAppMessage, error)
// - HandleIncomingMessage(ctx, sessionID, message) error
// - ReconnectSession(ctx, sessionID) error
```

### 4. Adapter Layer Components

#### WhatsAppClientManager Interface

```go
package adapter

type WhatsAppClientManager interface {
    CreateClient(sessionID, userID string) error
    GetClient(sessionID string) (WhatsAppClient, error)
    DeleteClient(sessionID string) error
    RestoreClients() error // Called on app startup
}

type WhatsAppClient interface {
    Connect() error
    Disconnect() error
    GetQRCode() (string, error) // Returns base64 PNG
    SendTextMessage(to, content string) (messageID string, err error)
    SendMediaMessage(to string, media MediaMessage) (messageID string, err error)
    IsConnected() bool
    GetConnectedPhone() string
    RegisterEventHandlers(handlers EventHandlers)
}

type MediaMessage struct {
    Type     MediaType // image, video, audio, document
    Data     []byte    // Media file content
    MimeType string    // e.g., "image/jpeg", "video/mp4"
    Caption  string    // Optional caption for media
    FileName string    // Required for documents
}

type MediaType string

const (
    MediaTypeImage    MediaType = "image"
    MediaTypeVideo    MediaType = "video"
    MediaTypeAudio    MediaType = "audio"
    MediaTypeDocument MediaType = "document"
)

type EventHandlers struct {
    OnConnected    func(phone string)
    OnDisconnected func()
    OnQRCode       func(qrCode string)
    OnMessage      func(message IncomingMessage)
}

type IncomingMessage struct {
    MessageID   string
    From        string
    To          string
    Content     string
    MessageType string // text, image, video, audio, document
    MediaURL    string // For media messages
    Caption     string // For media with captions
    Timestamp   time.Time
}
```

#### WhatsApp Client Implementation (whatsmeow wrapper)

```go
package whatsapp

import (
    "go.mau.fi/whatsmeow"
    "go.mau.fi/whatsmeow/store/sqlstore"
)

type ClientManager struct {
    clients    map[string]*Client // sessionID -> Client
    mu         sync.RWMutex
    pgStore    *PostgresStore // Custom PostgreSQL store for whatsmeow
    logger     logger.Logger
    eventBus   EventBus // For publishing events to service layer
}

// PostgresStore implements whatsmeow's store.Store interface
type PostgresStore struct {
    db     *sql.DB
    logger logger.Logger
}

// Implements store.Store interface methods:
// - GetDevice(jid) (*store.Device, error)
// - PutDevice(device *store.Device) error
// - GetIdentity(address string) (*store.Identity, error)
// - PutIdentity(address string, identity *store.Identity) error
// - GetPreKeys(count int) ([]*store.PreKey, error)
// - PutPreKey(preKey *store.PreKey) error
// - GetSession(address string) (*store.Session, error)
// - PutSession(address string, session *store.Session) error
// - GetSenderKey(group, sender string) (*store.SenderKey, error)
// - PutSenderKey(group, sender string, key *store.SenderKey) error

type Client struct {
    sessionID    string
    userID       string
    waClient     *whatsmeow.Client
    device       *store.Device // Device from PostgresStore
    eventHandler *EventHandler
    connected    bool
    mu           sync.RWMutex
}
```

### 5. HTTP Handlers

#### WhatsAppHandler

```go
package handler

type WhatsAppHandler struct {
    whatsappService *service.WhatsAppService
}

// Endpoints:
// POST   /api/v1/whatsapp/sessions              - Create new session
// GET    /api/v1/whatsapp/sessions              - List user's sessions
// GET    /api/v1/whatsapp/sessions/:id          - Get session details
// DELETE /api/v1/whatsapp/sessions/:id          - Delete session
// GET    /api/v1/whatsapp/sessions/:id/qr       - Get QR code for pairing
// POST   /api/v1/whatsapp/sessions/:id/messages - Send text message
// POST   /api/v1/whatsapp/sessions/:id/media    - Send media message
// GET    /api/v1/whatsapp/sessions/:id/messages - Get message history
// POST   /api/v1/whatsapp/sessions/:id/reconnect - Reconnect session
```

## Data Models

### Database Schema (PostgreSQL)

#### whatsapp_sessions table

```sql
CREATE TABLE whatsapp_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_name VARCHAR(255) NOT NULL,
    phone_number VARCHAR(20),
    status VARCHAR(50) NOT NULL DEFAULT 'disconnected',
    qr_code TEXT,
    qr_code_expiry TIMESTAMP,
    device_info JSONB,
    last_activity TIMESTAMP NOT NULL DEFAULT NOW(),
    webhook_url TEXT,
    webhook_secret TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_by UUID REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, session_name)
);

CREATE INDEX idx_whatsapp_sessions_user_id ON whatsapp_sessions(user_id);
CREATE INDEX idx_whatsapp_sessions_status ON whatsapp_sessions(status);
CREATE INDEX idx_whatsapp_sessions_phone_number ON whatsapp_sessions(phone_number);
```

#### whatsapp_messages table

```sql
CREATE TABLE whatsapp_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES whatsapp_sessions(id) ON DELETE CASCADE,
    message_id VARCHAR(255) NOT NULL,
    direction VARCHAR(20) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    from_number VARCHAR(20) NOT NULL,
    to_number VARCHAR(20) NOT NULL,
    content TEXT NOT NULL,
    message_type VARCHAR(50) NOT NULL DEFAULT 'text',
    timestamp TIMESTAMP NOT NULL,
    webhook_delivered BOOLEAN NOT NULL DEFAULT false,
    webhook_attempts INTEGER NOT NULL DEFAULT 0,
    error_message TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(session_id, message_id)
);

CREATE INDEX idx_whatsapp_messages_session_id ON whatsapp_messages(session_id);
CREATE INDEX idx_whatsapp_messages_direction ON whatsapp_messages(direction);
CREATE INDEX idx_whatsapp_messages_timestamp ON whatsapp_messages(timestamp DESC);
CREATE INDEX idx_whatsapp_messages_status ON whatsapp_messages(status);
```

#### whatsapp_devices table (whatsmeow store)

```sql
CREATE TABLE whatsapp_devices (
    jid VARCHAR(255) PRIMARY KEY, -- WhatsApp JID (phone@s.whatsapp.net)
    session_id UUID NOT NULL REFERENCES whatsapp_sessions(id) ON DELETE CASCADE,
    registration_id INTEGER NOT NULL,
    identity_key BYTEA NOT NULL,
    signed_pre_key BYTEA NOT NULL,
    signed_pre_key_id INTEGER NOT NULL,
    signed_pre_key_signature BYTEA NOT NULL,
    adv_secret_key BYTEA,
    platform VARCHAR(50) DEFAULT 'go',
    business_name VARCHAR(255),
    push_name VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(session_id)
);

CREATE INDEX idx_whatsapp_devices_session_id ON whatsapp_devices(session_id);
```

#### whatsapp_identity_keys table

```sql
CREATE TABLE whatsapp_identity_keys (
    address VARCHAR(255) PRIMARY KEY, -- recipient JID
    session_id UUID NOT NULL REFERENCES whatsapp_sessions(id) ON DELETE CASCADE,
    identity_key BYTEA NOT NULL,
    trust_level INTEGER DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_whatsapp_identity_keys_session_id ON whatsapp_identity_keys(session_id);
```

#### whatsapp_pre_keys table

```sql
CREATE TABLE whatsapp_pre_keys (
    id SERIAL PRIMARY KEY,
    session_id UUID NOT NULL REFERENCES whatsapp_sessions(id) ON DELETE CASCADE,
    key_id INTEGER NOT NULL,
    public_key BYTEA NOT NULL,
    private_key BYTEA NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(session_id, key_id)
);

CREATE INDEX idx_whatsapp_pre_keys_session_id ON whatsapp_pre_keys(session_id);
```

#### whatsapp_sessions_keys table (Signal protocol sessions)

```sql
CREATE TABLE whatsapp_session_keys (
    id SERIAL PRIMARY KEY,
    session_id UUID NOT NULL REFERENCES whatsapp_sessions(id) ON DELETE CASCADE,
    address VARCHAR(255) NOT NULL, -- recipient JID
    session_data BYTEA NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(session_id, address)
);

CREATE INDEX idx_whatsapp_session_keys_session_id ON whatsapp_session_keys(session_id);
CREATE INDEX idx_whatsapp_session_keys_address ON whatsapp_session_keys(address);
```

#### whatsapp_sender_keys table (Group encryption)

```sql
CREATE TABLE whatsapp_sender_keys (
    id SERIAL PRIMARY KEY,
    session_id UUID NOT NULL REFERENCES whatsapp_sessions(id) ON DELETE CASCADE,
    group_id VARCHAR(255) NOT NULL,
    sender_id VARCHAR(255) NOT NULL,
    sender_key BYTEA NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(session_id, group_id, sender_id)
);

CREATE INDEX idx_whatsapp_sender_keys_session_id ON whatsapp_sender_keys(session_id);
CREATE INDEX idx_whatsapp_sender_keys_group_id ON whatsapp_sender_keys(group_id);
```

#### permissions table (additions)

```sql
-- Add WhatsApp permissions
INSERT INTO permissions (id, name, description) VALUES
    (gen_random_uuid(), 'whatsapp:read', 'View WhatsApp sessions and messages'),
    (gen_random_uuid(), 'whatsapp:write', 'Create sessions and send messages'),
    (gen_random_uuid(), 'whatsapp:delete', 'Delete WhatsApp sessions');
```

### PostgreSQL Storage (Custom Implementation)

**Storage Strategy**: Full PostgreSQL implementation for Kubernetes-native deployment with true horizontal scaling.

Instead of using whatsmeow's default SQLite store, we'll implement a **custom PostgreSQL store** that satisfies the whatsmeow store interface. This enables:
- ✅ True horizontal scaling in Kubernetes
- ✅ No file storage dependencies
- ✅ High availability with PostgreSQL clustering
- ✅ Standard backup/restore procedures
- ✅ No session affinity requirements

**What gets stored in PostgreSQL**:
1. **Device Keys** (whatsmeow requirement):
   - Identity keys (public/private key pairs)
   - Pre-keys and signed pre-keys
   - Session keys
   - Sender keys
   
2. **Application Data**:
   - Session metadata
   - Message history
   - User associations
   - Webhooks configuration

**Implementation**: We'll create a custom `PostgresStore` that implements the `store.Store` interface from whatsmeow, replacing the default `sqlstore.Container`.

## Error Handling

### Error Types

```go
package errors

var (
    ErrSessionNotFound        = NewAppError("SESSION_NOT_FOUND", "WhatsApp session not found", 404)
    ErrSessionNotConnected    = NewAppError("SESSION_NOT_CONNECTED", "WhatsApp session is not connected", 400)
    ErrSessionAlreadyExists   = NewAppError("SESSION_ALREADY_EXISTS", "Session with this name already exists", 409)
    ErrUnauthorizedSession    = NewAppError("UNAUTHORIZED_SESSION", "You don't have access to this session", 403)
    ErrQRCodeExpired          = NewAppError("QR_CODE_EXPIRED", "QR code has expired, request a new one", 410)
    ErrInvalidPhoneNumber     = NewAppError("INVALID_PHONE_NUMBER", "Invalid phone number format", 400)
    ErrMessageSendFailed      = NewAppError("MESSAGE_SEND_FAILED", "Failed to send message", 500)
    ErrWhatsAppServiceDown    = NewAppError("WHATSAPP_SERVICE_DOWN", "WhatsApp service is unavailable", 503)
)
```

### Error Handling Strategy

1. **Connection Errors**: Automatic retry with exponential backoff (3 attempts)
2. **Authentication Errors**: Mark session as requiring re-authentication
3. **Rate Limiting**: Return 429 with retry-after header
4. **Webhook Failures**: Retry up to 3 times with exponential backoff, then log and continue
5. **Storage Errors**: Log and return 500, ensure data consistency

## Testing Strategy

### Unit Tests

1. **Domain Layer**
   - Entity methods (session lifecycle, message creation)
   - Service business logic (session management, message handling)
   - Mock repository interfaces

2. **Adapter Layer**
   - Repository implementations (PostgreSQL queries)
   - HTTP handlers (request/response handling)
   - Mock whatsmeow client

### Integration Tests

1. **Database Integration**
   - Session CRUD operations
   - Message storage and retrieval
   - Transaction handling

2. **WhatsApp Client Integration**
   - QR code generation
   - Message sending (using test mode if available)
   - Event handling

### End-to-End Tests

1. **Complete Flows**
   - Session creation → QR scan → message send → message receive
   - Session reconnection after disconnect
   - Webhook delivery

## Security Considerations

### Authentication & Authorization

1. **JWT Authentication**: All endpoints require valid JWT token
2. **RBAC Integration**: Check `whatsapp:read`, `whatsapp:write`, `whatsapp:delete` permissions
3. **Session Ownership**: Users can only access their own sessions (except admins)
4. **Webhook Security**: Optional HMAC signature for webhook authentication

### Data Protection

1. **Encryption at Rest**: 
   - Webhook secrets encrypted in database
   - SQLite databases protected by file system permissions
2. **Encryption in Transit**: HTTPS for all API calls
3. **Sensitive Data**: Phone numbers and message content handled securely
4. **Audit Trail**: Track session creation, updates, and deletions

### Rate Limiting

1. **API Rate Limits**: Apply existing rate limiting middleware
2. **WhatsApp Rate Limits**: Respect WhatsApp's rate limits (handled by whatsmeow)
3. **Per-Session Limits**: Configurable message sending limits per session

## Configuration

### Environment Variables

```env
# WhatsApp Configuration
WHATSAPP_ENABLED=true
WHATSAPP_SQLITE_PATH=/data/whatsapp/whatsmeow_store.db
WHATSAPP_QR_TIMEOUT=60s
WHATSAPP_RECONNECT_INTERVAL=30s
WHATSAPP_MAX_RECONNECT_ATTEMPTS=5
WHATSAPP_MESSAGE_RETENTION_DAYS=90

# Webhook Configuration
WHATSAPP_WEBHOOK_TIMEOUT=10s
WHATSAPP_WEBHOOK_MAX_RETRIES=3
WHATSAPP_WEBHOOK_RETRY_DELAY=5s
```

### Configuration Structure

```go
package config

type WhatsAppConfig struct {
    Enabled                bool
    QRTimeout              time.Duration
    ReconnectInterval      time.Duration
    MaxReconnectAttempts   int
    MessageRetentionDays   int
    MediaMaxSizeMB         int    // Maximum media file size in MB
    MediaStoragePath       string // Path to store media files
    WebhookTimeout         time.Duration
    WebhookMaxRetries      int
    WebhookRetryDelay      time.Duration
    // Database connection uses existing config.DatabaseConfig
}
```

## Kubernetes Deployment

### Architecture Benefits for Kubernetes

The full PostgreSQL implementation provides true cloud-native capabilities:

1. **Horizontal Scaling**: Scale pods independently without data conflicts
2. **Stateless Pods**: No local file storage requirements
3. **High Availability**: PostgreSQL clustering for redundancy
4. **Rolling Updates**: Zero-downtime deployments
5. **Load Balancing**: Any pod can handle any request (no session affinity needed)

### Deployment Configuration

#### Deployment YAML

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: whatsapp-service
  namespace: production
spec:
  replicas: 3  # Horizontal scaling supported!
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: whatsapp-service
  template:
    metadata:
      labels:
        app: whatsapp-service
    spec:
      containers:
      - name: whatsapp-api
        image: your-registry/whatsapp-service:latest
        ports:
        - containerPort: 8080
          name: http
        env:
        - name: DB_HOST
          value: postgres-service
        - name: DB_PORT
          value: "5432"
        - name: DB_NAME
          value: whatsapp_db
        - name: DB_USER
          valueFrom:
            secretKeyRef:
              name: postgres-credentials
              key: username
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: postgres-credentials
              key: password
        - name: WHATSAPP_ENABLED
          value: "true"
        - name: WHATSAPP_MEDIA_STORAGE_PATH
          value: "/media"
        resources:
          requests:
            memory: "512Mi"
            cpu: "500m"
          limits:
            memory: "2Gi"
            cpu: "2"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        volumeMounts:
        - name: media-storage
          mountPath: /media
      volumes:
      - name: media-storage
        persistentVolumeClaim:
          claimName: whatsapp-media-pvc
---
apiVersion: v1
kind: Service
metadata:
  name: whatsapp-service
  namespace: production
spec:
  type: ClusterIP
  selector:
    app: whatsapp-service
  ports:
  - port: 80
    targetPort: 8080
    protocol: TCP
    name: http
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: whatsapp-media-pvc
  namespace: production
spec:
  accessModes:
    - ReadWriteMany  # Multiple pods can access
  storageClassName: nfs-storage  # Or your storage class
  resources:
    requests:
      storage: 50Gi
```

#### HorizontalPodAutoscaler

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: whatsapp-service-hpa
  namespace: production
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: whatsapp-service
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

### PostgreSQL High Availability

```yaml
# Using a managed PostgreSQL service (recommended)
# - AWS RDS PostgreSQL
# - Google Cloud SQL
# - Azure Database for PostgreSQL

# Or self-hosted with replication:
apiVersion: postgresql.cnpg.io/v1
kind: Cluster
metadata:
  name: whatsapp-postgres
  namespace: production
spec:
  instances: 3
  storage:
    size: 100Gi
    storageClass: fast-ssd
  postgresql:
    parameters:
      max_connections: "200"
      shared_buffers: "256MB"
  backup:
    barmanObjectStore:
      destinationPath: s3://backups/postgres
      s3Credentials:
        accessKeyId:
          name: s3-credentials
          key: ACCESS_KEY_ID
        secretAccessKey:
          name: s3-credentials
          key: SECRET_ACCESS_KEY
```

### Scaling Characteristics

**Vertical Scaling** (per pod):
- 1 vCPU + 512MB RAM: ~50 concurrent sessions
- 2 vCPU + 2GB RAM: ~200 concurrent sessions
- 4 vCPU + 4GB RAM: ~500 concurrent sessions

**Horizontal Scaling**:
- 3 pods: ~600 concurrent sessions
- 5 pods: ~1000 concurrent sessions
- 10 pods: ~2000 concurrent sessions

### Monitoring & Observability

```yaml
apiVersion: v1
kind: ServiceMonitor
metadata:
  name: whatsapp-service
  namespace: production
spec:
  selector:
    matchLabels:
      app: whatsapp-service
  endpoints:
  - port: http
    path: /metrics
    interval: 30s
```

**Key Metrics to Monitor**:
- Active WhatsApp sessions per pod
- Message send/receive rate
- WebSocket connection health
- PostgreSQL connection pool usage
- API response times
- Error rates by type

## Deployment Considerations

### Resource Requirements

- **Memory**: ~50MB per active WhatsApp session
- **Storage**: 
  - PostgreSQL: ~2-3MB per session (metadata + message history + cryptographic keys)
  - No file storage dependencies
- **CPU**: Minimal, event-driven architecture
- **Network**: Persistent WebSocket connection per session to WhatsApp servers

### Scaling Strategy

1. **Vertical Scaling**: Increase resources for more concurrent sessions
2. **Session Limits**: Configurable max sessions per user
3. **Database Optimization**: Indexes on frequently queried fields
4. **Message Cleanup**: Periodic cleanup of old messages based on retention policy

### Monitoring

1. **Metrics**:
   - Active sessions count
   - Messages sent/received per minute
   - Session connection/disconnection events
   - Webhook delivery success rate
   - Error rates by type

2. **Logging**:
   - Session lifecycle events
   - Message send/receive events
   - Connection errors
   - Webhook delivery attempts

### Backup & Recovery

1. **PostgreSQL**: Single source of truth for all data (sessions, messages, cryptographic keys)
2. **Standard PostgreSQL Backup**: Use pg_dump, WAL archiving, or managed backup solutions
3. **Session Recovery**: Automatic reconnection on application restart using stored credentials
4. **No File Dependencies**: All data in PostgreSQL enables easy disaster recovery

## PostgreSQL vs SQLite: Why This Matters

### Comparison Table

| Feature | SQLite (Default) | PostgreSQL (Our Implementation) |
|---------|------------------|----------------------------------|
| **Kubernetes Scaling** | ❌ Single pod only | ✅ Unlimited horizontal scaling |
| **High Availability** | ❌ Single point of failure | ✅ PostgreSQL clustering |
| **Concurrent Access** | ⚠️ Limited (file locks) | ✅ Excellent (MVCC) |
| **Backup/Restore** | ⚠️ File-based | ✅ Standard PostgreSQL tools |
| **Cloud Native** | ❌ Requires persistent volumes | ✅ Fully stateless pods |
| **Data Consistency** | ⚠️ Risk with multiple pods | ✅ ACID guarantees |
| **Monitoring** | ⚠️ Limited | ✅ Rich PostgreSQL metrics |
| **Disaster Recovery** | ⚠️ Complex | ✅ Standard procedures |
| **Development Effort** | ✅ Zero (built-in) | ⚠️ Custom implementation |

### Production Benefits

1. **True Horizontal Scaling**: Add pods during peak hours, remove during off-hours
2. **Zero Downtime Deployments**: Rolling updates without service interruption
3. **Geographic Distribution**: Deploy across multiple regions with PostgreSQL replication
4. **Simplified Operations**: One database system to manage (PostgreSQL)
5. **Better Observability**: Standard PostgreSQL monitoring tools work out of the box

## Migration Path

### Phase 1: Core Infrastructure & Custom Store
- Database schema creation (including whatsmeow tables)
- Domain entities and repositories
- **Custom PostgreSQL store implementation**
- Basic service layer

### Phase 2: WhatsApp Integration
- whatsmeow client wrapper with PostgreSQL store
- Session management
- QR code generation
- Device registration flow

### Phase 3: Messaging
- Send text message functionality
- Send media message functionality
- Receive message handling
- Message persistence

### Phase 4: Advanced Features
- Webhook integration
- Message history API
- Session monitoring
- Auto-reconnection logic

### Phase 5: Production Readiness
- Comprehensive testing (unit, integration, E2E)
- Performance optimization
- Kubernetes deployment manifests
- Documentation
- Monitoring and alerting setup

## API Examples

### Create Session

```bash
POST /api/v1/whatsapp/sessions
Authorization: Bearer {jwt_token}
Content-Type: application/json

{
  "session_name": "My WhatsApp",
  "webhook_url": "https://example.com/webhook",
  "webhook_secret": "secret123"
}

Response:
{
  "id": "uuid",
  "session_name": "My WhatsApp",
  "status": "disconnected",
  "created_at": "2025-01-10T10:00:00Z"
}
```

### Get QR Code

```bash
GET /api/v1/whatsapp/sessions/{id}/qr
Authorization: Bearer {jwt_token}

Response:
{
  "qr_code": "data:image/png;base64,iVBORw0KGgoAAAANS...",
  "expires_at": "2025-01-10T10:01:00Z"
}
```

### Send Text Message

```bash
POST /api/v1/whatsapp/sessions/{id}/messages
Authorization: Bearer {jwt_token}
Content-Type: application/json

{
  "to": "+1234567890",
  "content": "Hello from API!"
}

Response:
{
  "id": "uuid",
  "message_id": "whatsapp_msg_id",
  "status": "sent",
  "timestamp": "2025-01-10T10:00:00Z"
}
```

### Send Media Message

```bash
POST /api/v1/whatsapp/sessions/{id}/media
Authorization: Bearer {jwt_token}
Content-Type: multipart/form-data

Form fields:
- to: "+1234567890"
- type: "image" (image|video|audio|document)
- file: [binary file data]
- caption: "Check this out!" (optional)
- filename: "document.pdf" (required for documents)

Response:
{
  "id": "uuid",
  "message_id": "whatsapp_msg_id",
  "status": "sent",
  "message_type": "image",
  "timestamp": "2025-01-10T10:00:00Z"
}
```

### Webhook Payload (Incoming Message)

```json
POST {webhook_url}
X-WhatsApp-Signature: hmac_signature
Content-Type: application/json

{
  "session_id": "uuid",
  "message_id": "whatsapp_msg_id",
  "from": "+1234567890",
  "to": "+0987654321",
  "content": "Hello!",
  "timestamp": "2025-01-10T10:00:00Z",
  "message_type": "text"
}
```

## Custom PostgreSQL Store Implementation

### Store Interface

The custom PostgreSQL store must implement whatsmeow's `store.Store` interface:

```go
package whatsapp

import (
    "database/sql"
    "go.mau.fi/whatsmeow/store"
    "go.mau.fi/whatsmeow/types"
)

type PostgresStore struct {
    db        *sql.DB
    sessionID string
    log       logger.Logger
}

// Core interface methods to implement:

// Device management
func (s *PostgresStore) GetDevice() (*store.Device, error)
func (s *PostgresStore) PutDevice(device *store.Device) error

// Identity keys (for end-to-end encryption)
func (s *PostgresStore) GetIdentity(address string) (*store.Identity, error)
func (s *PostgresStore) PutIdentity(address string, identity *store.Identity) error

// Pre-keys (for initial key exchange)
func (s *PostgresStore) GetPreKeys(count int) ([]*store.PreKey, error)
func (s *PostgresStore) PutPreKey(preKey *store.PreKey) error
func (s *PostgresStore) RemovePreKey(keyID uint32) error

// Session keys (for ongoing conversations)
func (s *PostgresStore) GetSession(address string) (*store.Session, error)
func (s *PostgresStore) PutSession(address string, session *store.Session) error
func (s *PostgresStore) HasSession(address string) (bool, error)

// Sender keys (for group encryption)
func (s *PostgresStore) GetSenderKey(group, sender string) (*store.SenderKey, error)
func (s *PostgresStore) PutSenderKey(group, sender string, key *store.SenderKey) error

// App state sync (for multi-device sync)
func (s *PostgresStore) GetAppStateSyncKey(keyID []byte) (*store.AppStateSyncKey, error)
func (s *PostgresStore) PutAppStateSyncKey(keyID []byte, key *store.AppStateSyncKey) error
```

### Implementation Strategy

1. **Binary Data Storage**: Store cryptographic keys as BYTEA in PostgreSQL
2. **Serialization**: Use Protocol Buffers for complex structures
3. **Indexing**: Proper indexes on lookup fields (JID, address, key_id)
4. **Transactions**: Use transactions for atomic operations
5. **Connection Pooling**: Leverage existing PostgreSQL connection pool

### Example Implementation

```go
func (s *PostgresStore) GetDevice() (*store.Device, error) {
    var device store.Device
    var identityKey, signedPreKey, signedPreKeySig []byte
    
    query := `
        SELECT registration_id, identity_key, signed_pre_key, 
               signed_pre_key_id, signed_pre_key_signature, 
               adv_secret_key, platform, business_name, push_name
        FROM whatsapp_devices 
        WHERE session_id = $1
    `
    
    err := s.db.QueryRow(query, s.sessionID).Scan(
        &device.RegistrationID,
        &identityKey,
        &signedPreKey,
        &device.SignedPreKeyID,
        &signedPreKeySig,
        &device.AdvSecretKey,
        &device.Platform,
        &device.BusinessName,
        &device.PushName,
    )
    
    if err == sql.ErrNoRows {
        return nil, nil // No device yet
    }
    if err != nil {
        return nil, err
    }
    
    // Deserialize binary data
    device.IdentityKey = identityKey
    device.SignedPreKey = signedPreKey
    device.SignedPreKeySignature = signedPreKeySig
    
    return &device, nil
}

func (s *PostgresStore) PutDevice(device *store.Device) error {
    query := `
        INSERT INTO whatsapp_devices (
            session_id, jid, registration_id, identity_key, 
            signed_pre_key, signed_pre_key_id, signed_pre_key_signature,
            adv_secret_key, platform, business_name, push_name
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        ON CONFLICT (session_id) DO UPDATE SET
            registration_id = EXCLUDED.registration_id,
            identity_key = EXCLUDED.identity_key,
            signed_pre_key = EXCLUDED.signed_pre_key,
            signed_pre_key_id = EXCLUDED.signed_pre_key_id,
            signed_pre_key_signature = EXCLUDED.signed_pre_key_signature,
            adv_secret_key = EXCLUDED.adv_secret_key,
            platform = EXCLUDED.platform,
            business_name = EXCLUDED.business_name,
            push_name = EXCLUDED.push_name,
            updated_at = NOW()
    `
    
    _, err := s.db.Exec(query,
        s.sessionID,
        device.JID.String(),
        device.RegistrationID,
        device.IdentityKey,
        device.SignedPreKey,
        device.SignedPreKeyID,
        device.SignedPreKeySignature,
        device.AdvSecretKey,
        device.Platform,
        device.BusinessName,
        device.PushName,
    )
    
    return err
}
```

## Dependencies

### Go Modules

```go
require (
    go.mau.fi/whatsmeow v0.0.0-latest
    go.mau.fi/libsignal v0.1.0
    google.golang.org/protobuf v1.31.0
    github.com/skip2/go-qrcode v0.0.0-20200617195104-da1b6568686e
    github.com/lib/pq v1.10.9  // PostgreSQL driver
)
```

### External Services

- PostgreSQL: Session and message metadata
- SQLite: WhatsApp authentication data (per session)
- Redis: Optional caching for session status

## Future Enhancements

1. **Group Messaging**: Send/receive group messages
3. **Message Templates**: Predefined message templates
4. **Bulk Messaging**: Send messages to multiple recipients
5. **Analytics Dashboard**: Message statistics and insights
6. **Message Scheduling**: Schedule messages for future delivery
7. **Contact Management**: Store and manage WhatsApp contacts
8. **Message Search**: Full-text search across message history
