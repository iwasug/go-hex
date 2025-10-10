package whatsapp

import (
	"context"
	"sync"
	"time"

	"golang-chi-hexagonal-base/internal/domain/entity"
	"golang-chi-hexagonal-base/internal/domain/repository"
	"golang-chi-hexagonal-base/internal/pkg/logger"
)

// EventType represents the type of WhatsApp event
type EventType string

const (
	EventTypeConnected    EventType = "connected"
	EventTypeDisconnected EventType = "disconnected"
	EventTypeQRCode       EventType = "qr_code"
	EventTypeMessage      EventType = "message"
	EventTypePairSuccess  EventType = "pair_success"
)

// Event represents a WhatsApp event
type Event struct {
	Type      EventType
	SessionID string
	UserID    string
	Data      interface{}
	Timestamp time.Time
}

// ConnectedEventData contains data for connected events
type ConnectedEventData struct {
	Phone string
}

// DisconnectedEventData contains data for disconnected events
type DisconnectedEventData struct {
	Reason string
}

// QRCodeEventData contains data for QR code events
type QRCodeEventData struct {
	QRCode string
	Expiry time.Time
}

// MessageEventData contains data for message events
type MessageEventData struct {
	Message IncomingMessage
}

// PairSuccessEventData contains data for pair success events
type PairSuccessEventData struct {
	Phone string
}

// EventBus manages event publishing and subscription
type EventBus struct {
	sessionRepo repository.WhatsAppSessionRepository
	messageRepo repository.WhatsAppMessageRepository
	logger      logger.LoggerInterface
	mu          sync.RWMutex
	subscribers map[EventType][]EventSubscriber
}

// EventSubscriber is a function that handles events
type EventSubscriber func(ctx context.Context, event Event) error

// NewEventBus creates a new EventBus instance
func NewEventBus(
	sessionRepo repository.WhatsAppSessionRepository,
	messageRepo repository.WhatsAppMessageRepository,
	log logger.LoggerInterface,
) *EventBus {
	return &EventBus{
		sessionRepo: sessionRepo,
		messageRepo: messageRepo,
		logger:      log.WithFields(logger.Fields{"component": "event_bus"}),
		subscribers: make(map[EventType][]EventSubscriber),
	}
}

// Subscribe adds a subscriber for a specific event type
func (eb *EventBus) Subscribe(eventType EventType, subscriber EventSubscriber) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.subscribers[eventType] = append(eb.subscribers[eventType], subscriber)
	eb.logger.Debug("subscriber added", logger.Fields{
		"event_type": eventType,
	})
}

// Publish publishes an event to all subscribers
func (eb *EventBus) Publish(ctx context.Context, event Event) {
	eb.mu.RLock()
	subscribers := eb.subscribers[event.Type]
	eb.mu.RUnlock()

	eb.logger.Debug("publishing event", logger.Fields{
		"event_type": event.Type,
		"session_id": event.SessionID,
		"user_id":    event.UserID,
	})

	// Call subscribers asynchronously
	for _, subscriber := range subscribers {
		go func(sub EventSubscriber) {
			if err := sub(ctx, event); err != nil {
				eb.logger.Error("subscriber error", logger.Fields{
					"event_type": event.Type,
					"session_id": event.SessionID,
					"error":      err.Error(),
				})
			}
		}(subscriber)
	}
}

// HandleConnectedEvent handles connected events
func (eb *EventBus) HandleConnectedEvent(ctx context.Context, sessionID, userID, phone string) error {
	eb.logger.Info("handling connected event", logger.Fields{
		"session_id": sessionID,
		"user_id":    userID,
		"phone":      phone,
	})

	// Update session status to connected
	if err := eb.sessionRepo.UpdateStatus(ctx, sessionID, entity.SessionStatusConnected); err != nil {
		eb.logger.Error("failed to update session status to connected", logger.Fields{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return err
	}

	// Reset reconnection attempts since we're now connected
	if err := eb.sessionRepo.ResetReconnectAttempts(ctx, sessionID); err != nil {
		eb.logger.Error("failed to reset reconnect attempts", logger.Fields{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		// Don't return error, this is not critical
	}

	// Update phone number if provided
	if phone != "" {
		if err := eb.sessionRepo.UpdatePhoneNumber(ctx, sessionID, phone); err != nil {
			eb.logger.Error("failed to update phone number", logger.Fields{
				"session_id": sessionID,
				"phone":      phone,
				"error":      err.Error(),
			})
			return err
		}
	}

	// Clear QR code since we're now connected
	if err := eb.sessionRepo.ClearQRCode(ctx, sessionID); err != nil {
		eb.logger.Error("failed to clear QR code", logger.Fields{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		// Don't return error, this is not critical
	}

	// Update last activity
	if err := eb.sessionRepo.UpdateLastActivity(ctx, sessionID); err != nil {
		eb.logger.Error("failed to update last activity", logger.Fields{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		// Don't return error, this is not critical
	}

	// Publish event to subscribers
	event := Event{
		Type:      EventTypeConnected,
		SessionID: sessionID,
		UserID:    userID,
		Data: ConnectedEventData{
			Phone: phone,
		},
		Timestamp: time.Now(),
	}
	eb.Publish(ctx, event)

	eb.logger.Info("connected event handled successfully", logger.Fields{
		"session_id": sessionID,
		"phone":      phone,
	})

	return nil
}

// HandleDisconnectedEvent handles disconnected events with reconnection logic
func (eb *EventBus) HandleDisconnectedEvent(ctx context.Context, sessionID, userID string, clientManager ClientManagerInterface) error {
	eb.logger.Warn("handling disconnected event", logger.Fields{
		"session_id": sessionID,
		"user_id":    userID,
	})

	// Update session status to disconnected
	if err := eb.sessionRepo.UpdateStatus(ctx, sessionID, entity.SessionStatusDisconnected); err != nil {
		eb.logger.Error("failed to update session status to disconnected", logger.Fields{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return err
	}

	// Update last activity
	if err := eb.sessionRepo.UpdateLastActivity(ctx, sessionID); err != nil {
		eb.logger.Error("failed to update last activity", logger.Fields{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		// Don't return error, this is not critical
	}

	// Publish event to subscribers
	event := Event{
		Type:      EventTypeDisconnected,
		SessionID: sessionID,
		UserID:    userID,
		Data: DisconnectedEventData{
			Reason: "connection_lost",
		},
		Timestamp: time.Now(),
	}
	eb.Publish(ctx, event)

	// Attempt automatic reconnection
	go eb.attemptReconnection(context.Background(), sessionID, userID, clientManager)

	eb.logger.Info("disconnected event handled successfully", logger.Fields{
		"session_id": sessionID,
	})

	return nil
}

// attemptReconnection attempts to reconnect a disconnected session with exponential backoff
func (eb *EventBus) attemptReconnection(ctx context.Context, sessionID, userID string, clientManager ClientManagerInterface) {
	if clientManager == nil {
		eb.logger.Error("client manager is nil, cannot attempt reconnection", logger.Fields{
			"session_id": sessionID,
		})
		return
	}

	// Get session to check reconnection attempts
	session, err := eb.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		eb.logger.Error("failed to get session for reconnection", logger.Fields{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return
	}

	// Configuration for reconnection
	const (
		maxAttempts       = 5
		initialDelay      = 5 * time.Second
		maxDelay          = 5 * time.Minute
		backoffMultiplier = 2
	)

	// Check if we should attempt reconnection
	if !session.ShouldAttemptReconnect(maxAttempts) {
		eb.logger.Warn("max reconnection attempts reached, marking session as failed", logger.Fields{
			"session_id":         sessionID,
			"reconnect_attempts": session.ReconnectAttempts,
			"max_attempts":       maxAttempts,
		})

		// Mark session as failed
		if err := eb.sessionRepo.UpdateStatus(ctx, sessionID, entity.SessionStatusFailed); err != nil {
			eb.logger.Error("failed to update session status to failed", logger.Fields{
				"session_id": sessionID,
				"error":      err.Error(),
			})
		}
		return
	}

	eb.logger.Info("starting reconnection attempt", logger.Fields{
		"session_id":         sessionID,
		"current_attempts":   session.ReconnectAttempts,
		"max_attempts":       maxAttempts,
	})

	// Calculate delay with exponential backoff
	delay := initialDelay
	for i := 0; i < session.ReconnectAttempts; i++ {
		delay = delay * backoffMultiplier
		if delay > maxDelay {
			delay = maxDelay
			break
		}
	}

	eb.logger.Info("waiting before reconnection attempt", logger.Fields{
		"session_id": sessionID,
		"delay":      delay,
		"attempt":    session.ReconnectAttempts + 1,
	})

	// Wait before attempting reconnection
	time.Sleep(delay)

	// Increment reconnection attempts
	if err := eb.sessionRepo.IncrementReconnectAttempts(ctx, sessionID); err != nil {
		eb.logger.Error("failed to increment reconnect attempts", logger.Fields{
			"session_id": sessionID,
			"error":      err.Error(),
		})
	}

	// Update session status to connecting
	if err := eb.sessionRepo.UpdateStatus(ctx, sessionID, entity.SessionStatusConnecting); err != nil {
		eb.logger.Error("failed to update session status to connecting", logger.Fields{
			"session_id": sessionID,
			"error":      err.Error(),
		})
	}

	// Get the client
	client, err := clientManager.GetClient(sessionID)
	if err != nil {
		eb.logger.Error("failed to get client for reconnection", logger.Fields{
			"session_id": sessionID,
			"error":      err.Error(),
		})

		// Update session status to disconnected
		if err := eb.sessionRepo.UpdateStatus(ctx, sessionID, entity.SessionStatusDisconnected); err != nil {
			eb.logger.Error("failed to update session status to disconnected", logger.Fields{
				"session_id": sessionID,
				"error":      err.Error(),
			})
		}

		// Schedule next reconnection attempt
		go eb.attemptReconnection(context.Background(), sessionID, userID, clientManager)
		return
	}

	// Check if already connected (might have reconnected automatically)
	if client.IsConnected() {
		eb.logger.Info("client already connected, resetting reconnection attempts", logger.Fields{
			"session_id": sessionID,
		})

		// Reset reconnection attempts
		if err := eb.sessionRepo.ResetReconnectAttempts(ctx, sessionID); err != nil {
			eb.logger.Error("failed to reset reconnect attempts", logger.Fields{
				"session_id": sessionID,
				"error":      err.Error(),
			})
		}

		// Update session status to connected
		if err := eb.sessionRepo.UpdateStatus(ctx, sessionID, entity.SessionStatusConnected); err != nil {
			eb.logger.Error("failed to update session status to connected", logger.Fields{
				"session_id": sessionID,
				"error":      err.Error(),
			})
		}
		return
	}

	// Attempt to connect
	if err := client.Connect(); err != nil {
		eb.logger.Error("reconnection attempt failed", logger.Fields{
			"session_id": sessionID,
			"attempt":    session.ReconnectAttempts + 1,
			"error":      err.Error(),
		})

		// Update session status to disconnected
		if err := eb.sessionRepo.UpdateStatus(ctx, sessionID, entity.SessionStatusDisconnected); err != nil {
			eb.logger.Error("failed to update session status to disconnected", logger.Fields{
				"session_id": sessionID,
				"error":      err.Error(),
			})
		}

		// Schedule next reconnection attempt
		go eb.attemptReconnection(context.Background(), sessionID, userID, clientManager)
		return
	}

	// Reconnection successful
	eb.logger.Info("reconnection successful", logger.Fields{
		"session_id": sessionID,
		"attempt":    session.ReconnectAttempts + 1,
	})

	// Reset reconnection attempts
	if err := eb.sessionRepo.ResetReconnectAttempts(ctx, sessionID); err != nil {
		eb.logger.Error("failed to reset reconnect attempts", logger.Fields{
			"session_id": sessionID,
			"error":      err.Error(),
		})
	}

	// Update session status to connected
	if err := eb.sessionRepo.UpdateStatus(ctx, sessionID, entity.SessionStatusConnected); err != nil {
		eb.logger.Error("failed to update session status after reconnection", logger.Fields{
			"session_id": sessionID,
			"error":      err.Error(),
		})
	}
}

// HandleQRCodeEvent handles QR code events
func (eb *EventBus) HandleQRCodeEvent(ctx context.Context, sessionID, userID, qrCode string) error {
	eb.logger.Info("handling QR code event", logger.Fields{
		"session_id": sessionID,
		"user_id":    userID,
	})

	// Set QR code expiry (60 seconds from now)
	expiry := time.Now().Add(60 * time.Second)

	// Update QR code in database
	if err := eb.sessionRepo.UpdateQRCode(ctx, sessionID, qrCode, expiry); err != nil {
		eb.logger.Error("failed to update QR code", logger.Fields{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return err
	}

	// Update last activity
	if err := eb.sessionRepo.UpdateLastActivity(ctx, sessionID); err != nil {
		eb.logger.Error("failed to update last activity", logger.Fields{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		// Don't return error, this is not critical
	}

	// Publish event to subscribers
	event := Event{
		Type:      EventTypeQRCode,
		SessionID: sessionID,
		UserID:    userID,
		Data: QRCodeEventData{
			QRCode: qrCode,
			Expiry: expiry,
		},
		Timestamp: time.Now(),
	}
	eb.Publish(ctx, event)

	eb.logger.Info("QR code event handled successfully", logger.Fields{
		"session_id": sessionID,
		"expiry":     expiry,
	})

	return nil
}

// HandleMessageEvent handles incoming message events
func (eb *EventBus) HandleMessageEvent(ctx context.Context, sessionID, userID string, message IncomingMessage) error {
	eb.logger.Info("handling message event", logger.Fields{
		"session_id": sessionID,
		"user_id":    userID,
		"from":       message.From,
		"message_id": message.MessageID,
		"type":       message.MessageType,
	})

	// Create message entity
	msg := entity.NewInboundMessage(
		sessionID,
		message.MessageID,
		message.From,
		message.Content,
		message.MessageType,
		message.Timestamp,
	)

	// Set the 'To' field
	msg.To = message.To

	// Store message in database
	if err := eb.messageRepo.Create(ctx, msg); err != nil {
		eb.logger.Error("failed to store incoming message", logger.Fields{
			"session_id": sessionID,
			"message_id": message.MessageID,
			"error":      err.Error(),
		})
		return err
	}

	// Update session last activity
	if err := eb.sessionRepo.UpdateLastActivity(ctx, sessionID); err != nil {
		eb.logger.Error("failed to update last activity", logger.Fields{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		// Don't return error, this is not critical
	}

	// Publish event to subscribers (for webhook delivery, etc.)
	event := Event{
		Type:      EventTypeMessage,
		SessionID: sessionID,
		UserID:    userID,
		Data: MessageEventData{
			Message: message,
		},
		Timestamp: time.Now(),
	}
	eb.Publish(ctx, event)

	eb.logger.Info("message event handled successfully", logger.Fields{
		"session_id": sessionID,
		"message_id": message.MessageID,
		"db_id":      msg.ID,
	})

	return nil
}

// HandlePairSuccessEvent handles successful pairing events
func (eb *EventBus) HandlePairSuccessEvent(ctx context.Context, sessionID, userID, phone string) error {
	eb.logger.Info("handling pair success event", logger.Fields{
		"session_id": sessionID,
		"user_id":    userID,
		"phone":      phone,
	})

	// Update session status to connected
	if err := eb.sessionRepo.UpdateStatus(ctx, sessionID, entity.SessionStatusConnected); err != nil {
		eb.logger.Error("failed to update session status to connected", logger.Fields{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return err
	}

	// Reset reconnection attempts since pairing is successful
	if err := eb.sessionRepo.ResetReconnectAttempts(ctx, sessionID); err != nil {
		eb.logger.Error("failed to reset reconnect attempts", logger.Fields{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		// Don't return error, this is not critical
	}

	// Update phone number
	if err := eb.sessionRepo.UpdatePhoneNumber(ctx, sessionID, phone); err != nil {
		eb.logger.Error("failed to update phone number", logger.Fields{
			"session_id": sessionID,
			"phone":      phone,
			"error":      err.Error(),
		})
		return err
	}

	// Clear QR code since pairing is successful
	if err := eb.sessionRepo.ClearQRCode(ctx, sessionID); err != nil {
		eb.logger.Error("failed to clear QR code", logger.Fields{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		// Don't return error, this is not critical
	}

	// Update last activity
	if err := eb.sessionRepo.UpdateLastActivity(ctx, sessionID); err != nil {
		eb.logger.Error("failed to update last activity", logger.Fields{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		// Don't return error, this is not critical
	}

	// Publish event to subscribers
	event := Event{
		Type:      EventTypePairSuccess,
		SessionID: sessionID,
		UserID:    userID,
		Data: PairSuccessEventData{
			Phone: phone,
		},
		Timestamp: time.Now(),
	}
	eb.Publish(ctx, event)

	eb.logger.Info("pair success event handled successfully", logger.Fields{
		"session_id": sessionID,
		"phone":      phone,
	})

	return nil
}
