# Implementation Plan

- [x] 1. Database schema and migrations





  - Create PostgreSQL tables for WhatsApp sessions, messages, and cryptographic keys
  - Write migration files for whatsapp_sessions, whatsapp_messages, whatsapp_devices, whatsapp_identity_keys, whatsapp_pre_keys, whatsapp_session_keys, whatsapp_sender_keys tables
  - Add WhatsApp permissions to permissions table (whatsapp:read, whatsapp:write, whatsapp:delete)
  - Create indexes for performance optimization
  - _Requirements: 1.1, 1.2, 1.3, 6.1, 7.2, 7.3, 8.3, 10.1_

- [x] 2. Domain layer - Entities





  - Create WhatsAppSession entity with status management methods
  - Create WhatsAppMessage entity with direction and status enums
  - Add entity methods for session lifecycle (connect, disconnect, update status)
  - Add entity methods for message creation and status updates
  - _Requirements: 1.1, 3.1, 3.3, 4.1, 5.1, 8.3_

- [x] 3. Domain layer - Repository interfaces





  - Create WhatsAppSessionRepository interface with CRUD operations
  - Create WhatsAppMessageRepository interface with query methods
  - Define ListSessionsParams and ListMessagesParams structs
  - Add methods for session status updates and QR code management
  - _Requirements: 1.1, 1.3, 3.1, 4.1, 5.1, 8.1, 8.2_

- [x] 4. Custom PostgreSQL store for whatsmeow





  - Create PostgresStore struct implementing store.Store interface
  - Implement GetDevice() and PutDevice() methods for device management
  - Implement GetIdentity() and PutIdentity() methods for identity keys
  - Implement GetPreKeys(), PutPreKey(), RemovePreKey() for pre-key management
  - Implement GetSession(), PutSession(), HasSession() for session keys
  - Implement GetSenderKey() and PutSenderKey() for group encryption
  - Implement GetAppStateSyncKey() and PutAppStateSyncKey() for multi-device sync
  - Add proper error handling and logging
  - _Requirements: 1.2, 1.6, 6.2, 9.1, 9.2, 10.1_

- [x] 5. PostgreSQL repository implementations








  - Implement WhatsAppSessionRepository with PostgreSQL
  - Implement Create, FindByID, FindByUserID, Update, Delete methods
  - Implement List method with filtering, sorting, and pagination
  - Implement UpdateStatus and UpdateQRCode methods
  - Implement WhatsAppMessageRepository with PostgreSQL
  - Implement message CRUD operations with proper indexing
  - Implement List method with date range filtering and pagination
  - Add UpdateStatus and IncrementWebhookAttempts methods
  - _Requirements: 1.1, 1.3, 3.1, 4.1, 5.1, 8.1, 8.2, 8.3_

- [x] 6. WhatsApp client manager





  - Create ClientManager struct with client lifecycle management
  - Implement CreateClient() method with PostgresStore initialization
  - Implement GetClient() method with thread-safe client retrieval
  - Implement DeleteClient() method with proper cleanup
  - Implement RestoreClients() method for application startup
  - Add connection pool management for whatsmeow clients
  - _Requirements: 1.1, 1.2, 1.6, 5.2, 5.5, 9.5_

- [x] 7. WhatsApp client wrapper





  - Create Client struct wrapping whatsmeow.Client
  - Implement Connect() method with WebSocket connection
  - Implement Disconnect() method with graceful shutdown
  - Implement GetQRCode() method returning base64 PNG
  - Implement IsConnected() and GetConnectedPhone() methods
  - Add RegisterEventHandlers() for event subscription
  - _Requirements: 1.1, 1.6, 2.1, 2.2, 2.3, 5.1, 5.3_

- [x] 8. Event handling system





  - Create EventHandlers struct with callback functions
  - Implement OnConnected handler to update session status
  - Implement OnDisconnected handler with reconnection logic
  - Implement OnQRCode handler to store QR code in database
  - Implement OnMessage handler for incoming messages
  - Add event bus for publishing events to service layer
  - _Requirements: 2.3, 2.4, 2.5, 4.1, 4.2, 5.2, 5.4_

- [x] 9. WhatsApp service - Session management





  - Create WhatsAppService struct with dependencies
  - Implement CreateSession() method with user validation
  - Implement GetSession() method with ownership check
  - Implement ListSessions() method with admin/user filtering
  - Implement DeleteSession() method with cleanup
  - Implement GenerateQRCode() method with expiry handling
  - Add session ownership validation for security
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 2.1, 2.2, 2.6, 6.1, 6.3, 6.4_

- [x] 10. WhatsApp service - Messaging





  - Implement SendTextMessage() method with validation
  - Implement SendMediaMessage() method with file handling
  - Add phone number validation (E.164 format)
  - Add session connection status check before sending
  - Implement message persistence after sending
  - Add error handling for send failures
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 3.6, 3.7, 3.8, 3.9, 3.10, 3.11, 9.1, 9.2_

- [x] 11. WhatsApp service - Message receiving





  - Implement HandleIncomingMessage() method
  - Store received messages in database
  - Extract sender, content, timestamp, and message type
  - Handle text and media messages
  - Add message deduplication logic
  - _Requirements: 4.1, 4.2, 8.3, 8.4, 8.5_

- [x] 12. Webhook integration





  - Create WebhookClient for HTTP delivery
  - Implement SendWebhook() method with retry logic
  - Add exponential backoff for failed deliveries
  - Implement HMAC signature for webhook authentication
  - Add webhook delivery tracking in database
  - Implement retry mechanism (up to 3 attempts)
  - Log failed webhooks after max retries
  - _Requirements: 4.3, 4.4, 4.5, 4.6, 9.5_

- [x] 13. HTTP handlers - Session management





  - Create WhatsAppHandler struct
  - Implement CreateSession handler (POST /api/v1/whatsapp/sessions)
  - Implement ListSessions handler (GET /api/v1/whatsapp/sessions)
  - Implement GetSession handler (GET /api/v1/whatsapp/sessions/:id)
  - Implement DeleteSession handler (DELETE /api/v1/whatsapp/sessions/:id)
  - Implement GetQRCode handler (GET /api/v1/whatsapp/sessions/:id/qr)
  - Add request validation and error responses
  - _Requirements: 1.1, 1.3, 1.4, 2.1, 2.2, 6.3, 7.1, 7.4, 9.1_

- [x] 14. HTTP handlers - Messaging





  - Implement SendTextMessage handler (POST /api/v1/whatsapp/sessions/:id/messages)
  - Implement SendMediaMessage handler (POST /api/v1/whatsapp/sessions/:id/media)
  - Implement GetMessages handler (GET /api/v1/whatsapp/sessions/:id/messages)
  - Add multipart form handling for media uploads
  - Add file size validation for media
  - Add MIME type validation
  - Implement pagination for message history
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 3.6, 3.7, 3.8, 8.1, 8.2, 9.1_

- [x] 15. HTTP handlers - Session operations





  - Implement ReconnectSession handler (POST /api/v1/whatsapp/sessions/:id/reconnect)
  - Add session status endpoint for monitoring
  - Implement proper HTTP status codes for all responses
  - Add structured error responses
  - _Requirements: 5.2, 5.4, 9.1, 9.2, 9.3_

- [x] 16. Middleware and RBAC integration





  - Add whatsapp:read permission check to read endpoints
  - Add whatsapp:write permission check to write endpoints
  - Add whatsapp:delete permission check to delete endpoints
  - Implement session ownership validation middleware
  - Add admin bypass for viewing all sessions
  - Ensure JWT authentication on all endpoints
  - _Requirements: 6.1, 6.3, 6.4, 6.5, 7.1, 7.2, 7.3, 7.4, 7.5, 7.6_

- [x] 17. Router configuration





  - Register WhatsApp routes in Chi router
  - Apply authentication middleware to all routes
  - Apply RBAC middleware with appropriate permissions
  - Group routes under /api/v1/whatsapp prefix
  - Add rate limiting to prevent abuse
  - _Requirements: 7.1, 7.2, 9.3_

- [x] 18. Configuration management





  - Add WhatsAppConfig struct to config package
  - Load WhatsApp configuration from environment variables
  - Add validation for required configuration
  - Set sensible defaults for optional config
  - Add configuration for QR timeout, reconnect interval, media settings
  - _Requirements: 10.1, 10.2, 10.3, 10.4, 10.5, 10.6_

- [x] 19. Media file handling





  - Create media storage directory structure
  - Implement file upload handling for images, videos, audio, documents
  - Add file size validation based on configuration
  - Add MIME type validation
  - Implement file storage with unique naming
  - Add media cleanup for deleted sessions
  - _Requirements: 3.4, 3.5, 3.6, 3.7, 3.8_

- [x] 20. Session reconnection logic





  - Implement automatic reconnection on disconnect
  - Add exponential backoff for reconnection attempts
  - Track reconnection attempts and max limit
  - Update session status during reconnection
  - Mark session as failed after max attempts
  - _Requirements: 5.2, 5.5, 9.5_

- [x] 21. Application startup integration





  - Initialize WhatsApp service in main.go
  - Call RestoreClients() on application startup
  - Add graceful shutdown for WhatsApp clients
  - Close all active connections on shutdown
  - Add health check for WhatsApp service
  - _Requirements: 1.6, 5.6, 9.2_

- [x] 22. Error handling and logging







  - Define WhatsApp-specific error types
  - Add structured logging for all operations
  - Log session lifecycle events
  - Log message send/receive events
  - Log webhook delivery attempts
  - Add error context for debugging
  - _Requirements: 9.1, 9.2, 9.3, 9.4, 9.5, 9.6_

- [ ] 23. Swagger documentation
  - Add Swagger annotations to all WhatsApp endpoints
  - Document request/response schemas
  - Document authentication requirements
  - Document permission requirements
  - Add example requests and responses
  - Generate Swagger documentation
  - _Requirements: 7.1, 7.2, 7.4_

- [ ] 24. Integration testing
  - Write integration tests for session creation flow
  - Write integration tests for QR code generation
  - Write integration tests for message sending
  - Write integration tests for message receiving
  - Write integration tests for webhook delivery
  - Test session reconnection scenarios
  - Test concurrent session handling
  - _Requirements: All requirements_

- [ ] 25. Kubernetes deployment configuration
  - Create Deployment YAML with horizontal scaling
  - Create Service YAML for load balancing
  - Create PersistentVolumeClaim for media storage
  - Create HorizontalPodAutoscaler configuration
  - Add resource requests and limits
  - Add liveness and readiness probes
  - Create ConfigMap for WhatsApp configuration
  - Create Secret for sensitive configuration
  - _Requirements: 10.1, 10.2, 10.3_
