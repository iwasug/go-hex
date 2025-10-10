package entity

import "time"

// MessageDirection represents the direction of a WhatsApp message
type MessageDirection string

const (
	MessageDirectionInbound  MessageDirection = "inbound"
	MessageDirectionOutbound MessageDirection = "outbound"
)

// MessageStatus represents the status of a WhatsApp message
type MessageStatus string

const (
	MessageStatusPending   MessageStatus = "pending"
	MessageStatusSent      MessageStatus = "sent"
	MessageStatusDelivered MessageStatus = "delivered"
	MessageStatusRead      MessageStatus = "read"
	MessageStatusFailed    MessageStatus = "failed"
)

// WhatsAppMessage represents a WhatsApp message entity in the domain
type WhatsAppMessage struct {
	ID               string
	SessionID        string
	MessageID        string
	Direction        MessageDirection
	Status           MessageStatus
	From             string
	To               string
	Content          string
	MessageType      string
	Timestamp        time.Time
	WebhookDelivered bool
	WebhookAttempts  int
	ErrorMessage     string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// NewOutboundMessage creates a new outbound WhatsApp message
func NewOutboundMessage(sessionID, to, content, messageType string) *WhatsAppMessage {
	now := time.Now()
	return &WhatsAppMessage{
		SessionID:        sessionID,
		Direction:        MessageDirectionOutbound,
		Status:           MessageStatusPending,
		To:               to,
		Content:          content,
		MessageType:      messageType,
		Timestamp:        now,
		WebhookDelivered: false,
		WebhookAttempts:  0,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

// NewInboundMessage creates a new inbound WhatsApp message
func NewInboundMessage(sessionID, messageID, from, content, messageType string, timestamp time.Time) *WhatsAppMessage {
	now := time.Now()
	return &WhatsAppMessage{
		SessionID:        sessionID,
		MessageID:        messageID,
		Direction:        MessageDirectionInbound,
		Status:           MessageStatusDelivered,
		From:             from,
		Content:          content,
		MessageType:      messageType,
		Timestamp:        timestamp,
		WebhookDelivered: false,
		WebhookAttempts:  0,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

// UpdateStatus updates the message status
func (m *WhatsAppMessage) UpdateStatus(status MessageStatus) {
	m.Status = status
	m.UpdatedAt = time.Now()
}

// MarkAsSent marks the message as sent and sets the message ID
func (m *WhatsAppMessage) MarkAsSent(messageID string) {
	m.MessageID = messageID
	m.Status = MessageStatusSent
	m.UpdatedAt = time.Now()
}

// MarkAsDelivered marks the message as delivered
func (m *WhatsAppMessage) MarkAsDelivered() {
	m.Status = MessageStatusDelivered
	m.UpdatedAt = time.Now()
}

// MarkAsRead marks the message as read
func (m *WhatsAppMessage) MarkAsRead() {
	m.Status = MessageStatusRead
	m.UpdatedAt = time.Now()
}

// MarkAsFailed marks the message as failed with an error message
func (m *WhatsAppMessage) MarkAsFailed(errorMessage string) {
	m.Status = MessageStatusFailed
	m.ErrorMessage = errorMessage
	m.UpdatedAt = time.Now()
}

// SetFrom sets the sender phone number (for outbound messages after sending)
func (m *WhatsAppMessage) SetFrom(from string) {
	m.From = from
	m.UpdatedAt = time.Now()
}

// IncrementWebhookAttempts increments the webhook delivery attempt counter
func (m *WhatsAppMessage) IncrementWebhookAttempts() {
	m.WebhookAttempts++
	m.UpdatedAt = time.Now()
}

// MarkWebhookDelivered marks the webhook as successfully delivered
func (m *WhatsAppMessage) MarkWebhookDelivered() {
	m.WebhookDelivered = true
	m.UpdatedAt = time.Now()
}

// IsInbound checks if the message is inbound
func (m *WhatsAppMessage) IsInbound() bool {
	return m.Direction == MessageDirectionInbound
}

// IsOutbound checks if the message is outbound
func (m *WhatsAppMessage) IsOutbound() bool {
	return m.Direction == MessageDirectionOutbound
}

// IsPending checks if the message is pending
func (m *WhatsAppMessage) IsPending() bool {
	return m.Status == MessageStatusPending
}

// IsFailed checks if the message has failed
func (m *WhatsAppMessage) IsFailed() bool {
	return m.Status == MessageStatusFailed
}

// ShouldRetryWebhook checks if webhook delivery should be retried
func (m *WhatsAppMessage) ShouldRetryWebhook(maxAttempts int) bool {
	return !m.WebhookDelivered && m.WebhookAttempts < maxAttempts
}
