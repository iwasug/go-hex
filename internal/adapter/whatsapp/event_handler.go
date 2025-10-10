package whatsapp

import (
	"fmt"

	"golang-chi-hexagonal-base/internal/pkg/logger"

	"go.mau.fi/whatsmeow/types/events"
)

// EventHandlers contains callback functions for WhatsApp events
type EventHandlers struct {
	OnConnected    func(sessionID, userID, phone string)
	OnDisconnected func(sessionID, userID string)
	OnQRCode       func(sessionID, userID, qrCode string)
	OnMessage      func(sessionID, userID string, message IncomingMessage)
	OnPairSuccess  func(sessionID, userID, phone string)
}

// EventHandler handles events from the WhatsApp client
type EventHandler struct {
	sessionID     string
	userID        string
	logger        logger.LoggerInterface
	eventHandlers EventHandlers
}

// NewEventHandler creates a new EventHandler instance
func NewEventHandler(
	sessionID, userID string,
	log logger.LoggerInterface,
	eventHandlers EventHandlers,
) *EventHandler {
	return &EventHandler{
		sessionID:     sessionID,
		userID:        userID,
		logger:        log.WithFields(logger.Fields{"session_id": sessionID}),
		eventHandlers: eventHandlers,
	}
}

// HandleEvent is the main event handler that dispatches to specific handlers
func (h *EventHandler) HandleEvent(evt interface{}) {
	switch v := evt.(type) {
	case *events.Connected:
		h.handleConnected(v)
	case *events.Disconnected:
		h.handleDisconnected(v)
	case *events.QR:
		h.handleQRCode(v)
	case *events.PairSuccess:
		h.handlePairSuccess(v)
	case *events.Message:
		h.handleMessage(v)
	case *events.Receipt:
		h.handleReceipt(v)
	case *events.LoggedOut:
		h.handleLoggedOut(v)
	default:
		// Log unknown events for debugging
		h.logger.Debug("received unknown event", logger.Fields{
			"event_type": fmt.Sprintf("%T", v),
		})
	}
}

// handleConnected handles the Connected event
func (h *EventHandler) handleConnected(evt *events.Connected) {
	h.logger.Info("WhatsApp connected")

	if h.eventHandlers.OnConnected != nil {
		// Phone number will be available after pairing
		h.eventHandlers.OnConnected(h.sessionID, h.userID, "")
	}
}

// handleDisconnected handles the Disconnected event
func (h *EventHandler) handleDisconnected(evt *events.Disconnected) {
	h.logger.Warn("WhatsApp disconnected")

	if h.eventHandlers.OnDisconnected != nil {
		h.eventHandlers.OnDisconnected(h.sessionID, h.userID)
	}
}

// handleQRCode handles the QR code event
func (h *EventHandler) handleQRCode(evt *events.QR) {
	h.logger.Info("QR code received", logger.Fields{
		"codes_count": len(evt.Codes),
	})

	if h.eventHandlers.OnQRCode != nil && len(evt.Codes) > 0 {
		// Use the first QR code
		h.eventHandlers.OnQRCode(h.sessionID, h.userID, evt.Codes[0])
	}
}

// handlePairSuccess handles successful pairing
func (h *EventHandler) handlePairSuccess(evt *events.PairSuccess) {
	phone := evt.ID.User

	h.logger.Info("pairing successful", logger.Fields{
		"phone": phone,
	})

	if h.eventHandlers.OnPairSuccess != nil {
		h.eventHandlers.OnPairSuccess(h.sessionID, h.userID, phone)
	}
}

// handleMessage handles incoming messages
func (h *EventHandler) handleMessage(evt *events.Message) {
	// Skip messages from self
	if evt.Info.IsFromMe {
		return
	}

	h.logger.Info("message received", logger.Fields{
		"from":       evt.Info.Sender.User,
		"message_id": evt.Info.ID,
		"timestamp":  evt.Info.Timestamp,
	})

	// Extract message content
	content := ""
	messageType := "text"
	mediaURL := ""
	caption := ""

	if evt.Message.Conversation != nil {
		content = *evt.Message.Conversation
	} else if evt.Message.ExtendedTextMessage != nil && evt.Message.ExtendedTextMessage.Text != nil {
		content = *evt.Message.ExtendedTextMessage.Text
	} else if evt.Message.ImageMessage != nil {
		messageType = "image"
		if evt.Message.ImageMessage.Caption != nil {
			caption = *evt.Message.ImageMessage.Caption
		}
		if evt.Message.ImageMessage.URL != nil {
			mediaURL = *evt.Message.ImageMessage.URL
		}
	} else if evt.Message.VideoMessage != nil {
		messageType = "video"
		if evt.Message.VideoMessage.Caption != nil {
			caption = *evt.Message.VideoMessage.Caption
		}
		if evt.Message.VideoMessage.URL != nil {
			mediaURL = *evt.Message.VideoMessage.URL
		}
	} else if evt.Message.AudioMessage != nil {
		messageType = "audio"
		if evt.Message.AudioMessage.URL != nil {
			mediaURL = *evt.Message.AudioMessage.URL
		}
	} else if evt.Message.DocumentMessage != nil {
		messageType = "document"
		if evt.Message.DocumentMessage.FileName != nil {
			content = *evt.Message.DocumentMessage.FileName
		}
		if evt.Message.DocumentMessage.URL != nil {
			mediaURL = *evt.Message.DocumentMessage.URL
		}
	}

	// Create incoming message
	incomingMsg := IncomingMessage{
		MessageID:   evt.Info.ID,
		From:        evt.Info.Sender.User,
		To:          evt.Info.Chat.User,
		Content:     content,
		MessageType: messageType,
		MediaURL:    mediaURL,
		Caption:     caption,
		Timestamp:   evt.Info.Timestamp,
	}

	// Call the message handler
	if h.eventHandlers.OnMessage != nil {
		h.eventHandlers.OnMessage(h.sessionID, h.userID, incomingMsg)
	}
}

// handleReceipt handles message receipts (delivery/read status)
func (h *EventHandler) handleReceipt(evt *events.Receipt) {
	h.logger.Debug("receipt received", logger.Fields{
		"type":        evt.Type,
		"message_ids": evt.MessageIDs,
		"timestamp":   evt.Timestamp,
	})

	// TODO: Update message status in database based on receipt type
	// This will be implemented in the service layer
}

// handleLoggedOut handles the logged out event
func (h *EventHandler) handleLoggedOut(evt *events.LoggedOut) {
	h.logger.Warn("logged out from WhatsApp", logger.Fields{
		"reason": evt.Reason,
	})

	// Treat logout as disconnection
	if h.eventHandlers.OnDisconnected != nil {
		h.eventHandlers.OnDisconnected(h.sessionID, h.userID)
	}
}
