package whatsapp

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"golang-chi-hexagonal-base/internal/domain/entity"
	"golang-chi-hexagonal-base/internal/domain/repository"
	"golang-chi-hexagonal-base/internal/pkg/logger"

	"go.mau.fi/whatsmeow"
)

// ClientManager manages the lifecycle of WhatsApp clients
type ClientManager struct {
	clients       map[string]*Client // sessionID -> Client
	mu            sync.RWMutex
	db            *sql.DB
	sessionRepo   repository.WhatsAppSessionRepository
	messageRepo   repository.WhatsAppMessageRepository
	logger        logger.LoggerInterface
	eventBus      *EventBus
	eventHandlers EventHandlers
}

// NewClientManager creates a new ClientManager instance
func NewClientManager(
	db *sql.DB,
	sessionRepo repository.WhatsAppSessionRepository,
	messageRepo repository.WhatsAppMessageRepository,
	log logger.LoggerInterface,
) *ClientManager {
	// Create event bus
	eventBus := NewEventBus(sessionRepo, messageRepo, log)

	// Create event handlers that use the event bus
	eventHandlers := EventHandlers{
		OnConnected: func(sessionID, userID, phone string) {
			ctx := context.Background()
			if err := eventBus.HandleConnectedEvent(ctx, sessionID, userID, phone); err != nil {
				log.Error("failed to handle connected event", logger.Fields{
					"session_id": sessionID,
					"error":      err.Error(),
				})
			}
		},
		OnDisconnected: func(sessionID, userID string) {
			// We'll pass the client manager reference later
			ctx := context.Background()
			if err := eventBus.HandleDisconnectedEvent(ctx, sessionID, userID, nil); err != nil {
				log.Error("failed to handle disconnected event", logger.Fields{
					"session_id": sessionID,
					"error":      err.Error(),
				})
			}
		},
		OnQRCode: func(sessionID, userID, qrCode string) {
			ctx := context.Background()
			if err := eventBus.HandleQRCodeEvent(ctx, sessionID, userID, qrCode); err != nil {
				log.Error("failed to handle QR code event", logger.Fields{
					"session_id": sessionID,
					"error":      err.Error(),
				})
			}
		},
		OnMessage: func(sessionID, userID string, message IncomingMessage) {
			ctx := context.Background()
			if err := eventBus.HandleMessageEvent(ctx, sessionID, userID, message); err != nil {
				log.Error("failed to handle message event", logger.Fields{
					"session_id": sessionID,
					"error":      err.Error(),
				})
			}
		},
		OnPairSuccess: func(sessionID, userID, phone string) {
			ctx := context.Background()
			if err := eventBus.HandlePairSuccessEvent(ctx, sessionID, userID, phone); err != nil {
				log.Error("failed to handle pair success event", logger.Fields{
					"session_id": sessionID,
					"error":      err.Error(),
				})
			}
		},
	}

	cm := &ClientManager{
		clients:       make(map[string]*Client),
		db:            db,
		sessionRepo:   sessionRepo,
		messageRepo:   messageRepo,
		logger:        log.WithFields(logger.Fields{"component": "client_manager"}),
		eventBus:      eventBus,
		eventHandlers: eventHandlers,
	}

	// Update the OnDisconnected handler to pass the client manager reference
	cm.eventHandlers.OnDisconnected = func(sessionID, userID string) {
		ctx := context.Background()
		if err := eventBus.HandleDisconnectedEvent(ctx, sessionID, userID, cm); err != nil {
			log.Error("failed to handle disconnected event", logger.Fields{
				"session_id": sessionID,
				"error":      err.Error(),
			})
		}
	}

	return cm
}

// CreateClient creates a new WhatsApp client for a session
func (cm *ClientManager) CreateClient(ctx context.Context, sessionID, userID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Check if client already exists
	if _, exists := cm.clients[sessionID]; exists {
		cm.logger.Warn("client already exists for session", logger.Fields{
			"session_id": sessionID,
			"user_id":    userID,
		})
		return fmt.Errorf("client already exists for session %s", sessionID)
	}

	// Create PostgresStore for this session
	pgStore := NewPostgresStore(cm.db, sessionID)

	// Create a logger adapter for whatsmeow
	waLogger := NewLoggerAdapter(cm.logger.WithFields(logger.Fields{
		"session_id": sessionID,
		"component":  "whatsmeow",
	}))

	// Create a container that holds the device with our store implementations
	container := NewStoreContainer(pgStore, waLogger)

	// Get or create device from container
	device, err := container.GetFirstDevice()
	if err != nil {
		cm.logger.Error("failed to get device", logger.Fields{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return fmt.Errorf("failed to get device: %w", err)
	}

	// Create whatsmeow client with device and logger
	waClient := whatsmeow.NewClient(device, waLogger)

	// Create our client wrapper
	client := &Client{
		sessionID:     sessionID,
		userID:        userID,
		waClient:      waClient,
		store:         pgStore,
		logger:        cm.logger.WithFields(logger.Fields{"session_id": sessionID}),
		eventHandlers: cm.eventHandlers,
		connected:     false,
	}

	// Register event handlers
	client.RegisterEventHandlers()

	// Store the client
	cm.clients[sessionID] = client

	cm.logger.Info("client created successfully", logger.Fields{
		"session_id": sessionID,
		"user_id":    userID,
	})

	return nil
}

// GetClient retrieves a WhatsApp client by session ID (thread-safe)
func (cm *ClientManager) GetClient(sessionID string) (*Client, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	client, exists := cm.clients[sessionID]
	if !exists {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	return client, nil
}

// DeleteClient removes a WhatsApp client and performs cleanup
func (cm *ClientManager) DeleteClient(ctx context.Context, sessionID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	client, exists := cm.clients[sessionID]
	if !exists {
		cm.logger.Warn("client not found for deletion", logger.Fields{
			"session_id": sessionID,
		})
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	// Disconnect the client if connected
	if client.IsConnected() {
		if err := client.Disconnect(); err != nil {
			cm.logger.Error("failed to disconnect client during deletion", logger.Fields{
				"session_id": sessionID,
				"error":      err.Error(),
			})
			// Continue with deletion even if disconnect fails
		}
	}

	// Remove from map
	delete(cm.clients, sessionID)

	cm.logger.Info("client deleted successfully", logger.Fields{
		"session_id": sessionID,
	})

	return nil
}

// RestoreClients restores all connected sessions on application startup
func (cm *ClientManager) RestoreClients(ctx context.Context) error {
	cm.logger.Info("restoring WhatsApp clients from database")

	// Get all connected sessions from database
	sessions, err := cm.sessionRepo.FindConnectedSessions(ctx)
	if err != nil {
		cm.logger.Error("failed to fetch connected sessions", logger.Fields{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to fetch connected sessions: %w", err)
	}

	cm.logger.Info("found sessions to restore", logger.Fields{
		"count": len(sessions),
	})

	// Restore each session
	var restoredCount int
	var failedCount int

	for _, session := range sessions {
		if err := cm.restoreSession(ctx, session); err != nil {
			cm.logger.Error("failed to restore session", logger.Fields{
				"session_id": session.ID,
				"user_id":    session.UserID,
				"error":      err.Error(),
			})
			failedCount++

			// Update session status to disconnected
			if err := cm.sessionRepo.UpdateStatus(ctx, session.ID, entity.SessionStatusDisconnected); err != nil {
				cm.logger.Error("failed to update session status", logger.Fields{
					"session_id": session.ID,
					"error":      err.Error(),
				})
			}
			continue
		}

		restoredCount++
	}

	cm.logger.Info("client restoration completed", logger.Fields{
		"restored": restoredCount,
		"failed":   failedCount,
		"total":    len(sessions),
	})

	return nil
}

// restoreSession restores a single session
func (cm *ClientManager) restoreSession(ctx context.Context, session *entity.WhatsAppSession) error {
	// Create the client
	if err := cm.CreateClient(ctx, session.ID, session.UserID); err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// Get the client
	client, err := cm.GetClient(session.ID)
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}

	// Connect the client
	if err := client.Connect(); err != nil {
		// Clean up the client if connection fails
		_ = cm.DeleteClient(ctx, session.ID)
		return fmt.Errorf("failed to connect client: %w", err)
	}

	cm.logger.Info("session restored successfully", logger.Fields{
		"session_id":   session.ID,
		"user_id":      session.UserID,
		"phone_number": session.PhoneNumber,
	})

	return nil
}

// GetAllClients returns all active clients (for monitoring/debugging)
func (cm *ClientManager) GetAllClients() map[string]*Client {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// Return a copy to prevent external modification
	clients := make(map[string]*Client, len(cm.clients))
	for k, v := range cm.clients {
		clients[k] = v
	}

	return clients
}

// GetClientCount returns the number of active clients
func (cm *ClientManager) GetClientCount() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return len(cm.clients)
}

// GetEventBus returns the event bus instance
func (cm *ClientManager) GetEventBus() *EventBus {
	return cm.eventBus
}

// SubscribeToEvent subscribes to a specific event type
func (cm *ClientManager) SubscribeToEvent(eventType EventType, subscriber EventSubscriber) {
	cm.eventBus.Subscribe(eventType, subscriber)
}

// Shutdown gracefully shuts down all clients
func (cm *ClientManager) Shutdown(ctx context.Context) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.logger.Info("shutting down all WhatsApp clients", logger.Fields{
		"count": len(cm.clients),
	})

	var errors []error

	for sessionID, client := range cm.clients {
		if client.IsConnected() {
			if err := client.Disconnect(); err != nil {
				cm.logger.Error("failed to disconnect client during shutdown", logger.Fields{
					"session_id": sessionID,
					"error":      err.Error(),
				})
				errors = append(errors, err)
			}
		}
	}

	// Clear all clients
	cm.clients = make(map[string]*Client)

	if len(errors) > 0 {
		return fmt.Errorf("encountered %d errors during shutdown", len(errors))
	}

	cm.logger.Info("all WhatsApp clients shut down successfully")
	return nil
}
