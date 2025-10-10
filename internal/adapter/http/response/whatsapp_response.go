package response

import "time"

// WhatsAppSessionResponse represents a WhatsApp session in API responses
type WhatsAppSessionResponse struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	SessionName   string    `json:"session_name"`
	PhoneNumber   string    `json:"phone_number,omitempty"`
	Status        string    `json:"status"`
	QRCode        string    `json:"qr_code,omitempty"`
	QRCodeExpiry  *time.Time `json:"qr_code_expiry,omitempty"`
	DeviceInfo    string    `json:"device_info,omitempty"`
	LastActivity  time.Time `json:"last_activity"`
	WebhookURL    string    `json:"webhook_url,omitempty"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// QRCodeResponse represents a QR code response
type QRCodeResponse struct {
	SessionID string     `json:"session_id"`
	QRCode    string     `json:"qr_code"`
	ExpiresAt time.Time  `json:"expires_at"`
	Status    string     `json:"status"`
}

// WhatsAppMessageResponse represents a WhatsApp message in API responses
type WhatsAppMessageResponse struct {
	ID               string    `json:"id"`
	SessionID        string    `json:"session_id"`
	MessageID        string    `json:"message_id,omitempty"`
	Direction        string    `json:"direction"`
	Status           string    `json:"status"`
	From             string    `json:"from"`
	To               string    `json:"to"`
	Content          string    `json:"content"`
	MessageType      string    `json:"message_type"`
	Timestamp        time.Time `json:"timestamp"`
	WebhookDelivered bool      `json:"webhook_delivered"`
	WebhookAttempts  int       `json:"webhook_attempts"`
	ErrorMessage     string    `json:"error_message,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// MessageSentResponse represents the response after sending a message
type MessageSentResponse struct {
	MessageID string    `json:"message_id"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Message   WhatsAppMessageResponse `json:"message"`
}

// SessionStatusResponse represents the status of a WhatsApp session
type SessionStatusResponse struct {
	SessionID      string    `json:"session_id"`
	Status         string    `json:"status"`
	PhoneNumber    string    `json:"phone_number,omitempty"`
	ConnectedPhone string    `json:"connected_phone,omitempty"`
	LastActivity   time.Time `json:"last_activity"`
	IsActive       bool      `json:"is_active"`
	IsConnected    bool      `json:"is_connected"`
	IsLoggedIn     bool      `json:"is_logged_in"`
}
